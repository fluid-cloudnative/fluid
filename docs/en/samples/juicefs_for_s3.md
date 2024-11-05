# DEMO - How to sync data of S3 using JuiceFS in Fluid

JuiceFS implements a distributed file system by adopting the architecture that separates "data" and "metadata" storage.
When using JuiceFS to store data, the data itself is persisted in object storage (e.g., Amazon S3), and the
corresponding metadata can be persisted in various databases such as Redis, MySQL, TiKV, etc., based on the scenarios
and requirements.

Therefore, to use JuiceFS in Fluid to synchronize existing data in S3, it is necessary to initialize the data in S3 into
JuiceFS first.

## Deploy JuiceFSRuntime

For the specific deployment method, please refer to the document ["How to use JuiceFS in Fluid"](./juicefs_runtime.md).

After the JuiceFSRuntime and Dataset are successfully created, wait for the worker pod to start successfully, and then
refer to the following steps.

## Data initialization

In the worker pod, you can use `juicefs sync` to synchronize the data of the specified bucket to juicefs.
For example, there is a file `JuiceFS_logo.png` in bucket `jfs-test-tmp`, in worker pod the mount point
is `/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse`, sync it into JuiceFS as follows.

### Check mount point in worker

The mount point format of worker pod is: `/runtime-mnt/juicefs/<namespace>/<runtimeName>/juicefs-fuse`. You can check it
in the worker pod with the following command:

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

### Sync data with `juicefs sync`

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

Address syntax of `juicefs sync` follows `[NAME://][ACCESS_KEY:SECRET_KEY@]BUCKET[.ENDPOINT][/PREFIX]`.
Refer to [this document](https://juicefs.com/docs/community/administration/sync) for more details.

### Check result

```shell
root@jfsdemo-worker-0:~# ls /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse/
JuiceFS_logo.png
root@jfsdemo-worker-0:~#
```

You can see that the files in the bucket have been synced to JuiceFS.

Finally, create a business Pod, where the Pod uses the `Dataset` created above to specify a PVC with the same name. This
step is consistent with the document ["How to use JuiceFS in Fluid"](./juicefs_runtime.md), and will not be repeated here.
