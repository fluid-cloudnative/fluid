# CacheRuntime Spec Field Update Capabilities

## 1. Overview

This document describes which `spec` fields of the CacheRuntime's Master and Worker components (backed by **AdvancedStatefulSet**) can be updated after creation, and how the CacheRuntimeClass a CacheRuntime uses is fixed when the CacheRuntime is set up.

The current version syncs the following three fields to existing components:
- **Container image** (`runtimeVersion`)
- **Resource limits** (`resources`)
- **Replica count** (`replicas`)

Modifications to other fields **will not be propagated to the AdvancedStatefulSet** and require redeployment to take effect.

---

## 2. Supported Components

| Component | Workload Type | Field Update Support |
|-----------|---------------|---------------------|
| **Master** | AdvancedStatefulSet | ✅ Supports `runtimeVersion`, `resources` and `replicas` |
| **Worker** | AdvancedStatefulSet | ✅ Supports `runtimeVersion`, `resources` and `replicas` |
| **Client** | DaemonSet | ❌ Not supported (any changes require redeployment) |

**Notes**:
- Modifying `runtimeVersion`, `resources` and `replicas` on Master and Worker components automatically propagates to the underlying AdvancedStatefulSet.
- The AdvancedStatefulSet controller bundled with Fluid only updates an image in place, restarting the container. A `resources` change recreates the Pods, which drops the cache held in memory and `emptyDir` volumes.
- The Client component uses a DaemonSet and does not support dynamic updates.

---

## 3. Supported Update Fields

### 3.1 Container Image (`runtimeVersion`)

**Field path**: `spec.{master,worker}.runtimeVersion`

**Supported sub-fields**:
- `image`: Image name
- `imageTag`: Image tag

**Example**:
```yaml
spec:
  worker:
    runtimeVersion:
      image: fluid-cache
      imageTag: v1.1.0
```

**Resolution**: on creation and on every reconcile, the image of the component's first container in the CacheRuntimeClass template the CacheRuntime is bound to (see section 4) is the baseline, and the `image` and `imageTag` set in the CacheRuntime override it.

**Completion rules**:
- `image` and `imageTag` may be set on their own. The missing half is taken from the template image. With a template image of `fluid-cache:v1.0.0`:
  - Only `imageTag: v1.1.0`: the template's image name is kept, giving `fluid-cache:v1.1.0`
  - Only `image: registry.example.com/fluid-cache`: the template's tag is kept (**not `latest`**), giving `registry.example.com/fluid-cache:v1.0.0`
  - Neither: the template image `fluid-cache:v1.0.0` is used
- A `runtimeVersion` setting both `image` and `imageTag` is used as is, even when the template image is pinned by digest.
- When only one half is set and the other cannot be completed (only `image` is set and the template image has no tag, or the template image or `image` is pinned by digest (`@sha256:...`)), the component uses the template image on creation; on update the edit has no effect and the component keeps its current image. In both cases the controller emits a `RuntimeVersionIncomplete` Warning event on the CacheRuntime and logs the edit; the client, which is not updated in place, reports it on creation only. `runtimeVersion` cannot express a digest reference; to pin a digest, put it in the CacheRuntimeClass template.
- Data operations that name no image of their own (such as DataLoad) use the same completed image as the worker. When the worker's version cannot be completed they use the template image, so they can differ from a worker that kept its current image, for example after the template image is changed to one pinned by digest.

**Limitations**:
- ⚠️ **When setting only `image`, make sure the template's tag exists in the new repository.** Otherwise the pods fail with `ImagePullBackOff`; on an update the component is unavailable until the new image can be pulled.
- ⚠️ **An image change restarts the containers** of the component: the AdvancedStatefulSet updates the image in place.
- ⚠️ **Image edits in the CacheRuntimeClass do not reach existing CacheRuntimes**: they are bound to the class as it was when they were set up (see section 4). To move an existing CacheRuntime to a new image, set `runtimeVersion` on it.
- ⚠️ Removing `runtimeVersion` from a CacheRuntime returns the component to the bound template image, except when that image has no tag or is pinned by digest: the component then keeps its current image, and you need to set a full `runtimeVersion` or recreate the CacheRuntime.

---

### 3.2 Resource Limits (`resources`)

**Field path**: `spec.{master,worker}.resources`

