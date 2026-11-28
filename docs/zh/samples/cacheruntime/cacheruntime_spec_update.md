# CacheRuntime Spec 字段更新能力说明

## 1. 概述

本文档说明 CacheRuntime 的 Master 和 Worker 组件（基于 **AdvancedStatefulSet**）在创建后支持更新的 spec 字段，
以及 CacheRuntime 所用的 CacheRuntimeClass 如何在 CacheRuntime 首次部署时被固定下来。

当前版本会把以下三个字段同步到已有组件：
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
- Fluid 自带的 AdvancedStatefulSet 控制器只能原地更新镜像（会重启容器）。修改 `resources` 会重建 Pod，内存和
  `emptyDir` 中的缓存随之丢失
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

**取值规则**：创建时和每次 reconcile 时，都以 CacheRuntime 所绑定的 CacheRuntimeClass（见第 4 节）模板中该组件第一个容器的
镜像为基线，用 CacheRuntime 中设置的 `image` 和 `imageTag` 覆盖。

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
- ⚠️ **只设置 `image` 时，请确认新仓库中存在模板的 tag**：否则 Pod 会因 `ImagePullBackOff` 无法启动；
  更新时，该组件在新镜像拉取成功前不可用。
- ⚠️ **修改镜像会重启组件的容器**：AdvancedStatefulSet 原地更新镜像。
- ⚠️ **CacheRuntimeClass 中镜像的修改不会同步到已有 CacheRuntime**：已有 CacheRuntime 绑定的是其首次部署时的
  CacheRuntimeClass（见第 4 节）。如需让已有 CacheRuntime 使用新镜像，请在其上设置 `runtimeVersion`。
- ⚠️ 从 CacheRuntime 中删除 `runtimeVersion` 会让组件回到所绑定模板的镜像。该镜像没有 tag 或使用 digest 固定时
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

创建时和每次 reconcile 时，`resources` 按以下方式取值：

1. 以 CacheRuntime 所绑定的 CacheRuntimeClass（见第 4 节）模板中声明的 `resources` 为基线；
2. 将 CacheRuntime 上设置的 `resources` 按 key 叠加到该基线上——`requests` 与 `limits` 按资源名
   合并，`claims` 按 claim 名合并；
3. 两者都未声明的 key 会从工作负载上清除。

也就是说，CacheRuntime 只需要表达相对模板的差异。只设置 `limits.memory` 时，仅该值被修改，模板中的
`requests.cpu`、`requests.memory` 和 `limits.cpu` 保持不变。

从 CacheRuntime 中删除某个 key 也遵循同样的规则：所绑定的模板中有该 key 时回到该值，没有时从工作负载上清除。
之后修改 CacheRuntimeClass 不会改变这个值。
`limits: {}` 这样的空列表不包含任何 key，效果与不写 `limits` 相同。

**限制**：
- ⚠️ **修改 `resources` 会重建组件的 Pod**：Fluid 自带的 AdvancedStatefulSet 控制器不支持原地调整容器资源，
  因此会重建 Pod，内存或 `emptyDir` 中的缓存随之丢失
- ⚠️ 不能超过节点可用资源
- ⚠️ CacheRuntimeClass 模板中声明的 key **可以覆盖，但不能删除**：从 CacheRuntime 中省略某个 key 不会
  让组件变为不受限，而是保留模板中的值。如需调整资源要求，请在 CacheRuntime 中显式设置目标值。
- ⚠️ **CacheRuntimeClass 中 resources 的修改不会同步到已有 CacheRuntime**（见第 4 节）。如需修改已有
  CacheRuntime，请在其上设置目标值。
- ⚠️ **以 CacheRuntime 和其绑定的模板为准**：master 和 worker 容器的 resources 由 Fluid 管理，直接修改
  AdvancedStatefulSet 上的 resources，会在下一次 reconcile 时被改回。
