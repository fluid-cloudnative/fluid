## 示例 - Fluid访问AWS S3所需的特殊配置

如果选择AWS S3作为Alluxio的底层存储系统，Alluxio需要进行额外配置，以使得Alluxio能够正常访问挂载的S3存储系统。

本文档展示了如何在Fluid中以声明式的方式完成Alluxio所需的特殊配置。更多信息请参考[在Amazon AWS S3上配置Alluxio](https://docs.alluxio.io/os/user/stable/en/ufs/S3.html)

## 前提条件

- [Fluid](https://github.com/fluid-cloudnative/fluid)
- 配置完成的S3 bucket及有权限访问该bucket的AWS证书

请参考[Fluid安装文档](../guide/install.md) 完成安装

## 运行示例

为了保证安全，Fluid建议使用Secret来配置敏感信息，如`aws.accessKeyId`和`aws.secretKey`。更多关于Secret在fluid中的使用，请参考
[使用Secret配置Dataset敏感信息](./use_encryptoptions.md)

**创建Dataset资源对象**

```yaml
$ cat << EOF > dataset.yaml
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
EOF
```
注意: 不同的云厂商对象存储，region的配置要替换为`alluxio.underfs.s3.endpoint.region=<S3_ENDPOINT_REGION>`,具体详细信息，参考[在Amazon AWS S3上配置Alluxio](https://docs.alluxio.io/os/user/stable/en/ufs/S3.html)

```
$ kubectl create -f dataset.yaml
```

**创建Secret**

在要创建的Secret中，需要写明在上面创建Dataset时需要配置的敏感信息。

```yaml
$ cat<<EOF >mysecret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: mysecret
stringData:
  aws.accessKeyId: <AWS_ACESS_KEY_ID>
  aws.secretKey: <AWS_SECRET_KEY>
EOF
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