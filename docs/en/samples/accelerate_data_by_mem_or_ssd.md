# DEMO - Example for Accelerate Data Access by MEM or SSD

This demo introduces examples for accelerate data by memory or ssd.
Fluid supports different speed up options such as memory, ssd, hdd and so on. 
We give an example for accelerate data by mem or ssd using AlluxioRumtime.

## Prerequisites
Before everything we are going to do, please refer to [Installation Guide](../userguide/install.md) to install Fluid on your Kubernetes Cluster, and make sure all the components used by Fluid are ready like this:
```shell
$ kubectl get pod -n fluid-system
NAME                                        READY   STATUS    RESTARTS      AGE
alluxioruntime-controller-7c54d9c76-vsrxg   1/1     Running   2 (17h ago)   18h
csi-nodeplugin-fluid-ggtjp                  2/2     Running   0             18h
csi-nodeplugin-fluid-krkbz                  2/2     Running   0             18h
dataset-controller-bdfbccd8c-8zds6          1/1     Running   0             18h
fluid-webhook-5984784577-m2xr4              1/1     Running   0             18h
fluidapp-controller-564dcd469-8dggv         1/1     Running   0             18h
```
## Example
[Alluxio](https://github.com/Alluxio/alluxio) supports tieredstores to store cached data in different location, for example different directories with different storage types.
Fluid leverages tieredstores of Alluxio to achieve accelerating by mem or ssd.

### Accelerate data by Mem
**Set Up Workspace**
```shell
$ mkdir <any-path>/mem
$ cd <any-path>/mem
```

Here is an typical example for accelerating data by **MEM** using AlluxioRuntime:
```yaml
cat<<EOF >runtime-mem.yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: hbase-mem
spec:
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 2Gi
EOF
```
> Note that `mediumtype` is `MEM`，which means accelerate data by mem.  
> `quota: 2Gi` specifies maximium cache capacity.

Create the corresponding dataset bound to the above AlluxioRuntime:
```yaml
cat<<EOF >dataset-mem.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: hbase-mem
spec:
  mounts:
    - mountPoint: https://downloads.apache.org/hbase/stable/
      name: hbase-mem
EOF
```
```shell
$ kubectl create -f dataset-mem.yaml
$ kubectl create -f runtime-mem.yaml
```

**data warm-up**（more details about data warmup please refer to [data warmup](./data_warmup.md)）：
```yaml
cat<<EOF >dataload-mem.yaml
apiVersion: data.fluid.io/v1alpha1
kind: DataLoad
metadata:
  name: hbase-dataload
spec:
  dataset:
    name: hbase-mem
    namespace: default
EOF
```
```shell
$ kubectl create -f dataload-mem.yaml
```

Wait a moment and data has all been loaded into the cache：
```shell
$ kubectl get dataset
NAME        UFS TOTAL SIZE   CACHED      CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
hbase-mem   569.12MiB        569.12MiB   2.00GiB          100.0%              Bound   5m15s
```

**Create a job to test accelerate data by mem：**
```yaml
cat<<EOF >app-mem.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: fluid-mem-copy-test
  labels:
    fluid.io/dataset.hbase-mem.sched: required
spec:
  template:
    spec:
      restartPolicy: OnFailure
      containers:
        - name: busybox
          image: busybox
          command: ["/bin/sh"]
          args: ["-c", "set -x; time cp -r /data/hbase-mem ./"]
          volumeMounts:
            - mountPath: /data
              name: hbase-vol
      volumes:
        - name: hbase-vol
          persistentVolumeClaim:
            claimName: hbase-mem
EOF
```
```shell
$ kubectl apply -f app-mem.yaml
```
Under the hood, the test job executes a shell command `time cp -r /data/hbase ./` and prints its result.
Wait for a while and make sure the job has completed. You can check its runnning status by:
```shell
$ kubectl get pod
NAME                              READY   STATUS      RESTARTS   AGE
fluid-mem-copy-test-r5vqg         0/1     Completed   0          18s
...
------
$ kubectl logs fluid-mem-copy-test-r5vqg
+ time cp -r /data/hbase-mem ./
real    0m 4.22s
user    0m 0.00s
sys     0m 1.34s
```
The read option using memory accelerate uses 4.22s.

**Clean up:**
```shell
$ kubectl delete -f .
```

### Accelerate data by SSD
**Set Up Workspace**
```
$ mkdir <any-path>/ssd
$ cd <any-path>/ssd
```

Here is an typical example for accelerating data by **SSD** using AlluxioRuntime:
```yaml
cat<<EOF >runtime-ssd.yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: hbase-ssd
spec:
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: SSD
        path: /mnt/ssd
        quota: 2Gi
EOF
```
Note that `mediumtype` is `SSD`，which means accelerate data by SSD.

Create the corresponding dataset bound to the above AlluxioRuntime:
```yaml
cat<<EOF >dataset-ssd.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: hbase-ssd
spec:
  mounts:
    - mountPoint: https://downloads.apache.org/hbase/stable/
      name: hbase-ssd
EOF
```
```shell
$ kubectl create -f runtime-ssd.yaml
$ kubectl create -f dataset-ssd.yaml
```


**data warmup：**
```yaml
cat<<EOF >dataload-ssd.yaml
apiVersion: data.fluid.io/v1alpha1
kind: DataLoad
metadata:
  name: hbase-dataload
spec:
  dataset:
    name: hbase-ssd
    namespace: default
EOF
```
```shell
$ kubectl create -f dataload-ssd.yaml
```

Wait a moment and data has all been loaded into the cache：
```shell
$ kubectl get dataset
NAME        UFS TOTAL SIZE   CACHED      CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
hbase-ssd   569.12MiB        569.12MiB   2.00GiB          100.0%              Bound   5m28s
```

**Create a job to test accelerate data by ssd:**
```yaml
cat<<EOF >app-ssd.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: fluid-ssd-copy-test
  labels:
    fluid.io/dataset.hbase-ssd.sched: required
spec:
  template:
    spec:
      restartPolicy: OnFailure
      containers:
        - name: busybox
          image: busybox
          command: ["/bin/sh"]
          args: ["-c", "set -x; time cp -r /data/hbase-ssd ./"]
          volumeMounts:
            - mountPath: /data
              name: hbase-vol
      volumes:
        - name: hbase-vol
          persistentVolumeClaim:
            claimName: hbase-ssd
EOF
```
```shell
$ kubectl apply -f app-ssd.yaml
```
Wait for a while and make sure the job has completed. You can check its runnning status by:
```shell
$ kubectl get pod
NAME                              READY   STATUS      RESTARTS   AGE
fluid-ssd-copy-test-b4bwv         0/1     Completed   0          18s
...

$ kubectl logs fluid-ssd-copy-test-b4bwv
+ time cp -r /data/hbase-ssd ./
real    0m 4.84s
user    0m 0.00s
sys     0m 1.80s
```
The read option using ssd accelerate uses 4.84s.

**Clean up:**
```shell
$ kubectl delete -f .
```

More detailed Configuration about AlluxioRuntime, please refer to [Alluxio Tieredstore Configuration](./tieredstore_config.md).
