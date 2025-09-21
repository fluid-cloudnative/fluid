# Fluid Release Notes

## v1.0.7
### Highlights
*   **JindoFS 6.9.1** by default – higher read-bandwidth utilisation; file-prefetcher tag now auto-matches JindoFS image.  
*   **Sidecar coexistence** – multiple sidecar Pods can share one node without mount conflict; webhook allocates unique hostPath automatically.  
*   **JuiceFSRuntime** can disable metadata-sync & dataset-size calculation to cut control-plane load.  
*   **ThinRuntimeProfile** supports custom Pod lifecycle hooks.  

### Bug Fixes
*   Fix JuiceFSRuntime fields unexpectedly overwritten after update.  
*   Fix missing tolerations when JindoRuntime sets multiple ones.  
*   Fix Fuse-Pod creation failure on very long Dataset names.  
*   Suppress excessive Runtime status updates.  
*   Correct volume generation when TieredStore contains no data.

---

## v1.0.6
### Highlights
*   **Prefetch Sidecar** – accelerates model-serving cold start and improves bandwidth usage.  
*   **JuiceFSRuntime Fuse seamless upgrade** – idle Pods auto-cleaned via `cleanPolicy: onFuseChange`.  
*   **JuiceFSRuntime preheat** – exposes advanced warming parameters in DataLoad.  
*   **Control-plane tunables** – runtimeWorkers, kubeClient QPS/Burst, work-queue, env all user-configurable.  

### Enhancements
*   Security: upgrade Helm → 3.17.3, Alpine → 3.19.1.  
*   ThinRuntime controller optimised to 35 instances/s.  
*   Fuse mount-check ConfigMap shared at namespace level to cut webhook CPU.  
*   Batch API detection added to lower DatasetController APIServer pressure.

### Bug Fixes
*   ThinRuntime `containerNetwork` setting now takes effect.  
*   JuiceFSRuntime no longer triggers spurious updates during sync.

---

## v1.0.5
### Highlights
*  **Enhanced Binding Efficiency in Kubernetes Clusters**  - Optimized ThinRuntimeController’s binding performance for large-scale Dataset deployments, reducing resource scheduling latency in Kubernetes clusters.  
* **Integration with Third-Party Storage via HostPath**  - Enabled Dataset to mount third-party storage systems (e.g., NFS, Ceph) through node-local HostPath volumes, extending compatibility with heterogeneous data sources.  
* **Distributed Fault Tolerance for High Availability**  - Isolated single-node FUSE failures from impacting Dataset-wide availability, ensuring cluster-level stability in distributed environments.  
* **Extended Support for Long Naming Conventions**  - Permitted creation of Datasets with lengthy names and namespaces, avoiding Kubernetes label length constraints (63-character limit).   

### Bug Fixes
*   Webhook development-mode log panic removed.

---

## v1.0.4
### Highlights
* Added support to configure the `maxConcurrentReconciles` parameter for the `ThinRuntimeController`, allowing for improved concurrency management.
* Introduced support for resource configuration within the Fluid webhook, enabling more precise control over resource allocation.
* Added support for configuring `Burst` (default 30) and `Qps` (default 20) settings for client interactions with the APIServer through the Fluid Webhook, enhancing request handling capabilities.
* Implemented `mountMode` configuration to allow skipping the check for mount readiness, streamlining deployment processes.
 

### Bug Fixes
*   JindoRuntime master StatefulSet anti-affinity added for HA.

---

## v1.0.3
### Highlights
*   **Dynamic mount status** – exposes real-time mount progress to users.  
*   **Dataflow affinity** – new operations can inherit node affinity from previous ones.  
*   **JindoCache Prometheus** – native data-plane metrics scraping.  
*   **imagePullSecret** – supported at Runtime level (Thin, Jindo, JuiceFS).

### Bug Fixes
*   CronDataLoad no longer fails when affinity / nodeSelector / schedulerName / imagePullSecrets present.  
*   Upgrade-safe – DataOps with same Dataset name keep working after Fluid version bump.

---

## v1.0.2
### Highlights
*   **DataProcess resources** – requests/limits now configurable.  
*   **Dataset-in-use info** – injected into mutated app Pods for observability.  
*   **Hierarchical imagePullSecrets** – AlluxioRuntime supports repo secrets at multiple scopes.  
*   **Sub-path dynamic mount** – JuiceFS and others can mount `subPath` on the fly.

