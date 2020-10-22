# 示例 - 单机多Dataset远程Web文件访问加速
通过[Alluxio](https://www.alluxio.io)和[Fuse](https://github.com/libfuse/libfuse)，Fluid为用户提供了一种更为简单的文件访问接口，使得任意运行在Kubernetes集群上的程序能够像访问本地文件一样轻松访问存储在远程文件系统中的文件。Fluid 针对数据集进行全生命周期的管理和隔离，尤其对于短生命周期应用（e.g 数据分析任务、机器学习任务），用户可以在集群中大规模部署。

本文档通过一个简单的例子演示了上述功能特性

## 前提条件
在运行该示例之前，请参考[安装文档](../userguide/install.md)完成安装，并检查Fluid各组件正常运行：
```shell
$ kubectl get pod -n fluid-system
NAME                                  READY   STATUS    RESTARTS   AGE
controller-manager-74d4f88b55-x99r8   1/1     Running   0          17h
csi-nodeplugin-fluid-brxph            2/2     Running   0          17h
csi-nodeplugin-fluid-gfcxm            2/2     Running   0          17h
```
通常来说，你会看到一个名为`controller-manager`的Pod和多个名为`csi-nodeplugin`的Pod正在运行。其中，`csi-nodeplugin`这些Pod的数量取决于你的Kubernetes集群中结点的数量。

## 运行示例

**对某个节点打标签** 
```shell
$ kubectl  label node cn-beijing.192.168.1.174 fluid=multi-dataset
```
> 在接下来的步骤中，我们将使用 `NodeSelector` 来管理Dataset调度的节点，这里仅做试验使用。 

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
  name: spark
spec:
  mounts:
    - mountPoint: https://mirror.bit.edu.cn/apache/spark/spark-3.0.1/
      name: spark
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
> 注意: 上述mountPoint中使用了Apache清华镜像源进行演示，如果你的环境位于海外，请更换为`https://downloads.apache.org/hbase/2.2.5/`和`https://downloads.apache.org/spark/spark-3.0.1/`进行尝试

在这里，我们将要创建一个kind为`Dataset`的资源对象(Resource object)。`Dataset`是Fluid所定义的一个Custom Resource Definition(CRD)，该CRD被用来告知Fluid在哪里可以找到你所需要的数据。Fluid将该CRD对象中定义的`mountPoint`属性挂载到Alluxio之上，因此该属性可以是任何合法的能够被Alluxio识别的UFS地址。在本示例中，为了简单，我们使用[WebUFS](https://docs.alluxio.io/os/user/stable/cn/ufs/WEB.html)进行演示。

更多有关UFS的信息，请参考[Alluxio文档-底层存储系统](https://docs.alluxio.io/os/user/stable/cn/ufs/OSS.html)部分。

> 本示例将以Apache镜像站点上的Hbase v2.25和Spark v3.0.1相关资源作为演示中使用的远程文件。这个选择并没有任何特殊之处，你可以将这个远程文件修改为任意你喜欢的远程文件。但是，如果你想要和我们一样使用WebUFS进行操作的话，最好还是选择一个Apache镜像源站点( e.g. [清华镜像源](https://mirrors.tuna.tsinghua.edu.cn/apache) )，因为根据目前WebUFS的实现，如果你选择其他更加复杂的网页作为WebUFS，你可能需要进行更多[更复杂的配置](https://docs.alluxio.io/os/user/stable/cn/ufs/WEB.html#%E9%85%8D%E7%BD%AEalluxio)

**创建Dataset资源对象**
```shell
$ kubectl apply -f dataset.yaml
dataset.data.fluid.io/hbase created
$ kubectl apply -f dataset1.yaml
dataset.data.fluid.io/spark created
```

**查看Dataset资源对象状态**
```shell
$ kubectl get dataset
NAME    UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE      AGE
hbase                                                                 NotBound    30s
spark                                                                 NotBound    27s
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
  replicas: 1
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

$ cat<<EOF >runtime1.yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: spark
spec:
  replicas: 1
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

**创建AlluxioRuntime资源对象**
```shell
$ kubectl create -f runtime.yaml
alluxioruntime.data.fluid.io/hbase created

# 注意等待 Dataset hbase 全部组件 Running 
$ kubectl get pod -o wide | grep fuse
NAME                 READY   STATUS    RESTARTS   AGE   IP              NODE                       NOMINATED NODE   READINESS GATES
hbase-fuse-56mmc     1/1     Running   0          26s   192.168.1.174   cn-beijing.192.168.1.174   <none>           <none>
hbase-master-0       2/2     Running   0          59s   192.168.1.175   cn-beijing.192.168.1.175   <none>           <none>
hbase-worker-bzlkn   2/2     Running   0          26s   192.168.1.174   cn-beijing.192.168.1.174   <none>           <none>

$ kubectl create -f runtime1.yaml
alluxioruntime.data.fluid.io/spark created
```

**检查AlluxioRuntime资源对象是否已经创建**
```shell
$ kubectl get alluxioruntime
NAME    MASTER PHASE   WORKER PHASE   FUSE PHASE   AGE
hbase   Ready          Ready          Ready        12m
spark   Ready          Ready          Ready        77s
```

`AlluxioRuntime`是另一个Fluid定义的CRD。一个`AlluxioRuntime`资源对象描述了在Kubernetes集群中运行一个Alluxio实例所需要的配置信息。

等待一段时间，让AlluxioRuntime资源对象中的各个组件得以顺利启动，你会看到类似以下状态：
```shell
$ kubectl get pod | grep hbase 
NAME                 READY   STATUS    RESTARTS   AGE   IP              NODE                       NOMINATED NODE   READINESS GATES
hbase-fuse-56mmc     1/1     Running   0          11m   192.168.1.174   cn-beijing.192.168.1.174   <none>           <none>
hbase-master-0       2/2     Running   0          11m   192.168.1.175   cn-beijing.192.168.1.175   <none>           <none>
hbase-worker-bzlkn   2/2     Running   0          11m   192.168.1.174   cn-beijing.192.168.1.174   <none>           <none>
spark-fuse-5zxmf     1/1     Running   0          11s   192.168.1.174   cn-beijing.192.168.1.174   <none>           <none>
spark-master-0       2/2     Running   0          43s   192.168.1.175   cn-beijing.192.168.1.175   <none>           <none>
spark-worker-cksdm   2/2     Running   0          11s   192.168.1.174   cn-beijing.192.168.1.174   <none>           <none>
```
注意上面的不同的 Dataset 的 worker 和 fuse 组件可以正常的调度到相同的节点 `cn-beijing.192.168.1.174`。

**再次查看Dataset资源对象状态**
```shell
$ kubectl get dataset 
NAME    UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
hbase   443.49MiB        0.00B    2.00GiB          0.0%                Bound   14m
spark   998.22MiB        0.00B    2.00GiB          0.0%                Bound   113s
```
因为已经与一个成功启动的AlluxioRuntime绑定，该Dataset资源对象的状态得到了更新，此时`PHASE`属性值已经变为`Bound`状态。通过上述命令可以获知有关资源对象的基本信息

**查看AlluxioRuntime状态**
```shell
$ kubectl get alluxioruntime -o wide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE   AGE
hbase   1               1                 Ready          1               1                 Ready          1             1               Ready        12m
spark   1               1                 Ready          1               1                 Ready          1             1               Ready        2m5s
```
`AlluxioRuntime`资源对象的`status`中包含了更多更详细的信息

**查看与远程文件关联的PersistentVolume以及PersistentVolumeClaim**
```shell
$ kubectl get pv
NAME    CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM           STORAGECLASS   REASON   AGE
hbase   100Gi      RWX            Retain           Bound    default/hbase                           12m
spark   100Gi      RWX            Retain           Bound    default/spark                           97s
```

```shell
$ kubectl get pvc
NAME    STATUS   VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS   AGE
hbase   Bound    hbase    100Gi      RWX                           12m
spark   Bound    spark    100Gi      RWX                           99s
```
`Dataset`资源对象准备完成后（即与Alluxio实例绑定后），与该资源对象关联的PV, PVC已经由Fluid生成，应用可以通过该PVC完成远程文件在Pod中的挂载，并通过挂载目录实现远程文件访问

## 远程文件访问

**查看待创建的应用**
```shell
$ cat<<EOF >nginx.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-hbase
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
  nodeName: cn-beijing.192.168.1.174

EOF

$ cat<<EOF >nginx1.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-spark
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
        claimName: spark
  nodeName: cn-beijing.192.168.1.174

EOF
```

**启动应用进行远程文件访问**
```shell
$ kubectl create -f nginx.yaml
$ kubectl create -f nginx1.yaml
```

登录Nginx hbase Pod:
```shell
$ kubectl exec -it nginx-hbase -- bash
```

查看远程文件挂载情况：
```shell
$ ls -lh /data/dataset
total 444M
-r--r----- 1 root root 174K May 26 07:30 CHANGES.md
-r--r----- 1 root root 106K May 26 07:30 RELEASENOTES.md
-r--r----- 1 root root 115K May 26 07:30 api_compare_2.2.5RC0_to_2.2.4.html
-r--r----- 1 root root 211M May 26 07:30 hbase-2.2.5-bin.tar.gz
-r--r----- 1 root root 200M May 26 07:30 hbase-2.2.5-client-bin.tar.gz
-r--r----- 1 root root  34M May 26 07:30 hbase-2.2.5-src.tar.gz
```

登录Nginx spark Pod:
```shell
$ kubectl exec -it nginx-spark -- bash
```

查看远程文件挂载情况：
```shell
$ ls -lh /data/spark/
total 999M
-r--r----- 1 root root 322K Aug 28 09:25 SparkR_3.0.1.tar.gz
-r--r----- 1 root root 195M Aug 28 09:25 pyspark-3.0.1.tar.gz
-r--r----- 1 root root 210M Aug 28 09:25 spark-3.0.1-bin-hadoop2.7-hive1.2.tgz
-r--r----- 1 root root 210M Aug 28 09:25 spark-3.0.1-bin-hadoop2.7.tgz
-r--r----- 1 root root 214M Aug 28 09:25 spark-3.0.1-bin-hadoop3.2.tgz
-r--r----- 1 root root 150M Aug 28 09:25 spark-3.0.1-bin-without-hadoop.tgz
-r--r----- 1 root root  22M Aug 28 09:25 spark-3.0.1.tgz

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
  name: fluid-copy-test-hbase
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
      nodeName: cn-beijing.192.168.1.174

EOF

$ cat<<EOF >app1.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: fluid-copy-test-spark
spec:
  template:
    spec:
      restartPolicy: OnFailure
      containers:
        - name: busybox
          image: busybox
          command: ["/bin/sh"]
          args: ["-c", "set -x; time cp -r /data/spark ./"]
          volumeMounts:
            - mountPath: /data
              name: spark-vol
      volumes:
        - name: spark-vol
          persistentVolumeClaim:
            claimName: spark
      nodeName: cn-beijing.192.168.1.174

EOF
```

**启动测试作业**
```shell
$ kubectl create -f app.yaml
job.batch/fluid-copy-test-hbase created
$ kubectl create -f app1.yaml
job.batch/fluid-copy-test-spark created
```

hbase任务程序会执行`time cp -r /data/hbase ./`的shell命令，其中`/data/hbase`是远程文件在Pod中挂载的位置，该命令完成后会在终端显示命令执行的时长。

spark任务程序会执行`time cp -r /data/spark ./`的shell命令，其中`/data/spark`是远程文件在Pod中挂载的位置，该命令完成后会在终端显示命令执行的时长。

等待一段时间,待该作业运行完成,作业的运行状态可通过以下命令查看:
```shell
$ kubectl get pod -o wide | grep copy 
fluid-copy-test-hbase-gnqfb   0/1     Completed   0          2m50s   172.25.0.20     cn-beijing.192.168.1.174   <none>           <none>
fluid-copy-test-spark-bprqh   0/1     Completed   0          2m47s   172.25.0.21     cn-beijing.192.168.1.174   <none>           <none>
```
如果看到如上结果,则说明该作业已经运行完成

> 注意: `luid-copy-test-hbase-gnqfb `中的`gnqfb`为作业生成的标识,在你的环境中,这个标识可能不同,接下来的命令中涉及该标识的地方请以你的环境为准

**查看测试作业完成时间**
```shell
$ kubectl  logs fluid-copy-test-hbase-gnqfb
+ time cp -r /data/hbase ./
real	0m 59.84s
user	0m 0.00s
sys	0m 1.43s
$ kubectl  logs fluid-copy-test-spark-bprqh
+ time cp -r /data/spark ./
real	1m 52.88s
user	0m 0.00s
sys	0m 3.12s
```

可见，第一次远程文件的读取hbase耗费了接近60s的时间，读取spark耗费接近1m53s时间。

**查看Dataset资源对象状态**
```shell
$ kubectl get dataset
NAME    UFS TOTAL SIZE   CACHED      CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
hbase   443.49MiB        443.49MiB   2.00GiB          100.0%              Bound   27m
spark   998.22MiB        998.22MiB   2.00GiB          100.0%              Bound   14m
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
$ kubectl get pod | grep fluid
fluid-copy-test-hbase-5lvzs   0/1     Completed   0          24s   172.25.0.22     cn-beijing.192.168.1.174   <none>           <none>
fluid-copy-test-spark-rvclt   0/1     Completed   0          19s   172.25.0.23     cn-beijing.192.168.1.174   <none>           <none>
```

```shell
$ kubectl  logs fluid-copy-test-hbase-5lvzs
+ time cp -r /data/hbase ./
real	0m 0.39s
user	0m 0.00s
sys	0m 0.38s
$ kubectl  logs fluid-copy-test-spark-rvclt
+ time cp -r /data/spark ./
real	0m 0.87s
user	0m 0.00s
sys	0m 0.86s

```
同样的文件访问操，hbase仅耗费了0.39s，spark仅耗费了0.87s。

这种大幅度的加速效果归因于Alluxio所提供的强大的缓存能力，这种缓存能力意味着，只要你访问某个远程文件一次，该文件就会被缓存在Alluxio中，你的所有接下来的重复访问都不再需要进行远程文件读取，而是从Alluxio中直接获取数据，因此对于数据的访问加速也就不难解释了。
> 注意： 上述文件的访问速度与示例运行环境的网络条件有关，如果文件访问速度过慢，请更换更小的远程文件尝试

同样登录主机节点（如果可以）
```shell
$ ssh root@192.168.1.174
$ ls /dev/shm/default/
hbase  spark
$ ls -lh /dev/shm/default/hbase/alluxioworker/
总用量 444M
-rwxrwxrwx 1 root root 174K 10月 22 18:45 100663296
-rwxrwxrwx 1 root root 115K 10月 22 18:45 16777216
-rwxrwxrwx 1 root root 200M 10月 22 18:44 33554432
-rwxrwxrwx 1 root root 106K 10月 22 18:44 50331648
-rwxrwxrwx 1 root root 211M 10月 22 18:45 67108864
-rwxrwxrwx 1 root root  34M 10月 22 18:45 83886080
$ ls -lh /dev/shm/default/spark/alluxioworker/
总用量 999M
-rwxrwxrwx 1 root root 150M 10月 22 18:46 100663296
-rwxrwxrwx 1 root root 322K 10月 22 18:46 117440512
-rwxrwxrwx 1 root root 210M 10月 22 18:45 16777216
-rwxrwxrwx 1 root root 210M 10月 22 18:46 33554432
-rwxrwxrwx 1 root root 195M 10月 22 18:45 50331648
-rwxrwxrwx 1 root root 214M 10月 22 18:45 67108864
-rwxrwxrwx 1 root root  22M 10月 22 18:45 83886080

```
可以看到不同Dataset缓存的block文件根据Dataset的namespace和name进行了隔离。

## 环境清理
```shell
$ kubectl delete -f .
$ kubectl label node cn-beijing.192.168.1.174 fluid-
```
