# 示例 - 使用 CacheRuntime 部署 Mooncake

[Mooncake](https://github.com/kvcache-ai/Mooncake) 是一个面向大模型推理场景的分布式 KVCache 存储系统。与 Alluxio、JuiceFS 等缓存系统不同，Mooncake 不提供 POSIX 挂载语义，应用通过其自带的客户端直接与缓存服务通信，而不是通过挂载点读写文件。

Fluid 的通用 CacheRuntime 支持这类缓存系统：在 CacheRuntimeClass 的 `topology` 中只声明 master 和 worker 组件，不声明负责挂载的 client 组件。本文档演示这一最小实现。

关于无 client 架构的说明，参见[通用缓存系统接入指南](../../dev/generic_cache_runtime_integration.md)。

## 前提条件

在运行该示例之前，请参考[安装文档](../../userguide/install.md)完成安装，并检查 Fluid 各组件正常运行：

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

通常来说，你会看到 `dataset-controller`、`cacheruntime-controller`、`fluid-webhook`、`fluidapp-controller` 各一个 Pod，以及每个节点一个的 `csi-nodeplugin` Pod 正在运行。其中 `cacheruntime-controller` 是本示例必需的，它由 Helm 安装时的 `--set runtime.cacheruntime.enabled=true` 开启。

> 注意：本示例需要包含 [#6157](https://github.com/fluid-cloudnative/fluid/pull/6157) 的 Fluid 版本。较早版本在 `topology` 中省略 client 组件时 controller 会发生 panic。

### 示例镜像

Fluid 并不发布 Mooncake 镜像。本示例使用的是一个演示镜像，它在 Mooncake 的 Python 发行版之上加入了两个供 Fluid 调用的脚本：

| 路径 | 作用 |
|---|---|
| `/custom-entrypoint.sh` | 组件启动入口，按角色（master/worker）启动对应进程 |
| `/reportSummary.sh` | 采集缓存用量并按 Fluid 要求的 JSON 格式输出 |

该镜像的构建上下文已随本仓库提供，位于 [`samples/mooncake/docker`](../../../../samples/mooncake/docker)，你可以据此构建等价镜像：

```shell
$ docker build -t <your-registry>/mooncake:v3 samples/mooncake/docker
$ docker push <your-registry>/mooncake:v3
```

注意 `apt-get` 和 `pip` 在构建时拉取的都是当时的最新版本，因此重新构建得到的是功能等价的镜像，并不能字节级复现下面的摘要。

**推荐的做法是自行构建镜像，并替换下文各处清单中的镜像地址。** 为方便试用，也提供了一份预构建镜像，并固定到不可变的镜像摘要，这样即使标签被覆盖，本示例仍然可用：

```
btxu/mooncake:v3@sha256:067614b70d25b496e3edc3480747d558ee8a364ef47a67f669f5d96ca5098552
```

国内网络环境下也可使用阿里云镜像。它是同一 manifest 的副本，摘要与上面完全相同：

```
crpi-4hkqof7tc9brc6d5.cn-hongkong.personal.cr.aliyuncs.com/mooncake1314/mooncake:v3@sha256:067614b70d25b496e3edc3480747d558ee8a364ef47a67f669f5d96ca5098552
```

> 注意：以上两个仓库均为本示例作者的个人账号，并非项目管控的基础设施，不提供任何可用性保证。它们仅用于快速试用本示例；正式使用请基于 `samples/mooncake/docker` 自行构建。

这两个脚本并非该镜像独有：你也可以基于任意 Mooncake 发行版自行构建等价镜像，只要它提供同样的两个入口即可，具体约定参见[通用缓存系统接入指南](../../dev/generic_cache_runtime_integration.md)。

## 运行示例

### 创建 CacheRuntimeClass

**查看待创建的 CacheRuntimeClass 资源对象**

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

`CacheRuntimeClass` 是 Fluid 定义的 CRD，用于描述一类缓存系统如何在 Kubernetes 上运行——包括各组件使用的镜像、启动方式、端口，以及 Fluid 如何采集其运行状态。

本示例中需要留意的几点：

- `topology` 下只声明了 `master` 和 `worker`。master、worker、client 三个组件在 API 中均为可选字段，Mooncake 无需 client 组件。
- 注意 `fileSystemType` 和 `topology` 是顶层字段，不在 `spec` 之下。
- master 组件暴露三个端口：`50051` 用于 RPC，`8080` 用于元数据服务，`9003` 用于 metrics。应用将直接连接前两个。
- `reportSummary` 指向镜像中的 `/reportSummary.sh`，Fluid 会周期性地在组件 Pod 中执行它，并将结果更新到 Dataset 的 `status` 字段。其输出格式要求参见[通用缓存系统接入指南](../../dev/generic_cache_runtime_integration.md)。

**创建 CacheRuntimeClass 资源对象**

```shell
$ kubectl apply -f mooncake-cacheruntimeclass.yaml
cacheruntimeclass.data.fluid.io/mooncake-demo created

$ kubectl get cacheruntimeclass
NAME            AGE
mooncake-demo   0s
```

> 提示：如果之后修改了 CacheRuntimeClass，需要删除并重建 CacheRuntime 才会生效。CacheRuntime 在创建时就已将当时的 CacheRuntimeClass 渲染为工作负载，后续修改不会回溯更新，仅删除 Pod 也无效。

### 创建 Dataset 与 CacheRuntime

**查看待创建的资源对象**

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

`Dataset` 描述数据集本身，`CacheRuntime` 描述为该数据集提供缓存服务的运行时实例——本示例中即一个 master 副本和两个 worker 副本。

关于 `mounts` 字段：Mooncake 不挂载任何底层存储（UFS），数据由客户端直接写入缓存，因此这里的 `mooncakefs:///` 只是一个占位挂载点。Fluid 会把 `mounts` 的内容透传到 CacheRuntime 的配置 ConfigMap 中，供组件容器读取；但由于本示例的 CacheRuntimeClass 未声明 `mountUfs` 执行入口，Fluid 不会真正执行任何挂载动作（详见[通用缓存系统接入指南](../../dev/generic_cache_runtime_integration.md)中关于"可以不设置 MountUFS"的说明）。API 上 `mounts` 是可选字段，但一旦声明就至少需要一项。

**创建资源对象**

```shell
$ kubectl apply -f mooncake-dataset-runtime.yaml
dataset.data.fluid.io/mooncake-demo created
cacheruntime.data.fluid.io/mooncake-demo created
```

**查看 Dataset 资源对象状态**

```shell
$ kubectl get dataset
NAME            UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE      AGE
mooncake-demo                                                                  NotBound   6s
```

如上所示，`PHASE` 属性值为 `NotBound`，这意味着该 `Dataset` 资源对象目前还未与缓存运行时绑定。等待各组件启动：

```shell
$ kubectl get pod -o wide
NAME                     READY   STATUS    RESTARTS   AGE   IP            NODE                     NOMINATED NODE   READINESS GATES
mooncake-demo-master-0   1/1     Running   0          15s   10.244.2.13   fluid-mooncake-worker    <none>           <none>
mooncake-demo-worker-0   1/1     Running   0          15s   10.244.2.14   fluid-mooncake-worker    <none>           <none>
mooncake-demo-worker-1   1/1     Running   0          15s   10.244.1.7    fluid-mooncake-worker2   <none>           <none>
```

**再次查看 Dataset 资源对象状态**

```shell
$ kubectl get dataset
NAME            UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
mooncake-demo   2.00GiB          0B       2.00GiB          0.0                 Bound   11s
```

此时 `PHASE` 已变为 `Bound`，缓存服务可以使用了。

> 注意：与其他 Runtime 一样，Fluid 也会为该 Dataset 创建 PV 和 PVC，且状态显示为 Bound。但由于没有 client 组件，不存在 FUSE 挂载点，**业务 Pod 不应挂载该 PVC**。详见文末[常见问题](#常见问题)。

## 访问缓存

Mooncake 的应用通过其 Python 客户端直连 master 服务读写数据，不经过挂载点。

**查看待创建的应用**

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

注意该 Pod 的 `spec` 中没有 `volumes` 和 `volumeMounts`——这是本示例与其他 Runtime 示例最主要的区别。

几点说明：

- 本示例直接复用 Mooncake 镜像作为客户端环境，它已包含 Python 客户端库；`command` 覆盖为 `sleep infinity`，避免镜像的默认入口把它启动成 master 或 worker。
- `POD_IP` 通过 downward API 注入，客户端的 `local_hostname` 必须是 Pod 自身 IP，其他节点才能回连取数。
- `nodeName` 显式指定节点，是为了让后面的跨节点读取有确定的效果：写入方固定在一个节点，读取方固定在另一个节点。请把 `fluid-mooncake-worker2` 换成你集群中的实际节点名（`kubectl get nodes`），并确保与后面 client-2 使用的节点不同。

**启动应用并写入数据**

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

> 本示例使用 master Pod 的稳定 DNS 名 `mooncake-demo-master-0.svc-mooncake-demo-master` 作为连接地址，从而固定访问某一个副本。这里用 Service 名 `svc-mooncake-demo-master` 同样可行：Fluid 创建的组件 Service 是 headless 的（`clusterIP: None`）且未声明 `ports`，但这只影响 SRV 记录，A 记录仍会直接解析到后端 Pod IP，客户端随后直连容器的 `8080` 和 `50051` 端口。之所以推荐使用 Pod 级 DNS 名，是因为它在 master 扩容后仍然固定指向单个副本。

`setup()` 的输出（截取关键部分）：

```
I0815 05:30:15.076617 transfer_metadata_plugin.cpp:1293] Found active interface eth0 with IP 10.244.1.8
I0815 05:30:15.077410 client_service.cpp:747] Transfer engine auto discovery is disabled for protocol: tcp
I0815 05:30:15.078089 real_client.cpp:734] Successfully created client on port 12699 after 1 attempt(s)
I0815 05:30:15.079909 real_client.cpp:767] Registering local memory: 134217728 bytes
I0815 05:30:15.080171 real_client.cpp:932] Global segment size is 0, skip mounting segment
0
```

> `global_segment_size` 设为 0 时，客户端日志中会出现 `Global segment size is 0, skip mounting segment`，表示客户端不再分配本地内存段，数据全部由 Fluid 管理的 worker 承载。
>
> 启动过程中出现的 `http=404 body: metadata not found` 属于正常现象：客户端首次向元数据服务注册自身时，对应的 key 尚不存在。

读写结果：

```
put md5: e114c0f1fb62bb7b1df45dbb1ab1d201
get md5: e114c0f1fb62bb7b1df45dbb1ab1d201
match: True
```

> 注意：Mooncake 的 `put` 对**已存在**的 key 是静默跳过的——不覆盖旧数据，且仍然返回 `0`，从返回值上看不出区别。因此如果你重复执行上面这段代码（或在后面的 client-2 里也执行了写入），第二次的 `put` 不会生效，`get` 读回的仍是第一次写入的内容，于是出现 `match: False`。这不是缓存故障。重跑本示例前，请换一个 key 名，或删除并重建 CacheRuntime 让 worker 的缓存清空。

**跨 Pod 跨节点读取**

先退出第一个客户端的 Python 交互环境，让写入方进程结束：

```python
>>> exit()
```

再创建第二个客户端 Pod，`nodeName` 指向**另一个**节点（client-1 在 `fluid-mooncake-worker2`，这里用 `fluid-mooncake-worker`）：

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

第二个 Pod 中是一个全新的 Python 进程，需要先进入它的 Python 环境，并重复上文的导入语句和完整的 `MooncakeDistributedStore` 初始化（`setup` 参数与上文完全相同，`POD_IP` 同样取当前 Pod 自身的 IP）：

```shell
$ kubectl exec -it mooncake-client-2 -- python3
```

`store` 和 `hashlib` 初始化完成后，再做读取和校验：

```python
got = store.get("demo_key")
print("len:", len(got))
print("md5:", hashlib.md5(got).hexdigest())
```

输出：

```
len: 4194304
md5: e114c0f1fb62bb7b1df45dbb1ab1d201
```

md5 与写入方完全一致。需要强调的是，此时写入数据的那个客户端进程已经退出，且读取方位于另一个节点上——这说明缓存数据由 Fluid 管理的 worker 承载，既不依赖写入它的客户端进程，也不依赖所在节点。

## 查看缓存状态

```shell
$ kubectl get dataset mooncake-demo -o yaml
```

> 注意：Fluid 是周期性执行 ReportSummary 脚本的，写入完成后 Dataset 的 `status` 不会立刻更新。实测约需 1 分钟左右才会刷新，在此之前看到的仍是 `cached: 0B`、`fileNum: "0"`，属于正常现象，稍等再查即可。若想确认缓存系统侧的即时状态，可以直接在 master Pod 中执行该脚本：
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

`cached: 4.00MiB` 与上文写入的 4 MiB 数据一致，`fileNum: 1` 对应写入的一个 key。`cacheCapacity: 2.00GiB` 则来自两个 worker 各 1Gi 的 `tieredStore` 配额。

其余两个字段需要结合无 UFS 的场景理解：

- `ufsTotal` 在其他 Runtime 中表示底层存储的数据总量。Mooncake 没有 UFS，示例镜像的 `reportSummary.sh` 直接用缓存总容量填充该字段（脚本中即 `UFS_TOTAL="$CACHE_CAPACITY"`），因此它与 `cacheCapacity` 相等。这是脚本的上报选择，不是异常。
- `cacheHitRatio` 由脚本从 master 的 `Get` 请求成功率近似估算。注意 master `/metrics/summary` 中的 `Requests (Success/Total per sec)` 是**每秒瞬时速率**而非累计计数，采样时若没有正在进行的 Get 请求，读到的就是 `0.00/0.00`，因此该字段通常显示为 `0`。它反映的是采样瞬间的请求成功率，并不是严格意义上的缓存命中率。

这些数据均由 CacheRuntimeClass 中配置的 ReportSummary 脚本采集上报，Fluid 只做透传，各字段的语义和输出格式要求参见[通用缓存系统接入指南](../../dev/generic_cache_runtime_integration.md)。

## 常见问题

### 业务 Pod 一直处于 ContainerCreating

**现象**：业务 Pod 挂载了该 Dataset 对应的 PVC 后一直无法启动，事件中出现 `FailedMount`。

```shell
$ kubectl describe pod <pod-name>
...
Events:
  Type     Reason       Age   From               Message
  ----     ------       ----  ----               -------
  Normal   Scheduled    60s   default-scheduler  Successfully assigned default/<pod-name> to fluid-mooncake-worker
  Warning  FailedMount  29s   kubelet            MountVolume.SetUp failed for volume "default-mooncake-demo" : rpc error: code = Internal desc = timeout waiting for FUSE mount point to be ready
```

**原因**：本示例的 CacheRuntimeClass 未声明 client 组件，因此不存在 FUSE 挂载点。CSI 插件会一直等待挂载点就绪直到超时，于是报出上面的 `timeout waiting for FUSE mount point to be ready`。虽然 Fluid 仍会创建 PVC/PV 且状态为 Bound，但业务 Pod 无法通过 `volumeMounts` 使用它。

**处理**：从业务 Pod 中移除该 PVC 的挂载，改为按上文方式直连缓存服务，然后删除卡住的 Pod。

```shell
$ kubectl delete pod <pod-name>
```

## 环境清理

```shell
$ kubectl delete -f mooncake-client-2.yaml
$ kubectl delete -f mooncake-client.yaml
$ kubectl delete -f mooncake-dataset-runtime.yaml
$ kubectl delete -f mooncake-cacheruntimeclass.yaml
```
