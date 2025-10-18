#!/bin/bash

function syslog() {
    echo ">>> $1"
}

function check_control_plane_status() {
    echo "=== Unique image tags used by Fluid control plane ==="
    kubectl get pod -n fluid-system -o jsonpath='
      {range .items[*]}{range .spec.containers[*]}{.image}{"\n"}{end}{range .spec.initContainers[*]}{.image}{"\n"}{end}{end}' \
      | sed 's/.*://' \
      | sort -u

    # Timeout counter (30 minutes = 360*5 seconds)
    local timeout=360
    local counter=0
    # Status check interval (36 iterations * 5s = 180s = 3 minutes)
    local status_interval=36
    
    while true; do
        total_pods=$(kubectl get pod -n fluid-system --no-headers | grep -cv "Completed")
        running_pods=$(kubectl get pod -n fluid-system --no-headers | grep -c "Running")
        not_running_pods=$(($total_pods - $running_pods))

        # Print status every 3 minutes
        if ((counter % status_interval == 0)); then
            syslog "[Status Check $((counter/status_interval))] Pod status: $running_pods/$total_pods running ($not_running_pods not ready)"
            
            # Get details for non-running pods
            if [[ $not_running_pods -gt 0 ]]; then
                echo "=== Not running pods ==="
                kubectl get pods -n fluid-system \
                    --field-selector=status.phase!=Running \
                    -o=custom-columns='NAME:.metadata.name,STATUS:.status.phase,REASON:.status.reason'
                
                # Get events for problem pods
                local problem_pods=$(kubectl get pods -n fluid-system \
                    --field-selector=status.phase!=Running \
                    -o=jsonpath='{.items[*].metadata.name}')
                
                for pod in $problem_pods; do
                    echo "--- Events for $pod ---"
                    # Extract events section from pod description
                    kubectl describe pod -n fluid-system $pod | awk '/Events:/,/^ *$/{if($0!~/^ *$/&&$0!~/Events:/)print}'
                done
            fi
        fi

        # Exit loop when all pods are running
        if [[ $total_pods -ne 0 ]] && [[ $total_pods -eq $running_pods ]]; then
            break
        fi
        
        # Handle timeout after 30 minutes
        if ((counter >= timeout)); then
            syslog "Timeout waiting for control plane after $counter checks!"
            
            # Final pod status
            echo "=== Final pod status ==="
            kubectl get pods -n fluid-system -o wide
            
            # Container logs (last 100 lines)
            local all_pods=$(kubectl get pods -n fluid-system -o jsonpath='{.items[*].metadata.name}')
            for pod in $all_pods; do
                echo "--- Logs for $pod (last 100 lines) ---"
                kubectl logs -n fluid-system $pod --all-containers --tail=100
            done
            
            # Additional diagnostics
            echo "=== Node resource usage ==="
            kubectl top nodes
            
            echo "=== Persistent volume claims ==="
            kubectl get pvc -n fluid-system
            
            echo "=== Fluid system events ==="
            kubectl get events -n fluid-system --sort-by=.metadata.creationTimestamp
            
            exit 1
        fi
        
        sleep 5
        ((counter++))
    done
    syslog "Fluid control plane is ready after $counter checks!"
}

function alluxio_e2e() {
    set -e
    bash test/gha-e2e/alluxio/test.sh
}

function jindo_e2e() {
    set -e
    bash test/gha-e2e/jindo/test.sh
}

function juicefs_e2e() {
    set -e
    bash test/gha-e2e/juicefs/test.sh
}

check_control_plane_status
alluxio_e2e
jindo_e2e
juicefs_e2e
