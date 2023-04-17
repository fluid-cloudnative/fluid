# Fluid 版本升级

## 使用Helm将Fluid更新到最新版本(v0.7)

> 建议您从v0.6升级到最新版本v0.7。如果您安装的是更旧版本的Fluid，建议重新进行安装。

如果您此前已经安装过旧版本的Fluid，可以使用Helm进行更新。
更新前，建议确保各Runtime资源对象中的各个组件已经顺利启动完成，也就是类似以下状态：

```shell
$ kubectl get pod
NAME                 READY   STATUS    RESTARTS   AGE
hbase-fuse-chscz     1/1     Running   0          9h
hbase-fuse-fmhr5     1/1     Running   0          9h
hbase-master-0       2/2     Running   0          9h
hbase-worker-bdbjg   2/2     Running   0          9h
hbase-worker-rznd5   2/2     Running   0          9h
```

由于helm upgrade不会更新CRD，需要先对其手动进行更新：

```shell
$ tar zxvf fluid-0.8.0.tgz ./
$ kubectl apply -f fluid/crds/.
```

更新：
```shell
$ helm upgrade fluid fluid/
Release "fluid" has been upgraded. Happy Helming!
NAME: fluid
LAST DEPLOYED: Fri Mar 12 09:22:32 2021
NAMESPACE: default
STATUS: deployed
REVISION: 2
TEST SUITE: None
```

> 对于Kubernetes v1.17及以下环境，请使用`helm upgrade --set runtime.criticalFusePod=false fluid fluid/`
