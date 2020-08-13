# Fluid

## What is Fluid?

An open source Kubernetes-native Distributed Dataset Manager and Orchestrator for Data Analysis and Machine Learning.

![architecture.png](./static/architecture.png)

## Features

- __Data Accessing Accelerates for OSS, HDFS, CEPH and Other underlayer storages__

  Fluid empowers Distributed Cache Capaicty(with Alluixo or other engines) in Kubernetes with  
  			- **Observability**  
  			- **Portability**  
  			- **Horizontal scalability**  
  on the cloud. 

- __Schedule data and compute with Cache Co-locality__

  Bring the data close to compute, and bring the compute close to data

- __Prefetch the data to cache automatically__

  Warm up the cache in Kubernetes automaticaly

- __Multi-tenant support__

  Users can create and manage multiple dataset in multiple namespaces

- __Unify the Data access from different underlayer storages__

  The data from the different storage can be consumed together 

## Prerequisites

- Kubernetes version > 1.14, and support CSI
- Golang 1.12+
- Helm 3

## Quick Start

You can follow our [Get Started](docs/installation/installation_cn/README.md) guide to quickly start a testing Kubernetes cluster.

## Documentation

You can see our documentation at [docs]() for more in-depth installation and instructions for production:

- [English]()
- [简体中文]()

All the Fluid documentation is maintained in the [docs-fluid repository](https://github.com/fluid-cloudnative/docs-fluid). 

* [Get started here](docs/quick-start.md)
* [How to write Runtime specs](examples/README.md)
* [How to develop Fluid](docs/configure-artifact-repository.md)

## Demo

### Demo 1: Unification

### Demo 2: Dawnbench

## Community

Feel free to reach out if you have any questions. The maintainers of this project are reachable via:

## Adopters

If you are intrested in Fluid and would like to share your experiences with others, you are warmly welcome to add your information on [ADOPTERS.md](ADOPTERS.md) page. We will continuousely discuss new requirements and feature design with you in advance.

## Contributing

Contributions are welcome and greatly appreciated. See [CONTRIBUTING.md](CONTRIBUTING.md) for details on submitting patches and the contribution workflow.

## License

Fluid is under the Apache 2.0 license. See the [LICENSE](./LICENSE) file for details.