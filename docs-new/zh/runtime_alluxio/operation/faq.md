# 常见问题

## 1.为什么我运行例子[远程文件访问加速](../task/dataset_usage/accelerate_data_accessing.md)，执行第一次拷贝文件时会遇到`Input/output error`的错误。类似如下：

```
time cp ./pyspark-2.4.6.tar.gz /tmp/
cp: error reading ‘./pyspark-2.4.6.tar.gz’: Input/output error
cp: failed to extend ‘/tmp/pyspark-2.4.6.tar.gz’: Input/output error

real	3m15.795s
user	0m0.001s
sys	0m0.092s
```

这个原因是什么？

**回答**: 这个例子的目的是让用户在无需搭建UFS（underlayer file system）的情况下，利用现有的基于Http协议的Apache软件镜像下载地址演示数据拷贝加速的能力。而实际场景中，一般不会使用WebUFS的实现。但是这个例子有三个限制：

1.Apache软件镜像下载地址的可用性和访问速度

2.WebUFS来源于Alluxio的社区贡献，并不是最优实现。比如实现并不是基于offset的断点续载，这就导致每次远程读操作都需要触发WebUFS大量数据块读

3.由于拷贝行为基于Fuse实现，每一次Fuse的chunk读由于Linux Kernel的上限都是128KB；从而导致文件越大，在初次拷贝时，就会触发大量的读操作

针对该问题，我们提拱了优化的方案：

1.配置读时，将Block size和chunk size读设置的大于文件的大小，这样就可以避免Fuse实现中频繁读的影响。

```
alluxio.user.block.size.bytes.default: 256MB
alluxio.user.streaming.reader.chunk.size.bytes: 256MB
alluxio.user.local.reader.chunk.size.bytes: 256MB
alluxio.worker.network.reader.buffer.size: 256MB
```

2.为了保障目标文件可以被下载成功，可以调整block下载的超时。例子中的超时时间是5分钟，如果您的网络状况不佳，可以酌情设置更长时间。

```
alluxio.user.streaming.data.timeout: 300sec
```

3.您可以尝试手动加载该文件

```
kubectl exec -it hbase-master-0 bash
time alluxio fs  distributedLoad --replication 1 /
```

## 2. 为什么在运行示例 [Nonroot access](../task/security/nonroot_access.md)时，遇到了 mkdir 权限被拒绝的问题？

**回答**:在非 root 用户情况下，你首先必须要检查是否将正确的用户信息传递给了 runtime。其次你应该检查 Alluxio master pod 的状态，并使用 journalctl 去查看 Alluxio master pod 节点对应 kubelet 的日志。
当将 hostpath 挂载到容器中，可能会造成无法创建文件的问题，因此我们必须要去检查 root 是否具有权限。例如在如下的情况中 root 是有权限使用 /dir。
```
$ stat /dir
  File: ‘/dir’
  Size: 32              Blocks: 0          IO Block: 4096   directory
Device: fd00h/64768d    Inode: 83          Links: 3
Access: (0755/drwxr-xr-x)  Uid: (    0/    root)   Gid: (    0/    root)
Access: 2021-04-14 23:35:47.928805350 +0800
Modify: 2021-01-19 00:16:21.539559082 +0800
Change: 2021-01-19 00:16:21.539559082 +0800
 Birth: -

```

## 3. 为什么我在创建了Dataset和AlluxioRuntime后，alluxio master Pod进入了CrashLoopBackOff状态？

**回答**:首先需要使用命令`kubectl describe pod <dataset_name>-master-0 `查看Pod错误退出的原因。

alluxio master Pod由两个容器alluxio-master和alluxio-job-master组成，如果是其中某一容器执行失败退出，则查看它异常退出前输出的日志。

例如，某次alluxio-job-master容器异常退出前，输出的日志为：

