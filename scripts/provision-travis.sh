#!/bin/bash -e
# This script provisions a single-node Mesos cluster that can be used for
# running integration tests for the Mesos plugin in Travis CI.

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
            --ip_address)     IP_ADDRESS="${2:-127.0.0.1}"      ; shift  ;;
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
    echo "zk://${IP_ADDRESS}:2181/mesos" > /etc/mesos/zk
    echo "${IP_ADDRESS}"                 > /etc/mesos-master/hostname
    echo "${IP_ADDRESS}"                 > /etc/mesos-master/ip
    service mesos-master restart

    # Agent
    # Note: there is a known bug in Mesos when using the 'cgroups/perf_event' isolator
    # on specific kernels and platforms. For more info, see MESOS-4705.
    echo "mesos"                                                 > /etc/mesos-slave/containerizers
    echo "${IP_ADDRESS}"                                         > /etc/mesos-slave/hostname
    echo "${IP_ADDRESS}"                                         > /etc/mesos-slave/ip
    echo "cgroups/cpu,cgroups/mem,cgroups/perf_event,posix/disk" > /etc/mesos-slave/isolation
    echo "cpu-clock,task-clock,context-switches"                 > /etc/mesos-slave/perf_events
    echo "/var/lib/mesos"                                        > /etc/mesos-slave/work_dir
    service mesos-slave restart

    cat << END
--------------------------------------------------
Mesos version ${MESOS_RELEASE} has been installed.

  * Master: http://${IP_ADDRESS}:5050
  * Agent: http://${IP_ADDRESS}:5051

--------------------------------------------------
END
}

function main {
    parse_args "$@"
    install_prereqs

    install_zookeeper
    install_mesos
    configure_mesos
}

main "$@"
