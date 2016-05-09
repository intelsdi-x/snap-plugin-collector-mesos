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
	"reflect"
	"strings"
	"time"

	"github.com/intelsdi-x/snap-plugin-collector-mesos/mesos/client"
	"github.com/intelsdi-x/snap-plugin-collector-mesos/mesos/mesos_pb2"
)

// The "/monitor/statistics" endpoint returns an array of JSON objects. Its top-level structure isn't defined by a
// protobuf, but the "statistics" object (and everything under it) is. For the actual Mesos implementation, see
// https://github.com/apache/mesos/blob/0.28.1/src/slave/monitor.cpp#L130-L148
type Executor struct {
	ID         string `json:"executor_id"`
	Name       string `json:"executor_name"`
	Source     string `json:"source"`
	Framework  string `json:"framework_id"`
	Statistics *mesos_pb2.ResourceStatistics
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
	data := map[string]float64{}

	c := client.NewClient(host, "/metrics/snapshot", time.Duration(5))
	if err := c.Fetch(&data); err != nil {
		return nil, err
	}

	return data, nil
}

// Collect metrics from the '/monitor/statistics' endpoint on the agent. This endpoint returns JSON, and all metrics
// contained in the endpoint use a string as the key. Depending on features enabled on the Mesos agent, additional
// metrics might be available under either the "statistics" object, or additional nested objects (e.g. "perf") as
// defined by the Executor structure, and the structures in mesos_pb2.ResourceStatistics.
func GetMonitoringStatistics(host string) ([]Executor, error) {
	var executors []Executor

	c := client.NewClient(host, "/monitor/statistics", time.Duration(30))
	if err := c.Fetch(&executors); err != nil {
		return nil, err
	}

	return executors, nil
}

// Recursively traverse the Executor struct, building "/"-delimited strings that resemble snap metric types.
func GetMonitoringStatisticsMetricTypes() []string {
	namespaces := []string{}

	// To prevent reflect from returning nil, define a skeleton of all possible nested structs
	// TODO(roger): is it possible to (easily) query the Mesos agent for enabled features?
	e := &Executor{
		Statistics: &mesos_pb2.ResourceStatistics{
			// TODO(roger): implement NetSnmpStatistics and NetTrafficControlStatistics in a future version
			Perf: &mesos_pb2.PerfStatistics{},
		},
	}

	var buildNamespaceRecursively func(ns string, v reflect.Value)
	buildNamespaceRecursively = func(ns string, v reflect.Value) {
		switch v.Kind() {
		case reflect.Ptr:
			buildNamespaceRecursively(ns, v.Elem())
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				fieldInfo := v.Type().Field(i)
				tag := strings.Split(fieldInfo.Tag.Get("json"), ",")[0] // ignore "omitempty" if it exists
				if tag == "-" {
					continue
				}
				if tag == "" {
					tag = strings.ToLower(fieldInfo.Name)
				}

				// Only consider valid metric namespaces. For example: "/statistics" should not be
				// considered, because it's a pointer to mesos_pb2.ResourceStatistics, but
				// "/statistics/cpus_user_time_secs" should be considered because it's a float64.
				f := fieldInfo.Type
				if v.Field(i).Kind() == reflect.Ptr {
					f = fieldInfo.Type.Elem()
				}

				nsNext := ""
				if ns == "" {
					nsNext = tag
				} else {
					nsNext = strings.Join([]string{ns, tag}, "/")
				}

				switch f.Kind() {
				case reflect.Uint32, reflect.Uint64, reflect.Float64:
					namespaces = append(namespaces, nsNext)
				}

				buildNamespaceRecursively(nsNext, v.Field(i))
			}
		}
	}

	buildNamespaceRecursively("", reflect.ValueOf(e))
	return namespaces
}
