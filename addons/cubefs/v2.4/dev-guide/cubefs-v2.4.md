# Simple example of CubeFS access to ThinRuntime

## Prerequisites

### Deploy CubeFS Cluster

#### Prerequisite

* Kubernetes 1.14+
* CSI spec version 1.1.0
* Helm 3

#### Deploy CubeFS

Deploy CubeFS v2.4.0 according to [cubefs-helm](https://github.com/cubefs/cubefs-helm).

The CSI driver also should be deployed according [cubefs-csi](https://github.com/cubefs/cubefs-csi).


### Use Remote CubeFS Cluster as backend storage

#### Create PV
```shell
$ cat << EOF > pv-static.yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: cfs-pv-static
spec:
  accessModes:
    - ReadWriteMany
  capacity:
    storage: 5Gi
  claimRef:
    apiVersion: v1
    kind: PersistentVolumeClaim
    name: cfs-pvc-static
    namespace: default
  csi:
    driver: csi.cubefs.com
    fsType: ext4
    volumeAttributes:
      masterAddr: <master-service.cubefs.svc.cluster.local:17010>
      owner: <cubefs>
      volName: <test>
      logDir: /cfs/logs/
      logLevel: error
      accessKey: "xxxxxxxxxxxxxxxx"
      secretKey: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
    volumeHandle: pvc-static-vol
  persistentVolumeReclaimPolicy: Retain
  volumeMode: Filesystem
EOF

$ kubectl create -f pv-static.yaml
```
> where `masterAddr` is the `<IP: port>` of CubeFS master0, `owner` is the user who create `volName` in CubeFS, and replace corresponding `accessKey, secretKey`。

#### Create PVC
```shell
$ cat << EOF > pvc-static.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: cfs-pvc-static
  namespace: default
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 5Gi
EOF

$ kubectl create -f pvc-static.yaml
```

## Prepare CubeFS-FUSE Client Image

1. Parameter Resolution Script

In the FUSE container, you need to extract the configuration information of the remote file system from the relevant **ThinRuntimeProfile, Dataset, and ThinRuntime** resources. The relevant information is saved to the FUSE container in the form of JSON strings in **/etc/fluid/config.json** file.

```python
import json

rawStr = ""
with open("/etc/fluid/config.json", "r") as f:
    rawStr = f.readlines()

print(rawStr[0])

script = """
#!/bin/sh
MNT_POINT=$targetPath

echo $MNT_POINT

if test -e ${MNT_POINT}
then
    echo "MNT_POINT exist"
else
    mkdir -p ${MNT_POINT}
fi

/cfs/bin/cfs-client -c /cfs/fuse.json

sleep inf
"""

obj = json.loads(rawStr[0])
pvAttrs = obj['persistentVolumeAttrs']['cfs-pvc-static']['volumeAttributes']

print("pvAttrs", pvAttrs)

fuse = {}
fuse["mountPoint"] = obj["targetPath"]
fuse["volName"] = pvAttrs["volName"]
fuse["owner"] = pvAttrs["owner"]
fuse["masterAddr"] = pvAttrs["masterAddr"]
fuse["logDir"] = pvAttrs["logDir"]
fuse["logLevel"] = pvAttrs["logLevel"]
fuse["accessKey"] = pvAttrs["accessKey"]
fuse["secretKey"] = pvAttrs["secretKey"]

print("fuse.json: ", fuse)

with open("/cfs/fuse.json", "w") as f:
    f.write(json.dumps(fuse))

with open("mount-cubefs.sh", "w") as f:
    f.write("targetPath=\"%s\"\n" % obj['targetPath'])
    f.write(script)
```
The Python script injects the parameters into the shell script in the form of variables after extraction.

2. Mount script

After the parameters are parsed and injected into the shell script, the generated script is as follows
```shell
targetPath="/runtime-mnt/thin/default/cubefs-test/thin-fuse"

#!/bin/sh
MNT_POINT=$targetPath

echo $MNT_POINT

if test -e ${MNT_POINT}
then
    echo "MNT_POINT exist"
else
    mkdir -p ${MNT_POINT}
fi

/cfs/bin/cfs-client -c /cfs/fuse.json

sleep inf
```
The shell script creates the mounted folder and mounts the remote file system to the target location（targetPath）.**To avoid the FUSE pod from restarting repeatedly，sleep inf is required to keep the process alive**.


3. Build FUSE Client Image

Package parameter resolution scripts, mount scripts, and related libraries into the image.

```dockerfile
FROM chubaofs/cfs-client:v2.4.0
ADD fluid_config_init.py /
```

cfs-client is needed to mount CubeFS volume, so we use CubeFS client image(chubaofs/cfs-client:2.4.0) here.

At the same time, the client image has integrated the Python environment (Python2.7), which will be used to perform parameter resolution script.

## Demo

### Create and Deploy ThinRuntimeProfile Resource
```shell
$ cat <<EOF > cubefs-profile.yaml
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntimeProfile
metadata:
  name: cubefs-profile
spec:
  fileSystemType: cubefs
  fuse:
    image: <IMG_REPO>
    imageTag: <IMG_TAG>
    imagePullPolicy: IfNotPresent 
    command:
      - sh
      - -c 
      - "python /fluid_config_init.py && chmod u+x /mount-cubefs.sh && /mount-cubefs.sh"
EOF

$ kubectl apply -f runtime-profile.yaml
```
Replace the above <IMG_ REPO> to the repository name of the image you created, <IMG_ TAG>is modified to the TAG of your image.

### Create and Deploy Dataset and ThinRuntime Resource
```shell
$ cat <<EOF > dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: cubefs-test
spec:
  mounts:
    - mountPoint: pvc://cfs-pvc-static
      name: cfs-pvc-static
      path: "/"
---
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntime
metadata:
  name: cubefs-test
spec:
  profileName: cubefs
EOF

$ kubectl apply -f dataset.yaml
```
`mountPoint` is the PVC you want to mount.

### Deploy Application


```shell
$ cat <<EOF > app.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
    - name: nginx
      image: nginx
      command: ["bash"]
      args:
        - -c
        - sleep 9999
      volumeMounts:
        - mountPath: /data
          name: data-vol
  volumes:
    - name: data-vol
      persistentVolumeClaim:
        claimName: cubefs-test
EOF

$ kubectl apply -f app.yaml
```

After the application using the remote file system is deployed, the corresponding FUSE pod is also scheduled to the same node.

```shell
$ kubectl get pods
NAME                    READY   STATUS    RESTARTS   AGE
cubefs-test-fuse-lf8r4  1/1     Running   0        2m56s
nginx                   1/1     Running   0        2m56s
```
The remote file system is mounted to the /data directory of nginx pod.

```
$ kubectl exec -it nginx bash

root@nginx:/# df -h
Filesystem      Size  Used Avail Use% 
...
chubaofs-fluid  5.0G  4.0K  5.0G   1% /data
...
```