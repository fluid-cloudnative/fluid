#!/bin/bash

# Mooncake is a cache system without POSIX mount semantics; its
# CacheRuntimeClass declares only master and worker, and no client. This case
# guards exactly that scenario: when the client component is omitted the
# controller must not panic and the Dataset must still reach Bound. The curvine
# case ships a client component and does not cover this path.

testname="mooncake cache runtime (client-less topology) e2e"

dataset_name="mooncake-demo"
rw_job_name="mooncake-rw-job"
bad_mount_pod_name="mooncake-bad-mount"
testdir="test/gha-e2e/mooncake"

function syslog() {
    echo ">>> $1"
}

function panic() {
    local err_msg=$1
    syslog "test \"$testname\" failed: $err_msg"
    exit 1
}

function create_dataset() {
    kubectl create -f $testdir/cacheruntimeclass.yaml
    kubectl create -f $testdir/dataset.yaml
    kubectl create -f $testdir/cacheruntime.yaml

    if [[ -z "$(kubectl get cacheruntimeclass $dataset_name -oname)" ]]; then
        panic "failed to create mooncake cache runtime class $dataset_name"
    fi

    if [[ -z "$(kubectl get dataset $dataset_name -oname)" ]]; then
        panic "failed to create dataset $dataset_name"
    fi

    if [[ -z "$(kubectl get cacheruntime $dataset_name -oname)" ]]; then
        panic "failed to create mooncake cache runtime $dataset_name"
    fi
}

# Regression guard: a client-less topology used to make cacheruntime-controller
# panic on a nil pointer. If it regresses the Dataset gets stuck in NotBound, but
# the panic itself surfaces earlier and says more, so check for it while waiting
# for Bound and report the root cause directly on failure.
function check_controller_not_panicked() {
    local logs=""
    logs=$(kubectl logs -n fluid-system -l control-plane=cacheruntime-controller \
        -c manager --tail=200 2>/dev/null || true)
    if echo "$logs" | grep -qE "panic:|invalid memory address or nil pointer dereference"; then
        syslog "--- cacheruntime-controller panic detected ---"
        echo "$logs" | grep -A 20 -E "panic:|nil pointer dereference" || true
        panic "cacheruntime-controller panicked on a client-less topology (regression)"
    fi
    syslog "No panic found in cacheruntime-controller logs"
}

function wait_dataset_bound() {
    local deadline=600 # 10 minutes
    local last_state=""
    local counter=0
    while true; do
        last_state=$(kubectl get dataset $dataset_name -ojsonpath='{@.status.phase}')
        if [[ "$last_state" == "Bound" ]]; then
            break
        fi

        counter=$((counter + 1))
        if [[ $((counter % 3)) -eq 0 ]]; then
            syslog "checking dataset.status.phase==Bound (already $((counter * 5))s, last state: ${last_state:-<empty>})"
            check_controller_not_panicked
        fi
        if [[ $((counter * 5)) -ge $deadline ]]; then
            panic "timeout ${deadline}s waiting for dataset $dataset_name to be Bound (last state: ${last_state:-<empty>})"
        fi
        sleep 5
    done
    syslog "Found dataset $dataset_name status.phase==Bound"
}

function wait_cache_worker_ready() {
    local deadline=180 # 3 minutes
    local worker_component_name="${dataset_name}-worker"
    local worker_selector="cacheruntime.fluid.io/component-name=${worker_component_name}"
    local last_phase=""
    local ready_replicas=""
    local desired_replicas=""
    local counter=0

    while true; do
        last_phase=$(kubectl get cacheruntime "$dataset_name" -ojsonpath='{@.status.worker.phase}')
        ready_replicas=$(kubectl get cacheruntime "$dataset_name" -ojsonpath='{@.status.worker.readyReplicas}')
        desired_replicas=$(kubectl get cacheruntime "$dataset_name" -ojsonpath='{@.status.worker.desiredReplicas}')

        if [[ "$last_phase" == "Ready" ]] && \
            [[ -n "$desired_replicas" ]] && \
            [[ "$desired_replicas" != "0" ]] && \
            [[ "$ready_replicas" == "$desired_replicas" ]] && \
            kubectl wait --for=condition=Ready --timeout=5s pod -l "$worker_selector" >/dev/null 2>&1; then
            break
        fi

        counter=$((counter + 1))
        if [[ $((counter % 3)) -eq 0 ]]; then
            syslog "checking cache worker readiness (already $((counter * 5))s, phase: ${last_phase:-<empty>}, ready/desired: ${ready_replicas:-<empty>}/${desired_replicas:-<empty>})"
        fi
        if [[ $((counter * 5)) -ge $deadline ]]; then
            panic "timeout ${deadline}s waiting for cache worker pod ready"
        fi
        sleep 5
    done
    syslog "Found ready cache worker pod for $dataset_name"
}

