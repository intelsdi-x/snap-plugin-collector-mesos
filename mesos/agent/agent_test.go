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

	"github.com/golang/protobuf/proto"
	"github.com/intelsdi-x/snap-plugin-collector-mesos/mesos/mesos_pb2"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetFlags(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testData := Flags{
			Flags: map[string]string{
				"isolation": "cgroups/cpu,cgroups/mem",
				"port":      "5051",
			},
		}
		td, err := json.Marshal(testData)
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

	Convey("When getting flags from the Mesos agent", t, func() {
		flags, err := GetFlags(host)

		Convey("Should return a map of the configuration flags", func() {
			So(err, ShouldBeNil)
			So(flags["isolation"], ShouldEqual, "cgroups/cpu,cgroups/mem")
			So(flags["port"], ShouldEqual, "5051")
			So(flags["perf_events"], ShouldEqual, "")
		})
	})
}

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

func TestGetMonitoringStatistics(t *testing.T) {
	testData := []Executor{
		Executor{
			ID:        "id1",
			Name:      "name1",
			Source:    "source1",
			Framework: "frame1",
			Statistics: &mesos_pb2.ResourceStatistics{
				CpusLimit:     proto.Float64(1.1),
				MemTotalBytes: proto.Uint64(1000),
				Perf: &mesos_pb2.PerfStatistics{
					ContextSwitches: proto.Uint64(10),
				},
			},
		},
		Executor{
			ID:        "id2",
			Name:      "name2",
			Source:    "source2",
			Framework: "frame2",
			Statistics: &mesos_pb2.ResourceStatistics{
				CpusLimit:     proto.Float64(1.1),
				MemTotalBytes: proto.Uint64(2000),
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		td, err := json.Marshal(testData)
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

	Convey("When monitoring statistics are requested", t, func() {
		execs, err := GetMonitoringStatistics(host)

		Convey("Then no error should be reported", func() {
			So(err, ShouldBeNil)
		})

		Convey("Then list of executors is returned", func() {
			So(execs, ShouldNotBeNil)
			So(len(execs), ShouldEqual, len(testData))
		})

		Convey("Then proper stats are returned", func() {
			for _, exec := range execs {
				switch exec.ID {
				case "id1":
					So(*exec.Statistics.CpusLimit, ShouldEqual, 1.1)
					So(*exec.Statistics.MemTotalBytes, ShouldEqual, 1000)
					So(*exec.Statistics.Perf.ContextSwitches, ShouldEqual, 10)
				case "id2":
					So(*exec.Statistics.CpusLimit, ShouldEqual, 1.1)
					So(*exec.Statistics.MemTotalBytes, ShouldEqual, 2000)
				}
			}
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
