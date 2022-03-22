# Demo - How to use Fuse NodeSelector 

In Fluid, the remote files defined in the `Dataset` resource object are schedulable, which means that you can manage where remote files are cached on the Kubernetes cluster just like your pods. Pods that perform computations can access data files through the Fuse client. Fuse client can be deployed globally in Kubernetes cluster, we can limit the deployment scope of Fuse client through `nodeSelector`.

This document will briefly show you the above features. Before that, you can visit [toc.md](https://github.com/fluid-cloudnative/fluid/blob/master/docs/en/TOC.md) to view the Fluid documentation to learn about the relevant knowledge details.

## Installation

Before running this example, please refer to the [Installation Documentation](https://github.com/fluid-cloudnative/fluid/blob/master/docs/en/userguide/install.md) to complete the installation, and check that each Fluid component is normal run:

```
$ kubectl get pod -n fluid-system
alluxioruntime-controller-f87f54fd6-pqj77   1/1     Running   0          14h
csi-nodeplugin-fluid-4pfmk                  2/2     Running   0          13h
csi-nodeplugin-fluid-qwlm4                  2/2     Running   0          13h
dataset-controller-bb95bb754-dsxww          1/1     Running   0          14h
fluid-webhook-66b77ccb8f-vvqjb              1/1     Running   0          13h
```

Typically, you will see a pod named `dataset-controller`, a pod named `alluxioruntime-controller`, and several pods named `csi-nodeplugin` running. Among them, the number of `csi-nodeplugin` these pods depends on the number of nodes in your Kubernetes cluster.

## Create new work environment

```
$ mkdir <any-path>/fuse-nodeselector-use
$ cd <any-path>/fuse-nodeselector-use
```

## Demo

**Check all nodes**

```
$ kubectl get nodes
NAME                      STATUS   ROLES    AGE   VERSION
cn-beijing.172.16.0.101   Ready    <none>   13h   v1.20.11-aliyun.1
cn-beijing.172.16.0.99    Ready    <none>   23h   v1.20.11-aliyun.1
```

**Use labels to identify nodes**

```
$ kubectl label nodes cn-beijing.172.16.0.101 select-node=true
node/cn-beijing.172.16.0.101 labeled
```

In the next steps, we will use the `NodeSelector` to decide which node to deploy the Fuse client. Firstly, we label the desired node here.

**Check nodes again**

```
$ kubectl get node -L select-node
NAME                      STATUS   ROLES    AGE   VERSION             SELECT-NODE
cn-beijing.172.16.0.101   Ready    <none>   13h   v1.20.11-aliyun.1   true
cn-beijing.172.16.0.99    Ready    <none>   23h   v1.20.11-aliyun.1 
```

Currently, out of all 2 nodes, only one node has the `select-node=true` label added, and next, we hope that the Fuse client will only be deployed on this node.

**Check the `Dataset` object to be created**

```
$ cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: hbase
spec:
  mounts:
    - mountPoint: https://mirrors.tuna.tsinghua.edu.cn/apache/hbase/stable/
      name: hbase
EOF
```

**Create the dataset object**

```
$ kubectl create -f dataset.yaml
dataset.data.fluid.io/hbase created
```

**Check the `AlluxioRuntime` object to be created**

```
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
        quota: 2Gi
        high: "0.95"
        low: "0.7"
  fuse:
    nodeSelector:
      select-node: "true"
EOF
```

In this configuration snippet, use `spec.fuse.nodeSelector` to set AlluxioRuntime to the node just marked with `select-node=true`

**Create `AlluxioRuntime` object **

```
$ kubectl create -f runtime.yaml
alluxioruntime.data.fluid.io/hbase created

$ kubectl get po -owide
NAME                    READY   STATUS    RESTARTS   AGE   IP             NODE                      NOMINATED NODE   READINESS GATES
hbase-master-0          2/2     Running   0          64m   172.16.0.101   cn-beijing.172.16.0.101   <none>           <none>
hbase-worker-0          2/2     Running   0          64m   172.16.0.101   cn-beijing.172.16.0.101   <none>           <none>
```

As you can see here, an Alluxio Worker has successfully started and is running on the node with the specified label (ie `select-node=true`).

**Check status of the `AlluxioRuntime` object**

```
$ kubectl get alluxioruntime hbase -o wide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE   AGE
hbase   1               1                 Ready          1               1                 Ready          0             0               Ready        12m
```

Here you can see that the number of Alluxio Workers is one, and the number of Alluxio Fuse is zero, because Alluxio Fuse needs to rely on a job application to start, and the following will show how to start Fuse through pod.

**Check the Deployment to be created**

Below is the YAML snippet for a simple Deployment with two copies and the selector tag `app=nginx-test`. Deployments are configured with `PodAntiAffinity` to ensure that the scheduler does not schedule all replicas on the same node.

```
$ cat<<EOF >nginx.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  selector:
    matchLabels:
      app: nginx-test
  replicas: 2
  template:
    metadata:
      labels:
        app: nginx-test
    spec:
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values:
                - nginx-test
            topologyKey: "kubernetes.io/hostname"
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

**Create the application**

```
$ kubectl create -f nginx.yaml
deployment.apps/nginx created
```

**Check the status of pods**

```
$ kubectl get po -owide
NAME                    READY   STATUS    RESTARTS   AGE   IP             NODE                      NOMINATED NODE   READINESS GATES
hbase-fuse-4jfgl        1/1     Running   0          41m   172.16.0.101   cn-beijing.172.16.0.101   <none>           <none>
hbase-master-0          2/2     Running   0          64m   172.16.0.101   cn-beijing.172.16.0.101   <none>           <none>
hbase-worker-0          2/2     Running   0          64m   172.16.0.101   cn-beijing.172.16.0.101   <none>           <none>
nginx-766564fc7-8vz4s   0/1     Pending   0          41m   <none>         <none>                    <none>           <none>
nginx-766564fc7-rtmwh   1/1     Running   0          41m   10.73.0.135    cn-beijing.172.16.0.101   <none>           <none>
```

It can be seen that the Fuse client is only started on the node `cn-beijing.172.16.0.101` with the `select-node=true` label.

One of the two pods (`nginx-766564fc7-rtmwh`) is started on node `cn-beijing.172.16.0.101` . The other pod (`nginx-766564fc7-8vz4s`), which can only be scheduled on another node `cn-beijing.172.16.0.99` due to the `PodAntiAffinity` constraint,  is in pending state because of the `spec.fuse.nodeSelector` constraint.

**start all pods**

In order to prove that the pod is not scheduled because of the limitation of `spec.fuse.nodeSelector`, we also label the node `cn-beijing.172.16.0.99` with `select-node=true`, you can find the Fuse Pod and Nginx Pod of this node All started.

```
$ kubectl label nodes cn-beijing.172.16.0.99 select-node=true
$ kubectl get po -owide
NAME                    READY   STATUS    RESTARTS   AGE    IP             NODE                      NOMINATED NODE   READINESS GATES
hbase-fuse-4jfgl        1/1     Running   0          138m   172.16.0.101   cn-beijing.172.16.0.101   <none>           <none>
hbase-fuse-k95g5        1/1     Running   0          28s    172.16.0.99    cn-beijing.172.16.0.99    <none>           <none>
hbase-master-0          2/2     Running   0          161m   172.16.0.101   cn-beijing.172.16.0.101   <none>           <none>
hbase-worker-0          2/2     Running   0          161m   172.16.0.101   cn-beijing.172.16.0.101   <none>           <none>
nginx-766564fc7-8vz4s   1/1     Running   0          138m   10.73.0.22     cn-beijing.172.16.0.99    <none>           <none>
nginx-766564fc7-rtmwh   1/1     Running   0          138m   10.73.0.135    cn-beijing.172.16.0.101   <none>           <none>
```

It can be found that the Fuse on the node `cn-beijing.172.16.0.99` is also started, and the Nginx Pod (`nginx-766564fc7-8vz4s`) can also be scheduled on this node.

## Clean up the environment

```
$ kubectl delete -f .

$ kubectl label node cn-beijing.172.16.0.101 select-node-
$ kubectl label node cn-beijing.172.16.0.99 select-node-
```