

# Demo 

## 展示的价值

1.可以管理多个存储来源的数据集的生命周期  
2.数据集本身可以调度  
3.数据集的总量和缓存可以观测  
4.应用调度  


## 步骤

### 部署模式

1. 下载fluid

要部署fluid， 请确保安装了Helm 3。

```
wget http://kubeflow.oss-cn-beijing.aliyuncs.com/fluid-0.1.0.tgz
tar -xvf fluid-0.1.0.tgz
```


2. 使用Helm 3安装

```
helm install fluid fluid
NAME: fluid
LAST DEPLOYED: Tue Jul  7 11:22:07 2020
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
```


3. 查看运行结果


```
kubectl get po -n fluid-system
NAME                                  READY     STATUS    RESTARTS   AGE
controller-manager-6b864dfd4f-995gm   1/1       Running   0          32h
csi-nodeplugin-fluid-c6pzj          2/2       Running   0          32h
csi-nodeplugin-fluid-wczmq          2/2       Running   0          32h
```

4. 卸载

```
helm del fluid
kubectl delete crd `kubectl get crd | grep data.fluid.io| awk '{print $1}'` 
```


## 使用方法

1. 首先查看通过kubectl，查看当前的节点数

```
# kubectl get no
NAME                         STATUS    ROLES     AGE       VERSION
cn-hongkong.172.31.136.194   Ready     <none>    18d       v1.16.9-aliyun.1
cn-hongkong.172.31.136.196   Ready     <none>    4d18h     v1.16.9-aliyun.1
```


2. 查看其中包含GPU的节点

```
kubectl get no -l aliyun.accelerator/nvidia_name=Tesla-V100-SXM2-16GB
NAME                         STATUS    ROLES     AGE       VERSION
cn-hongkong.172.31.136.196   Ready     <none>    4d18h     v1.16.9-aliyun.1
```

3. 创建一个Dataset对象, 其中描述了数据集的来源

```yaml
cat << EOF > dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: cifar10
  #namespace: fluid-system
spec:
  mounts:
  - mountPoint: https://downloads.apache.org/hadoop/common/hadoop-3.2.1/
    name: hadoop
  - mountPoint: https://downloads.apache.org/spark/spark-2.4.6/
    name: spark
  - mountPoint: https://downloads.apache.org/hbase/2.2.5/
    name: hbase
  nodeAffinity:
    required:
       nodeSelectorTerms:
          - matchExpressions:
            - key: aliyun.accelerator/nvidia_name
              operator: In
              values:
              - Tesla-V100-SXM2-16GB
EOF
```

执行安装

```
kubectl create -f dataset.yaml
```

4. 此时查看状态, 可以看到这个dataset处于not bound状态，表示这个dataset无法使用

```yaml
apiVersion: v1
items:
- apiVersion: data.fluid.io/v1alpha1
  kind: Dataset
  metadata:
    creationTimestamp: 2020-07-09T11:34:39Z
    finalizers:
    - fluid-dataset-controller-finalizer
    generation: 1
    name: cifar10
    namespace: default
    resourceVersion: "633951652"
    selfLink: /apis/data.fluid.io/v1alpha1/namespaces/default/datasets/cifar10
    uid: 19db3223-7f40-4ed7-830e-1a36ec2713ee
  spec:
    mounts:
    - mountPoint: https://downloads.apache.org/hadoop/common/hadoop-3.2.1/
      name: hadoop
    - mountPoint: https://downloads.apache.org/spark/spark-2.4.6/
      name: spark
    - mountPoint: https://downloads.apache.org/hbase/2.2.5/
      name: hbase
    nodeAffinity:
      required:
        nodeSelectorTerms:
        - matchExpressions:
          - key: aliyun.accelerator/nvidia_name
            operator: In
            values:
            - Tesla-V100-SXM2-16GB
  status:
    conditions: []
    phase: NotBound
```

5. 创建一个Alluxio Runtime对象，用来描述支持这个数据集的runtime, 其中worker的数量为2， 

```yaml
cat << EOF > runtime.yaml
---
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: cifar10
  #namespace: fluid-system
spec:
  # Add fields here
  dataCopies: 3
  alluxioVersion:
    image: alluxio/alluxio
    imageTag: "2.3.0-SNAPSHOT"
    imagePullPolicy: Always
  tieredstore:
    levels:
    - mediumtype: MEM
      path: /dev/shm
      quota: 2Gi
      high: "0.95"
      low: "0.7"
      storageType: Memory
    - mediumtype: SSD
      path: /var/lib/docker/alluxio
      quota: 10Gi
      high: "0.95"
      low: "0.7"
      storageType: Disk
  properteies:
    alluxio.user.file.writetype.default: MUST_CACHE
    alluxio.master.journal.folder: /journal
    alluxio.master.journal.type: UFS
  master:
    replicas: 1
    jvmOptions:
      - "-Xmx4G"
  worker:
    replicas: 2
    jvmOptions:
      - "-Xmx4G"
  fuse:
    image: alluxio/alluxio-fuse
    imageTag: "2.3.0-SNAPSHOT"
    imagePullPolicy: Always
    jvmOptions:
      - "-Xmx4G "
      - "-Xms4G "
    # For now, only support local
    shortCircuitPolicy: local
    args:
      - fuse
      - --fuse-opts=ro,max_read=131072,attr_timeout=7200,entry_timeout=7200
EOF
kubectl create -f runtime.yaml
``` 

