# CacheRuntime TieredStore 配置示例

本文档展示了如何为 CacheRuntime 配置分层存储（TieredStore）。

## 概述

CacheRuntime 支持三种类型的存储介质：
1. **ProcessMemory**：使用进程内存作为缓存存储
2. **EmptyDir**：使用 Kubernetes EmptyDir Volume 作为缓存存储
3. **HostPath**：使用 Kubernetes HostPath Volume 作为缓存存储

系统会根据配置自动计算 `mediumType`，用于缓存策略优化：
- **MEM**：内存介质
- **HDD**：硬盘介质

## 示例 1：使用 ProcessMemory

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: CacheRuntime
metadata:
  name: my-cache
  namespace: default
spec:
  runtimeClassName: my-runtime-class
  worker:
    tieredStore:
      levels:
        - processMemory:
            quota: 8Gi
          high: "0.95"
          low: "0.7"
```

**说明：**
- `processMemory.quota` 设置内存配额，会自动添加到容器的 resource requests/limits
- `high` 和 `low` 是水位线配置，用于控制缓存驱逐

## 示例 2：使用 EmptyDir Volume（磁盘）

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: CacheRuntime
metadata:
  name: my-cache
  namespace: default
spec:
  runtimeClassName: my-runtime-class
  worker:
    tieredStore:
      levels:
        - emptyDir:
            quota: 100Gi
            medium: ""  # 使用磁盘空间，如果要使用内存设置为 "Memory"
          high: "0.95"
          low: "0.7"
```

**说明：**
- `emptyDir` 会创建一个临时目录，生命周期与 Pod 相同
- `quota` 会设置为 EmptyDir 的 sizeLimit
- 缓存数据会在 Pod 重启后丢失

## 示例 3：使用 HostPath Volume（单路径）

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: CacheRuntime
metadata:
  name: my-cache
  namespace: default
spec:
  runtimeClassName: my-runtime-class
  worker:
    tieredStore:
      levels:
        - hostPath:
            paths:
              - /data/cache
            quotas:
              - 100Gi
            type: DirectoryOrCreate
          high: "0.95"
          low: "0.7"
```

**说明：**
- `hostPath` 使用节点上的持久化目录
- `paths` 指定缓存路径列表
- `quotas` 指定每个路径的配额
- 缓存数据在 Pod 重启后仍然保留
- 需要确保节点上有相应的目录权限

## 示例 4：多层级存储

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: CacheRuntime
metadata:
  name: my-cache
  namespace: default
spec:
  runtimeClassName: my-runtime-class
  worker:
    tieredStore:
      levels:
        # 第一层：高速内存缓存（mediumType 自动设置为 MEM）
        - processMemory:
            quota: 4Gi
          high: "0.95"
          low: "0.7"
        
        # 第二层：大容量磁盘缓存（mediumType 自动设置为 HDD）
        - emptyDir:
            quota: 200Gi
          high: "0.90"
          low: "0.6"
```

**说明：**
- 可以配置多个层级，按优先级从高到低排列
- 每层可以有不同的介质类型、配额和水位线
- 系统会优先使用高层级的存储

## 示例 5：HostPath 多路径配置

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: CacheRuntime
metadata:
  name: my-cache
  namespace: default
spec:
  runtimeClassName: my-runtime-class
  worker:
    tieredStore:
      levels:
        - hostPath:
            paths:
              - /data/cache1
              - /data/cache2
            quotas:
              - 100Gi
              - 200Gi
            type: DirectoryOrCreate
          high: "0.95"
          low: "0.7"
