# JuiceFSRuntime Worker 的配置

如何在 Fluid 中使用 JuiceFS，请参考文档[示例 - 如何在 Fluid 中使用 JuiceFS](juicefs_runtime.md)。本文讲述所有在 Fluid 中有关
JuiceFS worker 的相关配置。

## JuiceFS Worker

JuiceFSRuntime 会创建 JuiceFS 的缓存 worker，以 StatefulSet 的形式运行在 Kubernetes 集群中。其在 JuiceFS 商业版与社区版中作用不同，具体如下：

- 社区版：
  - JuiceFS worker pod 与 FUSE pod 共享本地缓存，以加速数据访问；
  - 应用被调度时，Fluid 会优先将其调度到有缓存 worker 的节点上。
- 商业版：
  - JuiceFS 的 worker pod 组成分布式缓存集群，为 FUSE pod 提供分布式缓存；
  - JuiceFS worker pod 与 FUSE pod 共享本地缓存，以加速数据访问；
  - 应用被调度时，Fluid 会优先将其调度到有缓存 worker 的节点上。

## Worker 基础配置

worker 的基础配置如下：

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: JuiceFSRuntime
metadata:
  name: jfsdemo
spec:
  replicas: 1
  juicefsVersion:
    image: registry.cn-hangzhou.aliyuncs.com/juicefs/juicefs-fuse
    imageTag: v1.0.0-4.8.0
    imagePullPolicy: IfNotPresent
  worker:
    options:
      "attr-cache"： "10"
    podMetadata:
      labels:
        juicefs: "worker"
      annotations:
        juicefs: "worker"
    networkMode: ContainerNetwork
    env:
      - name: "GOOGLE_CLOUD_PROJECT"
        value: "xxx"
    resources:
      limits:
        cpu: 2
        memory: 5Gi
      requests:
        cpu: 1
        memory: 1Gi
```

其中：

- `spec.replicas`：指定 worker 的副本数；
- `spec.juicefsVersion`：指定 JuiceFS 的镜像版本；
- `spec.worker.options`：指定 JuiceFS worker 的挂载参数，具体参数请参考 [JuiceFS 社区版文档](https://juicefs.com/docs/community/command_reference/#mount) 或 [JuiceFS 云服务文档](https://juicefs.com/docs/cloud/reference/commands_reference#mount)；
- `spec.worker.podMetadata`：指定 JuiceFS worker 的 pod 元数据，包括 labels 和 annotations；
- `spec.worker.networkMode`：指定 JuiceFS worker 的网络模式，目前支持 `HostNetwork` 和 `ContainerNetwork`，默认为 `HostNetwork`；
- `spec.worker.env`：指定 JuiceFS worker 的环境变量；
- `spec.worker.resources`：指定 JuiceFS worker 的资源限制。

## Worker 的 Volume 配置

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: JuiceFSRuntime
metadata:
  name: jfsdemo
spec:
  worker:
    networkMode: ContainerNetwork
    volumeMounts:
      - name: "juicefs-cache"
        mountPath: "/var/jfsCache"
  volumes:
    - name: "juicefs-cache"
      hostPath:
        path: "/var/jfsCache"
        type: DirectoryOrCreate
```

其中：
- `spec.worker.volumeMounts`：指定 JuiceFS worker 的 volumeMounts，可以指定多个 volumeMounts，其 name 与 `spec.volumes` 中的对应；
- `spec.volumes`：指定 volumes，可以指定多个 volumes。

## Worker 的调度策略配置

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: JuiceFSRuntime
metadata:
  name: jfsdemo
spec:
  worker:
    nodeSelector:
      fluid.io/dataset: jfsdemo
```

其中：

- `spec.worker.nodeSelector`：指定 JuiceFS worker 的 nodeSelector；

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: jfsdemo
spec:
  tolerations:
    - key: "fluid.io/dataset"
      operator: "Equal"
      value: "jfsdemo"
      effect: "NoSchedule"
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: "fluid.io/dataset"
              operator: "In"
              values:
                - "jfsdemo"
```

其中：

- `spec.tolerations`：指定 JuiceFS worker pod 的 tolerations；
- `spec.nodeAffinity`：指定 JuiceFS worker pod 的 nodeAffinity。
