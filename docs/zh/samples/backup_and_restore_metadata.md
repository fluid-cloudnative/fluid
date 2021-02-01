# 示例-Dataset数据的的备份与恢复

## 前提条件

在运行该示例之前，请参考 [安装文档](https://github.com/fluid-cloudnative/fluid/blob/master/docs/zh/userguide/install.md) 完成安装，并检查Fluid各组件正常运行：

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
hbase-fuse-krxlb       1/1     Running       0          2m44s
hbase-fuse-mtdmc       1/1     Running       0          2m44s
hbase-master-0         2/2     Running       0          3m16s
hbase-worker-sqrzc     2/2     Running       0          2m44s
hbase-worker-whmnv     2/2     Running       0          2m44s
```
片刻后，该Pod变为Completed状态：
```bash
$ kubectl get pods
NAME                   READY   STATUS        RESTARTS   AGE
hbase-databackup-pod   0/1     Completed     0          23s
hbase-fuse-krxlb       1/1     Running       0          3m4s
hbase-fuse-mtdmc       1/1     Running       0          3m4s
hbase-master-0         2/2     Running       0          3m36s
hbase-worker-sqrzc     2/2     Running       0          3m4s
hbase-worker-whmnv     2/2     Running       0          3m4s
```
该DataBackup同样变为Complete状态：
```bash
$ kubectl get databackup
NAME           DATASET   PHASE      BACKUPPATH                           BACKUPNODENAME     AGE
hbase-backup   hbase     Complete   pvc://<pvcName>/subpath1/subpath2/                      30s
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
NAME           DATASET   PHASE      BACKUPPATH                BACKUPNODENAME             AGE
hbase-backup   hbase     Complete   local:///data/subpath1/   cn-beijing.192.168.1.146   30s
```
发现该DataBackup变为Complete状态，同时会显示保存位置所在的路径和NodeName

查询NodeName对应的主机IP：
```bash
$ kubectl describe node cn-beijing.192.168.1.146
```
获得IP地址后，就可以从主机的对应目录取走文件刚刚的两个文件


## 恢复