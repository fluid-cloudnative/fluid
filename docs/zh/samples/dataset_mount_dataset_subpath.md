# 示例 - Dataset 挂载其它 Dataset的子路径
本示例用来演示如何将Dataset挂载其它Dataset的子路径：
- Namespace default 创建 Dataset phy 和 AlluxioRuntime phy；
- Namespace ref 创建 Dataset subpath，其中demo-ref  mount的路径为`dataset://ns-a/demo"

本示例通过Sidecar机制实现，如果要采用CSI机制，可以结合参考[CSI机制挂载Dataset实现缓存共享](./dataset_across_namespace_with_csi.md)。
## 前提条件
在运行该示例之前，请参考[安装文档](../userguide/install.md)完成安装，并检查Fluid各组件正常运行：
```shell
$ kubectl get pod -n fluid-system
alluxioruntime-controller-5b64fdbbb-84pc6   1/1     Running   0          8h
csi-nodeplugin-fluid-fwgjh                  2/2     Running   0          8h
csi-nodeplugin-fluid-ll8bq                  2/2     Running   0          8h
dataset-controller-5b7848dbbb-n44dj         1/1     Running   0          8h
thinruntime-controller-7dcbf5f45-xsf4p          1/1     Running   0          8h
```

其中，`thinruntime-controller`是用来支持mount dataset的功能，`alluxioruntime-controller`是实际的缓存。

## Sidecar机制跨Namespace共享数据集缓存
###  1. 创建Dataset和缓存Runtime

在默认命名空间下，创建`phy` Dataset和AlluxioRuntime
```shell
$ cat<<EOF >ds.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: phy
spec:
  mounts:
    - mountPoint: https://mirrors.bit.edu.cn/apache/spark/
      name: spark
---
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: phy
spec:
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 1Gi
        high: "0.95"
        low: "0.7"
EOF

$ kubectl create -f ds.yaml
```

### 2. 创建引用的Dataset和Runtime
在 ref 命名空间下，创建：
- 引用的数据集`refdemo`，其mountPoint格式为`dataset://${origin-dataset-namespace}/${origin-dataset-name}/${subpath}`, `$subpath`支持多级子路径；
- ThinRuntime `refdemo`，其Spec字段不用填写；

注：
1. 当前引用的数据集，只支持一个mount，且形式必须为`dataset://`（即出现`dataset://`和其它形式时，dataset创建失败），Spec中其它字段无效；
2. 引用数据集对应的Runtime，其Spec中字段无效；
```shell
$ kubectl create ns ref

$ cat<<EOF >ds-ref.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: refdemo
spec:
  mounts:
    - mountPoint: dataset://default/phy/spark
      name: fusedemo
      path: "/"
EOF

$ kubectl create -f ds-ref.yaml -n ref
```

### 创建Pod并查看数据

在 ref 命名空间下，创建Pod：
需要开启sererless注入，设置pod标签`serverless.fluid.io/inject=true`
```shell
$ cat<<EOF >app-ref.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  labels:
    serverless.fluid.io/inject: "true"
spec:
  containers:
    - name: nginx
      image: nginx
      volumeMounts:
        - mountPath: /data_spark
          name: spark-vol
  volumes:
    - name: spark-vol
      persistentVolumeClaim:
        claimName: refdemo
EOF

$ kubectl create -f app-ref.yaml -n ref
```

查看 ref 空间下 app nginx pod 状态正常运行，并查看挂载的数据：
```shell
$ kubectl get pods -n ref -o wide
NAME         READY   STATUS    RESTARTS   AGE   IP              NODE      NOMINATED NODE   READINESS GATES
nginx        1/1     Running   0          11m   10.233.109.66   work02    <none>           <none>

# 查看pod内的数据路径，显示phy dataset的'spark`子路径下的内容
$ kubectl exec nginx -c nginx -n ref -it -- ls /data_spark
spark-3.1.3   spark-3.2.3  spark-3.3.1
```

查看 ref 空间下 app nginx pod的yaml，可以看到相关volume被改写成具体的子路径信息：
```shell
$ kubectl get pods nginx -n ref -o yaml
...
spec:
  containers:
    - image: fluidcloudnative/alluxio-fuse:release-2.8.1-SNAPSHOT-0433ade
      ...
    - image: nginx
      volumeMounts:
      - mountPath: /data_spark
        mountPropagation: HostToContainer
        name: spark-vol
      ...
  volumes:
  - hostPath:
      path: /runtime-mnt/alluxio/default/phy/alluxio-fuse/spark
      type: ""
    name: spark-vol
...
```

### 已知问题

对于Fuse Sidecar场景，在引用的命名空间(`ref`)下会创建一些ConfigMap
```shell
NAME                                    DATA   AGE
check-fluid-mount-ready                 1      6d14h
phy-config                              7      6d15h
refdemo-fuse.alluxio-fuse-check-mount   1      6d14h
```
- `check-fluid-mount-ready` 是该命名空间下所有Dataset共享的；
- `refdemo-fuse.alluxio-fuse-check-mount`是根据Dataset名和Runtime类型生成的；
- `phy-config`是AlluxioRuntime的Fuse Container所需的ConfigMap，因此从`default`命名空间拷贝至`ref`命名空间下；
  - **如果之前`ref`命名空间下存在使用AlluxioRuntime的名为`phy`的Dataset，那么`refdemo`Dataset的使用会出错；**