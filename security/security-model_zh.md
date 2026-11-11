# Fluid 安全模型

本文描述 Fluid 的信任边界、哪些角色可以创建和使用其资源，以及哪些行为属于
安全漏洞。下述安全属性适用于遵循本模型的部署。其他部署配置需要单独进行
安全评估。

配套文档：[self-assessment.md](self-assessment.md)（CNCF TAG-Security
自评报告）、[SECURITY.md](../SECURITY.md)（漏洞报告方式）、
[security-model.md](security-model.md)（英文版本）。

代码引用使用路径，必要时注明函数名。引用失效属于文档缺陷。

## 1. 为什么 Fluid 工作负载需要特权

Fluid 将远端数据作为文件系统提供给 Pod。标准的节点级 FUSE 工作负载使用
`privileged: true`（FUSE 本身至少需要 `CAP_SYS_ADMIN` 和 `/dev/fuse`），
以及设置了 `mountPropagation: Bidirectional` 的 `hostPath` 卷。这样，FUSE
容器内创建的挂载会出现在宿主机的挂载命名空间中，供同节点的其他 Pod 使用。

这些工作负载运行在物理 Dataset 所在的命名空间。提供 `spec.fuse.networkMode`
的 Runtime API 默认使用 `HostNetwork`；设置
`spec.fuse.networkMode: ContainerNetwork` 可改用容器网络
（`api/v1alpha1/container_network.go`）。`CacheRuntime` 使用其 class 提供的
客户端 Pod 模板，是否启用特权和宿主机网络取决于该模板（§4）。Sidecar 模式
另有非特权选项，其限制见 S4。

配置这些 FUSE 工作负载的权限包含节点级特权。运维人员在委派访问权限前，
必须了解哪些 Fluid 资源提供了这项权限。

## 2. 角色

| 角色 | 信任级别 | 职责 |
| --- | --- | --- |
| 集群管理员 | 拥有完整的集群权限 | 安装 Fluid，管理集群级 Fluid 资源和 PersistentVolume，决定哪些节点池运行缓存工作负载，管理 Pod Security Admission（PSA）标签 |
| 数据管理员 | 受信任，可在其工作负载运行的节点上使用节点级特权 | 创建 `Dataset` 和 Runtime 资源，包括使用者命名空间中的引用 Dataset；持有存储凭据；调整缓存拓扑和主机级 FUSE 选项 |
| 数据使用者 | 普通租户，不可信 | 运行挂载 Dataset PVC 的应用 Pod；可选择读取 Dataset 状态 |
| Fluid 控制面 | 具有集群特权的系统组件 | 位于 `fluid-system` 的控制器、webhook 和 CSI 插件，代表其他三类角色执行操作 |

数据管理员的信任级别与获准创建特权 Pod 的人相同。数据使用者位于此信任边界
之外，可能具有恶意。命名空间级授权无法将特权工作负载的影响限制在该命名
空间内。

## 3. 基本规则

> 授予会启动特权 FUSE 工作负载的 Runtime 写权限，必须视为授予在其节点上
> 运行特权 Pod 的能力。这些 Pod 可以使用主机路径和双向挂载传播。

对于提供 FUSE spec 的 Runtime API，写入者可以选择镜像（`spec.fuse.image`）、
入口命令（`spec.fuse.command`、`spec.fuse.args`）、环境变量和卷。这些字段
让数据管理员能够配置存储系统所需的客户端。`CacheRuntime` 提供
`spec.client.runtimeVersion.image` 和 `spec.client.args` 等客户端设置，
安全上下文则由 class 的 Pod 模板提供。它仍仅限管理员配置，工作负载的权限
范围需要根据该模板评估（§4）。

`Dataset` 写权限也应由管理员持有，但原因不同。物理 Dataset 需要匹配的
Runtime 才能启动 FUSE 工作负载。其挂载选项可以配置该工作负载，Secret 引用
用于选择存储凭据。引用 Dataset 则授权访问其他命名空间的数据（§6.2）。因此，
仅写入 Dataset 并不无条件等同于创建任意特权 Pod。