**Supported sub-fields**:
- `requests.cpu`: CPU request
- `requests.memory`: Memory request
- `limits.cpu`: CPU limit
- `limits.memory`: Memory limit
- `claims`: Dynamic Resource Allocation (DRA) claims; they can only reference pod-level `resourceClaims` the template declares, see the limitations below

**Example**:
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

**Resolution order**:

`resources` is resolved as follows on creation and on every reconcile:

1. the `resources` declared by the template of the CacheRuntimeClass the CacheRuntime is bound to (see section 4) form the baseline;
2. the `resources` set on the CacheRuntime are overlaid on that baseline key by key —
   `requests` and `limits` merge by resource name, `claims` by claim name;
3. a key neither declares is cleared from the workload.

The CacheRuntime therefore expresses only the differences it wants from the template. Naming
`limits.memory` alone moves that one value and leaves the template's `requests.cpu`,
`requests.memory` and `limits.cpu` in place.

Removing a key from the CacheRuntime works the same way: it falls back to the bound template's value,
or is cleared from the workload when the template does not declare it. Editing the CacheRuntimeClass
afterwards does not change that value. An empty list such as
`limits: {}` names no key, so it reads the same as leaving `limits` out.

**Limitations**:
- ⚠️ **A `resources` change recreates the component's Pods**: the bundled AdvancedStatefulSet controller does not resize containers in place, so the Pods are recreated and the cache they held in memory or `emptyDir` volumes is lost.
- ⚠️ Cannot exceed the node's available resources.
- ⚠️ A key the CacheRuntimeClass template declares can be **overridden but not removed**: omitting it
  from the CacheRuntime leaves the template's value in place rather than leaving the component
  unconstrained. To adjust resource requirements, explicitly set the desired values in the CacheRuntime.
- ⚠️ **Resources edits in the CacheRuntimeClass do not reach existing CacheRuntimes** (see section 4).
  To change an existing CacheRuntime, set the values on it.
- ⚠️ **The CacheRuntime and its bound template are the source of truth**: Fluid owns the master and
  worker container resources, so resources edited directly on the AdvancedStatefulSet are reverted
  on the next reconcile.
- ⚠️ **A request must not exceed its limit after the merge**: the template and the CacheRuntime
  can each be valid and still merge into a request above its limit, for example a template
  `requests.cpu: 1` with a CacheRuntime `limits.cpu: 500m`. Huge pages (`hugepages-*`) and
  extended resources (such as `nvidia.com/gpu`) are stricter: their request must equal their
  limit, and a request needs a limit, so a template setting `nvidia.com/gpu` to 1 for both with
  a CacheRuntime raising only the limit to 2 conflicts as well. The controller then does not sync
  that component's `resources`, keeps the workload's current values and emits a
  `RuntimeResourcesConflict` Warning event on the CacheRuntime naming the conflicting keys; a new
  CacheRuntime with such a pair fails to set up. Set the matching request on the CacheRuntime too
  (equal to the limit for huge pages and extended resources). Changing the CacheRuntimeClass only
  helps CacheRuntimes created afterwards.
- ⚠️ **`claims` can only reference pod-level `resourceClaims` the workload declares**: the CacheRuntime
  sets the container's `resources.claims` but cannot add pod-level `spec.resourceClaims`, and claims
  added to the CacheRuntimeClass after a CacheRuntime was set up do not reach it (see section 4). On creation, a claim name the template does not declare makes the API
  server reject every master or worker Pod (`must be one of the names in pod.spec.resourceClaims`), and
  the component stays NotReady; clusters without `DynamicResourceAllocation` enabled (off by default
  before K8s 1.34) drop `claims` silently instead. On update, a claim that the workload neither
  declares nor already carries is
  not synced: the controller leaves it out, still syncs the rest of `resources`, and emits a
  `RuntimeResourcesConflict` Warning event on the CacheRuntime naming the claim. To use a new claim,
  recreate the CacheRuntime.

#### Check resource synchronization status

Look for the condition with `type: ResourcesSynced` in the CacheRuntime's `.status.conditions`.
During setup, the controller checks the resource settings of every enabled component, including
the master, worker, and client. After setup, the condition tracks workload resources only for
the enabled master and worker components. To check Pod readiness separately, inspect each
component's `readyReplicas`.

