# 示例 - 用Fluid加速机器学习训练
[![](https://fluid-imgs.oss-cn-shanghai.aliyuncs.com/public/imgs/machineLearning.png)](https://fluid-imgs.oss-cn-shanghai.aliyuncs.com/public/video/machineLearning.mp4)

本文介绍如何使用Fluid部署[阿里云OSS](https://cn.aliyun.com/product/oss)云端[ImageNet](http://www.image-net.org/)数据集到Kubernetes集群，并使用[Arena](https://github.com/kubeflow/arena)在此数据集上训练ResNet-50模型。本文以四机八卡测试环境为例。

## 前提条件

- [Fluid](https://github.com/fluid-cloudnative/fluid) (version >= 0.1.0)
- [Arena](https://github.com/kubeflow/arena)（version >= 0.4.0）
- [Horovod](https://github.com/horovod/horovod) (version=0.18.1)
- [Benchmark](https://github.com/tensorflow/benchmarks/tree/cnn_tf_v1.14_compatible)

> **注意**：
>
> 1. 本文要求在Kubernetes集群中已安装好Fluid，如果您还没部署Fluid，请参考[Fluid安装手册](../guide/get_started.md)在您的Kubernetes集群上安装Fluid。
>
> 2. `Arena`是一个方便数据科学家运行和监视机器学习任务的CLI, 本文使用`Arena`提交机器学习任务，安装教程可参考[Arena安装教程](https://github.com/kubeflow/arena/blob/master/docs/installation/index.md)。
>
> 3. 本演示中使用的Horovod， TensorFlow和Benchmark代码均已经开源，您可以从上述链接中获取。  


## 用Fluid部署云端数据集

### 创建Dataset和Runtime

如下的`dataset.yaml`文件中定义了一个`Dataset`和`Runtime`，并`---`符号将它们的定义分割。

数据集存储在[阿里云OSS](https://cn.aliyun.com/product/oss)，为保证Alluxio能够成功挂载OSS上的数据集，请确保`dataset.yaml`文件中设置了正确的`mountPoint`、`fs.oss.accessKeyId`、`fs.oss.accessKeySecret`和`fs.oss.endpoint`。

> 你可以参考Alluxio的官方文档示例[Aliyun Object Storage Service](https://docs.alluxio.io/os/user/stable/en/ufs/OSS.html)，了解更多在Alluxio中使用OSS的例子。
>
> 如果您希望自己准备数据集，可以访问ImageNet官方网站 [http://image-net.org/download-images](http://image-net.org/download-images)。
>
> 如果你希望使用我们提供的数据集重现这个实验，请在社区开Issue申请数据集下载。

本文档以阿里云的V100四机八卡为例，所以在`dataset.yaml`中设置`spec.replicas=4`。为了保证数据被缓存在V100机器上，配置了`nodeAffinity`。此外，`dataset.yaml`文件还根据我们的测试经验设置了许多参数以优化Alluxio的IO性能（包括Alluxio、Fuse和JVM等层次），您可以自行根据机器配置和任务需求调整参数。

```shell
$ cat << EOF >> dataset.yaml
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
---
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: imagenet
spec:
  replicas: 4
  data:
    replicas: 1
#  alluxioVersion:
#    image: registry.cn-huhehaote.aliyuncs.com/alluxio/alluxio
#    imageTag: "2.3.0-SNAPSHOT-bbce37a"
#    imagePullPolicy: Always
  tieredstore:
    levels:
      - mediumtype: SSD
        path: /var/lib/docker/alluxio
        quota: 50Gi
        high: "0.99"
        low: "0.8"
EOF
```

创建Dataset和Runtime：

```shell
$ kubectl create -f dataset.yaml
```

检查Alluxio Runtime，可以看到`1`个Master，`4`个Worker和`4`个Fuse已成功部署：

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
    Cache Capacity:     200GiB
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
```

同时，检查到Dataset也绑定到Alluxio Runtime：

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
    Cache Capacity:     200GiB
    Cached:             0B
    Cached Percentage:  0%
  Conditions:
    Last Transition Time:  2020-08-18T11:01:09Z
    Last Update Time:      2020-08-18T11:02:48Z
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

检查pv和pvc，名为imagenet的pv和pvc被成功创建：

```shell
$ kubectl get pv,pvc
NAME                        CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM              STORAGECLASS   REASON   AGE
persistentvolume/imagenet   100Gi      RWX            Retain           Bound    default/imagenet                           7m11s

NAME                             STATUS   VOLUME     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/imagenet   Bound    imagenet   100Gi      RWX                           7m11s
```

至此，OSS云端数据集已成功部署到Kubernetes集群中。

## 示例：使用Arena提交深度学习任务

`Arena`提供了便捷的方式帮助用户提交和监控机器学习任务。在本文中，我们使用`Arena`简化机器学习任务的部署流程。

如果您已经安装`Arena`，并且云端数据集已成功部署到本地集群中，只需要简单执行以下命令便能提交ResNet50四机八卡训练任务：

```shell
arena submit mpi \
    --name horovod-resnet50-v2-4x8-fluid \
    --gpus=8 \
    --workers=4 \
    --working-dir=/horovod-demo/tensorflow-demo/ \
    --data imagenet:/data \
    -e DATA_DIR=/data/imagenet \
    -e num_batch=1000 \
    -e datasets_num_private_threads=8 \
    --image=registry.cn-hangzhou.aliyuncs.com/tensorflow-samples/horovod-benchmark-dawnbench-v2:0.18.1-tf1.14.0-torch1.2.0-mxnet1.5.0-py3.6 \
    ./launch-example.sh 4 8
```

Arena参数说明：

- `--name`：指定job的名字
- `--workers`：指定参与训练的节点（worker）数
- `--gpus`：指定每个worker使用的GPU数
- `--working-dir`：指定工作路径
- `--data`：挂载Volume `imagenet`到worker的`/data`目录
- `-e DATA_DIR`：指定数据集位置
- `./launch-example.sh 4 8`：运行脚本启动四机八卡测试

检查任务是否正常执行：

```shell
$ arena get horovod-resnet50-v2-4x8-fluid -e
STATUS: RUNNING
NAMESPACE: default
PRIORITY: N/A
TRAINING DURATION: 16s

NAME                           STATUS   TRAINER  AGE  INSTANCE                                      NODE
horovod-resnet50-v2-4x8-fluid  RUNNING  MPIJOB   16s  horovod-resnet50-v2-4x8-fluid-launcher-czlfn  192.168.1.21
horovod-resnet50-v2-4x8-fluid  RUNNING  MPIJOB   16s  horovod-resnet50-v2-4x8-fluid-worker-0        192.168.1.16
horovod-resnet50-v2-4x8-fluid  RUNNING  MPIJOB   16s  horovod-resnet50-v2-4x8-fluid-worker-1        192.168.1.21
horovod-resnet50-v2-4x8-fluid  RUNNING  MPIJOB   16s  horovod-resnet50-v2-4x8-fluid-worker-2        192.168.1.25
horovod-resnet50-v2-4x8-fluid  RUNNING  MPIJOB   16s  horovod-resnet50-v2-4x8-fluid-worker-3        192.168.3.29
```

如果您看到`4`个处于`RUNNING`状态的worker，说明您已经成功启动训练。

如果您想知道训练进行到哪一步了，请检查Arena日志：

```shell
$ arena logs --tail 100 -f horovod-resnet50-v2-4x8-fluid
```

## 环境清理

```shell
$ kubectl delete -f dataset.yaml
```