审批 Runtime 写权限申请时，应与审批创建特权 Pod 的申请同样谨慎。不要向
不可信租户授予 Runtime 或 Dataset 写权限。§4 的资源分级说明了各项授权允许
执行的操作。

只校验 `command` 无法约束能够替换镜像的人。RBAC 决定谁可以配置特权工作
负载。Fluid 没有为这些字段提供 validating webhook。管理员可以增加准入策略
作为纵深防御，但仅覆盖部分字段的策略不能替代访问权限边界（§6，S3）。

承载特权 FUSE 工作负载的命名空间需要设置
`pod-security.kubernetes.io/enforce: privileged`，除非配置了等效的 PSA 豁免。
这些命名空间不是租户沙箱。不可信使用者应位于独立的命名空间，并对其 Pod
强制执行限制（§6，S2）。

## 4. 资源分级

Fluid 提供 14 个面向用户的 `data.fluid.io` CRD，以及一个内部
`workload.fluid.io` CRD。各级说明写入相应资源所授予的权限。

### Tier 0：仅限集群管理员（集群级）

| 资源 | 原因 |
| --- | --- |
| `CacheRuntimeClass` | 顶层 `topology.{master,worker,client}.template` 是完整的 `corev1.PodTemplateSpec`，会被深拷贝到渲染后的工作负载中（`api/v1alpha1/cacheruntimeclass_types.go`；`pkg/ddc/cache/engine/transform_common.go`）。写入者可以设置 `privileged`、`hostPID`、`hostNetwork`、`capabilities`、`serviceAccountName` 和任意 `hostPath`。class 还定义了 `dataOperationSpecs[].image`。这是 Fluid API 中权限范围最广的资源 |
| `ThinRuntimeProfile` | 为所有引用它的 `ThinRuntime` 提供 FUSE 镜像、命令、参数、卷和 `hostNetwork`。`spec.nodePublishSecretPolicy` 控制 `pvc://` 挂载的 Secret 处理方式（§5） |

### Tier 1：数据管理员（命名空间级，涉及特权或敏感配置）

| 资源 | 原因 |
| --- | --- |
| `AlluxioRuntime`、`JuiceFSRuntime`、`JindoRuntime`、`ThinRuntime`、`EFCRuntime`、`VineyardRuntime`、`CacheRuntime` | 配置节点级 FUSE 工作负载。标准特权工作负载遵循 §3；`CacheRuntime` 配置 class 提供的 Pod 模板中的客户端字段，其工作负载是否具有特权由模板的安全上下文决定（`api/v1alpha1/cacheruntime_types.go`） |
| `Dataset` | 持有 `mountPoint` 和 `encryptOptions`（引用本命名空间中的 Secret）。`spec.mounts[].options` 配置 FUSE 工作负载：`fluid.fuse.hostvolume` 注入 `hostPath`（`pkg/ddc/thin/transform_fuse.go`），`local://` 形式的 mountPoint 在 Jindo 中转换为 `hostPath`。`dataset://ns/name` 形式的 mountPoint 定义引用 Dataset，向使用者开放其他命名空间的数据（§6.2）。物理 Dataset 需要匹配的 Runtime 才能启动 FUSE 工作负载，但其凭据引用和数据访问范围本身就要求管理员级信任 |
| `DataProcess` | `spec.processor.job.podSpec` 是完整的 `corev1.PodSpec`，且 `spec.processor.serviceAccountName` 可设置（`api/v1alpha1/dataprocess_types.go`）。写入它等同于在该命名空间中以任意 ServiceAccount 创建任意 Pod |
| `DataMigrate` | 内联 `VersionSpec`（`api/v1alpha1/datamigrate_types.go`），因此镜像由调用方选择，且 `encryptOptions` 引用 Secret |

### Tier 2：可委派给受信任用户

