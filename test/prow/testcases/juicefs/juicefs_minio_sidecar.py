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
from framework.step import SimpleStep, StatusCheckStep, dummy_back, currying_fn
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

    name = "jfsdemo-sidecar"
    datamigrate_name = "jfsdemo"
    namespace = "default"
    test_write_job = "demo-write-sidecar"
    test_read_job = "demo-read-sidecar"

    dataset = fluidapi.assemble_dataset("juicefs-minio") \
        .set_namespaced_name(namespace, name)
    runtime = fluidapi.assemble_runtime("juicefs-minio") \
        .set_namespaced_name(namespace, name)

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
            step_name="create data write job",
            forth_fn=funcs.create_job_fn(script="dd if=/dev/zero of=/data/allzero.file bs=100M count=10 && sha256sum /data/allzero.file", dataset_name=name, name=test_write_job, namespace=namespace, serverless=True),
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
            step_name="create data read job",
            forth_fn=funcs.create_job_fn(script="time sha256sum /data/allzero.file && rm /data/allzero.file", dataset_name=name, name=test_read_job, namespace=namespace, serverless=True),
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
            forth_fn=currying_fn(create_datamigrate, datamigrate_name=datamigrate_name, dataset_name=name, namespace=namespace),
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

    try:
        flow.run()
    except Exception as e:
        print(e)
        exit(1)


if __name__ == '__main__':
    main()