| Status | Reason | Meaning |
| --- | --- | --- |
| `False` | `RuntimeResourceCreationConflict` | During setup, a conflict in the merged resource settings prevents component creation. |
| `False` | `RuntimeResourcesConflict` | During sync, a resource conflict or a claim the workload does not declare causes the controller to skip the affected settings. |
| `Unknown` | `WaitingForSync` | A previous creation conflict has been resolved. The corrected resource settings have passed validation and are awaiting the first resource sync. This clears the creation failure message even if Pods are still pending. |
| `True` | `RuntimeResourcesSynced` | All enabled master and worker workloads have the desired resource settings. |

In both conflict cases, the condition's `message` identifies the component and cause. The matching Warning event uses the reason `RuntimeResourcesConflict` in both phases.

For example, the following command waits for the CacheRuntime named `my-cache` to report that its resources are in sync.

```bash
kubectl wait --for=condition=ResourcesSynced=True cacheruntime/my-cache --timeout=5m
```

---

### 3.3 Replica Count (`replicas`)

**Field path**: `spec.{master,worker}.replicas`

**Example**:
```yaml
spec:
  worker:
    replicas: 3
```

```bash
kubectl patch cacheruntime my-cache --type='merge' -p '{"spec":{"worker":{"replicas":3}}}'
```

The new value is synced to the AdvancedStatefulSet's `spec.replicas`, which then performs the scaling.

**Limitations**:
- ⚠️ **Scaling in discards cached data**: the Worker Pods removed are deleted outright, and the data they had cached is not migrated to the remaining Workers — it has to be loaded from the underlying storage again. Scaling out leaves the existing cache untouched.
- ⚠️ Scaling writes no RuntimeCondition and emits no Kubernetes Event. Progress can be observed through `status.{master,worker}.{readyReplicas,desiredReplicas}` or the controller logs:
  ```bash
  # Current replica status
  kubectl get cacheruntime my-cache -o jsonpath='{.status.worker.readyReplicas}/{.status.worker.desiredReplicas}'

  # Or from the controller logs
  kubectl -n fluid-system logs deploy/cacheruntime-controller | grep "replicas changed"
  ```
- ⚠️ If the replica count exceeds the number of schedulable nodes, the surplus Worker Pods stay `Pending`.

## 4. CacheRuntimeClass Snapshot

When a CacheRuntime is first set up, Fluid copies the CacheRuntimeClass it names into an immutable
ConfigMap `fluid-runtime-class-snapshot-<runtime name>` in the CacheRuntime's namespace, owned by the
CacheRuntime. From then on the CacheRuntime reads its class only from that snapshot: component
templates, `options`, `executionEntries` (such as the mountUFS and reportSummary commands),
`dataOperationSpecs` and `extraResources`.

On new snapshots, the ConfigMap annotations `cacheruntime.fluid.io/runtime-class-generation` and
`cacheruntime.fluid.io/runtime-class-resource-version` record the generation and resource version
of the CacheRuntimeClass that Fluid copied. Older snapshots may lack these annotations.

The whole class is fixed, not just part of it, because these fields are written for one another:
following some of them live would hand new commands or options to components still running the old
images. And a CacheRuntimeClass is cluster-scoped, so following it live would let one edit restart
or recreate the components of every CacheRuntime using it, in every namespace.

- **Editing or deleting a CacheRuntimeClass** only affects CacheRuntimes created afterwards. Existing
  CacheRuntimes keep working even if their class is deleted. To move one to the new class, recreate it.
- **Changing `runtimeClassName`** on an existing CacheRuntime has no effect: it stays bound to its
  snapshot, and Fluid emits a `RuntimeClassChanged` Warning event. Recreate the CacheRuntime to use
  another class.

### Recovering an invalid snapshot

If Fluid cannot decode the snapshot, it leaves the ConfigMap intact and emits a
`RuntimeClassSnapshotInvalid` Warning event. Reconciliation and data operations such as DataLoad
wait for a valid snapshot. **Delete the snapshot ConfigMap only as a last resort to recover from
a damaged or unreadable snapshot.** Fluid rebuilds it from the current version of the
CacheRuntimeClass named by `spec.runtimeClassName`.

If that class has changed since setup, recovery can leave components using different class versions.

- **Master and worker workloads.** Fluid reconciles the main container's image and `resources` using
  the replacement snapshot and the CacheRuntime's overrides. An image change restarts the container.
  A resource change recreates the Pods. All other PodTemplate fields, including `args`, `env`, and
  `volumes`, keep their existing values.