| 资源 | 可委派的依据 | 仍需关注的输入 |
| --- | --- | --- |
| `DataLoad` | 调用方不能覆盖镜像，也不能指定 `podSpec` 或 `volumes`；镜像由 Runtime 选择（`pkg/utils/docker/image.go`，`GetWorkerImage`）。对于 CacheRuntime，class 可以通过 `dataOperationSpecs[].image` 提供操作镜像（`pkg/ddc/cache/engine/dataload.go`） | `spec.options` 和 `spec.target[].path` 传入 Runtime 的加载脚本。JuiceFS 将选项拼接为 `--key=value` 参数字符串（`pkg/ddc/juicefs/data_load.go`，`genDataLoadValue`）。若参数或命令注入使这些输入超出预期操作，属于漏洞范围（§7） |
| `DataBackup` | 不能定义任意 Pod，但安全委派需要限制备份目标和执行身份 | Alluxio 本地备份将调用方指定的 `spec.backupPath` 用作 master 节点上的可写 `hostPath`，且 `spec.runAs` 可以覆盖 Runtime 的身份（`pkg/ddc/alluxio/backup_data.go`）。除非准入策略限制了这些字段，否则此操作应仅限管理员。超出获授权备份操作的注入仍属于漏洞范围 |

两者都会执行 I/O 并消耗容量。对缓存操作的信任不包含运行任意命令或访问
无关数据的权限。DataBackup 没有 `podSpec` 并不意味着可以安全委派；命名
空间 RBAC 无法限制其备份目标。§6.3 的角色示例仅授予 DataLoad 写权限。

### Tier 3：数据使用者

| 操作 | 所需权限 |
| --- | --- |
| 挂载 Dataset | 无需 Fluid RBAC。应用引用 Fluid 在应用命名空间中供应的 PVC，其名称与 Dataset 相同 |
| 观察缓存状态 | 对 `datasets` 的 `get`/`list`/`watch`；使用状态端点的工具可能还需要对 `datasets/status` 的 `get`。两个读取端点都会暴露完整的 Dataset 对象，包括 spec（§5） |

### 内部资源

`AdvancedStatefulSet`（`workload.fluid.io`）属于控制面。用户不应直接写入。

### 委派汇总

| 角色 | 可以创建 / 写入 | 可以使用 |
| --- | --- | --- |
| 集群管理员 | Tier 0；Fluid 安装；PV；节点池；PSA 标签 | 所有资源 |
| 数据管理员 | 授权命名空间中的 Tier 1 资源，包括使用者命名空间中的引用 Dataset | 这些资源引用的凭据；缓存拓扑；主机级 FUSE 选项（`networkMode`、`fluid.fuse.hostvolume`、`runtime.fluid.io/fuse.hostpid`、`tieredStore` 路径） |
| 受信任的缓存操作员 | 针对已开放 Dataset 的 Tier 2 资源；只有强制实施备份目标和身份限制后，才可授予 DataBackup 写权限 | 读取这些 Dataset 及其操作 |
| 数据使用者 | 不可写入任何 `data.fluid.io` 资源 | 已供应的 PVC；可选择读取 Dataset 状态 |

使用数据不需要修改供应该数据的资源。

### 关于 `config/rbac/` 的说明

`config/rbac/` 下由 kubebuilder 脚手架生成的 `*-editor-role.yaml` /
`*-viewer-role.yaml` 文件没有被 `config/rbac/kustomization.yaml` 或 Helm chart
引用。这些安装路径不会创建这些角色。其中一些角色授予 Runtime 类型的
写权限，属于管理员级授权。不要将它们应用于租户。

## 5. 安全属性及其限制

这些属性定义了预期的安全边界。选择 Runtime 和配置部署时，需要考虑下述
限制条件。

### 凭据查找与处理

挂载级 `encryptOptions` 和 Dataset 级 `sharedEncryptOptions` 在 Dataset
所在命名空间中解析 Secret 引用（`pkg/ddc/alluxio/ufs_internal.go`、
`pkg/ddc/thin/ufs_internal.go`）。任何能够读取 Dataset 的人都能看到
`spec.mounts[].options` 或 `spec.sharedOptions` 中的明文值。凭据应使用 Secret
引用。

