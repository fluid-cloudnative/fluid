# JuiceFSRuntime 缓存配置

如何在 Fluid 中使用 JuiceFS，请参考文档[示例 - 如何在 Fluid 中使用 JuiceFS](juicefs_runtime.md)。本文讲述所有在 Fluid 中有关 JuiceFS 的缓存相关配置。

## 设置多个路径缓存

缓存路径在 JuiceFSRuntime 中的 tieredstore 设置，worker 和 fuse pod 共享相同的配置。

注意：JuiceFS 支持多路径缓存，不支持多级缓存。

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: JuiceFSRuntime
metadata:
  name: jfsdemo
spec:
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: SSD
        path: /mnt/cache1:/mnt/cache2
        quota: 40Gi
        low: "0.1"
```

其中：
- `spec.tiredstore.levels.path` 可设置为多个路径，以 `:` 分隔，缓存会被分配在这里设置的所有路径下；但不支持通配符；
- `spec.tiredstore.levels.quota` 为缓存对象的总大小，与路径多少无关；
- `spec.tiredstore.levels.low` 为缓存路径的最小剩余空间比例，无论缓存是否达到限额，都会保证缓存路径的剩余空间；
- `spec.tiredstore.levels.mediumtype` 为缓存路径的类型，目前支持 `SSD` 和 `MEM`。


## fuse 的缓存

默认情况下，worker 和 fuse 的缓存路径都在 `spec.tiredstore.levels` 中设置，但是也可以单独设置 fuse 的缓存。

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: JuiceFSRuntime
metadata:
  name: jfsdemo
spec:
  fuse:
    options:
      "cache-size": "0"
      "cache-dir": "/fuse/cache"
```

其中：
- `spec.fuse.options` 为 worker 的挂载参数，缓存路径以 `cache-dir` 为 key，以 `:` 分隔的多个路径；缓存大小以 `cache-size` 为 key，单位为 MiB。

若希望设置 fuse pod 不使用缓存，可以设置 `spec.fuse.options.cache-size` 为 `0`。
