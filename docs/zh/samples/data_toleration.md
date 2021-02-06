# 示例 - 数据容忍污点调度

## 背景介绍

节点亲和性是 Kubernetes Pod 的一种属性，它使 Pod 被吸引到一类特定的节点。 这可能出于一种偏好，也可能是硬性要求。 Taint（污点）则相反，它使节点能够排斥一类特定的 Pod。

容忍度（Tolerations）是应用于 Pod 上的，允许（但并不要求）Pod 调度到带有与之匹配的污点的节点上。

污点和容忍度（Toleration）相互配合，可以用来避免 Pod 被分配到不合适的节点上。 每个节点上都可以应用一个或多个污点，这表示对于那些不能容忍这些污点的 Pod，是不会被该节点接受的。

而在Fluid中，考虑到`Dataset`的可调度性，资源对象中也需要定义toleration，这意味着你能够像调度你的Pod一样调度缓存在Kubernetes集群上的存放位置。
本文档将向你简单地展示上述特性


## 新建工作环境
```shell
$ mkdir <any-path>/tolerations
$ cd <any-path>/tolerations
```

## 运行示例
**查看全部结点**
```shell
$ kubectl get nodes
NAME                       STATUS   ROLES    AGE     VERSION
cn-beijing.192.168.1.146   Ready    <none>   7d14h   v1.16.9-aliyun.1
cn-beijing.192.168.1.147   Ready    <none>   7d14h   v1.16.9-aliyun.1
```

kubectl taint nodes node1 key=value:NoSchedule

**使用标签标识结点**
```shell
$ kubectl taint nodes cn-beijing.192.168.1.146 hbase=true:NoSchedule
```
在接下来的步骤中，我们将看到`NodeSelector`来管理集群中存放数据的位置，所以在这里标记期望的结点

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
  properties:
    alluxio.user.block.size.bytes.default: 256MB
    alluxio.user.streaming.reader.chunk.size.bytes: 256MB
    alluxio.user.local.reader.chunk.size.bytes: 256MB
    alluxio.worker.network.reader.buffer.size: 256MB
    alluxio.user.streaming.data.timeout: 300sec
  fuse:
    args:
      - fuse
      - --fuse-opts=kernel_cache,ro,max_read=131072,attr_timeout=7200,entry_timeout=7200,nonempty,max_readahead=0
EOF
```
该配置文件片段中，包含了许多与Alluxio相关的配置信息，这些信息将被Fluid用来启动一个Alluxio实例。上述配置片段中的`spec.replicas`属性被设置为2,这表明Fluid将会启动一个包含1个Alluxio Master和2个Alluxio Worker的Alluxio实例

**创建AlluxioRuntime资源并查看状态**
```shell
$ kubectl create -f runtime.yaml
alluxioruntime.data.fluid.io/hbase created

