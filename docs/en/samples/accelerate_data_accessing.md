# DEMO - Speed Up Accessing Remote Files
Powered by [Alluxio](https://www.alluxio.io) and [Fuse](https://github.com/libfuse/libfuse), Fluid provides a simple way for users to access files stored in remote filesystems, just like accessing some ordinary file in local filesystems. 
What's more, with a powerful caching capability provided, users can enjoy a great speedup on accessing remote files especially for those that have a frequent access pattern.

This demo aims to show you an overview of all the features mentioned above.

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

Normally, you shall see a Pod named "controller-manager" and several Pods named "csi-nodeplugin". 
The num of "csi-nodeplugin" Pods depends on how many nodes your Kubernetes cluster have(e.g. 2 in this demo), so please make sure all "csi-nodeplugin" Pods are working properly.

## Set Up Workspace
```shell
$ mkdir <any-path>/accelerate
$ cd <any-path>/accelerate
```

## Install Resources to Kubernetes

**Check the `Dataset` object to be created**
```shell
$ cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: hbase
spec:
  mounts:
    - mountPoint: https://mirrors.tuna.tsinghua.edu.cn/apache/hbase/stable/
      name: hbase
EOF
```
> Notes: Here, we use THU's tuna Apache mirror site as our `mountPoint`. If your environment isn't in Chinese mainland, please replace it with `https://downloads.apache.org/hbase/stable/`.

Here, we'd like to create a resource object with kind `Dataset`. `Dataset` is a Custom Resource Definition(CRD) defined by Fluid and used to tell Fluid where to find all the data you'd like to access.
Under the hood, Fluid uses Alluxio to do some mount operations, so `mountPoint` property can be any legal UFS path acknowledged by Alluxio. Here, we use [WebUFS](https://docs.alluxio.io/os/user/stable/en/ufs/WEB.html) for its simplicity.

For more information about UFS, please refer to [Alluxio Docs - Storage Integrations](https://docs.alluxio.io/os/user/stable/en/ufs/HDFS.html)

For more information about properties in `Dataset`, please refer to our [API doc](../dev/api_doc.md) 

> We use hbase v2.2.5 on a mirror site of Apache downloads as an example of remote file. It's nothing special, you can change it to any remote file you like. But please note that, if you are going to use WebUFS like we do, files on Apache sites are highly recommended because you might need some [advanced configurations](https://docs.alluxio.io/os/user/stable/en/ufs/WEB.html#configuring-alluxio) due to current implementation of WebUFS.

**Create the `Dataset` object**
```shell
$ kubectl create -f dataset.yaml
dataset.data.fluid.io/hbase created
```

**Check status of the `Dataset` object**
```shell
$ kubectl get dataset hbase
NAME    UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE      AGE
hbase                                                                  NotBound   13s
```

With a `NotBound` phase in status, the dataset is not ready cause there isn't any `AlluxioRuntime` object supporting it. We'll create one in the following steps.

**Check the `AlluxioRuntime` object to be created**
```shell 
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

**Create a `AlluxioRuntime` object**
```shell
$ kubectl create -f runtime.yaml
alluxioruntime.data.fluid.io/hbase created
```

**Get the `AlluxioRuntime` object**
```shell
$ kubectl get alluxioruntime
NAME    AGE
hbase   55s
```

`AlluxioRuntime` is another CRD defined by Fluid. An `AluxioRuntime` object describes specifications used to run an Alluxio instance.

Wait for a while, and make sure all components defined in the `AlluxioRuntime` object are ready. You shall see something like this:
```shell
$ kubectl get pod
NAME                 READY   STATUS    RESTARTS   AGE
hbase-master-0       2/2     Running   0          62s
hbase-worker-0       2/2     Running   0          27s
hbase-worker-1       2/2     Running   0          27s
```

**Check status of the `Dataset` object again**
```shell
$ kubectl get dataset hbase
NAME    UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
hbase   443.5MiB         0B       4GiB             0%                  Bound   2m39s
```
`Dataset` object has been updated since a related Alluxio instance is ready and successfully bounded to the `Dataset` object.

**Check status of the `AlluxioRuntime` object**
```shell
$ kubectl get alluxioruntime hbase -o wide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE   AGE
hbase   1               1                 Ready          2               2                 Ready          0             0               Ready        2m50s
```
Detailed information about the Alluxio instance is provided here.

**Check related PersistentVolume and PersistentVolumeClaim**
```shell
$ kubectl get pv
NAME    CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM           STORAGECLASS   REASON   AGE
hbase   100Gi      RWX            Retain           Bound    default/hbase                           18m
```

```shell
$ kubectl get pvc
NAME    STATUS   VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS   AGE
hbase   Bound    hbase    100Gi      RWX                           18m
```

Related PV and PVC have been created by Fluid since the `Dataset` object is ready(bounded).
Workloads are now able to access remote files by mounting PVC.

## Remote File Access

**Check the app to be created**

```shell
$ cat<<EOF >nginx.yaml
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
          name: hbase-vol
  volumes:
    - name: hbase-vol
      persistentVolumeClaim:
        claimName: hbase
EOF
```

**Run a demo app to access remote files**
```shell
$ kubectl create -f nginx.yaml
```

Login to nginx Pod:
```shell
$ kubectl exec -it nginx -- bash
```

Check file status:
```shell
$ ls -1 /data/hbase
CHANGES.md
RELEASENOTES.md
api_compare_2.2.5RC0_to_2.2.4.html
hbase-2.2.5-bin.tar.gz
hbase-2.2.5-client-bin.tar.gz
hbase-2.2.5-src.tar.gz
```

```shell
$ du -h /data/hbase/*
174K    /data/hbase/CHANGES.md
106K    /data/hbase/RELEASENOTES.md
115K    /data/hbase/api_compare_2.2.5RC0_to_2.2.4.html
211M    /data/hbase/hbase-2.2.5-bin.tar.gz
1.0K    /data/hbase/hbase-2.2.5-bin.tar.gz.asc
512     /data/hbase/hbase-2.2.5-bin.tar.gz.sha512
200M    /data/hbase/hbase-2.2.5-client-bin.tar.gz
1.0K    /data/hbase/hbase-2.2.5-client-bin.tar.gz.asc
512     /data/hbase/hbase-2.2.5-client-bin.tar.gz.sha512
34M     /data/hbase/hbase-2.2.5-src.tar.gz
1.0K    /data/hbase/hbase-2.2.5-src.tar.gz.asc
512     /data/hbase/hbase-2.2.5-src.tar.gz.sha512
```

Logout:
```shell 
$ exit
```

As you may have seen, all the files on the WebUFS(e.g. hbase-related files on Apache mirror in our case) appear no differences from any other file in the local filesystem of the nginx Pod.

## Speed Up Accessing Remote Files
To demonstrate how great speedup you may enjoy when accessing remote files, here is a demo job:

**Check the test job to be launched**
```shell
$ cat<<EOF >app.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: fluid-copy-test
spec:
  template:
    spec:
      restartPolicy: OnFailure
      containers:
        - name: busybox
          image: busybox
          command: ["/bin/sh"]
          args: ["-c", "set -x; time cp -r /data/hbase ./"]
          volumeMounts:
            - mountPath: /data
              name: hbase-vol
      volumes:
        - name: hbase-vol
          persistentVolumeClaim:
            claimName: hbase
EOF
```

**Launch a test job**
```shell
$ kubectl create -f app.yaml
job.batch/fluid-test created
```
Under the hood, the test job executes a shell command `time cp -r /data/hbase ./` and prints its result.
Wait for a while and make sure the job has completed. You can check its runnning status by:

```shell
$ kubectl get pod
NAME                    READY   STATUS      RESTARTS   AGE
fluid-copy-test-h59w9   0/1     Completed   0          1m25s
...
```

> Note: the `h59w9` in `fluid-copy-test-h59w9` is a specifier generated by the Job we created. It's highly possible that you may have different specifier in your environment. Please remember replace it with your own specifier in the following steps

**Check running time of the test job**
```shell
$ kubectl logs fluid-copy-test-h59w9
+ time cp -r /data/hbase ./
real  1m 2.74s
user  0m 0.00s
sys   0m 1.35s
```
It's our first time to read such a file, and it takes us about 63s. It may be not as fast as you expected but:

**Check status of the dataset**
```shell
$ kubectl get dataset hbase
NAME    UFS TOTAL SIZE   CACHED     CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
hbase   443.5MiB         443.5MiB   4GiB             100%                Bound   9m27s
```
Now, all the remote files have been cached in Alluxio.

**Re-Launch the test job**
```shell
$ kubectl delete -f app.yaml
$ kubectl create -f app.yaml
```

It'll finish very soon after creation this time:
```shell
$ kubectl get pod
NAME                    READY   STATUS      RESTARTS   AGE
fluid-copy-test-d9h2x   0/1     Completed   0          24s
...
```

```shell
$ kubectl logs fluid-copy-test-d9h2x
+ time cp -r /data/hbase ./
real  0m 2.94s
user  0m 0.00s
sys   0m 1.27s
```
The same read operation takes only 3s this time.

The great speedup attributes to the powerful caching capability provided by Alluxio. That means that once you access some remote file, it will be cached in Alluxio, and your next following operations will enjoy a local access instead of a remote one, and thus a great speedup.

> Note: Time spent for the test job depends on your network environment. If it takes too long for you to complete the job, changing a mirror or some smaller file might help.

## Clean Up
```shell
$ kubectl delete -f .
```
