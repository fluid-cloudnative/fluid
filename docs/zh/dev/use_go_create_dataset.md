## fluid使用Go客户端创建和删除Dataset以及AlluxioRuntime

假定我们要依据`dataset.yaml`和`runtime.yaml`创建和删除相应的Dataset和AlluxioRuntime对象。

**Dataset对应的dataset.yaml文件**

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: spark
spec:
  mounts:
    - mountPoint: https://mirrors.bit.edu.cn/apache/spark/
      name: spark
```

**AlluxioRuntime对应的runtime.yaml文件**

```yaml
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
        quota: 8Gi
        high: "0.95"
        low: "0.7"
  properties:
    alluxio.user.block.size.bytes.default: 256MB
    alluxio.user.streaming.reader.chunk.size.bytes: 256MB
    alluxio.user.local.reader.chunk.size.bytes: 256MB
    alluxio.worker.network.reader.buffer.size: 256MB
    alluxio.user.streaming.data.timeout: 300sec
  fuse:
    args:
      - fuse
      - --fuse-opts=kernel_cache,ro,max_read=131072,attr_timeout=7200,entry_timeout=7200,nonempty,max_readahead=0
```

**go客户端代码**

采用go客户端创建Dataset，在go代码中，将Dataset和AlluxioRuntime的各个属性按照之前的yaml文件填充完整。

>注意：这里只列举了创建Dataset和Runtime对象的部分必要属性，Dataset和AlluxioRuntime可配置的属性远不只这些，如果想要配置其它更多属性，可仿照以下例子，结合`api/v1alpha1/dataset_types.go`以及`api/v1alpha1/alluxioruntime_types.go`文件进行配置，这两个文件中的代码详细写明了Dataset和AlluxioRuntime各个属性的名称及类型。

```go
dataset := &v1alpha1.Dataset{
    ObjectMeta: metav1.ObjectMeta{
        Name: "spark",
        Namespace: "default",
    },
    Spec: v1alpha1.DatasetSpec{
        Mounts: [] v1alpha1.Mount{
            {
                MountPoint:     "https://mirrors.bit.edu.cn/apache/spark/",
                Name:           "spark",
            },
        },
    },
}

quantity := resource.MustParse("8Gi")
alluxioRuntime := &v1alpha1.AlluxioRuntime{
    ObjectMeta: metav1.ObjectMeta{
        Name: "spark",
        Namespace: "default",
    },
    Spec: v1alpha1.AlluxioRuntimeSpec{
        Replicas: 1,
        Tieredstore: v1alpha1.Tieredstore{
            Levels: []v1alpha1.Level{
                {
                    MediumType: "MEM",
                    Path:       "/dev/shm",
                    Quota:      &quantity,
                    High:       "0.95",
                    Low:        "0.7",
                },
            },
        },
        Properties: map[string]string{
            "alluxio.user.block.size.bytes.default" : "256MB",
            "alluxio.user.streaming.reader.chunk.size.bytes" : "256MB",
            "alluxio.user.local.reader.chunk.size.bytes" : "256MB",
            "alluxio.worker.network.reader.buffer.size" : "256MB",
            "alluxio.user.streaming.data.timeout" : "300sec",
        },
        Fuse: v1alpha1.AlluxioFuseSpec{
            Args: []string{
                "fuse",
                "--fuse-opts=kernel_cache,ro,max_read=131072,attr_timeout=7200,entry_timeout=7200,nonempty,max_readahead=0",
            },
        },
    },

}
```

上面的代码相当于创建了Dataset和AlluxioRuntime的对象描述，分别保存在变量`dataset`以及`alluxioRuntime`中，在获取到`controller-runtime`的`client`之后，就可以依据这两个对象描述来创建和删除Dataset和AlluxioRuntime了。

- 创建Dataset

```go
err = client.Create(context.TODO(), dataset)
if err != nil {
    // things to do after creating the Dataset successfully.
} else {
    // things to do after creating the Dataset unsuccessfully.
}
```

- 删除Dataset

```go
err = client.Delete(context.TODO(), dataset)
if err != nil {
    // things to do after deleting the Dataset successfully.
} else {
    // things to do after deleting the Dataset unsuccessfully.
}
```

- 创建AlluxioRuntime

```go
err = client.Create(context.TODO(), alluxioRuntime)
if err != nil {
    // things to do after creating the AlluxioRuntime successfully.
} else {
    // things to do after creating the AlluxioRuntime unsuccessfully.
}
```

- 删除AlluxioRuntime

```go
err = client.Delete(context.TODO(), alluxioRuntime)
if err != nil {
    // things to do after deleting the AlluxioRuntime successfully.
} else {
    // things to do after deleting the AlluxioRuntime unsuccessfully.
}
```

