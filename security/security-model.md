# Fluid Security Model

This document describes Fluid's trust boundaries, who may create and use its
resources, and the behavior we consider a security vulnerability. The security
properties below apply to deployments that follow this model. Other deployment
configurations need their own security assessment.

Companion documents: [self-assessment.md](self-assessment.md) (CNCF TAG-Security
self-assessment), [SECURITY.md](../SECURITY.md) (vulnerability reporting), and
[security-model_zh.md](security-model_zh.md) (Chinese version).

Code references use paths and, where useful, function names. A reference that
no longer resolves is a documentation bug.

## 1. Why Fluid workloads are privileged

Fluid makes remote data available as a filesystem inside a Pod. Its standard
node-level FUSE workloads use `privileged: true` (FUSE itself needs at least
`CAP_SYS_ADMIN` and `/dev/fuse`) and a `hostPath` volume with
`mountPropagation: Bidirectional`. This lets a mount created inside the FUSE
container appear in the host mount namespace and become available to other
Pods on the node.

These workloads run in the physical Dataset's namespace. Runtime APIs with
`spec.fuse.networkMode` default to `HostNetwork`; set
`spec.fuse.networkMode: ContainerNetwork` to use container networking
(`api/v1alpha1/container_network.go`). `CacheRuntime` uses a class-provided
client Pod template instead: its privilege and host networking depend on
that template (§4). Sidecar mode has a separate non-privileged option, with
limitations described in S4.

Permission to configure these FUSE workloads carries node-level privilege.
Operators must know which Fluid resources expose that permission before
delegating access.

## 2. The actors

| Actor | Trust level | Responsibility |
| --- | --- | --- |
| Cluster Administrator | Full cluster authority | Installs Fluid, manages cluster-scoped Fluid resources and PersistentVolumes, decides which node pools run cache workloads, owns Pod Security Admission (PSA) labels |
| Data Administrator | Trusted with node-level privilege where their workloads run | Creates `Dataset` and Runtime resources, including reference Datasets in consumer namespaces; holds storage credentials; tunes cache topology and host-level FUSE options |
| Data Consumer | Ordinary tenant, untrusted | Runs application Pods that mount a Dataset's PVC; optionally reads Dataset status |
| Fluid Control Plane | Cluster-privileged system component | Controllers, webhook and CSI plugin in `fluid-system`, acting on behalf of the other three |

The Data Administrator is as trusted as someone allowed to create privileged
Pods. The Data Consumer is outside this trust boundary and may be hostile.
Namespace-scoped authorization does not limit a privileged workload's effects
to that namespace.

## 3. The basic rule

> Granting write access to a Runtime that launches a privileged FUSE workload
> must be treated as granting the ability to run privileged Pods with host
> paths and bidirectional mount propagation on its nodes.

For Runtime APIs with a FUSE spec, writers choose the image (`spec.fuse.image`),
entrypoint (`spec.fuse.command`, `spec.fuse.args`), environment and volumes.
These fields let data administrators configure the client their storage
system needs. `CacheRuntime` exposes client settings such as
`spec.client.runtimeVersion.image` and `spec.client.args`; the class's Pod
template supplies the security context. It remains administrator-only, and
its workload authority must be assessed against that template (§4).

`Dataset` write access also belongs to administrators, but for different
reasons. A physical Dataset needs a matching Runtime to launch its FUSE
workload. Its mount options can configure that workload, and its Secret
references select storage credentials. A reference Dataset authorizes access
to another namespace's data (§6.2). Writing a Dataset alone is therefore not
an unconditional equivalent of creating an arbitrary privileged Pod.

Review Runtime write requests with the same care as privileged-Pod creation
requests. Keep both Runtime and Dataset write permissions away from untrusted
tenants; the resource tiers in §4 explain what each grant allows.

Validating `command` alone cannot contain someone who can replace the image.
RBAC determines who may configure privileged workloads. Fluid does not ship a
validating webhook for these fields. Administrators can add admission policies
for defense in depth, but a partial field policy does not replace the access
boundary (§6, S3).

