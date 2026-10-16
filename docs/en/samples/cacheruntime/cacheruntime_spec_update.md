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

**Limitations**:
- ⚠️ In **cgroupv1** environments, this field cannot be updated simultaneously with `resources` (perform step-by-step; see the resource field section below).

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

1. the `resources` declared by the CacheRuntimeClass template form the baseline;
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
- ⚠️ Cannot exceed the node's available resources.
- ⚠️ A key the CacheRuntimeClass template declares can be **overridden but not removed**: omitting it
  from the CacheRuntime leaves the template's value in place rather than leaving the component
  unconstrained. To relax a limit, set the value you want explicitly; to drop a requirement
  altogether, change the CacheRuntimeClass template, which is where the runtime's own requirements
  are described.
- ⚠️ **Edits to the template's `resources` reach existing instances**: the CacheRuntimeClass template
  is the baseline on every reconcile, so after it changes, the keys a CacheRuntime does not override
  are synced to the master and worker workloads of every CacheRuntime that uses that class, without
  recreating the CacheRuntime. The controller does not watch CacheRuntimeClass, so the change lands
  on each CacheRuntime's next reconcile rather than immediately. The client is not synced, and
  template fields other than `resources` are not synced either. A key removed from the template is
  cleared from the workloads that do not override it.
- ⚠️ **The CacheRuntime and its template are the source of truth**: resources edited directly on
  the master or worker AdvancedStatefulSet are reverted on the next reconcile.
- ⚠️ **Upgrade note**: earlier releases dropped the template's other requirements from a
  CacheRuntime that set only part of `resources` (see
  [#6173](https://github.com/fluid-cloudnative/fluid/issues/6173)). The first reconcile after
  upgrading restores them, so the affected master and worker Pods go through a one-time resources
  update, which recreates the Pods where the cluster cannot resize them in place. A CacheRuntime
  that sets no `resources` is not affected, unless its template declares none either and its
  workloads carry resources from an earlier override or a manual edit: those are cleared.
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
- ⚠️ **`claims` can only reference pod-level `resourceClaims` the template declares**: the CacheRuntime
  sets the container's `resources.claims` but cannot add pod-level `spec.resourceClaims`. A claim name
  the CacheRuntimeClass template does not declare makes the API server reject every master or worker
  Pod that is created or recreated (`must be one of the names in pod.spec.resourceClaims`), and the
  component stays NotReady. Clusters without `DynamicResourceAllocation` enabled (off by default
  before K8s 1.34) drop `claims` silently instead. Earlier releases ignored a CacheRuntime that set
  only `claims`.
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

**Notes**:
- Such a field takes effect only after the CacheRuntime is redeployed.
- The system does not automatically detect or propagate these changes.

---

## 5. Summary

| Field | Supported | Update Method |
|-------|-----------|---------------|
| `runtimeVersion` | ✅ Yes | Automatically synced to AdvancedStatefulSet |
| `resources` | ✅ Yes | Automatically synced to AdvancedStatefulSet (requires K8s >= 1.27) |
| `replicas` | ✅ Yes | Automatically synced to AdvancedStatefulSet (scaling in discards that Worker's cache) |
| All other fields | ❌ No | Not synced; redeployment required |

**Key Takeaways**:
1. The current version supports dynamic updates for `runtimeVersion`, `resources` and `replicas`.
2. In cgroupv1 environments, `runtimeVersion` and `resources` must be updated separately (step-by-step); `replicas` is not subject to this limitation.
3. Modifications to other fields will not take effect and require redeployment.
