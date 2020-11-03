# 示例 - HDFS Client文件访问加速

本文介绍如何使用HDFS Client，在Fluid中通过Alluxio协议访问远程文件，并借助Alluxio的文件缓存能力，实现访问远程文件加速。

## 前提条件

在运行该示例之前，请参考[安装文档](https://github.com/fluid-cloudnative/fluid/blob/master/docs/zh/userguide/install.md)完成安装，并检查Fluid各组件正常运行：

```shell
$ kubectl get pod -n fluid-system
NAME                                  READY   STATUS    RESTARTS   AGE
controller-manager-7fd6457ccf-jnkvn   1/1     Running   0          60s
csi-nodeplugin-fluid-6rhpt            2/2     Running   0          60s
csi-nodeplugin-fluid-6zwgl            2/2     Running   0          60s
```

## 新建工作环境

```shell
$ mkdir <any-path>/hdfs
$ cd <any-path>/hdfs
```

## 运行示例

**查看待创建的Dataset资源对象**

```shell
$ cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: hadoop
spec:
  mounts:
    - mountPoint: https://mirrors.tuna.tsinghua.edu.cn/apache/hadoop/core/current/
      name: hadoop
EOF
```

在这里，我们将要创建一个kind为`Dataset`的资源对象(Resource object)。`Dataset`是Fluid所定义的一个Custom Resource Definition(CRD)，该CRD被用来告知Fluid在哪里可以找到你所需要的数据。Fluid将该CRD对象中定义的`mountPoint`属性挂载到Alluxio之上，因此该属性可以是任何合法的能够被Alluxio识别的UFS地址。在本示例中，为了简单，我们使用[WebUFS](https://docs.alluxio.io/os/user/stable/cn/ufs/WEB.html)进行演示。

**创建Dataset资源对象**

```shell
$ kubectl create -f dataset.yaml
dataset.data.fluid.io/hadoop created
```

**查看Dataset资源对象状态**

```shell
$ kubectl get dataset hadoop
NAME     UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE      AGE
hadoop                                                                  NotBound   1m
```

如上所示，`status`中的`phase`属性值为`NotBound`，这意味着该`Dataset`资源对象目前还未与任何`AlluxioRuntime`资源对象绑定，接下来，我们将创建一个`AlluxioRuntime`资源对象。

**查看待创建的AlluxioRuntime资源对象**

```shell
$ cat<<EOF >runtime.yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: hadoop
spec:
  replicas: 2
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
```

**创建AlluxioRuntime资源对象**

```shell
$ kubectl apply -f runtime.yaml
alluxioruntime.data.fluid.io/hadoop created
```

**检查AlluxioRuntime资源对象状态**

`AlluxioRuntime`是另一个Fluid定义的CRD。一个`AlluxioRuntime`资源对象描述了在Kubernetes集群中运行一个Alluxio实例所需要的配置信息。

等待一段时间，让AlluxioRuntime资源对象中的各个组件得以顺利启动，你会看到类似以下状态：

```shell
$ hdfs kubectl get pods
NAME                            READY   STATUS    RESTARTS   AGE
hadoop-fuse-749fs               1/1     Running   0          64s
hadoop-fuse-khdrb               1/1     Running   0          64s
hadoop-master-0                 2/2     Running   0          106s
hadoop-worker-cn9fg             2/2     Running   0          64s
hadoop-worker-tlldq             2/2     Running   0          64s
```

```shell
$ kubectl get alluxioruntime hadoop
NAME     MASTER PHASE   WORKER PHASE   FUSE PHASE   AGE
hadoop   Ready          Ready          Ready        116s
```

然后，再查看Dataset状态，发现已经与AlluxioRuntime绑定。

```shell
$ kubectl get dataset hadoop
NAME     UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
hadoop   390.2MiB         0B       4GiB             0%                  Bound   55m
```


Dataset资源对象准备完成后（即与Alluxio实例绑定后），与该资源对象关联的PV, PVC已经由Fluid生成，应用可以通过该PVC完成远程文件在Pod中的挂载，并通过挂载目录实现远程文件访问

## Access files through HDFS Client

**Prepare test program**

This example uses HDFS Java Client to access files, and the following dependencies need to be introduced when writing client code

```xml
<dependency>
  <groupId>org.apache.hadoop</groupId>
  <artifactId>hadoop-client</artifactId>
  <version>${hadoop.version}</version>
</dependency>
<dependency>
  <groupId>org.apache.hadoop</groupId>
  <artifactId>hadoop-hdfs</artifactId>
  <version>${hadoop.version}</version>
</dependency>
<dependency>
  <groupId>org.alluxio</groupId>
  <artifactId>alluxio-core-client</artifactId>
  <version>${alluxio.version}</version>
  <type>pom</type>
</dependency>
<dependency>
  <groupId>org.alluxio</groupId>
  <artifactId>alluxio-core-client-hdfs</artifactId>
  <version>${alluxio.version}</version>
</dependency>
```

At the same time, add the configuration of alluxio in core-site.xml. For details and troubleshooting, please refer to [Running Hadoop MapReduce on Alluxio](https://docs.alluxio.io/os/user/stable/en/compute/Hadoop-MapReduce.html)。

```xml
<property>
  <name>fs.alluxio.impl</name>
  <value>alluxio.hadoop.FileSystem</value>
  <description>The Alluxio FileSystem</description>
</property>
```

When accessing files through the HDFS client, you need to specify the HDFS server address

```java
final String HDFS_URL = "alluxio://hadoop-master-0.default:19998/hadoop"
Configuration conf = new Configuration();
FileSystem fs = FileSystem.get(URI.create(HDFS_URL), conf);
```

Note that the HDFS_URL domain name rule here is:

```shell
alluxio://{HCFS URL}/{DATASET_NAME}
```

Where DATASET_NAME is the name of the Dataset created earlier, in this case it is hadoop. The Endpoint can be obtained through the following command to obtain the HCFS (Hadoop Compatible FileSystem) URL

```shell
 kubectl get datasets.data.fluid.io -owide
NAME    UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   HCFS URL                                 AGE
hbase   443.49MiB        0.00B    4.00GiB          0.0%                Bound   alluxio://hbase-master-0.default:19998   97s
```

For the complete test code, please refer to [samples/hdfs](../../../samples/hdfs). We made the test code into a mirror to facilitate the next test. The mirror address is registry.cn-beijing.aliyuncs.com/yukong/fluid-hdfs-demo:1.0.0.

**查看待创建的测试作业**

```shell
$ cat<<EOF >app.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: fluid-hdfs-demo
spec:
  template:
    spec:
      restartPolicy: OnFailure
      containers:
        - name: fluid-hdfs-demo
          image: registry.cn-beijing.aliyuncs.com/yukong/fluid-hdfs-demo:1.0.0 #可以替换为自己构建的镜像
          imagePullPolicy: Always
EOF
```

**启动测试作业**

```shell
$ kubectl apply -f app.yaml
job.batch/fluid-hdfs-demo created
```

在测试程序中我们先遍历Dataset查看有哪些文件，然后把这些文件复制到本地，查看访问远程文件的加速效果。

等待一段时间,待该作业运行完成,作业的运行状态可通过以下命令查看:

```shell
$ kubectl get pods
NAME                            READY   STATUS      RESTARTS   AGE
fluid-hdfs-demo-8q9b7           0/1     Completed   0          14m
```

**查看任务执行时间**

```shell
$ kubectl logs fluid-hdfs-demo-8q9b7
## RELEASENOTES.md
## hadoop-3.1.3-src.tar.gz
## CHANGES.md
## hadoop-3.1.3-site.tar.gz
## hadoop-3.1.3-rat.txt
## hadoop-3.1.3.tar.gz
copy directory cost:67520ms
```

第一次执行作业，耗时67秒多时间。

**查看Dataset资源对象状态**

```shell
$ kubectl get dataset hadoop
NAME     UFS TOTAL SIZE   CACHED     CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
hadoop   390.2MiB         388.4MiB   4GiB             99%                 Bound   88m
```

可以看到所有远程文件已经被缓存在Alluxio中了。

**再次启动测试作业**

```shell
$ kubectl delete -f app.yaml
$ kubectl create -f app.yaml
```

由于远程文件都已经被缓存，这次作业耗时大大减少。

```shell
$ kubectl logs fluid-hdfs-demo-pxt45
## RELEASENOTES.md
## hadoop-3.1.3-src.tar.gz
## CHANGES.md
## hadoop-3.1.3-site.tar.gz
## hadoop-3.1.3-rat.txt
## hadoop-3.1.3.tar.gz
copy directory cost:1300ms
```

可以看到第二次作业，同样的文件访问，仅耗时1.3秒。

这种大幅度的加速效果归因于Alluxio所提供的强大的缓存能力，这种缓存能力意味着，只要你访问某个远程文件一次，该文件就会被缓存在Alluxio中，你的所有接下来的重复访问都不再需要进行远程文件读取，而是从Alluxio中直接获取数据，因此对于数据的访问加速也就不难解释了。

## 环境清理

```shell
$ kubectl delete -f .
```


