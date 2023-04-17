# DEMO - Single-Machine Multiple-Dataset Speed Up Accessing Remote Files
Powered by [Alluxio](https://www.alluxio.io) and [Fuse](https://github.com/libfuse/libfuse), Fluid provides a simple way for users to access files stored in remote filesystems, just like accessing some ordinary files in local filesystem. Fluid manages and isolates the entire life cycle of data sets, especially for short life cycle applications (e.g data analysis tasks, machine learning tasks), users can deploy them on a large scale in a cluster.

This demo aims to show you an overview of all the features mentioned above.

## Prerequisites

Before everything we are going to do, please refer to [Installation Guide](https://github.com/fluid-cloudnative/fluid/blob/master/docs/en/userguide/install.md) to install Fluid on your Kubernetes Cluster, and make sure all the components used by Fluid are ready like this:
```shell
$ kubectl get pod -n fluid-system
NAME                                  READY   STATUS    RESTARTS   AGE
alluxioruntime-controller-5b64fdbbb-84pc6   1/1     Running   0          8h
csi-nodeplugin-fluid-fwgjh                  2/2     Running   0          8h
csi-nodeplugin-fluid-ll8bq                  2/2     Running   0          8h
dataset-controller-5b7848dbbb-n44dj         1/1     Running   0          8h
```
Normally, you shall see a Pod named `controller-manager` and several Pods named `csi-nodeplugin`. The num of `csi-nodeplugin` Pods depends on how many nodes your Kubernetes cluster have, so please make sure all `csi-nodeplugin` Pods are working properly.

## Install Resources to Kubernetes

**Label a node**

```shell
$ kubectl  label node cn-beijing.192.168.0.199 fluid=multi-dataset
```
> In the next steps, we will use `NodeSelector` to manage the nodes scheduled by the Dataset. There, it is only for experiment use.

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
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: fluid
              operator: In
              values:
                - "multi-dataset"
  placement: "Shared" // set Exclusive or empty means dataset exclusive
EOF

$ cat<<EOF >dataset1.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: spark
spec:
  mounts:
    - mountPoint: https://mirrors.bit.edu.cn/apache/spark/
      name: spark
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: fluid
              operator: In
              values:
                - "multi-dataset"
  placement: "Shared" 
EOF        
```
> Notes: Here, we use THU's tuna Apache mirror site as our `mountPoint`. If your environment isn't in Chinese mainland, please replace it with `https://downloads.apache.org/hbase/stable/` and `https://downloads.apache.org/spark/`.

Here, we'd like to create a resource object with kind `Dataset`. `Dataset` is a Custom Resource Definition(CRD) defined by Fluid and used to tell Fluid where to find all the data you'd like to access. Under the hood, Fluid uses Alluxio to do some mount operations, so `mountPoint` property can be any legal UFS path acknowledged by Alluxio. Here, we use [WebUFS](https://docs.alluxio.io/os/user/stable/en/ufs/WEB.html) for its simplicity.

For more information about UFS, please refer to [Alluxio Docs - Storage Integrations](https://docs.alluxio.io/os/user/stable/en/ufs/HDFS.html).

> We use Hbase stable and Spark on a mirror site of Apache downloads as an example of remote file. It's nothing special, you can change it to any remote file you like. But please note that, if you are going to use WebUFS like we do, files on Apache sites are highly recommended because you might need some [advanced configurations](https://docs.alluxio.io/os/user/stable/en/ufs/WEB.html#configuring-alluxio) due to current implementation of WebUFS.

**Create the `Dataset` object**

```shell
$ kubectl apply -f dataset.yaml
dataset.data.fluid.io/hbase created
$ kubectl apply -f dataset1.yaml
dataset.data.fluid.io/spark created
```

**Check status of the `Dataset` object**

```shell
$ kubectl get dataset
NAME    UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE      AGE
hbase                                                                  NotBound   6s
spark                                                                  NotBound   4s
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
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 2Gi
        high: "0.95"
        low: "0.7"
EOF

$ cat<<EOF >runtime1.yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: spark
spec:
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 4Gi
        high: "0.95"
        low: "0.7"
        
EOF
```

**Create a `AlluxioRuntime` object**
```shell
$ kubectl create -f runtime.yaml
alluxioruntime.data.fluid.io/hbase created

# Pay attention to waiting for all components of Dataset hbase Running
$ kubectl get pod -o wide | grep hbase
NAME                 READY   STATUS    RESTARTS   AGE   IP              NODE                       NOMINATED NODE   READINESS GATES
hbase-master-0       2/2     Running   0          2m55s   192.168.0.200   cn-beijing.192.168.0.200   <none>           <none>
hbase-worker-0       2/2     Running   0          2m24s   192.168.0.199   cn-beijing.192.168.0.199   <none>           <none>

$ kubectl create -f runtime1.yaml
alluxioruntime.data.fluid.io/spark created
```

**Get the `AlluxioRuntime` object**

```shell
$ kubectl get alluxioruntime
NAME    MASTER PHASE   WORKER PHASE   FUSE PHASE   AGE
hbase   Ready          Ready          Ready        2m14s
spark   Ready          Ready          Ready        58s
```

`AlluxioRuntime` is another CRD defined by Fluid. An `AluxioRuntime` object describes specifications used to run an Alluxio instance.

Wait for a while, and make sure all components defined in the `AlluxioRuntime` object are ready. You shall see something like this:

```shell
$ kubectl get pod -o wide
NAME                 READY   STATUS    RESTARTS   AGE     IP              NODE                       NOMINATED NODE   READINESS GATES
hbase-master-0       2/2     Running   0          2m55s   192.168.0.200   cn-beijing.192.168.0.200   <none>           <none>
hbase-worker-0       2/2     Running   0          2m24s   192.168.0.199   cn-beijing.192.168.0.199   <none>           <none>
spark-master-0       2/2     Running   0          50s     192.168.0.200   cn-beijing.192.168.0.200   <none>           <none>
spark-worker-0       2/2     Running   0          19s     192.168.0.199   cn-beijing.192.168.0.199   <none>           <none>
```
Note that the worker and fuse components of the different Datasets above can be dispatched to the same node `cn-beijing.192.168.0.199` normally .

**Check status of the `Dataset` object again**

```shell
$ kubectl get dataset 
NAME    UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
hbase   443.89MiB        0.00B    2.00GiB          0.0%                Bound   11m
spark   1.92GiB          0.00B    4.00GiB          0.0%                Bound   9m38s
```
Because it has been bound to a successfully started AlluxioRuntime, the state of the Dataset resource object has been updated, and the value of the `PHASE` attribute has changed to the `Bound` state. The basic information about the resource object can be obtained through the above command.

**Check status of the `AlluxioRuntime` object**
```shell
$ kubectl get alluxioruntime -o wide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE   AGE
hbase   1               1                 Ready          1               1                 Ready          0             0               Ready        11m
spark   1               1                 Ready          1               1                 Ready          0             0               Ready        9m52s
```
Detailed information about the Alluxio instance is provided here.

**Check related PersistentVolume and PersistentVolumeClaim**

```shell
$ kubectl get pv
NAME    CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM           STORAGECLASS   REASON   AGE
hbase   100Gi      RWX            Retain           Bound    default/hbase                           4m55s
spark   100Gi      RWX            Retain           Bound    default/spark                           51s
```

```shell
$ kubectl get pvc
NAME    STATUS   VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS   AGE
hbase   Bound    hbase    100Gi      RWX                           4m57s
spark   Bound    spark    100Gi      RWX                           53s
```
Related PV and PVC have been created by Fluid since the `Dataset` object is ready(bounded). Workloads are now able to access remote files by mounting PVC.

## Remote File Access

**Check the app to be created**

```shell
$ cat<<EOF >nginx.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-hbase
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
  nodeName: cn-beijing.192.168.0.199

EOF

$ cat<<EOF >nginx1.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-spark
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
        claimName: spark
  nodeName: cn-beijing.192.168.0.199

EOF
```

**Run a demo app to access remote files**

```shell
$ kubectl create -f nginx.yaml
$ kubectl create -f nginx1.yaml
```

Login to nginx hbase Pod:
```shell
$ kubectl exec -it nginx-hbase -- bash
```

Check file status:
```shell
$ ls -lh /data/hbase
total 444M
-r--r----- 1 root root 193K Sep 16 00:53 CHANGES.md
-r--r----- 1 root root 112K Sep 16 00:53 RELEASENOTES.md
-r--r----- 1 root root  26K Sep 16 00:53 api_compare_2.2.6RC2_to_2.2.5.html
-r--r----- 1 root root 211M Sep 16 00:53 hbase-2.2.6-bin.tar.gz
-r--r----- 1 root root 200M Sep 16 00:53 hbase-2.2.6-client-bin.tar.gz
-r--r----- 1 root root  34M Sep 16 00:53 hbase-2.2.6-src.tar.gz
```

Login to nginx spark Pod:
```shell
$ kubectl exec -it nginx-spark -- bash
```

Check file status:
```shell
$ ls -lh /data/spark/
total 1.0K
dr--r----- 1 root root 7 Oct 22 12:21 spark-2.4.7
dr--r----- 1 root root 7 Oct 22 12:21 spark-3.0.1
$ du -h /data/spark/
999M	/data/spark/spark-3.0.1
968M	/data/spark/spark-2.4.7
2.0G	/data/spark/
```

Loginout nginx Pod:
```shell
$ exit
```

As you may have seen, all the files on the WebUFS appear no differences from any other file in the local filesystem of the nginx Pod.

## Speed Up Accessing Remote Files

To demonstrate how great speedup you may enjoy when accessing remote files, here is a demo job:

**Check the test job to be launched**

```shell
$ cat<<EOF >app.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: fluid-copy-test-hbase
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
      nodeName: cn-beijing.192.168.0.199

EOF

$ cat<<EOF >app1.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: fluid-copy-test-spark
spec:
  template:
    spec:
      restartPolicy: OnFailure
      containers:
        - name: busybox
          image: busybox
          command: ["/bin/sh"]
          args: ["-c", "set -x; time cp -r /data/spark ./"]
          volumeMounts:
            - mountPath: /data
              name: spark-vol
      volumes:
        - name: spark-vol
          persistentVolumeClaim:
            claimName: spark
      nodeName: cn-beijing.192.168.0.199

EOF
```

**Launch a test job**

```shell
$ kubectl create -f app.yaml
job.batch/fluid-copy-test-hbase created
$ kubectl create -f app1.yaml
job.batch/fluid-copy-test-spark created
```

The hbase job executes a shell command `time cp -r /data/hbase ./`, and `/data/hbase` is the location where the remote file is mounted in the pod. After the command is completed, the execution time of the command will be displayed on the terminal.

The spark job executes a shell command `time cp -r /data/spark ./`, and `/data/hbase` is the location where the remote file is mounted in the pod. After the command is completed, the execution time of the command will be displayed on the terminal.

Wait for a while and make sure the job has completed. You can check its runnning status by:

```shell
$ kubectl get pod -o wide | grep copy 
ffluid-copy-test-hbase-r8gxp   0/1     Completed   0          4m16s   172.29.0.135    cn-beijing.192.168.0.199   <none>           <none>
fluid-copy-test-spark-54q8m   0/1     Completed   0          4m14s   172.29.0.136    cn-beijing.192.168.0.199   <none>           <none>
```
If you see the above result, it means the job has been completed.

> Note: `r8gxp` in `fluid-copy-test-hbase-r8gxp` is a specifier generated by the Job we created. It's highly possible that you may have different specifier in your environment. Please remember replace it with your own specifier in the following steps

**Check running time of the test job**

```shell
$ kubectl  logs fluid-copy-test-hbase-r8gxp
+ time cp -r /data/hbase ./
real    3m 34.08s
user    0m 0.00s
sys     0m 1.24s
$ kubectl  logs fluid-copy-test-spark-54q8m
+ time cp -r /data/spark ./
real    3m 25.47s
user    0m 0.00s
sys     0m 5.48s
```

It can be seen that the first remote file read hbase took nearly 3m34s, and the spark took nearly 3m25s.

**Check status of the dataset**

```shell
$ kubectl get dataset
NAME    UFS TOTAL SIZE   CACHED      CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
hbase   443.89MiB        443.89MiB   2.00GiB          100.0%              Bound   30m
spark   1.92GiB          1.92GiB     4.00GiB          100.0%              Bound   28m
```
Now, all the remote files have been cached in Alluxio.

**Re-Launch the test job**

```shell
$ kubectl delete -f app.yaml
$ kubectl create -f app.yaml
$ kubectl delete -f app1.yaml
$ kubectl create -f app1.yaml
```

Since the remote file has been cached, the test job can be completed quickly:
```shell
$ kubectl get pod -o wide| grep fluid
fluid-copy-test-hbase-sf5md   0/1     Completed   0          53s   172.29.0.137    cn-beijing.192.168.0.199   <none>           <none>
fluid-copy-test-spark-fwp57   0/1     Completed   0          51s   172.29.0.138    cn-beijing.192.168.0.199   <none>           <none>
```

```shell
$ kubectl  logs fluid-copy-test-hbase-sf5md
+ time cp -r /data/hbase ./
real    0m 0.36s
user    0m 0.00s
sys     0m 0.36s
$ kubectl  logs fluid-copy-test-spark-fwp57
+ time cp -r /data/spark ./
real    0m 1.57s
user    0m 0.00s
sys     0m 1.57s
```
Doing the same read operation, hbase takes only 0.36s this time and spark takes only 1.57s.

The great speedup attributes to the powerful caching capability provided by Alluxio. That means that once you access some remote file, it will be cached in Alluxio, and your next following operations will enjoy a local access instead of a remote one, and thus a great speedup.
> Note: Time spent for the test job depends on your network environment. If it takes too long for you to complete the job, changing a mirror or some smaller file might help.

Also login to the host node (if possible)
```shell
$ ssh root@192.168.0.199
$ ls /dev/shm/default/
hbase  spark
$ ls -lh /dev/shm/default/hbase/alluxioworker/
total 444M
-rwxrwxrwx 1 root root 174K Oct 22 20:27 100663296
-rwxrwxrwx 1 root root 115K Oct 22 20:27 16777216
-rwxrwxrwx 1 root root 200M Oct 22 20:26 33554432
-rwxrwxrwx 1 root root 106K Oct 22 20:26 50331648
-rwxrwxrwx 1 root root 211M Oct 22 20:27 67108864
-rwxrwxrwx 1 root root  34M Oct 22 20:27 83886080
$ ls -lh /dev/shm/default/spark/alluxioworker/
total 2.0G
-rwxrwxrwx 1 root root 210M Oct 22 21:06 100663296
-rwxrwxrwx 1 root root  16M Oct 22 21:07 117440512
-rwxrwxrwx 1 root root 195M Oct 22 21:05 134217728
-rwxrwxrwx 1 root root 214M Oct 22 21:05 150994944
-rwxrwxrwx 1 root root 140M Oct 22 21:08 16777216
-rwxrwxrwx 1 root root  22M Oct 22 21:05 167772160
-rwxrwxrwx 1 root root 221M Oct 22 21:07 184549376
-rwxrwxrwx 1 root root 150M Oct 22 21:06 201326592
-rwxrwxrwx 1 root root 311K Oct 22 21:07 218103808
-rwxrwxrwx 1 root root 322K Oct 22 21:06 234881024
-rwxrwxrwx 1 root root 210M Oct 22 21:06 33554432
-rwxrwxrwx 1 root root 161M Oct 22 21:07 50331648
-rwxrwxrwx 1 root root 223M Oct 22 21:07 67108864
-rwxrwxrwx 1 root root 208M Oct 22 21:07 83886080
```
It can be seen that the block files cached by different Datasets are isolated according to the namespace and name of the Dataset.

## Clean Up

```shell
$ kubectl delete -f .
$ kubectl label node cn-beijing.192.168.0.199 fluid-
```
