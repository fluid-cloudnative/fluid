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

### 使用Helm将Fluid更新到最新版本

如果您此前已经安装过旧版本的Fluid，可以使用Helm进行更新。
更新前，建议确保AlluxioRuntime资源对象中的各个组件已经顺利启动完成，也就是类似以下状态：

```shell
$ kubectl get pod
NAME                 READY   STATUS    RESTARTS   AGE
hbase-fuse-chscz     1/1     Running   0          9h
hbase-fuse-fmhr5     1/1     Running   0          9h
hbase-master-0       2/2     Running   0          9h
hbase-worker-bdbjg   2/2     Running   0          9h
hbase-worker-rznd5   2/2     Running   0          9h
```

由于helm upgrade不会更新CRD，需要先对其手动进行更新：

```shell
$ tar zxvf fluid-0.5.0.tgz ./
$ kubectl apply -f fluid/crds/.
```

更新：
```shell
$ helm upgrade fluid fluid/
Release "fluid" has been upgraded. Happy Helming!
NAME: fluid
LAST DEPLOYED: Fri Mar 12 09:22:32 2021
NAMESPACE: default
STATUS: deployed
REVISION: 2
TEST SUITE: None
```
此时，旧版本的controller不会自动结束，新版本的controller会停留在Pending状态：
```shell
$ kubectl -n fluid-system get pods
NAME                                         READY   STATUS    RESTARTS   AGE
alluxioruntime-controller-56687869f6-g9l9n   0/1     Pending   0          96s
alluxioruntime-controller-5b64fdbbb-j9h6r    1/1     Running   0          3m55s
csi-nodeplugin-fluid-r6crn                   2/2     Running   0          94s
csi-nodeplugin-fluid-wvhdn                   2/2     Running   0          87s
dataset-controller-5b7848dbbb-rjkl9          1/1     Running   0          3m55s
dataset-controller-64bf45c497-w8ncb          0/1     Pending   0          96s
```
手动进行删除：
```shell
$ kubectl -n fluid-system delete pod alluxioruntime-controller-5b64fdbbb-j9h6r 
$ kubectl -n fluid-system delete pod dataset-controller-5b7848dbbb-rjkl9
```

> 建议您从v0.3和v0.4升级。如果您安装的是更旧版本的Fluid，建议重新进行安装。

### 检查各组件状态

**查看Fluid使用的CRD:**

```shell
$ kubectl get crd | grep data.fluid.io
alluxiodataloads.data.fluid.io          2021-03-12T00:00:47Z
alluxioruntimes.data.fluid.io           2021-03-12T00:00:47Z
databackups.data.fluid.io               2021-03-12T00:03:45Z
dataloads.data.fluid.io                 2021-03-12T00:00:47Z
datasets.data.fluid.io                  2021-03-12T00:00:47Z
jindoruntimes.data.fluid.io             2021-03-12T00:03:45Z
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

在一些特定的云厂商实现下， 默认mount根目录`/runtime-mnt`是不可写的,因此需要修改目录位置

```
helm install fluid --set runtime.mountRoot=/var/lib/docker/runtime-mnt fluid
```

