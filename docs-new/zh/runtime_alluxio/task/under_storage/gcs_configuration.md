## 示例 - Fluid访问 Google Cloud Storage (GCS) 所需的特殊配置

如果选择 Google Cloud Storage (GCS) 作为Alluxio的底层存储系统，Alluxio需要进行额外配置，以使得Alluxio能够正常访问挂载的GCS存储系统。

本文档展示了如何在Fluid中以声明式的方式完成Alluxio所需的特殊配置。更多信息请参考[Alluxio集成 Google Cloud Storage (GCS) 作为底层存储](https://docs.alluxio.io/os/user/stable/cn/ufs/GCS.html)

## 前提条件

- [Fluid](https://github.com/fluid-cloudnative/fluid)
- 配置完成的 [GCS bucket](https://cloud.google.com/storage/docs/creating-buckets)
- 有权限访问该GCS bucket的[Google Application Credentials](https://cloud.google.com/docs/authentication/getting-started)

请参考[Fluid安装文档](../guide/install.md) 完成安装

## 运行示例

在这个例子中，我们会使用 `GCS UFS version 2`。 它会使用 `Google Application Credentials` 来访问 GCS。这是Google推荐的方法。

在 Alluxio 中访问 GCS 的方法与其他存储系统的访问不同，比如S3。你需要挂载你的 `Google Application Credentials` 在你的Alluxio master 还有 workers 上。然后提供 `Google Application Credentials` 的路径在你的 Alluxio 配置中。

**创建Secret**

首先你要准备好你的 `Google Application Credentials`, 然后运行一下的命令创建一个叫 `gcscreds` 的 secret， 其中 key 是 `gcloud-application-credentials.json`，值是你的 json 文件内容。

```sh
kubectl create secret generic gcscreds \
    --from-file=gcloud-application-credentials.json=<name of your google application credentials>.json
```


**创建AlluxioRuntime资源对象**

这里你将创建你的 Alluxio Runtime, 并且挂载上一步中创建的 Secret 在 Alluxio master 还有 workers 上。路径必须相同。

```yaml
$ cat << EOF > runtime.yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: gcs-data
spec:
  ...
  volumes:
    - name: secret-volume
      secret:
        secretName: gcscreds # secret name you created
  master:
    volumeMounts:
      - name: secret-volume
        readOnly: true
        mountPath: "/var/secrets"
  worker:
    volumeMounts:
      - name: secret-volume
        readOnly: true
        mountPath: "/var/secrets"
EOF
```

创建你的 Runtime

```
$ kubectl create -f runtime.yaml
```

**创建Dataset资源对象**

这里你在 `mountPoint` 指定你的 GCS bucket 还有文件目录。然后在 `options` 中提供你挂载的 `google application credentials` 路径。

```yaml
$ cat << EOF > dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: gcs-data
spec:
  mounts:
    - mountPoint: gs://<bucket-name>/<path-to-data>
      name: gcs
      options:
        fs.gcs.credential.path: /var/secrets/gcloud-application-credentials.json
  accessModes:
    - ReadOnlyMany
EOF
```

创建你的 Dataset

```
$ kubectl create -f dataset.yaml
```

`dataset.yaml`中指定的bucket将被挂载到Alluxio中`/gcs`目录下。