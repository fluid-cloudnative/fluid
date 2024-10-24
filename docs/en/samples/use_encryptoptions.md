# DEMO - Use Secret to configure Dataset sensitive information

When creating a Dataset in Fluid, sometimes we need to configure some sensitive information in the `mounts`. To ensure security, Fluid provides the ability to configure these sensitive information using Secret. The following takes access to the [Aliyun OSS](https://cn.aliyun.com/product/oss) data set as an example to illustrate how to configure.

## Deploy Dataset with sensitive information

### Create Dataset and Runtime

```shell
$ cat << EOF >> dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: mydata
spec:
  mounts:
  - mountPoint: oss://<OSS_BUCKET>/<OSS_DIRECTORY>/
    name: mydata
    options:
      fs.oss.endpoint: <OSS_ENDPOINT>
    encryptOptions:
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

As you can see, in the above configuration, unlike the direct configuration of `fs.oss.endpoint`, we changed the configuration of `fs.oss.accessKeyId` and `fs.oss.accessKeySecret` to read from Secret to ensure safety.

> It should be noted that if the same key is configured both in `options` and `encryptOptions`, then the value in `encryptOptions` will override the corresponding value in `options`.

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