$ kubectl get pod -o wide
NAME                 READY   STATUS    RESTARTS   AGE    IP              NODE                       NOMINATED NODE   READINESS GATES
hbase-fuse-42csf     1/1     Running   0          104s   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
hbase-master-0       2/2     Running   0          3m3s   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
hbase-worker-l62m4   2/2     Running   0          104s   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
```
在此处可以看到，尽管我们期望看见两个AlluxioWorker被启动，但仅有一组Alluxio Worker成功启动，并且运行在具有指定标签（即`hbase-cache=true`）的结点之上。

**检查AlluxioRuntime状态**
```shell
$ kubectl get alluxioruntime hbase -o wide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE     AGE
hbase   1               1                 Ready          1               2                 PartialReady   1             2               PartialReady   4m3s
```
与预想一致，`Worker Phase`状态此时为`PartialReady`，并且`Ready Workers: 1`小于`Desired Workers: 2`

**查看待创建的应用**

我们提供了一个样例应用来演示Fluid是如何进行数据缓存亲和性调度的，首先查看该应用：

```shell
$ cat<<EOF >app.yaml
apiVersion: apps/v1beta1
kind: StatefulSet
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  replicas: 2
  serviceName: "nginx"
  podManagementPolicy: "Parallel"
  selector: # define how the deployment finds the pods it manages
    matchLabels:
      app: nginx
  template: # define the pods specifications
    metadata:
      labels:
        app: nginx
    spec:
      affinity:
        # prevent two Nginx Pod from being scheduled at the same Node
        # just for demonstrating tolerations demo
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values:
                - nginx
            topologyKey: "kubernetes.io/hostname"
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
EOF
```
其中的`podAntiAffinity`可能会让人有一点疑惑，关于这个属性的解释如下：`podAntiAffinity`属性将会确保属于相同应用的多个Pod被分散到多个不同的结点，这样的配置能够让我们更加清晰的观察到Fluid的数据缓存亲和性调度是怎么进行的。所以简单来说，这只是一个专用于演示的属性，你不必太过在意它

**运行应用**

```shell
$ kubectl create -f app.yaml
statefulset.apps/nginx created
```

**查看应用运行状态**
```shell
$ kubectl get pod -o wide -l app=nginx
NAME      READY   STATUS    RESTARTS   AGE    IP              NODE                       NOMINATED NODE   READINESS GATES
nginx-0   1/1     Running   0          2m5s   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
nginx-1   0/1     Pending   0          2m5s   <none>          <none>                     <none>           <none>
```
仅有一个Nginx Pod成功启动，并且运行在满足`nodeSelectorTerm`的结点之上

**查看应用启动失败原因**
```shell
$ kubectl describe pod nginx-1
...
Events:
  Type     Reason            Age        From               Message
  ----     ------            ----       ----               -------
  Warning  FailedScheduling  <unknown>  default-scheduler  0/2 nodes are available: 1 node(s) didn't match pod affinity/anti-affinity, 1 node(s) didn't satisfy existing pods anti-affinity rules, 1 node(s) had volume node affinity conflict.
  Warning  FailedScheduling  <unknown>  default-scheduler  0/2 nodes are available: 1 node(s) didn't match pod affinity/anti-affinity, 1 node(s) didn't satisfy existing pods anti-affinity rules, 1 node(s) had volume node affinity conflict.
```
如上所示，一方面，为了满足`PodAntiAffinity`属性的要求，使得两个Nginx Pod无法被调度到同一节点。另一方面，由于目前满足Dataset资源对象亲和性要求的结点仅有一个，因此仅有一个Nginx Pod被成功调度

**为另一个结点添加标签**
```shell
$ kubectl label node cn-beijing.192.168.1.147 hbase-cache=true
```
现在全部两个结点都具有相同的标签了，此时重新检查各个组件的运行状态
```shell
$ kubectl get pod -o wide
NAME                 READY   STATUS    RESTARTS   AGE   IP              NODE                       NOMINATED NODE   READINESS GATES
hbase-fuse-42csf     1/1     Running   0          44m   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
hbase-fuse-kth4g     1/1     Running   0          10m   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
hbase-master-0       2/2     Running   0          46m   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
hbase-worker-l62m4   2/2     Running   0          44m   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
hbase-worker-rvncl   2/2     Running   0          10m   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
```
两个Alluxio Worker都成功启动，并且分别运行在两个结点上

```shell
$ kubectl get alluxioruntime hbase -o wide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE   AGE
hbase   1               1                 Ready          2               2                 Ready          2             2               Ready        46m43s
```

```shell
$ kubectl get pod -l app=nginx -o wide
NAME      READY   STATUS    RESTARTS   AGE   IP              NODE                       NOMINATED NODE   READINESS GATES
nginx-0   1/1     Running   0          21m   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
nginx-1   1/1     Running   0          21m   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
```
另一个nginx Pod不再处于`Pending`状态，已经成功启动并运行在另一个结点上

可见，可调度的数据缓存以及对应用的数据缓存亲和性调度都是被Fluid所支持的特性。在绝大多数情况下，这两个特性协同工作，为用户提供了一种更灵活更便捷的方式在Kubernetes集群中管理数据。

可见，Fluid支持数据缓存的调度策略，这些调度策略为用户提供了更加灵活的数据缓存管理能力

## 环境清理
```shell
$ kubectl delete -f .

$ kubectl label node cn-beijing.192.168.1.146 hbase-cache-
$ kubectl label node cn-beijing.192.168.1.147 hbase-cache-
```
