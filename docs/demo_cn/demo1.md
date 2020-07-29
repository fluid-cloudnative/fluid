# 示例 - 数据集访问加速

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

**创建Dataset资源**
```shell script
$ kubectl create -f samples/demo1/demo_dataset.yaml 
dataset.data.fluid.io/cifar10 created
```

**查看Dataset状态**
```shell script
$ kubectl get dataset cifar10 -o yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  creationTimestamp: "2020-07-26T14:37:23Z"
  finalizers:
  - fluid-dataset-controller-finalizer
  generation: 1
  name: cifar10
  namespace: default
  resourceVersion: "35765870"
  selfLink: /apis/data.fluid.io/v1alpha1/namespaces/default/datasets/cifar10
  uid: 0cc6fa22-7e96-4b51-9e1f-8ce4b72c5c6c
spec:
  mounts:
  - mountPoint: https://mirrors.tuna.tsinghua.edu.cn/apache/hbase/2.2.5/
    name: hbase
status:
  conditions: []
  phase: NotBound
```

Dataset目前还未与一个配置完成的AlluxioRuntime绑定，因此在`status`中`phase`属性值为`NotBound`

**创建AlluxioRuntime**
```shell script
$ kubectl create -f samples/demo1/demo_runtime.yaml
alluxioruntime.data.fluid.io/cifar10 created
```

等待一段时间，让AlluxioRuntime中的各个组件得以顺利启动，看到类似以下状态：
```shell script
$ kubectl get pod
NAME                   READY   STATUS    RESTARTS   AGE
cifar10-fuse-sf44m     1/1     Running   0          58s
cifar10-fuse-w85vj     1/1     Running   0          58s
cifar10-master-0       2/2     Running   0          2m59s
cifar10-worker-2vsbz   2/2     Running   0          58s
cifar10-worker-znl8k   2/2     Running   0          58s
```

**查看Dataset状态**
```shell script
$ kubectl get dataset cifar10 -o yaml
...
...
status:
  cacheStates:
    cacheCapacity: 4GiB
    cached: 0B
    cachedPercentage: 0%
  conditions:
  - lastTransitionTime: "2020-07-26T14:41:45Z"
    lastUpdateTime: "2020-07-26T14:41:47Z"
    message: The ddc runtime is ready.
    reason: DatasetReady
    status: "True"
    type: Ready
  phase: Bound
  runtimes:
  - category: Accelerate
    name: cifar10
    namespace: default
    type: alluxio
  ufsTotal: 443.5MiB
```
因为已经与一个成功启动的AlluxioRuntime绑定，该Dataset的`Status`得到了更新，从这些信息中可以获得一些基本信息

**查看AlluxioRuntime状态**
```shell script
$ kubectl get alluxioruntime cifar10 -o yaml
...
...
status:
  cacheStates:
    cacheCapacity: 4GiB
    cached: 0B
    cachedPercentage: 0%
  conditions:
  - lastProbeTime: "2020-07-26T14:39:01Z"
    lastTransitionTime: "2020-07-26T14:39:01Z"
    message: The master is initialized.
    reason: Master is initialized
    status: "True"
    type: MasterInitialized
  - lastProbeTime: "2020-07-26T14:41:42Z"
    lastTransitionTime: "2020-07-26T14:39:21Z"
    message: The master is ready.
    reason: Master is ready
    status: "True"
    type: MasterReady
  - lastProbeTime: "2020-07-26T14:40:42Z"
    lastTransitionTime: "2020-07-26T14:40:42Z"
    message: The workers are initialized.
    reason: Workers are initialized
    status: "True"
    type: WorkersInitialized
  - lastProbeTime: "2020-07-26T14:40:42Z"
    lastTransitionTime: "2020-07-26T14:40:42Z"
    message: The fuses are initialized.
    reason: Fuses are initialized
    status: "True"
    type: FusesInitialized
  - lastProbeTime: "2020-07-26T14:41:42Z"
    lastTransitionTime: "2020-07-26T14:41:42Z"
    message: The workers are ready.
    reason: Workers are ready
    status: "True"
    type: WorkersReady
  - lastProbeTime: "2020-07-26T14:41:42Z"
    lastTransitionTime: "2020-07-26T14:41:42Z"
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
  valueFile: cifar10-alluxio-values
  workerNumberAvailable: 2
  workerNumberReady: 2
  workerPhase: Ready
```

**查看与数据集关联的PersistentVolume以及PersistentVolumeClaim**
```shell script
$ kubectl get pv
NAME      CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM             STORAGECLASS   REASON   AGE
cifar10   100Gi      RWX            Retain           Bound    default/cifar10                           2m20s
```

```shell script
$ kubectl get pvc
NAME      STATUS   VOLUME    CAPACITY   ACCESS MODES   STORAGECLASS   AGE
cifar10   Bound    cifar10   100Gi      RWX                           2m38s
```
与数据集关联的PV,PVC已经由Fluid成功生成，应用可以通过该PVC实现对于该数据集的访问

## 数据集访问加速
**启动测试作业**
```shell script
$ kubectl create -f samples/demo1/demo_test.yaml
job.batch/fluid-test created
```
该测试程序会尝试访问数据集，并打印出访问所耗费的时间
```shell script
$ kubectl logs fluid-test-cqmwj
real    1m 9.55s
user    0m 0.00s
sys     0m 0.64s
```
此过程耗费了接近70s的时间

**再次启动测试作业**
```shell script
kubectl delete -f samples/demo1/demo_test.yaml
kubectl create -f samples/demo1/demo_test.yaml
```
由于数据集已经被缓存，此次测试作业能够迅速完成：
```shell script
$ kubectl logs fluid-test-hpzqc
real    0m 2.03s
user    0m 0.00s
sys     0m 0.63s
```
同样的数据访问操作仅耗费了2s

因为该数据已经在Alluxio中被缓存，因此数据访问的速度大大加快，可见，Fluid利用Alluxio实现了数据集访问的加速

> 注意： 上述数据集的访问速度与示例运行环境的网络条件有关，如果数据访问速度过慢，请更换更小的数据集尝试

## 环境清理
```shell script
kubectl delete -f samples/demo1/
```