Secret 引用使凭据值不出现在 Dataset spec 中，但不保证所有 Runtime 都不会将
这些值写入生成的配置。JuiceFS 使用 `env.valueFrom.secretKeyRef`。Alluxio 和
Thin 提供 Secret 卷处理路径，但其他配置模式可能将解析后的值写入生成的配置
或 Helm values。较旧的 Jindo 处理路径也会将解析后的值渲染到配置中；JindoCache
和 JindoFSx 的行为取决于所选镜像版本是否支持 Secret 挂载。在认定配置可以
安全公开之前，应检查 Runtime、配置模式和镜像版本
（`pkg/ddc/juicefs/transform_fuse.go`，`genValue`；
`pkg/ddc/alluxio/transform.go`，`transformCommonPart`；
`pkg/ddc/thin/transform_config.go`，`extractMountOptions`；
`pkg/ddc/jindo/transform.go`，`transformMaster`）。

Kubernetes CRD 的状态读取返回完整对象。对 `datasets/status` 的 `get` 权限
并非仅能读取状态、按字段过滤的权限。如果不应让使用者看到 Dataset spec，
则两个读取端点都不能授权；应另行提供经过过滤的可观测性接口。

### 显式跨命名空间复制 Secret

设置 `ThinRuntimeProfile.spec.nodePublishSecretPolicy:
CopyNodePublishSecretAndMountIfNotExists` 时，ThinRuntime 控制器会将 `pvc://`
挂载所对应 PV 的 `nodePublishSecretRef` 引用的 Secret 复制到 Runtime 所在
命名空间，命名为 `<runtime>-thin-publish-secret`
（`pkg/ddc/thin/transform_pvc_mounts.go`，
`transfromSecretsForPersistentVolumeClaimMounts`；
`pkg/utils/kubeclient/secret.go`，`CopySecretToNamespace`）。

源 Secret 来自集群级 PV 中的引用；Secret 命名空间为空时使用 `default`。
PVC 必须位于 Dataset 所在命名空间。因此，查找范围限于已绑定到该命名空间
内 PVC 的 PV。默认 Helm 配置没有向 `thinruntime-controller` 授予复制所需的
Secret API 权限。管理员除了选择复制策略，还必须显式授予这些权限。

默认策略 `MountNodePublishSecretIfExists` 按名称挂载已存在的 Secret，不进行
复制。本小节描述的是 PV 节点发布 Secret 的复制路径，并不涵盖所有可能的
Runtime 凭据处理流程。

### 引用 Dataset 与 CSI 挂载边界

`dataset://ns/name/subpath` 引用必须恰好指向一个物理 Dataset。它不能混用
挂载协议，也不能形成引用链。其子路径不能逃出物理 Dataset 的根目录
（`pkg/ddc/thin/referencedataset/engine.go`，`checkDatasetMountSupport`；
`pkg/ddc/base/dataset.go`，`GetPhysicalDatasetSubPath`）。

在 Linux 上，对于非空子路径，默认的 `bindMount` 发布路径通过
`resolveMountSource` 使用 `O_NOFOLLOW` 逐个遍历路径分量，拒绝符号链接，
并通过文件描述符固定已检查的 inode（`pkg/csi/plugins/subpath_linux.go`，
`resolveMountSource`）。FUSE 挂载下的符号链接不能将该子路径的绑定挂载
重定向到任意主机路径。`NodePublishVolume` 通过 `cmdguard` 调用 `mount`
（`pkg/csi/plugins/nodeserver.go`，`NodePublishVolume`）。

非默认的 `symlink` 发布路径根据原始路径字符串创建链接，不使用固定的 inode，
因此不提供相同的 inode 固定保证。S2 使用 `bindMount`；依赖此保证前，应同时
检查 `csi.nodePublishMethod` 和 PV 上任何 `node_publish_method` 覆盖设置。

### 数据操作的目标

`DataLoad` 和 `DataProcess` 会拒绝位于其他命名空间的目标 Dataset
（`pkg/controllers/v1alpha1/dataload/implement.go`、
`pkg/controllers/v1alpha1/dataprocess/implement.go`）。

### 控制面权限

各组件使用独立的 ServiceAccount 和角色。Helm chart 的 RBAC 规则不使用
通配的 apiGroups、资源或动词。

