## 示例 - Fluid访问AWS S3所需的特殊配置

如果选择AWS S3作为Alluxio的底层存储系统,Alluxio需要进行额外配置,以使得Alluxio能够正常访问挂载的S3存储系统.

本文档展示了如何在Fluid中以声明式的方式完成Alluxio所需的特殊配置. 更多信息请参考[在Amazon AWS S3上配置Alluxio](https://docs.alluxio.io/os/user/stable/cn/ufs/S3.html)

## 前提条件

- [Fluid](https://github.com/fluid-cloudnative/fluid)
- 配置完成的S3 bucket及有权限访问该bucket的AWS证书

请参考[Fluid安装文档](https://github.com/fluid-cloudnative/fluid/blob/master/docs/zh/userguide/install.md)完成安装

## 运行示例

**创建Dataset资源对象**

```yaml
$ cat << EOF > dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: my-s3
spec:
  mounts:
    - mountPoint: s3://<bucket-name>/<path-to-data>
      name: s3
      options:
        aws.accessKeyId: XXXXXXXXXXXXX
        aws.secretKey: XXXXXXXXXXXXXXXXXXXXXXX
EOF
```

```
$ kubectl create -f dataset.yaml
```

**创建AlluxioRuntime资源对象**

```yaml
$ cat << EOF > runtime.yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: my-s3
spec:
  ...
EOF
```

```
$ kubectl create -f runtime.yaml
```

`dataset.yaml`中指定的bucket将被挂载到Alluxio中`/s3`目录下。
s