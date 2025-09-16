# CacheRuntime 对接社区文档

# 安装

*   安装支持cacheRutnime版本Fluid。


```shell
helm repo add fluid https://fluid-cloudnative.github.io/charts

helm repo update

helm search repo fluid --devel

helm install fluid fluid/fluid --devel --version xxx -nfluid-system
```

# 对接

## 步骤 1. 规划集群拓扑

首先需要规划一个集群拓扑：：

*   确定拓扑类型，包含哪些组件：


*   MasterSlave：Master/Worker/Client

*   P2P/DHT：Worker/Client

*   ClientOnly：Client


*   每个组件是什么形态以及组件的配置：


*   有状态/无状态 - 决定了工作负载类型

*   单机版/主备版/集群版


下表中展示了部署几个主要缓存拓扑类型所需要的基本信息示例。

*   MasterSlave：CubeFS/Alluxio


| 拓扑 |  | 设置 |
| --- | --- | --- |
| Master |  | *   workLoadType：appv1/StatefulSet<br>    <br>*   镜像配置<br>    <br>*   启动命令<br>    <br>*   需要创建HeadlessService<br>    <br>*   需要挂载认证密钥 |
| Woker：用于单一worker角色定义 |  | *   workLoadType：appv1/StatefulSet<br>    <br>*   镜像配置<br>    <br>*   启动命令<br>    <br>*   需要创建HeadlessService<br>    <br>*   不需要挂载认证密钥<br>    <br>*   需要配置TieredStore |
| Client | Fuse | *   角色：Posix客户端<br>    <br>*   workLoadType：appv1/Daemonset<br>    <br>*   镜像配置<br>    <br>*   启动命令<br>    <br>*   不需要挂载认证参数<br>    <br>*   不支持TieredStore |

*   P2P Worker：JuiceFS


| 拓扑 | 设置 |
| --- | --- |
| Woker：用于单一worker角色定义 | *   workLoadType：appv1/StatefulSet<br>    <br>*   镜像配置<br>    <br>*   启动命令<br>    <br>*   HeadlessService<br>    <br>*   需要挂载认证参数<br>    <br>*   支持TieredStore |
| Client | *   角色：Fuse客户端<br>    <br>*   workLoadType：appv1/Daemonset<br>    <br>*   镜像配置<br>    <br>*   启动命令<br>    <br>*   无需Service<br>    <br>*   需要挂载认证参数<br>    <br>*   支持TieredStore |

## 步骤 2. 准备缓存系统模板

一个缓存系统在Fluid中的缓存模板，包含以下几个部分

```yaml
├── Name #runtimeClassName由CacheRuntime中指定
├── FileSystemType #文件系统类型，用于做挂载就绪校验
├── Topology
│   ├── Master[component]
│   ├── Worker[component]
│   └── client[component]
└── ExtraResources
    └── ConfigMaps
```

Topology中comopent主要包含以下内容

| 内容 | 说明 | 建议 |
| --- | --- | --- |
| WorkloadType | 该组件的负载类型 | Master/Worker作为有状态应用，采取StatefulSet是最为常见的选择，好处在于可以更方便的配合Headless Service提供的格式化DNS域名进行访问<br>Client如果为Fuse客户端，需要负责为节点上的pod提供Posix访问能力，一般采取Daemonset<br>Client如果为SDK poxy，作为中心化的无状态应用，一般采用Deployment配合ClusterIP类型的Service使用 |
| Options | 默认options，会被用户设置覆盖 |  |
| PodTemplateSpec | workload原生字段 |  |
| Service | 目前仅支持Headless |  |
| Dependencies | EncryptOption | 该组件是否需要Fluid为其挂载Dataset中定义的用于访问数据源的访问密钥 |
|  | ExtraResources | 该组件是否需要extraResources【当前版本暂未支持】 |

### 步骤2.1 准备K8s适配的原生镜像及明确组件workloadType和PodTemplate

可以先利用原生镜像，配置组件**workloadType**和**PodTemplate**，在k8s集群中手动拉起一个固定的缓存系统，在pod里把缓存系统手动拉起，本地可访问。此步骤主要用于明确需要哪些K8s资源，以及准备基础镜像。

### 步骤2.2 明确组件需要哪些由CacheRuntime为其提供哪些配置

主要明确以下内容设置

*   Service

*   Dependencies


### 步骤2.3 确认Fluid CacheRuntime为组件提供的默认ENV，可被容器内脚本所应用

| ENV | 说明 |
| --- | --- |
| FLUID\_DATASET\_NAME | dataset名称，一般用于缓存组概念中用于group间的隔离 |
| FLUID\_DATASET\_NAMESPACE | dataset所在ns |
| FLUID\_RUNTIME\_CONFIG\_PATH | 由fluid提供的runtime配置路径 |
| FLUID\_RUNTIME\_MOUNT\_PATH | 常被Client所使用，client执行mount动作的目标路径 |
| FLUID\_RUNTIME\_COMPONENT\_TYPE | 表明当前组件是master，worker，还是client |

