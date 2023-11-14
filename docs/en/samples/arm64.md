# Demo - Using Fluid on ARM64 platform

## Prerequisites

### Get the latest code of Fluid
The official image for multiple platforms is provided by Fluid community. You can refer to the [document](../userguide/get_started.md) to deploy Fluid diretly and jump to the step of *Example*.
```shell
$ mkdir -p $GOPATH/src/github.com/fluid-cloudnative/
$ cd $GOPATH/src/github.com/fluid-cloudnative
$ git clone https://github.com/fluid-cloudnative/fluid.git
$ cd $GOPATH/src/github.com/fluid-cloudnative/fluid
```

### Build AMD64 and ARM64 container images

```shell
$ export DOCKER_CLI_EXPERIMENTAL=enabled
$ docker run --rm --privileged docker/binfmt:66f9012c56a8316f9244ffd7622d7c21c1f6f28d
$ docker buildx create --use --name mybuilder
$ export IMG_REPO=<your docker image repo>
$ make docker-buildx-all-push
```

> Get the built image address and image version

### Change the image version in helm chart

```shell
$ cd $GOPATH/src/github.com/fluid-cloudnative/fluid/charts/fluid/fluid
$ vim values.yaml
```

### Create an ARM64 based Kubernetes cluster and install helm chart

#### View ARM64 based nodes

```shell
kubectl get no -l kubernetes.io/arch=arm64
NAME                       STATUS   ROLES    AGE    VERSION
cn-beijing.192.168.3.183   Ready    <none>   6d3h   v1.22.10
cn-beijing.192.168.3.184   Ready    <none>   6d3h   v1.22.10
cn-beijing.192.168.3.185   Ready    <none>   6d3h   v1.22.10
```

#### Installation

```
$ kubectl create ns fluid-system
$ helm install fluid fluid
NAME: fluid
LAST DEPLOYED: Sat Aug 20 21:43:27 2022
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

## Example

This example is based on JuiceFSRuntime 


1. Creating an ARM64 based Kubernetes cluster

2. Refer to the [document](juicefs_setup.md) to prepare JuiceFS Community Edition

Before using JuiceFS, you need to provide parameters for metadata services (such as Redis) and object storage services (such as MinIO), and create corresponding secrets:

```shell
kubectl create secret generic jfs-secret \
    --from-literal=metaurl=redis://redis:6379/0 \
    --from-literal=access-key=minioadmin \
    --from-literal=secret-key=minioadmin
```

- `metaurl`: Connection URL for metadata engine (e.g. Redis). Read [this document](https://juicefs.com/docs/community/databases_for_metadata/) for more information.
- `access-key`: Access key of object storage.
- `secret-key`: Secret key of object storage.

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
      mountPoint: "juicefs:///"
      options:
        bucket: "http://minio:9000/minio/test"
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


- `mountPoint`: Refers to the subdirectory of JuiceFS, which is the directory where users store data in the JuiceFS file system, starts with `juicefs://`. For example, `juicefs:///demo` is the `/demo` subdirectory of the JuiceFS file system.
- `bucket`：Bucket URL。For example, using S3 as object storage, and the name of the bucket is test.
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
        quota: 4Gi
        low: "0.1"
EOF
```

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
$ kubectl get juicefs jfsdemo
NAME      WORKER PHASE   FUSE PHASE   AGE
jfsdemo   Ready          Ready        2m15s
```


Then, check the `Dataset` status again and find that it has been bound with `JuiceFSRuntime`.

```shell
$ kubectl get dataset jfsdemo
NAME      UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
jfsdemo   0.00B            0.00B    4.00GiB          0.0%                Bound   3m16s
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