# GlusterFS

该插件用于 [GlusterFS](https://www.gluster.org/).

## Install

```shell
kubectl apply -f runtime-profile.yaml
```

## How to use

### Create and Deploy Dataset and ThinRuntime Resource
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
Modify the above mountPoint to the address of the remote glusterfs you want to use.

### Run pod with Fluid PVC

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
After the application using the remote file system is deployed, the corresponding FUSE pod is also scheduled to the same node.

```shell
$ kubectl get pods
NAME                        READY   STATUS    RESTARTS   AGE
glusterfs-demo-fuse-wx7ns   1/1     Running   0          12s
nginx                       1/1     Running   0          26s
```
The remote file system is mounted to the /data directory of nginx pod.

## How to develop

Please check [doc](dev-guide/glusterfs.md)
