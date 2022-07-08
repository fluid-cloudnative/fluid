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
**查看全部节点**
```shell
$ kubectl get no
NAME                       STATUS   ROLES    AGE    VERSION
cn-beijing.192.168.1.146   Ready    <none>   200d   v1.16.9-aliyun.1
```

kubectl taint nodes node1 key=value:NoSchedule

**为节点配置污点（taint）**

```shell
$ kubectl taint nodes cn-beijing.192.168.1.146 hbase=true:NoSchedule
```
在接下来的步骤中，我们将看到节点上的污点配置

**再次查看节点**

```shell
$ kubectl get node cn-beijing.192.168.1.146 -oyaml | grep taints -A3
  taints:
  - effect: NoSchedule
    key: hbase
    value: "true"
```

目前，节点增加了taints配置NoSchedule，这样默认数据集就无法放置到该节点上

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
  tolerations:
    - key: hbase 
      operator: Equal 
      value: "true" 
EOF
```
在该`Dataset`资源对象的`spec`属性中，我们定义了一个`tolerations`的属性，该子属性要求数据缓存可以放置到配置污点的节点

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
EOF
```
该配置文件片段中，包含了许多与Alluxio相关的配置信息，这些信息将被Fluid用来启动一个Alluxio实例。上述配置片段中的`spec.replicas`属性被设置为1,这表明Fluid将会启动一个包含1个Alluxio Master和1个Alluxio Worker的Alluxio实例

**创建AlluxioRuntime资源并查看状态**
```shell
$ kubectl create -f runtime.yaml
alluxioruntime.data.fluid.io/hbase created

$ kubectl get pod -o wide
NAME                 READY   STATUS    RESTARTS   AGE   IP              NODE                       NOMINATED NODE   READINESS GATES
hbase-master-0       2/2     Running   0          85m   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
hbase-worker-0       2/2     Running   0          63m   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
```
在此处可以看到，AlluxioWorker被启动，并且运行在具有污点的节点之上。

**检查AlluxioRuntime状态**
```shell
$ kubectl get alluxioruntime hbase -o wide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE     AGE
hbase   1               1                 Ready          1               1                 Ready          0             0               Ready   4m3s
```

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
  replicas: 1
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
      tolerations:
      - key: hbase 
        operator: Equal 
        value: "true" 
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
```

可以看到Nginx Pod成功启动，并且运行在配置taint的节点之上

## 环境清理
```shell
$ kubectl delete -f .

$ kubectl taint nodes cn-beijing.192.168.1.146 hbase=true:NoSchedule-
```
