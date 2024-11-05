# Demo - Pod Scheduling Optimization
To help users better use Fluid, we provide a series of scheduling plugins. By automatically injecting affinity-related information for Pods, we optimize the scheduling results of Pods and improve the overall efficiency of cluster usage.

Specifically, Fluid, combined with Pod scheduling policies based on datasets layout, can achieve the following functions by injecting scheduling information into Pods through the webhook mechanism.

1. Support K8s native scheduler, as well as Volcano, Yunikorn, etc. to achieve Pod data affinity scheduling  
2. Scheduling Pods to nodes with data caching capability first  
3. Scheduling Pods forcibly to nodes with data caching capability by specifying the Pod Label
4. When Pods do not use data sets, they can avoid scheduling to nodes with cache as much as possible

## Prerequisites
The version of k8s you are using needs to support admissionregistration.k8s.io/v1 (Kubernetes version > 1.16 )
Enabling allowed controllers needs to be configured by passing a flag to the Kubernetes API server. Make sure that your cluster is properly configured.
```yaml
--enable-admission-plugins=MutatingAdmissionWebhook
```
Note that if your cluster has been previously configured with other allowed controllers, you only need to add the MutatingAdmissionWebhook parameter.

## Usage
**Check all nodes**
```shell
$ kubectl get nodes
NAME                      STATUS   ROLES    AGE   VERSION
node.172.16.0.16   Ready    <none>   16d   v1.20.4-aliyun.1
node.172.16.1.84   Ready    <none>   16d   v1.20.4-aliyun.1
```

**Check the Dataset resource object to be created**

```shell
$ cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: hbase
spec:
  mounts:
    - mountPoint: https://mirrors.tuna.tsinghua.edu.cn/apache/hbase/stable/
      name: hbase
---
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: hbase
spec:
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 2Gi
        high: "0.95"
        low: "0.7"
EOF
```

**Create Dataset resource objects**
```shell
$ kubectl create -f dataset-global.yaml
dataset.data.fluid.io/hbase created
```

This configuration file fragment contains a lot of Alluxio-related configuration information that will be used by Fluid to start an Alluxio instance.
The `spec.replicas` property in the above configuration fragment is set to 1, which indicates that Fluid will start an Alluxio instance with one Alluxio Master and one Alluxio Worker.

**Create AlluxioRuntime resources and check status**

```shell
$ kubectl create -f runtime.yaml
alluxioruntime.data.fluid.io/hbase created

$  kubectl get po -owide
NAME                 READY   STATUS    RESTARTS   AGE   IP             NODE                      NOMINATED NODE   READINESS GATES
hbase-master-0       2/2     Running   0          11m   172.16.0.16    node.172.16.0.16   <none>           <none>
hbase-worker-0       2/2     Running   0          10m   172.16.1.84    node.172.16.1.84   <none>           <none>
```
Here you can see that there is an Alluxio Worker successfully started and running on node 172.16.1.84.

## Running Demo 1: Create a Pod without a mounted dataset, it will be scheduled to a node as far away from the dataset as possible

**Create a Pod**
```shell
$ cat<<EOF >nginx-1.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-1
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
              - key: fluid.io/dataset-num
                operator: DoesNotExist
          weight: 100
```

As affected by affinity, the Pod is scheduled to the node.172.16.0.16 where there is no cache (i.e. no Alluxio Worker Pod running):

```shell
$ kubectl get pods nginx-1 -o  custom-columns=NAME:metadata.name,NODE:.spec.nodeName
NAME    NODE
nginx-1   node.172.16.0.16
```

## Running Demo 2: Create a Pod with mounted dataset, which will try to schedule to the node where the mounted dataset exists

**Create a Pod**

```shell
$ cat<<EOF >nginx-2.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-2
spec:
  containers:
    - name: nginx-2
      image: nginx
      volumeMounts:
        - mountPath: /data
          name: hbase-vol
  volumes:
    - name: hbase-vol
      persistentVolumeClaim:
        claimName: hbase
EOF
$ kubectl create -f nginx-2.yaml
```

**Check the Pod**

Checking the yaml file of Pod, shows that the following information has been injected:

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
```

Through the Webhook mechanism, the application Pod is injected with preferred affinity to the cache worker.

```shell
$ kubectl get pods nginx-2 -o  custom-columns=NAME:metadata.name,NODE:.spec.nodeName
NAME    NODE
nginx-1   node.172.16.1.84
```

From the results, we can see that the pod is scheduled to the node with the data cache (i.e., running the Alluxio Worker Pod).

## Running Demo 3: Create a Pod with mounted dataset, and schedule pod to the nodes with mounted dataset through specifying Pod Label

**Create a Pod**

The Label should be specified in metadata (in format `fluid.io/dataset.{dataset_name}.sched: required`). Such as `fluid.io/dataset.hbase.sched: required`, which indicates that this pod need to be scheduled to the node with dataset hbase cache. 

```shell
$ cat<<EOF >nginx-3.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-3
  labels:
    fuse.serverful.fluid.io/inject: "true"
    fluid.io/dataset.hbase.sched: required
spec:
  containers:
    - name: nginx-3
      image: nginx
      volumeMounts:
        - mountPath: /data
          name: hbase-vol
  volumes:
    - name: hbase-vol
      persistentVolumeClaim:
        claimName: hbase
EOF
$ kubectl create -f nginx-3.yaml
```

**Check the Pod**

Checking the yaml file of Pod, shows that the following information has been injected:

```yaml
spec:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: fluid.io/s-default-hbase
            operator: In
            values:
            - "true"
```

Through the Webhook mechanism, the application Pod is injected with preferred affinity to the cache worker.

```shell
$ kubectl get pods nginx-3 -o  custom-columns=NAME:metadata.name,NODE:.spec.nodeName
NAME    NODE
nginx-3   node.172.16.1.84
```

From the results, we can see that the pod is scheduled to the node with the data cache (i.e., running the Alluxio Worker Pod).