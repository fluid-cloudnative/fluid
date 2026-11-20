# Example - Cache Model Weights with CacheRuntime (Curvine) and Serve Them with vLLM

This example shows how to cache large model weights with Fluid's CacheRuntime and [Curvine](https://github.com/curvine-io/curvine), and how to run a [vLLM](https://github.com/vllm-project/vllm) inference service that loads the model directly from the PVC created by Fluid.

Inference services usually keep model weights in object storage. Every replica has to download the weights to local disk before it can start; the larger the model and the more replicas there are, the slower this step becomes and the more disk it uses. With Fluid:

- The weights are preloaded into the cache once, and all replicas read them through the same PVC instead of downloading their own copy.
- vLLM uses the model as a local directory. It does not need to know about the object storage, and no download logic is needed in the image or the startup script.

## Prerequisites

1. Fluid is installed on the Kubernetes cluster with CacheRuntime enabled. See the [installation guide](../../userguide/install.md). CacheRuntime is disabled by default; add `--set runtime.cacheruntime.enabled=true` to the Helm installation. After installation, `kubectl get pod -n fluid-system` should list `cacheruntime-controller`.
2. The cluster has at least one node with an NVIDIA GPU, and the [NVIDIA device plugin](https://github.com/NVIDIA/k8s-device-plugin) is installed so that the node reports the `nvidia.com/gpu` resource.
3. A MinIO or S3-compatible object storage service is available.
4. `kubectl` (1.23 or later, for `kubectl wait --for=jsonpath`) and `curl` (7.75.0 or later, for `--aws-sigv4`) are available locally.

This example was verified with the images `vllm/vllm-openai:v0.33.0` and `curvine/curvine:v0.3.6-alpha`. All resources are created in the `default` namespace, and the commands assume that `default` is the current namespace.

## Overview

1. Deploy MinIO, create a bucket, and create the `CacheRuntimeClass`
2. Upload the model to object storage
3. Create the `Dataset` and the `CacheRuntime`
4. Run a DataLoad to preload the model into the cache
5. Deploy vLLM with the PVC created by Fluid mounted
6. Call the inference API to verify
7. Scale out vLLM replicas that share the same cache

## Step 1: Prepare MinIO and the CacheRuntimeClass

This step is identical to [Use Curvine as CacheRuntime for Data Caching](./curvine_cache_runtime.md). Follow Step 1, Step 2 and Step 4 of that document in order:

- Step 1: Deploy MinIO (this also creates the credential Secret `curvine-secret`). After deploying, run `kubectl wait --for=condition=Ready pod -l app=minio --timeout=120s` and wait for MinIO to be ready before the next step; otherwise the bucket Job may exhaust its retries while MinIO is still starting
- Step 2: Create the bucket `test`
- Step 4: Create the `CacheRuntimeClass` named `curvine-demo`

> Note: MinIO in this example has no persistent volume, and Curvine's metadata and cache use `emptyDir`. This is for demonstration only. If the MinIO Pod is recreated, the model has to be uploaded again; if the Curvine Pods are recreated, the model has to be preloaded again. Use persistent object storage for the model in production.

If you use an existing S3-compatible object store, skip the MinIO deployment. In that case, change the following accordingly: the access credentials in the Secret (keep the name `curvine-secret` and the keys `access-key` and `secret-key`, or update the references in the Dataset), `mountPoint`, `endpoint_url` and `region_name` in the Dataset in Step 3, and the upload address, bucket and credentials in Step 2.

## Step 2: Upload the Model to Object Storage

This example uses [Qwen2.5-0.5B-Instruct](https://modelscope.cn/models/Qwen/Qwen2.5-0.5B-Instruct), whose weight file is about 1GB. Download the model to your local machine first:

```shell
$ pip install modelscope
$ modelscope download --model Qwen/Qwen2.5-0.5B-Instruct --local_dir ./Qwen2.5-0.5B-Instruct
```

Forward the MinIO port to your local machine, then upload each file in the model directory to `Qwen2.5-0.5B-Instruct/` in the bucket `test`. The upload runs in a subshell, so you stay in the current directory afterwards; it aborts as soon as any file fails to upload:

```shell
$ kubectl port-forward svc/minio 19000:9000 &
$ PF_PID=$!   # remember the port-forward process so it can be stopped later
$ timeout 60 sh -c 'until curl -s -o /dev/null http://127.0.0.1:19000/minio/health/live; do sleep 1; done' \
    || echo "MinIO is not reachable through the port-forward after 60s"   # wait for the port-forward

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

$ kill "$PF_PID"   # stop the port-forward
```

You should see one `uploaded ...` line per file. If a file such as `config.json` is missing, vLLM fails to start with `Invalid repository ID or local directory specified`.

If an HTTP proxy is configured on your machine, add `--noproxy '*'` to `curl`. You can also upload with any other S3 client, such as `mc` or `aws s3 cp`.

## Step 3: Create the Dataset and the CacheRuntime

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

Compared with the basic example, only one thing changes here: the `quota` of `worker.tieredStore` is raised from 1Gi to 4Gi. The cache capacity must be larger than the total size of the model, with headroom for the watermarks; raise it accordingly for larger models. For the choice of cache media, see [CacheRuntime TieredStore](../../userguide/cache_runtime_tieredstore.md).

Wait for the Dataset to reach the `Bound` phase and check that the Curvine components are running:

```shell
$ kubectl get dataset curvine-demo
NAME           UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
curvine-demo   0                0.00B    4.00GiB          0                   Bound   38s

$ kubectl get pod -l cacheruntime.fluid.io/name=curvine-demo
NAME                    READY   STATUS    RESTARTS   AGE
curvine-demo-master-0   1/1     Running   0          37s
curvine-demo-worker-0   1/1     Running   0          37s
```

Once the Dataset is bound, Fluid creates a PVC with the same name, `curvine-demo`, for applications to mount:

```shell
$ kubectl get pvc curvine-demo
NAME           STATUS   VOLUME                 CAPACITY   ACCESS MODES   STORAGECLASS   AGE
curvine-demo   Bound    default-curvine-demo   100Pi      RWX            fluid          40s
```

## Step 4: Preload the Model into the Cache

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

`/minio` in `target.path` corresponds to `mounts[].name` in the Dataset. The preload is complete when the DataLoad `PHASE` becomes `Complete`. At that point the `CACHED` column of the Dataset should be close to the total size of the model, which serves as an additional check:

```shell
$ kubectl get dataload curvine-dataload
NAME               DATASET        PHASE      AGE   DURATION
curvine-dataload   curvine-demo   Complete   4s    4s

$ kubectl get dataset curvine-demo
NAME           UFS TOTAL SIZE   CACHED      CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
curvine-demo   0                953.28MiB   4.00GiB          0                   Bound   68s
```

> Note: With Curvine v0.3.6-alpha, the `UFS TOTAL SIZE` and `CACHED PERCENTAGE` columns show 0. These two values come from Curvine's cache summary report.

## Step 5: Deploy vLLM

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

Key fields:

- **Model path**: The PVC is mounted at `/data`, the mount name in the Dataset is `minio`, and the model was uploaded to `Qwen2.5-0.5B-Instruct/` in the bucket, so the model path inside the container is `/data/minio/Qwen2.5-0.5B-Instruct`. The image entrypoint is `vllm serve`, and its first argument is the model path.
- **`volumes.model`**: References the PVC `curvine-demo` created by Fluid, mounted read-only.
- **`enableServiceLinks: false` and the Service name**: If a Service named `vllm` exists in the same namespace and service environment variable injection is enabled for the Pod (the Kubernetes default), the Pod gets environment variables such as `VLLM_PORT=tcp://...`, which collide with vLLM's own `VLLM_PORT` setting, and vLLM fails to start with `VLLM_PORT ... appears to be a URI`. This example disables service environment variable injection and names the Service `vllm-svc`; either one is enough to avoid the problem.
- **`/dev/shm`**: vLLM needs shared memory for inter-process communication, so an in-memory volume is mounted here.
- **`--gpu-memory-utilization`**: The fraction of GPU memory that vLLM reserves. The model in this example is small, so it is set to 0.4; adjust it to your model size and GPU.

## Step 6: Verify

Check that the model directory is a Curvine FUSE mount and that vLLM has finished loading the model (the durations in the log are illustrative and vary by environment):

```shell
$ kubectl exec deploy/vllm -- sh -c 'mount | grep " /data "'
curvinefs on /data type fuse (ro,relatime,user_id=0,group_id=0,default_permissions,allow_other)

$ kubectl logs deploy/vllm | grep -E "Loading weights took|Model loading took|Starting vLLM server"
(EngineCore pid=204) INFO [default_loader.py:506] Loading weights took 0.15 seconds
(EngineCore pid=204) INFO [model_runner.py:428] Model loading took 0.93 GiB memory and 1.909626 seconds
(APIServer pid=1) INFO [entry.py:139] Starting vLLM server on http://0.0.0.0:8000
```

Call the OpenAI-compatible API:

```shell
$ kubectl run curl-test --rm -i --restart=Never --image=curlimages/curl:8.22.0 -q -- \
    -s http://vllm-svc:8000/v1/chat/completions \
    -H 'Content-Type: application/json' \
    -d '{"model":"qwen","messages":[{"role":"user","content":"Introduce Kubernetes in one sentence."}],"max_tokens":80}'
```

The response looks like this:

```json
{
  "object": "chat.completion",
  "model": "qwen",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Kubernetes is an open-source platform for managing and deploying containerized applications across multiple nodes, enabling efficient resource utilization and high availability."
      }
    }
  ],
  "usage": {"prompt_tokens": 36, "total_tokens": 63, "completion_tokens": 27}
}
```

## Step 7: Scale Out vLLM Replicas

```shell
$ kubectl scale deploy/vllm --replicas=2
$ kubectl rollout status deployment/vllm --timeout=600s
$ kubectl get pod -l app=vllm
NAME                   READY   STATUS    RESTARTS   AGE
vllm-d969fc667-bwhsc   1/1     Running   0          59s
vllm-d969fc667-z4plq   1/1     Running   0          3h32m
```

The new replica mounts the same PVC and shares the single copy of the model in Curvine with the existing replica, so it does not take up another copy on local disk the way a download-per-replica setup would. You can check the `Loading weights took` line in the new replica's log to confirm that it has finished loading.

> Note: Each replica requests one `nvidia.com/gpu`. If the cluster has only one GPU, you can enable [time-slicing](https://github.com/NVIDIA/k8s-device-plugin#shared-access-to-gpus) in the NVIDIA device plugin so that several replicas share the card for functional verification. With time-slicing, the GPU memory used by all replicas together must fit in the physical memory of the card; if `renameByDefault: true` is set in the time-slicing configuration, the resource is advertised as `nvidia.com/gpu.shared` and the resource request in the Deployment must be changed accordingly. Time-slicing neither isolates GPU memory nor adds compute capacity, so it is not suitable for production.

## Cleanup

```shell
$ kubectl delete -f vllm.yaml
$ kubectl delete -f dataload.yaml
$ kubectl delete -f dataset.yaml
$ kubectl delete -f cacheruntime.yaml
$ kubectl delete cacheruntimeclass curvine-demo
$ kubectl delete -f minio_create_bucket.yaml
$ kubectl delete -f minio.yaml
```