- **Client workload.** Every field of the existing client DaemonSet stays as deployed, including
  its image and PodTemplate.
- **Configuration and operations.** Fluid regenerates the runtime configuration from the replacement
  snapshot's `options` and the CacheRuntime's overrides, including options for the client.
  Subsequent controller commands use its `executionEntries`, and new data operation jobs such as
  DataLoad use its `dataOperationSpecs`. New commands or options can therefore reach an old image.

**To adopt a newer CacheRuntimeClass, recreate the CacheRuntime** so that all components are created
from the same class version. Do not use snapshot deletion as an upgrade procedure.

```bash
# The class a CacheRuntime is bound to
kubectl get cm fluid-runtime-class-snapshot-my-cache -o jsonpath='{.data.runtimeclass\.json}'
```

---

## 5. Unsupported Update Fields

**Any field not listed in section 3 cannot be updated after creation**; changes to it are not propagated to the AdvancedStatefulSet.

**Notes**:
- Such a field takes effect only after the CacheRuntime is redeployed.
- The system does not automatically detect or propagate these changes (except `spec.worker.tieredStore` pod-volume drift below).

### 5.1 Worker Tiered Storage (`spec.worker.tieredStore`)

**Field path**: `spec.worker.tieredStore`

- **Why in-place updates are not supported**: Tiered storage volumes (`emptyDir.sizeLimit` / `emptyDir.medium`, memory tmpfs `/dev/shm`, `hostPath` mounts / types) and container memory adjustments are generated during initial creation (`Setup`). In Kubernetes, pod volume definitions cannot be modified dynamically in running pods.
- **Drift & validation reporting**: Fluid checks `spec.worker.tieredStore` against the worker PodSpec during reconciliation via `TransformRuntimeTieredStore`:
  - **Volume drift** (editing `processMemory.quota`, `emptyDir.quota` / `medium`, `hostPath.paths` / `type`, or adding/removing levels): sets condition `TieredStoreSynced=False` (`Reason: RuntimeTieredStoreConflict`) and emits a `RuntimeTieredStoreConflict` Warning event on transition:
    ```bash
    kubectl get events -n <namespace> --field-selector type=Warning,reason=RuntimeTieredStoreConflict
    ```
  - **Invalid spec** (such as multiple media on one level or multiple `processMemory` levels): sets condition `TieredStoreSynced=False` (`Reason: RuntimeTieredStoreInvalid`) and emits a `RuntimeTieredStoreInvalid` Warning event instructing you to fix `spec.worker.tieredStore`.
  - **ConfigMap-only fields**: Edits to `hostPath.quotas`, `high`, and `low` do not alter worker pod volumes, so they update the runtime ConfigMap without triggering pod-volume drift warnings.
- **How to apply changes**: Changes to `spec.worker.tieredStore` pod-volume settings require deleting and recreating the CacheRuntime to take effect.

---

## 6. Summary

| Field | Supported | Update Method |
|-------|-----------|---------------|
| `runtimeVersion` | ✅ Yes | Automatically synced to AdvancedStatefulSet (restarts the containers) |
| `resources` | ✅ Yes | Automatically synced to AdvancedStatefulSet (recreates the Pods) |
| `replicas` | ✅ Yes | Automatically synced to AdvancedStatefulSet (scaling in discards that Worker's cache) |
| `spec.worker.tieredStore` | ❌ No | Sets `TieredStoreSynced=False` and emits `RuntimeTieredStoreConflict` / `RuntimeTieredStoreInvalid` on transition; requires deleting and recreating CacheRuntime for volume changes |
| All other fields | ❌ No | Not synced; redeployment required |

**Key Takeaways**:
1. The current version supports dynamic updates for `runtimeVersion`, `resources` and `replicas`.
2. An image-only `runtimeVersion` update restarts the affected container in the same Pod. Its in-process cache is lost, and its `emptyDir` cache survives. A `resources` change recreates the Pod, so both caches are lost.
3. `spec.worker.tieredStore` pod-volume fields cannot be updated in place; modifying them sets `TieredStoreSynced=False`, emits a `RuntimeTieredStoreConflict` Warning event, and requires deleting and recreating the CacheRuntime.
4. Modifications to other fields will not take effect and require redeployment.
5. A CacheRuntime is bound to its CacheRuntimeClass as it was when the CacheRuntime was set up; editing the class only affects CacheRuntimes created afterwards.
