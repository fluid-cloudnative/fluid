## How to configure the cache size of Vineyard Fuse

Before you begin, we'd better have a basic understanding of the [Vineyard Fuse](https://github.com/v6d-io/v6d/tree/main/docker/vineyard-fluid-fuse). It mainly creates the vineyard client configurations in the application pod and defines the priority of different vineyard sockets.

For example, suppose we have enabled the cache of vineyard fuse, a local vineyardd will be started in the fuse container, and it will connect to the vineyard master. The new vineyardd will create a vineyard socket named `vineyard-local.sock` in the pod, and the vineyard fuse will also create the symbolic link `vineyard-worker.sock` to pod if the fuse pod runs on the same node as the vineyard worker. In this case, we will put the vineyard socket `vineyard-local.sock` in the first place in the vineyard client configurations YAML. When the application connects to the vineyard, it will first try to put and get data from the local vineyardd. If it's out of memory, it will fall back to the vineyard worker. Thus, we can regard the vineyard fuse as a local cache of the external vineyard workers.

## Install fluid

Refer to the [Installation Documentation](../../userguide/install.md) to complete the installation.

## Create Vineyard Runtime and set the fuse cache size to 0

If you don't set the cache size of vineyard fuse, the default value is 0. To make it
clear, we set the cache size to 0 explicitly.

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
  fuse:
    options:
      # Not to start the local vineyardd in the vineyard fuse container
      size: "0"
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
vineyard   Ready          PartialReady   Ready        59s
```

Then look at the status of the `Vineyard Dataset` and notice that it has been bound to the `Vineyard Runtime`:

```shell
$ kubectl get dataset vineyard
NAME       UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
vineyard                                                                  Bound   73s
```

### [Serverful Mode] Create an Application Pod and Mount the Vineyard Dataset

Add the label `fuse.serverful.fluid.io/inject: "true"` to make the pod scheduled to the node where the vineyard worker is running.

```shell
$ cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: demo-app
  labels:
    fuse.serverful.fluid.io/inject: "true"
spec:
  containers:
    - name: demo
      image: python:3.10
      command:
        - /bin/bash
        - -c
        - |
          pip3 install vineyard;
          sleep infinity
      volumeMounts:
        - mountPath: /var/run/vineyard
          name: client-configuration
  volumes:
    - name: client-configuration
      persistentVolumeClaim:
        claimName: vineyard
EOF
```

Then we can check the client configurations in the pod, and notice there is only one vineyard socket named `vineyard-worker.sock`.

```shell
$ kubectl exec -it demo-app -- ls /var/run/vineyard
rpc-conf  vineyard-config.yaml  vineyard-worker.sock

$ kubectl exec -it demo-app -- cat /var/run/vineyard/vineyard-config.yaml                    
Vineyard:
  IPCSocket: vineyard-worker.sock
  RPCEndpoint: vineyard-worker-0.vineyard-worker.default:9600,vineyard-worker-1.vineyard-worker.default:9600
```

Connect to the vineyard worker in the pod.

```shell
$  kubectl exec -it demo-app -- python
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

### [Serverless Mode] Create an Application Pod and Mount the Vineyard Dataset

Please note that the `serverless.fluid.io/inject` label shoule be set to `true` in the pod metadata when you want to run the application in serverless mode.

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
        - /bin/bash
        - -c
        - |
          pip3 install vineyard;
          sleep infinity
      volumeMounts:
        - mountPath: /var/run/vineyard
          name: client-configuration
  volumes:
    - name: client-configuration
      persistentVolumeClaim:
        claimName: vineyard
EOF
```

Then you can get the application pod with injected vineyard client configuration. Different from the serverful mode, the vineyard client configurations only contain the RPC endpoint of the vineyard workers.

```shell
$ kubectl get pod demo-app -oyaml
# here only show the main fields of the pod
apiVersion: v1
kind: Pod
metadata:
  labels:
    done.sidecar.fluid.io/inject: "true"
    serverless.fluid.io/inject: "true"
  name: demo-app
  namespace: default
spec:
  containers:
  - command:
    - /bin/bash
    - -c
    - |
      pip3 install vineyard;
      sleep infinity
    env:
    - name: VINEYARD_RPC_ENDPOINT
      valueFrom:
        configMapKeyRef:
          key: VINEYARD_RPC_ENDPOINT
          name: vineyard-rpc-conf
    image: python:3.10
    imagePullPolicy: IfNotPresent
    name: demo
  volumes:
  - configMap:
      defaultMode: 420
      name: vineyard-rpc-conf
    name: vineyard-rpc-conf
