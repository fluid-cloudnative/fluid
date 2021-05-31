# DEMO - Speed Up Accessing Minio Files

Start a standalone Minio locally as a remote S3 service. This example is for demonstration purposes only, not production

### start minio demo

```shell
docker run -ti -p 9000:9000 --name minio minio/minio server /data
```
```
Endpoint: http://172.17.0.8:9000  http://127.0.0.1:9000 
RootUser: minioadmin 
RootPass: minioadmin 

Browser Access:
   http://172.17.0.8:9000  http://127.0.0.1:9000

Command-line Access: https://docs.min.io/docs/minio-client-quickstart-guide
   $ mc alias set myminio http://172.17.0.8:9000 minioadmin minioadmin

Object API (Amazon S3 compatible):
   Go:         https://docs.min.io/docs/golang-client-quickstart-guide
   Java:       https://docs.min.io/docs/java-client-quickstart-guide
   Python:     https://docs.min.io/docs/python-client-quickstart-guide
   JavaScript: https://docs.min.io/docs/javascript-client-quickstart-guide
   .NET:       https://docs.min.io/docs/dotnet-client-quickstart-guide
Detected default credentials 'minioadmin:minioadmin', please change the credentials immediately using 'MINIO_ROOT_USER' and 'MINIO_ROOT_PASSWORD'
IAM initialization complete
```

### mock minio data
```shell
# create a new bucket
$ mc mb myminio/fluid
# there are some PDFs in my local folder fluid
$ mc cp fluid/* myminio/fluid/
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
create minio accessKeyId and accessKey
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

### check the data
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