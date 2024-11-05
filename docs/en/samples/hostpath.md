# 示例 - 使用Fluid加速主机目录

## 测试场景：ResNet50 模型训练

- 测试机型： V100 x8
- nfs地址：38037492dc-pol25.cn-shanghai.nas.aliyuncs.com

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

## 已知约束

- 通过主机目录实现挂载并不是推荐的使用方式，因为该方式依赖于Kubernetes意外的挂载点维护方式，实际上并不可靠，可能引发数据不一致的问题。

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

2. 将NFS挂载到主机目录上

```
$ sudo mount -t nfs -o vers=3,nolock,proto=tcp,rsize=1048576,wsize=1048576,hard,timeo=600,retrans=2,noresvport <YOUR_NFS_SERVER>:<YOUR_PATH_TO_DATASET> /mnt/nfs-imagenet
```

3. 查看NFS是否已成功挂载

```
$ mount | grep nfs
<YOUR_NFS_SERVER>:<YOUR_PATH_TO_DATASET> on /mnt/nfs-imagenet type nfs (rw,relatime,vers=3,rsize=1048576,wsize=1048576,namlen=255,hard,nolock,noresvport,proto=tcp,timeo=600,retrans=2,sec=sys,mountaddr=192.168.1.28,mountvers=3,mountport=2049,mountproto=tcp,local_lock=all,addr=192.168.1.28)
```

> **NOTE:**
>
> 修改上述命令中的`<YOUR_NFS_SERVER>`和`<YOUR_PATH_TO_DATASET>`为您的nfs server地址和挂载路径。


### dawnbench

#### 单机八卡

```
arena submit mpijob \
--name horovod-v2-nfs-hostpath-1x8-093000 \
--gpus=8 \
--workers=1 \
--working-dir=/horovod-demo/tensorflow-demo/ \
--data-dir /mnt/nfs-imagenet:/data \
-e DATA_DIR=/data/imagenet \
-e num_batch=1000 \
-e datasets_num_private_threads=8 \
--image=registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/horovod-benchmark-dawnbench-v2:0.18.1-tf1.14.0-torch1.2.0-mxnet1.5.0-py3.6 \
./launch-example.sh 1 8
```

#### 四机八卡
```
arena submit mpi \
--name horovod-v2-nfs-hostpath-4x8-092921 \
--gpus=8 \
--workers=4 \ 
--working-dir=/horovod-demo/tensorflow-demo/ \ 
--data-dir /mnt/nfs-imagenet:/data \
-e DATA_DIR=/data/imagenet \
-e num_batch=1000 \
-e datasets_num_private_threads=8 \
--image=registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/horovod-benchmark-dawnbench-v2:0.18.1-tf1.14.0-torch1.2.0-mxnet1.5.0-py3.6 \
./launch-example.sh 4 8
```

## Fluid加速主机目录

### 部署数据集
1. 按照前述步骤完成NFS的挂载
2. 部署Fluid加速NFS挂载的主机目录

```yaml
$ cat <<EOF > dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: imagenet
spec:
  mounts:
  - mountPoint: local:///mnt/nfs-imagenet
    name: imagenet
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
  name: imagenet
spec:
  replicas: 4
  data:
    replicas: 1
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /alluxio/ram 
        quota: 50Gi
        high: "0.99"
        low: "0.8"
EOF
```

> **NOTE:**
>
> - `mounts.mountPoint`通过`local://`的前缀来指明要挂载的主机目录(e.g. `/mnt/nfs-imagenet`)
> - `spec.replicas`和dawnbench测试的worker数量保持一致。比如：单机八卡为1，四机八卡为4
> - `nodeSelectorTerms`作用是限制在有V100显卡的机器上部署数据集，此处应根据实验环境具体调节

```
$ kubectl create -f dataset.yaml
```

3. 检查部署情况

```
$ kubectl get pv,pvc
NAME                        CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM              STORAGECLASS   REASON   AGE
persistentvolume/imagenet   100Gi      RWX            Retain           Bound    default/imagenet                           3h28m

NAME                             STATUS   VOLUME     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/imagenet   Bound    imagenet   100Gi      RWX                           3h28m
```

### dawnbench

#### 单机八卡

```
arena submit mpi \
--name horovod-v2-nfs-fluid-1x8-093009 \
--gpus=8 \
--workers=1 \
--working-dir=/horovod-demo/tensorflow-demo/ \
--data imagenet:/data \
-e DATA_DIR=/data/imagenet/imagenet \
-e num_batch=1000 \
-e datasets_num_private_threads=8 \
--image=registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/horovod-benchmark-dawnbench-v2:0.18.1-tf1.14.0-torch1.2.0-mxnet1.5.0-py3.6 \
./launch-example.sh 1 8
```

#### 四机八卡

```
arena submit mpi \
--name horovod-v2-nfs-fluid-4x8-092910 \
--gpus=8 \
--workers=4 \
--working-dir=/horovod-demo/tensorflow-demo/ \
--data imagenet:/data \
-e DATA_DIR=/data/imagenet/imagenet \
-e num_batch=1000 \
-e datasets_num_private_threads=8 \
--image=registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/horovod-benchmark-dawnbench-v2:0.18.1-tf1.14.0-torch1.2.0-mxnet1.5.0-py3.6 \
./launch-example.sh 4 8
```



## Experiment Results

### horovod-1x8

|                         | nfs-hostpath | fluid (cold) | fluid (warm) |
| ----------------------- | -------- | ------------ | ------------ |
| Training time                 | 4h20m36s | 4h21m56s     | 4h2m16s     |
| Speed at the 1000 step(images/second) | 2426.4  | 2467.2      | 8959.7      |
| Speed at the last step(images/second) | 8218.1   | 8219.8      | 8275.8      |
| steps                   | 56300    | 56300       | 56300       |
| Accuracy @ 5            | 0.9280   | 0.9288       | 0.9291      |

### horovod-4x8

|                         | nfs-hostpath | fluid (cold) | fluid (warm) |
| ----------------------- | ---------- | ------------ | ------------ |
| Training time                  | 2h9m21s   | 1h40m15s     | 1h29m55s     |
| Speed at the 1000 step(images/second) | 3219.2    | 11067.2      | 21951.3      |
| Speed at the last step(images/second) | 15855.7   | 20964.4      | 21869.8      |
| steps                   | 14070      | 14070        | 14070        |
| Accuracy @ 5            | 0.9227     | 0.9232       | 0.9228       |

## 结果分析


From the test results, the Fluid acceleration on 1x8 has no obvious enhancement, 
while in the improvements of 4x8, the effect is very obvious.
In warm data scenario, the training time can be shortened **(129-89)/129 = 31 %**;
In cold data scenario, training time can be shortened **（129-100）/129 = 22 %**.
This is because NFS bandwidth became a bottleneck under 4x8;
Fluid based on Alluxio provides distributed cache data reading capability for P2P data.