Namespaces hosting privileged FUSE workloads need
`pod-security.kubernetes.io/enforce: privileged` unless an equivalent PSA
exemption is configured. They are not tenant sandboxes. Keep untrusted
consumers in separate namespaces with enforced Pod restrictions (§6, S2).

## 4. Resource tiers

Fluid ships 14 user-facing `data.fluid.io` CRDs and one internal
`workload.fluid.io` CRD. The tiers describe the authority that writing each
resource grants.

### Tier 0: Cluster Administrator only (cluster-scoped)

| Resource | Why |
| --- | --- |
| `CacheRuntimeClass` | `topology.{master,worker,client}.template` is a full `corev1.PodTemplateSpec` deep-copied into the rendered workload (`api/v1alpha1/cacheruntimeclass_types.go`; `pkg/ddc/cache/engine/transform_common.go`). The writer sets `privileged`, `hostPID`, `hostNetwork`, `capabilities`, `serviceAccountName` and any `hostPath`. The class also defines `dataOperationSpecs[].image`. This is the broadest privilege in the Fluid API |
| `ThinRuntimeProfile` | Supplies FUSE image, command, args, volumes and `hostNetwork` for every `ThinRuntime` that references it. `spec.nodePublishSecretPolicy` controls Secret handling for `pvc://` mounts (§5) |

### Tier 1: Data Administrator (namespaced, privileged or sensitive configuration)

| Resource | Why |
| --- | --- |
| `AlluxioRuntime`, `JuiceFSRuntime`, `JindoRuntime`, `ThinRuntime`, `EFCRuntime`, `VineyardRuntime`, `CacheRuntime` | Configure node-level FUSE workloads. Standard privileged workloads follow §3; `CacheRuntime` configures client fields on a class-provided Pod template, whose security context determines the workload's privilege (`api/v1alpha1/cacheruntime_types.go`) |
| `Dataset` | Holds `mountPoint` and `encryptOptions` (Secret references in its namespace). `spec.mounts[].options` configure the FUSE workload: `fluid.fuse.hostvolume` injects a `hostPath` (`pkg/ddc/thin/transform_fuse.go`), and a `local://` mountPoint becomes a `hostPath` on Jindo. A `dataset://ns/name` mountPoint defines a reference Dataset that exposes another namespace's data (§6.2). A physical Dataset needs a matching Runtime to launch a FUSE workload, but its credential references and data-access scope already require administrator trust |
| `DataProcess` | `spec.processor.job.podSpec` is a full `corev1.PodSpec`, and `spec.processor.serviceAccountName` is settable (`api/v1alpha1/dataprocess_types.go`). Equivalent to creating an arbitrary Pod under an arbitrary ServiceAccount in the namespace |
| `DataMigrate` | Inlines `VersionSpec` (`api/v1alpha1/datamigrate_types.go`), so the image is caller-chosen, and `encryptOptions` reference Secrets |

### Tier 2: delegatable to trusted users

| Resource | What keeps delegation reasonable | Remaining input surface |
| --- | --- | --- |
| `DataLoad` | No caller image override, `podSpec` or `volumes`; the runtime chooses the image (`pkg/utils/docker/image.go`, `GetWorkerImage`). For CacheRuntime, the class can supply an operation image through `dataOperationSpecs[].image` (`pkg/ddc/cache/engine/dataload.go`) | `spec.options` and `spec.target[].path` feed the runtime's loader script. JuiceFS joins options into a `--key=value` argument string (`pkg/ddc/juicefs/data_load.go`, `genDataLoadValue`). Argument or command injection that lets these inputs exceed the intended operation is in scope (§7) |
| `DataBackup` | No arbitrary Pod definition, but safe delegation requires restrictions on destination and execution identity | Alluxio local backup uses caller-specified `spec.backupPath` as a writable `hostPath` on the master node, and `spec.runAs` can override the Runtime identity (`pkg/ddc/alluxio/backup_data.go`). Keep this operation administrator-only unless admission policy constrains these fields. Injection beyond the authorized backup remains in scope |

