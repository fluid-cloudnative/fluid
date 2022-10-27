# Demo - Set FUSE clean policy

FUSE clean policy is set under the property `spec.fuse.cleanPolicy` of `Runtime`. There are 2 clean policies for FUSE Pod. `OnRuntimeDeleted` means that the FUSE Pod is cleaned only when the cache runtime is deleted, and `OnDemand` means that when the FUSE Pod is not used by any application Pod, the FUSE Pod will be cleaned.
By default, the cleanup policy for FUSE Pod is `OnRuntimeDeleted` 

## Prerequisites

Before running this example, please refer to the [Installation Documentation](https://github.com/fluid-cloudnative/fluid/blob/master/docs/en/userguide/install.md) to complete the installation, and check that each Fluid component is normal run:

```shell
$ kubectl get pod -n fluid-system
NAME                                        READY   STATUS    RESTARTS        AGE
alluxioruntime-controller-86ddc878d-pc6g5   1/1     Running   7 (2m19s ago)   24h
csi-nodeplugin-fluid-ccbk8                  2/2     Running   6 (2m19s ago)   24h
dataset-controller-67bcb77b89-6xw7p         1/1     Running   4 (2m16s ago)   24h
fluid-webhook-648ccc89c6-bq5rd              1/1     Running   4 (2m18s ago)   24h
```

Typically, you will see a pod named `dataset-controller`, a pod named `alluxioruntime-controller`, and several pods named `csi-nodeplugin` running. Among them, the number of `csi-nodeplugin` these pods depends on the number of nodes in your Kubernetes cluster.

## Demo

**Check `DataSet` and `AlluxioRuntime` Objects to be Created**
```shell
$ cat dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: hbase
spec:
  mounts:
    - mountPoint: https://mirrors.bit.edu.cn/apache/hbase/stable/
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
  fuse:
    cleanPolicy: OnDemand
```
We set FUSE's cleanup policy to `OnDemand`. When the FUSE Pod is not used by any application Pod, the FUSE Pod will be cleaned.

**Create `DataSet` and `AlluxioRuntime` Objects**
```shell
$ kubectl craete -f dataset.yaml
dataset.data.fluid.io/hbase created
alluxioruntime.data.fluid.io/hbase created
$ kubectl get pods
NAME             READY   STATUS    RESTARTS   AGE
hbase-master-0   2/2     Running   0          74s
hbase-worker-0   2/2     Running   0          45s
```

**Create Application Pod**
```shell
$ cat nginx.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
    - name: nginx
      image: nginx
      volumeMounts:
        - mountPath: /data
          name: hbase-vol
  volumes:
    - name: hbase-vol
      persistentVolumeClaim:
        claimName: hbase
$ kubectl create -f nginx.yaml
pod/nginx created
```

**Check Pods**
```shell
$ kubectl get pods
NAME               READY   STATUS    RESTARTS   AGE
hbase-fuse-889ts   1/1     Running   0          29s
hbase-master-0     2/2     Running   0          4m27s
hbase-worker-0     2/2     Running   0          3m58s
nginx              1/1     Running   0          30s
```
After creating the application Pod, we find that the FUSE Pod is successfully created.

**Delete Application Pod**
```shell
$ kubectl delete -f nginx.yaml
pod "nginx" deleted
```

**Check Pods Again**    
```shell
$ kubectl get pods
NAME             READY   STATUS    RESTARTS   AGE
hbase-master-0   2/2     Running   0          6m57s
hbase-worker-0   2/2     Running   0          6m28s
```
Note that the FUSE clean policy is set to `OnDemand`. After the application Pod is deleted, the FUSE Pod is no longer used by it, so the FUSE Pod is cleaned.