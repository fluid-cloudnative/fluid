# 示例 - Fuse Sidecar 挂载模式

## 背景介绍

Fluid 为第三方存储在 Kubernetes 集群内提供 PVC 的挂载能力，支持两种挂载方式：**CSI 模式**和 **Sidecar 模式**。

### CSI 模式（Fuse Pod Mounting）

CSI 模式是 Fluid 默认的挂载方式，通过标准的 Kubernetes CSI 机制来完成数据挂载：

1. 当业务 Pod 挂载 Fluid 的 PVC 时，Fluid 会识别出业务 Pod 所在的节点
2. 通过为节点添加对应的 label，触发 Fuse DaemonSet 的调度逻辑
3. Fuse DaemonSet 会在该节点上扩容出一个 Fuse Pod
4. Fuse Pod 负责将 PVC 对应的数据源挂载到节点上的特定路径
5. CSI Driver 通过 bind mount 的方式，将节点上的挂载点传播给 kubelet 指定的 volume 挂载路径
6. 最终业务 Pod 通过 CSI 挂载点访问数据

**CSI 模式的特点：**
- 符合 Kubernetes CSI 标准，兼容性好
- Fuse Pod 以 DaemonSet 形式部署，独立于业务 Pod
- 适用于常规的 Kubernetes 工作负载场景

### Sidecar 模式（Fuse Sidecar Mounting）

Sidecar 模式通过 Webhook 机制将 Fuse 容器注入到业务 Pod 中，提供了更灵活的挂载方式：

1. 用户在业务 Pod 上添加特定的 label：`serverless.fluid.io/inject: "true"`
2. Fluid Webhook 拦截 Pod 创建请求，识别需要注入 Sidecar 的 Pod
3. Webhook 根据 Pod 挂载的 PVC 找到对应的 Fuse DaemonSet 配置
4. 将原本的 PVC 挂载方式 mutate 成一个 Fuse Sidecar 容器
5. Sidecar 容器与业务容器共享一个 HostPath 或 EmptyDir Volume
6. Sidecar 容器内执行数据源的挂载动作，将挂载点通过共享 Volume 传播至业务容器
7. 业务容器可以直接访问 Sidecar 挂载的数据

**Sidecar 模式的特点：**
- 无需依赖 CSI Driver，适合 Serverless 环境（如 Virtual Kubelet、Knative、ECI 等）
- Fuse 容器与业务容器生命周期一致，每个 Pod 独立 Fuse 容器，资源完全隔离
- 资源隔离性强，一个 Pod 的挂载点异常不会影响节点上其他 Pod
- 支持跨有节点环境和 Serverless 环境的兼容性调度
- 配合 Application Controller，可以在业务容器退出后自动清理 Fuse Sidecar
- **Trade-off**：每个 Pod 独立 Fuse 客户端会增加到缓存系统/后端数据源的连接数，需要权衡隔离性与连接数开销

### 两种模式的对比

| 特性 | CSI 模式 | Sidecar 模式 |
|------|---------|-------------|
| 挂载机制 | CSI + DaemonSet | Webhook + Sidecar Container |
| 适用场景 | 常规 Kubernetes 工作负载 | Serverless 环境、需要资源隔离的场景、跨环境兼容调度 |
| Fuse 生命周期 | 独立于业务 Pod | 与业务 Pod 一致 |
| 资源隔离 | 共享 DaemonSet Pod，节点级隔离 | 每个业务 Pod 独立 Sidecar，Pod 级隔离 |
| 故障影响 | 一个 Fuse Pod 异常影响节点所有业务 Pod | 一个 Fuse 容器异常只影响当前 Pod |
| 客户端连接数 | 节点级共享，连接数较少 | 每个 Pod 独立客户端，连接数随 Pod 数量增加 |
| 启用方式 | 默认启用 | 需添加 label |
| CSI 依赖 | 需要 CSI Driver | 无需 CSI Driver |

## 前提条件

在运行该示例之前，请参考[安装文档](../userguide/install.md)完成安装，并检查 Fluid 各组件正常运行：

```shell
$ kubectl get pod -n fluid-system
NAME                                  READY   STATUS    RESTARTS   AGE
alluxioruntime-controller-xxx         1/1     Running   0          5m
dataset-controller-xxx                1/1     Running   0          5m
fluid-webhook-xxx                     1/1     Running   0          5m
```

通常来说，你会看到一个名为 `dataset-controller` 的 Pod、一个或多个 Runtime Controller（如 `alluxioruntime-controller`）的 Pod，以及一个名为 `fluid-webhook` 的 Pod 正在运行。

