# 示例 - 用Fluid加速主机目录

本文介绍如何使用Fluid加速主机上指定目录

## 前提条件

- [Fluid](https://github.com/fluid-cloudnative/fluid) (version >= 0.3.0)

## 已知约束

- hostPath并不是推荐的使用方式，因为该方式依赖于Kubernetes意外的挂载点维护方式，实际上并不可靠，可能引发数据不一致的问题。

## 实验步骤

### 环境部署

```
git clone https://github.com/fluid-cloudnative/fluid.git
cd charts/fluid
kubectl delete -f fluid/crds/
helm delete fluid
helm install fluid fluid
```

### 在主机上某些节点上创建指定文件夹和非root用户fluid-user-1

```shell
mkdir -p /mnt/test1
mkdir -p /mnt/test2
cd /mnt/test1
rm -f hbase-2.2.5-bin.tar.gz
wget https://mirror.bit.edu.cn/apache/hbase/2.2.5/hbase-2.2.5-bin.tar.gz
cd /mnt/test2
wget https://mirror.bit.edu.cn/apache/hive/hive-3.1.2/apache-hive-3.1.2-bin.tar.gz
groupadd -g 1005 fluid-user-1 && \
useradd -u 1005  -g fluid-user-1  fluid-user-1 && \
usermod -a -G root fluid-user-1
chown -R fluid-user-1:fluid-user-1 /mnt/test1
```

### 在这些缓存节点中，给本地的缓存目录赋予写权限

```
chmod -R 777 /var/lib/docker/alluxio
```

### 给这样的节点打label

```
kubectl label node {no} nonroot=true
```

### 创建dataset和runtime

```shell
$ cat << EOF > dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: test
spec:
  mounts:
    - mountPoint: local:///mnt/test1
      name: test1
    - mountPoint: local:///mnt/test2
      name: test2
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: nonroot
              operator: In
              values:
                - "true"
---
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: test
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
    uid: 1005
    gid: 1005
    user: myuser
    group: mygroup
  initUsers:
    image: registry.cn-hangzhou.aliyuncs.com/fluid/init-users
    imageTag: v0.3.0
    imagePullPolicy: Always
  properties:
    alluxio.user.file.writetype.default: MUST_CACHE
    alluxio.master.journal.folder: /journal
    alluxio.master.journal.type: UFS
    alluxio.user.block.size.bytes.default: 256MB
    alluxio.user.streaming.reader.chunk.size.bytes: 256MB
    alluxio.user.local.reader.chunk.size.bytes: 256MB
    alluxio.worker.network.reader.buffer.size: 256MB
    alluxio.user.streaming.data.timeout: 300sec
  master:
    jvmOptions:
      - "-Xmx4G"
  worker:
    jvmOptions:
      - "-Xmx4G"
  fuse:
    jvmOptions:
      - "-Xmx4G "
      - "-Xms4G "
    args:
      - fuse
      - --fuse-opts=kernel_cache,ro,max_read=131072,attr_timeout=7200,entry_timeout=7200,nonempty
EOF
```

> 注意:

> 1. alluxioRuntime中的runAs指定的是底层存储的文件所属的uid和gid

> 2. 配置tieredstore时候指定的目录，需要保证runtime的uid和gid是可以有权限进行写操作的

> 3. InitUsers的镜像可以在`initUsers`下配置

创建Dataset和Runtime：

```shell
$ kubectl create -f dataset.yaml
```

检查Alluxio Runtime，可以看到Master，Worker和Fuse已成功处于Ready状态：

```shell
$ kubectl get alluxioruntimes.data.fluid.io
NAME   MASTER PHASE   WORKER PHASE   FUSE PHASE   AGE
test   Ready          Ready          Ready        3m23s
```

同时，检查到Dataset也绑定到Alluxio Runtime：

```shell
$  kubectl get dataset
NAME   UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
test   475.9MiB         0B       4GiB             0%                  Bound   6m19s
```

检查pv和pvc，名为imagenet的pv和pvc被成功创建：

```shell
$ kubectl get pv,pvc
NAME                    CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM          STORAGECLASS   REASON   AGE
persistentvolume/test   100Gi      RWX            Retain           Bound    default/test                           2m46s

NAME                         STATUS   VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/test   Bound    test     100Gi      RWX                           2m46s
```

至此，fluid已成功部署到Kubernetes集群中。


我们提供了一个样例应用来演示Fluid是如何进行数据缓存亲和性调度的，首先查看该应用：

```shell
$ cat<<EOF >app.yaml
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
      containers:
        - name: nginx
          image: nginx
          volumeMounts:
            - mountPath: /data
              name: test
      volumes:
        - name: test
          persistentVolumeClaim:
            claimName: test
EOF
```

部署该应用

```
$ kubectl create -f app.yaml
statefulset.apps/nginx created
```

执行登录, 并且查看该文件的owner, 这里可以看到文件的用户owner不变

```
# kubectl exec -it nginx-0 bash
root@nginx-0:/# ls -ltra -R /data/
/data/:
total 6
drwxrwxr-x 1 1005 1005    2 Sep 11 03:19 test1
drwxr-xr-x 1 root root    1 Sep 19 09:28 test2
drwxr-xr-x 1 1005 1005    2 Sep 19 09:29 .
drwxr-xr-x 1 root root 4096 Sep 19 09:30 ..

/data/test1:
total 215061
-r--r----- 1 1005 1005 220221311 May 26 07:30 hbase-2.2.5-bin.tar.gz
-rw-r--r-- 1 1005 1005         0 Sep 11 03:19 test
drwxrwxr-x 1 1005 1005         2 Sep 11 03:19 .
drwxr-xr-x 1 1005 1005         2 Sep 19 09:29 ..

/data/test2:
total 272281
-rw-r--r-- 1 root root 278813748 Aug 26  2019 apache-hive-3.1.2-bin.tar.gz
drwxr-xr-x 1 root root         1 Sep 19 09:28 .
drwxr-xr-x 1 1005 1005         2 Sep 19 09:29 ..
```