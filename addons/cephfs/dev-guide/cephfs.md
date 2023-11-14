# Simple example of Ceph access to ThinRuntime

## Prerequisites

should deploy [Ceph](https://ceph.com/en/) on the machine that K8s cluster can access. And configure read and write permissions, and ensure that the Ceph service can be accessed on the K8s cluster node.

## Prepare Ceph-FUSE Client Image

 ### 1.Parameter Resolution Script

In the FUSE container, you need to extract the configuration information of the remote file system from the relevant **ThinRuntimeProfile, Dataset, and ThinRuntime** resources. The relevant information is saved to the FUSE container in the form of JSON strings in **/etc/fluid/config.json** file.

~~~ python
# fluid_config_init.py
import json


def write_conf(pvAttrs: dict):
    confAttrs = pvAttrs
    with open("/etc/ceph/ceph.conf", "w") as f:
        f.write("[global]\n")
        f.write("fsid=%s\n" % confAttrs["fsid"])
        f.write("mon_initial_members=%s\n" % confAttrs["mon_initial_members"])
        f.write("mon_host=%s\n" % confAttrs["mon_host"])
        f.write("auth_cluster_required=%s\n" % confAttrs["auth_cluster_required"])
        f.write("auth_service_required=%s\n" % confAttrs["auth_service_required"])
        f.write("auth_client_required=%s\n" % confAttrs["auth_client_required"])


def write_keyring(pvAttrs: dict):
    keyringAttrs = pvAttrs
    with open("/etc/ceph/ceph.client.admin.keyring", "w+") as f:
        f.write("[client.admin]\n")
        f.write("key=%s\n" % keyringAttrs["key"])


def read_json():
    with open("/etc/fluid/config.json", "r") as f:
        rawStr = f.readlines()
    rawStr = "".join(rawStr)
    obj = json.loads(rawStr)
    return obj


def write_cmd(mon_url: str, target_path):
    mon_url = mon_url.replace("ceph://", "")
    script = """#!/bin/sh
mkdir -p {}
exec ceph-fuse -n client.admin -k /etc/ceph/ceph.client.admin.keyring -c /etc/ceph/ceph.conf  {}
"""
    with open("/mount_ceph.sh", "w+") as f:
        f.write(script.format(target_path, target_path))


if __name__ == '__main__':
    pvAttrs = read_json()
    write_conf(pvAttrs['mounts'][0]['options'])
    write_keyring(pvAttrs['mounts'][0]['options'])
    write_cmd(pvAttrs['mounts'][0]['mountPoint'], pvAttrs['targetPath'])
~~~

This Python script will extract parameters from **/etc/fluid/config.json** and generate config file **/etc/ceph/ceph.conf**, keyring file **/etc/ceph/ceph.client.admin.keyring** and mount script **/mount_cpeh.sh**.



### 2.Ceph-FUSE Mount Script

After the parameters are parsed and injected into the shell script, the generated script is as follows:

~~~ shell
#!/bin/sh
mkdir -p /runtime-mnt/thin/default/my-storage/thin-fuse
exec ceph-fuse -n client.admin -k /etc/ceph/ceph.client.admin.keyring -c /etc/ceph/ceph.conf  /runtime-mnt/thin/default/my-storage/thin-fuse
sleep inf
~~~

The shell script creates the mounted folder and mounts the remote file system to the target location(targetPath). **Since the mount command will return immediately, in order to keep the process running continuously (to prevent the FUSE pod from restarting repeatedly), sleep inf is required to keep the process alive.**



### 3.Container Entrypoint Script

Edit the script `entrypoint.sh` that needs to be executed when the container is started. This script executes the `fluid_config_init.py` to generate the mount script `mount_ceph.sh`, and then execute the `mount_ceph.sh` script to mount.

~~~ shell
#!/usr/bin/env sh
set +x

python3 /fluid_config_init.py

chmod u+x /mount_ceph.sh

sh /mount_ceph.sh
~~~





### 4.Build Ceph-FUSE Client Image

Package parameter resolution scripts, mount scripts, and related libraries into the image.

~~~ dockerfile
FROM alpine@sha256:124c7d2707904eea7431fffe91522a01e5a861a624ee31d03372cc1d138a3126
# use alpine:3.18

RUN mkdir /etc/ceph
RUN apk add ceph ceph-fuse python3

ADD fluid_config_init.py /
ADD entrypoint.sh /usr/local/bin

RUN chmod u+x /usr/local/bin/entrypoint.sh

CMD ["/usr/local/bin/entrypoint.sh"]
~~~

In addition to Python scripts and shell scripts, you also need to **install the python environment and nfs utils Ceph client** on the base image.



# Demo

### 1.Create and Deploy ThinRuntimeProfile Resource

~~~ shell
$ cat <<EOF > profile.yaml
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntimeProfile
metadata:
  name: ceph-profile
spec:
  fileSystemType: ceph-fuse
  fuse:
    image: <IMG_REPO>
    imageTag: <IMG_TAG>
    imagePullPolicy: IfNotPresent
    command:
    - "/usr/local/bin/entrypoint.sh"
EOF

$ kubectl apply -f profile.yaml
~~~

Replace the above <IMG_ REPO> to the repository name of the image you created, <IMG_ TAG>is modified to the TAG of your image. The **fileSystemType** corresponds to the file type to be mounted, which can be viewd by using the mount command in the mounted pod. For example,  we could find this record after using mount command in the pod mounted by ceph-fuse:

~~~ shell
ceph-fuse on /runtime-mnt/thin/default/ceph-demo/thin-fuse type fuse.ceph-fuse (rw,nosuid,nodev,relatime,user_id=0,group_id=0,allow_other)
~~~

 In the record, "fuse.ceph-fuse" is mount type, we can choose any sub string of "fuse.ceph-fuse" for  **fileSystemType**, like "ceph-fuse".



### 2.Create and Deploy Dataset and ThinRuntime Resource

This is the yaml file for **Dataset** and **Thinruntime**:

~~~ yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: ceph-demo
spec:
  mounts:
  - mountPoint: ceph://<IP:Port>
    name: ceph-pvc
    options:
      fsid: <fsid>
      mon_initial_members: <mon_initial_members>
      mon_host: <mon_host>
      auth_cluster_required: <auth_cluster_required>
      auth_service_required: <auth_service_required>
      auth_client_required: <auth_client_required>
      key: <key>
---
apiVersion: data.fluid.io/v1alpha1
kind: ThinRuntime
metadata:
  name: ceph-demo
spec:
  profileName: ceph-profile
~~~

The **mounts** part will be chuansfer to **json** file and mounted to **/etc/fluid/config.json** path. 

The **json** file for above **yaml** file is:

~~~ json
{
  "mounts": [
    {
      "mountPoint": "ceph://<IP:Port>",
      "name": "ceph-pvc",
      "options": {
        "key": "<key>",
        "fsid": "<fsid>",
        "mon_initial_members": "<mon_initial_members>",
        "mon_host": "<mon_host>",
        "auth_cluster_required": "<auth_cluster_required>",
        "auth_service_required": "<auth_service_required>",
        "auth_client_required": "<auth_client_required>"
      }
    }
  ],
  "targetPath": "/runtime-mnt/thin/default/ceph-demo/thin-fuse"
}
~~~

The **targetPath** is the mount point we used in fuse-pod.



### 3.部署应用

~~~ shell
$ cat <<EOF > app.yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-ceph
spec:
  containers:
  - name: nginx
    image: nginx
    command: ["bash"]
    args:
    - -c
    - ls /data && sleep inf
    volumeMounts:
    - mountPath: /data
      name: data-vol
  volumes:
  - name: data-vol
    persistentVolumeClaim:
      claimName: ceph-demo

$ kubectl apply -f app.yaml
~~~

After the application using the remote file system is deployed, the corresponding FUSE pod is also scheduled to the same node.

~~~ shell
$ kubectl get pods
NAME                  READY   STATUS    RESTARTS   AGE
ceph-demo-fuse-7kfdx  1/1     Running   0          34s
nginx                 1/1     Running   0          47s
~~~

The remote file system is mounted to the /data directory of nginx pod.
