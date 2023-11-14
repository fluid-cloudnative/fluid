# NFS

该插件用于 [NFS](https://nfs.sourceforge.net/).

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
  profileName: nfs
EOF

$ kubectl apply -f dataset.yaml
```
Modify the above mountPoint to the address of the remote NFS you want to use.

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
          name: nfs-demo
  volumes:
    - name: nfs-demo
      persistentVolumeClaim:
        claimName: nfs-demo
EOF

$ kubectl apply -f app.yaml
```
使用远程文件系统的应用部署完成后，对应的FUSE pod也会被调度到同一个节点上。

```shell
$ kubectl get pods
NAME                  READY   STATUS    RESTARTS   AGE
nfs-demo-fuse-wx7ns   1/1     Running   0          12s
nginx                 1/1     Running   0          26s
```
The remote file system is mounted to the /data directory of nginx pod.


## 如何开发
