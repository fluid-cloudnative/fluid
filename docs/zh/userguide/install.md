# 在Kubernetes集群上部署Fluid

## 前提条件

- Git
- Kubernetes集群（version >= 1.14）, 并且支持CSI功能
- kubectl（version >= 1.14）
- Helm（version >= 3.0）

接下来的文档假设您已经配置好上述所有环境。

对于`kubectl`的安装和配置，请参考[此处](https://kubernetes.io/docs/tasks/tools/install-kubectl/)。

对于Helm 3的安装和配置，请参考[此处](https://v3.helm.sh/docs/intro/install/)。

## Fluid安装步骤

### 获取Fluid Chart

您可以从[Fluid Releases](https://github.com/fluid-cloudnative/fluid/releases)下载最新的Fluid安装包。


### 使用Helm安装Fluid

创建命名空间：

```shell
$ kubectl create ns fluid-system
```

安装Fluid：

```shell
$ helm install fluid fluid.tgz
NAME: fluid
LAST DEPLOYED: Fri Jul 24 16:10:18 2020
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

> `helm install`命令的一般格式是`helm install <RELEASE_NAME> <SOURCE>`，在上面的命令中，第一个`fluid`指定了安装的release名字，这可以自行更改，第二个`fluid.tgz`指定了helm chart所在路径。

### 使用Helm更新Fluid

更新前，建议确保AlluxioRuntime资源对象中的各个组件得以顺利启动，也就是类似以下状态：

```shell
$ kubectl get pod
NAME                 READY   STATUS    RESTARTS   AGE
hbase-fuse-chscz     1/1     Running   0          9h
hbase-fuse-fmhr5     1/1     Running   0          9h
hbase-master-0       2/2     Running   0          9h
hbase-worker-bdbjg   2/2     Running   0          9h
hbase-worker-rznd5   2/2     Running   0          9h
```
如果您此前已经安装过旧版本的Fluid，可以使用Helm进行更新。

```shell
$ helm upgrade fluid fluid-0.4.0.tgz
Release "fluid" has been upgraded. Happy Helming!
NAME: fluid
LAST DEPLOYED: Wed Nov  4 09:19:58 2020
NAMESPACE: default
STATUS: deployed
REVISION: 2
TEST SUITE: None
```
> 我们目前只尝试过从v0.3更新到v0.4，如果您从更旧的版本直接升级到v0.4，可能会出现未知类型的错误。

### 检查各组件状态

**查看Fluid使用的CRD:**

```shell
$ kubectl get crd | grep data.fluid.io
alluxiodataloads.data.fluid.io          2020-07-24T06:54:50Z
alluxioruntimes.data.fluid.io           2020-07-24T06:54:50Z
datasets.data.fluid.io                  2020-07-24T06:54:50Z
```

**查看各Pod的状态:**

```shell
$ kubectl get pod -n fluid-system
NAME                                         READY   STATUS    RESTARTS   AGE
alluxioruntime-controller-5dfb5c7966-mkgzb   1/1     Running   0          2d1h
csi-nodeplugin-fluid-64h69                   2/2     Running   0          2d1h
csi-nodeplugin-fluid-tc7fx                   2/2     Running   0          2d1h
dataset-controller-7c4bc68b96-26mcb          1/1     Running   0          2d1h
```

如果Pod状态如上所示，那么Fluid就可以正常使用了！

**查看各Pod内程序的版本:**

csi-nodeplugin、alluxioruntime-controller、dataset-controller在启动时，会将自身的版本信息打印到日志中。  
如果您使用我们提供的charts进行安装，它们的版本应该是完全一致的。  
如果您是手动安装部署，它们的版本可能不完全一致，可以分别依次查看：  
```bash
$ kubectl logs csi-nodeplugin-fluid-tc7fx -c plugins  -n fluid-system | head -n 9 | tail -n 6
$ kubectl logs alluxioruntime-controller-5dfb5c7966-mkgzb -n fluid-system | head -n 6
$ kubectl logs dataset-controller-7c4bc68b96-26mcb  -n fluid-system | head -n 6
```
打印出的日志如下格式：
```bash
2020/10/27 10:16:02 BuildDate: 2020-10-26_14:04:22
2020/10/27 10:16:02 GitCommit: f2c3a3fa1335cb0384e565f17a4f3284a6507cef
2020/10/27 10:16:02 GitTreeState: dirty
2020/10/27 10:16:02 GoVersion: go1.14.2
2020/10/27 10:16:02 Compiler: gc
2020/10/27 10:16:02 Platform: linux/amd64
```
若Pod打印的日志已经被清理掉，可以执行下列命令查看版本： 
```bash
$ kubectl exec csi-nodeplugin-fluid-tc7fx -c plugins  fluid-csi version -n fluid-system
$ kubectl exec alluxioruntime-controller-5dfb5c7966-mkgzb alluxioruntime-controller version -n fluid-system
$ kubectl exec dataset-controller-7c4bc68b96-26mcb dataset-controller version -n  fluid-system 
```

### Fluid使用示例

有关Fluid的使用示例,可以参考我们提供的示例文档:
- [远程文件访问加速](../samples/accelerate_data_accessing.md)
- [数据缓存亲和性调度](../samples/data_co_locality.md)
- [用Fluid加速机器学习训练](../samples/machinelearning.md)

### 卸载Fluid

```shell
$ helm delete fluid
$ kubectl delete -f fluid/crds
$ kubectl delete ns fluid-system
```

> `helm delete`命令中的`fluid`对应安装时指定的<RELEASE_NAME>。


### 高级配置

在一些特定的云厂商实现下， 默认mount根目录`/alluxio-mnt`是不可写的,因此需要修改目录位置

```
helm install fluid --set runtime.mountRoot=/var/lib/docker/alluxio-mnt fluid
```