默认 Helm 设置为 `helmDriver: configmap`。使用 `fluid.helmDriver.rbacs` 的
Runtime 和 Dataset 控制器规则，仅在 `helmDriver: secret` 时获得集群范围的
Secret API 权限（`charts/fluid/fluid/templates/_helpers.tpl`）。Webhook 具有
独立的、命名空间级的证书 Secret 权限。非 Helm 清单可能授予不同的权限；
应检查实际安装路径渲染出的 RBAC。没有直接的 Secret API 授权，并不意味着
Fluid 控制面没有特权。

设置 `csi.useNodeAuthorization: true`（默认）时，CSI 插件使用 kubelet 凭据
读取和补丁更新 Node。其他 API 操作仍使用其 ServiceAccount。集群必须启用
Node authorizer 和 NodeRestriction 准入插件，其保护机制才会生效
（`pkg/csi/plugins/nodeserver.go`，`getNode`、`patchNodeWithLabel`）。

### 准入与 sidecar 特权

通过 Helm 安装的 `sidecar.fuse.fluid.io` 和 `fuse.serverful.fluid.io` webhook
使用 `failurePolicy: Fail`。调用出错时，会拒绝匹配的 Pod CREATE 请求
（`charts/fluid/fluid/templates/webhook/webhookconfiguration.yaml`）。标签
选择器限定了它们覆盖的 Pod；这不是针对所有 Pod 或 Fluid CRD 的准入策略。

变更准入先于 PSA 校验执行。因此，注入特权 sidecar 的 Pod 会被强制执行的
`baseline` 或 `restricted` PSA 拒绝，除非请求获得豁免。这一边界仍由集群
强制执行。

Serverless 路径提供非特权 sidecar 模式。当 Pod 同时具有
`serverless.fluid.io/inject: "true"` 和
`unprivileged.sidecar.fluid.io/inject: "true"` 标签时，mutator 会设置
`privileged: false`，从 `Capabilities.Add` 中移除 `SYS_ADMIN`，并从模板中
移除已知的主机 FUSE 挂载卷和设备卷
（`pkg/application/inject/fuse/mutator/mutator_unprivileged.go`，
`transformTemplateWithUnprivilegedSidecarEnabled`）。它不会将 `SYS_ADMIN`
加入 `Capabilities.Drop`，也不会移除所有可能的主机卷。

此模式不保证符合 PSA。当前变更路径仍会将应用的 Dataset PVC 卷转换为
`hostPath`（`pkg/application/inject/fuse/mutator/mutator_default.go`，
`defaultMutateDatasetVolumes`）。标准的 `baseline` 和 `restricted` PSA
会拒绝 `hostPath`。为不可信使用者选择此模式前，请参阅 S4。

## 6. 部署场景

### S1：单租户 / 受信任团队

所有用户都受信任，可以使用相关存储权限和节点特权。默认安装不需要额外的
租户 RBAC 分离。

### S2：多租户，由管理员供应数据集（推荐）

本模型通过默认 CSI `bindMount` 路径支持不可信使用者。数据由管理员供应，
使用者 Pod 受到限制。管理员还必须确认复制到使用者命名空间的配置不包含
超出预期授权的凭据或数据（§6.2）。

#### 6.1 布局

| 命名空间 | PSA `enforce` | 内容 | 写入者 |
| --- | --- | --- | --- |
| `fluid-system` | `privileged` | Fluid 控制面 | 集群管理员 |
| `fluid-datasets` | `privileged` | 物理 `Dataset` + Runtime，缓存 master/worker/FUSE | 数据管理员 |
| `team-a`、`team-b`、... | `restricted` | 引用 `Dataset`（管理员创建）+ 使用者 Pod | 管理员：Dataset；使用者：仅 Pod |

使用者不获得任何 `data.fluid.io` 资源的写动词。

#### 6.2 通过引用 Dataset 跨命名空间访问

Kubernetes PVC 不能被跨命名空间引用。管理员在使用者命名空间中创建引用
Dataset：

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: imagenet
  namespace: team-a
spec:
  mounts:
  - mountPoint: dataset://fluid-datasets/imagenet   # optionally /subpath
    name: imagenet
