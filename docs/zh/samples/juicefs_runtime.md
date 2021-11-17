# 示例 - 如何在 Fluid 中使用 JuiceFS

## 背景介绍

JuiceFS 是一款面向云环境设计的高性能共享文件系统，提供完备的 POSIX 兼容性，可将海量低价的云存储作为本地磁盘使用，亦可同时被多台主机同时挂载读写。

如何使用 JuiceFS 可以参考文档[JuiceFS 快速上手](https://github.com/juicedata/juicefs/blob/main/docs/zh_cn/quick_start_guide.md)

## 安装

您可以从 [Fluid Releases](https://github.com/fluid-cloudnative/fluid/releases) 下载最新的 Fluid 安装包。

在 Fluid 的安装 chart values.yaml 中将 `runtime.juicefs.enable` 设置为 `true` ，再参考 [安装文档](../userguide/install.md) 完成安装。并检查Fluid各组件正常运行：

```shell
kubectl get po -n fluid-system
NAME                                         READY   STATUS              RESTARTS   AGE
csi-nodeplugin-fluid-ctc4l                   2/2     Running             0          113s
csi-nodeplugin-fluid-k7cqt                   2/2     Running             0          113s
csi-nodeplugin-fluid-x9dfd                   2/2     Running             0          113s
dataset-controller-57ddd56b54-9vd86          1/1     Running             0          113s
fluid-webhook-84467465f8-t65mr               1/1     Running             0          113s
juicefsruntime-controller-56df96b75f-qzq8x   1/1     Running             0          113s
```

确保 `juicefsruntime-controller`、`dataset-controller`、`fluid-webhook` 的 pod 以及若干 `csi-nodeplugin` pod 正常运行。

## 新建工作环境

```shell
$ mkdir <any-path>/juicefs
$ cd <any-path>/juicefs
```

## 运行示例

在使用 JuiceFS 之前，您需要提供元数据服务（如 redis）及对象存储服务（如 minio）的参数，并创建对应的 secret:

```shell
kubectl create secret generic jfs-secret \
    --from-literal=access-key=<accesskey> \
    --from-literal=secret-key=<secretkey>
```

**查看待创建的 Dataset 资源对象**

```shell
cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: jfsdemo
spec:
  mounts:
    - name: minio
      mountPoint: "juicefs:///demo"
      options:
        metaurl: "<metaurl>"
        bucket: "<bucket>"
        storage: "minio"
      encryptOptions:
        - name: access-key
          valueFrom:
            secretKeyRef:
              name: jfs-secret
              key: access-key
        - name: secret-key
          valueFrom:
            secretKeyRef:
              name: jfs-secret
              key: secret-key
EOF
```

> 说明：demo 指的是 JuiceFS 的 Subpath,是用户在 JuiceFS 文件系统中存储数据的目录
> 注意：只有 name 和 metaurl 为必填项，若 juicefs 已经 format 过，只需要填 name 和 metaurl 即可。

由于 JuiceFS 采用的是本地缓存，对应的 Dataset 只支持一个 mount，且 JuiceFS 没有 UFS，mountpoint 中可以指定需要挂载的子目录 ("juicefs:///" 为根路径)，会作为根目录挂载到容器内。

**创建 Dataset 资源对象**
```shell
$ kubectl create -f dataset.yaml
dataset.data.fluid.io/jfsdemo created
```

**查看Dataset资源对象状态**
```shell
$ kubectl get dataset jfsdemo
NAME      UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE      AGE
jfsdemo                                                                  NotBound   44s
```

如上所示，`status` 中的 `phase` 属性值为 `NotBound`，这意味着该 `Dataset` 资源对象目前还未与任何 `JuiceFSRuntime` 资源对象绑定，接下来，我们将创建一个 `JuiceFSRuntime` 资源对象。

**查看待创建的 JuiceFSRuntime 资源对象**

```shell
$ cat<<EOF >runtime.yaml
apiVersion: data.fluid.io/v1alpha1
kind: JuiceFSRuntime
metadata:
  name: jfsdemo
spec:
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 40960
        low: "0.1"
EOF
```
> 注意：JuiceFS 中 quota 的最小单位是 MiB

**创建 JuiceFSRuntime 资源对象**

```shell
$ kubectl create -f runtime.yaml
juicefsruntime.data.fluid.io/jfsdemo created
```

**检查 JuiceFSRuntime 资源对象是否已经创建**
```shell
$ kubectl get juicefsruntime
NAME      AGE
jfsdemo   34s
```

等待一段时间，让 JuiceFSRuntime 资源对象中的各个组件得以顺利启动，你会看到类似以下状态：

```shell
$ kubectl get po |grep jfs
jfsdemo-worker-mjplw                                           1/1     Running   0          4m2s
```

JuiceFSRuntime 没有 master 组件，而 Fuse 组件实现了懒启动，会在 pod 使用时再创建。

```shell
$ kubectl get juicefsruntime jfsdemo
NAME      AGE
jfsdemo   6m13s
```

然后，再查看 Dataset 状态，发现已经与 JuiceFSRuntime 绑定。

```shell
$ kubectl get dataset jfsdemo
NAME      UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
jfsdemo   4.00KiB          -        40.00GiB         -                   Bound   9m28s
```

**查看待创建的 Pod 资源对象**，其中 Pod 使用上面创建的 Dataset 的方式为指定同名的 PVC。

```yaml
$ cat<<EOF >sample.yaml
apiVersion: v1
kind: Pod
metadata:
  name: demo-app
spec:
  containers:
    - name: demo
      image: nginx
      volumeMounts:
        - mountPath: /data
          name: demo
  volumes:
    - name: demo
      persistentVolumeClaim:
        claimName: jfsdemo
EOF
```

**创建 Pod 资源对象**

```shell
$ kubectl create -f sample.yaml
pod/demo-app created
```

**检查 Pod 资源对象是否已经创建**
```shell
$ kubectl get po |grep demo
demo-app                                                       1/1     Running   0          31s
jfsdemo-fuse-fx7np                                             1/1     Running   0          31s
jfsdemo-worker-mjplw                                           1/1     Running   0          10m
```

可以看到 pod 已经创建成功，同时 JuiceFS 的 Fuse 组件也启动成功。
