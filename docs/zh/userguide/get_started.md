# Fluid 快速上手    
本文档介绍了如何创建 Kubernetes 集群环境，通过 Helm 完成 Fluid 安装部署，并使用 Fluid 创建数据集。  

## 创建 Kubernetes 集群  
Fluid 需要 Kubernetes 环境，根据你的使用经历选择最适合你的方案:  

- 你已经有了一个 Kubernetes 环境，并满足 Kubernetes :版本>=1.14，可以直接[部署Fluid](#部署Fluid) 
- 你之前没有使用过 Kubernetes，可以使用 Minikube 创建 Kubernetes 集群.     
[Minikube](https://kubernetes.io/docs/setup/minikube/)可以在虚拟机中创建一个 Kubernetes 集群，可在 macOS, Linux 和 Windows 上运行。

请确保满足以下要求:      
  - [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) :版本 1.0.0+   
  - [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl) :  版本  1.14+               

安装好Minikube之后:
```shell
minikube start
```
如果安装成功的话,会出现类似的提示信息:
```shell
  Darwin 10.14.5 上的 minikube v1.12.1
```
使用 `kubectl`访问新创建的 Kubernetes 集群
```shell
$ kubectl get pods
NAME                                READY   STATUS    RESTARTS   AGE
nginx-deployment-558fc78868-kvjnf   1/1     Running   1          4d12h
nginx-deployment-558fc78868-kx9gt   1/1     Running   1          4d12h
```

## 部署Fluid
开始之前，确保已满足以下要求：

- 使用 `kubectl` 可以成功访问到 Kubernetes 集群
- [Helm](https://helm.sh/docs/intro/install/) : Helm 3 已安装

1. 获取 Fluid  
```shell
git clone https://github.com/fluid-cloudnative/fluid.git 
```  
2. 使用 Helm 安装 Fluid
```shell
helm install fluid fluid
NAME: fluid
LAST DEPLOYED: Tue Jul  7 11:22:07 2020
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
```  
3. 查看安装结果 
```shell
kubectl get pod -n fluid-system
NAME                                  READY     STATUS    RESTARTS   AGE
controller-manager-6b864dfd4f-995gm   1/1       Running   0          32h
csi-nodeplugin-fluid-c6pzj          2/2       Running   0          32h
csi-nodeplugin-fluid-wczmq          2/2       Running   0          32h
```

## 创建dataset
Fluid提供了云原生的数据加速和管理能力，并抽象出了`数据集`概念方便用户管理，接下来将演示如何用 Fluid 创建一个数据集。   

1. 通过CRD文件创建一个Dataset对象，其中描述了数据集的来源。
```yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: demo
spec:
  mounts:
    - mountPoint: https://mirror.bit.edu.cn/apache/spark/spark-3.0.0/
      name: spark
```  
执行安装

```
kubectl create -f dataset.yaml
```
dataset创建以后处于 *not bound* 状态，需要绑定 runtime 才能使用。


2. 同样根据 alluxioRuntime的CRD文件创建一个 *Alluxio* Runtime 对象，用来描述支持这个数据集的 runtime。
```yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: demo
spec:
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 2Gi
        high: "0.95"
        low: "0.7"
        storageType: Memory
  properties:
    alluxio.user.file.writetype.default: MUST_CACHE
    alluxio.master.journal.folder: /journal
    alluxio.master.journal.type: UFS
    alluxio.user.block.size.bytes.default: 256MB
    alluxio.user.streaming.reader.chunk.size.bytes: 256MB
    alluxio.user.local.reader.chunk.size.bytes: 256MB
    alluxio.worker.network.reader.buffer.size: 256MB
    alluxio.user.streaming.data.timeout: 300sec
  master:
    jvmOptions:
      - "-Xmx4G"
  worker:
    jvmOptions:
      - "-Xmx4G"
  fuse:
    jvmOptions:
      - "-Xmx4G "
      - "-Xms4G "
    # For now, only support local
    shortCircuitPolicy: local
    args:
      - fuse
      - --fuse-opts=direct_io,ro,max_read=131072
```
使用`kubectl`完成创建  

```shell
kubectl create -f runtime.yaml  
``` 

3. 接下来，我们创建一个应用容器来使用该数据集，我们将多次访问同一数据，并比较访问时间来展示 Fluid 的加速效果。
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: demo-app
spec:
  containers:
    - name: demo
      image: nginx
      volumeMounts:
        - mountPath: /data
          name: demo
  volumes:
    - name: demo
      persistentVolumeClaim:
        claimName: demo
```

4. 登录到应用容器中访问数据，初次访问会花费更长时间。
```shell
kubectl exec -it demo-app -- bash
#  du -sh /data/spark/spark-3.0.0-bin-without-hadoop.tgz
150M	/data/spark/spark-3.0.0-bin-without-hadoop.tgz
# time cp /data/spark/spark-3.0.0-bin-without-hadoop.tgz /dev/null
real	0m13.171s
user	0m0.002s
sys	0m0.028s
```

5. 为了避免其他因素(比如 page cache )对结果造成影响，我们将删除之前的容器，新建相同的应用，尝试访问同样的文件。由于此时文件已经被 *alluxio* 缓存，可以看到第二次访问所需时间远小于第一次。
```shell
kubectl delete -f app.yaml && kubectl create -f app.yaml
...
# time cp /data/spark/spark-3.0.0-bin-without-hadoop.tgz /dev/null
real	0m0.344s
user	0m0.002s
sys	0m0.020s
```

到这里，我们已经成功创建了一个数据集并完成了加速，关于数据集的进一步使用和管理可以参考[accelerate](
../samples/accelerate_data_accessing.md)和[co-locality](../samples/data_co_locality.md)这两个例子。