```

Fluid 在 `team-a` 中供应 PVC `imagenet`。使用者挂载这个本地 PVC。创建引用
会授予对物理 Dataset 或所选子路径的访问权限，属于管理员的授权决策。读写
权限单独设置：引用默认使用 `ReadOnlyMany`，物理 PV 的访问模式限制了可选
模式（`pkg/ddc/thin/referencedataset/volume.go`，`accessModesForVirtualDataset`）。
需要只读访问时，应显式设置 Dataset 的 `spec.accessModes: [ReadOnlyMany]`，
并将使用者的卷挂载设置为 `readOnly: true`；适用时还应使用只读后端凭据。

通过 CSI，使用者获得由 `fluid-datasets` 中物理 FUSE 工作负载提供的绑定
挂载。使用者命名空间可以保持 `restricted`。

Fluid 还会将物理 FUSE DaemonSet 复制到引用命名空间，由引用 Dataset 持有，
供 sidecar 模板使用（`pkg/ddc/thin/referencedataset/cm.go`，
`copyFuseDaemonSetForRefDataset`）。副本增加
`nodeSelector: {fluid.io/fuse-balloon: "true"}`，以避免被调度。管理员不得为
节点设置这个保留标签。复制的模板保留原有特权设置；sidecar 注入不会使其
符合使用者命名空间的 PSA（§5，S4）。

建立引用时，Fluid 还会将 Runtime 配置 ConfigMap 复制到使用者命名空间，
即使使用者采用 CSI 也是如此（`pkg/ddc/thin/referencedataset/cm.go`，
`createConfigMapForRefDataset`）。某些旧版或其他 Runtime 配置包含解析后的
凭据（§5）。获准创建 Pod 的使用者无需 ConfigMap 读取 RBAC 权限，就能挂载
本命名空间中的 ConfigMap；`restricted` PSA 也允许这种卷。因此，不授予
Dataset 读取权限无法保护复制过来的凭据。开放引用前，应检查复制的配置。
如果某种 Runtime 模式复制了超出数据授权范围的后端凭据，就不符合 S2 的
前提。授权引用 Dataset 或其子路径，本身并不代表授权披露其后端凭据。

#### 6.3 使用者角色

Dataset 读取权限用于可选的可观测性功能，挂载不需要此权限。只有在允许公开
完整 spec 时才应授予读取权限，且不要将凭据写入明文挂载选项：

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: fluid-dataset-reader
  namespace: team-a
rules:
- apiGroups: ["data.fluid.io"]
  resources: ["datasets"]
  verbs: ["get", "list", "watch"]
```

如果工具使用状态端点，可增加以下规则。它同样允许读取完整对象，不会隐藏
spec：

```yaml
- apiGroups: ["data.fluid.io"]
  resources: ["datasets/status"]
  verbs: ["get"]
```

受信任的缓存操作员（Tier 2）：

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: fluid-cache-operator
  namespace: fluid-datasets
