# 示例 - Fuse NodeSelector使用

在Fluid中，`Dataset`资源对象中所定义的远程文件是可被调度的，这意味着你能够像管理你的Pod一样管理远程文件缓存在Kubernetes集群上的存放位置。 而执行计算的Pod可以通过Fuse客户端访问数据文件。Fuse客户端可以在Kubernetes集群中全局部署,我们可以通过`nodeSelector`来限制Fuse客户端的部署范围。

本文档将向你简单地展示上述特性。在此之前，你可以访问[toc.md](https://github.com/fluid-cloudnative/fluid/blob/master/docs/zh/TOC.md)来查看Fluid文档来了解相关知识细节。

## 前提条件

在运行该示例之前，请参考[安装文档](https://github.com/fluid-cloudnative/fluid/blob/master/docs/zh/userguide/install.md)完成安装，并检查Fluid各组件正常运行：

```
$ kubectl get pod -n fluid-system
alluxioruntime-controller-f87f54fd6-pqj77   1/1     Running   0          14h
csi-nodeplugin-fluid-4pfmk                  2/2     Running   0          13h
csi-nodeplugin-fluid-qwlm4                  2/2     Running   0          13h
dataset-controller-bb95bb754-dsxww          1/1     Running   0          14h
fluid-webhook-66b77ccb8f-vvqjb              1/1     Running   0          13h
```

通常来说，你会看到一个名为`dataset-controller`的Pod、一个名为`alluxioruntime-controller`的Pod和多个名为`csi-nodeplugin`的Pod正在运行。其中，`csi-nodeplugin`这些Pod的数量取决于你的Kubernetes集群中结点的数量。



## 新建工作环境

```
$ mkdir <any-path>/fuse-nodeselector-use
$ cd <any-path>/fuse-nodeselector-use
```



## 运行示例

**查看全部结点**

```
$ kubectl get nodes
NAME                      STATUS   ROLES    AGE   VERSION
cn-beijing.172.16.0.101   Ready    <none>   13h   v1.20.11-aliyun.1
cn-beijing.172.16.0.99    Ready    <none>   23h   v1.20.11-aliyun.1
```

**使用标签标识结点**

```
$ kubectl label nodes cn-beijing.172.16.0.101 select-node=true
node/cn-beijing.172.16.0.101 labeled
```

在接下来的步骤中，我们将使用`NodeSelector`来管理Fuse客户端的部署范围，所以在这里标记期望的结点

**再次查看结点**

```
$ kubectl get node -L select-node
NAME                      STATUS   ROLES    AGE   VERSION             SELECT-NODE
cn-beijing.172.16.0.101   Ready    <none>   13h   v1.20.11-aliyun.1   true
cn-beijing.172.16.0.99    Ready    <none>   23h   v1.20.11-aliyun.1 
```

目前，在全部2个结点中，仅有一个结点添加了`select-node=true`的标签，接下来，我们希望Fuse客户端仅会被部署在该结点之上。

**检查待创建的Dataset资源对象**

```
$ cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: hbase
spec:
  mounts:
    - mountPoint: https://mirrors.tuna.tsinghua.edu.cn/apache/hbase/stable/
      name: hbase
EOF
```

**创建Dataset资源对象**

```
$ kubectl create -f dataset.yaml
dataset.data.fluid.io/hbase created
```

**检查待创建的AlluxioRuntime资源对象**

```
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
    nodeSelector:
      select-node: "true"
EOF
```

该配置文件片段中，通过 `spec.fuse.nodeSelector` 将AlluxioRuntime设置到刚刚被标注`select-node=true`的节点上

**创建AlluxioRuntime资源并查看状态**

```
$ kubectl create -f runtime.yaml
alluxioruntime.data.fluid.io/hbase created

$ kubectl get po -owide
NAME                    READY   STATUS    RESTARTS   AGE   IP             NODE                      NOMINATED NODE   READINESS GATES
hbase-master-0          2/2     Running   0          64m   172.16.0.101   cn-beijing.172.16.0.101   <none>           <none>
hbase-worker-0          2/2     Running   0          64m   172.16.0.101   cn-beijing.172.16.0.101   <none>           <none>
```

在此处可以看到，有一个Alluxio Worker成功启动，并且运行在具有指定标签（即`select-node=true`）的结点之上。

**检查AlluxioRuntime状态**

```
$ kubectl get alluxioruntime hbase -o wide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE   AGE
hbase   1               1                 Ready          1               1                 Ready          0             0               Ready        12m
```

这里可以看到Alluxio Worker的数量为1，而Alluxio Fuse的数量为0，因为Alluxio Fuse需要依靠一个作业应用来启动，下边将会展示通过pod启动Fuse。

**查看待创建的应用**

下面是一个简单的Deployment 的 YAML 代码段，它有两个个副本和选择器标签 `app=nginx-test`。 Deployment 配置了 `PodAntiAffinity`，用来确保调度器不会将所有副本调度到同一节点上。

```
$ cat<<EOF >nginx.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  selector:
    matchLabels:
      app: nginx-test
  replicas: 2
  template:
    metadata:
      labels:
        app: nginx-test
    spec:
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values:
                - nginx-test
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

**启动应用**

```
$ kubectl create -f nginx.yaml
deployment.apps/nginx created
```

**查看pod状态**

```
$ kubectl get po -owide
NAME                    READY   STATUS    RESTARTS   AGE   IP             NODE                      NOMINATED NODE   READINESS GATES
hbase-fuse-4jfgl        1/1     Running   0          41m   172.16.0.101   cn-beijing.172.16.0.101   <none>           <none>
hbase-master-0          2/2     Running   0          64m   172.16.0.101   cn-beijing.172.16.0.101   <none>           <none>
hbase-worker-0          2/2     Running   0          64m   172.16.0.101   cn-beijing.172.16.0.101   <none>           <none>
nginx-766564fc7-8vz4s   0/1     Pending   0          41m   <none>         <none>                    <none>           <none>
nginx-766564fc7-rtmwh   1/1     Running   0          41m   10.73.0.135    cn-beijing.172.16.0.101   <none>           <none>
```

可以看到Fuse客户端只在有`select-node=true`标签的节点`cn-beijing.172.16.0.101`上启动了。

两个pod的其中一个（`nginx-766564fc7-rtmwh`）在节点`cn-beijing.172.16.0.101`上启动了，因为pod间具有 `PodAntiAffinity`，另一个pod（`nginx-766564fc7-8vz4s`）应当在另一个节点`cn-beijing.172.16.0.99`上启动，但是因为`spec.fuse.nodeSelector`的限制，他不能被调度而处于pending状态。

**启动所有pod**

为了证明pod不被调度是因为`spec.fuse.nodeSelector`的限制，我们将节点`cn-beijing.172.16.0.99`也标注`select-node=true`，可以发现该节点的Fuse Pod和Nginx Pod都启动了。

```
$ kubectl label nodes cn-beijing.172.16.0.99 select-node=true
$ kubectl get po -owide
NAME                    READY   STATUS    RESTARTS   AGE    IP             NODE                      NOMINATED NODE   READINESS GATES
hbase-fuse-4jfgl        1/1     Running   0          138m   172.16.0.101   cn-beijing.172.16.0.101   <none>           <none>
hbase-fuse-k95g5        1/1     Running   0          28s    172.16.0.99    cn-beijing.172.16.0.99    <none>           <none>
hbase-master-0          2/2     Running   0          161m   172.16.0.101   cn-beijing.172.16.0.101   <none>           <none>
hbase-worker-0          2/2     Running   0          161m   172.16.0.101   cn-beijing.172.16.0.101   <none>           <none>
nginx-766564fc7-8vz4s   1/1     Running   0          138m   10.73.0.22     cn-beijing.172.16.0.99    <none>           <none>
nginx-766564fc7-rtmwh   1/1     Running   0          138m   10.73.0.135    cn-beijing.172.16.0.101   <none>           <none>
```

可以发现节点`cn-beijing.172.16.0.99`上的Fuse也启动了，同时Nginx Pod（`nginx-766564fc7-8vz4s`）也可以在该节点上进行调度。

## 环境清理

```
$ kubectl delete -f .

$ kubectl label node cn-beijing.172.16.0.101 select-node-
$ kubectl label node cn-beijing.172.16.0.99 select-node-
```