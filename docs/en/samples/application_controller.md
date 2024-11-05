# Demo - How to ensure the completion of Fluid's serverless tasks

## Background

In the serverless scenario, Workload such as Job, when the user container of the Pod completes the task and exits, the
Fuse Sidecar can also actively exit.
This enables the Job Controller to correctly determine the completion status of the Pod. However, the fuse container
itself does not have an exit mechanism, and the Fluid Application Controller will detect the pods with the fluid label
in the cluster.
After the user container exits, the fuse container is exited normally to reach the state where the job is completed.

## Installation

You can download the latest Fluid installation package
from [Fluid Releases](https://github.com/fluid-cloudnative/fluid/releases).
Refer to the [Installation Documentation](../userguide/install.md) to complete the installation. And check that the
components of Fluid are running normally (here takes JuiceFSRuntime as an example):

```shell
$ kubectl -n fluid-system get po
NAME                                         READY   STATUS    RESTARTS   AGE
dataset-controller-86768b56fb-4pdts          1/1     Running   0          36s
fluid-webhook-f77465869-zh8rv                1/1     Running   0          62s
fluidapp-controller-597dbd77dd-jgsbp         1/1     Running   0          81s
juicefsruntime-controller-65d54bb48f-vnzpj   1/1     Running   0          99s
```

Typically, you will see a Pod named `dataset-controller`, a Pod named `juicefsruntime-controller`, a Pod
named `fluid-webhook` and a Pod named `fluidapp-controller`.

## Demo

**Create dataset and runtime**

Create corresponding Runtime resources and Datasets with the same name for different types of runtimes. Take JuiceFSRuntime as an example here. For details, please refer to [Documentation](juicefs_runtime.md), as follows:

```shell
$ kubectl get juicefsruntime
NAME      WORKER PHASE   FUSE PHASE   AGE
jfsdemo   Ready          Ready        2m58s
$ kubectl get dataset
NAME      UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
jfsdemo   [Calculating]    N/A                       N/A                 Bound   2m55s
```

**Create Job**

To use Fluid in a serverless scenario, you need to add the `serverless.fluid.io/inject: "true"` and `fluid.io/managed-by: fluid` label to the application pod. as follows:

```yaml
$ cat<<EOF >sample.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: demo-app
spec:
  template:
    metadata:
      labels:
        serverless.fluid.io/inject: "true"
        fluid.io/managed-by: fluid
    spec:
      containers:
        - name: demo
          image: busybox
          args:
            - -c
            - echo $(date -u) >> /data/out.txt
          command:
            - /bin/sh
          volumeMounts:
            - mountPath: /data
              name: demo
      restartPolicy: Never
      volumes:
        - name: demo
          persistentVolumeClaim:
            claimName: jfsdemo
  backoffLimit: 4
EOF
$ kubectl create -f sample.yaml
job.batch/demo-app created
```

**Check if the Pod is completed**

```shell
$ kubectl get job
NAME       COMPLETIONS   DURATION   AGE
demo-app   1/1           14s        46s
$ kubectl get po
NAME               READY   STATUS      RESTARTS      AGE
demo-app-wdfr8     0/2     Completed   0             25s
jfsdemo-worker-0   1/1     Running     0             14m
```

It can be seen that the job has been completed, and its pod has two containers, both of which have been completed.
