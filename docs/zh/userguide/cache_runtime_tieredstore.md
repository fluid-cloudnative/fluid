# CacheRuntime TieredStore 配置示例

本文档展示了如何为 CacheRuntime 配置分层存储（TieredStore）。

## 概述

CacheRuntime 支持两种类型的存储介质：
1. **ProcessMemory**：使用进程内存作为缓存存储
2. **Volume**：使用 Kubernetes Volume 作为缓存存储（支持 hostPath、emptyDir、ephemeral）

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
        - medium:
            processMemory: {}
          path:
            - /dev/shm
          quota:
            - 8Gi
          high: "0.95"
          low: "0.7"
```

**说明：**
- `processMemory: {}` 表示使用进程内存
- `path` 指定缓存路径（对于内存，通常是 /dev/shm）
- `quota` 设置内存配额，会自动添加到容器的 resource requests/limits
- `high` 和 `low` 是水位线配置，用于控制缓存驱逐

## 示例 2：使用 EmptyDir Volume

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
        - medium:
            volume:
              emptyDir:
                medium: ""  # 使用磁盘空间，如果要使用内存设置为 "Memory"
          path:
            - /mnt/cache
          quota:
            - 100Gi
          high: "0.95"
          low: "0.7"
```

**说明：**
- `emptyDir` 会创建一个临时目录，生命周期与 Pod 相同
- `quota` 会设置为 EmptyDir 的 sizeLimit
- 缓存数据会在 Pod 重启后丢失

## 示例 3：使用 HostPath Volume

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
        - medium:
            volume:
              hostPath:
                path: /data/cache
                type: DirectoryOrCreate
          path:
            - /mnt/cache
          high: "0.95"
          low: "0.7"
```

**说明：**
- `hostPath` 使用节点上的持久化目录
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
        # 第一层：高速内存缓存
        - medium:
            processMemory: {}
          path:
            - /dev/shm
          quota:
            - 4Gi
          high: "0.95"
          low: "0.7"
        
        # 第二层：大容量磁盘缓存
        - medium:
            volume:
              emptyDir: {}
          path:
            - /mnt/ssd
          quota:
            - 200Gi
          high: "0.90"
          low: "0.6"
```

**说明：**
- 可以配置多个层级，按优先级从高到低排列
- 每层可以有不同的介质类型、配额和水位线
- 系统会优先使用高层级的存储

## 示例 5：多路径配置

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
        - medium:
            processMemory: {}
          path:
            - /dev/shm/cache1
            - /dev/shm/cache2
          quota:
            - 2Gi
            - 3Gi
          high: "0.95"
          low: "0.7"
```

**说明：**
- 可以为同一层级指定多个路径
- 每个路径可以有独立的配额
- 总内存配额 = 2Gi + 3Gi = 5Gi

## 示例 6：使用 Ephemeral Volume

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
        - medium:
            volume:
              ephemeral:
                volumeClaimTemplate:
                  spec:
                    accessModes: ["ReadWriteOnce"]
                    storageClassName: "fast-ssd"
          path:
            - /mnt/cache
          quota:
            - 100Gi
          high: "0.95"
          low: "0.7"
```

**说明：**
- `ephemeral` 会创建临时的 PVC，生命周期与 Pod 相同
- `quota` 会设置为 PVC 的 storage request
- 可以使用不同的 storageClassName 来获得不同的存储性能
- 缓存数据会在 Pod 删除后丢失，但比 EmptyDir 更灵活

## 实现细节

### ProcessMemory 处理

当使用 `processMemory` 时：
1. 系统会计算所有路径的配额总和
2. 将总配额添加到容器的 `resources.requests.memory` 和 `resources.limits.memory`
3. 如果容器已有其他资源请求，会累加而不是覆盖

### Volume 处理