```

Connect to the vineyard worker in the pod.

```shell
$  kubectl exec -it demo-app -- python
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

## Create Vineyard Runtime and set the fuse cache size to 1Gi

Add the option `size: "1Gi"` to enable the cache of vineyard fuse and set the cache size to 1Gi.

```shell
$ cat <<EOF | kubectl apply -f -
apiVersion: data.fluid.io/v1alpha1
kind: VineyardRuntime
metadata:
  name: vineyard-with-fuse-cache
spec:
  replicas: 2
  tieredstore:
    levels:
    - mediumtype: MEM
      quota: 20Gi
  fuse:
    options:
      # Start the local vineyardd with 1Gi memory in the vineyard fuse container
      size: "1Gi"
---
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: vineyard-with-fuse-cache
EOF
```

### [Serverful Mode] Create an Application Pod and Mount the Vineyard Dataset

```shell
$ cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: demo-app
  labels:
    fuse.serverful.fluid.io/inject: "true"
spec:
  containers:
    - name: demo
      image: python:3.10
      command:
        - /bin/bash
        - -c
        - |
          pip3 install vineyard;
          sleep infinity
      volumeMounts:
        - mountPath: /var/run/vineyard
          name: client-configuration
  volumes:
    - name: client-configuration
      persistentVolumeClaim:
        claimName: vineyard-with-fuse-cache
EOF
```

Check the status of Vineyard FUSE:

```shell
$ kubectl get po | grep vineyard-with-fuse-cache-fuse
vineyard-with-fuse-cache-fuse-chd9w   1/1     Running   0          110s
```

Then we can check the client configurations in the pod, and notice there are two vineyard sockets named `vineyard-local.sock` and `vineyard-worker.sock`. In the client
configurations YAML, the `vineyard-local.sock` is in the first place, which means the application will first try to put and get data from the local vineyardd.

```shell
$ kubectl exec -it demo-app -- ls /var/run/vineyard
rpc-conf  vineyard-config.yaml  vineyard-local.sock  vineyard-worker.sock

# Check the vineyard client configurations
$ kubectl exec -it demo-app -- cat /var/run/vineyard/vineyard-config.yaml                    
Vineyard:
  IPCSocket: vineyard-local.sock
  RPCEndpoint: vineyard-with-fuse-cache-worker-0.vineyard-with-fuse-cache-worker.default:9600,vineyard-with-fuse-cache-worker-1.vineyard-with-fuse-cache-worker.default:9600
```

We can find if we connect to the vineyard with the default configurations, the local
vineyardd will be used first.

```shell
$  kubectl exec -it demo-app -- python
Python 3.10.14 (main, Mar 25 2024, 21:45:25) [GCC 12.2.0] on linux
Type "help", "copyright", "credits" or "license" for more information.
>>> import vineyard
>>> client = vineyard.connect()
>>> client.status
{
    instance_id: 2,
    deployment: local,
    memory_usage: 0,
    memory_limit: 1073741824, # 1Gi, not 20Gi in vineyard worker
    deferred_requests: 0,
    ipc_connections: 1,
    rpc_connections: 0
}
```

## [Serverless Mode] Create an Application Pod and Mount the Vineyard Dataset

Add the label `serverless.fluid.io/inject: "true"` to make the pod run in serverless mode.

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
        - /bin/bash
        - -c
        - |
          pip3 install vineyard;
          sleep infinity
      volumeMounts:
        - mountPath: /var/run/vineyard
          name: client-configuration
  volumes:
    - name: client-configuration
      persistentVolumeClaim:
        claimName: vineyard-with-fuse-cache
EOF
```

Different from the situation when the cache size is 0, the vineyard fuse container will be injected into the application pod.

```shell
$ kubectl get pod demo-app -oyaml
# here only show the main fields of the pod
apiVersion: v1
kind: Pod
metadata:
  labels:
    done.sidecar.fluid.io/inject: "true"
    serverless.fluid.io/inject: "true"
  name: demo-app
  namespace: default
