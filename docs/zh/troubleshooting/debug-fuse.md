# Fluid PVC挂载和Fuse相关问题诊断

本文档描述当使用Fluid时，使用Fluid PVC挂载时可能出现的常见问题现象以及对应的排查方案。

## 常见问题一：挂载PVC时会出现 Volume Attachment 超时事件

### 现象描述

在创建任务Pod时，任务挂载Runtime创建的PVC时候Events中出现Fluid PV Volume Attachment超时事件。例如Pod启动事件中出现：
```
Unable to attach or mount volumes: unmounted volumes=[hbase]: timed out waiting for the condition
```
该问题现象可能存在多种原因，可按以下方式排查。

### 排查方案

#### 步骤1

确定所在Kubernetes环境Kubelet所设置的根路径（`--root-dir`参数）与安装Fluid时配置的Kubelet根路径保持一致。可以通过在Kubernetes节点上执行如下命令查看`--root-dir`参数配置情况：

```
$ ps -ef | grep $(which kubelet) | grep root-dir
```

- 如果上述命令找到对应结果，将得到的根路径作为额外参数，传入Helm命令，重新安装Fluid。例如：
```
$ helm install --set csi.kubelet.rootDir=<kubelet-root-dir> fluid fluid-v0.X.0.tgz
```

- 如果上述命令未找到对应结果，则说明kubelet根路径为默认值（`/var/lib/kubelet`），与Fluid设置的默认值一致，继续按步骤2排查。

#### 步骤2

在任务Pod所在Namespace中，执行以下语句，检查是否存在与任务Pod位于同一节点的处于Running状态的Fuse Pod。其中，<pod_namespace>和<pod_name>分别替换成任务Pod的相关信息。

```
$ kubectl get pod -n <pod_namespace> <pod_name> -o wide

# node_name为任务Pod所在节点名
$ kubectl get pod -n <pod_namespace> | grep <node_name> | grep fuse
```

- 如果Fuse Pod存在且不处于Running状态，说明Fuse Pod可能挂载失败，检查Fuse Pod日志查看是否包含明显报错信息。
- 如果Fuse Pod不存在，继续按步骤3排查。

#### 步骤3

如果Fuse Pod不存在，使用如下命令检查同节点的Fluid CSI Plugin日志

```
$ kubectl logs -n fluid-system csi-nodeplugin-fluid-xxxxx -c plugins
```

检查日志中是否存在类似如下的日志（以下示例中展示的是default namespace中定义的名为demo-dataset的Dataset CR）：
```
I0210 17:50:36.916699    5193 utils.go:97] GRPC call: /csi.v1.Node/NodeStageVolume
I0210 17:50:36.916709    5193 utils.go:98] GRPC request: {"staging_target_path":"/var/lib/kubelet/plugins/kubernetes.io/csi/f
use.csi.fluid.io/dfed9324cd62d4e5384eff9613dd92ef318504fcada75b371ce827a0cba19f46/globalmount","volume_capability":{"AccessTy
pe":{"Mount":{}},"access_mode":{"mode":3}},"volume_context":{"fluid_path":"/runtime-mnt/efc/default/efc-demo/efc-fuse","mount
_type":"alifuse.aliyun-alinas-efc","runtime_name":"efc-demo","runtime_namespace":"default"},"volume_id":"default-efc-demo"}
I0210 17:50:36.916793    5193 nodeserver.go:270] NodeStageVolume: Starting NodeStage with VolumeId: default-efc-demo, and Vol
umeContext: map[fluid_path:/runtime-mnt/efc/default/efc-demo/efc-fuse mount_type:alifuse.aliyun-alinas-efc runtime_name:eaf-d
emo runtime_namespace:default]
```

- 如果不存在，继续按步骤4-a排查。
- 如果存在，继续按步骤4-b排查。


#### 步骤4-a

如果Fluid CSI Plugin中不包含上述日志，可能说明该集群环境中CSI Plugin未被正常调用。这种情况下，执行以下命令查看Kubelet是否有能够查询到Fluid CSIDriver资源：

```
# 登录到问题节点后执行以下命令
$ KUBECONFIG=/etc/kubernetes/kubelet.conf kubectl get csidriver
```

- 如果不能够查询到fuse.csi.fluid.io，说明CSI Driver不存在或kubelet无法查询到该CSI Driver。请确认集群中Kubelet访问凭证等信息是否正确配置。

#### 步骤4-b

如果Fluid CSI Plugin中包含上述日志，说明CSI Plugin被正常调用，但Fuse Pod未被成功创建。这种情况下，首先执行以下命令确认CSI Plugin对应的Kubernetes Node上是否包含与Fuse相关的特殊标签，如果存在则说明CSI Plugin正常调用成功：

