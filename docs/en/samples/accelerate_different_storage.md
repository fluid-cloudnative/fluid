# DEMO - Accelerate different Stowith Fluid

We need to add additional configuration to Alluxio for accessing the storage normally, when using different **underlying storage services** as the underlying storage system of Alluxio.

This document shows how to declaratively complete the special configuration required by Alluxio in Fluid to access different storage services such as S3, HDFS, Ceph S3, PV, and MinIo. Please visit [Amazon AWS S3 - Alluxio v2.8.1 (stable) Documentation](https://docs.alluxio.io/os/user/stable/en/ufs/S3.html) for more  information.

## Prerequisites

- Before everything we are going to do, please refer to [Installation Guide](../userguide/install.md) to install Fluid on your Kubernetes Cluster, and make sure all the components used by Fluid are ready like this:

    ~~~ shell
    $ kubectl get pod -n fluid-system
    NAME                                  READY   STATUS    RESTARTS   AGE
    alluxioruntime-controller-5b64fdbbb-84pc6   1/1     Running   0          8h
    csi-nodeplugin-fluid-fwgjh                  2/2     Running   0          8h
    csi-nodeplugin-fluid-ll8bq                  2/2     Running   0          8h
    dataset-controller-5b7848dbbb-n44dj         1/1     Running   0          8h
    ~~~

- The Storage which can be visited.

## Configuration

**Create Dataset Resource Object**

``` yaml
$ cat << EOF > dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: my-hdfs
spec:
  mounts:
    - mountPoint: hdfs://<namenode>:<port>/path1
      name: hdfs-file1
    - mountPoint: hdfs://<namenode>:<port>/path2
      name: hdfs-file2
EOF
```

```
$ kubectl create -f dataset.yaml
```


Fluid mounts the `mountPoint` attribute defined in the CRD object to Alluxio, so this attribute can be any legal **UFS address** that can be recognized by Alluxio.

In addition, multiple `mountPoint` can be set for each `Dataset`, so that you can mount all of the `mountPoint` to the specified directory. At the same time, you can also set `subPath` when mounting `PVC` to specify a `mountPoint` or its subdirectory set in the mount dataset. For example, in the above example, when mounting `PVC` to your pod, you can set `subPath: hdfs-file1`, so that only the `hdfs://<namenode>:<port>/path1` directory will be mounted.

You can modify the `spec.mounts` field as required. It is generally set to the access path of the underlying storage, for example:

* HDFS：`- mountPoint: hdfs://<namenode>:<port>`；

* AWS S3：

    ``` yaml
    apiVersion: data.fluid.io/v1alpha1
    kind: Dataset
    metadata:
      name: my-s3
    spec:
      mounts:
        - mountPoint: s3://<bucket-name>/<path-to-data>/
          name: s3
          options:
            alluxio.underfs.s3.region: <s3-bucket-region>
            alluxio.underfs.s3.endpoint: <s3-endpoint>
            encryptOptions:
            - name: aws.accessKeyId
              valueFrom:
                secretKeyRef:
                  name: mysecret
                  key: aws.accessKeyId
            - name: aws.secretKey
              valueFrom:
                secretKeyRef:
                  name: mysecret
                  key: aws.secretKey
    ```

* PVC：`- mountPoint: pvc://nfs-imagenet`；

* local path：`- mountPoint: local:///mnt/nfs-imagenet`；

* GCS：`- mountPoint: gs://<bucket-name>/<path-to-data>`

You need to specify the storage location in `spec.mounts.mountPoint`; In `spec.mounts.options`, specify the region, endpoint, and key required to access the storage（Refer to [List of Configuration Properties - Alluxio v2.8.1 (stable) Documentation](https://docs.alluxio.io/os/user/stable/en/reference/Properties-List.html) for more options）



**Create AlluxioRuntime Resource Object**

``` yaml
$ cat << EOF > runtime.yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: my-hdfs
spec:
  ...
EOF
```

No additional configuration required in **AlluxioRuntime** for different underlying storage(Except HDFS，Please refer [HDFS](accelerate_data_accessing_by_hdfs.md)).

```
$ kubectl create -f runtime.yaml
```

So far, Alluxio can normally access different types of underlying storage according to the user specified configuration file.



## Examples

* [AWS S3](s3_configuration.md)
* [HDFS](accelerate_data_accessing_by_hdfs.md)
* [PVC](accelerate_pvc.md)
* [Minio](accelerate_s3_minio.md)
* [GCS](gcs_configuration.md)
