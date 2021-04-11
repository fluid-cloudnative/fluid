# DEMO - Cache Co-locality for Workload Scheduling
In Fluid, remote files specified in `Dataset` object are schedulable, which means you are able to control where to put your data in a k8s cluster, 
just like what you may have done to Pods. Also, Fluid is able to make cache co-locality scheduling decisions for workloads to minimize overhead costs.

This demo will show you an overview about features mentioned above.

## Prerequisites
Before everything we are going to do, please refer to [Installation Guide](../userguide/install.md) to install Fluid on your Kubernetes Cluster, and make sure all the components used by Fluid are ready like this:
```shell
$ kubectl get pod -n fluid-system
NAME                                  READY   STATUS    RESTARTS   AGE
alluxioruntime-controller-5b64fdbbb-84pc6   1/1     Running   0          8h
csi-nodeplugin-fluid-fwgjh                  2/2     Running   0          8h
csi-nodeplugin-fluid-ll8bq                  2/2     Running   0          8h
dataset-controller-5b7848dbbb-n44dj         1/1     Running   0          8h
```

Normally, you shall see a Pod named "dataset-controller", a Pod named "alluxioruntime-controller" and several Pods named "csi-nodeplugin". 
The num of "csi-nodeplugin" Pods depends on how many nodes your Kubernetes cluster have(e.g. 2 in this demo), so please make sure all "csi-nodeplugin" Pods are working properly.

## Set Up Workspace
```shell
$ mkdir <any-path>/co-locality
$ cd <any-path>/co-locality
```

## Install Resources to Kubernetes
**Check all nodes in your Kubernetes cluster**
```shell
$ kubectl get nodes
NAME                       STATUS   ROLES    AGE     VERSION
cn-beijing.192.168.1.146   Ready    <none>   7d14h   v1.16.9-aliyun.1
cn-beijing.192.168.1.147   Ready    <none>   7d14h   v1.16.9-aliyun.1
```

**Label one of the nodes**
```shell
$ kubectl label nodes cn-beijing.192.168.1.146 hbase-cache=true
```
Since we'll use `NodeSelector` to manage where to put our data, we mark the desired node by labeling it.


**Check all nodes again**
```shell
$ kubectl get node -L hbase-cache
NAME                       STATUS   ROLES    AGE     VERSION            HBASE-CACHE
cn-beijing.192.168.1.146   Ready    <none>   7d14h   v1.16.9-aliyun.1   true
cn-beijing.192.168.1.147   Ready    <none>   7d14h   v1.16.9-aliyun.1   
```
Only one of the two nodes holds a label `hbase-cache=true`. In the following steps, we are going to make sure it's the only location the data cache can be put on.

**Check the `Dataset` object to be created**
```shell
$ cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: hbase
spec:
  mounts:
    - mountPoint: https://downloads.apache.org/hbase/stable/
      name: hbase
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: hbase-cache
              operator: In
              values:
                - "true"
EOF
```
We defined a `nodeSelectorTerm` in `Dataset` object's `spec` to make sure only nodes with label `hbase-cache=true` are considered to be available for the dataset. 

**Create the dataset object**
```shell
$ kubectl create -f dataset.yaml
dataset.data.fluid.io/hbase created
```

**Check the `AlluxioRuntime` object to be created**
```shell
$ cat<<EOF >runtime.yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: hbase
spec:
  replicas: 2
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 2Gi
        high: "0.95"
        low: "0.7"
EOF
```
In this snippet of yaml, there are many specifications used by Fluid to launch an Alluxio instance. The `spec.replicas` in the yaml above is set to 2, which means an Alluxio instance with 1 master and 2 workers is expected to be launched.

**Create the `AlluxioRuntime` object**
```shell
$ kubectl create -f runtime.yaml
alluxioruntime.data.fluid.io/hbase created

$ kubectl get pod -o wide
NAME                 READY   STATUS    RESTARTS   AGE    IP              NODE                       NOMINATED NODE   READINESS GATES
hbase-fuse-42csf     1/1     Running   0          104s   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
hbase-master-0       2/2     Running   0          3m3s   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
hbase-worker-l62m4   2/2     Running   0          104s   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
```
While two running workers are expected, there's only one running on the node with label `hbase-cache=true`. The `nodeSelectorTerm` stops another worker from being deployed.

