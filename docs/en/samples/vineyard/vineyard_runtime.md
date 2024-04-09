# How to Use Vineyard Runtime in Fluid

## Background

Vineyard is an open-source in-memory data management system designed to provide high-performance data sharing and data exchange. Vineyard achieves zero-copy data sharing by storing data in shared memory, providing high-performance data sharing and data exchange capabilities.

For more information on how to use Vineyard, see the [Vineyard Quick Start Guide](https://v6d.io/notes/getting-started.html).

## Install Fluid

Refer to the [Installation Documentation](../../userguide/install.md) to complete the installation.

## Create Vineyard Runtime and Dataset

```shell
$ cat <<EOF | kubectl apply -f -
apiVersion: data.fluid.io/v1alpha1
kind: VineyardRuntime
metadata:
  name: vineyard
spec:
  replicas: 2
  tieredstore:
    levels:
    - mediumtype: MEM
      quota: 20Gi
---
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: vineyard
EOF
```

In VineyardRuntime:

- `spec.replicas`: Specifies the number of Vineyard Workers.
- `spec.tieredstore`: Specifies the storage configuration of Vineyard Worker, including storage levels and storage capacity. Here, a memory storage level with a capacity of 20Gi is configured.

In Dataset:

- `metadata.name`: Specifies the name of the Dataset, which must be consistent with the `metadata.name` in VineyardRuntime.

Check if the `Vineyard Runtime` is created successfully:

```shell
$ kubectl get vineyardRuntime vineyard 
NAME       MASTER PHASE   WORKER PHASE   FUSE PHASE   AGE
vineyard   Ready          PartialReady   Ready        3m4s
```

Then look at the status of the `Vineyard Dataset` and notice that it has been bound to the `Vineyard Runtime`:

```shell
$ kubectl get dataset vineyard
NAME       UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
vineyard                                                                  Bound   3m9s
```

## Create an Application Pod and Mount the Vineyard Dataset

```shell
$ cat <<EOF | kubectl apply -f -
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
        claimName: vineyard
EOF
```

Check if the Pod is created successfully:

```shell
$ kubectl get pod demo-app
NAME       READY   STATUS    RESTARTS   AGE
demo-app   1/1     Running   0          25s
```

Check the status of Vineyard FUSE:

```shell
$ kubectl get po | grep vineyard-fuse
vineyard-fuse-9dv4d                    1/1     Running   0               1m20s
```
