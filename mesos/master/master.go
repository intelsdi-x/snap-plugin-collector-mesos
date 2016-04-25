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

package master

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Collect metrics from the '/metrics/snapshot' endpoint on the master.  The '/metrics/snapshot' endpoint returns JSON,
// and all metrics contained in the endpoint use a string as the key, and a double (float64) for the value. For example:
//
//   {
//     "master/cpus_total": 2.0
//   }
//
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

// Determine if a given host is currently the leader, based on the location provided by the '/master/redirect' endpoint.
func IsLeader(host string) (bool, error) {
	req, err := http.NewRequest("HEAD", "http://"+host+"/master/redirect", nil)
	if err != nil {
		return false, fmt.Errorf("request error: %s", err)
	}

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		if resp.StatusCode == 307 {
			// Do nothing, this is expected
		} else if resp.StatusCode != 307 {
			return false, fmt.Errorf("error: expected HTTP 307, got %d", resp.StatusCode)
		} else {
			return false, fmt.Errorf("client error: %s", err)
		}
	}

	location, err := resp.Location()
	if err != nil {
		return false, err
	}

	if strings.Contains(location.Host, host) {
		return true, nil
	} else {
		return false, nil
	}
}
