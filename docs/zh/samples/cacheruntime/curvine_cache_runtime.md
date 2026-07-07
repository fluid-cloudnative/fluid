# 示例 - 使用 Curvine 作为 CacheRuntime 进行数据缓存

本示例演示如何在 Fluid 中使用 [Curvine](https://github.com/curvine-io/curvine) 作为 CacheRuntime，加速从对象存储（如 S3/MinIO）访问数据。Curvine 是一个高性能的云原生分布式文件系统，可与 Fluid 的 CacheRuntime 抽象无缝集成。

## 前提条件

在运行此示例之前，请确保：

1. 已在 Kubernetes 集群上安装支持 CacheRuntime 的 Fluid。请参考[安装文档](../../userguide/install.md)。
2. 集群中已安装 AdvancedStatefulSet 控制器（Curvine 使用 AdvancedStatefulSet 部署 worker 节点）。
3. 有一个可用的 MinIO 或 S3 兼容的对象存储服务。

## 概览

完整工作流程包含以下步骤：

1. 部署 MinIO 作为底层对象存储
2. 创建 MinIO 凭证 Secret
3. 定义指向 S3 存储桶的 `Dataset`
4. 定义包含 Curvine 拓扑和组件的 `CacheRuntimeClass`
5. 创建 `CacheRuntime` 实例化 Curvine 集群
6. 写入数据（通过向挂载的 PVC 写入数据的 Job）
7. 运行 DataLoad 预热数据到 Curvine 缓存中
8. 通过缓存层读取数据（获得更快的访问速度）
9. （可选）使用引用 Dataset 实现跨命名空间数据共享

## 步骤 1：部署 MinIO

创建一个 MinIO Deployment 作为 S3 兼容后端：

```shell
$ cat<<EOF >minio.yaml
apiVersion: v1
kind: Secret
metadata:
  name: curvine-secret
stringData:
  access-key: minioadmin
  secret-key: minioadmin
---
apiVersion: v1
kind: Service
metadata:
  name: minio
spec:
  type: ClusterIP
  ports:
    - port: 9000
      targetPort: 9000
      protocol: TCP
  selector:
    app: minio
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: minio
spec:
  selector:
    matchLabels:
      app: minio
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: minio
    spec:
      containers:
      - name: minio
        image: minio/minio
        imagePullPolicy: IfNotPresent
        resources:
          limits:
            memory: "512Mi"
        args:
        - server
        - /data
        env:
        - name: MINIO_ROOT_USER
          value: "minioadmin"
        - name: MINIO_ROOT_PASSWORD
          value: "minioadmin"
        ports:
        - containerPort: 9000
          hostPort: 9000
EOF
$ kubectl create -f minio.yaml
```

## 步骤 2：创建 MinIO 存储桶

Curvine 需要在挂载前目标存储桶已存在。运行一次性 Job 来创建它：

```shell
$ cat<<EOF >minio_create_bucket.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: minio-bucket-create
spec:
  template:
    spec:
      containers:
      - name: mc
        image: minio/mc
        imagePullPolicy: IfNotPresent
        resources:
          limits:
            memory: "512Mi"
        command:
          - /bin/sh
          - -c
          - "mc alias set myminio http://minio:9000 $MINIO_ROOT_USER $MINIO_ROOT_PASSWORD && mc mb myminio/test"
        env:
        - name: MINIO_ROOT_USER
          value: "minioadmin"
        - name: MINIO_ROOT_PASSWORD
          value: "minioadmin"
      restartPolicy: OnFailure
  backoffLimit: 4
EOF
$ kubectl create -f minio_create_bucket.yaml
$ kubectl wait --for=condition=complete job/minio-bucket-create --timeout=120s
```

## 步骤 3：定义 Dataset

创建一个指向 MinIO 存储桶的 Dataset：

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

关键字段说明：
- `mountPoint`: S3 存储桶路径（`s3://test` 映射到 MinIO 上的 `test` 存储桶）
- `options`: S3 兼容后端的连接参数
- `encryptOptions`: 引用包含认证凭据的 Secret

## 步骤 4：定义 CacheRuntimeClass

`CacheRuntimeClass` 定义了 Curvine 集群的拓扑结构，包括 master、worker 和 client 组件：

```shell
$ cat<<EOF >cacheruntimeclass.yaml
apiVersion: data.fluid.io/v1alpha1
kind: CacheRuntimeClass
metadata:
  name: curvine-demo
fileSystemType: curvinefs
dataOperationSpecs:
  - name: DataLoad
    command:
      - "/bin/bash"
      - "-c"
    args:
      - |
        CURVINE_HOME="/app/curvine"
        CURVINE_CONF_DIR="${CURVINE_HOME}/conf"
        CURVINE_DATA_DIR="${CURVINE_HOME}/data"
        CURVINE_LOG_DIR="${CURVINE_HOME}/logs"
        
        mkdir -p "$CURVINE_CONF_DIR"
        mkdir -p "$CURVINE_DATA_DIR"
        mkdir -p "$CURVINE_LOG_DIR"
        
        python3 "$CURVINE_HOME/generate_config.py" || {
          echo "Error: generate_config.py failed."
          exit 1
        }
        
        export CURVINE_CONF_FILE="${CURVINE_CONF_DIR}/curvine-cluster.toml"
        
        IFS=: read -ra paths <<< "$FLUID_DATALOAD_DATA_PATH"
        for p in "${paths[@]}"; do
          "$CURVINE_HOME/bin/cv" load "$p" --watch --conf "$CURVINE_CONF_FILE" || {
            echo "Error: load $p failed."
            exit 1
          }
        done
topology:
  master:
    service:
      headless: {}
    executionEntries:
      mountUFS:
        command:
          - bash
          - -c
          - /app/curvine/mountUfs.sh
        timeout: 120
    template:
      spec:
        restartPolicy: Always
        containers:
          - name: master
            image: curvine/curvine-fluid:latest
            command:
              - /entrypoint.sh
            args:
              - master
              - start
            imagePullPolicy: IfNotPresent
            readinessProbe:
              tcpSocket:
                port: 8995
              initialDelaySeconds: 35
              periodSeconds: 5
              failureThreshold: 12
            env:
              - name: POD_NAME
                valueFrom:
                  fieldRef:
                    fieldPath: metadata.name
  worker:
    service:
      headless: {}
    template:
      spec:
        restartPolicy: Always
        containers:
          - name: worker
            image: curvine/curvine-fluid:latest
            command:
              - /entrypoint.sh
            args:
              - worker
              - start
            imagePullPolicy: IfNotPresent
            readinessProbe:
              tcpSocket:
                port: 8997
              initialDelaySeconds: 5
              periodSeconds: 5
              failureThreshold: 12
            env:
              - name: POD_NAME
                valueFrom:
                  fieldRef:
                    fieldPath: metadata.name
  client:
    template:
      spec:
        restartPolicy: Always
        containers:
          - name: client
            image: curvine/curvine-fluid:latest
            command:
              - /entrypoint.sh
            args:
              - client
              - start
            imagePullPolicy: IfNotPresent
            securityContext:
              privileged: true
              runAsUser: 0
            lifecycle:
              preStop:
                exec:
                  command:
                    - /bin/sh
                    - -c
                    - |
                      echo "PreStop: Cleaning up FUSE mount..."
                      target_path="${CURVINE_TARGET_PATH:-${FLUID_RUNTIME_MOUNT_PATH:-}}"
                      if [ -n "$target_path" ] && mountpoint -q "$target_path" 2>/dev/null; then
                        fusermount -u "$target_path" 2>/dev/null || umount -f "$target_path" 2>/dev/null || true
                      else
                        echo "Mount point $target_path is not mounted, skipping"
                      fi
                      echo "PreStop: Cleanup completed"
EOF
$ kubectl create -f cacheruntimeclass.yaml
```

关键部分说明：
- **`fileSystemType`**: 标识此为 Curvine 文件系统（`curvinefs`）
- **`dataOperationSpecs`**: 定义 DataLoad 操作的执行方式——生成配置文件，然后使用 `cv` CLI 从 UFS 预热数据到 Curvine
- **`topology.master`**: Curvine master 节点，包含 Headless Service、8995 端口的就绪探针和 MountUFS 脚本
- **`topology.worker`**: Curvine worker 节点，包含 8997 端口的就绪探针
- **`topology.client`**: FUSE 客户端 DaemonSet，以特权模式运行，包含优雅的 pre-stop 卸载清理逻辑

## 步骤 5：创建 CacheRuntime

通过创建引用 `CacheRuntimeClass` 的 `CacheRuntime` 来实例化 Curvine 集群：

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
            quota: 1Gi
  client:
    options:
      debug: "false"
EOF
$ kubectl create -f cacheruntime.yaml
```

关键字段说明：
- **`runtimeClassName`**: 关联到步骤 4 中定义的 `CacheRuntimeClass`
- **`master.options`**: Curvine master 配置（端口、日志路径等）
- **`worker.tieredStore`**: 缓存层级配置——1Gi emptyDir，水位线 50%-80%
- **`worker.volumeMounts`**: 日志和数据目录的挂载点

## 步骤 6：验证资源就绪

检查 Dataset 是否已绑定以及运行时组件是否健康：

```shell
# 等待 Dataset 进入 Bound 阶段
$ kubectl get dataset curvine-demo
NAME           UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
curvine-demo                  0B       1Gi              0%               Bound    2m

# 检查 CacheRuntime 状态
$ kubectl get cacheruntime curvine-demo
NAME            PHASE   AGE
curvine-demo    Ready   2m

# 验证所有 Pod 都在运行
$ kubectl get pod -l cacheruntime.fluid.io/runtime-name=curvine-demo
NAME                            READY   STATUS    RESTARTS   AGE
curvine-demo-master-0           1/1     Running   0          1m
curvine-demo-worker-0           1/1     Running   0          1m
```

## 步骤 7：向缓存写入数据

创建一个 Job，将数据写入挂载的 PVC。这会填充底层的 MinIO 存储桶：

```shell
$ cat<<EOF >write_job.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: write-job
  namespace: default
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: write-job
        image: busybox
        imagePullPolicy: IfNotPresent
        resources:
          limits:
            ephemeral-storage: "5Gi"
            memory: "512Mi"
        command: ['sh', '-c', 'echo helloworld > /data/minio/bar']
        volumeMounts:
        - name: data-vol
          mountPath: /data
      volumes:
      - name: data-vol
        persistentVolumeClaim:
          claimName: curvine-demo
EOF
$ kubectl create -f write_job.yaml
$ kubectl wait --for=condition=complete job/write-job --timeout=120s
```

此 Job 在 Curvine FUSE 挂载的 `/data/minio/bar` 处写入包含 "helloworld" 的文件 `bar`。数据会持久化到 MinIO S3 存储桶中。

## 步骤 8：运行 DataLoad 预热数据

触发 DataLoad 操作，将数据从 MinIO 预热到 Curvine 缓存中：

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
$ kubectl wait dataload/curvine-dataload --for=condition=Complete --timeout=300s
```

DataLoad 目标为数据集内的 `/minio` 路径。`CacheRuntimeClass` 中定义的 `dataOperationSpecs` 将执行 Curvine `cv load` 命令，从 S3 拉取数据到缓存层。

验证缓存填充情况：

```shell
$ kubectl get dataset curvine-demo
NAME             UFS TOTAL SIZE   CACHED    CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
curvine-demo     12B              12B       1Gi              100%                Bound   10m
```

## 步骤 9：从缓存读取数据

运行一个从缓存读取数据的 Job。首次读取（DataLoad 之前）数据来自远程 S3 存储；DataLoad 之后，数据从本地缓存提供服务：

```shell
$ cat<<EOF >read_job.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: read-job
  namespace: default
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: read-job
        image: busybox
        imagePullPolicy: IfNotPresent
        resources:
          limits:
            memory: "512Mi"
            ephemeral-storage: "5Gi"
        command: ['sh']
        args:
        - -c
        - set -ex; test -n "$(cat /data/minio/bar)"
        volumeMounts:
        - name: data-vol
          mountPath: /data
      volumes:
      - name: data-vol
        persistentVolumeClaim:
          claimName: curvine-demo
EOF
$ kubectl create -f read_job.yaml
$ kubectl wait --for=condition=complete job/read-job --timeout=120s
```

## 步骤 10（可选）：引用 Dataset 实现跨命名空间共享

创建一个引用 Dataset，指向 Curvine 缓存，使其他命名空间或工作负载能够共享相同的缓存数据：

```shell
$ cat<<EOF >ref-dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: curvine-demo-ref
spec:
  mounts:
    - mountPoint: "dataset://default/curvine-demo"
      name: curvine-cache
EOF
$ kubectl create -f ref-dataset.yaml
```

然后一个 Job 可以挂载引用 Dataset：

```shell
$ cat<<EOF >read_ref_job.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: read-ref-job
  namespace: default
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: read-ref-job
        image: busybox
        imagePullPolicy: IfNotPresent
        resources:
          limits:
            memory: "512Mi"
            ephemeral-storage: "5Gi"
        command: ['sh']
        args:
        - -c
        - set -ex; test -n "$(cat /data/minio/bar)"
        volumeMounts:
        - name: data-vol
          mountPath: /data
      volumes:
      - name: data-vol
        persistentVolumeClaim:
          claimName: curvine-demo-ref
EOF
$ kubectl create -f read_ref_job.yaml
$ kubectl wait --for=condition=complete job/read-ref-job --timeout=120s
```

引用 Dataset 会自动创建一个 ThinRuntime，连接到原始 Curvine 缓存，实现工作负载间零拷贝数据共享。

## 扩缩 Worker

可以通过 Patch CacheRuntime 来扩缩 Curvine worker 池：

```shell
# 扩容到 2 个 worker
$ kubectl patch cacheruntime curvine-demo --type merge -p '{"spec":{"worker":{"replicas":2}}}'

# 验证
$ kubectl get cacheruntime curvine-demo -o jsonpath='{.status.worker.readyReplicas}/{.status.worker.desiredReplicas}'
2/2
```

## 环境清理

移除本示例中创建的所有资源：

```shell
$ kubectl delete -f write_job.yaml
$ kubectl delete -f read_job.yaml
$ kubectl delete -f read_ref_job.yaml
$ kubectl delete -f dataload.yaml
$ kubectl delete -f ref-dataset.yaml
$ kubectl delete -f dataset.yaml
$ kubectl delete -f cacheruntime.yaml
$ kubectl delete -f cacheruntimeclass.yaml
$ kubectl delete -f minio.yaml
$ kubectl delete -f minio_create_bucket.yaml
```
