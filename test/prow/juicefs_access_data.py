#   Copyright 2022 The Fluid Authors.
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.

"""
TestCase: Pod accesses Juicefs data
DDC Engine: Juicefs(Community) with local redis and minio

Prerequisite:
1. docker run -d -p 9000:9000 \
  --name minio \
  -e "MINIO_ROOT_USER=minioadmin" \
  -e "MINIO_ROOT_PASSWORD=minioadmin" \
  minio/minio server /data
2. docker run -itd --name redis -p 6379:6379 redis
3. Write down the node IP
4. Apply the following secret
```
apiVersion: v1
kind: Secret
metadata:
  name: jfs-secret
stringData:
  metaurl: redis://<node_ip>:6379/0
  access-key: minioadmin
  secret-key: minioadmin
```

Steps:
1. create Dataset & Runtime
2. check if dataset is bound
3. check if PVC & PV is created
4. submit data write job
5. wait until data write job completes
6. submit data read job
7. check if data content consistent
8. clean up
"""

import time

from kubernetes import client, config

NODE_IP = "minio"
APP_NAMESPACE = "default"
SECRET_NAME = "jfs-secret"


def create_redis_secret():
    api = client.CoreV1Api()
    jfs_secret = {
        "apiVersion": "v1",
        "kind": "Secret",
        "metadata": {"name": SECRET_NAME},
        "stringData": {"metaurl": "redis://redis:6379/0", "access-key": "minioadmin", "secret-key": "minioadmin"}
    }

    api.create_namespaced_secret(namespace=APP_NAMESPACE, body=jfs_secret)
    print("Created secret {}".format(SECRET_NAME))


def create_dataset_and_runtime(dataset_name):
    api = client.CustomObjectsApi()
    my_dataset = {
        "apiVersion": "data.fluid.io/v1alpha1",
        "kind": "Dataset",
        "metadata": {"name": dataset_name, "namespace": APP_NAMESPACE},
        "spec": {
            "mounts": [{
                "mountPoint": "juicefs:///",
                "name": "juicefs-community",
                "options": {"bucket": "http://%s:9000/minio/test" % NODE_IP, "storage": "minio"},
                "encryptOptions": [
                    {"name": "metaurl", "valueFrom": {"secretKeyRef": {"name": SECRET_NAME, "key": "metaurl"}}},
                    {"name": "access-key", "valueFrom": {"secretKeyRef": {"name": SECRET_NAME, "key": "access-key"}}},
                    {"name": "secret-key", "valueFrom": {"secretKeyRef": {"name": SECRET_NAME, "key": "secret-key"}}}
                ]
            }],
            "accessModes": ["ReadWriteMany"]
        }
    }

    my_juicefsruntime = {
        "apiVersion": "data.fluid.io/v1alpha1",
        "kind": "JuiceFSRuntime",
        "metadata": {"name": dataset_name, "namespace": APP_NAMESPACE},
        "spec": {
            "replicas": 1,
            "tieredstore": {"levels": [{"mediumtype": "MEM", "path": "/dev/shm", "quota": "40960", "low": "0.1"}]}
        }
    }

    api.create_namespaced_custom_object(
        group="data.fluid.io",
        version="v1alpha1",
        namespace="default",
        plural="datasets",
        body=my_dataset,
    )
    print("Create dataset {}".format(dataset_name))

    api.create_namespaced_custom_object(
        group="data.fluid.io",
        version="v1alpha1",
        namespace="default",
        plural="juicefsruntimes",
        body=my_juicefsruntime
    )
    print("Create juicefs runtime {}".format(dataset_name))


def check_dataset_bound(dataset_name):
    api = client.CustomObjectsApi()

    while True:
        resource = api.get_namespaced_custom_object(
            group="data.fluid.io",
            version="v1alpha1",
            name=dataset_name,
            namespace=APP_NAMESPACE,
            plural="datasets"
        )

        if "status" in resource:
            print(resource["status"])
            if "phase" in resource["status"]:
                if resource["status"]["phase"] == "Bound":
                    break
        time.sleep(1)