当使用 `volume` 时：
1. 为每个路径创建一个独立的 Volume
2. Volume 名称格式：`tieredstore-level{N}-path{M}`
3. 根据 Volume 类型（hostPath/emptyDir/ephemeral）创建对应的 VolumeSource
4. 为容器添加对应的 VolumeMount
5. 配额处理方式因类型而异：
   - **EmptyDir**：quota 设置为 `sizeLimit`
   - **Ephemeral**：quota 设置为 `volumeClaimTemplate.spec.resources.requests.storage`
   - **HostPath**：quota **仅供参考**，不会强制限制（需要依赖外部机制如 XFS quota）

### 多路径配置的重要说明

**不同 Volume 类型对多路径的支持程度：**

| Volume 类型 | 多路径合理性 | Quota 是否生效 | 物理隔离 | 建议 |
|------------|------------|--------------|---------|------|
| **EmptyDir** | ✅ 合理 | ✅ 是（通过 sizeLimit） | ✅ 独立 tmpfs/disk | 推荐使用 |
| **Ephemeral** | ✅ 合理 | ✅ 是（通过 PVC storage request） | ✅ 独立 PVC | 推荐使用 |
| **HostPath** | ❌ 不推荐 | ❌ 否（共享节点存储） | ❌ 无隔离 | 建议使用单路径 |
| **ProcessMemory** | ⚠️ 有限 | ⚠️ 部分（累加到容器 limit） | ❌ 无隔离 | 谨慎使用多路径 |

**详细说明：**

1. **EmptyDir + 多路径**：
   - ✅ 每个路径创建独立的 EmptyDir volume
   - ✅ 每个 volume 有独立的 `sizeLimit`
   - ✅ 真正的物理隔离和配额控制
   - 示例：可以使用多个磁盘分区或 tmpfs

2. **Ephemeral + 多路径**：
   - ✅ 每个路径创建独立的 PVC
   - ✅ 每个 PVC 可以有独立的 storage request（通过 quota 设置）
   - ⚠️ **所有 PVC 共享相同的 volumeClaimTemplate 配置**（包括 storageClassName、accessModes 等）
   - ⚠️ **无法在同一个 Level 内混合使用不同的 storage class**
   - 💡 多路径的主要用途：**扩展同一类型存储的总容量**
   - 示例：使用多个 SSD PVC，每个 100Gi，总容量 200Gi
   - 📌 如需使用不同类型的存储（SSD + HDD），应配置**不同的 Level**

3. **HostPath + 多路径**：
   - ❌ 多个路径可能指向同一节点的存储空间
   - ❌ quota 无法直接生效，只是配置元数据
   - ❌ 没有真正的配额限制和物理隔离
   - ⚠️ 如果需要配额控制，请使用操作系统级别的 quota 机制
   - **建议**：HostPath 场景下只配置单个路径

4. **ProcessMemory + 多路径**：
   - ⚠️ 所有路径的 quota 会累加到容器的 memory requests/limits
   - ⚠️ 路径之间没有物理隔离（都在 /dev/shm 下）
   - ⚠️ 应用程序可以随意使用总内存，不受单个路径 quota 限制
   - **建议**：除非有特殊的逻辑分区需求，否则使用单路径即可

## 注意事项

1. **配额单位**：支持 Kubernetes 标准单位（Ki, Mi, Gi, Ti 等）
2. **水位线**：`high` 和 `low` 是 0-1 之间的小数，表示使用率的百分比
3. **路径要求**：`path` 数组和 `quota` 数组长度必须一致
4. **介质选择**：每个层级只能选择一种介质类型（ProcessMemory 或 Volume）
5. **Volume 限制**：每个层级的 Volume 只能选择一种类型（hostPath、emptyDir 或 ephemeral）
6. **多路径建议**：
   - EmptyDir 和 Ephemeral 类型推荐使用多路径，可以实现真正的配额控制和物理隔离
   - HostPath 类型建议使用单路径，因为 quota 无法生效
   - ProcessMemory 类型谨慎使用多路径，只有逻辑分区意义，无物理隔离
7. **配额生效机制**：
   - EmptyDir：通过 `sizeLimit` 强制限制
   - Ephemeral：通过 PVC 的 storage request 限制
   - HostPath：需要依赖外部机制（如 XFS project quota、Cgroup 等）
   - ProcessMemory：通过容器 memory limit 限制总量，但路径间无隔离