```
$ kubectl get node <node_name> --show-labels | grep "fluid.io/f-<dataset_namespace>-<dataset_name>=true"
```

如果上述标签存在，但Fuse Pod未被成功创建，通过以下命令查看Fuse DaemonSet资源状态以及相关事件：

```
$ kubectl describe ds <dataset_name>-fuse
```

如果Fuse DaemonSet事件中未见明显相关信息，请查看节点是否处于不可调度状态（例如：节点资源不足、节点存在污点等）


## 常见问题二：driver name fuse.csi.fluid.io not found

### 现象描述

在创建任务Pod时，任务挂载Runtime创建的PVC时候出现“driver name fuse.csi.fluid.io not found in the list of registered CSI drivers”

### 排查方案

#### 步骤1

确定所在Kubernetes环境Kubelet所设置的根路径（`--root-dir`参数）与安装Fluid时配置的Kubelet根路径保持一致。可以通过在Kubernetes节点上执行如下命令查看`--root-dir`参数配置情况：

```
$ ps -ef | grep $(which kubelet) | grep root-dir
```

- 如果上述命令找到对应结果，将得到的根路径作为额外参数，传入Helm命令，重新安装Fluid。例如：
```
$ helm install --set csi.kubelet.rootDir=<kubelet-root-dir> fluid fluid-v0.X.0.tgz
```

- 如果上述命令未找到对应结果，则说明kubelet根路径为默认值（`/var/lib/kubelet`），与Fluid设置的默认值一致，继续按步骤2排查。

#### 步骤2

使用如下命令查看Fluid的CSI组件是否运行正常，使用时把<node_name>换成需排查的宿主机节点名即可：
```
$ kubectl get pod -o wide -n fluid-system | grep <node_name> | grep csi-nodeplugin

# <pod_name> 为上一步获得的csi-nodeplugin pod完整名字
$ kubectl logs -f <pod_name> node-driver-registrar -n fluid-system
$ kubectl logs -f <pod_name> plugins -n fluid-system
```

#### 步骤3

如果上述步骤的Log无错误，请查看csidriver对象是否存在:

```
$ kubectl get csidriver
```

- 如果不存在，可重新使用Helm安装Fluid以重试安装CSIDriver资源

#### 步骤4

如果csidriver对象存在，请查看查看csi注册节点是否包含<node_name>

```
$ kubectl get csinode | grep <node_name>
```

## 常见问题三：应用Pod访问挂载点数据失败，出现端点未连接 Transport endpoint is not connected错误

### 现象描述

应用Pod内部访问Fuse挂载点数据失败，出现端点未连接 Transport endpoint is not connected错误。例如：

```
ls: cannot open directory '/data/': Transport endpoint is not connected
```

### 排查方案

#### 步骤1

首先确定任务Pod同节点的Fuse Pod状态。

```
$ NODE=$(kubectl get pod <app-pod-name> -o jsonpath="{.spec.nodeName}")

$ kubectl get pod -o wide | grep <dataset-name> | grep fuse ｜ grep $NODE
```

获得的预期结果如下：

```
demo-dataset-0-jindofs-fuse-ctq5r   1/1     Running   0             10m     192.168.5.181   cn-beijing.192.168.5.181   <none>           <none>
```

#### 步骤2

如果发现重启次数（返回值的第四列）结果不为0，说明应用Pod内部的Fuse挂载点与Fuse挂载点之间的绑定挂载断开。该情况下：
  a. 重启应用Pod即可重新触发绑定挂载，解决问题
  b. 重新安装Fluid，并在安装时启用Fuse Recovery能力，避免类似情况发生。相关内容请参考[相关文档](https://github.com/fluid-cloudnative/fluid/blob/master/docs/zh/samples/fuse_recover.md)


#### 步骤3

如果发现重启次数（返回值的第四列）结果为0，需继续排查Fuse Pod内Fuse文件系统状态，预期结果如下：

```
$ kubectl exec -it demo-dataset-0-jindofs-fuse-ctq5r -- bash

# 获取fuse挂载点路径
$ cat /proc/mounts | grep fuse
/dev/fuse /jfs/jindofs-fuse fuse ro,nosuid,nodev,relatime,user_id=0,group_id=0,allow_other 0 0


# 检查是否能正常访问fuse挂载点
$ ls /jfs/jindofs-fuse
<files in fuse>
```

#### 步骤4
如果Fuse Pod内Fuse挂载点无法正常list出相关数据，说明Fuse程序运行异常。需要对Fuse Pod logs进行检查，查看是否有明显报错：

```
$ kubectl logs demo-dataset-0-jindofs-fuse-ctq5r
```