def check_volume_resources_ready(dataset_name):
    pv_name = "{}-{}".format(APP_NAMESPACE, dataset_name)
    pvc_name = dataset_name
    while True:
        try:
            client.CoreV1Api().read_persistent_volume(name=pv_name)
        except client.exceptions.ApiException as e:
            if e.status == 404:
                time.sleep(1)
                continue

        try:
            client.CoreV1Api().read_namespaced_persistent_volume_claim(name=pvc_name, namespace=APP_NAMESPACE)
        except client.exceptions.ApiException as e:
            if e.status == 404:
                time.sleep(1)
                continue

        print("PersistentVolume {} & PersistentVolumeClaim {} Ready.".format(pv_name, pvc_name))
        break


def create_eci_data_write_job(job_name, use_sidecar=False):
    api = client.BatchV1Api()

    container = client.V1Container(
        name="demo",
        image="debian:buster",
        command=["/bin/bash"],
        args=["-c", "dd if=/dev/zero of=/data/allzero.file bs=100M count=10 && sha256sum /data/allzero.file"],
        volume_mounts=[client.V1VolumeMount(mount_path="/data", name="demo")]
    )

    template = client.V1PodTemplateSpec(
        metadata=client.V1ObjectMeta(
            labels={"app": "datawrite"},
            annotations={
                "k8s.aliyun.com/eci-image-cache": "true",
                "k8s.aliyun.com/eci-use-specs": "ecs.g6e.4xlarge"
            }
        ),
        spec=client.V1PodSpec(
            restart_policy="Never",
            containers=[container],
            volumes=[client.V1Volume(
                name="demo",
                persistent_volume_claim=client.V1PersistentVolumeClaimVolumeSource(claim_name="jfsdemo")
            )]
        )
    )
    if use_sidecar:
        template.metadata.labels["alibabacloud.com/eci"] = "true"
        template.metadata.labels["alibabacloud.com/fluid-sidecar-target"] = "eci"

    spec = client.V1JobSpec(template=template, backoff_limit=4)

    job = client.V1Job(
        api_version="batch/v1",
        kind="Job",
        metadata=client.V1ObjectMeta(name=job_name, namespace=APP_NAMESPACE),
        spec=spec
    )

    api.create_namespaced_job(namespace=APP_NAMESPACE, body=job)
    print("Job {} created.".format(job_name))


def create_eci_data_read_job(job_name, use_sidecar=False):
    api = client.BatchV1Api()

    container = client.V1Container(
        name="demo",
        image="debian:buster",
        command=["/bin/bash"],
        args=["-c", "time sha256sum /data/allzero.file && rm /data/allzero.file"],
        volume_mounts=[client.V1VolumeMount(mount_path="/data", name="demo")]
    )

    template = client.V1PodTemplateSpec(
        metadata=client.V1ObjectMeta(
            labels={"app": "dataread"},
            annotations={
                "k8s.aliyun.com/eci-image-cache": "true",
                "k8s.aliyun.com/eci-use-specs": "ecs.g6e.4xlarge"
            }
        ),
        spec=client.V1PodSpec(
            restart_policy="Never",
            containers=[container],
            volumes=[client.V1Volume(
                name="demo",
                persistent_volume_claim=client.V1PersistentVolumeClaimVolumeSource(claim_name="jfsdemo")
            )]
        )
    )
    if use_sidecar:
        template.metadata.labels["alibabacloud.com/eci"] = "true"
        template.metadata.labels["alibabacloud.com/fluid-sidecar-target"] = "eci"

    spec = client.V1JobSpec(template=template, backoff_limit=4)

    job = client.V1Job(
        api_version="batch/v1",
        kind="Job",
        metadata=client.V1ObjectMeta(name=job_name, namespace=APP_NAMESPACE),
        spec=spec
    )

    api.create_namespaced_job(namespace=APP_NAMESPACE, body=job)
    print("Data Read Job {} created.".format(job_name))


