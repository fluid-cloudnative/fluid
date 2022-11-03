# DEMO - Example for Accelerate Data Access by MEM or SSD

Fluid支持使用不同的数据加速访问选项，如内存，SSD或HDD等。
本文档使用AlluxioRumtime提供一个简单的示例区分使用内存或者SSD来加速数据访问。


## Prerequisites
在运行该示例之前，请参考[安装文档](https://github.com/fluid-cloudnative/fluid/blob/master/docs/zh/userguide/install.md)完成安装，并检查Fluid各组件正常运行：
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
## 示例
[Alluxio](https://github.com/Alluxio/alluxio)支持多层存储并且可以将缓存数据存储在不同位置。
Fluid利用Alluxio支持多层存储的特性，实现通过不同介质（内存、SSD或HDD）来加速缓存数据访问。


### 使用内存加速数据访问
创建内存加速示例目录：
```
$ mkdir <any-path>/mem
$ cd <any-path>/mem
```
这里通过一个例子来演示使用AlluxioRuntime通过内存加速数据：
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
> 注意这里`mediumtype`类型为`MEM`，即通过内存来加速数据访问。  
> `quota: 2Gi`指最大缓存容量。

创建相应的dataset与上述AlluxioRuntime绑定：
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
创建dataset和runtime：
```shell
$ kubectl create -f dataset-mem.yaml
$ kubectl create -f runtime-mem.yaml
```

进行数据预热（详见[数据预加载](./data_warmup.md)）：
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
执行数据预热：
```shell
$ kubectl create -f dataload-mem.yaml
```
此时数据已经全部加载到缓存中：
```shell
$ kubectl get dataset
NAME        UFS TOTAL SIZE   CACHED      CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
hbase-mem   569.12MiB        569.12MiB   2.00GiB          100.0%              Bound   5m15s
```


创建作业测试内存加速效果：
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

执行作业查看内存加速效果：
```shell
$ kubectl apply -f app-mem.yaml
```
测试作业执行shell命令`time cp -r /data/hbase ./ `并打印结果。

查看测试作业完成时间：
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
可以看出使用内存加速数据读取需要4.22s.

清理环境：
```shell
$ kubectl delete -f .
```


### Accelerate data by SSD
创建SSD加速示例目录：
```
$ mkdir <any-path>/ssd
$ cd <any-path>/ssd
```
这里通过一个例子来演示使用AlluxioRuntime通过SSD加速数据：
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
注意这里`mediumtype`类型为`SSD`，即通过SSD来加速数据访问。

创建相应的dataset与上述AlluxioRuntime绑定：
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
创建dataset和runtime：
```shell
$ kubectl create -f runtime-ssd.yaml
$ kubectl create -f dataset-ssd.yaml
```


进行数据预热：
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
执行数据预测：
```shell
$ kubectl create -f dataload-ssd.yaml
```

数据已经全部加载到缓存中：
```shell
$ kubectl get dataset
NAME        UFS TOTAL SIZE   CACHED      CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
hbase-ssd   569.12MiB        569.12MiB   2.00GiB          100.0%              Bound   5m28s
```

创建作业测试SSD加速效果：
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
执行作业:
```shell
$ kubectl apply -f app-ssd.yaml
```

查看结果：
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
使用SSD加速数据读取需要4.84s，慢于内存加速。

清理环境：
```shell
$ kubectl delete -f .
```

更多关于AlluxioRuntime的配置可以参见[Alluxio Tieredstore Configuration](./tieredstore_config.md)。