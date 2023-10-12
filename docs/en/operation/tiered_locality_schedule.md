# Demo - Pod Scheduling Base on Runtime Tiered Locality

In [Pod Scheduling Optimization](./pod_schedule_optimization.md), we introduce how to schedule application Pods to nodes
with cached data.

However, in some cases, if the data cached nodes cannot be scheduled with the application Pod, the Pod will be scheduled
to a node closer to the data cached nodes, such as on the same zone, its read and write performance will be better than in different zones.

Fluid supports configuring tiered locality information in K8s clusters, which can be found in the 'tiered conf.yaml' 
file of Fluid's Helm Chart.

The following is a specific example, assuming that the K8s cluster has locality information for zones and regions, achieving the following goals:
- When the application Pod is not configured with required dataset scheduling, prefer to schedule pod to data cached nodes.
If pods can not be scheduled in data cached nodes, prefer to be scheduled in the same zone.
If pods can not be scheduled in the same zone nodes too, then prefer to be scheduled in the same region;
- When using Pod to configure required dataset scheduling, require pod to be scheduled in the same zone of data cached nodes instead of the data cached nodes.

## 0. Prerequisites
The version of k8s you are using needs to support admissionregistration.k8s.io/v1 (Kubernetes version > 1.16 )
Enabling allowed controllers needs to be configured by passing a flag to the Kubernetes API server. Make sure that your cluster is properly configured.
```yaml
--enable-admission-plugins=MutatingAdmissionWebhook
```
Note that if your cluster has been previously configured with other allowed controllers, you only need to add the MutatingAdmissionWebhook parameter.

## 1. Configure Tiered Locality in Fluid

1）Configure before installing Fluid

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