```
$ kubectl logs imagenet-master-0  -c alluxio-job-master -p
2021-06-08 12:03:47,611 INFO  MetricsSystem - Starting sinks with config: {}.
2021-06-08 12:03:47,616 INFO  MetricsHeartbeatContext - Created metrics heartbeat with ID app-1642528563209467270. This ID will be used for identifying info from the client. It can be set manually through the alluxio.user.app.id property
2021-06-08 12:03:47,656 INFO  TieredIdentityFactory - Initialized tiered identity TieredIdentity(node=132.252.41.86, rack=null)
2021-06-08 12:03:47,697 INFO  ExtensionFactoryRegistry - Loading core jars from /opt/alluxio-release-2.5.0-2-SNAPSHOT/lib
2021-06-08 12:03:47,784 INFO  ExtensionFactoryRegistry - Loading extension jars from /opt/alluxio-release-2.5.0-2-SNAPSHOT/extensions
2021-06-08 12:03:50,767 ERROR AlluxioJobMasterProcess - java.net.UnknownHostException: jfk8snode43: jfk8snode43: Temporary failure in name resolution
java.lang.RuntimeException: java.net.UnknownHostException: jfk8snode43: jfk8snode43: Temporary failure in name resolution
        at alluxio.util.network.NetworkAddressUtils.getLocalIpAddress(NetworkAddressUtils.java:514)
        at alluxio.util.network.NetworkAddressUtils.getLocalHostName(NetworkAddressUtils.java:436)
        at alluxio.util.network.NetworkAddressUtils.getConnectHost(NetworkAddressUtils.java:313)
        at alluxio.underfs.JobUfsManager.connectUfs(JobUfsManager.java:55)
        at alluxio.underfs.AbstractUfsManager.getOrAdd(AbstractUfsManager.java:150)
        at alluxio.underfs.AbstractUfsManager.lambda$addMount$0(AbstractUfsManager.java:171)
        at alluxio.underfs.UfsManager$UfsClient.acquireUfsResource(UfsManager.java:61)
        at alluxio.master.journal.ufs.UfsJournal.<init>(UfsJournal.java:150)
        at alluxio.master.journal.ufs.UfsJournalSystem.createJournal(UfsJournalSystem.java:83)
        at alluxio.master.journal.ufs.UfsJournalSystem.createJournal(UfsJournalSystem.java:53)
        at alluxio.master.AbstractMaster.<init>(AbstractMaster.java:73)
        at alluxio.master.job.JobMaster.<init>(JobMaster.java:157)
        at alluxio.master.AlluxioJobMasterProcess.<init>(AlluxioJobMasterProcess.java:92)
        at alluxio.master.AlluxioJobMasterProcess$Factory.create(AlluxioJobMasterProcess.java:269)
        at alluxio.master.AlluxioJobMaster.main(AlluxioJobMaster.java:45)
Caused by: java.net.UnknownHostException: jfk8snode43: jfk8snode43: Temporary failure in name resolution
        at java.net.InetAddress.getLocalHost(InetAddress.java:1506)
        at alluxio.util.network.NetworkAddressUtils.getLocalIpAddress(NetworkAddressUtils.java:472)
        ... 14 more
Caused by: java.net.UnknownHostException: jfk8snode43: Temporary failure in name resolution
        at java.net.Inet4AddressImpl.lookupAllHostAddr(Native Method)
        at java.net.InetAddress$2.lookupAllHostAddr(InetAddress.java:929)
        at java.net.InetAddress.getAddressesFromNameService(InetAddress.java:1324)
        at java.net.InetAddress.getLocalHost(InetAddress.java:1501)
        ... 15 more
2021-06-08 12:03:50,773 ERROR AlluxioJobMaster - Failed to create job master process
java.lang.RuntimeException: java.net.UnknownHostException: jfk8snode43: jfk8snode43: Temporary failure in name resolution
        at alluxio.util.network.NetworkAddressUtils.getLocalIpAddress(NetworkAddressUtils.java:514)
        at alluxio.util.network.NetworkAddressUtils.getLocalHostName(NetworkAddressUtils.java:436)
        at alluxio.util.network.NetworkAddressUtils.getConnectHost(NetworkAddressUtils.java:313)
        at alluxio.underfs.JobUfsManager.connectUfs(JobUfsManager.java:55)
        at alluxio.underfs.AbstractUfsManager.getOrAdd(AbstractUfsManager.java:150)
        at alluxio.underfs.AbstractUfsManager.lambda$addMount$0(AbstractUfsManager.java:171)
        at alluxio.underfs.UfsManager$UfsClient.acquireUfsResource(UfsManager.java:61)
        at alluxio.master.journal.ufs.UfsJournal.<init>(UfsJournal.java:150)
        at alluxio.master.journal.ufs.UfsJournalSystem.createJournal(UfsJournalSystem.java:83)
        at alluxio.master.journal.ufs.UfsJournalSystem.createJournal(UfsJournalSystem.java:53)
        at alluxio.master.AbstractMaster.<init>(AbstractMaster.java:73)
        at alluxio.master.job.JobMaster.<init>(JobMaster.java:157)
        at alluxio.master.AlluxioJobMasterProcess.<init>(AlluxioJobMasterProcess.java:92)
        at alluxio.master.AlluxioJobMasterProcess$Factory.create(AlluxioJobMasterProcess.java:269)
        at alluxio.master.AlluxioJobMaster.main(AlluxioJobMaster.java:45)
Caused by: java.net.UnknownHostException: jfk8snode43: jfk8snode43: Temporary failure in name resolution
        at java.net.InetAddress.getLocalHost(InetAddress.java:1506)
        at alluxio.util.network.NetworkAddressUtils.getLocalIpAddress(NetworkAddressUtils.java:472)
        ... 14 more
Caused by: java.net.UnknownHostException: jfk8snode43: Temporary failure in name resolution
        at java.net.Inet4AddressImpl.lookupAllHostAddr(Native Method)
        at java.net.InetAddress$2.lookupAllHostAddr(InetAddress.java:929)
        at java.net.InetAddress.getAddressesFromNameService(InetAddress.java:1324)
        at java.net.InetAddress.getLocalHost(InetAddress.java:1501)
        ... 15 more
2021-06-08 12:03:50,917 INFO  NettyUtils - EPOLL_MODE is available
2021-06-08 12:03:51,319 WARN  MetricsHeartbeatContext - Failed to heartbeat to the metrics master before exit
```

此错误一般是由于alluxio master Pod所在主机，未在dns服务器或本地/etc/hosts文件中配置hostname和ip的映射关系，导致alluxio-job-master无法解析hostname。
此时，你需要登陆alluxio master Pod所在主机，执行命令`hostname`查询主机名，将其与ip的映射关系写入/etc/hosts文件。

你可以搜索在本项目的issue中进行搜索，寻找你遇到的报错信息的解决方案。如果没有找到可以解决你问题的issue，也可以提出新的issue。