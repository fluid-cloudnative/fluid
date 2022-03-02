# Deploy Fluid on Your Kubernetes Cluster

## Prerequisites

- Git
- Kubernetes cluster（version >= 1.16）, and support CSI
- kubectl（version >= 1.16）
- [Helm](https://helm.sh/)（version >= 3.5）

The following documents assume that you have installed all the above requirements.

For the installation and configuration of kubectl, please refer to [here](https://kubernetes.io/docs/tasks/tools/install-kubectl/).

For the installation and configuration of Helm 3, please refer to [here](https://v3.helm.sh/docs/intro/install/).

## How to Deploy

### Download Fluid Chart

You can download the latest Fluid installation package from [Fluid Releases](https://github.com/fluid-cloudnative/fluid/releases).

### Install Fluid with Helm

Create namespace:

```shell
$ kubectl create ns fluid-system
```

Install Fluid with:

```shell
$ helm install fluid fluid.tgz
NAME: fluid
LAST DEPLOYED: Fri Jul 24 16:10:18 2020
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

> For Kubernetes version lower than v1.17(included), please use `helm install --set runtime.criticalFusePod=false fluid fluid.tgz`

> The general format of the `helm install` command is like: `helm install <RELEASE_NAME> <SOURCE>`. In the above command,  the first `fluid` means the release name, and the second  `fluid` specified the path to the helm chart, i.e. the directory just unpacked.


### Upgrade Fluid to the latest version(v0.7) with Helm

If you have installed an older version of Fluid before, you can use Helm to upgrade it.
Before upgrading, it is recommended to ensure that all components in the AlluxioRuntime resource object have been started completely, which is similar to the following state:

```shell
$ kubectl get pod
NAME                 READY   STATUS    RESTARTS   AGE
hbase-fuse-chscz     1/1     Running   0          9h
hbase-fuse-fmhr5     1/1     Running   0          9h
hbase-master-0       2/2     Running   0          9h
hbase-worker-bdbjg   2/2     Running   0          9h
hbase-worker-rznd5   2/2     Running   0          9h
```

The command "helm upgrade" will not upgrade CRDs，we need to upgrade them manually：

```shell
$ tar zxvf fluid-0.7.0.tgz ./
$ kubectl apply -f fluid/crds/.
```

upgrade fluid：
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

> For Kubernetes version lower than v1.17(included), please use `helm install --set runtime.criticalFusePod=false fluid fluid.tgz`

> We recommend you to update Fluid v0.7 from v0.6. If you have an older version, our suggestion is to reinstall it to ensure everything works fine.

### Check Status of Component

**Check CRD used by Fluid:**

```shell
$ kubectl get crd | grep data.fluid.io
alluxioruntimes.data.fluid.io                          2022-02-28T08:14:45Z
databackups.data.fluid.io                              2022-02-28T08:14:45Z
dataloads.data.fluid.io                                2022-02-28T08:14:45Z
datasets.data.fluid.io                                 2022-02-28T08:14:45Z
goosefsruntimes.data.fluid.io                          2022-02-28T08:14:45Z
jindoruntimes.data.fluid.io                            2022-02-28T08:14:45Z
```

**Check the status of pods:**

```shell
$ kubectl get pod -n fluid-system
NAME                                  READY   STATUS    RESTARTS   AGE
alluxioruntime-controller-66bf8cbdf4-k6cxt   1/1     Running   0          6m50s
csi-nodeplugin-fluid-pq2zd                   2/2     Running   0          4m30s
csi-nodeplugin-fluid-rkv7h                   2/2     Running   0          6m41s
dataset-controller-558c5c7785-mtgfh          1/1     Running   0          6m50s
fluid-webhook-7b6cbf558-lw6lq                1/1     Running   0          6m50s
```

If the Pod status is as shown above, then Fluid is installed on your Kubernetes cluster successfully!

### Check version info of Component

When csi-nodeplugin, alluxioruntime-controller and dataset-controller start，they will print their own version info into logs.  
If you installed with the charts provided by us，their version info will be fully consistent.  
If you installed manually, their version info may be not consistent. You can check it with the following command:

```bash
$ kubectl exec csi-nodeplugin-fluid-pq2zd -n fluid-system -c plugins fluid-csi version
$ kubectl exec alluxioruntime-controller-66bf8cbdf4-k6cxt -n fluid-system -- alluxioruntime-controller version
$ kubectl exec dataset-controller-558c5c7785-mtgfh -n fluid-system -- dataset-controller version
```

The output should be like:
```
BuildDate: 2022-02-20_09:43:43
GitCommit: 808c72e3c5136152690599d187a76849d03ea448
GitTreeState: dirty
GoVersion: go1.16.8
Compiler: gc
Platform: linux/amd64
```

### Fluid use cases
For more use cases about Fluid, please refer to our demos:
- [Speed Up Accessing Remote Files](../samples/accelerate_data_accessing.md)
- [Cache Co-locality for Workload Scheduling](../samples/data_co_locality.md)
- [Accelerate Machine Learning Training with Fluid](../samples/machinelearning.md)

### Uninstall Fluid

To uninstall fluid safely, we should check weather Custom Resource Objects about fluid have been deleted completely first:
```shell
kubectl get crds -o custom-columns=NAME:.metadata.name | grep data.fluid.io  | sed ':t;N;s/\n/,/;b t' | xargs kubectl get --all-namespaces
```
If you confirm that all Custom resource objects about fluid have been deleted, you can safely uninstall fluid:

```shell
$ helm delete fluid
$ kubectl delete -f fluid/crds
$ kubectl delete ns fluid-system
```

> The `fluid` in command `helm delete` means the <RELEASE_NAME> during installation.


### Advanced Configuration

In some cloud vendors, the default mount root directory `/runtime-mnt` is not writable, so you have to modify the directory location

```
helm install fluid --set runtime.mountRoot=/var/lib/docker/runtime-mnt fluid
```