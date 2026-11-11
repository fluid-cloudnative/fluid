# CacheRuntime Spec Field Update Capabilities

## 1. Overview

This document describes which `spec` fields of the CacheRuntime's Master and Worker components (backed by **AdvancedStatefulSet**) can be updated after the CacheRuntime is created, and how edits to a CacheRuntimeClass affect the CacheRuntimes that use it.

The current version syncs the following three fields to the running components:
- **Container image** (`runtimeVersion`): the containers restart in place
- **Resource limits** (`resources`): the Pods are recreated
- **Replica count** (`replicas`): the component is scaled

Modifications to other fields **will not be propagated to the AdvancedStatefulSet** and require redeployment to take effect.

A CacheRuntime is bound to its CacheRuntimeClass as the class is when the CacheRuntime is set up: editing a CacheRuntimeClass later does not change existing CacheRuntimes. See [section 4](#4-changes-to-the-cacheruntimeclass).

---

## 2. Supported Components

| Component | Workload Type | Field Update Support |
|-----------|---------------|---------------------|
| **Master** | AdvancedStatefulSet | ✅ Supports `runtimeVersion`, `resources` and `replicas` |
| **Worker** | AdvancedStatefulSet | ✅ Supports `runtimeVersion`, `resources` and `replicas` |
| **Client** | DaemonSet | ❌ Not supported (any changes require redeployment) |

**Notes**:
- Modifying `runtimeVersion`, `resources` and `replicas` on Master and Worker components automatically propagates to the underlying AdvancedStatefulSet.
- The AdvancedStatefulSet controller bundled with Fluid updates only container images in place. A change of `runtimeVersion` restarts the containers and keeps the Pods; a change of `resources` recreates the Pods one by one.
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

**Resolution**: on creation and on every reconcile, the image of the component's first container in the CacheRuntimeClass template, as recorded when the CacheRuntime was set up (see [section 4](#4-changes-to-the-cacheruntimeclass)), is the baseline, and the `image` and `imageTag` set in the CacheRuntime override it. "Template image" below refers to that recorded image.

**Completion rules**:
- `image` and `imageTag` may be set on their own. The missing half is taken from the template image. With a template image of `fluid-cache:v1.0.0`:
  - Only `imageTag: v1.1.0`: the template's image name is kept, giving `fluid-cache:v1.1.0`
  - Only `image: registry.example.com/fluid-cache`: the template's tag is kept (**not `latest`**), giving `registry.example.com/fluid-cache:v1.0.0`
  - Neither: the template image `fluid-cache:v1.0.0` is used
- A `runtimeVersion` setting both `image` and `imageTag` is used as is, even when the template image is pinned by digest.
- When only one half is set and the other cannot be completed (only `image` is set and the template image has no tag, or the template image or `image` is pinned by digest (`@sha256:...`)), the component uses the template image on creation; on update the edit has no effect and the component keeps its current image. In both cases the controller emits a `RuntimeVersionIncomplete` Warning event on the CacheRuntime and logs the edit; the client, which is not updated in place, reports it on creation only. `runtimeVersion` cannot express a digest reference; to pin a digest, put it in the CacheRuntimeClass template.
- Data operations that name no image of their own (such as DataLoad) use the same completed image as the worker. When the worker's version cannot be completed they use the template image, which can differ from the image a worker kept after such an edit.

**Limitations**:
- ⚠️ **An image change restarts the containers**: the master or worker containers are restarted in place with the new image. The Pods are kept, but the cache processes start over.
- ⚠️ **When setting only `image`, make sure the template's tag exists in the new repository.** Otherwise the containers fail with `ImagePullBackOff`, and the component is unavailable until the new image can be pulled.
- ⚠️ Editing the image in the CacheRuntimeClass template does not change existing CacheRuntimes, see [section 4](#4-changes-to-the-cacheruntimeclass).
- ⚠️ Removing `runtimeVersion` from a CacheRuntime returns the component to the template image, except when the template image has no tag or is pinned by digest: the component then keeps its current image, and you need to set a full `runtimeVersion` or recreate the CacheRuntime.

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

`resources` is resolved as follows on every reconcile:

1. the `resources` the CacheRuntimeClass template declared when the CacheRuntime was set up form
   the baseline (see [section 4](#4-changes-to-the-cacheruntimeclass));
2. the `resources` set on the CacheRuntime are overlaid on that baseline key by key —
   `requests` and `limits` merge by resource name, `claims` by claim name;
3. a key neither declares is cleared from the workload.

The CacheRuntime therefore expresses only the differences it wants from the template. Naming
`limits.memory` alone moves that one value and leaves the template's `requests.cpu`,
`requests.memory` and `limits.cpu` in place.

Removing a key from the CacheRuntime works the same way: it falls back to the template's value,
or is cleared from the workload when the template does not declare it. An empty list such as
`limits: {}` names no key, so it reads the same as leaving `limits` out.

**Limitations**:
- ⚠️ **A resources change recreates the Pods**: the AdvancedStatefulSet controller bundled with Fluid
  updates only container images in place, so any change to the resolved resources of the master or
  worker, including one that only adds or removes a key, is rolled out by recreating its Pods one by
  one. Cached data held in memory or in `emptyDir` volumes is lost and has to be loaded again.
- ⚠️ Cannot exceed the node's available resources.
- ⚠️ A key the CacheRuntimeClass template declares can be **overridden but not removed**: omitting it
  from the CacheRuntime leaves the template's value in place rather than leaving the component
  unconstrained. To adjust resource requirements, explicitly set the desired values in the CacheRuntime.
- ⚠️ Editing `resources` in the CacheRuntimeClass template does not change existing CacheRuntimes, see
  [section 4](#4-changes-to-the-cacheruntimeclass).
- ⚠️ **The CacheRuntime and its template are the source of truth**: resources edited directly on
  the master or worker AdvancedStatefulSet are reverted on the next reconcile.
- ⚠️ **A request must not exceed its limit after the merge**: the template and the CacheRuntime
  can each be valid and still merge into a request above its limit, for example a template
  `requests.cpu: 1` with a CacheRuntime `limits.cpu: 500m`. Huge pages (`hugepages-*`) and
  extended resources (such as `nvidia.com/gpu`) are stricter: their request must equal their
  limit, and a request needs a limit, so a template setting `nvidia.com/gpu` to 1 for both with
  a CacheRuntime raising only the limit to 2 conflicts as well. The controller then does not sync
  that component's `resources`, keeps the workload's current values and emits a
  `RuntimeResourcesConflict` Warning event on the CacheRuntime naming the conflicting keys; a new
  CacheRuntime with such a pair fails to set up. Set the matching request on the CacheRuntime too
  (equal to the limit for huge pages and extended resources), or change the template.
- ⚠️ **`claims` can only reference pod-level `resourceClaims` the workload declares**: the CacheRuntime
  sets the container's `resources.claims` but cannot add pod-level `spec.resourceClaims`. On creation,
  a claim name the template does not declare makes the API server reject every master or worker Pod
  (`must be one of the names in pod.spec.resourceClaims`), and the component stays NotReady; clusters
  without `DynamicResourceAllocation` enabled (off by default before K8s 1.34) drop `claims` silently
  instead. On update, a claim that the workload neither declares nor already carries is not synced:
  the controller leaves it out, still syncs the rest of `resources`, and emits a
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

## 4. Changes to the CacheRuntimeClass

A CacheRuntime is bound to its CacheRuntimeClass as the class is when the CacheRuntime is first set
up. Fluid stores a snapshot of the class in the ConfigMap `fluid-runtime-class-snapshot-<CacheRuntime name>`
in the CacheRuntime's namespace, owned by the CacheRuntime, and from then on reads the class only
from that snapshot: the component templates, images and resources, `options`, `executionEntries` and
`dataOperationSpecs` alike. A class bundles these for one version of a cache system, so a
CacheRuntime keeps using all of them together.

- **Editing or deleting a CacheRuntimeClass only affects the CacheRuntimes created afterwards.**
  Existing CacheRuntimes keep running on, and keep being reconciled against, their snapshot.
- **To move an existing CacheRuntime to the current version of its class, recreate it.** Changing
  `runtimeClassName` has no effect: the CacheRuntime stays bound to its snapshot and a
  `RuntimeClassChanged` Warning event is emitted.
- **Do not edit the snapshot.** Its data is immutable. A snapshot that cannot be read is reported
  with a `RuntimeClassSnapshotInvalid` Warning event and is left alone; the CacheRuntime is not
  reconciled until the snapshot is deleted, which binds the CacheRuntime again as described below.
- **CacheRuntimes created by an earlier Fluid version** have no snapshot and are bound on their first
  reconcile after the upgrade. The snapshot is taken from the current CacheRuntimeClass, except for
  the resources and image of the master and worker, which are taken from what their workloads run,
  so that binding changes no running component. A key the CacheRuntime sets is not recorded as a
  template default, so removing it from such a CacheRuntime later clears it.
  - If a component's workload does not run what the CacheRuntime declares, for example because the
    CacheRuntime was edited while the controller was being upgraded, Fluid does not apply the
    difference, as that would recreate its Pods or restart its containers. It lists the component in
    the `cacheruntime.fluid.io/adoption-pending` annotation of the snapshot and emits a
    `RuntimeAdoptionPending` Warning event; replicas are still synced. Remove the component from the
    annotation to apply the difference. The entry is dropped on its own once the workload runs what
    the CacheRuntime declares.
  - Until a CacheRuntime is bound, data operations such as DataLoad read the current class.

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
| `runtimeVersion` | ✅ Yes | Automatically synced to AdvancedStatefulSet (restarts the containers in place) |
| `resources` | ✅ Yes | Automatically synced to AdvancedStatefulSet (recreates the Pods) |
| `replicas` | ✅ Yes | Automatically synced to AdvancedStatefulSet (scaling in discards that Worker's cache) |
| All other fields | ❌ No | Not synced; redeployment required |

**Key Takeaways**:
1. The current version supports dynamic updates for `runtimeVersion`, `resources` and `replicas`.
2. A `resources` change recreates the master or worker Pods, losing the cache they hold in memory or `emptyDir` volumes; an image change restarts their containers.
3. Editing a CacheRuntimeClass does not change existing CacheRuntimes; recreate a CacheRuntime to pick up a new version of its class.
4. Modifications to other fields will not take effect and require redeployment.
