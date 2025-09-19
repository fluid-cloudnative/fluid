# Curvine

This addon integrates [Curvine](https://github.com/CurvineIO/curvine) with Fluid via ThinRuntime.

## Install

Apply the `runtime-profile.yaml` file in this directory:

```shell
kubectl apply -f runtime-profile.yaml
```

## Usage

### Prerequisites
- A Curvine cluster is deployed and reachable from the Kubernetes cluster
- Fluid is installed in the cluster

### Create and Deploy ThinRuntimeProfile Resource

```shell
$ cat <<EOF > runtime-profile.yaml
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntimeProfile
metadata:
  name: curvine
spec:
  fileSystemType: fuse
  fuse:
    image: fluid-cloudnative/curvine-thinruntime
    imageTag: v1.0.0
    imagePullPolicy: IfNotPresent
EOF

$ kubectl apply -f runtime-profile.yaml
```

### Create and Deploy Dataset and ThinRuntime Resource

```shell
$ cat <<EOF > dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: curvine-demo
spec:
  mounts:
    - mountPoint: curvine:///data
      name: curvine
      options:
        master-endpoints: "<CURVINE_MASTER_IP:PORT>"
---
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntime
metadata:
  name: curvine-demo
spec:
  profileName: curvine
EOF

$ kubectl apply -f dataset.yaml
```
Modify `master-endpoints` to your Curvine Master address and `mountPoint` to the Curvine path you want to mount.

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
        claimName: curvine-demo
EOF

$ kubectl apply -f app.yaml
```

After the application is deployed, the corresponding FUSE pod is scheduled to the same node. The Curvine filesystem is mounted to `/data` in the application container.

## How to develop

Please check [dev guide](./dev-guide/curvine.md).


