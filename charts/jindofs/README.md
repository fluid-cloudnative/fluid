# 使用 Kubernetes 部署 JindoFS



<a name="Dm24U"></a>
# 一、快速安装
<a name="L0ARG"></a>
## 1、创建kubernetes集群


![image.png](https://smartdata-binary.oss-cn-shanghai.aliyuncs.com/docs/docs-001.png)
<a name="Rmizg"></a>
#### 集群配置

- 实例规格要求内存不小于16GB。
- JindoFS使用集群内的数据盘存储缓存数据。生产环境上，建议挂载一个以上容量较大的数据盘。本文中，我们直接使用默认配置，使用单个系统盘（仅限开发测试使用）。

![image.png](https://smartdata-binary.oss-cn-shanghai.aliyuncs.com/docs/docs-002.png)

- 绑定弹性公网IP，或者用户有其它跳板机，可以ssh登录到集群中的其中一个节点。

<br />
<a name="oiYMG"></a>
## 2、安装JindoFS服务
<a name="JqK9T"></a>
### 2.1 前往容器服务->应用目录，进入“JindoFS”安装配置页面。
![image.png](https://smartdata-binary.oss-cn-shanghai.aliyuncs.com/docs/docs-003.png)
<a name="vdrsi"></a>
### 2.2 配置参数
完整的配置模板如下：<br />这里举例的镜像资源是cn-shanghai区的地址，版本号为3.3.5。**请使用安装配置页面展示的默认的image地址和tag版本。**

```yaml
image: registry.cn-shanghai.aliyuncs.com/jindofs/smartdata
imageTag: "3.3.5"
imagePullPolicy: Always
fuseImage: registry.cn-shanghai.aliyuncs.com/jindofs/jindo-fuse
fuseImageTag: "3.3.5"
user: 0
group: 0
fsGroup: 0
useHostNetwork: true
useHostPID: true
properties:
 logDir: /mnt/disk1/bigboot/log
master:
 replicaCount: 1
 resources:
   limits:
     cpu: "4"
     memory: "16G" # increase memory corresponding to filelet(blocklet) cache size
   requests:
     cpu: "1"
     memory: "1G"
 nodeSelector: {}
 properties:
   namespace.rpc.port: 8101
   namespace.meta-dir: /mnt/disk1/bigboot/server
   namespace.filelet.cache.size: 100000
   namespace.blocklet.cache.size: 1000000
   namespace.backend.type: rocksdb
   jfs.namespaces: test
   jfs.namespaces.test.mode :  cache
   jfs.namespaces.test.oss.uri :  oss://chengli-sh-test.oss-cn-shanghai-internal.aliyuncs.com/fuyue/k8s_c1
   jfs.namespaces.test.oss.access.key :  xx
   jfs.namespaces.test.oss.access.secret :  xx
worker:
 resources:
   limits:
     cpu: "4"
     memory: "8G" # increase memory corresponding to the number of concurrent reading/writing files
   requests:
     cpu: "1"
     memory: "1G"
 nodeSelector: {}
 properties:
   storage.rpc.port: 6101
   storage.data-dirs: /mnt/disk1/bigboot, /mnt/disk2/bigboot, /mnt/disk3/bigboot
   storage.temp-data-dirs: /mnt/disk1/bigboot/tmp
   storage.watermark.high.ratio: 0.4
   storage.watermark.low.ratio: 0.2
   storage.data-dirs.capacities: 80g,80g,80g
   storage.meta-dir: /mnt/disk1/bigboot/bignode
fuse:
 args:
 hostPath: /mnt/jfs
 properties:
   client.storage.rpc.port: 6101
   client.oss.retry: 5
   client.oss.upload.threads: 4
   client.oss.upload.queue.size: 5
   client.oss.upload.max.parallelism: 16
   client.oss.timeout.millisecond: 30000
   client.oss.connection.timeout.millisecond: 3000
mounts:
  master:
    # 1: /mnt/disk1
  workersAndClients:
    # 1: /mnt/disk1
    # 2: /mnt/disk2
    # 3: /mnt/disk3
```


- 配置OSS Bucket和AK，参考文档使用[JFS Scheme](https://help.aliyun.com/document_detail/164209.html#title-s6x-h1p-cg4)的部署方式。我们需要修改以下配置项：

```yaml
   jfs.namespaces: test
   jfs.namespaces.test.mode :  cache
   jfs.namespaces.test.oss.uri :  oss://chengli-sh-test.oss-cn-shanghai-internal.aliyuncs.com/fuyue/k8s_c1
   jfs.namespaces.test.oss.access.key :  xx
   jfs.namespaces.test.oss.access.secret :  xx
```

通过这些配置项，我们创建了一个名为test的命名空间，指向了chengli-sh-test这个OSS bucket的fuyue/k8s_c1目录。后续我们通过JindoFS操作test命名空间的时候，就等同于操作该OSS目录。<br />

- 其余配置保持默认值。

本文中，我们直接使用默认值进行演示。<br />生产环境上，建议参考章节“配置数据目录”，持久化缓存数据。<br />更多配置（如磁盘水位、性能相关）请参考文档[JindoFS使用说明](https://help.aliyun.com/document_detail/164209.html)。<br />

<a name="1nFTk"></a>
### 3.3 安装服务
![image.png](https://smartdata-binary.oss-cn-shanghai.aliyuncs.com/docs/docs-004.png)<br />

<a name="B6j2g"></a>
## 3. 验证安装成功

```
# kubectl get pods
NAME                               READY   STATUS      RESTARTS   AGE
jindofs-fuse-267vq                 1/1     Running     0          143m
jindofs-fuse-8qwdv                 1/1     Running     0          143m
jindofs-fuse-v6q7r                 1/1     Running     0          143m
jindofs-master-0                   1/1     Running     0          143m
jindofs-worker-mncqd               1/1     Running     0          143m
jindofs-worker-pk7j4               1/1     Running     0          143m
jindofs-worker-r2k99               1/1     Running     0          143m
```

在宿主机上访问/mnt/jfs/目录，即等同于访问JindoFS的文件

```bash
# ls /mnt/jfs/test/
15885689452274647042-0  17820745254765068290-0  entrypoint.sh
```

<a name="65H9t"></a>
# 二、体验JindoFS加速服务
通过上一章节的操作，我们成功启动了JindoFS集群。这一章节，我们将用各种方式来访问JindoFS集群，并体验JindoFS带来的加速效果。<br />JindoFS提供了Hadoop FileSystem接口、FUSE两种客户端来访问JindoFS集群，在容器环境下，同样支持这两种方式。<br />

<a name="TJxEV"></a>
## 方式1：在宿主机访问JindoFS集群（FUSE方式）
通过ssh登录到集群的一个节点，在Console内直接访问JindoFS集群<br />ls 目录

```bash
# ls /mnt/jfs/test/
15885689452274647042-0  17820745254765068290-0  entrypoint.sh  derby.log
```

读取文件

```bash
[hadoop@emr-header-1 ~]$ cat /mnt/jfs/test/derby.log
Thu Mar 05 23:25:53 CST 2020:
Booting Derby version The Apache Software Foundation - Apache Derby - 10.12.1.1 - (1704137)
```

<br />
<a name="rXyRe"></a>
## 方式2：在pod内访问JindoFS集群（FUSE方式）
为了演示，我们创建一个临时的pod（您也可以创建Deployment），然后将fuse目录挂载到pod内。首先，创建文件 jindofs-demo-app.yaml

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: jindofs-demo-app
  labels:
    app: jindofs-demo-app
spec:
  containers:
    - name: jindofs-demo-app
      image: alpine:3.7
      command: ["tail"]
      args: ["-f", "/dev/null"]
      volumeMounts:
        - name: jindofs-fuse-mount
          mountPath: /jfs
  volumes:
    - name: jindofs-fuse-mount
      hostPath:
        path: /mnt/jfs
        type: DirectoryOrCreate
```

执行命令

```bash
kubectl apply -f jindofs-demo-app.yaml
```

进入pod内的shell终端

```bash
kubectl exec -it jindofs-demo-app sh
```

在pod内读写JindoFS文件

```bash
/ # ls /jfs/test/
15885689452274647042-0  17820745254765068290-0  LICENSE.txt             bigboot.cfg             entrypoint.sh           fuse.cfg
/ # echo hello world > /jfs/test/my_first_file
/ # cat /jfs/test/my_first_file
hello world
```

这种方式适合于容器内跑机器学习计算。只需要将训练数据预先放到oss bucket上，然后在容器内读取/jfs目录下的文件即可。<br />

<a name="hK3YM"></a>
## 方式3：在pod内访问JindoFS集群（FileSystem接口）
这种方式，我们需要准备一个带有JindoFS SDK包的Hadoop Client、MapReduce、Hive或Spark镜像。请参考章节《制作带有JindoFS SDK的镜像》。<br />为了演示，我们直接使用jindo-fuse镜像（里面包含了Hadoop Client和JindoFS SDK)。<br />首先，我们创建一个临时的pod（您也可以创建Deployment），先创建文件 jindofs-demo-app2.yaml

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: jindofs-demo-app2
  labels:
    app: jindofs-demo-app2
spec:
    hostNetwork: true
    dnsPolicy: ClusterFirstWithHostNet
    containers:
    - name: jindofs-demo-app2
      image: "registry.cn-shanghai.aliyuncs.com/jindofs/jindo-fuse:3.3.5"
      command: ["tail"]
      args: ["-f", "/dev/null"]
      env:
      - name: CLIENT_NAMESPACE_RPC_ADDRESS
        value: jindofs-master:8101
      - name: CLIENT_STORAGE_RPC_PORT
        value: "6101"
      - name: CLIENT_STORAGE_RPC_HOST
        valueFrom:
          fieldRef:
            fieldPath: status.hostIP
      - name: JFS_CACHE_DATA_CACHE_ENABLE
        value: "1"
      volumeMounts:
      - mountPath: /etc/localtime
        name: volume-localtime
      - name: datavolume-1
        mountPath: "/mnt/disk1"
      - name: datavolume-2
        mountPath: "/mnt/disk2"
      - name: datavolume-3
        mountPath: "/mnt/disk3"
    volumes:
    - hostPath:
        path: /etc/localtime
        type: ''
      name: volume-localtime
    - hostPath:
        path:  "/mnt/disk1"
        type: DirectoryOrCreate
      name: datavolume-1
    - hostPath:
        path:  "/mnt/disk2"
        type: DirectoryOrCreate
      name: datavolume-2
    - hostPath:
        path:  "/mnt/disk3"
        type: DirectoryOrCreate
      name: datavolume-3

```

执行命令

```bash
kubectl apply -f jindofs-demo-app2.yaml
```

进入pod内的shell终端

```bash
kubectl exec -it jindofs-demo-app2 bash
```

在pod内读写JindoFS文件

```bash
root@iZuf6:/# hadoop fs -ls jfs://test/
Found 7 items
-rw-rw-rw-   1       3807 2020-06-18 12:31 jfs://test/15885689452274647042-0
-rw-rw-rw-   1         79 2020-06-28 15:46 jfs://test/17820745254765068290-0
-rw-rw-rw-   1     150569 2020-07-02 12:01 jfs://test/LICENSE.txt
-rw-rw-rw-   1        306 2020-07-02 11:58 jfs://test/bigboot.cfg
-rw-rw-rw-   1         79 2020-06-29 03:32 jfs://test/entrypoint.sh
-rw-rw-rw-   1        360 2020-07-07 11:31 jfs://test/fuse.cfg
-rw-rw-rw-   1         12 2020-07-16 03:16 jfs://test/my_first_file

# 您还可以尝试使用hadoop fs -put和-get读写文件
```

这种方式适合于容器内跑MapReduce、Hive或Spark等大数据计算任务。
<a name="5OyWp"></a>
## 方式4：在宿主机或其它机器访问JindoFS集群（FileSystem接口）
将下面环境变量添加到/etc/profile文件中<br />export B2SDK_CONF_DIR=/root/b2sdk/conf<br />创建文件 /root/b2sdk/conf/bigboot.cfg  包含以下主要内容

```
[bigboot]
logger.dir = /tmp/bigboot-log

[bigboot-client]
client.storage.rpc.port=6101
client.namespace.rpc.address=192.168.0.xxx:8101
client.oss.retry=5
client.oss.upload.threads=4
client.oss.upload.queue.size=5
client.oss.upload.max.parallelism=16
client.read.blocklet.prefetch.count=16
client.oss.timeout.millisecond=30000
client.oss.connection.timeout.millisecond=3000
client.read.cache-on-read.enable=1
```

主要修改client.namespace.rpc.address为master服务所在节点的ip<br />
<br />下载b2sdk包进行解压

```
wget https://smartdata-binary.oss-cn-shanghai.aliyuncs.com/b2sdk-3.3.5.tar.gz
tar -zxvf b2sdk-3.3.5.tar.gz
```

将JindoFS SDK的jar包复制到spark目录下

```
cp sdk/lib/jindofs-sdk-*.jar /usr/lib/spark-current/jars/
```

将JindoFS SDK的jar包复制到hive目录下<br />

```
cp sdk/lib/jindofs-sdk-*.jar /usr/lib/hive-current/lib/
```

将JindoFS SDK的jar包复制到Hadoop目录下<br />

```
cp sdk/lib/jindofs-sdk-*.jar /usr/lib/hadoop-current/share/hadoop/common/lib/
```

如果您跑的是Presto作业，那么将JindoFS SDK的jar包复制到Presto目录下

```
/usr/lib/presto-current/plugin/hive-hadoop2/
```

<br />**如果您搭建的是多节点spark、hive集群，那么集群的所有节点都要进行上述配置。**

<a name="RLiWy"></a>
# 三、进阶配置
<a name="eGHfX"></a>
## 3.1 数据持久化、数据locality
通过数据持久化我们可以使得缓存在pod重建后依然有效。JindoFS会维护缓存数据的生命周期，进行自动清理，您无需担心数据持久化后缓存遗留问题。<br />通过实现数据locality，缓存才能写入到本地磁盘，并且读取缓存时如果命中本地节点可以提高读取性能。<br />我们需要以下配置：

- 网络上使用宿主机网络

```yaml
    hostNetwork: true
    dnsPolicy: ClusterFirstWithHostNet
```

- 容器内路径，建议集群配置保持以下惯例

```yaml
properties:
 logDir: /mnt/disk1/bigboot/log

 master:
 properties:
   namespace.meta-dir: /mnt/disk1/bigboot/server

worker:
 properties:
   storage.data-dirs: /mnt/disk1/bigboot, /mnt/disk2/bigboot, /mnt/disk3/bigboot
   storage.temp-data-dirs: /mnt/disk1/bigboot/tmp
   storage.data-dirs.capacities: 80g,80g,80g
   storage.meta-dir: /mnt/disk1/bigboot/bignode
```

a. logDir、 namespace.meta-dir、 storage.meta-dir放置在容器内/mnt/disk1路径下，建议保持默认值。<br />
b. 如果宿主机有3块盘，那么分别用volume方式映射到容器内/mnt/disk{1..N}的路径，配置到storage.data-dirs用逗号分隔。同时storage.data-dirs.capacities配置N个数值，对应每块盘的磁盘容量。mounts里面配置宿主机path（按惯例，按数字顺序填写/mnt/disk{1..N})<br />

- 如何挂载多块磁盘

由于云盘有配额限制，因此我们可以在创建集群时，除了系统盘(vda)，我们额外挂载4块ESSD云盘（vdb,vdc,vdd,vde)。由于最后一块盘会自动挂载到/var/containers作为容器系统使用。因此我们选择使用（系统盘vda，云盘vdb，云盘vdc，云盘vdd）作为缓存数据的存储。<br />在每一台机器上，执行以下命令(格式化磁盘操作有风险,请确认后再操作)

```
mkfs.ext4 /dev/vdb
mkfs.ext4 /dev/vdc
mkfs.ext4 /dev/vdd
mkdir /mnt/disk{1..4}
mount -t ext4 /dev/vdb /mnt/disk2
mount -t ext4 /dev/vdc /mnt/disk3
mount -t ext4 /dev/vdd /mnt/disk4
```

其中/mnt/disk1使用了系统盘vda的空间，/mnt/disk{2..4}使用了3块云盘的空间。<br />

<a name="maGAo"></a>
## 3.2 使用内存缓存
JindoFS使用基于ramdisk的内存缓存，因此我们先挂载ramdisk（推荐使用tmpfs而不是ramfs）<br />在每一台机器上，执行以下命令

```
mkdir /mnt/disk1; mount -t tmpfs -o size=120G  mytmpfs /mnt/disk1
```

由于目前版本JindoFS只会使用约80%左右的配额，因此ramdisk创建时大小可以按预估数据量的120%进行创建。<br />
容器内路径，logDir使用临时路径，避免占用内存空间。然后master、worker服务的数据目录使用一块盘/mnt/disk1即可。

```yaml
properties:
 logDir: /tmp/bigboot-log/

 master:
 properties:
   namespace.meta-dir: /mnt/disk1/bigboot/server

worker:
 properties:
   storage.data-dirs: /mnt/disk1/bigboot
   storage.temp-data-dirs: /mnt/disk1/bigboot/tmp
   storage.watermark.high.ratio: 1
   storage.watermark.low.ratio: 0.95
   storage.data-dirs.capacities: 120g
   storage.meta-dir: /mnt/disk1/bigboot/bignode
```


<a name="tmZXm"></a>
## 3.2 使用Node Label将Master、Worker分开部署
一个Master需要服务于多个Worker，会成为集群中的瓶颈之一。因此建议预留至少8 vcore，16gb内存给Master。

- 对于某些场景，我们使用同构机型，选择少数几台相同的超高性能的机器（比如用4台ecs.gn6v-c8g1.16xlarge做机器学习训练），那么单机拥有充足的cpu、内存资源，因此可以使用默认配置，在master所在节点上也部署上Worker服务。



- 对于大部分场景（比如大数据分析场景），我们通常使用异构机型，选择1~3台cpu、内存较高的header节点，以及多台core节点用于执行计算任务，这种方式避免了计算作业任务对master服务造成影响。此时我们建议将JindoFS Master服务部署在header节点上，将Worker服务部署在core节点上。我们使用Kubernetes的Node Level功能来实现分开部署。操作方式如下：

    (1)登录容器服务 Kubernetes 版控制台。<br />    (2)在控制台左侧导航栏中，选择集群 - 节点。<br />    (3)在节点列表页面，选择目标集群并单击页面右上角标签管理。<br />    (4)在标签管理页面，批量选择节点，然后单击添加标签。<br />        可以批量选择多个core节点(排除1个header节点），在弹出的添加对话框中，填写标签名称和值。请确保名称设置为role，值设置为jindofs-worker。<br />       在同一个页面，再次选择1个header节点，然后单击添加标签。请确保名称设置为role，值设置为jindofs-master。
    <br />![docs-005.png](https://smartdata-binary.oss-cn-shanghai.aliyuncs.com/docs/docs-005.png)<br />![docs-006.png](https://smartdata-binary.oss-cn-shanghai.aliyuncs.com/docs/docs-006.png)


当然你可以执行kubectl命令，比如

```bash
kubectl  label no cn-beijing.192.168.8.240 role=jindofs-master
node/cn-beijing.192.168.8.240 labeled
kubectl  label no cn-beijing.192.168.8.241 role=jindofs-worker
node/cn-beijing.192.168.8.241 labeled
```

最后在JindoFS集群时，使用以下配置：

```yaml
master:
 nodeSelector:
   role: jindofs-master

worker:
 nodeSelector:
   role: jindofs-worker
```

<br />
<a name="kWhC7"></a>
# 四、附录
<a name="CrJCP"></a>
## 4.1 制作带有JindoFS SDK的镜像
制作镜像2种办法：<br />方法一：从镜像市场下载已有镜像，将sdk/lib/jindofs-sdk-*.jar拷贝到镜像内程序所在classpath<br />方法二：只接将JindoFS SDK和开源Spark（或Hadoop）一起打包成镜像。

本文演示用方法二制作Spark镜像。<br />首先我们下载b2sdk包进行解压

```
wget https://smartdata-binary.oss-cn-shanghai.aliyuncs.com/b2sdk-3.3.5.tar.gz
tar -zxvf b2sdk-3.3.5.tar.gz
```

从[spark下载页面](https://spark.apache.org/downloads.html)下载所需的spark版本，本次实验选择的saprk版本为2.4.7。运行如下命令下载spark并解压：

```
$ cd /root
$ wget https://mirrors.bit.edu.cn/apache/spark/spark-2.4.7/spark-2.4.7-bin-hadoop2.7.tgz
$ tar -xf spark-2.4.7-bin-hadoop2.7.tgz
$ export SPARK_HOME=/root/spark-2.4.7-bin-hadoop2.7
```

<br />将JindoFS SDK拷贝到Spark目录下

```
$ cp sdk/lib/jindofs-sdk-2.*.jar spark-2.4.7-bin-hadoop2.7/jars/
```

开始构建镜像：

```
cd ./spark-2.4.7-bin-hadoop2.7
docker build -t spark-jindofs:2.4.7 -f kubernetes/dockerfiles/spark/Dockerfile
```

请记住镜像名称“spark-jindofs:2.4.7”，在向k8s提交spark job中会用到这个信息。<br />镜像构建完成以后，对镜像的处理有两种方式：

- 如果有私有镜像仓库，将该镜像推送到私有镜像仓库中，同时保证k8s集群节点能够pull该镜像
- 如果没有私有镜像仓库，那么需要使用docker save命令将该镜像导出，然后scp到k8s集群的各个节点，在每个节点上使用docker load命令将镜像导入，这样就能保证每个节点上都存在该镜像。

<a name="Y3aW4"></a>
## 4.2 配置项列表

下列表单展示此chart的配置项与默认值，如需修改，请在参数列表中修改。

| Parameter | Description | Default |
| --- | --- | --- |
| properties.logDir | 容器内服务的日志目录，按照惯例请保存在默认位置，并且可以将该目录映射到宿主机，方便查看日志。 | /mnt/disk1/bigboot/log |
| <br /> | <br /> | <br /> |
| namespace.rpc.port | namespace的rpc端口，请保留默认值。 | 8101 |
| namespace.meta-dir | 容器内master服务的元数据目录，按照惯例请保存在默认位置，并且可以将该目录映射到宿主机，持久化该数据。 | /mnt/disk1/bigboot/server |
| namespace.filelet.cache.size | Master服务上内存中Inode缓存数量，当内存足够时适当调大该值，可以利用内存缓存提高性能。 | 100000 |
| namespace.blocklet.cache.size | Master服务上内存中Blocklet缓存数量，当内存足够时适当调大该值，可以利用内存缓存提高性能。 | 1000000 |
| namespace.backend.type | Master服务的元数据存储类型。目前仅支持rocksdb的方式。请保留默认值。 | rocksdb |
| jfs.namespaces | test表示当前JindoFS支持的命名空间，多个命名空间时以逗号隔开。 | test |
| jfs.namespaces.test.mode | cache表示test命名空间为缓存模式。block表示块模式。 | cache |
| jfs.namespaces.test.oss.uri | 表示test命名空间的后端存储。 | oss://xxx/ |
| jfs.namespaces.test.oss.access.key | 表示存储后端OSS的AccessKey ID | xxx |
| jfs.namespaces.test.oss.access.secret | 表示存储后端OSS的AccessKey Secret | xxx |
| <br /> | <br /> | <br /> |
| storage.rpc.port | worker的rpc端口，请保留默认值。 | 6101 |
| storage.data-dirs | worker容器内的缓存数据目录，多个目录用逗号隔开。 | /mnt/disk1/bigboot, /mnt/disk2/bigboot, /mnt/disk3/bigboot |
| storage.temp-data-dirs | worker容器内的临时文件目录，多个目录用逗号隔开。 | /mnt/disk1/bigboot/tmp |
| storage.watermark.high.ratio | worker使用的磁盘空间的水位上限百分比。假设500GB磁盘，0.4表示最大使用200GB | 0.4 |
| storage.watermark.low.ratio | worker使用的磁盘空间的水位下限百分比。假设500GB磁盘，0.2表示最少使用100GB | 0.2 |
| storage.data-dirs.capacities | 每块盘的容量大小，多个盘用逗号隔开。与storage.data-dirs的个数相对应。 | 80g,80g,80g |
| storage.meta-dir | worker的索引数据。按照惯例请保存在默认位置，并且可以将该目录映射到宿主机，方便持久化缓存信息。 |  /mnt/disk1/bigboot/bignode |
| <br /> | <br /> | <br /> |
| client.storage.rpc.port | worker的rpc端口，请保留默认值。 | 6101 |
| client.oss.retry | 客户端连接OSS失败时的重试次数 | 5 |
| client.oss.upload.threads | 客户端并行上传OSS的线程数 | 4 |
| client.oss.upload.queue.size | 客户端上传OSS的队列个数 | 5 |
| client.oss.upload.max.parallelism | 客户端并行上传OSS的最大线程数 | 16 |
| client.oss.timeout.millisecond | 客户端发送OSS请求的超时时间 | 30000 |
| client.oss.connection.timeout.millisecond | 客户端连接OSS的超时时间 | 3000 |
| <br /> | <br /> | <br /> |
| mounts.master | master服务挂载的宿主机hostPath和容器内的mountPath，如需持久化，请按惯例请填写/mnt/disk1 | 无 |
| mounts.workersAndClients | worker服务挂载的宿主机hostPath和容器内的mountPath，如需持久化，请按惯例请填写/mnt/disk1到/mnt/diskN | 无 |



