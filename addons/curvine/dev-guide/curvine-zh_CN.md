# Curvine 接入 ThinRuntime（开发指南）

## 前置条件

- Kubernetes 1.18+
- 已安装 Fluid
- 已运行的 Curvine 集群，且有可访问的路径

## 准备 Curvine FUSE 客户端镜像

ThinRuntime 使用 Curvine 的常规运行时镜像，镜像中已包含 Fluid 集成脚本。Curvine 将其发布为 `ghcr.io/curvineio/curvine`。

1）使用已发布的镜像，或从源码构建

```bash
# 在 Curvine 仓库根目录：编译 Curvine 并打包为运行时镜像
make docker-build
```

`make docker-build` 会生成名为 `curvine:latest` 的本地镜像。如需使用该镜像而不是已发布的镜像，请将其打标签并推送到集群可以拉取的镜像仓库（或加载到节点上，例如使用 `kind load docker-image`），并将下文 ThinRuntimeProfile 中的 `spec.fuse.image` 和 `spec.fuse.imageTag` 设置为该镜像：

```bash
docker tag curvine:latest <your-registry>/curvine:<tag>
docker push <your-registry>/curvine:<tag>
```

2）镜像工作机制

- 以 `/entrypoint.sh fluid-thin-runtime` 启动（或检测到 Fluid 挂载的配置）时，转交给 `/fluid-entrypoint.sh`
- 从 /etc/fluid/config/config.json 读取 Fluid 注入的配置
- 通过 `/app/curvine/config-parse.py` 解析 Dataset/ThinRuntime/ThinRuntimeProfile
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
    image: ghcr.io/curvineio/curvine
    imageTag: latest
    imagePullPolicy: IfNotPresent
    command:
    - /entrypoint.sh
    args:
    - fluid-thin-runtime
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
- 完整示例参见 Curvine 仓库中的 `curvine-docker/fluid/thin-runtime/curvine-thinruntime.yaml`
