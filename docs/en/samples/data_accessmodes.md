# Demo - Set Dataset Access Mode

## Prerequisites
Before everything we are going to do, please refer to [Installation Guide](../userguide/install.md) to install Fluid on your Kubernetes Cluster, and make sure all the components used by Fluid are ready like this:
```shell
$ kubectl get pod -n fluid-system
NAME                                  READY   STATUS    RESTARTS   AGE
alluxioruntime-controller-5b64fdbbb-84pc6   1/1     Running   0          8h
csi-nodeplugin-fluid-fwgjh                  2/2     Running   0          8h
csi-nodeplugin-fluid-ll8bq                  2/2     Running   0          8h
dataset-controller-5b7848dbbb-n44dj         1/1     Running   0          8h
```

Normally, you shall see a Pod named "dataset-controller", a Pod named "alluxioruntime-controller" and several Pods named "csi-nodeplugin". 
The num of "csi-nodeplugin" Pods depends on how many nodes your Kubernetes cluster have(e.g. 2 in this demo), so please make sure all "csi-nodeplugin" Pods are working properly.

## Set dataset access mode
The access mode of the dataset is set to **ReadOnlyMany** when user doesn`t specif the access mode. If there is a need to modify the default access mode, you need to specify it in spec.accessModes[] before creating it.

The currently supported access modes are：
- `ReadWriteOnce` : the volume can be mounted as read-write by a single node. ReadWriteOnce access mode still can allow multiple pods to access the volume when the pods are running on the same node
- `ReadOnlyMany` : the volume can be mounted as read-only by many nodes
- `ReadWriteMany` : the volume can be mounted as read-write by many nodes
- `ReadWriteOncePod` : the volume can be mounted as read-write by a single Pod. Use ReadWriteOncePod access mode if you want to ensure that only one pod across whole cluster can read that PVC or write to it. This is only supported for CSI volumes and Kubernetes version 1.22+

Refer to the [document](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#access-modes) for more infomation.


## Demo
This demo sets the access method of the dataset to ReadWriteOncePod.
```
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: serverless-data
spec:
  mounts:
    - mountPoint: https://mirrors.bit.edu.cn/apache/hbase/stable/
      name: hbase
      path: "/"
  accessModes:
    - ReadWriteOncePod
```
