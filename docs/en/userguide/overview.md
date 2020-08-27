# Fluid Overview

## Target Functions and Scenarios

Fluid is an open source cloud native infrastructure component. In the treand of computation and stroage separation, the goal of Fluid is to build an efficient data abstraction layer for AI and Big Data cloud native applications. It provides the following functions:
1. Through data affinity scheduling and distributed cache acceleration, realizing the fusion of data and computation, to speed up the data access from computation.
2. Storing and managing data independently, and isolating the resource by Kubernetes namespace, to realize data isolation safely.
3. Unifying the data from different storage, to avoid the islanding effect of data.

Through the data abstraction layer served on Kubernetes, the data will just be like the fluid, waving across the storage sources(such as HDFS, OSS, Ceph) and the cloud native applications on Kubernetes. It can be moved, copied, evicted, transformed and managed flexibly. Besides, All the data operations are transparent to users. Users do not need to worry about the efficiency of remote data access nor the convenience of data source management. User just need to access the data abstracted from the Kubernetes native data volume, and all the left tasks and details are handled by Fluid.

Fluid currently mainly focus on the dataset orchestration and application orchestration these two important scenarios. The dataset orchestration can arrange the cached dataset to the specific Kubernetes node, while the application orchestration can arrange the the applications to nodes with the pre-loaded datasets. These two can work together to form the co-orchestration scenario, which take both the dataset specifications and application characteristics into consideration during resouce scheduling.

云原生环境与更早的大数据处理框架在设计理念和机制上存在天然分歧。深受Google三篇论文GFS、MapReduce、BigTable影响的Hadoop大数据生态，从诞生之初即信奉和实践“移动计算而不是数据”的理念。因此以Spark，Hive，MapReduce为代表的数据密集型计算框架及其应用为减少数据传输，其设计更多地考虑数据本地化架构。但随着时代的变迁，为兼顾资源扩展的灵活性与使用成本，计算和存储分离的架构在更新兴的云原生环境中大行其道。因此云原生环境里需要类似Fluid这样的一款组件来补充大数据框架拥抱云原生以后的数据本地性的缺失。

## Why Cloud Native needs Fluid

此外，在云原生环境中，应用通常以无状态（Stateless）微服务化方式部署并不以数据处理为中心；而数据密集型框架和应用通常以数据抽象为中心，开展相关计算作业和任务的分配执行。当数据密集型框架融入云原生环境后，也需要像Fluid这样以数据抽象为中心的调度和分配框架来协同。

针对当前Kubernetes缺乏对应用数据的感知和优化，以及像Alluxio这样的数据编排引擎难以直接驱动云原生应用等架构层的局限，Fluid提出将数据应用协同编排、智能感知、联合优化等一系列创新方法，并且基于Alluxio形成一套云原生场景下数据密集型应用的高效支撑平台。


具体的架构参见下图：
<div>
  <img src="http://kubeflow.oss-cn-beijing.aliyuncs.com/Static/architecture.png" title="architecture">
</div>

## 演示
我们提供了视频的Demo，为您展示如何通过Fluid提升云上AI模型训练的速度。

### 演示 1: 多数据源联合访问

### 演示 2: Dawnbench性能测试

## 快速体验Fluid
Fluid需要运行在 Kubernetes v1.14 及以上版本，并且需要支持CSI存储。Fluid Operator的部署和管理是通过 Kubernetes 平台上的包管理工具 Helm v3实现的。运行 Fluid前请确保 Helm 已经正确安装在 Kubernetes 集群里。

你可以参照参考文档 [docs](https://github.com/fluid-cloudnative/docs-fluid)，安装和使用Fluid。
- [English](docs/en/TOC.md)
- [简体中文](docs/zh/TOC.md)