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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

func GetAgentStatistics(url string) ([]Executor, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(resp.Status)
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var executors []Executor
	err = json.Unmarshal(content, &executors)
	if err != nil {
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

	// TODO(roger): abstract the http client for consistent use throughout this plugin
	url := "http://" + host + "/metrics/snapshot"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return data, fmt.Errorf("fetch error: %s", resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return data, fmt.Errorf("read error: %s: %v\n", url, err)
	}

	if err := json.Unmarshal(b, &data); err != nil {
		return data, fmt.Errorf("unmarshal error: %s: %v\n", b, err)
	}

	return data, nil
}
