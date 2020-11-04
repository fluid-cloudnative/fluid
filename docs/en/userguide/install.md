# Deploy Fluid on Your Kubernetes Cluster

## Prerequisites

- Git
- Kubernetes cluster（version >= 1.14）, and support CSI
- kubectl（version >= 1.14）
- [Helm](https://helm.sh/)（version >= 3.0）

The following documents assume that you have installed all the above requirements.

For the installation and configuration of kubectl, please refer to [here](https://kubernetes.io/docs/tasks/tools/install-kubectl/).

For the installation and configuration of Helm 3, please refer to [here](https://v3.helm.sh/docs/intro/install/).

## How to Deploy

### Download Fluid Chart

You can download the latest Fluid installation package from [Fluid Releases](https://github.com/fluid-cloudnative/fluid/releases).

Untar the Fluid package you just downloaded:

```shell
$ tar -zxf fluid.tgz
```

### Install Fluid with Helm

Create namespace:

```shell
$ kubectl create ns fluid-system
```

Install Fluid with:

```shell
$ helm install fluid fluid
NAME: fluid
LAST DEPLOYED: Fri Jul 24 16:10:18 2020
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

> The general format of the `helm install` command is like: `helm install <RELEASE_NAME> <SOURCE>`. In the above command,  the first `fluid` means the release name, and the second  `fluid` specified the path to the helm chart, i.e. the directory just unpacked.


### Upgrade Fluid with Helm

Before updating, it is recommended to ensure that all components in the AlluxioRuntime resource object have been started completely, which is similar to the following state:

```shell
$ kubectl get pod
NAME                 READY   STATUS    RESTARTS   AGE
hbase-fuse-chscz     1/1     Running   0          9h
hbase-fuse-fmhr5     1/1     Running   0          9h
hbase-master-0       2/2     Running   0          9h
hbase-worker-bdbjg   2/2     Running   0          9h
hbase-worker-rznd5   2/2     Running   0          9h
```
If you have installed an older version of Fluid before, you can use Helm to update it.

```shell
$ helm upgrade fluid fluid-0.4.0.tgz
Release "fluid" has been upgraded. Happy Helming!
NAME: fluid
LAST DEPLOYED: Wed Nov  4 09:19:58 2020
NAMESPACE: default
STATUS: deployed
REVISION: 2
TEST SUITE: None
```
> We have only tried to update from v0.3 to v0.4. If you upgrade directly to v0.4 from an older version, unknown types of errors may occur.

### Check Status of Component

**Check CRD used by Fluid:**

```shell
$ kubectl get crd | grep data.fluid.io
alluxiodataloads.data.fluid.io          2020-07-24T06:54:50Z
alluxioruntimes.data.fluid.io           2020-07-24T06:54:50Z
datasets.data.fluid.io                  2020-07-24T06:54:50Z
```

**Check the status of pods:**

```shell
$ kubectl get pod -n fluid-system
NAME                                  READY   STATUS    RESTARTS   AGE
controller-manager-7f99c884dd-894g9   1/1     Running   0          5m28s
csi-nodeplugin-fluid-dm9b8            2/2     Running   0          5m28s
csi-nodeplugin-fluid-hwtvh            2/2     Running   0          5m28s
```

If the Pod status is as shown above, then Fluid is installed on your Kubernetes cluster successfully!

### Check version info of Component

When csi-nodeplugin, alluxioruntime-controller and dataset-controller start，they will print their own version info into logs.  
If you installed with the charts provided by us，their version info will be fully consistent.  
If you installed manually, their version info may be not consistent. You can view in turn:
```bash
$ kubectl logs csi-nodeplugin-fluid-tc7fx -c plugins  -n fluid-system | head -n 9 | tail -n 6
$ kubectl logs alluxioruntime-controller-5dfb5c7966-mkgzb -n fluid-system | head -n 6
$ kubectl logs dataset-controller-7c4bc68b96-26mcb  -n fluid-system | head -n 6
```
The printed logs are in the following format:
```bash
2020/10/27 10:16:02 BuildDate: 2020-10-26_14:04:22
2020/10/27 10:16:02 GitCommit: f2c3a3fa1335cb0384e565f17a4f3284a6507cef
2020/10/27 10:16:02 GitTreeState: dirty
2020/10/27 10:16:02 GoVersion: go1.14.2
2020/10/27 10:16:02 Compiler: gc
2020/10/27 10:16:02 Platform: linux/amd64
```
If the logs printed by Pod have been cleaned up, you can run the following command to view the version:
```bash
$ kubectl exec csi-nodeplugin-fluid-tc7fx -c plugins  fluid-csi version -n fluid-system
$ kubectl exec alluxioruntime-controller-5dfb5c7966-mkgzb alluxioruntime-controller version -n fluid-system
$ kubectl exec dataset-controller-7c4bc68b96-26mcb dataset-controller version -n  fluid-system 
```

### Fluid use cases
For more use cases about Fluid, please refer to our demos:
- [Speed Up Accessing Remote Files](../samples/accelerate_data_accessing.md)
- [Cache Co-locality for Workload Scheduling](../samples/data_co_locality.md)
- [Accelerate Machine Learning Training with Fluid](../samples/machinelearning.md)

### Uninstall Fluid

```shell
$ helm delete fluid
$ kubectl delete -f fluid/crds
$ kubectl delete ns fluid-system
```

> The `fluid` in command `helm delete` means the <RELEASE_NAME> during installation.


### Advanced Configuration

In some cloud vendors, the default mount root directory `/alluxio-mnt` is not writable, so you have to modify the directory location

```
helm install fluid --set runtime.mountRoot=/var/lib/docker/alluxio-mnt fluid
```