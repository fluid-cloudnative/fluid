# Fluid开发文档

## 环境需求
- [golang 1.13+](https://golang.org/dl/)
- [docker 19.03+](https://docs.docker.com/engine/install/)
- GNU Make 4.1+

## 下载源码到本地

如果您习惯使用GOPATH进行项目开发(`GO111MODULE="off"`)：
```shell script
mkdir -p $GOPATH/src/github.com/cloudnativefluid/
cd $GOPATH/src/github.com/cloudnativefluid
git clone https://github.com/cheyang/fluid.git
```

如果您习惯使用Go Module进行项目开发(`GO111MODULE="on"`)：
```shell script
cd <any-place-you-like>
git clone https://github.com/cheyang/fluid.git
```

## 编译&镜像构建
Fluid项目根目录下的`Makefile`文件已经包含了项目开发中的编译、构建、部署等基本逻辑
```shell script
# 构建Controller Manager Binary
make manager
# 构建CSI Binary
make csi
```
构建得到的Binary程序位于`./bin`目录下

>**注意：如果您正在使用Go Module进行项目开发，那么可能需要将Makefile文件中的相关目标的`GO111MODULE=off`修改为`GO111MODULE=on`以使得编译成功**

构建Docker镜像：
```shell script
# 构建Controller Manager Docker Image
make docker-build
# 构建CSI Docker Image
make docker-build-csi
```

或者一键完成镜像构建和Push操作：
```shell script
make docker-push

make docker-push-csi
```

## 运行
接下来的内容将假设您在本地环境中已经通过`KUBECONFIG`环境变量或是在`~/.kube/config`文件中配置好了可以访问的Kubernetes集群，您可以通过`kubectl cluster-info`对该配置进行快速检查

**步骤**

0.将构建得到的镜像上传到Kubernetes集群可以访问的镜像仓库，并修改`config/fluid/patches`中各文件的镜像名

```yaml
# config/fluid/patches/image_in_manager.yaml
...
...
containers:
  - name: manager
    image: <your-manager-image>
```

```yaml
# config/fluid/patches/image_in_csi-plugin.yaml
...
...
containers:
  - name: plugins
    image: <your-csi-plugin-image>
```

1.创建CRD
```shell script
kubectl apply -k config/crd
```

2.创建Fluid各组件
```shell script
kubectl apply -k config/fluid
```

3.编写样例或使用我们提供的样例
```shell script
kubectl apply -k config/samples
```

4.查看各组件的运行情况,确保各组件和样例资源正常运行
```shell script
$ kubectl get pod -n fluid-system
NAME                                  READY   STATUS    RESTARTS   AGE
controller-manager-7fd6457ccf-p7j2x   1/1     Running   0          84s
csi-nodeplugin-fluid-pj9tv            2/2     Running   0          84s
csi-nodeplugin-fluid-t8ctj            2/2     Running   0          84s
```
```shell script
$ kubectl get pod
NAME                   READY   STATUS    RESTARTS   AGE
cifar10-fuse-vb6l4     1/1     Running   0          6m15s
cifar10-fuse-vtqpx     1/1     Running   0          6m15s
cifar10-master-0       2/2     Running   0          8m24s
cifar10-worker-729xz   2/2     Running   0          6m15s
cifar10-worker-d6kmd   2/2     Running   0          6m15s
nginx-0                1/1     Running   0          8m30s
nginx-1                1/1     Running   0          8m30s
```
**注意**: 上述命令可能随您组件的不同实现或是不同的样例产生不同的结果

5.通过日志等方法查看您的组件是否运作正常(e.g. `kubectl logs -n fluid-system controller-manager`)

6.环境清理
```shell script
kubectl delete -k config/samples

kubectl delete -k config/fluid

kubectl delete -k config/crd
```