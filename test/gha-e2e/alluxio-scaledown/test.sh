#!/bin/bash

testname="alluxioruntime graceful scale-down e2e"

dataset_name="scaledown-demo"
worker_sts_name="scaledown-demo-worker"
controller_deployment="alluxioruntime-controller"
controller_namespace="fluid-system"
read_before_job_name="read-before-scaledown"
read_after_job_name="read-after-scaledown"

function syslog() {
    echo ">>> $1"
}

function panic() {
    local err_msg=$1
    syslog "test \"$testname\" failed: $err_msg"
    exit 1
}

# GracefulWorkerScaleDown is Alpha and disabled by default; the controller
# binary has no Helm value wired up for it yet, so enable it directly on the
# running deployment for this scenario.
function enable_graceful_scale_down() {
    if kubectl get deployment "$controller_deployment" -n "$controller_namespace" \
        -ojsonpath='{.spec.template.spec.containers[0].args}' | grep -q "feature-gates=GracefulWorkerScaleDown=true"; then
        syslog "GracefulWorkerScaleDown feature gate already enabled"
        return
    fi

    kubectl patch deployment "$controller_deployment" -n "$controller_namespace" --type=json \
        -p '[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--feature-gates=GracefulWorkerScaleDown=true"}]'

    kubectl rollout status deployment/"$controller_deployment" -n "$controller_namespace" --timeout=120s \
        || panic "alluxioruntime-controller did not roll out after enabling the feature gate"

    syslog "Enabled GracefulWorkerScaleDown feature gate on $controller_deployment"
}

function create_dataset() {
    kubectl create -f test/gha-e2e/alluxio-scaledown/dataset.yaml

    if [[ -z "$(kubectl get dataset $dataset_name -oname)" ]]; then
        panic "failed to create dataset $dataset_name"
    fi

    if [[ -z "$(kubectl get alluxioruntime $dataset_name -oname)" ]]; then
        panic "failed to create alluxioruntime $dataset_name"
    fi
}

function wait_dataset_bound() {
    local deadline=300 # 5 minutes
    local last_state=""
    local counter=0
    while true; do
        last_state=$(kubectl get dataset $dataset_name -ojsonpath='{@.status.phase}')
        if [[ "$last_state" == "Bound" ]]; then
            break
        fi

        if [[ $((counter % 3)) -eq 0 ]]; then
            syslog "checking dataset.status.phase==Bound (already $((counter * 5))s, last state: $last_state)"
        fi

        counter=$((counter + 1))
        if [[ $((counter * 5)) -ge $deadline ]]; then
            panic "timeout for ${deadline}s!"
        fi
        sleep 5
    done
    syslog "Found dataset $dataset_name status.phase==Bound"
}

function wait_worker_replicas() {
    local expected=$1
    # Generous enough to cover both a normal drain and the
    # defaultWorkerDecommissionDeadline (10m) forced-proceed fallback.
    local deadline=900
    local spec_replicas=""
    local status_replicas=""
    local counter=0
    while true; do
        spec_replicas=$(kubectl get statefulset "$worker_sts_name" -ojsonpath='{@.spec.replicas}' 2>/dev/null)
        status_replicas=$(kubectl get statefulset "$worker_sts_name" -ojsonpath='{@.status.replicas}' 2>/dev/null)

        if [[ "$spec_replicas" == "$expected" ]] && [[ "$status_replicas" == "$expected" ]]; then
            break
        fi

        if [[ $((counter % 6)) -eq 0 ]]; then
            syslog "waiting for $worker_sts_name to reach $expected replicas (already $((counter * 5))s, spec=$spec_replicas status=$status_replicas)"
        fi

        counter=$((counter + 1))
        if [[ $((counter * 5)) -ge $deadline ]]; then
            panic "timeout ${deadline}s waiting for $worker_sts_name to reach $expected replicas"
        fi
        sleep 5
    done
    syslog "$worker_sts_name reached $expected replicas"
}

function scale_down() {
    kubectl patch alluxioruntime "$dataset_name" --type=merge -p '{"spec":{"replicas":1}}'
    syslog "Requested scale-down of $dataset_name to 1 replica"
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
    local deadline=300
    local counter=0
    local succeed=""
    local job_failed=""
    while true; do
        succeed=$(kubectl get job "$job_name" -ojsonpath='{@.status.succeeded}')
        [[ -z "$succeed" ]] && succeed=0

        if [[ "$succeed" -ge "1" ]]; then
            break
        fi

        job_failed=$(kubectl get job "$job_name" \
            -ojsonpath='{.status.conditions[?(@.type=="Failed")].status}' 2>/dev/null || true)
        if [[ "$job_failed" == "True" ]]; then
            panic "job $job_name failed when accessing data (all retries exhausted)"
        fi

        counter=$((counter + 1))
        if [[ $((counter * 5)) -ge $deadline ]]; then
            panic "timeout ${deadline}s waiting for job $job_name to complete"
        fi
        sleep 5
    done
    syslog "Found succeeded job $job_name"
}

function dump_env_and_clean_up() {
    bash tools/diagnose-fluid-alluxio.sh collect --name $dataset_name --namespace default --collect-path ./e2e-tmp/testcase-alluxio-scaledown.tgz
    syslog "Cleaning up resources for testcase $testname"
    kubectl delete --ignore-not-found -f test/gha-e2e/alluxio-scaledown/read_after_job.yaml
    kubectl delete --ignore-not-found -f test/gha-e2e/alluxio-scaledown/read_before_job.yaml
    kubectl delete --ignore-not-found -f test/gha-e2e/alluxio-scaledown/dataset.yaml
}

function main() {
    syslog "[TESTCASE $testname STARTS AT $(date)]"
    trap dump_env_and_clean_up EXIT
    enable_graceful_scale_down
    create_dataset
    wait_dataset_bound
    wait_worker_replicas 2
    create_job test/gha-e2e/alluxio-scaledown/read_before_job.yaml $read_before_job_name
    wait_job_completed $read_before_job_name
    scale_down
    wait_worker_replicas 1
    create_job test/gha-e2e/alluxio-scaledown/read_after_job.yaml $read_after_job_name
    wait_job_completed $read_after_job_name
    syslog "[TESTCASE $testname SUCCEEDED AT $(date)]"
}

main
