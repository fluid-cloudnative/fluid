# Demo - Automatic Cleanup Data Operation

## Background

Fluid's universal data operations describe operations such as data prefetch, data migration, elastic scaling, cache cleaning, metadata backup, and recovery.
Similar to the Kubernetes Job's automatic cleaning mechanism, we also provides automatic cleanup data operation, utilizing the Time-to-Live (TTL) mechanism to limit the lifecycle of data operations that have finished execution. This document will briefly demonstrate the utilization of these features.


## Prerequisites
Before everything we are going to do, please refer to [Installation Guide](../userguide/install.md) to install Fluid on your Kubernetes Cluster, and make sure all the components used by Fluid are ready like this:
```shell
$ kubectl get pod -n fluid-system
alluxioruntime-controller-5b64fdbbb-84pc6   1/1     Running   0          8h
csi-nodeplugin-fluid-fwgjh                  2/2     Running   0          8h
csi-nodeplugin-fluid-ll8bq                  2/2     Running   0          8h
dataset-controller-5b7848dbbb-n44dj         1/1     Running   0          8h
```

## Automatic cleanup data operation

### 1.Set up demo dataset

**Check the Dataset and AlluxioRuntime objects to be created**

```shell
$ cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: hbase
spec:
  mounts:
    - mountPoint: https://mirrors.tuna.tsinghua.edu.cn/apache/hbase/stable/
      name: hbase
---
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: hbase
spec:
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 2Gi
        high: "0.95"
        low: "0.7"
EOF
```
**Create the Dataset and AlluxioRuntime**

```shell
$ kubectl create -f dataset.yaml
```

**Wait for the Dataset and AlluxioRuntime to be ready**
You can check their status by running:
```shell
$ kubectl get datasets hbase
```

Dataset and Runtime are all ready if you see something like this:
```shell
NAME    UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
hbase   1.21GiB          0.00B    2.00GiB          0.0%                Bound      75s
```

### 2.Set up data operation
We use Dataload to show the automatic cleanup of data operations here.

**Check the Dataload objects to be created**

```shell
$ cat <<EOF > dataload.yaml
apiVersion: data.fluid.io/v1alpha1
kind: DataLoad
metadata:
  name: hbase-dataload
spec:
  dataset:
    name: hbase
    namespace: default
  ttlSecondsAfterFinished: 300
EOF
```

Here, we use the `spec.ttlSecondsAfterFinished` field to indicate how many seconds the data operation will be cleaned up after the job is Complete or Failed, in seconds.

**Create the Dataload**
```shell
$ kubectl apply -f dataload.yaml
```

**Watch the Dataload status**

```shell
$ kubectl get dataload -w 
NAME             DATASET   PHASE       AGE   DURATION
hbase-dataload   hbase     Executing   7s    Unfinished
hbase-dataload   hbase     Complete    29s   7s
hbase-dataload   hbase     Complete    5m29s   7s

$ kubectl get dataload hbase-dataload
Error from server (NotFound): dataloads.data.fluid.io "hbase-dataload" not found
```

It can be seen that 300s after the execution of `hbase-dataload` is completed, the dataload will be automatically cleaned.


## Warning: Time skew
Because the TTL-after-finished controller (Fluid dataset-controller) uses timestamps stored in the Data Operation to determine whether the TTL has expired or not, this feature is sensitive to time skew in your cluster, which may cause the control plane to clean up Job objects at the wrong time. Please be aware of this risk when setting a non-zero TTL.