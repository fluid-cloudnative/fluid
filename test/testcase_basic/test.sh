#!/bin/bash
set -x

dataset_path="./dataset.yaml"
runtime_path="./runtime.yaml"
dataload_path="./dataload.yaml"
fluid_git="https://github.com/fluid-cloudnative/fluid.git"

dataset_name="spark"
dataload_name="spark-dataload"

get_fluid()
{
    echo "get fluid lastest chart..."
    if [ -d "/fluid" ]
    then
        echo "fluid repository already exists."
    else
        echo "clone from ${fluid_git}."
        git clone $fluid_git /fluid
    fi

    echo "update fluid from master branch..."
    cd /fluid && 
    git checkout master && 
    git pull origin master:master
    if [[ $? -ne 0 ]]
    then
        echo "ERROR: failed to update fluid"
        exit 1
    else
        echo "fluid updated."
    fi

    cd -
}

uninstall_fluid()
{
    local fluid=$(helm list | awk '{print $1}' | grep ^fluid$)
    if [[ $fluid == "fluid" ]]
    then
        echo "delete crd..."
        kubectl delete crd $(kubectl get crd | grep data.fluid.io | awk '{print $1}')
        local crd=$(kubectl get crd | grep data.fluid.io)
        if [[ $crd == "" ]]
        then
            echo "delete fluid crd successfully."
        else
            echo "ERROR: can not delete fluid crd."
            exit 1
        fi
    fi

    echo "uninstall fluid..."
    helm delete fluid
    fluid=$(helm list | awk '{print $1}' | grep ^fluid$)
    if [[ $fluid == "" ]]
    then
        echo "uninstall fluid successfully."
    else
        echo "ERROR: can not uninstall fluid."
        exit 1
    fi
}

install_fluid()
{
    echo "create namespace..."
    local namespace=$(kubectl get namespace | awk '{print $1}' | grep ^fluid-system$)
    if [[ $namespace == "" ]]
    then
        kubectl create namespace fluid-system
    else
        echo "namespace $namespace already exists."
    fi

    echo "install fluid..."
    helm install fluid /fluid/charts/fluid/fluid/

    local fluid=$(helm list | awk '{print $1}' | grep ^fluid$)
    if [[ $fluid == "fluid" ]]
    then
        echo "fluid has been installed successfully. check its running status..."
        while :
        do
            local alluxioruntime_controller_status=$(kubectl get pod -n fluid-system | grep alluxioruntime-controller | awk '{print $3}')
            local dataset_controller_status=$(kubectl get pod -n fluid-system | grep dataset-controller | awk '{print $3}')
            local node_num=$(expr $(kubectl get nodes | wc -l) - 1)
            local csi_nodeplugin_num=$(kubectl get pod -n fluid-system | grep csi-nodeplugin | awk '$3=="Running"' | wc -l)

            if [[ $alluxioruntime_controller_status == "Running" && $dataset_controller_status == "Running" && $csi_nodeplugin_num -eq $node_num ]]
            then
                echo "fluid runs successfully."
                break
            else
                echo "fluid does not run, wait 10 seconds..."
                sleep 10
            fi
        done
    else
        echo "ERROR: can not install fluid."
        exit 1
    fi
}

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

create_dataload()
{
    echo "create dataload..."
    kubectl create -f $dataload_path
    local result=$(kubectl get dataload | awk '{print $1}' | grep ^spark-dataload$)
    if [[ $result == $dataload_name ]]
    then
        echo "create dataload $dataload_name successfully!"
        sleep 5
    else
        echo "ERROR: can not create dataload ${dataload_name}."
        exit 1
    fi
}

check_dataload()
{
    echo "check dataload running status..."
    local job=$(kubectl get job | awk '$1=="spark-dataload-loader-job"')
    if [[ $job == "" ]]
    then
        echo "ERROR: the dataload job is not created successfully."
        exit 1
    else
        echo "the dataload job is created successfully."
    fi

    local dataload_status=$(kubectl get dataload | awk '$1=="spark-dataload" {print $3}')
    if [[ $dataload_status == "Pending" || $dataload_status == "Loading" || $dataload_status == "Complete" || $dataload_status == "Failed" ]]
    then
        echo "dataload is running properly."
    else
        echo "ERROR: dataload is not running properly"
        exit 1
    fi

    echo "check if dataload is finished..."
    while :
    do
        dataload_status=$(kubectl get dataload | awk '$1=="spark-dataload" {print $3}')
        if [[ $dataload_status == "Complete" || $dataload_status == "Failed" ]]
        then
            echo "dataload is finished."
            if [[ $dataload_status == "Complete" ]]
            then
                local cache_percent=$(kubectl get dataset | awk '$1=="spark" {print $5}')
                echo "data is loaded successfully, the cache percent is ${cache_percent}."
            else
                echo "failed to load data."
            fi
            break
        else
            echo "dataload is still running, wait 20 seconds..."
            sleep 20
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
            sleep 5
    fi
    done

    while :
    do
        local dataload_status=$(kubectl get dataload | awk '$1=="spark-dataload"')
        if [[ $dataload_status == "" ]]
        then
            echo "delete dataload $dataload_name successfully!"
            break
        else
            echo "dataload ${dataload_name} has not deleted, wait 5 seconds."
            sleep 5
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
            echo "runtime ${dataset_name} has not deleted, wait 10 seconds."
            sleep 10
    fi
    done

}

main()
{
    echo "begin to test..."
    get_fluid && \
    uninstall_fluid && \
    install_fluid
    create_dataset && \
    create_runtime && \
    check_runtime_pod && \
    check_pvc && \
    check_dataset_bound && \
    create_dataload && \
    check_dataload && \
    delete_dataset
    echo "pass the test."
}

main "$@"