# 示例 - 如何开启 FUSE 自动恢复能力

## 安装

您可以从 [Fluid Releases](https://github.com/fluid-cloudnative/fluid/releases) 下载最新的 Fluid 安装包。

在 Fluid 的安装 chart values.yaml 中将 `csi.featureGates` 设置为 `FuseRecovery=true`，表示开启 FUSE 自动恢复功能。
再参考 [安装文档](../guide/install.md) 完成安装。并检查 Fluid 各组件正常运行（这里以 JuiceFSRuntime 为例）：

```shell
$ kubectl -n fluid-system get po
NAME                                        READY   STATUS    RESTARTS   AGE
csi-nodeplugin-fluid-2gtsz                  2/2     Running   0          20m
csi-nodeplugin-fluid-2h79g                  2/2     Running   0          20m
csi-nodeplugin-fluid-sc459                  2/2     Running   0          20m
dataset-controller-57fb4569cd-k2jb7         1/1     Running   0          20m
fluid-webhook-844dcb995f-nfmjl              1/1     Running   0          20m
juicefsruntime-controller-7d9c964b4-jnbtf   1/1     Running   0          20m
```

通常来说，你会看到一个名为 `dataset-controller` 的 Pod、一个名为 `juicefsruntime-controller` 的 Pod、一个名为 `fluid-webhook` 的 Pod
和多个名为 `csi-nodeplugin` 的 Pod 正在运行。其中，`csi-nodeplugin` 这些 Pod 的数量取决于你的 Kubernetes 集群中结点的数量。

## 运行示例

**为 namespace 开启 webhook**

FUSE 挂载点自动恢复功能需要 pod 的 mountPropagation 设置为 `HostToContainer` 或 `Bidirectional`，才能将挂载点信息在容器和宿主机之间传递。而 `Bidirectional` 需要容器为特权容器。
Fluid webhook 提供了自动将 pod 的 mountPropagation 设置为 `HostToContainer`，为了开启该功能，需要将对应的 namespace 打上 `fluid.io/enable-injection=true` 的标签。操作如下：

```shell
$ kubectl patch ns default -p '{"metadata": {"labels": {"fluid.io/enable-injection": "true"}}}'
namespace/default patched
$ kubectl get ns default --show-labels
NAME      STATUS   AGE     LABELS
default   Active   4d12h   fluid.io/enable-injection=true,kubernetes.io/metadata.name=default
```

**创建 dataset 和 runtime**

针对不同类型的 runtime 创建相应的 Runtime 资源，以及同名的 Dataset。这里以 JuiceFSRuntime 为例，具体可参考 [文档](./juicefs_runtime.md)，如下：

```shell
$ kubectl get juicefsruntime
NAME      WORKER PHASE   FUSE PHASE   AGE
jfsdemo   Ready          Ready        2m58s
$ kubectl get dataset
NAME      UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
jfsdemo   [Calculating]    N/A                       N/A                 Bound   2m55s
```

**创建 Pod 资源对象**

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
$ kubectl create -f sample.yaml
pod/demo-app created
```

**查看 Pod 是否创建，并检查其 mountPropagation**

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

**删除 FUSE pod**

删除 FUSE pod 后，并等待其重启：

```shell
$ kubectl delete po jfsdemo-fuse-g9pvp
pod "jfsdemo-fuse-g9pvp" deleted
$ kubectl get po
NAME                 READY   STATUS    RESTARTS   AGE
demo-app             1/1     Running   0          5m7s
jfsdemo-fuse-bdsdt   1/1     Running   0          6s
jfsdemo-worker-0     1/1     Running   0          7m56s
````

新的 FUSE pod 创建后，再查看 demo pod 中的挂载点情况：

```shell
$ kubectl exec -it demo-app bash
kubectl exec [POD] [COMMAND] is DEPRECATED and will be removed in a future version. Use kubectl exec [POD] -- [COMMAND] instead.
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
  Normal  FuseRecoverSucceed  2m34s (x5 over 11m)  FuseRecover  Fuse recover /var/lib/kubelet/pods/6c1e0318-858b-4ead-976b-37ccce26edfe/volumes/kubernetes.io~csi/default-jfsdemo/mount succeed
```

可以看到 Dataset 的 event 有一条 `FuseRecover` 的事件，表明 Fluid 已经对挂载做过一次恢复操作。

## 注意

在 FUSE pod crash 的时候，挂载点恢复的时间依赖 FUSE pod 自身的恢复以及 csi 轮询 kubelet 的周期大小（环境变量 `RECOVER_FUSE_PERIOD`），在恢复之前挂载点会出现 `Transport endpoint is not connected` 的错误，这是符合预期的。
另外，挂载点恢复是通过重复 bind 的方法实现的，对于 FUSE pod crash 之前应用已经打开的文件描述符，挂载点恢复后该 fd 亦不可恢复，需要应用自身实现错误重试，增强应用自身的鲁棒性。
