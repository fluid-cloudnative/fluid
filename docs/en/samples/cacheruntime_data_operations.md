# Example - CacheRuntime Data Operations

## Prerequisites

This document is an extension of the [CacheRuntime Integration Guide](../dev/generic_cache_runtime_integration.md) and **assumes you have already completed the basic CacheRuntime integration** (including defining topology, configuring components, etc.).

This document only explains how to **add data operation support** to an existing `CacheRuntimeClass`. The core change involves just one field: `dataOperationSpecs`.

## Background

Fluid's CacheRuntime provides a generic cache runtime abstraction that allows users to define implementation details for different caching systems through `CacheRuntimeClass`. Starting from the latest version of Fluid, CacheRuntime natively supports data operations, including DataLoad (data preloading) and DataProcess (data processing).

This document demonstrates how to configure and use the DataLoad feature for CacheRuntime.

> Note: The DataProcess Spec defines Pod information and mounts the Dataset as a PVC, so no modifications to the caching system are required to use the DataProcess feature.

## Environment Verification

Before running this example, please refer to the [Installation Guide](../userguide/install.md) to complete the Fluid installation and verify that all Fluid components are running properly:

```shell
$ kubectl get pod -n fluid-system
cacheruntime-controller-xxxxx              1/1     Running   0          8h
csi-nodeplugin-fluid-fwgjh                  2/2     Running   0          8h
dataset-controller-5b7848dbbb-n44dj         1/1     Running   0          8h
```

Ensure that your cluster has installed a CacheRuntime controller that supports data operations.

## Core Concepts

### The Only Change Compared to Basic Integration

Compared to the basic CacheRuntime integration, supporting data operations **only requires adding one top-level field** to the CacheRuntimeClass:

- **Only add** the `dataOperationSpecs` field
- **No need to modify** any existing fields (topology, fileSystemType, extraResources, etc.)
- **Backward compatible**: CacheRuntimeClass without this field configured can still use basic caching functionality normally

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: CacheRuntimeClass
metadata:
  name: curvine-demo
fileSystemType: curvinefs

# [NEW] Only this field is added; other configurations (topology, etc.) remain unchanged
dataOperationSpecs:
  - name: DataLoad
    command: ["/bin/bash", "-c"]
    args: ["..."]

# [ORIGINAL CONFIGURATION] Fields like topology and extraResources require no modifications
topology:
  master:
    # ... Exactly the same as basic integration
  worker:
    # ... Exactly the same as basic integration
  client:
    # ... Exactly the same as basic integration
```

### Detailed Explanation of dataOperationSpecs Field

`dataOperationSpecs` is an array where each element defines the execution specification for a type of data operation.

#### Field Structure

```yaml
dataOperationSpecs:
  - name: <operation type>
    command: [<command>, <parameters>]
    args: [<script or parameters>]
    image: <optional: dedicated image>
```

#### Field Description

| Field Name | Type | Required | Description                                                                                                                        |
|--------|------|----|---------------------------------------------------------------------------------------------------------------------------|
| `name` | string | Yes | Operation type identifier. Currently supported values:<br>• `DataLoad`: Data preloading operation<br>• `DataMigrate`: Data migration operation (not yet supported)<br>• `DataBackup`: Data backup operation (not yet supported) |
| `command` | []string | Yes | Command to execute in the container (entrypoint), typically set to `["/bin/bash", "-c"]` to support script execution                                                                  |
| `args` | []string | Yes | Arguments for the command, usually containing the complete execution script. The script can use environment variables injected by Fluid (see below)                                                                               |
| `image` | string | No | Container image used for the operation.<br>• **If not specified**: Defaults to using the `worker` component image from `CacheRuntimeClass`<br>• **If specified**: Uses a custom dedicated image (suitable for scenarios requiring special tools)                |

### Available Environment Variables

During data operation execution, Fluid automatically injects the following environment variables into the container:

#### DataLoad-Specific Environment Variables

| Environment Variable Name | Description                             | Example Value |
|-----------|--------------------------------|--------|
| `FLUID_DATALOAD_METADATA` | Whether to load metadata                        | `"true"` or `"false"` |
| `FLUID_DATALOAD_DATA_PATH` | Data paths to be loaded (multiple paths separated by colons)           | `/spark/spark-3.0.1:/spark/spark-2.4.7` |
| `FLUID_DATALOAD_PATH_REPLICAS` | Number of replicas for each path (separated by colons, corresponding one-to-one with DATA_PATH) | `1:2` |

The underlying caching system writes data preloading scripts based on the above environment variables and packages them into the image. When users define DataLoad operations, they can specify the script through the `command` and `args` fields.
