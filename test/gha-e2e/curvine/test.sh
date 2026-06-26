#!/bin/bash

testname="curvine cache runtime basic e2e"

dataset_name="curvine-demo"
ref_dataset_name="curvine-demo-ref"
write_job_name="write-job"
read_job_name="read-job"
read_ref_job_name="read-ref-job"
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
    local deadline=600 # 10 minutes
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

function wait_cache_client_ready() {
    local deadline=180 # 3 minutes
    local client_component_name="${dataset_name}-client"
    local client_selector="cacheruntime.fluid.io/component-name=${client_component_name}"
    local last_phase=""
    local runtime_ready_replicas=""
    local runtime_desired_replicas=""
    local ds_ready_replicas=""
    local ds_desired_replicas=""
    local pod_states=""
    local log_interval=0
    local log_times=0

    while true; do
        last_phase=$(kubectl get cacheruntime "$dataset_name" -ojsonpath='{@.status.client.phase}')
        runtime_ready_replicas=$(kubectl get cacheruntime "$dataset_name" -ojsonpath='{@.status.client.readyReplicas}')
        runtime_desired_replicas=$(kubectl get cacheruntime "$dataset_name" -ojsonpath='{@.status.client.desiredReplicas}')
        ds_ready_replicas=$(kubectl get daemonset "$client_component_name" -ojsonpath='{@.status.numberReady}' 2>/dev/null)
        ds_desired_replicas=$(kubectl get daemonset "$client_component_name" -ojsonpath='{@.status.desiredNumberScheduled}' 2>/dev/null)
        pod_states=$(kubectl get pod -l "$client_selector" -ojsonpath='{range .items[*]}{.metadata.name}:{range .status.containerStatuses[*]}{.ready}{end}:{.status.phase}{" "}{end}' 2>/dev/null)

        if [[ $log_interval -eq 3 ]]; then
            log_times=$((log_times + 1))
            syslog "checking cache client readiness (already $((log_times * log_interval * 5))s, runtime phase: ${last_phase:-<empty>}, runtime ready/desired: ${runtime_ready_replicas:-<empty>}/${runtime_desired_replicas:-<empty>}, ds ready/desired: ${ds_ready_replicas:-<empty>}/${ds_desired_replicas:-<empty>}, pods: ${pod_states:-<empty>})"
            if [[ $((log_times * log_interval * 5)) -ge $deadline ]]; then
                panic "timeout waiting for cache client pod ready after ${deadline}s"
            fi
            log_interval=0
        fi

        if kubectl rollout status daemonset/"$client_component_name" --timeout=5s >/dev/null 2>&1 && \
            kubectl wait --for=condition=Ready --timeout=5s pod -l "$client_selector" >/dev/null 2>&1; then
            break
        fi

        log_interval=$((log_interval + 1))
        sleep 5
    done

    syslog "Found ready cache client pod for $dataset_name"
}

