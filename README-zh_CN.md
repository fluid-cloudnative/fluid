# Fluid

[English](./README.md) | 简体中文

## 什么是FLuid

一个开源的Kubernetes原生的分布式数据集编排和加速引擎，主要服务于在线数据分析和机器学习。通过定义数据集这种自定义资源的抽象，

<div>
  <img src="http://kubeflow.oss-cn-beijing.aliyuncs.com/Static/architecture.png" title="architecture">
</div>

## 核心功能

- __数据加速__

	Fluid通过使用分布式缓存引擎（Alluxio Inside）提供数据加速，同时可以保障数据的**可观测性**，**可迁移性**和**自动化的水平扩展**

- __调度应用时考虑数据集的本地性__

  	Bring the data close to compute, and bring the compute close to data

- __自动数据预热__

  	提供自动的预热Kubernetes数据缓存的机制

- __多用户支持__

	用户可以创建和管理不同namespace的数据集，

- __一次性统一访问不同来源的底层数据（对象存储，HDFS和Ceph等存储)，适用于混合云场景__


## 先决条件

- Kubernetes version > 1.14, 支持CSI
- Golang 1.12+
- Helm 3

## 快速开始

你可以通过 [快速开始](docs/installation/installation_cn/README.md) 在Kubernetes集群中测试Fluid.

## 文档

如果需要详细了解Fluid的使用，请参考文档 [docs](https://github.com/fluid-cloudnative/docs-fluid)：

- [English](https://github.com/fluid-cloudnative/docs-fluid/blob/master/en/TOC.md)
- [简体中文](https://github.com/fluid-cloudnative/docs-fluid/blob/master/zh/TOC.md)

Fluid的文档维护在 [docs-fluid repository](https://github.com/fluid-cloudnative/docs-fluid). 

## 演示

### 演示 1: 多数据源联合访问

### 演示 2: Dawnbench性能测试

## 如何贡献

## 社区

随时提出您关心的问题，您可以和项目的维护者通过：

1.Slack

2.钉钉讨论群

<div>
  <img src="http://kubeflow.oss-cn-beijing.aliyuncs.com/Static/dingtalk.png" width="280" title="dingtalk">
</div>

## License

Fluid is under the Apache 2.0 license. See the [LICENSE](./LICENSE) file for details.