6. 查看alluxio runtime的状态， 可以看到目前只有1个节点符合条件， 可以提供的缓存能力为12GiB

```yaml
kubectl get alluxioruntime -oyaml
status:
    cacheStates:
      cacheCapacity: 12GiB
      cached: 0B
      cachedPercentage: 0%
    conditions:
    - lastProbeTime: 2020-07-09T07:48:17Z
      lastTransitionTime: 2020-07-09T07:48:17Z
      message: The master is initialized.
      reason: Master is initialized
      status: "True"
      type: MasterInitialized
    - lastProbeTime: 2020-07-09T07:49:58Z
      lastTransitionTime: 2020-07-09T07:48:37Z
      message: The master is ready.
      reason: Master is ready
      status: "True"
      type: MasterReady
    - lastProbeTime: 2020-07-09T07:52:43Z
      lastTransitionTime: 2020-07-09T07:49:38Z
      message: The workers are initialized.
      reason: Workers are initialized
      status: "True"
      type: WorkersInitialized
    - lastProbeTime: 2020-07-09T07:52:43Z
      lastTransitionTime: 2020-07-09T07:49:38Z
      message: The fuses are initialized.
      reason: Fuses are initialized
      status: "True"
      type: FusesInitialized
    - lastProbeTime: 2020-07-09T07:52:43Z
      lastTransitionTime: 2020-07-09T07:49:58Z
      message: The workers are partially ready.
      reason: Workers are ready
      status: "True"
      type: WorkersReady
    - lastProbeTime: 2020-07-09T07:52:43Z
      lastTransitionTime: 2020-07-09T07:49:58Z
      message: The fuses are partially ready.
      reason: Fuses are ready
      status: "True"
      type: FusesReady
    currentFuseNumberScheduled: 1
    currentMasterNumberScheduled: 1
    currentWorkerNumberScheduled: 1
    desiredFuseNumberScheduled: 2
    desiredMasterNumberScheduled: 1
    desiredWorkerNumberScheduled: 2
    fuseNumberAvailable: 1
    fuseNumberReady: 1
    fusePhase: Ready
    masterNumberReady: 1
    masterPhase: Ready
    valueFile: cifar10-alluxio-values
    workerNumberAvailable: 1
    workerNumberReady: 1
    workerPhase: PartialReady
```

7. 进而再查看dataset这个对象的状态, 此时数据集已经处于可以使用的状态

```yaml
status:
    cacheStates:
      cacheCapacity: 12GiB
      cached: 0B
      cachedPercentage: 0%
    conditions:
    - lastTransitionTime: 2020-07-09T10:34:36Z
      lastUpdateTime: 2020-07-09T11:19:53Z
      message: The ddc runtime is ready.
      reason: DatasetReady
      status: "True"
      type: Ready
    phase: Bound
    ufsTotal: 1.742GiB
```

8. 此时我们自动创建了pv和pvc， 这是Kubernetes中接入存储的标准方式

```
kubectl get pv,pvc
NAME                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS    CLAIM             STORAGECLASS   REASON    AGE
persistentvolume/cifar10   100Gi      RWX            Retain           Bound     default/cifar10                            5m57s

NAME                            STATUS    VOLUME    CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/cifar10   Bound     cifar10   100Gi      RWX                           5m57s
```

9. 再进一步, 创建使用该数据集的应用

```yaml
cat << EOF > test.yaml
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
      hostNetwork: true
      containers:
      - name: nginx
        image: nginx
        ports:
        - containerPort: 80
          protocol: TCP
        volumeMounts:
          - mountPath: /data
            name: cifar10
      volumes:
      - name: cifar10
        persistentVolumeClaim:
          claimName: cifar10
EOF
kubectl create -f test.yaml
statefulset.apps/nginx created
```

10. 这时我们看到只有一个Pod被部署下去

```
kubectl get po -owide -l app=nginx
NAME      READY     STATUS    RESTARTS   AGE       IP               NODE                         NOMINATED NODE   READINESS GATES
nginx-0   1/1       Running   0          78s       172.31.136.196   cn-hongkong.172.31.136.196   <none>           <none>
nginx-1   0/1       Pending   0          78s       <none>           <none>                       <none>           <none>
```

而另一个pod没有部署下去的原因, 是只有一个节点满足条件可以部署数据集，而和这个数据集相关的应用无法被部署下去

