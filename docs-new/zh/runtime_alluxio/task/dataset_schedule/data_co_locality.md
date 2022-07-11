# 示例 - 数据缓存亲和性调度
在Fluid中，`Dataset`资源对象中所定义的远程文件是可被调度的，这意味着你能够像管理你的Pod一样管理远程文件缓存在Kubernetes集群上的存放位置。

本文档将向你简单地展示上述特性

## 前提条件
在运行该示例之前，请参考[安装文档](../guide/install.md)完成安装，并检查Fluid各组件正常运行：
```shell
$ kubectl get pod -n fluid-system
alluxioruntime-controller-5b64fdbbb-84pc6   1/1     Running   0          8h
csi-nodeplugin-fluid-fwgjh                  2/2     Running   0          8h
csi-nodeplugin-fluid-ll8bq                  2/2     Running   0          8h
dataset-controller-5b7848dbbb-n44dj         1/1     Running   0          8h
```

通常来说，你会看到一个名为`dataset-controller`的Pod、一个名为`alluxioruntime-controller`的Pod和多个名为`csi-nodeplugin`的Pod正在运行。其中，`csi-nodeplugin`这些Pod的数量取决于你的Kubernetes集群中结点的数量。

## 新建工作环境
```shell
$ mkdir <any-path>/co-locality
$ cd <any-path>/co-locality
```

## 运行示例
**查看全部结点**
```shell
$ kubectl get nodes
NAME                       STATUS   ROLES    AGE     VERSION
cn-beijing.192.168.1.146   Ready    <none>   7d14h   v1.16.9-aliyun.1
cn-beijing.192.168.1.147   Ready    <none>   7d14h   v1.16.9-aliyun.1
```

**使用标签标识结点**
```shell
$ kubectl label nodes cn-beijing.192.168.1.146 hbase-cache=true
```
在接下来的步骤中，我们将使用`NodeSelector`来管理集群中存放数据的位置，所以在这里标记期望的结点

**再次查看结点**
```shell
$ kubectl get node -L hbase-cache
NAME                       STATUS   ROLES    AGE     VERSION            HBASE-CACHE
cn-beijing.192.168.1.146   Ready    <none>   7d14h   v1.16.9-aliyun.1   true
cn-beijing.192.168.1.147   Ready    <none>   7d14h   v1.16.9-aliyun.1   
```
目前，在全部2个结点中，仅有一个结点添加了`hbase-cache=true`的标签，接下来，我们希望数据缓存仅会被放置在该结点之上

**检查待创建的Dataset资源对象**
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
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: hbase-cache
              operator: In
              values:
                - "true"
EOF
```
在该`Dataset`资源对象的`spec`属性中，我们定义了一个`nodeSelectorTerm`的子属性，该子属性要求数据缓存必须被放置在具有`hbase-cache=true`标签的结点之上

**创建Dataset资源对象**
```shell
$ kubectl create -f dataset.yaml
dataset.data.fluid.io/hbase created
```

**检查待创建的AlluxioRuntime资源对象**
```shell
$ cat<<EOF >runtime.yaml
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
EOF
```
该配置文件片段中，包含了许多与Alluxio相关的配置信息，这些信息将被Fluid用来启动一个Alluxio实例。上述配置片段中的`spec.replicas`属性被设置为2,这表明Fluid将会启动一个包含1个Alluxio Master和2个Alluxio Worker的Alluxio实例

**创建AlluxioRuntime资源并查看状态**
```shell
$ kubectl create -f runtime.yaml
alluxioruntime.data.fluid.io/hbase created

$ kubectl get pod -o wide
NAME                 READY   STATUS    RESTARTS   AGE    IP              NODE                       NOMINATED NODE   READINESS GATES
hbase-master-0       2/2     Running   0          3m3s   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
hbase-worker-0       2/2     Running   0          104s   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
hbase-worker-1       0/2     Pending   0          104s   <none>          <none>                     <none>           <none>
```
在此处可以看到，尽管我们期望看见两个AlluxioWorker被启动，但仅有一个Alluxio Worker Pod成功被调度，并且运行在具有指定标签（即`hbase-cache=true`）的结点之上。

**检查AlluxioRuntime状态**
```shell
$ kubectl get alluxioruntime hbase -o wide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE     AGE
hbase   1               1                 Ready          1               2                 PartialReady   0             0               Ready          4m3s
```
与预想一致，`Worker Phase`状态此时为`PartialReady`，并且`Ready Workers: 1`小于`Desired Workers: 2`

**为另一个结点添加标签**
```shell
$ kubectl label node cn-beijing.192.168.1.147 hbase-cache=true
```
现在全部两个结点都具有相同的标签了，此时重新检查各个组件的运行状态
```shell
$ kubectl get pod -o wide
NAME                 READY   STATUS    RESTARTS   AGE   IP              NODE                       NOMINATED NODE   READINESS GATES
hbase-master-0       2/2     Running   0          46m   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
hbase-worker-0       2/2     Running   0          44m   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
hbase-worker-1       2/2     Running   0          44m   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
```
两个Alluxio Worker都成功启动，并且分别运行在两个结点上

```shell
$ kubectl get alluxioruntime hbase -o wide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE   AGE
hbase   1               1                 Ready          2               2                 Ready          0             0               Ready        46m43s
```

可见，Fluid将数据缓存视作Kubernetes可调度的资源，支持数据缓存的调度策略,这些调度策略为用户提供了更加灵活的数据缓存管理能力。

## 环境清理
```shell
$ kubectl delete -f .

$ kubectl label node cn-beijing.192.168.1.146 hbase-cache-
$ kubectl label node cn-beijing.192.168.1.147 hbase-cache-
```
