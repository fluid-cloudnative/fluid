# Example - Use Curvine as CacheRuntime for Data Caching

This example demonstrates how to use [Curvine](https://github.com/curvine-io/curvine) as a CacheRuntime in Fluid to accelerate data access from object storage (e.g., S3/MinIO). Curvine is a high-performance, cloud-native distributed file system that integrates seamlessly with Fluid's CacheRuntime abstraction.

## Prerequisites

Before running this example, make sure:

1. Fluid with CacheRuntime support is installed on your Kubernetes cluster. Refer to the [Installation Guide](../../userguide/install.md).
2. Your cluster has the AdvancedStatefulSet controller installed (Curvine uses AdvancedStatefulSet for its worker nodes).
3. You have a MinIO or S3-compatible object storage service available.

## Overview

The complete workflow consists of these steps:

1. Deploy MinIO as the underlying object storage
2. Create a Secret for MinIO credentials
3. Define a `Dataset` pointing to the S3 bucket
4. Define a `CacheRuntimeClass` with Curvine topology and components
5. Create a `CacheRuntime` to instantiate the Curvine cluster
6. Write data into the cache (via a job that writes to the mounted PVC)
7. Run a DataLoad to preload data into Curvine
8. Read data through the cached layer (faster access)
9. (Optional) Use a reference Dataset for cross-namespace data sharing

For detailed documentation on how Curvine integrates with Fluid, including how Curvine parameters are parsed and used, refer to the [Curvine Fluid Integration Guide](https://curvineio.github.io/zh-cn/docs/Architecture/fluid-integration/).

## Step 1: Deploy MinIO

Create a MinIO deployment to serve as the S3-compatible backend:

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

## Step 2: Create a MinIO Bucket

Curvine requires the target bucket to exist before mounting. Run a one-time job to create it:

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

## Step 3: Define the Dataset

Create a Dataset that points to the MinIO bucket:

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

Key fields:
- `mountPoint`: The S3 bucket path (`s3://test` maps to the `test` bucket on MinIO)
- `options`: Connection parameters for the S3-compatible backend
- `encryptOptions`: References to secrets containing authentication credentials

## Step 4: Define the CacheRuntimeClass

The `CacheRuntimeClass` defines the Curvine cluster topology including master, worker, and client components:

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

Key sections:
- **`fileSystemType`**: Identifies this as a Curvine filesystem (`curvinefs`)
- **`dataOperationSpecs`**: Defines how DataLoad operations execute — generates config, then uses the `cv` CLI to preload data from the UFS into Curvine
- **`topology.master`**: Curvine master node with headless service, readiness probe on RPC port 8995, and a MountUFS script
- **`topology.worker`**: Curvine worker nodes with readiness probe on port 8997
- **`topology.client`**: FUSE client DaemonSet running in privileged mode, with graceful pre-stop unmount cleanup

## Step 5: Create the CacheRuntime

Instantiate the Curvine cluster by creating a `CacheRuntime` that references the `CacheRuntimeClass`:

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

Key fields:
- **`runtimeClassName`**: Links to the `CacheRuntimeClass` defined in Step 4
- **`master.options`**: Curvine master configuration (ports, journal paths, etc.)
- **`worker.tieredStore`**: Cache tier configuration — 1Gi emptyDir with 50%-80% low/high watermarks
- **`worker.volumeMounts`**: Mounts for logs and data directories

## Step 6: Verify Resources Are Ready

Check that the Dataset is bound and the runtime components are healthy:

```shell
# Wait for Dataset to reach Bound phase
$ kubectl get dataset curvine-demo
NAME           UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
curvine-demo                  0B       1Gi              0%               Bound    2m

# Check CacheRuntime status
$ kubectl get cacheruntime curvine-demo
NAME            PHASE   AGE
curvine-demo    Ready   2m

# Verify all pods are running
$ kubectl get pod -l cacheruntime.fluid.io/runtime-name=curvine-demo
NAME                            READY   STATUS    RESTARTS   AGE
curvine-demo-master-0           1/1     Running   0          1m
curvine-demo-worker-0           1/1     Running   0          1m
```

## Step 7: Write Data to the Cache

Create a job that writes data to the mounted PVC. This populates the underlying S3 bucket:

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

This job writes a file `bar` containing "helloworld" to `/data/minio/bar` inside the Curvine FUSE mount. The data is persisted to the MinIO S3 bucket.

## Step 8: Run DataLoad to Preload Data

Trigger a DataLoad operation to preload data from MinIO into the Curvine cache:

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

The DataLoad targets the `/minio` path within the dataset. The `dataOperationSpecs` defined in the `CacheRuntimeClass` will execute the Curvine `cv load` command to pull data from S3 into the cache tier.

Verify the cache population:

```shell
$ kubectl get dataset curvine-demo
NAME             UFS TOTAL SIZE   CACHED    CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
curvine-demo     12B              12B       1Gi              100%                Bound   10m
```

## Step 9: Read Data from Cache

Run a job that reads the cached data. On first read (before DataLoad), data comes from the remote S3 store. After DataLoad, data is served from the local cache:

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

## Step 10 (Optional): Reference Dataset for Cross-Namespace Sharing

Create a reference Dataset that points to the Curvine cache, allowing other namespaces or workloads to share the same cached data:

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

Then a job can mount the reference dataset:

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

The reference Dataset automatically creates a ThinRuntime that connects to the original Curvine cache, enabling zero-copy data sharing across workloads.

## Scaling Workers

You can scale the Curvine worker pool by patching the CacheRuntime:

```shell
# Scale up to 2 workers
$ kubectl patch cacheruntime curvine-demo --type merge -p '{"spec":{"worker":{"replicas":2}}}'

# Verify
$ kubectl get cacheruntime curvine-demo -o jsonpath='{.status.worker.readyReplicas}/{.status.worker.desiredReplicas}'
2/2
```

## Cleanup

Remove all resources created in this example:

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
