# DEMO - Use SharedOptions in Dataset

When creating a dataset in Fluid, sometimes we need to configure some of the same information in multiple 'mounts', In order to reduce duplicate configurations, Fluid provides SharedOptions and SharedEncryptOptions to provide configurations for all mounts. The following takes access to the [Aliyun OSS](https://cn.aliyun.com/product/oss) data set as an example to illustrate how to configure.

## Create a Dataset with SharedOptions

### Create Dataset and Runtime

```shell
$ cat << EOF >> dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: mydata
spec:
  sharedOptions:
    fs.oss.endpoint: <OSS_ENDPOINT>
  sharedEncryptOptions:
    - name: fs.oss.accessKeyId
    valueFrom:
      secretKeyRef:
        name: mysecret
        key: fs.oss.accessKeyId
  - name: fs.oss.accessKeySecret
    valueFrom:
      secretKeyRef:
        name: mysecret
        key: fs.oss.accessKeySecret
  mounts:
  - mountPoint: oss://<OSS_BUCKET>/<OSS_DIRECTORY>/
    name: mydata
---
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: mydata
spec:
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

As you can see, in the above configuration, we configured 'sharedOptions' and' sharedEncryptOptions', which will take effect in all mounts. This eliminates the need to configure in each mount.

> It should be noted that if a key with the same name is configured in `options` and`sharedOptions`, or a key with the same name is configured in `encryptOptions` and`sharedencryptOptions`, the value in `options` will overwrite the corresponding value in`sharedOptions`, and the value in `encryptOptions` will overwrite the corresponding value in`sharedrEncryptOptions`.

> In addition, if a key with the same name is configured in `options` and`encryptOptions`, or a key with the same name is configured in `options` and`sharedencryptOptions`, the creation will fail. The keys of the two should not be the same.
### Create a dataset with the same name as options and encryptOptions
```shell
$ cat << EOF >> dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: mydata
spec:
  sharedOptions:
    fs.oss.endpoint: <OSS_ENDPOINT>
  sharedEncryptOptions:
    - name: fs.oss.accessKeyId
    valueFrom:
      secretKeyRef:
        name: mysecret
        key: fs.oss.accessKeyId
  - name: fs.oss.accessKeySecret
    valueFrom:
      secretKeyRef:
        name: mysecret
        key: fs.oss.accessKeySecret
  mounts:
  - mountPoint: oss://<OSS_BUCKET>/<OSS_DIRECTORY>/
    name: mydata
    sharedOptions:
      fs.oss.endpoint: <OSS_ENDPOINT>
    sharedEncryptOptions:
        - name: fs.oss.accessKeyId
        valueFrom:
          secretKeyRef:
            name: mysecret
            key: fs.oss.accessKeyId
      - name: fs.oss.accessKeySecret
        valueFrom:
          secretKeyRef:
            name: mysecret
            key: fs.oss.accessKeySecret
```

A key with the same name is configured in `options` and `'`encryptOptions`, which will cause failure and error.

> In short, `options` and`sharedOptions`belong to the same type of configuration, and there is an override relationship between them. However, `options` and`encryptOption` are not of the same type. There should be no overlap relationship between them, so the same key should not be configured.

### Create Secret

In the Secret to be created, you need to specify the sensitive information that needs to be configured in the above Dataset.

```shell
$ cat<<EOF >mysecret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: mysecret
stringData:
  fs.oss.accessKeyId: <OSS_ACCESS_KEY_ID>
  fs.oss.accessKeySecret: <OSS_ACCESS_KEY_SECRET>
EOF
```

As you can see, the specific contents of `fs.oss.accessKeySecret` and `fs.oss.accessKeyId` are written in Secret, and Dataset reads the corresponding value by looking for the Secret and key according to its configuration, instead of reading them in its configuration directly. So the security of some data is guaranteed.

