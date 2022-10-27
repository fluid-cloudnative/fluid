# 示例 - 设置 FUSE 清理策略

FUSE清理策略在`Runtime`的`spec.fuse.cleanPolicy`下设置。FUSE Pod的清理策略有两种：`OnRuntimeDeleted`表示FUSE Pod仅在Runtime被删除时回收，`OnDemand`表示当FUSE Pod不被任何应用Pod使用时，FUSE Pod会被回收。
默认情况下，FUSE Pod的清理策略为`OnRuntimeDeleted`

## 前提条件

在运行该示例之前，请参考[安装文档](https://github.com/fluid-cloudnative/fluid/blob/master/docs/zh/userguide/install.md) 完成安装，并检查Fluid各组件正常运行：

```shell
$ kubectl get pod -n fluid-system
NAME                                        READY   STATUS    RESTARTS        AGE
alluxioruntime-controller-86ddc878d-pc6g5   1/1     Running   7 (2m19s ago)   24h
csi-nodeplugin-fluid-ccbk8                  2/2     Running   6 (2m19s ago)   24h
dataset-controller-67bcb77b89-6xw7p         1/1     Running   4 (2m16s ago)   24h
fluid-webhook-648ccc89c6-bq5rd              1/1     Running   4 (2m18s ago)   24h     13h
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
我们将FUSE的清理策略设置为`OnDemand`。当FUSE Pod不被任何应用Pod使用时，FUSE Pod会被回收。

**创建资源DataSet和AlluxioRuntime资源对象**
```shell
$ kubectl craete -f dataset.yaml
dataset.data.fluid.io/hbase created
alluxioruntime.data.fluid.io/hbase created
$ kubectl get pods
NAME             READY   STATUS    RESTARTS   AGE
hbase-master-0   2/2     Running   0          74s
hbase-worker-0   2/2     Running   0          45s
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
hbase-fuse-889ts   1/1     Running   0          29s
hbase-master-0     2/2     Running   0          4m27s
hbase-worker-0     2/2     Running   0          3m58s
nginx              1/1     Running   0          30s
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
hbase-master-0   2/2     Running   0          6m57s
hbase-worker-0   2/2     Running   0          6m28s
```
由于清理策略设置的是`OnDemand`，删除应用Pod后，FUSE Pod不再被其使用，故FUSE Pod被清理。