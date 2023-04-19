# 示例 - JuiceFSRuntime 数据迁移

JuiceFS 本质是一个 POSIX 兼容的文件系统，当存量数据存在对象存储中，需要用 JuiceFS 访问数据时，需要先将数据导入到 JuiceFS
中。本文讲述如何 Fluid 中针对 JuiceFSRuntime 的 Dataset 做数据迁移。

## Dataset & JuiceFSRuntime

数据迁移是在数据集的基础上的，所以首先需要先创建一个 Dataset 和
JuiceFSRuntime。具体请参考文档[示例 - 如何在 Fluid 中使用 JuiceFS](juicefs_runtime.md)，这里不再赘述。

## 数据迁移

在 Dataset 可用（Bound 状态）之后，接下来可以通过 DataMigrate 来加速数据访问。以下是一个 DataMigrate 的示例：

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: DataMigrate
metadata:
  name: jfs-migrate
spec:
  image: registry.cn-hangzhou.aliyuncs.com/juicefs/juicefs-fuse
  imageTag: nightly
  from:
    externalStorage:
      uri: minio://minio.default.svc.cluster.local:9000/test/
      encryptOptions:
        - name: access-key
          valueFrom:
            secretKeyRef:
              name: jfs-secret
              key: accesskey
        - name: secret-key
          valueFrom:
            secretKeyRef:
              name: jfs-secret
              key: secretkey
  to:
    dataset:
      name: jfsdemo
      namespace: default
      path: /dir1
  options:
    "exclude": "4.png"
```

其中：

- `spec.image`：指定数据迁移的 job 使用的镜像；
- `spec.imageTag`：指定数据迁移的 job 使用的镜像的 tag；
- `spec.from`/`spec.to`：指定数据迁移的源和目标；
- `spec.from(/to).externalStorage`：表明需要迁移的数据是外部的存储，可以是 oss、s3 等；
- `spec.from(/to).externalStorage.uri`：外部存储的 uri；
- `spec.from(/to).externalStorage.encryptOptions`：外部存储的加密参数，access-key、secret-key 是固定格式，表示对象存储的
  aksk；
- `spec.from(/to).dataset`：表明需要迁移数据的 dataset；
- `spec.from(/to).dataset.name`：需要迁移数据的 dataset 的 name；
- `spec.from(/to).dataset.namespace`：需要迁移数据的 dataset 的 namespace；
- `spec.from(/to).dataset.path`：需要迁移数据的 dataset 的子路径；
- `spec.options`：juicefs sync 的参数，具体请参考[文档](https://juicefs.com/docs/zh/community/command_reference#juicefs-sync)。

## DataMigrate 生命周期

DataMigrate 的生命周期如下图所示：

![](images/fluid-datamigration-state.jpg)

### DataMigration 状态转换流程

1. DataMigrate 创建后，状态为 Pending；
2. 只有当 DataSet 状态为 Bound 时，或当前环境满足 policy，DataMigration 才能运行；
3. 运行数据迁移时，状态从 Pending 置为 Excuting；
4. 运行成功状态变为 Complete；运行失败状态变为 Failed；

### DataSet 状态转换流程

1. DataMigrate 执行时，DataSet 状态为 Migrating，表示此时正在迁移数据，不可进行数据更新，否则会有数据不一致的风险。
2. DataMigrate 执行完，DataSet 状态恢复为 Bound。
