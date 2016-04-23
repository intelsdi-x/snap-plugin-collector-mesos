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
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

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
		u, err := url.Parse(ts1.URL)
		if err != nil {
			panic(err)
		}

		host, p, _ := net.SplitHostPort(u.Host)
		port, err := strconv.Atoi(p)
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
			u, err := url.Parse(ts2.URL)
			if err != nil {
				panic(err)
			}

			host, p, _ := net.SplitHostPort(u.Host)
			port, err := strconv.Atoi(p)
			if err != nil {
				panic(err)
			}

			hostIsLeader, err := IsLeader(host, port)
			So(hostIsLeader, ShouldBeFalse)
			So(err, ShouldBeNil)
		})
	})
}
