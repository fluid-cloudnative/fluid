## 使用Vineyard Runtime加速Kubeflow pipeline

Vineyard相比现有方法如本地文件或S3服务, 可以通过利用共享内存加速数据共享。在本文中，我们将展示如何使用Vineyard Runtime来加速Fluid平台上现有的Kubeflow pipeline。

### 前提条件

- 安装 argo CLI 工具，参见[官方指南](https://github.com/argoproj/argo-workflows/releases/).

### pipline概述

目前我们使用的pipeline是一个简单的pipeline，它是在虚拟的波士顿房价数据集上训练一个线性回归模型。它包含三个步骤：[数据预处理](../../../../samples/vineyard/preprocess-data/preprocess-data.y), [模型训练](../../../../samples/vineyard/train-data/train-data.py), 和 [模型测试](../../../../samples/vineyard/test-data/test-data.py).


### 准备环境

- 准备一个kubernetes集群。如果您没有现成的kubernetes集群，您可以使用以下命令通过kind(v0.20.0+)创建一个kubernetes集群：

```shell
cat <<EOF | kind create cluster -v 5 --name=kind --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  image: kindest/node:v1.23.0
- role: worker
  image: kindest/node:v1.23.0
- role: worker
  image: kindest/node:v1.23.0
- role: worker
  image: kindest/node:v1.23.0
EOF
```

- 准备kubeflow pipeline的docker镜像。您可以使用以下命令构建docker镜像：

```shell
$ make docker-build REGISTRY="test-registry"(Replace with your custom registry)
```

接下来，您可以将这些镜像推送到您的kubernetes集群可以访问的镜像仓库，或者如果您的kubernetes集群是通过kind创建的，则可以将这些镜像加载到集群中。

```shell
$ make load-images REGISTRY="test-registry"(Replace with your custom registry)
```

or

```shell
$ make push-images REGISTRY="test-registry"(Replace with your custom registry)
```

### 安装fluid平台和kubeflow组件

- 参考[安装文档](../../../userguide/install.md)完成安装。

- 安装argo workflow控制器，它可以作为kubeflow pipeline的后端。您可以使用以下命令安装argo workflow控制器：

```shell
$ kubectl create ns argo
$ kubectl apply -n argo -f https://github.com/argoproj/argo-workflows/releases/download/v3.4.8/install.yaml
```

- 安装vineyard runtime和dataset。您可以使用以下命令安装vineyard runtime和dataset：

```shell
$ cat <<EOF | kubectl apply -f -
apiVersion: data.fluid.io/v1alpha1
kind: VineyardRuntime
metadata:
  name: vineyard
spec:
  replicas: 2
  tieredstore:
    levels:
    - mediumtype: MEM
      quota: 20Gi
---
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: vineyard
EOF
```

### 运行pipeline

请确保`pipeline.yaml`和`pipeline-with-vineyard.yaml`文件是原始的文件，因为它们已经被修改为使用特权容器来清除页面缓存。然后您可以按照下面的步骤运行pipeline：

**注意** 您需要将**NAS路径**挂载到kubernetes节点。在这里，我们将NAS路径挂载到所有kubernetes节点的`/mnt/csi-benchmark`(在`prepare-data.yaml`中显示)路径上。接下来，我们需要通过运行以下命令来准备数据集：

```shell
$ kubectl apply -f samples/vineyard/prepare-data.yaml
```

数据集将存储在主机路径中。此外，您可能需要等待一段时间以生成数据集，您可以使用以下命令来检查状态：

```shell
$ while ! kubectl logs -l app=prepare-data | grep "preparing data time" >/dev/null; do echo "dataset unready, waiting..."; sleep 5; done && echo "dataset ready"
```

在运行pipeline之前，我们需要为pipeline创建一些rbac roles，如下所示。

```shell
$ kubectl apply -f samples/vineyard/rbac.yaml
```

之后，您可以通过以下命令运行不使用vineyard的pipeline：

```shell
$ argo submit samples/vineyard/pipeline.yaml -p data_mu
ltiplier=2000 -p registry="test-registry" 
Name:                machine-learning-pipeline-z72gm
Namespace:           default
ServiceAccount:      pipeline-runner
Status:              Pending
Created:             Wed Apr 03 11:46:43 +0800 (now)
Progress:            
Parameters:          
  data_multiplier:   2000
  registry:          test-registry
```

结果如下所示。

```shell
$ argo get machine-learning-pipeline-z72gm                                           
Name:                machine-learning-pipeline-z72gm
Namespace:           default
ServiceAccount:      pipeline-runner
Status:              Succeeded
Conditions:          
 PodRunning          False
 Completed           True
Created:             Wed Apr 03 11:46:43 +0800 (3 minutes ago)
Started:             Wed Apr 03 11:46:43 +0800 (3 minutes ago)
Finished:            Wed Apr 03 11:49:23 +0800 (49 seconds ago)
Duration:            2 minutes 40 seconds
Progress:            3/3
ResourcesDuration:   4m8s*(1 cpu),4m8s*(100Mi memory)
Parameters:          
  data_multiplier:   2000
  registry:          test-registry

STEP                                TEMPLATE                   PODNAME                                                     DURATION  MESSAGE
 ✔ machine-learning-pipeline-z72gm  machine-learning-pipeline                                                                          
 ├─✔ preprocess-data                preprocess-data            machine-learning-pipeline-z72gm-preprocess-data-4229626381  1m          
 ├─✔ train-data                     train-data                 machine-learning-pipeline-z72gm-train-data-1389575193       45s         
 └─✔ test-data                      test-data                  machine-learning-pipeline-z72gm-test-data-2535188255        13s
```

在运行使用vineyard的pipeline之前，您需要为vineyard runtime启用最佳调度策略，如下所示：

```shell
# 开启fuse亲和性调度
$ kubectl edit configmap webhook-plugins -n fluid-system
data:
  pluginsProfile: |
    pluginConfig:
    - args: |
        preferred:
          - name: fluid.io/fuse
            weight: 100
    ...

# 重启fluid-webhook pod
$ kubectl delete pod -lcontrol-plane=fluid-webhook -n fluid-system
```

接下来，您可以通过以下命令运行使用vineyard的pipeline：

```shell
$ argo submit samples/vineyard/pipeline-with-vineyard.yaml -p data_multiplier=2000 -p registry="test-registry"
Name:                machine-learning-pipeline-with-vineyard-q4tfr
Namespace:           default
ServiceAccount:      pipeline-runner
Status:              Pending
Created:             Wed Apr 03 12:00:45 +0800 (now)
Progress:            
Parameters:          
  data_multiplier:   2000
  registry:          test-registry
```

结果如下所示。

```shell
$ argo get machine-learning-pipeline-with-vineyard-q4tfr                               
Name:                machine-learning-pipeline-with-vineyard-q4tfr
Namespace:           default
ServiceAccount:      pipeline-runner
Status:              Succeeded
Conditions:          
 PodRunning          False
 Completed           True
Created:             Wed Apr 03 12:00:45 +0800 (2 minutes ago)
Started:             Wed Apr 03 12:00:45 +0800 (2 minutes ago)
Finished:            Wed Apr 03 12:02:36 +0800 (34 seconds ago)
Duration:            1 minute 51 seconds
Progress:            3/3
ResourcesDuration:   2m40s*(1 cpu),2m40s*(100Mi memory)
Parameters:          
  data_multiplier:   2000
  registry:          test-registry

STEP                                              TEMPLATE                                 PODNAME                                                                  DURATION  MESSAGE
 ✔ machine-learning-pipeline-with-vineyard-q4tfr  machine-learning-pipeline-with-vineyard                                                                                       
 ├─✔ preprocess-data                              preprocess-data                          machine-learning-pipeline-with-vineyard-q4tfr-preprocess-data-869469459  55s         
 ├─✔ train-data                                   train-data                               machine-learning-pipeline-with-vineyard-q4tfr-train-data-4177295571      26s         
 └─✔ test-data                                    test-data                                machine-learning-pipeline-with-vineyard-q4tfr-test-data-1755965473       8s
```

从结果中可以看出，使用vineyard的pipeline比不使用vineyard的pipeline减少了大约30%的运行时间。这种改进是由于vineyard利用共享内存加速数据共享，比传统方法如NFS更有效。

### Modifications to use vineyard

相比原始的kubeflow pipeline，您可以使用以下命令来查看使用Vineyard需要的修改：

```shell
$ git diff --no-index --unified=40 samples/vineyard/pipeline.py samples/vineyard/pipeline-with-vineyard.py
```

主要的修改如下：
- 为pipeline添加vineyard持久卷。这个持久卷用于将vineyard客户端配置文件挂载到pipeline中。

此外，您可以按照以下步骤检查源代码的修改。

- [Save data to vineyard](../../../../samples/vineyard/preprocess-data/preprocess-data.py#L32-L35).
- [Load data from vineyard](../../../../samples/vineyard/train-data/train-data.py#L15-L16).
- [load data from vineyard](../../../../samples/vineyard/test-data/test-data.py#L14-L15).

其中主要的修改是使用vineyard来加载和保存数据，而不是使用文件。
