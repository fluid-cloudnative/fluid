# 示例-Dataset数据的的备份与恢复

## 前提条件

在运行该示例之前，请参考 [安装文档](../guide/install.md) 完成安装，并检查Fluid各组件正常运行：

```shell
$ kubectl get pod -n fluid-system
NAME                                  READY   STATUS    RESTARTS   AGE
alluxioruntime-controller-5b64fdbbb-84pc6   1/1     Running   0          8h
csi-nodeplugin-fluid-fwgjh                  2/2     Running   0          8h
csi-nodeplugin-fluid-ll8bq                  2/2     Running   0          8h
dataset-controller-5b7848dbbb-n44dj         1/1     Running   0          8h
```
在部署Fluid时，可以通过修改charts中values.yaml中的workdir，配置Fluid工作目录
工作目录主要用于暂存备份过程中的一些中间文件，默认为/tmp目录：
```yaml
workdir: /tmp
```

请保证Dataset已经进入Bound状态：

```shell
$ kubectl get dataset
NAME    UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
hbase   443.89MiB        0.00B    4.00GiB          0.0%                Bound   23m
```

## 备份
目前支持两种数据的备份：
* 数据集的metadata，包括文件系统metadata（例如文件系统inode tree）等
* 数据集的一些关键统计信息，包括数据量大小和文件数量

### 备份到PVC
首先创建DataBackup文件：
```bash
$ cat <<EOF > backup.yaml
apiVersion: data.fluid.io/v1alpha1
kind: DataBackup
metadata:
  name: hbase-backup
spec:
  dataset: hbase
  backupPath: pvc://<pvcName>/subpath1/subpath2/
EOF
```
用户需要将backPath中的pvcName设置为自己的PVC（需要有写的权限），将subpath设置为需要存储的子路径

创建后，会看到运行了一个backup Pod：
```bash
$ kubectl get pods
NAME                   READY   STATUS        RESTARTS   AGE
hbase-databackup-pod   1/1     Running       0          3s
hbase-master-0         2/2     Running       0          3m16s
hbase-worker-0         2/2     Running       0          2m44s
hbase-worker-1         2/2     Running       0          2m44s
```
片刻后，该Pod变为Completed状态：
```bash
$ kubectl get pods
NAME                   READY   STATUS        RESTARTS   AGE
hbase-databackup-pod   0/1     Completed     0          23s
hbase-master-0         2/2     Running       0          3m36s
hbase-worker-0         2/2     Running       0          3m4s
hbase-worker-1         2/2     Running       0          3m4s
```
该DataBackup同样变为Complete状态：
```bash
$ kubectl get databackup
NAME           DATASET   PHASE      PPATH                                NODENAME     AGE
hbase-backup   hbase     Complete   pvc://<pvcName>/subpath1/subpath2/   NA           30s
```

进入该PVC，查看对应子目录，发现生成的两个备份文件：
```bash
$ ls subpath1/subpath2/
hbase-default.yaml  metadata-backup-hbase-default.gz
```
其中，gz压缩包中是数据集的metadata，yaml文件中是数据集的一些关键统计信息

### 备份到本地

首先创建DataBackup文件：
```bash
$ cat <<EOF > backup.yaml
apiVersion: data.fluid.io/v1alpha1
kind: DataBackup
metadata:
  name: hbase-backup
spec:
  dataset: hbase
  backupPath: local:///data/subpath1/
EOF
```
用户需要将local://后的路径替换为需要保存备份文件的本地路径

片刻后，查看该DataBackup的状态：

```bash
$ kubectl get databackup
NAME           DATASET   PHASE      PATH                      NODENAME                   AGE
hbase-backup   hbase     Complete   local:///data/subpath1/   cn-beijing.192.168.1.146   30s
```
发现该DataBackup变为Complete状态，同时会显示保存位置所在的路径和NodeName

查询NodeName对应的主机IP：
```bash
$ kubectl describe node cn-beijing.192.168.1.146
```
获得IP地址后，就可以从主机的对应目录取走文件刚刚的两个文件

