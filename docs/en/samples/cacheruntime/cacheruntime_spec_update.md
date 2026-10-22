# CacheRuntime Spec Field Update Capabilities

## 1. Overview

This document describes which `spec` fields of the CacheRuntime's Master and Worker components (backed by **AdvancedStatefulSet**) can be updated in-place.

The current version supports in-place updates for the following three fields:
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

**Resolution**: on creation and on every reconcile, the image of the component's first container in the CacheRuntimeClass template is the baseline, and the `image` and `imageTag` set in the CacheRuntime override it.

**Completion rules**:
- `image` and `imageTag` may be set on their own. The missing half is taken from the template image. With a template image of `fluid-cache:v1.0.0`:
  - Only `imageTag: v1.1.0`: the template's image name is kept, giving `fluid-cache:v1.1.0`
  - Only `image: registry.example.com/fluid-cache`: the template's tag is kept (**not `latest`**), giving `registry.example.com/fluid-cache:v1.0.0`
  - Neither: the template image `fluid-cache:v1.0.0` is used
- A `runtimeVersion` setting both `image` and `imageTag` is used as is, even when the template image is pinned by digest.
- When only one half is set and the other cannot be completed (only `image` is set and the template image has no tag, or the template image or `image` is pinned by digest (`@sha256:...`)), the component uses the template image on creation; on update the edit has no effect and the component keeps its current image. In both cases the controller emits a `RuntimeVersionIncomplete` Warning event on the CacheRuntime and logs the edit; the client, which is not updated in place, reports it on creation only. `runtimeVersion` cannot express a digest reference; to pin a digest, put it in the CacheRuntimeClass template.
- Data operations that name no image of their own (such as DataLoad) use the same completed image as the worker. When the worker's version cannot be completed they use the template image, so they can differ from a worker that kept its current image, for example after the template image is changed to one pinned by digest.

**Limitations**:
- ⚠️ In **cgroupv1** environments, this field cannot be updated simultaneously with `resources` (perform step-by-step; see the resource field section below).
- ⚠️ **When setting only `image`, make sure the template's tag exists in the new repository.** Otherwise the pods fail with `ImagePullBackOff`; on an in-place update the component is unavailable until the new image can be pulled. The same holds the other way round: an instance that sets only `imageTag` moves to that tag in the new repository when the template's image repository changes, so make sure the tag exists there too.
- ⚠️ **Image edits in the CacheRuntimeClass template reach existing instances**: after the template image changes, each enabled master and worker of a CacheRuntime using that CacheRuntimeClass that does not fully set `runtimeVersion` is re-resolved against the new template image, and rolls only if the resolved image changes (one that sets only `imageTag` keeps its tag, one that sets only `image` keeps its image name; an instance that cannot be resolved to a `name:tag` image keeps its current image, for example when the new template image is pinned by digest). The controller does not watch CacheRuntimeClass, so the change takes effect on each CacheRuntime's next reconcile rather than immediately. The client is not updated.
- ⚠️ Removing `runtimeVersion` from a CacheRuntime returns the component to the template image, except when the template image has no tag or is pinned by digest: the component then keeps its current image, and you need to set a full `runtimeVersion` or recreate the CacheRuntime.

---

### 3.2 Resource Limits (`resources`)

**Field path**: `spec.{master,worker}.resources`

**Supported sub-fields**:
- `requests.cpu`: CPU request
- `requests.memory`: Memory request
- `limits.cpu`: CPU limit
- `limits.memory`: Memory limit

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

1. the value set on the CacheRuntime, if it declares any `requests` or `limits`;
2. otherwise the value declared by the CacheRuntimeClass template;
3. otherwise nothing is synced and the workload keeps its current resources.

**Limitations**:
- ⚠️ Cannot exceed the node's available resources.
- ⚠️ When the CacheRuntimeClass template declares `resources`, removing `resources` from the
  CacheRuntime does **not** leave the component unconstrained — it falls back to the template value.
  To relax a limit, set the value you want explicitly instead of removing the field.
