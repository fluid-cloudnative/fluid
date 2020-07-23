# 从源码构建Fluid

## 环境需求
- golang 1.14
- docker 19.03.11
- GNU Make 4.1
- *TODO 测试最低支持版本？*

## 下载源码到本地
```shell script
mkdir -p $GOPATH/src/fluid
cd $GOPATH/src/fluid
git clone https://github.com/cheyang/fluid.git
```

## 构建
Fluid项目目录下已经为用户提供了基本的构建逻辑`./Makefile`
```shell script
# 构建Controller Manager Binary
make manager
# 构建CSI Binary
make csi
```
构建得到的Binary程序位于`./bin`目录下


如果您希望生成对应的Docker镜像：
```shell script
# 构建Controller Manager Docker Image
make docker-build
# 构建CSI Docker Image
make docker-build-csi
```

- TODO: 考虑GO111MODULE=on/off的不同情况（目前Makefile中使用GO111MODULE=off构建）


## demo
- todo