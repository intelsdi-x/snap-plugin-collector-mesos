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
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap-plugin-collector-mesos/mesos/client"
)

// The "/monitor/statistics" endpoint returns an array of JSON objects. Its top-level structure isn't defined by a
// protobuf, but the "statistics" object (and everything under it) is. For the actual Mesos implementation, see
// https://github.com/apache/mesos/blob/0.28.1/src/slave/monitor.cpp#L130-L148
type Executor struct {
	ID         string                 `json:"executor_id"`
	Name       string                 `json:"executor_name"`
	Source     string                 `json:"source"`
	Framework  string                 `json:"framework_id"`
	Statistics map[string]interface{} `json:"statistics"`
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
