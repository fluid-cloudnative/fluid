# Demo - Cache Runtime Manually Scaling 

## Prerequisites

- [Fluid](https://github.com/fluid-cloudnative/fluid)(version >= 0.5.0)

Please refer to the [Fluid installation](../userguide/install.md) to complete the installation

## Set Up Workspace
```shell
$ mkdir <any-path>/dataset_scale
$ cd <any-path>/dataset_scale
```

## Running

**Create Dataset and AlluxioRuntime resource objects**
```yaml
$ cat << EOF > dataset.yaml
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

In the above example, we set `AlluxioRuntime.spec.replicas` to 1, which means we will start an Alluxio cluster with one worker to cache the data in the dataset.

```
$ kubectl create -f dataset.yaml
dataset.data.fluid.io/hbase created
alluxioruntime.data.fluid.io/hbase created
```
After the Alluxio cluster has started normally, we can see that the Dataset and AlluxioRuntime created before are in the following state:

Status of the Alluxio components:
```
$ kubectl get pod
NAME                 READY   STATUS    RESTARTS   AGE
hbase-master-0       2/2     Running   0          3m50s
hbase-worker-0       2/2     Running   0          3m15s
```

Status of the Dataset：
```
$ kubectl get dataset hbase
NAME    UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
hbase   544.77MiB        0.00B    2.00GiB          0.0%                Bound   3m28s
```

Status of the AlluxioRuntime:
```
$ kubectl get alluxioruntime hbase -o wide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE   AGE
hbase   1               1                 Ready          1               1                 Ready          0             0               Ready        4m55s
```

**Scale-out Cache Runtime**

```
$ kubectl scale alluxioruntime hbase --replicas=2
alluxioruntime.data.fluid.io/hbase scaled
```
Directly use the `kubectl scale` command to complete the scale-out of Cache Runtime. After successfully executing the above command and waiting for a while, you can see that the status of both Dataset and AlluxioRuntime has changed.

A new Alluxio Worker and the corresponding Alluxio Fuse component have been successfully started:
```
$ kubectl get pod
NAME                 READY   STATUS    RESTARTS   AGE
hbase-master-0       2/2     Running   0          13m
hbase-worker-1       2/2     Running   0          6m49s
hbase-worker-0       2/2     Running   0          13m
```

The `Cache Capacity` in the Dataset changes from `2.00GiB` to `4.00GiB`, indicating an increase in the available cache capacity of the Dataset:
```
$ kubectl get dataset hbase
NAME    UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
hbase   544.77MiB        0.00B    4.00GiB          0.0%                Bound   15m
```

```
$ kubectl get alluxioruntime hbase -o wide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE   AGE
hbase   1               1                 Ready          2               2                 Ready          0             0               Ready        17m
```

Check the AlluxioRuntime description for the latest scaling information:
```
$ kubectl describe alluxioruntime hbase
...
  Conditions:
    ...
    Last Probe Time:                2021-04-23T07:54:03Z
    Last Transition Time:           2021-04-23T07:54:03Z
    Message:                        The workers are scale out.
    Reason:                         Workers scaled out
    Status:                         True
    Type:                           WorkersScaledOut
...
Events:
  Type    Reason   Age   From            Message
  ----    ------   ----  ----            -------
  Normal  Succeed  2m2s  AlluxioRuntime  Runtime scaled out. current replicas: 2, desired replicas: 2.
```

**Scale-in Cache Runtime**

Similar to scale-out, the number of workers in the Runtime can also scale-in using `kubectl scale`.
```
$ kubectl scale alluxioruntime hbase --replicas=1
alluxioruntime.data.fluid.io/hbase scaled
```

After successful execution of the above command, **if no application is trying to access the dataset currently**, then the Cache Runtime scale-in will be triggered.

Runtime Workers that exceed the specified number of `replicas` will be stopped:
```
NAME                 READY   STATUS        RESTARTS   AGE
hbase-master-0       2/2     Running       0          22m
hbase-worker-1       2/2     Terminating   0          17m32s
hbase-worker-0       2/2     Running       0          21m
```

Dataset's cache capacity (`Cache Capacity`) is returned to `2.00GiB`:
```
$ kubectl get dataset hbase
NAME    UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
hbase   544.77MiB        0.00B    2.00GiB          0.0%                Bound   30m
```

> Note: In the current version of Fluid, there is a delay of a few minutes in the change of the `Cache Capacity` property when scale-in, so you may not be able to observe the change of this property quickly.

The `Ready Workers` and `Ready Fuses` fields in AlluxioRuntime also become `1`:
```
$ kubectl get alluxioruntime hbase -o wide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE   AGE
hbase   1               1                 Ready          1               1                 Ready          0             0               Ready        30m
```

Check the AlluxioRuntime description for the latest scaling information:
```
$ kubectl describe alluxioruntime hbase
...
  Conditions:
    ...
    Last Probe Time:                2021-04-23T08:00:55Z
    Last Transition Time:           2021-04-23T08:00:55Z
    Message:                        The workers scaled in.
    Reason:                         Workers scaled in
    Status:                         True
    Type:                           WorkersScaledIn
...
Events:
  Type     Reason               Age    From            Message
  ----     ------               ----   ----            -------
  Normal   Succeed              6m56s  AlluxioRuntime  Alluxio runtime scaled out. current replicas: 2, desired replicas: 2.
  Normal   Succeed              4s     AlluxioRuntime  Alluxio runtime scaled in. current replicas: 1, desired replicas: 1.
```

The scaling capability provided by Fluid helps users or cluster administrators to adjust the resources occupied by the dataset cache in a timely manner, reducing the cache capacity of an infrequently used dataset (scale-in) or increasing the cache capacity of a dataset on demand (scale-out) to achieve a more fine-grained resource allocation and improve resource utilization.

## Clean your environment
```shell
$ kubectl delete -f dataset.yaml
```
