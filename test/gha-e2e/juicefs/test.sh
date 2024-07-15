#!/bin/bash

testname="juicefsruntime basic e2e"

dataset_name="jfsdemo"
write_job_name="write-job"
read_job_name="read-job"

function syslog() {
    echo ">>> $1"
}

function panic() {
    err_msg=$1
    syslog "test \"$testname\" failed: $err_msg"
    exit 1
}

function setup_redis() {
    kubectl create -f test/gha-e2e/juicefs/redis.yaml
}

function setup_minio() {
    kubectl create -f test/gha-e2e/juicefs/minio.yaml
}

function create_dataset() {
    kubectl create -f test/gha-e2e/juicefs/dataset.yaml

    if [[ -z "$(kubectl get dataset $dataset_name -oname)" ]]; then
        panic "failed to create dataset $dataset_name"
    fi

    if [[ -z "$(kubectl get juicefsruntime $dataset_name -oname)" ]]; then
        panic "failed to create juicefsruntime $dataset_name"
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
            kubectl describe pod
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
    job_file=$1
    job_name=$2
    kubectl create -f $job_file

    if [[ -z "$(kubectl get job $job_name -oname)" ]]; then
        panic "failed to create job $job_name"
    fi
}

function wait_job_completed() {
    job_name=$1
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

function clean_up() {
    syslog "Cleaning up resources for testcase $testname"
    kubectl delete -f test/gha-e2e/juicefs/
}

function main() {
    syslog "[TESTCASE $testname STARTS AT $(date)]"
    setup_redis
    setup_minio
    create_dataset
    trap clean_up EXIT
    wait_dataset_bound
    create_job test/gha-e2e/juicefs/write_job.yaml $write_job_name
    wait_job_completed $write_job_name
    create_job test/gha-e2e/juicefs/read_job.yaml $read_job_name
    wait_job_completed $read_job_name
    syslog "[TESTCASE $testname SUCCEEDED AT $(date)]"
}

main



