<div align="left">
    <img src="http://kubeflow.oss-cn-beijing.aliyuncs.com/Static/fluid_logo.jpg" title="architecture" height="11%" width="11%" alt="">
</div>


[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![CircleCI](https://circleci.com/gh/circleci/circleci-docs.svg?style=svg)](https://circleci.com/gh/fluid-cloudnative/fluid)
[![Build Status](https://travis-ci.org/fluid-cloudnative/fluid.svg?branch=master)](https://travis-ci.org/fluid-cloudnative/fluid)
[![codecov](https://codecov.io/gh/fluid-cloudnative/fluid/branch/master/graph/badge.svg)](https://codecov.io/gh/fluid-cloudnative/fluid)
[![Go Report Card](https://goreportcard.com/badge/github.com/fluid-cloudnative/fluid)](https://goreportcard.com/report/github.com/fluid-cloudnative/fluid)

# Fluid
English | [简体中文](./README-zh_CN.md)

|![notification](http://kubeflow.oss-cn-beijing.aliyuncs.com/Static/bell-outline-badge.svg) What is NEW!|
|------------------|
|Nov 6th, 2020. Fluid v0.4.0 is **RELEASED**! It provides various features and bugfix, such as Prefetch Dataset automatically before using it, please check the [CHANGELOG](CHANGELOG.md) for details.|
|Oct 1st, 2020. Fluid v0.3.0 is **RELEASED**! It provides various features and bugfix, such as Data Access Acceleration For Persistent Volume and Hostpath mode in K8s, please check the [CHANGELOG](CHANGELOG.md) for details.|

## What is Fluid?

Fluid is an open source Kubernetes-native Distributed Dataset Orchestrator and Accelerator for data-intesive applications, such as big data and AI applications.
<div align="center">
    <img src="http://kubeflow.oss-cn-beijing.aliyuncs.com/Static/architecture.png" title="architecture" height="60%" width="60%" alt="">
</div>

## Features

- __Native Support for DataSet Abstraction__

	Make the abilities needed by data-intensive applictions as navtive-supported functions, to achieve efficient data access and reduce the cost of multidimensional management.

- __Cloud Data Warming up and Accessing Acceleration__

    Fluid empowers Distributed Cache Capaicty(Alluixo inside) in Kubernetes with  **Observability**, **Portability**, **Horizontal Scalability**

- __Co-Orchestration for Data and Application__

    During application scheduling and data placement on cloud, taking both the app's characteristics and data location into consideration, to improve the performance.

- __Support Multiple Namespaces Management__

  	User can create and manage datasets in multiple namespaces

- __Support Heterogeneous Data Source Management__

  	Unify the Data access for OSS, HDFS, CEPH and Other underlayer storages

## Key Concepts

**Dataset**:  A set of logically related data that will be used by a computing engine, such as Spark for big data and TensorFlow for AI scenarios. The management of dataset has many metrics, has multiple dimensions, such as security, version management and data acceleration. And we hope to start with data acceleration and provide support for the management of data sets.

**Runtime**:  Security, version management and data acceleration, and defines a series of life cycle interfaces. You can implement them.

**AlluxioRuntime**: From [Alluixo](https://www.alluxio.org/), 
Fluid manages and schedules Alluxio Runtime to achieve dataset visibility, elastic scaling, and data migration. It is an engine which supports data management and caching of datasets.

## Prerequisites

- Kubernetes version > 1.14, and support CSI
- Golang 1.12+
- Helm 3

## Quick Start

You can follow our [Get Started](docs/en/userguide/get_started.md) guide to quickly start a testing Kubernetes cluster.

## Documentation

You can see our documentation at [docs](docs/README.md) for more in-depth installation and instructions for production:

- [English](docs/en/TOC.md)
- [简体中文](docs/zh/TOC.md)

## Qucik Demo

<details>
<summary>Demo 1: Accelerate Remote File Accessing with Fluid</summary>
<pre>
<a href="http://cloud.video.taobao.com/play/u/2987821887/p/1/e/6/t/1/277753111709.mp4" rel="nofollow"><img src="https://camo.githubusercontent.com/2ee9ef7de9eeb386f365a5d10f5defd12f08457f/687474703a2f2f6b756265666c6f772e6f73732d636e2d6265696a696e672e616c6979756e63732e636f6d2f5374617469632f72656d6f74655f66696c655f616363657373696e672e706e67" alt="" data-canonical-src="http://kubeflow.oss-cn-beijing.aliyuncs.com/Static/remote_file_accessing.png" style="max-width:100%;"></a>
</pre>
</details>

<details>
<summary>Demo 2: Machine Learning with Fluid</summary>
<pre>
<a href="http://cloud.video.taobao.com/play/u/2987821887/p/1/e/6/t/1/277528130570.mp4" rel="nofollow"><img src="https://camo.githubusercontent.com/5688ab788da9f8cd057e32f3764784ce616ff0fd/687474703a2f2f6b756265666c6f772e6f73732d636e2d6265696a696e672e616c6979756e63732e636f6d2f5374617469632f6d616368696e655f6c6561726e696e672e706e67" alt="" data-canonical-src="http://kubeflow.oss-cn-beijing.aliyuncs.com/Static/machine_learning.png" style="max-width:100%;"></a>
</pre>
</details>

<details>
<summary>Demo 3: Accelerate PVC with Fluid</summary>
<pre>
<a href="http://cloud.video.taobao.com/play/u/2987821887/p/1/e/6/t/1/281779782703.mp4" rel="nofollow"><img src="https://camo.githubusercontent.com/7343be344cfebfd53619c1c8a70530ffd43d3d96/68747470733a2f2f696d672e616c6963646e2e636f6d2f696d6765787472612f69342f363030303030303030333331352f4f31434e303164386963425031614d4a614a576a5562725f2121363030303030303030333331352d302d7462766964656f2e6a7067" alt="" data-canonical-src="https://img.alicdn.com/imgextra/i4/6000000003315/O1CN01d8icBP1aMJaJWjUbr_!!6000000003315-0-tbvideo.jpg" style="max-width:100%;"></a>
</pre>
</details>

<details open>
<summary>Demo 4: Preload dataset with Fluid</summary>
<pre>
<a href="http://cloud.video.taobao.com/play/u/2987821887/p/1/e/6/t/1/287213603893.mp4" rel="nofollow"><img src="https://img.alicdn.com/imgextra/i4/6000000005626/O1CN01JJ9Fb91rQktps7K3R_!!6000000005626-0-tbvideo.jpg" alt="" style="max-width:100%;"></a>
</pre>
</details>

## Community

Feel free to reach out if you have any questions. The maintainers of this project are reachable via:

DingTalk:

<div>
  <img src="http://kubeflow.oss-cn-beijing.aliyuncs.com/Static/dingtalk.png" width="280" title="dingtalk">
</div>


## Contributing

Contributions are welcome and greatly appreciated. See [CONTRIBUTING.md](CONTRIBUTING.md) for details on submitting patches and the contribution workflow.

## Open Srouce License

Fluid is under the Apache 2.0 license. See the [LICENSE](./LICENSE) file for details.
