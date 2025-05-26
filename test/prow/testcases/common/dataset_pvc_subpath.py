"""
TestCase: Dataset with PVC Subpath Verification
DDC Engine: Alluxio
Steps:
1. create PV and PVC
2. create Dataset(PVC subpath) & Runtime
3. check if dataset is bound
4. check if persistentVolumeClaim & PV is created
5. check if app pod is running and if data read success
6. clean up
"""

import os
import shutil
import sys

project_root = os.path.dirname(os.path.dirname(os.path.dirname(__file__)))
sys.path.insert(0, project_root)

import fluid.fluidapi as fluidapi
import fluid.step_funcs as funcs
from framework.step import SimpleStep, StatusCheckStep, currying_fn, dummy_back
from framework.testflow import TestFlow
from kubernetes import client, config
from kubernetes.stream import stream


def checkPVCSubpathDatasetAccess(pod_name, namespace, file):
    # check dirs in /data
    exec_command = ["/bin/sh",
                    "-c",
                    "ls -l /data"]
    resp = stream(
        client.CoreV1Api().connect_get_namespaced_pod_exec, pod_name, namespace,
        command=exec_command, stderr=True, stdin=False,
        stdout=True, tty=False)
    print("Response: " + resp)
    if file not in resp:
        print("checkPVCSubpathDatasetAccess Failed")
        return False
    
    return True


def main():
    if os.getenv("KUBERNETES_SERVICE_HOST") is None:
        config.load_kube_config()
    else:
        config.load_incluster_config()

    namespace = "default"

    base_name = "hbase"
    base_mount = fluidapi.Mount()
    base_mount.set_mount_info(
        base_name, "https://mirrors.ustc.edu.cn/apache/zookeeper/stable/")
    base_dataset = fluidapi.Dataset(base_name)
    base_dataset.add_mount(base_mount.dump())
    base_dataset.set_placement("Shared")
    base_runtime = fluidapi.assemble_runtime("alluxio-webufs").set_namespaced_name(namespace, base_name)

    name = "pvc-subpath"
    mount = fluidapi.Mount()
    mount.set_mount_info(name, "pvc://{}/hbase".format(base_name), "/")

    dataset = fluidapi.Dataset("pvc-subpath")
    dataset.set_namespaced_name(namespace, name)
    dataset.add_mount(mount.dump())
    dataset.set_placement("Shared")

    runtime = fluidapi.Runtime("AlluxioRuntime", name)
    runtime.set_replicas(1)
    runtime.set_tieredstore("MEM", "/dev/shm", "4Gi")
    runtime.set_namespaced_name(namespace, name)

    flow = TestFlow("Alluxio - Test PVC subpath Dataset")

    flow.append_step(
        SimpleStep(
            step_name="create base dataset",
            forth_fn=funcs.create_dataset_fn(base_dataset.dump()),
            back_fn=funcs.delete_dataset_and_runtime_fn(base_runtime.dump(), base_name)
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="create runtime",
            forth_fn=funcs.create_runtime_fn(base_runtime.dump()),
            back_fn=dummy_back
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check if base dataset is bound",
            forth_fn=funcs.check_dataset_bound_fn(base_name, namespace)
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check if base PV & PVC is ready",
            forth_fn=funcs.check_volume_resource_ready_fn(base_name, namespace)
        )
    )

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
            step_name="check if dataset is bound",
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

    flow.append_step(
        StatusCheckStep(
            step_name="checkPVCSubpathDatasetAccess",
            forth_fn=currying_fn(checkPVCSubpathDatasetAccess,
                                 pod_name="nginx-test", namespace=namespace, file="zookeeper"),
            timeout=10
        )
    )

    try:
        flow.run()
    except Exception as e:
        print(e)
        exit(1)

if __name__=="__main__":
    main()