# 如何在 Fluid 中使用 Vineyard Runtime

## 背景介绍

Vineyard 是一个开源的内存数据管理系统，旨在提供高性能的数据共享和数据交换。Vineyard 通过将数据存储在共享内存中，实现了数据的零拷贝共享，从而提供了高性能的数据共享和数据交换能力。

如何使用 Vineyard 可以参考文档 [Vineyard 快速上手指南](https://v6d.io/notes/getting-started.html)。

## 安装 Fluid

参考 [安装文档](../../userguide/install.md) 完成安装。

## 创建Vineyard Runtime 及 Dataset

```shell
$ cat <<EOF | kubectl apply -f -
apiVersion: data.fluid.io/v1alpha1
kind: VineyardRuntime
metadata:
  name: vineyard
spec:
  replicas: 2
  master:
    image: registry.aliyuncs.com/vineyard/vineyardd
    imageTag: v0.22.2
  tieredstore:
    levels:
    - mediumtype: MEM
      quota: 20Gi
---
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: vineyard
EOF
```

在 VineyardRuntime 中:

- `spec.replicas`：指定 Vineyard Worker 的副本数；
- `spec.tieredstore`：指定 Vineyard Worker 的存储配置，包括存储层级和存储容量。这里配置了一个内存存储层级，容量为 20Gi。

在 Dataset 中:

- `metadata.name`：指定 Dataset 的名称，需要与 VineyardRuntime 中的 `metadata.name` 保持一致。


检查 `Vineyard Runtime` 是否创建成功：

```shell
$ kubectl get vineyardRuntime vineyard 
NAME       MASTER PHASE   WORKER PHASE   FUSE PHASE   AGE
vineyard   Ready          PartialReady   Ready        3m4s
```

再查看 `Vineyard Dataset` 的状态，发现已经与 `Vineyard Runtime` 绑定：

```shell
$ kubectl get dataset vineyard
NAME       UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
vineyard                                                                  Bound   3m9s
```

## 创建一个应用 Pod 并挂载 Vineyard Dataset

默认情况下，Vineyard Dataset 的挂载路径为 `/var/run/vineyard`。然后您可以通过默认配置连接到 vineyard worker。如果更改挂载路径，需要在连接到 vineyard worker 时指定配置。

```shell
$ cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: demo-app
spec:
  containers:
    - name: demo
      image: python:3.10
      command:
      - bash
      - -c
      - |
        pip install vineyard;
        sleep infinity;
      volumeMounts:
        - mountPath: /var/run/vineyard
          name: client-config
  volumes:
    - name: client-config
      persistentVolumeClaim:
        claimName: vineyard
EOF
```

检查 Pod 资源对象是否创建成功：

```shell
$ kubectl get pod demo-app
NAME       READY   STATUS    RESTARTS   AGE
demo-app   1/1     Running   0          25s
```

查看 Vineyard FUSE 的状态：

```shell
$ kubectl get po | grep vineyard-fuse
vineyard-fuse-9dv4d                    1/1     Running   0               1m20s
```

检查 vineyard 客户端配置是否已经挂载到应用 Pod：

```shell
$ kubectl exec demo-app -- ls /var/run/vineyard/
rpc-conf
vineyard-config.yaml
```

```shell
$ kubectl exec demo-app -- cat /var/run/vineyard/vineyard-config.yaml
Vineyard:
  IPCSocket: vineyard.sock
  RPCEndpoint: vineyard-worker-0.vineyard-worker.default:9600,vineyard-worker-1.vineyard-worker.default:9600
```

在应用pod中连接 vineyard worker：

```shell
$ kubectl exec -it demo-app -- python
Python 3.10.14 (main, Mar 25 2024, 21:45:25) [GCC 12.2.0] on linux
Type "help", "copyright", "credits" or "license" for more information.
>>> import vineyard
>>> client = vineyard.connect()
>>> client.status
{
    instance_id: 1,
    deployment: local,
    memory_usage: 0,
    memory_limit: 21474836480,
    deferred_requests: 0,
    ipc_connections: 0,
    rpc_connections: 1
}
```

## 使用 Vineyard Runtime 在pod之间共享数据

在这个示例中，我们将展示如何在不同的工作负载之间共享数据。假设我们有两个 Pod，一个是生产者，另一个是消费者。

创建生产者 Pod：

```shell
$ cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: producer
spec:
  containers:
    - name: producer
      image: python:3.10
      command:
      - bash
      - -c
      - |
        pip install vineyard numpy pandas;
        cat << EOF >> producer.py
        import vineyard
        import numpy as np
        import pandas as pd
        rng = np.random.default_rng(seed=42)
        vineyard.put(pd.DataFrame(rng.standard_normal((100, 4)), columns=list('ABCD')), persist=True, name="test_dataframe")
        vineyard.put((1, 1.2345, 'xxxxabcd'), persist=True, name="test_basic_data_unit");
        EOF
        python producer.py;
        sleep infinity;
      volumeMounts:
        - mountPath: /var/run/vineyard
          name: client-config
  volumes:
    - name: client-config
      persistentVolumeClaim:
        claimName: vineyard
EOF
```

接下来创建消费者 Pod：

```shell
$ cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: consumer
spec:
  containers:
    - name: consumer
      image: python:3.10
      command:
      - bash
      - -c
      - |
        pip install vineyard numpy pandas;
        cat << EOF >> consumer.py
        import vineyard
        print(vineyard.get(name="test_dataframe",fetch=True).sum())
        print(vineyard.get(name="test_basic_data_unit",fetch=True))
        EOF
        python consumer.py;
        sleep infinity;
      volumeMounts:
        - mountPath: /var/run/vineyard
          name: client-config
  volumes:
    - name: client-config
      persistentVolumeClaim:
        claimName: vineyard
EOF

检查消费者 Pod 的日志：

```shell
$  kubectl logs consumer --tail 6
A    2.260771
B   -2.690233
C   -1.523646
D    7.208424
dtype: float64
(1, 1.2345000505447388, 'xxxxabcd')
```
