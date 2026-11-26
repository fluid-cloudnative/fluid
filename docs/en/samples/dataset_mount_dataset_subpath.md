# DEMO - Mount a Subpath of Another Dataset
This demo shows how a Dataset can mount a subpath of another Dataset:
- In namespace `default`, create the Dataset `phy` and the AlluxioRuntime `phy`;
- In namespace `ref`, create the Dataset `refdemo`, which mounts `dataset://default/phy/spark`.

This demo uses the Sidecar mechanism. To use the CSI mechanism instead, see also [Share data across namespace (CSI mode)](./dataset_across_namespace_with_csi.md).
## Prerequisites
Before running this demo, please refer to the [Installation Guide](../userguide/install.md) to install Fluid, and check that all Fluid components are running:
```shell
$ kubectl get pod -n fluid-system
alluxioruntime-controller-5b64fdbbb-84pc6   1/1     Running   0          8h
csi-nodeplugin-fluid-fwgjh                  2/2     Running   0          8h
csi-nodeplugin-fluid-ll8bq                  2/2     Running   0          8h
dataset-controller-5b7848dbbb-n44dj         1/1     Running   0          8h
thinruntime-controller-7dcbf5f45-xsf4p          1/1     Running   0          8h
```

`thinruntime-controller` supports mounting a Dataset, and `alluxioruntime-controller` provides the actual cache.

## Share the Dataset cache across namespaces with the Sidecar mechanism
###  1. Create the Dataset and the cache Runtime

In the default namespace, create the `phy` Dataset and AlluxioRuntime:
```shell
$ cat<<EOF >ds.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: phy
spec:
  mounts:
    - mountPoint: https://mirrors.bit.edu.cn/apache/spark/
      name: spark
---
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: phy
spec:
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 1Gi
        high: "0.95"
        low: "0.7"
EOF

$ kubectl create -f ds.yaml
```

### 2. Create the referencing Dataset and Runtime
In the `ref` namespace, create:
- the referencing dataset `refdemo`, whose mountPoint format is `dataset://${origin-dataset-namespace}/${origin-dataset-name}/${subpath}`; `$subpath` can have several levels;
- ThinRuntime `refdemo`, whose Spec fields don't need to be filled.

Note:
1. Currently, the referencing Dataset supports only one mount, and it must use the `dataset://` form (creating a dataset that mixes `dataset://` with other forms fails). The other fields in its Spec have no effect.
2. The fields in the Spec of the Runtime of the referencing Dataset have no effect.
```shell
$ kubectl create ns ref

$ cat<<EOF >ds-ref.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: refdemo
spec:
  mounts:
    - mountPoint: dataset://default/phy/spark
      name: fusedemo
      path: "/"
EOF

$ kubectl create -f ds-ref.yaml -n ref
```

### Create a Pod and check the data

In the `ref` namespace, create a Pod.
Serverless injection must be enabled: set the Pod label `serverless.fluid.io/inject=true`.
```shell
$ cat<<EOF >app-ref.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  labels:
    serverless.fluid.io/inject: "true"
spec:
  containers:
    - name: nginx
      image: nginx
      volumeMounts:
        - mountPath: /data_spark
          name: spark-vol
  volumes:
    - name: spark-vol
      persistentVolumeClaim:
        claimName: refdemo
EOF

$ kubectl create -f app-ref.yaml -n ref
```

Check that the nginx Pod in the `ref` namespace is running, and look at the mounted data:
```shell
$ kubectl get pods -n ref -o wide
NAME         READY   STATUS    RESTARTS   AGE   IP              NODE      NOMINATED NODE   READINESS GATES
nginx        1/1     Running   0          11m   10.233.109.66   work02    <none>           <none>

# The data path in the Pod shows the content of the `spark` subpath of the phy dataset
$ kubectl exec nginx -c nginx -n ref -it -- ls /data_spark
spark-3.1.3   spark-3.2.3  spark-3.3.1
```

The YAML of the nginx Pod in the `ref` namespace shows that the volume was rewritten to the specific subpath:
```shell
$ kubectl get pods nginx -n ref -o yaml
...
spec:
  containers:
    - image: fluidcloudnative/alluxio-fuse:release-2.8.1-SNAPSHOT-0433ade
      ...
    - image: nginx
      volumeMounts:
      - mountPath: /data_spark
        mountPropagation: HostToContainer
        name: spark-vol
      ...
  volumes:
  - hostPath:
      path: /runtime-mnt/alluxio/default/phy/alluxio-fuse/spark
      type: ""
    name: spark-vol
...
```

### Known issues

With FUSE Sidecar, some ConfigMaps are created in the referencing namespace (`ref`):
```shell
NAME                                    DATA   AGE
check-fluid-mount-ready                 1      6d14h
phy-config                              7      6d15h
refdemo-fuse.alluxio-fuse-check-mount   1      6d14h
```
- `check-fluid-mount-ready` is shared by all Datasets in the namespace;
- `refdemo-fuse.alluxio-fuse-check-mount` is generated from the Dataset name and the Runtime type;
- `phy-config` is the ConfigMap that the FUSE container of the AlluxioRuntime needs, so it is copied from the `default` namespace to the `ref` namespace;
  - **If the `ref` namespace already had a Dataset named `phy` that uses an AlluxioRuntime, using the `refdemo` Dataset fails.**
