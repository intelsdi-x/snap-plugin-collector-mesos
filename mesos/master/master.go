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
	"net"
	"net/http"
	"strconv"
	"strings"
)

// Determine if a given host is currently the leader, based on the location provided by the '/master/redirect' endpoint.
func IsLeader(host string, port int) (bool, error) {
	master := net.JoinHostPort(host, strconv.Itoa(port))

	req, err := http.NewRequest("HEAD", "http://"+master+"/master/redirect", nil)
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

	if strings.Contains(location.Host, master) {
		return true, nil
	} else {
		return false, nil
	}
}
