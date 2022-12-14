# 示例 - 如何在 Fluid 中使用 EAC

## 背景介绍

EAC 是一款针对 NAS 的用户态客户端，并在提供分布式缓存的同时，保证多客户端之间的缓存一致性。

如何在 ACK（Alibaba Cloud Container Service for Kubernetes）场景中使用 EAC 可以参考文档 [开启 CNFS 文件存储客户端加速特性](https://help.aliyun.com/document_detail/440307.html)。

## 安装

您可以从 [Fluid Releases](https://github.com/fluid-cloudnative/fluid/releases) 下载最新的 Fluid 安装包。

在 Fluid 的安装 chart values.yaml 中将 `runtime.eac.enable` 设置为 `true` ，再参考 [安装文档](../userguide/install.md) 完成安装。并检查Fluid各组件正常运行：

```shell
$ kubectl get po -n fluid-system
```

确保 `eacruntime-controller`、`dataset-controller`、`fluid-webhook` 的 pod 以及若干 `csi-nodeplugin` pod 正常运行。

## 新建工作环境

```shell
$ mkdir <any-path>/eac
$ cd <any-path>/eac
```

## 运行示例

在使用 EAC 之前，您需要拥有一个[通用型 NAS](https://www.aliyun.com/product/nas?spm=5176.19720258.J_3207526240.80.e93976f4Ps3XxX)，以及一个和 NAS 处在同一 [VPC](https://www.aliyun.com/product/vpc?spm=5176.59209.J_3207526240.35.253f76b9hZAU4x) 的 [ACK](https://www.aliyun.com/product/kubernetes?spm=5176.7937172.J_3207526240.54.7f51751avPxHwi) 集群，并确保集群节点操作系统使用 Alibaba Cloud Linux 2.1903。

**查看待创建的 `Dataset` 资源对象**

```yaml
$ cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: mydemo
spec:
  mounts:
    - mountPoint: "eac://nas-mount-point-address/sub/path"
EOF
```

其中：

- `mountPoint`：指的是 EAC 的子目录，是用户在 NAS 文件系统中存储数据的目录，以 `eac://` 开头；如 `eac://nas-mount-point-address/sub/path` 为 `nas-mount-point-address` 文件系统的 `/sub/path` 子目录。

**创建 `Dataset` 资源对象**
```shell
$ kubectl create -f dataset.yaml
dataset.data.fluid.io/mydemo created
```

**查看 `Dataset` 资源对象状态**
```shell
$ kubectl get dataset mydemo
```

如上所示，`status` 中的 `phase` 属性值为 `NotBound`，这意味着该 `Dataset` 资源对象目前还未与任何 `EACRuntime` 资源对象绑定，接下来，我们将创建一个 `EACRuntime` 资源对象。

**查看待创建的 `EACRuntime` 资源对象**

```yaml
$ cat<<EOF >runtime.yaml
apiVersion: data.fluid.io/v1alpha1
kind: EACRuntime
metadata:
  name: mydemo
spec:
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: MEM
        volumeType: emptyDir
        path: /dev/eac-worker-cache-path
        quota: 2Gi
  fuse:
    properties:
      g_unas_InodeAttrExpireTimeoutSec: "100"
      g_unas_InodeEntryExpireTimeoutSec: "100"
EOF
```

**创建 `EACRuntime` 资源对象**

```shell
$ kubectl create -f runtime.yaml
```

**检查 `EACRuntime` 资源对象是否已经创建**
```shell
$ kubectl get eacruntime
```

等待一段时间，让 `EACRuntime` 资源对象中的各个组件得以顺利启动，你会看到类似以下状态：

```shell
$ kubectl get po |grep mydemo
```

`EACRuntime` 的 FUSE 组件实现了懒启动，会在 pod 使用时再创建。

然后，再查看 `Dataset` 状态，发现已经与 `EACRuntime` 绑定。

```shell
$ kubectl get dataset mydemo
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
```

**检查 StatefulSet 资源对象是否已经创建**
```shell
$ kubectl get po |grep mydemo
```

可以看到 StatefulSet 已经创建成功，同时 EAC 的 FUSE 组件也启动成功。

