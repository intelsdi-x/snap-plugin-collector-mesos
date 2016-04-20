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
