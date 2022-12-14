"""
TestCase: Resources setting for alluxio runtime 
DDC Engine: Alluxio
Steps:
1. create Dataset(WebUFS) & Runtime with specified resource
2. check if dataset is bound
3. check if persistentVolumeClaim & PV is created
4. submit data read job
5. check if alluxio runtime resources are consistent with expected
6. clean up
"""

import time

from kubernetes import client, config


def createDatasetAndRuntime():
    api = client.CustomObjectsApi()
    my_dataset = {
        "apiVersion": "data.fluid.io/v1alpha1",
        "kind": "Dataset",
        "metadata": {"name": "hbase"},
        "spec": {
            "mounts": [{"mountPoint": "https://mirrors.bit.edu.cn/apache/hbase/stable/", "name": "hbase"}]
        }
    }
    my_alluxioruntime = {
        "apiVersion": "data.fluid.io/v1alpha1",
        "kind": "AlluxioRuntime",
        "metadata": {
            "name": "hbase"
        },
        "spec": {
            "replicas": 1,
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
                "resources": {
                    "requests": {
                        "cpu": "1000m",
                        "memory": "4Gi"
                    },
                    "limits": {
                        "cpu": "2000m",
                        "memory": "8Gi"
                    }
                }
            },
            "jobMaster": {
                "resources": {
                    "requests": {
                        "cpu": "1500m",
                        "memory": "4Gi"
                    },
                    "limits": {
                        "cpu": "2000m",
                        "memory": "8Gi"
                    }
                }
            },
            "worker": {
                "resources": {
                    "requests": {
                        "cpu": "1000m",
                        "memory": "4Gi"
                    },
                    "limits": {
                        "cpu": "2000m",
                        "memory": "8Gi"
                    }
                }
            },
            "jobWorker": {
                "resources": {
                    "requests": {
                        "cpu": "1000m",
                        "memory": "4Gi"
                    },
                    "limits": {
                        "cpu": "2000m",
                        "memory": "8Gi"
                    }
                }
            },
            "fuse": {
                "resources": {
                    "requests": {
                        "cpu": "1000m",
                        "memory": "4Gi"
                    },
                    "limits": {
                        "cpu": "2000m",
                        "memory": "8Gi"
                    }
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


def createDataReadJob():
    api = client.BatchV1Api()

    container = client.V1Container(
        name="busybox",
        image="busybox",
        command=["/bin/sh"],
        volume_mounts=[client.V1VolumeMount(
            mount_path="/data", name="hbase-vol")]
    )

    template = client.V1PodTemplateSpec(
        metadata=client.V1ObjectMeta(labels={"app": "dataread"}),
        spec=client.V1PodSpec(restart_policy="Never", containers=[container], volumes=[client.V1Volume(name="hbase-vol",
                                                                                                       persistent_volume_claim=client.V1PersistentVolumeClaimVolumeSource(
                                                                                                           claim_name="hbase"))])
    )

    spec = client.V1JobSpec(
        template=template,
        backoff_limit=4
    )

    job = client.V1Job(
        api_version="batch/v1",
        kind="Job",
        metadata=client.V1ObjectMeta(name="fluid-copy-test"),
        spec=spec
    )

    api.create_namespaced_job(namespace="default", body=job)
    print("Job created.")


def checkAlluxioruntimeResource():
    api = client.CoreV1Api()

    response = api.read_namespaced_pod(name="hbase-master-0", namespace="default")
    master_resource_check = True
    for container in response.spec.containers:
        if container.name == "master": # master
            if container.resources.limits["cpu"] == "2" and container.resources.limits["memory"] == "8Gi" and \
                container.resources.requests["cpu"] == "1" and container.resources.requests["memory"] == "4Gi":
                continue
            else:
                master_resource_check = False
        elif container.name == "jobmaster": # jobmaster
            if container.resources.limits["cpu"] == "2" and container.resources.limits["memory"] == "8Gi" and \
                container.resources.requests["cpu"] == "1500m" and container.resources.requests["memory"] == "4Gi":
                continue
            else:
                master_resource_check = False
    
    if master_resource_check:
        print("Master Resource Check Pass")

    response = api.read_namespaced_pod(name="hbase-worker-0", namespace="default")
    worker_resource_check = True
    for container in response.spec.containers:
        if container.resources.limits["cpu"] == "2" and container.resources.limits["memory"] == "8Gi" and \
            container.resources.requests["cpu"] == "1" and container.resources.requests["memory"] == "4Gi":
            continue
        else:
            worker_resource_check = False

    if worker_resource_check:
        print("Worker Resource Check Pass")

    # pod_list = api.list_namespaced_pod(namespace="default")
    # fuse_resource_check = True
    # for pod in pod_list.items:
    #     if "fuse" in pod.metadata.name:
    #         for container in pod.spec.containers:
    #             if container.resources.limits["cpu"] == "2" and container.resources.limits["memory"] == "8Gi" and \
    #                 container.resources.requests["cpu"] == "1" and container.resources.requests["memory"] == "4Gi":
    #                 continue
    #             else:
    #                 fuse_resource_check = False
    # if fuse_resource_check:
    #     print("Fuse Resource Check Pass")

    if not master_resource_check & worker_resource_check:
        return 1

def cleanUp():
    batch_api = client.BatchV1Api()

    # Delete Data Read Job

    # See https://github.com/kubernetes-client/python/issues/234
    body = client.V1DeleteOptions(propagation_policy='Background')
    batch_api.delete_namespaced_job(
        name="fluid-copy-test", namespace="default", body=body)

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
    createDataReadJob()
    res = checkAlluxioruntimeResource()
    cleanUp()
    if res == 1:
        exit(-1)


if __name__ == '__main__':
    main()
