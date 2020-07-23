# Fluid开发文档

## 环境需求
- golang 1.13+
- docker 19.03+
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

## 本地运行与调试
在实际部署之前，您可能希望能够在本地K8S上测试组件是否运转正常

使用go-delve进行程序调试：
```shell script
make debug

make debug-csi
```

或者基于`$HOME/.kube/config`中的默认K8S集群配置快速运行以观察组件行为：
```shell script
make run
```