- ⚠️ **合并后 request 不能超过 limit**：模板和 CacheRuntime 各自合法，合并后仍可能出现 request 大于 limit，
  例如模板 `requests.cpu: 1` 加上 CacheRuntime `limits.cpu: 500m`。hugepages（`hugepages-*`）和扩展资源
  （如 `nvidia.com/gpu`）要求更严：request 必须等于 limit，且设置了 request 就必须设置 limit，例如模板
  `nvidia.com/gpu` 的 request 和 limit 都为 1，CacheRuntime 只把 limit 改成 2，合并后同样冲突。
  此时控制器不同步该组件的 `resources`，工作负载保持当前值，并在 CacheRuntime 上发出
  `RuntimeResourcesConflict` Warning event，写明冲突的 key；新建的 CacheRuntime 若带有这样的组合，会创建失败。
  请在 CacheRuntime 上同时设置对应的 request（hugepages 和扩展资源设为与 limit 相等）。修改 CacheRuntimeClass
  只对之后创建的 CacheRuntime 生效。
- ⚠️ **`claims` 只能引用工作负载已声明的 Pod 级 `resourceClaims`**：CacheRuntime 只能设置容器的
  `resources.claims`，不能添加 Pod 级的 `spec.resourceClaims`；CacheRuntime 部署后在 CacheRuntimeClass 中
  新增的 claim 也不会到达它（见第 4 节）。创建时，引用模板中未声明的 claim 名会导致
  master/worker Pod 被 API server 拒绝（`must be one of the names in pod.spec.resourceClaims`），组件会
  一直处于 NotReady；未启用 `DynamicResourceAllocation` 的集群（K8s 1.34 之前默认关闭）则会静默丢弃
  `claims`。更新时，若某个 claim 既未在工作负载中声明、工作负载上也还没有，控制器不同步这个 claim，`resources` 的其余部分照常同步，并在 CacheRuntime 上发出
  `RuntimeResourcesConflict` Warning event，写明该 claim。如需使用新的 claim，请重建 CacheRuntime。

#### 查看资源同步状态

查看 CacheRuntime 的 `.status.conditions` 中 `type` 为 `ResourcesSynced` 的条目。
创建阶段会检查所有已启用组件的资源配置，包括 master、worker 和 client。创建完成后，
该 Condition 只跟踪已启用的 master 和 worker 的工作负载资源。Pod 的就绪情况可通过组件的 `readyReplicas` 单独查看。

| Status | Reason | 含义 |
| --- | --- | --- |
| `False` | `RuntimeResourceCreationConflict` | 创建阶段合并后的资源配置存在冲突，阻止组件创建。 |
| `False` | `RuntimeResourcesConflict` | 同步时，资源冲突或工作负载未声明的资源 claim 会导致部分资源设置被跳过。 |
| `Unknown` | `WaitingForSync` | 先前的创建冲突已修复，新的资源配置已通过校验，正在等待首次资源同步。即使 Pod 仍未就绪，也会清除旧的创建失败提示。 |
| `True` | `RuntimeResourcesSynced` | 所有已启用的 master 和 worker 工作负载都已同步期望的资源设置。 |

两种冲突的 Condition 都会在 `message` 中指明受影响的组件和具体原因。对应 Warning 事件的 Reason 均为 `RuntimeResourcesConflict`。

例如，使用以下命令等待名为 `my-cache` 的 CacheRuntime 报告资源已同步。

```bash
kubectl wait --for=condition=ResourcesSynced=True cacheruntime/my-cache --timeout=5m
```

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

## 4. CacheRuntimeClass 快照

CacheRuntime 首次部署时，Fluid 会把它引用的 CacheRuntimeClass 复制到 CacheRuntime 所在命名空间中一个不可变的
ConfigMap `fluid-runtime-class-snapshot-<runtime 名称>`，该 ConfigMap 归属于这个 CacheRuntime。此后 CacheRuntime
只从这份快照读取 class：组件模板、`options`、`executionEntries`（如 mountUFS、reportSummary 命令）、
`dataOperationSpecs` 和 `extraResources`。

新快照的 ConfigMap 注解 `cacheruntime.fluid.io/runtime-class-generation` 和
`cacheruntime.fluid.io/runtime-class-resource-version` 分别记录 Fluid 复制的 CacheRuntimeClass 对象的
`generation` 和 `resourceVersion`。旧快照可能没有这两个注解。

