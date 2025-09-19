# Curvine

该插件通过 Fluid 的 ThinRuntime 将 [Curvine](https://github.com/CurvineIO/curvine) 集成到 Kubernetes 集群中。

## 安装

应用本目录下的 `runtime-profile.yaml`：

```shell
kubectl apply -f runtime-profile.yaml
```

## 使用

### 前置条件
- 已部署并可访问的 Curvine 集群
- 集群中已安装 Fluid

### 创建并部署 ThinRuntimeProfile 资源

```shell
$ cat <<EOF > runtime-profile.yaml
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntimeProfile
metadata:
  name: curvine
spec:
  fileSystemType: fuse
  fuse:
    image: fluid-cloudnative/curvine-thinruntime
    imageTag: v1.0.0
    imagePullPolicy: IfNotPresent
EOF

$ kubectl apply -f runtime-profile.yaml
```

### 创建并部署 Dataset 和 ThinRuntime 资源

```shell
$ cat <<EOF > dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: curvine-demo
spec:
  mounts:
    - mountPoint: curvine:///data
      name: curvine
      options:
        master-endpoints: "<CURVINE_MASTER_IP:PORT>"
---
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntime
metadata:
  name: curvine-demo
spec:
  profileName: curvine
EOF

$ kubectl apply -f dataset.yaml
```
将 `master-endpoints` 修改为 Curvine Master 地址，`mountPoint` 修改为需要挂载的 Curvine 路径。

### 数据访问应用示例

```shell
$ cat <<EOF > app.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
    - name: nginx
      image: nginx
      command: ["bash"]
      args:
        - -c
        - sleep 9999
      volumeMounts:
        - mountPath: /data
          name: data-vol
  volumes:
    - name: data-vol
      persistentVolumeClaim:
        claimName: curvine-demo
EOF

$ kubectl apply -f app.yaml
```

应用部署后，Fluid 会根据 Profile 自动创建 FUSE Pod 并调度到同一节点，Curvine 文件系统挂载到应用容器内的 `/data`。

## 如何开发

请参考 [开发指南](./dev-guide/curvine-zh_CN.md)。

Curvine 项目地址: https://github.com/CurvineIO/curvine.git
