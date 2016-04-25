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
	return []plugin.PluginMetricType{}, nil
}

func (m *Mesos) CollectMetrics(mts []plugin.PluginMetricType) ([]plugin.PluginMetricType, error) {
	return []plugin.PluginMetricType{}, nil
}
