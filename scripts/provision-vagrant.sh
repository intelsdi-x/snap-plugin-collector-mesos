#!/bin/bash
# This script provisions a single-node Mesos cluster that can be used for
# developing and testing the Mesos plugin for the Snap telemetry framework.
set -e

if [[ $(id -u) -ne 0 ]]; then
    echo "Please re-run this script as root."
    exit 1
fi

DISTRO=$(lsb_release -is | tr '[:upper:]' '[:lower:]')
CODENAME=$(lsb_release -cs)

function parse_args {
    while [[ $# > 1 ]]; do
        case "$1" in
            --mesos_release)  MESOS_RELEASE="$2"                ; shift  ;;
            --golang_release) GOLANG_RELEASE="$2"               ; shift  ;;
            --snap_release)   SNAP_RELEASE="$2"                 ; shift  ;;
            --*)              echo "Error: invalid option '$1'" ; exit 1 ;;
        esac
        shift
    done
}

function _install_pkg_with_version {
    local name="$1"
    local ver="$2"

    if [[ $ver =~ "latest" ]]; then
        echo "Installing ${name}..."
        apt-get -y install "${name}"
    else
        echo "Installing ${name} version ${ver}..."
        apt-get -y install "${name}=${ver}"
    fi
}

function install_prereqs {
    echo "Updating metadata and installing prerequisites..."
    apt-get -y update
    apt-get -y install apt-transport-https ca-certificates git \
        linux-tools-common linux-tools-generic linux-tools-$(uname -r)
}

function install_zookeeper {
    echo "Installing ZooKeeper..."
    apt-get -y install zookeeperd
    echo "1" > /etc/zookeeper/conf/myid
    service zookeeper restart
}

function install_mesos {
    echo "Installing Mesosphere repository..."
    apt-key adv --keyserver keyserver.ubuntu.com --recv E56151BF
    echo "deb http://repos.mesosphere.io/${DISTRO} ${CODENAME} main" \
        | tee /etc/apt/sources.list.d/mesosphere.list

    echo "Refreshing metadata..."
    apt-get -y update

    _install_pkg_with_version mesos $MESOS_RELEASE
}

function configure_mesos {
    mkdir -p /etc/{mesos,mesos-master,mesos-slave}

    # Master
    echo "zk://10.180.10.180:2181/mesos" > /etc/mesos/zk
    echo "10.180.10.180"                 > /etc/mesos-master/hostname
    echo "10.180.10.180"                 > /etc/mesos-master/ip
    service mesos-master restart

    # Agent
    # Note: there is a known bug in Mesos when using the 'cgroups/perf_event' isolator
    # on specific kernels and platforms. For more info, see MESOS-4705.
    echo "mesos"                                      > /etc/mesos-slave/containerizers
    echo "10.180.10.180"                              > /etc/mesos-slave/hostname
    echo "10.180.10.180"                              > /etc/mesos-slave/ip
    echo "cgroups/cpu,cgroups/mem,cgroups/perf_event" > /etc/mesos-slave/isolation
    echo "cpu-clock,task-clock,context-switches"      > /etc/mesos-slave/perf_events
    echo "/var/lib/mesos"                             > /etc/mesos-slave/work_dir
    service mesos-slave restart

    cat << END
--------------------------------------------------
Mesos version ${MESOS_RELEASE} has been installed.

  * Master: http://10.180.10.180:5050
  * Agent: http://10.180.10.180:5051

--------------------------------------------------
END
}

function install_golang {
    local GOROOT="/usr/local"

    if [[ -d "${GOROOT}/go" ]]; then
        echo "Found an existing Go installation at ${GOROOT}/go. Skipping install..."
        return
    fi

    echo "Installing Go ${GOLANG_RELEASE}..."
    local GOLANG_URL="https://storage.googleapis.com/golang"
    local GOLANG_FILENAME="go${GOLANG_RELEASE}.linux-amd64.tar.gz"

    local GOPATH="/home/vagrant/work"

    curl -sLO "${GOLANG_URL}/${GOLANG_FILENAME}"
    tar zxf $GOLANG_FILENAME -C $GOROOT

    echo "export GOPATH=${GOPATH}"                              >> /etc/profile
    echo "export PATH=\${PATH}:\${GOPATH}/bin:${GOROOT}/go/bin" >> /etc/profile

    mkdir -p "${GOPATH}/src/github.com/intelsdi-x" && chown -R vagrant:vagrant $GOPATH
    ln -fs /vagrant "${GOPATH}/src/github.com/intelsdi-x/snap-plugin-collector-mesos"

    cat << END
--------------------------------------------------------------
Go ${GOLANG_RELEASE} has been installed to /usr/local/go.
/usr/local/go/bin has been appended to \$PATH in /etc/profile.
/home/vagrant/work has been set as the \$GOPATH.

For more information on getting started with Go, see
https://golang.org/doc/code.html

--------------------------------------------------------------
END
}

function install_snap {
    local SNAP_PATH="/usr/local/snap"
    if [[ -d $SNAP_PATH ]]; then
        echo "Found an existing Snap installation at ${SNAP_PATH}. Skipping install..."
        return
    fi

    echo "Installing snap ${SNAP_RELEASE}..."
    local SNAP_URL="https://github.com/intelsdi-x/snap/releases/download"
    local SNAP_FILENAME="snap-${SNAP_RELEASE}-linux-amd64.tar.gz"
    local SNAP_PLUGINS_FILENAME="snap-plugins-${SNAP_RELEASE}-linux-amd64.tar.gz"
    curl -sLO "${SNAP_URL}/${SNAP_RELEASE}/${SNAP_FILENAME}"
    curl -sLO "${SNAP_URL}/${SNAP_RELEASE}/${SNAP_PLUGINS_FILENAME}"

    echo "export SNAP_PATH=${SNAP_PATH}"        >> /etc/profile
    echo 'export PATH=${PATH}:${SNAP_PATH}/bin' >> /etc/profile
    mkdir -p $SNAP_PATH
    tar zxf $SNAP_FILENAME --strip-components=1 -C $SNAP_PATH
    tar zxf $SNAP_PLUGINS_FILENAME --strip-components=1 -C $SNAP_PATH

    cat << END
------------------------------------------------------------------------
Snap version ${SNAP_RELEASE} has been installed to ${SNAP_PATH}.
${SNAP_PATH}/bin has been appended to \$PATH in /etc/profile.
The Snap plugins have also been installed to ${SNAP_PATH}/plugin.

When you're ready, you can start the snap daemon by running:

  snapd --plugin-trust 0 --log-level 1 --auto-discover \\
    "${SNAP_PATH}/plugin" >> /var/log/snap.log 2>&1 &

or something similar.

------------------------------------------------------------------------
END
}

function main {
    parse_args "$@"
    install_prereqs

    install_zookeeper
    install_mesos
    configure_mesos

    install_golang
    install_snap
}

main "$@"
