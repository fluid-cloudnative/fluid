# CubeFS 3.2 接入 ThinRuntime 的简单示例

## 前期准备

### CubeFS集群搭建

#### 前置条件

* Kubernetes 1.14+
* CSI spec version 1.1.0
* Helm 3

#### CubeFS 集群部署

根据 [cubefs-helm](https://github.com/cubefs/cubefs-helm) 部署 CubeFS v3.2.0 。


### 使用CubeFS作为后端存储

在CubeFS集群内创建需要挂载的存储卷。

## 准备 CubeFS-FUSE 客户端镜像

1. 挂载参数解析脚本

在 FUSE 容器内需要提取相关的 **ThinRuntimeProfile、Dataset、ThinRuntime**资源中对远程文件系统的配置信息，相关信息以 JSON 字符串的方式保存到 FUSE 容器的 **/etc/fluid/config.json** 文件内。

```python
import json

rawStr = ""
with open("/etc/fluid/config.json", "r") as f:
    rawStr = f.readlines()

print(rawStr[0])

script = """
#!/bin/sh
MNT_POINT=$targetPath

echo $MNT_POINT

if test -e ${MNT_POINT}
then
    echo "MNT_POINT exist"
else
    mkdir -p ${MNT_POINT}
fi

/cfs/bin/cfs-client -c /cfs/fuse.json

sleep inf
"""

obj = json.loads(rawStr[0])
volAttrs = obj['mounts'][0]

print("pvAttrs", volAttrs)

fuse = {}
fuse["mountPoint"] = obj["targetPath"]
fuse["volName"] = volAttrs["name"]
fuse["masterAddr"] = volAttrs["mountPoint"]
fuse["owner"] = "root"
fuse["logDir"] = "/cfs/logs/"
fuse["logLevel"] = "error"

print("fuse.json: ", fuse)

with open("/cfs/fuse.json", "w") as f:
    f.write(json.dumps(fuse))

with open("mount-cubefs.sh", "w") as f:
    f.write("targetPath=\"%s\"\n" % obj['targetPath'])
    f.write(script)
```
该 Python 脚本，将参数提取后以变量的方式注入 shell 脚本。其中挂载的存储卷为CubeFS集群中root用户创建，`logDir`和`logLevel`为默认。

2. 挂载远程文件系统脚本

在将参数解析并注入shell脚本后，生成的脚本如下
```shell
targetPath="/runtime-mnt/thin/default/cubefs-test/thin-fuse"

#!/bin/sh
MNT_POINT=$targetPath

echo $MNT_POINT

if test -e ${MNT_POINT}
then
    echo "MNT_POINT exist"
else
    mkdir -p ${MNT_POINT}
fi

/cfs/bin/cfs-client -c /cfs/fuse.json

sleep inf
```
该 shell 脚本创建挂载的文件夹并将远程文件系统挂载到目标位置（targetPath）。**为了防⽌ fuse pod 反复重启，需要 sleep inf 来保持进程的存在**。


3. 创建 FUSE 客户端镜像

将参数解析脚本、挂载脚本和相关的库打包入镜像。

```dockerfile
FROM chubaofs/cfs-client:v3.2.0
ADD fluid_config_init.py /
```

由于Cubefs挂载需要使用官方提供的二进制文件 cfs-client，该文件存放于cubefs客户端镜像(chubaofs/cfs-client:v3.2.0) /cfs/bin中。

同时该客户端镜像已经集成了python环境(python2.7)，将用于执行参数解析脚本。

## 使用示例

### 创建并部署 ThinRuntimeProfile 资源

```shell
$ cat <<EOF > cubefs-profile.yaml
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntimeProfile
metadata:
  name: cubefs3.2
spec:
  fileSystemType: cubefs
  fuse:
    image: <IMG_REPO>
    imageTag: <IMG_TAG>
    imagePullPolicy: IfNotPresent 
    command:
      - sh
      - -c 
      - "python /fluid_config_init.py && chmod u+x /mount-cubefs.sh && /mount-cubefs.sh"
EOF

$ kubectl apply -f runtime-profile.yaml
```
将上述 <IMG_REPO> 改为您制作的镜像的仓库名称，<IMG_TAG> 修改为该镜像的 TAG。

### 创建并部署 Dataset 和 ThinRuntime 资源
```shell
$ cat <<EOF > dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: cubefs-test
spec:
  mounts:
    - mountPoint: <IP:Port>
      name: fluid-test
---
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntime
metadata:
  name: cubefs-test
spec:
  profileName: cubefs3.2
EOF

$ kubectl apply -f dataset.yaml
```
将上述 `mountPoint` 修改为您需要使用的CuebFS集群Master的地址, `name`修改为需要挂载的存储卷的名字。

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
        claimName: cubefs-test
EOF

$ kubectl apply -f app.yaml
```

查看应用Pod，可发现其正常运行。Fluid自动根据ThinRuntimeProfile中的Fuse配置，创建Fuse Pod，并调度到应用同一个节点上。

```shell
$ kubectl get pods
NAME                    READY   STATUS    RESTARTS   AGE
cubefs-test-fuse-wsd26  1/1     Running   0        2m56s
nginx                   1/1     Running   0        2m56s
```
远程的⽂件系统被挂载到 nginx pod 的 /data ⽬录下。

```
$ kubectl exec -it nginx bash

root@nginx:/# df -h
Filesystem      Size  Used Avail Use% 
...
chubaofs-fluid  5.0G  4.0K  5.0G   1% /data
...
```