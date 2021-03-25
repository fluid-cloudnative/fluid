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


### Upgrade Fluid to the latest version with Helm

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
$ tar zxvf fluid-0.5.0.tgz ./
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

The old version of controller will not terminate automatically, the new version will stay in the Pending status：
```shell
$ kubectl -n fluid-system get pods
NAME                                         READY   STATUS    RESTARTS   AGE
alluxioruntime-controller-56687869f6-g9l9n   0/1     Pending   0          96s
alluxioruntime-controller-5b64fdbbb-j9h6r    1/1     Running   0          3m55s
csi-nodeplugin-fluid-r6crn                   2/2     Running   0          94s
csi-nodeplugin-fluid-wvhdn                   2/2     Running   0          87s
dataset-controller-5b7848dbbb-rjkl9          1/1     Running   0          3m55s
dataset-controller-64bf45c497-w8ncb          0/1     Pending   0          96s
```
delete them manually：
```shell
$ kubectl -n fluid-system delete pod alluxioruntime-controller-5b64fdbbb-j9h6r 
$ kubectl -n fluid-system delete pod dataset-controller-5b7848dbbb-rjkl9
```

> We recommend you to update from v0.3 and v0.4. If you have an older version, you'd better to reinstall it.

### Check Status of Component

**Check CRD used by Fluid:**

```shell
$ kubectl get crd | grep data.fluid.io
alluxiodataloads.data.fluid.io          2020-07-24T06:54:50Z
alluxioruntimes.data.fluid.io           2020-07-24T06:54:50Z
datasets.data.fluid.io                  2020-07-24T06:54:50Z
dataloads.data.fluid.io                 2021-03-12T00:00:47Z
datasets.data.fluid.io                  2021-03-12T00:00:47Z
jindoruntimes.data.fluid.io             2021-03-12T00:03:45Z
```

**Check the status of pods:**

```shell
$ kubectl get pod -n fluid-system
NAME                                  READY   STATUS    RESTARTS   AGE
alluxioruntime-controller-5dfb5c7966-mkgzb   1/1     Running   0          2d1h
csi-nodeplugin-fluid-64h69                   2/2     Running   0          2d1h
csi-nodeplugin-fluid-tc7fx                   2/2     Running   0          2d1h
dataset-controller-7c4bc68b96-26mcb          1/1     Running   0          2d1h
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

For uninstalling fluid safely, we should check weather Custom Resource Objects about fluid have been deleted completely first:
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