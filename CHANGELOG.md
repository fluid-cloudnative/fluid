# Fluid Release Notes

## v1.0.0

## v0.9.0

### Breaking Changes

- Change matching pod requests mode of webhook from namespaceSelector to objectSelector

### Features

- Add thinRuntime to simplify integration with third-party storage systems
- Addon component for Fluid's open source CubeFS, NFS
- Support for accessing data across namespaces
- Support for subDataset
- Native acceleration system EFCRuntime for distributed file systems NFS, GPFS
- Support dataMigrate for data migration operations (currently only supported by JuiceFSRuntime)
- Add customizable configuration for cache cleanup timeout and maximum retry times, Webhook timeout limits
- Add Dataload configuration for ImagePullSecrets, node affinity
- RBAC permission reduction
- Upgrade to golang 1.18
- Support for installing Fluid via Helm Repo

### Refactoring

- Use data operation framework to construct data migrate, load, backup behaviors

### Bug Fix

- [JindoRuntime should support configurable env variables](https://github.com/fluid-cloudnative/fluid/issues/3154)
- [Could not set nodeAffinity in dataset](https://github.com/fluid-cloudnative/fluid/issues/2772)
- [Runtime helm release stuck in "pending-install" status](https://github.com/fluid-cloudnative/fluid/issues/2764)
- [CSI failed to recover FUSE mount point for AlluxioRuntime ](https://github.com/fluid-cloudnative/fluid/issues/2719)
- [[JuiceFS] FUSE pod scheduled failed because of conflict port](https://github.com/fluid-cloudnative/fluid/issues/2668)
- [Fluid csi on rke2 k8s(1.22) use mount output empty,it caused app pod not work](https://github.com/fluid-cloudnative/fluid/issues/2613)

### Runtime Upgrade

- AlluxioRuntime is upgrade from v2.8.2 to v2.9.1
- JindoRuntime is upgraded from from  4.5.1 to 4.6.7
- JuicefsRuntime is upgraded from v1.0.0 to v1.0.4


## v0.8.0

### Features

- Lifecycle management of Serverless Job with fluid sidecar support
- Enabling Runtime Controller on demand
- Arm64 support with JuicefsRuntime
- Container Network with short-circuit read support
- Leader election support for Controllers and Webhook
- Automatic CRD upgrader
- Restrict Pod scheduling to dataset cache nodes
- Tens of thousands of nodes support
- Image pull secrets support
- GCS support for Alluxio Runtime

### Refactorings

- Port Allocation  with different strategies: bitmap and random

### Bug Fix

- [Runtime cannot complete deletion when restarting controller](https://github.com/fluid-cloudnative/fluid/issues/1970) 
- [Pod update failed with fluid webhook injection enabled](https://github.com/fluid-cloudnative/fluid/issues/2053)
- [Unhandled exception in gopkg.in/yaml.v3](https://github.com/fluid-cloudnative/fluid/issues/1869)
- [Webhook failed to load root certificates: unable to parse bytes as PEM block](https://github.com/fluid-cloudnative/fluid/issues/1399)
- [Plugin delete the csi socket when restarting unexpectly](https://github.com/fluid-cloudnative/fluid/issues/2088)

### Runtime Upgrade 

- AlluxioRuntime is upgrade from v2.7.2 to v2.8
- JindoRuntime is upgraded from Jindo Engine to JindoFSX Engine, and the version is from 3.8 to  4.5.1
- JuiceRuntime is upgraded from v0.11.0 to v1.0.0


## v0.7.0

### Breaking Changes

- Update Kubernetes v1.20.12 dependencies and use Go 1.16
- Update CustomResourceDefinition to apiextensions.k8s.io/v1
- Update MutatingWebhookConfiguration to admissionregistration.k8s.io/v1
- Update CSIDriver to storage.k8s.io/v1

### Features

- Support fuse sidecar auto injection for all the runtimes，it’s helpful for no CSI environment
- Support fuse auto recovery and upgrade
- Support lazy fuse mount mode
- Support New cache runtime: JuiceFS

### Refactorings

- Change cache worker deployment mode from DaemonSet to StatefulSet to use K8s Native schedule mechanism

### Bug Fix
- Fix “[Failed to update status of dataload](https://github.com/fluid-cloudnative/fluid/issues/1436)”
- Fix “[Failed to delete dataload when target dataset is removed](https://github.com/fluid-cloudnative/fluid/issues/1419)”
- Fix “[node-driver-registrar will not receive any volume umount in subdirectories of kubelet-dir](https://github.com/fluid-cloudnative/fluid/issues/1048)”


## v0.6.0

### Features

- Support dataset cache autoscaling and cronscaling
- Add dataset mount point dynamically update feature
- Enhance dataset cache aware Pod scheduling 
- Enhance HA support for cache Runtime
- Support new cache Runtime：GooseFS

### Bugs

- Fix [if alluxioruntime is nonroot, databackup will fail](https://github.com/fluid-cloudnative/fluid/issues/745)
- Fix [Node labels exceeds maximum length limit for long namespace and name](https://github.com/fluid-cloudnative/fluid/issues/704)


## v0.5.0

### Features

- Support on-the-fly dataset cache scale out/in
- Add Metadata backup and restore operation
- Support Fuse global mode and toleration
- Enhance Prometheus monitoring support for AlluxioRuntime
- Support new Runtime：JindoFS
- Support HDFS configuration

### Bugs

- [Fix compatibality issue of K8s 1.19+](https://github.com/fluid-cloudnative/fluid/issues/603)

## v0.4.0

### Features

- Warm up Dataset automatically before using it
- Support managing 4,000,000 files
- Support deploying multiple dataset in the same node
- Support showing the HCFS Access Endpoint in Dataset status

## v0.3.0

### Features

- Accelerate data access in Host-path mode in K8s
- Accelerate data access in Persistent Volume mode in K8s
- Support the underlay storage(NFS, Lustre) which is configured only with non-root
- Make the Alluxio Runtime’s settings optimized by default


### Bugs

- [MountVolume.SetUp failed in GKE](https://github.com/fluid-cloudnative/fluid/issues/222)
- [fuse.csi.fluid.io not found in the list registered CSI drivers when node restart in k8s 1.15.1](https://github.com/fluid-cloudnative/fluid/issues/220)

## v0.2.0

### Features

- Add code coverage badge  
- Update chart v0.2.0  
- Update docs  
- Refactor the package name  

## v0.1.0

### Features

- Add 'Dataset', 'AlluxioRuntime' CRD for managing the data
- Add PV and PVC for Posix interface
- Documentation
- Helm Chart