> **注意**：Sidecar 模式依赖 `fluid-webhook` 组件进行 Pod 的动态注入，请确保 webhook 正常运行。

## 运行示例

### 1. 创建 Dataset 和 Runtime

首先创建一个 Dataset 和对应的 Runtime。这里以 AlluxioRuntime 为例：

```shell
$ cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: demo-data
spec:
  mounts:
    - mountPoint: https://mirrors.bit.edu.cn/apache/spark/
      name: spark
---
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: demo-data
spec:
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 2Gi
        high: "0.95"
        low: "0.7"
EOF

$ kubectl create -f dataset.yaml
```

等待 Dataset 和 Runtime 准备就绪：

```shell
$ kubectl get dataset demo-data
NAME        UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
demo-data   [Calculating]    0.00B    2.00GiB          0.0%                Bound   1m

$ kubectl get alluxioruntime demo-data
NAME        MASTER PHASE   WORKER PHASE   FUSE PHASE   AGE
demo-data   Ready          Ready          Ready        1m
```

### 2. 创建使用 Sidecar 模式的应用 Pod

创建一个 Pod，并通过 label `serverless.fluid.io/inject: "true"` 启用 Sidecar 模式：

```shell
$ cat<<EOF >app-sidecar.yaml
apiVersion: v1
kind: Pod
metadata:
  name: demo-app
  labels:
    serverless.fluid.io/inject: "true"
spec:
  containers:
    - name: demo
      image: nginx
      command: ["sh", "-c"]
      args:
        - |
          echo "Listing files in /data:"
          ls -lh /data/spark
          echo "Sleeping..."
          sleep 3600
      volumeMounts:
        - mountPath: /data
          name: data-vol
  volumes:
    - name: data-vol
      persistentVolumeClaim:
        claimName: demo-data
EOF

$ kubectl create -f app-sidecar.yaml
```

### 3. 验证 Sidecar 注入

查看 Pod 状态，确认 Fuse Sidecar 已经被注入：

```shell
$ kubectl get po demo-app
NAME       READY   STATUS    RESTARTS   AGE
demo-app   2/2     Running   0          30s
```

可以看到 Pod 显示 `2/2`，说明除了业务容器外，还有一个 Fuse Sidecar 容器。

查看 Pod 的详细信息：

```shell
$ kubectl get po demo-app -o jsonpath='{.spec.containers[*].name}'
demo fluid-fuse
```

可以看到 Pod 中包含两个容器：业务容器 `demo` 和 Fuse Sidecar 容器 `fluid-fuse`。

查看 Fuse Sidecar 的镜像：

```shell
$ kubectl get po demo-app -o jsonpath='{.spec.containers[?(@.name=="fluid-fuse")].image}'
fluidcloudnative/alluxio-fuse:xxx
```

### 4. 验证数据访问

在业务容器中验证数据是否可以正常访问：

```shell
$ kubectl logs demo-app -c demo
Listing files in /data:
total 0
drwxr-xr-x 1 root root 0 Jan  1 00:00 spark
Sleeping...
```

进入容器查看挂载的数据：

```shell
$ kubectl exec demo-app -c demo -it -- ls /data/spark
spark-3.3.0  spark-3.3.1  spark-3.3.2
```

### 5. 对比 CSI 模式

为了对比，我们创建一个使用 CSI 模式的 Pod（不添加 Sidecar 注入的 label）：

```shell
$ cat<<EOF >app-csi.yaml
apiVersion: v1
kind: Pod
metadata:
  name: demo-app-csi
spec:
  containers:
    - name: demo
      image: nginx
      command: ["sh", "-c", "sleep 3600"]
      volumeMounts:
        - mountPath: /data
          name: data-vol
  volumes:
    - name: data-vol
      persistentVolumeClaim:
        claimName: demo-data
EOF

$ kubectl create -f app-csi.yaml
```

查看 Pod 状态：

```shell
$ kubectl get po demo-app-csi
NAME           READY   STATUS    RESTARTS   AGE
demo-app-csi   1/1     Running   0          20s
```

CSI 模式下，Pod 只有 `1/1` 个容器，没有 Sidecar 容器被注入。

此时可以看到 Fuse DaemonSet 在对应节点上启动了 Fuse Pod：

```shell
$ kubectl get po -l app=alluxio,role=alluxio-fuse,fluid.io/dataset=demo-data
NAME                   READY   STATUS    RESTARTS   AGE
demo-data-fuse-xxxxx   1/1     Running   0          25s
```

