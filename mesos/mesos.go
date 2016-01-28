// +build linux

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


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
	"strings"
	"time"

	"github.com/intelsdi-x/snap-plugin-utilities/config"
	"github.com/intelsdi-x/snap-plugin-utilities/ns"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"

	"github.com/intelsdi-x/snap-plugin-collector-mesos/agent"
	"github.com/intelsdi-x/snap/core"
)

const (
	VENDOR  = "intel"
	PLUGIN  = "mesos"
	SERVICE = "agent"
	VERSION = 1
)

type mesosPlugin struct {
	host string
}

func New() *mesosPlugin {
	host, err := os.Hostname()
	if err != nil {
		host = "localhost"
	}
	return &mesosPlugin{host: host}
}

func (mp *mesosPlugin) GetMetricTypes(cfg plugin.PluginConfigType) ([]plugin.PluginMetricType, error) {
	metricTypes := []plugin.PluginMetricType{}
	url, err := config.GetConfigItem(cfg, "url")
	if err != nil {
		return nil, fmt.Errorf("url config parameter not provided")
	}
	content, err := agent.GetStats(url.(string))
	if err != nil {
		return nil, err
	}

	executors := agent.ParseStats(content)
	namespace := []string{}
	for _, executor := range executors {
		ns.FromMap(
			executor.Statistics,
			strings.Join(
				[]string{VENDOR, PLUGIN, SERVICE, executor.ID, executor.Framework}, "/"),
			&namespace,
		)
	}

	for _, n := range namespace {
		metricTypes = append(metricTypes, plugin.PluginMetricType{
			Namespace_: strings.Split(n, "/"),
			Config_:    cfg.ConfigDataNode,
		})
	}

	metricTypes = append(metricTypes, plugin.PluginMetricType{
		Namespace_: []string{VENDOR, PLUGIN, SERVICE, "*"},
		Config_:    cfg.ConfigDataNode,
	})
	return metricTypes, nil
}

func (mp *mesosPlugin) CollectMetrics(metricTypes []plugin.PluginMetricType) ([]plugin.PluginMetricType, error) {
	metrics := []plugin.PluginMetricType{}
	url, err := config.GetConfigItem(metricTypes[0], "url")
	if err != nil {
		return nil, fmt.Errorf("url config parameter not provided")
	}

	content, err := agent.GetStats(url.(string))
	if err != nil {
		// TODO - workaround for mesos cluster
		// TODO - Handling case where there is not connection to mesos agent
		return nil, nil
	}
	executors := agent.ParseStats(content)

	for _, metricType := range metricTypes {
		namespace := metricType.Namespace()
		if len(namespace) < 4 {
			return nil, fmt.Errorf("Namespace length is to short. Is %d, expected >=4", len(namespace))
		}
		for _, executor := range executors {
			tag := map[string]string{
				"executor_name":   executor.Name,
				"executor_id":     executor.ID,
				"framework_id":    executor.Framework,
				"executor_source": executor.Source,
			}
			metric := plugin.PluginMetricType{
				Source_:    mp.host,
				Tags_:      tag,
				Timestamp_: time.Now(),
				Labels_: []core.Label{
					{Index: 3, Name: executor.ID},
					{Index: 4, Name: executor.Framework},
				},
			}
			if namespace[3] == "*" {
				for stat, val := range executor.Statistics {
					nspace := []string{
						VENDOR,
						PLUGIN,
						SERVICE,
						executor.ID,
						executor.Framework,
						stat,
					}
					if perf, ok := val.(map[string]interface{}); ok {
						for perfStat, perfVal := range perf {
							metric.Namespace_ = append(nspace, perfStat)
							metric.Data_ = perfVal
							metrics = append(metrics, metric)
						}
					} else {
						metric.Namespace_ = nspace
						metric.Data_ = val
						metrics = append(metrics, metric)
					}
				}
			} else {
				eId := namespace[3]
				fId := namespace[4]
				if eId == executor.ID && fId == executor.Framework {
					// TODO - finish populating for specific metrics
					fmt.Println("foo")
				}
			}
		}
	}
	return metrics, nil
}

func (mp *mesosPlugin) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	return cpolicy.New(), nil
}
