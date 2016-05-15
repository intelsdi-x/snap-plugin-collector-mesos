// +build integration

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

package mesos

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/intelsdi-x/snap-plugin-utilities/config"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMesos_GetMetricTypes(t *testing.T) {
	cfg := setupCfg()
	mc := NewMesosCollector()
	mts, err := mc.GetMetricTypes(cfg)
	if err != nil {
		panic(err)
	}

	Convey("Should return metric types for the '/metrics/snapshot' endpoint", t, func() {
		So(err, ShouldBeNil)
		So(len(mts), ShouldBeGreaterThan, 1)
	})

	Convey("Should return metric types for the '/monitor/statistics' endpoint", t, func() {
		So(err, ShouldBeNil)
		So(len(mts), ShouldBeGreaterThan, 1)
	})
}

func TestMesos_CollectMetrics(t *testing.T) {
	cfg := setupCfg()
	master, err := config.GetConfigItem(cfg, "master")
	if err != nil {
		panic(err)
	}

	go launchTask(master.(string))
	time.Sleep(time.Duration(10)) // TODO(roger): do a status check instead of sleeping for an arbitrary duration

	Convey("Collect metrics from a Mesos master and agent", t, func() {
		mc := NewMesosCollector()
		mc.GetMetricTypes(cfg)

		Convey("Should collect requested metrics from the master", func() {
			mts := []plugin.MetricType{
				plugin.MetricType{
					Namespace_: core.NewNamespace("intel", "mesos", "master", "master", "tasks_running"),
					Config_:    cfg.ConfigDataNode,
				},
				plugin.MetricType{
					Namespace_: core.NewNamespace("intel", "mesos", "master", "registrar", "state_store_ms", "p99"),
					Config_:    cfg.ConfigDataNode,
				},
				plugin.MetricType{
					Namespace_: core.NewNamespace("intel", "mesos", "master", "system", "load_5min"),
					Config_:    cfg.ConfigDataNode,
				},
			}

			metrics, err := mc.CollectMetrics(mts)
			So(err, ShouldBeNil)
			So(metrics, ShouldNotBeNil)
			So(len(metrics), ShouldEqual, 3)
		})

		// NOTE: in future versions of Mesos, the term "slave" will change to "agent". This has the potential
		// to break this test if the Mesos version is bumped in CI and this test isn't updated at the same time.
		Convey("Should collect requested metrics from the agent", func() {
			mts := []plugin.MetricType{
				plugin.MetricType{
					Namespace_: core.NewNamespace("intel", "mesos", "agent", "slave", "tasks_running"),
					Config_:    cfg.ConfigDataNode,
				},
				plugin.MetricType{
					Namespace_: core.NewNamespace("intel", "mesos", "agent", "system", "load_5min"),
					Config_:    cfg.ConfigDataNode,
				},
				plugin.MetricType{
					Namespace_: core.NewNamespace("intel", "mesos", "agent", "*", "*", "mem_total_bytes"),
					Config_:    cfg.ConfigDataNode,
				},
			}

			metrics, err := mc.CollectMetrics(mts)
			So(err, ShouldBeNil)
			So(metrics, ShouldNotBeNil)
			So(len(metrics), ShouldEqual, 3)
		})

		Convey("Should return an error if an invalid metric was requested", func() {
			mts := []plugin.MetricType{
				plugin.MetricType{
					Namespace_: core.NewNamespace("intel", "mesos", "master", "foo", "bar", "baz"),
					Config_:    cfg.ConfigDataNode,
				},
			}

			metrics, err := mc.CollectMetrics(mts)
			So(metrics, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
	})
}

// setupCfg builds a new ConfigDataNode that specifies the Mesos master and agent host / port
// to use in the integration test(s).
func setupCfg() plugin.ConfigType {
	master := os.Getenv("SNAP_MESOS_MASTER")
	if master == "" {
		master = "127.0.0.1:5050"
	}

	agent := os.Getenv("SNAP_MESOS_AGENT")
	if agent == "" {
		agent = "127.0.0.1:5051"
	}

	node := cdata.NewNode()
	node.AddItem("master", ctypes.ConfigValueStr{Value: master})
	node.AddItem("agent", ctypes.ConfigValueStr{Value: agent})

	return plugin.ConfigType{ConfigDataNode: node}

}

// Launch a Mesos task
func launchTask(master string) {
	cmd := "mesos"
	id := time.Now().Unix()
	args := []string{
		"execute", fmt.Sprintf("--master=%s", master),
		fmt.Sprintf("--name=%v", id),
		"--command=sleep 60",
	}

	if err := exec.Command(cmd, args...).Run(); err != nil {
		panic(err)
	}
}
