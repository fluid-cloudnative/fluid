# 示例 - 数据集亲和性调度
通常情况下，执行模型训练时需要使用到特殊的硬件（e.g. GPU/TPU）对整个训练过程进行加速。为了能够将待训练的数据集放置在具有特殊硬件的集群结点上，Fluid提供了针对数据集的亲和性调度，本文档通过一个简单的例子演示了该特性。

## 前提条件
在运行该示例之前，请参考[安装文档](../installation_cn/README.md)完成安装，并检查Fluid各组件正常运行：
```shell script
$ kubectl get pod -n fluid-system
NAME                                  READY   STATUS    RESTARTS   AGE
controller-manager-7fd6457ccf-jnkvn   1/1     Running   0          60s
csi-nodeplugin-fluid-6rhpt            2/2     Running   0          60s
csi-nodeplugin-fluid-6zwgl            2/2     Running   0          60s
```

## 运行示例
> 为了演示，在接下来的示例中将使用结点上的Label来标识结点上是否具有特殊硬件

**查看具有特殊硬件的结点**
```shell script
$ kubectl get node -L aliyun.accelerator/nvidia_name
NAME                       STATUS   ROLES    AGE     VERSION            NVIDIA_NAME
cn-beijing.192.168.1.146   Ready    <none>   4d13h   v1.16.9-aliyun.1   Tesla-V100-SXM2-16GB
cn-beijing.192.168.1.147   Ready    <none>   4d13h   v1.16.9-aliyun.1   
```
目前，在全部2个结点中，仅有一个结点包含型号为Nvidia Tesla V100的计算加速硬件设备

**检查待创建的Dataset资源**
```shell script
$ cat samples/demo2/demo_dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: cifar10
spec:
  ...
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: aliyun.accelerator/nvidia_name
              operator: In
              values:
                - Tesla-V100-SXM2-16GB
```
在该Dataset的`spec.nodeAffinity`属性中定义了亲和性调度的相关配置，该配置要求将数据集缓存在具有上述型号特殊硬件的结点上

**创建Dataset资源**
```shell script
$ kubectl create -f samples/demo2/demo_dataset.yaml
dataset.data.fluid.io/cifar10 created
```

**检查待创建的AlluxioRuntime资源**
```shell script
cat samples/demo2/demo_runtime.yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: cifar10
spec:
  ...
  master:
    replicas: 1
    ...
  worker:
    replicas: 2
    ...
  fuse:
    image: alluxio/alluxio-fuse
    imageTag: "2.3.0-SNAPSHOT"
    imagePullPolicy: Always
    ...
```
该配置文件表明希望创建一个AlluxioRuntime资源，其中包含1个AlluxioMaster和2个AlluxioWorker，并且2个AlluxioWorker均有对应的Alluxio Fuse

**创建AlluxioRuntime资源并查看状态**
```shell script
$ kubectl create -f samples/demo2/demo_runtime.yaml
alluxioruntime.data.fluid.io/cifar10 created

$ kubectl get pod -o wide
NAME                   READY   STATUS    RESTARTS   AGE     IP              NODE                       NOMINATED NODE   READINESS GATES
cifar10-fuse-qtxl7     1/1     Running   0          3m24s   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
cifar10-master-0       2/2     Running   0          4m57s   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
cifar10-worker-n87mf   2/2     Running   0          3m24s   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
```
仅有一组Alluxio Worker/Alluxio Fuse成功运行，并且均执行在具有特殊硬件的结点（即`cn-beijing.192.168.1.146`）之上。

