#!/bin/bash

testname="alluxioruntime basic e2e"

dataset_name="zookeeper"
job_name="fluid-test"

function syslog() {
    echo ">>> $1"
}

function panic() {
    err_msg=$1
    syslog "test \"$testname\" failed: $err_msg"
    exit 1
}

function create_dataset() {
    kubectl create -f test/gha-e2e/alluxio/dataset.yaml

    if [[ -z "$(kubectl get dataset $dataset_name -oname)" ]]; then
        panic "failed to create dataset"
    fi

    if [[ -z "$(kubectl get alluxioruntime $dataset_name -oname)" ]]; then
        panic "failed to create alluxioruntime"
    fi
}

function wait_dataset_bound() {
    last_state=""
    log_interval=0
    log_times=0
    while true; do
        last_state=$(kubectl get dataset $dataset_name -ojsonpath='{@.status.phase}')
        if [[ $log_interval -eq 3 ]]; then
            log_times=$(expr $log_times + 1)
            syslog "checking dataset.status.phase==Bound (already $(expr $log_times \* $log_interval \* 5)s, last state: $last_state)"
            log_interval=0
        fi

        if [[ "$last_state" == "Bound" ]]; then
            break
        fi
        log_interval=$(expr $log_interval + 1)
        sleep 5
    done
    syslog "Found dataset $dataset_name status.phase==Bound"
}

function create_job() {
    kubectl create -f test/gha-e2e/alluxio/job.yaml

    if [[ -z "$(kubectl get job $job_name -oname)" ]]; then
        panic "failed to create job"
    fi
}

function wait_job_completed() {
    while true; do
        succeed=$(kubectl get job $job_name -ojsonpath='{@.status.succeeded}')
        failed=$(kubectl get job $job_name -ojsonpath='{@.status.failed}')
        if [[ "$failed" -ne "0" ]]; then
            panic "job failed when accessing data"
        fi
        if [[ "$succeed" -eq "1" ]]; then
            break
        fi
        sleep 5
    done
    syslog "Found succeeded job $job_name"
}

function dump_env_and_clean_up() {
    bash tools/diagnose-fluid-alluxio.sh collect --name $dataset_name --namespace default --collect-path ./e2e-tmp/testcase-alluxio.tgz
    syslog "Cleaning up resources for testcase $testname"
    kubectl delete -f test/gha-e2e/alluxio/
}

function main() {
    syslog "[TESTCASE $testname STARTS AT $(date)]"
    create_dataset
    trap dump_env_and_clean_up EXIT
    wait_dataset_bound
    create_job
    wait_job_completed
    syslog "[TESTCASE $testname SUCCEEDED AT $(date)]"
}

main
