# Troubleshooting

## Overview

If the application Pod is in the `ContainerCreating` or `Pending` state for a long time and cannot be created due to the CSI Plugin issue, it is usually caused by the state of the Fuse Pod in the same node is not correct. You can follow these steps to troubleshoot.

###  Check Application Pods Events

The first step in debugging Fuse is to check the application Pod information. Use the following command to check the current status and recent events of the Pod.

```shell
kubectl describe pods ${POD_NAME}
```

Similar to the following information, but may not be identical, but all point to the reason for FailedMount.

```shell
kubectl describe po nginx-0
...
Events:
  Type     Reason       Age   From               Message
  ----     ------       ----  ----               -------
  Normal   Scheduled    30s   default-scheduler  Successfully assigned default/nginx-0 to testnode
  Warning  FailedMount  1s    kubelet            MountVolume.MountDevice failed for volume "default-shared-data" : rpc error: code = Unknown desc = fuse pod on node testnode is not ready
```

###  Get the information of the Fuse Pod corresponding to this application Pod

At this point you need to check the status of the Fuse Pod of this node, first use the following command to get the information of the node where the current application Pod is located, the output corresponding to the NODE column is the node where the application Pod is located.


```shell
kubectl get pods ${POD_NAME} -owide
```

For example, the following node where Fuse Pod is located is testnode:

```shell
kubectl get pods  nginx-0 -owide
NAME      READY   STATUS              RESTARTS   AGE   IP       NODE                       NOMINATED NODE   READINESS GATES
nginx-0   0/1     ContainerCreating   0          17m   <none>   testnode   <none>           <none>
```


Get this Fuse Pod's detail information:

```shell
 kubectl get po -owide | grep ${NODE_NAME} | grep -i fuse
```


In the following example, the Fuse Pod is `shared-data-alluxio-fuse-w6lcp`, and you can see that the Pod is in a failed state:

```shell
kubectl get po -owide | grep testnode | grep fuse
shared-data-alluxio-fuse-w6lcp   0/1     CrashLoopBackOff    10         29m   192.168.0.233   testnode   <none>           <none>
```

### Troubleshoot the Fuse Pod

You can follow the [Kubernetes documentation](https://kubernetes.io/docs/tasks/debug/debug-application/debug-running-pod/) to troubleshoot the issue with this Fuse Pod.