## 进阶配置

### 在有节点环境启用 Sidecar 模式

**背景说明：**

Sidecar 模式通过共享 HostPath 来实现挂载点传播。在有节点环境中，如果多个挂载相同 Dataset 的 Pod 使用相同的 HostPath，会导致多个业务 Pod 中的 Fuse Sidecar 对 HostPath 上挂载点的操作产生冲突。

为了解决这个问题，Fluid 提供了 `random-suffix` 模式，为每个业务 Pod 生成独立的 HostPath 路径。

**配置方法：**

为业务 Pod 添加 annotation：`default.fuse-sidecar.fluid.io/host-mount-path-mode: "random-suffix"`

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: demo-app-node
  labels:
    serverless.fluid.io/inject: "true"
  annotations:
    default.fuse-sidecar.fluid.io/host-mount-path-mode: "random-suffix"  # 为有节点环境启用独立 HostPath
spec:
  containers:
    - name: demo
      image: nginx
      volumeMounts:
        - mountPath: /data
          name: data-vol
  volumes:
    - name: data-vol
      persistentVolumeClaim:
        claimName: demo-data
```

设置此 annotation 后，Fluid Webhook 在为开启了 Sidecar 模式的业务 Pod 注入用于挂载点传播的 HostPath Volume 时，会根据数据源信息和时间戳生成一个随机的 HostPath 路径，避免在有节点环境中的冲突问题。

**兼容性调度：**

此配置在 Serverless 环境中同样生效，通过这个配置，用户可以使用一套部署 YAML，在有节点环境和 Serverless 环境之间实现兼容性调度。

**注意事项：**

由于 Kubernetes 的原生机制，HostPath 创建后在 Pod 删除后不会自动清理，因此每次挂载都会有一个 HostPath 残留。建议定期运行清理脚本来清理残留的 HostPath。

> **清理脚本**：Fluid 后续会提供安全的清理脚本，用于清理主机上残留的HostPath。

### 禁用 PostStart 阻塞

**背景说明：**

在 Sidecar 模式中，Fluid 会为 Fuse Sidecar 容器注入一个 PostStart 的 Lifecycle Hook，用于检查挂载点是否就绪。这个 PostStart Hook 会阻塞业务容器的启动，确保业务容器启动时数据已经准备好。

在某些场景下，如果希望 Sidecar 容器的挂载动作不阻塞业务容器的启动（例如业务容器有自己的重试机制，或者希望业务容器和 Fuse 容器并行启动），可以禁用 PostStart 注入。

**配置方法：**

为业务 Pod 添加 label：`fuse.sidecar.poststart.fluid.io/inject: "false"`

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: demo-app-no-poststart
  labels:
    serverless.fluid.io/inject: "true"
    fuse.sidecar.poststart.fluid.io/inject: "false"  # 禁用 PostStart 阻塞
spec:
  containers:
    - name: demo
      image: nginx
      command: ["sh", "-c"]
      args:
        - |
          # 业务容器可能需要自己检查挂载点就绪
          while [ ! -d /data/spark ]; do
            echo "Waiting for mount point..."
            sleep 1
          done
          echo "Mount point ready, starting application..."
          sleep 3600
      volumeMounts:
        - mountPath: /data
          name: data-vol
  volumes:
    - name: data-vol
      persistentVolumeClaim:
        claimName: demo-data
```

设置此 label 后，Fuse Sidecar 容器将不会注入 PostStart Hook，业务容器可以立即启动，无需等待挂载点就绪。业务容器需要自己处理挂载点可能尚未就绪的情况。

### 跨 Namespace 访问

Sidecar 模式同样支持跨 Namespace 访问 Dataset。具体使用方法请参考[跨 Namespace 共享数据（Sidecar 模式）](dataset_across_namespace_with_sidecar.md)。

## 适用场景

Sidecar 模式特别适用于以下场景：

### 1. Serverless 容器服务

- **阿里云 ECI（Elastic Container Instance）**
- **AWS Fargate**
- **Azure Container Instances**
- **Virtual Kubelet** 环境

这些环境通常不支持或限制 CSI Driver 的使用，Sidecar 模式可以无缝集成。

### 2. Knative Serving

在 Knative 环境中运行的 Serverless 应用，可以通过 Sidecar 模式实现数据加速。详见[如何在 Knative 环境运行](knative.md)。

### 3. 临时任务和批处理作业

- **Kubernetes Job**
- **CronJob**
- **Argo Workflows**
- **Spark on Kubernetes**

