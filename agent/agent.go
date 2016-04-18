// +build linux

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


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
	ID         string
	Name       string
	Source     string
	Framework  string
	Statistics map[string]interface{}
}

func GetStats(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(resp.Status)
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return contents, nil
}

func ParseStats(input []byte) []Executor {
	jsonOutput := []map[string]interface{}{}
	if err := json.Unmarshal(input, &jsonOutput); err != nil {
		panic(err)
	}

	executors := []Executor{}
	for _, o := range jsonOutput {
		e := Executor{
			ID:         o["executor_id"].(string),
			Name:       o["executor_name"].(string),
			Source:     o["source"].(string),
			Framework:  o["framework_id"].(string),
			Statistics: o["statistics"].(map[string]interface{}),
		}
		executors = append(executors, e)
	}
	return executors
}
