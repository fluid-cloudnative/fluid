# Fluid

## What is Fluid?

An open source Kubernetes-native Distributed Dataset Manager and Orchestrator for Data Analysis and Machine Learning.

![architecture.png](./static/architecture.png)

## Features


- Data Accessing Accelerates for OSS, HDFS for free
- Portable and Scalable Dataset in Kubernetes with Infrastructure Knowledge
- Cache co-locality for workload scheduling
- Unify Dataset Access from different storage source
- Manage the dataset by preloading the data automatically and visibility of cache state

## Why Fluid?



## 前置条件

需要CSI支持


## Quickstart

1. 下载fluid

要部署fluid， 请确保安装了Helm 3。

```
wget http://kubeflow.oss-cn-beijing.aliyuncs.com/fluid-0.4.0.tgz
tar -xvf fluid-0.4.0.tgz
```


2. 创建namespace

```
kubectl create ns fluid-system
```

3. 使用Helm 3安装

```
helm install fluid fluid
NAME: fluid
LAST DEPLOYED: Tue Jul  7 11:22:07 2020
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
```


4. 查看运行结果

```
kubectl get po -n fluid-system
NAME                                  READY     STATUS    RESTARTS   AGE
controller-manager-6b864dfd4f-995gm   1/1       Running   0          32h
csi-nodeplugin-fluid-c6pzj          2/2       Running   0          32h
csi-nodeplugin-fluid-wczmq          2/2       Running   0          32h
```

5. 卸载

```
helm del fluid
kubectl delete crd `kubectl get crd | grep data.fluid.io| awk '{print $1}'` 
kubectl patch crd/datasets.data.fluid.io -p '{"metadata":{"finalizers":[]}}' --type=merge
```

## Who uses Fluid?


## Documentation
* [Get started here](docs/quick-start.md)
* [How to write Runtime specs](examples/README.md)
* [How to develop Fluid](docs/configure-artifact-repository.md)

## Features
* 极致的数据加速体验 （没有额外费用）
* 可以调度的数据集
* 缓存亲和性调度
* 可观测数据缓存


oproj.github.io/community/join-slack)