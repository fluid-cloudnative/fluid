# Alluxio S3 High-Concurrency Read Tuning

This document provides a tuning profile for high-concurrency read workloads that use AlluxioRuntime with an S3-compatible backend.

This profile was validated while investigating [issue #5802](https://github.com/fluid-cloudnative/fluid/issues/5802), where fio reads over an S3-backed AlluxioRuntime could hang at high concurrency. It does not change Alluxio internals. Users can apply the configuration through `spec.properties` and FUSE args.

## Scenario

The issue was reproduced with an environment close to:

- Kubernetes v1.26.7
- Fluid v1.0.8 and Fluid master at the time of investigation
- Alluxio 2.9.5
- SeaweedFS 3.80 as an S3-compatible backend
- One Alluxio master, one worker, and FUSE
- 64 files in S3, each about 5GiB

The fio command was:

```bash
FILES=$(seq -f "/data/file%g" 0 63 | paste -sd:)
fio -iodepth=1 -rw=read -ioengine=libaio -bs=256k \
  -numjobs=<numjobs> -group_reporting -size=5G \
  --filename="$FILES" -name=read_test --readonly -direct=1 --runtime=60
```

Observed behavior without this tuning profile:

- `numjobs=8` and `numjobs=16` completed.
- Higher concurrency, such as `numjobs=32` or `numjobs=64`, could hang.
- The test Pod could fail to delete normally after the hang.
- Force deletion could leave fio or FUSE state stuck on the node.

The validation suggests this tuning mainly mitigates Alluxio 2.9.5 FUSE/client read-path pressure under high-concurrency S3 reads. In the reproduced environment, JNI-FUSE could hit path-lock timeout symptoms. When using JNR/libfuse2, S3 thread/client-pool tuning and disabling direct memory IO were also required to make repeated `numjobs=64` stable.

## Recommended Runtime Configuration

Use this profile only for S3 or S3-compatible high-concurrency read workloads. Keep the default behavior for other workloads unless you have validated the same tuning in your own environment.

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: my-s3
spec:
  replicas: 1
  master:
    resources:
      requests:
        cpu: 8
        memory: 32Gi
      limits:
        cpu: 8
        memory: 32Gi
  worker:
    resources:
      requests:
        cpu: 8
        memory: 32Gi
      limits:
        cpu: 8
        memory: 64Gi
  fuse:
    jvmOptions:
      - "-Xmx16G"
      - "-Xms16G"
      - "-XX:+UseG1GC"
      - "-XX:MaxDirectMemorySize=32g"
      - "-XX:+UnlockExperimentalVMOptions"
      - "-XX:ActiveProcessorCount=16"
    resources:
      requests:
        cpu: 16
        memory: 32Gi
      limits:
        cpu: 16
        memory: 64Gi
    args:
      - fuse
      - --fuse-opts=kernel_cache,rw,allow_other,entry_timeout=60,attr_timeout=60,max_background=256,congestion_threshold=256
  properties:
    alluxio.fuse.jnifuse.enabled: "false"
    alluxio.fuse.jnifuse.libfuse.version: "2"
    alluxio.underfs.s3.threads.max: "2048"
    alluxio.user.block.worker.client.pool.max: "8192"
    alluxio.user.block.size.bytes.default: "64MB"
    alluxio.user.streaming.reader.chunk.size.bytes: "64MB"
    alluxio.user.local.reader.chunk.size.bytes: "64MB"
    alluxio.worker.network.reader.buffer.size: "64MB"
    alluxio.user.direct.memory.io.enabled: "false"
  tieredstore:
    levels:
      - mediumtype: SSD
        path: /home/work/fluid_test
        quota: 100G
        high: "0.95"
        low: "0.6"
```

Important details:

- Set `alluxio.fuse.jnifuse.enabled=false` and `alluxio.fuse.jnifuse.libfuse.version=2` to use JNR/libfuse2.
- Remove `max_idle_threads=*` from FUSE args when using libfuse2. `max_idle_threads` is a libfuse3 option.
- Increase S3 threads and worker client pool size for high-concurrency reads.
- Use larger read chunks and buffers to reduce request fragmentation.
- Set `alluxio.user.direct.memory.io.enabled=false`. In the reproduced environment, this was required for repeated `numjobs=64` stability.

## Dataset Example

Store access keys in a Kubernetes Secret instead of hardcoding them in YAML.

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: my-s3
spec:
  mounts:
    - mountPoint: s3://<bucket-name>/<path-to-data>/
      name: s3
      options:
        alluxio.underfs.s3.endpoint: <s3-endpoint>
        alluxio.underfs.s3.endpoint.region: <s3-endpoint-region>
      encryptOptions:
        - name: aws.accessKeyId
          valueFrom:
            secretKeyRef:
              name: mysecret
              key: aws.accessKeyId
        - name: aws.secretKey
          valueFrom:
            secretKeyRef:
              name: mysecret
              key: aws.secretKey
```

## Test Pod Example

Mount the dataset and run fio from `/data`.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: fio-reader
spec:
  restartPolicy: Never
  containers:
    - name: client
      image: alluxio/alluxio:2.9.5
      securityContext:
        runAsUser: 0
      command: ["/bin/bash", "-lc", "sleep infinity"]
      volumeMounts:
        - mountPath: /data
          name: data
          readOnly: true
          subPath: s3
  volumes:
    - name: data
      persistentVolumeClaim:
        claimName: my-s3
        readOnly: true
```

## Validation Result

In the validation environment, after applying the above profile through Fluid-generated AlluxioRuntime configuration:

```text
numjobs=8: passed
numjobs=16: passed
numjobs=32: passed
numjobs=64: passed
repeat numjobs=64: passed
test Pod deletion: passed
Alluxio master/worker/fuse restart count: 0
```

The following error symptoms were not observed after applying the profile:

- `DeadlineExceededRuntimeException`
- `Timer expired`
- `OutOfDirectMemoryError`

`TempBlockMeta not found` warnings could still appear in Alluxio logs, but fio completed successfully, test Pods deleted normally, and Runtime components stayed healthy in the validation environment.

## Risks and Scope

- This is a tuning/configuration profile, not an upstream Alluxio internal fix.
- The values were validated for the reproduced S3-compatible workload in issue #5802. Different S3 backends, object sizes, network latency, and concurrency levels may still require tuning.
- Disabling direct memory IO improves stability for this workload, but it may affect performance.
- If the same symptoms continue after applying this profile, collect FUSE logs, worker logs, node process states, mount information, and kubelet logs before force-deleting Pods.
