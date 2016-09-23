#!/bin/bash

set -e
set -u
set -o pipefail

TRAVIS=${TRAVIS:-}

if [[ $TRAVIS == "true" ]]; then
    _info "Provisioning Mesos master and agent..."
    export SNAP_MESOS_MASTER="127.0.0.1:5050"
    export SNAP_MESOS_AGENT="127.0.0.1:5051"

    sudo ./scripts/provision-travis.sh --mesos_release "${MESOS_RELEASE}" --ip_address 127.0.0.1

    UNIT_TEST="go_test"
    test_unit
else
    _info "Not running in Travis CI. Skipping medium test."
fi
