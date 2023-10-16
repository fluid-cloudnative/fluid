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

1) Configure before installing Fluid

Define the tiered locality configuration in the 'tiered conf.yaml' file of Helm Charts as below.
- fluid.io/node is the built-in name of Fluid, used to schedule pods to the data cached node
```yaml
tieredLocality:
  preferred:
    # fluid built-in name, used to schedule pods to the data cached node
    - name: fluid.io/node
      weight: 100
    # runtime worker's zone label name, can be changed according to k8s environment.
    - name: topology.kubernetes.io/zone
      weight: 50
    # runtime worker's region label name, can be changed according to k8s environment.
    - name: topology.kubernetes.io/region
      weight: 10
  required:
    # If Pod is configured with required affinity, then schedule the pod to nodes match the label.
    # Multiple names is the And relation.
    - topology.kubernetes.io/zone
```

Install Fluid following the document [Installation](../userguide/install.md). After installation, a configmap
named `tiered-locality-config` storing above configuration will exist in Fluid namespace(default `fluid-system`).

2) Modify tiered locality configuration in the existing Fluid cluster

Modify tiered location configuration (content see point 1) in the configMap named 'tiered local configuration' 
in the Fluid namespace (default `fluid-system`), the latest configuration will be read for Pod scheduling during the next webhook mutation.

## 2. Configure the tiered locality information for the Runtime
Tiered location information can be configured through the NodeAffinity field of the Dataset or the NodeSelector field of the Runtime.

The following is the configuration of tiered location information defined in the yaml of the Dataset. 
And the workers of the Runtime will be deployed on nodes matching these labels.
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

## 3. Application Pod Scheduling

### 3.1 Preferred Affinity Scheduling
**Creating the Pod**
```shell
$ cat<<EOF >nginx-1.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-1
  labels:
    # enable Fluid's scheduling optimization for the pod
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

**Check the Pod**

Checking the yaml file of Pod, shows that the following affinity constraint information has been injected:

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

These affinity will achieve the following effects:：
- If the data cached node (a node with the label 'fluid.io/s-default-hbase') is schedulable, schedule Pod to that node;
- If the data cached node is un-schedulable, prefer to schedule pod to nodes in the same zone ("zone-a");
- If the same zone nodes are un-schedulable, prefer to schedule pod to nodes in the same region ("region-a");
- All of the above are not met, schedule to other schedulable nodes in the cluster.


### 3.2 Required Affinity Scheduling

If sets pod with required dataset scheduling as below :
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-1
  labels:
    # required dataset scheduling
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
Pod will be injected with required node affinity, as shown below, forcing scheduling to nodes with value "zone-a" for label "topology.kubernetes.io/zone"  .
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

### 3.3 Note

1. If the application Pod specifies the affinity about tiered locality information (defined in 'spec.affinity' or 'spec.nodeselector'), webhook will
not inject the relevant location affinity, and the user's configuration will be kept:
2. The affinity scheduling of tiered location is a global configuration that takes effect for all datasets and does not support different affinity configurations for different datasets;
