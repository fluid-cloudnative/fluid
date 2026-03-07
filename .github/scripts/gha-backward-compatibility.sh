#!/bin/bash

syslog() {
    echo ">>> ${1}"
    return 0
}

panic() {
    local err_msg="${1}"
    syslog "backward compatibility test failed: ${err_msg}"
    exit 1
}

check_control_plane_status() {
    echo "=== Unique image tags used by Fluid control plane ==="
    kubectl get pod -n fluid-system -o jsonpath='
      {range .items[*]}{range .spec.containers[*]}{.image}{"\n"}{end}{range .spec.initContainers[*]}{.image}{"\n"}{end}{end}' \
      | sed 's/.*://' \
      | sort -u

    # Timeout counter (30 minutes = 360*5 seconds)
    local timeout=360
    local counter=0
    local status_interval=36
    
    while true; do
        total_pods=$(kubectl get pod -n fluid-system --no-headers | grep -cv "Completed")
        running_pods=$(kubectl get pod -n fluid-system --no-headers | grep -c "Running")
        not_running_pods=$((total_pods - running_pods))

        if ((counter % status_interval == 0)); then
            syslog "[Status Check $((counter / status_interval))] Pod status: ${running_pods}/${total_pods} running (${not_running_pods} not ready)"
            if [[ "${not_running_pods}" -gt 0 ]]; then
                echo "=== Not running pods ==="
                kubectl get pods -n fluid-system \
                    --field-selector=status.phase!=Running \
                    -o=custom-columns='NAME:.metadata.name,STATUS:.status.phase,REASON:.status.reason'
            fi
        fi

        if [[ "${total_pods}" -ne 0 ]] && [[ "${total_pods}" -eq "${running_pods}" ]]; then
            break
        fi
        
        if [[ "${counter}" -ge "${timeout}" ]]; then
            panic "Timeout waiting for control plane after ${counter} checks!"
        fi
        
        sleep 5
        ((counter++))
    done
    syslog "Fluid control plane is ready after ${counter} checks!"
}

wait_dataset_bound() {
    local dataset_name="${1}"
    local deadline=180
    local log_interval=0
    local log_times=0
    
    syslog "Waiting for dataset ${dataset_name} to be Bound..."
    
    while true; do
        # We don't use 'set -e' here so we can handle the case where the object or field is missing
        last_state=$(kubectl get dataset "${dataset_name}" -n default -ojsonpath='{.status.phase}' 2>/dev/null || echo "Unknown")
        
        if [[ "${last_state}" == "Bound" ]]; then
            break
        fi

        if [[ "${log_interval}" -eq 3 ]]; then
            ((log_times++))
            syslog "checking dataset.status.phase==Bound (elapsed: $((log_times * 3 * 5))s, current state: ${last_state})"
            if [[ $((log_times * 3 * 5)) -ge "${deadline}" ]]; then
                panic "timeout for ${deadline}s waiting for dataset ${dataset_name} to become bound!"
            fi
            log_interval=0
        fi

        ((log_interval++))
        sleep 5
    done
    syslog "Found dataset ${dataset_name} status.phase==Bound"
}

wait_job_completed() {
    local job_name="${1}"
    local deadline=600 # 10 minutes
    local counter=0
    while true; do
        # Handle missing fields gracefully
        succeed=$(kubectl get job "${job_name}" -ojsonpath='{.status.succeeded}' 2>/dev/null || echo "0")
        failed=$(kubectl get job "${job_name}" -ojsonpath='{.status.failed}' 2>/dev/null || echo "0")
        
        # Ensure variables are treated as integers
        [[ -z "${succeed}" ]] && succeed=0
        [[ -z "${failed}" ]] && failed=0

        if [[ "${failed}" -gt 0 ]]; then
            panic "job ${job_name} failed when accessing data"
        fi
        if [[ "${succeed}" -gt 0 ]]; then
            break
        fi
        
        ((counter++))
        if [[ $((counter * 5)) -ge "${deadline}" ]]; then
            panic "timeout for ${deadline}s waiting for job ${job_name} completion!"
        fi
        sleep 5
    done
    syslog "Found succeeded job ${job_name}"
}

setup_old_fluid() {
    syslog "Setting up older version of Fluid from charts"
    helm repo add fluid https://fluid-cloudnative.github.io/charts
    helm repo update fluid
    
    # We ignore errors in case namespace exists
    kubectl create ns fluid-system || true
    
    helm install fluid fluid/fluid --namespace fluid-system --wait
    check_control_plane_status
}

create_dataset() {
    syslog "Creating alluxio dataset..."
    kubectl apply -f test/gha-e2e/alluxio/dataset.yaml
    # give it 15s to let the CRDs and controllers settle
    sleep 15
    wait_dataset_bound "zookeeper"
}

upgrade_fluid() {
    syslog "Upgrading Fluid to the locally built current version..."
    ./.github/scripts/deploy-fluid-to-kind.sh
    check_control_plane_status
}

verify_backward_compatibility() {
    syslog "Verifying backward compatibility..."
    # Ensure the dataset created earlier is still bound
    wait_dataset_bound "zookeeper"
    
    # create job to access data over the runtime
    kubectl apply -f test/gha-e2e/alluxio/job.yaml
    wait_job_completed "fluid-test"
    
    # Clean up
    kubectl delete -f test/gha-e2e/alluxio/
}

main() {
    syslog "[BACKWARD COMPATIBILITY TEST STARTS AT $(date)]"
    
    setup_old_fluid
    create_dataset
    upgrade_fluid
    verify_backward_compatibility
    
    syslog "[BACKWARD COMPATIBILITY TEST SUCCEEDED AT $(date)]"
}

main
