# CacheRuntime TieredStore Configuration Example

This document demonstrates how to configure tiered storage (TieredStore) for CacheRuntime.

## Overview

CacheRuntime supports three types of storage media:
1. **ProcessMemory**: Uses process memory as cache storage
2. **EmptyDir**: Uses Kubernetes EmptyDir Volume as cache storage
3. **HostPath**: Uses Kubernetes HostPath Volume as cache storage

The system automatically calculates `mediumType` based on the configuration for cache strategy optimization:
- **MEM**: Memory medium
- **HDD**: Hard disk medium

## Example 1: Using ProcessMemory

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: CacheRuntime
metadata:
  name: my-cache
  namespace: default
spec:
  runtimeClassName: my-runtime-class
  worker:
    tieredStore:
      levels:
        - processMemory:
            quota: 8Gi
          high: "0.95"
          low: "0.7"
```

**Notes:**
- `processMemory.quota` sets the memory quota, which will be automatically added to the container's resource requests/limits
- `high` and `low` are watermark configurations used to control cache eviction

## Example 2: Using EmptyDir Volume (Disk)

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: CacheRuntime
metadata:
  name: my-cache
  namespace: default
spec:
  runtimeClassName: my-runtime-class
  worker:
    tieredStore:
      levels:
        - emptyDir:
            quota: 100Gi
            medium: ""  # Use disk space, set to "Memory" for tmpfs
          high: "0.95"
          low: "0.7"
```

**Notes:**
- `emptyDir` creates a temporary directory with the same lifecycle as the Pod
- `quota` is set as the EmptyDir's sizeLimit
- Cache data will be lost after Pod restart

## Example 3: Using HostPath Volume (Single Path)

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: CacheRuntime
metadata:
  name: my-cache
  namespace: default
spec:
  runtimeClassName: my-runtime-class
  worker:
    tieredStore:
      levels:
        - hostPath:
            paths:
              - /data/cache
            quotas:
              - 100Gi
            type: DirectoryOrCreate
          high: "0.95"
          low: "0.7"
```

**Notes:**
- `hostPath` uses a persistent directory on the node
- `paths` specifies the list of cache paths
- `quotas` specifies the quota for each path
- Cache data persists after Pod restart
- Ensure proper directory permissions on the node

## Example 4: Multi-tier Storage

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: CacheRuntime
metadata:
  name: my-cache
  namespace: default
spec:
  runtimeClassName: my-runtime-class
  worker:
    tieredStore:
      levels:
        # Level 1: High-speed memory cache (mediumType automatically set to MEM)
        - processMemory:
            quota: 4Gi
          high: "0.95"
          low: "0.7"
        
        # Level 2: Large-capacity disk cache (mediumType automatically set to HDD)
        - emptyDir:
            quota: 200Gi
          high: "0.90"
          low: "0.6"
```

**Notes:**
- Multiple tiers can be configured, ordered from highest to lowest priority
- Each tier can have different media types, quotas, and watermarks
- The system prioritizes higher tiers

## Example 5: HostPath Multi-path Configuration

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: CacheRuntime
metadata:
  name: my-cache
  namespace: default
spec:
  runtimeClassName: my-runtime-class
  worker:
    tieredStore:
      levels:
        - hostPath:
            paths:
              - /data/cache1
              - /data/cache2
            quotas:
              - 100Gi
              - 200Gi
            type: DirectoryOrCreate
          high: "0.95"
          low: "0.7"
```

**Notes:**
- Only HostPath type supports multi-path configuration
- The lengths of `paths` and `quotas` arrays must match
- Total quota = 100Gi + 200Gi = 300Gi

## Implementation Details

### mediumType Auto Calculation Logic

`mediumType` is automatically calculated by the system based on the configuration, no manual configuration required:

| Configuration Type | mediumType Value | Description |
|-------------------|------------------|-------------|
| `processMemory` | MEM | Memory medium |
| `emptyDir` (medium="Memory") | MEM | Uses tmpfs memory |
| `emptyDir` (medium="") | HDD | Uses node's default storage (disk) |
| `hostPath` | HDD | Hard disk medium |

### ProcessMemory Handling

When using `processMemory`:
1. Adds `quota` to the container's `resources.requests.memory` and `resources.limits.memory`
2. Accumulates if the container already has other resource requests
3. `mediumType` is automatically set to MEM
4. Creates an EmptyDir Volume with memory medium
5. Volume name format: `tiered-store-level-{N}-memory`
6. Mount path in container: `/dev/shm`

### EmptyDir Handling

When using `emptyDir`:
1. Creates an EmptyDir Volume
2. Volume name format: `tiered-store-level-{N}-index-{M}` (M is always 0 for EmptyDir)
3. Mount path in container: `/etc/fluid/mount/tiered-store/level-{N}-index-{M}-emptydir`
4. `quota` is set as EmptyDir's `sizeLimit`
5. If `medium` is set to `"Memory"`, uses tmpfs, `mediumType` is MEM
6. If `medium` is empty, uses node's default storage, `mediumType` is HDD

### HostPath Handling

When using `hostPath`:
1. Creates an independent HostPath Volume for each path
2. Volume name format: `tiered-store-level-{N}-index-{M}`
3. Mount path in container: `/etc/fluid/mount/tiered-store/level-{N}-index-{M}-hostpath`
4. `paths` list specifies directory paths on the node
5. `quotas` list specifies quotas for each path (for reference only, not enforced)
6. `type` specifies the hostPath type (e.g., `DirectoryOrCreate`)
7. `mediumType` is automatically set to HDD

> **Note:** Volume names are implementation details and may change. Users should not depend on specific volume name formats.

### Important Notes on Multi-path Configuration

According to the latest API definition, **only HostPath type supports multi-path configuration**:

| Media Type | Multi-path Support | Quota Enforcement | Description |
|------------|-------------------|------------------|-------------|
| **ProcessMemory** | ❌ Not supported | Via container memory limit | Single quota configuration |
| **EmptyDir** | ❌ Not supported | Via `sizeLimit` | Single quota configuration |
| **HostPath** | ✅ Supported | For reference only (requires external mechanism) | Multiple paths configurable |

**HostPath Multi-path Details:**
- ✅ Each path creates an independent HostPath volume
- ✅ `paths` and `quotas` arrays must have matching lengths
- ⚠️ `quotas` are for reference only, external mechanisms (e.g., XFS quota, Cgroup) are required for enforcement
- 💡 Main purpose of multi-path: Expand total cache capacity or distribute across multiple disks

## Notes

1. **Quota Units**: Supports Kubernetes standard units (Ki, Mi, Gi, Ti, etc.)
2. **Watermarks**: `high` and `low` are decimals between 0-1, representing usage percentage
3. **Media Selection**: Each tier can only select one media type (ProcessMemory, EmptyDir, or HostPath)
4. **Multi-path Support**: Only HostPath type supports multi-path configuration; ProcessMemory and EmptyDir only support single path
5. **HostPath Path Requirements**: `paths` and `quotas` arrays must have matching lengths
6. **mediumType**: Automatically calculated by the system, no manual configuration needed, only supports MEM and HDD values
7. **Quota Enforcement Mechanism**:
   - ProcessMemory: Enforced via container memory limit
   - EmptyDir: Enforced via `sizeLimit`
   - HostPath: `quotas` are for reference only, requires external mechanisms (e.g., XFS project quota, Cgroup)
