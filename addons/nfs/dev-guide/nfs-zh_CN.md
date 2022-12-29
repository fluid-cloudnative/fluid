# NFS 接入 ThinRuntime 的简单示例

## 前期准备
首先需要在 K8s 集群能够访问到的机器上部署 [NFS 服务端](https://nfs.sourceforge.net/) 并配置读写权限，并且保证在 K8s 集群的节点上能够访问到该 NFS 服务。

## 准备 NFS-FUSE 客户端镜像
1. 挂载参数解析脚本

在 FUSE 容器内需要提取相关的 **ThinRuntimeProfile、Dataset、ThinRuntime**资源中对远程文件系统的配置信息，相关信息以 JSON 字符串的方式保存到 FUSE 容器的 **/etc/fluid/config.json** 文件内。


```python
# fluid_config_init.py
import json

rawStr = ""
with open("/etc/fluid/config.json", "r") as f:
    rawStr = f.readlines()

rawStr = rawStr[0]

# 挂载脚本
script = """
#!/bin/sh
set -ex
MNT_FROM=$mountPoint
MNT_TO=$targetPath


trap "umount ${MNT_TO}" SIGTERM
mkdir -p ${MNT_TO}
mount -t nfs ${MNT_FROM} ${MNT_TO}
sleep inf
"""

obj = json.loads(rawStr)


with open("mount-nfs.sh", "w") as f:
    f.write("mountPoint=\"%s\"\n" % obj['mounts'][0]['mountPoint'])
    f.write("targetPath=\"%s\"\n" % obj['targetPath'])

    f.write(script)

```
该 Python 脚本，将参数提取后以变量的方式注入 shell 脚本。

2. 挂载远程文件系统脚本

在将参数解析并注入shell脚本后，生成的脚本如下
```shell
mountPoint="xx.xx.xx.xx:/xxx/nfs"
targetPath="/runtime-mnt/thin/default/my-storage/thin-fuse"

#!/bin/sh
set -ex
MNT_FROM=$mountPoint
MNT_TO=$targetPath


trap "umount ${MNT_TO}" SIGTERM
mkdir -p ${MNT_TO}
mount -t nfs ${MNT_FROM} ${MNT_TO}

sleep inf
```
该 shell 脚本创建挂载的文件夹并将远程文件系统挂载到目标位置（targetPath）。**由于 mount 命令会立即返回，为了保持该进程的持续运行（防止FUSE pod 反复重启），需要 sleep inf 来保持进程的存在。同时为了在 FUSE pod 被删除前，将挂载的远程存储系统卸载，需要捕捉 SIGTERM 信号执行卸载命令**。

3. 创建 FUSE 客户端镜像


将参数解析脚本、挂载脚本和相关的库打包入镜像。

```dockerfile
FROM alpine
RUN apk add python3 bash nfs-utils
ADD ./fluid_config_init.py /
```

除了 Python 脚本外，还需要在基镜像上**安装 python 环境和 nfs-utils NFS 客户端**。

## 使用示例
### 创建并部署 ThinRuntimeProfile 资源
```shell
$ cat <<EOF > profile.yaml
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntimeProfile
metadata:
  name: nfs-profile
spec:
  fileSystemType: nfs
  
  fuse:
    image: <IMG_REPO>
    imageTag: <IMG_TAG>
    imagePullPolicy: IfNotPresent
    command:
      - sh
      - -c
      - "python3 /fluid_config_init.py && chmod u+x /mount-nfs.sh && /mount-nfs.sh"
EOF

$ kubectl apply -f profile.yaml
```
将上述 <IMG_REPO> 改为您制作的镜像的仓库名称，<IMG_TAG> 修改为该镜像的 TAG。
### 创建并部署 Dataset 和 ThinRuntime 资源
```shell
$ cat <<EOF > data.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: nfs-demo
spec:
  mounts:
  - mountPoint: <IP:PATH>
    name: nfs-demo
---
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntime
metadata:
  name: nfs-demo
spec:
  profileName: nfs-profile
EOF

$ kubectl apply -f data.yaml
```
将上述 mountPoint 修改为您需要使用的远程 NFS 的地址。

### 部署应用
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
      volumeMounts:
        - mountPath: /data
          name: nfs-demo
  volumes:
    - name: nfs-demo
      persistentVolumeClaim:
        claimName: nfs-demo
EOF

$ kubectl apply -f app.yaml
```
在将使用该远程文件系统的应用部署后，相应的 FUSE pod 也被调度到相同节点。

```shell
$ kubectl get pods
NAME                  READY   STATUS    RESTARTS   AGE
nfs-demo-fuse-wx7ns   1/1     Running   0          12s
nginx                 1/1     Running   0          26s
```
远程的文件系统被挂载到 nginx pod 的 /data 目录下。