# 示例 - 单机多Dataset远程Web文件访问加速
通过[Alluxio](https://www.alluxio.io)和[Fuse](https://github.com/libfuse/libfuse)，Fluid为用户提供了一种更为简单的文件访问接口，使得任意运行在Kubernetes集群上的程序能够像访问本地文件一样轻松访问存储在远程文件系统中的文件。Fluid 针对数据集进行全生命周期的管理和隔离，尤其对于短生命周期应用（e.g 数据分析任务、机器学习任务），用户可以在集群中大规模部署。

本文档通过一个简单的例子演示了上述功能特性

## 前提条件
在运行该示例之前，请参考[安装文档](../userguide/install.md)完成安装，并检查Fluid各组件正常运行：
```shell
$ kubectl get pod -n fluid-system
NAME                                  READY   STATUS    RESTARTS   AGE     USED-GPU   IP             NODE                NOMINATED NODE   READINESS GATES
controller-manager-7f7c44b97f-4zqxc   1/1     Running   0          15h     0/0        10.10.10.136   9f-03-00-27-fd-24   <none>           <none>
csi-nodeplugin-fluid-299r9            2/2     Running   0          2d14h   0/0        10.10.10.162   9f-03-00-ab-aa-bc   <none>           <none>
csi-nodeplugin-fluid-2wqfx            2/2     Running   0          2d14h   0/0        10.10.10.124   9b-03-00-cc-15-c4   <none>           <none>
csi-nodeplugin-fluid-4dt5c            2/2     Running   0          2d14h   0/0        10.10.10.165   9f-03-00-ab-aa-d4   <none>           <none>
csi-nodeplugin-fluid-4sdqf            2/2     Running   0          2d14h   0/0        10.10.10.134   9f-03-00-27-fd-3c   <none>           <none>
```
通常来说，你会看到一个名为`controller-manager`的Pod和多个名为`csi-nodeplugin`的Pod正在运行。其中，`csi-nodeplugin`这些Pod的数量取决于你的Kubernetes集群中结点的数量。

## 运行示例

**对某个节点打标签** 
```shell
$ kubectl label node 9f-03-00-05-96-fc fluid=multi-dataset
```
> 在接下来的步骤中，我们将使用 `NodeSelector` 来管理Dataset调度的节点，这里仅做试验使用。 

**查看待创建的Dataset资源对象**
```shell
$ cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: dataset-0
spec:
  mounts:
    - mountPoint: http://dl.unisound.ai/atlas/lustre-fs/lustre-server-latest/
      name: dataset
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: fluid
              operator: In
              values:
                - "multi-dataset"

EOF

$ cat<<EOF >dataset1.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: dataset-1
spec:
  mounts:
    - mountPoint: http://dl.unisound.ai/atlas/lustre-fs/lustre-server-latest/
      name: dataset
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: fluid
              operator: In
              values:
                - "multi-dataset"

EOF        
```
> 注意: 上述`mountPoint`中使用了内部源，非Apache。用户试验可使用Apache清华镜像源进行替换，如果你的环境位于海外，请更换为`https://downloads.apache.org/hbase/2.2.5/`进行尝试

在这里，我们将要创建一个kind为`Dataset`的资源对象(Resource object)。`Dataset`是Fluid所定义的一个Custom Resource Definition(CRD)，该CRD被用来告知Fluid在哪里可以找到你所需要的数据。Fluid将该CRD对象中定义的`mountPoint`属性挂载到Alluxio之上，因此该属性可以是任何合法的能够被Alluxio识别的UFS地址。在本示例中，为了简单，我们使用[WebUFS](https://docs.alluxio.io/os/user/stable/cn/ufs/WEB.html)进行演示。

更多有关UFS的信息，请参考[Alluxio文档-底层存储系统](https://docs.alluxio.io/os/user/stable/cn/ufs/OSS.html)部分。

> 本示例将以内部源 lustre 安装文件相关资源作为演示中使用的远程文件。这个选择并没有任何特殊之处，你可以将这个远程文件修改为任意你喜欢的远程文件。但是，如果你想要和我们一样使用WebUFS进行操作的话，最好还是选择一个Apache镜像源站点( e.g. [清华镜像源](https://mirrors.tuna.tsinghua.edu.cn/apache) )，因为根据目前WebUFS的实现，如果你选择其他更加复杂的网页作为WebUFS，你可能需要进行更多[更复杂的配置](https://docs.alluxio.io/os/user/stable/cn/ufs/WEB.html#%E9%85%8D%E7%BD%AEalluxio)

**创建Dataset资源对象**
```shell
$ kubectl apply -f dataset.yaml
dataset.data.fluid.io/dataset-0 created
$ kubectl apply -f dataset1.yaml
dataset.data.fluid.io/dataset-1 created
```

**查看Dataset资源对象状态**
```shell
$ kubectl get dataset
NAME        UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE      AGE
dataset-0                                                                  NotBound   17s
dataset-1                                                                  NotBound   13s
```

如上所示，`status`中的`phase`属性值为`NotBound`，这意味着该`Dataset`资源对象目前还未与任何`AlluxioRuntime`资源对象绑定，接下来，我们将创建一个`AlluxioRuntime`资源对象。

**查看待创建的AlluxioRuntime资源对象**
```shell
$ cat<<EOF >runtime.yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: dataset-0
spec:
  replicas: 1
  alluxioVersion:
    image: harbor.unisound.ai/xieydd/alluxio
    imageTag: "2.3.0-SNAPSHOT-e0feba3"
    imagePullPolicy: Always
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 2Gi
        high: "0.95"
        low: "0.7"
  runAs:
    uid: 831
    gid: 831
    user: xieyuandong
    group: xieyuandong
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
    image: harbor.unisound.ai/xieydd/alluxio-fuse
    imageTag: "2.3.0-SNAPSHOT-e0feba3"
    imagePullPolicy: Always
    jvmOptions:
      - "-Xmx4G "
      - "-Xms4G "
    args:
      - fuse
      - --fuse-opts=direct_io,ro,max_read=131072,attr_timeout=7200,entry_timeout=7200,nonempty

EOF

cat<<EOF >runtime1.yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: dataset-1
spec:
  replicas: 1
  alluxioVersion:
    image: harbor.unisound.ai/xieydd/alluxio
    imageTag: "2.3.0-SNAPSHOT-e0feba3"
    imagePullPolicy: Always
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 2Gi
        high: "0.95"
        low: "0.7"
  runAs:
    uid: 831
    gid: 831
    user: xieyuandong
    group: xieyuandong
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
    image: harbor.unisound.ai/xieydd/alluxio-fuse
    imageTag: "2.3.0-SNAPSHOT-e0feba3"
    imagePullPolicy: Always
    jvmOptions:
      - "-Xmx4G "
      - "-Xms4G "
    args:
      - fuse
      - --fuse-opts=direct_io,ro,max_read=131072,attr_timeout=7200,entry_timeout=7200,nonempty

EOF
```
> runAs中设置的用户,为non-root访问设置，详细可参考[nonroot_access](./nonroot_access.md)文档

**创建AlluxioRuntime资源对象**
```shell
$ kubectl create -f runtime.yaml
alluxioruntime.data.fluid.io/dataset-0 created

# 注意等待 dataset-0 组件全部 Running 
$ kubectl get pod | grep dataset
dataset-0-fuse-7qtvx                1/1     Running   0          112s          10.10.10.127    9f-03-00-05-96-fc
dataset-0-master-0                  2/2     Running   0          2m25s         10.10.10.126    9b-03-00-cc-15-f4
dataset-0-worker-ls59l              2/2     Running   0          112s          10.10.10.127    9f-03-00-05-96-fc 

$ kubectl create -f runtime1.yaml
alluxioruntime.data.fluid.io/dataset-1 created
```

**检查AlluxioRuntime资源对象是否已经创建**
```shell
$ kubectl get alluxioruntime
NAME        MASTER PHASE   WORKER PHASE   FUSE PHASE   AGE
dataset-0   Ready          Ready          Ready        17m
dataset-1   Ready          Ready          Ready        101s
```

`AlluxioRuntime`是另一个Fluid定义的CRD。一个`AlluxioRuntime`资源对象描述了在Kubernetes集群中运行一个Alluxio实例所需要的配置信息。

等待一段时间，让AlluxioRuntime资源对象中的各个组件得以顺利启动，你会看到类似以下状态：
```shell
$ kubectl get pod | grep dataset
dataset-0-fuse-7qtvx                1/1     Running   0          15m    0/0        10.10.10.127    9f-03-00-05-96-fc   <none>           <none>
dataset-0-master-0                  2/2     Running   0          16m    0/0        10.10.10.126    9b-03-00-cc-15-f4   <none>           <none>
dataset-0-worker-ls59l              2/2     Running   0          15m    0/0        10.10.10.127    9f-03-00-05-96-fc   <none>           <none>
dataset-1-fuse-plw5g                1/1     Running   0          13s    0/0        10.10.10.127    9f-03-00-05-96-fc   <none>           <none>
dataset-1-master-0                  2/2     Running   0          42s    0/0        10.10.10.153    9f-03-00-56-82-c0   <none>           <none>
dataset-1-worker-cmgrz              2/2     Running   0          13s    0/0        10.10.10.127    9f-03-00-05-96-fc   <none>           <none>
```
注意上面的不同的 Dataset 的 worker 和 fuse 组件可以正常的调度到相同的节点。

**再次查看Dataset资源对象状态**
```shell
$ kubectl get dataset 
NAME        UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
dataset-0   50.99MiB         0.00B    2.00GiB          0.0%                Bound   46m
dataset-1   50.99MiB         0.00B    2.00GiB          0.0%                Bound   46m
```
因为已经与一个成功启动的AlluxioRuntime绑定，该Dataset资源对象的状态得到了更新，此时`PHASE`属性值已经变为`Bound`状态。通过上述命令可以获知有关资源对象的基本信息

**查看AlluxioRuntime状态**
```shell
$ kubectl get alluxioruntime -o wide
NAME        READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE   AGE
dataset-0   1               1                 Ready          1               1                 Ready          1             1               Ready        51m
dataset-1   1               1                 Ready          1               1                 Ready          1             1               Ready        35m
```
`AlluxioRuntime`资源对象的`status`中包含了更多更详细的信息

**查看与远程文件关联的PersistentVolume以及PersistentVolumeClaim**
```shell
$ kubectl get pv
NAME                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                      STORAGECLASS   REASON   AGE
dataset-0                  100Gi      RWX            Retain           Bound    default/dataset-0                                                  50m
dataset-1                  100Gi      RWX            Retain           Bound    default/dataset-1                                                  35m
```

```shell
$ kubectl get pvc
NAME        STATUS   VOLUME      CAPACITY   ACCESS MODES   STORAGECLASS   AGE
dataset-0   Bound    dataset-0   100Gi      RWX                           50m
dataset-1   Bound    dataset-1   100Gi      RWX                           35m
```
`Dataset`资源对象准备完成后（即与Alluxio实例绑定后），与该资源对象关联的PV, PVC已经由Fluid生成，应用可以通过该PVC完成远程文件在Pod中的挂载，并通过挂载目录实现远程文件访问

## 远程文件访问

**查看待创建的应用**
```shell
$ cat<<EOF >nginx.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx0
spec:
  containers:
    - name: nginx
      image: harbor.unisound.ai/xieydd/nginx
      volumeMounts:
        - mountPath: /data
          name: dataset
  volumes:
    - name: dataset
      persistentVolumeClaim:
        claimName: dataset-0
  nodeName: 9f-03-00-05-96-fc
EOF

$ cat<<EOF >nginx1.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx1
spec:
  containers:
    - name: nginx
      image: harbor.unisound.ai/xieydd/nginx
      volumeMounts:
        - mountPath: /data
          name: dataset1
  volumes:
    - name: dataset1
      persistentVolumeClaim:
        claimName: dataset-1
  nodeName: 9f-03-00-05-96-fc
EOF
```

**启动应用进行远程文件访问**
```shell
$ kubectl create -f nginx.yaml
$ kubectl create -f nginx1.yaml
```

登录Nginx Pod:
```shell
$ kubectl exec -it nginx0 -- bash
```

查看远程文件挂载情况：
```shell
$ ls -1 /data/dataset
r--r-----. 1 831 831 48260160 May 22  2018 kernel-3.10.0-862.2.3.el7_lustre.x86_64.rpm
-r--r-----. 1 831 831  3963752 May 22  2018 kmod-lustre-2.10.4-1.el7.x86_64.rpm
-r--r-----. 1 831 831   474132 May 22  2018 kmod-lustre-osd-ldiskfs-2.10.4-1.el7.x86_64.rpm
-r--r-----. 1 831 831   708372 May 22  2018 lustre-2.10.4-1.el7.x86_64.rpm
-r--r-----. 1 831 831    41772 May 22  2018 lustre-iokit-2.10.4-1.el7.x86_64.rpm
-r--r-----. 1 831 831    14924 May 22  2018 lustre-osd-ldiskfs-mount-2.10.4-1.el7.x86_64.rpm
```

```shell
$ du -h /data/dataset/*
47M /data/dataset/kernel-3.10.0-862.2.3.el7_lustre.x86_64.rpm
3.8M  /data/dataset/kmod-lustre-2.10.4-1.el7.x86_64.rpm
464K  /data/dataset/kmod-lustre-osd-ldiskfs-2.10.4-1.el7.x86_64.rpm
692K  /data/dataset/lustre-2.10.4-1.el7.x86_64.rpm
41K /data/dataset/lustre-iokit-2.10.4-1.el7.x86_64.rpm
15K /data/dataset/lustre-osd-ldiskfs-mount-2.10.4-1.el7.x86_64.rpm
```

登出Nginx Pod:
```shell
$ exit
```

正如你所见，WebUFS上所存储的全部文件,可以和本地文件完全没有区别的方式存在于某个Pod中，并且可以被该Pod十分方便地访问

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
          image: harbor.unisound.ai/xieydd/busybox:1.28.3
          command: ["/bin/sh"]
          args: ["-c", "set -x; time cp -r /data/dataset ./"]
          volumeMounts:
            - mountPath: /data
              name: dataset
      volumes:
        - name: dataset
          persistentVolumeClaim:
            claimName: dataset-0
      nodeName: 9f-03-00-05-96-fc

EOF

$ cat<<EOF >app1.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: fluid-copy-test1
spec:
  template:
    spec:
      restartPolicy: OnFailure
      containers:
        - name: busybox
          image: harbor.unisound.ai/xieydd/busybox:1.28.3
          command: ["/bin/sh"]
          args: ["-c", "set -x; time cp -r /data/dataset ./"]
          volumeMounts:
            - mountPath: /data
              name: dataset
      volumes:
        - name: dataset
          persistentVolumeClaim:
            claimName: dataset-1
      nodeName: 9f-03-00-05-96-fc

EOF
```

**启动测试作业**
```shell
$ kubectl create -f app.yaml
job.batch/fluid-copy-test1 created
$ kubectl create -f app1.yaml
job.batch/fluid-copy-test1 created
```

该测试程序会执行`time cp -r /data/dataset ./`的shell命令，其中`/data/dataset`是远程文件在Pod中挂载的位置，该命令完成后会在终端显示命令执行的时长：

等待一段时间,待该作业运行完成,作业的运行状态可通过以下命令查看:
```shell
$ kubectl get pod | grep copy
NAME                    READY   STATUS      RESTARTS   AGE
fluid-copy-test-gcj5v               0/1     Completed   0          72s     0/0        10.99.118.4     9f-03-00-05-96-fc   <none>           <none>
fluid-copy-test1-9pw7d              0/1     Completed   0          15s     0/0        10.99.118.42    9f-03-00-05-96-fc   <none>           <none>
```
如果看到如上结果,则说明该作业已经运行完成

> 注意: `ffluid-copy-test-gcj5v`中的`gcj5v`为作业生成的标识,在你的环境中,这个标识可能不同,接下来的命令中涉及该标识的地方请以你的环境为准

**查看测试作业完成时间**
```shell
$ kubectl logs fluid-copy-test-gcj5v
+ time cp -r /data/dataset ./
real  0m 1.58s
user  0m 0.00s
sys 0m 0.14s
$ kubectl logs fluid-copy-test1-9pw7d
+ time cp -r /data/dataset ./
real  0m 1.43s
user  0m 0.00s
sys 0m 0.14s
```

可见，第一次远程文件的读取耗费了接近1.4s的时间。

**查看Dataset资源对象状态**
```shell
$ kubectl get dataset
NAME        UFS TOTAL SIZE   CACHED     CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
dataset-0   50.99MiB         50.99MiB   2.00GiB          100.0%              Bound   99m
dataset-1   50.99MiB         50.99MiB   2.00GiB          100.0%              Bound   99m
```
现在，所有远程文件都已经被缓存在了Alluxio中

**再次启动测试作业**
```shell
$ kubectl delete -f app.yaml
$ kubectl create -f app.yaml
$ kubectl delete -f app1.yaml
$ kubectl create -f app1.yaml
```

由于远程文件已经被缓存，此次测试作业能够迅速完成：
```shell
$ kubectl get pod
fluid-copy-test-d6qk9               0/1     Completed   0          20s     0/0        10.99.118.48    9f-03-00-05-96-fc   <none>           <none>
fluid-copy-test1-n97m5              0/1     Completed   0          17s     0/0        10.99.118.58    9f-03-00-05-96-fc   <none>           <none>
```

```shell
$ kubectl logs fluid-copy-test-d6qk9
+ time cp -r /data/dataset ./
real  0m 0.53s
user  0m 0.00s
sys 0m 0.12s
$ kubectl logs fluid-copy-test1-n97m5
+ time cp -r /data/dataset ./
real  0m 0.47s
user  0m 0.00s
sys 0m 0.12s
```
同样的文件访问操作仅耗费了0.47s

这种大幅度的加速效果归因于Alluxio所提供的强大的缓存能力，这种缓存能力意味着，只要你访问某个远程文件一次，该文件就会被缓存在Alluxio中，你的所有接下来的重复访问都不再需要进行远程文件读取，而是从Alluxio中直接获取数据，因此对于数据的访问加速也就不难解释了。
> 注意： 上述文件的访问速度与示例运行环境的网络条件有关，如果文件访问速度过慢，请更换更小的远程文件尝试

同样登录主机节点（如果可以）
```shell
$ ssh root@9f-03-00-05-96-fc
[root@9f-03-00-05-96-fc ~]# ls /dev/shm/default/
dataset-0    dataset-1
[root@9f-03-00-05-96-fc ~]# ls /dev/shm/default/dataset-0/alluxioworker/
100663296  16777216  33554432  50331648  67108864  83886080
[root@9f-03-00-05-96-fc ~]# ls /dev/shm/default/dataset-1/alluxioworker/
100663296  16777216  33554432  50331648  67108864  83886080
```
可以看到不同Dataset缓存的block文件根据Dataset的namespace和name进行了隔离，因为缓存相同的文件，所以block名相同。

## 环境清理
```shell
$ kubectl delete -f .
$ kubectl label node 9f-03-00-05-96-fc fluid-
```
