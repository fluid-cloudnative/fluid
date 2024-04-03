# 在Kubernetes集群上部署Fluid

## 前提条件

- Git
- Kubernetes集群（version >= 1.18）, 并且支持CSI功能
- kubectl（version >= 1.18）
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

为您本地Helm仓库添加并且更新“fluid”源到最新版本

```shell
$ helm repo add fluid https://fluid-cloudnative.github.io/charts
$ helm repo update
```

安装Fluid：

```shell
$ helm install fluid fluid/fluid
NAME: fluid
LAST DEPLOYED: Wed May 24 18:17:16 2023
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

> 对于Kubernetes v1.17及以下环境，请使用`helm install --set runtime.criticalFusePod=false fluid fluid.tgz`

> `helm install`命令的一般格式是`helm install <RELEASE_NAME> <SOURCE>`，在上面的命令中，第一个`fluid`指定了安装的release名字，这可以自行更改，第二个`fluid.tgz`指定了helm chart所在路径。

### 使用Helm将Fluid更新到最新版本

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
$ helm upgrade fluid fluid/fluid
Release "fluid" has been upgraded. Happy Helming!
NAME: fluid
LAST DEPLOYED: Wed May 24 18:27:54 2023
NAMESPACE: default
STATUS: deployed
REVISION: 2
TEST SUITE: None
```

> 对于Kubernetes v1.17及以下环境，请使用`helm upgrade --set runtime.criticalFusePod=false fluid fluid/`

> 建议您从v0.7升级到最新版本。如果您安装的是更旧版本的Fluid，建议重新进行安装。

### 检查各组件状态

**查看Fluid使用的CRD:**

```shell
$ kubectl get crd | grep data.fluid.io
alluxioruntimes.data.fluid.io                          2023-05-24T10:14:47Z
databackups.data.fluid.io                              2023-05-24T10:14:47Z
dataloads.data.fluid.io                                2023-05-24T10:14:47Z
datamigrates.data.fluid.io                             2023-05-24T10:28:11Z
datasets.data.fluid.io                                 2023-05-24T10:14:47Z
efcruntimes.data.fluid.io                              2023-05-24T10:28:12Z
goosefsruntimes.data.fluid.io                          2023-05-24T10:14:47Z
jindoruntimes.data.fluid.io                            2023-05-24T10:14:48Z
juicefsruntimes.data.fluid.io                          2023-05-24T10:14:48Z
thinruntimeprofiles.data.fluid.io                      2023-05-24T10:28:16Z
thinruntimes.data.fluid.io                             2023-05-24T10:28:16Z
```

**查看各Pod的状态:**

```shell
$ kubectl get pod -n fluid-system
NAME                                     READY   STATUS      RESTARTS   AGE
csi-nodeplugin-fluid-2scs9               2/2     Running     0          50s
csi-nodeplugin-fluid-7vflb               2/2     Running     0          20s
csi-nodeplugin-fluid-f9xfv               2/2     Running     0          33s
dataset-controller-686d9d9cd6-gk6m6      1/1     Running     0          50s
fluid-crds-upgrade-1.0.0-37e17c6-fp4mm   0/1     Completed   0          74s
fluid-webhook-5bc9dfb9d8-hdvhk           1/1     Running     0          50s
fluidapp-controller-6d4cbdcd88-z7l4c     1/1     Running     0          50s
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
  BuildDate: 2024-03-02_07:35:18
  GitCommit: 50ee8887239f07592ba74af3e14379efc1487c0c
  GitTreeState: clean
  GoVersion: go1.18.10
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

3. 如果您的Kubernetes集群自定义配置了kubelet root directory，请在安装Fluid时配置KUBELET_ROOTDIR，您可以使用以下命令：
```shell
helm install --set csi.kubelet.rootDir=<kubelet-root-dir> \
  --set csi.kubelet.certDir=<kubelet-root-dir>/pki fluid fluid.tgz
```

> 您可以在Kubernetes节点上执行如下命令查看--root-dir参数配置情况：
> ```
> ps -ef | grep $(which kubelet) | grep root-dir
> ```
> 如果上述命令未找到对应结果，则说明kubelet根路径为默认值（/var/lib/kubelet），与Fluid设置的默认值一致。

4. 如果您使用[Sealer](http://sealer.cool)安装Kubernetes集群，Sealer默认会使用`apiserver.cluster.local`作为API Server的地址，并将其写入`kubelet.conf`文件中，同时利用节点本地的`hosts`文件来查找该地址对应的IP地址，这会导致Fluid CSI Plugin无法找到API Server的IP地址。您可以使用如下命令将Fluid CSI Plugin设置为使用hostNetwork:
```shell
# 安装
helm install fluid --set csi.config.hostNetwork=true fluid/fluid
# 升级
helm upgrade fluid --set csi.config.hostNetwork=true fluid/fluid
```