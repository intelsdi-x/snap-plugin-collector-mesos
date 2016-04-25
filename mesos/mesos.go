// +build linux

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
	"strings"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"

	"github.com/intelsdi-x/snap-plugin-utilities/config"
	"github.com/intelsdi-x/snap-plugin-utilities/ns"

	"github.com/intelsdi-x/snap-plugin-collector-mesos/mesos/agent"
)

const (
	pluginVendor  = "intel"
	pluginName    = "mesos"
	pluginVersion = 1
	pluginType    = plugin.CollectorPluginType
)

func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(
		pluginName,
		pluginVersion,
		pluginType,
		[]string{plugin.SnapGOBContentType},
		[]string{plugin.SnapGOBContentType})
}

func NewMesosCollector() *Mesos {
	return &Mesos{}
}

type Mesos struct {
}

func (m *Mesos) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	return cpolicy.New(), nil
}

func (m *Mesos) GetMetricTypes(cfg plugin.PluginConfigType) ([]plugin.PluginMetricType, error) {
	metricTypes := []plugin.PluginMetricType{}
	items, err := config.GetConfigItems(cfg, "agent_url", "master_url")
	if err != nil {
		return nil, fmt.Errorf("Mesos master and/or agent url not provided")
	}

	namespaces := []string{}
	// build namespace from mesos agent endpoint
	{
		agentURL := items["agent_url"].(string)
		executors, err := agent.GetAgentStatistics(agentURL)
		if err != nil {
			return nil, fmt.Errorf("Can't get Mesos agent statistics")
		}
		// include double wildcard for executor_id and framework_id
		// TODO: with incoming changes to namespace creation it will need some rework
		for _, executor := range executors {
			ns.FromMap(
				executor.Statistics,
				strings.Join(
					[]string{pluginVendor, pluginName, "agent", "*", "*"}, "/"),
				&namespaces,
			)
		}
	}

	// build namespace from mesos master endpoint
	{
		//TODO
	}

	// build metric types from available namespaces
	for _, namespace := range namespaces {
		metricTypes = append(metricTypes, plugin.PluginMetricType{
			Namespace_: strings.Split(namespace, "/"),
		})
	}

	return metricTypes, nil
}

func (m *Mesos) CollectMetrics(mts []plugin.PluginMetricType) ([]plugin.PluginMetricType, error) {
	// TODO
	return []plugin.PluginMetricType{}, nil
}
