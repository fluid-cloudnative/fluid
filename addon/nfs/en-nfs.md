# Simple example of NFS access to ThinRuntime

## Prerequisites
First, deploy [NFS Server](https://nfs.sourceforge.net/) on the machine that K8s cluster can access. And configure read and write permissions, and ensure that the NFS service can be accessed on the K8s cluster node.

## Prepare NFS-FUSE Client Image
1. Parameter Resolution Script

In the FUSE container, you need to extract the configuration information of the remote file system from the relevant **ThinRuntimeProfile, Dataset, and ThinRuntime** resources. The relevant information is saved to the FUSE container in the form of JSON strings in **/etc/fluid/config.json** file.


```python
# fluid_config_init.py
import json

rawStr = ""
with open("/etc/fluid/config.json", "r") as f:
    rawStr = f.readlines()

rawStr = rawStr[0]

# Mount script
script = """
#!/bin/sh
set -ex
MNT_FROM=$mountPoint
MNT_TO=$targetPath


trap "umount ${MNT_TO}" SIGTERM
mkdir -p ${MNT_TO}
mount -t nfs ${MNT_FROM} ${MNT_TO}
sleep inf
"""

obj = json.loads(rawStr)


with open("mount-nfs.sh", "w") as f:
    f.write("mountPoint=\"%s\"\n" % obj['mounts'][0]['mountPoint'])
    f.write("targetPath=\"%s\"\n" % obj['targetPath'])

    f.write(script)

```
The Python script injects the parameters into the shell script in the form of variables after extraction.

2. Mount script

After the parameters are parsed and injected into the shell script, the generated script is as follows
```shell
mountPoint="xx.xx.xx.xx:/xxx/nfs"
targetPath="/runtime-mnt/thin/default/my-storage/thin-fuse"

#!/bin/sh
set -ex
MNT_FROM=$mountPoint
MNT_TO=$targetPath


trap "umount ${MNT_TO}" SIGTERM
mkdir -p ${MNT_TO}
mount -t nfs ${MNT_FROM} ${MNT_TO}

sleep inf
```

The shell script creates the mounted folder and mounts the remote file system to the target location. **Since the mount command will return immediately, in order to keep the process running continuously (to prevent the FUSE pod from restarting repeatedly), sleep inf is required to keep the process alive. At the same time, in order to uninstall the attached remote storage system before the FUSE pod is deleted, you need to capture the SIGTERM signal and execute the uninstall command.**

3. Build FUSE Client Image


Package parameter resolution scripts, mount scripts, and related libraries into the image.

```dockerfile
FROM alpine
RUN apk add python3 bash nfs-utils
ADD ./fluid_config_init.py /
```
In addition to Python scripts, you also need to **install the python environment and nfs utils NFS client** on the base image.

## Demo
### Create and Deploy ThinRuntimeProfile Resource
```shell
$ cat <<EOF > profile.yaml
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntimeProfile
metadata:
  name: nfs-profile
spec:
  fileSystemType: nfs
  
  fuse:
    image: <IMG_REPO>
    imageTag: <IMG_TAG>
    imagePullPolicy: IfNotPresent
    command:
      - sh
      - -c
      - "python3 /fluid_config_init.py && chmod u+x /mount-nfs.sh && /mount-nfs.sh"
EOF

$ kubectl apply -f profile.yaml
```
Replace the above <IMG_ REPO> to the repository name of the image you created, <IMG_ TAG>is modified to the TAG of your image.
### Create and Deploy Dataset and ThinRuntime Resource
```shell
$ cat <<EOF > data.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: nfs-demo
spec:
  mounts:
  - mountPoint: <IP:PATH>
    name: nfs-demo
---
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntime
metadata:
  name: nfs-demo
spec:
  profileName: nfs-profile
EOF

$ kubectl apply -f data.yaml
```
Modify the above mountPoint to the address of the remote NFS you want to use.

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
      volumeMounts:
        - mountPath: /data
          name: nfs-demo
  volumes:
    - name: nfs-demo
      persistentVolumeClaim:
        claimName: nfs-demo
EOF

$ kubectl apply -f app.yaml
```
After the application using the remote file system is deployed, the corresponding FUSE pod is also scheduled to the same node.

```shell
$ kubectl get pods
NAME                  READY   STATUS    RESTARTS   AGE
nfs-demo-fuse-wx7ns   1/1     Running   0          12s
nginx                 1/1     Running   0          26s
```
The remote file system is mounted to the /data directory of nginx pod.