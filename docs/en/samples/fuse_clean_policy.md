# Demo - Set FUSE clean policy

FUSE clean policy is set under the property `spec.fuse.cleanPolicy` of `Runtime`. There are 2 clean policies for FUSE Pod. `OnRuntimeDeleted` means that the FUSE Pod is cleaned only when the cache runtime is deleted, and `OnDemand` means that when the FUSE Pod is not used by any application Pod, the FUSE Pod will be cleaned.
By default, the cleanup policy for FUSE Pod is `OnRuntimeDeleted` 

## Prerequisites

Before running this example, please refer to the [Installation Documentation](https://github.com/fluid-cloudnative/fluid/blob/master/docs/en/userguide/install.md) to complete the installation, and check that each Fluid component is normal run:

```shell
$ kubectl get pod -n fluid-system
NAME                                   READY   STATUS    RESTARTS   AGE
csi-nodeplugin-fluid-5w7gk             2/2     Running   0          4m50s
csi-nodeplugin-fluid-h6wm7             2/2     Running   0          4m50s
csi-nodeplugin-fluid-nlkc4             2/2     Running   0          4m50s
dataset-controller-74554dfc4f-gwxmb    1/1     Running   0          4m50s
fluid-webhook-5c77b8b4f9-xgpv8         1/1     Running   0          4m50s
fluidapp-controller-7bb7bdb5d7-k7hdc   1/1     Running   0          4m50s
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
  replicas: 2
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
hbase-master-0   2/2     Running   0          2m32s
hbase-worker-0   2/2     Running   0          2m3s
hbase-worker-1   2/2     Running   0          2m2s
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
hbase-fuse-4ncs2   1/1     Running   0          85s
hbase-master-0     2/2     Running   0          4m31s
hbase-worker-0     2/2     Running   0          4m2s
hbase-worker-1     2/2     Running   0          4m1s
nginx              1/1     Running   0          85s
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
hbase-master-0   2/2     Running   0          5m
hbase-worker-0   2/2     Running   0          4m31s
hbase-worker-1   2/2     Running   0          4m30s
```
Note that the FUSE clean policy is set to `OnDemand`. After the application Pod is deleted, the FUSE Pod is no longer used by it, so the FUSE Pod is cleaned.

**Set `cleanPolicy` to  `OnRuntimeDeleted`**

Set the property `cleanPolicy` of `AlluxioRuntime` to `OnRuntimeDeleted`
```shell
$ cat dataset.yaml
...
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: hbase
spec:
  replicas: 2
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 2Gi
        high: "0.95"
        low: "0.7"
  fuse:
    cleanPolicy: OnRuntimeDeleted
$ kubectl apply -f dataset.yaml
```

**Create Application Pod Again**
```shell
$ kubectl create -f nginx.yaml
pod/nginx created
$ kubectl get pod
NAME               READY   STATUS    RESTARTS   AGE
hbase-fuse-bl9w6   1/1     Running   0          7s
hbase-master-0     2/2     Running   0          12m
hbase-worker-0     2/2     Running   0          11m
hbase-worker-1     2/2     Running   0          11m
nginx              1/1     Running   0          7s
```

**Delete Application Pod**
```shell
$ kubectl delete -f nginx.yaml
pod "nginx" deleted
$ kubectl get pod
NAME               READY   STATUS    RESTARTS   AGE
hbase-fuse-bl9w6   1/1     Running   0          92s
hbase-master-0     2/2     Running   0          13m
hbase-worker-0     2/2     Running   0          13m
hbase-worker-1     2/2     Running   0          12m
```
Since `cleanPolicy` was set to `OnRuntimeDeleted`, after deleting the application pod, we found that the FUSE pod was not cleaned.

**Delete `AlluxioRuntime`**
```shell
$ kubectl delete alluxioruntime hbase
alluxioruntime.data.fluid.io "hbase" deleted
$ kubectl get pod
No resources found in default namespace.
$ kubectl get alluxioruntime
No resources found in default namespace.
```
FUSE Pods are also cleaned after removing `AlluxioRuntime`.

**Clean up the Environment**
```shell
$ kubectl delete dataset hbase
dataset.data.fluid.io "hbase" deleted
```