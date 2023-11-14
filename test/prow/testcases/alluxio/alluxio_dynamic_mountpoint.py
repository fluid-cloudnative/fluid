"""
TestCase: Alluxio dynamic changes mountpoints
DDC Engine: Alluxio
Steps:
1. create Dataset(WebUFS) & Runtime with two mountpoint
2. check if dataset is bound
3. check if persistentVolumeClaim & PV is created
4. check alluxioruntime mountpoint and data
5. change dataset mountpoint and update
6. check dataset is bound and mountpoint change
7. check if alluxio master recover after crash
8. clean up
"""

import os
import sys
import time

project_root = os.path.dirname(os.path.dirname(os.path.dirname(__file__)))
sys.path.insert(0, project_root)

import fluid.fluidapi as fluidapi
import fluid.step_funcs as funcs
from framework.testflow import TestFlow
from framework.step import SimpleStep, StatusCheckStep, dummy_back, currying_fn
from framework.exception import TestError

from kubernetes import client, config
from kubernetes.stream import stream

def checkAlluxioruntimeMountpoint(dataset_name, namespace, mp1, mp2):
    exec_command = ["/bin/sh",
                    "-c",
                    "alluxio fs mount"]
    resp = stream(
        client.CoreV1Api().connect_get_namespaced_pod_exec, "{}-master-0".format(dataset_name), namespace,
        command=exec_command, stderr=True, stdin=False,
        stdout=True, tty=False, container='alluxio-master')
    print("Response: " + resp)
    if mp1 not in resp or mp2 not in resp:
        print("checkAlluxioruntimeMountpoint Failed")
        return False
    
    return True

def change_dataset_mount_point(new_dataset, name, namespace):
    client.CustomObjectsApi().patch_namespaced_custom_object(
        name=name,
        group="data.fluid.io",
        version="v1alpha1",
        namespace=namespace,
        plural="datasets",
        body=new_dataset,
    )

    print("new dataset patched")

def check_dataset_mount_change(name, namespace):
    resource = client.CustomObjectsApi().get_namespaced_custom_object(
            group="data.fluid.io",
            version="v1alpha1",
            name=name,
            namespace=namespace,
            plural="datasets"
        )

    # print(resource)

    if "status" in resource:
        if "mounts" in resource["status"]:
            print(resource["status"]["mounts"])
            if resource["status"]["mounts"][0]["name"] == "zookeeper" or resource["status"]["mounts"][1]["name"] == "zookeeper":
                return True
    
    return False

def checkRecoverAfterCrash():
    # exec the master pod and kill
    exec_command = ["/bin/sh",
                    "-c",
                    "kill 1"]
    resp = stream(
        client.CoreV1Api().connect_get_namespaced_pod_exec, "hbase-master-0", "default",
        command=exec_command, stderr=True, stdin=False,
        stdout=True, tty=False, container='alluxio-master')
    print("Response: " + resp)

    api = client.CoreV1Api()
    time.sleep(1)
    response = api.read_namespaced_pod(
        name="hbase-master-0", namespace="default")
    while response.status.phase != "Running":
        time.sleep(1)
        response = api.read_namespaced_pod(
            name="hbase-master-0", namespace="default")
        print(response)


def cleanUp():
    custom_api = client.CustomObjectsApi()

    custom_api.delete_namespaced_custom_object(
        group="data.fluid.io",
        version="v1alpha1",
        name="hbase",
        namespace="default",
        plural="datasets"
    )

    runtimeDelete = False
    while not runtimeDelete:
        print("runtime still exists...")
        try:
            runtime = custom_api.get_namespaced_custom_object(
                group="data.fluid.io",
                version="v1alpha1",
                name="hbase",
                namespace="default",
                plural="alluxioruntimes"
            )
        except client.exceptions.ApiException as e:
            if e.status == 404:
                runtimeDelete = True
                continue

        time.sleep(1)


def main():
    if os.getenv("KUBERNETES_SERVICE_HOST") is None:
        config.load_kube_config()
    else:
        config.load_incluster_config()

    name = "hbase-mount"
    namespace = "default"

    mount = fluidapi.Mount()
    mount.set_mount_info("hadoop", "https://mirrors.bit.edu.cn/apache/hadoop/common/stable/")

    dataset = fluidapi.assemble_dataset("alluxio-webufs") \
        .set_namespaced_name(namespace, name) \
        .add_mount(mount.dump())

    runtime = fluidapi.assemble_runtime("alluxio-webufs") \
        .set_namespaced_name(namespace, name)

    flow = TestFlow("Alluxio - Test Dynamically Change Mountpoints")

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
        StatusCheckStep(
            step_name="check alluxio runtime mount point",
            forth_fn=currying_fn(checkAlluxioruntimeMountpoint, dataset_name=name, namespace=namespace, mp1="zookeeper", mp2="hadoop"),
            timeout=10
        )
    )

    new_mount = fluidapi.Mount()
    new_mount.set_mount_info("hbase", "https://mirrors.bit.edu.cn/apache/hbase/stable/")


    new_dataset = fluidapi.assemble_dataset("alluxio-webufs") \
        .set_namespaced_name(namespace, name) \
        .add_mount(new_mount.dump())
    
    flow.append_step(
        SimpleStep(
            step_name="patch new mount point to dataset",
            forth_fn=currying_fn(change_dataset_mount_point, new_dataset=new_dataset.dump(), name=name, namespace=namespace),
            back_fn=dummy_back
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check if mount point changed",
            forth_fn=currying_fn(check_dataset_mount_change, name=name, namespace=namespace)
        )
    )

    try:
        flow.run()
    except Exception as e:
        print(e)
        exit(1)

    # createDatasetAndRuntime()
    # checkDatasetBound()
    # checkVolumeResourcesReady()
    # res_check_mountpoint0 = checkAlluxioruntimeMountpoint("hbase", "hadoop")
    # changeDatasetMountpoint()
    # res_check_mountpoint1 = checkAlluxioruntimeMountpoint("hbase", "zookeeper")
    # checkRecoverAfterCrash()
    # cleanUp()
    # if res_check_mountpoint0 == 1 or res_check_mountpoint1 == 1:
    #     exit(-1)


if __name__ == "__main__":
    main()
