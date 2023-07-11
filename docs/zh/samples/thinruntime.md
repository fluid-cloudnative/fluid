# 示例 - 如何使用ThinRuntime支持通用存储系统接入Fluid

除了Fluid原生集成的存储/缓存系统外，Fluid提供了ThinRuntime CRD。ThinRuntime CRD允许用户描述任何自定义的存储系统，并将其对接到Fluid中。本文档以[Minio](https://github.com/minio/minio)为例，展示如何通过Fluid管理并访问Minio存储中的数据。

## 前提条件

- [Fluid](https://github.com/fluid-cloudnative/fluid)版本需至少为0.9.0

请参考[Fluid安装文档](https://github.com/fluid-cloudnative/fluid/blob/master/docs/zh/userguide/install.md)完成安装

## 运行示例

**部署Minio存储到集群中**

有如下YAML文件`minio.yaml`：
```
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
        image: bitnami/minio
        env:
        # Minio access key and secret key
        - name: MINIO_ROOT_USER
          value: "minioadmin"
        - name: MINIO_ROOT_PASSWORD
          value: "minioadmin"
        - name: MINIO_DEFAULT_BUCKETS
          value: "my-first-bucket:public"
        ports:
        - containerPort: 9000
          hostPort: 9000
```

部署上述资源到Kubernetes集群：
```
$ kubectl create -f minio.yaml
```
部署成功后，Kubernetes集群内的其他Pod即可通过`http://minio:9000`的Minio API端点访问Minio存储系统中的数据。上述YAML配置中，我们设置Minio的用户名与密码均为`minioadmin`，并在启动Minio存储时默认创建一个名为`my-first-bucket`的存储桶，在接下来的示例中，我们将会访问`my-first-bucket`这个存储桶中的数据。在执行以下步骤前，首先执行以下命令，在`my-first-bucket`中存储示例文件：

```
$ kubectl exec -it minio-69c555f4cf-np59j -- bash -c "echo fluid-minio-test > testfile"

$ kubectl exec -it minio-69c555f4cf-np59j -- bash -c "mc cp ./testfile local/my-first-bucket/" 

$ kubectl exec -it  minio-69c555f4cf-np59j -- bash -c "mc cat local/my-first-bucket/testfile"
fluid-minio-test
```

**准备包含Minio Fuse客户端的容器镜像**

Fluid将会把ThinRuntime中Fuse所需的运行参数、Dataset中描述数据路径的挂载点等参数传入到ThinRuntime Fuse Pod容器中。在容器内部，用户需要执行参数解析脚本，并将解析完的运行时参数传递给Fuse客户端程序，由客户端程序完成Fuse文件系统在容器内的挂载。

因此，使用ThinRuntime CRD描述存储系统时，需要使用**特制的容器镜像**，镜像中需要包括以下两个程序：
- Fuse客户端程序
- Fuse客户端程序所需的运行时参数解析脚本

对于Fuse客户端程序，在本示例中选择S3协议兼容的[goofys](https://github.com/kahing/goofys)客户端连接并挂载minio存储系统。

对于运行时所需的参数解析脚本，定义如下python脚本`fluid-config-parse.py`：

```python
import json

with open("/etc/fluid/config.json", "r") as f:
    lines = f.readlines()

rawStr = lines[0]
print(rawStr)


script = """
#!/bin/sh
set -ex
export AWS_ACCESS_KEY_ID=`cat $akId`
export AWS_SECRET_ACCESS_KEY=`cat $akSecret`

mkdir -p $targetPath

exec goofys -f --endpoint "$url" "$bucket" $targetPath
"""

obj = json.loads(rawStr)

with open("mount-minio.sh", "w") as f:
    f.write("targetPath=\"%s\"\n" % obj['targetPath'])
    f.write("url=\"%s\"\n" % obj['mounts'][0]['options']['minio-url'])
    if obj['mounts'][0]['mountPoint'].startswith("minio://"):
      f.write("bucket=\"%s\"\n" % obj['mounts'][0]['mountPoint'][len("minio://"):])
    else:
      f.write("bucket=\"%s\"\n" % obj['mounts'][0]['mountPoint'])
    f.write("akId=\"%s\"\n" % obj['mounts'][0]['options']['minio-access-key'])
    f.write("akSecret=\"%s\"\n" % obj['mounts'][0]['options']['minio-access-secret'])

    f.write(script)
```

上述python脚本按以下步骤执行：
1. 读取`/etc/fluid/config.json`文件中的json字符串，Fluid会将Fuse客户端挂载所需的参数存储并挂载到Fuse容器的`/etc/fluid/config.json`文件。
2. 解析json字符串，从中提取Fuse客户端挂载所需的参数。例如，上述示例中的`url`、`bucket`、`minio-access-key`、`minio-access-secret`等参数。其中, `url`、`bucket`等参数为实际值，`minio-access-key`、`minio-access-secret`等涉及敏感信息的参数则为存储了实际值的文件路径，因此，对于此类参数需要额外执行文件读取操作（例如：上述实例中的"export AWS_ACCESS_KEY_ID=\`cat $akId\`"）。
3. 提取出所需参数后，输出挂载脚本到文件`mount-minio.sh`

接着，使用如下Dockerfile制作镜像，这里我们直接选择包含`goofys`客户端程序的镜像（i.e. `cloudposse/goofys`）作为Dockerfile的基镜像：

```
FROM cloudposse/goofys

RUN apk add python3

COPY ./fluid-config-parse.py /fluid-config-parse.py
```

使用以下命令构建并推送镜像到镜像仓库：

```
$ IMG_REPO=<your image repo>

$ docker build -t $IMG_REPO/fluid-minio-goofys:demo .

$ docker push $IMG_REPO/fluid-minio-goofys:demo
```

**创建ThinRuntimeProfile**

在创建Fluid Dataset和ThinRuntime CR以挂载Minio存储系统前，首先需要创建ThinRuntimeProfile CR资源。ThinRuntimeProfile是一种Cluster-level的 Fluid CRD资源，它描述了一类需要与Fluid对接的存储系统的基础配置（例如：容器镜像信息、Pod Spec描述信息等）。集群管理员需提前在集群中定义若干ThinRuntimeProfile CR资源，在这之后，集群用户需要显示声明使用哪一个ThinRuntimeProfile CR来创建ThinRuntime，从而完成对应存储系统的挂载。

下面展示了一个Minio存储系统的ThinRuntimeProfile CR示例（`profile.yaml`）:

```
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntimeProfile
metadata:
  name: minio-profile
spec:
  fileSystemType: fuse
  fuse:
    image: $IMG_REPO/fluid-minio-goofys
    imageTag: demo
    imagePullPolicy: IfNotPresent
    command:
    - sh
    - -c
    - "python3 /fluid-config-parse.py && chmod u+x ./mount-minio.sh && ./mount-minio.sh"
```
在上述CR示例中：
- `fileSystemType`描述了ThinRuntime Fuse所挂载的文件系统类型(fsType)。需要根据使用的存储系统Fuse客户端程序填写，例如，goofys挂载的挂载点fsType为fuse, s3fs挂载的挂载点fsType为fuse.s3fs）
- `fuse`描述了ThinRuntime Fuse的容器信息，包括镜像信息（`image、imageTag、imagePullPolicy`）以及容器启动命令(`command`)等。


创建ThinRuntimeProfile CR `minio-profile`到Kubernetes集群。
```
$ kubectl create -f profile.yaml
```

**创建Dataset和ThinRuntime**

首先将访问minio所需的访问凭证存储于Secret：
```
$ kubectl create secret generic minio-secret \                                                                                   
  --from-literal=minio-access-key=minioadmin \ 
  --from-literal=minio-access-secret=minioadmin
```

集群用户可通过创建Dataset和ThinRuntime CR来挂载访问Minio存储系统中的数据。下面展示了Dataset和ThinRuntime CR的一个示例（`dataset.yaml`）:

```
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: minio-demo
spec:
  mounts:
  - mountPoint: minio://my-first-bucket   # minio://<bucket name>
    name: minio
    options:
      minio-url: http://minio:9000  # minio service <url>:<port>
    encryptOptions:
      - name: minio-access-key
        valueFrom:
          secretKeyRef:
            name: minio-secret
            key: minio-access-key
      - name: minio-access-secret
        valueFrom:
          secretKeyRef:
            name: minio-secret
            key: minio-access-secret
---
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntime
metadata:
  name: minio-demo
spec:
  profileName: minio-profile
```

- `Dataset.spec.mounts[*].mountPoint`指定所需访问的数据桶(e.g. `my-frist-bucket`)
- `Dataset.spec.mounts[*].options.minio-url`指定minio在集群可访问的URL（e.g. `http://minio:9000`）
- `ThinRuntime.spec.profileName`指定已创建的ThinRuntimeProfile（e.g. `minio-profile`）

创建Dataset CR和ThinRuntime CR：

```
$ kubectl create -f dataset.yaml
```

检查Dataset状态，一段时间后，可发现Dataset和`Phase`状态变为`Bound`，Dataset可正常挂载使用：

```
$ kubectl get dataset minio-demo
NAME         UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
minio-demo                    N/A      N/A              N/A                 Bound   2m18s
```

**创建Pod访问Minio存储系统中数据**

下述展示了一个示例Pod Spec的YAML文件（`pod.yaml`）:

```
apiVersion: v1
kind: Pod
metadata:
  name: test-minio
spec:
  containers:
    - name: nginx
      image: nginx:latest
      command: ["bash"]
      args:
      - -c
      - ls -lh /data && cat /data/testfile && sleep 3600
      volumeMounts:
        - mountPath: /data
          name: data-vol
  volumes:
    - name: data-vol
      persistentVolumeClaim:
        claimName: minio-demo
```

创建数据访问Pod：

```
$ kubectl create -f pod.yaml
```

查看数据访问Pod结果：

```
$ kubectl logs test-minio      
total 512
-rw-r--r-- 1 root root 17 Dec 15 07:58 testfile
fluid-minio-test
```
可以看到，Pod `test-minio`可正常访问Minio存储系统中的数据。

## 环境清理

```
$ kubectl delete -f pod.yaml
$ kubectl delete -f dataset.yaml
$ kubectl delete -f profile.yaml
$ kubectl delete -f minio.yaml
```




