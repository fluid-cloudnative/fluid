# CacheRuntime Integration Guide

# Installation

*   Install Fluid version that supports CacheRuntime.


```shell
helm repo add fluid https://fluid-cloudnative.github.io/charts

helm repo update

helm search repo fluid --devel

helm install fluid fluid/fluid --devel --version xxx -n fluid-system
```

# Integration

## Step 1. Plan Cluster Topology

First, you need to plan a cluster topology:

*   Determine the topology type and which components are included:


*   MasterSlave: Master/Worker/Client

*   P2P/DHT: Worker/Client

*   ClientOnly: Client


*   Determine the form and configuration of each component:


*   Stateful/Stateless - Determines the workload type

*   Standalone/Active-Standby/Cluster


The table below shows basic information examples for deploying several major cache topology types.

*   MasterSlave: CubeFS/Alluxio


| Topology |  | Settings                                                                                                                                                                                                                                                                  |
| --- | --- |---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Master |  | *   workLoadType: apps/v1/StatefulSet<br>    <br>*   Image configuration<br>   <br>*   Startup command<br>  <br>*   UFS mount command<br>   <br>*   HeadlessService needs to be created<br>    <br>*   Authentication keys need to be mounted                             |
| Worker: Used for single worker role definition |  | *   workLoadType: apps/v1/StatefulSet<br>    <br>*   Image configuration<br>    <br>*   Startup command<br>    <br>*   HeadlessService needs to be created<br>    <br>*   Authentication keys do NOT need to be mounted<br>    <br>*   TieredStore needs to be configured |
| Client | Fuse | *   Role: Posix client<br>    <br>*   workLoadType: apps/v1/DaemonSet<br>    <br>*   Image configuration<br>    <br>*   Startup command<br>    <br>*   Authentication parameters do NOT need to be mounted<br>    <br>*   TieredStore is NOT supported                    |

*   P2P Worker: JuiceFS


| Topology | Settings |
| --- | --- |
| Worker: Used for single worker role definition | *   workLoadType: apps/v1/StatefulSet<br>    <br>*   Image configuration<br>    <br>*   Startup command<br>    <br>*   HeadlessService<br>    <br>*   Authentication parameters need to be mounted<br>    <br>*   TieredStore is supported |
| Client | *   Role: Fuse client<br>    <br>*   workLoadType: apps/v1/DaemonSet<br>    <br>*   Image configuration<br>    <br>*   Startup command<br>    <br>*   Service is NOT required<br>    <br>*   Authentication parameters need to be mounted<br>    <br>*   TieredStore is supported |

## Step 2. Prepare Cache System Template

A cache system template in Fluid contains the following parts:

```yaml
├── Name # runtimeClassName is specified in CacheRuntime
├── FileSystemType # File system type, used for mount readiness verification
├── Topology
│   ├── Master[component]
│   ├── Worker[component]
│   └── client[component]
└── ExtraResources
    └── ConfigMaps
```

The component in Topology mainly contains the following content:

| Content | Description                 | Recommendation                                                                                                                                                                                                                      |
| --- |--------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| WorkloadType | The workload type of this component           | For stateful applications like Master/Worker, StatefulSet is the most common choice, as it can more easily cooperate with formatted DNS domain names provided by Headless Service for access<br>If Client is a Fuse client responsible for providing Posix access capability for pods on nodes, DaemonSet is generally used<br>If Client is an SDK proxy as a centralized stateless application, Deployment with ClusterIP type Service is generally used |
| Options | Default options, will be overridden by user settings |                                                                                                                                                                                                                         |
| Template | PodTemplateSpec native field   |                                                                                                                                                                                                                         |
| Service | Currently only supports Headless      |                                                                                                                                                                                                                         |
| Dependencies | ExtraResources     | Whether this component needs to mount additional ConfigMaps (the dependent ConfigMap information is defined in the ExtraResources field of CacheRuntimeClass).                                                                                                                                     |
| ExecutionEntries| MountUFS           | For Master-Worker architecture, when Master is Ready, the underlying file system mount operation needs to be executed.                                                                                                                                                                       |
| ExecutionEntries| ReportSummary      | How the cache system defines operations to obtain cache information metrics [Not supported in current version].                                                                                                                                                                                            |

### Step 2.1 Prepare K8s-adapted Native Images and Define Component workloadType and PodTemplate

You can first use native images, configure component **workloadType** and **PodTemplate**, manually start a fixed cache system in the K8s cluster, manually start the cache system in the pod, and make it locally accessible. This step is mainly used to clarify what K8s resources are needed and to prepare base images.

### Step 2.2 Clarify What Configurations CacheRuntime Should Provide for Components

Mainly clarify the following settings:

*   Service

*   Dependencies


### Step 2.3 Confirm Default ENV Provided by Fluid CacheRuntime for Components, Applicable by Scripts Inside Containers

| ENV | Description                               |
| --- |----------------------------------|
| FLUID_DATASET_NAME | Dataset name, generally used for isolation between groups in cache group concepts  |
| FLUID_DATASET_NAMESPACE | Namespace where the dataset is located               |
| FLUID_RUNTIME_CONFIG_PATH | Runtime configuration path provided by fluid             |
| FLUID_RUNTIME_MOUNT_PATH | Often used by Client, the target path where client performs mount action |
| FLUID_RUNTIME_COMPONENT_TYPE | Indicates whether the current component is master, worker, or client    |
| FLUID_RUNTIME_COMPONENT_SVC_NAME | If the component defines a service, this value is the service name    |

