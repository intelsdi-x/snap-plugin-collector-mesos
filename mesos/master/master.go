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
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/intelsdi-x/snap-plugin-collector-mesos/mesos/client"
	"github.com/intelsdi-x/snap-plugin-utilities/ns"
)

type Frameworks struct {
	ActiveFrameworks []*Framework `json:"frameworks"`
}

type Framework struct {
	ID               string     `json:"id"`
	OfferedResources *Resources `json:"offered_resources"`
	Resources        *Resources `json:"resources"`
	UsedResources    *Resources `json:"used_resources"`
}

type Resources struct {
	CPUs float64 `json:"cpus"`
	Disk float64 `json:"disk"`
	Mem  float64 `json:"mem"`
}

// Recursively traverse the Frameworks struct, building "/"-delimited strings that resemble snap metric types.
func GetFrameworksMetricTypes() ([]string, error) {
	namespaces := []string{}
	if err := ns.FromCompositeObject(Framework{}, "", &namespaces); err != nil {
		return nil, err
	}
	for i := 0; i < len(namespaces); i++ {
		if namespaces[i] == "id" {
			namespaces = append(namespaces[:i], namespaces[i+1:]...)
			break
		}
	}
	return namespaces, nil

}

// Get metrics from the '/master/frameworks' endpoint on the master. This endpoint returns JSON about the overall
// state and resource utilization of the frameworks running on the cluster.
func GetFrameworks(host string) ([]*Framework, error) {
	var frameworks Frameworks

	c := client.NewClient(host, "/master/frameworks", time.Duration(10))
	if err := c.Fetch(&frameworks); err != nil {
		return nil, err
	}

	return frameworks.ActiveFrameworks, nil
}

// Collect metrics from the '/metrics/snapshot' endpoint on the master.  The '/metrics/snapshot' endpoint returns JSON,
// and all metrics contained in the endpoint use a string as the key, and a double (float64) for the value. For example:
//
//   {
//     "master/cpus_total": 2.0
//   }
//
func GetMetricsSnapshot(host string) (map[string]float64, error) {
	data := map[string]float64{}

	c := client.NewClient(host, "/metrics/snapshot", time.Duration(5))
	if err := c.Fetch(&data); err != nil {
		return nil, err
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
