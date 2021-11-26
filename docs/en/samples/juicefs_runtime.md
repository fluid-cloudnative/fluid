# DEMO - How to use JuiceFS in Fluid

## Background introduction

JuiceFS is a high-performance POSIX file system released under GNU Affero General Public License v3.0. It is specially optimized for the cloud-native environment. Provides complete POSIX compatibility, which can use massive and low-cost cloud storage as a local disk, and can also be mounted and read by multiple hosts at the same time.

About how to use JuiceFS you can refer to the document [JuiceFS Quick Start Guide](https://juicefs.com/docs/community/quick_start_guide).

## Installation

You can download the latest Fluid installation package from [Fluid Releases](https://github.com/fluid-cloudnative/fluid/releases).

Set `runtime.juicefs.enable` to `true` in Fluid chart, then refer to the [Installation document](../userguide/install.md) to complete the installation

```shell
$ kubectl get po -n fluid-system
NAME                                         READY   STATUS              RESTARTS   AGE
csi-nodeplugin-fluid-ctc4l                   2/2     Running             0          113s
csi-nodeplugin-fluid-k7cqt                   2/2     Running             0          113s
csi-nodeplugin-fluid-x9dfd                   2/2     Running             0          113s
dataset-controller-57ddd56b54-9vd86          1/1     Running             0          113s
fluid-webhook-84467465f8-t65mr               1/1     Running             0          113s
juicefsruntime-controller-56df96b75f-qzq8x   1/1     Running             0          113s
```

Make sure `juicefsruntime-controller`, `dataset-controller`, `fluid-webhook` pod and `csi-nodeplugin` pods work well.

## Create new work environment

```shell
$ mkdir <any-path>/juicefs
$ cd <any-path>/juicefs
```

## Demo

Before using JuiceFS, you need to provide parameters for metadata services (such as Redis) and object storage services (such as MinIO), and create corresponding secrets:

```shell
kubectl create secret generic jfs-secret \
    --from-literal=metaurl=redis://192.168.169.168:6379/1 \
    --from-literal=access-key=<accesskey> \
    --from-literal=secret-key=<secretkey>
```

**Check `Dataset` to be created**

```yaml
$ cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: jfsdemo
spec:
  mounts:
    - name: minio
      mountPoint: "juicefs:///demo"
      options:
        bucket: "<bucket>"
        storage: "minio"
      encryptOptions:
        - name: metaurl
          valueFrom:
            secretKeyRef:
              name: jfs-secret
              key: metaurl
        - name: access-key
          valueFrom:
            secretKeyRef:
              name: jfs-secret
              key: access-key
        - name: secret-key
          valueFrom:
            secretKeyRef:
              name: jfs-secret
              key: secret-key
EOF
```

> **Note**: `/demo` refers to the subpath of JuiceFS, which is the directory of the JuiceFS file system where users store data in.

> **Attention**: Only `name` and `metaurl` are required. If the JuiceFS has been formatted, you only need to fill in the `name` and `metaurl`.

Since JuiceFS uses local cache, the corresponding `Dataset` supports only one mount, and JuiceFS does not have UFS, you can specify subdirectory in `mountPoint` (`juicefs:///` represents root directory), and it will be mounted as the root directory into the container.

**Create `Dataset`**
```shell
$ kubectl create -f dataset.yaml
dataset.data.fluid.io/jfsdemo created
```

**Check `Dataset` status**
```shell
$ kubectl get dataset jfsdemo
NAME      UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE      AGE
jfsdemo                                                                  NotBound   44s
```

As shown above, the value of the `phase` in `status` is `NotBound`, which means that the `Dataset` resource is not currently bound to any `JuiceFSRuntime` resource. Next, we will create `JuiceFSRuntime` resource.

**Check `JuiceFSRuntime` resource to be create**

```yaml
$ cat<<EOF >runtime.yaml
apiVersion: data.fluid.io/v1alpha1
kind: JuiceFSRuntime
metadata:
  name: jfsdemo
spec:
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 40960
        low: "0.1"
EOF
```

> Note: The smallest unit of `quota` in JuiceFS is the MiB

**Create `JuiceFSRuntime`**

```shell
$ kubectl create -f runtime.yaml
juicefsruntime.data.fluid.io/jfsdemo created
```

**Check `JuiceFSRuntime`**
```shell
$ kubectl get juicefsruntime
NAME      AGE
jfsdemo   34s
```

Wait a while for the various components of `JuiceFSRuntime` to start smoothly, and you will see status similar to the following:

```shell
$ kubectl get po |grep jfs
jfsdemo-worker-mjplw                                           1/1     Running   0          4m2s
```

`JuiceFSRuntime` does not have master, but the FUSE component implements lazy startup and will be created when the pod is used.

```shell
$ kubectl get juicefsruntime jfsdemo
NAME      AGE
jfsdemo   6m13s
```

Then, check the `Dataset` status again and find that it has been bound with `JuiceFSRuntime`.

```shell
$ kubectl get dataset jfsdemo
NAME      UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
jfsdemo   4.00KiB          -        40.00GiB         -                   Bound   9m28s
```

**Check Pod to be create**, the Pod uses the `Dataset` created above to specify the PVC with the same name.

```yaml
$ cat<<EOF >sample.yaml
apiVersion: v1
kind: Pod
metadata:
  name: demo-app
spec:
  containers:
    - name: demo
      image: nginx
      volumeMounts:
        - mountPath: /data
          name: demo
  volumes:
    - name: demo
      persistentVolumeClaim:
        claimName: jfsdemo
EOF
```

**Create Pod**

```shell
$ kubectl create -f sample.yaml
pod/demo-app created
```

**Check Pod**
```shell
$ kubectl get po |grep demo
demo-app                                                       1/1     Running   0          31s
jfsdemo-fuse-fx7np                                             1/1     Running   0          31s
jfsdemo-worker-mjplw                                           1/1     Running   0          10m
```

You can see that the pod has been created successfully, and the FUSE component of JuiceFS has also started successfully.
