# Deploy Fluid on Your Kubernetes Cluster

## Prerequisites

- Git
- Kubernetes cluster（version >= 1.18）, and support CSI
- kubectl（version >= 1.18）
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

Add Fluid repository to Helm repos and keep it up-to-date

```shell
$ helm repo add fluid https://fluid-cloudnative.github.io/charts
$ helm repo update
```


Install Fluid with:

```shell
$ helm install fluid fluid/fluid
NAME: fluid
LAST DEPLOYED: Wed May 24 18:17:16 2023
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

> For Kubernetes version lower than v1.17(included), please use `helm install --set runtime.criticalFusePod=false fluid fluid.tgz`

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

upgrade fluid：
```shell
$ helm upgrade fluid fluid/fluid
Release "fluid" has been upgraded. Happy Helming!
NAME: fluid
LAST DEPLOYED: Fri Sep  2 18:54:18 2022
NAMESPACE: default
STATUS: deployed
REVISION: 2
TEST SUITE: None
```

> For Kubernetes version lower than v1.17(included), please use `helm install --set runtime.criticalFusePod=false fluid fluid.tgz`

> We recommend you to update Fluid latest version from v0.7. If you have an older version, our suggestion is to reinstall it to ensure everything works fine.

### Check Status of Component

**Check CRD used by Fluid:**

```shell
$ kubectl get crd | grep data.fluid.io
alluxioruntimes.data.fluid.io                          2023-05-24T10:14:47Z
databackups.data.fluid.io                              2023-05-24T10:14:47Z
dataloads.data.fluid.io                                2023-05-24T10:14:47Z
datamigrates.data.fluid.io                             2023-05-24T10:28:11Z
datasets.data.fluid.io                                 2023-05-24T10:14:47Z
efcruntimes.data.fluid.io                              2023-05-24T10:28:12Z
goosefsruntimes.data.fluid.io                          2023-05-24T10:14:47Z
jindoruntimes.data.fluid.io                            2023-05-24T10:14:48Z
juicefsruntimes.data.fluid.io                          2023-05-24T10:14:48Z
thinruntimeprofiles.data.fluid.io                      2023-05-24T10:28:16Z
thinruntimes.data.fluid.io                             2023-05-24T10:28:16Z
```

**Check the status of pods:**

```shell
$ kubectl get pod -n fluid-system
NAME                                     READY   STATUS      RESTARTS   AGE
csi-nodeplugin-fluid-2scs9               2/2     Running     0          50s
csi-nodeplugin-fluid-7vflb               2/2     Running     0          20s
csi-nodeplugin-fluid-f9xfv               2/2     Running     0          33s
dataset-controller-686d9d9cd6-gk6m6      1/1     Running     0          50s
fluid-crds-upgrade-1.0.0-37e17c6-fp4mm   0/1     Completed   0          74s
fluid-webhook-5bc9dfb9d8-hdvhk           1/1     Running     0          50s
fluidapp-controller-6d4cbdcd88-z7l4c     1/1     Running     0          50s
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
  BuildDate: 2024-03-02_07:35:18
  GitCommit: 50ee8887239f07592ba74af3e14379efc1487c0c
  GitTreeState: clean
  GoVersion: go1.18.10
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

1. In some cloud vendors, the default mount root directory `/runtime-mnt` is not writable, so you have to modify the directory location

```
helm install fluid --set runtime.mountRoot=/var/lib/docker/runtime-mnt fluid
```

2. The feature `Fuse Recovery` is not enable by default, to enable this:

```
helm install fluid --set csi.featureGates='FuseRecovery=true' fluid
```

3. If your Kubernetes cluster has a custom configured kubelet root directory, please configure the KUBELET_ROOTDIR when installing Fluid with the following command: 
```shell
helm install --set csi.kubelet.rootDir=<kubelet-root-dir> \
  --set csi.kubelet.certDir=<kubelet-root-dir>/pki fluid fluid.tgz
```

> You can execute the following command on the Kubernetes node to view the --root-dir parameter configuration:
> ```
> ps -ef | grep $(which kubelet) | grep root-dir
> ```
> If the above command has no output, the kubelet root path is the default value (/var/lib/kubelet), which is the default value set by Fluid.

4. When you install a Kubernetes cluster using [Sealer](http://sealer.cool), it by default uses `apiserver.cluster.local` as the address of the API Server. At the same time, it writes this address to the `kubelet.conf` file and the corresponding IP address to the `hosts` file. This will cause the Fluid CSI Plugin fail to find the IP address of the API Server. You can set the Fluid CSI Plugin to use hostNetwork via the following command:
```shell
# install
helm install fluid --set csi.config.hostNetwork=true fluid/fluid
# upgrade
helm upgrade fluid --set csi.config.hostNetwork=true fluid/fluid
```

5. Under the default configuration, Fluid does not set `spec.resources` for the installed Controller Pods. If you are deploying Fluid in a production environment, it is recommended to configure `spec.resources` for each Controller Pod based on your . For how to configure for different Controller Pods, you can refer to the examples in the [`values.yaml`](https://github.com/fluid-cloudnative/fluid/blob/master/charts/fluid/fluid/values.yaml) file. For instance, if you'd like to modify the resource requests and limits for the Dataset Controller Pod, you should first define a custom values file (e.g., my-resource-file.yaml):

```yaml
dataset:
  resources:
    requests:
      cpu: 500m
      memory: 256Mi
    limits:
      cpu: 1000m
      memory: 512Mi
```

Override the default resources with the following command:

```shell
helm upgrade fluid --install -f my-resource-file.yaml fluid/fluid
```
