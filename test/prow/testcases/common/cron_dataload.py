"""
TestCase: Cron Dataload Verification
DDC Engine: Alluxio
Steps:
1. create Dataset & Runtime
2. check if dataset is bound
3. check if PVC & PV is created
4. submit DataLoad CR
5. check dataload status
6. wait until DataLoad completes
7. check dataset cache percentage
8. create app pod
9. check app pod is running
10. clean up
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

def check_untriggered_dataload(dataload_name, start_time, dataload_namespace="default"):
    localtime = time.localtime(time.time())
    # if current time is later than cron start time (cur time >= start time, 
    # expect start time is the beginning of next hour, but cur time is the end of cur hour), 
    # dataload is already triggered, pass check untriggered dataload
    if localtime.tm_min >= start_time and not (start_time < 5 and localtime.tm_min > 55):
        return True
    
    api = client.CustomObjectsApi()
    resource = api.get_namespaced_custom_object(
        group="data.fluid.io",
        version="v1alpha1",
        name=dataload_name,
        namespace=dataload_namespace,
        plural="dataloads"
    )

    if "status" in resource:
        if "phase" in resource["status"]:
            if resource["status"]["phase"] != "Executing":
                return False

    pods = client.CoreV1Api().list_namespaced_pod(dataload_namespace)
    for pod in pods.items:
        if dataload_name in pod.metadata.name:
            return False

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
    
    dataload_name = "jindo-dataload"
    localtime = time.localtime(time.time())
    start_time = (localtime.tm_min + 5) % 60
    # cron dataload will start executing after 5 minutes
    cron_dataload = fluidapi.DataLoad(name=dataload_name, namespace=namespace) \
        .set_target_dataset(name, namespace) \
        .set_load_metadata(True) \
        .set_cron("{} * * * *".format(start_time))
    
    flow = TestFlow("Common - Test Cron Dataload")

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
            forth_fn=funcs.create_dataload_fn(cron_dataload.dump()),
            back_fn=dummy_back,  # DataLoad should have ownerReference of Dataset
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="check untriggered dataload: no job pod is created",
            forth_fn=currying_fn(check_untriggered_dataload, 
                                 dataload_name=dataload_name, start_time=start_time),
            back_fn=dummy_back,
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check if dataload job completes",
            forth_fn=funcs.check_dataload_job_status_fn(dataload_name, namespace),
            timeout=600
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check if the whole dataset is warmed up",
            forth_fn=funcs.check_dataset_cached_percentage_fn(name, namespace)
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="create data read job",
            forth_fn=funcs.create_pod_fn(dataset_name=name, name="nginx-test"), 
            back_fn=funcs.delete_pod_fn(name="nginx-test")
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check_pod_running",
            forth_fn=funcs.check_pod_running_fn(name="nginx-test")
        )
    )

    try:
        flow.run()
    except Exception as e:
        print(e)
        exit(1)


if __name__ == '__main__':
    main()