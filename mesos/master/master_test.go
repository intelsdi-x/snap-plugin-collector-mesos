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

package master

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetFrameworks(t *testing.T) {
	testData := Frameworks{ActiveFrameworks: []*Framework{
		&Framework{
			ID: "id1",
			OfferedResources: &Resources{
				CPUs: 1.0,
				Mem:  1024.0,
				Disk: 512.0,
			},
			Resources: &Resources{
				CPUs: 1.0,
				Mem:  1024.0,
				Disk: 512.0,
			},
			UsedResources: &Resources{
				CPUs: 1.0,
				Mem:  1024.0,
				Disk: 512.0,
			},
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

	Convey("When framework resource utilization is requested", t, func() {
		frameworks, err := GetFrameworks(host)

		Convey("Then no error should be reported", func() {
			So(err, ShouldBeNil)
		})

		Convey("Then list of frameworks is returned", func() {
			So(frameworks, ShouldNotBeNil)
			So(len(frameworks), ShouldEqual, len(testData.ActiveFrameworks))
		})

		Convey("Then proper stats are returned", func() {
			for _, framework := range frameworks {
				switch framework.ID {
				case "id1":
					So(framework.Resources.CPUs, ShouldEqual, 1.0)
					So(framework.Resources.Disk, ShouldEqual, 512.0)
					So(framework.Resources.Mem, ShouldEqual, 1024.0)
					So(framework.OfferedResources.CPUs, ShouldEqual, 1.0)
					So(framework.OfferedResources.Disk, ShouldEqual, 512.0)
					So(framework.OfferedResources.Mem, ShouldEqual, 1024.0)
					So(framework.UsedResources.CPUs, ShouldEqual, 1.0)
					So(framework.UsedResources.Disk, ShouldEqual, 512.0)
					So(framework.UsedResources.Mem, ShouldEqual, 1024.0)

				}
			}
		})
	})
}

func TestGetFrameworksMetricTypes(t *testing.T) {
	Convey("When building metric types for Frameworks on the master", t, func() {
		namespaces, err := GetFrameworksMetricTypes()
		Convey("No errors should be reported", func() {
			So(err, ShouldBeNil)
		})
		Convey("Valid namespace parts should be returned as a slice of strings", func() {
			So(len(namespaces), ShouldBeGreaterThan, 0)
			So(namespaces, ShouldContain, "resources/disk")
			So(namespaces, ShouldNotContain, "resources/foo")
		})
		Convey("Should not contain non-metrics namespaces, e.g. 'id'", func() {
			So(namespaces, ShouldNotContain, "id")
		})
	})
}

func TestGetMetricsSnapshot(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		td, err := json.Marshal(map[string]float64{
			"allocator/event_queue_dispatches": 0.0,
			"master/cpus_percent":              0.0,
			"registrar/queued_operations":      0.0,
			"system/cpus_total":                2.0,
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
			So(len(res), ShouldEqual, 4)
			So(res["system/cpus_total"], ShouldEqual, 2.0)
			So(err, ShouldBeNil)
		})
	})
}

func TestIsLeader(t *testing.T) {

	// ts1 simulates a host that is the leader
	ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", r.URL.String())
		w.WriteHeader(307)
	}))

	// ts2 simulates a host that is not the leader
	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "//mesos-master-2.example.com:5050")
		w.WriteHeader(307)
	}))

	defer ts1.Close()
	defer ts2.Close()

	Convey("Determine if master is leader", t, func() {
		host, err := extractHostFromURL(ts1.URL)
		if err != nil {
			panic(err)
		}

		Convey("No error should be reported", func() {
			_, err := IsLeader(host)
			So(err, ShouldBeNil)
		})

		Convey("Should return true when leading", func() {
			hostIsLeader, err := IsLeader(host)
			So(hostIsLeader, ShouldBeTrue)
			So(err, ShouldBeNil)
		})

		Convey("Should return false when not leading", func() {
			host, err := extractHostFromURL(ts2.URL)
			if err != nil {
				panic(err)
			}

			hostIsLeader, err := IsLeader(host)
			So(hostIsLeader, ShouldBeFalse)
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
