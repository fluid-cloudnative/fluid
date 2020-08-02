# 示例 - 数据缓存亲和性调度
Fluid提供了针对数据缓存的调度机制，这意味着用户能够像管理Pod一样管理数据缓存在Kubernetes集群中的存放位置，这些存放位置同样也会间接地影响相关应用的调度策略。本文档通过一个简单的示例来演示上述功能特性，该示例将会尝试将远程文件的数据缓存分布在指定的集群结点之上，并启动应用使用这些数据缓存

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
**查看全部结点**
```shell script
$ kubectl get nodes
NAME                       STATUS   ROLES    AGE     VERSION
cn-beijing.192.168.1.146   Ready    <none>   7d14h   v1.16.9-aliyun.1
cn-beijing.192.168.1.147   Ready    <none>   7d14h   v1.16.9-aliyun.1
```

**使用标签标识结点**
```shell script
$ kubectl label nodes cn-beijing.192.168.1.146 hbase-cache=true
```

**再次查看结点**
```shell script
$ kubectl get node -L hbase-cache
NAME                       STATUS   ROLES    AGE     VERSION            HBASE-CACHE
cn-beijing.192.168.1.146   Ready    <none>   7d14h   v1.16.9-aliyun.1   true
cn-beijing.192.168.1.147   Ready    <none>   7d14h   v1.16.9-aliyun.1   
```
目前，在全部2个结点中，仅有一个结点添加了`hbase-cache=true`的标签，接下来将使用该标签作为依据进行数据缓存的调度

**检查待创建的Dataset资源对象**
```shell script
$ cat samples/co-locality/dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: hbase
spec:
  mounts:
    - mountPoint: https://mirrors.tuna.tsinghua.edu.cn/apache/hbase/2.2.5/
      name: hbase
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: hbase-cache
              operator: In
              values:
                - "true"
```
在该Dataset资源对象的`spec.nodeAffinity`属性中定义了亲和性调度的相关配置，该配置要求将数据缓存放置在具有`hbase-cache=true`标签的结点之上

**创建Dataset资源对象**
```shell script
$ kubectl create -f samples/co-locality/dataset.yaml
dataset.data.fluid.io/hbase created
```

**检查待创建的AlluxioRuntime资源对象**
```shell script
cat samples/co-locality/runtime.yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: hbase
spec:
  ...
  replicas: 2
  master:
    replicas: 1
    ...
  worker:
    ...
  fuse:
    image: alluxio/alluxio-fuse
    imageTag: "2.3.0-SNAPSHOT"
    imagePullPolicy: Always
    ...
```
该配置文件表明希望创建一个AlluxioRuntime资源，其中包含1个Alluxio Master和2个Alluxio Worker，并且对于任意一个Alluxio Worker均会启动一个Alluxio Fuse组件与其协同工作

**创建AlluxioRuntime资源并查看状态**
```shell script
$ kubectl create -f samples/co-locality/runtime.yaml
alluxioruntime.data.fluid.io/hbase created

$ kubectl get pod -o wide
NAME                 READY   STATUS    RESTARTS   AGE    IP              NODE                       NOMINATED NODE   READINESS GATES
hbase-fuse-42csf     1/1     Running   0          104s   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
hbase-master-0       2/2     Running   0          3m3s   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
hbase-worker-l62m4   2/2     Running   0          104s   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
```
仅有一组Alluxio Worker/Alluxio Fuse成功启动，并且均运行在具有指定标签的结点（即`cn-beijing.192.168.1.146`）之上。

**检查AlluxioRuntime状态**
```shell script
$ kubectl get alluxioruntime hbase -o yaml
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
  valueFile: hbase-alluxio-values
  workerNumberAvailable: 1
  workerNumberReady: 1
  workerPhase: PartialReady
```
与预想一致，无论是Alluxio Worker还是Alluxio Fuse，其状态均为PartialReady，这是另一个结点无法满足Dataset资源对象的亲和性要求所致

**查看待创建的应用**
```shell script
$ cat samples/co-locality/app.yaml
...
spec:
  ...
  template: # define the pods specifications
    ...
    spec:
      affinity:
        # prevent two Nginx Pod from being scheduled at the same Node
        # just for demonstrating co-locality demo
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values:
                - nginx
            topologyKey: "kubernetes.io/hostname"
...
```
该应用定义了`PodAntiAffinity`的相关配置，这些配置将确保属于相同应用的多个Pod不会被调度到同一结点，通过这样的配置，能够更加清楚地演示数据缓存的调度对使用该数据缓存的应用的影响

**运行应用**

```shell script
$ kubectl create -f samples/co-locality/app.yaml
statefulset.apps/nginx created
```

**查看应用运行状态**
```shell script
kubectl get pod -o wide -l app=nginx
NAME      READY   STATUS    RESTARTS   AGE    IP              NODE                       NOMINATED NODE   READINESS GATES
nginx-0   1/1     Running   0          2m5s   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
nginx-1   0/1     Pending   0          2m5s   <none>          <none>                     <none>           <none>
```
仅有一个Nginx Pod成功启动，并且运行在具有指定标签的结点上

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
一方面，由于`samples/co-locality/app.yaml`中对于`PodAntiAffinity`的配置，使得两个Nginx Pod无法被调度到同一节点。**另一方面，由于目前满足Dataset资源对象亲和性要求的结点仅有一个，因此仅有一个Nginx Pod被成功调度**

**为结点添加标签**
```shell script
kubectl label node cn-beijing.192.168.1.147 hbase-cache=true
```
现在两个结点都具有相同的标签了，此时重新检查各个组件的运行状态
```shell script
$ kubectl get pod -o wide
NAME                 READY   STATUS    RESTARTS   AGE   IP              NODE                       NOMINATED NODE   READINESS GATES
hbase-fuse-42csf     1/1     Running   0          44m   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
hbase-fuse-kth4g     1/1     Running   0          10m   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
hbase-master-0       2/2     Running   0          46m   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
hbase-worker-l62m4   2/2     Running   0          44m   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
hbase-worker-rvncl   2/2     Running   0          10m   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
```
两个Alluxio Worker和Alluxio Fuse都成功启动，并且分别运行在两个结点上

```shell script
$ kubectl get pod -l app=nginx -o wide
NAME      READY   STATUS    RESTARTS   AGE   IP              NODE                       NOMINATED NODE   READINESS GATES
nginx-0   1/1     Running   0          21m   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
nginx-1   1/1     Running   0          21m   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
```
两个Nginx Pod均成功启动，并且分别运行在两个结点上

可见，Fluid支持数据缓存的调度策略，这些调度策略为用户提供了更加灵活的数据缓存管理能力

## 环境清理
```shell script
kubectl delete -f samples/co-locality

kubectl label node cn-beijing.192.168.1.146 hbase-cache-
kubectl label node cn-beijing.192.168.1.147 hbase-cache-
```
