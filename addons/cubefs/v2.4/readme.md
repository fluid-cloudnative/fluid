# CubeFS 2.4

This addon is built based for [CubeFS](https://cubefs.io/) v2.4.0.

## Install

To install the addon, apply the `runtime-profile.yaml` file:

```shell
kubectl apply -f runtime-profile.yaml
```

## Usage

### Prerequisites
CubeFS 2.4 has been deployed in the K8s cluster and can be accessed normally.

### Create and Deploy ThinRuntimeProfile Resource

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

### Create and Deploy Dataset and ThinRuntime Resource
To create and deploy the `Dataset` and `ThinRuntime` resources for the addon, run the following commands:

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
Modify the above `mountPoint` to the address of the Master of CubeFS you want to use. Modify `name` to the name of the storage volume to be mounted

### Run pod with Fluid PVC

To run a Pod with a Fluid PVC that uses the addon, run the following commands:

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

After the application using the remote file system is deployed, the corresponding FUSE pod is also scheduled to the same node.

To verify that the addon is working correctly, run the following command:

```shell
$ kubectl get pods
NAME                    READY   STATUS    RESTARTS   AGE
cubefs-test-fuse-lf8r4  1/1     Running   0        2m56s
nginx                   1/1     Running   0        2m56s
```
The remote file system is mounted to the /data directory of nginx pod.

```
$ kubectl exec -it nginx bash

root@nginx:/# df -h
Filesystem      Size  Used Avail Use% 
...
chubaofs-fluid  5.0G  4.0K  5.0G   1% /data
...
```

## How to develop

Please check [doc](./dev-guide/cubefs-v2.4.md).
