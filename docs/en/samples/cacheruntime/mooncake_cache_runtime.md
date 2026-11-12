# Example - Deploy Mooncake with CacheRuntime

[Mooncake](https://github.com/kvcache-ai/Mooncake) is a distributed KVCache store built for LLM inference workloads. Unlike cache systems such as Alluxio or JuiceFS, Mooncake does not provide POSIX mount semantics: applications talk to the cache service directly through its own client library instead of reading and writing files through a mount point.

Fluid's generic CacheRuntime supports this kind of cache system: the CacheRuntimeClass `topology` declares only the master and worker components, and omits the client component that would otherwise be responsible for mounting. This document walks through that minimal setup.

For background on the client-less architecture, see the [Generic Cache System Integration Guide](../../dev/generic_cache_runtime_integration.md).

## Prerequisites

Before running this example, follow the [Installation Guide](../../userguide/install.md) to install Fluid, and verify that its components are running:

```shell
$ kubectl get pod -n fluid-system
NAME                                      READY   STATUS    RESTARTS   AGE
cacheruntime-controller-58c775584-x9q47   1/1     Running   0          10m
csi-nodeplugin-fluid-74bpw                2/2     Running   0          10m
csi-nodeplugin-fluid-cjvc6                2/2     Running   0          10m
csi-nodeplugin-fluid-x4wwp                2/2     Running   0          10m
dataset-controller-58497b968b-q6ssv       1/1     Running   0          10m
fluid-webhook-74db64fdf4-7rq69            1/1     Running   0          10m
fluidapp-controller-7dbdc7696b-86pph      1/1     Running   0          10m
```

Typically you should see one pod each for `dataset-controller`, `cacheruntime-controller`, `fluid-webhook` and `fluidapp-controller`, plus one `csi-nodeplugin` pod per node. `cacheruntime-controller` is the one this example requires; it is enabled with `--set runtime.cacheruntime.enabled=true` when installing the Helm chart.

> Note: this example requires a Fluid build that includes [#6157](https://github.com/fluid-cloudnative/fluid/pull/6157). On earlier versions the controller panics when the client component is omitted from `topology`.

### The demo image

Fluid does not publish a Mooncake image. This example uses a demo image built from the Mooncake Python distribution plus two small scripts that Fluid invokes:

| Path | Purpose |
|---|---|
| `/custom-entrypoint.sh` | Component entrypoint; starts the right process based on the role (master/worker) |
| `/reportSummary.sh` | Collects cache usage and emits it as JSON in the format Fluid expects |

The build context for that image lives in this repository under [`samples/mooncake/docker`](../../../../samples/mooncake/docker), so you can build an equivalent image yourself:

```shell
$ docker build -t <your-registry>/mooncake:v3 samples/mooncake/docker
$ docker push <your-registry>/mooncake:v3
```

Note that `apt-get` and `pip` resolve to whatever versions are current at build time, so a fresh build produces a functionally equivalent image, not a bit-for-bit reproduction of the digest below.

**Building your own image and substituting it in the manifests below is the recommended path.** For convenience, a prebuilt copy is also available, pinned to an immutable digest so this example keeps working even if the tag is moved:

```
btxu/mooncake:v3@sha256:067614b70d25b496e3edc3480747d558ee8a364ef47a67f669f5d96ca5098552
```

An Alibaba Cloud mirror is available for networks in mainland China. It is a copy of the same manifest, so it carries the identical digest:

```
crpi-4hkqof7tc9brc6d5.cn-hongkong.personal.cr.aliyuncs.com/mooncake1314/mooncake:v3@sha256:067614b70d25b496e3edc3480747d558ee8a364ef47a67f669f5d96ca5098552
```

> Note: both registries are personal accounts belonging to the author of this example, not project-controlled infrastructure, and they carry no availability guarantee. Treat them as a convenience for trying the example out, and build from `samples/mooncake/docker` for anything beyond that.

Neither script is specific to this image: you can build an equivalent image on top of any Mooncake distribution, as long as it provides the same two entry points. See the conventions in the [Generic Cache System Integration Guide](../../dev/generic_cache_runtime_integration.md).

## Running the Example

### Create the CacheRuntimeClass

**Review the CacheRuntimeClass to be created**

```shell
$ cat<<EOF >mooncake-cacheruntimeclass.yaml
apiVersion: data.fluid.io/v1alpha1
kind: CacheRuntimeClass
metadata:
  name: mooncake-demo
fileSystemType: mooncakefs
topology:
  master:
    service:
      headless: {}
    executionEntries:
      reportSummary:
        command:
          - bash
          - -c
          - /reportSummary.sh
        timeout: 30
    template:
      spec:
        restartPolicy: Always
        containers:
          - name: master
            image: btxu/mooncake:v3@sha256:067614b70d25b496e3edc3480747d558ee8a364ef47a67f669f5d96ca5098552
            command:
              - /custom-entrypoint.sh
            args:
              - master
              - start
            imagePullPolicy: IfNotPresent
            readinessProbe:
              tcpSocket:
                port: 50051
              initialDelaySeconds: 10
              periodSeconds: 5
              failureThreshold: 12
            env:
              - name: POD_NAME
                valueFrom:
                  fieldRef:
                    fieldPath: metadata.name
            ports:
              - containerPort: 50051
                name: rpc
              - containerPort: 8080
                name: metadata
              - containerPort: 9003
                name: metrics
  worker:
    service:
      headless: {}
    template:
      spec:
        restartPolicy: Always
        containers:
          - name: worker
            image: btxu/mooncake:v3@sha256:067614b70d25b496e3edc3480747d558ee8a364ef47a67f669f5d96ca5098552
            command:
              - /custom-entrypoint.sh
            args:
              - worker
              - start
            imagePullPolicy: IfNotPresent
            readinessProbe:
              tcpSocket:
                port: 50052
              initialDelaySeconds: 5
              periodSeconds: 5
              failureThreshold: 12
            env:
              - name: POD_NAME
                valueFrom:
                  fieldRef:
                    fieldPath: metadata.name
            ports:
              - containerPort: 50052
                name: data
              - containerPort: 9300
                name: http
EOF
```

`CacheRuntimeClass` is a CRD defined by Fluid that describes how a class of cache system runs on Kubernetes — the image each component uses, how it starts, which ports it exposes, and how Fluid collects its runtime status.

A few things worth noting in this example:

- Only `master` and `worker` are declared under `topology`. All three components (master, worker, client) are optional in the API, and Mooncake does not need a client component.
- Note that `fileSystemType` and `topology` are top-level fields — they do not live under `spec`.
- The master component exposes three ports: `50051` for RPC, `8080` for the metadata service, and `9003` for metrics. Applications connect directly to the first two.
- `reportSummary` points at `/reportSummary.sh` inside the image. Fluid executes it periodically in the component pod and writes the result to the Dataset's `status` field. For the required output format, see the [Generic Cache System Integration Guide](../../dev/generic_cache_runtime_integration.md).

**Create the CacheRuntimeClass**

```shell
$ kubectl apply -f mooncake-cacheruntimeclass.yaml
cacheruntimeclass.data.fluid.io/mooncake-demo created

$ kubectl get cacheruntimeclass
NAME            AGE
mooncake-demo   0s
```

> Tip: if you modify the CacheRuntimeClass afterwards, you must delete and recreate the CacheRuntime for the change to take effect. A CacheRuntime renders the CacheRuntimeClass into workloads at creation time; later edits are not applied retroactively, and deleting the pods alone does not help.

### Create the Dataset and CacheRuntime

**Review the resources to be created**

```shell
$ cat<<EOF >mooncake-dataset-runtime.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: mooncake-demo
  namespace: default
spec:
  placement: Shared
  accessModes:
    - ReadWriteMany
  mounts:
    - name: mc
      mountPoint: "mooncakefs:///"
---
apiVersion: data.fluid.io/v1alpha1
kind: CacheRuntime
metadata:
  name: mooncake-demo
  namespace: default
spec:
  runtimeClassName: mooncake-demo
  master:
    replicas: 1
  worker:
    replicas: 2
    tieredStore:
      levels:
        - emptyDir:
            quota: 1Gi
          high: "0.8"
          low: "0.5"
EOF
```

The `Dataset` describes the dataset itself, while the `CacheRuntime` describes the runtime instance providing cache service for it — here, one master replica and two worker replicas.

About the `mounts` field: Mooncake does not mount any underlying storage (UFS); data is written directly into the cache by the client, so `mooncakefs:///` here is only a placeholder mount point. Fluid passes the contents of `mounts` through into the CacheRuntime config ConfigMap for the component containers to read, but because this example's CacheRuntimeClass declares no `mountUfs` execution entry, Fluid never performs an actual mount (see the notes on skipping MountUFS in the [Generic Cache System Integration Guide](../../dev/generic_cache_runtime_integration.md)). In the API `mounts` is optional, but if you do declare it, it must contain at least one entry.

**Create the resources**

```shell
$ kubectl apply -f mooncake-dataset-runtime.yaml
dataset.data.fluid.io/mooncake-demo created
cacheruntime.data.fluid.io/mooncake-demo created
```

**Check the Dataset status**

```shell
$ kubectl get dataset
NAME            UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE      AGE
mooncake-demo                                                                  NotBound   6s
```

As shown above, `PHASE` is `NotBound`, meaning the `Dataset` is not yet bound to a cache runtime. Wait for the components to start:

```shell
$ kubectl get pod -o wide
NAME                     READY   STATUS    RESTARTS   AGE   IP            NODE                     NOMINATED NODE   READINESS GATES
mooncake-demo-master-0   1/1     Running   0          15s   10.244.2.13   fluid-mooncake-worker    <none>           <none>
mooncake-demo-worker-0   1/1     Running   0          15s   10.244.2.14   fluid-mooncake-worker    <none>           <none>
mooncake-demo-worker-1   1/1     Running   0          15s   10.244.1.7    fluid-mooncake-worker2   <none>           <none>
```

**Check the Dataset status again**

```shell
$ kubectl get dataset
NAME            UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
mooncake-demo   2.00GiB          0B       2.00GiB          0.0                 Bound   11s
```

`PHASE` is now `Bound` and the cache service is ready to use.

> Note: as with other runtimes, Fluid also creates a PV and PVC for this Dataset, and they show as Bound. But since there is no client component, there is no FUSE mount point, so **application pods must not mount this PVC**. See [FAQ](#faq) at the end of this document.

## Accessing the Cache

Mooncake applications read and write data by connecting directly to the master service through its Python client, not through a mount point.

**Review the application to be created**

```shell
$ cat<<EOF >mooncake-client.yaml
apiVersion: v1
kind: Pod
metadata:
  name: mooncake-client-1
spec:
  nodeName: fluid-mooncake-worker2
  containers:
    - name: client
      image: btxu/mooncake:v3@sha256:067614b70d25b496e3edc3480747d558ee8a364ef47a67f669f5d96ca5098552
      imagePullPolicy: IfNotPresent
      command: ["sleep", "infinity"]
      env:
        - name: POD_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
EOF
```

Note that this pod's `spec` has no `volumes` or `volumeMounts` — this is the main difference between this example and the other runtime examples.

A few notes:

- This example reuses the Mooncake image as the client environment since it already ships the Python client library. `command` is overridden with `sleep infinity` so the image's default entrypoint does not start it as a master or worker.
- `POD_IP` is injected via the downward API. The client's `local_hostname` must be the pod's own IP so that other nodes can connect back to fetch data.
- `nodeName` pins the pod to a specific node so that the cross-node read later is deterministic: the writer stays on one node and the reader on another. Replace `fluid-mooncake-worker2` with a real node name from your cluster (`kubectl get nodes`), and make sure it differs from the node used by client-2 below.

**Start the application and write data**

```shell
$ kubectl apply -f mooncake-client.yaml
$ kubectl exec -it mooncake-client-1 -- python3
```

```python
import os, hashlib
from mooncake.store import MooncakeDistributedStore

MASTER = "mooncake-demo-master-0.svc-mooncake-demo-master"

store = MooncakeDistributedStore()
store.setup(
    local_hostname=os.environ["POD_IP"],
    metadata_server=f"http://{MASTER}:8080/metadata",
    master_server_addr=f"{MASTER}:50051",
    global_segment_size=0,
    local_buffer_size=128 * 1024 * 1024,
    protocol="tcp",
    rdma_devices="",
)

payload = os.urandom(4 * 1024 * 1024)
store.put("demo_key", payload)
got = store.get("demo_key")
print("put md5:", hashlib.md5(payload).hexdigest())
print("get md5:", hashlib.md5(got).hexdigest())
print("match:", hashlib.md5(payload).hexdigest() == hashlib.md5(got).hexdigest())
```

> The example connects through the master pod's stable DNS name `mooncake-demo-master-0.svc-mooncake-demo-master`, which pins the client to one specific replica. The Service name `svc-mooncake-demo-master` works just as well here: Fluid creates the component Service as headless (`clusterIP: None`) and declares no `ports` on it, but that only affects SRV records — the A record still resolves straight to the backing pod IPs, and the client connects to container ports `8080` and `50051` directly. The per-pod name is the safer habit because it stays pinned to a single replica if the master is ever scaled out.

Output of `setup()` (key lines only):

```
I0815 05:30:15.076617 transfer_metadata_plugin.cpp:1293] Found active interface eth0 with IP 10.244.1.8
I0815 05:30:15.077410 client_service.cpp:747] Transfer engine auto discovery is disabled for protocol: tcp
I0815 05:30:15.078089 real_client.cpp:734] Successfully created client on port 12699 after 1 attempt(s)
I0815 05:30:15.079909 real_client.cpp:767] Registering local memory: 134217728 bytes
I0815 05:30:15.080171 real_client.cpp:932] Global segment size is 0, skip mounting segment
0
```

> With `global_segment_size` set to 0, the client log shows `Global segment size is 0, skip mounting segment`, meaning the client no longer allocates a local memory segment and all data is held by the Fluid-managed workers.
>
> The `http=404 body: metadata not found` message during startup is expected: the client is registering itself with the metadata service for the first time, so the corresponding key does not exist yet.

Read/write result:

```
put md5: e114c0f1fb62bb7b1df45dbb1ab1d201
get md5: e114c0f1fb62bb7b1df45dbb1ab1d201
match: True
```

> Note: Mooncake's `put` silently skips a key that **already exists** — it does not overwrite the old data, and it still returns `0`, so the return value gives nothing away. If you run the snippet above a second time (or also run a write in client-2 later on), the second `put` has no effect and `get` returns the content written the first time, producing `match: False`. This is not a cache failure. Before re-running this example, use a different key name, or delete and recreate the CacheRuntime so the workers' cache is cleared.

**Reading across pods and nodes**

First exit the Python session in the first client so the writer process terminates:

```python
>>> exit()
```

Then create a second client pod, with `nodeName` pointing at a **different** node (client-1 is on `fluid-mooncake-worker2`, so this one uses `fluid-mooncake-worker`):

```shell
$ cat<<EOF >mooncake-client-2.yaml
apiVersion: v1
kind: Pod
metadata:
  name: mooncake-client-2
spec:
  nodeName: fluid-mooncake-worker
  containers:
    - name: client
      image: btxu/mooncake:v3@sha256:067614b70d25b496e3edc3480747d558ee8a364ef47a67f669f5d96ca5098552
      imagePullPolicy: IfNotPresent
      command: ["sleep", "infinity"]
      env:
        - name: POD_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
EOF

$ kubectl apply -f mooncake-client-2.yaml
$ kubectl get pod mooncake-client-1 mooncake-client-2 -o wide
NAME                READY   STATUS    RESTARTS   AGE     IP            NODE                     NOMINATED NODE   READINESS GATES
mooncake-client-1   1/1     Running   0          3m44s   10.244.1.8    fluid-mooncake-worker2   <none>           <none>
mooncake-client-2   1/1     Running   0          79s     10.244.2.15   fluid-mooncake-worker    <none>           <none>
```

The second pod runs a brand-new Python process, so open a session in it and repeat the imports and the full `MooncakeDistributedStore` initialization shown above — the `setup` parameters are identical, and `POD_IP` again resolves to this pod's own IP:

```shell
$ kubectl exec -it mooncake-client-2 -- python3
```

Once `store` and `hashlib` are initialized, read and verify:

```python
got = store.get("demo_key")
print("len:", len(got))
print("md5:", hashlib.md5(got).hexdigest())
```

Output:

```
len: 4194304
md5: e114c0f1fb62bb7b1df45dbb1ab1d201
```

The md5 matches the writer exactly. Note that by this point the client process that wrote the data has already exited, and the reader is on a different node — showing that the cached data is held by the Fluid-managed workers, depending neither on the client process that wrote it nor on the node it ran on.

## Inspecting Cache Status

```shell
$ kubectl get dataset mooncake-demo -o yaml
```

> Note: Fluid runs the ReportSummary script periodically, so the Dataset `status` does not update the moment a write completes — in practice it takes about a minute to refresh. Until then you will still see `cached: 0B` and `fileNum: "0"`, which is expected; just check again shortly. To see the cache system's live state directly, run the script inside the master pod:
>
> ```shell
> $ kubectl exec mooncake-demo-master-0 -- bash -c /reportSummary.sh
> ```

```yaml
  cacheStates:
    cacheCapacity: 2.00GiB
    cacheHitRatio: "0"
    cached: 4.00MiB
    cachedPercentage: "0.2"
    fileNum: "1"
    ufsTotal: 2.00GiB
  conditions:
  - lastTransitionTime: "2026-08-15T05:28:03Z"
    lastUpdateTime: "2026-08-15T05:28:03Z"
    message: The ddc runtime is ready.
    reason: DatasetReady
    status: "True"
    type: Ready
```

`cached: 4.00MiB` matches the 4 MiB written above, and `fileNum: 1` corresponds to the single key. `cacheCapacity: 2.00GiB` comes from the 1Gi `tieredStore` quota on each of the two workers.

The remaining two fields need to be read in the context of having no UFS:

- `ufsTotal` normally reports the total data size in the underlying storage. Mooncake has no UFS, so the example image's `reportSummary.sh` fills the field with the total cache capacity instead (`UFS_TOTAL="$CACHE_CAPACITY"` in the script), which is why it equals `cacheCapacity`. This is a choice made by the script, not an anomaly.
- `cacheHitRatio` is approximated by the script from the master's `Get` request success rate. Note that `Requests (Success/Total per sec)` in the master's `/metrics/summary` is a **per-second instantaneous rate**, not a cumulative counter, so if no Get is in flight when the script samples, it reads `0.00/0.00` — which is why this field usually shows `0`. It reflects request success rate at the sampling instant, not a strict cache hit ratio.

All of these values are collected and reported by the ReportSummary script configured in the CacheRuntimeClass; Fluid only passes them through. For the meaning of each field and the required output format, see the [Generic Cache System Integration Guide](../../dev/generic_cache_runtime_integration.md).

## FAQ

### An application pod is stuck in ContainerCreating

**Symptom**: an application pod that mounts this Dataset's PVC never starts, and its events show `FailedMount`.

```shell
$ kubectl describe pod <pod-name>
...
Events:
  Type     Reason       Age   From               Message
  ----     ------       ----  ----               -------
  Normal   Scheduled    60s   default-scheduler  Successfully assigned default/<pod-name> to fluid-mooncake-worker
  Warning  FailedMount  29s   kubelet            MountVolume.SetUp failed for volume "default-mooncake-demo" : rpc error: code = Internal desc = timeout waiting for FUSE mount point to be ready
```

**Cause**: this example's CacheRuntimeClass declares no client component, so there is no FUSE mount point. The CSI plugin waits for a mount point that never appears until it times out, which produces the `timeout waiting for FUSE mount point to be ready` above. Fluid still creates the PVC/PV and reports them as Bound, but application pods cannot consume them through `volumeMounts`.

**Resolution**: remove the PVC mount from the application pod, connect to the cache service directly as shown above, and delete the stuck pod.

```shell
$ kubectl delete pod <pod-name>
```

## Cleanup

```shell
$ kubectl delete -f mooncake-client-2.yaml
$ kubectl delete -f mooncake-client.yaml
$ kubectl delete -f mooncake-dataset-runtime.yaml
$ kubectl delete -f mooncake-cacheruntimeclass.yaml
```
