# JuiceFSRuntime Dataset 的配置

如何在 Fluid 中使用 JuiceFS，请参考文档[示例 - 如何在 Fluid 中使用 JuiceFS](juicefs_runtime.md)。本文讲述所有在 Fluid 中有关
JuiceFS Dataset 的相关配置。

## Dataset AccessMode

Dataset 的默认访问模式为 `ReadOnlyMany`，参考文档[《示例 - 修改Dataset的访问模式》](../data_accessmodes.md) 可进行修改。

## Dataset 子目录配额设置

在 Dataset 中可以设置访问 JuiceFS 文件系统的子目录，同时可以为子目录设置配额。以下是一个 Dataset 的示例：

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: jfsdemo
spec:
  mounts:
    - name: minio
      mountPoint: "juicefs:///demo"
      options:
        quota: "2Gi"
```

其中：
- `spec.mounts.mountPoint`：指定 Dataset 的挂载点，目前只支持一个挂载点，以 `juicefs://` 开头；如 `juicefs:///demo` 为 JuiceFS 文件系统的 `/demo` 子目录。
- `spec.mounts.options.quota`：指定 Dataset 子目录的配额，至少为 1Gi。

注意：
JuiceFS 社区版 v1.1.0 以上版本才支持子目录配额设置；商业版 4.9.2 以上支持子目录配额设置。设置子目录配额时，需要确保 JuiceFS 的版本满足要求， 
在 JuiceFSRuntime 中指定 fuse 的镜像为 `juicedata/juicefs-fuse:v1.0.4-4.9.2`，具体镜像参考 [JuiceFS 镜像仓库](https://hub.docker.com/r/juicedata/juicefs-fuse/tags?page=1&name=v)。
