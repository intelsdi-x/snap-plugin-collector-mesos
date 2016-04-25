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
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

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

	host, port, err := extractHostAndPortFromURL(ts.URL)
	if err != nil {
		panic(err)
	}

	Convey("Get metrics snapshot from the master", t, func() {
		res, err := GetMetricsSnapshot(host, port)

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
		host, port, err := extractHostAndPortFromURL(ts1.URL)
		if err != nil {
			panic(err)
		}

		Convey("No error should be reported", func() {
			_, err := IsLeader(host, port)
			So(err, ShouldBeNil)
		})

		Convey("Should return true when leading", func() {
			hostIsLeader, err := IsLeader(host, port)
			So(hostIsLeader, ShouldBeTrue)
			So(err, ShouldBeNil)
		})

		Convey("Should return false when not leading", func() {
			host, port, err := extractHostAndPortFromURL(ts2.URL)
			if err != nil {
				panic(err)
			}

			hostIsLeader, err := IsLeader(host, port)
			So(hostIsLeader, ShouldBeFalse)
			So(err, ShouldBeNil)
		})
	})
}

func extractHostAndPortFromURL(u string) (string, int, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", 0, err
	}

	host, p, _ := net.SplitHostPort(parsed.Host)
	port, err := strconv.Atoi(p)
	if err != nil {
		return "", 0, err
	}

	return host, port, nil
}
