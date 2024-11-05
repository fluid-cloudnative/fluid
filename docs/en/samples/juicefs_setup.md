# Steps to build Juicefs open source environment

In order to get started with Fluid + JuiceFS, you can quickly build community version of JuiceFS environment in Kubernetes by following the steps below. This environment is only used for functional verification, without any tuning, and is not recommended for production environments.

The following example uses MinIO as the back-end object storage and Redis as the metadata service.

1. Create Minio Service

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

2. Create Deployment and Service

```bash
$ kubectl create -f minio.yaml
service/minio created
deployment.apps/minio created
````

3. Check the result

```bash
$ kubectl  get deploy minio
NAME    READY   UP-TO-DATE   AVAILABLE   AGE
minio   0/1     1            0           40s
$ kubectl  get svc minio
NAME    TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
minio   ClusterIP   172.16.159.15   <none>        9000/TCP   77s
```

4. Create Redis service

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

At this point, the basic environment of open source Juicefs has been prepared. You can see the corresponding configuration at this time.

| Name                             | Value                                      | Description                                |
|----------------------------------|--------------------------------------------|--------------------------------------------|
| `metaurl`                        | redis://redis:6379/0                       | Access URL of metadata service (such as Redis). For more information, refer to [this document](https://juicefs.com/docs/community/databases_for_metadata/)                  | 
| `access-key`                     | minoadmin                                  |Access key for object storage                       |
| `access-secret`                  | minoadmin                                  |Access secret for object storage                     |
| `storage type`                   | minio                                      | Type of object storage                                |



