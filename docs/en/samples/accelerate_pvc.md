# Demo - Accelerate PVC with Fluid

## Test scenario: ResNet50 model training

- Machine： V100 x8
- NFS Server：38037492dc-pol25.cn-shanghai.nas.aliyuncs.com

## Settings

### Hardware

| Cluster | Alibaba Cloud Kubernetes. v1.16.9-aliyun.1             |
| ------- | ------------------------------------------------------ |
| ECS Instance | ECS   specifications：ecs.gn6v-c10g1.20xlarge<br />    CPU：82 cores |
| Distributed Storage|    NAS                                          |

### Software

Software version: 0.18.1-tf1.14.0-torch1.2.0-mxnet1.5.0-py3.6

## Prerequisites

- [Fluid](https://github.com/fluid-cloudnative/fluid) (version >= 0.3.0)
- [Arena](https://github.com/kubeflow/arena)（version >= 0.4.0）
- [Horovod](https://github.com/horovod/horovod) (version=0.18.1)
- [Benchmark](https://github.com/tensorflow/benchmarks/tree/cnn_tf_v1.14_compatible)

## Prepare Dataset

1. Download

```bash
$ wget http://imagenet-tar.oss-cn-shanghai.aliyuncs.com/imagenet.tar.gz
```

2. Unpack

```bash
$ tar -I pigz -xvf imagenet.tar.gz
```

## NFS dawnbench

### Deploy Dataset

1. Export Dataset on Your NFS Server

2. Create Volume using Kubernetes

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
> Please replace `YOUR_PATH_TO_DATASET` and `YOUR_NFS_SERVER` 
> with your own nfs server address and path to dataset.

```bash
$ kubectl create -f nfs.yaml
```

3. Check Volume

```bash
$ kubectl get pv,pvc
NAME                            CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                  STORAGECLASS   REASON   AGE
persistentvolume/nfs-imagenet   150Gi      ROX            Retain           Bound    default/nfs-imagenet   nfs                     45s

NAME                                 STATUS   VOLUME         CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/nfs-imagenet   Bound    nfs-imagenet   150Gi      ROX            nfs            45s
```

### Dawnbench

#### 1x8

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

#### 4x8

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
> If you find that nfs volume cannot be deleted,
> this is because Arena will leave a launcher pod after training finished,
> and Kubernetes still thinks that volume is in using.
>
> Just execute following command to force deleting volume:
> ```bash
> $ kubectl patch pvc nfs-imagenet  -p '{"metadata":{"finalizers": []}}' --type=merge
> ```

## Accelerate PVC with Fluid

### Deploy Dataset

1. Follow Previous Steps to Create NFS Volume
2. Deploy Fluid to Accelerate NFS Volume

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
> - Please keep `spec.replicas` consistent with the number of machines you are going to use for machine learning。
> - `nodeSelectorTerms` is used to restrict scheduling on machines with V100 GPU only.

```bash
$ kubectl create -f dataset.yaml
```

3. Check Volume

```bash
$ kubectl get pv,pvc
NAME                              CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                    STORAGECLASS   REASON   AGE
persistentvolume/fluid-imagenet   100Gi      RWX            Retain           Bound    default/fluid-imagenet                           1s
persistentvolume/nfs-imagenet     150Gi      ROX            Retain           Bound    default/nfs-imagenet     nfs                     16m

NAME                                   STATUS   VOLUME           CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/fluid-imagenet   Bound    fluid-imagenet   100Gi      RWX                           0s
persistentvolumeclaim/nfs-imagenet     Bound    nfs-imagenet     150Gi      ROX            nfs            16m
```

### Dawnbench

#### 1x8

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

#### 4x8

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

## Experiment Results

### horovod-1x8

|                         | nfs      | fluid (cold) | fluid (warm) |
| ----------------------- | -------- | ------------ | ------------ |
| Training time                | 3h49m10s | 3h50m40s     | 3h34m15s     |
| Speed at the 1000 step(images/second) | 2400.8   | 2378.4       | 9327.6       |
| Speed at the last step(images/second) | 8696.8   | 8692.8       | 9301.6       |
| steps                   | 56300    | 56300        | 56300        |
| Accuracy @ 5            | 0.9282   | 0.9286       | 0.9285       |

### horovod-4x8

|                         | nfs      | fluid (cold) | fluid (warm) |
| ----------------------- | -------- | ------------ | ------------ |
| Training time                | 2h15m59s | 1h43m43s     | 1h32m22s     |
| Speed at the 1000 step(images/second) | 3136     | 8889.6       | 20859.5      |
| Speed at the last step(images/second) | 15024    | 20506.3      | 21329        |
| steps                   | 14070    | 14070        | 14070        |
| Accuracy @ 5            | 0.9228   | 0.9204       | 0.9243       |


## 结果分析

From the test results, the Fluid acceleration effect on 1x8 has no obvious effect,
but in the scenario of 4x8, the effect is very obvious.
In warm data scenario, the training time can be shortened (135-92)/135 = 31%;
In cold data scenario, training time can be shortened (135-103) /135 = 23%.
This is because NFS bandwidth became a bottleneck under 4x8;
Fluid based on Alluxio provides distributed cache data reading capability for P2P data.
