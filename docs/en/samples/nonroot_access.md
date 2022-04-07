# Example - Using Fluid to access non-root user's data

If the user data could only be access by specific uid, Runtime's 'RunAs' parameter should be set to let specific user run distributed data caching engine, to access underlying data.

This document demonstrates the above features with a simple example.

## Prerequisites

- [Fluid](https://github.com/fluid-cloudnative/fluid)(version >= 0.3.0)

Please refer to [Fluid installation documentation](https://github.com/fluid-cloudnative/fluid/blob/master/docs/zh/userguide/install.md) to complete installation.

## Running Example

**Create a non-root user**
```
$ groupadd -g 1201 fluid-user-1 && \
useradd -u 1201 -g fluid-user-1 fluid-user-1
```
The above command creates a non-root user`fluid-user-1`

**Create a directory that belongs to the user**
```
$ mkdir -p /mnt/nonroot/user1_data && \
echo "This is fluid-user-1's data" > /mnt/nonroot/user1_data/data1 && \
chown -R fluid-user-1:fluid-user-1 /mnt/nonroot/user1_data && \
chmod -R 0750 /mnt/nonroot/user1_data
```
The above command creates a directory `user1_data` belonging to `fluid-user-1` in the `/mnt/nonroot` directory, We will use the `data1` file in the `user1_data` directory to represent the data owned by `fluid-user-1`

```
$ ls -ltR /mnt/nonroot
```
Using the above command, you will see the following results
```
/mnt/nonroot/:
total 4
drwxr-x--- 2 fluid-user-1 fluid-user-1 4096 9月  27 16:45 user1_data

/mnt/nonroot/user1_data:
total 4
-rwxr-x--- 1 fluid-user-1 fluid-user-1 28 9月  27 16:45 data1
```

**Create Dataset and AlluxioRuntime resource object**

```yaml
$ cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: nonroot
spec:
  mounts:
    # Specify the directory you just created as the mount point
    - mountPoint: local:///mnt/nonroot/
      name: nonroot
  # Ensure that the data cache is placed at the node where the /mnt/nonroot directory exists
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: nonroot
              operator: In
              values:
                - "true"
---
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: nonroot
spec:
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: SSD
        path: /var/lib/docker/alluxio
        quota: 2Gi
        high: "0.95"
        low: "0.7"
  # start Alluxio as the fluid-user-1 user
  runAs:
    uid: 1201
    gid: 1201
    user: fluid-user-1
    group: fluid-user-1
  fuse:
    args:
    - fuse
    - --fuse-opts=kernel_cache,ro,max_read=131072,attr_timeout=7200,entry_timeout=7200,max_readahead=0
EOF
```

In the above yaml configuration file, we will mount the directory we just created (/mnt/nonroot ') in the same way as the Fluid host directory. For more information about Fluid mount host directory, please refer to [example - use Fluid to accelerate host directory](./hostpath.md)

In addition, in `spec.runAs` we have set user information such as `uid`, which means that we are going to start the caching engine as a `fluid-user-1` user to provide distributed caching capabilities

**Log into the Application**

```
$ kubectl exec -it nginx -- bash
```

```
$ id
```
Using the above command, you will see the following results:
```
uid=1201 gid=1201 groups=1201
```
This indicates that we started the application as a user with `uid` of 1201

**Access Data**

```
$ ls -ltR /data
```
Using the above command, you will see the following results:
```
/data/:
total 1
drwxr-xr-x 1 root root 1 Sep 27 08:45 nonroot

/data/nonroot:
total 1
drwxr-x--- 1 1201 1201 1 Sep 27 08:45 user1_data

/data/nonroot/user1_data:
total 1
-rwxr-x--- 1 1201 1201 28 Sep 27 08:45 data1
```

As you can see, Fluid exposes data belonging to a non-root user to applications that require it in the manner of **passthrough**, and the file information for the user's data does not change

Of course, the user is free to access the data:

```
$ cat /data/nonroot/user1_data/data1
```

Using the above command, you will see the following results:
```
This is fluid-user-1's data
```

## Environment Cleanup

```
$ kubectl delete -f .
$ rm -rf /mnt/nonroot
$ userdel fluid-user-1
```