function wait_cache_worker_ready() {
    local deadline=180 # 3 minutes
    local worker_component_name="${dataset_name}-worker"
    local worker_selector="cacheruntime.fluid.io/component-name=${worker_component_name}"
    local last_phase=""
    local runtime_ready_replicas=""
    local runtime_desired_replicas=""
    local asts_ready_replicas=""
    local asts_desired_replicas=""
    local worker_pod=""
    local worker_registered="false"
    local pod_states=""
    local log_interval=0
    local log_times=0

    while true; do
        last_phase=$(kubectl get cacheruntime "$dataset_name" -ojsonpath='{@.status.worker.phase}')
        runtime_ready_replicas=$(kubectl get cacheruntime "$dataset_name" -ojsonpath='{@.status.worker.readyReplicas}')
        runtime_desired_replicas=$(kubectl get cacheruntime "$dataset_name" -ojsonpath='{@.status.worker.desiredReplicas}')
        asts_ready_replicas=$(kubectl get advancedstatefulset "$worker_component_name" -ojsonpath='{@.status.readyReplicas}' 2>/dev/null)
        asts_desired_replicas=$(kubectl get advancedstatefulset "$worker_component_name" -ojsonpath='{@.spec.replicas}' 2>/dev/null)
        worker_pod=$(kubectl get pod -l "$worker_selector" -ojsonpath='{.items[0].metadata.name}' 2>/dev/null)
        worker_registered="false"
        if [[ -n "$worker_pod" ]] && kubectl logs "$worker_pod" -c worker --tail=200 2>/dev/null | grep -q "worker register success"; then
            worker_registered="true"
        fi
        pod_states=$(kubectl get pod -l "$worker_selector" -ojsonpath='{range .items[*]}{.metadata.name}:{range .status.containerStatuses[*]}{.ready}{end}:{.status.phase}{" "}{end}' 2>/dev/null)

        if [[ $log_interval -eq 3 ]]; then
            log_times=$((log_times + 1))
            syslog "checking cache worker readiness (already $((log_times * log_interval * 5))s, runtime phase: ${last_phase:-<empty>}, runtime ready/desired: ${runtime_ready_replicas:-<empty>}/${runtime_desired_replicas:-<empty>}, advanced sts ready/desired: ${asts_ready_replicas:-<empty>}/${asts_desired_replicas:-<empty>}, registered: ${worker_registered}, pods: ${pod_states:-<empty>})"
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

function create_reference_dataset() {
    # Note: ThinRuntime will be automatically created by Dataset Controller
    # when a reference dataset (mountPoint: dataset://...) is created
    kubectl create -f test/gha-e2e/curvine/ref-dataset.yaml

    if [[ -z "$(kubectl get dataset $ref_dataset_name -oname)" ]]; then
        panic "failed to create reference dataset $ref_dataset_name"
    fi
}

function wait_reference_dataset_bound() {
    local deadline=600 # 10 minutes
    local last_state=""
    local log_interval=0
    local log_times=0
    while true; do
        last_state=$(kubectl get dataset $ref_dataset_name -ojsonpath='{@.status.phase}')
        if [[ $log_interval -eq 3 ]]; then
            log_times=$((log_times + 1))
            syslog "checking reference dataset.status.phase==Bound (already $((log_times * log_interval * 5))s, last state: $last_state)"
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
    syslog "Found reference dataset $ref_dataset_name status.phase==Bound"
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

        # Only fail when the job controller itself marks the job as Failed
        # (i.e. all backoffLimit retries are exhausted), not on first pod failure.
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
    local exit_code=$?
    if [[ $exit_code -ne 0 ]]; then
        bash tools/diagnose-fluid-curvine.sh collect --name $dataset_name --namespace default --collect-path ./e2e-tmp/testcase-curvine.tgz
        syslog "=== Diagnostic logs for failed test ==="
        syslog "--- cacheruntime-controller logs (last 100 lines) ---"
        kubectl logs -n fluid-system -l control-plane=cacheruntime-controller -c manager --tail=100 2>&1 || true
        syslog "--- CacheRuntime describe ---"
        kubectl describe cacheruntime $dataset_name 2>&1 || true
        syslog "--- Dataset describe ---"
        kubectl describe dataset $dataset_name 2>&1 || true
        syslog "--- Pods in default namespace ---"
        kubectl get pods -n default -owide 2>&1 || true
        syslog "--- Events in default namespace ---"
        kubectl get events -n default --sort-by='.lastTimestamp' 2>&1 || true
        syslog "=== End of diagnostic logs ==="
    fi
    syslog "Cleaning up resources for testcase $testname"
    kubectl delete --ignore-not-found -f test/gha-e2e/curvine/read_ref_job.yaml
    kubectl delete --ignore-not-found -f test/gha-e2e/curvine/read_job.yaml
    kubectl delete --ignore-not-found -f test/gha-e2e/curvine/write_job.yaml
    kubectl delete --ignore-not-found -f test/gha-e2e/curvine/dataload.yaml
    kubectl delete --ignore-not-found -f test/gha-e2e/curvine/ref-dataset.yaml
    # ThinRuntime will be automatically deleted when reference dataset is deleted
    kubectl delete --ignore-not-found -f test/gha-e2e/curvine/dataset.yaml
    kubectl delete --ignore-not-found -f test/gha-e2e/curvine/cacheruntime.yaml
    kubectl delete --ignore-not-found -f test/gha-e2e/curvine/cacheruntimeclass.yaml
    kubectl delete --ignore-not-found -f test/gha-e2e/curvine/minio.yaml
    kubectl delete --ignore-not-found -f test/gha-e2e/curvine/mount.yaml
    kubectl delete --ignore-not-found -f test/gha-e2e/curvine/minio_create_bucket.yaml
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
function scale_worker_and_verify() {
    local target_replicas=$1
    local worker_component_name="${dataset_name}-worker"
    local deadline=180
    local counter=0

    syslog "Scaling worker to $target_replicas replicas"
    kubectl patch cacheruntime $dataset_name --type merge -p "{\"spec\":{\"worker\":{\"replicas\":$target_replicas}}}"

    while true; do
        local ready_replicas=""
        local desired_replicas=""
        ready_replicas=$(kubectl get cacheruntime $dataset_name -ojsonpath='{@.status.worker.readyReplicas}' 2>/dev/null)
        desired_replicas=$(kubectl get cacheruntime $dataset_name -ojsonpath='{@.status.worker.desiredReplicas}' 2>/dev/null)

        if [[ "$ready_replicas" == "$target_replicas" ]] && [[ "$desired_replicas" == "$target_replicas" ]]; then
            break
        fi

        counter=$((counter + 1))
        if [[ $((counter * 5)) -ge $deadline ]]; then
            panic "timeout ${deadline}s waiting for worker scale to $target_replicas (ready: ${ready_replicas:-<empty>}, desired: ${desired_replicas:-<empty>})"
        fi
        sleep 5
    done

    # verify AdvancedStatefulSet replicas match
    local asts_replicas=""
    asts_replicas=$(kubectl get advancedstatefulset "$worker_component_name" -ojsonpath='{@.spec.replicas}' 2>/dev/null)
    if [[ "$asts_replicas" != "$target_replicas" ]]; then
        panic "AdvancedStatefulSet replicas mismatch: expected $target_replicas, got $asts_replicas"
    fi

    local asts_ready=""
    asts_ready=$(kubectl get advancedstatefulset "$worker_component_name" -ojsonpath='{@.status.readyReplicas}' 2>/dev/null)
    if [[ "$asts_ready" != "$target_replicas" ]]; then
        panic "AdvancedStatefulSet readyReplicas mismatch: expected $target_replicas, got $asts_ready"
    fi

    syslog "Worker scaled to $target_replicas replicas successfully (asts replicas=$asts_replicas, ready=$asts_ready)"
}

function delete_dataset_and_runtime() {
    kubectl delete --ignore-not-found -f test/gha-e2e/curvine/dataload.yaml
    kubectl delete --ignore-not-found -f test/gha-e2e/curvine/ref-dataset.yaml
    kubectl delete -f test/gha-e2e/curvine/dataset.yaml
    kubectl delete -f test/gha-e2e/curvine/cacheruntime.yaml
}

function wait_runtime_deleted() {
    local deadline=120
    local counter=0
    while true; do
        local remaining=""
        remaining=$(kubectl get advancedstatefulset,daemonset,svc -l fluid.io/managed-by=fluid -n default -oname 2>/dev/null)
        if [[ -z "$remaining" ]]; then
            break
        fi
        counter=$((counter + 1))
        if [[ $((counter * 5)) -ge $deadline ]]; then
            syslog "remaining resources after deletion: $remaining"
            panic "timeout ${deadline}s waiting for runtime resources to be garbage collected"
        fi
        sleep 5
    done
    syslog "All runtime resources (AdvancedStatefulSet, DaemonSet, Service) garbage collected"

    # verify node labels are cleaned up
    local cache_labels=""
    cache_labels=$(kubectl get nodes -ojsonpath='{range .items[*]}{.metadata.labels}' 2>/dev/null | grep -o "fluid.io/s-default-${dataset_name}[^ ]*" || true)
    if [[ -n "$cache_labels" ]]; then
        panic "node labels not cleaned up: $cache_labels"
    fi
    syslog "Node labels cleaned up successfully"

    # verify PV/PVC are cleaned up
    if kubectl get pvc $dataset_name -n default >/dev/null 2>&1; then
        panic "PVC $dataset_name still exists after deletion"
    fi
    if kubectl get pv default-$dataset_name >/dev/null 2>&1; then
        panic "PV default-$dataset_name still exists after deletion"
    fi
    syslog "PV/PVC cleaned up successfully"
}

function main() {
    syslog "[TESTCASE $testname STARTS AT $(date)]"
    trap dump_env_and_clean_up EXIT
    setup
    create_dataset
    wait_dataset_bound
    create_reference_dataset
    wait_reference_dataset_bound
    wait_cache_worker_ready
    create_job test/gha-e2e/curvine/write_job.yaml $write_job_name
    wait_job_completed $write_job_name
    create_dataload
    wait_dataload_completed "curvine-dataload"
    wait_cache_client_ready
    create_job test/gha-e2e/curvine/read_job.yaml $read_job_name
    wait_job_completed $read_job_name
    create_job test/gha-e2e/curvine/read_ref_job.yaml $read_ref_job_name
    wait_job_completed $read_ref_job_name

    # verify scale-up and scale-down (exercises patch on advancedstatefulsets and delete on pods)
    # skip on single-node clusters where anti-affinity prevents scheduling the second pod
    local node_count=""
    node_count=$(kubectl get nodes --no-headers 2>/dev/null | wc -l | tr -d ' ')
    if [[ "$node_count" -ge 2 ]]; then
        scale_worker_and_verify 2
        scale_worker_and_verify 1
    else
        syslog "Skipping scale test: single-node cluster ($node_count node)"
    fi

    # verify deletion and cleanup
    delete_dataset_and_runtime
    wait_runtime_deleted

    syslog "[TESTCASE $testname SUCCEEDED AT $(date)]"
}

main