### Step 2.4 Create RuntimeClass Example and Field Description:

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: CacheRuntimeClass
metadata:
  name: demofs
fileSystemType: $fsType
topology:
  master:
    workloadType: # Create master with StatefulSet workload
      apiVersion: apps/v1
      kind: StatefulSet
    service: # Need to create Headless Service for master, only supported when workloadType is StatefulSet
      headless: {}
    dependencies:
      encryptOption: {} # Current not support
    podTemplateSpec:
      spec:
        restartPolicy: Always
        containers:
        - name: master
          image: $image
          args:
          - /bin/sh
          - -c
          - custom-endpoint.sh
          imagePullPolicy: IfNotPresent
  worker:
    workloadType: # Create worker with StatefulSet workload
      apiVersion: apps/v1
      kind: StatefulSet
    service:
      headless: {} # Need to create Headless Service for worker, only supported when workloadType is StatefulSet
    dependencies: {} 
    podTemplateSpec:
      spec:
        restartPolicy: Always
        containers:
        - name: worker
          image: $image
          args:
          - /bin/sh
          - -c
          - custom-endpoint.sh
          imagePullPolicy: IfNotPresent
  client:
    workloadType: # Create client with DaemonSet workload
      apiVersion: apps/v1
      kind: DaemonSet
    dependencies:
      encryptOption: {} # Need to provide encryptOption declared by user in dataset for client
    podTemplateSpec:
      spec:
        restartPolicy: Always
        containers:
        - name: client
          image: $image 
          securityContext: # Usually client needs to configure privileged for operating fuse device
            privileged: true
            runAsUser: 0
          args:
          - /bin/sh
          - -c
          - custom-endpoint.sh
          imagePullPolicy: IfNotPresent
```

### Step 2.5 User Creates Runtime

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: demofs
  namespace: default
spec:
  placement: Shared
  accessModes:
  - ReadWriteMany
  mounts:
  - name: demo
    mountPoint: "demofs:///"
    options:
      key1: value1
      key2: value2
    encryptOptions:
    - name: token
      valueFrom:
        secretKeyRef:
          name: jfs-secret
          key: token
    - name: access-key
      valueFrom:
        secretKeyRef:
          name: jfs-secret
          key: access-key
    - name: secret-key
      valueFrom:
        secretKeyRef:
          name: jfs-secret
          key: secret-key
---
apiVersion: data.fluid.io/v1alpha1
kind: CacheRuntime
metadata:
  name: demofs
  namespace: default
spec:
  runtimeClassName: demofs
  master:
    options: # master option
      key1: value1
      key2: value2
    replicas: 2 # master replica count
  worker:
    options: # worker option
      key1: value1
      key2: value2
    replicas: 2 # worker
    tieredStore:
      levels: # worker cache configuration 
      - quota: 40Gi
        low: "0.5"
        high: "0.8"
        path: "/cache-data"
        medium:
          emptyDir: # Use tmpfs as cache medium
            medium: Memory
  client:
    options:
      key1: value1
      key2: value2
    volumeMounts: # Can configure volumes and corresponding volumeMounts
    - name: demo
      mountPath: /mnt
  volumes:
  - name: demo
    persistentVolumeClaim:
      claimName: test

```

### Step 2.6 Confirm RuntimeConfig Provided by Fluid CacheRuntime for Components, Parse Parameters to Start Containers
> You can modify the entryPoint script based on the native image, first parse RuntimeConfig, generate corresponding configuration files, and then start the container.
> You can refer to the integration example in test/gha-e2e/curvine in the official repository.

In cacheruntime, all control plane processes are handled by Fluid. However, as a data caching engine, when providing services, the entire cache system requires **topology**, **data source**, **authentication**, and **cache information**. Fluid will provide this information to components through configuration files based on different Component roles. The component's internal process is responsible for parsing this configuration to perform environment variable configuration, data engine configuration file generation, and other operations. After preparation is complete, the data engine process can be started. For specific parsing details, please refer to the table below:

*   Taking the above resources as an example, the Config examples mounted by Master/Worker/Client and maintained by Fluid are as follows:
the `mounts`, `accessModes`, and `targetPath` fields in the JSON are all derived from the Dataset's Spec definition.

```json
{
  "mounts": [
    {
      "mountPoint": "s3://test",
      "options": {
        "access": "minioadmin",
        "endpoint_url": "http://minio:9000",
        "path_style": "true",
        "region_name": "us-east-1",
        "secret": "minioadmin"
      },
      "encryptOptions": {
        "access-key": "/etc/fluid/secrets/minio-secret/access-key",
        "secret-key": "/etc/fluid/secrets/minio-secret/secret-key"
      },
      "name": "minio",
      "path": "/minio"
    }
  ],
  "accessModes": [
    "ReadWriteMany"
  ],
  "targetPath": "/runtime-mnt/cache/default/curvine-demo/cache-fuse",
  "master": {
    "enabled": true,
    "name": "curvine-demo-master",
    "options": {
      "key1": "master-value1"
    },
    "replicas": 1,
    "service": {
      "name": "svc-curvine-demo-master"
    }
  },
  "worker": {
    "enabled": true,
    "name": "curvine-demo-worker",
    "options": {
      "key1": "worker-value1"
    },
    "replicas": 1,
    "service": {
      "name": "svc-curvine-demo-worker"
    }
  },
  "client": {
    "enabled": true,
    "name": "curvine-demo-client",
    "options": {
      "key1": "value1"
    },
    "service": {
      "name": ""
    }
  }
}
```
