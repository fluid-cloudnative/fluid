#!/bin/bash

testname="jindoruntime basic e2e"

dataset_name="jindo-demo"
job_name="fluid-test"

function syslog() {
    echo ">>> $1"
}

function panic() {
    err_msg=$1
    syslog "test \"$testname\" failed: $err_msg"
    exit 1
}

function setup_minio() {
    kubectl create -f test/gha-e2e/jindo/minio.yaml
    minio_pod=$(kubectl get pod -oname | grep minio) 
    kubectl wait --for=condition=Ready $minio_pod

    kubectl exec -it $minio_pod -- /bin/bash -c 'mc alias set myminio http://127.0.0.1:9000 minioadmin minioadmin && mc mb myminio/mybucket && echo "helloworld" > testfile && mc mv testfile myminio/mybucket/subpath/testfile && mc cat myminio/mybucket/subpath/testfile'
}

function create_dataset() {
    kubectl create -f test/gha-e2e/jindo/dataset.yaml

    if [[ -z "$(kubectl get dataset $dataset_name -oname)" ]]; then
        panic "failed to create dataset $dataset_name"
    fi

    if [[ -z "$(kubectl get jindoruntime $dataset_name -oname)" ]]; then
        panic "failed to create jindoruntime $dataset_name"
    fi
}

function wait_dataset_bound() {
    deadline=180 # 3 minutes
    last_state=""
    log_interval=0
    log_times=0
    while true; do
        last_state=$(kubectl get dataset $dataset_name -ojsonpath='{@.status.phase}')
        if [[ $log_interval -eq 3 ]]; then
            log_times=$(expr $log_times + 1)
            syslog "checking dataset.status.phase==Bound (already $(expr $log_times \* $log_interval \* 5)s, last state: $last_state)"
            if [[ "$(expr $log_times \* $log_interval \* 5)" -ge "$deadline" ]]; then
                panic "timeout for ${deadline}s!"
            fi
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
    kubectl create -f test/gha-e2e/jindo/job.yaml

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
    bash tools/diagnose-fluid-jindo.sh collect --name $dataset_name --namespace default --collect-path ./e2e-tmp/testcase-jindo.tgz
    syslog "Cleaning up resources for testcase $testname"
    kubectl delete -f test/gha-e2e/jindo/
}

function main() {
    syslog "[TESTCASE $testname STARTS AT $(date)]"
    setup_minio
    create_dataset
    trap dump_env_and_clean_up EXIT
    wait_dataset_bound
    create_job
    wait_job_completed
    syslog "[TESTCASE $testname SUCCEEDED AT $(date)]"
}

main
