# CacheRuntime Spec 字段更新能力说明

## 1. 概述

本文档说明 CacheRuntime 的 Master 和 Worker 组件（基于 **AdvancedStatefulSet**）支持的 spec 字段更新能力。

当前版本仅支持以下两个字段的原地更新：
- **容器镜像** (`runtimeVersion`)
- **资源限制** (`resources`)

其他字段的修改**不会同步到 AdvancedStatefulSet**，需要重新部署才能生效。

---

## 2. 支持的组件

| 组件 | 工作负载类型 | 字段更新支持 |
|------|-------------|------------|
| **Master** | AdvancedStatefulSet | ✅ 支持 `runtimeVersion` 和 `resources` |
| **Worker** | AdvancedStatefulSet | ✅ 支持 `runtimeVersion` 和 `resources` |
| **Client** | DaemonSet | ❌ 不支持（任何变更都需重新部署） |

**说明**：
- Master 和 Worker 的 `runtimeVersion` 和 `resources` 字段修改会自动同步到 AdvancedStatefulSet
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
- `claims`: 动态资源分配（DRA）的资源声明，只能引用模板中声明的 Pod 级 `resourceClaims`，见下方限制

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

**取值方式**：

每次 reconcile 时，`resources` 按以下方式取值：

1. 以 CacheRuntimeClass 模板中声明的 `resources` 为基线；
2. 将 CacheRuntime 上设置的 `resources` 按 key 叠加到该基线上——`requests` 与 `limits` 按资源名
   合并，`claims` 按 claim 名合并；
3. 两者都未声明的 key 会从工作负载上清除。

也就是说，CacheRuntime 只需要表达相对模板的差异。只设置 `limits.memory` 时，仅该值被修改，模板中的
`requests.cpu`、`requests.memory` 和 `limits.cpu` 保持不变。

从 CacheRuntime 中删除某个 key 也遵循同样的规则：模板中有该 key 时回到模板的值，没有时从工作负载上清除。
`limits: {}` 这样的空列表不包含任何 key，效果与不写 `limits` 相同。

**限制**：
- ⚠️ 不能超过节点可用资源
- ⚠️ CacheRuntimeClass 模板中声明的 key **可以覆盖，但不能删除**：从 CacheRuntime 中省略某个 key 不会
  让组件变为不受限，而是保留模板中的值。如需放宽限制，请显式设置目标值；如需彻底去掉某项资源要求，
  请修改 CacheRuntimeClass 模板——运行时自身的资源要求本就描述在那里。
- ⚠️ **CacheRuntimeClass 模板中 `resources` 的修改会同步到已有实例**：模板是每次 reconcile 的基线，
  修改后，所有引用该 CacheRuntimeClass 的 CacheRuntime 中未被覆盖的 key 都会同步到 master 和 worker
  的工作负载，无需重建 CacheRuntime。控制器不监听 CacheRuntimeClass，因此修改在各 CacheRuntime 的下一次
  reconcile 时生效，而不是立即生效。client 不同步，模板中 `resources` 以外的字段也不同步。从模板中删除的 key，
  会从未覆盖该 key 的工作负载上清除。
- ⚠️ **以 CacheRuntime 和模板为准**：直接修改 master 或 worker 的 AdvancedStatefulSet 上的 resources，
  会在下一次 reconcile 时被改回。
- ⚠️ **升级说明**：在旧版本中，只设置了部分 `resources` 的 CacheRuntime 会丢失模板中的其余资源要求
  （见 [#6173](https://github.com/fluid-cloudnative/fluid/issues/6173)）。升级后的第一次 reconcile 会补回
  这些值，相应的 master 和 worker Pod 会经历一次性的资源更新；集群不支持原地更新时，Pod 会被重建。完全未设置
  `resources` 的 CacheRuntime 不受影响；但如果模板也未声明，而工作负载上有之前覆盖遗留或手动修改的值，
  这些值会被清除。
- ⚠️ **`claims` 只能引用模板中声明的 Pod 级 `resourceClaims`**：CacheRuntime 只能设置容器的
  `resources.claims`，不能添加 Pod 级的 `spec.resourceClaims`。引用 CacheRuntimeClass 模板中未声明的
  claim 名时，新建或重建的 master/worker Pod 会被 API server 拒绝（`must be one of the names in
  pod.spec.resourceClaims`），组件会一直处于 NotReady。未启用 `DynamicResourceAllocation` 的集群
  （K8s 1.34 之前默认关闭）会静默丢弃 `claims`。旧版本会忽略只设置了 `claims` 的 CacheRuntime。
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

## 4. 不支持更新的字段

除 `runtimeVersion` 和 `resources` 外，**修改 CacheRuntime spec 中的其他字段不会同步到 AdvancedStatefulSet**，包括但不限于：

- `env`: 环境变量
- `podMetadata`: Pod 元数据（labels/annotations）
- `args`/`command`: 容器启动参数
- `ports`: 容器端口
- `volumeMounts`/`volumes`: 存储配置
- `nodeSelector`/`tolerations`/`affinity`: 调度配置
- `securityContext`: 安全上下文
- `replicas`: 副本数
- `disabled`: 组件启用状态

**说明**：
- 修改这些字段后，需要重新部署 CacheRuntime 才能生效
- 系统不会自动检测或同步这些字段的变更

---

## 5. 总结

| 字段 | 是否支持更新 | 更新方式 |
|------|------------|---------|
| `runtimeVersion` | ✅ 支持 | 自动同步到 AdvancedStatefulSet |
| `resources` | ✅ 支持 | 自动同步到 AdvancedStatefulSet（需 K8s >= 1.27） |
| 其他所有字段 | ❌ 不支持 | 不会同步，需重新部署 |

**关键要点**：
1. 当前版本仅支持 `runtimeVersion` 和 `resources` 两个字段的动态更新
2. Cgroupv1 环境需分步更新这两个字段
3. 其他字段的修改不会生效，必须重新部署
