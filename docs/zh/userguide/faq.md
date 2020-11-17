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

如果上述步骤的Log无错误，请查看任务被调度节点所在的 kubelet 配置是否为默认`/var/lib/kubelet`。因为Fluid的CSI组件通过固定地址的socket注册到kubelet，默认为`--csi-address=/var/lib/kubelet/csi-plugins/fuse.csi.fluid.io/csi.sock --kubelet-registration-path=/var/lib/kubelet/csi-plugins/fuse.csi.fluid.io/csi.sock`。