# 系统架构

Fluid的整体架构：

![](../../../static/concepts/architecture.png)


Fluid有两个核心概念：Dataset和Runtime。为了支持这两个概念，Fluid的架构被逻辑地划分为控制平面和数据平面。


- 控制平面 

  - **Dataset/Runtime Manager**: 主要关心数据集和支持数据集的运行时在Kubernetes中的调度和编排。负责数据集的调度，迁移和缓存运行时的弹性扩缩容；同时支持数据集的自动化运维操作，比如控制细粒度的数据预热，比如可以指定预热某个指定文件夹；控制元数据备份和恢复，提升对于海量小文件场景的数据访问性能；设置缓存数据的pin策略，避免数据驱逐导致的性能震荡。


  - **Application Manager**: 主要关心使用数据集的应用Pod的调度和运行，分为两个核心组件：Scheduler和Webhook.

    - Scheduler: 结合和Runtime获取的数据集排布信息，对于Kubernetes集群中的Pod进行调度。需要使用数据集的应用优先调度到含有数据缓存的节点，并且应用和所需要具体数据的匹配度，进行优选；而不使用该数据集则会避免调度到这样的节点。另一方面在调度应用的同时可以结合应用的调度决策信息，将数据预热到应用未来会运行的节点上，实现在深度学习场景中对于第一个epoch的缓存加速。



    - Webhook:






Fluid控制平面包括Dataset Controller，负责管理Dataset的通用操作；Runtime Controller，负责管理各种Runtime的生命周期；Scheduler plugin负责调度使用Dataset的Pod。

在数据平面上，我们通过CSI插件支持运行在ECS上的Pod，通过FUSE sidecar支持运行在ECI上的Pod。开发人员就不需要担心数据平面的实现。他们只需按照Runtime的开发指南集成Fluid即可。值得注意的是，Fluid也实现了CSI插件接口，但是与传统的CSI插件相比，有两个主要区别：1）FUSE客户端实现了容器化，以Pod形式运行，使其具有更好的可观测性，并可以单独设置资源配额；2）FUSE客户端与CSI插件完全解耦，不再需要将FUSE客户端构建到CSI插件容器镜像中。这使得FUSE客户端和CSI插件可以各自独立地演进。

