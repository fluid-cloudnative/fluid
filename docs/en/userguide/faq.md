## 1. Why do I fail to install fluid with Helm？

**Answer**: It is recommended to follow the [Fluid installation document](./install.md) to confirm whether the Fluid components are operating normally.

The Fluid installation document is deployed based on `Helm 3` as an example. If you use a version below `Helm 3` to deploy Fluid and encounter the situation of `CRD not starting normally`. This may be because `Helm 3` and above versions will automatically install CRD during `helm install` but the lower version of Helm will not. See the [Helm official documentation](https://helm.sh/docs/chart_best_practices/custom_resource_definitions/).

In this case, you need to install CRD manually：
```bash
$ kubectl create -f fluid/crds
```

## 2. Why can't I delete Runtime？

**Answer**: Please check the related Pod running status and Runtime Events.

As long as any active Pod is still using the Volume created by Fluid, Fluid will not complete the delete operation.

The following commands can quickly find these active Pods. When using it, replace `<dataset_name>` and `<dataset_namespace>` with yours:
```bash
kubectl describe pvc <dataset_name> -n <dataset_namespace> | \
	awk '/^Mounted/ {flag=1}; /^Events/ {flag=0}; flag' | \
	awk 'NR==1 {print $3}; NR!=1 {print $1}' | \
	xargs -I {} kubectl get po {} | \
	grep -E "Running|Terminating|Pending" | \
	cut -d " " -f 1
```


## 3.Why do I run the example [Speed up Accessing Remote Files](../samples/accelerate_data_accessing.md), and I will encounter Input/output error when copying files for the first time. Similar to the following:

```
time cp ./pyspark-2.4.6.tar.gz /tmp/
cp: error reading ‘./pyspark-2.4.6.tar.gz’: Input/output error
cp: failed to extend ‘/tmp/pyspark-2.4.6.tar.gz’: Input/output error

real	3m15.795s
user	0m0.001s
sys	0m0.092s
```

What caused this to happen？

**Answer**: The purpose of this example is to allow users to use the existing mirror download address of Apache software based on Http protocol to demonstrate the ability of data copy acceleration without building UFS (underlayer file system). However, in actual scenarios, the implementation of WebUFS is generally not used。But there are three limitations in this example：

1.Availability and access speed of Apache software mirror download address.

2.WebUFS is derived from Alluxio's community contribution and it is not the optimal implementation. For example, the implementation is not based on offset-based breakpoint resuming, which leads to the need to trigger WebUFS to read a large number of data blocks for each remote read operation.

3.Since the copy behavior is implemented based on Fuse, the upper limit of each Fuse chunk read is 128KB under the Linux Kernel; The larger the file is, the first copy will trigger a large number of read operations.

In response to this problem, we proposed an optimized solution:

1.When configuring read, set the block size and chunk size to be larger than the file size, so as to avoid the influence of frequent reading in Fuse implementation.

```
alluxio.user.block.size.bytes.default: 256MB
alluxio.user.streaming.reader.chunk.size.bytes: 256MB
alluxio.user.local.reader.chunk.size.bytes: 256MB
alluxio.worker.network.reader.buffer.size: 256MB
```

2.In order to ensure that the target file can be downloaded successfully, the timeout of the block download can be adjusted. The timeout period in the example is 5 minutes. If your network condition is not good, you can set a longer time as appropriate.

```
alluxio.user.streaming.data.timeout: 300sec
```

3.You can try to load the file manually.

```
kubectl exec -it hbase-master-0 bash
time alluxio fs  distributedLoad --replication 1 /
```

## 4. Why does the `driver name fuse.csi.fluid.io not found in the list of registered CSI drivers` error appear when I create a task to mount the PVC created by Runtime?

**Answer**: Please check whether the kubelet configuration of the node which the task is scheduled on is the default value `/var/lib/kubelet`.

First, check whether Fluid's CSI component is normal using the command.

The following command can find the Pod quickly. When using it, replace the `<node_name>` and `<fluid_namespace>` with yours:
```bash
kubectl get pod -n <fluid_namespace> | grep <node_name>

# <pod_name> the pod name of last step
kubectl logs <pod_name> node-driver-registrar -n <fluid_namespace>
kubectl logs <pod_name> plugins -n <fluid_namespace>
```

If there is no error in the Log of the above steps, check whether the csidriver object exists:
```
kubectl get csidriver
```
If the csidriver object exists, please check if the csi registered node contains `<node_name>`:
```
kubectl get csinode | grep <node_name>
```
If the above command has no output, check whether the kubelet configuration of the node which the task is scheduled on is the default value `/var/lib/kubelet`. Because Fluid's CSI component is registered to kubelet through a fixed address socket, the default value is `--csi-address=/var/lib/kubelet/csi-plugins/fuse.csi.fluid.io/csi.sock --kubelet-registration-path=/var/lib/kubelet/csi-plugins/fuse.csi.fluid.io/csi.sock`.

## 5. After upgrading fluid，why does the dataset created in older version miss some fields compared to a newly created dataset, when querying them via `kubectl get` command？

**Answer**: During the upgrading, we perhaps have upgraded the CRDs. The dataset created in older version，will set the new fields in the CRDs to null
For example, if you upgrade from v0.4 or before, the dataset did not have a 'FileNum' field at that time
After upgrading fluid, if you use the `kubectl get` command, you cannot query the FileNum of the dataset

You can recreate the dataset, and the new dataset will display these fields normally

## 6. Why do I run the example [Nonroot access](../samples/nonroot_access.md), and I  encounter mkdir permission denied error

**Answer**: In nonroot scenirio,Firstly,you have to check that you pass the right useinfo to runtime,Secondly,you should check the alluxio master pod status,and use journalctl to see the kubelet logs in the node of alluxio master pod.the mkdir error was cause when mount the hostpath to container,So we have to check the root have right permission to exec the directory.We can use stat to check if root can exec the directory.For example in the below root have permission to operator /dir
```
$ stat /dir
  File: ‘/dir’
  Size: 32              Blocks: 0          IO Block: 4096   directory
Device: fd00h/64768d    Inode: 83          Links: 3
Access: (0755/drwxr-xr-x)  Uid: (    0/    root)   Gid: (    0/    root)
Access: 2021-04-14 23:35:47.928805350 +0800
Modify: 2021-01-19 00:16:21.539559082 +0800
Change: 2021-01-19 00:16:21.539559082 +0800
 Birth: -

```
