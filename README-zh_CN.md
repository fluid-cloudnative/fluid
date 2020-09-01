[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![CircleCI](https://circleci.com/gh/circleci/circleci-docs.svg?style=svg)](https://circleci.com/gh/fluid-cloudnative/fluid)
[![codecov](https://codecov.io/gh/fluid-cloudnative/fluid/branch/master/graph/badge.svg)](https://codecov.io/gh/fluid-cloudnative/fluid)
[![Go Report Card](https://goreportcard.com/badge/github.com/fluid-cloudnative/fluid)](https://goreportcard.com/report/github.com/fluid-cloudnative/fluid)
# Fluid

[English](./README.md) | 简体中文

## 什么是Fluid

Fluid是一个开源的Kubernetes原生的分布式数据集编排和加速引擎，主要服务于云原生场景下的数据密集型应用，例如大数据应用、AI应用等。通过定义数据集资源的抽象，实现如下功能：

<div align="center">
  <img src="http://kubeflow.oss-cn-beijing.aliyuncs.com/Static/architecture.png" title="architecture" width="60%" height="60%" alt="">
</div>

## 核心功能

- __数据集抽象原生支持__

  	将数据密集型应用所需基础支撑能力功能化，实现数据高效访问并降低多维管理成本

- __云上数据预热与加速__

	Fluid通过使用分布式缓存引擎（Alluxio inside）为云上应用提供数据预热与加速，同时可以保障缓存数据的**可观测性**，**可迁移性**和**自动化的水平扩展**

- __数据应用协同编排__

  	在云上调度应用和数据时候，同时考虑两者特性与位置，实现协同编排，提升性能

- __多命名空间管理支持__

	用户可以创建和管理不同namespace的数据集

- __异构数据源管理__

	一次性统一访问不同来源的底层数据（对象存储，HDFS和Ceph等存储)，适用于混合云场景

## 先决条件

- Kubernetes version > 1.14, 支持CSI
- Golang 1.12+
- Helm 3

## 快速开始

你可以通过 [快速开始](docs/zh/userguide/get_started.md) 在Kubernetes集群中测试Fluid.

## 文档

如果需要详细了解Fluid的使用，请参考文档 [docs](docs/README_zh.md)：

- [English](docs/en/TOC.md)
- [简体中文](docs/zh/TOC.md)

你也可以访问[南京大学Fluid项目主页](http://pasa-bigdata.nju.edu.cn/project/Fluid.html)来获取有关文档.

## 快速演示

### 演示 1: 加速文件访问

[![](http://kubeflow.oss-cn-beijing.aliyuncs.com/Static/remote_file_accessing.png)](http://cloud.video.taobao.com/play/u/2987821887/p/1/e/6/t/1/277753111709.mp4)


### 演示 2: 加速机器学习

[![](http://kubeflow.oss-cn-beijing.aliyuncs.com/Static/machine_learning.png)](http://cloud.video.taobao.com/play/u/2987821887/p/1/e/6/t/1/277528130570.mp4)

## 如何贡献

欢迎您的贡献，如何贡献请参考[CONTRIBUTING.md](CONTRIBUTING.md).

## 欢迎加入与反馈

Fluid让Kubernetes真正具有分布式数据缓存的基础能力，开源只是一个起点，需要大家的共同参与。大家在使用过程发现Bug或需要的Feature，都可以直接在 [GitHub](https://github.com/fluid-cloudnative/fluid)上面提 issue 或 PR，一起参与讨论。另外我们有一个钉钉群，欢迎您的参与和讨论。

钉钉讨论群
<div>
  <img src="http://kubeflow.oss-cn-beijing.aliyuncs.com/Static/dingtalk.png" width="280" title="dingtalk">
</div>

## 开源协议

Fluid采用Apache 2.0 license开源协议，详情参见[LICENSE](./LICENSE)文件。
