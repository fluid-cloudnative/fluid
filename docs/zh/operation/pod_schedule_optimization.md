# 示例 - Pod 调度优化
为了帮助用户更好的使用Fluid，我们提供了一系列调度插件。
通过为Pod自动注入亲和性相关信息，优化Pod的调度结果，提高集群的整体使用效率。

具体来说，Fluid结合根据数据集排布的Pod调度策略，通过webhook机制将调度信息注入到Pod可以实现以下功能：

1.支持K8s原生调度器,以及Volcano, Yunikorn等实现Pod数据亲和性调度  
2.将Pod优先调度到有数据缓存能力的节点  
3.可以通过指定Pod Label的形式，将Pod强制调度到有数据缓存的节点
4.当Pod不使用数据集时，可以尽量避免调度到有缓存的节点

## 前提条件

您使用的k8s版本需要支持 admissionregistration.k8s.io/v1（ Kubernetes version > 1.16 )
启用允许控制器集需要通过向 Kubernetes API 服务器传递一个标志来配置，确保你的集群进行了正常的配置
```yaml
--enable-admission-plugins=MutatingAdmissionWebhook
```
注意如果您的集群之前已经配置了其他的准入控制器，只需要增加 MutatingAdmissionWebhook 这个参数

## 使用方法

**查看全部结点**
```shell
$ kubectl get nodes
NAME                      STATUS   ROLES    AGE   VERSION
node.172.16.0.16   Ready    <none>   16d   v1.20.4-aliyun.1
node.172.16.1.84   Ready    <none>   16d   v1.20.4-aliyun.1
```

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

> 这里通过nodeselector指定了

**创建Dataset资源对象**
```shell
$ kubectl create -f dataset-global.yaml
dataset.data.fluid.io/hbase created
```

该配置文件片段中，包含了许多与Alluxio相关的配置信息，这些信息将被Fluid用来启动一个Alluxio实例。
上述配置片段中的`spec.replicas`属性被设置为1,这表明Fluid将会启动一个包含1个Alluxio Master和1个Alluxio Worker的Alluxio实例。

**创建AlluxioRuntime资源并查看状态**

```shell
$ kubectl create -f runtime.yaml
alluxioruntime.data.fluid.io/hbase created

$  kubectl get po -owide
NAME                 READY   STATUS    RESTARTS   AGE   IP             NODE                      NOMINATED NODE   READINESS GATES
hbase-master-0       2/2     Running   0          11m   172.16.0.16    node.172.16.0.16   <none>           <none>
hbase-worker-0       2/2     Running   0          10m   172.16.1.84    node.172.16.1.84   <none>           <none>
```
在此处可以看到，有一个Alluxio Worker成功启动，并且运行在结点172.16.1.84上。

## 运行示例1: 创建没有挂载数据集的Pod，它将尽量被调度到远离数据集的节点

**创建Pod**
```shell
$ cat<<EOF >nginx-1.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-1
  labels:
    fuse.serverful.fluid.io/inject: "true"
spec:
  containers:
    - name: nginx-1
      image: nginx
      volumeMounts:
        - mountPath: /data
          name: hbase-vol
  volumes:
    - name: hbase-vol
      persistentVolumeClaim:
        claimName: hbase
EOF
$ kubectl create -f nginx-1.yaml
```

示例中`metadata.labels`中新增`fuse.serverful.fluid.io/inject=true`以对该Pod开启Fluid的调度优化功能。

**查看Pod**

查看Pod的yaml文件，发现被注入了如下亲和性约束信息：

```yaml
spec:
  affinity:
    nodeAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
        - preference:
            matchExpressions:
              - key: fluid.io/dataset-num
                operator: DoesNotExist
          weight: 100
```

正如亲和性所影响的，Pod调度到了没有缓存(即无Alluxio Worker Pod运行)的node.172.16.0.16节点。

```shell
$ kubectl get pods nginx-1 -o  custom-columns=NAME:metadata.name,NODE:.spec.nodeName
NAME    NODE
nginx-1   node.172.16.0.16
```

## 运行示例2: 创建挂载数据集的Pod，它将尽量往存在所挂载数据集的节点调度

**创建Pod**

```shell
$ cat<<EOF >nginx-2.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-2
  labels:
    fuse.serverful.fluid.io/inject: "true"
spec:
  containers:
    - name: nginx-2
      image: nginx
      volumeMounts:
        - mountPath: /data
          name: hbase-vol
  volumes:
    - name: hbase-vol
      persistentVolumeClaim:
        claimName: hbase
EOF
$ kubectl create -f nginx-2.yaml
```

示例中`metadata.labels`中新增`fuse.serverful.fluid.io/inject=true`以对该Pod开启Fluid的调度优化功能。

**查看Pod**

查看Pod的yaml文件，发现被注入了如下信息：

```yaml
spec:
  affinity:
    nodeAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
      - preference:
          matchExpressions:
          - key: fluid.io/s-default-hbase
            operator: In
            values:
            - "true"
        weight: 100
```

通过Webhook机制，应用Pod被注入和缓存worker的弱亲和性配置。


```shell
$ kubectl get pods nginx-2 -o  custom-columns=NAME:metadata.name,NODE:.spec.nodeName
NAME    NODE
nginx-1   node.172.16.1.84
```

从结果上看, 可以看到pod被调度到了有数据缓存（即运行Alluxio Worker Pod）的节点。

## 运行示例3: 创建挂载数据集的Pod，通过指定Label将Pod调度到挂载数据集的节点

**创建Pod**
metadata中指定label（格式`fluid.io/dataset.{dataset_name}.sched: required`），如`fluid.io/dataset.hbase.sched: required`表明该Pod需要被调度到数据集 hbase 的缓存节点上。

```shell
$ cat<<EOF >nginx-3.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-3
  labels:
    fuse.serverful.fluid.io/inject: "true"
    fluid.io/dataset.hbase.sched: required
spec:
  containers:
    - name: nginx-3
      image: nginx
      volumeMounts:
        - mountPath: /data
          name: hbase-vol
  volumes:
    - name: hbase-vol
      persistentVolumeClaim:
        claimName: hbase
EOF
$ kubectl create -f nginx-3.yaml
```

**查看Pod**

查看Pod的yaml文件，发现被注入了如下信息：

```yaml
spec:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: fluid.io/s-default-hbase
            operator: In
            values:
            - "true"
```

通过Webhook机制，应用Pod被注入和缓存worker的强亲和性配置。


```shell
$ kubectl get pods nginx-3 -o  custom-columns=NAME:metadata.name,NODE:.spec.nodeName
NAME    NODE
nginx-3   node.172.16.1.84
```

从结果上看, 可以看到pod被调度到了有数据缓存（即运行Alluxio Worker Pod）的节点。