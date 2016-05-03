# Snap Collector Plugin - Apache Mesos

[![Build Status](https://travis-ci.com/intelsdi-x/snap-plugin-collector-mesos.svg?token=mxqCYyjxtayP5XBp4JEu&branch=master)](https://travis-ci.com/intelsdi-x/snap-plugin-collector-mesos)

This snap plugin collects metrics from an [Apache Mesos][mesos-home] cluster.
It gathers information about cluster resource allocation and utilization, as
well as metrics about running containers.

1. [Getting Started](#getting-started)
    * [System Requirements](#system-requirements)
    * [Installation](#installation)
    * [Configuration and Usage](#configuration-and-usage)
2. [Documentation](#documentation)
    * [Collected Metrics](#collected-metrics)
    * [Examples](#examples)
    * [Known Issues and Caveats](#known-issues-and-caveats)
    * [Roadmap](#roadmap)
3. [Community Support](#community-support)
4. [Contributing](#contributing)
5. [License](#license)
6. [Acknowledgements](#acknowledgements)

## Getting Started
### System Requirements
### Installation
### Configuration and Usage

## Documentation
### Collected Metrics
### Examples
There are examples of the snap global configuration and various tasks located in the [examples/](examples) directory.
To get started with these examples and collect Mesos metrics and publish them to a file, you'll need to perform the
following steps.

*Note: these steps will work with the Vagrant development environment included in this repo. For more info on how
to get started with Vagrant, please see [CONTRIBUTING.md](CONTRIBUTING.md).*

Start the snap daemon in the background:

```
$ snapd --plugin-trust 0 --log-level 1 --config examples/configs/snap-config-example.json \
    > /tmp/snap.log 2>&1 &
```

Assuming you're in the working directory for this plugin, load the Mesos collector plugin:

```
$ snapctl plugin load build/rootfs/snap-plugin-collector-mesos
```

Get the available metrics for your system:

```
$ snapctl metric list
```

Load the `passthru` processor plugin, and the `file` publisher plugin:

```
$ snapctl plugin load ${SNAP_PATH}/plugin/snap-processor-passthru
$ snapctl plugin load ${SNAP_PATH}/plugin/snap-publisher-file
```

Create a new snap task:

```
$ snapctl task create -t examples/tasks/mesos-all-file.json
```

Stop the task:

```
$ snapctl task stop <task ID>
```

### Known Issues and Caveats
  * Snap's metric catalog is populated only once, when the Mesos collector plugin is loaded. A configuration change on
  the master or agent could alter the metrics reported by Mesos. Therefore, if you modify the configuration of a Mesos
  master or agent, you should reload this Snap plugin at the same time.
  * Due to a bug in Mesos, the parsing logic for the `perf` command was incorrect on certain platforms and kernels. When
  the `cgroups/perf_event` isolator was enabled on an agent, the `perf` object would appear in the JSON returned by the
  agent's `/monitor/statistics` endpoint, but it would contain no data. This issue was resolved in Mesos 0.29.0, and was
  backported to Mesos 0.28.2, 0.27.3, and 0.26.2. For more information, see [MESOS-4705][mesos-4705-jira].
  * There is an ongoing effort to rename the Mesos "slave" service to "agent". As of Mesos 0.28.x, this work is still
  in progress. This plugin uses the newer "agent" terminology, but some metrics returned by Mesos may still use the
  older "slave" term. For more information, see [MESOS-1478][mesos-1478-jira].

### Roadmap

## Community Support

## Contributing
We love contributions!

There's more than one way to give back, from examples to blog posts to code updates. See our recommended process in
[CONTRIBUTING.md](CONTRIBUTING.md).

## License
[snap][snap-github], along with this plugin, is open source software released
under the [Apache Software License, version 2.0](LICENSE).

## Acknowledgements
  * Authors: [Marcin Krolik][marcin-github], [Roger Ignazio][roger-github]


[marcin-github]: https://github.com/marcin-krolik
[mesos-1478-jira]: https://issues.apache.org/jira/browse/MESOS-1478
[mesos-4705-jira]: https://issues.apache.org/jira/browse/MESOS-4705
[mesos-home]: http://mesos.apache.org
[roger-github]: https://github.com/rji
[snap-github]: http://github.com/intelsdi-x/snap