def check_data_job_status(job_name):
    api = client.BatchV1Api()

    job_completed = False
    while not job_completed:
        response = api.read_namespaced_job_status(name=job_name, namespace=APP_NAMESPACE)
        if response.status.succeeded is not None or response.status.failed is not None:
            job_completed = True
        time.sleep(1)
    print("Data Write Job {} done.".format(job_name))


def clean_job(job_name):
    batch_api = client.BatchV1Api()
    # Delete Data Read Job
    # See https://github.com/kubernetes-client/python/issues/234
    body = client.V1DeleteOptions(propagation_policy='Background')
    batch_api.delete_namespaced_job(name=job_name, namespace=APP_NAMESPACE, body=body)

    job_delete = False
    while not job_delete:
        print("job {} still exists...".format(job_name))
        try:
            batch_api.read_namespaced_job(name=job_name, namespace=APP_NAMESPACE)
        except client.exceptions.ApiException as e:
            if e.status == 404:
                job_delete = True
                continue
        time.sleep(1)
    print("job {} deleted".format(job_name))


def clean_up_dataset_and_runtime(dataset_name):
    custom_api = client.CustomObjectsApi()
    custom_api.delete_namespaced_custom_object(
        group="data.fluid.io",
        version="v1alpha1",
        name=dataset_name,
        namespace=APP_NAMESPACE,
        plural="datasets"
    )
    print("Dataset [] deleted".format(dataset_name))

    runtime_delete = False
    while not runtime_delete:
        print("runtime {} still exists...".format(dataset_name))
        try:
            custom_api.get_namespaced_custom_object(
                group="data.fluid.io",
                version="v1alpha1",
                name=dataset_name,
                namespace=APP_NAMESPACE,
                plural="juicefsruntimes"
            )
        except client.exceptions.ApiException as e:
            if e.status == 404:
                runtime_delete = True
                continue
        time.sleep(1)
    print("Runtime CR {} is cleaned up".format(dataset_name))


def clean_up_secret():
    core_api = client.CoreV1Api()
    core_api.delete_namespaced_secret(name=SECRET_NAME, namespace=APP_NAMESPACE)
    print("secret {} is cleaned up".format(SECRET_NAME))


def main():
    config.load_kube_config()

    # ------- create secret -------
    create_redis_secret()

    # ********************************
    # ------- test normal mode -------
    # ********************************
    # 1. create dataset and runtime
    dataset_name = "jfsdemo"
    create_dataset_and_runtime(dataset_name)
    check_dataset_bound(dataset_name)
    check_volume_resources_ready(dataset_name)

    # 2. create write & read data job
    test_write_job = "demo-write"
    create_eci_data_write_job(test_write_job)
    check_data_job_status(test_write_job)
    test_read_job = "demo-read"
    create_eci_data_read_job(test_read_job)
    check_data_job_status(test_read_job)

    # 3. clean up write & read data job
    clean_job(test_write_job)
    clean_job(test_read_job)

    # 4. clean up dataset and runtime
    clean_up_dataset_and_runtime(dataset_name)

    # ********************************
    # ------- test sidecar mode -------
    # ********************************
    # 1. create dataset and runtime
    dataset_name = "jfsdemo-sidecar"
    create_dataset_and_runtime(dataset_name)
    check_dataset_bound(dataset_name)
    check_volume_resources_ready(dataset_name)

    # 2. create write & read data job
    test_write_job = "demo-write-sidecar"
    create_eci_data_write_job(test_write_job, use_sidecar=True)
    check_data_job_status(test_write_job)
    test_read_job = "demo-read-sidecar"
    create_eci_data_read_job(test_read_job, use_sidecar=True)
    check_data_job_status(test_read_job)

    # 3. clean up write & read data job
    clean_job(test_write_job)
    clean_job(test_read_job)

    # 4. clean up dataset and runtime
    clean_up_dataset_and_runtime(dataset_name)

    # ------- clean up secret -------
    clean_up_secret()


if __name__ == '__main__':
    main()
