# 示例 - 基于Restful的文件系统客户端

AlluxioRuntime提供了基于[Restful API](https://docs.alluxio.io/os/user/stable/en/api/FS-API.html)的访问接口，方便用户通过Python，Java和Golang等语言进行二次开发。AlluxioRuntime也提供了这方面的支持，该功能默认并没有开启。可以通过CRD的声明打开。


## 运行示例

1. 创建Dataset和AlluxioRuntime资源对象
```yaml
$ cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: spark
spec:
  mounts:
    - mountPoint: https://mirrors.bit.edu.cn/apache/spark/
      name: spark
---
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
        quota: 1Gi
        high: "0.95"
        low: "0.7"
  properties:
    alluxio.user.streaming.data.timeout: 300sec
  apiGateway:
    enabled: true
EOF
```

> 注意： 只需要将apiGateway设置为true，就代表开启该能力


2. 查询访问端点

```
$  kubectl get alluxioruntimes.data.fluid.io  -owide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE   API GATEWAY                    AGE
spark   1               1                 Ready          1               1                 Ready          0             0               Ready        spark-master-0.default:20009   110s
```

可以看到API Gateway的访问地址为spark-master-0.default:20009。 您可以通过这个地址进行访问。