Both drive I/O and consume capacity. Cache-operation trust does not include
permission to run arbitrary commands or access unrelated data. In particular,
DataBackup is not safe to delegate merely because it has no `podSpec`;
namespace RBAC does not constrain its backup destination. The role example
in §6.3 grants DataLoad writes only.

### Tier 3: Data Consumer

| Action | Permission required |
| --- | --- |
| Mount a Dataset | No Fluid RBAC. Reference the PVC Fluid provisions in the application's namespace, named after the Dataset |
| Observe cache state | `get`/`list`/`watch` on `datasets`; tools using the status endpoint may need `get` on `datasets/status`. Either read endpoint exposes the full Dataset object, including spec (§5) |

### Internal

`AdvancedStatefulSet` (`workload.fluid.io`) belongs to the control plane.
Users should not write it directly.

### Delegation summary

| Actor | May create / write | May use |
| --- | --- | --- |
| Cluster Administrator | Tier 0; Fluid install; PVs; node pools; PSA labels | Everything |
| Data Administrator | Tier 1 in authorized namespaces, including reference Datasets in consumer namespaces | Credentials referenced by those resources; cache topology; host-level FUSE knobs (`networkMode`, `fluid.fuse.hostvolume`, `runtime.fluid.io/fuse.hostpid`, `tieredStore` paths) |
| Trusted cache operator | Tier 2 against exposed Datasets; DataBackup writes only after destination and identity restrictions are enforced | Read on those Datasets and operations |
| Data Consumer | No `data.fluid.io` resource writes | The provisioned PVC; optionally read Dataset state |

Consuming data does not require permission to change the resources that
provision it.

### A note on `config/rbac/`

The kubebuilder-scaffolded `*-editor-role.yaml` / `*-viewer-role.yaml` files
under `config/rbac/` are not referenced by `config/rbac/kustomization.yaml`
or the Helm chart. Those installation paths do not create these roles. Some
grant write on Runtime kinds, which is an administrator-level grant. Do not
apply them to tenants.

## 5. Security properties and their limits

These properties define the expected boundaries. The qualifications below
matter when selecting a Runtime and configuring a deployment.

### Credential lookup and handling

Mount-level `encryptOptions` and Dataset-level `sharedEncryptOptions` resolve
Secret references in the Dataset's namespace
(`pkg/ddc/alluxio/ufs_internal.go`, `pkg/ddc/thin/ufs_internal.go`). Plain values
in `spec.mounts[].options` or `spec.sharedOptions` are visible to anyone who
can read the Dataset. Use Secret references for credentials.

Secret references keep values out of the Dataset spec. They do not guarantee
that every Runtime keeps those values out of generated configuration.
JuiceFS uses `env.valueFrom.secretKeyRef`. Alluxio and Thin have Secret-volume
paths, but alternative configuration modes can resolve values into generated
configuration or Helm values. Older Jindo paths also render resolved values
into configuration; JindoCache and JindoFSx behavior depends on support for
Secret mounts in the chosen image version. Check the Runtime, configuration
mode and image version before treating its configuration as safe to disclose
(`pkg/ddc/juicefs/transform_fuse.go`, `genValue`;
`pkg/ddc/alluxio/transform.go`, `transformCommonPart`;
`pkg/ddc/thin/transform_config.go`, `extractMountOptions`;
`pkg/ddc/jindo/transform.go`, `transformMaster`).

Kubernetes CRD status reads return the whole object. `get` on
`datasets/status` is not a status-only, field-filtered permission. If consumers
must not see Dataset spec, do not grant either read endpoint; provide a
separately filtered observability interface.

### Explicit cross-namespace Secret copying

With `ThinRuntimeProfile.spec.nodePublishSecretPolicy:
CopyNodePublishSecretAndMountIfNotExists`, the ThinRuntime controller copies
the `nodePublishSecretRef` of the PV behind a `pvc://` mount into the Runtime's
namespace as `<runtime>-thin-publish-secret`
(`pkg/ddc/thin/transform_pvc_mounts.go`,
`transfromSecretsForPersistentVolumeClaimMounts`;
`pkg/utils/kubeclient/secret.go`, `CopySecretToNamespace`).

