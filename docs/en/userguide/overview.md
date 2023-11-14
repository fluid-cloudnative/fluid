# Fluid Overview

## Target Functions and Scenarios

Fluid is an open source cloud native infrastructure component. In the treand of computation and stroage separation, the goal of Fluid is to build an efficient data abstraction layer for AI and Big Data cloud native applications. It provides the following functions:
1. Through data affinity scheduling and distributed cache acceleration, realizing the fusion of data and computation, to speed up the data access from computation.
2. Storing and managing data independently, and isolating the resource by Kubernetes namespace, to realize data isolation safely.
3. Unifying the data from different storage, to avoid the islanding effect of data.

Through the data abstraction layer served on Kubernetes, the data will just be like the fluid, waving across the storage sources(such as HDFS, OSS, Ceph) and the cloud native applications on Kubernetes. It can be moved, copied, evicted, transformed and managed flexibly. Besides, All the data operations are transparent to users. Users do not need to worry about the efficiency of remote data access nor the convenience of data source management. User just need to access the data abstracted from the Kubernetes native data volume, and all the left tasks and details are handled by Fluid.

Fluid currently mainly focuses on the dataset orchestration and application orchestration these two important scenarios. The dataset orchestration can arrange the cached dataset to the specific Kubernetes node, while the application orchestration can arrange the the applications to nodes with the pre-loaded datasets. These two can work together to form the co-orchestration scenario, which take both the dataset specifications and application characteristics into consideration during resouce scheduling.

## Why Cloud Native Needs Fluid

There exist a nature divergence between the cloud native environment and the earlier big data processing framework. Deeply affected by Google's GFS, MapReduce, BigTable influential papers, the open souce big data ecosystem keeps the concept of 'moving data but not moving computation' during system design. Therefore, data-intensive computing frameworks, such as Spark, Hive, MapReduce, aim to reduce data transmission, and consider more data locality architecture during the design. However, as time changes, for both consider the flexibility of the resource scalability and usage cost, compution and storage separation architecture has been widely used in the cloud native environment. sThus, the cloud native ecosystem need an component like Fluid to make up the lost data locality when the big data architecture embraces cloud native architecture.

Besides, in the cloud native environment, applications are usually deployed in the stateless micro-service style, but focus on data processing. However, the data-intensive frameworks and applications always focus on data abstraction, and schedules and executes the computing jobs and tasks. When data-intensive frameworks are deployed in cluod native environment, it needs component like Fluid to handle the data scheduling in cloud.

To resolve the issue that Kubernetes lacks the awareness and optimization for application data, Fluid put forward a series of innovative methods suach as co-orchestration, intelligent awareness, join-optimization, to form an efficient supporting platform for data-intensive applications in cloud native environment.


The architecture of Fluid in Kubernetes is as following:
<div align="center">
  <img src="../../../static/architecture.png" title="architecture" width="60%" height="60%" alt="">
</div>

## Concept

**Dataset**:  A set of logically related data that will be used by a computing engine, such as Spark for big data and TensorFlow for AI scenarios. The management of dataset has many metrics, has multiple dimensions, such as security, version management and data acceleration. And we hope to start with data acceleration and provide support for the management of data sets.

**Runtime**:  Security, version management and data acceleration, and defines a series of life cycle interfaces. You can implement them.

**AlluxioRuntime**: From [Alluixo](https://www.alluxio.org/), 
Fluid manages and schedules Alluxio Runtime to achieve dataset visibility, elastic scaling, and data migration. It is an engine which supports data management and caching of datasets.


## Demo
We provide demo to show how to improve the AI model traning speed in Cloud by using Fluid.

### Demo 1: Accelerate Remote File Accessing with Fluid

[![](../../../static/remote_file_accessing.png)](http://cloud.video.taobao.com/play/u/2987821887/p/1/e/6/t/1/277753111709.mp4)


### Demo 2: Machine Learning with Fluid

[![](../../../static/machine_learning.png)](http://cloud.video.taobao.com/play/u/2987821887/p/1/e/6/t/1/277528130570.mp4)

## Quick Start
Fluid needs to run on Kubernetes v1.14 or above version, also needs to support CSI storage. The deployment and management of Fluid Operator is through Helm v3 which is the package mangement tool on Kubernetes platform. Please make sure the Helm is correctly installed in the Kubernetes cluster before running Fluid.

You can refer to the following documents to insall and use Fluid.
- [English](docs/en/TOC.md)
- [简体中文](docs/zh/TOC.md)