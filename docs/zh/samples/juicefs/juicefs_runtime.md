# 示例 - 如何在 Fluid 中使用 JuiceFS

## 背景介绍

JuiceFS 是一款面向云环境设计的开源高性能共享文件系统，提供完备的 POSIX 兼容性，可将海量低价的云存储作为本地磁盘使用，亦可同时被多台主机同时挂载读写。

如何使用 JuiceFS 可以参考文档 [JuiceFS 快速上手指南](https://juicefs.com/docs/zh/community/quick_start_guide)。

## 安装

您可以从 [Fluid Releases](https://github.com/fluid-cloudnative/fluid/releases) 下载最新的 Fluid 安装包。参考 [安装文档](../../userguide/install.md) 完成安装。并检查 Fluid 各组件正常运行：

```shell
$ kubectl get po -n fluid-system
NAME                                         READY   STATUS              RESTARTS   AGE
csi-nodeplugin-fluid-ctc4l                   2/2     Running             0          113s
csi-nodeplugin-fluid-k7cqt                   2/2     Running             0          113s
csi-nodeplugin-fluid-x9dfd                   2/2     Running             0          113s
dataset-controller-57ddd56b54-9vd86          1/1     Running             0          113s
fluid-webhook-84467465f8-t65mr               1/1     Running             0          113s
```

确保 `dataset-controller`、`fluid-webhook` 的 pod 以及若干 `csi-nodeplugin` pod 正常运行。 `juicefs-runtime-controller` 会在使用 JuiceFSRuntime 的时候动态创建。

## 新建工作环境

```shell
$ mkdir <any-path>/juicefs
$ cd <any-path>/juicefs
```

## 运行示例

使用 JuiceFS 社区版和云服务版所需字段有所区别，下面分别介绍其使用方法：

### 社区版

在使用 JuiceFS 之前，您需要提供元数据服务（如 Redis）及对象存储服务（如 MinIO）的参数，并创建对应的 secret:

```shell
kubectl create secret generic jfs-secret \
    --from-literal=metaurl=redis://192.168.169.168:6379/1 \
    --from-literal=access-key=<accesskey> \
    --from-literal=secret-key=<secretkey>
```

其中：

- `metaurl`：元数据服务的访问 URL (比如 Redis)。更多信息参考[这篇文档](https://juicefs.com/docs/zh/community/databases_for_metadata/) 。
- `access-key`：对象存储的 access key。
- `secret-key`：对象存储的 secret key。

**查看待创建的 `Dataset` 资源对象**

```yaml
$ cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: jfsdemo
spec:
  mounts:
    - name: minio
      mountPoint: "juicefs:///demo"
      options:
        bucket: "<bucket>"
        storage: "minio"
      encryptOptions:
        - name: metaurl
          valueFrom:
            secretKeyRef:
              name: jfs-secret
              key: metaurl
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

其中：

- `mountPoint`：指的是 JuiceFS 的子目录，是用户在 JuiceFS 文件系统中存储数据的目录，以 `juicefs://` 开头；如 `juicefs:///demo` 为 JuiceFS 文件系统的 `/demo` 子目录。
- `bucket`：Bucket URL。例如使用 S3 作为对象存储，bucket 为 `https://myjuicefs.s3.us-east-2.amazonaws.com`；更多信息参考[这篇文档](https://juicefs.com/docs/zh/community/how_to_setup_object_storage/) 。
- `storage`：对象存储类型，比如 `s3`，`gs`，`oss`。更多信息参考[这篇文档](https://juicefs.com/docs/zh/community/how_to_setup_object_storage/) 。

> **注意**：只有 `name` 和 `metaurl` 为必填项，若 JuiceFS 已经格式化过，只需要填写 `name` 和 `metaurl` 即可。

由于 JuiceFS 采用的是本地缓存，对应的 `Dataset` 只支持一个 mount，且 JuiceFS 没有 UFS，`mountPoint` 中可以指定需要挂载的子目录（`juicefs:///` 为根路径），会作为根目录挂载到容器内。

**创建 `Dataset` 资源对象**
```shell
$ kubectl create -f dataset.yaml
dataset.data.fluid.io/jfsdemo created
```

**查看 `Dataset` 资源对象状态**
```shell
$ kubectl get dataset jfsdemo
NAME      UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE      AGE
jfsdemo                                                                  NotBound   44s
```

如上所示，`status` 中的 `phase` 属性值为 `NotBound`，这意味着该 `Dataset` 资源对象目前还未与任何 `JuiceFSRuntime` 资源对象绑定，接下来，我们将创建一个 `JuiceFSRuntime` 资源对象。

**查看待创建的 `JuiceFSRuntime` 资源对象**

```yaml
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
        quota: 40Gi
        low: "0.1"
EOF
```

**创建 `JuiceFSRuntime` 资源对象**

```shell
$ kubectl create -f runtime.yaml
juicefsruntime.data.fluid.io/jfsdemo created
```

**检查 `JuiceFSRuntime` 资源对象是否已经创建**
```shell
$ kubectl get juicefsruntime
NAME      AGE
jfsdemo   34s
```

等待一段时间，让 `JuiceFSRuntime` 资源对象中的各个组件得以顺利启动，你会看到类似以下状态：

```shell
$ kubectl get po |grep jfs
jfsdemo-worker-0                                          1/1     Running   0          4m2s
```

`JuiceFSRuntime` 没有 master 组件，而 FUSE 组件实现了懒启动，会在 pod 使用时再创建。

```shell
$ kubectl get juicefsruntime jfsdemo
NAME      AGE
jfsdemo   6m13s
```

然后，再查看 `Dataset` 状态，发现已经与 `JuiceFSRuntime` 绑定。

```shell
$ kubectl get dataset jfsdemo
NAME      UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
jfsdemo   4.00KiB          -        40.00GiB         -                   Bound   9m28s
```

**查看待创建的 Pod 资源对象**，其中 Pod 使用上面创建的 `Dataset` 的方式为指定同名的 PVC。

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
jfsdemo-worker-0                                               1/1     Running   0          10m
```

可以看到 pod 已经创建成功，同时 JuiceFS 的 FUSE 组件也启动成功。

### 云服务版

在使用 JuiceFS 之前，您需要提供 JuiceFS 控制台管理的 Token（更多信息参考[这篇文档](https://juicefs.com/docs/zh/cloud/metadata/#%E4%BB%A4%E7%89%8C%E7%AE%A1%E7%90%86)），及对象存储服务（如 MinIO）的参数，并创建对应的 secret：

```shell
kubectl create secret generic jfs-secret \
    --from-literal=token=<token> \
    --from-literal=access-key=<accesskey>  \
    --from-literal=secret-key=<secretkey>
```

其中：

- `token`：JuiceFS 管理 token。更多信息参考[这篇文档](https://juicefs.com/docs/zh/cloud/metadata/#%E4%BB%A4%E7%89%8C%E7%AE%A1%E7%90%86)。
- `access-key`：对象存储的 access key。
- `secret-key`：对象存储的 secret key。

**查看待创建的 `Dataset` 资源对象**

```yaml
$ cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: jfsdemo
spec:
  mounts:
    - name: minio
      mountPoint: "juicefs:///demo"
      options:
        bucket: "<bucket>"
      encryptOptions:
        - name: token
          valueFrom:
            secretKeyRef:
              name: jfs-secret
              key: token
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

其中：

- `name`：需要与在 JuiceFS 控制台创建的 volume 名一致。
- `mountPoint`：指的是 JuiceFS 的子目录，是用户在 JuiceFS 文件系统中存储数据的目录，以 `juicefs://` 开头；如 `juicefs:///demo` 为 JuiceFS 文件系统的 `/demo` 子目录。
- `bucket`：Bucket URL。例如使用 S3 作为对象存储，bucket 为 `https://myjuicefs.s3.us-east-2.amazonaws.com`；更多信息参考[这篇文档](https://juicefs.com/docs/zh/community/how_to_setup_object_storage/) 。

> **注意**：其中 `name` 和 `token` 为必填项。

JuiceFS 对应的 `Dataset` 只支持一个 mount，且 JuiceFS 没有 UFS，`mountPoint` 中可以指定需要挂载的子目录（`juicefs:///` 为根路径），会作为根目录挂载到容器内。

**创建 `Dataset` 资源对象**
```shell
$ kubectl create -f dataset.yaml
dataset.data.fluid.io/jfsdemo created
```

**查看 `Dataset` 资源对象状态**
```shell
$ kubectl get dataset jfsdemo
NAME      UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE      AGE
jfsdemo                                                                  NotBound   44s
```

如上所示，`status` 中的 `phase` 属性值为 `NotBound`，这意味着该 `Dataset` 资源对象目前还未与任何 `JuiceFSRuntime` 资源对象绑定，接下来，我们将创建一个 `JuiceFSRuntime` 资源对象。

**查看待创建的 `JuiceFSRuntime` 资源对象**

```yaml
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
        quota: 40Gi
        low: "0.1"
EOF
```

**创建 `JuiceFSRuntime` 资源对象**

```shell
$ kubectl create -f runtime.yaml
juicefsruntime.data.fluid.io/jfsdemo created
```

**检查 `JuiceFSRuntime` 资源对象是否已经创建**
```shell
$ kubectl get juicefsruntime
NAME      AGE
jfsdemo   34s
```

等待一段时间，让 `JuiceFSRuntime` 资源对象中的各个组件得以顺利启动，你会看到类似以下状态：

```shell
$ kubectl get po |grep jfs
jfsdemo-worker-0                                           1/1     Running   0          4m2s
```

`JuiceFSRuntime` 没有 master 组件，而 FUSE 组件实现了懒启动，会在 pod 使用时再创建；JuiceFS 云服务版的 worker 提供了分布式独立缓存。

```shell
$ kubectl get juicefsruntime jfsdemo
NAME      AGE
jfsdemo   6m13s
```

然后，再查看 `Dataset` 状态，发现已经与 `JuiceFSRuntime` 绑定。

```shell
$ kubectl get dataset jfsdemo
NAME      UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
jfsdemo   4.00KiB          -        40.00GiB         -                   Bound   9m28s
```

**查看待创建的 Pod 资源对象**，其中 Pod 使用上面创建的 `Dataset` 的方式为指定同名的 PVC。

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
demo-app                 1/1     Running   0          83s
jfsdemo-fuse-9xgkc       1/1     Running   0          82s
jfsdemo-worker-0         1/1     Running   0          4m56s
```

可以看到 pod 已经创建成功，同时 JuiceFS 的 FUSE 组件也启动成功。

