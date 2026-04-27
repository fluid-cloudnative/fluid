# 示例 - CacheRuntime 数据操作

## 前提说明

本文档是 [CacheRuntime 对接社区文档](../dev/generic_cache_runtime_integration.md) 的扩展，**假设您已经完成了基础的 CacheRuntime 集成**（包括定义 topology、配置组件等）。

本文档仅说明如何在已有的 `CacheRuntimeClass` 基础上**新增数据操作支持**，核心改动只有一个字段：`dataOperationSpecs`。

## 背景介绍

Fluid 的 CacheRuntime 提供了一种通用的缓存运行时抽象，允许用户通过 `CacheRuntimeClass` 定义不同缓存系统的实现细节。从 Fluid 最新版本开始，CacheRuntime 原生支持数据操作（Data Operations），包括数据预热（DataLoad）和数据处理（DataProcess）。

本文档将阐述如何为 CacheRuntime 配置和使用 DataLoad 功能。

> 注意： DataProcess Spec 中定义了 Pod 信息，是把 DataSet 当作 PVC 进行挂载使用，因此不需要缓存系统做改动即可使用 DataProcess 功能。

## 前提条件

在运行该示例之前，请参考[安装文档](../userguide/install.md)完成 Fluid 安装，并检查 Fluid 各组件正常运行：

```shell
$ kubectl get pod -n fluid-system
cacheruntime-controller-xxxxx              1/1     Running   0          8h
csi-nodeplugin-fluid-fwgjh                  2/2     Running   0          8h
dataset-controller-5b7848dbbb-n44dj         1/1     Running   0          8h
```

确保你的集群中已安装了支持数据操作的 CacheRuntime 控制器。

## 核心概念

### 相对于基础集成的唯一改动

与基础的 CacheRuntime 集成相比，支持数据操作**仅需在 CacheRuntimeClass 中添加一个顶层字段**：

- **只需添加** `dataOperationSpecs` 字段
- **无需修改** 任何现有字段（topology、fileSystemType、extraResources 等）
- **向后兼容**：未配置此字段的 CacheRuntimeClass 仍可正常使用基础缓存功能

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: CacheRuntimeClass
metadata:
  name: curvine-demo
fileSystemType: curvinefs

# 【新增】仅此一个字段，其他配置（topology等）完全保持不变
dataOperationSpecs:
  - name: DataLoad
    command: ["/bin/bash", "-c"]
    args: ["..."]

# 【原有配置】topology、extraResources 等字段无需任何修改
topology:
  master:
    # ... 与基础集成完全一致
  worker:
    # ... 与基础集成完全一致
  client:
    # ... 与基础集成完全一致
```

### dataOperationSpecs 字段详解

`dataOperationSpecs` 是一个数组，每个元素定义一种数据操作的执行规范。

#### 字段结构

```yaml
dataOperationSpecs:
  - name: <操作类型>
    command: [<命令>, <参数>]
    args: [<脚本或参数>]
    image: <可选：专用镜像>
```

#### 字段说明

| 字段名 | 类型 | 必填 | 说明                                                                                                                        |
|--------|------|----|---------------------------------------------------------------------------------------------------------------------------|
| `name` | string |  是 | 操作类型标识符，当前支持的值：<br>• `DataLoad`：数据预热操作<br>•  `DataMigrate`：数据迁移操作（暂未支持)<br>• `DataBackup`：数据备份操作（暂未支持) |
| `command` | []string |  是 | 容器中执行的命令（entrypoint），通常设置为 `["/bin/bash", "-c"]` 以支持脚本执行                                                                  |
| `args` | []string |  是 | 命令的参数，通常包含完整的执行脚本。脚本中可使用 Fluid 注入的环境变量（见下文）                                                                               |
| `image` | string |  否 | 操作使用的容器镜像。<br>• **如果不指定**：默认使用 `CacheRuntimeClass` 中 `worker` 组件的镜像<br>• **如果指定**：使用自定义的专用镜像（适用于需要特殊工具的场景）                |

### 可用环境变量

在数据操作执行过程中，Fluid 会自动向容器中注入以下环境变量：

#### DataLoad 专属环境变量

| 环境变量名 | 说明                             | 示例值 |
|-----------|--------------------------------|--------|
| `FLUID_DATALOAD_METADATA` | 是否加载元数据                        | `"true"` 或 `"false"` |
| `FLUID_DATALOAD_DATA_PATH` | 需要加载的数据路径（多个路径用冒号分隔）           | `/spark/spark-3.0.1:/spark/spark-2.4.7` |
| `FLUID_DATALOAD_PATH_REPLICAS` | 每个路径的副本数（用冒号分隔，与 DATA_PATH 一一对应） | `1:2` |

底层的缓存系统根据上面的环境变量，编写数据预热的脚本并打包到镜像中，用户在定义 DataLoad 操作时，即可通过`command` 和 `args` 字段指定脚本。
