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

**取值规则**：创建时和每次 reconcile 时，都以 CacheRuntimeClass 模板中该组件第一个容器的镜像为基线，用 CacheRuntime 中设置的
`image` 和 `imageTag` 覆盖。

**补全规则**:
- `image` 和 `imageTag` 可以只设置其中一个，缺少的一半从模板镜像补全。
  以模板镜像 `fluid-cache:v1.0.0` 为例：
  - 只设置 `imageTag: v1.1.0`：沿用模板的镜像名，得到 `fluid-cache:v1.1.0`
  - 只设置 `image: registry.example.com/fluid-cache`：沿用模板的 tag（**而不是 `latest`**），得到
    `registry.example.com/fluid-cache:v1.0.0`
  - 都不设置：使用模板镜像 `fluid-cache:v1.0.0`
- 同时设置了 `image` 和 `imageTag` 的 `runtimeVersion` 直接使用，即使模板镜像使用 digest 固定。
- 只设置了其中一半、而另一半无法补全时（只设置了 `image` 而模板镜像没有 tag；模板镜像或 `image` 使用 digest
  （`@sha256:...`）固定），创建时组件使用模板镜像；更新时修改不会生效，组件保持原有镜像。这两种情况下，控制器都会在
  CacheRuntime 上发出 `RuntimeVersionIncomplete` Warning event，并输出日志；client 不参与更新同步，只在创建时报告。`runtimeVersion` 无法表达 digest 引用，如需固定到 digest，请写在 CacheRuntimeClass 模板中。
- 未单独指定镜像的数据操作（如 DataLoad）使用与 worker 相同的补全结果。worker 的版本无法补全时，数据操作使用模板镜像，
  因此可能与保持原有镜像的 worker 不一致，例如模板镜像改为使用 digest 固定之后。

**限制**:
- ⚠️ Cgroupv1 环境中，不能与 `resources` 字段同时更新（需分步操作，见下方资源字段说明）
- ⚠️ **只设置 `image` 时，请确认新仓库中存在模板的 tag**：否则 Pod 会因 `ImagePullBackOff` 无法启动；
  原地更新时，该组件在新镜像拉取成功前不可用。反过来也一样：只设置了 `imageTag` 的实例，在模板的镜像仓库
  被修改后，会改用新仓库中的这个 tag，请确认新仓库中也存在它。
- ⚠️ **CacheRuntimeClass 模板中镜像的修改会同步到已有实例**：修改模板中的镜像后，
  所有引用该 CacheRuntimeClass 的 CacheRuntime 中，未禁用、且未完整设置 `runtimeVersion` 的 master 和 worker
  会按新模板镜像重新计算，结果变化时滚动到新镜像（只设置了 `imageTag` 的保留自己的 tag，只设置了 `image` 的
  保留自己的镜像名；无法补全为 `镜像名:tag` 形式的实例保持原有镜像，例如新模板镜像使用 digest 固定时）。控制器不监听
  CacheRuntimeClass，因此修改在各 CacheRuntime 的下一次 reconcile 时生效，而不是立即生效。client 不同步。
- ⚠️ 从 CacheRuntime 中删除 `runtimeVersion` 会让组件回到模板镜像。模板镜像没有 tag 或使用 digest 固定时
  例外：组件保持当前镜像，需要手动设置完整的 `runtimeVersion` 或重建 CacheRuntime。

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
  让组件变为不受限，而是保留模板中的值。如需调整资源要求，请在 CacheRuntime 中显式设置目标值。
- ⚠️ CacheRuntimeClass 的修改如何影响已有 CacheRuntime，由 #6213 跟踪；请勿依赖这些修改会传递到已有实例。
- ⚠️ **以 CacheRuntime 和模板为准**：直接修改 master 或 worker 的 AdvancedStatefulSet 上的 resources，
  会在下一次 reconcile 时被改回。
- ⚠️ **合并后 request 不能超过 limit**：模板和 CacheRuntime 各自合法，合并后仍可能出现 request 大于 limit，
  例如模板 `requests.cpu: 1` 加上 CacheRuntime `limits.cpu: 500m`。hugepages（`hugepages-*`）和扩展资源
  （如 `nvidia.com/gpu`）要求更严：request 必须等于 limit，且设置了 request 就必须设置 limit，例如模板
  `nvidia.com/gpu` 的 request 和 limit 都为 1，CacheRuntime 只把 limit 改成 2，合并后同样冲突。
  此时控制器不同步该组件的 `resources`，工作负载保持当前值，并在 CacheRuntime 上发出
  `RuntimeResourcesConflict` Warning event，写明冲突的 key；新建的 CacheRuntime 若带有这样的组合，会创建失败。
  请在 CacheRuntime 上同时设置对应的 request（hugepages 和扩展资源设为与 limit 相等），或修改模板。
- ⚠️ **`claims` 只能引用工作负载已声明的 Pod 级 `resourceClaims`**：CacheRuntime 只能设置容器的
  `resources.claims`，不能添加 Pod 级的 `spec.resourceClaims`；创建后在 CacheRuntimeClass 模板中新增的
  Pod 级 `resourceClaims` 也不会同步到已有的工作负载。创建时，引用模板中未声明的 claim 名会导致
  master/worker Pod 被 API server 拒绝（`must be one of the names in pod.spec.resourceClaims`），组件会
  一直处于 NotReady；未启用 `DynamicResourceAllocation` 的集群（K8s 1.34 之前默认关闭）则会静默丢弃
  `claims`。更新时，若某个 claim 既未在工作负载中声明、工作负载上也还没有（无论来自 CacheRuntime 还是模板
  修改），控制器不同步这个 claim，`resources` 的其余部分照常同步，并在 CacheRuntime 上发出
  `RuntimeResourcesConflict` Warning event，写明该 claim。如需使用新的 claim，请重建 CacheRuntime。
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
