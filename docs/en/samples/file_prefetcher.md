# DEMO - Speed Up Application File Reads with Fluid File Prefetching

A distributed cache gives applications a data source with lower latency and higher throughput, but that does not mean reading a cached file always becomes much faster. Besides the "hardware conditions" of the data access path (for example, the network bandwidth and latency between the client and the distributed cache), the access pattern of the application itself also decides how long reading a file takes.

For example, readahead is a common optimization in file access clients. Readahead caches file data that the application is likely to read soon in the client, which lowers the latency of read requests and makes better use of the local network bandwidth. Because it has to predict future read requests, readahead usually works well only when a file is read **sequentially**. If the application tends to read files randomly, readahead can even cause read amplification and make reading slower.

This document describes how to use Fluid file prefetching to solve the under-used bandwidth caused by file access patterns in some specific scenarios.

## When to use Fluid file prefetching

Fluid file prefetching loads remote files into memory on the node of the application Pod in advance. After the application container starts, its random read requests all hit the local memory cache. Since local memory is much faster than reaching the distributed cache over the network, even random reads return quickly, and the application reads its files faster.

Fluid file prefetching fits scenarios where **the list of files the application will read is known in advance, and the memory available to the application container can hold all files in the list**.
- Example 1: speeding up model loading for AI inference services. Loading model parameter files in Safetensors format usually causes random reads. Prefetching loads the model files into memory in advance and speeds up model loading.
- Example 2: speeding up DataFrame analysis. When analyzing DataFrame data stored as Numpy files, SQL statements can cause random reads on the Numpy files. Prefetching loads the Numpy files into memory in advance and speeds up the analysis.

## Example of Fluid file prefetching

### Prerequisites

- Fluid (version >= 1.0.6)
Please refer to the [Fluid installation guide](../userguide/install.md) to install Fluid.

### Example

1. Create the Dataset and the Runtime

Fluid prefetching works with all types of Runtime. This example uses AlluxioRuntime:

```yaml
cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: mydataset
spec:
  mounts:
    - mountPoint: https://mirrors.tuna.tsinghua.edu.cn/apache/zookeeper/stable/
      name: zookeeper
      path: /
---
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: mydataset
spec:
  replicas: 2
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 2Gi
        high: "0.95"
        low: "0.7"
EOF
```

2. Wait until the Runtime is initialized and bound to the Dataset

```
$ kubectl wait --for=condition=Ready dataset mydataset --timeout=120s
```

After a while the command returns the following result, which means the Dataset and the Runtime are ready:
```
dataset.data.fluid.io/mydataset condition met
```

3. Create the application Pod and enable file prefetching for it

Create a sample application Pod with the nginx image:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: demo
  labels:
    fuse.serverful.fluid.io/inject: "true"
  annotations:
    file-prefetcher.fluid.io/inject: "true"
    ### optional annotations:
    # file-prefetcher.fluid.io/file-list: "pvc://mydataset/*.tar.gz"
    # file-prefetcher.fluid.io/async-prefetch: "false"
    # file-prefetcher.fluid.io/extra-envs: "FOO=BAR"
    # file-prefetcher.fluid.io/image: "<file-prefetcher-image>"
spec:
  restartPolicy: Always
  containers:
    - name: demo
      image: nginx
      command:
      - "bash"
      - "-c"
      args:
      - "sleep inf"
      volumeMounts:
        - mountPath: /data/
          name: data-vol
  volumes:
    - name: data-vol
      persistentVolumeClaim:
        claimName: mydataset
```

The parameters in the example mean:

| Type | Parameter | Default | Description |
| ---- | ---- | ---- | ---- |
| labels | `fuse.serverful.fluid.io/inject` | `false`| Whether Fluid processes this Pod. Only Pods with `'true'` are processed by Fluid |
| annotations | `file-prefetcher.fluid.io/inject` | `false` | Whether file prefetching is enabled for this Pod. When enabled, Fluid injects a file prefetcher sidecar container into the Pod |
| annotations | `file-prefetcher.fluid.io/file-list` | All files under all Fluid PVCs mounted in the Pod, i.e. `pvc://<pvc1>/**;pvc://<pvc2>/**;pvc://<pvc3>/**` | Optional. The list of files to prefetch. Several items can be given, separated by semicolons (;). Each item must have the form pvc://<pvc_name>/<glob_path>, where <pvc_name> is the PVC of a Fluid Dataset mounted by the Pod and <glob_path> is a string in glob syntax. For example, `pvc://zookeeper/*.tar.gz` prefetches all files ending in .tar.gz in the `zookeeper` PVC |
| annotations| `file-prefetcher.fluid.io/async-prefetch` | `false` | Optional. Whether the start of the file prefetcher sidecar blocks the start of the application containers. With `true`, all files are not guaranteed to be prefetched when the application container starts. |
| annotations| `file-prefetcher.fluid.io/prefetch-timeout-seconds` | `120` | Only takes effect when async-prefetch=false. The longest time the main container waits for prefetching to finish. |
| annotations| `file-prefetcher.fluid.io/extra-envs` | `<none>` | Optional. Extra environment variables for the file prefetcher sidecar container, in the form `ENV1=value1 ENV2=value2` |
| annotations| `file-prefetcher.fluid.io/image` | Fluid's built-in prefetcher image | Optional. The image of the file prefetcher sidecar container. |


Create the Pod
```
$ kubectl create -f pod.yaml
```
    
4. Check the effect of file prefetching

Check the Pod status. With the configuration above, the file prefetcher sidecar blocks the start of the application container, so the application container may take longer to start than in a normal Pod.
```
$ kubectl wait --for=condition=Ready pod demo
```
The result is:
```
pod/demo condition met
```

Log in to the Pod and use the vmtouch tool to check that each file has been prefetched into local memory:
```
$ kubectl exec -it demo -c demo  -- /bin/bash
$ apt update && apt install -y vmtouch
$ vmtouch /data/*
```
