# 示例 - 远程文件访问加速
通过[Alluxio](https://www.alluxio.io)和[Fuse](https://github.com/libfuse/libfuse)，Fluid为用户提供了一种更为简单的文件访问接口，使得任意运行在Kubernetes集群上的程序能够像访问本地文件一样轻松访问存储在远程文件系统中的文件。更为重要的是，Fluid借助Alluxio提供了强大的文件缓存能力，这意味着用户在访问远程文件时，尤其是那些具有较高访问频率的远程文件时，用户可以享受到大幅度的文件访问速度的提升。

本文档通过一个简单的例子演示了上述功能特性

## 前提条件
在运行该示例之前，请参考[安装文档](../userguide/install.md)完成安装，并检查Fluid各组件正常运行：
```shell
$ kubectl get pod -n fluid-system
NAME                                  READY   STATUS    RESTARTS   AGE
controller-manager-7fd6457ccf-jnkvn   1/1     Running   0          60s
csi-nodeplugin-fluid-6rhpt            2/2     Running   0          60s
csi-nodeplugin-fluid-6zwgl            2/2     Running   0          60s
```
通常来说，你会看到一个名为“controller-manager”的Pod和多个名为“csi-nodeplugin”的Pod正在运行。其中，“csi-nodeplugin”这些Pod的数量取决于你的Kubernetes集群中结点的数量。

## 新建工作环境
```shell
$ mkdir <any-path>/accelerate
$ cd <any-path>/accelerate
```

## 运行示例

**查看待创建的Dataset资源对象**
```shell
$ cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: hbase
spec:
  mounts:
    - mountPoint: https://mirrors.tuna.tsinghua.edu.cn/apache/hbase/2.2.5/
      name: hbase
EOF
```

在这里，我们将要创建一个kind为`Dataset`的资源对象(Resource object)。`Dataset`是Fluid所定义的一个Custom Resource Definition(CRD)，该CRD被用来告知Fluid在哪里可以找到你所需要的数据。Fluid将该CRD对象中定义的`mountPoint`属性挂载到Alluxio之上，因此该属性可以是任何合法的能够被Alluxio识别的UFS地址。在本示例中，为了简单，我们使用[WebUFS](https://docs.alluxio.io/os/user/stable/en/ufs/WEB.html)进行演示。

更多有关UFS的信息，请参考[Alluxio文档-底层存储系统](https://docs.alluxio.io/os/user/stable/cn/ufs/OSS.html)部分。

> 本示例将以Apache镜像站点上的Hbase v2.25相关资源作为演示中使用的远程文件。这个选择并没有任何特殊之处，你可以将这个远程文件修改为任意你喜欢的远程文件。但是，如果你想要和我们一样使用WebUFS进行操作的话，最好还是选择一个Apache镜像源站点( e.g. [清华镜像源](https://mirrors.tuna.tsinghua.edu.cn/apache) )，因为基于目前WebUFS的实现，如果你选择其他更加复杂的网页作为WebUFS，你可能需要进行更多更复杂的配置。

**创建Dataset资源对象**
```shell
$ kubectl create -f dataset.yaml
dataset.data.fluid.io/hbase created
```

**查看Dataset资源对象状态**
```shell
$ kubectl get dataset hbase -o yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
...
status:
  conditions: []
  phase: NotBound
```

如上所示，`status`中的`phase`属性值为`NotBound`，这意味着该`Dataset`资源对象目前还未与任何`AlluxioRuntime`资源对象绑定，接下来，我们将创建一个`AlluxioRuntime`资源对象。

**查看待创建的AlluxioRuntime资源对象**
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
        storageType: Memory
  properties:
    alluxio.user.file.writetype.default: MUST_CACHE
    alluxio.master.journal.folder: /journal
    alluxio.master.journal.type: UFS
    alluxio.user.block.size.bytes.default: 256MB
    alluxio.user.streaming.reader.chunk.size.bytes: 256MB
    alluxio.user.local.reader.chunk.size.bytes: 256MB
    alluxio.worker.network.reader.buffer.size: 256MB
    alluxio.user.streaming.data.timeout: 300sec
  master:
    jvmOptions:
      - "-Xmx4G"
  worker:
    jvmOptions:
      - "-Xmx4G"
  fuse:
    jvmOptions:
      - "-Xmx4G "
      - "-Xms4G "
    # For now, only support local
    shortCircuitPolicy: local
    args:
      - fuse
      - --fuse-opts=direct_io,ro,max_read=131072,attr_timeout=7200,entry_timeout=7200,nonempty
EOF
```

**创建AlluxioRuntime资源对象**
```shell
$ kubectl create -f runtime.yaml
alluxioruntime.data.fluid.io/hbase created
```

`AlluxioRuntime`是另一个Fluid定义的CRD。一个`AlluxioRuntime`资源对象描述了在Kubernetes集群中运行一个Alluxio实例所需要的配置信息。

等待一段时间，让AlluxioRuntime资源对象中的各个组件得以顺利启动，你会看到类似以下状态：
```shell
$ kubectl get pod
NAME                 READY   STATUS    RESTARTS   AGE
hbase-fuse-hvxgh     1/1     Running   0          27s
hbase-fuse-sjhxk     1/1     Running   0          27s
hbase-master-0       2/2     Running   0          62s
hbase-worker-92cln   2/2     Running   0          27s
hbase-worker-rlb5w   2/2     Running   0          27s
```



**再次查看Dataset资源对象状态**
```shell
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
因为已经与一个成功启动的AlluxioRuntime绑定，该Dataset资源对象的`Status`得到了更新，此时`phase`属性值已经变为`Bound`状态。从上述状态中可以获知有关资源对象的基本信息

**查看AlluxioRuntime状态**
```shell
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
`AlluxioRuntime`资源对象的`status`中包含了更多更详细的信息

**查看与远程文件关联的PersistentVolume以及PersistentVolumeClaim**
```shell
$ kubectl get pv
NAME    CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM           STORAGECLASS   REASON   AGE
hbase   100Gi      RWX            Retain           Bound    default/hbase                           18m
```

```shell
$ kubectl get pvc
NAME    STATUS   VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS   AGE
hbase   Bound    hbase    100Gi      RWX                           18m
```
`Dataset`资源对象准备完成后（即与Alluxio实例绑定后），与该资源对象关联的PV, PVC已经由Fluid生成，应用可以通过该PVC完成远程文件在Pod中的挂载，并通过挂载目录实现远程文件访问

## 远程文件访问

**查看待创建的应用**
```shell
$ cat<<EOF >nginx.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
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

**启动应用进行远程文件访问**
```shell
$ kubectl create -f nginx.yaml
```

登录Nginx Pod:
```shell
$ kubectl exec -it nginx -- bash
```

查看远程文件挂载情况：
```shell
$ ls -1 /data/hbase
CHANGES.md
RELEASENOTES.md
api_compare_2.2.5RC0_to_2.2.4.html
hbase-2.2.5-bin.tar.gz
hbase-2.2.5-client-bin.tar.gz
hbase-2.2.5-src.tar.gz
```

```shell
$ du -h /data/hbase/*
174K    /data/hbase/CHANGES.md
106K    /data/hbase/RELEASENOTES.md
115K    /data/hbase/api_compare_2.2.5RC0_to_2.2.4.html
211M    /data/hbase/hbase-2.2.5-bin.tar.gz
200M    /data/hbase/hbase-2.2.5-client-bin.tar.gz
34M     /data/hbase/hbase-2.2.5-src.tar.gz
```

登出Nginx Pod:
```shell
$ exit
```

正如你所见，WebUFS上所存储的全部文件(也就是hbase v2.2.5的相关文件)可以以和本地文件完全没有区别的方式存在于某个Pod中，并且可以被该Pod十分方便地访问

## 远程文件访问加速

为了演示在访问远程文件时，你能获得多大的加速效果，我们提供了一个测试作业的样例:

**查看待创建的测试作业**
```shell
$ cat<<EOF >app.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: fluid-copy-test
spec:
  template:
    spec:
      restartPolicy: OnFailure
      containers:
        - name: busybox
          image: busybox
          command: ["/bin/sh"]
          args: ["-c", "set -x; time cp -r /data/hbase ./"]
          volumeMounts:
            - mountPath: /data
              name: hbase-vol
      volumes:
        - name: hbase-vol
          persistentVolumeClaim:
            claimName: hbase
EOF
```

**启动测试作业**
```shell
$ kubectl create -f app.yaml
job.batch/fluid-test created
```

该测试程序会执行`time cp -r /data/hbase ./`的shell命令，其中`/data/hbase`是远程文件在Pod中挂载的位置，该命令完成后会在终端显示命令执行的时长：

```shell
kubectl logs fluid-copy-test-h59w9
+ time cp -r /data/hbase ./
real  1m 2.74s
user  0m 0.00s
sys   0m 1.35s
```
可见，第一次远程文件的读取耗费了接近63s的时间。当然，你可能会觉得这并没有你预期的那么快，但是：

**再次启动测试作业**
```shell
$ kubectl delete -f app.yaml
$ kubectl create -f app.yaml
```

由于远程文件已经被缓存，此次测试作业能够迅速完成：
```shell
$ kubectl logs fluid-copy-test-d9h2x
+ time cp -r /data/hbase ./
real  0m 2.94s
user  0m 0.00s
sys   0m 1.27s
```
同样的文件访问操作仅耗费了3s

这种大幅度的加速效果归因于Alluxio所提供的强大的缓存能力，这种缓存能力意味着，只要你访问某个远程文件一次，该文件就会被缓存在Alluxio中，你的所有接下来的重复访问都不再需要进行远程文件读取，而是从Alluxio中直接获取数据，因此对于数据的访问加速也就不难解释了。

> 注意： 上述文件的访问速度与示例运行环境的网络条件有关，如果文件访问速度过慢，请更换更小的远程文件尝试

## 环境清理
```shell
$ kubectl delete -f .
```

