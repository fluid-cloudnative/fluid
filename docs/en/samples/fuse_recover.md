# Demo - How to enable FUSE auto-recovery

## Installation

You can download the latest Fluid installation package from [Fluid Releases](https://github.com/fluid-cloudnative/fluid/releases).

In Fluid chart values.yaml, set `csi.featureGates` to `FuseRecovery=true`, indicating enable FUSE auto-recovery.
Refer to the [Installation Documentation](../userguide/install.md) to complete the installation. And check that the components of Fluid are running normally (here takes JuiceFSRuntime as an example):

```shell
$ kubectl -n fluid-system get po
NAME                                        READY   STATUS    RESTARTS   AGE
csi-nodeplugin-fluid-2gtsz                  2/2     Running   0          20m
csi-nodeplugin-fluid-2h79g                  2/2     Running   0          20m
csi-nodeplugin-fluid-sc459                  2/2     Running   0          20m
dataset-controller-57fb4569cd-k2jb7         1/1     Running   0          20m
fluid-webhook-844dcb995f-nfmjl              1/1     Running   0          20m
juicefsruntime-controller-7d9c964b4-jnbtf   1/1     Running   0          20m
```

Typically, you will see a Pod named `dataset-controller`, a Pod named `juicefsruntime-controller`, a Pod named `fluid-webhook`
and multiple pods named `csi-nodeplugin` are running. Among them, the number of `csi-nodeplugin` these Pods depends on the number of nodes in your Kubernetes cluster.

## Demo

**Create dataset and runtime**

Create corresponding Runtime resources and Datasets with the same name for different types of runtimes. Take JuiceFSRuntime as an example here. For details, please refer to [Documentation](juicefs_runtime.md), as follows:

```shell
$ kubectl get juicefsruntime
NAME      WORKER PHASE   FUSE PHASE   AGE
jfsdemo   Ready          Ready        2m58s
$ kubectl get dataset
NAME      UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
jfsdemo   [Calculating]    N/A                       N/A                 Bound   2m55s
```

**Create Pod**

```yaml
$ cat<<EOF >sample.yaml
apiVersion: v1
kind: Pod
metadata:
  name: demo-app
  labels:
    fuse.serverful.fluid.io/inject: "true"
spec:
  containers:
    - name: demo
      image: nginx
      volumeMounts:
        - mountPath: /data
          name: demo
  volumes:
    - name: demo
      persistentVolumeClaim:
        claimName: jfsdemo
  EOF
$ kubectl create -f sample.yaml
pod/demo-app created
```

The FUSE mount point auto-recovery feature requires the pod's mountPropagation to be set to `HostToContainer` or `Bidirectional` to pass the mount point information between the container and the host. 
And `Bidirectional` requires the container to be a privileged container.
Fluid webhook helps automatically set the pod's mountPropagation to `HostToContainer`. To enable this function, you need to set label `fuse.serverful.fluid.io/inject=true` on the corresponding Pod's metadata (See the sample mentioned above).


**See if the Pod is created and check its mountPropagation**

```shell
$ kubectl get po |grep demo
demo-app             1/1     Running   0          96s
jfsdemo-fuse-g9pvp   1/1     Running   0          95s
jfsdemo-worker-0     1/1     Running   0          4m25s
$ kubectl get po demo-app -oyaml |grep volumeMounts -A 3
    volumeMounts:
    - mountPath: /data
      mountPropagation: HostToContainer
      name: demo
```

## Test FUSE mount point auto recovery

**Delete FUSE pod**

Delete the FUSE pod, and waiting for it to restart:

```shell
$ kubectl delete po jfsdemo-fuse-g9pvp
pod "jfsdemo-fuse-g9pvp" deleted
$ kubectl get po
NAME                 READY   STATUS    RESTARTS   AGE
demo-app             1/1     Running   0          5m7s
jfsdemo-fuse-bdsdt   1/1     Running   0          6s
jfsdemo-worker-0     1/1     Running   0          7m56s
````

After the new FUSE pod is created, check the mount points in the demo pod:

```shell
$ kubectl exec -it demo-app bash
kubectl exec [POD] [COMMAND] is DEPRECATED and will be removed in a future version. Use kubectl exec [POD] -- [COMMAND] instead.
[root@demo-app /]# df -h
Filesystem      Size  Used Avail Use% Mounted on
overlay         100G  9.4G   91G  10% /
tmpfs            64M     0   64M   0% /dev
tmpfs           2.0G     0  2.0G   0% /sys/fs/cgroup
JuiceFS:minio   1.0P   64K  1.0P   1% /data
/dev/sdb1       100G  9.4G   91G  10% /etc/hosts
shm              64M     0   64M   0% /dev/shm
tmpfs           3.8G   12K  3.8G   1% /run/secrets/kubernetes.io/serviceaccount
tmpfs           2.0G     0  2.0G   0% /proc/acpi
tmpfs           2.0G     0  2.0G   0% /proc/scsi
tmpfs           2.0G     0  2.0G   0% /sys/firmware
```

It can be seen that there is no `Transport endpoint is not connected` error in the container, indicating that the mount point has been restored.

**Check dataset events**

```shell
$ kubectl describe dataset jfsdemo
Name:         jfsdemo
Namespace:    default
...
Events:
  Type    Reason              Age                  From         Message
  ----    ------              ----                 ----         -------
  Normal  FuseRecoverSucceed  2m34s (x5 over 11m)  FuseRecover  Fuse recover /var/lib/kubelet/pods/6c1e0318-858b-4ead-976b-37ccce26edfe/volumes/kubernetes.io~csi/default-jfsdemo/mount succeed
```

You can see that there is a `FuseRecover` event in the Dataset event, indicating that Fluid has performed a recovery operation on the mount.

## Notice

When the FUSE pod crashes, the recovery time of the mount point depends on the recovery of the FUSE pod itself and the period of the csi polling kubelet (env `RECOVER_FUSE_PERIOD`).
Before the recovery, the mount point will still have a `Transport endpoint is not connected` error, which is expected.
In addition, the mount point recovery is accomplished by bind mount repeatedly. For the file descriptors that have been opened by the application before the FUSE pod crash,
cannot be recovered even after the mount point recovered. The application itself needs to retry when an error occurs and enhance its robustness.
