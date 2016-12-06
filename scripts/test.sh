#!/bin/bash -e
# The script does automatic checking on a Go package and its sub-packages, including:
#   - gofmt         (https://golang.org/cmd/gofmt)
#   - goimports     (https://godoc.org/cmd/goimports)
#   - go vet        (https://golang.org/cmd/vet)
#   - test coverage (https://blog.golang.org/cover)

COVERALLS_MAX_ATTEMPTS=5
TEST_DIRS="main.go mesos/"
PKG_DIRS=". ./mesos/..."
IGNORE_PKGS="mesos_pb2"

function _gofmt {
    echo "Running 'gofmt'"
    test -z "$(gofmt -l -d $TEST_DIRS | tee /dev/stderr)"
}

function _goimports {
    echo "Running 'goimports'"
    go get golang.org/x/tools/cmd/goimports
    test -z "$(goimports -l -d $TEST_DIRS | tee /dev/stderr)"
}

function _govet {
    echo "Running 'go vet'"
    go vet $PKG_DIRS
}

function _unit_test_with_coverage {
    echo "Running unit tests..."

    go get github.com/smartystreets/goconvey/convey
    go get golang.org/x/tools/cmd/cover

    # As of Go 1.6, we cannot use the test profile flag with multiple packages.
    # Therefore, we run 'go test' for each package, and concatenate the results
    # into 'profile.cov'.
    echo "mode: count" > profile.cov
    mkdir -p ./tmp

    for import_path in $(go list -f={{.ImportPath}} ${PKG_DIRS}); do
        package=$(basename ${import_path})
        [[ "$IGNORE_PKGS" =~ $package ]] && continue
        go test -v --tags=small -covermode=count -coverprofile=./tmp/profile_${package}.cov $import_path
    done

    for f in ./tmp/profile_*.cov; do
        tail -n +2 ${f} >> profile.cov
    done

    rm -rf ./tmp
    go tool cover -func profile.cov
}

function _submit_to_coveralls {
    # Only submit to Coveralls.io if we're running in Travis CI. We don't want
    # this happening on dev machines! Note that the Coveralls repo token is
    # available via the $COVERALLS_REPO_TOKEN environment variable, which is
    # configured for the project in the Travis CI web interface.
    if [[ $TRAVIS == "true" && $COVERALLS_REPO_TOKEN != "" ]]; then
        go get github.com/mattn/goveralls

        for attempt in {1..${COVERALLS_MAX_ATTEMPTS}}; do
            echo "Posting test coverage to Coveralls, attempt ${attempt} of ${COVERALLS_MAX_ATTEMPTS}"
            goveralls -v -coverprofile=profile.cov -service=travis-ci -repotoken ${COVERALLS_REPO_TOKEN} && break
        done
    else
        echo "Not running in Travis CI (environment variables unset), not posting test coverage to Coveralls!"
    fi
}

function _integration_test {
    echo "Running integration tests..."

    if [[ $TRAVIS == "true" ]]; then
        echo "Detected that we're running in Travis CI. Provisioning Mesos master and agent..."
        export SNAP_MESOS_MASTER="127.0.0.1:5050"
        export SNAP_MESOS_AGENT="127.0.0.1:5051"

        sudo ./scripts/provision-travis.sh --mesos_release ${MESOS_RELEASE} --ip_address 127.0.0.1

        sleep 10
        set +e
        sudo service mesos-master status
        sudo service mesos-slave status
        sudo tail -100 /var/log/syslog
        curl http://127.0.0.1:5050
        curl http://127.0.0.1:5051
        set -e
    else
        echo "Detected that we aren't running in Travis CI. Skipping provisioning of Mesos master and agent..."
    fi

    go test -v --tags=medium ./...
}

function main {
    TEST_SUITE="$1"

    if [[ $TEST_SUITE == "small" ]]; then
        _gofmt
        _goimports
        _govet
        _unit_test_with_coverage
        _submit_to_coveralls
    elif [[ $TEST_SUITE == "medium" ]]; then
        _integration_test
    elif [[ $TEST_SUITE == "build" ]]; then
        ./scripts/build.sh
    else
        echo "Error: unknown test suite ${TEST_SUITE}"
        exit 1
    fi
}

main "$@"
