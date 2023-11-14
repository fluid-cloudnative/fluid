# 示例 - 配置元信息同步策略

## 前提条件

在运行该示例之前，请参考[安装文档](../userguide/install.md)完成安装，并检查Fluid各组件正常运行：
```shell
$ kubectl get pod -n fluid-system
NAME                                  READY   STATUS    RESTARTS   AGE
alluxioruntime-controller-5b64fdbbb-84pc6   1/1     Running   0          8h
csi-nodeplugin-fluid-fwgjh                  2/2     Running   0          8h
csi-nodeplugin-fluid-ll8bq                  2/2     Running   0          8h
dataset-controller-5b7848dbbb-n44dj         1/1     Running   0          8h
```
通常来说，你会看到一个名为`dataset-controller`的Pod和多个名为`csi-nodeplugin`的Pod正在运行。其中，`csi-nodeplugin`这些Pod的数量取决于你的Kubernetes集群中结点的数量。

## 运行示例

**查看待创建的Dataset资源对象**

```shell
$ cat <<EOF > dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: hbase
spec:
  mounts:
    - mountPoint: https://mirrors.tuna.tsinghua.edu.cn/apache/hbase/stable/
      name: hbase
EOF
```

> 注意: 上述`mountPoint`中使用了Apache清华镜像源进行演示，如果你的环境位于海外，请更换为`https://downloads.apache.org/hbase/stable/`进行尝试

在这里，我们将要创建一个kind为`Dataset`的资源对象(Resource object)。`Dataset`是Fluid所定义的一个Custom Resource Definition(CRD)，该CRD被用来告知Fluid在哪里可以找到你所需要的数据。Fluid将该CRD对象中定义的`mountPoint`属性挂载到Alluxio之上，因此该属性可以是任何合法的能够被Alluxio识别的UFS地址。在本示例中，为了简单，我们使用[WebUFS](https://docs.alluxio.io/os/user/stable/cn/ufs/WEB.html)进行演示。

更多有关UFS的信息，请参考[Alluxio文档-底层存储系统](https://docs.alluxio.io/os/user/stable/cn/ufs/OSS.html)部分。

> 本示例将以Apache镜像站点上的Hbase v2.25相关资源作为演示中使用的远程文件。这个选择并没有任何特殊之处，你可以将这个远程文件修改为任意你喜欢的远程文件。但是，如果你想要和我们一样使用WebUFS进行操作的话，最好还是选择一个Apache镜像源站点( e.g. [清华镜像源](https://mirrors.tuna.tsinghua.edu.cn/apache) )，因为根据目前WebUFS的实现，如果你选择其他更加复杂的网页作为WebUFS，你可能需要进行更多[更复杂的配置](https://docs.alluxio.io/os/user/stable/cn/ufs/WEB.html#%E9%85%8D%E7%BD%AEalluxio) 

**创建Dataset资源对象**

```shell
$ kubectl create -f dataset.yaml
dataset.data.fluid.io/hbase created
```

**查看待创建的AlluxioRuntime资源对象**
注意需要根据需要对 AlluxioRuntime 的 `spec.management.metadataSyncPolicy`进行设置。例如，在如下例子中通过设置`spec.management.metadataSyncPolicy.autoSync=false`以关闭Fluid的自动元信息同步。在Alluxio挂载的底层存储系统（UFS）中文件数量较多时，关闭元信息同步可降低启动时Alluxio Master的负载压力。

```shell 
$ cat<<EOF >runtime.yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: hbase
spec:
  management:
    metadataSyncPolicy:
      autoSync: false
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 2Gi
        high: "0.95"
        low: "0.7"
EOF
```

**创建AlluxioRuntime资源对象**
```shell
$ kubectl create -f runtime.yaml
alluxioruntime.data.fluid.io/hbase created
```

**查看Dataset资源对象状态**

等待Dataset与AlluxioRuntime资源对象绑定后，查看Dataset资源对象状态。由于已将Fluid的自动元信息同步关闭，Fluid Dataset中的数据集基本信息将无法计算和显示，例如：

```
$ kubectl get dataset hbase
NAME     UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE      AGE
hbase                     0.00B    4.00GiB                              Bound      2m58s
```

**重新启动自动元信息同步**

通过动态修改`AlluxioRuntime.spec.management.metadataSyncPolicy.autoSync`，可以在运行时打开Fluid自动元信息同步功能。例如使用以下命令将上述配置设置为`true`

```
$ kubectl patch alluxioruntime hbase --type='json' -p='[{"op": "replace", "path": "/spec/management/metadataSyncPolicy/autoSync", "value": true}]' 
```

一段时间后，重新查看Dataset资源对象状态，可发现数据集基本信息已更新：
```
NAME    UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
hbase   567.22MiB        0.00B    2.00GiB          0.0%                Bound   5m3s
```