rules:
- apiGroups: ["data.fluid.io"]
  resources: ["datasets", "datasets/status"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["data.fluid.io"]
  resources: ["dataloads"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["data.fluid.io"]
  resources: ["dataloads/status"]
  verbs: ["get"]
```

此角色有意不授予 DataBackup 写权限。只有在强制限制允许的备份目标和
`spec.runAs` 身份后，才应增加该权限。否则，创建备份仍应仅由数据管理员
负责（§4）。

不要向不可信租户授予 `alluxioruntimes`、`juicefsruntimes`、`jindoruntimes`、
`thinruntimes`、`efcruntimes`、`vineyardruntimes`、`cacheruntimes`、`datasets`、
`dataprocesses`、`datamigrates`、`thinruntimeprofiles` 或 `cacheruntimeclasses`
的写权限。

#### 6.4 节点本地缓存数据

磁盘层级缓存（`spec.tieredStore.levels[].path`）和 FUSE 挂载点使用节点
`hostPath` 存储。同节点上的其他特权工作负载可以读取这些路径。Fluid 不会
将缓存内容与这些工作负载隔离。

对于敏感级别不同的 Dataset，应为每个敏感域分配独立节点池，为 Runtime
设置不同的 `tieredStore` 路径，并由管理员强制执行节点放置约束。如果担心
磁盘缓存残留，优先使用 `MEM` 层；仅分离路径不能实现隔离。

### S3：自建 Dataset 和 Runtime 的自助租户

Runtime 写权限赋予租户节点级特权。命名空间 RBAC 无法约束这种特权。对于
互不信任的租户，应使用独立集群，或由管理员强制执行数据面隔离，例如租户
专用节点。放置约束必须由宿主集群管理员的准入策略和节点配置强制执行，不能
依赖租户自行选择的 `nodeSelector` 或容忍度。

vCluster 等虚拟控制面本身无法隔离共享宿主节点上的特权工作负载，还需要
隔离的数据面。

准入策略可以限制镜像覆盖，作为纵深防御。以下 Kubernetes v1.30+ 示例检查
显式覆盖的 ThinRuntime FUSE 镜像的仓库前缀。Binding 通过 `Deny` 强制执行
该检查：

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicy
metadata:
  name: fluid-thinruntime-image-allowlist
spec:
  failurePolicy: Fail
  matchConstraints:
    resourceRules:
    - apiGroups: ["data.fluid.io"]
      apiVersions: ["v1alpha1"]
      operations: ["CREATE", "UPDATE"]
      resources: ["thinruntimes"]
  validations:
  - expression: >-
      !has(object.spec.fuse) || !has(object.spec.fuse.image) ||
      object.spec.fuse.image.startsWith("registry.example.com/fluid/")
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicyBinding
metadata:
  name: fluid-thinruntime-image-allowlist
spec:
  policyName: fluid-thinruntime-image-allowlist
  validationActions: [Deny]
```

此示例允许省略镜像，不校验从 ThinRuntimeProfile 继承的镜像、镜像标签、其他
Runtime 类型或 Dataset 的 `fluid.fuse.hostvolume` 选项。管理员必须控制继承
镜像的来源，并为其他需要限制的字段添加策略。此策略只覆盖部分字段，不能
使 Runtime 写权限成为可安全授予租户的权限。

### S4：Serverless / sidecar 注入

Serverless webhook 路径向带有 `serverless.fluid.io/inject: "true"` 标签的
Pod 注入 FUSE sidecar。Pod 创建者选择该标签，还可以增加
`unprivileged.sidecar.fluid.io/inject: "true"` 标签来请求非特权模式。

默认 sidecar 具有特权。在允许特权工作负载的命名空间中，不应让不可信使用者
仅通过给 Pod 添加标签就获得特权 sidecar。PSA 必须校验变更后的完整 Pod，
管理员不得让不可信使用者豁免这些检查。

当前非特权 mutator 仍会引入 Dataset `hostPath` 卷，因此也会被标准的
`baseline` 或 `restricted` PSA 拒绝。移除 sidecar 的特权标志不能提供 S2
基于 CSI 的隔离。不可信使用者应采用 S2，不应为使 sidecar 注入成功而放宽
PSA。使用 sidecar 模式的环境，需要对其设备支持和变更后的完整 Pod 进行
单独的安全评估。

独立的 `fuse.serverful.fluid.io/inject: "true"` 选择器使用 serverful 调度
路径。仅添加非特权标签，不会使该路径切换到 serverless 非特权 mutator
（`pkg/utils/annotations.go`，`GetServerlessPlatform`；
`pkg/webhook/handler/mutating/mutating_handler.go`）。

处理函数仍识别已弃用的 `fuse.sidecar.fluid.io/inject` 标签，但仅有此标签
不会匹配默认 Helm 配置中任何一个 webhook 选择器。新部署应使用当前标签。

## 7. 漏洞分诊：哪些行为属于漏洞

### 范围内

以下情况适用于上述部署前提：

1. 没有 Tier 0/1 写权限的主体提权。其权限仅限于挂载 PVC、读取 Dataset、
   为自己的 Pod 添加标签或创建 Tier 2 操作，却获得了主机访问权限、逃出
   Pod，或读取了未获授权的租户数据。
2. 通过 Tier 2 字段注入。经由 `DataLoad` 字段（`options`、`target[].path`）
   或 DataBackup 的 `backupPath` 进行命令或参数注入，超出预期的加载或备份
   操作。
3. 通过 webhook 获得特权。注入绕过使用者请求的非特权模式，或使特权突破
   已配置的 PSA 边界而通过准入。
4. 跨命名空间或跨租户泄漏。Runtime 或 Dataset 未经管理员授权，暴露其他
   命名空间的 Secret、卷或缓存数据，或引用 Dataset 逃出其声明的子路径。
   管理员创建引用，或授权复制 PV 节点发布 Secret，均属于对指定数据或
   Secret 的显式授权。
5. 控制面成为混淆代理。控制器、webhook 或 CSI 插件利用其权限（创建/删除
   PV、`pods/exec`、复制 Secret、补丁更新节点），执行请求角色无权通过
   Fluid API 请求执行的操作。
6. 节点插件边界被突破。Pod 获得未授权的卷、逃出 FUSE 挂载根目录，或损坏
   其他 Pod 的挂载。
7. 租户间主机路径冲突。租户可以预测 FUSE 挂载路径、缓存路径或 sidecar
   主机挂载中的冲突，并利用它们执行超出授权范围的操作。
8. 默认配置回归。默认 `helm install` 向非管理员授予 Tier 0/1 动词，未经
   授权开放特权操作，或将注入 webhook 的 `failurePolicy` 改为 `Ignore`。
9. 常规控制面漏洞：未经身份认证的 webhook 路径、不可信输入可触发的崩溃、
   TLS 或凭据处理缺陷，以及存在可触达路径的依赖 CVE。

### 范围外

1. 受信任的 Tier 0/1 写入者使用资源授权本就提供的节点级或工作负载权限
   （§3 和 §4）。这不意味着访问无关租户数据或未授权 Secret 也被排除在
   漏洞范围外。
2. 仅限管理员配置的工作负载字段缺少校验：`spec.fuse.command`、
   `spec.fuse.args`、`spec.volumes`、`spec.tieredStore.levels[].path`、
   `fluid.fuse.hostvolume`、`runtime.fluid.io/fuse.hostpid` 及等价字段，且影响
   限于写入者获授的权限范围。
3. 非默认选项执行了文档所述、经管理员授权的操作。例如，管理员启用复制
   策略并授予所需 RBAC 后，复制 PV 引用的 Secret。仅启用选项并不授权
   超出该操作范围的行为。
4. 部署向不可信租户授予 Tier 0/1 写权限，且仅依赖命名空间隔离（S3）；或
   允许不可信的 sidecar 使用者创建特权 Pod（S4）。
5. 管理员将凭据存为明文挂载选项而非 Secret 引用，导致读取 Dataset 时可以
   看到凭据。
6. 特权工作负载读取管理员放置在同节点上的节点本地缓存（§6.4）。

欢迎将范围外报告作为普通 issue 提交，用于加固、修正文档或改进默认配置。

## 8. 报告方式

按照 [SECURITY.md](../SECURITY.md) 进行私密披露。报告应以 `Role`/`ClusterRole`
给出攻击者所需的最小 RBAC，说明攻击者命名空间的 PSA 级别，以及涉及的角色
（§2）和资源级别（§4）。说明默认 `helm install fluid fluid/fluid` 能否复现，
并列出必需的非默认设置或集群级资源。

§7 范围内的第 1、2、3 类报告会优先处理。

## 9. 验证

§5 的属性应有自动化检查。分别使用两种 `helmDriver` 设置渲染 Helm chart，
检查 RBAC 是否存在通配的 apiGroups、资源或动词，以及非预期的集群范围 Secret
授权。验证两个注入 webhook 均渲染出 `failurePolicy: Fail` 和预期的选择器。

单元测试应覆盖引用 Dataset 子路径逃逸、CSI 符号链接处理，以及 DataLoad 和
DataProcess 对跨命名空间目标的拒绝；其中部分测试已存在。Tier 2 选项的渲染
还需要单元测试或模糊测试。集成测试应验证 PSA 对变更后完整 Pod 的判定，以及
CSI 使用者布局，不能假定非特权模式会通过 PSA。

缺少测试属于覆盖缺口，不会取消本文所述的安全边界承诺。
