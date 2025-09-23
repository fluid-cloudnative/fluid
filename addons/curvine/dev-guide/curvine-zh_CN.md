# Curvine 接入 ThinRuntime（开发指南）

## 前置条件

- Kubernetes 1.18+
- 已安装 Fluid
- 已运行的 Curvine 集群，且有可访问的路径

## 准备 Curvine FUSE 客户端镜像

本指南使用 Curvine 提供的 ThinRuntime 构建脚本。

1）编译 Curvine

```bash
# 在 Curvine 仓库根目录
make all
# 期望产物包含 build/dist/lib/curvine-fuse
```

2）构建 ThinRuntime 镜像

```bash
# 在 Curvine 仓库
cd curvine-docker/fluid/thin-runtime
./build-image.sh
```

镜像示例：
- fluid-cloudnative/curvine-thinruntime:v1.0.0

3）镜像工作机制

- 从 /etc/fluid/config/config.json 读取 Fluid 注入的配置
- 通过 fluid-config-parse.py 解析 Dataset/ThinRuntime/ThinRuntimeProfile
- 启动 curvine-fuse 并保持前台运行

## 创建 ThinRuntimeProfile

```yaml
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
```

## 创建 Dataset 与 ThinRuntime

```yaml
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
```

## 运行应用 Pod

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
    - name: nginx
      image: nginx
      command: ["bash"]
      args: ["-c", "sleep 9999"]
      volumeMounts:
        - mountPath: /data
          name: data-vol
  volumes:
    - name: data-vol
      persistentVolumeClaim:
        claimName: curvine-demo
```

## 说明

- 确保 master-endpoints 可从 FUSE Pod 访问
- 确保镜像中存在 curvine-fuse，且 glibc 版本兼容
- 更多细节参考 Curvine 仓库中 ThinRuntime 的 README
