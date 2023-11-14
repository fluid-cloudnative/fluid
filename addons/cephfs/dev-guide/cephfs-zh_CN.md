# Ceph接入ThinRuntime的简单示例

# 前期准备

⾸先需要在 K8s 集群能够访问到的机器上部署[Ceph系统](https://ceph.com/en/)并配置读写权限，并且保证在 K8s 集群的节点上 能够访问到该Cephfs服务。

# 制作Ceph-FUSE客户端镜像

### 1.编写参数解析脚本

在 FUSE 容器内需要提取相关的 **ThinRuntimeProfile**、**Dataset**、**ThinRuntime**资源中对远程⽂件系统的配置信息，相关信息以 JSON 字符串的⽅式保存到 FUSE 容器的 **/etc/fluid/config.json** ⽂件内。

~~~ python
# fluid_config_init.py
import json


def write_conf(pvAttrs: dict):
    confAttrs = pvAttrs
    with open("/etc/ceph/ceph.conf", "w") as f:
        f.write("[global]\n")
        f.write("fsid=%s\n" % confAttrs["fsid"])
        f.write("mon_initial_members=%s\n" % confAttrs["mon_initial_members"])
        f.write("mon_host=%s\n" % confAttrs["mon_host"])
        f.write("auth_cluster_required=%s\n" % confAttrs["auth_cluster_required"])
        f.write("auth_service_required=%s\n" % confAttrs["auth_service_required"])
        f.write("auth_client_required=%s\n" % confAttrs["auth_client_required"])


def write_keyring(pvAttrs: dict):
    keyringAttrs = pvAttrs
    with open("/etc/ceph/ceph.client.admin.keyring", "w+") as f:
        f.write("[client.admin]\n")
        f.write("key=%s\n" % keyringAttrs["key"])


def read_json():
    with open("/etc/fluid/config.json", "r") as f:
        rawStr = f.readlines()
    rawStr = "".join(rawStr)
    obj = json.loads(rawStr)
    return obj


def write_cmd(mon_url: str, target_path):
    mon_url = mon_url.replace("ceph://", "")
    script = """#!/bin/sh
mkdir -p {}
exec ceph-fuse -n client.admin -k /etc/ceph/ceph.client.admin.keyring -c /etc/ceph/ceph.conf  {}
"""
    with open("/mount_ceph.sh", "w+") as f:
        f.write(script.format(target_path, target_path))


if __name__ == '__main__':
    pvAttrs = read_json()
    write_conf(pvAttrs['mounts'][0]['options'])
    write_keyring(pvAttrs['mounts'][0]['options'])
    write_cmd(pvAttrs['mounts'][0]['mountPoint'], pvAttrs['targetPath'])
~~~

该python脚本会从**/etc/fluid/config.json**文件中提取并生成配置文件**/etc/ceph/ceph.conf**和密钥文件**/etc/ceph/ceph.client.admin.keyring**，以及挂载脚本**/mount_ceph.sh**。



### 2.挂载ceph-fuse脚本

在将参数解析并注⼊shell脚本后，⽣成的脚本如下：

~~~ shell
#!/bin/sh
mkdir -p /runtime-mnt/thin/default/my-storage/thin-fuse
exec ceph-fuse -n client.admin -k /etc/ceph/ceph.client.admin.keyring -c /etc/ceph/ceph.conf  /runtime-mnt/thin/default/my-storage/thin-fuse
sleep inf
~~~

该 shell 脚本创建挂载的⽂件夹并将远程⽂件系统挂载到⽬标位置（targetPath）。**由于 mount 命令会⽴即返回，为了保持该进程的持续运⾏（防⽌FUSE pod 反复重启），需要 sleep inf 来保持进程的存在。**



### 3.编辑启动脚本

编辑容器启动时需要执行的脚本`entrypoint.sh`，主要执行`fluid_config_init.py`生成挂载脚本`mount_ceph.sh`，然后执行`mount_ceph.sh`脚本进行挂载。

~~~ shell
#!/usr/bin/env sh
set +x

python3 /fluid_config_init.py

chmod u+x /mount_ceph.sh

sh /mount_ceph.sh
~~~



### 4.创建FUSE客户端镜像

将参数解析脚本、挂载脚本和相关的库打包⼊镜像。

~~~ dockerfile
FROM alpine@sha256:124c7d2707904eea7431fffe91522a01e5a861a624ee31d03372cc1d138a3126
# use alpine:3.18

RUN mkdir /etc/ceph
RUN apk add ceph ceph-fuse python3

ADD fluid_config_init.py /
ADD entrypoint.sh /usr/local/bin

RUN chmod u+x /usr/local/bin/entrypoint.sh

CMD ["/usr/local/bin/entrypoint.sh"]
~~~

除了 Python 脚本和 Shell 脚本外，还需要在基镜像上**安装 python 环境和 ceph-fuse 客户端**。



# 使用示例

### 1.创建并部署ThinRuntimeProfile资源

~~~ shell
$ cat <<EOF > profile.yaml
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntimeProfile
metadata:
  name: ceph-profile
spec:
  fileSystemType: ceph-fuse
  fuse:
    image: <IMG_REPO>
    imageTag: <IMG_TAG>
    imagePullPolicy: IfNotPresent
    command:
      - "/usr/local/bin/entrypoint.sh"
EOF

$ kubectl apply -f profile.yaml
~~~

将上述<IMG_REPO>改为您制作的镜像的仓库名称，<IMG_TAG>修改为该镜像的TAG。其中**fileSystemType**要对应挂载的文件类型，可通过在挂载的pod中使用mount命令查看。例如，我在ceph-fuse挂载完成后的pod中执行mount，可以看到如下的一条记录：

~~~ shell
ceph-fuse on /runtime-mnt/thin/default/ceph-demo/thin-fuse type fuse.ceph-fuse (rw,nosuid,nodev,relatime,user_id=0,group_id=0,allow_other)
~~~

其中的fuse.ceph-fuse就是mount类型，可以选择其任意子字符串填写到fileSystemType。



### 2.创建并部署 Dataset 和 ThinRuntime 资源

下面是**Dataset**和**Thinruntime**的yaml文件。

~~~ yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: ceph-demo
spec:
  mounts:
  - mountPoint: ceph://<IP:PORT>
    name: ceph-pvc
    options:
      fsid: <fsid>
      mon_initial_members: <mon_initial_members>
      mon_host: <mon_host>
      auth_cluster_required: <auth_cluster_required>
      auth_service_required: <auth_service_required>
      auth_client_required: <auth_client_required>
      key: <key>
---
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntime
metadata:
  name: ceph-demo
spec:
  profileName: ceph-profile
~~~

其中**mounts**字段会被转换成**json**挂载到**/etc/fluid/config.json**，然后通过上述的python脚本解析后转换成配置文件和挂载脚本。

上述**yaml**文件对于json文件为：

~~~ json
{
  "mounts": [
    {
      "mountPoint": "ceph://<IP:PORT>",
      "name": "ceph-pvc",
      "options": {
        "key": "<key>",
        "fsid": "<fsid>",
        "mon_initial_members": "<mon_initial_members>",
        "mon_host": "<mon_host>",
        "auth_cluster_required": "<auth_cluster_required>",
        "auth_service_required": "<auth_service_required>",
        "auth_client_required": "<auth_client_required>"
      }
    }
  ],
  "targetPath": "/runtime-mnt/thin/default/ceph-demo/thin-fuse"
}
~~~

**targetPath**即为我们在fuse-pod中执行**mount**脚本的本地挂载点。



### 3.部署应用

~~~ shell
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
    - ls /data && sleep inf
    volumeMounts:
    - mountPath: /data
      name: data-vol
  volumes:
  - name: data-vol
    persistentVolumeClaim:
      claimName: ceph-demo

$ kubectl apply -f app.yaml
~~~

在将使⽤该远程⽂件系统的应⽤部署后，相应的 FUSE pod 也被调度到相同节点。

~~~ shell
$ kubectl get pods
NAME                  READY   STATUS    RESTARTS   AGE
ceph-demo-fuse-7kfdx  1/1     Running   0          34s
nginx                 1/1     Running   0          47s
~~~

远程的文件系统被挂载到 nginx pod 的 /data 目录下。
