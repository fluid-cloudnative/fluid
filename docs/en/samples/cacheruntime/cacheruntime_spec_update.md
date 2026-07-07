# CacheRuntime Spec Field Update Capabilities

## 1. Overview

This document describes which `spec` fields of the CacheRuntime's Master and Worker components (backed by **AdvancedStatefulSet**) can be updated in-place.

The current version supports in-place updates for only the following two fields:
- **Container image** (`runtimeVersion`)
- **Resource limits** (`resources`)

Modifications to other fields **will not be propagated to the AdvancedStatefulSet** and require redeployment to take effect.

---

## 2. Supported Components

| Component | Workload Type | Field Update Support |
|-----------|---------------|---------------------|
| **Master** | AdvancedStatefulSet | ✅ Supports `runtimeVersion` and `resources` |
| **Worker** | AdvancedStatefulSet | ✅ Supports `runtimeVersion` and `resources` |
| **Client** | DaemonSet | ❌ Not supported (any changes require redeployment) |

**Notes**:
- Modifying `runtimeVersion` and `resources` on Master and Worker components automatically propagates to the underlying AdvancedStatefulSet.
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

**Limitations**:
- ⚠️ Cannot exceed the node's available resources.
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

## 4. Unsupported Update Fields

Aside from `runtimeVersion` and `resources`, **modifying any other fields in the CacheRuntime spec will not propagate to the AdvancedStatefulSet**, including but not limited to:

- `env`: Environment variables
- `podMetadata`: Pod metadata (labels/annotations)
- `args`/`command`: Container startup arguments
- `ports`: Container ports
- `volumeMounts`/`volumes`: Storage configuration
- `nodeSelector`/`tolerations`/`affinity`: Scheduling configuration
- `securityContext`: Security context
- `replicas`: Replica count
- `disabled`: Component enabled/disabled status

**Notes**:
- After modifying these fields, you must redeploy the CacheRuntime for changes to take effect.
- The system does not automatically detect or propagate changes to these fields.

---

## 5. Summary

| Field | Supported | Update Method |
|-------|-----------|---------------|
| `runtimeVersion` | ✅ Yes | Automatically synced to AdvancedStatefulSet |
| `resources` | ✅ Yes | Automatically synced to AdvancedStatefulSet (requires K8s >= 1.27) |
| All other fields | ❌ No | Not synced; redeployment required |

**Key Takeaways**:
1. The current version supports dynamic updates for only `runtimeVersion` and `resources`.
2. In cgroupv1 environments, these two fields must be updated separately (step-by-step).
3. Modifications to other fields will not take effect and require redeployment.
