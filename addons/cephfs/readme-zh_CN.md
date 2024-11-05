# Ceph

该插件用于 [Ceph](https://ceph.com/).

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
  name: ceph-demo
spec:
  mounts:
  - mountPoint: ceph://<IP:Port>
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
EOF

$ kubectl apply -f dataset.yaml
```
将上面的`mountPoint`修改为实际Ceph存储的地址，并将相关验证信息填在`options`字段。

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
          name: ceph-demo
  volumes:
    - name: ceph-demo
      persistentVolumeClaim:
        claimName: ceph-demo
EOF

$ kubectl apply -f app.yaml
```
使用远程文件系统的应用部署完成后，对应的FUSE pod也会被调度到同一个节点上。

```shell
$ kubectl get pods
NAME                  READY   STATUS    RESTARTS   AGE
ceph-demo-fuse-7kfdx  1/1     Running   0          34s
nginx                 1/1     Running   0          47s
```
远程的文件系统被挂载到 nginx pod 的 /data 目录下。

## 如何开发

请参考文档[doc](dev-guide/cephfs-zh_CN.md)


## 版本

* 0.1

初始支持Ceph。
