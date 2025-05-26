# 示例 - JuiceFSRuntime 加速数据访问

如何在 Fluid 中使用 JuiceFS，请参考文档[示例 - 如何在 Fluid 中使用 JuiceFS](juicefs_runtime.md)。本文讲述如何 Fluid 中使用 JuiceFSRuntime 加速数据访问。

## Dataset & JuiceFSRuntime

加速数据访问是在数据集的基础上的，所以首先需要先创建一个 Dataset 和 JuiceFSRuntime。具体请参考文档[示例 - 如何在 Fluid 中使用 JuiceFS](juicefs_runtime.md)，这里不再赘述。

## 加速数据访问

在 Dataset 可用（Bound 状态）之后，接下来可以通过 DataLoad 来加速数据访问。以下是一个 DataLoad 的示例：

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: DataLoad
metadata:
  name: jfs-load
spec:
  dataset:
    name: jfsdemo
    namespace: default
  target:
    - path: /dir1
  options:
    threads: "50"
```

其中：
- `spec.dataset.name`：指定要加速数据访问的 Dataset 名称；
- `spec.dataset.namespce`：指定要加速数据访问的 Dataset 的 namespace；
- `spec.target.path`：指定要加速数据访问的数据路径，可以是目录或文件；target 是列表，可以填多个 path。
- `spec.options`：指定缓存加速的参数，可用参数可参考 [JuiceFS 的缓存加速文档](https://juicefs.com/docs/zh/community/command_reference#warmup)，注意 key 和 value 均为字符串类型。
