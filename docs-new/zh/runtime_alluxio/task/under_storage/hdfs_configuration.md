## 示例 - Fluid访问HDFS所需的特殊配置

如果选择HDFS作为Alluxio的底层存储系统,并且HDFS具有特殊的配置时,Alluxio同样需要进行额外配置,以使得Alluxio能够
正常访问挂载的HDFS集群.

本文档展示了如何在Fluid中以声明式的方式完成Alluxio所需的特殊配置. 更多信息请参考[在HDFS上配置Alluxio](https://docs.alluxio.io/os/user/stable/cn/ufs/HDFS.html)

## 前提条件

- [Fluid](https://github.com/fluid-cloudnative/fluid)
- 可与Kubernetes环境网络连通的HDFS集群

请参考[Fluid安装文档](../guide/install.md)完成安装

## 运行示例

**从HDFS配置文件创建ConfigMap**
```
$ kubectl create configmap <configmap-name> --from-file=/path/to/hdfs-site.xml --from-file=/path/to/core-site.xml
```

**创建Dataset资源对象**

```yaml
$ cat << EOF > dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: my-hdfs
spec:
  mounts:
    - mountPoint: hdfs://<namenode>:<port>
      name: hdfs
EOF
```

```
$ kubectl create -f dataset.yaml
```

**创建AlluxioRuntime资源对象**

```yaml
$ cat << EOF > runtime.yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: my-hdfs
spec:
  ...
  hadoopConfig: <configmap-name>
  ...
EOF
```

`configmap-name`须为之前创建的ConfigMap资源对象, 该ConfigMap必须与创建的AlluxioRuntime同处于同一Namespace. 其中必须包含以`hdfs-site.xml`或`core-site.xml`为键的键值对. Fluid会检查
该ConfigMap中的内容,并从中选取`hdfs-site.xml`以及`core-site.xml`的键值对并以文件的形式挂载到Alluxio的各个Pod中.

```
$ kubectl create -f runtime.yaml
```

**在Alluxio中查看HDFS配置文件**

待Alluxio集群启动后,ConfigMap中的HDFS配置文件将被挂载到 alluxio-master Pod 和 alluxio-worker Pod中各容器的`/hdfs-config`目录下:

```
$ kubectl exec my-hdfs-master-0 -- ls /hdfs-config
hdfs-site.xml  core-site.xml
``` 

至此, Alluxio能够根据用户指定的配置文件正常地访问HDFS集群.
