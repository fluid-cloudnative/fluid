# DEMO - Dataset Toleration of Node Taints

## Background

Node affinity is a property of Kubernetes Pods that attracts them to a set of nodes, either as a preference or as a hard requirement. Taints are the opposite: they allow a node to repel a set of Pods.

Tolerations are applied to Pods. They allow, but do not require, the Pods to be scheduled onto nodes with matching taints.

Taints and tolerations work together to keep Pods away from unsuitable nodes. One or more taints can be applied to a node; the node does not accept any Pod that does not tolerate them.

In Fluid, a `Dataset` is schedulable too, so the resource object can also define tolerations. This means you can schedule where the cache is placed in the Kubernetes cluster just like you schedule your Pods.
This document gives a short demonstration of this feature.

## Set up the working directory
```shell
$ mkdir <any-path>/tolerations
$ cd <any-path>/tolerations
```

## Run the Demo
**List the nodes**
```shell
$ kubectl get no
NAME                       STATUS   ROLES    AGE    VERSION
cn-beijing.192.168.1.146   Ready    <none>   200d   v1.16.9-aliyun.1
```

**Taint the node**

```shell
$ kubectl taint nodes cn-beijing.192.168.1.146 hbase=true:NoSchedule
```
In the next step we can see the taint on the node.

**Check the node again**

```shell
$ kubectl get node cn-beijing.192.168.1.146 -oyaml | grep taints -A3
  taints:
  - effect: NoSchedule
    key: hbase
    value: "true"
```

The node now has a `NoSchedule` taint, so by default a dataset cannot be placed on it.

**Check the Dataset to create**
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
  tolerations:
    - key: hbase 
      operator: Equal 
      value: "true" 
EOF
```
The `spec` of this `Dataset` defines `tolerations`, which allow the data cache to be placed on the tainted node.

**Create the Dataset**
```shell
$ kubectl create -f dataset.yaml
dataset.data.fluid.io/hbase created
```

**Check the AlluxioRuntime to create**
```shell
$ cat<<EOF >runtime.yaml
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
        quota: 4Gi
        high: "0.95"
        low: "0.7"
EOF
```
This snippet contains Alluxio settings that Fluid uses to start an Alluxio instance. `spec.replicas` is set to 1, so Fluid starts an Alluxio instance with one Alluxio master and one Alluxio worker.

**Create the AlluxioRuntime and check its status**
```shell
$ kubectl create -f runtime.yaml
alluxioruntime.data.fluid.io/hbase created

$ kubectl get pod -o wide
NAME                 READY   STATUS    RESTARTS   AGE   IP              NODE                       NOMINATED NODE   READINESS GATES
hbase-master-0       2/2     Running   0          85m   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
hbase-worker-0       2/2     Running   0          63m   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
```
The Alluxio worker has started and runs on the tainted node.

**Check the AlluxioRuntime status**
```shell
$ kubectl get alluxioruntime hbase -o wide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE     AGE
hbase   1               1                 Ready          1               1                 Ready          0             0               Ready   4m3s
```

**Check the application to create**

We provide a sample application to show how Fluid schedules with the data cache. First check the application:

```shell
$ cat<<EOF >app.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  replicas: 1
  serviceName: "nginx"
  podManagementPolicy: "Parallel"
  selector: # define how the deployment finds the pods it manages
    matchLabels:
      app: nginx
  template: # define the pods specifications
    metadata:
      labels:
        app: nginx
    spec:
      tolerations:
      - key: hbase 
        operator: Equal 
        value: "true" 
      containers:
        - name: nginx
          image: nginx
          volumeMounts:
            - mountPath: /data
              name: hbase-vol
      volumes:
        - name: hbase-vol
          persistentVolumeClaim:
            claimName: hbase
EOF
```


**Run the application**

```shell
$ kubectl create -f app.yaml
statefulset.apps/nginx created
```

**Check the application status**
```shell
$ kubectl get pod -o wide -l app=nginx
NAME      READY   STATUS    RESTARTS   AGE    IP              NODE                       NOMINATED NODE   READINESS GATES
nginx-0   1/1     Running   0          2m5s   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
```

The Nginx Pod started and runs on the tainted node.

## Clean up
```shell
$ kubectl delete -f .

$ kubectl taint nodes cn-beijing.192.168.1.146 hbase=true:NoSchedule-
```
