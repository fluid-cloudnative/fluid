# 示例 - 如何运行在以Knative为例的Serverless环境中

本示例以开源框架Knative为例子，演示如何在Serverless环境中通过Fluid进行统一的数据加速，本例子以AlluxioRuntime为例，实际上Fluid支持所有的Runtime运行在Serverless环境。

## 安装

1.根据[Knative文档](https://knative.dev/docs/install/serving/install-serving-with-yaml/)安装Knative Serving v1.2，需要开启[kubernetes.Deploymentspec-persistent-volume-claim](https://github.com/knative/serving/blob/main/config/core/configmaps/features.yaml#L156)。

检查 Knative的组件是否正常运行

```
kubectl get Deployments -n knative-serving
```

> 注：本文只是作为演示目的，关于Knative的生产系统安装请参考Knative文档最佳实践进行部署。另外由于Knative的容器镜像都在gcr.io镜像仓库，请确保镜像可达。
如果您使用的是阿里云，您也可以直接使用[阿里云ACK的托管服务](https://help.aliyun.com/document_detail/121508.html)降低配置Knative的复杂度。

2. 请参考[安装文档](../guide/install.md)安装Fluid最新版, 安装后检查 Fluid 各组件正常运行（本文以 AlluxioRuntime 为例）：

```shell
$ kubectl get deploy -n fluid-system
NAME                        READY   UP-TO-DATE   AVAILABLE   AGE
alluxioruntime-controller   1/1     1            1           18m
dataset-controller          1/1     1            1           18m
fluid-webhook               1/1     1            1           18m
```

通常来说，可以看到一个名为 `dataset-controller` 的 Deployment、一个名为 `alluxioruntime-controller` 的 Deployment以及一个名为 `fluid-webhook` 的 Deployment。

## 配置

**为namespace添加标签**

为namespace添加标签fluid.io/enable-injection后，可以开启此namespace下Pod的调度优化功能

```bash
$ kubectl label namespace default fluid.io/enable-injection=true
```

## 运行示例

**创建 dataset 和 runtime**

针对不同类型的 runtime 创建相应的 Runtime 资源，以及同名的 Dataset。这里以 AlluxioRuntime 为例, 以下为Dataset内容

```yaml
$ cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: serverless-data
spec:
  mounts:
    - mountPoint: https://mirrors.bit.edu.cn/apache/hbase/stable/
      name: hbase
      path: "/"
  accessModes:
    - ReadOnlyMany
---
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: serverless-data
spec:
  replicas: 2
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 2Gi
        high: "0.95"
        low: "0.7"
EOF
```

执行创建Dataset操作

```
$ kubectl create -f dataset.yaml
```

查看Dataset状态


```shell
$ kubectl get alluxio
NAME              MASTER PHASE   WORKER PHASE   FUSE PHASE   AGE
serverless-data   Ready          Ready          Ready        4m52s
$ kubectl get dataset
NAME              UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
serverless-data   566.22MiB        0.00B    4.00GiB          0.0%                Bound   4m52s
```

**创建 Knative Serving 资源对象**

```yaml
$ cat<<EOF >serving.yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: model-serving
spec:
  template:
    metadata:
      labels:
        app: model-serving
        serverless.fluid.io/inject: "true"
      annotations:
        autoscaling.knative.dev/target: "10"
        autoscaling.knative.dev/scaleDownDelay: "30m"
        autoscaling.knative.dev/minScale: "1"
    spec:
      containers:
        - image: fluidcloudnative/serving
          ports:
            - name: http1
              containerPort: 8080
          env:
            - name: TARGET
              value: "World"
          volumeMounts:
            - mountPath: /data
              name: data
              readOnly: true
      volumes:
        - name: data
          persistentVolumeClaim:
            claimName: serverless-data
            readOnly: true
  EOF
$ kubectl create -f serving.yaml
service.serving.knative.dev/model-serving created
```

请在podSpec或者podTemplateSpec中的label中配置`serverless.fluid.io/inject: "true"`


**查看 Knative Serving 是否创建，并检查 fuse-container 是否注入**

```shell
$ kubectl get po
NAME                                              READY   STATUS    RESTARTS   AGE
model-serving-00001-deployment-64d674d75f-46vvf   3/3     Running   0          76s
serverless-data-master-0                          2/2     Running   0          16m
serverless-data-worker-0                          2/2     Running   0          16m
serverless-data-worker-1                          2/2     Running   0          16m
$ kubectl get po model-serving-00001-deployment-64d674d75f-46vvf -oyaml| grep -i fluid-fuse -B 3
          - /opt/alluxio/integration/fuse/bin/alluxio-fuse
          - unmount
          - /runtime-mnt/alluxio/default/serverless-data/alluxio-fuse
    name: fluid-fuse
```

查看 Knative Serving 启动速度,可以看到启动加载数据的时间是**92s**

```shell
$ kubectl logs model-serving-00001-deployment-64d674d75f-46vvf -c user-container
Begin loading models at 16:29:02

real  1m32.639s
user  0m0.001s
sys 0m1.305s
Finish loading models at 16:29:45
2022-02-15 16:29:45 INFO Hello world sample started.
``****`

清理knative serving实例

```
$ kubectl delete -f serving.yaml
```

**执行数据预热**

创建dataload对象，并查看状态：

```yaml
$ cat<<EOF >dataload.yaml
apiVersion: data.fluid.io/v1alpha1
kind: DataLoad
metadata:
  name: serverless-dataload
spec:
  dataset:
    name: serverless-data
    namespace: default
  EOF
$ kubectl create -f dataload.yaml
dataload.data.fluid.io/serverless-dataload created
$ kubectl get dataload
NAME                  DATASET           PHASE      AGE     DURATION
serverless-dataload   serverless-data   Complete   2m43s   34s
```

检查此时的缓存状态, 目前已经将数据完全缓存到集群中

```
$ kubectl get dataset
NAME              UFS TOTAL SIZE   CACHED      CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
serverless-data   566.22MiB        566.22MiB   4.00GiB          100.0%              Bound   33m
```

再次创建Knative服务：

```shell
$ kubectl create -f serving.yaml
service.serving.knative.dev/model-serving created
```

此时查看启动时间发现当前启动加载数据的时间是**3.66s**, 变成没有预热的情况下性能的**1/20**

```
kubectl logs model-serving-00001-deployment-6cb54f94d7-dbgxf -c user-container
Begin loading models at 18:38:23

real  0m3.666s
user  0m0.000s
sys 0m1.367s
Finish loading models at 18:38:25
2022-02-15 18:38:25 INFO Hello world sample started.
```

> 注： 本例子使用的是Knative serving，如果您没有Knative环境，也可以使用Deployment进行实验。

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: model-serving
spec:
  selector:
    matchLabels:
      app: model-serving
  template:
    metadata:
      labels:
        app: model-serving
        serverless.fluid.io/inject: "true"
    spec:
      containers:
        - image: fluidcloudnative/serving
          name: serving
          ports:
            - name: http1
              containerPort: 8080
          env:
            - name: TARGET
              value: "World"
          volumeMounts:
            - mountPath: /data
              name: data
      volumes:
        - name: data
          persistentVolumeClaim:
            claimName: serverless-data
```

> 注：默认的sidecar注入模式是不会开启缓存目录短路读，如果您需要开启该能力，可以在labels中通过配置参数`cachedir.sidecar.fluid.io/inject`为`true`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: model-serving
spec:
  selector:
    matchLabels:
      app: model-serving
  template:
    metadata:
      labels:
        app: model-serving
        serverless.fluid.io/inject: "true"
        cachedir.sidecar.fluid.io/inject: "true"
    spec:
      containers:
        - image: fluidcloudnative/serving
          name: serving
          ports:
            - name: http1
              containerPort: 8080
          env:
            - name: TARGET
              value: "World"
          volumeMounts:
            - mountPath: /data
              name: data
      volumes:
        - name: data
          persistentVolumeClaim:
            claimName: serverless-data
```
