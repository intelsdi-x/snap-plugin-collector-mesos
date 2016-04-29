// +build unit

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
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetMetricsSnapshot(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		td, err := json.Marshal(map[string]float64{
			"containerizer/mesos/container_destroy_errors": 0.0,
			"slave/cpus_percent":                           0.0,
			"system/cpus_total":                            2.0,
		})
		if err != nil {
			panic(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(td)
	}))
	defer ts.Close()

	host, err := extractHostFromURL(ts.URL)
	if err != nil {
		panic(err)
	}

	Convey("Get metrics snapshot from the master", t, func() {
		res, err := GetMetricsSnapshot(host)

		Convey("Should return a map of metrics", func() {
			So(len(res), ShouldEqual, 3)
			So(res["system/cpus_total"], ShouldEqual, 2.0)
			So(err, ShouldBeNil)
		})
	})
}

func extractHostFromURL(u string) (string, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	return parsed.Host, nil
}
