# CacheRuntime Spec 字段更新能力说明

## 1. 概述

本文档说明 CacheRuntime 的 Master 和 Worker 组件（基于 **AdvancedStatefulSet**）支持的 spec 字段更新能力。

当前版本支持以下三个字段的原地更新：
- **容器镜像** (`runtimeVersion`)
- **资源限制** (`resources`)
- **副本数** (`replicas`)

其他字段的修改**不会同步到 AdvancedStatefulSet**，需要重新部署才能生效。

---

## 2. 支持的组件

| 组件 | 工作负载类型 | 字段更新支持 |
|------|-------------|------------|
| **Master** | AdvancedStatefulSet | ✅ 支持 `runtimeVersion`、`resources` 和 `replicas` |
| **Worker** | AdvancedStatefulSet | ✅ 支持 `runtimeVersion`、`resources` 和 `replicas` |
| **Client** | DaemonSet | ❌ 不支持（任何变更都需重新部署） |

**说明**：
- Master 和 Worker 的 `runtimeVersion`、`resources` 和 `replicas` 字段修改会自动同步到 AdvancedStatefulSet
- Client 组件使用 DaemonSet，不支持动态更新

---

## 3. 支持更新的字段

### 3.1 容器镜像 (`runtimeVersion`)

**字段路径**: `spec.{master,worker}.runtimeVersion`

**支持字段**:
- `image`: 镜像名称
- `imageTag`: 镜像标签

**示例**:
```yaml
spec:
  worker:
    runtimeVersion:
      image: fluid-cache
      imageTag: v1.1.0
```

**限制**:
- ⚠️ Cgroupv1 环境中，不能与 `resources` 字段同时更新（需分步操作，见下方资源字段说明）

---

### 3.2 资源限制 (`resources`)

**字段路径**: `spec.{master,worker}.resources`

**支持字段**:
- `requests.cpu`: CPU 请求
- `requests.memory`: 内存请求
- `limits.cpu`: CPU 限制
- `limits.memory`: 内存限制

**示例**:
```yaml
spec:
  worker:
    resources:
      requests:
        cpu: "4"
        memory: 8Gi
      limits:
        cpu: "8"
        memory: 16Gi
```

**取值优先级**：

每次 reconcile 时，`resources` 按以下顺序取值：

1. CacheRuntime 上设置的值（只要声明了 `requests` 或 `limits`）；
2. 否则取 CacheRuntimeClass 模板中声明的值；
3. 两者都未声明时不做同步，工作负载保持当前的资源配置。

**限制**：
- ⚠️ 不能超过节点可用资源
- ⚠️ 当 CacheRuntimeClass 模板声明了 `resources` 时，从 CacheRuntime 中删除 `resources` **不会**
  让组件变为不受限，而是回退到模板中的值。如需放宽限制，请显式设置目标值，而不是删除该字段。
- ⚠️ **Kubernetes 版本要求**：需要 K8s >= 1.27 且启用 `InPlacePodVerticalScaling` Feature Gate
  ```bash
  # 检查 Feature Gate 是否启用
  kubectl get nodes -o jsonpath='{.items[0].status.config.kubeletConfig.featureGates.InPlacePodVerticalScaling}'
  ```
- ⚠️ **Cgroupv1 环境限制**：不能与 `runtimeVersion` 字段同时更新
  - 原因：Cgroupv1 不支持在同一操作中同时更新容器镜像和资源限制
  - 解决方案：分步操作，先更新资源，等待完成后再更新镜像
  ```bash
  # 第一步：更新资源
  kubectl patch cacheruntime my-cache --type='merge' -p '{"spec":{"worker":{"resources":{"requests":{"cpu":"4","memory":"8Gi"}}}}}'
  kubectl wait pod -l fluid.io/cache-runtime-name=my-cache --for=condition=InPlaceUpdateReady --timeout=300s
  
  # 第二步：更新镜像
  kubectl patch cacheruntime my-cache --type='merge' -p '{"spec":{"worker":{"runtimeVersion":{"imageTag":"v1.1.0"}}}}'
  ```
  详见 [Kubernetes Issue #127356](https://github.com/kubernetes/kubernetes/issues/127356)。

---

### 3.3 副本数 (`replicas`)

**字段路径**: `spec.{master,worker}.replicas`

**示例**:
```yaml
spec:
  worker:
    replicas: 3
```

```bash
kubectl patch cacheruntime my-cache --type='merge' -p '{"spec":{"worker":{"replicas":3}}}'
```

修改后同步到 AdvancedStatefulSet 的 `spec.replicas`，由其完成扩缩容。

**限制**：
- ⚠️ **缩容会丢失缓存数据**：被缩掉的 Worker Pod 直接删除，其上已缓存的数据不会迁移到其余 Worker，需要重新从底层存储加载。扩容不影响已有缓存。
- ⚠️ 扩缩容不会写入 RuntimeCondition，也不会产生 Kubernetes Event。可以通过 `status.{master,worker}.{readyReplicas,desiredReplicas}` 或 controller 日志观察进度：
  ```bash
  # 查看当前副本状态
  kubectl get cacheruntime my-cache -o jsonpath='{.status.worker.readyReplicas}/{.status.worker.desiredReplicas}'

  # 或从 controller 日志观察
  kubectl -n fluid-system logs deploy/cacheruntime-controller | grep "replicas changed"
  ```
- ⚠️ 副本数超过可调度节点数时，多余的 Worker Pod 会一直处于 Pending。

## 4. 不支持更新的字段

**第 3 节未列出的字段一律不支持原地更新**，修改后不会同步到 AdvancedStatefulSet。

**说明**：
- 这类字段修改后，需要重新部署 CacheRuntime 才能生效
- 系统不会自动检测或同步这类变更

---

## 5. 总结

| 字段 | 是否支持更新 | 更新方式 |
|------|------------|---------|
| `runtimeVersion` | ✅ 支持 | 自动同步到 AdvancedStatefulSet |
| `resources` | ✅ 支持 | 自动同步到 AdvancedStatefulSet（需 K8s >= 1.27） |
| `replicas` | ✅ 支持 | 自动同步到 AdvancedStatefulSet（缩容会丢失该 Worker 上的缓存） |
| 其他所有字段 | ❌ 不支持 | 不会同步，需重新部署 |

**关键要点**：
1. 当前版本支持 `runtimeVersion`、`resources` 和 `replicas` 三个字段的动态更新
2. Cgroupv1 环境需分步更新 `runtimeVersion` 和 `resources`，`replicas` 不受此限制
3. 其他字段的修改不会生效，必须重新部署
