[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![CircleCI](https://circleci.com/gh/circleci/circleci-docs.svg?style=svg)](https://circleci.com/gh/fluid-cloudnative/fluid)
[![Build Status](https://travis-ci.org/fluid-cloudnative/fluid.svg?branch=master)](https://travis-ci.org/fluid-cloudnative/fluid)
[![codecov](https://codecov.io/gh/fluid-cloudnative/fluid/branch/master/graph/badge.svg)](https://codecov.io/gh/fluid-cloudnative/fluid)
[![Go Report Card](https://goreportcard.com/badge/github.com/fluid-cloudnative/fluid)](https://goreportcard.com/report/github.com/fluid-cloudnative/fluid)
# Fluid

[English](./README.md) | 简体中文

|![更新](http://kubeflow.oss-cn-beijing.aliyuncs.com/Static/bell-outline-badge.svg) 最新进展：|
|------------------|
|Mar. 16th, 2021. Fluid v0.5.0 **发布**! 提供一系列新功能，包括提供数据集缓存的在线弹性扩缩容，元数据备份与恢复，Fuse全局模式部署等，详情参见 [CHANGELOG](CHANGELOG.md)。|
|Nov. 6th, 2020. Fluid v0.4.0 **发布**! 提供一系列新功能和修复上一版本的遗留问题，包括提供主动的数据预热，详情参见 [CHANGELOG](CHANGELOG.md)。|
|Oct. 1st, 2020. Fluid v0.3.0 **发布**! 提供一系列新功能和修复上一版本的遗留问题，包括对于K8s通用的数据卷加速和主机目录加速，详情参见 [CHANGELOG](CHANGELOG.md)。|

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

## 重要概念

**Dataset**: 数据集是逻辑上相关的一组数据的集合，会被运算引擎使用，比如大数据的Spark，AI场景的TensorFlow。而这些数据智能的应用会创造工业界的核心价值。Dataset的管理实际上也有多个维度，比如安全性，版本管理和数据加速。我们希望从数据加速出发，对于数据集的管理提供支持。

**Runtime**: 实现数据集安全性，版本管理和数据加速等能力的执行引擎，定义了一系列生命周期的接口。可以通过实现这些接口，支持数据集的管理和加速。

**AlluxioRuntime**: 来源于[Alluixo](https://www.alluxio.org/)社区，是支撑Dataset数据管理和缓存的执行引擎实现。Fluid通过管理和调度Alluxio Runtime实现数据集的可见性，弹性伸缩， 数据迁移。

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

<details>
<summary>演示 1: 加速文件访问</summary>
<pre>
<a href="http://cloud.video.taobao.com/play/u/2987821887/p/1/e/6/t/1/277753111709.mp4" rel="nofollow"><img src="https://camo.githubusercontent.com/2ee9ef7de9eeb386f365a5d10f5defd12f08457f/687474703a2f2f6b756265666c6f772e6f73732d636e2d6265696a696e672e616c6979756e63732e636f6d2f5374617469632f72656d6f74655f66696c655f616363657373696e672e706e67" alt="" data-canonical-src="http://kubeflow.oss-cn-beijing.aliyuncs.com/Static/remote_file_accessing.png" style="max-width:100%;"></a>
</pre>
</details>

<details>
<summary>演示 2: 加速机器学习</summary>
<pre>
<a href="http://cloud.video.taobao.com/play/u/2987821887/p/1/e/6/t/1/277528130570.mp4" rel="nofollow"><img src="https://camo.githubusercontent.com/5688ab788da9f8cd057e32f3764784ce616ff0fd/687474703a2f2f6b756265666c6f772e6f73732d636e2d6265696a696e672e616c6979756e63732e636f6d2f5374617469632f6d616368696e655f6c6561726e696e672e706e67" alt="" data-canonical-src="http://kubeflow.oss-cn-beijing.aliyuncs.com/Static/machine_learning.png" style="max-width:100%;"></a>
</pre>
</details>

<details>
<summary>演示 3: 加速PVC</summary>
<pre>
<a href="http://cloud.video.taobao.com/play/u/2987821887/p/1/e/6/t/1/281779782703.mp4" rel="nofollow"><img src="https://camo.githubusercontent.com/7343be344cfebfd53619c1c8a70530ffd43d3d96/68747470733a2f2f696d672e616c6963646e2e636f6d2f696d6765787472612f69342f363030303030303030333331352f4f31434e303164386963425031614d4a614a576a5562725f2121363030303030303030333331352d302d7462766964656f2e6a7067" alt="" data-canonical-src="https://img.alicdn.com/imgextra/i4/6000000003315/O1CN01d8icBP1aMJaJWjUbr_!!6000000003315-0-tbvideo.jpg" style="max-width:100%;"></a>
</pre>
</details>

<details>
<summary>演示 4: 数据预热</summary>
<pre>
<a href="http://cloud.video.taobao.com/play/u/2987821887/p/1/e/6/t/1/287213603893.mp4" rel="nofollow"><img src="https://img.alicdn.com/imgextra/i4/6000000005626/O1CN01JJ9Fb91rQktps7K3R_!!6000000005626-0-tbvideo.jpg" alt="" style="max-width:100%;"></a>
</pre>
</details>

<details open>
<summary>演示 5: 在线不停机数据集缓存扩缩容</summary>
<pre>
<a href="http://cloud.video.taobao.com/play/u/2987821887/p/1/e/6/t/1/302459823704.mp4" rel="nofollow"><img src="https://img.alicdn.com/imgextra/i4/6000000004852/O1CN013kKkea1liGNWo2DOE_!!6000000004852-0-tbvideo.jpg" alt="" style="max-width:100%;"></a>
</pre>
</details>

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
