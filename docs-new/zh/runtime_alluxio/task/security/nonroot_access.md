# 示例 - 使用Fluid访问非root用户的数据

如果用户的数据只能以特定uid访问时，需要通过设置Runtime的RunAs参数指定特定用户运行分布式数据缓存引擎以访问底层数据。

本文档将通过一个简单的例子演示上述特性

## 前提条件

- [Fluid](https://github.com/fluid-cloudnative/fluid)(version >= 0.3.0)

请参考[Fluid安装文档](../guide/install.md)完成安装

## 运行示例

**创建一个非root用户**
```
$ groupadd -g 1201 fluid-user-1 && \
useradd -u 1201 -g fluid-user-1 fluid-user-1
```
上述命令创建了一个非root用户`fluid-user-1`

**创建属于该用户的目录**
```
$ mkdir -p /mnt/nonroot/user1_data && \
echo "This is fluid-user-1's data" > /mnt/nonroot/user1_data/data1 && \
chown -R fluid-user-1:fluid-user-1 /mnt/nonroot/user1_data && \
chmod -R 0750 /mnt/nonroot/user1_data
```
上述命令在`/mnt/nonroot`目录下创建了属于`fluid-user-1`的目录`user1_data`, 我们将以`user1_data`目录下的`data1`文件来模拟专属于`fluid-user-1`的数据

```
$ ls -ltR /mnt/nonroot
```
使用上述命令，你将看到以下结果：
```
/mnt/nonroot/:
total 4
drwxr-x--- 2 fluid-user-1 fluid-user-1 4096 9月  27 16:45 user1_data

/mnt/nonroot/user1_data:
total 4
-rwxr-x--- 1 fluid-user-1 fluid-user-1 28 9月  27 16:45 data1
```

**创建Dataset和AlluxioRuntime资源对象**
```yaml
$ cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: nonroot
spec:
  mounts:
    # 指定刚才建立的目录作为挂载位置
    - mountPoint: local:///mnt/nonroot/
      name: nonroot
  # 确保数据缓存被放置在存在/mnt/nonroot目录的结点上
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: nonroot
              operator: In
              values:
                - "true"
---
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: nonroot
spec:
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: SSD
        path: /var/lib/docker/alluxio
        quota: 2Gi
        high: "0.95"
        low: "0.7"
  # 以fluid-user-1的用户身份启动Alluxio
  runAs:
    uid: 1201
    gid: 1201
    user: fluid-user-1
    group: fluid-user-1
  fuse:
    args:
    - fuse
    - --fuse-opts=kernel_cache,ro,max_read=131072,attr_timeout=7200,entry_timeout=7200,max_readahead=0
EOF
```

在上述yaml配置文件中，我们将以挂载主机目录的方式挂载我们刚才创建的目录(`/mnt/nonroot`)，更多有关Fluid挂载主机目录的信息，请参考[示例 - 用Fluid加速主机目录](./hostpath.md)

另外，在`spec.runAs`中我们设置了`uid`等用户信息，这意味着我们将以`fluid-user-1`的用户身份启动缓存引擎，提供分布式缓存能力

**标记结点**
```
$ kubectl label node <node> nonroot=true
```

使用`nonroot=true`对刚才创建的`/mnt/nonroot`目录所在结点进行标记，确保缓存引擎在该结点上启动，使得其能够正确挂载指定的主机目录

**创建Dataset和AlluxioRuntime资源对象**
```
$ kubectl create -f dataset.yaml
```

**查看PV,PVC**
```
$ kubectl get pv,pvc
```

待缓存引擎正常启动后，上述命令将得到如下结果：
```
NAME                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM             STORAGECLASS   REASON   AGE
persistentvolume/nonroot   100Gi      RWX            Retain           Bound    default/nonroot                           3m18s

NAME                            STATUS   VOLUME    CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/nonroot   Bound    nonroot   100Gi      RWX                           3m18s
```


**创建应用使用Dataset**
```yaml
$ cat<<EOF >app.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
    - name: nginx
      image: nginx
      imagePullPolicy: IfNotPresent
      volumeMounts:
        - mountPath: /data
          name: nonroot-vol
      securityContext:
        runAsUser: 1201
        runAsGroup: 1201
      command:
        - tail
        - -f
        - /dev/null 
  volumes:
    - name: nonroot-vol
      persistentVolumeClaim:
        claimName: nonroot
EOF
```
上述配置意味着我们将以`uid`为1201的用户身份启动该应用, 并通过Fluid创建出的PVC将数据挂载到Pod上

```
$ kubectl create -f app.yaml
```

**登录应用**
```
$ kubectl exec -it nginx -- bash
```

```
$ id
```
上述命令将得到以下结果：
```
uid=1201 gid=1201 groups=1201
```
这表明我们以`uid`为1201的用户身份启动了该应用

**访问数据**

```
$ ls -ltR /data
```
上述命令将得到以下结果：
```
/data/:
total 1
drwxr-xr-x 1 root root 1 Sep 27 08:45 nonroot

/data/nonroot:
total 1
drwxr-x--- 1 1201 1201 1 Sep 27 08:45 user1_data

/data/nonroot/user1_data:
total 1
-rwxr-x--- 1 1201 1201 28 Sep 27 08:45 data1
```

可以看到，Fluid能够以**透传**的方式将所属某个非root用户的数据暴露给需要这些数据的应用，用户数据的各文件信息不会发生改变

当然，该用户可以自由地访问这些数据：

```
$ cat /data/nonroot/user1_data/data1
```

上述命令将得到以下结果：
```
This is fluid-user-1's data
```

## 环境清理

```
$ kubectl delete -f .
$ rm -rf /mnt/nonroot
$ userdel fluid-user-1
```
