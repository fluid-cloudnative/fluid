# 3FS 接入 ThinRuntime 的简单示例

## 前期准备

根据dev-guide里的Dockerfile制作3fs builder镜像为后续3fs binary的编译做准备， 在容器中编译的具体原因请[参照] (https://mp.weixin.qq.com/s/pGct2Yy-lO-mG9mT0jJ-Sw?click_id=1)。


## 准备 3FS-FUSE 客户端镜像
1. 挂载参数解析脚本

在 FUSE 容器内需要提取相关的 **ThinRuntimeProfile、Dataset、ThinRuntime**资源中对远程文件系统的配置信息，相关信息以 JSON 字符串的方式保存到 FUSE 容器的 **/etc/fluid/config.json** 文件内。


```python
# fluid_config_init.py
#!/usr/bin/env python

import json

rawStr = ""
try:
    with open("/etc/fluid/config/config.json", "r") as f:
        rawStr = f.readlines()
except:
    pass

if rawStr == "":
    try:
        with open("/etc/fluid/config.json", "r") as f:
            rawStr = f.readlines()
    except:
        pass

rawStr = rawStr[0]

# 挂载脚本
script = """
#!/bin/sh
set -ex
MNT_FROM=$mountPoint
TOKEN=$(echo $MNT_FROM | awk -F'@' '{print $1}')
RDMA=$(echo $MNT_FROM | awk -F'@' '{print $2}' | awk -F'://' '{print $2}')

echo $TOKEN > /opt/3fs/etc/token.txt

sed -i "s#RDMA://0.0.0.0:8000#${RDMA}#g" /opt/3fs/etc/hf3fs_fuse_main_launcher.toml

MNT_TO=$targetPath
trap "umount ${MNT_TO}" SIGTERM
mkdir -p ${MNT_TO}

sed -i "s#/3fs/stage#${MNT_TO}#g" /opt/3fs/etc/hf3fs_fuse_main_launcher.toml

LD_LIBRARY_PATH=/opt/3fs/bin /opt/3fs/bin/hf3fs_fuse_main --launcher_cfg /opt/3fs/etc/hf3fs_fuse_main_launcher.toml
"""

obj = json.loads(rawStr)

with open("/mount-3fs.sh", "w") as f:
    f.write('mountPoint="%s"\n' % obj["mounts"][0]["mountPoint"])
    f.write('targetPath="%s"\n' % obj["targetPath"])
    f.write(script)
```
该Python 脚本，将参数提取后以变量的方式注入 shell 脚本。

2. 复制docker文件夹下的目录结构设置所有相关配置和脚本 


3. 编译hf3fs_fuse_main 
```shell
docker run -it --rm -v $(pwd):/app shaowenchen/demo-3fsbuilder:latest bash
git clone https://github.com/deepseek-ai/3FS
cd 3FS
git submodule update --init --recursive
./patches/apply.sh
cmake -S . -B build -DCMAKE_CXX_COMPILER=clang++-14 -DCMAKE_C_COMPILER=clang-14 -DCMAKE_BUILD_TYPE=RelWithDebInfo -DCMAKE_EXPORT_COMPILE_COMMANDS=ON
cmake --build build -j 100
```

将参数解析脚本、挂载脚本和相关的库打包入镜像。

```dockerfile
# Build environment
FROM shaowenchen/demo-3fsbuilder:latest
RUN apt-get install -y python3
COPY bin /opt/3fs/bin
RUN chmod +x /opt/3fs/bin/*  && mkdir -p /var/log/3fs
COPY etc /opt/3fs/etc
COPY ./fluid_config_init.py /
COPY ./entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/entrypoint.sh
ENTRYPOINT []
```
用户需要首先下载 [该项目](https://github.com/sahlberg/fuse-nfs) 到本地并解压到 **<PATH>/fuse-nfs-master，然后将上述 Dockerfile 、参数解析脚本 fluid_config_init.py和binary 启动脚本 entrypoint.sh 复制到项目下父目录下**。
```shell
$ ls                                      
Dockerfile build.sh  entrypoint.sh  fluid_config_init.py  etc
```

## 使用示例
### 创建并部署 ThinRuntimeProfile 资源
```shell
$ cat <<EOF > profile.yaml
apiVersion:data.fluid.io/v1alpha1
kind:ThinRuntimeProfile
metadata:
name:3fs
spec:
fileSystemType:3fs
fuse:
    image:shaowenchen/demo-fluid-3fs
    imageTag:latest
    imagePullPolicy:Always
    command:
      -"/usr/local/bin/entrypoint.sh"
EOF

$ kubectl apply -f profile.yaml
```
将上述 <IMG_REPO> 改为您制作的镜像的仓库名称，<IMG_TAG> 修改为该镜像的 TAG。
### 创建并部署 Dataset 和 ThinRuntime 资源
```shell
$ cat <<EOF > data.yaml
apiVersion:data.fluid.io/v1alpha1
kind:Dataset
metadata:
name:demo-3fs
spec:
mounts:
-mountPoint:my3fsTOKEN@RDMA://x.x.x.x:8000
    name:demo-3fs
---
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntime
metadata:
  name: demo-3fs
spec:
  profileName: 3fs
EOF

$ kubectl apply -f data.yaml
```
这里的 mountPoint 是由 token 和 RDMA 地址组成的，token 是 3FS 的认证信息，RDMA 地址是 3FS 的服务地址。
### 部署应用
```shell
kubectl apply-f-<<EOF
apiVersion:v1
kind:Pod
metadata:
name:demo-3fs
spec:
containers:
    -name:demo-3fs
      image:shaowenchen/demo-ubuntu
      volumeMounts:
        -mountPath:/data
          name:demo-3fs
volumes:
    -name:demo-3fs
      persistentVolumeClaim:
        claimName:demo-3fs
EOF
```
在将使用该远程文件系统的应用部署后，相应的 FUSE pod 也被调度到相同节点。


