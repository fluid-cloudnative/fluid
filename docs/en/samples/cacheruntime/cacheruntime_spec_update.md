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

The whole class is fixed, not just part of it, because these fields are written for one another:
following some of them live would hand new commands or options to components still running the old
images. And a CacheRuntimeClass is cluster-scoped, so following it live would let one edit restart
or recreate the components of every CacheRuntime using it, in every namespace.

- **Editing or deleting a CacheRuntimeClass** only affects CacheRuntimes created afterwards. Existing
  CacheRuntimes keep working even if their class is deleted. To move one to the new class, recreate it.
- **Changing `runtimeClassName`** on an existing CacheRuntime has no effect: it stays bound to its
  snapshot, and Fluid emits a `RuntimeClassChanged` Warning event. Recreate the CacheRuntime to use
  another class.
- **A snapshot that cannot be decoded** is never overwritten. The CacheRuntime stops reconciling and
  Fluid emits a `RuntimeClassSnapshotInvalid` Warning event. Deleting the ConfigMap binds the
  CacheRuntime to its current class again; its master and worker are then synced to that class,
  which restarts or recreates them if their image or resources differ. Until the CacheRuntime is
  bound again, data operations such as DataLoad wait instead of reading the current class.

```bash
# The class a CacheRuntime is bound to
kubectl get cm fluid-runtime-class-snapshot-my-cache -o jsonpath='{.data.runtimeclass\.json}'
```

---

## 5. Unsupported Update Fields

**Any field not listed in section 3 cannot be updated after creation**; changes to it are not propagated to the AdvancedStatefulSet.

**Notes**:
- Such a field takes effect only after the CacheRuntime is redeployed.
- The system does not automatically detect or propagate these changes.

---

## 6. Summary

| Field | Supported | Update Method |
|-------|-----------|---------------|
| `runtimeVersion` | ✅ Yes | Automatically synced to AdvancedStatefulSet (restarts the containers) |
| `resources` | ✅ Yes | Automatically synced to AdvancedStatefulSet (recreates the Pods) |
| `replicas` | ✅ Yes | Automatically synced to AdvancedStatefulSet (scaling in discards that Worker's cache) |
| All other fields | ❌ No | Not synced; redeployment required |

**Key Takeaways**:
1. The current version supports dynamic updates for `runtimeVersion`, `resources` and `replicas`.
2. A `runtimeVersion` change restarts the containers and a `resources` change recreates the Pods; either loses the cache held in memory or `emptyDir` volumes.
3. Modifications to other fields will not take effect and require redeployment.
4. A CacheRuntime is bound to its CacheRuntimeClass as it was when the CacheRuntime was set up; editing the class only affects CacheRuntimes created afterwards.