### 步骤2.4 创建RuntimeClass示例及字段说明：

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: CacheRuntimeClass
metadata:
  name: demofs
fileSystemType: $fsType
topology:
  master:
    workloadType: #以StatefulSet workload创建master
      apiVersion: apps/v1
      kind: StatefulSet
    service: #需要为master创建Headless Service，仅当workloadType为Statefulset时支持
      headless: {}
    dependencies:
      encryptOption: {} #需要为master提供用户在dataset中声明的encryptOption
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
    workloadType: #以StatefulSet workload创建worker
      apiVersion: apps/v1
      kind: StatefulSet
    service:
      headless: {} #需要为worker创建Headless Service，仅当workloadType为Statefulset时支持
    dependencies: {} #此处与#14区别为fluid不会为worker提供用户在dataset中声明的encryptOption
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
    workloadType: #DaemonSet workload创建client
      apiVersion: apps/v1
      kind: DaemonSet
    dependencies:
      encryptOption: {} #需要为client提供用户在dataset中声明的encryptOption
    podTemplateSpec:
      spec:
        restartPolicy: Always
        containers:
        - name: client
          image: $image 
          securityContext: #通常client需要配置privileged，用于操作fuse设备
            privileged: true
            runAsUser: 0
          args:
          - /bin/sh
          - -c
          - custom-endpoint.sh
          imagePullPolicy: IfNotPresent
```

### 步骤2.5 用户创建Runtime

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
    options: #master option
      key1: value1
      key2: value2
    replicas: 2 #master副本数
  worker:
    options: #worker option
      key1: value1
      key2: value2
    replicas: 2 #worker
    tieredStore:
      levels: #worker缓存配置 
      - quota: 40Gi
        low: "0.5"
        high: "0.8"
        path: "/cache-data"
        medium:
          emptyDir: #使用tmpfs作为缓存介质
            medium: Memory
  client:
    options:
      key1: value1
      key2: value2
    volumeMounts: #可配置volume和对应的volumeMounts
    - name: demo
      mountPath: /mnt
  volumes:
  - name: demo
    persistentVolumeClaim:
      claimName: test

```

### 步骤2.6 确认Fluid CacheRuntime为组件提供的RuntimeConfig，基于原生镜像改造entryPoint脚本，进行参数解析

在cacheruntime中，控制面的所有流程全都有Fluid来负责，但作为数据缓存引擎，提供服务时，需要整个缓存系统中的**拓扑**、**数据源、认证、缓存信息，**Fluid会根据不同的Component角色来通过配置文件的方式提供至组件内部，由组件内部进程负责解析该配置，来进行环境变量配置、数据引擎配置文件生成等操作，准备就绪后，可拉起数据引擎进程，解析过程中具体可参考下表：

*   以上述资源为例，Master/Worker/Client挂载的由Fluid维护的Config示例如下：


```json
{
  "mounts": [
    {
      "mountPoint": "demofs:///",
      "options": {
        "key1": "value1",
        "key2": "value2"
      },
      "name": "demo"
    }
  ],
  "targetPath": "/runtime-mnt/cache/default/cachefs-demo/thin-fuse",
  "accessModes": [
    "ReadWriteMany"
  ],
  "master": {
    "enabled": true,
    "options": {
      "key1": "value1",
        "key2": "value2"
    },
  },
  "worker": {
    "enabled": true,
    "options": {
      "key1": "value1",
        "key2": "value2"
    },
    "tieredStore": {
      "levels": [
        {
          "path": "/cache-data",
          "high": "0.8",
          "low": "0.5",
          "quota": "40Gi"
        }
      ]
    },
  },
  "client": {
    "enabled": true,
    "options": {
      "key1": "value1",
        "key2": "value2"
    },
    "encryptOption": {
      "access-key": "/etc/fluid/secrets/jfs-secret/access-key",
      "secret-key": "/etc/fluid/secrets/jfs-secret/secret-key",
      "token": "/etc/fluid/secrets/jfs-secret/token"
    },
  },
  "topology": {
    "master": {
      "podConfigs": [
        {
          "podName": "cachefs-demo-master-0",
          "podIP": "10.xx.xx.xx"
        },
        {
          "podName": "cachefs-demo-master-1",
          "podIP": "10.xx.xx.xx"
        }
      ],
      "service": {
        "name": "cachefs-demo-master"
      }
    },
    "worker": {
      "podConfigs": [
        {
          "podName": "cachefs-demo-worker-0",
          "podIP": "10.xx.xx.xx"
        },
        {
          "podName": "cachefs-demo-worker-1",
          "podIP": "10.xx.xx.xx"
        }
      ],
      "service": {
        "name": ""
      }
    },
    "client": {
      "podConfigs": [
        {
          "podName": "cachefs-demo-client-xxxxx",
          "podIP": "10.xx.xx.xx"
        },
        {
          "podName": "cachefs-demo-client-xxxxx",
          "podIP": "10.xx.xx.xx"
        }
      ],
      "service": {
        "name": ""
      }
    }
  }
}
```