# Simple example of GlusterFS access to ThinRuntime

## Prerequisites
GlusterFS access to ThinRuntime needs to construct an GlusterFS-fuse client. This demo uses [this repository](https://github.com/gluster/gluster-containers) to build a FUSE image.

## Prepare GlusterFS-FUSE Client Image
1. Parameter Resolution Script

In the FUSE container, you need to extract the configuration information of the remote file system from the relevant **ThinRuntimeProfile, Dataset, and ThinRuntime** resources. The relevant information is saved to the FUSE container in the form of JSON strings in **/etc/fluid/config.json** file.

```python
import json
import os
import subprocess

obj = json.load(open("/etc/fluid/config.json"))

mount_point = obj["mounts"][0]["mountPoint"]
target_path = obj["targetPath"]

os.makedirs(target_path, exist_ok=True)

if len(mount_point.split(":")) != 2:
    print(
        f"The mountPoint format [{mount_point}] is wrong, should be server:volumeId")
    exit(1)

server, volume_id = mount_point.split(":")
args = ["glusterfs", "--volfile-server", server, "--volfile-id",
        volume_id, target_path, "--no-daemon", "--log-file", "/dev/stdout"]

# Available options are described in the following pages:
# https://manpages.ubuntu.com/manpages/trusty/en/man8/mount.glusterfs.8.html
# https://manpages.ubuntu.com/manpages/trusty/en/man8/glusterfs.8.html
if "options" in obj["mounts"][0]:
    options = obj["mounts"][0]["options"]
    for option in options:
        if option[0] == "ro":
            option[0] = "read-only"
        elif option[0] == "transport":
            option[0] = "volfile-server-transport" 
            
        if option[1].lower() == "true":
            args.append(f'--{option[0]}')
        elif option[1].lower() == "false":
            continue
        else:
            args.append(f"--{option[0]}={option[1]}")

subprocess.run(args)
```
The shell script creates the mounted folder and mounts the remote file system to the target location.

2. Build FUSE Client Image


Package parameter resolution scripts, mount scripts, and related libraries into the image.

```dockerfile
FROM gluster/glusterfs-client@sha256:66b1d51d327ab1c4c1d81e6bad28444e13e1746c2d6f009f9874dad2fba9836e

ADD entrypoint.py /

CMD ["python3", "/entrypoint.py"]
```

## Demo
### Create and Deploy ThinRuntimeProfile Resource
```shell
$ cat <<EOF > profile.yaml
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntimeProfile
metadata:
  name: glusterfs
spec:
  fileSystemType: mount.glusterfs
  
  fuse:
    image: <IMG_REPO>
    imageTag: <IMG_TAG>
    imagePullPolicy: IfNotPresent
EOF

$ kubectl apply -f profile.yaml
```
Replace the above <IMG_REPO> to the repository name of the image you created, <IMG_TAG>is modified to the TAG of your image.

### Create and Deploy Dataset and ThinRuntime Resource
```shell
$ cat <<EOF > data.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: glusterfs-demo
spec:
  mounts:
  - mountPoint: <IP:PATH>
    name: glusterfs-demo
---
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntime
metadata:
  name: glusterfs-demo
spec:
  profileName: glusterfs
EOF

$ kubectl apply -f data.yaml
```
Modify the above mountPoint(HOST_IP:PATH) to the address of the remote glusterfs you want to use.

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
          name: glusterfs-demo
  volumes:
    - name: glusterfs-demo
      persistentVolumeClaim:
        claimName: glusterfs-demo
EOF

$ kubectl apply -f app.yaml
```
After the application using the remote file system is deployed, the corresponding FUSE pod is also scheduled to the same node.

```shell
$ kubectl get pods
NAME                        READY   STATUS    RESTARTS   AGE
glusterfs-demo-fuse-wx7ns   1/1     Running   0          12s
nginx                       1/1     Running   0          26s
```
The remote file system is mounted to the /data directory of nginx pod.