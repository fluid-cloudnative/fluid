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