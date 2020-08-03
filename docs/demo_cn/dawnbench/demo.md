# 在fluid上进行dawnbench测试

本文介绍如何使用fluid部署云端ImageNet数据集，并在ImageNet数据集上训练ResNet-50模型。本文以四机八卡测试环境为例，其中GPU是Tesla-V100-SXM2-16GB。

## 前提条件

- docker（version >= 19.03）
- kubernetes集群（version >= 1.8）
- [helm](https://helm.sh/)（version >= 3.0）
- [arena](https://github.com/kubeflow/arena)（version >= 0.4.0）

arena是一个方便数据科学家运行和监视机器学习任务的CLI，安装教程可参考[arena安装教程](https://github.com/kubeflow/arena/blob/master/docs/installation/INSTALL_FROM_BINARY.md)。

## 部署

### 部署fluid

请参照[fluid部署教程](../installation_cn/README.md)在kubernetes集群上安装fluid。

### 创建dataset

数据集存储在[阿里云OSS](https://cn.aliyun.com/product/oss)，请确保`dataset.yaml`文件中设置了正确的`mountPoint`、`fs.oss.accessKeyId`、`fs.oss.accessKeySecret`和`fs.oss.endpoint`。

```shell
$ cat << EOF > dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: imagenet
spec:
  mounts:
    - mountPoint: oss://<OSS_BUCKET>/<OSS_DIRECTORY>/
      name: imagenet
      options:
        fs.oss.accessKeyId: <OSS_ACCESS_KEY_ID>
        fs.oss.accessKeySecret: <OSS_ACCESS_KEY_SECRET>
        fs.oss.endpoint: <OSS_ENDPOINT>
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: aliyun.accelerator/nvidia_name
              operator: In
              values:
                - Tesla-V100-SXM2-16GB
EOF
```

然后安装dataset

```shell
$ kubectl create -f dataset.yaml
```

查看dataset状态，为`NotBound`

```shell
$ kubectl describe dataset
Name:         imagenet
Namespace:    default
Labels:       <none>
Annotations:  <none>
API Version:  data.fluid.io/v1alpha1
Kind:         Dataset
Metadata:
  # more metadata
Spec:
  # more spec
Status:
  Conditions:
  Phase:  NotBound
Events:   <none>
```

### 创建Alluxio Runtime

Alluxio Runtime直接使用阿里云上的docker镜像`registry.cn-huhehaote.aliyuncs.com/alluxio/alluxio:2.3.0-SNAPSHOT-b7629dc`，这是目前为止支持AlluxioFuse `kernel_cache`模式的最为稳定的版本。本文档以四机八卡为例，所以设置`spec.worker.replicas=4`。此外，如下的`runtime.yaml`文件还设置了许多参数以优化dawnbench测试的IO性能。

```shell
$ cat << EOF > runtime.yaml
---
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: imagenet
  #namespace: fluid-system
spec:
  # Add fields here
  replicas: 4
  dataReplicas: 3
  alluxioVersion:
    image: registry.cn-huhehaote.aliyuncs.com/alluxio/alluxio
    imageTag: "2.3.0-SNAPSHOT-b7629dc"
    imagePullPolicy: Always
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /var/lib/docker/alluxio
        quota: 150Gi
        high: "0.99"
        low: "0.8"
        storageType: Memory
  properties:
    # jni-fuse related configurations
    alluxio.fuse.jnifuse.enabled: "true"
    alluxio.user.client.cache.enabled: "false"
    alluxio.user.client.cache.store.type: MEMORY
    alluxio.user.client.cache.dir: /alluxio/ram
    alluxio.user.client.cache.page.size: 2MB
    alluxio.user.client.cache.size: 1800MB
    # alluxio master
    alluxio.master.metastore: ROCKS
    alluxio.master.metastore.inode.cache.max.size: "10000000"
    alluxio.master.journal.log.size.bytes.max: 500MB
    alluxio.master.metadata.sync.concurrency.level: "128"
    alluxio.master.metadata.sync.executor.pool.size: "128"
    alluxio.master.metadata.sync.ufs.prefetch.pool.size: "128"
    # alluxio configurations
    alluxio.user.block.worker.client.pool.min: "512"
    alluxio.fuse.debug.enabled: "false"
    alluxio.web.ui.enabled: "false"
    alluxio.user.file.writetype.default: MUST_CACHE
    alluxio.user.ufs.block.read.location.policy: alluxio.client.block.policy.LocalFirstAvoidEvictionPolicy
    alluxio.user.block.write.location.policy.class: alluxio.client.block.policy.LocalFirstAvoidEvictionPolicy
    alluxio.worker.allocator.class: alluxio.worker.block.allocator.GreedyAllocator
    alluxio.user.block.size.bytes.default: 16MB
    alluxio.user.streaming.reader.chunk.size.bytes: 32MB
    alluxio.user.local.reader.chunk.size.bytes: 32MB
    alluxio.worker.network.reader.buffer.size: 32MB
    alluxio.worker.file.buffer.size: 320MB
    alluxio.user.metrics.collection.enabled: "false"
    alluxio.master.rpc.executor.max.pool.size: "1024"
    alluxio.master.rpc.executor.core.pool.size: "128"
    #alluxio.master.mount.table.root.readonly: "true"
    alluxio.user.update.file.accesstime.disabled: "true"
    alluxio.user.file.passive.cache.enabled: "false"
    alluxio.user.block.avoid.eviction.policy.reserved.size.bytes: 2GB
    alluxio.master.journal.folder: /journal
    alluxio.master.journal.type: UFS
    alluxio.user.block.master.client.pool.gc.threshold: 2day
    alluxio.user.file.master.client.threads: "1024"
    alluxio.user.block.master.client.threads: "1024"
    alluxio.user.file.readtype.default: CACHE
    alluxio.security.stale.channel.purge.interval: 365d
    alluxio.user.metadata.cache.enabled: "true"
    alluxio.user.metadata.cache.expiration.time: 2day
    alluxio.user.metadata.cache.max.size: "1000000"
    alluxio.user.direct.memory.io.enabled: "true"
    alluxio.fuse.cached.paths.max: "1000000"
    alluxio.job.worker.threadpool.size: "164"
    alluxio.user.worker.list.refresh.interval: 2min
    alluxio.user.logging.threshold: 1000ms
    alluxio.fuse.logging.threshold: 1000ms
    alluxio.worker.block.master.client.pool.size: "1024"
  master:
    replicas: 1
    jvmOptions:
      - "-Xmx6G"
      - "-XX:+UnlockExperimentalVMOptions"
      - "-XX:ActiveProcessorCount=8"
  worker:
    jvmOptions:
      - "-Xmx12G"
      - "-XX:+UnlockExperimentalVMOptions"
      - "-XX:MaxDirectMemorySize=32g"
      - "-XX:ActiveProcessorCount=8"
  fuse:
    image: registry.cn-huhehaote.aliyuncs.com/alluxio/alluxio-fuse
    imageTag: "2.3.0-SNAPSHOT-b7629dc"
    imagePullPolicy: Always
    env:
      SPENT_TIME: "1000"
    jvmOptions:
      - "-Xmx16G"
      - "-Xms16G"
      - "-XX:+UseG1GC"
      - "-XX:MaxDirectMemorySize=32g"
      - "-XX:+UnlockExperimentalVMOptions"
      - "-XX:ActiveProcessorCount=24"
    shortCircuitPolicy: local
    args:
      - fuse
      - --fuse-opts=kernel_cache,ro,max_read=131072,attr_timeout=7200,entry_timeout=7200
EOF
```

安装Alluxio Runtime

```shell
$ kubectl create -f runtime.yaml
```

检查Alluxio Runtime，`1`个Master，`4`个Worker和`4`个Fuse已成功部署

```shell
$ kubectl describe alluxioruntime imagenet 
Name:         imagenet
Namespace:    default
Labels:       <none>
Annotations:  <none>
API Version:  data.fluid.io/v1alpha1
Kind:         AlluxioRuntime
Metadata:
  # more metadata
Spec:
  # more spec
Status:
  Cache States:
    Cache Capacity:     600GiB
    Cached:             0B
    Cached Percentage:  0%
  Conditions:
    # more conditions
  Current Fuse Number Scheduled:    4
  Current Master Number Scheduled:  1
  Current Worker Number Scheduled:  4
  Desired Fuse Number Scheduled:    4
  Desired Master Number Scheduled:  1
  Desired Worker Number Scheduled:  4
  Fuse Number Available:            4
  Fuse Numb    Status:                True
    Type:                  Ready
  Phase:                   Bound
  Runtimes:
    Category:   Accelerate
    Name:       imagenet
    Namespace:  default
    Type:       alluxio
  Ufs Total:    143.7GiB
Events:         <none>
er Ready:                4
  Fuse Phase:                       Ready
  Master Number Ready:              1
  Master Phase:                     Ready
  Value File:                       imagenet-alluxio-values
  Worker Number Available:          4
  Worker Number Ready:              4
  Worker Phase:                     Ready
Events:                             <none>
```

同时，dataset也绑定到Alluxio Runtime

```shell
$ kubectl describe dataset
Name:         imagenet
Namespace:    default
Labels:       <none>
Annotations:  <none>
API Version:  data.fluid.io/v1alpha1
Kind:         Dataset
Metadata:
  # more metadata
Spec:
  # more spec
Status:
  Cache States:
    Cache Capacity:     600GiB
    Cached:             0B
    Cached Percentage:  0%
  Conditions:
    Last Transition Time:  2020-07-28T01:21:40Z
    Last Update Time:      2020-07-28T01:24:16Z
    Message:               The ddc runtime is ready.
    Reason:                DatasetReady
    Status:                True
    Type:                  Ready
  Phase:                   Bound
  Runtimes:
    Category:   Accelerate
    Name:       imagenet
    Namespace:  default
    Type:       alluxio
  Ufs Total:    143.7GiB
Events:         <none>
```

检查pv和pvc，名为imagenet的pv和pvc被成功创建。至此，云端数据集已成功部署到集群中

```shell
$ kubectl get pv,pvc
NAME                        CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM              STORAGECLASS   REASON   AGE
persistentvolume/imagenet   100Gi      RWX            Retain           Bound    default/imagenet                           7m11s

NAME                             STATUS   VOLUME     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/imagenet   Bound    imagenet   100Gi      RWX                           7m11s

```

## 执行测试

我们使用`arena`简化机器学习任务的部署流程。提交四机八卡训练任务：

```shell
arena submit mpi \
--name fluid-4x8-dawnbench-v2 \
--gpus=8 \
--workers=4 \
--working-dir=/perseus-demo/tensorflow-demo/ \
--data imagenet:/data \
-e DATA_DIR=/data/imagenet \
-e num_batch=1000  \
-e datasets_num_private_threads=8 \
--image=registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/perseus-benchmark-dawnbench-v2:centos7-cuda10.0-1.2.2-1.14-py36 \
"./launch-example.sh 4 8"
```

arena参数说明：

- `--name`：指定job的名字
- `--workers`：指定参与训练的机器（worker）数
- `--gpus`：指定每个worker使用的GPU数
- `--working-dir`：指定工作路径
- `--data`：挂载数据集（`imagenet`）到worker的`/data`目录
- `-e DATA_DIR`：指定数据集位置
- `./launch-example.sh 4 8`：启动四机八卡测试
