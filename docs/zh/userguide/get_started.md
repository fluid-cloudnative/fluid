# Fluid 快速上手    
本文档介绍了如何创建或使用 Kubernetes 集群环境，通过 Helm 完成 Fluid 安装部署，并使用 Fluid 创建数据集。  

## 前置需求

1. Kubernetes 1.14+
  
    如果你目前没有满足条件的 Kubernetes 环境, 那么我们推荐你选择官方认证的 Kubernetes 云服务, 通常情况下, 你仅需寥寥几步即可快速获得一个专属的 Kubernetes 环境, 以下列出了部分经过认证的 Kubernetes 云服务:
    - [阿里云容器服务](https://www.aliyun.com/product/kubernetes)
    - [Amazon Elastic Kubernetes Service](https://aws.amazon.com/eks/)
    - [Azure Kubernetes Service](https://docs.microsoft.com/en-us/azure/aks/tutorial-kubernetes-deploy-cluster)
    - [Google Kubernetes Engine](https://cloud.google.com/kubernetes-engine/)

    > 注意: 考虑到 Minikube 功能的局限性,我们不推荐使用 Minikube 进行接下来的步骤

2. Kubectl 1.14+

    请确保Kubectl已经正确配置使其能够与你的Kubernetes环境进行交互

3. [Helm 3](https://helm.sh/docs/intro/install/)

    在接下来的步骤中, 将使用Helm 3进行 Fluid 的快速安装


## 安装Fluid
1. 创建命名空间
    ```shell
    $ kubectl create ns fluid-system
    ```  
2. 从 Github 仓库[Release页面](https://github.com/fluid-cloudnative/fluid/releases)下载最新版本的Fluid
    
3. 使用 Helm 安装 Fluid
    ```shell
    $ helm install fluid fluid-<version>.tgz
    NAME: fluid
    LAST DEPLOYED: Tue Jul  7 11:22:07 2020
    NAMESPACE: default
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None
    ```

4. 查看Fluid的运行状态
    ```shell
    $ kubectl get po -n fluid-system
    NAME                                         READY   STATUS    RESTARTS   AGE
    alluxioruntime-controller-64948b68c9-zzsx2   1/1     Running   0          108s
    csi-nodeplugin-fluid-2mfcr                   2/2     Running   0          108s
    csi-nodeplugin-fluid-l7lv6                   2/2     Running   0          108s
    dataset-controller-5465c4bbf9-5ds5p          1/1     Running   0          108s
    ```

## 创建dataset
Fluid提供了云原生的数据加速和管理能力，并抽象出了`数据集(Dataset)`概念方便用户管理，接下来将演示如何用 Fluid 创建一个数据集。   

1. 创建一个Dataset CRD对象，其中描述了数据集的来源。
    ```shell 
    $ cat<<EOF >dataset.yaml
    apiVersion: data.fluid.io/v1alpha1
    kind: Dataset
    metadata:
      name: demo
    spec:
      mounts:
        - mountPoint: https://mirrors.bit.edu.cn/apache/spark/
          name: spark
    EOF
    ```  
    执行安装
    
    ```
    $ kubectl create -f dataset.yaml
    ```

2. 创建 `AlluxioRuntime` CRD对象，用来描述支持这个数据集的 Runtime, 在这里我们使用[Alluxio](https://www.alluxio.io/)作为其Runtime
    ```yaml
    $ cat<<EOF >runtime.yaml
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
    EOF
    ```
    使用`kubectl`完成创建  
    
    ```shell
    $ kubectl create -f runtime.yaml  
    ``` 

3. 接下来，我们创建一个应用容器来使用该数据集，我们将多次访问同一数据，并比较访问时间来展示 Fluid 的加速效果。
    ```shell
    $ cat<<EOF >app.yaml
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
    EOF
    ```
    使用`kubectl`完成创建  

    ```shell
    $ kubectl create -f app.yaml  
    ``` 

4. 登录到应用容器中访问数据，初次访问会花费更长时间。
    ```shell
    $ kubectl exec -it demo-app -- bash
    $ du -sh /data/spark/spark-3.0.1/spark-3.0.1-bin-without-hadoop.tgz
    150M	/data/spark/spark-3.0.1/spark-3.0.1-bin-without-hadoop.tgz
    $ time cp /data/spark/spark-3.0.1/spark-3.0.1-bin-without-hadoop.tgz /dev/null
    real	0m13.171s
    user	0m0.002s
    sys	0m0.028s
    ```

5. 为了避免其他因素(比如 page cache )对结果造成影响，我们将删除之前的容器，新建相同的应用，尝试访问同样的文件。由于此时文件已经被 `Alluxio` 缓存，可以看到第二次访问所需时间远小于第一次。
    ```shell
    $ kubectl delete -f app.yaml && kubectl create -f app.yaml
    $ kubectl exec -it demo-app -- bash
    $ time cp /data/spark/spark-3.0.1/spark-3.0.1-bin-without-hadoop.tgz /dev/null
    real	0m0.034s
    user	0m0.001s
    sys	0m0.032s
    ```

到这里，我们简单地创建了一个数据集并实现了数据集的抽象管理与加速, 更多有关 Fluid 的更详细的信息, 请参考以下示例文档:
- [远程文件访问加速](../samples/accelerate_data_accessing.md)
- [数据缓存亲和性调度](../samples/data_co_locality.md)
- [用Fluid加速机器学习训练](../samples/machinelearning.md)