spec:
  containers:
  - env:
    - name: MOUNT_DIR
      value: /runtime-mnt/vineyard/default/vineyard-with-fuse-cache
    - name: FUSE_DIR
      value: /runtime-mnt/vineyard/default/vineyard-with-fuse-cache/vineyard-fuse
    - name: RPC_CONF_DIR
      value: /runtime-mnt/vineyard/default/vineyard-with-fuse-cache/vineyard-fuse/rpc-conf
    - name: PRESTOP_MARKER
      value: /tmp/prestop-marker
    - name: CACHE_SIZE
      value: 1Gi
    - name: ETCD_ENDPOINT
      value: http://vineyard-with-fuse-cache-master-0.vineyard-with-fuse-cache-master.default:2379
    - name: ETCD_PREFIX
      value: /vineyard
    image: vineyardcloudnative/vineyard-fluid-fuse:v0.21.5
    imagePullPolicy: IfNotPresent
    lifecycle:
      postStart:
        exec:
          command:
          - bash
          - -c
          - time /check-mount.sh /runtime-mnt/vineyard/default/vineyard-with-fuse-cache
            vineyard-fuse  >> /proc/1/fd/1
      preStop:
        exec:
          command:
          - sh
          - -c
          - touch /tmp/prestop-marker && rm -f /runtime-mnt/vineyard/default/vineyard-with-fuse-cache/vineyard-fuse/vineyard-local.sock
            /runtime-mnt/vineyard/default/vineyard-with-fuse-cache/vineyard-fuse/vineyard-worker.sock
            && umount /runtime-mnt/vineyard/default/vineyard-with-fuse-cache/vineyard-fuse/rpc-conf
    name: fluid-fuse-0
    resources:
      requests:
        memory: 1Gi
    securityContext:
      privileged: true
    volumeMounts:
    - mountPath: /runtime-mnt/vineyard/default/vineyard-with-fuse-cache
      mountPropagation: Bidirectional
      name: vineyard-fuse-mount-0
    - mountPath: /runtime-mnt/vineyard/default/vineyard-with-fuse-cache/vineyard-fuse/rpc-conf
      name: vineyard-rpc-conf-0
    - mountPath: /check-mount.sh
      name: check-mount-0
      readOnly: true
      subPath: check-mount.sh
  - command:
    - /bin/bash
    - -c
    - |
      pip3 install vineyard;
      sleep infinity
    image: python:3.10
    imagePullPolicy: IfNotPresent
    name: demo
    volumeMounts:
    - mountPath: /var/run/vineyard
      mountPropagation: HostToContainer
      name: client-configuration
  volumes:
  - hostPath:
      path: /runtime-mnt/vineyard/default/vineyard-with-fuse-cache/vineyard-fuse
      type: ""
    name: client-configuration
  - hostPath:
      path: /runtime-mnt/vineyard/default/vineyard-with-fuse-cache
      type: DirectoryOrCreate
    name: vineyard-fuse-mount-0
  - configMap:
      defaultMode: 420
      name: vineyard-with-fuse-cache-rpc-conf
    name: vineyard-rpc-conf-0
  - configMap:
      defaultMode: 493
      name: vineyard-with-fuse-cache-vineyard-fuse-check-mount
    name: check-mount-0
```

Then we can check the client configurations in the pod, and notice there is
only one vineyard socket named `vineyard-local.sock` as it doesn't run on the same node as the vineyard worker.

```shell
$ kubectl exec -it demo-app -c demo -- ls /var/r
un/vineyard
rpc-conf  vineyard-config.yaml  vineyard-local.sock

# Check the vineyard client configurations
$ kubectl exec -it demo-app -c demo -- cat /var/
run/vineyard/vineyard-config.yaml
Vineyard:
  IPCSocket: vineyard-local.sock
  RPCEndpoint: vineyard-with-fuse-cache-worker-0.vineyard-with-fuse-cache-worker.default:9600,vineyard-with-fuse-cache-worker-1.vineyard-with-fuse-cache-worker.default:9600
```

Connect to the vineyard in the pod.

```shell
$ kubectl exec -it demo-app -c demo -- python
Python 3.10.14 (main, Mar 25 2024, 21:45:25) [GCC 12.2.0] on linux
Type "help", "copyright", "credits" or "license" for more information.
>>> import vineyard
>>> client = vineyard.connect()
kubectl exec -it demo-app -c demo -- python
Python 3.10.14 (main, Mar 25 2024, 21:45:25) [GCC 12.2.0] on linux
Type "help", "copyright", "credits" or "license" for more information.
>>> import vineyard
>>> client = vineyard.connect()
>>> client.status
{
    instance_id: 3,
    deployment: local,
    memory_usage: 0,
    memory_limit: 1073741824,
    deferred_requests: 0,
    ipc_connections: 1,
    rpc_connections: 0
}
```
