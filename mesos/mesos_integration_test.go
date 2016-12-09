// +build medium

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
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/intelsdi-x/snap-plugin-collector-mesos/mesos/agent"
	"github.com/intelsdi-x/snap-plugin-collector-mesos/mesos/master"
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

	var namespaces []string
	for i := 0; i < len(mts); i++ {
		namespaces = append(namespaces, mts[i].Namespace().String())
	}

	Convey("Should return metric types for the master and agent", t, func() {
		So(err, ShouldBeNil)
		So(len(mts), ShouldBeGreaterThan, 1)
		So(namespaces, ShouldContain, "/intel/mesos/master/system/cpus_total")
		So(namespaces, ShouldContain, "/intel/mesos/master/*/resources/cpus")
		So(namespaces, ShouldContain, "/intel/mesos/agent/system/cpus_total")
		So(namespaces, ShouldContain, "/intel/mesos/agent/*/*/cpus_limit")
	})

	// In both the Vagrant and Travis provisioning scripts, "network/port_mapping" is not
	// enabled because Mesos hasn't been compiled with that isolator. If that changes, this
	// test case will need to be updated.
	Convey("Should only return metrics for available features on a Mesos agent", t, func() {
		So(namespaces, ShouldNotContain, "/intel/mesos/agent/*/*/net_rx_bytes")
	})
}

func TestMesos_CollectMetrics(t *testing.T) {
	cfg := setupCfg()
	cfgItems, err := config.GetConfigItems(cfg, []string{"master", "agent"}...)
	if err != nil {
		panic(err)
	}

	// Clean slate
	teardown(cfgItems["master"].(string))
	launchTasks(cfgItems["master"].(string), cfgItems["agent"].(string))

	Convey("Collect metrics from a Mesos master and agent", t, func() {
		mc := NewMesosCollector()
		_, err := mc.GetMetricTypes(cfg)
		if err != nil {
			panic(err)
		}

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
				plugin.MetricType{
					Namespace_: core.NewNamespace("intel", "mesos", "master").
						AddDynamicElement("framework_id", "Framework ID").
						AddStaticElements("used_resources", "cpus"),
					Config_: cfg.ConfigDataNode,
				},
			}

			metrics, err := mc.CollectMetrics(mts)
			So(err, ShouldBeNil)
			So(metrics, ShouldNotBeNil)
			So(len(metrics), ShouldEqual, 5)
			So(metrics[0].Namespace().String(), ShouldEqual, "/intel/mesos/master/master/tasks_running")
			So(metrics[1].Namespace().String(), ShouldEqual, "/intel/mesos/master/registrar/state_store_ms/p99")
			So(metrics[2].Namespace().String(), ShouldEqual, "/intel/mesos/master/system/load_5min")
			So(metrics[3].Namespace().Strings()[4], ShouldEqual, "used_resources")
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
					Namespace_: core.NewNamespace("intel", "mesos", "agent").
						AddDynamicElement("framework_id", "Framework ID").
						AddDynamicElement("executor_id", "Executor ID").
						AddStaticElement("cpus_system_time_secs"),
					Config_: cfg.ConfigDataNode,
				},
				plugin.MetricType{
					Namespace_: core.NewNamespace("intel", "mesos", "agent").
						AddDynamicElement("framework_id", "Framework ID").
						AddDynamicElement("executor_id", "Executor ID").
						AddStaticElement("disk_used_bytes"),
					Config_: cfg.ConfigDataNode,
				},
				plugin.MetricType{
					Namespace_: core.NewNamespace("intel", "mesos", "agent").
						AddDynamicElement("framework_id", "Framework ID").
						AddDynamicElement("executor_id", "Executor ID").
						AddStaticElement("mem_total_bytes"),
					Config_: cfg.ConfigDataNode,
				},
			}

			metrics, err := mc.CollectMetrics(mts)
			So(err, ShouldBeNil)
			So(metrics, ShouldNotBeNil)
			So(len(metrics), ShouldEqual, 8)
			So(metrics[0].Namespace().String(), ShouldEqual, "/intel/mesos/agent/slave/tasks_running")
			So(metrics[1].Namespace().String(), ShouldEqual, "/intel/mesos/agent/system/load_5min")
			So(metrics[2].Namespace().Strings()[5], ShouldEqual, "cpus_system_time_secs")
			So(metrics[4].Namespace().Strings()[5], ShouldEqual, "disk_used_bytes")
			So(metrics[6].Namespace().Strings()[5], ShouldEqual, "mem_total_bytes")
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

	teardown(cfgItems["master"].(string))
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

// Launch some Mesos tasks
func launchTasks(masterHost string, agentHost string) {
	cmd := "mesos"
	id := time.Now().Unix()
	task1Args := []string{
		"execute", fmt.Sprintf("--master=%s", masterHost),
		fmt.Sprintf("--name=sleep.%v", id),
		"--resources=cpus:0.5;mem:64;disk:32",
		"--command=date && sleep 60",
	}
	task2Args := []string{
		"execute", fmt.Sprintf("--master=%s", masterHost),
		fmt.Sprintf("--name=sleep: %v", id),
		"--resources=cpus:0.5;mem:64;disk:32",
		"--command=date && sleep 60",
	}

	launch := func(cmd *exec.Cmd) {
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			fmt.Println(stderr.String())
			panic(err)
		}
	}

	go launch(exec.Command(cmd, task1Args...))
	go launch(exec.Command(cmd, task2Args...))

	// There is a delay until disk usage information is available for a newly-launched executor. See:
	//   * https://github.com/apache/mesos/blob/0.28.1/src/slave/containerizer/mesos/isolators/posix/disk.cpp#L352-L357
	//   * https://github.com/apache/mesos/blob/0.28.1/src/tests/disk_quota_tests.cpp#L496-L514
	//
	// Therefore, this function needs to fetch statistics for the new executors it just created and block until
	// the disk usage metrics are available. Otherwise, we'll see some flakiness in the integration tests:
	//
	//   * /home/vagrant/work/src/github.com/intelsdi-x/snap-plugin-collector-mesos/mesos/mesos_integration_test.go
	//   Line 157:
	//   Expected: '8'
	//   Actual:   '6'
	//   (Should be equal)
	//
	done := map[string]bool{}
	for len(done) != 2 {
		executors, err := agent.GetMonitoringStatistics(agentHost)
		if err != nil {
			panic(err)
		}
		if len(executors) != 2 {
			time.Sleep(1)
			continue
		}
		for _, exec := range executors {
			if done[exec.ID] != true {
				if exec.Statistics.DiskUsedBytes != nil {
					done[exec.ID] = true
				} else {
					time.Sleep(1)
				}
			}
		}
	}
}

// Get the system to a clean state by tearing down all active frameworks on the Mesos master, thus killing all tasks.
func teardown(host string) {
	u := url.URL{Scheme: "http", Host: host, Path: "/master/teardown"}
	frameworks, err := master.GetFrameworks(host)
	if err != nil {
		panic(err)
	}

	for _, framework := range frameworks {
		formData := url.Values{}
		formData.Add("frameworkId", framework.ID)
		resp, err := http.PostForm(u.String(), formData)
		if err != nil {
			panic(err)
		}
		if resp.StatusCode != http.StatusOK {
			panic(fmt.Errorf("Expected HTTP 200, got %d", resp.StatusCode))
		}
	}
	time.Sleep(1)
}
