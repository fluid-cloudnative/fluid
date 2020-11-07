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
