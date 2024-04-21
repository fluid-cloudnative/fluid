# How to Use Vineyard Runtime in Fluid

## Background

Vineyard is an open-source in-memory data management system designed to provide high-performance data sharing and data exchange. Vineyard achieves zero-copy data sharing by storing data in shared memory, providing high-performance data sharing and data exchange capabilities.

For more information on how to use Vineyard, see the [Vineyard Quick Start Guide](https://v6d.io/notes/getting-started.html).

## Install Fluid

Refer to the [Installation Documentation](../../userguide/install.md) to complete the installation.

## Create Vineyard Runtime and Dataset

```shell
$ cat <<EOF | kubectl apply -f -
apiVersion: data.fluid.io/v1alpha1
kind: VineyardRuntime
metadata:
  name: vineyard
spec:
  replicas: 2
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

In VineyardRuntime:

- `spec.replicas`: Specifies the number of Vineyard Workers.
- `spec.tieredstore`: Specifies the storage configuration of Vineyard Worker, including storage levels and storage capacity. Here, a memory storage level with a capacity of 20Gi is configured.

In Dataset:

- `metadata.name`: Specifies the name of the Dataset, which must be consistent with the `metadata.name` in VineyardRuntime.

Check if the `Vineyard Runtime` is created successfully:

```shell
$ kubectl get vineyardRuntime vineyard 
NAME       MASTER PHASE   WORKER PHASE   FUSE PHASE   AGE
vineyard   Ready          PartialReady   Ready        3m4s
```

Then look at the status of the `Vineyard Dataset` and notice that it has been bound to the `Vineyard Runtime`:

```shell
$ kubectl get dataset vineyard
NAME       UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
vineyard                                                                  Bound   3m9s
```

## Create an Application Pod and Mount the Vineyard Dataset

The default mount path of the Vineyard Dataset is `/var/run/vineyard`. Then you can
connect to the vineyard worker by the default configurations. If you change the mount path, you need to specify the configurations when connecting to the vineyard worker.

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

Check if the Pod is created successfully:

```shell
$ kubectl get pod demo-app
NAME       READY   STATUS    RESTARTS   AGE
demo-app   1/1     Running   0          25s
```

Check the status of Vineyard FUSE:

```shell
$ kubectl get po | grep vineyard-fuse
vineyard-fuse-9dv4d                    1/1     Running   0               1m20s
```

Check the vineyard client configurations
have been mounted to the Pod:

```shell
$ kubectl exec demo-app -- ls /data/
rpc-conf
vineyard-config.yaml
```

```shell
$ kubectl exec demo-app -- cat /data/vineyard-config.yaml
Vineyard:
  IPCSocket: vineyard.sock
  RPCEndpoint: vineyard-worker-0.vineyard-worker.default:9600,vineyard-worker-1.vineyard-worker.default:9600
```

Connect to the vineyard worker:

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

## Data sharing between pods with Vineyard Runtime

In this section, we will show you how to share data between different workloads with Vineyard Runtime. Assume we have two pods, one is a producer and the other is a consumer. 

Create the producer pod:

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
        vineyard.put(pd.DataFrame(np.random.randn(100, 4), columns=list('ABCD')), persist=True, name="test_dataframe")
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

Then create the consumer pod:

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

Check the logs of the consumer pod:

```shell
$  kubectl logs consumer --tail 6
A    2.260771
B   -2.690233
C   -1.523646
D    7.208424
dtype: float64
(1, 1.2345000505447388, 'xxxxabcd')
```
