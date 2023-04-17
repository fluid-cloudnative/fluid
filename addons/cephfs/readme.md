# Ceph

This addon is built based for [Ceph](https://ceph.com/).

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
Modify the above mountPoint to the address of the remote Ceph you want to use, and fill  the verification information in the `options` field.

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
          name: ceph-demo
  volumes:
    - name: ceph-demo
      persistentVolumeClaim:
        claimName: ceph-demo
EOF

$ kubectl apply -f app.yaml
```
After the application using the remote file system is deployed, the corresponding FUSE pod is also scheduled to the same node.

```shell
$ kubectl get pods
NAME                  READY   STATUS    RESTARTS   AGE
ceph-demo-fuse-7kfdx  1/1     Running   0          34s
nginx                 1/1     Running   0          47s
```
The remote file system is mounted to the /data directory of nginx pod.


## How to develop

Please check [doc](dev-guide/cephfs.md)


## Versions

* 0.1

Add init support for Ceph.
