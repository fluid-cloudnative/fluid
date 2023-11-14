# 示例 - 设置 FUSE 清理策略

FUSE清理策略在`Runtime`的`spec.fuse.cleanPolicy`下设置。FUSE Pod的清理策略有两种：`OnRuntimeDeleted`表示FUSE Pod仅在Runtime被删除时回收，`OnDemand`表示当FUSE Pod不被任何应用Pod使用时，FUSE Pod会被回收。
默认情况下，FUSE Pod的清理策略为`OnRuntimeDeleted`

## 前提条件

在运行该示例之前，请参考[安装文档](https://github.com/fluid-cloudnative/fluid/blob/master/docs/zh/userguide/install.md) 完成安装，并检查Fluid各组件正常运行：

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

通常来说，你会看到一个名为`dataset-controller`的Pod、一个名为`alluxioruntime-controller`的Pod和多个名为`csi-nodeplugin`的Pod正在运行。其中，`csi-nodeplugin`这些Pod的数量取决于你的Kubernetes集群中结点的数量。


## 示例

**查看DataSet和AlluxioRuntime资源对象**
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
我们将FUSE的清理策略设置为`OnDemand`。当FUSE Pod不被任何应用Pod使用时，FUSE Pod会被回收。

**创建资源DataSet和AlluxioRuntime资源对象**
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

**创建应用Pod**
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

**查看Pod**
```shell
$ kubectl get pods
NAME               READY   STATUS    RESTARTS   AGE
hbase-fuse-4ncs2   1/1     Running   0          85s
hbase-master-0     2/2     Running   0          4m31s
hbase-worker-0     2/2     Running   0          4m2s
hbase-worker-1     2/2     Running   0          4m1s
nginx              1/1     Running   0          85s
```
创建应用Pod后，我们发现FUSE Pod被成功创建。

**删除应用Pod**
```shell
$ kubectl delete -f nginx.yaml
pod "nginx" deleted
```

**再次查看Pod**
```shell
$ kubectl get pods
NAME             READY   STATUS    RESTARTS   AGE
hbase-master-0   2/2     Running   0          5m
hbase-worker-0   2/2     Running   0          4m31s
hbase-worker-1   2/2     Running   0          4m30s
s
```
由于清理策略设置的是`OnDemand`，删除应用Pod后，FUSE Pod不再被其使用，故FUSE Pod被清理。

**将cleanPolicy修改为OnDemandDeleted**

将`AlluxioRuntime`的cleanPolicy修改为`OnDemandDeleted`
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
    cleanPolicy: OnDemandDeleted
$ kubectl apply -f dataset.yaml
```

**再次创建应用Pod**
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

**删除应用Pod**
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
由于`cleanPolicy`被设置为`OnDemandDeleted`，删除应用Pod后，我们发现FUSE Pod并没有被清理。

**删除AlluxioRuntime**
```shell
$ kubectl delete alluxioruntime hbase
alluxioruntime.data.fluid.io "hbase" deleted
$ kubectl get pod
No resources found in default namespace.
$ kubectl get alluxioruntime
No resources found in default namespace.
```
删除`AlluxioRuntime`后，FUSE Pod也被清理

**清理环境**
```shell
$ kubectl delete dataset hbase
dataset.data.fluid.io "hbase" deleted
```