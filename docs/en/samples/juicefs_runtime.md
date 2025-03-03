# DEMO - How to use JuiceFS in Fluid

## Background introduction

JuiceFS is an open source high-performance POSIX file system released under GNU Affero General Public License v3.0. It is specially optimized for the cloud-native environment. Provides complete POSIX compatibility, which can use massive and low-cost cloud storage as a local disk, and can also be mounted and read by multiple hosts at the same time.

About how to use JuiceFS you can refer to the document [JuiceFS Quick Start Guide](https://juicefs.com/docs/community/quick_start_guide).

## Installation

You can download the latest Fluid installation package from [Fluid Releases](https://github.com/fluid-cloudnative/fluid/releases).

Refer to the [Installation document](../userguide/install.md) to complete the installation.

```shell
$ kubectl get po -n fluid-system
NAME                                         READY   STATUS              RESTARTS   AGE
csi-nodeplugin-fluid-ctc4l                   2/2     Running             0          113s
csi-nodeplugin-fluid-k7cqt                   2/2     Running             0          113s
csi-nodeplugin-fluid-x9dfd                   2/2     Running             0          113s
dataset-controller-57ddd56b54-9vd86          1/1     Running             0          113s
fluid-webhook-84467465f8-t65mr               1/1     Running             0          113s
```

Make sure `dataset-controller`, `fluid-webhook` pod and `csi-nodeplugin` pods work well. `juicefs-runtime-controller` will be installed automatically when JuiceFSRuntime created.

## Create new work environment

```shell
$ mkdir <any-path>/juicefs
$ cd <any-path>/juicefs
```

## Demo

The fields required for using JuiceFS Community Edition and Cloud Service Edition are different. The following describes how to use them:

### Community edition

Before using JuiceFS, you need to provide parameters for metadata services (such as Redis) and object storage services (such as MinIO), and create corresponding secrets:

```shell
kubectl create secret generic jfs-secret \
    --from-literal=metaurl=redis://192.168.169.168:6379/1 \
    --from-literal=access-key=<accesskey> \
    --from-literal=secret-key=<secretkey>
```

- `metaurl`: Connection URL for metadata engine (e.g. Redis), it is required. Read [this document](https://juicefs.com/docs/community/databases_for_metadata/) for more information.
- `access-key`: Access key of object storage, not required, if your filesystem is already formated, can be empty.
- `secret-key`: Secret key of object storage, not required, if your filesystem is already formated, can be empty.

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
      mountPoint: "juicefs:///demo"     # Refers to the subdirectory of JuiceFS, starts with `juicefs://`. Required.
      options:
        bucket: "<bucket>"              # Bucket URL. Not required, if your filesystem is already formated, can be empty.
        storage: "minio"
      encryptOptions:
        - name: metaurl                 # Connection URL for metadata engine. Required.
          valueFrom:
            secretKeyRef:
              name: jfs-secret
              key: metaurl
        - name: access-key              # Access key of object storage. Not required, if your filesystem is already formated, can be empty.
          valueFrom:
            secretKeyRef:
              name: jfs-secret
              key: access-key
        - name: secret-key              # Secret key of object storage. Not required, if your filesystem is already formated, can be empty.
          valueFrom:
            secretKeyRef:
              name: jfs-secret
              key: secret-key
EOF
```

- `mountPoint`: Refers to the subdirectory of JuiceFS, which is the directory where users store data in the JuiceFS file system, starts with `juicefs://`. For example, `juicefs:///demo` is the `/demo` subdirectory of the JuiceFS file system.
- `bucket`: Bucket URL. For example, using S3 as object storage, bucket is `https://myjuicefs.s3.us-east-2.amazonaws.com`. Read [this document](https://juicefs.com/docs/community/how_to_setup_object_storage/) to learn how to setup different object storage.
- `storage`: Specify the type of storage to be used by the file system, e.g. `s3`, `gs`, `oss`. Read [this document](https://juicefs.com/docs/community/how_to_setup_object_storage/) for more details.

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
jfsdemo-worker-0                                           1/1     Running   0          4m2s
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
jfsdemo-worker-0                                               1/1     Running   0          10m
```

You can see that the pod has been created successfully, and the FUSE component of JuiceFS has also started successfully.

### Cloud service edition

Before using JuiceFS, you need to provide token managed by JuiceFS and object storage services (such as MinIO), and create corresponding secrets:

```shell
kubectl create secret generic jfs-secret \
    --from-literal=token=<token> \
    --from-literal=access-key=<accesskey>  \
    --from-literal=secret-key=<secretkey>
```

- `token`: JuiceFS managed token. Not required, if JuiceFS connected to meta server via `initconfig`, can be empty. Read [this document](https://juicefs.com/docs/cloud/metadata/#token-management) for more details.
- `access-key`: Access key of object storage, not required, if it has been configured in JuiceFS console, can be empty. 
- `secret-key`: Secret key of object storage, not required, if it has been configured in JuiceFS console, can be empty.

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
      mountPoint: "juicefs:///demo"   # Refers to the subdirectory of JuiceFS, starts with `juicefs://`. Required.
      options:
        bucket: "<bucket>"            # Bucket URL. Not required, if no display of the specification is required, can be empty.
      encryptOptions:
        - name: token                 # JuiceFS managed token. Not required, if JuiceFS connected to meta server via `initconfig`, can be empty.
          valueFrom:
            secretKeyRef:
              name: jfs-secret
              key: token
        - name: access-key            # Access key of object storage. Not required, if it has been configured in JuiceFS console, can be empty.
          valueFrom:
            secretKeyRef:
              name: jfs-secret
              key: access-key
        - name: secret-key            # Secret key of object storage. Not required, if it has been configured in JuiceFS console, can be empty.
          valueFrom:
            secretKeyRef:
              name: jfs-secret
              key: secret-key
EOF
```

- `name`: It needs to be the same as the volume name created in the JuiceFS console.
- `mountPoint`: Refers to the subdirectory of JuiceFS, which is the directory where users store data in the JuiceFS file system, starts with `juicefs://`. For example, `juicefs:///demo` is the `/demo` subdirectory of the JuiceFS file system.
- `bucket`: Bucket URL. For example, using S3 as object storage, bucket is `https://myjuicefs.s3.us-east-2.amazonaws.com`. Read [this document](https://juicefs.com/docs/community/how_to_setup_object_storage/) to learn how to setup different object storage.

> **Attention**: `name` is required.

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
jfsdemo-worker-0                                           1/1     Running   0          4m2s
```

`JuiceFSRuntime` does not have master, but the FUSE component implements lazy startup and will be created when the pod is used. The worker of JuiceFS cloud service edition provides distributed independent cache.

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
jfsdemo-worker-0                                               1/1     Running   0          10m
```

You can see that the pod has been created successfully, and the FUSE component of JuiceFS has also started successfully.
