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

You can execute the following command in any folder to clone source code from [Fluid repository](https://github.com/fluid-cloudnative/fluid):

```shell
$ git clone https://github.com/fluid-cloudnative/fluid.git
```

[Helm Charts](https://github.com/fluid-cloudnative/fluid/tree/master/charts) used to deploy Fluid is included in source code.

### Install Fluid with Helm

Enter the cloned local repository:

```shell
$ cd fluid
```

Create namespace:

```shell
$ kubectl create ns fluid-system
```

Install Fluid with:

```shell
$ helm install fluid charts/fluid/fluid
NAME: fluid
LAST DEPLOYED: Fri Jul 24 16:10:18 2020
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

> The general format of the `helm install` command is like: `helm install <RELEASE_NAME> <SOURCE>`. In the above command,  `fluid` means the release name, and `charts/fluid/fluid` specify the path to the helm chart.

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

### Uninstall Fluid

```shell
$ helm delete fluid
$ kubectl delete -f charts/fluid/fluid/crds
```

> The `fluid` here means the <RELEASE_NAME> during installation.