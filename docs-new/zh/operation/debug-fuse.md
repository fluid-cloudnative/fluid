# FUSE Pod 问题诊断

## 概要

如果应用Pod由于CSI Plugin的问题处于长时间处于`ContainerCreating`或者`Pending`的状态无法创建，这通常是由同一个节点Fuse Pod状态不正常导致的。可以按照以下步骤进行排查

###  查看应用Pods事件

调试 Fuse 的第一步是查看应用 Pod 信息。用如下命令查看 Pod 的当前状态和最近的事件：

```shell
kubectl describe pods ${POD_NAME}
```

类似以下信息，但是可能不尽相同,但是都指向FailedMount原因：

```shell
kubectl describe po nginx-0
...
Events:
  Type     Reason       Age   From               Message
  ----     ------       ----  ----               -------
  Normal   Scheduled    30s   default-scheduler  Successfully assigned default/nginx-0 to testnode
  Warning  FailedMount  1s    kubelet            MountVolume.MountDevice failed for volume "default-shared-data" : rpc error: code = Unknown desc = fuse pod on node testnode is not ready
```

###  获取该应用Pod对应Fuse Pod的信息

此时需要查看此节点的Fuse Pod的状态，首先用如下命令获取该当前应用Pod所在的节点信息, NODE列对应的输出就是应用Pod所在的节点


```shell
kubectl get pods ${POD_NAME} -owide
```

例如，以下Fuse Pod所在的节点为testnode:

```shell
kubectl get pods  nginx-0 -owide
NAME      READY   STATUS              RESTARTS   AGE   IP       NODE                       NOMINATED NODE   READINESS GATES
nginx-0   0/1     ContainerCreating   0          17m   <none>   testnode   <none>           <none>
```


定位该Fuse Pod具体信息

```shell
 kubectl get po -owide | grep ${NODE_NAME} | grep -i fuse
```


例如，以下例子中，该Fuse Pod为 `shared-data-alluxio-fuse-w6lcp`， 可以看到该Pod处于失败的状态

```shell
kubectl get po -owide | grep testnode | grep fuse
shared-data-alluxio-fuse-w6lcp   0/1     CrashLoopBackOff    10         29m   192.168.0.233   testnode   <none>           <none>
```

###  排查该Fuse Pod

可以按照[Kubernetes文档](https://kubernetes.io/zh/docs/tasks/debug-application-cluster/debug-running-pod/)排查该Fuse Pod的问题。

