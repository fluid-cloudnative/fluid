# Fluid

English | [简体中文](./README-zh_CN.md)

## What is Fluid?

An open source Kubernetes-native Distributed Dataset Orchestrator and Accelerator for Data Analysis and Machine Learning.

<div>
  <img src="static/architecture.png" width="560" title="architecture">
</div>

## Features

- __Accelerate Data Accessing__

    Fluid empowers Distributed Cache Capaicty(Alluixo Inside) in Kubernetes with  **Observability**, **Portability**, **Horizontal scalability**

- __Schedule data and compute with Cache Co-locality__

  	Bring the data close to compute, and bring the compute close to data

- __Prefetch the data to cache automatically__

  	Warm up the cache in Kubernetes automaticaly

- __Multi-tenant support__

  	Users can create and manage multiple dataset in multiple namespaces

- __Unify the Data access for OSS, HDFS, CEPH and Other underlayer storages__

  	The data from the different storage can be consumed together 

## Prerequisites

- Kubernetes version > 1.14, and support CSI
- Golang 1.12+
- Helm 3

## Quick Start

You can follow our [Get Started](docs/installation/installation_cn/README.md) guide to quickly start a testing Kubernetes cluster.

## Documentation

You can see our documentation at [docs](https://github.com/fluid-cloudnative/docs-fluid) for more in-depth installation and instructions for production:

- [English](https://github.com/fluid-cloudnative/docs-fluid/blob/master/en/TOC.md)
- [简体中文](https://github.com/fluid-cloudnative/docs-fluid/blob/master/zh/TOC.md)

All the Fluid documentation is maintained in the [docs-fluid repository](https://github.com/fluid-cloudnative/docs-fluid). 

## Demo

### Demo 1: Unification

### Demo 2: Dawnbench

## Community

Feel free to reach out if you have any questions. The maintainers of this project are reachable via:

1.Slack:

2.DingTalk:

<div>
  <img src="static/dingtalk.png" width="280" title="dingtalk">
</div>


## Contributing

Contributions are welcome and greatly appreciated. See [CONTRIBUTING.md](CONTRIBUTING.md) for details on submitting patches and the contribution workflow.

## License

Fluid is under the Apache 2.0 license. See the [LICENSE](./LICENSE) file for details.