# DEMO - Configure the Metadata Sync Policy

## Prerequisites

Before running this demo, please refer to the [Installation Guide](../userguide/install.md) to install Fluid, and check that all Fluid components are running:
```shell
$ kubectl get pod -n fluid-system
NAME                                  READY   STATUS    RESTARTS   AGE
alluxioruntime-controller-5b64fdbbb-84pc6   1/1     Running   0          8h
csi-nodeplugin-fluid-fwgjh                  2/2     Running   0          8h
csi-nodeplugin-fluid-ll8bq                  2/2     Running   0          8h
dataset-controller-5b7848dbbb-n44dj         1/1     Running   0          8h
```
Normally, you shall see a Pod named `dataset-controller` and several Pods named `csi-nodeplugin`. The number of `csi-nodeplugin` Pods depends on the number of nodes in your Kubernetes cluster.

## Run the Demo

**Check the Dataset to create**

```shell
$ cat <<EOF > dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: hbase
spec:
  mounts:
    - mountPoint: https://mirrors.tuna.tsinghua.edu.cn/apache/hbase/stable/
      name: hbase
EOF
```

> Note: the `mountPoint` above uses the Tsinghua Apache mirror for the demo. If your environment is outside China, try `https://downloads.apache.org/hbase/stable/` instead.

Here we create a resource object of kind `Dataset`. `Dataset` is a Custom Resource Definition (CRD) defined by Fluid, used to tell Fluid where to find the data you need. Fluid mounts the `mountPoint` defined in this object onto Alluxio, so it can be any valid UFS address that Alluxio recognizes. For simplicity, this demo uses [WebUFS](https://docs.alluxio.io/os/user/stable/en/ufs/WEB.html).

For more information about UFS, please refer to the [Alluxio documentation - Storage Integrations](https://docs.alluxio.io/os/user/stable/en/ufs/OSS.html).

> This demo uses the HBase resources on an Apache mirror as the remote files. There is nothing special about this choice: you can use any remote files you like. If you want to use WebUFS as we do, an Apache mirror (e.g. the [Tsinghua mirror](https://mirrors.tuna.tsinghua.edu.cn/apache)) is the best choice, because with the current WebUFS implementation a more complex web page may need [more configuration](https://docs.alluxio.io/os/user/stable/en/ufs/WEB.html#configuring-alluxio).

**Create the Dataset**

```shell
$ kubectl create -f dataset.yaml
dataset.data.fluid.io/hbase created
```

**Check the AlluxioRuntime to create**

Set `spec.management.metadataSyncPolicy` of the AlluxioRuntime as needed. For example, the runtime below sets `spec.management.metadataSyncPolicy.autoSync=false` to turn off Fluid's automatic metadata sync. When the under storage system (UFS) mounted by Alluxio holds many files, turning off metadata sync lowers the load on the Alluxio master at startup.

```shell
$ cat<<EOF >runtime.yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: hbase
spec:
  management:
    metadataSyncPolicy:
      autoSync: false
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 2Gi
        high: "0.95"
        low: "0.7"
EOF
```

**Create the AlluxioRuntime**
```shell
$ kubectl create -f runtime.yaml
alluxioruntime.data.fluid.io/hbase created
```

**Check the Dataset status**

Once the Dataset and the AlluxioRuntime are bound, check the Dataset status. Because automatic metadata sync is off, Fluid cannot compute and show the basic information of the dataset, for example:

```
$ kubectl get dataset hbase
NAME     UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE      AGE
hbase                     0.00B    4.00GiB                              Bound      2m58s
```

**Turn automatic metadata sync back on**

You can turn Fluid's automatic metadata sync on at runtime by changing `AlluxioRuntime.spec.management.metadataSyncPolicy.autoSync`. For example, set it to `true` with:

```
$ kubectl patch alluxioruntime hbase --type='json' -p='[{"op": "replace", "path": "/spec/management/metadataSyncPolicy/autoSync", "value": true}]'
```

After a while, check the Dataset status again: the basic information of the dataset is now shown:
```
NAME    UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
hbase   567.22MiB        0.00B    2.00GiB          0.0%                Bound   5m3s
```
