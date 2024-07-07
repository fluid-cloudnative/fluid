#!/bin/bash

function syslog() {
    echo ">>> $1"
}

function check_control_plane_status() {
    while true; do
        total_pods=$(kubectl get pod -n fluid-system --no-headers | grep -cv "Completed")
        running_pods=$(kubectl get pod -n fluid-system --no-headers | grep -c "Running")

        if [[ $total_pods -ne 0 ]]; then
            if [[ $total_pods -eq $running_pods ]]; then
                break
            fi
        fi
        sleep 5
    done
    syslog "Fluid control plane is ready!"
}

function alluxio_e2e() {
    set -e
    docker pull alluxio/alluxio-dev:2.9.0
    kind load docker-image alluxio/alluxio-dev:2.9.0 --name ${KIND_CLUSTER}
    docker image prune -a -f
    bash test/gha-e2e/alluxio/test.sh
}

check_control_plane_status
alluxio_e2e
