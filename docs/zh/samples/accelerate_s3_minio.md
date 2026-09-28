# 示例 - 加速Minio文件访问

开启一个单机版的本地Minio作为远程的S3服务，这个示例只做举例使用，不适用于生产环境。

### 开启minio

```shell
# MinIO 官方已不再发布镜像，这里使用 ACK 镜像仓库中的副本
docker run -ti -p 9000:9000 --name minio registry-cn-hongkong.ack.aliyuncs.com/acs/minio:RELEASE.2022-10-24T18-35-07Z-update server /data
```
```
Formatting 1st pool, 1 set(s), 1 drives per set.
WARNING: Host local has more than 0 drives of set. A host failure will result in data becoming unavailable.
WARNING: Detected default credentials 'minioadmin:minioadmin', we recommend that you change these values with 'MINIO_ROOT_USER' and 'MINIO_ROOT_PASSWORD' environment variables
MinIO Object Storage Server
Copyright: 2015-2022 MinIO, Inc.
License: GNU AGPLv3 <https://www.gnu.org/licenses/agpl-3.0.html>
Version: RELEASE.2022-10-24T18-35-07Z (go1.19.2 linux/amd64)

Status:         1 Online, 0 Offline.
API: http://172.17.0.2:9000  http://127.0.0.1:9000
Console: http://172.17.0.2:41919 http://127.0.0.1:41919

Documentation: https://min.io/docs/minio/linux/index.html
```

### 生成minio数据
```shell
# MinIO 官方已不再发布 mc，任意 S3 客户端均可，例如 7.75 及以上版本的 curl（--aws-sigv4）
# 创建一个新的 bucket
$ curl -sSf -X PUT --aws-sigv4 aws:amz:us-east-1:s3 --user minioadmin:minioadmin http://127.0.0.1:9000/fluid
# 本地 fluid 目录中有一些 PDF 文件
$ for f in fluid/*; do curl -sSf -T "$f" --aws-sigv4 aws:amz:us-east-1:s3 --user minioadmin:minioadmin http://127.0.0.1:9000/fluid/; done
```

### dataset.yaml
```yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: demo
spec:
  mounts:
    - mountPoint: s3://spark/fluid-data
      name: spark
      options:
        alluxio.underfs.s3.endpoint: http://{{$demo-minio-addr}}:9000
        alluxio.underfs.s3.disable.dns.buckets: "true"
        alluxio.underfs.s3.inherit.acl: "false"
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
### secret.yaml
创建minio的accessKeyId和accessKey
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mysecret
stringData:
  aws.accessKeyId: minioadmin
  aws.secretKey: minioadmin
```
### runtime.yaml
```yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: demo
spec:
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 20M
        high: "0.95"
        low: "0.7"
```
### pod.yaml
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: demo-app
spec:
  containers:
    - name: demo
      image: nginx:latest
      volumeMounts:
        - mountPath: /data
          name: demo
  volumes:
    - name: demo
      persistentVolumeClaim:
        claimName: demo
```

### 检查数据
```shell
$ k apply -f dataset.yaml
$ k apply -f secret.yaml
$ k apply -f runtime.yaml
$ k apply -f pod.yaml
$ k exec -ti demo-app sh 
$ ls /data
data-mesh-in-practice-how-europes-leading-online-platform-for-fashion-goes-beyond-the-data-lake-iteblog.com.pdf
data-science-across-data-sources-with-apache-arrow-iteblog.com.pdf
from-hdfs-to-s3-migrate-pinterest-apache-spark-clusters-iteblog.com.pdf
running-apache-spark-jobs-using-kubernetes-iteblog.com.pdf
running-apache-spark-on-kubernetes-best-practices-and-pitfalls-iteblog.com.pdf
scaling-data-and-ml-with-apache-spark-and-feast-iteblog.com.pdf
using-ai-to-support-proliferating-merchant-changes-iteblog.com.pdf
```