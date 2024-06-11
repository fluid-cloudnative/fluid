# How to Use Vineyard Runtime in serverless mode

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

Check if the `Vineyard Runtime` is created successfully:

```shell
$ kubectl get vineyardRuntime vineyard
NAME       MASTER PHASE   WORKER PHASE   FUSE PHASE   AGE
vineyard   Ready          Ready          Ready        84s
```

Then look at the status of the `Vineyard Dataset` and notice that it has been bound to the `Vineyard Runtime`:

```shell
$ kubectl get dataset vineyard
NAME       UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
vineyard                                                                  Bound   96s
```

## Create an Application Pod with serverless label and Mount the Vineyard Dataset

Please note that the `serverless.fluid.io/inject` label should be set to `true` in the pod metadata when you want to run the application in serverless mode. During injection, the vineyard client configurations will be mounted to the Pod as environment variables.

```shell
$ cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: demo-app
  labels:
    serverless.fluid.io/inject: "true"
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

Check whether the vineyard client configurations have been mounted to the Pod:

```shell
$ kubectl get pod demo-app -oyaml
apiVersion: v1
kind: Pod
metadata:
  name: demo-app
  labels:
    serverless.fluid.io/inject: "true"
    done.sidecar.fluid.io/inject: "true"
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
      env:
        - name: VINEYARD_RPC_ENDPOINT
          valueFrom:
            configMapKeyRef:
              name: vineyard-rpc-conf
              key: VINEYARD_RPC_ENDPOINT
  volumes:
    - name: vineyard-rpc-conf
      configMap:
        name: vineyard-rpc-conf
```

Check the vineyard client configurations in the Pod:

```shell
$ kubectl exec demo-app -- env | grep VINEYARD_RPC_ENDPOINT
VINEYARD_RPC_ENDPOINT=vineyard-worker-0.vineyard-worker.default:9600,vineyard-worker-1.vineyard-worker.default:9600
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
    instance_id: 0,
    deployment: local,
    memory_usage: 0,
    memory_limit: 21474836480,
    deferred_requests: 0,
    ipc_connections: 0,
    rpc_connections: 1
}
```
