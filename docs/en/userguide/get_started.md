# Get Started With Fluid

This document mainly describes how to create a Kubernetes cluster environment, complete Fluid installation and deployment with Helm, and use Fluid to create a data set and speed up your application.  

## Create a Kubernetes Cluster:  
A Kubernetes environment is prerequisite for Fluid,choose the most suitable solution to get it based on your experience: 
 
- If you have already had a Kubernetes cluster, you can skip to step  [Deploy Fluid](#Deploy-Fluid).  
- If you have not used Kubernetes before, you can use Minikube to create a Kubernetes cluster.  
[Minikube](https://kubernetes.io/docs/setup/minikube/) can create a Kubernetes cluster in a virtual machine, which can run on macOS, Linux and Windows.  

Please ensure that the following requirements are met:  

  - [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) :version 1.0.0+   
  - [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl) :           version 1.14+       

After installing Minikube:
```shell
minikube start
```

If the installation is successful, you will get prompt message like this:
```shell
  minikube v1.12.1 on Darwin 10.14.5
```  

Use `kubectl` to access the newly created Kubernetes cluster  
```shell
$ kubectl get pods
NAME                                READY   STATUS    RESTARTS   AGE
nginx-deployment-558fc78868-kvjnf   1/1     Running   1          4d12h
nginx-deployment-558fc78868-kx9gt   1/1     Running   1          4d12h
```

## Deploy Fluid  
Before the installation, make sure that the following requirements have been met:

- You can access the Kubernetes cluster with `kubectl` successfully.   
- [Helm](https://helm.sh/docs/intro/install/): Helm 3 is installed.  
- [Git](): Git is installed
1. Download Fluid  
```shell
git clone https://github.com/fluid-cloudnative/fluid.git 
cd fluid/charts/fluid
```  
2. Install Fluid with Helm
```shell
helm install fluid fluid
NAME: fluid
LAST DEPLOYED: Tue Jul  7 11:22:07 2020
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
```  
3. Check installation results
```shell
kubectl get po -n fluid-system
NAME                                  READY     STATUS    RESTARTS   AGE
controller-manager-6b864dfd4f-995gm   1/1       Running   0          32h
csi-nodeplugin-fluid-c6pzj          2/2       Running   0          32h
csi-nodeplugin-fluid-wczmq          2/2       Running   0          32h
```

## Create a Dataset  
Fluid provides cloud-native data acceleration and management capabilities, and use *dataset* as a high-level abstraction to facilitate user management. Here we will show you how to create a dataset with Fluid.  
1. Create a Dataset object through the CRD file, which describes the source of the dataset.  
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
Create dataset with kubectl

```shell
kubectl create -f dataset.yaml
```
After the dataset is created, it is in the `not bound` state and needs to be bound to a runtime to use it.


2. Also we create an *Alluxio* Runtime object based on the alluxioRuntimeCRD file, which enables the dataset.

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

Create *Alluxio* Runtime with `kubectl`

```shell
kubectl create -f runtime.yaml  
``` 

3. Next, we create an application to access this dataset. Here we will access the same data multiple times and compare the time consumed by each access.

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

Create Pod with `kubectl`

```shell
kubectl create -f app.yaml
```

4. Dive into the container to access data, the first access will take longer.
```
kubectl exec -it demo-app -- bash
#  du -sh /data/spark/spark-3.0.0-bin-without-hadoop.tgz
150M	/data/spark/spark-3.0.0-bin-without-hadoop.tgz
# time cp /data/spark/spark-3.0.0-bin-without-hadoop.tgz /dev/null
real	0m13.171s
user	0m0.002s
sys	0m0.028s
```

5. In order to avoid the influence of other factors like page cache, we will delete the previous container, create the same application, and try to access the same file. Since the file has been cached by alluxio at this time, you can see that it takes significantly less time now.
```
kubectl delete -f app.yaml && kubectl create -f app.yaml
...
# time cp /data/spark/spark-3.0.0-bin-without-hadoop.tgz /dev/null
real	0m0.344s
user	0m0.002s
sys	0m0.020s
```

At this point, we have successfully created a data set and completed the acceleration. For the further use and management of the dataset, please refer to the two examples of [accelerate](../samples/accelerate_data_accessing.md) and [co-locality](../samples/data_co_locality.md).