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
	"time"

	"github.com/intelsdi-x/snap-plugin-collector-mesos/mesos/client"
)

type Executor struct {
	ID         string                 `json:"executor_id"`
	Name       string                 `json:"executor_name"`
	Source     string                 `json:"source"`
	Framework  string                 `json:"framework_id"`
	Statistics map[string]interface{} `json:"statistics"`
}

func (e *Executor) GetExecutorStatistic(stat string) (float64, error) {
	if val, ok := e.Statistics[stat]; ok {
		return val.(float64), nil
	} else if perf, ok := e.Statistics["perf"]; ok {
		if val, ok := perf.(map[string]interface{})[stat]; ok {
			return val.(float64), nil
		}
	}
	return 0, fmt.Errorf("Requested stat %s is not available for %s", stat, e.ID)
}

func GetAgentStatistics(host string) ([]Executor, error) {
	var executors []Executor

	c := client.NewClient(host, "/monitor/statistics", time.Duration(30))
	if err := c.Fetch(&executors); err != nil {
		return nil, err
	}

	return executors, nil
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
