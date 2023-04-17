# 在Kubernetes集群上部署Fluid

## 前提条件

- Git
- Kubernetes集群（version >= 1.16）, 并且支持CSI功能
- kubectl（version >= 1.16）
- Helm（version >= 3.5）

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
LAST DEPLOYED: Fri Sep  2 19:03:56 2022
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

> 对于Kubernetes v1.17及以下环境，请使用`helm install --set runtime.criticalFusePod=false fluid fluid.tgz`

> `helm install`命令的一般格式是`helm install <RELEASE_NAME> <SOURCE>`，在上面的命令中，第一个`fluid`指定了安装的release名字，这可以自行更改，第二个`fluid.tgz`指定了helm chart所在路径。

### 使用Helm将Fluid更新到最新版本(v0.8)

如果您此前已经安装过旧版本的Fluid，可以使用Helm进行更新。
更新前，建议确保各Runtime资源对象中的各个组件已经顺利启动完成，也就是类似以下状态：

```shell
$ kubectl get pod
NAME                 READY   STATUS    RESTARTS   AGE
hbase-fuse-chscz     1/1     Running   0          9h
hbase-fuse-fmhr5     1/1     Running   0          9h
hbase-master-0       2/2     Running   0          9h
hbase-worker-bdbjg   2/2     Running   0          9h
hbase-worker-rznd5   2/2     Running   0          9h
```

更新：
```shell
$ helm upgrade fluid fluid/
Release "fluid" has been upgraded. Happy Helming!
NAME: fluid
LAST DEPLOYED: Fri Sep  2 18:54:18 2022
NAMESPACE: default
STATUS: deployed
REVISION: 2
TEST SUITE: None
```

> 对于Kubernetes v1.17及以下环境，请使用`helm upgrade --set runtime.criticalFusePod=false fluid fluid/`

> 建议您从v0.7升级到最新版本v0.8。如果您安装的是更旧版本的Fluid，建议重新进行安装。

### 检查各组件状态

**查看Fluid使用的CRD:**

```shell
$ kubectl get crd | grep data.fluid.io
alluxioruntimes.data.fluid.io                          2022-06-28T02:43:52Z
databackups.data.fluid.io                              2022-06-28T02:43:52Z
dataloads.data.fluid.io                                2022-06-28T02:43:52Z
datasets.data.fluid.io                                 2022-06-28T02:43:52Z
goosefsruntimes.data.fluid.io                          2022-06-28T02:43:52Z
jindoruntimes.data.fluid.io                            2022-06-28T02:43:52Z
juicefsruntimes.data.fluid.io                          2022-06-28T02:43:52Z
```

**查看各Pod的状态:**

```shell
$ kubectl get pod -n fluid-system
NAME                                        READY   STATUS      RESTARTS   AGE
csi-nodeplugin-fluid-g6ggh                  2/2     Running     0          6m53s
csi-nodeplugin-fluid-tnj5r                  2/2     Running     0          5m50s
dataset-controller-5f56cc4f97-2lfqt         1/1     Running     0          6m54s
fluid-crds-upgrade-0.8.0-aa7fdca--1-gtpt9   0/1     Completed   0          7m23s
fluid-webhook-7d8c586f59-mxkwz              1/1     Running     0          6m54s
fluidapp-controller-86f5bfc4c5-ct25p        1/1     Running     0          6m54s
```

如果Pod状态如上所示，那么Fluid就可以正常使用了！

**查看各Pod内程序的版本:**

csi-nodeplugin、alluxioruntime-controller、dataset-controller在启动时，会将自身的版本信息打印到日志中。  
如果您使用我们提供的charts进行安装，它们的版本应该是完全一致的。  

可以执行下列命令查看版本： 
```bash
$ kubectl exec csi-nodeplugin-fluid-pq2zd -n fluid-system -c plugins fluid-csi version
$ kubectl exec alluxioruntime-controller-66bf8cbdf4-k6cxt -n fluid-system -- alluxioruntime-controller version
$ kubectl exec dataset-controller-558c5c7785-mtgfh -n fluid-system -- dataset-controller version
```

如果版本一致，您将看到如下信息：
```
kubectl exec dataset-controller-5f56cc4f97-2lfqt -n fluid-system -- dataset-controller version
  BuildDate: 2022-09-01_13:07:33
  GitCommit: aa7fdca4c4306762280570b7dc0c2a7c649ff785
  GitTreeState: clean
  GoVersion: go1.17.8
  Compiler: gc
  Platform: linux/amd64
```

### Fluid使用示例

有关Fluid的使用示例,可以参考我们提供的示例文档:
- [远程文件访问加速](../samples/accelerate_data_accessing.md)
- [数据缓存亲和性调度](../samples/data_co_locality.md)
- [用Fluid加速机器学习训练](../samples/machinelearning.md)

### 卸载Fluid

为了安全的卸载fluid，在卸载前，首先需要检查fluid相关的自定义资源对象是否已被清除：
```shell
kubectl get crds -o custom-columns=NAME:.metadata.name | grep data.fluid.io  | sed ':t;N;s/\n/,/;b t' | xargs kubectl get --all-namespaces
```
如果确认已经删除所有资源对象，则可以安全卸载fluid：

```shell
$ helm delete fluid
$ kubectl delete -f fluid/crds
$ kubectl delete ns fluid-system
```

> `helm delete`命令中的`fluid`对应安装时指定的<RELEASE_NAME>。


### 高级配置

1. 在一些特定的云厂商实现下， 默认mount根目录`/runtime-mnt`是不可写的,因此需要修改目录位置

```
helm install fluid --set runtime.mountRoot=/var/lib/docker/runtime-mnt fluid
```

2. 默认Fuse Recovery的功能并没有打开，如果需要开放该功能，需要按照以下配置

```
helm install fluid --set csi.featureGates='FuseRecovery=true' fluid
```
