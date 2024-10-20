# 示例 - Dataset公共option使用

在Fluid中创建Dataset时，有时候我们需要在多个`mounts`中配置一些相同的信息，为了减少重复配置，Fluid提供SharedOptions和SharedEncryptOptions来提供所有mounts共享的配置。下面以访问[阿里云OSS](https://cn.aliyun.com/product/oss)数据集为例说明如何配置。

## 创建带SharedOptions的Dataset

### 创建Dataset和Runtime

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

可以看到，在上面的配置中，我们配置了`sharedOptions`和`sharedEncryptOptions`，这两个会在所有mount中生效。这样就无须在每个mount中配置。

> 需要注意的是，如果在`options`和 `sharedOptions`中配置了同名的键，或者`encryptOptions`和`sharedencryptOptions`中配置了同名的键，那么`options`中的值会覆盖`sharedOptions`中对应的值，`encryptOptions`中的值会覆盖`sharedencryptOptions`中对应的值。

> 另外，如果在`options`和 `encryptOptions`中配置了同名的键，或者`options`和`sharedencryptOptions`中配置了同名的键，那么会创建失败，两者的key不应该相同。

### 创建options和encryptOptions同名的Dataset
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

这样`options`和 `encryptOptions`中配置了同名的键，会导致失败报错。

> 简单来说，`options`和 `sharedOptions`属于同一类配置，两者存在覆盖关系。而`options`和`encryptOption`不属于同一类，两者不应存在覆盖关系，所以不应配置相同的键。

### 创建Secret

在要创建的Secret中，需要写明在上面创建Dataset时需要配置的敏感信息。

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

可以看到，`fs.oss.accessKeySecret`和`fs.oss.accessKeyId`的具体内容写在Secret中，Dataset通过寻找配置中同名的Secret和key来读取对应的值，而不再是在Dataset直接写明，这样就保证了一些数据的安全性。