- ⚠️ **Kubernetes version requirement**: K8s >= 1.27 with the `InPlacePodVerticalScaling` Feature Gate enabled.
  ```bash
  # Check if the Feature Gate is enabled
  kubectl get nodes -o jsonpath='{.items[0].status.config.kubeletConfig.featureGates.InPlacePodVerticalScaling}'
  ```
- ⚠️ **Cgroupv1 limitation**: Cannot be updated simultaneously with the `runtimeVersion` field.
  - Reason: Cgroupv1 does not support updating container image and resource limits in the same operation.
  - Solution: Perform step-by-step updates — update resources first, wait for completion, then update the image.
  ```bash
  # Step 1: Update resources
  kubectl patch cacheruntime my-cache --type='merge' -p '{"spec":{"worker":{"resources":{"requests":{"cpu":"4","memory":"8Gi"}}}}}'
  kubectl wait pod -l fluid.io/cache-runtime-name=my-cache --for=condition=InPlaceUpdateReady --timeout=300s
  
  # Step 2: Update image
  kubectl patch cacheruntime my-cache --type='merge' -p '{"spec":{"worker":{"runtimeVersion":{"imageTag":"v1.1.0"}}}}'
  ```
  See [Kubernetes Issue #127356](https://github.com/kubernetes/kubernetes/issues/127356).

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

## 4. Unsupported Update Fields

**Any field not listed in section 3 cannot be updated in place**; changes to it are not propagated to the AdvancedStatefulSet.

### 4.1 Worker Tiered Storage (`spec.worker.tieredStore`)

**Field path**: `spec.worker.tieredStore`

- **Why in-place updates are not supported**: Tiered storage volumes (`emptyDir.sizeLimit`, memory tmpfs `/dev/shm`, hostPath mounts) and container memory adjustments are generated during initial creation (`Setup`). In Kubernetes, pod volume definitions and volume `sizeLimit` properties cannot be modified dynamically in running pods.
- **Drift warning behavior**: While editing `spec.worker.tieredStore` updates the runtime configuration ConfigMap, the running worker pods retain their creation-time volume configurations and limits. When `spec.worker.tieredStore` quotas or paths diverge from the worker pod volumes, the controller detects the split during reconciliation and emits a `RuntimeTieredStoreConflict` Warning event on the `CacheRuntime`:
  ```bash
  kubectl get events -n <namespace> --field-selector type=Warning,reason=RuntimeTieredStoreConflict
  ```
- **How to apply changes**: Changes to `spec.worker.tieredStore` require deleting and recreating the CacheRuntime to take effect.

---

## 5. Summary

| Field | Supported | Update Method |
|-------|-----------|---------------|
| `runtimeVersion` | ✅ Yes | Automatically synced to AdvancedStatefulSet |
| `resources` | ✅ Yes | Automatically synced to AdvancedStatefulSet (requires K8s >= 1.27) |
| `replicas` | ✅ Yes | Automatically synced to AdvancedStatefulSet (scaling in discards that Worker's cache) |
| `spec.worker.tieredStore` | ❌ No | Emits `RuntimeTieredStoreConflict` Warning event on drift; requires deleting and recreating CacheRuntime |
| All other fields | ❌ No | Not synced; redeployment required |

**Key Takeaways**:
1. The current version supports dynamic updates for `runtimeVersion`, `resources` and `replicas`.
2. In cgroupv1 environments, `runtimeVersion` and `resources` must be updated separately (step-by-step); `replicas` is not subject to this limitation.
3. `spec.worker.tieredStore` cannot be updated in place; modifying it causes volume drift and emits a `RuntimeTieredStoreConflict` Warning event. Deleting and recreating the CacheRuntime is required.
4. Modifications to other fields will not take effect and require redeployment.