固定的是整个 class，而不是其中一部分，因为这些字段是相互配套编写的：只让其中一部分实时跟随，会把新的命令或
参数交给仍在运行旧镜像的组件。而 CacheRuntimeClass 是集群级资源，如果实时跟随，一次修改就会重启或重建所有
命名空间中使用它的 CacheRuntime 的组件。

- **修改或删除 CacheRuntimeClass** 只影响之后创建的 CacheRuntime。已有 CacheRuntime 即使其 class 被删除也能
  继续工作。如需让已有 CacheRuntime 使用新的 class，请重建它。
- **修改已有 CacheRuntime 的 `runtimeClassName`** 不会生效：它仍绑定原来的快照，Fluid 会发出 `RuntimeClassChanged`
  Warning event。如需使用其他 class，请重建 CacheRuntime。

### 恢复损坏的快照

如果 Fluid 无法解析快照，它会保留原 ConfigMap 并发出 `RuntimeClassSnapshotInvalid` Warning 事件。
CacheRuntime 的协调和 DataLoad 等数据操作会等待有效快照。**仅在快照损坏或无法解析时，才将删除快照 ConfigMap
作为兜底恢复手段。** Fluid 会从 `spec.runtimeClassName` 指定的 CacheRuntimeClass 的当前版本重新生成快照。

如果该 class 自首次部署以来发生过变化，恢复后各组件可能混用不同版本的配置。

- 对于 **master 和 worker 工作负载**，Fluid 按新快照和 CacheRuntime 的覆盖配置同步主容器的镜像和 `resources`。
  镜像变化会重启容器，资源变化会重建 Pod。其余所有 PodTemplate 字段（包括 `args`、`env` 和 `volumes`）均保留原值。
- **现有 client DaemonSet** 的所有字段均保持原值，包括镜像和整个 PodTemplate。
- Fluid 根据新快照中的 `options` 和 CacheRuntime 的覆盖配置重新生成**运行时配置**，其中也包括
  client 的选项。控制器后续执行命令时使用新快照的 `executionEntries`，新建的 DataLoad 等数据操作作业使用其
  `dataOperationSpecs`。因此，旧镜像可能收到新版本的命令或选项。

**要让已有 CacheRuntime 采用新版 CacheRuntimeClass，请重建 CacheRuntime**，使所有组件都从同一版本的 class 创建。
请勿通过删除快照来升级 class。

```bash
# 查看 CacheRuntime 绑定的 class
kubectl get cm fluid-runtime-class-snapshot-my-cache -o jsonpath='{.data.runtimeclass\.json}'
```

---

## 5. 不支持更新的字段

**第 3 节未列出的字段一律不支持在创建后更新**，修改后不会同步到 AdvancedStatefulSet。

**说明**：
- 这类字段修改后，需要重新部署 CacheRuntime 才能生效
- 系统不会自动检测或同步这类变更

---

## 6. 总结

| 字段 | 是否支持更新 | 更新方式 |
|------|------------|---------|
| `runtimeVersion` | ✅ 支持 | 自动同步到 AdvancedStatefulSet（重启容器） |
| `resources` | ✅ 支持 | 自动同步到 AdvancedStatefulSet（重建 Pod） |
| `replicas` | ✅ 支持 | 自动同步到 AdvancedStatefulSet（缩容会丢失该 Worker 上的缓存） |
| 其他所有字段 | ❌ 不支持 | 不会同步，需重新部署 |

**关键要点**：
1. 当前版本支持 `runtimeVersion`、`resources` 和 `replicas` 三个字段的动态更新
2. 仅通过 `runtimeVersion` 更新镜像时，受影响的容器在原 Pod 内重启，进程内存中的缓存会丢失，`emptyDir` 中的缓存会保留。修改 `resources` 会重建 Pod，两处缓存都会丢失。
3. 其他字段的修改不会生效，必须重新部署
4. CacheRuntime 绑定的是其首次部署时的 CacheRuntimeClass；修改 class 只影响之后创建的 CacheRuntime
