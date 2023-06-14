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
1. apply minio service and deployment:
```
apiVersion: v1
kind: Service
metadata:
  name: minio
spec:
  type: ClusterIP
  ports:
    - port: 9000
      targetPort: 9000
      protocol: TCP
  selector:
    app: minio
---
apiVersion: apps/v1
kind: Deployment
metadata:
  # This name uniquely identifies the Deployment
  name: minio
spec:
  selector:
    matchLabels:
      app: minio
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        # Label is used as selector in the service.
        app: minio
    spec:
      containers:
      - name: minio
        # Pulls the default Minio image from Docker Hub
        image: minio/minio
        args:
        - server
        - /data
        env:
        # Minio access key and secret key
        - name: MINIO_ROOT_USER
          value: "minioadmin"
        - name: MINIO_ROOT_PASSWORD
          value: "minioadmin"
        ports:
        - containerPort: 9000
          hostPort: 9000
```
2. apply redis service and deployment:
```
apiVersion: v1
kind: Service
metadata:
  name: redis
spec:
  type: ClusterIP
  ports:
    - port: 6379
      targetPort: 6379
      protocol: TCP
  selector:
    app: redis
---
apiVersion: apps/v1
kind: Deployment
metadata:
  # This name uniquely identifies the Deployment
  name: redis
spec:
  selector:
    matchLabels:
      app: redis
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        # Label is used as selector in the service.
        app: redis
    spec:
      containers:
      - name: redis
        # Pulls the default Redis image from Docker Hub
        image: redis
        ports:
        - containerPort: 6379
          hostPort: 6379
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

import os
import sys
import time

project_root = os.path.dirname(os.path.dirname(os.path.dirname(__file__)))
sys.path.insert(0, project_root)

import fluid.fluidapi as fluidapi
import fluid.step_funcs as funcs
from framework.testflow import TestFlow
from framework.step import SimpleStep, StatusCheckStep, dummy_back, currying_fn, check
from framework.exception import TestError

from kubernetes import client, config

NODE_IP = "minio"
APP_NAMESPACE = "default"
SECRET_NAME = "jfs-secret"

NODE_NAME = ""


def create_redis_secret(namespace="default"):
    api = client.CoreV1Api()
    jfs_secret = {
        "apiVersion": "v1",
        "kind": "Secret",
        "metadata": {"name": SECRET_NAME},
        "stringData": {"metaurl": "redis://redis:6379/0", "accesskey": "minioadmin", "secretkey": "minioadmin"}
    }

    api.create_namespaced_secret(namespace=namespace, body=jfs_secret)
    print("Created secret {}".format(SECRET_NAME))


def create_datamigrate(datamigrate_name, dataset_name, namespace="default"):
    api = client.CustomObjectsApi()
    my_datamigrate = {
        "apiVersion": "data.fluid.io/v1alpha1",
        "kind": "DataMigrate",
        "metadata": {"name": datamigrate_name, "namespace": namespace},
        "spec": {
            "image": "registry.cn-hangzhou.aliyuncs.com/juicefs/juicefs-fuse",
            "imageTag": "nightly",
            "from": {
                "dataset": {"name": dataset_name, "namespace": namespace}
            },
            "to": {"externalStorage": {
                "uri": "minio://%s:9000/minio/test/" % NODE_IP,
                "encryptOptions": [
                    {"name": "access-key", "valueFrom": {"secretKeyRef": {"name": SECRET_NAME, "key": "accesskey"}}},
                    {"name": "secret-key", "valueFrom": {"secretKeyRef": {"name": SECRET_NAME, "key": "secretkey"}}},
                ]
            }}
        },
    }

    api.create_namespaced_custom_object(
        group="data.fluid.io",
        version="v1alpha1",
        namespace="default",
        plural="datamigrates",
        body=my_datamigrate,
    )
    print("Create datamigrate {}".format(datamigrate_name))


def create_cron_datamigrate(datamigrate_name, dataset_name, namespace="default"):
    api = client.CustomObjectsApi()
    my_datamigrate = {
        "apiVersion": "data.fluid.io/v1alpha1",
        "kind": "DataMigrate",
        "metadata": {"name": datamigrate_name, "namespace": namespace},
        "spec": {
            "image": "registry.cn-hangzhou.aliyuncs.com/juicefs/juicefs-fuse",
            "policy": "Cron",
            "schedule": "* * * * *",
            "imageTag": "nightly",
            "to": {"dataset": {"name": dataset_name, "namespace": namespace}},
            "from": {"externalStorage": {
                "uri": "minio://%s:9000/minio/test/" % NODE_IP,
                "encryptOptions": [
                    {"name": "access-key", "valueFrom": {"secretKeyRef": {"name": SECRET_NAME, "key": "accesskey"}}},
                    {"name": "secret-key", "valueFrom": {"secretKeyRef": {"name": SECRET_NAME, "key": "secretkey"}}},
                ]
            }}
        },
    }

    api.create_namespaced_custom_object(
        group="data.fluid.io",
        version="v1alpha1",
        namespace="default",
        plural="datamigrates",
        body=my_datamigrate,
    )
    print("Create datamigrate {}".format(datamigrate_name))


def check_cron_datamigrate(datamigrate_name, dataset_name, namespace="default"):
    api = client.CustomObjectsApi()

    for i in range(0, 60):
        resource = api.get_namespaced_custom_object(
            group="data.fluid.io",
            version="v1alpha1",
            name=datamigrate_name,
            namespace=namespace,
            plural="datamigrates"
        )
        dataset = api.get_namespaced_custom_object(
            group="data.fluid.io",
            version="v1alpha1",
            name=dataset_name,
            namespace=namespace,
            plural="datasets"
        )

        datamigrate_status = resource["status"]["phase"]
        dataset_status = dataset["status"]["phase"]
        opRef = dataset["status"]["operationRef"]
        if datamigrate_status == "Failed":
            print("Datamigrate {} is failed".format(datamigrate_name))
            return False
        if datamigrate_status == "Complete":
            if dataset_status != "Bound":
                print("Datamigrate {} is complete but dataset status {}".format(datamigrate_name, dataset_status))
                return False
            if opRef is not None and opRef["DataMigrate"] != "":
                print("Datamigrate {} is complete but dataset opRef {} is not None".format(datamigrate_name, opRef))
                return False
        if datamigrate_status == "Running":
            if dataset_status != "DataMigrating":
                print("Datamigrate {} is running but dataset status {}".format(datamigrate_name, dataset_status))
                return False
            if opRef is None or opRef["DataMigrate"] != datamigrate_name:
                print("Datamigrate {} is running but dataset opRef {}".format(datamigrate_name, opRef))
                return False
        time.sleep(1)

    return True


def check_datamigrate_complete(datamigrate_name, namespace="default"):
    api = client.CustomObjectsApi()

    resource = api.get_namespaced_custom_object(
        group="data.fluid.io",
        version="v1alpha1",
        name=datamigrate_name,
        namespace=namespace,
        plural="datamigrates"
    )

    if "status" in resource:
        if "phase" in resource["status"]:
            if resource["status"]["phase"] == "Complete":
                print("Datamigrate {} is complete.".format(datamigrate_name))
                return True

    return False


def get_worker_node_fn(dataset_name, namespace="default"):
    def check_internal():
        api = client.CoreV1Api()
        pod_name = "{}-worker-0".format(dataset_name)
        try:
            global NODE_NAME
            pod = api.read_namespaced_pod(name=pod_name, namespace=namespace)
            NODE_NAME = pod.spec.node_name
            return True
        except client.exceptions.ApiException as e:
            if e.status == 404:
                return False
            raise e

    return funcs.check(check_internal, retries=60, interval=1)


def create_check_cache_job(job_name, node_name, namespace="default"):
    if node_name == "":
        raise TestError("cannot check cache cleaned up given an empty node")

    print("Create check cache job")
    api = client.BatchV1Api()

    container = client.V1Container(
        name="demo",
        image="debian:buster",
        command=["/bin/bash"],
        args=["-c", "if [ $(find /dev/shm/* | grep chunks | wc -l) = 0 ]; then exit 0; else exit 1; fi"],
        volume_mounts=[client.V1VolumeMount(mount_path="/dev/shm/cache1", name="cache1"),
                       client.V1VolumeMount(mount_path="/dev/shm/cache2", name="cache2")]
    )

    template = client.V1PodTemplateSpec(
        metadata=client.V1ObjectMeta(labels={"app": "checkcache"}),
        spec=client.V1PodSpec(
            restart_policy="Never",
            containers=[container],
            volumes=[
                client.V1Volume(
                    name="cache1",
                    host_path=client.V1HostPathVolumeSource(path="/dev/shm/cache1")
                ),
                client.V1Volume(
                    name="cache2",
                    host_path=client.V1HostPathVolumeSource(path="/dev/shm/cache2")
                )
            ],
            node_name=node_name,
        )
    )

    spec = client.V1JobSpec(template=template, backoff_limit=4)

    job = client.V1Job(
        api_version="batch/v1",
        kind="Job",
        metadata=client.V1ObjectMeta(name=job_name, namespace=namespace),
        spec=spec
    )

    api.create_namespaced_job(namespace=namespace, body=job)
    print("Job {} created.".format("checkcache"))


def check_data_job_status(job_name, namespace="default"):
    api = client.BatchV1Api()

    count = 0
    while count < 300:
        count += 1
        response = api.read_namespaced_job_status(name=job_name, namespace=namespace)
        if response.status.succeeded is not None:
            print("Job {} completed.".format(job_name))
            return True
        if response.status.failed is not None:
            print("Job {} failed.".format(job_name))
            return False
        time.sleep(1)
    print("Job {} not completed within 300s.".format(job_name))
    return False


def clean_job(job_name, namespace="default"):
    batch_api = client.BatchV1Api()

    # See https://github.com/kubernetes-client/python/issues/234
    body = client.V1DeleteOptions(propagation_policy='Background')
    try:
        batch_api.delete_namespaced_job(name=job_name, namespace=namespace, body=body)
    except client.exceptions.ApiException as e:
        if e.status == 404:
            print("job {} deleted".format(job_name))
            return True

    count = 0
    while count < 300:
        count += 1
        print("job {} still exists...".format(job_name))
        try:
            batch_api.read_namespaced_job(name=job_name, namespace=namespace)
        except client.exceptions.ApiException as e:
            if e.status == 404:
                print("job {} deleted".format(job_name))
                return True
        time.sleep(1)

    print("job {} not deleted within 300s".format(job_name))
    return False


def check_datamigrate_clean_up(datamigrate_name, namespace="default"):
    def check_clean_up():
        api = client.CustomObjectsApi()
        print("datamigrate {} still exists...".format(datamigrate_name))
        try:
            api.get_namespaced_custom_object(
                group="data.fluid.io",
                version="v1alpha1",
                name=datamigrate_name,
                namespace=namespace,
                plural="datamigrates"
            )
        except client.exceptions.ApiException as e:
            if e.status == 404:
                return True

        return False

    timeout_check_fn = check(check_clean_up, 20, 5)
    timeout_check_fn()
    print("datamigrate {} deleted".format(datamigrate_name))


def clean_up_datamigrate(datamigrate_name, namespace):
    custom_api = client.CustomObjectsApi()
    custom_api.delete_namespaced_custom_object(
        group="data.fluid.io",
        version="v1alpha1",
        name=datamigrate_name,
        namespace=namespace,
        plural="datamigrates"
    )
    print("Datamigrate {} deleted".format(datamigrate_name))


def clean_up_secret(namespace="default"):
    core_api = client.CoreV1Api()
    try:
        core_api.delete_namespaced_secret(name=SECRET_NAME, namespace=namespace)
    except client.ApiException as e:
        if e.status != 404:
            raise e
    print("secret {} is cleaned up".format(SECRET_NAME))


def main():
    if os.getenv("KUBERNETES_SERVICE_HOST") is None:
        config.load_kube_config()
    else:
        config.load_incluster_config()

    name = "jfsdemo"
    datamigrate_name = "jfsdemo"
    cron_datamigrate_name = "cron-jfsdemo"
    dataload_name = "jfsdemo-warmup"
    namespace = "default"
    test_write_job = "demo-write"
    test_read_job = "demo-read"
    check_cache_job = "checkcache"

    dataset = fluidapi.assemble_dataset("juicefs-minio") \
        .set_namespaced_name(namespace, name)
    runtime = fluidapi.assemble_runtime("juicefs-minio") \
        .set_namespaced_name(namespace, name)
    dataload = fluidapi.DataLoad(name=dataload_name, namespace=namespace) \
        .set_target_dataset(name, namespace)

    flow = TestFlow("JuiceFS - Access Minio data")

    flow.append_step(
        SimpleStep(
            step_name="create jfs secrets",
            forth_fn=currying_fn(create_redis_secret, namespace=namespace),
            back_fn=currying_fn(clean_up_secret, namespace=namespace)
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="create dataset",
            forth_fn=funcs.create_dataset_fn(dataset.dump()),
            back_fn=funcs.delete_dataset_and_runtime_fn(runtime.dump(), name=name, namespace=namespace)
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
            step_name="get juicefs worker node info",
            forth_fn=get_worker_node_fn(dataset_name=name, namespace=namespace),
            back_fn=dummy_back
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="create data write job",
            forth_fn=funcs.create_job_fn(
                script="dd if=/dev/zero of=/data/allzero.file bs=100M count=10 && sha256sum /data/allzero.file",
                dataset_name=name, name=test_write_job, namespace=namespace),
            back_fn=funcs.delete_job_fn(name=test_write_job, namespace=namespace)
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check data write job status",
            forth_fn=funcs.check_job_status_fn(name=test_write_job, namespace=namespace)
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="create dataload",
            forth_fn=funcs.create_dataload_fn(dataload.dump()),
            back_fn=dummy_back
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check if dataload succeeds",
            forth_fn=funcs.check_dataload_job_status_fn(dataload_name=dataload_name, dataload_namespace=namespace)
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="create data read job",
            forth_fn=funcs.create_job_fn(script="time sha256sum /data/allzero.file && rm /data/allzero.file",
                                         dataset_name=name, name=test_read_job, namespace=namespace),
            back_fn=funcs.delete_job_fn(name=test_read_job, namespace=namespace)
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check data read job status",
            forth_fn=funcs.check_job_status_fn(name=test_read_job, namespace=namespace)
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="create DataMigrate job",
            forth_fn=currying_fn(create_datamigrate, datamigrate_name=datamigrate_name, dataset_name=name,
                                 namespace=namespace),
            back_fn=dummy_back
            # No need to clean up DataMigrate because of its ownerReference
            # back_fn=currying_fn(clean_up_datamigrate, datamigrate_name=datamigrate_name, namespace=namespace)
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check if DataMigrate succeeds",
            forth_fn=currying_fn(check_datamigrate_complete, datamigrate_name=datamigrate_name, namespace=namespace)
        )
    )

    flow.append_step(
        SimpleStep(
            step_name="create cron DataMigrate job",
            forth_fn=currying_fn(create_cron_datamigrate, datamigrate_name=cron_datamigrate_name, dataset_name=name,
                                 namespace=namespace),
            back_fn=dummy_back
        )
    )

    flow.append_step(
        StatusCheckStep(
            step_name="check if cron DataMigrate status correct",
            forth_fn=currying_fn(check_cron_datamigrate, datamigrate_name=cron_datamigrate_name,
                                 dataset_name=name, namespace=namespace)
        )
    )

    try:
        flow.run()
    except Exception as e:
        print(e)
        exit(1)

    print("> Post-Check: Cache cleaned up?")
    try:
        assert (NODE_NAME != "")
        create_check_cache_job(job_name=check_cache_job, node_name=NODE_NAME, namespace=namespace)
        if not check_data_job_status(check_cache_job, namespace=namespace):
            raise Exception("> FAIL: Job {} in normal mode failed.".format("checkcache"))
    except Exception as e:
        print(e)
        exit(1)
    finally:
        # clean up check cache job
        clean_job(check_cache_job, namespace=namespace)
    print("> Post-Check: PASSED")

    print("> Post-Check: Data Migrate deleted?")
    try:
        check_datamigrate_clean_up(datamigrate_name, name)
    except Exception as e:
        print(e)
        exit(1)
    print("> Post-Check: PASSED")


if __name__ == '__main__':
    main()
