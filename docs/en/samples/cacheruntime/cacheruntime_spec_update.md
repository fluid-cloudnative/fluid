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
