# Example - How to Use ThinRuntime to Support Generic Storage Systems in Fluid

In addition to the storage/cache systems natively integrated with Fluid, Fluid provides the ThinRuntime CRD. The ThinRuntime CRD allows users to describe any custom storage system and integrate it into Fluid. This document uses [Minio](https://github.com/minio/minio) as an example to demonstrate how to manage and access data in Minio storage through Fluid.

⚠️ **Note**:
- ThinRuntime supports two usage modes: **Normal Mode** (by specifying the `profileName` field to mount external storage systems) and **Reference Dataset Mode** (without specifying `profileName` to reference other Datasets)
- In reference dataset mode, the physical runtime (the runtime bound to the original dataset) must NOT be CacheRuntime

## Prerequisites

- [Fluid](https://github.com/fluid-cloudnative/fluid) version must be at least 0.9.0

Please refer to the [Fluid Installation Documentation](https://github.com/fluid-cloudnative/fluid/blob/master/docs/en/userguide/install.md) for installation instructions.

## Running the Example

**Deploy Minio Storage to the Cluster**

Here is the YAML file `minio.yaml`:
```yaml
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

Deploy the above resources to the Kubernetes cluster:
```bash
$ kubectl create -f minio.yaml
```

After successful deployment, other Pods in the Kubernetes cluster can access data in the Minio storage system through the Minio API endpoint `http://minio:9000`. In the above YAML configuration, we set both the Minio username and password to `minioadmin`, and create a bucket named `my-first-bucket` by default when starting Minio storage. In the following examples, we will access data in the `my-first-bucket` bucket. Before executing the following steps, first execute the following commands to store sample files in `my-first-bucket`:

```bash
$ kubectl exec -it minio-69c555f4cf-np59j -- bash -c "echo fluid-minio-test > testfile"

$ kubectl exec -it minio-69c555f4cf-np59j -- bash -c "mc cp ./testfile local/my-first-bucket/" 

$ kubectl exec -it  minio-69c555f4cf-np59j -- bash -c "mc cat local/my-first-bucket/testfile"
fluid-minio-test
```

**Prepare Container Image with Minio Fuse Client**

Fluid will pass the runtime parameters required by Fuse in ThinRuntime, mount points describing data paths in Dataset, and other parameters to the ThinRuntime Fuse Pod container. Inside the container, users need to execute a parameter parsing script and pass the parsed runtime parameters to the Fuse client program, which completes the mounting of the Fuse file system within the container.

Therefore, when using the ThinRuntime CRD to describe a storage system, you need to use a **specially crafted container image** that includes the following two programs:
- Fuse client program
- Runtime parameter parsing script required by the Fuse client program

For the Fuse client program, this example chooses the S3 protocol-compatible [goofys](https://github.com/kahing/goofys) client to connect and mount the Minio storage system.

For the runtime parameter parsing script, define the following Python script `fluid-config-parse.py`:

```python
import json

with open("/etc/fluid/config/config.json", "r") as f:
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

The above Python script executes in the following steps:
1. Read the JSON string from the `/etc/fluid/config.json` file. Fluid stores and mounts the parameters required for Fuse client mounting to the `/etc/fluid/config.json` file in the Fuse container.
2. Parse the JSON string and extract the parameters required for Fuse client mounting. For example, `url`, `bucket`, `minio-access-key`, `minio-access-secret`, and other parameters in the above example.
3. After extracting the required parameters, output the mount script to the file `mount-minio.sh`

| ⚠️ Note: Starting from Fluid v1.0.0, encryption parameters defined in `dataset.spec.mounts[*].encryptOptions` cannot be directly obtained from the `/etc/fluid/config.json` file. The `/etc/fluid/config.json` file only provides the storage paths for each encryption parameter value, so the parameter parsing script needs to perform additional file reading operations (e.g., "export AWS_ACCESS_KEY_ID=\`cat $akId\`" in the above example).
| ⚠️ Note: Starting from Fluid v1.1, Fluid uses `/etc/fluid/config/config.json` as the configuration file instead of the `/etc/fluid/config.json` file used in earlier versions.

Next, use the following Dockerfile to build the image. Here we directly choose the image containing the `goofys` client program (i.e., `cloudposse/goofys`) as the base image for the Dockerfile:

```dockerfile
FROM cloudposse/goofys

RUN apk add python3 bash

COPY ./fluid-config-parse.py /fluid-config-parse.py
```

Use the following commands to build and push the image to the image repository:

```bash
$ IMG_REPO=<your image repo>

$ docker build -t $IMG_REPO/fluid-minio-goofys:demo .

$ docker push $IMG_REPO/fluid-minio-goofys:demo
```

**Create ThinRuntimeProfile**

Before creating Fluid Dataset and ThinRuntime CR to mount the Minio storage system, you first need to create the ThinRuntimeProfile CR resource. ThinRuntimeProfile is a cluster-level Fluid CRD resource that describes the basic configuration of a class of storage systems that need to be integrated with Fluid (e.g., container image information, Pod Spec description information, etc.). Cluster administrators need to define several ThinRuntimeProfile CR resources in the cluster in advance. After that, cluster users need to explicitly declare which ThinRuntimeProfile CR to use to create ThinRuntime, thereby completing the mounting of the corresponding storage system.

The following shows an example of a ThinRuntimeProfile CR for the Minio storage system (`profile.yaml`):

```yaml
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

In the above CR example:
- `fileSystemType` describes the file system type (fsType) mounted by ThinRuntime Fuse. It needs to be filled in according to the storage system Fuse client program used. For example, the fsType of the mount point mounted by goofys is `fuse`, and the fsType of the mount point mounted by s3fs is `fuse.s3fs`.
- `fuse` describes the container information of ThinRuntime Fuse, including image information (`image`, `imageTag`, `imagePullPolicy`) and container startup command (`command`), etc.

Create the ThinRuntimeProfile CR `minio-profile` in the Kubernetes cluster:
```bash
$ kubectl create -f profile.yaml
```

**Create Dataset and ThinRuntime**

First, store the access credentials required to access Minio in a Secret:
```bash
$ kubectl create secret generic minio-secret \                                                                                   
  --from-literal=minio-access-key=minioadmin \ 
  --from-literal=minio-access-secret=minioadmin
```

Cluster users can mount and access data in the Minio storage system by creating Dataset and ThinRuntime CRs. The following shows an example of Dataset and ThinRuntime CRs (`dataset.yaml`):

```yaml
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

- `Dataset.spec.mounts[*].mountPoint` specifies the data bucket to be accessed (e.g., `my-first-bucket`)
- `Dataset.spec.mounts[*].options.minio-url` specifies the URL accessible by Minio in the cluster (e.g., `http://minio:9000`)
- `ThinRuntime.spec.profileName` specifies the created ThinRuntimeProfile (e.g., `minio-profile`)

Create the Dataset CR and ThinRuntime CR:

```bash
$ kubectl create -f dataset.yaml
```

Check the Dataset status. After a while, you can see that the Dataset's `Phase` status becomes `Bound`, and the Dataset can be mounted and used normally:

```bash
$ kubectl get dataset minio-demo
NAME         UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
minio-demo                    N/A      N/A              N/A                 Bound   2m18s
```

**Create Pod to Access Data in Minio Storage System**

The following shows an example Pod Spec YAML file (`pod.yaml`):

```yaml
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

Create the data access Pod:

```bash
$ kubectl create -f pod.yaml
```

View the data access Pod results:

```bash
$ kubectl logs test-minio      
total 512
-rw-r--r-- 1 root root 17 Dec 15 07:58 testfile
fluid-minio-test
```

As you can see, the Pod `test-minio` can normally access data in the Minio storage system.

## Cleanup

```bash
$ kubectl delete -f pod.yaml
$ kubectl delete -f dataset.yaml
$ kubectl delete -f profile.yaml
$ kubectl delete -f minio.yaml
```
