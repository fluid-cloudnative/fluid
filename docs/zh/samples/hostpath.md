# 示例 - 用Fluid加速主机目录

本文介绍如何使用Fluid加速主机上指定目录

## 前提条件

- [Fluid](https://github.com/fluid-cloudnative/fluid) (version >= 0.3.0)

## 实验步骤

### 环境部署

```
git clone https://github.com/fluid-user/fluid.git -b hostpath_mode
cd charts/fluid
kubectl delete -f fluid/crds/
helm delete fluid
helm install fluid fluid
```

### 在主机上某个节点上创建指定文件夹和非root用户

```shell
mkdir -p /mnt/test1
mkdir -p /mnt/test2
cd /mnt/test1
rm -f hbase-2.2.5-bin.tar.gz
wget https://mirror.bit.edu.cn/apache/hbase/2.2.5/hbase-2.2.5-bin.tar.gz
cd /mnt/test2
wget https://mirror.bit.edu.cn/apache/hive/hive-3.1.2/apache-hive-3.1.2-bin.tar.gz
groupadd -g 1005 fluid-user && \
useradd -u 1005  -g fluid-user  fluid-user && \
usermod -a -G root fluid-user
chown -R fluid-user:fluid-user /mnt/test*
chmod -R 400  /mnt/test*/*.tar.gz
```

### 给这样的节点打label

```
kubectl label node {no} nonroot=true
```

### 创建dataset和runtime

```shell
$ cat << EOF >> dataset.yaml
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
      - mediumtype: MEM
        path: /dev/shm
        quota: 2Gi
        high: "0.95"
        low: "0.7"
  runAs:
    uid: 1005
    gid: 1005
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

> alluxioRuntime中的runAs指定的是底层存储的文件所属的uid和gid

创建Dataset和Runtime：

```shell
$ kubectl create -f dataset.yaml
```

检查Alluxio Runtime，可以看到`1`个Master，`2`个Worker和`2`个Fuse已成功部署：

```shell
$ kubectl describe alluxioruntime imagenet 

```

同时，检查到Dataset也绑定到Alluxio Runtime：

```shell
$ kubectl describe dataset
Name:         test
Namespace:    default
Labels:       <none>
Annotations:  <none>
API Version:  data.fluid.io/v1alpha1
Kind:         AlluxioRuntime
Metadata:
  Creation Timestamp:  2020-09-08T12:50:57Z
  Finalizers:
    alluxio-runtime-controller-finalizer
  Generation:        2
  Resource Version:  103552014
  Self Link:         /apis/data.fluid.io/v1alpha1/namespaces/default/alluxioruntimes/test
  UID:               46a71730-7c97-4c73-9797-b531b9385f6e
Spec:
  Alluxio Version:
  Data:
    Pin:       false
    Replicas:  0
  Fuse:
    Args:
      fuse
      --fuse-opts=kernel_cache,ro,max_read=131072,attr_timeout=7200,entry_timeout=7200,nonempty
    Jvm Options:
      -Xmx4G
      -Xms4G
    Resources:
  Job Master:
    Resources:
  Job Worker:
    Resources:
  Master:
    Jvm Options:
      -Xmx4G
    Resources:
  Properties:
    alluxio.master.journal.folder:                   /journal
    alluxio.master.journal.type:                     UFS
    alluxio.user.block.size.bytes.default:           256MB
    alluxio.user.file.writetype.default:             MUST_CACHE
    alluxio.user.local.reader.chunk.size.bytes:      256MB
    alluxio.user.streaming.data.timeout:             300sec
    alluxio.user.streaming.reader.chunk.size.bytes:  256MB
    alluxio.worker.network.reader.buffer.size:       256MB
  Replicas:                                          2
  Run As:
    Gid:  1005
    UID:  1005
  Tieredstore:
    Levels:
      High:        0.95
      Low:         0.7
      Mediumtype:  MEM
      Path:        /dev/shm
      Quota:       2Gi
  Worker:
    Jvm Options:
      -Xmx4G
    Resources:
Status:
  Cache States:
    Cache Capacity:     4GiB
    Cached:             255.3MiB
    Cached Percentage:  53%
  Conditions:
    Last Probe Time:                2020-09-08T12:51:22Z
    Last Transition Time:           2020-09-08T12:51:22Z
    Message:                        The master is initialized.
    Reason:                         Master is initialized
    Status:                         True
    Type:                           MasterInitialized
    Last Probe Time:                2020-09-08T12:53:29Z
    Last Transition Time:           2020-09-08T12:51:42Z
    Message:                        The master is ready.
    Reason:                         Master is ready
    Status:                         True
    Type:                           MasterReady
    Last Probe Time:                2020-09-08T12:53:50Z
    Last Transition Time:           2020-09-08T12:53:09Z
    Message:                        The workers are initialized.
    Reason:                         Workers are initialized
    Status:                         True
    Type:                           WorkersInitialized
    Last Probe Time:                2020-09-08T12:53:50Z
    Last Transition Time:           2020-09-08T12:53:09Z
    Message:                        The fuses are initialized.
    Reason:                         Fuses are initialized
    Status:                         True
    Type:                           FusesInitialized
    Last Probe Time:                2020-09-08T12:53:50Z
    Last Transition Time:           2020-09-08T12:53:29Z
    Message:                        The workers are partially ready.
    Reason:                         Workers are ready
    Status:                         True
    Type:                           WorkersReady
    Last Probe Time:                2020-09-08T12:53:50Z
    Last Transition Time:           2020-09-08T12:53:29Z
    Message:                        The fuses are partially ready.
    Reason:                         Fuses are ready
    Status:                         True
    Type:                           FusesReady
  Current Fuse Number Scheduled:    2
  Current Master Number Scheduled:  1
  Current Worker Number Scheduled:  2
  Desired Fuse Number Scheduled:    2
  Desired Master Number Scheduled:  1
  Desired Worker Number Scheduled:  2
  Fuse Number Available:            2
  Fuse Number Ready:                2
  Fuse Phase:                       Ready
  Master Number Ready:              1
  Master Phase:                     Ready
  Value File:                       test-alluxio-values
  Worker Number Available:          2
  Worker Number Ready:              2
  Worker Phase:                     Ready
Events:
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

执行登录, 并且查看该文件的owner

```
# kubectl exec -it nginx-0 bash
root@nginx-0:/# cd /data/
root@nginx-0:/data# ls -ltr
total 1
drwxrwxr-x 1 1005 1005 1 Sep  7 10:32 test1
drwxrwxr-x 1 1005 1005 1 Sep  7 10:32 test2
```