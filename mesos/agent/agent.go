/*
Copyright 2016 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package agent

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/intelsdi-x/snap-plugin-collector-mesos/mesos/client"
	"github.com/intelsdi-x/snap-plugin-collector-mesos/mesos/mesos_pb2"
	"github.com/intelsdi-x/snap-plugin-utilities/ns"
	"github.com/intelsdi-x/snap-plugin-utilities/str"
	log "github.com/sirupsen/logrus"
)

// The "/monitor/statistics" endpoint returns an array of JSON objects. Its top-level structure isn't defined by a
// protobuf, but the "statistics" object (and everything under it) is. For the actual Mesos implementation, see
// https://github.com/apache/mesos/blob/0.28.1/src/slave/monitor.cpp#L130-L148
type Executor struct {
	ID         string                        `json:"executor_id"`
	Name       string                        `json:"executor_name"`
	Source     string                        `json:"source"`
	Framework  string                        `json:"framework_id"`
	Statistics *mesos_pb2.ResourceStatistics `json:"statistics"`
}

// The "/slave(1)/flags" endpoint on a Mesos agent returns an object that contains a single object "flags".
type Flags struct {
	Flags map[string]string
}

// Get the configuration flags from the Mesos agent and return them as a map.
func GetFlags(host string) (map[string]string, error) {
	log.Debug("Getting configuration flags from host ", host)
	flags := &Flags{}

	c := client.NewClient(host, "/slave(1)/flags", time.Duration(5))
	if err := c.Fetch(&flags); err != nil {
		log.Error(err)
		return nil, err
	}

	return flags.Flags, nil
}

// Collect metrics from the '/metrics/snapshot' endpoint on the agent.  The '/metrics/snapshot' endpoint returns JSON,
// and all metrics contained in the endpoint use a string as the key, and a double (float64) for the value. For example:
//
//   {
//     "slave/cpus_total": 2.0
//   }
//
// Note that, as of Mesos 0.28.x, "slave" is being renamed to "agent" and this effort isn't yet complete. For more
// information, see https://issues.apache.org/jira/browse/MESOS-1478.
func GetMetricsSnapshot(host string) (map[string]float64, error) {
	log.Debug("Getting metrics snapshot from host ", host)
	data := map[string]float64{}

	c := client.NewClient(host, "/metrics/snapshot", time.Duration(5))
	if err := c.Fetch(&data); err != nil {
		log.Error(err)
		return nil, err
	}

	return data, nil
}

// Collect metrics from the '/monitor/statistics' endpoint on the agent. This endpoint returns JSON, and all metrics
// contained in the endpoint use a string as the key. Depending on features enabled on the Mesos agent, additional
// metrics might be available under either the "statistics" object, or additional nested objects (e.g. "perf") as
// defined by the Executor structure, and the structures in mesos_pb2.ResourceStatistics.
func GetMonitoringStatistics(host string) ([]Executor, error) {
	log.Debug("Getting monitoring statistics from host ", host)
	var executors []Executor

	c := client.NewClient(host, "/monitor/statistics", time.Duration(30))
	if err := c.Fetch(&executors); err != nil {
		log.Error(err)
		return nil, err
	}

	return executors, nil
}

// Recursively traverse the Executor struct, building "/"-delimited strings that resemble snap metric types. If a given
// feature is not enabled on a Mesos agent (e.g. the network isolator), then those metrics will be removed from the
// metric types returned by this function.
func GetMonitoringStatisticsMetricTypes(host string) ([]string, error) {
	log.Debug("Getting monitoring statistics metrics type from host ", host)
	// TODO(roger): supporting NetTrafficControlStatistics means adding another dynamic metric to the plugin.
	// When we're ready to do this, remove ns.InspectEmptyContainers(ns.AlwaysFalse) so this defaults to true.
	namespaces := []string{}
	err := ns.FromCompositeObject(
		&mesos_pb2.ResourceStatistics{}, "", &namespaces, ns.InspectEmptyContainers(ns.AlwaysFalse))
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// Avoid returning a metric type that is impossible to collect on this system
	flags, err := GetFlags(host)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// Isolators are defined using a comma-separated string passed to the "--isolation" flag on the Mesos agent.
	// See https://github.com/apache/mesos/blob/0.28.1/src/slave/containerizer/mesos/containerizer.cpp#L196-L223
	isolators := strings.Split(flags["isolation"], ",")

	if str.Contains(isolators, "cgroups/perf_event") {
		log.Debug("Isolator cgroups/perf_event is enabled on host ", host)
		// Expects a perf event from the output of `perf list`. Mesos then normalizes the event name. See
		// https://github.com/apache/mesos/blob/0.28.1/src/linux/perf.cpp#L65-L71
		namespaces = deleteFromSlice(namespaces, "^perf/.*")
		var normalizedPerfEvents []string
		perfEvents := strings.Split(flags["perf_events"], ",")

		for _, event := range perfEvents {
			log.Debug("Adding perf event ", event, " to metrics catalog")
			event = fmt.Sprintf("perf/%s", normalizePerfEventName(event))
			normalizedPerfEvents = append(normalizedPerfEvents, event)
		}
		namespaces = append(namespaces, normalizedPerfEvents...)
	} else {
		log.Debug("Isolator cgroups/perf_event is not enabled on host ", host)
		namespaces = deleteFromSlice(namespaces, "^perf.*")
	}

	if !str.Contains(isolators, "posix/disk") {
		log.Debug("Isolator posix/disk is not enabled on host ", host)
		namespaces = deleteFromSlice(namespaces, "^disk_.*")
	}

	if !str.Contains(isolators, "network/port_mapping") {
		log.Debug("Isolator network/port_mapping is not enabled on host ", host)
		namespaces = deleteFromSlice(namespaces, "^net_.*")
	}

	return namespaces, nil
}

// Normalizes a perf event, based on https://github.com/apache/mesos/blob/0.28.1/src/linux/perf.cpp#L65-L71
func normalizePerfEventName(s string) string {
	normalized := strings.ToLower(s)
	return strings.Replace(normalized, "-", "_", -1)
}

// Delete a given string from a slice, without returning a new slice. Regex is allowed (e.g. '^perf/.*').
func deleteFromSlice(a []string, s string) []string {
	for i := 0; i < len(a); i++ {
		matched, _ := regexp.MatchString(s, a[i])
		if matched {
			a = append(a[:i], a[i+1:]...)
			i--
		}
	}
	return a
}
