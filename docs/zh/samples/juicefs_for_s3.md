# 示例 - 如何在 Fluid 中使用 JuiceFS 同步 S3 中的数据

JuiceFS 采用「数据」与「元数据」分离存储的架构，从而实现文件系统的分布式设计。使用 JuiceFS 存储数据，数据本身会被持久化在对象存储（例如，Amazon S3），相对应的元数据可以按需持久化在
Redis、MySQL、TiKV 等多种数据库中。

故而，在 Fluid 中使用 JuiceFS 同步 S3 中已有的数据，需要先将 S3 中的数据初始化到 JuiceFS 中。

## 部署 JuiceFSRuntime 环境

具体部署方法参考文档 [如何在 Fluid 中使用 JuiceFS](./juicefs_runtime.md)。

在 JuiceFSRuntime 和 Dataset 创建成功后，等待 worker pod 启动成功，再进行下面的步骤。

## 数据初始化

在 worker pod 中可使用 `juicefs sync` 将指定 bucket 的数据同步到 juicefs 中。
例如，在 bucket `jfs-test-tmp` 中有一个文件 `JuiceFS_logo.png`，在 worker pod 中挂载点为 `/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse`，
按如下操作，可以将其同步到 JuiceFS 中。

### 查看 worker 中的挂载点

Worker pod 的挂载点格式为：`/runtime-mnt/juicefs/<namespace>/<runtimeName>/juicefs-fuse`。可以在 worker pod 中通过如下命令查看：

```shell
$ kubectl exec -it jfsdemo-worker-0 bash
root@jfsdemo-worker-0:~# df -h
Filesystem      Size  Used Avail Use% Mounted on
overlay         100G   29G   72G  29% /
tmpfs            64M     0   64M   0% /dev
tmpfs           2.0G     0  2.0G   0% /sys/fs/cgroup
/dev/sdb1       100G   29G   72G  29% /etc/hosts
shm              64M     0   64M   0% /dev/shm
tmpfs           3.8G   12K  3.8G   1% /run/secrets/kubernetes.io/serviceaccount
JuiceFS:minio   1.0P  1.1G  1.0P   1% /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse
```

### 使用 `juicefs sync` 同步数据

```shell
$ kubectl exec -it jfsdemo-worker-0 bash
root@jfsdemo-worker-0:~# juicefs sync s3://jfs-test-tmp/ /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse/
2022/05/09 07:19:53.852821 juicefs[1995593] <INFO>: Syncing from s3://jfs-test-tmp/ to file:///runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse/ [sync.go:571]
Scanned objects count: 1 / 1 [==============================================================]  done
 Copied objects count: 1
 Copied objects bytes: 101.75 KiB (104192 Bytes)   305.10 KiB/s
Checked objects bytes: 0.00 b     (0 Bytes)        0.00 b/s
Deleted objects count: 0
Skipped objects count: 0
 Failed objects count: 0
2022/05/09 07:19:54.187294 juicefs[1995593] <INFO>: Found: 1, copied: 1 (101.75 KiB), checked: 0 B, deleted: 0, skipped: 0, failed: 0 [sync.go:795]
root@jfsdemo-worker-0:~#
```

`juicefs sync` 的格式为：`[NAME://][ACCESS_KEY:SECRET_KEY@]BUCKET[.ENDPOINT][/PREFIX]`，具体可参考[文档](https://juicefs.com/docs/zh/community/administration/sync)

### 检查文件同步结果

```shell
root@jfsdemo-worker-0:~# ls /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse/
JuiceFS_logo.png
root@jfsdemo-worker-0:~#
```

可以看到 bucket 中的文件已经被同步到了 JuiceFS 中。

最后创建业务 Pod，其中 Pod 使用上面创建的 `Dataset` 的方式为指定同名的 PVC。该步骤与文档 [如何在 Fluid 中使用 JuiceFS](./juicefs_runtime.md) 中一致，这里不再赘述。
