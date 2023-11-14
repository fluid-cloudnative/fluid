# DEMO - Accelerate Machine Learning Training with Fluid

This article describes how to deploy [ImageNet](http://www.image-net.org/) dataset stored on [Aliyun OSS](https://cn.aliyun.com/product/oss) to Kubernetes cluster with Fluid, and train a ResNet-50 model on this dataset using [Arena](https://github.com/kubeflow/arena). In this article, we perform machine learning training on 4 nodes, each node with 8 GPU cards.

## Prerequisites

- [Fluid](https://github.com/fluid-cloudnative/fluid) (version >= 0.1.0)
- [Arena](https://github.com/kubeflow/arena)（version >= 0.4.0）
- [TensorFlow](https://github.com/tensorflow/tensorflow) (version = 1.14)
- [Horovod](https://github.com/horovod/horovod) (version=0.18.1)
- [Benchmark](https://github.com/tensorflow/benchmarks/tree/cnn_tf_v1.14_compatible)

> **NOTE**:
>
> 1. The document requires Fluid installed on your Kubernetes cluster. Please refer to [Fluid Installation Guide](../userguide/install.md) to finish installation before going to the next step.
>
> 2. Arena is a CLI that is convenient for data scientists to run and monitor machine learning tasks. See [Arena Installation Tutorial](https://github.com/kubeflow/arena/blob/master/docs/installation/INSTALL_FROM_BINARY.md) for more information.
>
> 3. This Demo uses Horovod + TensorFlow + Benchmark, they are all open source version. 


## Deploy Dataset on Kubernetes Cluster with Fluid

### Create Dataset and Runtime

The following `dataset.yaml` file defined a `Dataset` and `Runtime`  separated by `---`.

The dataset is stored on [Alibaba Cloud OSS](https://cn.aliyun.com/product/oss). To ensure that Alluxio can successfully mount the dataset, please make sure that configurations in the `dataset.yaml` are correct set, including `mountPoint`, `fs.oss.accessKeyId`, `fs.oss.accessKeySecret` and `fs.oss.endpoint`. 

> See Alluxio's official document [Aliyun Object Storage Service](https://docs.alluxio.io/os/user/stable/en/ufs/OSS.html) for more examples of using OSS in Alluxio.
>
> If you'd like to prepare dataset by yourself， please download the dataset from [http://image-net.org/download-images](http://image-net.org/download-images).
>
> If you'd like to download the imagenet dataset from us, please open an issue in Fluid community to ask for it.

This document takes 4 machines to training machine learning tasks, so `spec.replicas` is set to `4`. To ensure that the data is cached on the V100 machine, `nodeAffinity` is configured. In addition, the following configuration file `dataset.yaml` also sets many parameters based on our experience to optimize the IO performance of Alluxio in machine learning tasks, including Alluxio, Fuse and JVM levels. You can adjust these parameters according to the test environment and task requirements.

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

Create Dataset and Alluxio Runtime with:

```shell
$ kubectl create -f dataset.yaml
```

Check the status Alluxio Runtime, and  there should be `1` Master，`4` Worker and `4` Fuse running:

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

At the same time, Dataset is bound to Alluxio Runtime:

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

A pv and pvc named `imagenet` are successfully created. So far, the dataset stored on cloud has been successfully deployed to the Kubernetes cluster.

```shell
$ kubectl get pv,pvc
NAME                        CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM              STORAGECLASS   REASON   AGE
persistentvolume/imagenet   100Gi      RWX            Retain           Bound    default/imagenet                           7m11s

NAME                             STATUS   VOLUME     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/imagenet   Bound    imagenet   100Gi      RWX                           7m11s
```

## Example: Run Deep Learning Frameworks Using Arena

`Arena` provides a convenient way to help users submit and monitor machine learning tasks. In this article, we use `Arena` to simplify the deployment process of machine learning tasks.

If you have installed `Arena` and dataset has been successfully deployed to the local cluster, you can start training a ResNet50 model by simply executing the following command:

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

Notes：

- `--name`：specify the name of job, `horovod-resnet50-v2-4x8-fluid` in this example
- `--workers`：specify the number of nodes (workers) participating in training
- `--gpus`：specify the number of GPUs used by each worker
- `--working-dir`：specify working directory
- `--data`：tell workers to mount a volume named `imagenet` to the directory `/data` 
- `-e DATA_DIR`：specify the directory where dataset locates
- `./launch-example.sh 4 8`：run shell script to launch training process

Check whether the task is executed normally:

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

If you find that `4` workers are in the `RUNNING` state, congratulations!  It means that you have successfully started the training.

If you want to know where the training is going, please check the Arena log:

```shell
$ arena logs --tail 100 -f horovod-resnet50-v2-4x8-fluid
```

## Clean Up

```shell
$ kubectl delete -f dataset.yaml
```
