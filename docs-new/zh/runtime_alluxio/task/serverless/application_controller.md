# 示例 - 如何保障 Fluid 的 Serverless 任务顺利完成

## 背景介绍

在 Serverless 场景中， Job 等 Workload，当 Pod 的 user container 完成任务并退出后，需要 Fuse Sidecar 也可以主动退出，
从而使 Job Controller 能够正确判断 Pod 所处的完成状态。然而，fuse container 自身并没有退出机制，Fluid Application Controller 会检测集群中带 fluid label 的 Pod， 
在 user container 退出后，将 fuse container 正常退出，以达到 Job 完成的状态。

## 安装

您可以从 [Fluid Releases](https://github.com/fluid-cloudnative/fluid/releases) 下载最新的 Fluid 安装包。
再参考 [安装文档](../guide/install.md) 完成安装。并检查 Fluid 各组件正常运行（这里以 JuiceFSRuntime 为例）：

```shell
$ kubectl -n fluid-system get po
NAME                                         READY   STATUS    RESTARTS   AGE
dataset-controller-86768b56fb-4pdts          1/1     Running   0          36s
fluid-webhook-f77465869-zh8rv                1/1     Running   0          62s
fluidapp-controller-597dbd77dd-jgsbp         1/1     Running   0          81s
juicefsruntime-controller-65d54bb48f-vnzpj   1/1     Running   0          99s
```

通常来说，你会看到一个名为 `dataset-controller` 的 Pod、一个名为 `juicefsruntime-controller` 的 Pod、一个名为 `fluid-webhook` 的 Pod 和一个名为 `fluidapp-controller` 的 Pod。

## 运行示例

**为 namespace 开启 webhook**

Fluid webhook 提供了在 Serverless 场景中为 pod 注入 FUSE Sidecar 的功能，为了开启该功能，需要将对应的 namespace 打上 `fluid.io/enable-injection=true` 的标签。操作如下：

```shell
$ kubectl patch ns default -p '{"metadata": {"labels": {"fluid.io/enable-injection": "true"}}}'
namespace/default patched
$ kubectl get ns default --show-labels
NAME      STATUS   AGE     LABELS
default   Active   4d12h   fluid.io/enable-injection=true,kubernetes.io/metadata.name=default
```

**创建 dataset 和 runtime**

针对不同类型的 runtime 创建相应的 Runtime 资源，以及同名的 Dataset。这里以 JuiceFSRuntime 为例，具体可参考 [文档](./juicefs_runtime.md)，如下：

```shell
$ kubectl get juicefsruntime
NAME      WORKER PHASE   FUSE PHASE   AGE
jfsdemo   Ready          Ready        2m58s
$ kubectl get dataset
NAME      UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
jfsdemo   [Calculating]    N/A                       N/A                 Bound   2m55s
```

**创建 Job 资源对象**

在 Serverless 场景使用 Fluid，需要在应用 Pod 中添加 `serverless.fluid.io/inject: "true"` label。如下：

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

**查看 job 是否完成**

```shell
$ kubectl get job
NAME       COMPLETIONS   DURATION   AGE
demo-app   1/1           14s        46s
$ kubectl get po
NAME               READY   STATUS      RESTARTS      AGE
demo-app-wdfr8     0/2     Completed   0             25s
jfsdemo-worker-0   1/1     Running     0             14m
```

可以看到，job 已经完成，其 pod 有两个 container，均已完成。
