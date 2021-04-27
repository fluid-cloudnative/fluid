## 1. 为什么我使用Helm安装fluid失败了？

**回答**:推荐按照[Fluid安装文档](./install.md)依次确认Fluid组件是否正常运行。

Fluid安装文档是以`Helm 3`为例进行部署的。如果您使用`Helm 3`以下的版本部署Fluid，
并且遇到了`CRD没有正常启动`的情况，这可能是因为`Helm 3`及其以上版本会在`helm install`的时候自动安装CRD，
而低版本的Helm则不会。
参见[Helm官方文档说明](https://helm.sh/docs/chart_best_practices/custom_resource_definitions/)。

在这种情况下，您需要手动安装CRD：
```bash
$ kubectl create -f fluid/crds
```

## 2. 为什么我无法删除Runtime？

**回答**:请检查相关Pod运行状态和Runtime的Events。

只要有任何活跃Pod还在使用Fluid创建的Volume，Fluid就不会完成删除操作。

如下的命令可以快速地找出这些活跃Pod，使用时把`<dataset_name>`和`<dataset_namespace>`换成自己的即可：
```bash
kubectl describe pvc <dataset_name> -n <dataset_namespace> | \
	awk '/^Mounted/ {flag=1}; /^Events/ {flag=0}; flag' | \
	awk 'NR==1 {print $3}; NR!=1 {print $1}' | \
	xargs -I {} kubectl get po {} | \
	grep -E "Running|Terminating|Pending" | \
	cut -d " " -f 1
```


## 3.为什么我运行例子[远程文件访问加速](../samples/accelerate_data_accessing.md)，执行第一次拷贝文件时会遇到`Input/output error`的错误。类似如下：

```
time cp ./pyspark-2.4.6.tar.gz /tmp/
cp: error reading ‘./pyspark-2.4.6.tar.gz’: Input/output error
cp: failed to extend ‘/tmp/pyspark-2.4.6.tar.gz’: Input/output error

real	3m15.795s
user	0m0.001s
sys	0m0.092s
```

这个原因是什么？

**回答**: 这个例子的目的是让用户在无需搭建UFS（underlayer file system）的情况下，利用现有的基于Http协议的Apache软件镜像下载地址演示数据拷贝加速的能力。而实际场景中，一般不会使用WebUFS的实现。但是这个例子有三个限制：

1.Apache软件镜像下载地址的可用性和访问速度

2.WebUFS来源于Alluxio的社区贡献，并不是最优实现。比如实现并不是基于offset的断点续载，这就导致每次远程读操作都需要触发WebUFS大量数据块读

3.由于拷贝行为基于Fuse实现，每一次Fuse的chunk读由于Linux Kernel的上限都是128KB；从而导致文件越大，在初次拷贝时，就会触发大量的读操作

针对该问题，我们提拱了优化的方案：

1.配置读时，将Block size和chunk size读设置的大于文件的大小，这样就可以避免Fuse实现中频繁读的影响。

```
alluxio.user.block.size.bytes.default: 256MB
alluxio.user.streaming.reader.chunk.size.bytes: 256MB
alluxio.user.local.reader.chunk.size.bytes: 256MB
alluxio.worker.network.reader.buffer.size: 256MB
```

2.为了保障目标文件可以被下载成功，可以调整block下载的超时。例子中的超时时间是5分钟，如果您的网络状况不佳，可以酌情设置更长时间。

```
alluxio.user.streaming.data.timeout: 300sec
```

3.您可以尝试手动加载该文件

```
kubectl exec -it hbase-master-0 bash
time alluxio fs  distributedLoad --replication 1 /
```

## 4. 为什么我在创建任务挂载 Runtime 创建的 PVC 的时候出现 `driver name fuse.csi.fluid.io not found in the list of registered CSI drivers` 错误？

**回答**:请查看任务被调度节点所在的 kubelet 配置是否为默认`/var/lib/kubelet`。

首先通过命令查看Fluid的CSI组件是否正常

如下的命令可以快速地找出Pod，使用时把`<node_name>`和`<fluid_namespace>`换成自己的即可：
```bash
kubectl get pod -n <fluid_namespace> | grep <node_name>

# <pod_name> 为上一步pod名
kubectl logs <pod_name> node-driver-registrar -n <fluid_namespace>
kubectl logs <pod_name> plugins -n <fluid_namespace>
```

如果上述步骤的Log无错误，请查看csidriver对象是否存在：
```
kubectl get csidriver
```
如果csidriver对象存在，请查看查看csi注册节点是否包含`<node_name>`：
```
kubectl get csinode | grep <node_name>
```
如果上述命令无输出，查看任务被调度节点所在的 kubelet 配置是否为默认`/var/lib/kubelet`。因为Fluid的CSI组件通过固定地址的socket注册到kubelet，默认为`--csi-address=/var/lib/kubelet/csi-plugins/fuse.csi.fluid.io/csi.sock --kubelet-registration-path=/var/lib/kubelet/csi-plugins/fuse.csi.fluid.io/csi.sock`。


## 5. 为什么更新了fluid后，使用 `kubectl get` 查询更新前创建的dataset，发现相比新建的dataset缺少了某些字段？

**回答**:由于我们在fluid的升级过程中可能更新了CRD，你在旧版本创建的dataset，会将CRD中新增的字段设置为空
例如，如果你从v0.4或更早版本升级，那时候的dataset没有FileNum字段
更新fluid后，如果你使用 `kubectl get` 命令，无法查询到该dataset的FileNum

你可以重建dataset，新建的dataset会正常显示这些字段

## 6. 为什么在运行示例 [Nonroot access](../samples/nonroot_access.md)时，遇到了 mkdir 权限被拒绝的问题？

**回答**:在非 root 用户情况下，你首先必须要检查是否将正确的用户信息传递给了 runtime。其次你应该检查 Alluxio master pod 的状态，并使用 journalctl 去查看 Alluxio master pod 节点对应 kubelet 的日志。
当将 hostpath 挂载到容器中，可能会造成无法创建文件的问题，因此我们必须要去检查 root 是否具有权限。例如在如下的情况中 root 是有权限使用 /dir。
```
$ stat /dir
  File: ‘/dir’
  Size: 32              Blocks: 0          IO Block: 4096   directory
Device: fd00h/64768d    Inode: 83          Links: 3
Access: (0755/drwxr-xr-x)  Uid: (    0/    root)   Gid: (    0/    root)
Access: 2021-04-14 23:35:47.928805350 +0800
Modify: 2021-01-19 00:16:21.539559082 +0800
Change: 2021-01-19 00:16:21.539559082 +0800
 Birth: -

```

## 7. 为什么在应用程序中使用 PVC 时会产生了 Volume Attachment 超时问题？
**回答**:Volume Attachment 超时问题是 Kubelet 进行请求 CSI Driver 时未收到 CSI Driver 的响应而造成的超时。
该问题是由于 Fluid 的 CSI Driver 没有安装，或者kubelet没有访问 CSI Driver 的权限导致的。
由于 CSI Driver 是由 Kubelet 进行回调，但是如果 Fluid 没有安装 CSI Driver 或者 Kubelet 没有权限查看 CSI Driver，就会导致 CSI Plugin 没有被正确触发。

首先需要使用命令`kubectl get csidriver`查看是否安装了 CSI Driver。
如果没有安装，使用命令`kubectl apply -f charts/fluid/fluid/templates/csi/driver.yaml`进行安装，然后观察 PVC 是否成功挂载到应用程序中。
如果仍未能解决，使用`export KUBECONFIG=/etc/kubernetes/kubelet.conf && kubectl get csidriver`来查看 Kubelet 能够具有权限看到 CSI Driver。
