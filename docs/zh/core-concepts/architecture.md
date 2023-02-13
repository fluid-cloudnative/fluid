# 系统架构

Fluid的整体架构：

![](../../../static/concepts/architecture.png)


Fluid有两个核心概念：Dataset和Runtime。为了支持这两个概念，Fluid的架构被逻辑地划分为控制平面和数据平面。


- 控制平面 

  - **Dataset/Runtime Manager**: 主要负责数据集和支持数据集的运行时在Kubernetes中的调度和编排。负责数据集的调度，迁移和缓存运行时的弹性扩缩容；同时支持数据集的自动化运维操作，比如控制细粒度的数据预热，比如可以指定预热某个指定文件夹；控制元数据备份和恢复，提升对于海量小文件场景的数据访问性能；设置缓存数据的pin策略，避免数据驱逐导致的性能震荡。


  - **Application Manager**: 主要关心使用数据集的应用Pod的调度和运行，分为两个核心组件：Scheduler和Webhook.

    - Scheduler: 结合从Runtime获取的数据集对应运行时的部署信息信息，对于Kubernetes集群中的Pod进行调度。将使用数据集的应用优先调度到含有数据缓存的节点。

    - Sidecar Webhook: 对于无法运行csi-pluign的Kubernetes环境， Sidecar webhook会将自动的将PVC替换成 FUSE sidecar，并且控制Pod中容器启动顺序，保证FUSE容器先启动。


 - 数据平面

   - **Runtime Plugin**: 可以扩展兼容多种分布式缓存引擎。
同时Fluid抽象出了共性特征，比如对于缓存描述：使用了什么缓存介质，缓存quota，缓存目录，这些都是共性的; 而分布式缓存引擎的拓扑抽象有一定的差异性，比如alluxiomaster和slave架构，Juice是只有worker P2P的架构，可以在Rungtime的CRD中进行配置。支持Alluxio，JuiceFS等Runtime。


   - **CSI Plugin**: 支持以






Fluid控制平面包括Dataset Controller，负责管理Dataset的通用操作；Runtime Controller，负责管理各种Runtime的生命周期；Scheduler plugin负责调度使用Dataset的Pod。

在数据平面上，我们通过CSI插件支持运行在ECS上的Pod，通过FUSE sidecar支持运行在ECI上的Pod。开发人员就不需要担心数据平面的实现。他们只需按照Runtime的开发指南集成Fluid即可。值得注意的是，Fluid也实现了CSI插件接口，但是与传统的CSI插件相比，有两个主要区别：1）FUSE客户端实现了容器化，以Pod形式运行，使其具有更好的可观测性，并可以单独设置资源配额；2）FUSE客户端与CSI插件完全解耦，不再需要将FUSE客户端构建到CSI插件容器镜像中。这使得FUSE客户端和CSI插件可以各自独立地演进。

