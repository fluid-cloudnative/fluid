# JuiceFS 开源环境搭建步骤

为了能够快速验证Fluid + JuiceFS，可以通过以下步骤可以在Kubernetes快速搭建开源版本的JuiceFS环境。该步骤仅用于功能验证，并没有任何调优，不推荐用于生产环境。

以下示例以minIO作为后端对象存储，以Redis作为元数据服务。

1.创建Minio服务

```yaml
$ cat << EOF > minio.yaml
apiVersion: v1
kind: Service
metadata:
  name: minio
spec:
  type: ClusterIP
  ports:
    - port: 9000
      targetPort: 9000
      protocol: TCP
  selector:
    app: minio
---
apiVersion: apps/v1 #  for k8s versions before 1.9.0 use apps/v1beta2  and before 1.8.0 use extensions/v1beta1
kind: Deployment
metadata:
  # This name uniquely identifies the Deployment
  name: minio
spec:
  selector:
    matchLabels:
      app: minio
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        # Label is used as selector in the service.
        app: minio
    spec:
      containers:
      - name: minio
        # Pulls the default Minio image from Docker Hub
        image: minio/minio
        args:
        - server
        - /storage
        env:
        # Minio access key and secret key
        - name: MINIO_ROOT_USER
          value: "minioadmin"
        - name: MINIO_ROOT_PASSWORD
          value: "minioadmin"
        ports:
        - containerPort: 9000
          hostPort: 9000
EOF
```

2.执行以下命令创建Deployment和Service，

```bash
$ kubectl create -f minio.yaml
service/minio created
deployment.apps/minio created
````

3.查看运行结果

```bash
$ kubectl  get deploy minio
NAME    READY   UP-TO-DATE   AVAILABLE   AGE
minio   0/1     1            0           40s
$ kubectl  get svc minio
NAME    TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
minio   ClusterIP   172.16.159.15   <none>        9000/TCP   77s
```

4.创建Redis服务

```yaml
$ cat << EOF > redis.yaml
---
apiVersion: v1
kind: Service
metadata:
  name: redis
  labels:
    app: redis
spec:
  type: ClusterIP
  ports:
    - name: redis
      port: 6379
  selector:
    app: redis
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  labels:
    app: redis
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
        - name: redis
          image: redis
          ports:
            - containerPort: 6379
EOF
```

此时开源Juicefs依赖的基础环境已经准备完毕，可以看到此时对应的配置

| Name                             | Value                                      | Description                                |
|----------------------------------|--------------------------------------------|--------------------------------------------|
| `metaurl`                        | redis://redis:6379/0                       | 元数据服务的访问 URL (比如 Redis)。更多信息参考[文档](https://juicefs.com/docs/zh/community/databases_for_metadata/)                  | 
| `access-key`                     | minoadmin                                  | 对象存储的 access key                        |
| `access-secret`                  | minoadmin                                  | 对象存储的 access secret                     |
| `storage type`                   | minio                                      | 对象存储的类型                                |



