# Demo - Data Preloading

In order to ensure the performance of the application when accessing the data, 
the data in the remote storage system can be pulled to the distributed cache engine
that is close to the computing node through **data preloading** beofre the application starts. 
Then the application that consumes the data can enjoy the acceleration effect brought by distributed cache even at the first time.

For the great benefit mentioned above, we provide **DataLoad CRD**. This is a CRD which offers you a clear and easy way to control data preloading behaviors.

This document will introduce you two different ways about how to use DataLoad CRD:
- [DataLoad Quick Usage](#dataload-quick-usage)
- [DataLoad Advanced Configurations](#dataload-advanced-configurations)

## Prerequisite

- [Fluid](https://github.com/fluid-cloudnative/fluid)(version >= 0.4.0)

Please refer to the [installation guide](https://github.com/fluid-cloudnative/fluid/blob/master/docs/zh/userguide/install.md) to complete the installation of fluid.

## Set Up Workspace
```
$ mkdir <any-path>/warmup
$ cd <any-path>/warmup
```

## DataLoad Quick Usage

**Check the Dataset and AlluxioRuntime objects to be created**
```yaml
cat << EOF > dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: spark
spec:
  mounts:
    - mountPoint: https://mirrors.tuna.tsinghua.edu.cn/apache/spark/
      name: spark
---
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: spark
spec:
  replicas: 2
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 2Gi
        high: "0.95"
        low: "0.7"
EOF
```

> Notes: Here, we use THU's tuna Apache mirror site as our `mountPoint`. If your environment isn't in Chinese mainland, please replace it with `https://downloads.apache.org/spark/`.

Here, we'd like to create a resource object with kind `Dataset`. `Dataset` is a Custom Resource Definition(CRD) defined by Fluid and used to tell Fluid where to find all the data you'd like to access.
In this guide, we'll use [WebUFS](https://docs.alluxio.io/os/user/stable/en/ufs/WEB.html) for its simplicity.

For more information about UFS, please refer to [Alluxio Docs - Storage Integrations](https://docs.alluxio.io/os/user/stable/en/ufs/HDFS.html)

> We use Apache Spark on a mirror site of Apache downloads as an example of remote file. It's nothing special, you can change it to any remote file you like. But please note that, if you are going to use WebUFS like we do, files on Apache sites are highly recommended because you might need some [advanced configurations](https://docs.alluxio.io/os/user/stable/en/ufs/WEB.html#configuring-alluxio) due to current implementation of WebUFS.

**Create the Dataset and AlluxioRuntime**

```
kubectl create -f dataset.yaml
```

**Wait for the Dataset and AlluxioRuntime to be ready**

You can check their status by running:

```
kubectl get datasets spark
```

Dataset and Runtime are all ready if you see something like this:

```
NAME    UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
spark   1.92GiB          0.00B    4.00GiB          0.0%                Bound   4m4s
```

**Check the DataLoad object to be created**

```yaml
cat <<EOF > dataload.yaml
apiVersion: data.fluid.io/v1alpha1
kind: DataLoad
metadata:
  name: spark-dataload
spec:
  dataset:
    name: spark
    namespace: default
EOF
```

`spec.dataset` specifies the target dataset that needs to be preloaded. In this example, our target is the Dataset named `spark` under the `default` namespace. 
Feel free to change the configuration above if it doesn't match your actual environment. ** note ** The namespace of your DataLoad must be consistent with the namespace of your dataset.

**By default, it'll preload all the data in the target dataset**. If you'd like to control the data preloading behaviors in a more find-grained way(e.g. preload data under some specified path only),
please refer to [DataLoad Advanced Configurations](#dataload-advanced-configurations)

**Create the DataLoad object**

```
kubectl create -f dataload.yaml
```

**Check DataLoad's status**

```
kubectl get dataload spark-dataload
```

You shall see something like:
```
NAME             DATASET   PHASE     AGE
spark-dataload   spark     Loading   2m13s
```

In addition, you can get detailed info about the DataLoad object by:

```
kubectl describe dataload spark-dataload
```

and you shall see something like this:

```
Name:         spark-dataload
Namespace:    default
Labels:       <none>
Annotations:  <none>
API Version:  data.fluid.io/v1alpha1
Kind:         DataLoad
...
Spec:
  Dataset:
    Name:       spark
    Namespace:  default
Status:
  Conditions:
  Phase:  Loading
Events:
  Type    Reason              Age   From      Message
  ----    ------              ----  ----      -------
  Normal  DataLoadJobStarted  80s   DataLoad  The DataLoad job spark-dataload-loader-job started
```

The data preloading process may take serveral minutes according to your network environment.

**Wait for the data preloading to complete**

Check its status by running:

```
kubectl get dataload spark-dataload
```

If the data preloading is already done, you should find that the `Phase` of the DataLoad has turned to `Complete`:

```
NAME             DATASET   PHASE      AGE
spark-dataload   spark     Complete   5m17s
```

Now check the status of the dataset again:

```
kubectl get dataset spark
```

You'll find that all data in the remote file storage has already been preloaded into the distributed cache engine:
```
NAME    UFS TOTAL SIZE   CACHED    CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
spark   1.92GiB          1.92GiB   4.00GiB          100.0%              Bound   7m41s
```

## DataLoad Advanced Configurations

Besides the basic data preloading feature showed in the above example, 
with a little bit more configurations, you can enable some advanced features that the DataLoad CRD offers, including:
- Preload data under some specified path only
- Set cache replicas when preloading data
- Sync metadata before preloading data

### Preload data under some specified path only
 
With some extra configurations, DataLoad will only preload data under some specified path (or file) instead of the whole dataset. For example:

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: DataLoad
metadata:
  name: spark-dataload
spec:
  dataset:
    name: spark
    namespace: default
  target:
    - path: /spark/spark-2.4.7
    - path: /spark/spark-3.0.1/pyspark-3.0.1.tar.gz
```

Instead of the whole dataset, the above DataLoad will only preload `/spark/spark-2.4.7` and `/spark/spark-3.0.1/pyspark-3.0.1.tar.gz`

### Set cache replicas when preloading data

When preloading data, you can set cache replicas by simple configuration. For example:

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: DataLoad
metadata:
  name: spark-dataload
spec:
  dataset:
    name: spark
    namespace: default
  target:
    - path: /spark/spark-2.4.7
      replicas: 1
    - path: /spark/spark-3.0.1/pyspark-3.0.1.tar.gz
      replicas: 2
```

The above DataLoad will preload all the files under `/spark/spark-2.4.7` with **only one** cache replicas in the distributed cache engine, while it will
preload the file `/spark/spark-3.0.1/pyspark-3.0.1.tar.gz` with **two** cache replicas.

### Sync metadata before preloading data

Under many circumstances, files in the remote storage system has changed. 
Distributed cache engine like Alluxio needs to sync metadata to update its view of the remote file storage.
It is very common to sync metadata before preloading data from remote file storage, DataLoad CRD offers you a simple way to do this:

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: DataLoad
metadata:
  name: spark-dataload
spec:
  dataset:
    name: spark
    namespace: default
  loadMetadata: true
  target:
    - path: /
      replicas: 1
```

By setting `loadMetadata` to true, you can sync metadata before the data preload starts.

> Notes: Syncing metadata from remote under storage is usually expensive. We do not suggest you enable it if it's not necessary.

## Clean up
```shell
$ kubectl delete -f .
```
