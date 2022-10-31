# 示例 - 用Fluid加速不同的底层存储

选择不同的**底层存储服务**作为Alluxio的底层存储系统，需要Alluxio进行额外的配置，以使得Alluxio能够正常访问存储。

本文档展示了如何在Fluid中以声明式的方式完成Alluxio所需的特殊配置，以访问S3、HDFS、Ceph S3、PV、Minio等不同的存储服务。更多信息请参考[Alluxio集成Amazon AWS S3作为底层存储 - Alluxio v2.8.1 (stable) Documentation](https://docs.alluxio.io/os/user/stable/cn/ufs/S3.html)

## 前提条件

- 在运行该示例之前，请参考[安装文档](https://github.com/fluid-cloudnative/fluid/blob/master/docs/zh/userguide/install.md)完成安装，并检查Fluid各组件正常运行：

    ~~~ shell
    $ kubectl get pod -n fluid-system
    NAME                                  READY   STATUS    RESTARTS   AGE
    alluxioruntime-controller-5b64fdbbb-84pc6   1/1     Running   0          8h
    csi-nodeplugin-fluid-fwgjh                  2/2     Running   0          8h
    csi-nodeplugin-fluid-ll8bq                  2/2     Running   0          8h
    dataset-controller-5b7848dbbb-n44dj         1/1     Running   0          8h
    ~~~

- 可访问的底层存储服务；

## 配置方式

**创建Dataset资源对象**

``` yaml
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

在这里，我们将要创建一个kind为`Dataset`的资源对象(Resource object)。`Dataset`是Fluid所定义的一个Custom Resource Definition(CRD)，该CRD被用来告知Fluid在哪里可以找到你所需要的数据。

Fluid将该CRD对象中定义的`mountPoint`属性挂载到Alluxio之上，因此该属性可以是任何合法的能够被Alluxio识别的UFS地址。

用户可以根据需要修改`spec.mounts`字段，一般设置为底层存储的访问路径，例如：

* HDFS：`- mountPoint: hdfs://<namenode>:<port>`；

* AWS S3：

    ~~~ yaml
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
    ~~~

* PVC：`- mountPoint: pvc://nfs-imagenet`；

* 本地目录：`- mountPoint: local:///mnt/nfs-imagenet`；

* GCS：`- mountPoint: gs://<bucket-name>/<path-to-data>`

用户需要在`spec.mounts.mountPoint`指定存储位置；在`spec.mounts.options`指定访问存储所需的region、endpoint以及密钥等（更多选项参考[配置项列表 - Alluxio v2.8.1 (stable) Documentation](https://docs.alluxio.io/os/user/stable/cn/reference/Properties-List.html)）



**创建AlluxioRuntime资源对象**

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

**AlluxioRuntime**无需为不同的底层存储进行额外的配置（除了HDFS，具体参考[HDFS](hdfs_configuration.md)）。

```
$ kubectl create -f runtime.yaml
```

至此, Alluxio能够根据用户指定的配置文件正常地访问不同类型的底层存储。



## 具体示例

* [AWS S3](s3_configuration.md)
* [HDFS](hdfs_configuration.md)
* [PVC](accelerate_pvc.md)
* [Minio](accelerate_s3_minio.md)
* [GCS](gcs_configuration.md)
