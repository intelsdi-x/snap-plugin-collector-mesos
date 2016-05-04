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

package mesos

import (
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMesosPlugin(t *testing.T) {
	Convey("Meta should return metadata for the plugin", t, func() {
		meta := Meta()
		So(meta.Name, ShouldResemble, pluginName)
		So(meta.Version, ShouldResemble, pluginVersion)
		So(meta.Type, ShouldResemble, pluginType)
	})

	Convey("Create Mesos collector", t, func() {
		mc := NewMesosCollector()
		Convey("MesosCollector should not be nil", func() {
			So(mc, ShouldNotBeNil)
		})
		Convey("MesosCollector should be of type Mesos", func() {
			So(mc, ShouldHaveSameTypeAs, &Mesos{})
		})
	})
}

func TestMesos_getConfig(t *testing.T) {
	log.SetLevel(log.ErrorLevel) // Suppress warning messages from getConfig

	Convey("Get plugin configuration from snap global config", t, func() {
		Convey("When only a master is provided, getConfig() should return only the master value", func() {
			node := cdata.NewNode()
			node.AddItem("master", ctypes.ConfigValueStr{Value: "mesos-master.example.com:5050"})
			snapCfg := plugin.ConfigType{ConfigDataNode: node}

			parsedCfg, err := getConfig(snapCfg)

			So(parsedCfg["master"], ShouldEqual, "mesos-master.example.com:5050")
			So(parsedCfg["agent"], ShouldEqual, "")
			So(err, ShouldBeNil)
		})

		Convey("When only an agent is provided, getConfig() should return only the agent value", func() {
			node := cdata.NewNode()
			node.AddItem("agent", ctypes.ConfigValueStr{Value: "mesos-agent.example.com:5051"})
			snapCfg := plugin.ConfigType{ConfigDataNode: node}

			parsedCfg, err := getConfig(snapCfg)

			So(parsedCfg["master"], ShouldEqual, "")
			So(parsedCfg["agent"], ShouldEqual, "mesos-agent.example.com:5051")
			So(err, ShouldBeNil)
		})

		Convey("When both a master and an agent are provided, getConfig() should return both values", func() {
			node := cdata.NewNode()
			node.AddItem("master", ctypes.ConfigValueStr{Value: "mesos-master.example.com:5050"})
			node.AddItem("agent", ctypes.ConfigValueStr{Value: "mesos-agent.example.com:5051"})
			snapCfg := plugin.ConfigType{ConfigDataNode: node}

			parsedCfg, err := getConfig(snapCfg)

			So(len(parsedCfg), ShouldEqual, 2)
			So(err, ShouldBeNil)
		})

		Convey("When both master and agent are missing, getConfig() should return an error", func() {
			node := cdata.NewNode()
			node.AddItem("foo", ctypes.ConfigValueStr{Value: "bar"})
			snapCfg := plugin.ConfigType{ConfigDataNode: node}

			parsedCfg, err := getConfig(snapCfg)

			So(len(parsedCfg), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})
	})
}
