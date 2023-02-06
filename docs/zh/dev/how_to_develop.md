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

### 运行ruanruan


接下来的内容将假设在本地环境中已经通过`KUBECONFIG`环境变量或是在`~/.kube/config`文件中配置好了可以访问的Kubernetes集群，您可以通过`kubectl cluster-info`对该配置进行快速检查。更多有关`kubeconfig`的信息可以参考 [Kubernetes官方文档](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/)

> 以下内容将使用`kustomize`，`kubectl 1.14+`已经内置了`kustomize`工具，正在使用`kubectl 1.14`版本以下的开发者请参考 [此处](https://kustomize.io/) 获取有关kustomize的更多信息

1. 将构建的镜像上传到Kubernetes集群可以访问的镜像仓库

   > 如果构建并上传的镜像在私有仓库中，请确保在kubernetes集群的各个结点上已经成功执行了`sudo docker login <docker-registry>`操作

2. 修改`config/fluid/patches`中各文件的镜像名

   ```shell
   # config/fluid/patches/controller/injections_in_alluxioruntime_controller.yaml
   ...
   ...
   containers:
     - name: manager
       image: <your-registry>/<your-namespace>/<img-name>:<img-tag>
       env:
         - name: DEFAULT_INIT_IMAGE_ENV
   	    value: <your-registry>/<your-namespace>/<img-name>:<img-tag>
   ...
   ...
   ```

   ```shell
   # config/fluid/patches/controller/injections_in_data_controller.yaml
   ...
   ...
   containers:
     - name: manager
       image: <your-registry>/<your-namespace>/<img-name>:<img-tag>
   ...
   ...
   ```

   ```shell
   # config/fluid/patches/csi/injections_in_csi_plugin.yaml
   ...
   ...
   containers:
     # change the following two images if necessary
     - name: node-driver-registrar
       image: registry.cn-hangzhou.aliyuncs.com/acs/csi-node-driver-registrar:v1.2.0
     - name: plugins
       image: <your-registry>/<your-namespace>/<img-name>:<img-tag>
   ...
   ...
   ```

   

3. 创建CRD

   ```shell
   $ kubectl apply -k config/crd
   ```

   检查CRD：

   ```shell
   $ kubectl get crd | grep fluid
   alluxioruntimes.data.fluid.io       2020-11-28T06:20:36Z
   dataloads.data.fluid.io             2020-11-28T06:20:36Z
   datasets.data.fluid.io              2020-11-28T06:20:36Z
   ```

4. 创建Fluid各组件

   ```shell
   $ kubectl apply -k config/fluid
   ```

   检查Fluid组件：

   ```shell
   $ kubectl get pod -n fluid-system
   NAME                                         READY   STATUS    RESTARTS   AGE
   alluxioruntime-controller-5f9d4b899f-6h8xp   1/1     Running   0          8s
   csi-nodeplugin-fluid-hngkl                   2/2     Running   0          8s
   dataset-controller-6bcf4fc7b9-9rm84          1/1     Running   0          8s
   ```

5. 编写样例或使用提供的样例

   ```shell
   $ kubectl apply -k config/samples
   ```

   检查样例pod：

   ```shell
   $ kubectl get pod
   NAME                   READY   STATUS    RESTARTS   AGE
   dataset-fuse-5sz2c             1/1     Running   0          61s
   dataset-master-0               2/2     Running   0          93s
   dataset-worker-nbvrm           2/2     Running   0          61s
   et-operator-769b7864d4-glk7v   1/1     Running   0          11d
   nginx-0                        1/1     Running   0          2m3s
   ```
   
   > 注意: 上述命令可能随您组件的不同实现或是不同的样例产生不同的结果。

6. 通过日志等方法查看您的组件是否运作正常

   ```shell
   $ kubectl logs -n fluid-system <controller_manager_name>
   ```

7. 环境清理

   ```shell
   $ kubectl delete -k config/samples
   $ kubectl delete -k config/fluid
   $ kubectl delete -k config/crd
   ```

### 单元测试

#### 基本测试

在项目根目录下执行如下命令运行基本单元测试(工具类测试和engine测试)：

```shell
$ make unit-test
```

如果您需要增加接口中的方法

下载`go-mock`, 大于Go 1.16

```shell
go install github.com/golang/mock/mockgen@v1.6.0
```

生产mock代码

```shell
cd fluid
mockgen --source=pkg/ddc/base/engine.go --destination pkg/ddc/base/mock/mock_engine.go --package base
```

修改`pkg/ddc/base/template_engine_test.go`

#### 集成测试

`kubebuilder`基于[envtest](https://godoc.org/sigs.k8s.io/controller-runtime/pkg/envtest)提供了controller测试的基本框架，如果您想运行controller测试，您需要执行如下命令安装`kubebuilder`：

```shell
$ os=$(go env GOOS)
$ arch=$(go env GOARCH)
$ curl -L https://go.kubebuilder.io/dl/2.3.1/${os}/${arch} | tar -xz -C /tmp/
$ sudo mv /tmp/kubebuilder_2.3.1_${os}_${arch} /usr/local/kubebuilder
$ export PATH=$PATH:/usr/local/kubebuilder/bin
```

接下来，您可以在项目根目录下运行所有的单元测试：

```shell
$ make test
```


### 调试

**前提条件**

确保环境中已经安装了go-delve，具体安装过程可以参考[go-delve安装手册](https://github.com/go-delve/delve/tree/master/Documentation/installation)

**本地调试**

```shell
$ dlv debug cmd/alluxio/main.go
```

**远程调试** 在开发Fluid时，通常情况下更为常用的方式是远程调试，确保本机和远程主机均已正确安装了go-delve

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