```
Events:
  Type     Reason            Age        From               Message
  ----     ------            ----       ----               -------
  Warning  FailedScheduling  <unknown>  default-scheduler  0/2 nodes are available: 1 node(s) didn't have free ports for the requested pod ports, 1 node(s) had volume node affinity conflict.
  Warning  FailedScheduling  <unknown>  default-scheduler  0/2 nodes are available: 1 node(s) didn't have free ports for the requested pod ports, 1 node(s) had volume node affinity conflict.
```

11. 此时我们再增加一下node label

```
kubectl label node cn-hongkong.172.31.136.194 aliyun.accelerator/nvidia_name=Tesla-V100-SXM2-16GB
```

此时我们发现符合要求的节点已经有两个了

```
kubectl get no -l aliyun.accelerator/nvidia_name=Tesla-V100-SXM2-16GB
NAME                         STATUS    ROLES     AGE       VERSION
cn-hongkong.172.31.136.194   Ready     <none>    19d       v1.16.9-aliyun.1
cn-hongkong.172.31.136.196   Ready     <none>    5d22h     v1.16.9-aliyun.1
```


12.  再次观测，可以发现cacheCapacity变成了24GB， 可用的节点变成了2个

```yaml
kubectl get alluxioruntime -oyaml
status:
    cacheStates:
      cacheCapacity: 24GiB
      cached: 0B
      cachedPercentage: 0%
    conditions:
    - lastProbeTime: 2020-07-10T09:38:07Z
      lastTransitionTime: 2020-07-10T09:38:07Z
      message: The master is initialized.
      reason: Master is initialized
      status: "True"
      type: MasterInitialized
    - lastProbeTime: 2020-07-10T09:40:08Z
      lastTransitionTime: 2020-07-10T09:38:27Z
      message: The master is ready.
      reason: Master is ready
      status: "True"
      type: MasterReady
    - lastProbeTime: 2020-07-10T11:36:37Z
      lastTransitionTime: 2020-07-10T09:39:28Z
      message: The workers are initialized.
      reason: Workers are initialized
      status: "True"
      type: WorkersInitialized
    - lastProbeTime: 2020-07-10T11:36:37Z
      lastTransitionTime: 2020-07-10T09:39:28Z
      message: The fuses are initialized.
      reason: Fuses are initialized
      status: "True"
      type: FusesInitialized
    - lastProbeTime: 2020-07-10T11:36:37Z
      lastTransitionTime: 2020-07-10T09:40:08Z
      message: The workers are partially ready.
      reason: Workers are ready
      status: "True"
      type: WorkersReady
    - lastProbeTime: 2020-07-10T11:36:37Z
      lastTransitionTime: 2020-07-10T09:40:08Z
      message: The fuses are partially ready.
      reason: Fuses are ready
      status: "True"
      type: FusesReady
    currentFuseNumberScheduled: 2
    currentMasterNumberScheduled: 1
    currentWorkerNumberScheduled: 2
    desiredFuseNumberScheduled: 2
    desiredMasterNumberScheduled: 1
    desiredWorkerNumberScheduled: 2
    fuseNumberAvailable: 2
    fuseNumberReady: 2
    fusePhase: Ready
    masterNumberReady: 1
    masterPhase: Ready
    valueFile: cifar10-alluxio-values
    workerNumberAvailable: 2
    workerNumberReady: 2
    workerPhase: Ready
```


13. 从dataset的维度, 也可以看到相应的信息


```yaml
kubectl get dataset -oyaml
status:
    cacheStates:
      cacheCapacity: 24GiB
      cached: 0B
      cachedPercentage: 0%
    conditions:
    - lastTransitionTime: 2020-07-10T11:36:58Z
      lastUpdateTime: 2020-07-10T12:03:30Z
      message: The ddc runtime is ready.
      reason: DatasetReady
      status: "True"
      type: Ready
    phase: Bound
    ufsTotal: 1.742GiB
```

14.  看到两个应用都启动了起来

```
kubectl get po -owide -l app=nginx
NAME      READY     STATUS    RESTARTS   AGE       IP               NODE                         NOMINATED NODE   READINESS GATES
nginx-0   1/1       Running   0          114m      172.31.136.196   cn-hongkong.172.31.136.196   <none>           <none>
nginx-1   1/1       Running   0          114m      172.31.136.194   cn-hongkong.172.31.136.194   <none>           <none>
```


15. 登录到Pod nginx-0 中

```
kubectl exec -it nginx-0 bash
# ls /data
hadoop	hbase  spark
# du -sh /data/hbase/hbase-2.2.5-client-bin.tar.gz
200M	/data/hbase/hbase-2.2.5-client-bin.tar.gz
# time cp /data/hbase/hbase-2.2.5-client-bin.tar.gz /dev/null
real	1m11.708s
user	0m0.002s
sys	0m0.047s
```


16. 登录到另一个pod nginx-1 中, 可以发现集群中其他节点访问数据都得了加速

```
time cp /data/hbase/hbase-2.2.5-client-bin.tar.gz /dev/null
real	0m1.040s
user	0m0.001s
sys	0m0.045s
```

