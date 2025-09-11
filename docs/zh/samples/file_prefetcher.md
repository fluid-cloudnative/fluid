# 示例 - 使用Fluid文件预取优化应用文件读取效率

分布式缓存为应用提供了一个延时更低吞吐上限更高的数据源，但是这并不意味着应用读取一个已缓存文件的耗时一定能够大幅缩短。应用读取文件的实际耗时除了与数据访问链路上的“硬件条件”相关（例如： 客户端与分布式缓存之间可用网络带宽、延时等），应用自身读取文件的访问模式（Access Pattern）同样对实际耗时起到了决定性的作用。

例如：文件访问客户端的一种常用优化手段为预读（readahead）。预读优化提前将应用可能在未来需要读取的文件数据缓存在客户端中，降低应用读请求的访问延迟，提高对本地网络带宽资源的利用率。由于需要预测应用未来的读请求，预读优化通常仅在**顺序读取**某个文件时有较好的效果，如果应用倾向于随机读取文件，那么此时开启预读优化甚至可能造成读放大问题，导致读取文件的耗时不降反增。

本文档介绍如何使用Fluid文件预取功能，解决部分特定场景下因文件访问模式导致的带宽资源利用不充分的问题。

## Fluid文件预取适用场景

Fluid文件预取功能将远程文件提前预取至应用Pod所在节点的内存中，应用容器启动后对文件的随机读请求均会命中本地内存缓存，由于访问本地内存的延时远低于通过网络访问分布式缓存的延时，即使是随机读请求也能迅速返回，从而加速应用读取文件的实际耗时。

Fluid文件预取功能适用于**可提前确定应用访问的文件列表，且应用容器可用内存能够存储列表中全部文件**的场景。
- 示例场景1: AI推理服务模型加载加速。以Safetensors格式存储的模型参数文件加载时通常会产生随机读请求，预取功能能够提前将模型参数文件内容加载到内存中，从而加速模型加载过程。
- 示例场景2: DataFrame数据分析加速。对Numpy格式存储的DataFrame数据文件作分析时，应用SQL语句可能产生对Numpy文件的随机读请求，预取功能能够提前将Numpy文件内容加载到内存中，从而加速数据分析过程。

## Fluid文件预取功能使用示例

### 前提条件

- Fluid(version >= 1.0.6)
请参考[Fluid安装文档](https://github.com/fluid-cloudnative/fluid/blob/master/docs/zh/userguide/install.md)完成安装

### 使用示例

1. 创建Dataset和Runtime

Fluid预取能力兼容所有类型的Runtime，在本示例中我们以AlluxioRuntime为例进行说明：

```yaml
cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: mydataset
spec:
  mounts:
    - mountPoint: https://mirrors.tuna.tsinghua.edu.cn/apache/zookeeper/stable/
      name: zookeeper
      path: /
---
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: mydataset
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

2. 等待Runtime初始化完成并与Dataset绑定

```
$ kubectl wait --for=condition=Ready dataset mydataset --timeout=120s
```

等待一段时间后，上述命令返回以下结果，说明Dataset和Runtime已经初始化成功：
```
dataset.data.fluid.io/mydataset condition met
```

3. 创建应用Pod，并在应用Pod上启用文件预取能力

这里使用nginx镜像创建一个示例应用Pod：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: demo
  labels:
    fuse.serverful.fluid.io/inject: "true"
  annotations:
    file-prefetcher.fluid.io/inject: "true"
    ### optional annotations:
    # file-prefetcher.fluid.io/file-list: "pvc://mydataset/*.tar.gz"
    # file-prefetcher.fluid.io/async-prefetch: "false"
    # file-prefetcher.fluid.io/extra-envs: "FOO=BAR"
    # file-prefetcher.fluid.io/image: "<file-prefetcher-image>"
spec:
  restartPolicy: Always
  containers:
    - name: demo
      image: nginx
      command:
      - "bash"
      - "-c"
      args:
      - "sleep inf"
      volumeMounts:
        - mountPath: /data/
          name: data-vol
  volumes:
    - name: data-vol
      persistentVolumeClaim:
        claimName: mydataset
```

上述示例中各个参数含义如下：

| 分类 | 参数 | 默认值 | 描述 |
| ---- | ---- | ---- | ---- |
| labels | `fuse.serverful.fluid.io/inject` | `false`| 控制该Pod是否经过Fluid处理的开关，仅当配置为'true'的Pod才会被Fluid处理 |
| annotations | `file-prefetcher.fluid.io/inject` | `false` | 控制是否对该Pod开启文件预取功能，功能开启后Fluid会为该Pod注入一个文件预取Sidecar容器 |
| annotations | `file-prefetcher.fluid.io/file-list` | 默认值为当前Pod中挂载的所有Fluid PVC下的所有文件。等同于`pvc://<pvc1>/**;pvc://<pvc2>/**;pvc://<pvc3>/**` | 指定待预读的文件列表，为可选参数。支持多个列表项，各个列表项之间可通过分号（;）分隔。每个列表项格式必须为pvc://<pvc_name>/<glob_path>，其中<pvc_name>必须是Pod挂载的Fluid Dataset对应的PVC，<glob_path>为支持glob语法的字符串。例如：`pvc://zookeeper/*.tar.gz`表示预取`zookeeper` PVC下所有以.tar.gz结尾的文件 |
| annotations| `file-prefetcher.fluid.io/async-prefetch` | `false` | 指定文件预取Sidecar容器的启动是否阻塞后续应用容器的启动，为可选参数。配置为`true`后，无法保证应用容器启动时所有文件都预取成功。 |
| annotations| `file-prefetcher.fluid.io/prefetch-timeout-seconds` | `120` | 仅在async-prefetch=false时生效。指定主容器等待预取完成的最长等待时间。 |
| annotations| `file-prefetcher.fluid.io/extra-envs` | `<none>` | 指定为文件预取Sidecar容器添加的额外环境变量，为可选参数。格式为`ENV1=value1 ENV2=value2` |
| annotations| `file-prefetcher.fluid.io/image` | fluid内置的预取镜像 | 指定文件预取Sidecar容器使用的镜像，为可选参数。 |


创建Pod
```
$ kubctl create -f pod.yaml
```
    
4. 查看文件预取优化效果

查看Pod运行状态。在上述Pod配置下，由于文件预取Sidecar容器将阻塞后续应用容器启动，因此应用容器的启动时间可能比正常Pod启动时间长。
```
$ kubectl wait --for=condition=Ready pod demo
```
查看到如下返回结果：
```
pod/demo condition met
```

登陆到Pod中，使用vmtouch工具验证各个文件是否已经预取到本地内存中：
```
$ kubectl exec -it demo -c demo  -- /bin/bash
$ apt update && apt install -y vmtouch
$ vmtouch /data/*
```
