# 示例 - HDFS Client文件访问加速

本文介绍如何使用HDFS Client，在Fluid中通过Alluxio协议访问远程文件，并借助Alluxio的文件缓存能力，实现访问远程文件加速。

## 前提条件

在运行该示例之前，请参考[安装文档](../guide/get_started.md)完成安装，并检查Fluid各组件正常运行：

```shell
$ kubectl get pod -n fluid-system
NAME                                  READY   STATUS    RESTARTS   AGE
alluxioruntime-controller-5b64fdbbb-84pc6   1/1     Running   0          8h
csi-nodeplugin-fluid-fwgjh                  2/2     Running   0          8h
csi-nodeplugin-fluid-ll8bq                  2/2     Running   0          8h
dataset-controller-5b7848dbbb-n44dj         1/1     Running   0          8h
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
hadoop-master-0                 2/2     Running   0          106s
hadoop-worker-0                 2/2     Running   0          64s
hadoop-worker-1                 2/2     Running   0          64s
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

## 通过HDFS Client访问文件

**准备测试程序**

本例采用HDFS Java Client访问文件，在编写客户端代码时需要引入以下依赖项

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

同时要在core-site.xml中添加alluxio的配置，详情和故障排查可参考[Running Hadoop MapReduce on Alluxio](https://docs.alluxio.io/os/user/stable/en/compute/Hadoop-MapReduce.html)。

```xml
<property>
  <name>fs.alluxio.impl</name>
  <value>alluxio.hadoop.FileSystem</value>
  <description>The Alluxio FileSystem</description>
</property>
```

在通过HDFS Client访问文件时，需要指定HDFS Server地址

```java
final String HDFS_URL = "alluxio://hadoop-master-0.default.svc.cluster.local:"+ System.getenv("HADOOP_PORT") + "/hadoop";
Configuration conf = new Configuration();
FileSystem fs = FileSystem.get(URI.create(HDFS_URL), conf);
```

注意这里的HDFS_URL域名规则为: 

```shell
alluxio://{HCFS URL}/{DATASET_NAME}
```

其中DATASET_NAME为前面创建的Dataset名称，本例中为hadoop。而Endpoint的获得完全可以通过如下命令获得HCFS(Hadoop Compatible FileSystem) URL

```shell
 kubectl get datasets.data.fluid.io -owide
NAME    UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   HCFS URL                                 AGE
hbase   443.49MiB        0.00B    4.00GiB          0.0%                Bound   alluxio://hbase-master-0.default:19998   97s
```

完整的测试代码可参考[samples/hdfs](../../../samples/hdfs)。我们把测试代码制作成镜像，方便接下来的测试，镜像地址为registry.cn-hangzhou.aliyuncs.com/qiulingwei/fluid-hdfs-demo:1.2.0

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
          image: registry.cn-hangzhou.aliyuncs.com/qiulingwei/fluid-hdfs-demo:1.3.0
          imagePullPolicy: Always
          env:
          - name: HADOOP_PORT
            value: "19998"
EOF
```
此处，需要将环境变量中的19998替换为刚刚查询得到的HCFS(Hadoop Compatible FileSystem) URL中实际的端口

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


