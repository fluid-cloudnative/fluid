# 从私有仓库拉取镜像

如果 fluid 镜像存储在私有镜像仓库, 部署 fluid 需要设置镜像拉取密钥  
关于镜像拉取密钥,请参考[从私有仓库拉取镜像](https://kubernetes.io/zh-cn/docs/tasks/configure-pod-container/pull-image-private-registry/)

使用 helm charts 部署 fluid 支持设置镜像拉取密钥

通过在 `values.yaml` 设置密钥, 如下
```yaml
# fluid helm charts values.yaml 
# imagePullSecrets 默认值是空
image:
  imagePullSecrets: []

# 设置密钥
# 假设有两个密钥 `test-1` 和 `test-2`
image:
  imagePullSecrets: 
  - name: test-1
  - name: test-2
```

通过 `values.yaml` 设置密钥, 完成 fluid 服务部署以后,  
可以看到在 controller 里面以后包含我们前面设置的密钥 

alluxio controller yaml 截取信息
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: alluxioruntime-controller
  namespace: fluid-system
spec:
  template:
    spec:
      containers:
      - image: fluidcloudnative/alluxioruntime-controller:
        name: manager
      dnsPolicy: ClusterFirst
      imagePullSecrets:
      - name: test-1
      - name: test-2
```


如果你想指定runtime镜像拉取密钥，可以通过编辑你已经创建的对应runtime controller对应的deployment中的环境变量。runtime controller是由Deployment资源创建和管理的，它主要用来管理Kubernetes应用的生命周期，以确保指定数量的应用副本在集群中始终正常运行。

这是一个短小的YAML配置示例，正确配置后deployment将作用于你指定的image pull secret。

```yaml
- name: IMAGE_PULL_SECRETS
  value: test-1,test-2
```

> 在上述YAML文件中， IMAGE_PULL_SECRETS 是环境变量名，test-1,test-2 是你想要指定的镜像拉取密钥，并且以逗号分隔。


同时 fluid 也支持 controller 在拉起对应 runtime 服务时, 使用前面配置的镜像拉取密钥   
controller 会传递密钥到 runtime 的服务中

alluxio runtime master yaml 截取信息
```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: demo-master
  namespace: test
spec:
  template:
    spec:
      containers:
      - image: fluidcloudnative/alluxio:release-2.8.1-SNAPSHOT-0433ade
        imagePullPolicy: IfNotPresent
        name: alluxio-master
      imagePullSecrets:
      - name: test1
      - name: test2

```

综上, 镜像拉取密钥设置成功