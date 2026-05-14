# Alluxio S3 高并发读调优

本文记录一个面向 AlluxioRuntime + S3 兼容后端高并发读场景的调优 profile。

这个配置来自 [issue #5802](https://github.com/fluid-cloudnative/fluid/issues/5802) 的排查：fio 通过 S3 后端的 AlluxioRuntime 读取数据时，在高并发下可能挂住。这个 profile 不修改 Alluxio 内部实现，而是提供一组已经验证过的 AlluxioRuntime 配置，用户可以通过 `spec.properties` 和 FUSE args 显式配置。

## 场景

问题在接近以下环境中复现：

- Kubernetes v1.26.7
- Fluid v1.0.8 和当前 Fluid master
- Alluxio 2.9.5
- SeaweedFS 3.80 作为 S3 兼容后端
- 1 个 Alluxio master，1 个 worker，以及 FUSE
- S3 中 64 个文件，每个约 5GiB

fio 命令如下：

```bash
FILES=$(seq -f "/data/file%g" 0 63 | paste -sd:)
fio -iodepth=1 -rw=read -ioengine=libaio -bs=256k \
  -numjobs=<numjobs> -group_reporting -size=5G \
  --filename="$FILES" -name=read_test --readonly -direct=1 --runtime=60
```

应用调优 profile 前观察到的现象：

- `numjobs=8` 和 `numjobs=16` 能完成。
- 更高并发，例如 `numjobs=32` 或 `numjobs=64`，可能挂住。
- 挂住后测试 Pod 可能无法正常删除。
- force delete 后，节点上的 fio 或 FUSE 状态仍可能残留卡住。

排查中最强的信号指向 Alluxio 2.9.5 在 S3 高并发读下的 FUSE/client 读路径。在复现环境中，JNI-FUSE 可能出现路径锁超时；切换到 JNR/libfuse2 后，也需要同时调大 S3 thread、worker client pool，并关闭 direct memory IO，才能让重复 `numjobs=64` 稳定通过。

## 推荐 Runtime 配置

仅建议在 S3 或 S3 兼容后端的高并发读场景中使用这组配置。其他 workload 建议保持默认行为，除非你已经在自己的环境中验证了同样的调优。

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

注意事项：

- 设置 `alluxio.fuse.jnifuse.enabled=false` 和 `alluxio.fuse.jnifuse.libfuse.version=2`，使用 JNR/libfuse2。
- 使用 libfuse2 时，需要从 FUSE args 中移除 `max_idle_threads=*`。`max_idle_threads` 是 libfuse3 参数。
- 调大 S3 threads 和 worker client pool，以承载高并发读。
- 增大 read chunk 和 buffer，减少请求碎片化。
- 设置 `alluxio.user.direct.memory.io.enabled=false`。在复现环境中，这是让重复 `numjobs=64` 稳定通过的关键。

## Dataset 示例

请使用 Kubernetes Secret 保存访问密钥，不要把 AK/SK 硬编码到 YAML 中。

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

## 测试 Pod 示例

挂载 Dataset，并在 `/data` 下运行 fio。

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

## 验证结果

在复现问题的云上环境中，通过 Fluid 生成的 AlluxioRuntime 配置应用上述 profile 后：

```text
numjobs=8: passed
numjobs=16: passed
numjobs=32: passed
numjobs=64: passed
repeat numjobs=64: passed
test Pod deletion: passed
Alluxio master/worker/fuse restart count: 0
```

应用 profile 后未再观察到以下致命症状：

- `DeadlineExceededRuntimeException`
- `Timer expired`
- `OutOfDirectMemoryError`

Alluxio 日志中仍可能出现 `TempBlockMeta not found` 警告，但在验证环境中，fio 可以成功完成，测试 Pod 可以正常删除，Runtime 组件也保持健康。

## 风险和适用范围

- 这是一个调优/configuration profile，不是 Alluxio 内部实现修复。
- 这些参数已在 issue #5802 的 S3 兼容 workload 中验证。不同 S3 后端、对象大小、网络延迟和并发级别可能仍需要调参。
- 关闭 direct memory IO 可以提升这个 workload 的稳定性，但可能影响性能。
- 如果应用这组配置后仍出现同类问题，请在 force delete Pod 前收集 FUSE 日志、worker 日志、节点进程状态、mount 信息和 kubelet 日志。
