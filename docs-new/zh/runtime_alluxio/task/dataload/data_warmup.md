# 示例 - 数据预加载
[![](https://fluid-imgs.oss-cn-shanghai.aliyuncs.com/public/imgs/dataWarmup.jfif)](https://fluid-imgs.oss-cn-shanghai.aliyuncs.com/public/video/dataWarmup.mp4)

为了保证应用在访问数据时的性能，可以通过**数据预加载**提前将远程存储系统中的数据拉取到靠近计算结点的分布式缓存引擎中，使得消费该数据集的应用能够在首次运行时即可享受到缓存带来的加速效果。

为此，我们提供了DataLoad CRD, 该CRD让你可通过简单的配置就能完成整个数据预加载过程，并对数据预加载的许多行为进行自定义控制。

本文档通过以下两个例子演示了DataLoad CRD的使用方法：
- [DataLoad快速使用](#dataload快速使用)
- [DataLoad进阶配置](#dataload进阶配置)

## 前提条件

- [Fluid](https://github.com/fluid-cloudnative/fluid)(version >= 0.4.0)

请参考[Fluid安装文档](../guide/install.md)完成安装


## 新建工作环境
```
$ mkdir <any-path>/warmup
$ cd <any-path>/warmup
```

## DataLoad快速使用

**配置待创建的Dataset和Runtime对象**

```yaml
cat << EOF > dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: spark
spec:
  mounts:
    - mountPoint: https://mirrors.tuna.tsinghua.edu.cn/apache/spark/
      name: spark
---
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: spark
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

> 注意: 上述`mountPoint`中使用了Apache清华镜像源进行演示，如果你的环境位于海外，请更换为`https://downloads.apache.org/spark/`进行尝试

在这里，我们将要创建一个kind为`Dataset`的资源对象(Resource object)。`Dataset`是Fluid所定义的一个Custom Resource Definition(CRD)，该CRD被用来告知Fluid在哪里可以找到你所需要的数据。Fluid将该CRD对象中定义的`mountPoint`属性挂载到Alluxio之上，因此该属性可以是任何合法的能够被Alluxio识别的UFS地址。在本示例中，为了简单，我们使用[WebUFS](https://docs.alluxio.io/os/user/stable/cn/ufs/WEB.html)进行演示。

更多有关UFS的信息，请参考[Alluxio文档-底层存储系统](https://docs.alluxio.io/os/user/stable/cn/ufs/OSS.html)部分。

> 本示例将以Apache镜像站点上的Spark相关资源作为演示中使用的远程文件。这个选择并没有任何特殊之处，你可以将这个远程文件修改为任意你喜欢的远程文件。但是，如果你想要和我们一样使用WebUFS进行操作的话，最好还是选择一个Apache镜像源站点( e.g. [清华镜像源](https://mirrors.tuna.tsinghua.edu.cn/apache) )，因为根据目前WebUFS的实现，如果你选择其他更加复杂的网页作为WebUFS，你可能需要进行更多[更复杂的配置](https://docs.alluxio.io/os/user/stable/cn/ufs/WEB.html#%E9%85%8D%E7%BD%AEalluxio) 

**创建Dataset和Runtime对象**

```
kubectl create -f dataset.yaml
```

**等待Dataset和Runtime准备就绪**

```
kubectl get datasets spark
```

如果看到类似以下结果，说明Dataset和Runtime均已准备就绪：
```
NAME    UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
spark   1.92GiB          0.00B    4.00GiB          0.0%                Bound   4m4s
```

**配置待创建的DataLoad对象**

```yaml
cat <<EOF > dataload.yaml
apiVersion: data.fluid.io/v1alpha1
kind: DataLoad
metadata:
  name: spark-dataload
spec:
  dataset:
    name: spark
    namespace: default
EOF
```

`spec.dataset`指明了需要进行数据预加载的目标数据集，在该例子中，我们的数据预加载目标为`default`命名空间下名为`spark`的数据集，如果该配置与你所在的实际环境不符，请根据你的实际环境对其进行调整。**注意** 你的DataLoad的namespace须和你的dataset的namespace保持一致。

**默认情况下，上述DataLoad配置将会尝试加载整个数据集中的全部数据**，如果你希望进行更细粒度的控制(例如：仅加载数据集下指定路径的数据)，请参考[DataLoad进阶配置](#dataload进阶配置)

**创建DataLoad对象**

```
kubectl create -f dataload.yaml
```


**查看创建的DataLoad对象状态**
```
kubectl get dataload spark-dataload
```

上述命令会得到类似以下结果：
```
NAME             DATASET   PHASE     AGE
spark-dataload   spark     Loading   2m13s
```

你也可以通过`kubectl describe`获取有关该DataLoad的更多详细信息：

```
kubectl describe dataload spark-dataload
```

得到以下结果：

```
Name:         spark-dataload
Namespace:    default
Labels:       <none>
Annotations:  <none>
API Version:  data.fluid.io/v1alpha1
Kind:         DataLoad
...
Spec:
  Dataset:
    Name:       spark
    Namespace:  default
Status:
  Conditions:
  Phase:  Loading
Events:
  Type    Reason              Age   From      Message
  ----    ------              ----  ----      -------
  Normal  DataLoadJobStarted  80s   DataLoad  The DataLoad job spark-dataload-loader-job started
```

上述数据加载过程根据你所在的网络环境不同，可能会耗费数分钟

**等待数据加载过程完成**

```
kubectl get dataload spark-dataload
```

你会看到该DataLoad的`Phase`状态已经从`Loading`变为`Complete`，这表明整个数据加载过程已经完成
```
NAME             DATASET   PHASE      AGE
spark-dataload   spark     Complete   5m17s
```

此时再次查看Dataset对象的缓存状态：
```
kubectl get dataset spark
```

可发现，远程存储系统中的全部数据均已成功缓存到分布式缓存引擎中
```
NAME    UFS TOTAL SIZE   CACHED    CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
spark   1.92GiB          1.92GiB   4.00GiB          100.0%              Bound   7m41s
```

## DataLoad进阶配置

除了上述示例中展示的数据预加载功能外，通过一些简单的配置，你可以对数据预加载进行更加细节的调整，这些调整包括：
- 指定一个或多个数据集子目录进行加载
- 设置数据加载时的缓存副本数量
- 数据加载前首先进行元数据同步

### 指定一个或多个数据集子目录进行加载

进行数据加载时可以加载指定的子目录(或文件)，而不是整个数据集，例如：

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: DataLoad
metadata:
  name: spark-dataload
spec:
  dataset:
    name: spark
    namespace: default
  target:
    - path: /spark/spark-2.4.7
    - path: /spark/spark-3.0.1/pyspark-3.0.1.tar.gz
```

上述DataLoad仅会加载`/spark/spark-2.4.7`目录下的全部文件，以及`/spark/spark-3.0.1/pyspark-3.0.1.tar.gz`文件

### 设置数据加载时的缓存副本数量

进行数据加载时,你也可以通过配置控制加载的数据副本数量，例如：

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: DataLoad
metadata:
  name: spark-dataload
spec:
  dataset:
    name: spark
    namespace: default
  target:
    - path: /spark/spark-2.4.7
      replicas: 1
    - path: /spark/spark-3.0.1/pyspark-3.0.1.tar.gz
      replicas: 2
```
上述DataLoad在进行数据加载时，对于`/spark/spark-2.4.7`目录下的文件仅会在分布式缓存引擎中保留**1份**数据缓存副本，而对于文件`/spark/spark-3.0.1/pyspark-3.0.1.tar.gz`，分布式缓存引擎将会保留**2份**缓存副本。

### 数据加载前首先进行元数据同步

在许多场景下，底层存储系统中的文件可能发生了变化，对于分布式缓存引擎来说，需要重新进行文件元信息的同步才能感知到底层存储系统中的变化。因此在进行数据加载前，你也可以通过设置DataLoad对象的`spec.loadMetadata`来预先完成元信息的同步操作，例如：

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: DataLoad
metadata:
  name: spark-dataload
spec:
  dataset:
    name: spark
    namespace: default
  loadMetadata: true
  target:
    - path: /
      replicas: 1
```

> 注意：元数据的同步会使整个数据加载过程变长，如非必要，我们不推荐开启该配置项

## 环境清理
```shell
$ kubectl delete -f .
```
