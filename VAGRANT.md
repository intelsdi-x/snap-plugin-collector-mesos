## Vagrant development environment

Included in this repo is a Vagrant development environment that provisions a single-node Mesos cluster for developing
and testing this plugin. To get started, you'll need to have:

  * VirtualBox: <https://www.virtualbox.org/wiki/Downloads>
  * Vagrant: <https://www.vagrantup.com/downloads.html>

Next, assuming that your current working directory is the root of this Git repository, simply run the following command:

```
$ vagrant up
```

The provisioning script will install Mesos, ZooKeeper, Go, and Snap. You can then connect to Mesos at the following
URLs:

  * Mesos master: <http://10.180.10.180:5050>
  * Mesos agent: <http://10.180.10.180:5051>
  * InfluxDB: <http://10.180.10.180:8083>
  * Grafana: <http://10.180.10.180:3000>

In order to launch a task on the cluster that you can then observe with Snap, try running the following command in a
separate terminal window / SSH session:

```
$ mesos execute --master="$(mesos resolve zk://10.180.10.180:2181/mesos)" \
    --name="PythonHttpServer" --command="python -m SimpleHTTPServer 9000"
```