The source Secret comes from the cluster-scoped PV's reference; an empty Secret
namespace resolves to `default`. The PVC must be in the Dataset's namespace.
This limits the lookup to a PV already bound to a PVC in that namespace.
The default Helm configuration does not grant `thinruntime-controller` the
Secret API permissions needed for copying. An administrator must grant those
permissions explicitly as well as selecting the copy policy.

The default policy, `MountNodePublishSecretIfExists`, mounts an existing
Secret by name without copying it. This subsection describes the PV
node-publish Secret copy path, not all possible Runtime credential flows.

### Reference Dataset and CSI mount boundaries

A `dataset://ns/name/subpath` reference must refer to exactly one physical
Dataset. It cannot mix mount schemes or chain references. Its subpath cannot
escape the physical Dataset's root (`pkg/ddc/thin/referencedataset/engine.go`,
`checkDatasetMountSupport`; `pkg/ddc/base/dataset.go`,
`GetPhysicalDatasetSubPath`).

For a non-empty subpath on Linux, the default `bindMount` publish path uses
`resolveMountSource` to walk components with `O_NOFOLLOW`, reject symlinks and
pin the checked inode with a file descriptor
(`pkg/csi/plugins/subpath_linux.go`, `resolveMountSource`). A symlink beneath
the FUSE mount cannot redirect that subpath bind to an arbitrary host path.
`NodePublishVolume` invokes `mount` through `cmdguard`
(`pkg/csi/plugins/nodeserver.go`, `NodePublishVolume`).

The non-default `symlink` publish path creates a link from the original path
string and does not use the pinned inode. It does not provide the same
pinning guarantee. S2 uses `bindMount`; verify both `csi.nodePublishMethod`
and any PV `node_publish_method` override before relying on that guarantee.

### Data-operation targets

`DataLoad` and `DataProcess` reject a target Dataset in another namespace
(`pkg/controllers/v1alpha1/dataload/implement.go`,
`pkg/controllers/v1alpha1/dataprocess/implement.go`).

### Control-plane permissions

Components have separate ServiceAccounts and roles. The Helm chart's RBAC
rules do not use wildcard apiGroups, resources or verbs.

The default Helm setting is `helmDriver: configmap`. Runtime and Dataset
controller rules using `fluid.helmDriver.rbacs` receive cluster-wide Secret
API permissions only with `helmDriver: secret`
(`charts/fluid/fluid/templates/_helpers.tpl`). The webhook has separate,
namespace-scoped certificate Secret permissions. Non-Helm manifests may grant
different permissions; inspect the rendered RBAC for the installation path
actually used. Absence of direct Secret API grants does not make the Fluid
control plane unprivileged.

With `csi.useNodeAuthorization: true` (default), the CSI plugin uses kubelet
credentials for its Node reads and patches. Other API operations still use
its ServiceAccount. The cluster must enable the Node authorizer and
NodeRestriction admission plugin for their protections to apply
(`pkg/csi/plugins/nodeserver.go`, `getNode`, `patchNodeWithLabel`).

### Admission and sidecar privilege

The Helm-installed `sidecar.fuse.fluid.io` and `fuse.serverful.fluid.io`
webhooks use `failurePolicy: Fail`. Invocation errors reject matching Pod
CREATE requests (`charts/fluid/fluid/templates/webhook/webhookconfiguration.yaml`).
Their label selectors limit which Pods are covered; this is not an admission
policy for all Pods or for Fluid CRDs.

Mutating admission runs before PSA validation. A privileged sidecar injected
into a Pod is therefore rejected by enforced `baseline` or `restricted` PSA,
unless the request is exempt. This remains a cluster-enforced boundary.

The serverless path offers an unprivileged sidecar mode. With both
`serverless.fluid.io/inject: "true"` and
`unprivileged.sidecar.fluid.io/inject: "true"` labels, the mutator sets
`privileged: false`, removes `SYS_ADMIN` from `Capabilities.Add`, and removes
the known host FUSE mount/device volumes from the template
(`pkg/application/inject/fuse/mutator/mutator_unprivileged.go`,
`transformTemplateWithUnprivilegedSidecarEnabled`). It does not add
`SYS_ADMIN` to `Capabilities.Drop` or remove every possible host volume.

