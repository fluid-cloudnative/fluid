#!/bin/bash
set -x

dataset_path="./dataset.yaml"
runtime_path="./runtime.yaml"

dataset_name="spark"

create_dataset()
{
    echo "create dataset..."
    kubectl create -f $dataset_path
    local result=$(kubectl get dataset | awk '{print $1}' | grep ^spark$)
    if [[ $result == $dataset_name ]]
    then
        echo "create dataset $dataset_name successfully!"
    else
        echo "ERROR: can not create dataset ${dataset_name}."
        exit 1
    fi
}

create_runtime()
{
    echo "create runtime..."
    kubectl create -f $runtime_path
    local result=$(kubectl get alluxioruntime | awk '{print $1}' | grep ^spark$)
    if [[ $result == $dataset_name ]]
    then
        echo "create runtime $dataset_name successfully!"
    else
        echo "ERROR: can not create runtime ${dataset_name}."
        exit 1
    fi
}

check_runtime_pod()
{
    echo "check runtime pods..."
    while :
    do
        local master_num=$(kubectl get pod | grep spark-master | awk '$3=="Running"' | wc -l)
        local worker_num=$(kubectl get pod | grep spark-worker | awk '$3=="Running"' | wc -l)
        local fuse_num=$(kubectl get pod | grep spark-fuse | awk '$3=="Running"' | wc -l)

        if [[ $master_num -gt 0 && $worker_num -gt 0 && $fuse_num -gt 0 ]]
        then
            echo "runtime pods are ready."
            break;
        else
            echo "runtime pods are not ready, wait 10 seconds..."
            sleep 10
        fi
    done
            
}

check_pvc()
{
    echo "check pv and pvc..."
    while :
    do
        local pv_status=$(kubectl get pv | awk '$1=="spark" && $7=="fluid" {print $5}')
        if [[ $pv_status == "Bound" ]]
        then
            echo "pv $spark_name has been created and bound."
            break
        else
            echo "pv is not created or bound, wait 5 seconds..."
        fi
    done
    
    while :
    do
        local pvc_status=$(kubectl get pvc | awk '$1=="spark" && $3=="spark" && $6=="fluid" {print $2}')
        if [[ $pvc_status == "Bound" ]]
        then
            echo "pvc $spark_name has been created and bound."
            break
        else
            echo "pvc is not created or bound, wait 5 seconds..."
        fi
    done
    
}

check_dataset_bound()
{
    echo "check whether dataset is bound..."
    while :
    do
        local master_status=$(kubectl get alluxioruntime | awk '$1=="spark"{print $2}')
        local worker_status=$(kubectl get alluxioruntime | awk '$1=="spark"{print $3}')
        local fuse_status=$(kubectl get alluxioruntime | awk '$1=="spark"{print $4}')

        if [[ $master_status == "Ready" && ($worker_status == "Ready" || $worker_status == "PartialReady") && ($fuse_status == "Ready" || $fuse_status == "PartialReady") ]]
        then
            echo "runtime is ready."
        else
            echo "runtime is not ready, wait 5 seconds..."
            continue
        fi

        local dataset_status=$(kubectl get dataset | awk '$1=="spark"{print $6}')
        if [[ $dataset_status == "Bound" ]]
        then
            echo "dataset is bound."
            break
        else
            echo "dataset is not bound, wait 5 seconds..."
            sleep 5
        fi
    done
    
}

delete_dataset()
{
    echo "delete dataset..."
    while :
    do
        kubectl delete dataset $dataset_name
        local dataset_status=$(kubectl get dataset | awk '$1=="spark"')
        if [[ $dataset_status == "" ]]
        then
            echo "delete dataset $dataset_name successfully!"
            break
        else
            echo "dataset ${dataset_name} has not deleted, wait 5 seconds."
    fi
    done
    
    while :
    do
        local runtime_status=$(kubectl get alluxioruntime | awk '$1=="spark"')
        if [[ $runtime_status == "" ]]
        then
            echo "delete runtime $dataset_name successfully!"
            break
        else
            echo "runtime ${dataset_name} has not deleted, wait 5 seconds."
    fi
    done

}

main()
{
    echo "begin to test..."
    create_dataset && \
    create_runtime && \
    check_runtime_pod && \
    check_pvc && \
    check_dataset_bound && \
    delete_dataset
    echo "pass the test."
}

main "$@"