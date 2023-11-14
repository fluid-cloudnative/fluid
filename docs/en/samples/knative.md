# Demo - How to run in a Serverless environment with Knative as an example

This example uses the open source framework Knative as an example to demonstrate how to perform unified data acceleration via Fluid in a Serverless environment. This example uses AlluxioRuntime as an example, and in fact Fluid supports all Runtime running in a Serverless environment.

## Installation

1.Install Knative Serving v1.2 according to the [Knative documentation](https://knative.dev/docs/install/serving/install-serving-with-yaml/), you need to enable the [kubernetes.Deploymentspec-persistent-volume-claim](https://github.com/knative/serving/blob/main/config/core/configmaps/features.yaml#L156) option.


Check if Knative's components are working properly

```
kubectl get Deployments -n knative-serving
```

> Note: This document is just for demonstration purpose, please refer to the best practices of Knative documentation for Knative deployment in production environment. Also, since the container images of Knative are in the gcr.io image repository, please make sure the images are reachable. If you are using AliCloud, you can also use [AliCloud ACK](https://help.aliyun.com/document_detail/121508.html) hosting service directly to reduce the complexity of configuring Knative.

2.Please refer to the [installation documentation](../userguide/install.md) to install the latest Fluid, and check that the Fluid components are working properly after installation (this document uses AlluxioRuntime as an example):

```shell
$ kubectl get deploy -n fluid-system
NAME                        READY   UP-TO-DATE   AVAILABLE   AGE
alluxioruntime-controller   1/1     1            1           18m
dataset-controller          1/1     1            1           18m
fluid-webhook               1/1     1            1           18m
```

Typically, you can see a Deployment named `dataset-controller`, a Deployment named `alluxioruntime-controller`, and a Deployment named `fluid-webhook`.

## Configuration

## Running

**Create dataset and runtime**

Create Runtime resources for different types of Runtime, as well as a Dataset with the same name. Here is the example of AlluxioRuntime, the following is the Dataset content:

```yaml
$ cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: serverless-data
spec:
  mounts:
    - mountPoint: https://mirrors.bit.edu.cn/apache/hbase/stable/
      name: hbase
      path: "/"
  accessModes:
    - ReadOnlyMany
---
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: serverless-data
spec:
  replicas: 2
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 2Gi
        high: "0.95"
        low: "0.7"
EOF
```

Execute the Create Dataset operation:

```
$ kubectl create -f dataset.yaml
```

Check Dataset Status:

```shell
$ kubectl get alluxio
NAME              MASTER PHASE   WORKER PHASE   FUSE PHASE   AGE
serverless-data   Ready          Ready          Ready        4m52s
$ kubectl get dataset
NAME              UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
serverless-data   566.22MiB        0.00B    4.00GiB          0.0%                Bound   4m52s
```

**Creating Knative Serving Resource Objects**

```yaml
$ cat<<EOF >serving.yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: model-serving
spec:
  template:
    metadata:
      labels:
        app: model-serving
        serverless.fluid.io/inject: "true"
      annotations:
        autoscaling.knative.dev/target: "10"
        autoscaling.knative.dev/scaleDownDelay: "30m"
        autoscaling.knative.dev/minScale: "1"
    spec:
      containers:
        - image: fluidcloudnative/serving
          ports:
            - name: http1
              containerPort: 8080
          env:
            - name: TARGET
              value: "World"
          volumeMounts:
            - mountPath: /data
              name: data
              readOnly: true
      volumes:
        - name: data
          persistentVolumeClaim:
            claimName: serverless-data
            readOnly: true
  EOF
$ kubectl create -f serving.yaml
service.serving.knative.dev/model-serving created
```

Please configure `serverless.fluid.io/inject: "true"` in the label of the podSpec or podTemplateSpec.


**Check if Knative Serving is created and check if fuse-container is injected**

```shell
$ kubectl get po
NAME                                              READY   STATUS    RESTARTS   AGE
model-serving-00001-deployment-64d674d75f-46vvf   3/3     Running   0          76s
serverless-data-master-0                          2/2     Running   0          16m
serverless-data-worker-0                          2/2     Running   0          16m
serverless-data-worker-1                          2/2     Running   0          16m
$ kubectl get po model-serving-00001-deployment-64d674d75f-46vvf -oyaml| grep -i fluid-fuse -B 3
          - /opt/alluxio/integration/fuse/bin/alluxio-fuse
          - unmount
          - /runtime-mnt/alluxio/default/serverless-data/alluxio-fuse
    name: fluid-fuse
```

Checking the Knative Serving startup speed, you can see that the startup data loading time is **92s**.

```shell
$ kubectl logs model-serving-00001-deployment-64d674d75f-46vvf -c user-container
Begin loading models at 16:29:02

real  1m32.639s
user  0m0.001s
sys 0m1.305s
Finish loading models at 16:29:45
2022-02-15 16:29:45 INFO Hello world sample started.
```

**Clean up Knative serving instances**

```
$ kubectl delete -f serving.yaml
```

**Execute data warm-up**

Create the dataload object and check its status:

```yaml
$ cat<<EOF >dataload.yaml
apiVersion: data.fluid.io/v1alpha1
kind: DataLoad
metadata:
  name: serverless-dataload
spec:
  dataset:
    name: serverless-data
    namespace: default
  EOF
$ kubectl create -f dataload.yaml
dataload.data.fluid.io/serverless-dataload created
$ kubectl get dataload
NAME                  DATASET           PHASE      AGE     DURATION
serverless-dataload   serverless-data   Complete   2m43s   34s
```

Check the cache status at this point, the data is now fully cached in the cluster.

```
$ kubectl get dataset
NAME              UFS TOTAL SIZE   CACHED      CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
serverless-data   566.22MiB        566.22MiB   4.00GiB          100.0%              Bound   33m
```

Create Knative service again：

```shell
$ kubectl create -f serving.yaml
service.serving.knative.dev/model-serving created
```

Checking the boot time at this point reveals that the current boot time for loading data is **3.66s**, which becomes **1/20** of the performance without warm-up.


```
kubectl logs model-serving-00001-deployment-6cb54f94d7-dbgxf -c user-container
Begin loading models at 18:38:23

real  0m3.666s
user  0m0.000s
sys 0m1.367s
Finish loading models at 18:38:25
2022-02-15 18:38:25 INFO Hello world sample started.
```

> Note: This example uses Knative serving. If you don't have a Knative environment, you can also experiment with Deployment.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: model-serving
spec:
  selector:
    matchLabels:
      app: model-serving
  template:
    metadata:
      labels:
        app: model-serving
        serverless.fluid.io/inject: "true"
    spec:
      containers:
        - image: fluidcloudnative/serving
          name: serving
          ports:
            - name: http1
              containerPort: 8080
          env:
            - name: TARGET
              value: "World"
          volumeMounts:
            - mountPath: /data
              name: data
      volumes:
        - name: data
          persistentVolumeClaim:
            claimName: serverless-data
```

> Note: The default sidecar injection mode does not enable cached directory short-circuit reads, if you need to enable this capability, you can configure the parameter `cachedir.sidecar.fluid.io/inject` to `true` in the labels. 

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: model-serving
spec:
  selector:
    matchLabels:
      app: model-serving
  template:
    metadata:
      labels:
        app: model-serving
        serverless.fluid.io/inject: "true"
        cachedir.sidecar.fluid.io/inject: "true"
    spec:
      containers:
        - image: fluidcloudnative/serving
          name: serving
          ports:
            - name: http1
              containerPort: 8080
          env:
            - name: TARGET
              value: "World"
          volumeMounts:
            - mountPath: /data
              name: data
      volumes:
        - name: data
          persistentVolumeClaim:
            claimName: serverless-data
```