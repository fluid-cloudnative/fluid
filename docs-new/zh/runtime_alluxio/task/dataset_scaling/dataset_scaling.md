# 示例 - Dataset手动扩缩容

## 前提条件

- [Fluid](https://github.com/fluid-cloudnative/fluid)(version >= 0.5.0)

请参考[Fluid安装文档](../guide/install.md)完成安

## 新建工作环境
```shell
$ mkdir <any-path>/dataset_scale
$ cd <any-path>/dataset_scale
```

## 运行示例

**创建Dataset和AlluxioRuntime资源对象**
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

在上述示例中，我们设置`AlluxioRuntime.spec.replicas`为1，这意味着我们将启动一个带有一个Worker的Alluxio集群来缓存数据集中的数据。

```
$ kubectl create -f dataset.yaml
dataset.data.fluid.io/hbase created
alluxioruntime.data.fluid.io/hbase created
```
待Alluxio集群正常启动后，可以看到此时创建出来的Dataset以及AlluxioRuntime处于如下状态：

Alluxio各组件运行状态：
```
$ kubectl get pod
NAME                 READY   STATUS    RESTARTS   AGE
hbase-master-0       2/2     Running   0          3m50s
hbase-worker-0       2/2     Running   0          3m15s
```

Dataset状态：
```
$ kubectl get dataset hbase
NAME    UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
hbase   544.77MiB        0.00B    2.00GiB          0.0%                Bound   3m28s
```

AlluxioRuntime状态：
```
$ kubectl get alluxioruntime hbase -o wide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE   AGE
hbase   1               1                 Ready          1               1                 Ready          0             0               Ready        4m55s
```

**Dataset扩容**

```
$ kubectl scale alluxioruntime hbase --replicas=2
alluxioruntime.data.fluid.io/hbase scaled
```
直接使用`kubectl scale`命令即可完成Dataset的扩容操作。在成功执行上述命令并等待一段时间后可以看到Dataset以及AlluxioRuntime的状态均发生了变化：

一个新的Alluxio Worker以及对应的Alluxio Fuse组件成功启动：
```
$ kubectl get pod
NAME                 READY   STATUS    RESTARTS   AGE
hbase-master-0       2/2     Running   0          13m
hbase-worker-0       2/2     Running   0          13m
hbase-worker-1       2/2     Running   0          6m49s
```

Dataset中的`Cache Capacity`从原来的`2.00GiB`变为`4.00GiB`，表明该Dataset的可用缓存容量增加：
```
$ kubectl get dataset hbase
NAME    UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
hbase   544.77MiB        0.00B    4.00GiB          0.0%                Bound   15m
```

AlluxioRuntime中的`Ready Workers`以及`Ready Fuses`属性均变为2：
```
$ kubectl get alluxioruntime hbase -o wide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE   AGE
hbase   1               1                 Ready          2               2                 Ready          0             0               Ready        17m
```

查看AlluxioRuntime的具体描述信息可以了解最新的扩缩容信息：
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

**Dataset缩容**

与扩容类似，缩容时同样可以使用`kubectl scale`对Runtime的Worker数量进行调整：
```
$ kubectl scale alluxioruntime hbase --replicas=1
alluxioruntime.data.fluid.io/hbase scaled
```

成功执行上述命令后，**如果目前环境中没有应用正在尝试访问该数据集**，那么就会触发Runtime的缩容。

超出指定`replicas`数量的Runtime Worker将会被停止：
```
NAME                 READY   STATUS        RESTARTS   AGE
hbase-master-0       2/2     Running       0          22m
hbase-worker-0       2/2     Running       0          21m
hbase-worker-1       2/2     Terminating   0          17m32s
```

Dataset的缓存容量(`Cache Capacity`)恢复到`2.00GiB`:
```
$ kubectl get dataset hbase
NAME    UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
hbase   544.77MiB        0.00B    2.00GiB          0.0%                Bound   30m
```

> 注意：在目前版本的Fluid中，缩容时Dataset中`Cache Capacity`属性字段的变化存在几分钟的延迟，因此您可能无法迅速观察到这一属性的变化

AlluxioRuntime中的`Ready Workers`以及`Ready Fuses`字段同样变为`1`：
```
$ kubectl get alluxioruntime hbase -o wide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE   AGE
hbase   1               1                 Ready          1               1                 Ready          0             0               Ready        30m
```

查看AlluxioRuntime的具体描述信息可以了解最新的扩缩容信息：
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
  Normal   Succeed              6m56s  AlluxioRuntime  Runtime scaled out. current replicas: 2, desired replicas: 1.
  Normal   Succeed              4s     AlluxioRuntime  Runtime scaled in. current replicas: 1, desired replicas: 2.
```

Fluid提供的这种扩缩容能力能够帮助用户或是集群管理员适时地调整数据集缓存所占用的集群资源，减少某个不频繁使用的数据集的缓存容量（缩容），或者按需增加某数据集的缓存容量（扩容），以实现更加精细的资源分配，提高资源利用率。

## 环境清理
```shell
$ kubectl delete -f dataset.yaml
```
