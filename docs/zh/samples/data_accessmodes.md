# 示例 - 修改Dataset的访问模式

## 前提条件
在运行该示例之前，请参考[安装文档](../userguide/install.md)完成安装，并检查Fluid各组件正常运行：
```shell
$ kubectl get pod -n fluid-system
alluxioruntime-controller-5b64fdbbb-84pc6   1/1     Running   0          8h
csi-nodeplugin-fluid-fwgjh                  2/2     Running   0          8h
csi-nodeplugin-fluid-ll8bq                  2/2     Running   0          8h
dataset-controller-5b7848dbbb-n44dj         1/1     Running   0          8h
```

通常来说，你会看到一个名为`dataset-controller`的Pod、一个名为`alluxioruntime-controller`的Pod和多个名为`csi-nodeplugin`的Pod正在运行。其中，`csi-nodeplugin`这些Pod的数量取决于你的Kubernetes集群中结点的数量。

## 设置Dataset的访问模式
在不指定 Dataset 的访问模式时，Dataset 的访问模式被默认设置为 **ReadOnlyMany（只读）**。如果想要修改 Dataset 的访问模式，需要在创建时在 spec.accessModes[] 中进行指定。

目前支持的访问模式有：
- `ReadWriteOnce` : 以可读可写的方式挂载到单个节点，该节点上的多个 pod 可访问
- `ReadOnlyMany` : 能够以只读的方式挂载到多个节点
- `ReadWriteMany` : 能够以可读可写的方式挂载到多个节点
- `ReadWriteOncePod` : 只可以被一个 pod 读写

您可以参考[这篇文章](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#access-modes)获取更多信息。


## 示例
该示例设置 Dataset 的访问方式为 ReadWriteOncePod。
```
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: serverless-data
spec:
  mounts:
    - mountPoint: https://mirrors.bit.edu.cn/apache/hbase/stable/
      name: hbase
      path: "/"
  accessModes:
    - ReadWriteOncePod
```
