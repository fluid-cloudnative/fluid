# GlusterFS

该插件用于 [GlusterFS](https://www.gluster.org/).

## 安装

```shell
kubectl apply -f runtime-profile.yaml
```

## 使用

### 创建 Dataset 和 ThinRuntime 
```shell
$ cat <<EOF > dataset.yaml
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

$ kubectl apply -f dataset.yaml
```
将上述`mountPoint`修改为实际挂载的glusterfs地址

### 运行Pod，并且使用Fluid PVC

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
使用远程文件系统的应用部署完成后，对应的FUSE pod也会被调度到同一个节点上。

```shell
$ kubectl get pods
NAME                        READY   STATUS    RESTARTS   AGE
glusterfs-demo-fuse-wx7ns   1/1     Running   0          12s
nginx                       1/1     Running   0          26s
```
远程文件系统被挂载到nginx容器的`/data`目录下


## 如何开发
请参考 [doc](dev-guide/glusterfs-zh_CN.md)
