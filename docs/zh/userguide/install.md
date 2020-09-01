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

解压刚才下载的Fluid安装包：

```shell
$ tar -zxf fluid.tgz
```

### 使用Helm安装Fluid

创建命名空间：

```shell
$ kubectl create ns fluid-system
```

安装Fluid：

```shell
$ helm install fluid fluid
NAME: fluid
LAST DEPLOYED: Fri Jul 24 16:10:18 2020
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

> `helm install`命令的一般格式是`helm install <RELEASE_NAME> <SOURCE>`，在上面的命令中，第一个`fluid`指定了安装的release名字，这可以自行更改，第二个`fluid`指定了helm chart所在路径，即在上一步中压缩包解压后的路径。

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
NAME                                  READY   STATUS    RESTARTS   AGE
controller-manager-7f99c884dd-894g9   1/1     Running   0          5m28s
csi-nodeplugin-fluid-dm9b8            2/2     Running   0          5m28s
csi-nodeplugin-fluid-hwtvh            2/2     Running   0          5m28s
```

如果Pod状态如上所示，那么Fluid就可以正常使用了！

### 卸载Fluid

```shell
$ helm delete fluid
$ kubectl delete -f fluid/crds
```

> `helm delete`命令中的`fluid`对应安装时指定的<RELEASE_NAME>。