### Bug Fixes
*   JindoRuntime syncs AK/SK secrets correctly.  
*   JuiceFS worker port conflict resolved.  
*   DataLoad cronJob template fixed for JindoCache.

---

## v1.0.1
### Highlights
*   **Data-flow affinity scheduling** – operations stick to cache nodes where previous jobs ran.  
*   **Dynamic mount-point updates** – Dataset mount list editable without restart.  
*   **Sidecar PVC acceleration** – Fluid PVC can be accelerated in sidecar mode.  

### Runtime Versions
*   JuiceFS community → v1.1.0, enterprise → v4.9.16  
*   JindoRuntime → v6.2.0  

### Bug Fixes
*   JuiceFS attr-cache / entry-cache options can be overridden.  
*   JindoRuntime name no longer restricted by “jindofs” substring.

---

## v1.0.0

### Highlights

*   Configurable tiered data locality scheduling capability to optimize affinity between application pod and the data cache.

*   Support three kinds of data operations with different modes: once, onEvent, Cron, including: DataLoad, DataMigrate and DataProcess.

*   Support defining dataflow for data operations.

*   Support Python SDK for data scientists and operators to interact with Fluid control plane.

*   A new runtime for sharing in-memory immutable data: VineyardRuntime is supported in Fluid.

*   Security Hardening: Define a more restricted minimum necessary cluster role permissions for Fluid components, including eliminating  all the secret-related  and some create/update/delete privileges.

*   Enhance CSI Plugin and the FUSE Recovery feature for production usage.

*   Fluid drop all the Secret-related privileges to enhance its security.


### New Features

#### Data Operation Features

*   Support cron for DataMigrate（#3224）

