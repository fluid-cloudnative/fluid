# Demo - Access Dataset cache across Namespace (Sidecar mechanism)
This demo is used to show how to use a Dataset cache across Namespace.
- In Namespace `ns-a`, create Dataset `demo` and AlluxioRuntime `demo`
- In Namespace `ns-b` create Dataset `demo-ref`. The mountPoint of `demo-ref` is `dataset://ns-a/demo`
 
## Prerequests
Before running this demo, please refer to the [installation documentation](../userguide/install.md) to complete the installation and check that the components of Fluid are working properly:
```shell
$ kubectl get pod -n fluid-system
alluxioruntime-controller-5b64fdbbb-84pc6   1/1     Running   0          8h
csi-nodeplugin-fluid-fwgjh                  2/2     Running   0          8h
csi-nodeplugin-fluid-ll8bq                  2/2     Running   0          8h
dataset-controller-5b7848dbbb-n44dj         1/1     Running   0          8h
thinruntime-controller-7dcbf5f45-xsf4p          1/1     Running   0          8h
```

Where `thinruntime-controller` is used to support Dataset sharing across Namespace and `alluxioruntime-controller` is the actual cache.

## Share Dataset cache across Namespace through CSI mechanism 
###  1. Create Dataset and Cache Runtime

In default Namespace，create `phy` Dataset and AlluxioRuntime.
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

### 2. Create referenced Dataset and Runtime
In Namespace `ref`, create:
- the referenced dataset `refdemo`, whose mountPoint format is `dataset://${origin-dataset-namespace}/${origin-dataset-name}`.
- ThinRuntime `refdemo`, and its Spec fields don't need to be filled.

Note:
1. Currently, the referenced Dataset only supports single mount and its form must be `dataset://` (i.e. the creation of a dataset fails when `dataset://` and other forms both appear), and other fields in the Spec are invalid.
2. The fields in Spec of the referenced Runtime corresponding to the Dataset are invalid.
```shell
$ kubectl create ns ref

$ cat<<EOF >ds-ref.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: refdemo
spec:
  mounts:
    - mountPoint: dataset://default/phy
      name: fusedemo
      path: "/"
EOF

$ kubectl create -f ds-ref.yaml -n ref
```

### Create Pod and Check the data

In Namespace `ref`, create a Pod：  
Need to enable serverless injection, set pod tag `serverless.fluid.io/inject=true`
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

In default Namespace, check the Pod information.
- Only one AlluxioRuntime cluster exists, i.e. there is only one cache.
- The sidecar mechanism is used, so there is no fuse pod.
```shell
$ kubectl get pods -o wide
NAME             READY   STATUS    RESTARTS   AGE     IP              NODE      NOMINATED NODE   READINESS GATES
phy-master-0     2/2     Running   0          7m2s    172.16.1.10     work02    <none>           <none>
phy-worker-0     2/2     Running   0          6m29s   172.16.1.10     work02    <none>           <none>
```

In Namespace `ref`, check the status of the app nginx pod.
```shell
$ kubectl get pods -n ref -o wide
NAME         READY   STATUS    RESTARTS   AGE   IP              NODE      NOMINATED NODE   READINESS GATES
nginx        1/1     Running   0          11m   10.233.109.66   work02    <none>           <none>
```

Check the yaml of the app nginx pod under the `ref` namespace and you can see that the `fuse` container is injected.
```shell
$ kubectl get pods nginx -n ref -o yaml
...
spec:
  containers:
    - image: fluidcloudnative/alluxio-fuse:release-2.8.1-SNAPSHOT-0433ade
      ...
    - image: nginx
...
```

### Known Issues

For the Fuse Sidecar scenario, some ConfigMap is created under the referenced namespace (`ref`).
```shell
NAME                                    DATA   AGE
check-fluid-mount-ready                 1      6d14h
phy-config                              7      6d15h
refdemo-fuse.alluxio-fuse-check-mount   1      6d14h
```
- `check-fluid-mount-ready` is shared by all Datasets under that namespace.
- `refdemo-fuse.alluxio-fuse-check-mount` is generated based on the Dataset name and Runtime type.
- `phy-config` is the ConfigMap required by AlluxioRuntime's Fuse Container and is therefore copied from the `default` namespace to the `ref` namespace.
  - **If a Dataset named `phy` using AlluxioRuntime previously existed under the `ref` namespace, then the use of the `refdemo` Dataset will be wrong.**