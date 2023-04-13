# 示例 - 如何在 Fluid 中使用 EFC

## 背景介绍

EFC 是一款针对分布式文件系统 NAS 的用户态客户端，并在提供分布式缓存的同时，保证多客户端之间的缓存一致性。EFC现阶段支持阿里云NAS，未来会支持通用NAS和GPFS等高速分布式文件系统。

如何开启NAS服务能力，可以参考[文档](https://help.aliyun.com/document_detail/148430.html)

## 安装

您可以从 [Fluid Releases](https://github.com/fluid-cloudnative/fluid/releases) 下载最新的 Fluid 安装包。

在 Fluid 的安装 chart values.yaml 中将 `runtime.eac.enable` 设置为 `true` ，再参考 [安装文档](../userguide/install.md) 完成安装。并检查Fluid各组件正常运行：

```shell
$ kubectl get po -n fluid-system
NAME                                     READY   STATUS    RESTARTS   AGE
csi-nodeplugin-fluid-4m2rq               2/2     Running   0          81s
csi-nodeplugin-fluid-8l6nr               2/2     Running   0          81s
csi-nodeplugin-fluid-t7hl2               2/2     Running   0          81s
dataset-controller-99bc4dcc8-sl6h7       1/1     Running   0          81s
eacruntime-controller-6fd48c77fc-k2hhr   1/1     Running   0          81s
fluid-webhook-d8c4dcc7-whq5k             1/1     Running   0          81s
fluidapp-controller-78c7ccd7fd-blw6w     1/1     Running   0          81s
```

确保 `eacruntime-controller`、`dataset-controller`、`fluid-webhook` 的 pod 以及若干 `csi-nodeplugin` pod 正常运行。

## 新建工作环境

```shell
$ mkdir <any-path>/efc
$ cd <any-path>/efc
```

## 运行示例

在使用 EFC 之前，您需要拥有一个[通用型 NAS](https://www.aliyun.com/product/nas?spm=5176.19720258.J_3207526240.80.e93976f4Ps3XxX)，以及一个和 NAS 处在同一 [VPC](https://www.aliyun.com/product/vpc?spm=5176.59209.J_3207526240.35.253f76b9hZAU4x) 的 [ACK](https://www.aliyun.com/product/kubernetes?spm=5176.7937172.J_3207526240.54.7f51751avPxHwi) 集群，并确保集群节点操作系统使用 Alibaba Cloud Linux 2.1903。

**查看待创建的 `Dataset` 资源对象**

```yaml
$ cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: mydemo
spec:
  mounts:
    - mountPoint: "eac://nas-mount-point-address:/sub/path"
EOF
```

其中：

- `mountPoint`：指的是 EFC 的子目录，是用户在 NAS 文件系统中存储数据的目录，以 `eac://` 开头；如 `eac://nas-mount-point-address:/sub/path` 为 `nas-mount-point-address` 文件系统的 `/sub/path` 子目录。

**创建 `Dataset` 资源对象**
```shell
$ kubectl create -f dataset.yaml
dataset.data.fluid.io/mydemo created
```

**查看 `Dataset` 资源对象状态**
```shell
$ kubectl get dataset mydemo
NAME     UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE      AGE
mydemo                                                                  NotBound   14s
```

如上所示，`status` 中的 `phase` 属性值为 `NotBound`，这意味着该 `Dataset` 资源对象目前还未与任何 `EFCRuntime` 资源对象绑定，接下来，我们将创建一个 `EFCRuntime` 资源对象。

**查看待创建的 `EFCRuntime` 资源对象**

```yaml
$ cat<<EOF >runtime.yaml
apiVersion: data.fluid.io/v1alpha1
kind: EFCRuntime
metadata:
  name: mydemo
spec:
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: MEM
        volumeType: emptyDir
        path: /mnt/efc-worker-cache-path
        quota: 2Gi
  fuse:
    properties:
      g_unas_InodeAttrExpireTimeoutSec: "100"
      g_unas_InodeEntryExpireTimeoutSec: "100"
EOF
```

**创建 `EFCRuntime` 资源对象**

```shell
$ kubectl create -f runtime.yaml
efcruntime.data.fluid.io/mydemo created
```

**检查 `EFCRuntime` 资源对象是否已经创建**
```shell
$ kubectl get efcruntime
NAME     MASTER PHASE   WORKER PHASE   FUSE PHASE   AGE
mydemo   NotReady                                   23s
```

等待一段时间，让 `EFCRuntime` 资源对象中的各个组件得以顺利启动，你会看到类似以下状态：

```shell
$ kubectl get po | grep mydemo
mydemo-master-0   2/2     Running   0          81s
mydemo-worker-0   1/1     Running   0          61s
```

```shell
$ kubectl get efcruntime
NAME     MASTER PHASE   WORKER PHASE   FUSE PHASE   AGE
mydemo   Ready          Ready          Ready        55s
```

`EFCRuntime` 的 FUSE 组件实现了懒启动，会在 pod 使用时再创建。

然后，再查看 `Dataset` 状态，发现已经与 `EFCRuntime` 绑定。

```shell
$ kubectl get dataset mydemo
NAME     UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
mydemo                                                                  Bound   106s
```

**查看待创建的 StatefulSet 资源对象**，其中 StatefulSet 使用上面创建的 `Dataset` 的方式为指定同名的 PVC。

```yaml
$ cat<<EOF >app.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: mydemo-app
  labels:
    app: busybox
spec:
  serviceName: "busybox"
  replicas: 2
  selector:
    matchLabels:
      app: busybox
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
        - name: busybox
          image: busybox
          command: ["/bin/sh"]
          args: ["-c", "sleep 3600"]
          volumeMounts:
            - mountPath: "/data"
              name: demo
      volumes:
        - name: demo
          persistentVolumeClaim:
            claimName: mydemo
EOF
```

**创建 StatefulSet 资源对象**

```shell
$ kubectl create -f app.yaml
statefulset.apps/mydemo-app created
```

**检查 StatefulSet 资源对象是否已经创建**
```shell
$ kubectl get po | grep mydemo
mydemo-app-0        1/1     Running   0          64s
mydemo-app-1        1/1     Running   0          20s
mydemo-fuse-dfswf   2/2     Running   0          20s
mydemo-fuse-gb4lm   2/2     Running   0          63s
mydemo-master-0     2/2     Running   0          3m12s
mydemo-worker-0     1/1     Running   0          2m52s
```

可以看到 StatefulSet 已经创建成功，同时 EFC 的 FUSE 组件也启动成功。

**测试缓存加速效果**
```shell
$ kubectl exec -it mydemo-app-0 -- /bin/sh -c  'ls -hl /data'
total 1G
-rw-r--r--    1 root     root          15 Dec 14 08:24 test.txt
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp1.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp10.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp11.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp12.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp13.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp14.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp15.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp16.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp17.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp18.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp19.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp2.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp20.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp21.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp22.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp23.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp24.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp25.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp26.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp27.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp28.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp29.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp3.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp30.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp31.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp32.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp33.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp34.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp35.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp36.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp37.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp38.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp39.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp4.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp40.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp5.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp6.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp7.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp8.32M
-rw-r--r--    1 root     root       32.0M Dec 14 08:24 tmp9.32M
```
可以看到被挂载目录下存在多个 32M 的普通文件，现通过在两个不同业务 pod 中进行 cp 操作来验证缓存加速效果。

```shell
$ kubectl exec -it mydemo-app-0 -- /bin/sh -c  'time cp /data/* /'
real  0m 15.25s
user  0m 0.00s
sys 0m 1.20s
```
先在 mydemo-app-0 pod 中执行 cp 操作耗时 15.25s。

```shell
$ kubectl exec -it mydemo-app-1 -- /bin/sh -c  'time cp /data/* /'
real  0m 5.27s
user  0m 0.00s
sys 0m 1.27s
```
后在 mydemo-app-1 pod 中执行 cp 操作耗时 5.27s，加速效果明显。

```shell
$ kubectl logs mydemo-worker-0
2022/12/14 17:23:57|INFO |th=0000000002B70450|photon/syncio/epoll.cpp:336|fd_events_epoll_init:init event engine: epoll
2022/12/14 17:23:57|INFO |th=0000000002B70450|server/server.cpp:323|Server:[cachedfs->get_pool()->defaultQuota()=-1]
2022/12/14 17:23:57|INFO |th=0000000002B70450|server/server.cpp:353|Server:Server Start.
2022/12/14 17:34:30|INFO |th=00007F05981F8300|server/server.cpp:335|operator():Accept 192.168.0.185:45966
2022/12/14 17:34:30|INFO |th=00007F05969E7280|server/server.cpp:335|operator():Accept 192.168.0.185:45968
2022/12/14 17:34:30|INFO |th=00007F05979F16C0|server/cache-adapter.cpp:50|add_throughPutCount:server current READ Throughput = 0MB/s, WRITE Throughput = 0MB/s
2022/12/14 17:34:30|INFO |th=00007F05961E4340|server/server.cpp:335|operator():Accept 192.168.0.185:45970
2022/12/14 17:34:31|INFO |th=00007F05959E0740|server/server.cpp:335|operator():Accept 192.168.0.185:45972
2022/12/14 17:34:31|INFO |th=00007F05930B7280|server/cache-adapter.cpp:50|add_throughPutCount:server current READ Throughput = 71MB/s, WRITE Throughput = 37MB/s
2022/12/14 17:34:32|INFO |th=00007F05926AEF40|server/cache-adapter.cpp:50|add_throughPutCount:server current READ Throughput = 177MB/s, WRITE Throughput = 82MB/s
2022/12/14 17:34:33|INFO |th=00007F05926AEF40|server/cache-adapter.cpp:50|add_throughPutCount:server current READ Throughput = 212MB/s, WRITE Throughput = 89MB/s
2022/12/14 17:34:34|INFO |th=00007F05926AEF40|server/cache-adapter.cpp:50|add_throughPutCount:server current READ Throughput = 204MB/s, WRITE Throughput = 92MB/s
2022/12/14 17:34:35|INFO |th=00007F05951D9AC0|server/cache-adapter.cpp:50|add_throughPutCount:server current READ Throughput = 191MB/s, WRITE Throughput = 86MB/s
2022/12/14 17:34:37|INFO |th=00007F05916A4B00|server/cache-adapter.cpp:50|add_throughPutCount:server current READ Throughput = 199MB/s, WRITE Throughput = 90MB/s
2022/12/14 17:34:38|INFO |th=00007F05938C0EC0|server/cache-adapter.cpp:50|add_throughPutCount:server current READ Throughput = 196MB/s, WRITE Throughput = 88MB/s
2022/12/14 17:34:39|INFO |th=00007F05926AEF40|server/cache-adapter.cpp:50|add_throughPutCount:server current READ Throughput = 191MB/s, WRITE Throughput = 87MB/s
2022/12/14 17:34:40|INFO |th=00007F058F57DA80|server/cache-adapter.cpp:50|add_throughPutCount:server current READ Throughput = 154MB/s, WRITE Throughput = 58MB/s
2022/12/14 17:34:41|INFO |th=00007F05979F4B40|server/cache-adapter.cpp:50|add_throughPutCount:server current READ Throughput = 157MB/s, WRITE Throughput = 60MB/s
2022/12/14 17:34:42|INFO |th=00007F0590D99EC0|server/cache-adapter.cpp:50|add_throughPutCount:server current READ Throughput = 210MB/s, WRITE Throughput = 77MB/s
2022/12/14 17:34:43|INFO |th=00007F05951D82C0|server/cache-adapter.cpp:50|add_throughPutCount:server current READ Throughput = 213MB/s, WRITE Throughput = 109MB/s
2022/12/14 17:34:44|INFO |th=00007F0590D99EC0|server/cache-adapter.cpp:50|add_throughPutCount:server current READ Throughput = 278MB/s, WRITE Throughput = 118MB/s
2022/12/14 17:34:45|INFO |th=00007F05979F3B00|server/cache-adapter.cpp:50|add_throughPutCount:server current READ Throughput = 282MB/s, WRITE Throughput = 113MB/s
2022/12/14 17:34:58|INFO |th=00007F05951D6680|server/server.cpp:335|operator():Accept 192.168.0.181:55794
2022/12/14 17:34:58|INFO |th=00007F05940C9EC0|server/cache-adapter.cpp:50|add_throughPutCount:server current READ Throughput = 5MB/s, WRITE Throughput = 2MB/s
2022/12/14 17:34:58|INFO |th=00007F05979F3300|server/server.cpp:335|operator():Accept 192.168.0.181:55796
2022/12/14 17:34:58|INFO |th=00007F05938C26C0|server/server.cpp:335|operator():Accept 192.168.0.181:55798
2022/12/14 17:34:58|INFO |th=00007F05930BCB00|server/server.cpp:335|operator():Accept 192.168.0.181:55800
2022/12/14 17:34:59|INFO |th=00007F0590D9AEC0|server/cache-adapter.cpp:50|add_throughPutCount:server current READ Throughput = 475MB/s, WRITE Throughput = 0MB/s
2022/12/14 17:35:00|INFO |th=00007F05926AF740|server/cache-adapter.cpp:50|add_throughPutCount:server current READ Throughput = 463MB/s, WRITE Throughput = 0MB/s
2022/12/14 17:35:01|INFO |th=00007F0590592EC0|server/cache-adapter.cpp:50|add_throughPutCount:server current READ Throughput = 493MB/s, WRITE Throughput = 0MB/s
2022/12/14 17:35:02|INFO |th=00007F058F585700|server/cache-adapter.cpp:50|add_throughPutCount:server current READ Throughput = 477MB/s, WRITE Throughput = 0MB/s
2022/12/14 17:35:03|INFO |th=00007F0591EA6280|server/cache-adapter.cpp:50|add_throughPutCount:server current READ Throughput = 562MB/s, WRITE Throughput = 0MB/s
```
观察 worker 的日志，可以发现首次操作会将缓存写入 worker，第二次操作将读取缓存。