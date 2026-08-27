#!/bin/bash
testname="cache runtime status.selector e2e"
dataset_name="selector-demo"
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
    kubectl create -f test/gha-e2e/cacheruntime-selector/minio.yaml
    kubectl create -f test/gha-e2e/cacheruntime-selector/minio_create_bucket.yaml
    wait_job_completed "$bucket_create_job_name"
    kubectl create -f test/gha-e2e/cacheruntime-selector/mount.yaml
}

function create_dataset() {
    kubectl create -f test/gha-e2e/cacheruntime-selector/cacheruntimeclass.yaml
    kubectl create -f test/gha-e2e/cacheruntime-selector/dataset.yaml
    kubectl create -f test/gha-e2e/cacheruntime-selector/cacheruntime.yaml
    if [[ -z "$(kubectl get dataset $dataset_name -oname)" ]]; then
        panic "failed to create dataset $dataset_name"
    fi
    if [[ -z "$(kubectl get cacheruntime $dataset_name -oname)" ]]; then
        panic "failed to create cache runtime $dataset_name"
    fi
    if [[ -z "$(kubectl get cacheruntimeclass $dataset_name -oname)" ]]; then
        panic "failed to create cache runtime class $dataset_name"
    fi
}

function wait_dataset_bound() {
    local deadline=600
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

function wait_cache_worker_ready() {
    local deadline=180
    local worker_component_name="${dataset_name}-worker"
    local worker_selector="cacheruntime.fluid.io/component-name=${worker_component_name}"
    local last_phase=""
    local runtime_ready_replicas=""
    local runtime_desired_replicas=""
    local worker_pod=""
    local worker_registered="false"
    local pod_states=""
    local log_interval=0
    local log_times=0
    while true; do
        last_phase=$(kubectl get cacheruntime "$dataset_name" -ojsonpath='{@.status.worker.phase}')
        runtime_ready_replicas=$(kubectl get cacheruntime "$dataset_name" -ojsonpath='{@.status.worker.readyReplicas}')
        runtime_desired_replicas=$(kubectl get cacheruntime "$dataset_name" -ojsonpath='{@.status.worker.desiredReplicas}')
        worker_pod=$(kubectl get pod -l "$worker_selector" -ojsonpath='{.items[0].metadata.name}' 2>/dev/null)
        worker_registered="false"
        if [[ -n "$worker_pod" ]] && kubectl logs "$worker_pod" -c worker --tail=200 2>/dev/null | grep -q "worker register success"; then
            worker_registered="true"
        fi
        pod_states=$(kubectl get pod -l "$worker_selector" -ojsonpath='{range .items[*]}{.metadata.name}:{range .status.containerStatuses[*]}{.ready}{end}:{.status.phase}{" "}{end}' 2>/dev/null)
        if [[ $log_interval -eq 3 ]]; then
            log_times=$((log_times + 1))
            syslog "checking cache worker readiness (already $((log_times * log_interval * 5))s, runtime phase: ${last_phase:-<empty>}, runtime ready/desired: ${runtime_ready_replicas:-<empty>}/${runtime_desired_replicas:-<empty>}, registered: ${worker_registered}, pods: ${pod_states:-<empty>})"
            if [[ $((log_times * log_interval * 5)) -ge $deadline ]]; then
                panic "timeout waiting for cache worker pod ready after ${deadline}s"
            fi
            log_interval=0
        fi
        if [[ "$last_phase" == "Ready" ]] && \
            [[ -n "$runtime_desired_replicas" ]] && \
            [[ "$runtime_desired_replicas" != "0" ]] && \
            [[ "$runtime_ready_replicas" == "$runtime_desired_replicas" ]] && \
            kubectl wait --for=condition=Ready --timeout=5s pod -l "$worker_selector" >/dev/null 2>&1 && \
            [[ "$worker_registered" == "true" ]]; then
            break
        fi
        log_interval=$((log_interval + 1))
        sleep 5
    done
    syslog "Found ready cache worker pod for $dataset_name"
}

function wait_job_completed() {
    local job_name=$1
    local succeed=""
    local deadline=600
    local counter=0
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
            panic "job $job_name failed (all retries exhausted)"
        fi
        counter=$((counter + 1))
        if [[ $((counter * 5)) -ge $deadline ]]; then
            panic "timeout ${deadline}s waiting for job $job_name to complete"
        fi
        sleep 5
    done
    syslog "Found succeeded job $job_name"
}

function assert_status_selector() {
    local selector
    selector=$(kubectl get cacheruntime "$dataset_name" -ojsonpath='{@.status.selector}')
    syslog "cacheruntime $dataset_name status.selector = ${selector:-<empty>}"

    if [[ -z "$selector" ]]; then
        panic "status.selector is empty"
    fi

    local expected="cacheruntime.fluid.io/component-name=${dataset_name}-worker,cacheruntime.fluid.io/name=${dataset_name}"
    if [[ "$selector" != "$expected" ]]; then
        panic "status.selector mismatch: got '$selector', expected '$expected'"
    fi
    syslog "status.selector matches expected value"
}

function assert_selector_resolves_worker_pods() {
    local selector
    selector=$(kubectl get cacheruntime "$dataset_name" -ojsonpath='{@.status.selector}')

    local pod_count
    pod_count=$(kubectl get pod -l "$selector" --no-headers 2>/dev/null | wc -l)
    if [[ "$pod_count" -lt 1 ]]; then
        panic "no worker pods found via selector '$selector'"
    fi
    syslog "selector '$selector' correctly resolved to $pod_count worker pod(s)"
}

function assert_scale_subresource() {
    local status_selector
    status_selector=$(kubectl get cacheruntime "$dataset_name" -ojsonpath='{@.status.selector}')

    local scale_json
    scale_json=$(kubectl get --raw "/apis/data.fluid.io/v1alpha1/namespaces/default/cacheruntimes/${dataset_name}/scale" 2>/dev/null)
    if [[ -z "$scale_json" ]]; then
        panic "failed to read scale subresource for cacheruntime $dataset_name"
    fi
    syslog "scale subresource: $scale_json"

    local scale_selector
    scale_selector=$(echo "$scale_json" | jq -r '.status.selector')
    if [[ "$scale_selector" != "$status_selector" ]]; then
        panic "scale subresource selector '$scale_selector' does not match status.selector '$status_selector'"
    fi
    syslog "scale subresource selector matches status.selector: $scale_selector"

    local scale_replicas
    scale_replicas=$(echo "$scale_json" | jq -r '.status.replicas')
    if [[ -z "$scale_replicas" || "$scale_replicas" == "null" || "$scale_replicas" -lt 1 ]]; then
        panic "scale subresource status.replicas is invalid: '$scale_replicas'"
    fi
    syslog "scale subresource status.replicas: $scale_replicas"
}

function dump_env_and_clean_up() {
    syslog "Cleaning up resources for testcase $testname"
    kubectl delete --ignore-not-found -f test/gha-e2e/cacheruntime-selector/dataset.yaml
    kubectl delete --ignore-not-found -f test/gha-e2e/cacheruntime-selector/cacheruntime.yaml
    kubectl delete --ignore-not-found -f test/gha-e2e/cacheruntime-selector/cacheruntimeclass.yaml
    kubectl delete --ignore-not-found -f test/gha-e2e/cacheruntime-selector/minio.yaml
    kubectl delete --ignore-not-found -f test/gha-e2e/cacheruntime-selector/mount.yaml
    kubectl delete --ignore-not-found -f test/gha-e2e/cacheruntime-selector/minio_create_bucket.yaml
}

function main() {
    syslog "[TESTCASE $testname STARTS AT $(date)]"
    trap dump_env_and_clean_up EXIT
    setup
    create_dataset
    wait_dataset_bound
    wait_cache_worker_ready
    assert_status_selector
    assert_selector_resolves_worker_pods
    assert_scale_subresource
    syslog "[TESTCASE $testname SUCCEEDED AT $(date)]"
}

main