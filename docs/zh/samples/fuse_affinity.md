# 示例 - Fuse客户端全局部署
在Fluid中，`Dataset`资源对象中所定义的远程文件是可被调度的，这意味着你能够像管理你的Pod一样管理远程文件缓存在Kubernetes集群上的存放位置。 而执行计算的Pod可以通过Fuse客户端访问数据文件。

Fuse客户端提供两种模式：

1. global为false，该模式为Fuse客户端和缓存数据强制亲和性，此时Fuse客户端的数量等于Runtime的replicas数量。此配置默认模式，无需显式声明，好处是可以发挥数据的亲和性优点，但是Fuse客户端的部署就变得比较固定。

2. global为true， 该模式为Fuse客户端可以在Kubernetes集群中全局部署，并不要求数据和Fuse客户端之间的强制亲和性，此时Fuse客户端的数量可能远远超Runtime的replicas数量。 建议此时可以通过nodeSelector来指定Fuse客户端的部署范围。

本文档将向你简单地展示上述特性

## 前提条件
在运行该示例之前，请参考[安装文档](../userguide/install.md)完成安装，并检查Fluid各组件正常运行：
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
$ mkdir <any-path>/fuse-global-deployment
$ cd <any-path>/fuse-global-deployment
```

## 运行示例1: 设置global为true
**查看全部结点**
```shell
$ kubectl get nodes
NAME                       STATUS   ROLES    AGE     VERSION
cn-beijing.192.168.1.146   Ready    <none>   7d14h   v1.16.9-aliyun.1
cn-beijing.192.168.1.147   Ready    <none>   7d14h   v1.16.9-aliyun.1
```

**使用标签标识结点**
```shell
$ kubectl label nodes cn-beijing.192.168.1.146 cache-node=true
```
在接下来的步骤中，我们将使用`NodeSelector`来管理集群中存放数据的位置，所以在这里标记期望的结点

**再次查看结点**
```shell
$ kubectl get node -L cache-node
NAME                       STATUS   ROLES    AGE     VERSION            cache-node
cn-beijing.192.168.1.146   Ready    <none>   7d14h   v1.16.9-aliyun.1   true
cn-beijing.192.168.1.147   Ready    <none>   7d14h   v1.16.9-aliyun.1   
```
目前，在全部2个结点中，仅有一个结点添加了`cache-node=true`的标签，接下来，我们希望数据缓存仅会被放置在该结点之上

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
            - key: cache-node
              operator: In
              values:
                - "true"
EOF
```
在该`Dataset`资源对象的`spec`属性中，我们定义了一个`nodeSelectorTerm`的子属性，该子属性要求数据缓存必须被放置在具有`cache-node=true`标签的结点之上

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
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 2Gi
        high: "0.95"
        low: "0.7"
  fuse:
    global: true
EOF
```

该配置文件片段中，包含了许多与Alluxio相关的配置信息，这些信息将被Fluid用来启动一个Alluxio实例。上述配置片段中的`spec.replicas`属性被设置为1,这表明Fluid将会启动一个包含1个Alluxio Master和1个Alluxio Worker的Alluxio实例。 另外一个值得注意的是Fuse包含`global: true`,  
这样意味着Fuse可以全局部署，而不依赖于数据缓存的位置。

**创建AlluxioRuntime资源并查看状态**
```shell
$ kubectl create -f runtime.yaml
alluxioruntime.data.fluid.io/hbase created

$ kubectl get po -owide
NAME                 READY   STATUS    RESTARTS   AGE     IP              NODE                       NOMINATED NODE   READINESS GATES
hbase-fuse-gfq7z     1/1     Running   0          3m47s   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
hbase-fuse-lmk5p     1/1     Running   0          3m47s   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
hbase-master-0       2/2     Running   0          3m47s   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
hbase-worker-hvbp2   2/2     Running   0          3m1s    192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
```
在此处可以看到，有一个Alluxio Worker成功启动，并且运行在具有指定标签（即`cache-node=true`）的结点之上。Alluixo Fuse的数量为2，运行在所有的子节点上。

**检查AlluxioRuntime状态**
```shell
$ kubectl get alluxioruntime hbase -o wide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE   AGE
hbase   1               1                 Ready          1               1                 Ready          2             2               Ready        12m
```

这里可以看到Alluxio Worker的数量为1，而Alluxio Fuse的数量为2。

**删除AlluxioRuntime**

```shell
kubectl delete alluxioruntime hbase
```

## 运行示例2: 设置global为true, 并且设置fuse的nodeSelector

下面，我们希望通过配置node selector配置Fuse客户端，将其指定到集群中某个节点上。在本例子中，既然我们已经选择节点cn-beijing.192.168.1.146作为缓存节点，为了形成对比，这里选择节点cn-beijing.192.168.1.147运行Alluxio Fuse。

```shell
$ cat<<EOF >runtime-node-selector.yaml
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
    global: true
    nodeSelector:
      kubernetes.io/hostname: cn-beijing.192.168.1.147
EOF
```

该配置文件片段中，和之前runtime.yaml相比，在Fuse包含`global: true`的前提下,  还增加了nodeSelector并且指向了节点cn-beijing.192.168.1.147。

**创建AlluxioRuntime资源并查看状态**
```shell
$ kubectl create -f runtime-node-selector.yaml
alluxioruntime.data.fluid.io/hbase created

$ kubectl get po -owide
NAME                 READY   STATUS    RESTARTS   AGE   IP              NODE                       NOMINATED NODE   READINESS GATES
hbase-fuse-xzbww     1/1     Running   0          1h   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
hbase-master-0       2/2     Running   0          1h   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
hbase-worker-vdxd5   2/2     Running   0          1h   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
```
在此处可以看到，有一个Alluxio Worker成功启动，并且运行在具有指定标签（即`cache-node=true`）的结点之上。Alluixo Fuse的数量为1，运行在节点cn-beijing.192.168.1.147上。

**检查AlluxioRuntime状态**
```shell
$ kubectl get alluxioruntimes.data.fluid.io -owide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE   AGE
hbase   1               1                 Ready          1               1                 Ready          1             1               Ready        1h
```

这里可以看到Alluxio Worker的数量为1，而Alluxio Fuse的数量也为1，这是因为AlluxioRuntime指定了nodeSelector，并且满足条件的节点只有一个。


可见，Fluid支持Fuse客户端的单独的调度策略，这些调度策略为用户提供了更加灵活的Fuse客户端调度策略

## 环境清理
```shell
$ kubectl delete -f .

$ kubectl label node cn-beijing.192.168.1.146 cache-node-
```
