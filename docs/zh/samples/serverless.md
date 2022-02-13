# 示例 - 如何运行在Serverless环境中

本示例以开源框架Knative为例子，演示如何在Serverless环境中通过Fluid进行统一的数据加速，本例子以AlluxioRuntime为例，实际上Fluid支持所有已经支持的Runtime运行在Serverless环境。


## 安装

1.根据[Knative文档](https://knative.dev/docs/install/serving/install-serving-with-yaml/)安装Knative Serving v1.2，需要开启[kubernetes.Deploymentspec-persistent-volume-claim](https://github.com/knative/serving/blob/main/config/core/configmaps/features.yaml#L156)和[kubernetes.podspec-persistent-volume-write](https://github.com/knative/serving/blob/main/config/core/configmaps/features.yaml#L161)。

检查 Knative的组件是否正常运行

```
kubectl get Deployments -n knative-serving
```

> 注：本文只是作为演示目的，关于Knative的生产系统安装请参考Knative文档最佳实践进行部署。另外由于Knative的容器镜像都在gcr.io镜像仓库，请确保镜像可达。
如果您使用的是阿里云，您也可以直接使用[阿里云ACK的托管服务](https://help.aliyun.com/document_detail/121508.html)降低配置Knative的复杂度。

2.下载、安装Fluid最新版

```
git clone https://github.com/fluid-cloudnative/fluid.git
cd fluid/charts
kubectl create ns fluid-system
helm install --set webhook.enabled=true  fluid fluid
```

检查 Fluid 各组件正常运行（这里以 AlluxioRuntime 为例）：

```shell
$ kubectl get deploy -n fluid-system
NAME                        READY   UP-TO-DATE   AVAILABLE   AGE
alluxioruntime-controller   1/1     1            1           18m
dataset-controller          1/1     1            1           18m
fluid-webhook               1/1     1            1           18m
```

通常来说，你会看到一个名为 `dataset-controller` 的 Deployment、一个名为 `AlluxioRuntime-controller` 的 Deployment、一个名为 `fluid-webhook` 的 Deployment。

## 配置

**为namespace添加标签**

为namespace添加标签fluid.io/enable-injection后，可以开启此namespace下Pod的调度优化功能

```bash
$ kubectl label namespace default fluid.io/enable-injection=true
```


## 运行示例

**为 namespace 开启 webhook**

FUSE 挂载点自动恢复功能需要 Deployment 的 mountPropagation 设置为 `HostToContainer` 或 `Bidirectional`，才能将挂载点信息在容器和宿主机之间传递。而 `Bidirectional` 需要容器为特权容器。
Fluid webhook 提供了自动将 Deployment 的 mountPropagation 设置为 `HostToContainer`，为了开启该功能，需要将对应的 namespace 打上 `fluid.io/enable-injection=true` 的标签。操作如下：

```shell
$ kubectl patch ns default -p '{"metadata": {"labels": {"fluid.io/enable-injection": "true"}}}'
namespace/default patched
$ kubectl get ns default --show-labels
NAME      STATUS   AGE     LABELS
default   Active   4d12h   fluid.io/enable-injection=true,kubernetes.io/metadata.name=default
```

**创建 dataset 和 runtime**

针对不同类型的 runtime 创建相应的 Runtime 资源，以及同名的 Dataset。这里以 AlluxioRuntime 为例，具体可参考 [文档](juicefs_runtime.md)，如下：

```shell
$ kubectl get AlluxioRuntime
NAME      WORKER PHASE   FUSE PHASE   AGE
jfsdemo   Ready          Ready        2m58s
$ kubectl get dataset
NAME      UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
jfsdemo   [Calculating]    N/A                       N/A                 Bound   2m55s
```

**创建 Deployment 资源对象**

```yaml
$ cat<<EOF >sample.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: shared-data
spec:
  mounts:
    - mountPoint: https://mirrors.bit.edu.cn/apache/hbase/stable/
      name: hbase
      path: "/"
  accessModes:
    - ReadOnlyMany
---
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: shared-data
spec:
  replicas: 2
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 2Gi
        high: "0.95"
        low: "0.7"
  EOF
$ kubectl create -f sample.yaml
Deployment/demo-app created
```

**查看 Deployment 是否创建，并检查其 mountPropagation**

```shell
$ kubectl get po |grep demo
demo-app             1/1     Running   0          96s
jfsdemo-fuse-g9pvp   1/1     Running   0          95s
jfsdemo-worker-0     1/1     Running   0          4m25s
$ kubectl get po demo-app -oyaml |grep volumeMounts -A 3
    volumeMounts:
    - mountPath: /data
      mountPropagation: HostToContainer
      name: demo
```

## 测试 FUSE 挂载点自动恢复

**删除 FUSE Deployment**

删除 FUSE Deployment 后，并等待其重启：

```shell
$ kubectl delete po jfsdemo-fuse-g9pvp
Deployment "jfsdemo-fuse-g9pvp" deleted
$ kubectl get po
NAME                 READY   STATUS    RESTARTS   AGE
demo-app             1/1     Running   0          5m7s
jfsdemo-fuse-bdsdt   1/1     Running   0          6s
jfsdemo-worker-0     1/1     Running   0          7m56s
````

新的 FUSE Deployment 创建后，再查看 demo Deployment 中的挂载点情况：

```shell
$ kubectl exec -it demo-app bash
kubectl exec [Deployment] [COMMAND] is DEPRECATED and will be removed in a future version. Use kubectl exec [Deployment] -- [COMMAND] instead.
[root@demo-app /]# df -h
Filesystem      Size  Used Avail Use% Mounted on
overlay         100G  9.4G   91G  10% /
tmpfs            64M     0   64M   0% /dev
tmpfs           2.0G     0  2.0G   0% /sys/fs/cgroup
JuiceFS:minio   1.0P   64K  1.0P   1% /data
/dev/sdb1       100G  9.4G   91G  10% /etc/hosts
shm              64M     0   64M   0% /dev/shm
tmpfs           3.8G   12K  3.8G   1% /run/secrets/kubernetes.io/serviceaccount
tmpfs           2.0G     0  2.0G   0% /proc/acpi
tmpfs           2.0G     0  2.0G   0% /proc/scsi
tmpfs           2.0G     0  2.0G   0% /sys/firmware
```

可以看到，容器中没有出现 `Transport endpoint is not connected` 的报错，表明挂载点已经恢复。

**查看 dataset 的 event**

```shell
$ kubectl describe dataset jfsdemo
Name:         jfsdemo
Namespace:    default
...
Events:
  Type    Reason              Age                  From         Message
  ----    ------              ----                 ----         -------
  Normal  FuseRecoverSucceed  2m34s (x5 over 11m)  FuseRecover  Fuse recover /var/lib/kubelet/Deployments/6c1e0318-858b-4ead-976b-37ccce26edfe/volumes/kubernetes.io~csi/default-jfsdemo/mount succeed
```

可以看到 Dataset 的 event 有一条 `FuseRecover` 的事件，表明 Fluid 已经对挂载做过一次恢复操作。

## 注意

在 FUSE Deployment crash 的时候，挂载点恢复的时间依赖 FUSE Deployment 自身的恢复以及 `recoverFusePeriod` 的大小，在恢复之前挂载点会出现 `Transport endpoint is not connected` 的错误，这是符合预期的。
另外，挂载点恢复是通过重复 bind 的方法实现的，对于 FUSE Deployment crash 之前应用已经打开的文件描述符，挂载点恢复后该 fd 亦不可恢复，需要应用自身实现错误重试，增强应用自身的鲁棒性。