# A client-less topology must produce no client-side artifacts. If the controller
# ever starts creating a fallback DaemonSet for the missing client component,
# this is what catches it first.
function check_no_client_component() {
    local client_component_name="${dataset_name}-client"
    local client_selector="cacheruntime.fluid.io/component-name=${client_component_name}"

    if kubectl get daemonset "$client_component_name" >/dev/null 2>&1; then
        panic "unexpected client DaemonSet $client_component_name created for a client-less topology"
    fi

    local client_pods=""
    client_pods=$(kubectl get pod -l "$client_selector" -oname 2>/dev/null)
    if [[ -n "$client_pods" ]]; then
        panic "unexpected client pods for a client-less topology: $client_pods"
    fi

    # Note: status.client itself does exist (as {"phase":""}) and spec.client is
    # filled in by the CRD defaults; neither means a client component was
    # actually started. The criterion is that phase stays empty.
    local client_phase=""
    client_phase=$(kubectl get cacheruntime "$dataset_name" -ojsonpath='{@.status.client.phase}' 2>/dev/null)
    if [[ -n "$client_phase" ]]; then
        panic "expected empty cacheruntime.status.client.phase for a client-less topology, got: $client_phase"
    fi

    syslog "Confirmed no client component was created"
}

# Evidence that the ReportSummary script ran: cacheStates gets populated.
# Mooncake has no UFS, so its reportSummary.sh fills ufsTotal with the cache
# capacity, which is why the two are expected to be equal.
function check_dataset_cache_state() {
    local deadline=180
    local counter=0
    local cache_capacity=""
    while true; do
        cache_capacity=$(kubectl get dataset ${dataset_name} -ojsonpath='{.status.cacheStates.cacheCapacity}' 2>/dev/null)
        if [[ -n "$cache_capacity" ]] && [[ "$cache_capacity" != "0B" ]]; then
            break
        fi
        counter=$((counter + 1))
        if [[ $((counter * 5)) -ge $deadline ]]; then
            panic "timeout ${deadline}s waiting for cacheStates.cacheCapacity to be reported (got: ${cache_capacity:-<empty>}), report summary may have failed"
        fi
        sleep 5
    done

    local ufs_total=""
    ufs_total=$(kubectl get dataset ${dataset_name} -ojsonpath='{.status.cacheStates.ufsTotal}' 2>/dev/null)
    if [[ "$ufs_total" != "$cache_capacity" ]]; then
        panic "expected ufsTotal to equal cacheCapacity for a UFS-less cache system, got ufsTotal=${ufs_total:-<empty>} cacheCapacity=${cache_capacity}"
    fi

    syslog "Found reported cacheStates (cacheCapacity=$cache_capacity, ufsTotal=$ufs_total)"
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
            syslog "--- logs of failed job $job_name ---"
            kubectl logs job/"$job_name" --tail=100 2>&1 || true
            panic "job $job_name failed when accessing data (all retries exhausted)"
        fi

        counter=$((counter + 1))
        if [[ $((counter * 5)) -ge $deadline ]]; then
            panic "timeout ${deadline}s waiting for job $job_name to complete"
        fi
        sleep 5
    done
    syslog "Found succeeded job $job_name"
    kubectl logs job/"$job_name" --tail=20 2>&1 || true
}

# After a write, cached should reflect the actual usage. ReportSummary runs
# periodically, so one round has to elapse first.
function check_cached_after_write() {
    local deadline=180
    local counter=0
    local cached=""
    local file_num=""
    while true; do
        cached=$(kubectl get dataset ${dataset_name} -ojsonpath='{.status.cacheStates.cached}' 2>/dev/null)
        file_num=$(kubectl get dataset ${dataset_name} -ojsonpath='{.status.cacheStates.fileNum}' 2>/dev/null)
        if [[ -n "$cached" ]] && [[ "$cached" != "0B" ]] && [[ "$file_num" != "0" ]]; then
            break
        fi
        counter=$((counter + 1))
        if [[ $((counter * 5)) -ge $deadline ]]; then
            panic "timeout ${deadline}s waiting for cacheStates to reflect written data (cached=${cached:-<empty>}, fileNum=${file_num:-<empty>})"
        fi
        sleep 5
    done
    syslog "Found cacheStates reflecting written data (cached=$cached, fileNum=$file_num)"
}

