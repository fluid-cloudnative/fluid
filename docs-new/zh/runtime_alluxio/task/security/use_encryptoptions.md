# 示例 - 使用Secret配置Dataset敏感信息

在Fluid中创建Dataset时，有时候我们需要在`mounts`中配置一些敏感信息，为了保证安全，Fluid提供使用Secret来配置这些敏感信息的能力。下面以访问[阿里云OSS](https://cn.aliyun.com/product/oss)数据集为例说明如何配置。

## 创建带敏感信息的Dataset

### 创建Dataset和Runtime

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

可以看到，在上面的配置中，与直接配置`fs.oss.endpoint`不同，我们把`fs.oss.accessKeyId`以及`fs.oss.accessKeySecret`的配置改为从Secret中读取，以此来保障安全性。

> 需要注意的是，如果在`options`和`encryptOptions`中配置了同名的键，例如都有`fs.oss.accessKeyId`的配置，那么`encryptOptions`中的值会覆盖`options`中对应的值。

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