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

	"github.com/intelsdi-x/snap-plugin-collector-mesos/mesos/agent"
	"github.com/intelsdi-x/snap-plugin-collector-mesos/mesos/master"
	"github.com/intelsdi-x/snap-plugin-utilities/config"
	"github.com/intelsdi-x/snap-plugin-utilities/ns"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	log "github.com/sirupsen/logrus"
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
	log.Debug("Created a new instance of the Mesos collector plugin")
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
		log.Error(err)
		return nil, err
	}

	metricTypes := []plugin.MetricType{}

	if configItems["master"] != "" {
		log.Info("Getting metric types for the Mesos master at ", configItems["master"])
		master_mts, err := master.GetMetricsSnapshot(configItems["master"])
		if err != nil {
			log.Error(err)
			return nil, err
		}

		for key, _ := range master_mts {
			namespace := core.NewNamespace(pluginVendor, pluginName, "master").
				AddStaticElements(strings.Split(key, "/")...)
			log.Debug("Adding metric to catalog: ", namespace.String())
			metricTypes = append(metricTypes, plugin.MetricType{Namespace_: namespace})
		}

		framework_mts, err := master.GetFrameworksMetricTypes()
		if err != nil {
			log.Error(err)
			return nil, err
		}

		for _, key := range framework_mts {
			namespace := core.NewNamespace(pluginVendor, pluginName, "master").
				AddDynamicElement("framework_id", "Framework ID").
				AddStaticElements(strings.Split(key, "/")...)
			log.Debug("Adding metric to catalog: ", namespace.String())
			metricTypes = append(metricTypes, plugin.MetricType{Namespace_: namespace})
		}
	}

	if configItems["agent"] != "" {
		log.Info("Getting metric types for the Mesos agent at ", configItems["agent"])
		agent_mts, err := agent.GetMetricsSnapshot(configItems["agent"])
		if err != nil {
			log.Error(err)
			return nil, err
		}

		for key, _ := range agent_mts {
			namespace := core.NewNamespace(pluginVendor, pluginName, "agent").
				AddStaticElements(strings.Split(key, "/")...)
			log.Debug("Adding metric to catalog: ", namespace.String())
			metricTypes = append(metricTypes, plugin.MetricType{Namespace_: namespace})
		}

		agent_stats, err := agent.GetMonitoringStatisticsMetricTypes(configItems["agent"])
		if err != nil {
			log.Error(err)
			return nil, err
		}

		for _, key := range agent_stats {
			namespace := core.NewNamespace(pluginVendor, pluginName, "agent").
				AddDynamicElement("framework_id", "Framework ID").
				AddDynamicElement("executor_id", "Executor ID").
				AddStaticElements(strings.Split(key, "/")...)
			log.Debug("Adding metric to catalog: ", namespace.String())
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

	requestedMaster := []core.Namespace{}
	requestedAgent := []core.Namespace{}

	for _, metricType := range mts {
		switch metricType.Namespace().Strings()[2] {
		case "master":
			requestedMaster = append(requestedMaster, metricType.Namespace())
		case "agent":
			requestedAgent = append(requestedAgent, metricType.Namespace())
		}
	}

	// Translate Mesos metrics into Snap PluginMetrics
	now := time.Now()
	metrics := []plugin.MetricType{}

	if configItems["master"] != "" && len(requestedMaster) > 0 {
		log.Info("Collecting ", len(requestedMaster), " metrics from the master")
		isLeader, err := master.IsLeader(configItems["master"])
		if err != nil {
			log.Error(err)
			return nil, err
		}
		if isLeader {
			snapshot, err := master.GetMetricsSnapshot(configItems["master"])
			if err != nil {
				log.Error(err)
				return nil, err
			}

			frameworks, err := master.GetFrameworks(configItems["master"])
			if err != nil {
				log.Error(err)
				return nil, err
			}

			tags := map[string]string{"source": configItems["master"]}

			for _, requested := range requestedMaster {
				isDynamic, _ := requested.IsDynamic()
				if isDynamic {
					n := requested.Strings()[4:]

					// Iterate through the array of frameworks returned by GetFrameworks()
					for _, framework := range frameworks {
						val := ns.GetValueByNamespace(framework, n)
						if val == nil {
							log.Warn("Attempted to collect metric ", requested.String(), " but it returned nil!")
							continue
						}
						// substituting "framework" wildcard with particular framework id
						rendered := cloneNamespace(requested)
						rendered[3].Value = framework.ID
						// TODO(roger): units
						metrics = append(metrics, *plugin.NewMetricType(rendered, now, tags, "", val))

					}
				} else {
					n := requested.Strings()[3:]
					val, ok := snapshot[strings.Join(n, "/")]
					if !ok {
						e := fmt.Errorf("error: requested metric %s not found", requested.String())
						log.Error(e)
						return nil, e
					}
					//TODO(kromar): is it possible to provide unit NewMetricType(ns, time, tags, unit, value)?
					// I'm leaving empty string for now...
					metrics = append(metrics, *plugin.NewMetricType(requested, now, tags, "", val))
				}
			}
		} else {
			log.Info("Attempted CollectMetrics() on ", configItems["master"], "but it isn't the leader. Skipping...")
		}
	}

	if configItems["agent"] != "" && len(requestedAgent) > 0 {
		log.Info("Collecting ", len(requestedAgent), " metrics from the agent")
		snapshot, err := agent.GetMetricsSnapshot(configItems["agent"])
		if err != nil {
			log.Error(err)
			return nil, err
		}

		executors, err := agent.GetMonitoringStatistics(configItems["agent"])
		if err != nil {
			log.Error(err)
			return nil, err
		}

		tags := map[string]string{"source": configItems["agent"]}

		for _, requested := range requestedAgent {
			n := requested.Strings()[5:]
			isDynamic, _ := requested.IsDynamic()
			if isDynamic {
				// Iterate through the array of executors returned by GetMonitoringStatistics()
				for _, exec := range executors {
					val := ns.GetValueByNamespace(exec.Statistics, n)
					if val == nil {
						log.Warn("Attempted to collect metric ", requested.String(), " but it returned nil!")
						continue
					}
					rendered := cloneNamespace(requested)
					// substituting "framework" wildcard with particular framework id
					rendered[3].Value = exec.Framework
					// substituting "executor" wildcard with particular executor id
					rendered[4].Value = exec.ID
					// TODO(roger): units
					metrics = append(metrics, *plugin.NewMetricType(rendered, now, tags, "", val))

				}
			} else {
				// Get requested metrics from the snapshot map
				n := requested.Strings()[3:]
				val, ok := snapshot[strings.Join(n, "/")]
				if !ok {
					e := fmt.Errorf("error: requested metric %v not found", requested.String())
					log.Error(e)
					return nil, e
				}

				//TODO(kromar): units here also?
				metrics = append(metrics, *plugin.NewMetricType(requested, now, tags, "", val))
			}
		}
	}

	log.Debug("Collected a total of ", len(metrics), " metrics.")
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
		e := fmt.Errorf("error: no global config specified for 'master' and 'agent'.")
		log.Error(e)
		return items, e
	}

	items["master"], ok = master_cfg.(string)
	if !ok {
		log.Warn("No global config specified for 'master', only 'agent' metrics will be collected.")
	}

	items["agent"], ok = agent_cfg.(string)
	if !ok {
		log.Warn("No global config specified for 'agent', only 'master' metrics will be collected.")
	}

	return items, nil
}

func cloneNamespace(ns core.Namespace) core.Namespace {
	nsCopy := make(core.Namespace, len(ns))
	copy(nsCopy, ns)

	return nsCopy
}
