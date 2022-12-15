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

import subprocess
import time

from kubernetes import client, config
from kubernetes.stream import stream


def createDatasetAndRuntime():
    api = client.CustomObjectsApi()
    my_dataset = {
        "apiVersion": "data.fluid.io/v1alpha1",
        "kind": "Dataset",
        "metadata": {
            "name": "hbase"
        },
        "spec": {
            "mounts": [
                {
                    "mountPoint": "https://mirrors.bit.edu.cn/apache/hbase/stable/",
                    "name": "hbase"
                },
                {
                    "mountPoint": "https://mirrors.bit.edu.cn/apache/hadoop/common/stable/",
                    "name": "hadoop"
                }
            ]
        }
    }

    # let alluxio master in current node
    hostname = subprocess.run("hostname", stdout=subprocess.PIPE)
    hostname = hostname.stdout.decode().replace('\n', '')

    my_alluxioruntime = {
        "apiVersion": "data.fluid.io/v1alpha1",
        "kind": "AlluxioRuntime",
        "metadata": {
            "name": "hbase"
        },
        "spec": {
            "replicas": 2,
            "tieredstore": {
                "levels": [
                    {
                        "mediumtype": "MEM",
                        "path": "/dev/shm",
                        "quota": "2Gi",
                        "high": "0.95",
                        "low": "0.7"
                    }
                ]
            },
            "master": {
                "nodeSelector": {
                    "kubernetes.io/hostname": hostname
                }
            }
        }
    }

    api.create_namespaced_custom_object(
        group="data.fluid.io",
        version="v1alpha1",
        namespace="default",
        plural="datasets",
        body=my_dataset,
    )

    print("Created dataset.")

    api.create_namespaced_custom_object(
        group="data.fluid.io",
        version="v1alpha1",
        namespace="default",
        plural="alluxioruntimes",
        body=my_alluxioruntime
    )

    print("Created alluxioruntime.")


def checkDatasetBound():
    api = client.CustomObjectsApi()

    while True:
        resource = api.get_namespaced_custom_object(
            group="data.fluid.io",
            version="v1alpha1",
            name="hbase",
            namespace="default",
            plural="datasets"
        )

        print(resource)

        if "status" in resource:
            if "phase" in resource["status"]:
                if resource["status"]["phase"] == "Bound":
                    break
        time.sleep(1)
        print(resource)


def checkVolumeResourcesReady():
    while True:
        try:
            client.CoreV1Api().read_persistent_volume(name="default-hbase")
        except client.exceptions.ApiException as e:
            if e.status == 404:
                time.sleep(1)
                continue

        try:
            client.CoreV1Api().read_namespaced_persistent_volume_claim(
                name="hbase", namespace="default")
        except client.exceptions.ApiException as e:
            if e.status == 404:
                time.sleep(1)
                continue

        print("PersistentVolume & PersistentVolumeClaim Ready.")
        break


def checkAlluxioruntimeMountpoint(dataset1, dataset2):
    exec_command = ["/bin/sh",
                    "-c",
                    "alluxio fs mount"]
    resp = stream(
        client.CoreV1Api().connect_get_namespaced_pod_exec, "hbase-master-0", "default",
        command=exec_command, stderr=True, stdin=False,
        stdout=True, tty=False, container='alluxio-master')
    print("Response: " + resp)
    if dataset1 not in resp or dataset2 not in resp:
        print("checkAlluxioruntimeMountpoint Failed")
        return 1


def changeDatasetMountpoint():
    new_dataset = {
        "apiVersion": "data.fluid.io/v1alpha1",
        "kind": "Dataset",
        "metadata": {
            "name": "hbase"
        },
        "spec": {
            "mounts": [
                {
                    "mountPoint": "https://mirrors.bit.edu.cn/apache/hbase/stable/",
                    "name": "hbase"
                },
                {
                    "mountPoint": "https://mirrors.bit.edu.cn/apache/zookeeper/stable/",
                    "name": "zookeeper"
                }
            ]
        }
    }

    hostname = subprocess.run("hostname", stdout=subprocess.PIPE)
    hostname = hostname.stdout.decode().replace('\n', '')
    new_alluxioruntime = {
        "apiVersion": "data.fluid.io/v1alpha1",
        "kind": "AlluxioRuntime",
        "metadata": {
            "name": "hbase"
        },
        "spec": {
            "replicas": 2,
            "tieredstore": {
                "levels": [
                    {
                        "mediumtype": "MEM",
                        "path": "/dev/shm",
                        "quota": "2Gi",
                        "high": "0.95",
                        "low": "0.7"
                    }
                ]
            },
            "master": {
                "nodeSelector": {
                    "kubernetes.io/hostname": hostname
                }
            }
        }
    }

    client.CustomObjectsApi().patch_namespaced_custom_object(
        name="hbase",
        group="data.fluid.io",
        version="v1alpha1",
        namespace="default",
        plural="datasets",
        body=new_dataset,
    )

    client.CustomObjectsApi().patch_namespaced_custom_object(
        name="hbase",
        group="data.fluid.io",
        version="v1alpha1",
        namespace="default",
        plural="alluxioruntimes",
        body=new_alluxioruntime,
    )

    time.sleep(1)
    while True:
        resource = client.CustomObjectsApi().get_namespaced_custom_object(
            group="data.fluid.io",
            version="v1alpha1",
            name="hbase",
            namespace="default",
            plural="datasets"
        )

        print(resource)

        if "status" in resource:
            if "mounts" in resource["status"]:
                print(resource["status"]["mounts"])
                if resource["status"]["mounts"][0]["name"] == "zookeeper" or resource["status"]["mounts"][1]["name"] == "zookeeper":
                    break

        time.sleep(1)


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
    response = api.read_namespaced_pod(name="hbase-master-0", namespace="default")
    while response.status.phase != "Running":
        time.sleep(1)
        response = api.read_namespaced_pod(
            name="hbase-master-0", namespace="default")


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
    # config.load_kube_config()
    config.load_incluster_config()

    createDatasetAndRuntime()
    checkDatasetBound()
    checkVolumeResourcesReady()
    res_check_mountpoint0 = checkAlluxioruntimeMountpoint("hbase", "hadoop")
    changeDatasetMountpoint()
    res_check_mountpoint1 = checkAlluxioruntimeMountpoint("hbase", "zookeeper")
    checkRecoverAfterCrash()
    cleanUp()
    if res_check_mountpoint0 == 1 or res_check_mountpoint1 == 1:
        exit(-1)


if __name__ == "__main__":
    main()
