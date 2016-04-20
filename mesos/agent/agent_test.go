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
	"fmt"
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

var server *httptest.Server = httptest.NewServer(
	AgentStatisticsHandler(testData),
)

var testData []Executor = []Executor{
	Executor{
		ID:        "id1",
		Name:      "name1",
		Source:    "source1",
		Framework: "frame1",
		Statistics: map[string]interface{}{
			"stat_a": 1.1,
			"stat_b": 2.2,
			"perf": map[string]interface{}{
				"perf_1": 3.3,
			},
		},
	},
	Executor{
		ID:        "id2",
		Name:      "name2",
		Source:    "source2",
		Framework: "frame2",
		Statistics: map[string]interface{}{
			"stat_c": 4.4,
			"stat_d": 5.5,
		},
	},
}

func AgentStatisticsHandler(executors []Executor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		rendered, err := json.Marshal(executors)
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(w, string(rendered))
	}
}

func TestGetAgentStatistics(t *testing.T) {
	Convey("When mesos agent statistics are requested", t, func() {
		execs, err := GetAgentStatistics(server.URL)

		Convey("Then no error should be reporeted", func() {
			So(err, ShouldBeNil)
		})

		Convey("Then list of executors is returned", func() {
			So(execs, ShouldNotBeNil)
			So(len(execs), ShouldEqual, len(testData))

			for i, exec := range execs {
				So(exec.ID, ShouldEqual, testData[i].ID)
				So(exec.Name, ShouldEqual, testData[i].Name)
				So(exec.Framework, ShouldEqual, testData[i].Framework)
				So(exec.Source, ShouldEqual, testData[i].Source)
			}
		})

		Convey("Then proper stats are set", func() {
			So(execs[0].Statistics["stat_a"], ShouldEqual, 1.1)
			So(execs[0].Statistics["stat_b"], ShouldEqual, 2.2)
			So(execs[1].Statistics["stat_c"], ShouldEqual, 4.4)
			So(execs[1].Statistics["stat_d"], ShouldEqual, 5.5)
		})
	})
}

func TestGetExecutorStatistics(t *testing.T) {
	tcs := []struct {
		Executor Executor
		Stat     string
		Expected float64
		Error    error
	}{
		{testData[0], "stat_a", 1.1, nil},
		{testData[0], "stat_b", 2.2, nil},
		{testData[0], "perf_1", 3.3, nil},
		{testData[1], "stat_c", 4.4, nil},
		{testData[1], "stat_d", 5.5, nil},
		{testData[0], "stat_c", 0, fmt.Errorf("Requested stat %s is not available for %s", "stat_c", testData[0].ID)},
	}

	for _, tc := range tcs {
		Convey("When executor statistics are requested", t, func() {
			value, err := tc.Executor.GetExecutorStatistic(tc.Stat)

			Convey("Then proper value is returned", func() {
				if tc.Error == nil {
					So(err, ShouldEqual, tc.Error)
				} else {
					So(err.Error(), ShouldEqual, tc.Error.Error())
				}
				So(value, ShouldEqual, tc.Expected)
			})
		})
	}
}
