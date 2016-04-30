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

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap-plugin-collector-mesos/mesos/agent"
	"github.com/intelsdi-x/snap-plugin-collector-mesos/mesos/master"
	"github.com/intelsdi-x/snap-plugin-utilities/config"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
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
	// Note: although config.GetConfigItems can accept multiple config parameter names, it appears that if
	// any of those names are missing, GetConfigItems() will `return nil, err`. Since this plugin will work
	// individually with master or agent (or both), we break this up into two separate lookups and then
	// test for the existence of the configuration parameter to determine which metric types are available.

	// We expect the value of "master" in the global config to follow the convention "192.168.99.100:5050"
	master_cfg, master_err := config.GetConfigItems(cfg, []string{"master"})

	// We expect the value of "agent" in the global config to follow the convention "192.168.99.100:5051"
	agent_cfg, agent_err := config.GetConfigItems(cfg, []string{"agent"})

	if master_err != nil && agent_err != nil {
		return nil, fmt.Errorf("error: no global config specified for \"master\" or \"agent\".")
	}
	if master_err != nil {
		log.Warn("no global config specified for \"master\". only \"agent\" metrics will be collected.")
	}
	if agent_err != nil {
		log.Warn("no global config specified for \"agent\". only \"master\" metrics will be collected.")
	}

	metricTypes := []plugin.PluginMetricType{}

	if master_err == nil {
		master_mts, err := master.GetMetricsSnapshot(master_cfg["master"].(string))
		if err != nil {
			return nil, err
		}

		for key, _ := range master_mts {
			namespace := append([]string{pluginVendor, pluginName, "master"}, strings.Split(key, "/")...)
			metricTypes = append(metricTypes, plugin.PluginMetricType{Namespace_: namespace})
		}
	}

	if agent_err == nil {
		agent_mts, err := agent.GetMetricsSnapshot(agent_cfg["agent"].(string))
		if err != nil {
			return nil, err
		}

		for key, _ := range agent_mts {
			namespace := append([]string{pluginVendor, pluginName, "agent"}, strings.Split(key, "/")...)
			metricTypes = append(metricTypes, plugin.PluginMetricType{Namespace_: namespace})
		}
	}

	return metricTypes, nil
}

func (m *Mesos) CollectMetrics(mts []plugin.PluginMetricType) ([]plugin.PluginMetricType, error) {
	return []plugin.PluginMetricType{}, nil
}
