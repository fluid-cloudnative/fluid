# Demo - Restful-based file system client

Alluxio provides an access interface based on the [Restful API](https://docs.alluxio.io/os/user/stable/en/api/FS-API.html) for secondary development in Python, Java and Golang. AlluxioRuntime also provides support for this, and the feature is not turned on by default. It can be turned on via a CRD declaration.


## Running

1. Create Dataset and AlluxioRuntime resource objects
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

> Note: Just set apiGateway to true to enable this capability.


2. Check Access Endpoint

```
$  kubectl get alluxioruntimes.data.fluid.io  -owide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE   API GATEWAY                    AGE
spark   1               1                 Ready          1               1                 Ready          0             0               Ready        spark-master-0.default:20009   110s
```

You can see that the API Gateway is accessed at spark-master-0.default:20009. You can access it from this address.