Sidecar 模式下，Fuse 容器随业务容器一起启动和退出，避免了 DaemonSet Fuse Pod 的残留问题。

### 4. 需要强资源隔离的场景

当需要为不同的业务 Pod 提供更强的资源隔离时，Sidecar 模式可以为每个 Pod 分配独立的 Fuse 容器，具有以下优势：

- **资源隔离**：每个 Pod 有独立的 Fuse 进程，内存、CPU 资源互不影响
- **故障隔离**：一个 Pod 的 Fuse 容器异常崩溃，不会影响节点上其他 Pod 的数据访问
- **安全隔离**：多租户场景下，不同租户的 Pod 不共享 Fuse 进程

### 5. 跨环境兼容性调度

对于需要在有节点环境和 Serverless 环境之间灵活调度的应用，Sidecar 模式配合 `random-suffix` 配置，可以使用统一的 YAML 部署文件，在不同环境中无缝切换：

- **有节点环境**：通过 `random-suffix` 模式避免 HostPath 冲突
- **Serverless 环境**：无 CSI 依赖，直接使用 Sidecar 注入
- **统一配置**：一套 YAML 配置文件，无需针对不同环境进行修改

## 注意事项

### 1. Webhook 依赖

Sidecar 模式依赖 Fluid Webhook 组件进行 Pod 的动态注入，请确保：

- `fluid-webhook` Deployment 正常运行
- Webhook 的 MutatingWebhookConfiguration 已正确配置
- API Server 能够正常访问 Webhook 服务

### 2. 资源消耗与连接数

Sidecar 模式下，每个业务 Pod 都会注入一个 Fuse Sidecar 容器，会增加：

- **内存消耗**：每个 Sidecar 容器需要额外的内存
- **CPU 消耗**：Fuse 进程的 CPU 开销
- **镜像拉取**：需要拉取 Fuse 容器镜像
- **连接数增加**：每个 Pod 独立的 Fuse 客户端会建立独立的连接到缓存系统（如 Alluxio Worker）和后端数据源

**Trade-off 考虑：**

- **CSI 模式**：节点上所有 Pod 共享一个 Fuse DaemonSet Pod，对缓存系统和数据源的连接数较少，适合大规模 Pod 部署
- **Sidecar 模式**：每个 Pod 有独立的 Fuse 客户端，连接数随 Pod 数量线性增长，可能对缓存系统和数据源造成连接压力

**建议：**

- 在大规模部署时，需要评估缓存系统和数据源的连接数承载能力
- 如果隔离性需求不强，优先考虑 CSI 模式以减少连接数开销
- 如果必须使用 Sidecar 模式，建议配置缓存系统和数据源的连接池大小，并监控连接数指标

### 3. 存储卷共享机制

Sidecar 模式下，业务容器和 Fuse 容器通过共享 Volume 传递挂载点，可能使用：

- **EmptyDir**：默认方式，简单但仅限于 Pod 内共享
- **HostPath**：某些场景下使用，需要注意节点权限

不同的实现方式可能影响性能和兼容性。

### 4. ConfigMap 共享

在某些场景下（如跨 Namespace 访问），Fluid 可能会在目标 Namespace 中创建 ConfigMap 用于配置传递。如果目标 Namespace 中已存在同名 ConfigMap，可能导致冲突。

详见[跨 Namespace 共享数据（Sidecar 模式）](dataset_across_namespace_with_sidecar.md)的已知问题部分。

## 总结

Sidecar 模式是 Fluid 为 Serverless 和云原生场景提供的灵活挂载方案：

- ✅ **无需 CSI 依赖**，适用于 Virtual Kubelet、ECI 等 Serverless 环境
- ✅ **生命周期一致**，与业务容器同生共死，避免资源残留
- ✅ **Pod 级资源隔离**，每个 Pod 独立 Fuse 容器，故障互不影响
- ✅ **跨环境兼容**，支持有节点和 Serverless 环境间的无缝调度
- ✅ **配置灵活**，支持 random-suffix、PostStart 控制等高级特性

通过简单的 label 配置（`serverless.fluid.io/inject: "true"`），即可启用 Sidecar 模式，为 Serverless 应用提供高效的数据访问能力。

## 相关文档

- [如何在 Knative 环境运行](knative.md)
- [如何保障 Serverless 任务顺利完成](application_controller.md)
- [跨 Namespace 共享数据（Sidecar 模式）](dataset_across_namespace_with_sidecar.md)
- [跨 Namespace 共享数据（CSI 模式）](dataset_across_namespace_with_csi.md)
