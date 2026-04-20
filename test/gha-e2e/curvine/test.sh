#!/bin/bash

testname="curvine cache runtime basic e2e"

dataset_name="curvine-demo"
write_job_name="write-job"
read_job_name="read-job"
bucket_create_job_name="minio-bucket-create"

function syslog() {
    echo ">>> $1"
}

function panic() {
    local err_msg=$1
    syslog "test \"$testname\" failed: $err_msg"
    exit 1
}

function setup() {
    # minio 需要有 bucket 才能被 curvine 挂载
    kubectl create -f test/gha-e2e/curvine/minio.yaml

    kubectl create -f test/gha-e2e/curvine/minio_create_bucket.yaml
    wait_job_completed "$bucket_create_job_name"

    kubectl create -f test/gha-e2e/curvine/mount.yaml
}

function create_dataset() {
    kubectl create -f test/gha-e2e/curvine/cacheruntimeclass.yaml
    kubectl create -f test/gha-e2e/curvine/dataset.yaml
    kubectl create -f test/gha-e2e/curvine/cacheruntime.yaml

    if [[ -z "$(kubectl get dataset $dataset_name -oname)" ]]; then
        panic "failed to create dataset $dataset_name"
    fi

    if [[ -z "$(kubectl get cacheruntime $dataset_name -oname)" ]]; then
        panic "failed to create curvine cache runtime $dataset_name"
    fi

    if [[ -z "$(kubectl get cacheruntimeclass $dataset_name -oname)" ]]; then
        panic "failed to create curvine cache runtime class $dataset_name"
    fi

}

function wait_dataset_bound() {
    local deadline=180 # 3 minutes
    local last_state=""
    local log_interval=0
    local log_times=0
    while true; do
        last_state=$(kubectl get dataset $dataset_name -ojsonpath='{@.status.phase}')
        if [[ $log_interval -eq 3 ]]; then
            log_times=$((log_times + 1))
            syslog "checking dataset.status.phase==Bound (already $((log_times * log_interval * 5))s, last state: $last_state)"
            if [[ $((log_times * log_interval * 5)) -ge $deadline ]]; then
                panic "timeout for ${deadline}s!"
            fi
            log_interval=0
        fi

        if [[ "$last_state" == "Bound" ]]; then
            break
        fi
        log_interval=$((log_interval + 1))
        sleep 5
    done
    syslog "Found dataset $dataset_name status.phase==Bound"
}

function create_job() {
    local job_file=$1
    local job_name=$2
    kubectl create -f "$job_file"

    if [[ -z "$(kubectl get job "$job_name" -oname)" ]]; then
        panic "failed to create job $job_name"
    fi
}

function wait_job_completed() {
    local job_name=$1
    while true; do
        succeed=$(kubectl get job "$job_name" -ojsonpath='{@.status.succeeded}')
        failed=$(kubectl get job "$job_name" -ojsonpath='{@.status.failed}')
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
    bash tools/diagnose-fluid-curvine.sh collect --name $dataset_name --namespace default --collect-path ./e2e-tmp/testcase-curvine.tgz
    syslog "Cleaning up resources for testcase $testname"
    kubectl delete -f test/gha-e2e/curvine/read_job.yaml
    kubectl delete -f test/gha-e2e/curvine/write_job.yaml
    kubectl delete -f test/gha-e2e/curvine/dataload.yaml
    kubectl delete -f test/gha-e2e/curvine/dataset.yaml
    kubectl delete -f test/gha-e2e/curvine/cacheruntime.yaml
    kubectl delete -f test/gha-e2e/curvine/cacheruntimeclass.yaml
    kubectl delete -f test/gha-e2e/curvine/minio.yaml
    kubectl delete -f test/gha-e2e/curvine/mount.yaml
    kubectl delete -f test/gha-e2e/curvine/minio_create_bucket.yaml
}

function create_dataload() {
    kubectl create -f test/gha-e2e/curvine/dataload.yaml
}

function wait_dataload_completed() {
    local dataload_name=$1
    local log_interval=0
    local status
    while true; do
        status=$(kubectl get dataload "$dataload_name" -ojsonpath='{@.status.phase}')
        if [[ "$status" == "Complete" ]]; then
            syslog "dataload $dataload_name status.phase==Complete"
            break
        fi
        # wait at most 60 seconds
        if [[ $log_interval -ge 12 ]]; then
            panic "dataload $dataload_name status is $status, not complete for 60s!"
        fi
        sleep 5
        log_interval=$((log_interval + 1))
    done
    syslog "Found succeeded dataload_name $dataload_name"
}

function main() {
    syslog "[TESTCASE $testname STARTS AT $(date)]"
    trap dump_env_and_clean_up EXIT
    setup
    create_dataset
    wait_dataset_bound

    create_job test/gha-e2e/curvine/write_job.yaml $write_job_name
    wait_job_completed $write_job_name

    create_dataload 
    wait_dataload_completed "curvine-dataload"

    create_job test/gha-e2e/curvine/read_job.yaml $read_job_name
    wait_job_completed $read_job_name

    syslog "[TESTCASE $testname SUCCEEDED AT $(date)]"
}

main