This mode does not guarantee PSA compliance. In particular, the current
mutation path still converts the application's Dataset PVC volume to
`hostPath` (`pkg/application/inject/fuse/mutator/mutator_default.go`,
`defaultMutateDatasetVolumes`). Standard `baseline` and `restricted` PSA
reject `hostPath`. See S4 before choosing this mode for untrusted consumers.

## 6. Deployment scenarios

### S1: single tenant / trusted team

All users are trusted with the relevant storage and node privileges. The
default installation requires no additional tenant RBAC separation.

### S2: multi-tenant, administrator-provisioned datasets (recommended)

This model supports untrusted consumers through the default CSI `bindMount`
path, with administrator-provisioned data and restricted consumer Pods. The
administrator must also verify that configuration copied into consumer
namespaces contains no credentials or data beyond the intended grant (§6.2).

#### 6.1 Layout

| Namespace | PSA `enforce` | Contents | Who writes |
| --- | --- | --- | --- |
| `fluid-system` | `privileged` | Fluid control plane | Cluster Administrator |
| `fluid-datasets` | `privileged` | Physical `Dataset` + Runtime, cache master/worker/FUSE | Data Administrator |
| `team-a`, `team-b`, ... | `restricted` | Reference `Dataset` (admin-created) + consumer Pods | Admin: Dataset; consumer: Pods only |

Consumers receive no write verb on any `data.fluid.io` resource.

#### 6.2 Cross-namespace access via reference Datasets

Kubernetes PVCs cannot be referenced across namespaces. An administrator
creates a reference Dataset in the consumer namespace:

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

Fluid provisions a PVC `imagenet` in `team-a`. Consumers mount that local PVC.
Creating the reference grants access to the physical Dataset or the selected
subpath and is an administrator authorization decision. Read/write access is
a separate setting: references default to `ReadOnlyMany`, and the physical
PV's access modes constrain the allowed modes
(`pkg/ddc/thin/referencedataset/volume.go`, `accessModesForVirtualDataset`).
Where read-only access is required, explicitly set Dataset
`spec.accessModes: [ReadOnlyMany]` and the consumer's volume mount
`readOnly: true`; use read-only backend credentials where applicable.

With CSI, the consumer receives a bind mount served by the physical FUSE
workload in `fluid-datasets`. Its namespace can remain `restricted`.

Fluid also copies the physical FUSE DaemonSet into the reference namespace,
owned by the reference Dataset, for sidecar-template use
(`pkg/ddc/thin/referencedataset/cm.go`, `copyFuseDaemonSetForRefDataset`).
The copy adds `nodeSelector: {fluid.io/fuse-balloon: "true"}` to avoid
scheduling it. Administrators must not put that reserved label on nodes.
The copied template retains the original privilege settings; sidecar
injection does not make it compliant with consumer PSA (§5, S4).

Reference setup also copies Runtime configuration ConfigMaps into the consumer
namespace, even when consumers use CSI
(`pkg/ddc/thin/referencedataset/cm.go`, `createConfigMapForRefDataset`). Some
legacy or alternative Runtime configurations contain resolved credentials
(§5). A consumer allowed to create Pods can mount a ConfigMap in its own
namespace without ConfigMap read RBAC; `restricted` PSA permits that volume
type. Withholding Dataset reads therefore does not protect copied credentials.
Before exposing a reference, verify the copied configuration. A Runtime mode
that copies backend credentials exceeding the data grant does not meet S2's
assumptions. Authorizing a Dataset or subpath reference does not by itself
authorize disclosure of its backend credentials.

#### 6.3 Consumer roles

Dataset reads are optional for observability and are unnecessary for mounting.
Only grant them if exposing the full spec is acceptable, and keep credentials
out of plain mount options:

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

