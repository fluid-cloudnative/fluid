# 示例 - Alluxio分层存储配置
Fluid所使用的底层存储引擎[Alluxio](https://github.com/Alluxio/alluxio)支持分层的多目录存储. 通过合理的分层存储配置, 能够为用户提升总体I/O吞吐量, 减小数据密集型引用出现数据访问瓶颈的可能性.

本文档将介绍在Fluid中如何对AlluxioRuntime进行声明式的配置,以开启Alluxio分层存储的相关支持.

更多与Alluxio分层存储的配置与技术细节,请参考[Alluxio存储官方文档](https://docs.alluxio.io/os/user/stable/cn/core-services/Caching.html)

## 前提条件

- [Fluid](https://github.com/fluid-cloudnative/fluid)

请参考[Fluid安装文档](../guide/install.md)完成安装


## 单层存储配置

以下是一个典型的AlluxioRuntime资源对象配置:

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: dataset
spec:
  ...
  tieredstore:
    levels:
      - path: /dev/shm
        mediumtype: MEM
        quota: 2Gi
```

在上述示例中,`spec.tieredstore.levels`目录下仅包含一层存储配置,这意味着Alluxio将以单层存储的配置方式运行. 

分层配置中各项属性解释如下:
- `path`: 数据缓存在Alluxio Worker所在结点的实际存储位置
- `mediumtype`: 取值只能为"MEM","SSD","HDD"三者之一,用于指明`path`目录所使用的存储设备类型.
- `quota`: 该层存储所允许的最大缓存容量(Cache Capacity)

## 单层多目录存储配置

介绍单层多目录存储配置的最好方式是以一个例子进行说明:

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: dataset
spec:
  ...
  tieredstore:
    levels:
      - path: /mnt/ssd0/cache,/mnt/ssd1/cache
        mediumtype: SSD
        quota: 4Gi
```

为了启用Alluxio的单层多目录存储配置,唯一要做的修改是在`path`属性中增加若干新的目录,这些目录之间需以逗号分隔.以上述AlluxioRuntime资源对象的yaml配置为例,`path`属性中包含了"/mnt/ssd0/cache"和"mnt/ssd1/cache",这意味着Alluxio将同时使用这两个目录作为数据缓存的实际存储位置.

上述yaml配置同样说明了单层多目录存储配置的一个实用场景: 当应用的数据访问瓶颈来自于存储设备本身时(例如:磁盘带宽上限),上述配置使得Alluxio可以将数据缓存放置在同层的多个不同的目录中,这些目录分别位于不同存储设备的挂载点下(例如上例中的`/mnt/ssd0`和`/mnt/ssd1`),于是大量数据访问请求带来的压力也将被均摊到多个存储设备之上,减轻数据访问造成的瓶颈.

> 注意: 目前Fluid仅支持同构存储设备类型的同层多目录配置.换言之,目前您将无法不同类型的存储介质混合使用

另外,值得注意的是, 如果在同层使用多目录配置,那么`quota`所指明的最大缓存容量上限同样会被多个目录均摊.以上述yaml配置为例进行说明:`quota`被设置为了"4Gi",那么`path`中设置的两个目录均会获得最大"2Gi"(4Gi / 2)的缓存容量.

如果上述"缓存容量均摊"的配置策略无法满足需求, Fluid同样提供了更加细粒度的缓存容量配置方式:

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: dataset
spec:
  properties:
    ...
    # [default property in fluid]
    # alluxio.worker.allocator.class: alluxio.worker.block.allocator.MaxFreeAllocator
  ...
  tieredstore:
    levels:
      - path: /mnt/ssd0/cache,/mnt/ssd1/cache
        mediumtype: SSD
        #quota: 4Gi
        quotaList: 3Gi,2Gi
```

通过设置`quotaList`属性可以实现多目录缓存容量的细粒度配置.`quotaList`必须与`path`中的目录数量一致, `quotaList`中的各个缓存容量上限以逗号(",")进行分隔,各缓存容量上限按先后顺序分配到对应的目录下.
例如,上述yaml配置意味着"/mnt/ssd0/cache"目录的最大缓存容量为"3Gi",而"/mnt/ssd1/cache"目录的最大缓存容量为"2Gi".

另一个与多目录存储配置相关的配置是缓存目录的使用策略.该策略由`alluxio.worker.allocator.class`这一Alluxio Property进行配置. Fluid默认使用`alluxio.worker.block.allocator.MaxFreeAllocator`作为数据缓存的放置策略.

Alluxio支持三种数据缓存的放置策略, 用户可通过设置Alluxio Property(`alluxio.worker.allocator.class`)修改该策略:
- "MaxFreeAllocator": 总是选择当前最大空余容量的缓存目录
- "RoundRobinAllocator": 以RoundRobin方式选择缓存目录
- "GreedyAllocator": 总是选择第一个可用的缓存目录

## 多层存储配置

以下为多层存储配置的一个示例:
```yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: dataset
spec:
  ...
  tieredstore:
    levels:
      - path: /dev/shm
        mediumtype: MEM
        quota: 2Gi
      - path: /mnt/ssd0/cache,/mnt/ssd1/cache
        mediumtype: SSD
        quotaList: 3Gi,2Gi
```

多层存储配置可以对每层进行各自的单目录或多目录配置.例如,在上述示例中,我们定义了两层数据缓存:其中第一层使用内存进行高速的数据访问,而第二层则使用SSD作为容量更大但访问速度稍慢的次级数据缓存.在第二层中,我们配置了多个目录(多块SSD磁盘)以均摊大量数据访问请求带来的压力.

`spec.tieredstore.levels`中定义的层级顺序不会影响Alluxio集群分层存储的层级顺序.在Alluxio集群启动前,Fluid会按照`mediumtype`对多个层级进行重排序,以保证数据访问速度快的存储介质("MEM" < "SSD" < "HDD")会被优先使用.

> 注意: 多层存储配置的Alluxio使用不同的方式计算存储使用量. 在目前的Fluid版本下,这会使得Alluxio已缓存比例(`Dataset.Cached`以及`Dataset.CachedPercentage`属性)受到一定的精确度影响. 