# Negative case: without a client component there is no FUSE mount point, so an
# application pod mounting this PVC is bound to fail. The docs' FAQ promises this
# behaviour (the PVC/PV do reach Bound, but cannot be mounted); pin it down here
# so the docs do not quietly go stale if the CSI-side behaviour ever changes.
function check_pvc_not_mountable() {
    # First confirm the PVC/PV really are Bound — this is the misleading part
    local pvc_phase=""
    pvc_phase=$(kubectl get pvc $dataset_name -ojsonpath='{@.status.phase}' 2>/dev/null)
    if [[ "$pvc_phase" != "Bound" ]]; then
        panic "expected PVC $dataset_name to be Bound, got: ${pvc_phase:-<empty>}"
    fi
    syslog "PVC $dataset_name is Bound as documented (but must not be mountable)"

    kubectl create -f $testdir/bad_mount_pod.yaml

    local deadline=150
    local counter=0
    local failed_mount=""
    local pod_phase=""
    while true; do
        failed_mount=$(kubectl get events \
            --field-selector "involvedObject.name=${bad_mount_pod_name},reason=FailedMount" \
            -ojsonpath='{.items[*].message}' 2>/dev/null)
        if [[ -n "$failed_mount" ]]; then
            break
        fi

        # If the pod ever becomes Running the mount actually succeeded, which is
        # a behaviour change
        pod_phase=$(kubectl get pod "$bad_mount_pod_name" -ojsonpath='{@.status.phase}' 2>/dev/null)
        if [[ "$pod_phase" == "Running" ]]; then
            panic "pod $bad_mount_pod_name mounted the PVC successfully; a client-less runtime is expected to have no FUSE mount point"
        fi

        counter=$((counter + 1))
        if [[ $((counter % 6)) -eq 0 ]]; then
            syslog "waiting for FailedMount event (already $((counter * 5))s, pod phase: ${pod_phase:-<empty>})"
        fi
        if [[ $((counter * 5)) -ge $deadline ]]; then
            syslog "--- describe of $bad_mount_pod_name ---"
            kubectl describe pod "$bad_mount_pod_name" 2>&1 || true
            panic "timeout ${deadline}s waiting for a FailedMount event on $bad_mount_pod_name"
        fi
        sleep 5
    done

    syslog "Found expected FailedMount event: $failed_mount"

    # The failure reason should point at the missing FUSE mount point, not some
    # other mount problem
    if ! echo "$failed_mount" | grep -qi "fuse mount point"; then
        syslog "WARNING: FailedMount message does not mention the FUSE mount point; the docs' FAQ wording may need updating"
    fi

    kubectl delete --ignore-not-found -f $testdir/bad_mount_pod.yaml --force --grace-period=0 >/dev/null 2>&1
    syslog "Confirmed the Dataset PVC cannot be mounted by application pods"
}

function delete_dataset_and_runtime() {
    kubectl delete -f $testdir/dataset.yaml
    kubectl delete -f $testdir/cacheruntime.yaml
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
    syslog "All runtime resources (AdvancedStatefulSet, Service) garbage collected"

    if kubectl get pvc $dataset_name -n default >/dev/null 2>&1; then
        panic "PVC $dataset_name still exists after deletion"
    fi
    if kubectl get pv default-$dataset_name >/dev/null 2>&1; then
        panic "PV default-$dataset_name still exists after deletion"
    fi
    syslog "PV/PVC cleaned up successfully"
}

function dump_env_and_clean_up() {
    local exit_code=$?
    if [[ $exit_code -ne 0 ]]; then
        syslog "=== Diagnostic logs for failed test ==="
        syslog "--- cacheruntime-controller logs (last 100 lines) ---"
        kubectl logs -n fluid-system -l control-plane=cacheruntime-controller -c manager --tail=100 2>&1 || true
        syslog "--- CacheRuntime describe ---"
        kubectl describe cacheruntime $dataset_name 2>&1 || true
        syslog "--- Dataset describe ---"
        kubectl describe dataset $dataset_name 2>&1 || true
        syslog "--- Pods in default namespace ---"
        kubectl get pods -n default -owide 2>&1 || true
        syslog "--- bad-mount pod describe ---"
        kubectl describe pod $bad_mount_pod_name 2>&1 || true
        syslog "--- Job logs ---"
        kubectl logs job/$rw_job_name --tail=100 2>&1 || true
        syslog "--- Events in default namespace ---"
        kubectl get events -n default --sort-by='.lastTimestamp' 2>&1 || true
        syslog "=== End of diagnostic logs ==="
    fi
    syslog "Cleaning up resources for testcase $testname"
    kubectl delete --ignore-not-found -f $testdir/bad_mount_pod.yaml --force --grace-period=0
    kubectl delete --ignore-not-found -f $testdir/rw_job.yaml
    kubectl delete --ignore-not-found -f $testdir/dataset.yaml
    kubectl delete --ignore-not-found -f $testdir/cacheruntime.yaml
    kubectl delete --ignore-not-found -f $testdir/cacheruntimeclass.yaml
}

function main() {
    syslog "[TESTCASE $testname STARTS AT $(date)]"
    trap dump_env_and_clean_up EXIT

    create_dataset
    wait_dataset_bound
    check_controller_not_panicked
    wait_cache_worker_ready
    check_no_client_component
    check_dataset_cache_state

    create_job $testdir/rw_job.yaml $rw_job_name
    wait_job_completed $rw_job_name
    check_cached_after_write

    check_pvc_not_mountable

    delete_dataset_and_runtime
    wait_runtime_deleted

    syslog "[TESTCASE $testname SUCCEEDED AT $(date)]"
}

main