*   Support cron for DataLoad (#3309)

*   Make toleration configurable in helm charts (#3354)

*   Support Data Process CRD (#3345) (#3360) (#3367)

*   Support dataflow (#3384)

*   Data operations support resources (#3571)

*   Support ssh password free pod for parallel data migrate (#3645)

#### Vineyard Runtime

*   Add the vineyard runtime CRD definitions. (#3555)

*   Add the helm chart for vineyard runtime. (#3624)

*   Add the controller and RBAC yaml for vineyard runtime. (#3659)

*   Implement vineyard runtime engine and controller (#3649)

*   Add a replicas field to the vineyard Runtime and delete the svc suffix in the vineyard helm chart. (#3700)

#### Others

*   Allow users to skip inject post start hook when using Fuse Sidecar (#3423)

*   Support worker of juicefs can be update by runtime (#3422)

*   Make juicefsruntime.spec.fuse configurale (#3539)

*   Feature: support symlink for NodePublish (#3440)

*   Support tiered locality scheduling for app pod (#3461) (#3489)

*   Allow setting kube client QPS and Burst (#3736)

*   Support configure rate limiter for controllers (#3758)

### Enhancements & Optimizations:

#### CSI Plugin & FUSE Recovery

*   Check path existence and lock in NodeUnpublishVolume (#3284)

*   UmountDuplicate larger than the threshold (#3429)

*   Fix duplicate mount issue, To #48332603 (#3433)

*   Bugfix: remove umountDuplicate and add warning event (#3403)

*   Bugfix: fix csi plugin concurrency issue on FuseRecovery and NodeUnpublishVolume (#3448)

*   Enhancement: recover fuse according to multiple peer group options (#3491)

*   Support symlink for NodePublish (#3440)

#### Security Enhancement

*   Upgrade Docker base image from Alpine 3.17 to 3.18（#3254）

*   Use configmap helm driver instead of secrets (#3272)

*   Del secret get of juicefs runtime clusterrole (#3305)

*   Enhance: optimize secret-related code logic in ThinRuntime controller (#3323)

*   Support alluxioruntime with no secret and generate idempotent mount scripts (#3378)

*   Refactor: safe exec commands (#3699)

*   Support pipe command (#3692)

*   Use utils.command  to replace exec.command (#3686)

*   Fix JuicefsRuntime: escape customized string before constructing commands (#3761)

*   Enhance: remove jindoruntime's fsGroup (#3632) (#3634) (#3635)

*   Support configure hostpid for runtime fuse. (#3755)

#### Others

*   Update runtime info in dataset's status during binding (#3270)

*   Print error message when failing to helm install (#3271)

*   Fail fast with wrong kubelet rootdir (#3331)

*   Auto clean up crd-upgrade pod  (#3500)

*   Change some fields from optional to required in CRD YAML (#3684)

*   Change enableServiceLinks from true to false (#3701)

*   Enhancement: auto discover runtime crds and dynamically enable/disable reconcilers (#3708)

*   Enhancement: add common labels to resources that managed by fluid (#3720)

*   Enhancement: Avoid heavy pod listing when worker statefulset has no replicas (#3771)

### Minor

#### Runtime Upgrade

*   JuiceFSRuntime's default version upgrades to v1.1.0(community) and v4.9.16(enterprise)

*   JindoRuntime's default version upgrades to v6.2.0

#### Bug Fix

*   Bugfix: Support batch/v1beta1 cronjobs for compatibility before Kubernetes v1.21 (#3280)

*   Bugfix: fix csi plugin loop mount bug (#3287)

*   Bugfix: reconcile thinruntime failed when dataset is deleted (#3300)

*   Bugfix: fix thinruntime stuck bug when deleting it (#3335)

*   Fix worker\&fuse options & worker tiredstore (#3344)

*   Del metadata sync && del duplicate metrics (#3380)

*   Bugfix: clean up orphaned thinruntime resources (#3393)

*   Fix Alluxio master in HA mode start error. (#3658)

*   Pass AccessModes to thinruntime fuse container (#3696)

*   Fix fatal error: concurrent map writes for runtime controllers (#3757)

*   Fix incorrect conversion between integer types (#3688)

#### Refactoring

*   Refactor fuse sidecar mutation with mutator (#3477) (#3487) (#3492)

*   Auto infer engine implementation to support multi-engine Runtimes (#3672)

#### Integration

*   Integrate Fluid with Kubeflow Pipline (#3694)



## v0.9.0

### Breaking Changes

*   Change matching pod requests mode of webhook from namespaceSelector to objectSelector

### Features

*   Add thinRuntime to simplify integration with third-party storage systems

*   Addon component for Fluid's open source CubeFS, NFS

*   Support for accessing data across namespaces

*   Support for subDataset

*   Native acceleration system EFCRuntime for distributed file systems NFS, GPFS

*   Support dataMigrate for data migration operations (currently only supported by JuiceFSRuntime)

*   Add customizable configuration for cache cleanup timeout and maximum retry times, Webhook timeout limits

*   Add Dataload configuration for ImagePullSecrets, node affinity

*   RBAC permission reduction

*   Upgrade to golang 1.18

*   Support for installing Fluid via Helm Repo

### Refactoring

*   Use data operation framework to construct data migrate, load, backup behaviors

### Bug Fix

*   [JindoRuntime should support configurable env variables](https://github.com/fluid-cloudnative/fluid/issues/3154)

*   [Could not set nodeAffinity in dataset](https://github.com/fluid-cloudnative/fluid/issues/2772)

*   [Runtime helm release stuck in "pending-install" status](https://github.com/fluid-cloudnative/fluid/issues/2764)

*   [CSI failed to recover FUSE mount point for AlluxioRuntime ](https://github.com/fluid-cloudnative/fluid/issues/2719)

*   [\[JuiceFS\] FUSE pod scheduled failed because of conflict port](https://github.com/fluid-cloudnative/fluid/issues/2668)

*   [Fluid csi on rke2 k8s(1.22) use mount output empty,it caused app pod not work](https://github.com/fluid-cloudnative/fluid/issues/2613)

### Runtime Upgrade

*   AlluxioRuntime is upgrade from v2.8.2 to v2.9.1

*   JindoRuntime is upgraded from from  4.5.1 to 4.6.7

*   JuicefsRuntime is upgraded from v1.0.0 to v1.0.4

## v0.8.0

### Features

*   Lifecycle management of Serverless Job with fluid sidecar support

*   Enabling Runtime Controller on demand

*   Arm64 support with JuicefsRuntime

*   Container Network with short-circuit read support

*   Leader election support for Controllers and Webhook

*   Automatic CRD upgrader

*   Restrict Pod scheduling to dataset cache nodes

*   Tens of thousands of nodes support

*   Image pull secrets support

*   GCS support for Alluxio Runtime

### Refactorings

*   Port Allocation  with different strategies: bitmap and random

### Bug Fix

*   [Runtime cannot complete deletion when restarting controller](https://github.com/fluid-cloudnative/fluid/issues/1970)

*   [Pod update failed with fluid webhook injection enabled](https://github.com/fluid-cloudnative/fluid/issues/2053)

*   [Unhandled exception in gopkg.in/yaml.v3](https://github.com/fluid-cloudnative/fluid/issues/1869)

*   [Webhook failed to load root certificates: unable to parse bytes as PEM block](https://github.com/fluid-cloudnative/fluid/issues/1399)

*   [Plugin delete the csi socket when restarting unexpectly](https://github.com/fluid-cloudnative/fluid/issues/2088)

### Runtime Upgrade

*   AlluxioRuntime is upgrade from v2.7.2 to v2.8

*   JindoRuntime is upgraded from Jindo Engine to JindoFSX Engine, and the version is from 3.8 to  4.5.1

*   JuiceRuntime is upgraded from v0.11.0 to v1.0.0

## v0.7.0

### Breaking Changes

*   Update Kubernetes v1.20.12 dependencies and use Go 1.16

*   Update CustomResourceDefinition to apiextensions.k8s.io/v1

*   Update MutatingWebhookConfiguration to admissionregistration.k8s.io/v1

*   Update CSIDriver to storage.k8s.io/v1

### Features

*   Support fuse sidecar auto injection for all the runtimes，it’s helpful for no CSI environment

*   Support fuse auto recovery and upgrade

*   Support lazy fuse mount mode

*   Support New cache runtime: JuiceFS

### Refactorings

*   Change cache worker deployment mode from DaemonSet to StatefulSet to use K8s Native schedule mechanism

### Bug Fix

*   Fix “[Failed to update status of dataload](https://github.com/fluid-cloudnative/fluid/issues/1436)”

*   Fix “[Failed to delete dataload when target dataset is removed](https://github.com/fluid-cloudnative/fluid/issues/1419)”

*   Fix “[node-driver-registrar will not receive any volume umount in subdirectories of kubelet-dir](https://github.com/fluid-cloudnative/fluid/issues/1048)”

## v0.6.0

### Features

*   Support dataset cache autoscaling and cronscaling

*   Add dataset mount point dynamically update feature

*   Enhance dataset cache aware Pod scheduling

*   Enhance HA support for cache Runtime

*   Support new cache Runtime：GooseFS

### Bugs

*   Fix [if alluxioruntime is nonroot, databackup will fail](https://github.com/fluid-cloudnative/fluid/issues/745)

*   Fix [Node labels exceeds maximum length limit for long namespace and name](https://github.com/fluid-cloudnative/fluid/issues/704)

## v0.5.0

### Features

*   Support on-the-fly dataset cache scale out/in

*   Add Metadata backup and restore operation

*   Support Fuse global mode and toleration

*   Enhance Prometheus monitoring support for AlluxioRuntime

*   Support new Runtime：JindoFS

*   Support HDFS configuration

### Bugs

*   [Fix compatibality issue of K8s 1.19+](https://github.com/fluid-cloudnative/fluid/issues/603)

## v0.4.0

### Features

*   Warm up Dataset automatically before using it

*   Support managing 4,000,000 files

*   Support deploying multiple dataset in the same node

*   Support showing the HCFS Access Endpoint in Dataset status

## v0.3.0

### Features

*   Accelerate data access in Host-path mode in K8s

*   Accelerate data access in Persistent Volume mode in K8s

*   Support the underlay storage(NFS, Lustre) which is configured only with non-root

*   Make the Alluxio Runtime’s settings optimized by default

### Bugs

*   [MountVolume.SetUp failed in GKE](https://github.com/fluid-cloudnative/fluid/issues/222)

*   [fuse.csi.fluid.io not found in the list registered CSI drivers when node restart in k8s 1.15.1](https://github.com/fluid-cloudnative/fluid/issues/220)

## v0.2.0

### Features

*   Add code coverage badge

*   Update chart v0.2.0

*   Update docs

*   Refactor the package name

## v0.1.0

### Features

*   Add 'Dataset', 'AlluxioRuntime' CRD for managing the data

*   Add PV and PVC for Posix interface

*   Documentation

*   Helm Chart

