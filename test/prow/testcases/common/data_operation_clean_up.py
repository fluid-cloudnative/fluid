"""
TestCase: CleanUp Data Operation Verification
DDC Engine: Alluxio
Steps:
1. create Dataset & Runtime
2. check if dataset is bound
3. check if PVC & PV is created
4. submit DataLoad CR
5. check dataload status
6. wait until DataLoad completes
7. wait Dataload to clean up
8. clean up
"""

import os
import sys
import time

project_root = os.path.dirname(os.path.dirname(os.path.dirname(__file__)))
sys.path.insert(0, project_root)

from kubernetes import client, config

import fluid.fluidapi as fluidapi
import fluid.step_funcs as funcs
from framework.testflow import TestFlow
from framework.step import SimpleStep, StatusCheckStep, dummy_back, currying_fn


def wait_clean_up(dataload_name, namespace, ttl):
    time.sleep(ttl)
    api = client.CustomObjectsApi()
    try:
        dataload = api.get_namespaced_custom_object(
            group="data.fluid.io",
            version="v1alpha1",
            name=dataload_name,
            namespace=namespace,
            plural="dataloads"
        )
    except Exception as e :
        print(e)
        if e.status == 404:
            return True
    
    return False


def check_cron_dataload(dataload_name, dataset_name, namespace):
    api = client.CustomObjectsApi()
    for i in range(0, 60):
        dataload = api.get_namespaced_custom_object(
            group="data.fluid.io",
            version="v1alpha1",
            name=dataload_name,
            namespace=namespace,
            plural="dataloads"
        )
        dataset = api.get_namespaced_custom_object(
            group="data.fluid.io",
            version="v1alpha1",
            name=dataset_name,
            namespace=namespace,
            plural="datasets"
        )
        
        dataload_status = dataload["status"]["phase"]
        opRef = dataset["status"].get("operationRef", {})
        if dataload_status == "Failed":
            print("Dataload {} is failed".format(dataload_name))
            return False
        elif dataload_status == "Complete":
            if opRef is not None and opRef.get("DataLoad", "") != "":
                print("DataLoad {} is complete but dataset opRef {} is not None".format(dataload_name, opRef))
                return False
        elif dataload_status == "Executing":
            if opRef is None:
                print("Dataload {} is running but dataset opRef None".format(dataload_name))
                return False
            if opRef.get("DataLoad", "") != dataload_name:
                print("DataLoad {} is running but dataset opRef {}".format(dataload_name, opRef))
                return False
        time.sleep(1)

    return True

def main():
    if os.getenv("KUBERNETES_SERVICE_HOST") is None:
        config.load_kube_config()
    else:
        config.load_incluster_config()

    name = "hbase"
    namespace = "default"

    dataset = fluidapi.assemble_dataset("alluxio-webufs") \
        .set_namespaced_name(namespace, name) \
        
    runtime = fluidapi.assemble_runtime("alluxio-webufs") \
        .set_namespaced_name(namespace, name) \
        .set_tieredstore(mediumtype="MEM", path="/dev/shm", quota="2Gi")
    
    dataload_name = "test-dataload"
    datalaod = fluidapi.DataLoad(name=dataload_name, namespace=namespace) \
        .set_target_dataset(name, namespace) \
        .set_load_metadata(True) \
        .set_ttlSecondsAfterFinished(20)
    
    flow = TestFlow("Common - Test Clean Up Dataoperation")

    flow.append_step(
        SimpleStep(
            step_name="create dataset",
            forth_fn=funcs.create_dataset_fn(dataset.dump()),
            back_fn=funcs.delete_dataset_and_runtime_fn(runtime.dump(), name, namespace)
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="create runtime",
            forth_fn=funcs.create_runtime_fn(runtime.dump()),
            back_fn=dummy_back
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check dataset bound",
            forth_fn=funcs.check_dataset_bound_fn(name, namespace)
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check if PV & PVC is ready",
            forth_fn=funcs.check_volume_resource_ready_fn(name, namespace)
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="create dataload",
            forth_fn=funcs.create_dataload_fn(datalaod.dump()),
            back_fn=dummy_back,  # DataLoad should have ownerReference of Dataset
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check if dataload job completes",
            forth_fn=funcs.check_dataload_job_status_fn(dataload_name, namespace),
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="wait dataload to clean up",
            forth_fn=currying_fn(wait_clean_up, dataload_name=dataload_name, namespace=namespace, ttl=20)
        )
    )

    try:
        flow.run()
    except Exception as e:
        print(e)
        exit(1)
    
    return 0

if __name__ == '__main__':
    main()
