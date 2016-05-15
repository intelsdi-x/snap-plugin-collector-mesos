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
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap-plugin-collector-mesos/mesos/agent"
	"github.com/intelsdi-x/snap-plugin-collector-mesos/mesos/master"
	"github.com/intelsdi-x/snap-plugin-utilities/config"
	"github.com/intelsdi-x/snap-plugin-utilities/ns"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
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

func (m *Mesos) GetMetricTypes(cfg plugin.ConfigType) ([]plugin.MetricType, error) {
	configItems, err := getConfig(cfg)
	if err != nil {
		return nil, err
	}

	metricTypes := []plugin.MetricType{}

	if configItems["master"] != "" {
		master_mts, err := master.GetMetricsSnapshot(configItems["master"])
		if err != nil {
			return nil, err
		}

		for key, _ := range master_mts {
			namespace := core.NewNamespace(pluginVendor, pluginName, "master").
				AddStaticElements(strings.Split(key, "/")...)
			metricTypes = append(metricTypes, plugin.MetricType{Namespace_: namespace})
		}
	}

	if configItems["agent"] != "" {
		agent_mts, err := agent.GetMetricsSnapshot(configItems["agent"])
		if err != nil {
			return nil, err
		}

		for key, _ := range agent_mts {
			namespace := core.NewNamespace(pluginVendor, pluginName, "agent").
				AddStaticElements(strings.Split(key, "/")...)
			metricTypes = append(metricTypes, plugin.MetricType{Namespace_: namespace})
		}

		agent_stats, err := agent.GetMonitoringStatisticsMetricTypes()
		for _, key := range agent_stats {
			namespace := core.NewNamespace(pluginVendor, pluginName, "agent").
				AddDynamicElement("framework_id", "Framework ID").
				AddDynamicElement("executor_id", "Executor ID").
				AddStaticElements(strings.Split(key, "/")...)

			metricTypes = append(metricTypes, plugin.MetricType{Namespace_: namespace})
		}
	}

	return metricTypes, nil
}

func (m *Mesos) CollectMetrics(mts []plugin.MetricType) ([]plugin.MetricType, error) {
	configItems, err := getConfig(mts[0])
	if err != nil {
		return nil, err
	}

	requestedMaster := []string{}
	requestedAgent := []string{}

	for _, metricType := range mts {
		// Mesos metrics are (mostly) returned in a flat JSON object and are '/' delimited, e.g.
		// "slave/cpus_percent". Where they aren't (e.g. perf metrics), we've normalized them into a "/"
		// string. Therefore, we need to return everything after the snap MetricType namespace (e.g.
		// "/intel/mesos/master") as a single string.
		svc := metricType.Namespace().Strings()[2]
		namespace := strings.Join(metricType.Namespace().Strings()[3:], "/")

		switch {
		case svc == "master":
			requestedMaster = append(requestedMaster, namespace)
		case svc == "agent":
			requestedAgent = append(requestedAgent, namespace)
		}
	}

	// Translate Mesos metrics into Snap PluginMetrics
	now := time.Now()
	metrics := []plugin.MetricType{}

	// TODO(roger): only return a master's metrics if master.IsLeader() returns true.
	// If master.IsLeader() is false, this should wait and periodically poll the master
	// to determine if leadership has changed and metrics should now be collected.
	if configItems["master"] != "" && len(requestedMaster) > 0 {
		snapshot, err := master.GetMetricsSnapshot(configItems["master"])
		if err != nil {
			return nil, err
		}

		tags := map[string]string{"source": configItems["master"]}

		for _, key := range requestedMaster {
			val, ok := snapshot[key]
			if !ok {
				return nil, fmt.Errorf("error: requested metric %s not found", val)
			}

			namespace := core.NewNamespace(pluginVendor, pluginName, "master", key)
			//TODO(kromar): is it possible to provide unit NewMetricType(ns, time, tags, unit, value)?
			// I'm leaving empty string for now...
			metric := *plugin.NewMetricType(namespace, now, tags, "", val)
			metrics = append(metrics, metric)
		}
	}

	if configItems["agent"] != "" && len(requestedAgent) > 0 {
		snapshot, err := agent.GetMetricsSnapshot(configItems["agent"])
		if err != nil {
			return nil, err
		}

		executors, err := agent.GetMonitoringStatistics(configItems["agent"])
		if err != nil {
			return nil, err
		}

		tags := map[string]string{"source": configItems["agent"]}

		for _, key := range requestedAgent {
			var val interface{}
			var ok bool = false

			if strings.Split(key, "/")[0] == "*" {
				// Iterate through the array of executors returned by GetMonitoringStatistics()
				for _, exec := range executors {
					sep := "/"
					m := strings.Split(key, sep)[2:]
					val = ns.GetValueByNamespace(exec.Statistics, m)
					if val == nil {
						return nil, fmt.Errorf("error: requested metric %v not found", key)
					}

					// TODO(roger): we can lookup the ID returned by exec.Framework and return a
					// human-readable name if that's desired. So instead of the user needing to
					// make sense of '1101bcf1-4b17-419d-8bbb-6d5b2c9e5eb3-0000', we could instead
					// return 'marathon' or 'chronos'.
					namespace := core.NewNamespace(
						pluginVendor, pluginName, "agent", exec.Framework, exec.ID,
						strings.Join(m, sep))
					// TODO(roger): units
					metric := *plugin.NewMetricType(namespace, now, tags, "", val)
					metrics = append(metrics, metric)

				}
			} else {
				// Get requested metrics from the snapshot map
				val, ok = snapshot[key]
				if !ok {
					return nil, fmt.Errorf("error: requested metric %v not found", key)
				}

				namespace := core.NewNamespace(pluginVendor, pluginName, "agent", key)
				//TODO(kromar): units here also?
				metric := *plugin.NewMetricType(namespace, now, tags, "", val)
				metrics = append(metrics, metric)
			}
		}
	}

	return metrics, nil
}

func getConfig(cfg interface{}) (map[string]string, error) {
	items := make(map[string]string)
	var ok bool

	// Note: although config.GetConfigItems can accept multiple config parameter names, it appears that if
	// any of those names are missing, GetConfigItems() will `return nil, err`. Since this plugin will work
	// individually with master or agent (or both), we break this up into two separate lookups and then
	// test for the existence of the configuration parameter to determine which metric types are available.

	// We expect the value of "master" in the global config to follow the convention "192.168.99.100:5050"
	master_cfg, master_err := config.GetConfigItem(cfg, "master")

	// We expect the value of "agent" in the global config to follow the convention "192.168.99.100:5051"
	agent_cfg, agent_err := config.GetConfigItem(cfg, "agent")

	if master_err != nil && agent_err != nil {
		return items, fmt.Errorf("error: no global config specified for \"master\" and \"agent\".")
	}

	items["master"], ok = master_cfg.(string)
	if !ok {
		log.Warn("no global config specified for \"master\". only \"agent\" metrics will be collected.")
	}

	items["agent"], ok = agent_cfg.(string)
	if !ok {
		log.Warn("no global config specified for \"agent\". only \"master\" metrics will be collected.")
	}

	return items, nil
}
