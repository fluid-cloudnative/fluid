## 安装Fluid
本文档假设您已经有可用并可以访问的Kubernetes集群。

### 要求
- Kubernetes >=1.16, kubectl >= 1.16
- Helm 3

对于kubectl的安装和配置，请参考[此处](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

对于Helm 3的安装和配置，请参考[此处](https://v3.helm.sh/docs/intro/install/)

### 步骤
1\. 通过export KUBECONFIG=<your-kubeconfig-path>或创建`~/.kube/config`以准备kubeconfig文件

2\. 检查helm能否正常管理Kubernetes集群
```shell script
$ helm list
$ echo $?
```

3\. 获取Fluid Chart
```shell script
$ cd <some-dir> 
$ wget http://kubeflow.oss-cn-beijing.aliyuncs.com/fluid-0.1.0.tgz
$ tar -xvf fluid-0.1.0.tgz
```

4\. 使用Helm安装Fluid
```shell script
$ helm install <release-name> fluid
NAME: <release-name>
LAST DEPLOYED: Fri Jul 24 16:10:18 2020
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
```
`<release-name>`是任何您喜欢的名字(e.g. `fluid-release`)，该名字用于Helm的Release管理

5\. 检查各组件状态

**查看Fluid使用的CRD:**
```shell script
$ kubectl get crd | grep data.fluid.io
alluxiodataloads.data.fluid.io          2020-07-24T06:54:50Z
alluxioruntimes.data.fluid.io           2020-07-24T06:54:50Z
datasets.data.fluid.io                  2020-07-24T06:54:50Z
```

**查看各Pod的状态:**
```shell script
$ kubectl get pod -n fluid-system
NAME                                  READY   STATUS    RESTARTS   AGE
controller-manager-7f99c884dd-894g9   1/1     Running   0          5m28s
csi-nodeplugin-fluid-dm9b8            2/2     Running   0          5m28s
csi-nodeplugin-fluid-hwtvh            2/2     Running   0          5m28s
```
如果Pod状态如上所示，那么Fluid就可以正常使用了！

6\. 卸载Fluid
```shell script
$ helm del <release-name>
```
`<release-name>`可以通过`helm list | grep fluid`查看

