# GlusterFS 接入 ThinRuntime 的简单示例

## 前期准备
glusterfs 接入 ThinRuntime 需要构造 GlusterFS-FUSE 客户端，本实例使用 [该项目](https://github.com/gluster/gluster-containers) 为基础构建 FUSE 镜像。

## 准备 GlusterFS-FUSE 客户端镜像
1. 挂载远程文件系统脚本

在 FUSE 容器内需要提取相关的 **ThinRuntimeProfile、Dataset、ThinRuntime**资源中对远程文件系统的配置信息，相关信息以 JSON 字符串的方式保存到 FUSE 容器的 **/etc/fluid/config.json** 文件内。

```python
import json
import os
import subprocess

obj = json.load(open("/etc/fluid/config/config.json"))

mount_point = obj["mounts"][0]["mountPoint"]
target_path = obj["targetPath"]

os.makedirs(target_path, exist_ok=True)

if len(mount_point.split(":")) != 2:
    print(
        f"The mountPoint format [{mount_point}] is wrong, should be server:volumeId")
    exit(1)

server, volume_id = mount_point.split(":")
args = ["glusterfs", "--volfile-server", server, "--volfile-id",
        volume_id, target_path, "--no-daemon", "--log-file", "/dev/stdout"]

# Available options are described in the following pages:
# https://manpages.ubuntu.com/manpages/trusty/en/man8/mount.glusterfs.8.html
# https://manpages.ubuntu.com/manpages/trusty/en/man8/glusterfs.8.html
if "options" in obj["mounts"][0]:
    options = obj["mounts"][0]["options"]
    for option in options:
        if option[0] == "ro":
            option[0] = "read-only"
        elif option[0] == "transport":
            option[0] = "volfile-server-transport" 
            
        if option[1].lower() == "true":
            args.append(f'--{option[0]}')
        elif option[1].lower() == "false":
            continue
        else:
            args.append(f"--{option[0]}={option[1]}")

subprocess.run(args)
```
该 shell 脚本创建挂载的文件夹并将远程文件系统挂载到目标位置（targetPath）。

2. 创建 FUSE 客户端镜像


将脚本打包入镜像。

```dockerfile
FROM gluster/glusterfs-client@sha256:66b1d51d327ab1c4c1d81e6bad28444e13e1746c2d6f009f9874dad2fba9836e

ADD entrypoint.py /

CMD ["python3", "/entrypoint.py"]
```

## 使用示例
### 创建并部署 ThinRuntimeProfile 资源
```shell
$ cat <<EOF > profile.yaml
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntimeProfile
metadata:
  name: glusterfs
spec:
  fileSystemType: mount.glusterfs
  
  fuse:
    image: <IMG_REPO>
    imageTag: <IMG_TAG>
    imagePullPolicy: IfNotPresent
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
  name: glusterfs-demo
spec:
  mounts:
  - mountPoint: <IP:PATH>
    name: glusterfs-demo
---
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntime
metadata:
  name: glusterfs-demo
spec:
  profileName: glusterfs
EOF

$ kubectl apply -f data.yaml
```
将上述 mountPoint(HOST_IP:PATH) 修改为您需要使用的远程 glusterfs 的地址。

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
          name: glusterfs-demo
  volumes:
    - name: glusterfs-demo
      persistentVolumeClaim:
        claimName: glusterfs-demo
EOF

$ kubectl apply -f app.yaml
```
在将使用该远程文件系统的应用部署后，相应的 FUSE pod 也被调度到相同节点。

```shell
$ kubectl get pods
NAME                        READY   STATUS    RESTARTS   AGE
glusterfs-demo-fuse-wx7ns   1/1     Running   0          12s
nginx                       1/1     Running   0          26s
```
远程的文件系统被挂载到 nginx pod 的 /data 目录下。