#!/bin/bash

testname="alluxioruntime graceful scale-down e2e"

dataset_name="scaledown-demo"
worker_sts_name="scaledown-demo-worker"
controller_deployment="alluxioruntime-controller"
controller_namespace="fluid-system"
read_before_job_name="read-before-scaledown"
read_after_job_name="read-after-scaledown"
bucket_create_job_name="scaledown-minio-bucket-create"

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
    local existing_args
    existing_args=$(kubectl get deployment "$controller_deployment" -n "$controller_namespace" \
        -ojson | jq -c '.spec.template.spec.containers[0].args // []')

    if echo "$existing_args" | grep -q "GracefulWorkerScaleDown=true"; then
        syslog "GracefulWorkerScaleDown feature gate already enabled"
        return
    fi

    # --feature-gates is a last-flag-wins string flag, so if one is already
    # present (e.g. via Helm values), merge into it instead of appending a
    # second occurrence that would silently disable whatever it already set.
    local new_args
    if echo "$existing_args" | grep -q -- '--feature-gates='; then
        new_args=$(echo "$existing_args" | jq -c 'map(
            if startswith("--feature-gates=") then . + ",GracefulWorkerScaleDown=true" else . end)')
    else
        new_args=$(echo "$existing_args" | jq -c '. + ["--feature-gates=GracefulWorkerScaleDown=true"]')
    fi

    kubectl patch deployment "$controller_deployment" -n "$controller_namespace" --type=json \
        -p "[{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/0/args\",\"value\":$new_args}]" \
        || panic "failed to patch $controller_deployment to enable GracefulWorkerScaleDown"

    kubectl rollout status deployment/"$controller_deployment" -n "$controller_namespace" --timeout=120s \
        || panic "alluxioruntime-controller did not roll out after enabling the feature gate"

    syslog "Enabled GracefulWorkerScaleDown feature gate on $controller_deployment"
}

function setup_minio() {
    kubectl create -f test/gha-e2e/alluxio-scaledown/minio.yaml
    # minio has no readiness probe, and the bucket-create job has a limited
    # backoffLimit; without waiting here, the job can exhaust all its retries
    # before minio finishes scheduling/starting on a slower runner.
    kubectl rollout status deployment/scaledown-minio --timeout=120s \
        || panic "scaledown-minio deployment did not become ready"
    kubectl create -f test/gha-e2e/alluxio-scaledown/minio_create_bucket.yaml
    wait_job_completed "$bucket_create_job_name"
}

# Tears minio down entirely so the post-scale-down read can only be served by
# data still cached on an Alluxio worker, never by a transparent UFS
# re-fetch: fixture.txt lives in the S3 UFS the whole test long, so with minio
# still up, read_after_job's "cat" would succeed identically whether or not
# graceful decommission actually preserved any cached block - it proves
# nothing about the drain. Called between the two read jobs so read_before
# (which primes the cache via Alluxio's default CACHE_PROMOTE read type) can
# still reach minio, but read_after cannot.
function teardown_minio() {
    kubectl delete --ignore-not-found -f test/gha-e2e/alluxio-scaledown/minio.yaml
    kubectl wait --for=delete pod -l app=scaledown-minio --timeout=60s 2>/dev/null || true
    syslog "Tore down scaledown-minio; the post-scale-down read can now only be served by data still cached in Alluxio workers"
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
    local decommission_condition=""
    local counter=0
    while true; do
        spec_replicas=$(kubectl get statefulset "$worker_sts_name" -ojsonpath='{@.spec.replicas}' 2>/dev/null)
        status_replicas=$(kubectl get statefulset "$worker_sts_name" -ojsonpath='{@.status.replicas}' 2>/dev/null)

        if [[ "$spec_replicas" == "$expected" ]] && [[ "$status_replicas" == "$expected" ]]; then
            break
        fi

        if [[ $((counter % 6)) -eq 0 ]]; then
            decommission_condition=$(kubectl get alluxioruntime "$dataset_name" \
                -ojsonpath='{.status.conditions[?(@.type=="WorkerDecommissioning")]}' 2>/dev/null)
            syslog "waiting for $worker_sts_name to reach $expected replicas (already $((counter * 5))s, spec=$spec_replicas status=$status_replicas, decommissionCondition=$decommission_condition)"
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
    kubectl patch alluxioruntime "$dataset_name" --type=merge -p '{"spec":{"replicas":1}}' \
        || panic "failed to patch alluxioruntime $dataset_name to replicas=1"
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
            syslog "dumping diagnostics for failed job $job_name"
            kubectl get pods -l job-name="$job_name" -o wide 2>&1 || true
            # Address pods by name (not job/$job_name) and bound every call
            # with --request-timeout: resolving logs via the job/deployment
            # helper can itself hang waiting for a pod to reach Running,
            # which is exactly the state we're trying to debug.
            for pod in $(kubectl get pods -l job-name="$job_name" -o jsonpath='{.items[*].metadata.name}' 2>/dev/null); do
                syslog "describing pod $pod"
                kubectl describe pod "$pod" --request-timeout=10s 2>&1 || true
                syslog "logs for pod $pod"
                if ! kubectl logs "$pod" --all-containers --prefix --request-timeout=10s 2>&1; then
                    # By the time backoffLimit is exceeded, the job
                    # controller may already be terminating the pod, racing
                    # our own attempt to read its logs. --previous targets
                    # the last *fully exited* container instance instead of
                    # the one mid-teardown, and often still works when the
                    # plain form above doesn't.
                    syslog "current logs unavailable for pod $pod, retrying with --previous"
                    kubectl logs "$pod" --all-containers --prefix --previous --request-timeout=10s 2>&1 || true
                fi
            done
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
    kubectl delete --ignore-not-found -f test/gha-e2e/alluxio-scaledown/minio_create_bucket.yaml
    kubectl delete --ignore-not-found -f test/gha-e2e/alluxio-scaledown/minio.yaml
}

# Note: the GHA Kind cluster is single-node, so both worker pods land on the
# same node and report the same HostIP. drainScalingDownWorkers in
# pkg/ddc/alluxio/replicas.go dedupes decommission addresses by HostIP:port,
# so this test cannot prove a *specific* worker was targeted by address - only
# that the decommission flow runs, the StatefulSet converges, and data isn't
# lost. Per-address targeting on distinct hosts is covered by the gomonkey
# unit tests in pkg/ddc/alluxio/replicas_drain_test.go instead.
function main() {
    syslog "[TESTCASE $testname STARTS AT $(date)]"
    trap dump_env_and_clean_up EXIT
    enable_graceful_scale_down
    setup_minio
    create_dataset
    wait_dataset_bound
    wait_worker_replicas 2
    create_job test/gha-e2e/alluxio-scaledown/read_before_job.yaml $read_before_job_name
    wait_job_completed $read_before_job_name
    teardown_minio
    scale_down
    wait_worker_replicas 1
    create_job test/gha-e2e/alluxio-scaledown/read_after_job.yaml $read_after_job_name
    wait_job_completed $read_after_job_name
    syslog "[TESTCASE $testname SUCCEEDED AT $(date)]"
}

main