**Check status of the `AlluxioRuntime` object**
```shell
$ kubectl get alluxioruntime hbase -o wide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE     AGE
hbase   1               1                 Ready          1               2                 PartialReady   1             2               PartialReady   4m3s
```
As expected, `Worker Phase` is `PartialReady` and `Ready Workers: 1` is less than `Desired Workers: 2`.

**Check the workload to be created**

A sample workload is provided to demonstrate how cache co-locality scheduling works. Let's check it out first:
```shell
$ cat<<EOF >app.yaml
apiVersion: apps/v1beta1
kind: StatefulSet
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  replicas: 2
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
      affinity:
        # prevent two Nginx Pod from being scheduled at the same Node
        # just for demonstrating co-locality demo
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values:
                - nginx
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
The `podAntiAffinity` property might be a little confusing. 
Here is the explanation: The `podAntiAffinity` property makes sure all pods created by the workload should be distributed across different nodes, which can provide us a clear view of how cache co-locality scheduling works. 
In short, it's just a property for demonstration, you don't need to put much focus on that :)


**Run the workload**

```shell
$ kubectl create -f app.yaml
statefulset.apps/nginx created
```

**Check status of the workload**
```shell
$ kubectl get pod -o wide -l app=nginx
NAME      READY   STATUS    RESTARTS   AGE    IP              NODE                       NOMINATED NODE   READINESS GATES
nginx-0   1/1     Running   0          2m5s   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
nginx-1   0/1     Pending   0          2m5s   <none>          <none>                     <none>           <none>
```
Only one Pod is ready, and running on the only node that matches the `nodeSelectorTerm`

**Check the reason why it's still not ready**
```shell
$ kubectl describe pod nginx-1
...
Events:
  Type     Reason            Age        From               Message
  ----     ------            ----       ----               -------
  Warning  FailedScheduling  <unknown>  default-scheduler  0/2 nodes are available: 1 node(s) didn't match pod affinity/anti-affinity, 1 node(s) didn't satisfy existing pods anti-affinity rules, 1 node(s) had volume node affinity conflict.
  Warning  FailedScheduling  <unknown>  default-scheduler  0/2 nodes are available: 1 node(s) didn't match pod affinity/anti-affinity, 1 node(s) didn't satisfy existing pods anti-affinity rules, 1 node(s) had volume node affinity conflict.
```
As you may have seen, for one reason, `podAntiAffinity` prevents `nginx-1` Pod from being scheduled together with `nginx-0`. For another, there's only one node satisfying the given affinity condition.

**Label another node**
```shell
$ kubectl label node cn-beijing.192.168.1.147 hbase-cache=true
```
Now all of the two nodes hold the same label `hbase-cache=true`, re-check all the pods:
```shell
$ kubectl get pod -o wide
NAME                 READY   STATUS    RESTARTS   AGE   IP              NODE                       NOMINATED NODE   READINESS GATES
hbase-fuse-42csf     1/1     Running   0          44m   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
hbase-fuse-kth4g     1/1     Running   0          10m   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
hbase-master-0       2/2     Running   0          46m   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
hbase-worker-l62m4   2/2     Running   0          44m   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
hbase-worker-rvncl   2/2     Running   0          10m   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
```
There're two running Alluxio workers now.

```shell
$ kubectl get alluxioruntime hbase -o wide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE   AGE
hbase   1               1                 Ready          2               2                 Ready          2             2               Ready        46m43s
```

```shell
$ kubectl get pod -l app=nginx -o wide
NAME      READY   STATUS    RESTARTS   AGE   IP              NODE                       NOMINATED NODE   READINESS GATES
nginx-0   1/1     Running   0          21m   192.168.1.146   cn-beijing.192.168.1.146   <none>           <none>
nginx-1   1/1     Running   0          21m   192.168.1.147   cn-beijing.192.168.1.147   <none>           <none>
```
Another nginx Pod is also no longer pending.

In conclusion, schedulable data cache and cache co-locality scheduling for workloads are both supported by Fluid. Usually, they work together and offer a more flexible way to users who need some data management in Kubernetes.

## Clean Up
```shell
$ kubectl delete -f .

# unlabel nodes
$ kubectl label node cn-beijing.192.168.1.146 hbase-cache-
$ kubectl label node cn-beijing.192.168.1.147 hbase-cache-
```
