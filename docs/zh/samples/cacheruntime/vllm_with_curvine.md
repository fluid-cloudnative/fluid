# 示例 - 使用 CacheRuntime（Curvine）缓存模型权重并运行 vLLM 推理服务

本示例演示如何用 Fluid 的 CacheRuntime 和 [Curvine](https://github.com/curvine-io/curvine) 缓存大模型权重，并让 [vLLM](https://github.com/vllm-project/vllm) 推理服务直接从 Fluid 创建的 PVC 加载模型。

推理服务通常把模型权重放在对象存储里。每个副本启动时都要先把权重下载到本地盘，模型越大、副本越多，这一步越慢，也越占磁盘。使用 Fluid 后：

- 权重只需预热到缓存一次，所有副本通过同一个 PVC 读取，不再各自下载。
- vLLM 把模型当作本地目录使用，不需要感知对象存储，也不需要在镜像或启动脚本里加下载逻辑。

## 前提条件

1. 已在 Kubernetes 集群上安装 Fluid，并开启了 CacheRuntime。请参考[安装文档](../../userguide/install.md)，CacheRuntime 默认关闭，需要在 Helm 安装时加上 `--set runtime.cacheruntime.enabled=true`。安装后 `kubectl get pod -n fluid-system` 应能看到 `cacheruntime-controller`。
2. 集群中至少有一个带 NVIDIA GPU 的节点，并已安装 [NVIDIA device plugin](https://github.com/NVIDIA/k8s-device-plugin)，节点上报了 `nvidia.com/gpu` 资源。
3. 有一个可用的 MinIO 或 S3 兼容的对象存储服务。
4. 本地可以执行 `kubectl` 和 `curl`（7.75.0 及以上，需要 `--aws-sigv4`）。

本示例使用镜像 `vllm/vllm-openai:v0.33.0` 和 `curvine/curvine:v0.3.6-alpha` 验证。所有资源都创建在 `default` 命名空间，文中的命令也假定当前命名空间是 `default`。

## 概览

1. 部署 MinIO、创建存储桶、创建 `CacheRuntimeClass`
2. 把模型上传到对象存储
3. 创建 `Dataset` 和 `CacheRuntime`
4. 运行 DataLoad 把模型预热到缓存
5. 部署 vLLM，挂载 Fluid 创建的 PVC
6. 调用推理接口验证
7. 扩容 vLLM 副本，共享同一份缓存

## 步骤 1：准备 MinIO 和 CacheRuntimeClass

这一步和 [使用 Curvine 作为 CacheRuntime 进行数据缓存](./curvine_cache_runtime.md) 完全相同。请按该文档的步骤 1、步骤 2 和步骤 4 依次完成：

- 步骤 1：部署 MinIO（同时创建凭证 Secret `curvine-secret`）。部署后先执行 `kubectl wait --for=condition=Ready pod -l app=minio --timeout=120s`，等 MinIO 就绪再做下一步，否则创建存储桶的 Job 可能在 MinIO 启动期间重试耗尽而失败
- 步骤 2：创建存储桶 `test`
- 步骤 4：创建 `CacheRuntimeClass`，名称为 `curvine-demo`

> 说明：示例中的 MinIO 没有挂载持久卷，Curvine 的元数据和缓存也使用 `emptyDir`，仅用于演示。MinIO 的 Pod 重建后需要重新上传模型，Curvine 的 Pod 重建后需要重新预热。生产环境请使用持久化的对象存储保存模型。

如果使用已有的 S3 兼容对象存储，可以跳过 MinIO 的部署。此时需要相应修改：Secret 中的访问凭证（保持名称 `curvine-secret` 和键名 `access-key`、`secret-key`，或同步修改 Dataset 中的引用），步骤 3 的 Dataset 中的 `mountPoint`、`endpoint_url` 和 `region_name`，以及步骤 2 中上传模型的地址、存储桶和凭证。

## 步骤 2：上传模型到对象存储

本示例使用 [Qwen2.5-0.5B-Instruct](https://modelscope.cn/models/Qwen/Qwen2.5-0.5B-Instruct)，权重文件约 1GB。先把模型下载到本地：

```shell
$ pip install modelscope
$ modelscope download --model Qwen/Qwen2.5-0.5B-Instruct --local_dir ./Qwen2.5-0.5B-Instruct
```

把 MinIO 的端口转发到本地，然后把模型目录下的文件逐个上传到存储桶 `test` 的 `Qwen2.5-0.5B-Instruct/` 目录。上传放在子 shell 中执行，结束后仍停留在当前目录；任何一个文件上传失败都会立即中止：

```shell
$ kubectl port-forward svc/minio 19000:9000 &
$ until curl -s -o /dev/null http://127.0.0.1:19000/minio/health/live; do sleep 1; done   # 等待端口转发建立

$ (
    cd Qwen2.5-0.5B-Instruct || exit 1
    for f in *; do
      [ -f "$f" ] || continue
      curl -sSf -X PUT -T "$f" \
        -H "x-amz-content-sha256: $(sha256sum "$f" | cut -d' ' -f1)" \
        --aws-sigv4 "aws:amz:us-east-1:s3" --user minioadmin:minioadmin \
        "http://127.0.0.1:19000/test/Qwen2.5-0.5B-Instruct/$f" && echo "uploaded $f" || exit 1
    done
  )

$ kill %1   # 结束端口转发
```

应看到每个文件各一行 `uploaded ...`。如果缺了 `config.json` 等文件，vLLM 启动时会报 `Invalid repository ID or local directory specified`。

如果本机配置了 HTTP 代理，请给 `curl` 加上 `--noproxy '*'`。也可以使用 `mc`、`aws s3 cp` 等任意 S3 客户端上传。

## 步骤 3：创建 Dataset 和 CacheRuntime

```shell
$ cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: curvine-demo
spec:
  accessModes: ["ReadWriteMany"]
  mounts:
    - mountPoint: "s3://test"
      name: minio
      options:
        endpoint_url: "http://minio:9000"
        region_name: "us-east-1"
        path_style: "true"
      encryptOptions:
        - name: access
          valueFrom:
            secretKeyRef:
              name: curvine-secret
              key: access-key
        - name: secret
          valueFrom:
            secretKeyRef:
              name: curvine-secret
              key: secret-key
EOF
$ kubectl create -f dataset.yaml
```

```shell
$ cat<<EOF >cacheruntime.yaml
apiVersion: data.fluid.io/v1alpha1
kind: CacheRuntime
metadata:
  name: curvine-demo
spec:
  runtimeClassName: curvine-demo
  volumes:
    - name: curvine-logs
      emptyDir: {}
    - name: curvine-master-data
      emptyDir: {}
  master:
    replicas: 1
    options:
      format_master: "true"
      rpc_port: "8995"
      journal_port: "8996"
      web_port: "9000"
      meta_dir: "/app/curvine/data/meta"
      journal_dir: "/app/curvine/data/journal"
    volumeMounts:
      - name: curvine-logs
        mountPath: /app/curvine/logs
      - name: curvine-master-data
        mountPath: /app/curvine/data
  worker:
    replicas: 1
    options:
      format_worker: "true"
      rpc_port: "8997"
      web_port: "9001"
      dir_reserved: "0"
    volumeMounts:
      - name: curvine-logs
        mountPath: /app/curvine/logs
    tieredStore:
      levels:
        - low: "0.5"
          high: "0.8"
          emptyDir:
            quota: 4Gi
  client:
    options:
      debug: "false"
EOF
$ kubectl create -f cacheruntime.yaml
```

和基础示例相比，这里只改了一处：`worker.tieredStore` 的 `quota` 从 1Gi 调大到 4Gi。缓存容量必须大于模型的总大小，并为水位线留出余量；使用更大的模型时请相应调大。关于缓存介质的选择，请参考 [CacheRuntime 分层存储](../../userguide/cache_runtime_tieredstore.md)。

等待 Dataset 进入 `Bound` 阶段，并确认 Curvine 组件都在运行：

```shell
$ kubectl get dataset curvine-demo
NAME           UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
curvine-demo   0                0.00B    4.00GiB          0                   Bound   38s

$ kubectl get pod -l cacheruntime.fluid.io/name=curvine-demo
NAME                    READY   STATUS    RESTARTS   AGE
curvine-demo-master-0   1/1     Running   0          37s
curvine-demo-worker-0   1/1     Running   0          37s
```

Dataset 绑定后，Fluid 会创建同名的 PVC `curvine-demo`，供应用挂载：

```shell
$ kubectl get pvc curvine-demo
NAME           STATUS   VOLUME                 CAPACITY   ACCESS MODES   STORAGECLASS   AGE
curvine-demo   Bound    default-curvine-demo   100Pi      RWX            fluid          40s
```

## 步骤 4：预热模型到缓存

```shell
$ cat<<EOF >dataload.yaml
apiVersion: data.fluid.io/v1alpha1
kind: DataLoad
metadata:
  name: curvine-dataload
spec:
  dataset:
    name: curvine-demo
    namespace: default
  target:
    - path: /minio
EOF
$ kubectl create -f dataload.yaml
$ kubectl wait dataload/curvine-dataload --for=jsonpath='{.status.phase}'=Complete --timeout=300s
```

`target.path` 中的 `/minio` 对应 Dataset 里 `mounts[].name`。DataLoad 的 `PHASE` 变为 `Complete` 表示预热完成。此时 Dataset 的 `CACHED` 一列应接近模型的总大小，可以作为辅助确认：

```shell
$ kubectl get dataload curvine-dataload
NAME               DATASET        PHASE      AGE   DURATION
curvine-dataload   curvine-demo   Complete   4s    4s

$ kubectl get dataset curvine-demo
NAME           UFS TOTAL SIZE   CACHED      CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
curvine-demo   0                953.28MiB   4.00GiB          0                   Bound   68s
```

> 说明：在 Curvine v0.3.6-alpha 上，`UFS TOTAL SIZE` 和 `CACHED PERCENTAGE` 两列显示为 0，这两个值来自 Curvine 的缓存摘要报告。

## 步骤 5：部署 vLLM

```shell
$ cat<<EOF >vllm.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: vllm
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vllm
  template:
    metadata:
      labels:
        app: vllm
    spec:
      enableServiceLinks: false
      containers:
        - name: vllm
          image: vllm/vllm-openai:v0.33.0
          imagePullPolicy: IfNotPresent
          args:
            - /data/minio/Qwen2.5-0.5B-Instruct
            - --served-model-name=qwen
            - --port=8000
            - --gpu-memory-utilization=0.4
            - --max-model-len=4096
          ports:
            - containerPort: 8000
          resources:
            limits:
              nvidia.com/gpu: 1
          readinessProbe:
            httpGet:
              path: /health
              port: 8000
            periodSeconds: 5
            failureThreshold: 120
          volumeMounts:
            - name: model
              mountPath: /data
              readOnly: true
            - name: shm
              mountPath: /dev/shm
      volumes:
        - name: model
          persistentVolumeClaim:
            claimName: curvine-demo
        - name: shm
          emptyDir:
            medium: Memory
            sizeLimit: 2Gi
---
apiVersion: v1
kind: Service
metadata:
  name: vllm-svc
spec:
  selector:
    app: vllm
  ports:
    - port: 8000
      targetPort: 8000
EOF
$ kubectl create -f vllm.yaml
$ kubectl rollout status deployment/vllm --timeout=600s
```

关键字段说明：

- **模型路径**：PVC 挂载在 `/data`，Dataset 的挂载点名称是 `minio`，模型上传到了存储桶的 `Qwen2.5-0.5B-Instruct/` 目录，所以容器内的模型路径是 `/data/minio/Qwen2.5-0.5B-Instruct`。镜像的入口是 `vllm serve`，第一个参数就是模型路径。
- **`volumes.model`**：直接引用 Fluid 创建的 PVC `curvine-demo`，以只读方式挂载。
- **`enableServiceLinks: false` 和 Service 名称**：如果同一命名空间里有名为 `vllm` 的 Service，且 Pod 开启了服务环境变量注入（Kubernetes 的默认行为），Pod 里会出现 `VLLM_PORT=tcp://...` 这样的环境变量，与 vLLM 自身的配置项 `VLLM_PORT` 冲突，vLLM 会启动失败并报 `VLLM_PORT ... appears to be a URI`。本示例关闭了服务环境变量注入，并把 Service 命名为 `vllm-svc`，两者任取其一即可避免该问题。
- **`/dev/shm`**：vLLM 的多进程通信需要共享内存，这里挂载了一个内存盘。
- **`--gpu-memory-utilization`**：vLLM 预占的显存比例。本示例的模型很小，设为 0.4；请按模型大小和显卡规格调整。

## 步骤 6：验证

确认模型目录是 Curvine 的 FUSE 挂载点，并且 vLLM 已完成模型加载（日志中的耗时仅为示例，因环境而异）：

```shell
$ kubectl exec deploy/vllm -- sh -c 'mount | grep " /data "'
curvinefs on /data type fuse (ro,relatime,user_id=0,group_id=0,default_permissions,allow_other)

$ kubectl logs deploy/vllm | grep -E "Loading weights took|Model loading took|Starting vLLM server"
(EngineCore pid=204) INFO [default_loader.py:506] Loading weights took 0.15 seconds
(EngineCore pid=204) INFO [model_runner.py:428] Model loading took 0.93 GiB memory and 1.909626 seconds
(APIServer pid=1) INFO [entry.py:139] Starting vLLM server on http://0.0.0.0:8000
```

调用 OpenAI 兼容接口：

```shell
$ kubectl run curl-test --rm -i --restart=Never --image=curlimages/curl:8.22.0 -q -- \
    -s http://vllm-svc:8000/v1/chat/completions \
    -H 'Content-Type: application/json' \
    -d '{"model":"qwen","messages":[{"role":"user","content":"用一句话介绍一下 Kubernetes。"}],"max_tokens":80}'
```

返回内容类似：

```json
{
  "object": "chat.completion",
  "model": "qwen",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Kubernetes（简称K8s）是一种开源的容器编排系统，它提供了一个平台来管理和调度应用在多个集群中的容器实例。..."
      }
    }
  ],
  "usage": {"prompt_tokens": 34, "total_tokens": 114, "completion_tokens": 80}
}
```

## 步骤 7：扩容 vLLM 副本

```shell
$ kubectl scale deploy/vllm --replicas=2
$ kubectl rollout status deployment/vllm --timeout=600s
$ kubectl get pod -l app=vllm
NAME                   READY   STATUS    RESTARTS   AGE
vllm-d969fc667-bwhsc   1/1     Running   0          59s
vllm-d969fc667-z4plq   1/1     Running   0          3h32m
```

新副本挂载同一个 PVC，和已有副本共用 Curvine 里的同一份模型，不需要像「每个副本各自下载一份」的做法那样再占用一份本地盘。可以在新副本的日志里查看 `Loading weights took` 一行，确认它已完成加载。

> 说明：每个副本需要一个 `nvidia.com/gpu`。如果集群只有一张 GPU，可以为 NVIDIA device plugin 开启 [time-slicing](https://github.com/NVIDIA/k8s-device-plugin#shared-access-to-gpus) 让多个副本共享这张卡以验证功能。开启后所有副本占用的显存之和不能超过这张卡的物理显存；如果在 time-slicing 配置中设置了 `renameByDefault: true`，资源名会变成 `nvidia.com/gpu.shared`，需要相应修改 Deployment 中的资源请求。time-slicing 不隔离显存，也不增加算力，不适合作为生产方案。

## 环境清理

```shell
$ kubectl delete -f vllm.yaml
$ kubectl delete -f dataload.yaml
$ kubectl delete -f dataset.yaml
$ kubectl delete -f cacheruntime.yaml
$ kubectl delete cacheruntimeclass curvine-demo
$ kubectl delete -f minio_create_bucket.yaml
$ kubectl delete -f minio.yaml
```
