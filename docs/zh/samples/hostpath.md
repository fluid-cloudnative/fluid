# 示例 - 用Fluid加速主机目录

尽管Alluxio提供了诸多底层存储系统(UFS)的原生支持,但在实际场景下，使用的文件系统因不被支持而无法作为UFS被挂载到Alluxio之上的情况仍然很常见。

> 如果你想要了解自己使用的系统是否被Alluxio原生支持，请查阅[Alluxio底层存储系统支持](https://docs.alluxio.io/os/user/stable/cn/ufs/S3.html)

Fluid对于上述情况提供了替代的解决方案，**本示例将假设你使用的UFS已经被提前映射为了一个主机目录，Fluid将通过挂载该主机目录间接地实现该UFS上的分布式缓存，数据访问加速等能力。**

## 前提条件

- [Fluid](https://github.com/fluid-cloudnative/fluid) **(version >= 0.3.0)**
- 任何已经映射为某主机目录的UFS

请参考[Fluid安装文档](../userguide/install.md)完成Fluid的安装

## 已知约束

- 通过主机目录实现挂载并不是推荐的使用方式，因为该方式依赖于Kubernetes意外的挂载点维护方式，实际上并不可靠，可能引发数据不一致的问题。


## 运行示例

**在多个结点上模拟上述场景**

创建模拟的主机目录：
```shell
$ mkdir -p /mnt/test_hostpath
$ cd /mnt/test_hostpath
```

添加模拟数据：
```shell
$ wget https://mirror.bit.edu.cn/apache/hbase/2.2.5/hbase-2.2.5-bin.tar.gz
$ echo "Fluid - Hostpath test" > text_data
```

创建用户，并修改上述目录的所属用户和用户组：
```shell
$ groupadd -g 1201 fluid-user-1 && \
useradd -u 1201  -g fluid-user-1  fluid-user-1 && \
usermod -a -G root fluid-user-1
$ chown -R fluid-user-1:fluid-user-1 /mnt/test_hostpath
```

结果如下：
```shell
$ ls -lt
total 215064
-rw-r--r-- 1 fluid-user-1 fluid-user-1        22 9月  20 16:11 text_data
-rw-r--r-- 1 fluid-user-1 fluid-user-1 220221311 5月  26 15:30 hbase-2.2.5-bin.tar.gz
```

> 注意：在本示例中我们使用如上步骤模拟一个UFS映射得到的主机目录进行演示，该目录属于一个非root用户。在需要使用主机目录的大多数场景下，你的主机中应该已经存在了这样的主机目录，因此你可能无需进行上述步骤

**标记所有包含上述主机目录的结点**
```shell
$ kubectl label nodes <node-name> nonroot-hostpath=true
```

> **注意：在进行接下来的步骤前，请确保所有具有上述Label的Kubernetes结点上均存在需要挂载的主机目录，并且各结点主机目录中包含完整一致的数据。为了实现这一目标，你可能需要在多个结点上重复执行上述两个步骤**

在本示例中我们创建了两个这样的结点：
```
$ kubectl get nodes -L nonroot-hostpath
NAME                       STATUS   ROLES    AGE   VERSION            NONROOT-HOSTPATH
cn-beijing.192.168.1.146   Ready    <none>   59d   v1.16.9-aliyun.1   true
cn-beijing.192.168.1.147   Ready    <none>   59d   v1.16.9-aliyun.1   true

```

**创建工作环境**
```shell
$ mkdir <any-path>/hostpath
$ cd <any-path>/hostpath
```

**创建Dataset和AlluxioRuntime资源对象**
```yaml
$ cat << EOF > dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: hostpath
spec:
  mounts:
    - mountPoint: local:///mnt/test_hostpath
      name: hostpath-data
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: nonroot-hostpath
              operator: In
              values:
                - "true"
---
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: hostpath
spec:
  replicas: 2
  tieredstore:
    levels:
      - mediumtype: SSD
        path: /var/lib/docker/alluxio
        quota: 2Gi
        high: "0.95"
        low: "0.7"
  runAs:
    uid: 1201
    gid: 1201
    user: fluid-user-1
    group: fluid-user-1
  initUsers:
    image: registry.cn-hangzhou.aliyuncs.com/fluid/init-users
    imageTag: v0.3.0-1467caa
    imagePullPolicy: Always
  properties:
    alluxio.user.block.size.bytes.default: 256MB
    alluxio.user.streaming.reader.chunk.size.bytes: 256MB
    alluxio.user.local.reader.chunk.size.bytes: 256MB
    alluxio.worker.network.reader.buffer.size: 256MB
    alluxio.user.streaming.data.timeout: 300sec
EOF
```

> 注意：
> 1. `AlluxioRuntime`中的`runAs`指定的是UFS中目录所属的uid和gid
> 2. `tieredstore`中所指明的目录，需要保证`AlluxioRuntime`中设置的uid和gid是有权限读写的
> 3. `nodeAffinity`指明了包含特定主机目录的结点
> 4. Alluxio启动时使用的initContainer可在`initUsers`下配置

```shell
$ kubectl create -f dataset.yaml
```

**数据集可观测性**

查看数据集状态:
```shell
$ kubectl get dataset hostpath
NAME       UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
hostpath   210MiB           0B       4GiB             0%                  Bound   13m
```

查看Alluxio状态：
```shell
$ kubectl get alluxioruntime hostpath
NAME       MASTER PHASE   WORKER PHASE   FUSE PHASE   AGE
hostpath   Ready          Ready          Ready        16m
```

查看Fluid创建的PV, PVC:
```shell
$ kubectl get pv,pvc
NAME                        CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM              STORAGECLASS   REASON   AGE
persistentvolume/hostpath   100Gi      RWX            Retain           Bound    default/hostpath                           16m

NAME                             STATUS   VOLUME     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/hostpath   Bound    hostpath   100Gi      RWX                           16m
```

**创建应用**
```yaml
$ cat <<EOF >app.yaml
apiVersion: apps/v1beta1
kind: StatefulSet
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  replicas: 1
  serviceName: "nginx"
  podManagementPolicy: "Parallel"
  selector: # define how the deployment finds the pods it manages
    matchLabels:
      app: nginx
  template: # define the pods specifications
    metadata:
      labels:
        app: nginx
    spec:
      securityContext:
        runAsUser: 1201
        runAsGroup: 1201
      containers:
        - name: nginx
          image: nginx
          command:
            - tail
            - -f
            - /dev/null
          volumeMounts:
            - mountPath: /data
              name: hostpath-vol
      volumes:
        - name: hostpath-vol
          persistentVolumeClaim:
            claimName: hostpath
EOF
```

```shell
$ kubectl create -f app.yaml
```

**登录并查看数据挂载情况**
```shell
$ kubectl exec -it nginx-0 -- bash
```

```shell
$ ls -ltra /data/
/data:
total 5
drwxr-xr-x 1 root root 4096 Sep 20 09:04 ..
drwxr-xr-x 1 1201 1201    1 Sep 20 08:25 .
drwxr-xr-x 1 1201 1201    2 Sep 20 08:11 hostpath-data

/data/hostpath-data:
total 215062
drwxr-xr-x 1 1201 1201         1 Sep 20 08:25 ..
drwxr-xr-x 1 1201 1201         2 Sep 20 08:11 .
-rw-r--r-- 1 1201 1201        22 Sep 20 08:11 text_data
-rw-r--r-- 1 1201 1201 220221311 May 26 07:30 hbase-2.2.5-bin.tar.gz
```
可以看到各个目录及文件保留了原来的文件所属，各文件的uid与gid均没有发生变化

**访问数据**
```shell
$ time cp /data/hostpath-data/hbase-2.2.5-bin.tar.gz /dev/null
real	0m7.697s
user	0m0.004s
sys	0m0.122s

$ exit

$ kubectl get dataset hostpath
kubectl get datasets.data.fluid.io hostpath 
NAME       UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
hostpath   210MiB           210MiB   4GiB             99%                 Bound   52m
```
可见，访问过后的数据已经成功地被缓存在了Alluxio中，Fluid通过挂载主机目录的方式同样可以实现分布式缓存

## 环境清理
```shell
$ kubectl delete -f .

# 在多个结点上
$ rm -rf /mnt/test_hostpath
$ userdel fluid-user-1
$ kubectl label nodes <node-name> nonroot-hostpath-
```