```

**说明：**
- 只有 HostPath 类型支持多路径配置
- `paths` 和 `quotas` 数组长度必须一致
- 总配额 = 100Gi + 200Gi = 300Gi

## 实现细节

### mediumType 自动计算逻辑

`mediumType` 是由系统根据配置自动计算的，用户无需手动配置：

| 配置类型 | mediumType 值 | 说明 |
|---------|--------------|------|
| `processMemory` | MEM | 内存介质 |
| `emptyDir`（medium="Memory"） | MEM | 使用 tmpfs 内存 |
| `emptyDir`（medium=""） | HDD | 使用节点默认存储（磁盘） |
| `hostPath` | HDD | 硬盘介质 |

### ProcessMemory 处理

当使用 `processMemory` 时：
1. 将 `quota` 添加到容器的 `resources.requests.memory` 和 `resources.limits.memory`
2. 如果容器已有其他资源请求，会累加而不是覆盖
3. `mediumType` 自动设置为 MEM
4. 创建一个使用内存介质的 EmptyDir Volume
5. Volume 名称格式：`tiered-store-level-{N}-memory`
6. 容器内挂载路径：`/dev/shm`

### EmptyDir 处理

当使用 `emptyDir` 时：
1. 创建一个 EmptyDir Volume
2. Volume 名称格式：`tiered-store-level-{N}-index-{M}`（EmptyDir 的 M 始终为 0）
3. 容器内挂载路径：`/etc/fluid/mount/tiered-store/level-{N}-index-{M}-emptydir`
4. `quota` 设置为 EmptyDir 的 `sizeLimit`
5. 如果 `medium` 设置为 `"Memory"`，则使用 tmpfs，`mediumType` 为 MEM
6. 如果 `medium` 为空，则使用节点默认存储，`mediumType` 为 HDD

### HostPath 处理

当使用 `hostPath` 时：
1. 为每个路径创建一个独立的 HostPath Volume
2. Volume 名称格式：`tiered-store-level-{N}-index-{M}`
3. 容器内挂载路径：`/etc/fluid/mount/tiered-store/level-{N}-index-{M}-hostpath`
4. `paths` 列表指定节点上的目录路径
5. `quotas` 列表指定每个路径的配额（仅供参考，不会强制限制）
6. `type` 指定 hostPath 类型（如 `DirectoryOrCreate`）
7. `mediumType` 自动设置为 HDD

> **注意**：Volume 名称属于实现细节，可能会发生变化。用户不应依赖特定的 Volume 名称格式。

### 多路径配置的重要说明

根据最新的 API 定义，**只有 HostPath 类型支持多路径配置**：

| 介质类型 | 多路径支持 | Quota 生效方式 | 说明 |
|---------|-----------|---------------|------|
| **ProcessMemory** | ❌ 不支持 | 通过容器 memory limit | 单配额配置 |
| **EmptyDir** | ❌ 不支持 | 通过 `sizeLimit` | 单配额配置 |
| **HostPath** | ✅ 支持 | 仅供参考（需外部机制） | 可配置多个路径 |

**HostPath 多路径详细说明：**
- ✅ 每个路径创建独立的 HostPath volume
- ✅ `paths` 和 `quotas` 数组长度必须一致
- ⚠️ `quotas` 仅供参考，需要依赖外部机制（如 XFS quota、Cgroup 等）进行强制限制
- 💡 多路径的主要用途：扩展缓存容量或分布到多个磁盘

## 注意事项

1. **配额单位**：支持 Kubernetes 标准单位（Ki, Mi, Gi, Ti 等）
2. **水位线**：`high` 和 `low` 是 0-1 之间的小数，表示使用率的百分比
3. **介质选择**：每个层级只能选择一种介质类型（ProcessMemory、EmptyDir 或 HostPath）
4. **多路径支持**：只有 HostPath 类型支持多路径配置，ProcessMemory 和 EmptyDir 仅支持单路径
5. **HostPath 路径要求**：`paths` 和 `quotas` 数组长度必须一致
6. **mediumType**：由系统自动计算，无需用户配置，仅支持 MEM 和 HDD 两种值
7. **配额生效机制**：
   - ProcessMemory：通过容器 memory limit 强制限制
   - EmptyDir：通过 `sizeLimit` 强制限制
   - HostPath：`quotas` 仅供参考，需要依赖外部机制（如 XFS project quota、Cgroup 等）
