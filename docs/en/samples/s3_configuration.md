## DEMO - The special configurations required for Fluid access AWS S3

If AWS S3 is selected as the underlying storage system of Alluxio, additional configuration is required for Alluxio to properly access the mounted S3 storage system.

This document shows how to do the special configuration required for Alluxio in Fluid in a declarative manner. For more information, see [Configuring Alluxio on Amazon AWS S3](https://docs.alluxio.io/os/user/stable/en/ufs/S3.html).

## Prerequisites

- [Fluid](https://github.com/fluid-cloudnative/fluid)
- The S3 bucket has been configured and the AWS certificate that has permission to access the bucket.

Please refer to the[Fluid installation documentation](https://github.com/fluid-cloudnative/fluid/blob/master/docs/zh/userguide/install.md) to complete the installation.

## Run the example

For security, Fluid recommends using Secret to configure sensitive information such as`aws.accessKeyId` and `aws.secretKey`。For more information about Secret's use in Fluid, see[Use Secret to configure sensitive Dataset information](https://github.com/fluid-cloudnative/fluid/blob/master/docs/en/samples/use_encryptoptions.md)

**Create Dataset Resource Object**

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
Note: For object storage of different cloud vendors, the region configuration must be replaced with`alluxio.underfs.s3.endpoint.region=<S3_ENDPOINT_REGION>`,For details, see [Configuring Alluxio on Amazon AWS S3](https://docs.alluxio.io/os/user/stable/en/ufs/S3.html)

```
$ kubectl create -f dataset.yaml
```

**Create Secret**

In the Secret to be created, specify the sensitive information that needs to be configured when the Dataset is created above.

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

**Create AlluxioRuntime Resource Object**

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

The bucket specified in 'dataset.yaml' will be mounted to the '/s3' directory in Alluxio.