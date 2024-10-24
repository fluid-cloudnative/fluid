# CubeFS 2.4

该插件用于 [CubeFS](https://cubefs.io/) v2.4.0.

## 安装

使用以下命令安装：

```shell
kubectl apply -f runtime-profile.yaml
```

## 使用

### 前置条件
K8s集群中已经部署CubeFS 2.4，并且可以正常访问。

### 创建并部署 ThinRuntimeProfile 资源

使用以下命令创建并部署ThinRuntimeProfile资源：
```shell
$ cat <<EOF > runtime-profile.yaml
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntimeProfile
metadata:
  name: cubefs2.4
spec:
  fileSystemType: cubefs
  fuse:
    image: fluidcloudnative/cubefs_v2.4
    imageTag: v0.1
    imagePullPolicy: IfNotPresent
    command:
      - "/usr/local/bin/entrypoint.sh"
EOF

$ kubectl apply -f runtime-profile.yaml
```


### 创建 Dataset 和 ThinRuntime 

使用以下命令创建Dataset和ThinRuntime：

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
  profileName: cubefs2.4
EOF

$ kubectl apply -f dataset.yaml
```
将上述 `mountPoint` 修改为您需要使用的CuebFS集群Master的地址，`name`修改为需要挂载的存储卷的名字。

### 数据访问应用示例

使用以下命令创建数据访问应用示例：

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

运行以上命令后，应用Pod将会自动创建Fuse Pod，并将其调度到与应用相同的节点上。您可以通过以下命令检查Pod状态：

```shell
$ kubectl get pods
NAME                    READY   STATUS    RESTARTS   AGE
cubefs-test-fuse-lf8r4  1/1     Running   0        2m56s
nginx                   1/1     Running   0        2m56s
```

您可以通过以下命令检查远程的⽂件系统被挂载到 nginx pod 的 /data ⽬录下。

```
$ kubectl exec -it nginx bash

root@nginx:/# df -h
Filesystem      Size  Used Avail Use% 
...
chubaofs-fluid  5.0G  4.0K  5.0G   1% /data
...
```

## 如何开发
请参考文档 [doc](./dev-guide/cubefs-v2.4-zh_CN.md).
