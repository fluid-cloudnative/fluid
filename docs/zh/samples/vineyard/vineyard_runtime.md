# 如何在 Fluid 中使用 Vineyard Runtime

## 背景介绍

Vineyard 是一个开源的内存数据管理系统，旨在提供高性能的数据共享和数据交换。Vineyard 通过将数据存储在共享内存中，实现了数据的零拷贝共享，从而提供了高性能的数据共享和数据交换能力。

如何使用 Vineyard 可以参考文档 [Vineyard 快速上手指南](https://v6d.io/notes/getting-started.html)。

## 安装 Fluid

您可以从 [Fluid Releases](https://github.com/fluid-cloudnative/fluid/releases) 下载最新的 Fluid 安装包。参考 [安装文档](../../userguide/install.md) 完成安装。并检查 Fluid 各组件正常运行：

```shell
$ kubectl get po -n fluid-system
NAME                                         READY   STATUS              RESTARTS   AGE
csi-nodeplugin-fluid-56d44                   2/2     Running             0          106s
csi-nodeplugin-fluid-5l78j                   2/2     Running             0          106s
csi-nodeplugin-fluid-5mghb                   2/2     Running             0          106s
dataset-controller-5cd87f8b9b-t7dv2          1/1     Running             0          106s
fluid-webhook-77d44f5fbc-wttzl               1/1     Running             0          106s
```

确保 `dataset-controller`、`fluid-webhook` 的 pod 以及若干 `csi-nodeplugin` pod 正常运行。 `vineyard-runtime-controller` 会在使用 VineyardRuntime 的时候动态创建。

## 创建Vineyard Runtime 及 Dataset

```shell
$ cat <<EOF | kubectl apply -f -
apiVersion: data.fluid.io/v1alpha1
kind: VineyardRuntime
metadata:
  name: vineyard
spec:
  replicas: 2
  tieredstore:
    levels:
    - mediumtype: MEM
      quota: 20Gi
---
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: vineyard
EOF
```

在 VineyardRuntime 中:

- `spec.replicas`：指定 Vineyard Worker 的副本数；
- `spec.tieredstore`：指定 Vineyard Worker 的存储配置，包括存储层级和存储容量。这里配置了一个内存存储层级，容量为 20Gi。

在 Dataset 中:

- `metadata.name`：指定 Dataset 的名称，需要与 VineyardRuntime 中的 `metadata.name` 保持一致。


检查 `Vineyard Runtime` 是否创建成功：

```shell
$ kubectl get vineyardRuntime vineyard 
NAME       MASTER PHASE   WORKER PHASE   FUSE PHASE   AGE
vineyard   Ready          PartialReady   Ready        3m4s
```

再查看 `Vineyard Dataset` 的状态，发现已经与 `Vineyard Runtime` 绑定：

```shell
$ kubectl get dataset vineyard
NAME       UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
vineyard                                                                  Bound   3m9s
```

## 创建一个应用 Pod 并挂载 Vineyard Dataset

```shell
$ cat <<EOF | kubectl apply -f -
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
        claimName: vineyard
EOF
```

检查 Pod 资源对象是否创建成功：

```shell
$ kubectl get pod demo-app
NAME       READY   STATUS    RESTARTS   AGE
demo-app   1/1     Running   0          25s
```

查看 Vineyard FUSE 的状态：

```shell
$ kubectl get po | grep vineyard-fuse
vineyard-fuse-9dv4d                    1/1     Running   0               1m20s
```
