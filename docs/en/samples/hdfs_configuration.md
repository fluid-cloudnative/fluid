## DEMO - Special Configuration for Fluid to Access HDFS

If HDFS is the underlying storage system of Alluxio and the HDFS cluster has a special configuration, Alluxio needs the same extra configuration to access the mounted HDFS cluster.

This document shows how to give Alluxio that configuration declaratively in Fluid. For more information, see [Configuring Alluxio with HDFS](https://docs.alluxio.io/os/user/stable/en/ufs/HDFS.html).

## Prerequisites

- [Fluid](https://github.com/fluid-cloudnative/fluid)
- An HDFS cluster reachable from the Kubernetes network

Please refer to the [Fluid installation guide](../userguide/install.md) to install Fluid.

## Run the Demo

**Create a ConfigMap from the HDFS configuration files**
```
$ kubectl create configmap <configmap-name> --from-file=/path/to/hdfs-site.xml --from-file=/path/to/core-site.xml
```

**Create the Dataset**

```yaml
$ cat << EOF > dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: my-hdfs
spec:
  mounts:
    - mountPoint: hdfs://<namenode>:<port>
      name: hdfs
EOF
```

```
$ kubectl create -f dataset.yaml
```

**Create the AlluxioRuntime**

```yaml
$ cat << EOF > runtime.yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: my-hdfs
spec:
  ...
  hadoopConfig: <configmap-name>
  ...
EOF
```

`configmap-name` must be the ConfigMap created above, in the same namespace as the AlluxioRuntime. It must contain at least one of the keys `hdfs-site.xml` and `core-site.xml`. Fluid checks the ConfigMap, picks the `hdfs-site.xml` and `core-site.xml` entries and mounts them as files into the Alluxio pods.

```
$ kubectl create -f runtime.yaml
```

**Check the HDFS configuration files in Alluxio**

Once the Alluxio cluster is up, the HDFS configuration files from the ConfigMap are mounted in the `/hdfs-config` directory of the containers in the alluxio-master and alluxio-worker pods:

```
$ kubectl exec my-hdfs-master-0 -- ls /hdfs-config
hdfs-site.xml  core-site.xml
```

Alluxio can now access the HDFS cluster with the configuration files you provided.
