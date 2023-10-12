# 示例 - 基于Runtime分层位置信息的应用Pod调度

在[Pod 调度优化](./pod_schedule_optimization.md)中介绍如何将应用Pod调度到具有数据缓存能力的节点。
但是在有些情况下，如果数据缓存的节点无法调度应用Pod，那么将Pod调度到离数据缓存节点比较近的节点，例如在同一个
zone中，其读写性能会比在不同的zone中要好。

Fluid 支持配置 K8s 集群中的分层位置信息，在Fluid 的 Helm Chart 中的`tiered-conf.yaml` 中

以下是具体的示例，假设 K8s 集群具有 zone 和 region 的位置信息，达到以下目标：
- 应用 Pod 未配置强制亲和调度时，优先调度到数据缓存的节点，如果不满足其次优先调度到同一个 zone，再其次调度到同一个 region；
- 应用 Pod 配置强制亲和调度时，只强制调度到同一个 zone 下，不需要强制调度到数据缓存的节点；

## 0. 前提条件
您使用的k8s版本需要支持 admissionregistration.k8s.io/v1（ Kubernetes version > 1.16 )
启用允许控制器集需要通过向 Kubernetes API 服务器传递一个标志来配置，确保你的集群进行了正常的配置
```yaml
--enable-admission-plugins=MutatingAdmissionWebhook
```
注意如果您的集群之前已经配置了其他的准入控制器，只需要增加 MutatingAdmissionWebhook 这个参数。


## 1. Fluid 配置分层位置信息

1） 在安装 Fluid 前配置

在 Helm Charts 的 `tiered-conf.yaml` 文件中定义分层位置的配置
- fluid.io/node 是 fluid 内置的亲和性，用于调度到数据缓存的节点
```yaml
tieredLocality:
  preferred:
    # fluid 内置的亲和性，用于调度到数据缓存的节点，名称不可修改
    - name: fluid.io/node
      weight: 100
    # zone 的 label 名称
    - name: topology.kubernetes.io/zone
      weight: 50
    # region 的 label 名称
    - name: topology.kubernetes.io/region
      weight: 10
  required:
    # 如果Pod 配置 强制亲和性，则强制亲和性匹配 zone
    # 配置多个，采用 And 语义
    - topology.kubernetes.io/zone
```

然后按照[Fluid 安装](../userguide/install.md) 安装 Fluid，安装好之后，在 Fluid namespace（默认fluid-system） 中存在
`tiered-locality-config` 的 ConfigMap，保存分层的位置信息配置。

2） 已经存在的 Fluid 集群，修改分层位置信息
对 Fluid namespace（默认fluid-system） 中存在`tiered-locality-config` 的 ConfigMap 进行修改，
添加相关的分层位置信息配置（见第一点），配置完成后，再下一次 webhook mutation 时会读取最新的配置进行Pod调度。

## 2. Runtime 配置相应的分层信息
可以通过 Dataset 的 nodeAffinity 或者 Runtime 的 NodeSelector 字段配置分层位置信息。

下面是通过 Dataset 的 nodeAffinity 配置分层位置信息，此时 Runtime 的 Worker 会部署在符合条件的节点上。
```yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: hbase
spec:
  mounts:
    - mountPoint: https://mirrors.tuna.tsinghua.edu.cn/apache/hbase/stable/
      name: hbase
  nodeAffinity:
    required:
      nodeSelectorTerms:
      	- matchExpressions:
          - key: topology.kubernetes.io/zone
            operator: In
            values: 
              - zone-a
          - key: topology.kubernetes.io/region
            operator: In
            values:
              - region-a
```

## 3. 应用 Pod 的调度

### 3.1 优先亲和性调度
**创建Pod**
```shell
$ cat<<EOF >nginx-1.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-1
  labels:
    fuse.serverful.fluid.io/inject: "true"
spec:
  containers:
    - name: nginx-1
      image: nginx
      volumeMounts:
        - mountPath: /data
          name: hbase-vol
  volumes:
    - name: hbase-vol
      persistentVolumeClaim:
        claimName: hbase
EOF
$ kubectl create -f nginx-1.yaml
```
示例中`metadata.labels`中新增`fuse.serverful.fluid.io/inject=true`以对该Pod开启Fluid的调度优化功能。

**查看Pod**

查看Pod的yaml文件，发现被注入了如下亲和性约束信息：

```yaml
spec:
  affinity:
    nodeAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
        - preference:
          matchExpressions:
            - key: fluid.io/s-default-hbase
              operator: In
              values:
                - "true"
          weight: 100
        - preference:
            matchExpressions:
              - key: topology.kubernetes.io/zone
                operator: In
                values:
                  - "zone-a"
          weight: 50
        - preference:
            matchExpressions:
              - key: topology.kubernetes.io/region
                operator: In
                values:
                  - "region-a"
          weight: 10         
```

该亲和性会达到以下效果：
- 如果数据缓存节点（具有`fluid.io/s-default-hbase`标签的节点）可调度，则将 Pod 调度到该节点；
- 如果数据缓存节点不可调度，则第一优先调度到同一个zone（“zone-a"）的节点，其次优先调度到到同一个region（”region-a")的节点。

### 3.2 强制亲和性调度

如果应用pod 指定强制指定数据集调度时
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-1
  labels:
    # 强制调度到 hbase dataset
    fluid.io/dataset.hbase.sched: required
    fuse.serverful.fluid.io/inject: "true"
spec:
  containers:
    - name: nginx-1
      image: nginx
      volumeMounts:
        - mountPath: /data
          name: hbase-vol
  volumes:
    - name: hbase-vol
      persistentVolumeClaim:
        claimName: hbase
```
pod 会被注入 required 节点亲和性，如下所示，强制调度到 "topology.kubernetes.io/zone" 为 "zone-a" 的节点
```yaml
spec:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
          - matchExpressions:
            - key: topology.kubernetes.io/zone
              operator: In
              values:
                - "zone-a"
```

### 3.3 注意事项

1. 如果应用 Pod 指定分层位置信息的亲和性（包括`spec.affinity` 和 `spec.nodeselector`），则 webhook 不会注入相关的位置亲和性，以用户的配置为准:
2. 分层位置信息的亲和性调度是全局性的配置，针对所有的Dataset 生效，不支持不同的Dataset的不同的亲和性配置；