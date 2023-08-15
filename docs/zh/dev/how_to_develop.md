# Fluid开发文档

## 环境需求

- Git
- Golang (version >= 1.16)
- Docker (version >= 19.03)
- Kubernetes (version >= 1.18)
- GNU Make

对于Golang的安装与配置，请参考[此处](https://golang.org/dl/)。

对于Docker的安装与配置，请参考[此处](https://docs.docker.com/engine/install/)。

Fluid需要使用`make`命令进行项目构建，使用以下命令安装`make`：

- Linux
  - `sudo apt-get install build-essential`

## 项目构建

### 获取项目源码
```
$ export GOPATH=$(go env GOPATH)

$ mkdir $GOPATH/src/github.com/fluid-cloudnative

$ cd $GOPATH/src/github.com/fluid-cloudnative 

$ git clone https://github.com/fluid-cloudnative/fluid.git

$ cd fluid
```

> **注意**：本文在非Go Module模式下完成Fluid的编译、运行和调试。
>
> 有关Go module可以参阅 [Golang 官方文档](https://github.com/golang/go/wiki/Modules) 获取更多信息。

### 安装`controller-gen`

首先，运行以下命令下载Fluid项目所需的代码生成工具`controller-gen`。

```shell
$ make controller-gen
```

检查`controller-gen`是否成功安装：
```
$ controller-gen --version 
Version: v0.8.0
```
> **注意**: controller-gen默认将下载到`$GOPATH/bin`路径下，请确保`$GOPATH/bin`已被添加在开发环境的`$PATH`环境变量中

### 二进制程序编译

Fluid项目根目录下的`Makefile`文件已经包含了项目开发中的编译、构建、部署等基本逻辑

```shell
# 构建Fluid各控制器组件、Fluid Webhook组件和Fluid CSI插件二进制程序
$ make build
```

```shell
# 如果只想编译一个组件，比如alluxioruntime-controller
$ make alluxioruntime-controller-build
```
构建得到的二进制程序位于Fluid项目`./bin`目录下。

### Fluid组件镜像构建&推送

1. 设置需要推送的私有镜像仓库（将以下`<docker-registry>`和`<my-repo>`替换为实际地址）

   ```shell
   export IMG_REPO=<docker-registry>/<my-repo>
   ```
   
2. 输入镜像仓库访问凭证：

   ```shell
   $ sudo docker login <docker-registry>
   ```

3. 构建全部Fluid组件镜像然后推送到仓库:

   ```shell
   $ make docker-push-all
   ```

   > 如果仅需要构建并推送特定Fluid组件的镜像（e.g. alluxio-runtime-controller镜像，fluid-webhook镜像）等，请参考`./Makefile`找到对应的Makefile target执行（e.g. `make docker-push-alluxioruntime-controller`，`make docker-push-webhook`等）

## 单元测试、运行和调试

### 单元测试

运行以下命令，执行Fluid项目单元测试：
```
make test
```

> 如果您在macOS等非linux系统开发，运行测试时若提示`exec format error`，则需要检查运行测试命令时是否设置了与开发环境不一致的`GOOS`环境变量，可通过`go env -w GOOS=darwin`进行覆盖。

### 本地运行Fluid控制器组件

Fluid控制器组件支持本地运行或调试。Fluid控制器组件包括Dataset Controller、各Runtime Controller以及Application Controller。在本地运行控制器组件前，需要在本地环境提前配置kubeconfig（通过`KUBECONFIG`环境变量配置或通过`$HOME/.kube/config`文件配置），并能正常访问一个Kubernetes集群。

> Fluid Webhook组件或Fluid CSI插件无法在本地运行与Kubernetes集群交互。运行此类组件需要首先进行镜像构建，手动替换`charts/fluid/fluid/values.yaml`的对应镜像地址后，使用helm部署到Kubernetes集群后运行。

Runtime Controller依赖于[`helm`](https://helm.sh/)以及相关Helm Charts以正常工作。因此在运行Runtime Controller前，执行如下命令进行环境配置：

1. 在本地创建helm命令的软链接程序
```
$ ln -s $(which helm) /usr/local/bin/ddc-helm
```

2. 在本地创建相关Charts的软链接目录
```
$ ln -s $GOPATH/src/github.com/fluid-cloudnative/fluid/charts $HOME/charts
```

3. 以Alluxio Runtime Controller为例，使用以下命令在本地运行该组件：
```
# 配置AlluxioRuntime相关环境变量参数
$ export ALLUXIO_RUNTIME_IMAGE_ENV="alluxio/alluxio-dev:2.9.0"
$ export ALLUXIO_FUSE_IMAGE_ENV="alluxio/alluxio-dev:2.9.0
$ export DEFAULT_INIT_IMAGE_ENV="fluidcloudnative/init-users:v0.8.0-5bb4677"
$ export MOUNT_ROOT="/runtime-mnt"
$ export HOME="$HOME"

# 打开development调试模式，打开leader-election，启动alluxioruntime-controller
$ ./bin/alluxioruntime-controller start --development=true --enable-leader-election
```

### 调试Fluid组件

Fluid控制器组件支持本地运行或调试。Fluid控制器组件包括Dataset Controller、各Runtime Controller以及Application Controller。在本地运行控制器组件前，需要在本地环境提前配置kubeconfig（通过`KUBECONFIG`环境变量配置或通过`$HOME/.kube/config`文件配置），并能正常访问一个Kubernetes集群。

> Fluid CSI插件无法在本地运行与Kubernetes集群交互。调试此类组件需要首先进行镜像构建，手动替换`charts/fluid/fluid/values.yaml`的对应镜像地址后，使用helm部署到Kubernetes集群后运行，并通过dlv远程调试进行此类组件的调试。
#### 本地命令行调试

确保环境中已经安装了go-delve，具体安装过程可以参考[go-delve安装手册](https://github.com/go-delve/delve/tree/master/Documentation/installation)

```shell
$ dlv debug cmd/alluxio/main.go
```

#### 本地VSCode调试
如果使用VSCode作为开发环境，可直接安装VSCode的[Go插件](https://marketplace.visualstudio.com/items?itemName=golang.go)并进行本地调试。

##### 调试控制器组件
以调试Alluxio Runtime Controller为例，可在`./.vscode/launch.json`中定义如下Go代码调试任务：

```json
{
    "version": "0.2.0",
    "configurations": [
       {
            "name": "Alluxio Runtime Controller",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "cmd/alluxio/main.go",
            "args": ["start", "--development=true", "--enable-leader-election"],
            "env": {
                "KUBECONFIG": "<path>/<to>/<kubeconfig>",
                "ALLUXIO_RUNTIME_IMAGE_ENV": "alluxio/alluxio-dev:2.9.0",
                "ALLUXIO_FUSE_IMAGE_ENV": "alluxio/alluxio-dev:2.9.0",
                "DEFAULT_INIT_IMAGE_ENV": "fluidcloudnative/init-users:v0.8.0-5bb4677",
                "MOUNT_ROOT": "/runtime-mnt",
                "HOME": "<HOME_PATH>"
            }
        },
    ]
}
```
##### 调试WebHook组件
将集群中对WebHook的访问代理到本机：

```shell
# 1. 安装kt-connect（https://github.com/alibaba/kt-connect）
$ curl -OL https://github.com/alibaba/kt-connect/releases/download/v0.3.7/ktctl_0.3.7_Linux_x86_64.tar.gz
$ tar zxf ktctl_0.3.7_Linux_x86_64.tar.gz
$ mv ktctl /usr/local/bin/ktctl
$ ktctl --version

# 2. 将对Webhook的访问代理到本机
$ ktctl exchange fluid-pod-admission-webhook  --kubeconfig <path>/<to>/<kubeconfig> --namespace fluid-system --expose 9443 
```
在`./.vscode/launch.json`中定义如下Go代码调试任务：
```json
{
    "version": "0.2.0",
    "configurations": [
       {
            "name": "Fluid Webhook",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "cmd/webhook/main.go",
            "args": ["start", "--development=true", "--full-go-profile=false", "--pprof-addr=:6060"],
            "env": {
                "TIME_TRACK_DEBUG": "true",
                "MY_POD_NAMESPACE": "fluid-system"
            }
        },
    ]
}
```


#### 远程调试

针对Fluid Webhook和Fluid CSI插件等组件，通常情况下更为常用的方式是远程调试，确保本机和组件镜像中均已正确安装了go-delve

在远程主机上:

```shell
$ dlv debug --headless --listen ":12345" --log --api-version=2 cmd/alluxio/main.go
```

这将使得远程主机的调试程序监听指定的端口(e.g. 12345)

在本机上:

```shell
$ dlv connect "<remote-addr>:12345" --api-version=2
```

> 注意：要进行远程调试，请确保远程主机指定的端口未被占用并且已经对远程主机的防火墙进行了适当的配置
