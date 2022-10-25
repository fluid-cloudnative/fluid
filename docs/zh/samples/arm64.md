# 示例 - 在ARM64平台上使用Fluid

## 环境准备

### 获取Fluid最新源码
Fluid 社区目前提供了多平台的 Fluid 官方镜像，您可以直接参考[这篇文档](../userguide/get_started.md)部署 Fluid 并且直接跳到*运行实例*步骤.
```shell
$ mkdir -p $GOPATH/src/github.com/fluid-cloudnative/
$ cd $GOPATH/src/github.com/fluid-cloudnative
$ git clone https://github.com/fluid-cloudnative/fluid.git
$ cd $GOPATH/src/github.com/fluid-cloudnative/fluid
```

### 构建amd64和arm64容器镜像

```shell
$ export DOCKER_CLI_EXPERIMENTAL=enabled
$ docker run --rm --privileged docker/binfmt:66f9012c56a8316f9244ffd7622d7c21c1f6f28d
$ docker buildx create --use --name mybuilder
$ export IMG_REPO=<your docker image repo>
$ make docker-buildx-all-push
```

> 获取到构建的镜像地址和镜像版本

### 修改helm chart中镜像版本

```shell
$ cd $GOPATH/src/github.com/fluid-cloudnative/fluid/charts/fluid/fluid
$ vim values.yaml
```

### 创建基于arm64的Kubernetes集群， 并且安装helm chart

#### 查看arm节点

```shell
kubectl get no -l kubernetes.io/arch=arm64
NAME                       STATUS   ROLES    AGE    VERSION
cn-beijing.192.168.3.183   Ready    <none>   6d3h   v1.22.10
cn-beijing.192.168.3.184   Ready    <none>   6d3h   v1.22.10
cn-beijing.192.168.3.185   Ready    <none>   6d3h   v1.22.10
```

#### 安装

```
$ kubectl create ns fluid-system
$ helm install fluid fluid
NAME: fluid
LAST DEPLOYED: Sat Aug 20 21:43:27 2022
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

## 运行示例

本文 JuiceFSRuntime 为例


1. 创建基于arm64的Kubernetes集群

2. 根据[文档](juicefs_setup.md)准备JuiceFS 社区版

在使用 JuiceFS 之前，您需要提供元数据服务（如 Redis）及对象存储服务（如 MinIO）的参数，并创建对应的 secret:

```shell
kubectl create secret generic jfs-secret \
    --from-literal=metaurl=redis://redis:6379/0 \
    --from-literal=access-key=minioadmin \
    --from-literal=secret-key=minioadmin
```

其中：

- `metaurl`：元数据服务的访问 URL (比如 Redis)。更多信息参考[这篇文档](https://juicefs.com/docs/zh/community/databases_for_metadata/) 。
- `access-key`：对象存储的 access key。
- `secret-key`：对象存储的 secret key。

**查看待创建的 `Dataset` 资源对象**

```yaml
$ cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: jfsdemo
spec:
  mounts:
    - name: minio
      mountPoint: "juicefs:///"
      options:
        bucket: "http://minio:9000/minio/test"
        storage: "minio"
      encryptOptions:
        - name: metaurl
          valueFrom:
            secretKeyRef:
              name: jfs-secret
              key: metaurl
        - name: access-key
          valueFrom:
            secretKeyRef:
              name: jfs-secret
              key: access-key
        - name: secret-key
          valueFrom:
            secretKeyRef:
              name: jfs-secret
              key: secret-key
EOF
```

其中：

- `mountPoint`：指的是 JuiceFS 的子目录，是用户在 JuiceFS 文件系统中存储数据的目录，以 `juicefs://` 开头；如 `juicefs:///demo` 为 JuiceFS 文件系统的 `/demo` 子目录。
- `bucket`：Bucket URL。例如使用 S3 作为对象存储，在本例子中bucket的名字为test 。
- `storage`：对象存储类型，比如 `s3`，`gs`，`oss`。更多信息参考[这篇文档](https://juicefs.com/docs/zh/community/how_to_setup_object_storage/) 。

> **注意**：只有 `name` 和 `metaurl` 为必填项，若 JuiceFS 已经格式化过，只需要填写 `name` 和 `metaurl` 即可。

由于 JuiceFS 采用的是本地缓存，对应的 `Dataset` 只支持一个 mount，且 JuiceFS 没有 UFS，`mountPoint` 中可以指定需要挂载的子目录（`juicefs:///` 为根路径），会作为根目录挂载到容器内。

**创建 `Dataset` 资源对象**
```shell
$ kubectl create -f dataset.yaml
dataset.data.fluid.io/jfsdemo created
```

**查看 `Dataset` 资源对象状态**
```shell
$ kubectl get dataset jfsdemo
NAME      UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE      AGE
jfsdemo                                                                  NotBound   44s
```

如上所示，`status` 中的 `phase` 属性值为 `NotBound`，这意味着该 `Dataset` 资源对象目前还未与任何 `JuiceFSRuntime` 资源对象绑定，接下来，我们将创建一个 `JuiceFSRuntime` 资源对象。

**查看待创建的 `JuiceFSRuntime` 资源对象**

```yaml
$ cat<<EOF >runtime.yaml
apiVersion: data.fluid.io/v1alpha1
kind: JuiceFSRuntime
metadata:
  name: jfsdemo
spec:
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 4Gi
        low: "0.1"
EOF
```

**创建 `JuiceFSRuntime` 资源对象**

```shell
$ kubectl create -f runtime.yaml
juicefsruntime.data.fluid.io/jfsdemo created
```

**检查 `JuiceFSRuntime` 资源对象是否已经创建**
```shell
$ kubectl get juicefsruntime
NAME      AGE
jfsdemo   34s
```

等待一段时间，让 `JuiceFSRuntime` 资源对象中的各个组件得以顺利启动，你会看到类似以下状态：

```shell
$ kubectl get juicefs jfsdemo
NAME      WORKER PHASE   FUSE PHASE   AGE
jfsdemo   Ready          Ready        2m15s
```


然后，再查看 `Dataset` 状态，发现已经与 `JuiceFSRuntime` 绑定。

```shell
$ kubectl get dataset jfsdemo
NAME      UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
jfsdemo   0.00B            0.00B    4.00GiB          0.0%                Bound   3m16s
```

**查看待创建的 Pod 资源对象**，其中 Pod 使用上面创建的 `Dataset` 的方式为指定同名的 PVC。

```yaml
$ cat<<EOF >sample.yaml
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
        claimName: jfsdemo
EOF
```

**创建 Pod 资源对象**

```shell
$ kubectl create -f sample.yaml
pod/demo-app created
```

**检查 Pod 资源对象是否已经创建**
```shell
$ kubectl get po |grep demo
demo-app                                                       1/1     Running   0          31s
jfsdemo-fuse-fx7np                                             1/1     Running   0          31s
jfsdemo-worker-0                                               1/1     Running   0          10m
```

可以看到 pod 已经创建成功，同时 JuiceFS 的 FUSE 组件也启动成功。

登陆到应用pod中可以看到挂载成功，并且该环境是ARM64

```bash
root@demo-app:/# mount | grep fuse.juicefs
JuiceFS:minio on /data type fuse.juicefs (rw,relatime,user_id=0,group_id=0,default_permissions,allow_other)
root@demo-app:/# arch
aarch64
root@demo-app:/# lscpu | grep -i arm
Vendor ID:                       ARM
```
