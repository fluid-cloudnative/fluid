## 1. 为什么我使用Helm安装fluid失败了？

推荐按照[Fluid安装文档](./install.md)依次确认Fluid组件是否正常运行。

Fluid安装文档是以`Helm 3`为例进行部署的。如果您使用`Helm 3`以下的版本部署Fluid，
并且遇到了`CRD没有正常启动`的情况，这可能是因为`Helm 3`及其以上版本会在`helm install`的时候自动安装CRD，
而低版本的Helm则不会。
参见[Helm官方文档说明](https://helm.sh/docs/chart_best_practices/custom_resource_definitions/)。

在这种情况下，您需要手动安装CRD：
```bash
$ kubectl create -f fluid/crds
```

## 2. 为什么我无法删除Runtime？

请检查相关Pod运行状态和Runtime的Events。

只要有任何活跃Pod还在使用Fluid创建的Volume，Fluid就不会完成删除操作。

如下的命令可以快速地找出这些活跃Pod，使用时把`<dataset_name>`和`<dataset_namespace>`换成自己的即可：
```bash
kubectl describe pvc <dataset_name> -n <dataset_namespace> | \
	awk '/^Mounted/ {flag=1}; /^Events/ {flag=0}; flag' | \
	awk 'NR==1 {print $3}; NR!=1 {print $1}' | \
	xargs -I {} kubectl get po {} | \
	grep -E "Running|Terminating|Pending" | \
	cut -d " " -f 1
```
