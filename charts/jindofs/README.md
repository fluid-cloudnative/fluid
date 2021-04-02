# JindoFS
## JindoFS概述：云原生的大数据计算存储分离方案
### JindoFS 之前
在 JindoFS 之前，云上客户主要使用 HDFS 和 OSS/S3 作为大数据存储。HDFS 是 Hadoop 原生的存储系统，10 年来，HDFS 已经成为大数据生态的存储标准，但是我们也可以看到 HDFS 虽然不断优化，但是 JVM 的瓶颈也始终无法突破，社区后来重新设计了 OZone。OSS/S3 作为云上对象存储的代表，也在大数据生态进行了适配，但是由于对象存储设计上的特点，元数据相关操作无法达到 HDFS 一样的效率；对象存储给客户的带宽不断增加，但是也是有限的，一些时候较难完全满足用户大数据使用上的需求。
### Jindo 的由来
EMR Jindo 是阿里云基于 Apache Spark / Apache Hadoop 在云上定制的分布式计算和存储引擎。Jindo 原是内部的研发代号，取自筋斗(云)的谐音，EMR Jindo 在开源基础上做了大量优化和扩展，深度集成和连接了众多阿里云基础服务。阿里云 EMR (E-MapReduce) 在 TPC 官方提交的 TPCDS 成绩，也是使用 Jindo 提交的。

http://www.tpc.org/tpcds/results/tpcds_perf_results.asp?resulttype=all
### JindoFS
EMR Jindo 有计算和存储两大部分，存储的部分叫 JindoFS。JindoFS 是阿里云针对云上存储定制的自研大数据存储服务，完全兼容 Hadoop 文件系统接口，给客户带来更加灵活、高效的计算存储方案，目前已验证支持阿里云 EMR 中所有的计算服务和引擎：Spark、Flink、Hive、MapReduce、Presto、Impala 等。Jindo FS 有两种使用模式，块存储模式和缓存模式。下面我们来分析下，JindoFS 是如何来解决大数据上的存储问题的。

<div align="center">
<img src="https://ucc.alicdn.com/pic/developer-ecology/d37dd8d3bdc246c59e4eae22dd5eef32.png" align="center">
</div>

### 块存储模式
计算和存储分离是业界的趋势，OSS 这样的云上存储能力是无限大的，成本上非常有优势，如何利用 OSS 提供的无限存储能力，同时又高效地操作文件系统的元数据。JindoFS 块存储模式提供了一套完整的云原生解决方案。
JindoFS 的块存储模式，在元数据上使用 JindoNameService 服务管理 Jindo 文件系统元数据，元数据操作的性能和体验上可以对标 HDFS NameNode。同时，JindoStorageService 保障了数据可以始终有一份存在 OSS 上，即使数据节点被释放，数据也可以随时从 OSS 上拉取，成本上也可以做到更加灵活。

JindoFS 的块存储模式，也支持多种存储策略，比如，本地存两份，OSS上存一份；本地存两份，OSS上不存储；本地不存，OSS上存一份等等。用户可以充分利用不同的存储策略根据业务或者数据冷热进行使用。

块存储使用了全新的 jfs:// 格式，原始 HDFS/OSS 数据通过 distcp 方式即可完成数据导入，同时，JindoFS 提供了 SDK，在 EMR 集群外部，用户也可以读写 Jindo FS。
### 缓存模式
缓存模式，正如“缓存”本身的含义，通过缓存的方式，在本地集群基于 JindoFS 的存储能力构建了一个分布式缓存服务，远端的数据可以保存在本地集群，使远端数据变成“本地化”。简单地描述 JindoFS 缓存模式解决的问题
就是“OSS / 远端HDFS 已经有了大量数据，每次读数据的时候网络带宽经常被打满，Jindo FS 就可以通过缓存模式优化网络带宽的限制。”

“原来的文件路径是 oss://bucket1/file1 或 hdfs://namenode/file2，不想改作业的路径可以吗？”。是的，不需要修改。EMR 对 OSS 进行了适配（后续会支持远端 HDFS 的场景），可以通过配置的方式使用缓存模式。缓存对于上层的作业做到了完全无感。

但是缓存模式也不是万能的，为了保证多端数据一致性，rename 这种操作一定要同步刷新到远端的 OSS / HDFS，特别是 OSS 的Rename 操作比较耗时，缓存模式对 rename这种文件元数据操作暂时不能优化。
### 附录：JindoFS参数说明

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
| storage.rpc.port | worker的rpc端口，请保留默认值。 | 6101 |
| storage.data-dirs | worker容器内的缓存数据目录，多个目录用逗号隔开。 | /mnt/disk1/bigboot, /mnt/disk2/bigboot, /mnt/disk3/bigboot |
| storage.temp-data-dirs | worker容器内的临时文件目录，多个目录用逗号隔开。 | /mnt/disk1/bigboot/tmp |
| storage.watermark.high.ratio | worker使用的磁盘空间的水位上限百分比。假设500GB磁盘，0.4表示最大使用200GB | 0.4 |
| storage.watermark.low.ratio | worker使用的磁盘空间的水位下限百分比。假设500GB磁盘，0.2表示最少使用100GB | 0.2 |
| storage.data-dirs.capacities | 每块盘的容量大小，多个盘用逗号隔开。与storage.data-dirs的个数相对应。 | 80g,80g,80g |
| storage.meta-dir | worker的索引数据。按照惯例请保存在默认位置，并且可以将该目录映射到宿主机，方便持久化缓存信息。 |  /mnt/disk1/bigboot/bignode |
| client.storage.rpc.port | worker的rpc端口，请保留默认值。 | 6101 |
| client.oss.retry | 客户端连接OSS失败时的重试次数 | 5 |
| client.oss.upload.threads | 客户端并行上传OSS的线程数 | 4 |
| client.oss.upload.queue.size | 客户端上传OSS的队列个数 | 5 |
| client.oss.upload.max.parallelism | 客户端并行上传OSS的最大线程数 | 16 |
| client.oss.timeout.millisecond | 客户端发送OSS请求的超时时间 | 30000 |
| client.oss.connection.timeout.millisecond | 客户端连接OSS的超时时间 | 3000 |
| mounts.master | master服务挂载的宿主机hostPath和容器内的mountPath，如需持久化，请按惯例请填写/mnt/disk1 | 无 |
| mounts.workersAndClients | worker服务挂载的宿主机hostPath和容器内的mountPath，如需持久化，请按惯例请填写/mnt/disk1到/mnt/diskN | 无 |