# Simple example of NFS access to ThinRuntime

## Prerequisites
NFS access to ThinRuntime needs to construct an NFS-FUSE client. This demo uses [this item](https://github.com/sahlberg/fuse-nfs) to build a FUSE image fo.

## Prepare NFS-FUSE Client Image
1. Parameter Resolution Script

In the FUSE container, you need to extract the configuration information of the remote file system from the relevant **ThinRuntimeProfile, Dataset, and ThinRuntime** resources. The relevant information is saved to the FUSE container in the form of JSON strings in **/etc/fluid/config.json** file.


```python
# fluid_config_init.py
import json

rawStr = ""
with open("/etc/fluid/config/config.json", "r") as f:
    rawStr = f.readlines()

rawStr = rawStr[0]

# Mount script
script = """
#!/bin/sh
set -ex
MNT_FROM=$mountPoint
MNT_TO=$targetPath


trap "fusermount -u ${MNT_TO}" SIGTERM
mkdir -p ${MNT_TO}
fuse-nfs -n nfs://${MNT_FROM} -m ${MNT_TO}
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
mountPoint="xx.xx.xx.xx/xxx/nfs"
targetPath="/runtime-mnt/thin/default/my-storage/thin-fuse"

#!/bin/sh
set -ex
MNT_FROM=$mountPoint
MNT_TO=$targetPath


trap "fusermount -u ${MNT_TO}" SIGTERM
mkdir -p ${MNT_TO}
fuse-nfs -n nfs://${MNT_FROM} ${MNT_TO}

sleep inf
```

The shell script creates the mounted folder and mounts the remote file system to the target location. **Since the mount command will return immediately, in order to keep the process running continuously (to prevent the FUSE pod from restarting repeatedly), sleep inf is required to keep the process alive. At the same time, in order to uninstall the attached remote storage system before the FUSE pod is deleted, you need to capture the SIGTERM signal and execute the uninstall command.**

3. Build FUSE Client Image


Package parameter resolution scripts, mount scripts, and related libraries into the image.

```dockerfile
# Build environment
FROM ubuntu:jammy as BUILD
RUN apt update && \
    apt install --yes automake libfuse-dev libnfs-dev libnfs-dev libnfs13 libtool libtool m4 make xsltproc


COPY ./fuse-nfs-master /src
WORKDIR /src
RUN ./setup.sh && \
    ./configure && \
    make

# Production image
FROM ubuntu:jammy
RUN apt update && \
    apt install --yes bash fuse libfuse2 libnfs13 python3 && \
    apt clean autoclean && \
    apt autoremove --yes && \
    rm -rf /var/lib/{apt,dpkg,cache,log}/
ADD ./fluid_config_init.py /
ADD ./entrypoint.sh /usr/local/bin
COPY --from=BUILD /src/fuse/fuse-nfs /bin/fuse-nfs
CMD ["/usr/local/bin/entrypoint.sh"]
```
In addition to Python scripts, you also need to **install the python environment and nfs utils NFS client** on the base image.
Users need to download [this item](https://github.com/sahlberg/fuse-nfs) to the local **<PATH>/fuse-nfs-master and then copy Dockerfile with the above example, the parameter resolution script fluid_config_init.py and the startup script entrypoint.sh to the parent directory**.

```shell
$ ls                                      
Dockerfile  entrypoint.sh  fluid_config_init.py  fuse-nfs-master
```

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
  - mountPoint: <IP/PATH>
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
Modify the above mountPoint(HOST_IP/PATH) to the address of the remote NFS you want to use.

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