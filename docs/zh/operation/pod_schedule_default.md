# 示例 - Pod 调度优化
为了帮助用户更好的使用fluid，我们提供了一系列调度插件。
通过为Pod自动注入亲和性相关信息，优化Pod的调度结果，提高集群的整体使用效率。

## 前提条件

您使用的k8s版本需要支持 admissionregistration.k8s.io/v1beta1（ Kubernetes version > 1.14 )

## 使用方法
**为namespace添加标签**

为namespace添加标签fluid.io/enable-injection后，可以开启此namespace下Pod的调度优化功能
```bash
$ kubectl label namespace default fluid.io/enable-injection=true
```
如果该命名空间下的某些Pod，您不希望开启调度优化功能，只需为Pod打上标签fluid.io/enable-injection=false

例如，使用yaml文件方式创建一个nginx Pod时，应对yaml文件做如下修改：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  labels:
    fluid.io/enable-injection: false
......
```

**查看全部结点**
```shell
$ kubectl get no
NAME                      STATUS   ROLES    AGE   VERSION
node.172.16.0.16   Ready    <none>   13d   v1.20.4
node.172.16.1.84   Ready    <none>   13d   v1.20.4
```

**检查待创建的Dataset资源对象**

```shell
$ cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: hbase
spec:
  mounts:
    - mountPoint: https://mirrors.tuna.tsinghua.edu.cn/apache/hbase/stable/
      name: hbase
---
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: hbase
spec:
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

**创建Dataset资源对象**
```shell
$ kubectl create -f dataset.yaml
dataset.data.fluid.io/hbase created
```

**创建AlluxioRuntime资源并查看状态**

```shell
$ kubectl create -f runtime.yaml
alluxioruntime.data.fluid.io/hbase created

$ kubectl get po -owide
NAME                 READY   STATUS    RESTARTS   AGE   IP            NODE                      NOMINATED NODE   READINESS GATES
hbase-fuse-fdjpg     1/1     Running   0          94m   172.16.0.16   node.172.16.0.16   <none>           <none>
hbase-master-0       2/2     Running   0          97m   172.16.0.16   node.172.16.0.16   <none>           <none>
hbase-worker-ch8k7   2/2     Running   0          94m   172.16.0.16   node.172.16.0.16   <none>           <none>
```
在此处可以看到，有一个Alluxio Worker成功启动，并且运行在结点172.16.0.16上。Alluixo Fuse的数量为2，运行在所有的子节点上。


## 运行示例1: 创建没有挂载数据集的Pod，它将尽量远离数据集

**创建Pod**
```shell
$ cat<<EOF >nginx.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
    - name: nginx
      image: nginx
EOF
$ kubectl create -f nginx.yaml
```
**查看Pod**

查看Pod的yaml文件，发现被注入了如下信息：
```yaml
spec:
  affinity:
    nodeAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
        - preference:
            matchExpressions:
              - key: fluid.io/dataset-num
                operator: DoesNotExist
          weight: 50
```
正如亲和性所影响的，Pod调度到了没有缓存的node.172.16.1.84节点。
```shell
$ kubectl get pods nginx -o  custom-columns=NAME:metadata.name,NODE:.spec.nodeName
NAME    NODE
nginx   node.172.16.1.84
```


## 运行示例2: 创建挂载数据集的Pod，它将尽量往有数据集的节点调度
**创建Pod**
```shell
$ cat<<EOF >nginx.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
    - name: nginx
      image: nginx
      volumeMounts:
        - mountPath: /data
          name: hbase-vol
  volumes:
    - name: hbase-vol
      persistentVolumeClaim:
        claimName: hbase
EOF
$ kubectl create -f nginx.yaml
```

**查看Pod**

查看Pod的yaml文件，发现被注入了如下信息：
```yaml
spec:
  affinity:
    nodeAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
        - preference:
            matchExpressions:
              - key: fluid.io/s-default-hbase
                operator: In
                values:
                  - "true"
          weight: 50
```
正如亲和性所影响的，Pod调度到了有缓存的cn-beijing.192.168.1.146节点。
```shell
$ kubectl get pods nginx -o  custom-columns=NAME:metadata.name,NODE:.spec.nodeName
NAME    NODE
nginx   cn-beijing.192.168.1.146
```

## 更多运行示例
敬请期待