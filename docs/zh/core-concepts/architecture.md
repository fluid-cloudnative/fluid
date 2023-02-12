# 系统架构

Fluid的整体架构：

![](../../../static/concepts/architecture.png)


Fluid有两个核心概念：Dataset和Runtime。为了支持这两个概念，Fluid的架构被逻辑地划分为控制平面和数据平面。








Fluid控制平面包括Dataset Controller，负责管理Dataset的通用操作；Runtime Controller，负责管理各种Runtime的生命周期；Scheduler plugin负责调度使用Dataset的Pod。

在数据平面上，我们通过CSI插件支持运行在ECS上的Pod，通过FUSE sidecar支持运行在ECI上的Pod。开发人员就不需要担心数据平面的实现。他们只需按照Runtime的开发指南集成Fluid即可。值得注意的是，Fluid也实现了CSI插件接口，但是与传统的CSI插件相比，有两个主要区别：1）FUSE客户端实现了容器化，以Pod形式运行，使其具有更好的可观测性，并可以单独设置资源配额；2）FUSE客户端与CSI插件完全解耦，不再需要将FUSE客户端构建到CSI插件容器镜像中。这使得FUSE客户端和CSI插件可以各自独立地演进。

