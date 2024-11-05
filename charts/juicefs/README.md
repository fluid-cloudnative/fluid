# JuiceFS

[JuiceFS](https://juicefs.com) is an open source high-performance POSIX file system released under GNU Affero General Public License v3.0. It is specially optimized for the cloud-native environment. Using the JuiceFS to store data, the data itself will be persisted in object storage (e.g. Amazon S3), and the metadata corresponding to the data can be persisted in various database engines such as Redis, MySQL, and TiKV according to the needs of the scene.

JuiceFS can simply and conveniently connect massive cloud storage directly to big data, machine learning, artificial intelligence, and various application platforms that have been put into production environment, without modifying the code, you can use massive cloud storage as efficiently as using local storage.

## Introduction

This chart bootstraps [JuiceFS](https://juicefs.com) client on a [Kubernetes](https://kubernetes.io/) cluster using the [Helm](https://helm.sh/) package manager.
