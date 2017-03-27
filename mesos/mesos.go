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
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap-plugin-collector-mesos/mesos/agent"
	"github.com/intelsdi-x/snap-plugin-collector-mesos/mesos/master"
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

const (
	PluginVendor  = "intel"
	PluginName    = "mesos"
	PluginVersion = 2
)

func NewMesosCollector() *Mesos {
	log.Debug("Created a new instance of the Mesos collector plugin")
	return &Mesos{LastRun: 0}
}

type Mesos struct {
	LastRun int64
}

func (m *Mesos) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	configKeyMaster := []string{"intel", "mesos", "master"}
	configKeyAgent := []string{"intel", "mesos", "agent"}

	policy.AddNewStringRule(configKeyMaster,
		"master",
		false,
		plugin.SetDefaultString("127.0.0.1:5050"))
	policy.AddNewBoolRule(configKeyMaster, "resolve_local_ips", false, plugin.SetDefaultBool(true))

	policy.AddNewStringRule(configKeyAgent,
		"agent",
		false,
		plugin.SetDefaultString("127.0.0.1:5051"))

	return *policy, nil
}

func (m *Mesos) GetMetricTypes(cfg plugin.Config) ([]plugin.Metric, error) {
	var metricTypes []plugin.Metric
	metricTypes = append(metricTypes, plugin.Metric{Namespace: plugin.NewNamespace(PluginVendor, PluginName, "master"), Version: PluginVersion})
	metricTypes = append(metricTypes, plugin.Metric{Namespace: plugin.NewNamespace(PluginVendor, PluginName, "agent"), Version: PluginVersion})
	return metricTypes, nil
}

func decodeTree(tree *interface{}, ret *map[string]interface{}, cpath string) {
	//	if reflect.ValueOf(tree).Kind() == reflect.Map {
	i2 := (*tree).(map[string]interface{})
	for k, v := range i2 {
		if reflect.ValueOf(v).Kind() == reflect.Map {
			key := cpath + "/" + k
			decodeTree(&v, ret, key)
		} else {
			//			ret2 := (*ret)
			key := cpath + "/" + k
			(*ret)[key] = v
		}
	}
}

func (m *Mesos) CollectMetrics(mts []plugin.Metric) ([]plugin.Metric, error) {
	metrics := []plugin.Metric{}
	timestamp := time.Now()

	for _, item := range mts {
		/*
			filter,err := item.Config.GetString("filter")
			if err != nil {
				return nil, err
			}
		*/

		switch item.Namespace.Strings()[2] {
		case "master":
			log.Debug("Start master collection ...")
			endpoint, err := item.Config.GetString("master")
			if err != nil {
				return nil, err
			}
			tags := map[string]string{"source": endpoint}
			resolve, err := item.Config.GetBool("resolve_local_ips")
			if err != nil {
				return nil, err
			}

			isLeader, err := master.IsLeader(endpoint, resolve)
			log.Debug("IsLeader: " + strconv.FormatBool(isLeader))
			if err != nil {
				log.Warning(err)
				isLeader = false
			}
			if isLeader {
				snapshot, err := master.GetMetricsSnapshot(endpoint)
				if err != nil {
					log.Error(err)
					return nil, err
				}
				log.Debug("Query frameworks ...")
				frameworks, err := master.GetFrameworks(endpoint)
				if err != nil {
					log.Error(err)
					return nil, err
				}
				log.Debug("Metrics snapshot ...")
				for k, v := range snapshot {
					ns := plugin.NewNamespace(PluginVendor, PluginName, "master")
					metric := plugin.Metric{
						Timestamp: timestamp,
						Namespace: ns.AddStaticElements(strings.Split(k, "/")...),
						Config:    item.Config,
						Data:      v,
						Tags:      tags,
						Version:   PluginVersion,
					}

					metrics = append(metrics, metric)
				}
				log.Debug("Parse framwork information ...")
				for _, framework := range frameworks {
					var tree interface{}
					data := make(map[string]interface{})
					bytes, _ := json.Marshal(framework)
					json.Unmarshal(bytes, &tree)
					decodeTree(&tree, &data, "")

					for k, v := range data {
						ns := plugin.NewNamespace(PluginVendor, PluginName, "master", "framework", framework.ID)
						metric := plugin.Metric{
							Timestamp: timestamp,
							Namespace: ns.AddStaticElements(strings.Split(k, "/")[1:]...),
							Config:    item.Config,
							Data:      v,
							Tags:      tags,
							Version:   PluginVersion,
						}
						metrics = append(metrics, metric)
					}
				}
			} else {
				log.Debug("Not leader.")
			}
		case "agent":
			endpoint, err := item.Config.GetString("agent")
			if err != nil {
				return nil, err
			}
			tags := map[string]string{"source": endpoint}
			snapshot, err := agent.GetMetricsSnapshot(endpoint)
			if err != nil {
				log.Warning(err)
			} else {
				executors, err := agent.GetMonitoringStatistics(endpoint)
				if err != nil {
					log.Warning(err)
				}

				for k, v := range snapshot {
					ns := plugin.NewNamespace(PluginVendor, PluginName, "agent")
					metric := plugin.Metric{
						Timestamp: timestamp,
						Namespace: ns.AddStaticElements(strings.Split(k, "/")...),
						Config:    item.Config,
						Data:      v,
						Tags:      tags,
						Version:   PluginVersion,
					}

					metrics = append(metrics, metric)
				}

				for _, executor := range executors {
					var tree interface{}
					data := make(map[string]interface{})
					bytes, _ := json.Marshal(executor)
					json.Unmarshal(bytes, &tree)
					decodeTree(&tree, &data, "")
					for k, v := range data {
						ns := plugin.NewNamespace(PluginVendor, PluginName, "agent", "executor", executor.ID)
						metric := plugin.Metric{
							Timestamp: timestamp,
							Namespace: ns.AddStaticElements(strings.Split(k, "/")[1:]...),
							Config:    item.Config,
							Data:      v,
							Tags:      tags,
							Version:   PluginVersion,
						}
						metrics = append(metrics, metric)
					}
					if m.LastRun > 0 {
						cTree := tree.(map[string]interface{})
						if sTree, ok := cTree["statistics"]; ok {
							dTree := sTree.(map[string]interface{})
							if cpus_limit, ok := dTree["cpus_limit"]; ok {
								cl, _ := cpus_limit.(float64)
								if cpus_system_time_secs, ok := dTree["cpus_system_time_secs"]; ok {
									cs, _ := cpus_system_time_secs.(float64)
									if cpus_user_time_secs, ok := dTree["cpus_user_time_secs"]; ok {
										cu, _ := cpus_user_time_secs.(float64)
										cd := (100000000000 * (cs + cu)) / (cl * (float64(timestamp.UnixNano()) - float64(m.LastRun)))
										ns := plugin.NewNamespace(PluginVendor, PluginName, "agent", "executor", executor.ID, "statistics", "cpus_util_pc_diff")
										metrics = append(metrics, plugin.Metric{
											Timestamp: timestamp,
											Namespace: ns,
											Config:    item.Config,
											Data:      cd,
											Tags:      tags,
											Version:   PluginVersion,
										})
									}
								}
							}
						}
					}
				}
				m.LastRun = timestamp.UnixNano()
			}
		}
	}
	log.Debug("Collected a total of ", len(metrics), " metrics.")
	return metrics, nil
}
