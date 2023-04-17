# 示例 - 使用Fluid加速PVC
[![](https://fluid-imgs.oss-cn-shanghai.aliyuncs.com/public/imgs/accelerate_pvc.jfif)](https://fluid-imgs.oss-cn-shanghai.aliyuncs.com/public/video/accelerate_pvc.mp4)

## 测试场景：ResNet50 模型训练

- 测试机型： V100 x8
- 容量型nfs

## 配置

### 硬件配置

| Cluster | Alibaba Cloud Kubernetes. v1.16.9-aliyun.1             |
| ------- | ------------------------------------------------------ |
| ECS实例 | ECS   规格：ecs.gn6v-c10g1.20xlarge<br />    CPU：82核 |
| 分布式存储|    容量型NAS                                          |

###  软件配置

软件版本： 0.18.1-tf1.14.0-torch1.2.0-mxnet1.5.0-py3.6

## 前提条件

- [Fluid](https://github.com/fluid-cloudnative/fluid) (version >= 0.3.0)
- [Arena](https://github.com/kubeflow/arena)（version >= 0.4.0）
- [Horovod](https://github.com/horovod/horovod) (version=0.18.1)
- [Benchmark](https://github.com/tensorflow/benchmarks/tree/cnn_tf_v1.14_compatible)

## 数据准备

1. 下载数据集

```bash
$ wget http://imagenet-tar.oss-cn-shanghai.aliyuncs.com/imagenet.tar.gz
```

2. 解压数据集

```bash
$ tar -I pigz -xvf imagenet.tar.gz
```

## NFS dawnbench测试

### 部署数据集

1. 在NFS Server中挂载数据集

2. 使用Kubernetes创建nfs的volume

```bash
$ cat <<EOF > nfs.yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: nfs-imagenet
spec:
  capacity:
    storage: 150Gi
  volumeMode: Filesystem
  accessModes:
  - ReadOnlyMany
  persistentVolumeReclaimPolicy: Retain
  storageClassName: nfs
  mountOptions:
  - vers=3
  - nolock
  - proto=tcp
  - rsize=1048576
  - wsize=1048576
  - hard
  - timeo=600
  - retrans=2
  - noresvport
  - nfsvers=4.1
  nfs:
    path: <YOUR_PATH_TO_DATASET>
    server: <YOUR_NFS_SERVER>
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: nfs-imagenet
spec:
  accessModes:
  - ReadOnlyMany
  resources:
    requests:
      storage: 150Gi
  storageClassName: nfs
EOF
```

> **NOTE:**
>
> 修改上述yaml文件中的nfs的server和path为您的nfs server地址和挂载路径。

```bash
$ kubectl create -f nfs.yaml
```

3. 检查Kubernetes是否正常创建volume

```bash
$ kubectl get pv,pvc
NAME                            CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                  STORAGECLASS   REASON   AGE
persistentvolume/nfs-imagenet   150Gi      ROX            Retain           Bound    default/nfs-imagenet   nfs                     45s

NAME                                 STATUS   VOLUME         CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/nfs-imagenet   Bound    nfs-imagenet   150Gi      ROX            nfs            45s
```

### dawnbench

#### 单机八卡

```bash
arena submit mpi \
    --name horovod-resnet50-v2-1x8-nfs \
    --gpus=8 \
    --workers=1 \
    --working-dir=/horovod-demo/tensorflow-demo/ \
    --data nfs-imagenet:/data \
    -e DATA_DIR=/data/imagenet \
    -e num_batch=1000 \
    -e datasets_num_private_threads=8 \
    --image=registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/horovod-benchmark-dawnbench-v2:0.18.1-tf1.14.0-torch1.2.0-mxnet1.5.0-py3.6 \
    ./launch-example.sh 1 8
```

#### 四机八卡

```bash
arena submit mpi \
    --name horovod-resnet50-v2-4x8-nfs \
    --gpus=8 \
    --workers=4 \
    --working-dir=/horovod-demo/tensorflow-demo/ \
    --data nfs-imagenet:/data \
    -e DATA_DIR=/data/imagenet \
    -e num_batch=1000 \
    -e datasets_num_private_threads=8 \
    --image=registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/horovod-benchmark-dawnbench-v2:0.18.1-tf1.14.0-torch1.2.0-mxnet1.5.0-py3.6 \
    ./launch-example.sh 4 8
```

> **NOTE:**
>
> 训练完成后，arena保留了laucher，可能导致nfs删不掉。请在提交nfs删除命令后执行如下命令：
>
> ```bash
> $ kubectl patch pvc nfs-imagenet  -p '{"metadata":{"finalizers": []}}' --type=merge
> ```

## Fluid加速PVC

### 部署数据集

1. 按照前述步骤创建NFS的volume 
2. 部署Fluid加速刚才创建的PVC

```bash
$ cat <<EOF > dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: fluid-imagenet
spec:
  mounts:
  - mountPoint: pvc://nfs-imagenet
    name: nfs-imagenet
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: aliyun.accelerator/nvidia_name
              operator: In
              values:
                - Tesla-V100-SXM2-16GB
---
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: fluid-imagenet
spec:
  replicas: 4
  data:
    replicas: 1
  tieredstore:
    levels:
      - mediumtype: SSD
        path: /var/lib/docker/alluxio
        quota: 150Gi
        high: "0.99"
        low: "0.8"
EOF
```

> **NOTE:**
>
> - `spec.replicas`和dawnbench测试的worker数量保持一致。比如：单机八卡为1，四机八卡为4。
> - `nodeSelectorTerms`作用是限制在有V100显卡的机器上部署数据集，此处应根据实验环境具体调节。

```bash
$ kubectl create -f dataset.yaml
```

3. 检查部署

```bash
$ kubectl get pv,pvc
NAME                              CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                    STORAGECLASS   REASON   AGE
persistentvolume/fluid-imagenet   100Gi      RWX            Retain           Bound    default/fluid-imagenet                           1s
persistentvolume/nfs-imagenet     150Gi      ROX            Retain           Bound    default/nfs-imagenet     nfs                     16m

NAME                                   STATUS   VOLUME           CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/fluid-imagenet   Bound    fluid-imagenet   100Gi      RWX                           0s
persistentvolumeclaim/nfs-imagenet     Bound    nfs-imagenet     150Gi      ROX            nfs            16m
```

### dawnbench

#### 单机八卡

```bash
arena submit mpi \
    --name horovod-resnet50-v2-1x8-fluid \
    --gpus=8 \
    --workers=1 \
    --working-dir=/horovod-demo/tensorflow-demo/ \
    --data fluid-imagenet:/data \
    -e DATA_DIR=/data/nfs-imagenet/imagenet \
    -e num_batch=1000 \
    -e datasets_num_private_threads=8 \
    --image=registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/horovod-benchmark-dawnbench-v2:0.18.1-tf1.14.0-torch1.2.0-mxnet1.5.0-py3.6 \
    ./launch-example.sh 1 8
```

#### 四机八卡

```bash
arena submit mpi \
    --name horovod-resnet50-v2-4x8-fluid \
    --gpus=8 \
    --workers=4 \
    --working-dir=/horovod-demo/tensorflow-demo/ \
    --data fluid-imagenet:/data \
    -e DATA_DIR=/data/nfs-imagenet/imagenet \
    -e num_batch=1000 \
    -e datasets_num_private_threads=8 \
    --image=registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/horovod-benchmark-dawnbench-v2:0.18.1-tf1.14.0-torch1.2.0-mxnet1.5.0-py3.6 \
    ./launch-example.sh 4 8
```

## 测试结果

### horovod-1x8

|                         | nfs      | fluid (cold) | fluid (warm) |
| ----------------------- | -------- | ------------ | ------------ |
| 训练时间                | 3h49m10s | 3h50m40s     | 3h34m15s     |
| 1000步速度(images/second) | 2400.8   | 2378.4       | 9327.6       |
| 最终速度(images/second) | 8696.8   | 8692.8       | 9301.6       |
| steps                   | 56300    | 56300        | 56300        |
| Accuracy @ 5            | 0.9282   | 0.9286       | 0.9285       |

### horovod-4x8

|                         | nfs      | fluid (cold) | fluid (warm) |
| ----------------------- | -------- | ------------ | ------------ |
| 训练时间                | 2h15m59s | 1h43m43s     | 1h32m22s     |
| 1000步速度(images/second) | 3136     | 8889.6       | 20859.5      |
| 最终速度(images/second) | 15024    | 20506.3      | 21329        |
| steps                   | 14070    | 14070        | 14070        |
| Accuracy @ 5            | 0.9228   | 0.9204       | 0.9243       |


## 结果分析

从测试结果来看，单机八卡通过Fluid加速效果并没有明显的效果，但是在四机八卡的场景下Fluid加速效果非常明显。在热数据的场景下，可以缩短训练时间 (135-92)/135 = 31 %; 在冷数据场景下可以缩短训练时间 （135-103）/135 = 23 % 。 这是由于四机八卡下，NFS的带宽成为了瓶颈；而Fluid基于Alluxio提供了分布式缓存的P2P数据读取能力。

