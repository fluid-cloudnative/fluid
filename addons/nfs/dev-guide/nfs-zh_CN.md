# NFS 接入 ThinRuntime 的简单示例

## 前期准备
NFS 接入 ThinRuntime 需要构造 NFS-FUSE 客户端，本实例使用 [该项目](https://github.com/sahlberg/fuse-nfs) 为基础构建 FUSE 镜像。

## 准备 NFS-FUSE 客户端镜像
1. 挂载参数解析脚本

在 FUSE 容器内需要提取相关的 **ThinRuntimeProfile、Dataset、ThinRuntime**资源中对远程文件系统的配置信息，相关信息以 JSON 字符串的方式保存到 FUSE 容器的 **/etc/fluid/config.json** 文件内。


```python
# fluid_config_init.py
import json

rawStr = ""
with open("/etc/fluid/config/config.json", "r") as f:
    rawStr = f.readlines()

rawStr = rawStr[0]

# 挂载脚本
script = """
#!/bin/sh
set -ex
MNT_FROM=$mountPoint
MNT_TO=$targetPath


trap "fusermount -u ${MNT_TO}" SIGTERM
mkdir -p ${MNT_TO}
fuse-nfs -n nfs://${MNT_FROM} -m ${MNT_TO}
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
mountPoint="xx.xx.xx.xx/xxx/nfs"
targetPath="/runtime-mnt/thin/default/my-storage/thin-fuse"

#!/bin/sh
set -ex
MNT_FROM=$mountPoint
MNT_TO=$targetPath


trap "fusermount -u ${MNT_TO}" SIGTERM
mkdir -p ${MNT_TO}
fuse-nfs -n nfs://${MNT_FROM} ${MNT_TO}

sleep inf
```
该 shell 脚本创建挂载的文件夹并将远程文件系统挂载到目标位置（targetPath）。**由于 mount 命令会立即返回，为了保持该进程的持续运行（防止FUSE pod 反复重启），需要 sleep inf 来保持进程的存在。同时为了在 FUSE pod 被删除前，将挂载的远程存储系统卸载，需要捕捉 SIGTERM 信号执行卸载命令**。

3. 创建 FUSE 客户端镜像


将参数解析脚本、挂载脚本和相关的库打包入镜像。

```dockerfile
# Build environment
FROM ubuntu:jammy as BUILD
RUN apt update && \
    apt install --yes libfuse-dev libnfs13 libnfs-dev libtool m4 automake libnfs-dev xsltproc make libtool


COPY ./fuse-nfs-master /src
WORKDIR /src
RUN ./setup.sh && \
    ./configure && \
    make

# Production image
FROM ubuntu:jammy
RUN apt update && \
    apt install --yes libnfs13 libfuse2 fuse python3 bash && \
    apt clean autoclean && \
    apt autoremove --yes && \
    rm -rf /var/lib/{apt,dpkg,cache,log}/
ADD ./fluid_config_init.py /
ADD ./entrypoint.sh /usr/local/bin
COPY --from=BUILD /src/fuse/fuse-nfs /bin/fuse-nfs
CMD ["/usr/local/bin/entrypoint.sh"]
```
用户需要首先下载 [该项目](https://github.com/sahlberg/fuse-nfs) 到本地并解压到 **<PATH>/fuse-nfs-master，然后将上述 Dockerfile 、参数解析脚本 fluid_config_init.py和启动脚本 entrypoint.sh 复制到项目下父目录下**。
```shell
$ ls                                      
Dockerfile  entrypoint.sh  fluid_config_init.py  fuse-nfs-master
```

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
      - /bin/sh
      - -c
      - "chmod +x /usr/local/bin/entrypoint.sh && /usr/local/bin/entrypoint.sh"
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
  - mountPoint: <IP/PATH>
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
将上述 mountPoint(HOST_IP/PATH) 修改为您需要使用的远程 NFS 的地址。

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