If a tool uses the status endpoint, add the following rule. It also allows
reading the full object; it does not hide spec:

```yaml
- apiGroups: ["data.fluid.io"]
  resources: ["datasets/status"]
  verbs: ["get"]
```

Trusted cache operators (Tier 2):

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

This role deliberately omits DataBackup writes. Add them only after enforcing
allowed backup destinations and `spec.runAs` identities. Otherwise keep
backup creation with the data administrator (§4).

Do not grant untrusted tenants writes on `alluxioruntimes`, `juicefsruntimes`,
`jindoruntimes`, `thinruntimes`, `efcruntimes`, `vineyardruntimes`,
`cacheruntimes`, `datasets`, `dataprocesses`, `datamigrates`,
`thinruntimeprofiles` or `cacheruntimeclasses`.

#### 6.4 Node-local cache data

Disk-backed tiered-store cache (`spec.tieredStore.levels[].path`) and FUSE
mount points use node `hostPath` storage. Other privileged workloads on the
same node can read these paths. Fluid does not isolate cache contents from
those workloads.

For Datasets of different sensitivity, assign each sensitivity domain its
own node pool, give Runtimes distinct `tieredStore` paths, and enforce node
placement administratively. Prefer `MEM` tiers when disk cache residue is a
concern; separate paths alone do not provide isolation.

### S3: self-service tenants with their own Datasets and Runtimes

Runtime write permission gives tenants node-level privilege. Namespace RBAC
cannot contain it. For mutually untrusted tenants, use separate clusters or
administratively enforced data-plane isolation, such as dedicated tenant
nodes. Placement constraints must be enforced by the host administrator's
admission policy and node configuration, not left to tenant-chosen
`nodeSelector` or tolerations.

A virtual control plane such as vCluster does not by itself isolate privileged
workloads on shared host nodes. It also needs an isolated data plane.

Admission policies can restrict image overrides for defense in depth. The
following Kubernetes v1.30+ example checks the repository prefix of an
explicit ThinRuntime FUSE image override. The Binding makes it enforce the
check with `Deny`:

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

This example allows omitted images and does not validate the image inherited
from a ThinRuntimeProfile, image tags, other Runtime types, or Dataset
`fluid.fuse.hostvolume` options. Administrators must control inherited image
sources and add policies for any other fields they intend to restrict. This
partial policy does not turn Runtime write access into a safe tenant grant.

### S4: serverless / sidecar injection

The serverless webhook path injects a FUSE sidecar into Pods labeled
`serverless.fluid.io/inject: "true"`. The Pod creator selects the label and
can request unprivileged mode with the additional
`unprivileged.sidecar.fluid.io/inject: "true"` label.

The default sidecar is privileged. Do not allow untrusted consumers to obtain
it simply by labeling a Pod in a namespace that admits privileged workloads.
PSA must validate the complete mutated Pod, and administrators must not exempt
untrusted consumers from those checks.

The current unprivileged mutator still introduces a Dataset `hostPath` volume,
so it is also rejected by standard `baseline` or `restricted` PSA. Removing
the sidecar's privileged flag does not provide the CSI-based isolation of S2.
For untrusted consumers, use S2 rather than weakening PSA to make sidecar
injection succeed. Environments that use sidecar mode need a separate security
assessment of their device support and complete mutated Pod.

The separate `fuse.serverful.fluid.io/inject: "true"` selector uses the
serverful scheduling path. The unprivileged label alone does not switch that
path to the serverless unprivileged mutator (`pkg/utils/annotations.go`,
`GetServerlessPlatform`; `pkg/webhook/handler/mutating/mutating_handler.go`).

The deprecated `fuse.sidecar.fluid.io/inject` label remains recognized by the
handler, but alone does not match either default Helm webhook selector. New
deployments should use the current labels.

## 7. Triage: what we consider a vulnerability

### In scope

The following apply within the deployment assumptions above:

1. Escalation without Tier 0/1 write access. A principal limited to mounting a
   PVC, reading a Dataset, labeling its own Pods, or creating a Tier 2
   operation obtains host access, escapes its Pod, or reads unauthorized
   tenant data.
