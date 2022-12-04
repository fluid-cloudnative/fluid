"""
TestCase: Access WebUFS data
DDC Engine: Alluxio
Steps:
1. create Dataset(WebUFS) & Runtime
2. check if dataset is bound
3. check if persistentVolumeClaim & PV is created
4. submit data read job
5. wait until data read job completes
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
            "mounts": [{"mountPoint": "https://mirrors.bit.edu.cn/apache/zookeeper/stable/", "name": "hbase"}]
        }
    }

    my_alluxioruntime = {
        "apiVersion": "data.fluid.io/v1alpha1",
        "kind": "AlluxioRuntime",
        "metadata": {"name": "hbase"},
        "spec": {
            "replicas": 1,
            "podMetadata": {
                "labels": {
                    "foo": "bar"
                }
            },
            "master": {
                "podMetadata": {
                    "labels": {
                        "foo": "bar2",
                        "test1": "master-value",
                    }
                }
            },
            "worker": {
                "podMetadata": {
                    "labels": {
                        "foo": "bar2",
                        "test1": "worker-value",
                    }
                }
            },
            "tieredstore": {
                "levels": [{
                    "mediumtype": "MEM",
                    "path": "/dev/shm",
                    "quota": "2Gi",
                    "high": "0.95",
                    "low": "0.7"
                }]
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
            client.CoreV1Api().read_namespaced_persistent_volume_claim(name="hbase", namespace="default")
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
        args=["-c", "set -x; time cp -r /data/hbase ./"],
        volume_mounts=[client.V1VolumeMount(mount_path="/data", name="hbase-vol")]
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


def checkDataReadJobStatus():
    api = client.BatchV1Api()

    job_completed = False
    while not job_completed:
        response = api.read_namespaced_job_status(
            name="fluid-copy-test",
            namespace="default"
        )

        if response.status.succeeded is not None or \
            response.status.failed is not None:
            job_completed = True

        time.sleep(1)

    print("Data Read Job done.")

def cleanUp():
    batch_api = client.BatchV1Api()

    # Delete Data Read Job

    # See https://github.com/kubernetes-client/python/issues/234
    body = client.V1DeleteOptions(propagation_policy='Background')
    batch_api.delete_namespaced_job(name="fluid-copy-test", namespace="default", body=body)

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
    checkDataReadJobStatus()
    cleanUp()


if __name__ == '__main__':
    main()

