# 示例 - 远程文件访问加速
Fluid使用[Alluxio](https://www.alluxio.io)为用户提供了极其便捷的远程文件访问接口，使得程序能够像访问本地文件一样访问远程文件，同时，借助Alluxio提供的文件缓存能力，程序对于已访问过的文件重复访问能够获得大幅度的速度提升。本文档通过一个简单的例子演示了上述功能特性

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
**查看待创建的Dataset资源对象**
```shell script
$ cat samples/accelerate/dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: hbase
spec:
  mounts:
    - mountPoint: https://mirrors.tuna.tsinghua.edu.cn/apache/hbase/2.2.5/
      name: hbase
```
> 本示例将以Apache镜像站点上的Hbase v2.25相关资源作为演示中使用的远程文件

**创建Dataset资源对象**
```shell script
$ kubectl create -f samples/accelerate/dataset.yaml
dataset.data.fluid.io/hbase created
```

**查看Dataset资源对象状态**
```shell script
$ kubectl get dataset hbase -o yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
...
status:
  conditions: []
  phase: NotBound
```

该Dataset资源对象目前还未与任何AlluxioRuntime资源对象绑定，因此其`status`中的`phase`属性值为`NotBound`，这意味着该Dataset资源对象仍然处于不可用状态

**创建AlluxioRuntime资源对象**
```shell script
$ kubectl create -f samples/accelerate/runtime.yaml
alluxioruntime.data.fluid.io/hbase created
```

等待一段时间，让AlluxioRuntime资源对象中的各个组件得以顺利启动，看到类似以下状态：
```shell script
$ kubectl get pod
NAME                 READY   STATUS    RESTARTS   AGE
hbase-fuse-hvxgh     1/1     Running   0          27s
hbase-fuse-sjhxk     1/1     Running   0          27s
hbase-master-0       2/2     Running   0          62s
hbase-worker-92cln   2/2     Running   0          27s
hbase-worker-rlb5w   2/2     Running   0          27s
```

**再次查看Dataset资源对象状态**
```shell script
$ kubectl get dataset hbase -o yaml
...
...
status:
  cacheStates:
    cacheCapacity: 4GiB
    cached: 0B
    cachedPercentage: 0%
  conditions:
  - lastTransitionTime: "2020-07-29T08:23:44Z"
    lastUpdateTime: "2020-07-29T08:26:29Z"
    message: The ddc runtime is ready.
    reason: DatasetReady
    status: "True"
    type: Ready
  phase: Bound
  runtimes:
  - category: Accelerate
    name: hbase
    namespace: default
    type: alluxio
  ufsTotal: 443.5MiB
```
因为已经与一个成功启动的AlluxioRuntime绑定，该Dataset资源对象的`Status`得到了更新，从上述状态中可以获知有关资源对象的基本信息

**查看AlluxioRuntime状态**
```shell script
$ kubectl get alluxioruntime hbase -o yaml
...
...
status:
  cacheStates:
    cacheCapacity: 4GiB
    cached: 0B
    cachedPercentage: 0%
  conditions:
  - lastProbeTime: "2020-07-29T08:23:05Z"
    lastTransitionTime: "2020-07-29T08:23:05Z"
    message: The master is initialized.
    reason: Master is initialized
    status: "True"
    type: MasterInitialized
  - lastProbeTime: "2020-07-29T08:23:40Z"
    lastTransitionTime: "2020-07-29T08:23:05Z"
    message: The master is ready.
    reason: Master is ready
    status: "True"
    type: MasterReady
  - lastProbeTime: "2020-07-29T08:23:20Z"
    lastTransitionTime: "2020-07-29T08:23:20Z"
    message: The workers are initialized.
    reason: Workers are initialized
    status: "True"
    type: WorkersInitialized
  - lastProbeTime: "2020-07-29T08:23:20Z"
    lastTransitionTime: "2020-07-29T08:23:20Z"
    message: The fuses are initialized.
    reason: Fuses are initialized
    status: "True"
    type: FusesInitialized
  - lastProbeTime: "2020-07-29T08:23:40Z"
    lastTransitionTime: "2020-07-29T08:23:40Z"
    message: The workers are partially ready.
    reason: Workers are ready
    status: "True"
    type: WorkersReady
  - lastProbeTime: "2020-07-29T08:23:40Z"
    lastTransitionTime: "2020-07-29T08:23:40Z"
    message: The fuses are ready.
    reason: Fuses are ready
    status: "True"
    type: FusesReady
  currentFuseNumberScheduled: 2
  currentMasterNumberScheduled: 1
  currentWorkerNumberScheduled: 2
  desiredFuseNumberScheduled: 2
  desiredMasterNumberScheduled: 1
  desiredWorkerNumberScheduled: 2
  fuseNumberAvailable: 2
  fuseNumberReady: 2
  fusePhase: Ready
  masterNumberReady: 1
  masterPhase: Ready
  valueFile: hbase-alluxio-values
  workerNumberAvailable: 2
  workerNumberReady: 2
  workerPhase: Ready
```

**查看与远程文件关联的PersistentVolume以及PersistentVolumeClaim**
```shell script
$ kubectl get pv
NAME    CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM           STORAGECLASS   REASON   AGE
hbase   100Gi      RWX            Retain           Bound    default/hbase                           18m
```

```shell script
$ kubectl get pvc
NAME    STATUS   VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS   AGE
hbase   Bound    hbase    100Gi      RWX                           18m
```
与远程文件关联的PV,PVC已经由Fluid生成，应用可以通过该PVC完成远程文件在Pod中的挂载，并通过挂载目录实现远程文件访问

## 远程文件访问

**启动应用进行远程文件访问**
```shell script
kubectl create -f samples/accelerate/nginx.yaml
```

登录Nginx Pod:
```shell script
kubectl exec -it nginx -- bash
```

查看远程文件挂载情况：
```shell script
# ls -1 /data/hbase
CHANGES.md
RELEASENOTES.md
api_compare_2.2.5RC0_to_2.2.4.html
hbase-2.2.5-bin.tar.gz
hbase-2.2.5-client-bin.tar.gz
hbase-2.2.5-src.tar.gz
```

```shell script
# du -sh /data/hbase/hbase-2.2.5-client-bin.tar.gz
200M    /data/hbase/hbase-2.2.5-client-bin.tar.gz
```

## 远程文件访问加速

**启动测试作业**
```shell script
$ kubectl create -f samples/accelerate/test.yaml
job.batch/fluid-test created
```
该测试程序会尝试读取一个远程文件(e.g. `hbase-2.2.5-client-bin.tar.gz`)，并打印出此过程所耗费的时间：
```shell script
$ kubectl logs fluid-test-cqmwj
real    1m 9.55s
user    0m 0.00s
sys     0m 0.64s
```
可见，第一次远程文件的读取耗费了接近70s的时间

**再次启动测试作业**
```shell script
kubectl delete -f samples/accelerate/test.yaml
kubectl create -f samples/accelerate/test.yaml
```
由于远程文件已经被缓存，此次测试作业能够迅速完成：
```shell script
$ kubectl logs fluid-test-hpzqc
real    0m 2.03s
user    0m 0.00s
sys     0m 0.63s
```
同样的文件访问操作仅耗费了2s

因为该文件已经在Alluxio中被缓存，因此访问的速度大大加快，可见，Fluid利用Alluxio实现了远程文件访问的加速

> 注意： 上述文件的访问速度与示例运行环境的网络条件有关，如果文件访问速度过慢，请更换更小的远程文件尝试

## 环境清理
```shell script
kubectl delete -f samples/accelerate
```