2. Injection through Tier 2 fields. Command or argument injection via
   `DataLoad` fields (`options`, `target[].path`) or DataBackup `backupPath`
   exceeds the intended loader or backup operation.
3. Webhook-mediated privilege. Injection bypasses the unprivileged mode a
   consumer requested or causes privileges to be admitted despite the
   configured PSA boundary.
4. Cross-namespace or cross-tenant leakage. A Runtime or Dataset exposes
   another namespace's Secret, volume or cached data without administrator
   authorization, or a reference Dataset escapes its declared subpath.
   An administrator-created reference or authorized PV node-publish Secret
   copy is explicit authorization for its specified data or Secret.
5. Confused deputy in the control plane. A controller, webhook or CSI plugin
   uses its authority (PV create/delete, `pods/exec`, Secret copy, node patch)
   for an action beyond what the requesting actor is authorized to obtain
   through the Fluid API.
6. Node plugin boundary violations. A Pod obtains an unauthorized volume,
   escapes the FUSE mount root, or corrupts another Pod's mount.
7. Host-path collisions between tenants in FUSE mount paths, cache paths or
   sidecar host mounts that a tenant can predict and abuse beyond its grant.
8. Defaults regressions. A default `helm install` grants non-administrators a
   Tier 0/1 verb, exposes a privileged operation without authorization, or
   changes the injection webhooks' `failurePolicy` to `Ignore`.
9. Ordinary control-plane vulnerabilities: unauthenticated webhook paths,
   crashes triggerable by untrusted input, TLS or credential handling flaws,
   and dependency CVEs with a reachable path.

### Out of scope

1. A trusted Tier 0/1 writer uses the node-level or workload authority their
   resource grant intentionally provides (§3 and §4). This does not exclude
   unrelated tenant data or unauthorized Secret access.
2. Missing validation on administrator-only workload configuration fields:
   `spec.fuse.command`, `spec.fuse.args`, `spec.volumes`,
   `spec.tieredStore.levels[].path`, `fluid.fuse.hostvolume`,
   `runtime.fluid.io/fuse.hostpid`, and equivalents, where the effect stays
   within the writer's granted authority.
3. A non-default option performs its documented, administrator-authorized
   operation, such as copying a PV-referenced Secret after the administrator
   enables the copy policy and grants the required RBAC. Merely enabling an
   option does not authorize behavior beyond that operation.
4. A deployment gives untrusted tenants Tier 0/1 writes and relies only on
   namespace isolation (S3), or allows untrusted sidecar consumers to create
   privileged Pods (S4).
5. Credentials are visible in Dataset reads because the administrator stored
   them as plain mount options instead of Secret references.
6. A privileged workload reads node-local cache that the administrator
   co-located on the same node (§6.4).

Out-of-scope reports are welcome as regular issues for hardening,
documentation fixes or safer defaults.

## 8. Reporting

Follow [SECURITY.md](../SECURITY.md) for private disclosure. Include the
attacker's minimum RBAC as a `Role`/`ClusterRole`, their namespace's PSA level,
and the actor (§2) and resource tier (§4) involved. State whether the default
`helm install fluid fluid/fluid` reproduces the issue, and list any required
non-default settings or cluster-scoped resources.

Reports in §7 in-scope classes 1, 2 and 3 receive priority.

## 9. Verification

The properties in §5 should have automated checks. Render the Helm chart
under both `helmDriver` settings and check that RBAC has no wildcard
apiGroups, resources or verbs, and no unexpected cluster-wide Secret grants.
Verify that both injection webhooks render `failurePolicy: Fail` and the
expected selectors.

Unit tests should cover reference Dataset subpath escape, CSI symlink
handling, and cross-namespace target rejection in DataLoad and DataProcess;
some tests already exist. Tier 2 option rendering also needs unit or fuzz
tests. Integration tests should verify PSA decisions on the fully mutated Pod
and the CSI consumer layout, rather than assuming non-privileged mode passes
PSA.

Missing tests are coverage gaps. They do not remove the boundary commitments
stated in this document.