### 使用non-root身份进行备份

如果用户指定的数据备份目录只能以特定uid访问时，需要通过设置RunAs参数指定特定用户来进行备份

如果您已经参考 [示例 - 使用Fluid访问非root用户的数据](./nonroot_access.md) 为AlluxioRuntime配置了RunAs参数，默认进行备份的用户与启动缓存引擎的用户相同

如果您没有为AlluxioRuntime配置RunAs参数，或者您希望以其他用户进行备份，可以通过为DataBackup配置RunAs参数进行设置

假如每台主机的/data/subpath1/目录都属于fluid-user-1用户

修改刚刚的DataBackup：
```bash
$ cat <<EOF > backup.yaml
apiVersion: data.fluid.io/v1alpha1
kind: DataBackup
metadata:
  name: hbase-backup
spec:
  dataset: hbase
  backupPath: local:///data/subpath1/
  runAs:
    uid: 1201
    gid: 1201
    user: fluid-user-1
    group: fluid-user-1
EOF
```

等待备份完成后，前往指定的备份位置，查看刚刚备份的文件：

```bash
$ ls -al
total 217108
drwxr-s--- 2 fluid-user-1 fluid-user-1      4096 Mar 24 18:34 .
drwxr-sr-x 5 root         root              4096 Mar 23 17:02 ..
-rw-r--r-- 1 fluid-user-1 fluid-user-1        79 Mar 24 18:34 hbase-default.yaml
-rw-r--r-- 1 fluid-user-1 fluid-user-1 222303277 Mar 24 18:28 metadata-backup-hbase-default.gz
```

这表明我们以fluid-user-1的用户身份保存了备份文件

## 恢复
要进行恢复，需要保证Dataset的名称与原来保持一致，否则会找不到备份文件

### 从PVC恢复
在创建Dataset时，在spec中添加dataRestoreLocation，填入刚刚查询到的路径作为恢复路径

如果备份文件被移动过，需要修改路径
```bash
$ cat <<EOF > dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: hbase
spec:
  dataRestoreLocation:
    path: pvc://pvc-local/subpath1/
  mounts:
    - mountPoint:  https://mirrors.tuna.tsinghua.edu.cn/apache/hbase/2.2.6/
EOF
```

创建Dataset资源对象：
```bash
$ kubectl create -f dataset.yaml
dataset.data.fluid.io/hbase created
```

AlluxioRuntime资源对象无需做特殊配置：
```bash
$ cat<<EOF >runtime.yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: hbase
spec:
  replicas: 2
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 2Gi
        high: "0.95"
        low: "0.7"
EOF
```
创建AlluxioRuntime资源对象：
```bash
$ kubectl create -f runtime.yaml
alluxioruntime.data.fluid.io/hbase created
```
数据集加载时，将不再从UFS中加载metadata并统计UFS TOTAL SIZE等信息，而是从备份文件中恢复

片刻后，Dataset进入Bound状态：
```bash
$ kubectl get dataset
NAME    UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED  PERCENTAGE   PHASE   AGE
hbase   443.86MiB        0.00B    4.00GiB          0.0%                 Bound   20h
```
### 从本地恢复
在创建Dataset时，在spec中添加dataRestoreLocation，填入刚刚查询到的路径和备份文件所在主机的nodeName

如果备份文件被移动过，需要修改路径和nodeName
```bash
$ cat <<EOF > dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: hbase
spec:
  dataRestoreLocation:
    path: local:///data/subpath1/
    nodeName: cn-beijing.192.168.1.146
  mounts:
    - mountPoint:  https://mirrors.tuna.tsinghua.edu.cn/apache/hbase/2.2.6/
EOF
```
创建AlluxioRuntime资源对象：
```bash
$ kubectl create -f runtime.yaml
alluxioruntime.data.fluid.io/hbase created
```
片刻后，Dataset进入Bound状态
```bash
$ kubectl get dataset
NAME    UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED  PERCENTAGE   PHASE   AGE
hbase   443.86MiB        0.00B    4.00GiB          0.0%                 Bound   20h
```