**检查AlluxioRuntime状态**
```shell script
$ kubectl get alluxioruntime cifar10 -o yaml
...
status:
  cacheStates:
    cacheCapacity: 2GiB
    cached: 0B
    cachedPercentage: 0%
  conditions:
  ...
  currentFuseNumberScheduled: 1
  currentMasterNumberScheduled: 1
  currentWorkerNumberScheduled: 1
  desiredFuseNumberScheduled: 2
  desiredMasterNumberScheduled: 1
  desiredWorkerNumberScheduled: 2
  fuseNumberAvailable: 1
  fuseNumberReady: 1
  fusePhase: PartialReady
  masterNumberReady: 1
  masterPhase: Ready
  valueFile: cifar10-alluxio-values
  workerNumberAvailable: 1
  workerNumberReady: 1
  workerPhase: PartialReady
```
与预想一致，无论是Alluxio Worker还是Alluxio Fuse，均只是PartialReady，这是另一个结点没有满足Dataset亲和性要求的特殊硬件所致

**运行应用模拟模型训练过程**
> 为了演示，接下来将使用Nginx服务器应用使用上述数据集。通常情况下，您不会这么做，但在本示例中为了简单，我们使用该应用演示数据集的亲和性调度特性

```shell script
$ kubectl create -f samples/demo2/demo_app.yaml 
statefulset.apps/nginx created
```

**查看应用运行状态**
```shell script
kubectl get pod -o wide -l app=nginx
NAME      READY   STATUS    RESTARTS   AGE    IP              NODE                       NOMINATED NODE   READINESS GATES
nginx-0   1/1     Running   0          2m5s   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
nginx-1   0/1     Pending   0          2m5s   <none>          <none>                     <none>           <none>
```
仅有一个nginx应用成功启动，并且运行在含有特殊硬件的结点上

**查看应用启动失败原因**
```shell script
$ kubectl describe pod nginx-1
...
Events:
  Type     Reason            Age        From               Message
  ----     ------            ----       ----               -------
  Warning  FailedScheduling  <unknown>  default-scheduler  0/2 nodes are available: 1 node(s) didn't match pod affinity/anti-affinity, 1 node(s) didn't satisfy existing pods anti-affinity rules, 1 node(s) had volume node affinity conflict.
  Warning  FailedScheduling  <unknown>  default-scheduler  0/2 nodes are available: 1 node(s) didn't match pod affinity/anti-affinity, 1 node(s) didn't satisfy existing pods anti-affinity rules, 1 node(s) had volume node affinity conflict.
```
一方面，由于`samples/demo2/demo_app.yaml`中对于`PodAntiAffinity`的配置，使得两个Nginx Pod无法被调度到同一节点。**另一方面，由于目前满足Dataset亲和性要求的结点仅有一个，因此仅有一个Nginx Pod被成功调度**

**增加特殊硬件**
```shell script
kubectl label node cn-beijing.192.168.1.147 aliyun.accelerator/nvidia_name=Tesla-V100-SXM2-16GB
```
现在两个结点都具有相同型号的特殊硬件了，此时重新检查各个组件的运行状态
```shell script
$ kubectl get pod -o wide
NAME                   READY   STATUS    RESTARTS   AGE   IP              NODE                       NOMINATED NODE   READINESS GATES
cifar10-fuse-qmjh5     1/1     Running   0          10m   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
cifar10-fuse-qtxl7     1/1     Running   0          44m   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
cifar10-master-0       2/2     Running   0          46m   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
cifar10-worker-n87mf   2/2     Running   0          44m   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
cifar10-worker-wmhkg   2/2     Running   0          10m   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
```
两个Alluxio Worker和Alluxio Fuse都成功启动，并且分别运行在两个结点上

```shell script
$ kubectl get pod -l app=nginx -o wide
NAME      READY   STATUS    RESTARTS   AGE   IP              NODE                       NOMINATED NODE   READINESS GATES
nginx-0   1/1     Running   0          21m   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
nginx-1   1/1     Running   0          21m   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
```
两个Nginx应用均成功启动，并且分别运行在两个结点上

可见，Fluid支持Dataset资源的亲和性调度，该亲和性调度的能力为数据密集作业在Kubernetes集群上的运行提供了更强的灵活性

## 环境清理
```shell script
kubectl delete -f samples/demo2
```













