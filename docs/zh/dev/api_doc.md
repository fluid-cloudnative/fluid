# API Reference

## Packages
- [data.fluid.io/v1alpha1](#datafluidiov1alpha1)


## data.fluid.io/v1alpha1



Package v1alpha1 is the v1alpha1 version of the API.

Package v1alpha1 contains API Schema definitions for the data v1alpha1 API group

zz_unit_test_scheme.go with a "zz" prefix to ensure its init function is called after all the other init function in the pacakge

### Resource Types
- [AlluxioRuntime](#alluxioruntime)
- [CacheRuntime](#cacheruntime)
- [CacheRuntimeClass](#cacheruntimeclass)
- [DataBackup](#databackup)
- [DataLoad](#dataload)
- [DataMigrate](#datamigrate)
- [DataProcess](#dataprocess)
- [Dataset](#dataset)
- [EFCRuntime](#efcruntime)
- [GooseFSRuntime](#goosefsruntime)
- [JindoRuntime](#jindoruntime)
- [JuiceFSRuntime](#juicefsruntime)
- [ThinRuntime](#thinruntime)
- [ThinRuntimeProfile](#thinruntimeprofile)
- [VineyardRuntime](#vineyardruntime)



#### APIGatewayStatus



API Gateway



_Appears in:_
- [RuntimeStatus](#runtimestatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `endpoint` _string_ | Endpoint for accessing |  |  |


#### AffinityPolicy

_Underlying type:_ _string_

AffinityPolicy the strategy for the affinity between Data Operation Pods.



_Appears in:_
- [AffinityStrategy](#affinitystrategy)

| Field | Description |
| --- | --- |
| `` |  |
| `Require` |  |
| `Prefer` |  |


#### AffinityStrategy







_Appears in:_
- [OperationRef](#operationref)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `dependOn` _[ObjectRef](#objectref)_ | Specifies the dependent preceding operation in a workflow. If not set, use the operation referred to by RunAfter. |  | Optional: \{\} <br /> |
| `policy` _[AffinityPolicy](#affinitypolicy)_ | Policy one of: "", "Require", "Prefer" |  | Optional: \{\} <br /> |
| `prefers` _[Prefer](#prefer) array_ |  |  |  |
| `requires` _[Require](#require) array_ |  |  |  |


#### AlluxioCompTemplateSpec



AlluxioCompTemplateSpec is a description of the Alluxio commponents



_Appears in:_
- [AlluxioRuntimeSpec](#alluxioruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `replicas` _integer_ | Replicas is the desired number of replicas of the given template.<br />If unspecified, defaults to 1.<br />replicas is the min replicas of dataset in the cluster |  | Minimum: 1 <br />Optional: \{\} <br /> |
| `jvmOptions` _string array_ | Options for JVM |  |  |
| `properties` _object (keys:string, values:string)_ | Configurable properties for the Alluxio component. <br><br />Refer to <a href="https://docs.alluxio.io/os/user/stable/en/reference/Properties-List.html">Alluxio Configuration Properties</a> for more info |  | Optional: \{\} <br /> |
| `ports` _object (keys:string, values:integer)_ | Ports used by Alluxio(e.g. rpc: 19998 for master) |  | Optional: \{\} <br /> |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#resourcerequirements-v1-core)_ | Resources that will be requested by the Alluxio component. <br><br /><br><br />Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources<br />already allocated to the pod. |  | Optional: \{\} <br /> |
| `env` _object (keys:string, values:string)_ | Environment variables that will be used by Alluxio component. <br> |  |  |
| `enabled` _boolean_ | Enabled or Disabled for the components. For now, only  API Gateway is enabled or disabled. |  | Optional: \{\} <br /> |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector which must be true for the master to fit on a node |  | Optional: \{\} <br /> |
| `networkMode` _[NetworkMode](#networkmode)_ | Whether to use hostnetwork or not |  | Enum: [HostNetwork  ContainerNetwork] <br />Optional: \{\} <br /> |
| `volumeMounts` _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#volumemount-v1-core) array_ | VolumeMounts specifies the volumes listed in ".spec.volumes" to mount into the alluxio runtime component's filesystem. |  | Optional: \{\} <br /> |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata defines labels and annotations that will be propagated to Alluxio's pods |  | Optional: \{\} <br /> |
| `imagePullSecrets` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#localobjectreference-v1-core) array_ | ImagePullSecrets that will be used to pull images |  |  |


#### AlluxioFuseSpec



AlluxioFuseSpec is a description of the Alluxio Fuse



_Appears in:_
- [AlluxioRuntimeSpec](#alluxioruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `image` _string_ | Image for Alluxio Fuse(e.g. alluxio/alluxio-fuse) |  |  |
| `imageTag` _string_ | Image Tag for Alluxio Fuse(e.g. 2.3.0-SNAPSHOT) |  |  |
| `imagePullPolicy` _string_ | One of the three policies: `Always`, `IfNotPresent`, `Never` |  |  |
| `imagePullSecrets` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#localobjectreference-v1-core) array_ | ImagePullSecrets that will be used to pull images |  |  |
| `jvmOptions` _string array_ | Options for JVM |  |  |
| `properties` _object (keys:string, values:string)_ | Configurable properties for Alluxio System. <br><br />Refer to <a href="https://docs.alluxio.io/os/user/stable/en/reference/Properties-List.html">Alluxio Configuration Properties</a> for more info |  |  |
| `env` _object (keys:string, values:string)_ | Environment variables that will be used by Alluxio Fuse |  |  |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#resourcerequirements-v1-core)_ | Resources that will be requested by Alluxio Fuse. <br><br /><br><br />Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources<br />already allocated to the pod. |  | Optional: \{\} <br /> |
| `args` _string array_ | Arguments that will be passed to Alluxio Fuse |  |  |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector which must be true for the fuse client to fit on a node,<br />this option only effect when global is enabled |  | Optional: \{\} <br /> |
| `cleanPolicy` _[FuseCleanPolicy](#fusecleanpolicy)_ | CleanPolicy decides when to clean Alluxio Fuse pods.<br />Currently Fluid supports two policies: OnDemand and OnRuntimeDeleted<br />OnDemand cleans fuse pod once the fuse pod on some node is not needed<br />OnRuntimeDeleted cleans fuse pod only when the cache runtime is deleted<br />Defaults to OnRuntimeDeleted |  | Optional: \{\} <br /> |
| `networkMode` _[NetworkMode](#networkmode)_ | Whether to use hostnetwork or not |  | Enum: [HostNetwork  ContainerNetwork] <br />Optional: \{\} <br /> |
| `volumeMounts` _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#volumemount-v1-core) array_ | VolumeMounts specifies the volumes listed in ".spec.volumes" to mount into the alluxio runtime component's filesystem. |  | Optional: \{\} <br /> |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata defines labels and annotations that will be propagated to Alluxio's fuse pods |  | Optional: \{\} <br /> |


#### AlluxioRuntime



AlluxioRuntime is the Schema for the alluxioruntimes API





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `data.fluid.io/v1alpha1` | | |
| `kind` _string_ | `AlluxioRuntime` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[AlluxioRuntimeSpec](#alluxioruntimespec)_ |  |  |  |




#### AlluxioRuntimeSpec



AlluxioRuntimeSpec defines the desired state of AlluxioRuntime



_Appears in:_
- [AlluxioRuntime](#alluxioruntime)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `alluxioVersion` _[VersionSpec](#versionspec)_ | The version information that instructs fluid to orchestrate a particular version of Alluxio. |  |  |
| `master` _[AlluxioCompTemplateSpec](#alluxiocomptemplatespec)_ | The component spec of Alluxio master |  |  |
| `jobMaster` _[AlluxioCompTemplateSpec](#alluxiocomptemplatespec)_ | The component spec of Alluxio job master |  |  |
| `worker` _[AlluxioCompTemplateSpec](#alluxiocomptemplatespec)_ | The component spec of Alluxio worker |  |  |
| `jobWorker` _[AlluxioCompTemplateSpec](#alluxiocomptemplatespec)_ | The component spec of Alluxio job Worker |  |  |
| `apiGateway` _[AlluxioCompTemplateSpec](#alluxiocomptemplatespec)_ | The component spec of Alluxio API Gateway |  |  |
| `initUsers` _[InitUsersSpec](#initusersspec)_ | The spec of init users |  |  |
| `fuse` _[AlluxioFuseSpec](#alluxiofusespec)_ | The component spec of Alluxio Fuse |  |  |
| `properties` _object (keys:string, values:string)_ | Configurable properties for Alluxio system. <br><br />Refer to <a href="https://docs.alluxio.io/os/user/stable/en/reference/Properties-List.html">Alluxio Configuration Properties</a> for more info |  |  |
| `jvmOptions` _string array_ | Options for JVM |  |  |
| `tieredstore` _[TieredStore](#tieredstore)_ | Tiered storage used by Alluxio |  |  |
| `data` _[Data](#data)_ | Management strategies for the dataset to which the runtime is bound |  |  |
| `replicas` _integer_ | The replicas of the worker, need to be specified |  |  |
| `runAs` _[User](#user)_ | Manage the user to run Alluxio Runtime |  |  |
| `disablePrometheus` _boolean_ | Disable monitoring for Alluxio Runtime<br />Prometheus is enabled by default |  | Optional: \{\} <br /> |
| `hadoopConfig` _string_ | Name of the configMap used to support HDFS configurations when using HDFS as Alluxio's UFS. The configMap<br />must be in the same namespace with the AlluxioRuntime. The configMap should contain user-specific HDFS conf files in it.<br />For now, only "hdfs-site.xml" and "core-site.xml" are supported. It must take the filename of the conf file as the key and content<br />of the file as the value. |  | Optional: \{\} <br /> |
| `volumes` _[Volume](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#volume-v1-core) array_ | Volumes is the list of Kubernetes volumes that can be mounted by the alluxio runtime components and/or fuses. |  | Optional: \{\} <br /> |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata defines labels and annotations that will be propagated to Alluxio's pods |  | Optional: \{\} <br /> |
| `management` _[RuntimeManagement](#runtimemanagement)_ | RuntimeManagement defines policies when managing the runtime |  | Optional: \{\} <br /> |
| `imagePullSecrets` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#localobjectreference-v1-core) array_ | ImagePullSecrets that will be used to pull images |  | Optional: \{\} <br /> |


#### CacheRuntime



CacheRuntime is the Schema for the CacheRuntimes API





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `data.fluid.io/v1alpha1` | | |
| `kind` _string_ | `CacheRuntime` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[CacheRuntimeSpec](#cacheruntimespec)_ |  |  |  |


#### CacheRuntimeClass



CacheRuntimeClass is the Schema for the cacheruntimeclasses API.
CacheRuntimeClass defines a class of cache runtime implementations with specific configurations.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `data.fluid.io/v1alpha1` | | |
| `kind` _string_ | `CacheRuntimeClass` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `fileSystemType` _string_ | FileSystemType is the file system type of the cache runtime (e.g., "alluxio", "juicefs") |  | Required: \{\} <br /> |
| `topology` _[RuntimeTopology](#runtimetopology)_ | Topology describes the topology of the CacheRuntime components (master, worker, client) |  | Optional: \{\} <br /> |
| `extraResources` _[RuntimeExtraResources](#runtimeextraresources)_ | ExtraResources specifies additional resources (e.g., ConfigMaps) used by the CacheRuntime components |  | Optional: \{\} <br /> |


#### CacheRuntimeClientSpec



CacheRuntimeClientSpec describes the desired state of CacheRuntime client component



_Appears in:_
- [CacheRuntimeSpec](#cacheruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `disabled` _boolean_ | Disabled indicates whether the component should be disabled.<br />If set to true, the component will not be created. |  | Optional: \{\} <br /> |
| `runtimeVersion` _[VersionSpec](#versionspec)_ | RuntimeVersion is the version information that instructs Fluid to orchestrate a particular version of the runtime. |  | Optional: \{\} <br /> |
| `options` _object (keys:string, values:string)_ | Options is a set of key-value pairs that provide additional configuration for the cache system. |  | Optional: \{\} <br /> |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata contains labels and annotations that will be propagated to the component's pods. |  | Optional: \{\} <br /> |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#resourcerequirements-v1-core)_ | Resources describes the compute resource requirements.<br />More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |  | Optional: \{\} <br /> |
| `env` _[EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#envvar-v1-core) array_ | Env is a list of environment variables to set in the container. |  | Optional: \{\} <br /> |
| `volumeMounts` _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#volumemount-v1-core) array_ | VolumeMounts are the volumes to mount into the container's filesystem.<br />Cannot be updated. |  | Optional: \{\} <br /> |
| `args` _string array_ | Args are arguments to the entrypoint.<br />The container image's CMD is used if this is not provided. |  | Optional: \{\} <br /> |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector which must be true for the pod to fit on a node.<br />Selector which must match a node's labels for the pod to be scheduled on that node.<br />More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/ |  | Optional: \{\} <br /> |
| `tolerations` _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#toleration-v1-core) array_ | Tolerations are the pod's tolerations.<br />If specified, the pod's tolerations.<br />More info: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/ |  | Optional: \{\} <br /> |
| `tieredStore` _[RuntimeTieredStore](#runtimetieredstore)_ | TieredStore describes the tiered storage configuration used by the client component. |  | Optional: \{\} <br /> |
| `cleanPolicy` _[FuseCleanPolicy](#fusecleanpolicy)_ | CleanPolicy determines when to clean up the FUSE client pods.<br />Currently supports two policies:<br />- "OnDemand": Clean up the FUSE pod when it is no longer needed on a node<br />- "OnRuntimeDeleted": Clean up the FUSE pod only when the CacheRuntime is deleted<br />Defaults to "OnRuntimeDeleted". | OnRuntimeDeleted | Enum: [OnRuntimeDeleted OnDemand] <br />Optional: \{\} <br /> |


#### CacheRuntimeMasterSpec



CacheRuntimeMasterSpec describes the desired state of CacheRuntime master component



_Appears in:_
- [CacheRuntimeSpec](#cacheruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `disabled` _boolean_ | Disabled indicates whether the component should be disabled.<br />If set to true, the component will not be created. |  | Optional: \{\} <br /> |
| `runtimeVersion` _[VersionSpec](#versionspec)_ | RuntimeVersion is the version information that instructs Fluid to orchestrate a particular version of the runtime. |  | Optional: \{\} <br /> |
| `options` _object (keys:string, values:string)_ | Options is a set of key-value pairs that provide additional configuration for the cache system. |  | Optional: \{\} <br /> |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata contains labels and annotations that will be propagated to the component's pods. |  | Optional: \{\} <br /> |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#resourcerequirements-v1-core)_ | Resources describes the compute resource requirements.<br />More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |  | Optional: \{\} <br /> |
| `env` _[EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#envvar-v1-core) array_ | Env is a list of environment variables to set in the container. |  | Optional: \{\} <br /> |
| `volumeMounts` _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#volumemount-v1-core) array_ | VolumeMounts are the volumes to mount into the container's filesystem.<br />Cannot be updated. |  | Optional: \{\} <br /> |
| `args` _string array_ | Args are arguments to the entrypoint.<br />The container image's CMD is used if this is not provided. |  | Optional: \{\} <br /> |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector which must be true for the pod to fit on a node.<br />Selector which must match a node's labels for the pod to be scheduled on that node.<br />More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/ |  | Optional: \{\} <br /> |
| `tolerations` _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#toleration-v1-core) array_ | Tolerations are the pod's tolerations.<br />If specified, the pod's tolerations.<br />More info: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/ |  | Optional: \{\} <br /> |
| `replicas` _integer_ | Replicas is the desired number of replicas of the master component.<br />If unspecified, defaults to 1. | 1 | Minimum: 0 <br />Optional: \{\} <br /> |


#### CacheRuntimeSpec



CacheRuntimeSpec describes the desired state of CacheRuntime



_Appears in:_
- [CacheRuntime](#cacheruntime)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `runtimeClassName` _string_ | RuntimeClassName is the name of the CacheRuntimeClass to use for this runtime.<br />The CacheRuntimeClass defines the implementation details of the cache runtime. |  | Required: \{\} <br /> |
| `master` _[CacheRuntimeMasterSpec](#cacheruntimemasterspec)_ | Master is the desired state of the master component. |  | Optional: \{\} <br /> |
| `worker` _[CacheRuntimeWorkerSpec](#cacheruntimeworkerspec)_ | Worker is the desired state of the worker component. |  | Optional: \{\} <br /> |
| `client` _[CacheRuntimeClientSpec](#cacheruntimeclientspec)_ | Client is the desired state of the client component. |  | Optional: \{\} <br /> |
| `options` _object (keys:string, values:string)_ | Options is a set of key-value pairs that provide additional configuration for the cache system.<br />These options will be propagated to all components. |  | Optional: \{\} <br /> |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata contains labels and annotations that will be propagated to all component pods. |  | Optional: \{\} <br /> |
| `imagePullSecrets` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#localobjectreference-v1-core) array_ | ImagePullSecrets is an optional list of references to secrets in the same namespace<br />to use for pulling any of the images used by this PodSpec.<br />More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod |  | Optional: \{\} <br /> |
| `volumes` _[Volume](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#volume-v1-core) array_ | Volumes is the list of volumes that can be mounted by containers belonging to the cache runtime components.<br />More info: https://kubernetes.io/docs/concepts/storage/volumes |  | Optional: \{\} <br /> |




#### CacheRuntimeWorkerSpec



CacheRuntimeWorkerSpec describes the desired state of CacheRuntime worker component



_Appears in:_
- [CacheRuntimeSpec](#cacheruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `disabled` _boolean_ | Disabled indicates whether the component should be disabled.<br />If set to true, the component will not be created. |  | Optional: \{\} <br /> |
| `runtimeVersion` _[VersionSpec](#versionspec)_ | RuntimeVersion is the version information that instructs Fluid to orchestrate a particular version of the runtime. |  | Optional: \{\} <br /> |
| `options` _object (keys:string, values:string)_ | Options is a set of key-value pairs that provide additional configuration for the cache system. |  | Optional: \{\} <br /> |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata contains labels and annotations that will be propagated to the component's pods. |  | Optional: \{\} <br /> |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#resourcerequirements-v1-core)_ | Resources describes the compute resource requirements.<br />More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |  | Optional: \{\} <br /> |
| `env` _[EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#envvar-v1-core) array_ | Env is a list of environment variables to set in the container. |  | Optional: \{\} <br /> |
| `volumeMounts` _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#volumemount-v1-core) array_ | VolumeMounts are the volumes to mount into the container's filesystem.<br />Cannot be updated. |  | Optional: \{\} <br /> |
| `args` _string array_ | Args are arguments to the entrypoint.<br />The container image's CMD is used if this is not provided. |  | Optional: \{\} <br /> |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector which must be true for the pod to fit on a node.<br />Selector which must match a node's labels for the pod to be scheduled on that node.<br />More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/ |  | Optional: \{\} <br /> |
| `tolerations` _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#toleration-v1-core) array_ | Tolerations are the pod's tolerations.<br />If specified, the pod's tolerations.<br />More info: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/ |  | Optional: \{\} <br /> |
| `replicas` _integer_ | Replicas is the desired number of replicas of the worker component.<br />If unspecified, defaults to 1. | 1 | Minimum: 0 <br />Optional: \{\} <br /> |
| `tieredStore` _[RuntimeTieredStore](#runtimetieredstore)_ | TieredStore describes the tiered storage configuration used by the worker component. |  | Optional: \{\} <br /> |


#### CacheableNodeAffinity



CacheableNodeAffinity defines constraints that limit what nodes this dataset can be cached to.



_Appears in:_
- [DatasetSpec](#datasetspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `required` _[NodeSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#nodeselector-v1-core)_ | Required specifies hard node constraints that must be met. |  |  |


#### CleanCachePolicy



CleanCachePolicy defines policies when cleaning cache



_Appears in:_
- [EFCRuntimeSpec](#efcruntimespec)
- [GooseFSRuntimeSpec](#goosefsruntimespec)
- [JindoRuntimeSpec](#jindoruntimespec)
- [RuntimeManagement](#runtimemanagement)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `gracePeriodSeconds` _integer_ | Optional duration in seconds the cache needs to clean gracefully. May be decreased in delete runtime request.<br />Value must be non-negative integer. The value zero indicates clean immediately via the timeout<br />command (no opportunity to shut down).<br />If this value is nil, the default grace period will be used instead.<br />The grace period is the duration in seconds after the processes running in the pod are sent<br />a termination signal and the time when the processes are forcibly halted with timeout command.<br />Set this value longer than the expected cleanup time for your process. | 60 | Optional: \{\} <br /> |
| `maxRetryAttempts` _integer_ | Optional max retry Attempts when cleanCache function returns an error after execution, runtime attempts<br />to run it three more times by default. With Maximum Retry Attempts, you can customize the maximum number<br />of retries. This gives you the option to continue processing retries. | 3 | Optional: \{\} <br /> |


#### ClientMetrics







_Appears in:_
- [JindoFuseSpec](#jindofusespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ | Enabled decides whether to expose client metrics. |  |  |
| `scrapeTarget` _string_ | ScrapeTarget decides which fuse component will be scraped by Prometheus.<br />It is a list separated by comma where supported items are [MountPod, Sidecar, All (indicates MountPod and Sidecar), None].<br />Defaults to None when it is not explicitly set. |  |  |


#### ComponentServiceConfig



ComponentServiceConfig defines the service configuration for runtime components.
Currently only headless service is supported, but this can be extended in the future
to support other service types such as ClusterIP, NodePort, LoadBalancer, etc.



_Appears in:_
- [RuntimeComponentService](#runtimecomponentservice)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `headless` _[HeadlessRuntimeComponentService](#headlessruntimecomponentservice)_ | Headless enables a headless service for the component.<br />A headless service does not allocate a cluster IP and allows direct pod-to-pod communication. |  | Optional: \{\} <br /> |


#### Condition



Condition explains the transitions on phase



_Appears in:_
- [OperationStatus](#operationstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[ConditionType](#conditiontype)_ | Type of condition, either `Complete` or `Failed` |  |  |
| `reason` _string_ | Reason for the condition's last transition |  |  |
| `message` _string_ | Message is a human-readable message indicating details about the transition |  |  |
| `lastProbeTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#time-v1-meta)_ | LastProbeTime describes last time this condition was updated. |  |  |
| `lastTransitionTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#time-v1-meta)_ | LastTransitionTime describes last time the condition transitioned from one status to another. |  |  |


#### ConfigMapDependencyConfig



ConfigMapDependencyConfig defines the ConfigMap mount configuration



_Appears in:_
- [ExtraResourcesComponentDependency](#extraresourcescomponentdependency)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the ConfigMap template name defined in extraResources.configMaps |  | Optional: \{\} <br /> |
| `mountPath` _string_ | MountPath is the path within the container at which the ConfigMap should be mounted.<br />Must not contain ':'. |  | Optional: \{\} <br /> |


#### ConfigMapRuntimeExtraResource



ConfigMapRuntimeExtraResource defines a ConfigMap template for CacheRuntime extra resources



_Appears in:_
- [RuntimeExtraResources](#runtimeextraresources)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the ConfigMap.<br />This will be used as the actual ConfigMap name when created in the runtime's namespace. |  | Optional: \{\} <br /> |
| `data` _object (keys:string, values:string)_ | Data contains the configuration data.<br />Each key must consist of alphanumeric characters, '-', '_' or '.'.<br />Values with non-UTF-8 byte sequences must use the BinaryData field.<br />The keys stored in Data must not overlap with the keys in<br />the BinaryData field, this is enforced during validation process. |  | Optional: \{\} <br /> |


#### Data



Data management strategies



_Appears in:_
- [AlluxioRuntimeSpec](#alluxioruntimespec)
- [GooseFSRuntimeSpec](#goosefsruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `replicas` _integer_ | The copies of the dataset |  | Optional: \{\} <br /> |
| `pin` _boolean_ | Pin the dataset or not. Refer to <a href="https://docs.alluxio.io/os/user/stable/en/operation/User-CLI.html#pin">Alluxio User-CLI pin</a> |  | Optional: \{\} <br /> |


#### DataBackup



DataBackup is the Schema for the backup API





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `data.fluid.io/v1alpha1` | | |
| `kind` _string_ | `DataBackup` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[DataBackupSpec](#databackupspec)_ |  |  |  |


#### DataBackupSpec



DataBackupSpec defines the desired state of DataBackup



_Appears in:_
- [DataBackup](#databackup)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `dataset` _string_ | Dataset defines the target dataset of the DataBackup |  |  |
| `backupPath` _string_ | BackupPath defines the target path to save data of the DataBackup |  |  |
| `runAs` _[User](#user)_ | Manage the user to run Alluxio DataBackup |  |  |
| `runAfter` _[OperationRef](#operationref)_ | Specifies that the preceding operation in a workflow |  | Optional: \{\} <br /> |
| `ttlSecondsAfterFinished` _integer_ | TTLSecondsAfterFinished is the time second to clean up data operations after finished or failed |  | Optional: \{\} <br /> |


#### DataLoad



DataLoad is the Schema for the dataloads API





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `data.fluid.io/v1alpha1` | | |
| `kind` _string_ | `DataLoad` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[DataLoadSpec](#dataloadspec)_ |  |  |  |


#### DataLoadSpec



DataLoadSpec defines the desired state of DataLoad



_Appears in:_
- [DataLoad](#dataload)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `dataset` _[TargetDataset](#targetdataset)_ | Dataset defines the target dataset of the DataLoad |  |  |
| `loadMetadata` _boolean_ | LoadMetadata specifies if the dataload job should load metadata |  |  |
| `target` _[TargetPath](#targetpath) array_ | Target defines target paths that needs to be loaded |  |  |
| `options` _object (keys:string, values:string)_ | Options specifies the extra dataload properties for runtime |  |  |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata defines labels and annotations that will be propagated to DataLoad pods |  |  |
| `affinity` _[Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#affinity-v1-core)_ | Affinity defines affinity for DataLoad pod |  | Optional: \{\} <br /> |
| `tolerations` _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#toleration-v1-core) array_ | Tolerations defines tolerations for DataLoad pod |  | Optional: \{\} <br /> |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector defiens node selector for DataLoad pod |  | Optional: \{\} <br /> |
| `schedulerName` _string_ | SchedulerName sets the scheduler to be used for DataLoad pod |  | Optional: \{\} <br /> |
| `policy` _[Policy](#policy)_ | including Once, Cron, OnEvent | Once | Enum: [Once Cron OnEvent] <br />Optional: \{\} <br /> |
| `schedule` _string_ | The schedule in Cron format, only set when policy is cron, see https://en.wikipedia.org/wiki/Cron. |  | Optional: \{\} <br /> |
| `runAfter` _[OperationRef](#operationref)_ | Specifies that the preceding operation in a workflow |  | Optional: \{\} <br /> |
| `ttlSecondsAfterFinished` _integer_ | TTLSecondsAfterFinished is the time second to clean up data operations after finished or failed |  | Optional: \{\} <br /> |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#resourcerequirements-v1-core)_ | Resources that will be requested by the DataLoad job. <br> |  | Optional: \{\} <br /> |


#### DataMigrate



DataMigrate is the Schema for the datamigrates API





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `data.fluid.io/v1alpha1` | | |
| `kind` _string_ | `DataMigrate` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[DataMigrateSpec](#datamigratespec)_ |  |  |  |


#### DataMigrateSpec



DataMigrateSpec defines the desired state of DataMigrate



_Appears in:_
- [DataMigrate](#datamigrate)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `image` _string_ | Image (e.g. alluxio/alluxio) |  |  |
| `imageTag` _string_ | Image tag (e.g. 2.3.0-SNAPSHOT) |  |  |
| `imagePullPolicy` _string_ | One of the three policies: `Always`, `IfNotPresent`, `Never` |  |  |
| `from` _[DataToMigrate](#datatomigrate)_ | data to migrate source, including dataset and external storage |  |  |
| `to` _[DataToMigrate](#datatomigrate)_ | data to migrate destination, including dataset and external storage |  |  |
| `block` _boolean_ | if dataMigrate blocked dataset usage, default is false |  | Optional: \{\} <br /> |
| `runtimeType` _string_ | using which runtime to migrate data; if none, take dataset runtime as default |  | Optional: \{\} <br /> |
| `options` _object (keys:string, values:string)_ | options for migrate, different for each runtime |  | Optional: \{\} <br /> |
| `policy` _[Policy](#policy)_ | policy for migrate, including Once, Cron, OnEvent | Once | Enum: [Once Cron OnEvent] <br />Optional: \{\} <br /> |
| `schedule` _string_ | The schedule in Cron format, only set when policy is cron, see https://en.wikipedia.org/wiki/Cron. |  | Optional: \{\} <br /> |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata defines labels and annotations that will be propagated to DataMigrate pods |  |  |
| `affinity` _[Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#affinity-v1-core)_ | Affinity defines affinity for DataMigrate pod |  | Optional: \{\} <br /> |
| `tolerations` _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#toleration-v1-core) array_ | Tolerations defines tolerations for DataMigrate pod |  | Optional: \{\} <br /> |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector defiens node selector for DataMigrate pod |  | Optional: \{\} <br /> |
| `schedulerName` _string_ | SchedulerName sets the scheduler to be used for DataMigrate pod |  | Optional: \{\} <br /> |
| `runAfter` _[OperationRef](#operationref)_ | Specifies that the preceding operation in a workflow |  | Optional: \{\} <br /> |
| `ttlSecondsAfterFinished` _integer_ | TTLSecondsAfterFinished is the time second to clean up data operations after finished or failed |  | Optional: \{\} <br /> |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#resourcerequirements-v1-core)_ | Resources that will be requested by the DataMigrate job. <br> |  | Optional: \{\} <br /> |
| `parallelism` _integer_ | Parallelism defines the parallelism tasks numbers for DataMigrate. If the value is greater than 1, the job acts<br />as a launcher, and users should define the WorkerSpec. | 1 | Minimum: 1 <br />Optional: \{\} <br /> |
| `parallelOptions` _object (keys:string, values:string)_ | ParallelOptions defines options like ssh port and ssh secret name when parallelism is greater than 1. |  | Optional: \{\} <br /> |


#### DataProcess



DataProcess is the Schema for the dataprocesses API





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `data.fluid.io/v1alpha1` | | |
| `kind` _string_ | `DataProcess` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[DataProcessSpec](#dataprocessspec)_ |  |  |  |


#### DataProcessSpec



DataProcessSpec defines the desired state of DataProcess



_Appears in:_
- [DataProcess](#dataprocess)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `dataset` _[TargetDatasetWithMountPath](#targetdatasetwithmountpath)_ | Dataset specifies the target dataset and its mount path. |  | Required: \{\} <br /> |
| `processor` _[Processor](#processor)_ | Processor specify how to process data. |  | Required: \{\} <br /> |
| `runAfter` _[OperationRef](#operationref)_ | Specifies that the preceding operation in a workflow |  | Optional: \{\} <br /> |
| `ttlSecondsAfterFinished` _integer_ | TTLSecondsAfterFinished is the time second to clean up data operations after finished or failed |  | Optional: \{\} <br /> |


#### DataRestoreLocation



DataRestoreLocation describes the spec restore location of  Dataset



_Appears in:_
- [DatasetSpec](#datasetspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `path` _string_ | Path describes the path of restore, in the form of  local://subpath or pvc://<pvcName>/subpath |  | Optional: \{\} <br /> |
| `nodeName` _string_ | NodeName describes the nodeName of restore if Path is  in the form of local://subpath |  | Optional: \{\} <br /> |


#### DataToMigrate







_Appears in:_
- [DataMigrateSpec](#datamigratespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `dataset` _[DatasetToMigrate](#datasettomigrate)_ | dataset to migrate |  |  |
| `externalStorage` _[ExternalStorage](#externalstorage)_ | external storage for data migrate |  |  |


#### Dataset



Dataset is the Schema for the datasets API





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `data.fluid.io/v1alpha1` | | |
| `kind` _string_ | `Dataset` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[DatasetSpec](#datasetspec)_ |  |  |  |


#### DatasetCondition



Condition describes the state of the cache at a certain point.



_Appears in:_
- [DatasetStatus](#datasetstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[DatasetConditionType](#datasetconditiontype)_ | Type of cache condition. |  |  |
| `reason` _string_ | The reason for the condition's last transition. |  |  |
| `message` _string_ | A human readable message indicating details about the transition. |  |  |
| `lastUpdateTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#time-v1-meta)_ | The last time this condition was updated. |  |  |
| `lastTransitionTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#time-v1-meta)_ | Last time the condition transitioned from one status to another. |  |  |


#### DatasetConditionType

_Underlying type:_ _string_

DatasetConditionType defines all kinds of types of cacheStatus.<br>
one of the three types: `RuntimeScheduled`, `Ready` and `Initialized`



_Appears in:_
- [DatasetCondition](#datasetcondition)

| Field | Description |
| --- | --- |
| `RuntimeScheduled` | RuntimeScheduled means the runtime CRD has been accepted by the system,<br />But master and workers are not ready<br /> |
| `Ready` | DatasetReady means the cache system for the dataset is ready.<br /> |
| `NotReady` | DatasetNotReady means the dataset is not bound due to some unexpected error<br /> |
| `UpdateReady` | DatasetUpdateReady means the cache system for the dataset is updated.<br /> |
| `Updating` | DatasetUpdating means the cache system for the dataset is updating.<br /> |
| `Initialized` | DatasetInitialized means the cache system for the dataset is Initialized.<br /> |


#### DatasetPhase

_Underlying type:_ _string_

DatasetPhase indicates whether the loading is behaving



_Appears in:_
- [DatasetStatus](#datasetstatus)

| Field | Description |
| --- | --- |
| `Pending` | TODO: add the Pending phase to Dataset<br /> |
| `Bound` | Bound to dataset, can't be released<br /> |
| `Failed` | Failed, can't be deleted<br /> |
| `NotBound` | Not bound to runtime, can be deleted<br /> |
| `Updating` | updating dataset, can't be released<br /> |
| `DataMigrating` | migrating dataset, can't be mounted<br /> |
| `` | the dataset have no phase and need to be judged<br /> |


#### DatasetSpec



DatasetSpec defines the desired state of Dataset



_Appears in:_
- [Dataset](#dataset)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `mounts` _[Mount](#mount) array_ | Mount Points to be mounted on cache runtime. <br><br />This field can be empty because some runtimes don't need to mount external storage (e.g.<br /><a href="https://v6d.io/">Vineyard</a>). |  | MinItems: 1 <br />UniqueItems: false <br />Optional: \{\} <br /> |
| `owner` _[User](#user)_ | The owner of the dataset |  | Optional: \{\} <br /> |
| `nodeAffinity` _[CacheableNodeAffinity](#cacheablenodeaffinity)_ | NodeAffinity defines constraints that limit what nodes this dataset can be cached to.<br />This field influences the scheduling of pods that use the cached dataset. |  | Optional: \{\} <br /> |
| `tolerations` _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#toleration-v1-core) array_ | If specified, the pod's tolerations. |  | Optional: \{\} <br /> |
| `accessModes` _[PersistentVolumeAccessMode](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#persistentvolumeaccessmode-v1-core) array_ | AccessModes contains all ways the volume backing the PVC can be mounted |  | Optional: \{\} <br /> |
| `runtimes` _[Runtime](#runtime) array_ | Runtimes for supporting dataset (e.g. AlluxioRuntime) |  |  |
| `placement` _[PlacementMode](#placementmode)_ | Manage switch for opening Multiple datasets single node deployment or not |  | Enum: [Exclusive  Shared] <br />Optional: \{\} <br /> |
| `dataRestoreLocation` _[DataRestoreLocation](#datarestorelocation)_ | DataRestoreLocation is the location to load data of dataset  been backuped |  | Optional: \{\} <br /> |
| `sharedOptions` _object (keys:string, values:string)_ | SharedOptions is the options to all mount |  | Optional: \{\} <br /> |
| `sharedEncryptOptions` _[EncryptOption](#encryptoption) array_ | SharedEncryptOptions is the encryptOption to all mount |  | Optional: \{\} <br /> |




#### DatasetToMigrate







_Appears in:_
- [DataToMigrate](#datatomigrate)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | name of dataset |  |  |
| `namespace` _string_ | namespace of dataset |  |  |
| `path` _string_ | path to migrate |  |  |


#### EFCCompTemplateSpec



EFCCompTemplateSpec is a description of the EFC components



_Appears in:_
- [EFCRuntimeSpec](#efcruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `replicas` _integer_ | Replicas is the desired number of replicas of the given template.<br />If unspecified, defaults to 1.<br />replicas is the min replicas of dataset in the cluster |  | Minimum: 1 <br />Optional: \{\} <br /> |
| `version` _[VersionSpec](#versionspec)_ | The version information that instructs fluid to orchestrate a particular version of EFC Comp |  |  |
| `properties` _object (keys:string, values:string)_ | Configurable properties for the EFC component. |  | Optional: \{\} <br /> |
| `ports` _object (keys:string, values:integer)_ | Ports used by EFC(e.g. rpc: 19998 for master). |  | Optional: \{\} <br /> |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#resourcerequirements-v1-core)_ | Resources that will be requested by the EFC component. <br><br /><br><br />Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources<br />already allocated to the pod. |  | Optional: \{\} <br /> |
| `disabled` _boolean_ | Enabled or Disabled for the components.<br />Default enable. |  | Optional: \{\} <br /> |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector which must be true for the component to fit on a node. |  | Optional: \{\} <br /> |
| `networkMode` _[NetworkMode](#networkmode)_ | Whether to use host network or not. |  | Enum: [HostNetwork  ContainerNetwork] <br />Optional: \{\} <br /> |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata defines labels and annotations that will be propagated to EFC's master and worker pods |  | Optional: \{\} <br /> |


#### EFCFuseSpec



EFCFuseSpec is a description of the EFC Fuse



_Appears in:_
- [EFCRuntimeSpec](#efcruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `version` _[VersionSpec](#versionspec)_ | The version information that instructs fluid to orchestrate a particular version of EFC Fuse |  |  |
| `properties` _object (keys:string, values:string)_ | Configurable properties for EFC fuse |  | Optional: \{\} <br /> |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#resourcerequirements-v1-core)_ | Resources that will be requested by EFC Fuse. <br><br /><br><br />Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources<br />already allocated to the pod. |  | Optional: \{\} <br /> |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector which must be true for the fuse client to fit on a node,<br />this option only effect when global is enabled |  | Optional: \{\} <br /> |
| `cleanPolicy` _[FuseCleanPolicy](#fusecleanpolicy)_ | CleanPolicy decides when to clean EFC Fuse pods.<br />Currently Fluid supports two policies: OnDemand and OnRuntimeDeleted<br />OnDemand cleans fuse pod once th fuse pod on some node is not needed<br />OnRuntimeDeleted cleans fuse pod only when the cache runtime is deleted<br />Defaults to OnRuntimeDeleted |  | Optional: \{\} <br /> |
| `networkMode` _[NetworkMode](#networkmode)_ | Whether to use hostnetwork or not |  | Enum: [HostNetwork  ContainerNetwork] <br />Optional: \{\} <br /> |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata defines labels and annotations that will be propagated to EFC's fuse pods |  | Optional: \{\} <br /> |


#### EFCRuntime



EFCRuntime is the Schema for the efcruntimes API





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `data.fluid.io/v1alpha1` | | |
| `kind` _string_ | `EFCRuntime` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[EFCRuntimeSpec](#efcruntimespec)_ |  |  |  |


#### EFCRuntimeSpec



EFCRuntimeSpec defines the desired state of EFCRuntime



_Appears in:_
- [EFCRuntime](#efcruntime)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `master` _[EFCCompTemplateSpec](#efccomptemplatespec)_ | The component spec of EFC master |  |  |
| `worker` _[EFCCompTemplateSpec](#efccomptemplatespec)_ | The component spec of EFC worker |  |  |
| `initFuse` _[InitFuseSpec](#initfusespec)_ | The spec of init alifuse |  |  |
| `fuse` _[EFCFuseSpec](#efcfusespec)_ | The component spec of EFC Fuse |  |  |
| `tieredstore` _[TieredStore](#tieredstore)_ | Tiered storage used by EFC worker |  |  |
| `replicas` _integer_ | The replicas of the worker, need to be specified |  |  |
| `osAdvise` _[OSAdvise](#osadvise)_ | Operating system optimization for EFC |  |  |
| `cleanCachePolicy` _[CleanCachePolicy](#cleancachepolicy)_ | CleanCachePolicy defines cleanCache Policy |  | Optional: \{\} <br /> |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata defines labels and annotations that will be propagated to all EFC's pods |  | Optional: \{\} <br /> |


#### EncryptOption







_Appears in:_
- [DatasetSpec](#datasetspec)
- [ExternalEndpointSpec](#externalendpointspec)
- [ExternalStorage](#externalstorage)
- [Mount](#mount)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | The name of encryptOption |  | Required: \{\} <br /> |
| `valueFrom` _[EncryptOptionSource](#encryptoptionsource)_ | The valueFrom of encryptOption |  | Optional: \{\} <br /> |


#### EncryptOptionComponentDependency



EncryptOptionComponentDependency defines the configuration for encrypt option dependency



_Appears in:_
- [RuntimeComponentDependencies](#runtimecomponentdependencies)



#### EncryptOptionSource







_Appears in:_
- [EncryptOption](#encryptoption)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `secretKeyRef` _[SecretKeySelector](#secretkeyselector)_ | The encryptInfo obtained from secret |  | Optional: \{\} <br /> |


#### ExternalEndpointSpec



ExternalEndpointSpec defines the configurations for external etcd cluster



_Appears in:_
- [MasterSpec](#masterspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `uri` _string_ | URI specifies the endpoint of external Etcd cluster<br />E,g. "etcd-svc.etcd-namespace.svc.cluster.local:2379"<br />Default is not set and use http protocol to connect to external etcd cluster |  | Optional: \{\} <br /> |
| `encryptOptions` _[EncryptOption](#encryptoption) array_ | encrypt info for accessing the external etcd cluster |  | Optional: \{\} <br /> |
| `options` _object (keys:string, values:string)_ | Configurable options for External Etcd cluster. |  | Optional: \{\} <br /> |


#### ExternalStorage







_Appears in:_
- [DataToMigrate](#datatomigrate)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `uri` _string_ | type of external storage, including s3, oss, gcs, ceph, nfs, pvc, etc. (related to runtime) |  |  |
| `encryptOptions` _[EncryptOption](#encryptoption) array_ | encrypt info for external storage |  | Optional: \{\} <br /> |


#### ExtraResourcesComponentDependency



ExtraResourcesComponentDependency defines the extra resources configuration for component dependencies



_Appears in:_
- [RuntimeComponentDependencies](#runtimecomponentdependencies)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `configMaps` _[ConfigMapDependencyConfig](#configmapdependencyconfig) array_ | ConfigMaps is a list of ConfigMaps in the same namespace to mount into the component |  | Optional: \{\} <br /> |


#### FuseCleanPolicy

_Underlying type:_ _string_





_Appears in:_
- [AlluxioFuseSpec](#alluxiofusespec)
- [CacheRuntimeClientSpec](#cacheruntimeclientspec)
- [EFCFuseSpec](#efcfusespec)
- [GooseFSFuseSpec](#goosefsfusespec)
- [JindoFuseSpec](#jindofusespec)
- [JuiceFSFuseSpec](#juicefsfusespec)
- [ThinFuseSpec](#thinfusespec)
- [VineyardClientSocketSpec](#vineyardclientsocketspec)

| Field | Description |
| --- | --- |
| `` | NoneCleanPolicy is the default clean policy. It will be transformed to OnRuntimeDeletedCleanPolicy automatically.<br /> |
| `OnDemand` | OnDemandCleanPolicy cleans fuse pod once the fuse pod on some node is not needed<br /> |
| `OnRuntimeDeleted` | OnRuntimeDeletedCleanPolicy cleans fuse pod only when the cache runtime is deleted<br /> |
| `OnFuseChanged` | OnFuseChangedCleanPolicy cleans fuse pod when the fuse in runtime is updated and the fuse pod on some node is not needed<br /> |


#### GooseFSCompTemplateSpec



GooseFSCompTemplateSpec is a description of the GooseFS commponents



_Appears in:_
- [GooseFSRuntimeSpec](#goosefsruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `replicas` _integer_ | Replicas is the desired number of replicas of the given template.<br />If unspecified, defaults to 1.<br />replicas is the min replicas of dataset in the cluster |  | Minimum: 1 <br />Optional: \{\} <br /> |
| `jvmOptions` _string array_ | Options for JVM |  |  |
| `properties` _object (keys:string, values:string)_ | Configurable properties for the GOOSEFS component. <br><br />Refer to <a href="https://cloud.tencent.com/document/product/436/56415">GOOSEFS Configuration Properties</a> for more info |  | Optional: \{\} <br /> |
| `ports` _object (keys:string, values:integer)_ | Ports used by GooseFS(e.g. rpc: 19998 for master) |  | Optional: \{\} <br /> |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#resourcerequirements-v1-core)_ | Resources that will be requested by the GooseFS component. <br><br /><br><br />Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources<br />already allocated to the pod. |  | Optional: \{\} <br /> |
| `env` _object (keys:string, values:string)_ | Environment variables that will be used by GooseFS component. <br> |  |  |
| `enabled` _boolean_ | Enabled or Disabled for the components. For now, only  API Gateway is enabled or disabled. |  | Optional: \{\} <br /> |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector which must be true for the master to fit on a node |  | Optional: \{\} <br /> |
| `annotations` _object (keys:string, values:string)_ | Annotations is an unstructured key value map stored with a resource that may be<br />set by external tools to store and retrieve arbitrary metadata. They are not<br />queryable and should be preserved when modifying objects.<br />More info: http://kubernetes.io/docs/user-guide/annotations |  | Optional: \{\} <br /> |


#### GooseFSFuseSpec



GooseFSFuseSpec is a description of the GooseFS Fuse



_Appears in:_
- [GooseFSRuntimeSpec](#goosefsruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `image` _string_ | Image for GooseFS Fuse(e.g. goosefs/goosefs-fuse) |  |  |
| `imageTag` _string_ | Image Tag for GooseFS Fuse(e.g. v1.0.1) |  |  |
| `imagePullPolicy` _string_ | One of the three policies: `Always`, `IfNotPresent`, `Never` |  |  |
| `jvmOptions` _string array_ | Options for JVM |  |  |
| `properties` _object (keys:string, values:string)_ | Configurable properties for the GOOSEFS component. <br><br />Refer to <a href="https://cloud.tencent.com/document/product/436/56415">GOOSEFS Configuration Properties</a> for more info |  |  |
| `env` _object (keys:string, values:string)_ | Environment variables that will be used by GooseFS Fuse |  |  |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#resourcerequirements-v1-core)_ | Resources that will be requested by GooseFS Fuse. <br><br /><br><br />Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources<br />already allocated to the pod. |  | Optional: \{\} <br /> |
| `args` _string array_ | Arguments that will be passed to GooseFS Fuse |  |  |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector which must be true for the fuse client to fit on a node,<br />this option only effect when global is enabled |  | Optional: \{\} <br /> |
| `cleanPolicy` _[FuseCleanPolicy](#fusecleanpolicy)_ | CleanPolicy decides when to clean GooseFS Fuse pods.<br />Currently Fluid supports two policies: OnDemand and OnRuntimeDeleted<br />OnDemand cleans fuse pod once th fuse pod on some node is not needed<br />OnRuntimeDeleted cleans fuse pod only when the cache runtime is deleted<br />Defaults to OnRuntimeDeleted |  | Optional: \{\} <br /> |
| `annotations` _object (keys:string, values:string)_ | Annotations is an unstructured key value map stored with a resource that may be<br />set by external tools to store and retrieve arbitrary metadata. They are not<br />queryable and should be preserved when modifying objects.<br />More info: http://kubernetes.io/docs/user-guide/annotations |  | Optional: \{\} <br /> |


#### GooseFSRuntime



GooseFSRuntime is the Schema for the goosefsruntimes API





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `data.fluid.io/v1alpha1` | | |
| `kind` _string_ | `GooseFSRuntime` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[GooseFSRuntimeSpec](#goosefsruntimespec)_ |  |  |  |


#### GooseFSRuntimeSpec



GooseFSRuntimeSpec defines the desired state of GooseFSRuntime



_Appears in:_
- [GooseFSRuntime](#goosefsruntime)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `goosefsVersion` _[VersionSpec](#versionspec)_ | The version information that instructs fluid to orchestrate a particular version of GooseFS. |  |  |
| `master` _[GooseFSCompTemplateSpec](#goosefscomptemplatespec)_ | The component spec of GooseFS master |  |  |
| `jobMaster` _[GooseFSCompTemplateSpec](#goosefscomptemplatespec)_ | The component spec of GooseFS job master |  |  |
| `worker` _[GooseFSCompTemplateSpec](#goosefscomptemplatespec)_ | The component spec of GooseFS worker |  |  |
| `jobWorker` _[GooseFSCompTemplateSpec](#goosefscomptemplatespec)_ | The component spec of GooseFS job Worker |  |  |
| `apiGateway` _[GooseFSCompTemplateSpec](#goosefscomptemplatespec)_ | The component spec of GooseFS API Gateway |  |  |
| `initUsers` _[InitUsersSpec](#initusersspec)_ | The spec of init users |  |  |
| `fuse` _[GooseFSFuseSpec](#goosefsfusespec)_ | The component spec of GooseFS Fuse |  |  |
| `properties` _object (keys:string, values:string)_ | Configurable properties for the GOOSEFS component. <br><br />Refer to <a href="https://cloud.tencent.com/document/product/436/56415">GOOSEFS Configuration Properties</a> for more info |  |  |
| `jvmOptions` _string array_ | Options for JVM |  |  |
| `tieredstore` _[TieredStore](#tieredstore)_ | Tiered storage used by GooseFS |  |  |
| `data` _[Data](#data)_ | Management strategies for the dataset to which the runtime is bound |  |  |
| `replicas` _integer_ | The replicas of the worker, need to be specified |  |  |
| `runAs` _[User](#user)_ | Manage the user to run GooseFS Runtime<br />GooseFS support POSIX-ACL and Apache Ranger to manager authorization |  |  |
| `disablePrometheus` _boolean_ | Disable monitoring for GooseFS Runtime<br />Prometheus is enabled by default |  | Optional: \{\} <br /> |
| `hadoopConfig` _string_ | Name of the configMap used to support HDFS configurations when using HDFS as GooseFS's UFS. The configMap<br />must be in the same namespace with the GooseFSRuntime. The configMap should contain user-specific HDFS conf files in it.<br />For now, only "hdfs-site.xml" and "core-site.xml" are supported. It must take the filename of the conf file as the key and content<br />of the file as the value. |  | Optional: \{\} <br /> |
| `cleanCachePolicy` _[CleanCachePolicy](#cleancachepolicy)_ | CleanCachePolicy defines cleanCache Policy |  | Optional: \{\} <br /> |


#### HCFSStatus



HCFS Endpoint info



_Appears in:_
- [DatasetStatus](#datasetstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `endpoint` _string_ | Endpoint for accessing |  |  |
| `underlayerFileSystemVersion` _string_ | Underlayer HCFS Compatible Version |  |  |


#### HeadlessRuntimeComponentService



HeadlessRuntimeComponentService defines the configuration for headless service



_Appears in:_
- [ComponentServiceConfig](#componentserviceconfig)
- [RuntimeComponentService](#runtimecomponentservice)



#### InitFuseSpec



InitFuseSpec is a description of initialize the fuse kernel module for runtime



_Appears in:_
- [EFCRuntimeSpec](#efcruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `version` _[VersionSpec](#versionspec)_ | The version information that instructs fluid to orchestrate a particular version of Alifuse |  |  |


#### InitUsersSpec



InitUsersSpec is a description of the initialize the users for runtime



_Appears in:_
- [AlluxioRuntimeSpec](#alluxioruntimespec)
- [GooseFSRuntimeSpec](#goosefsruntimespec)
- [JuiceFSRuntimeSpec](#juicefsruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `image` _string_ | Image for initialize the users for runtime(e.g. alluxio/alluxio-User init) |  |  |
| `imageTag` _string_ | Image Tag for initialize the users for runtime(e.g. 2.3.0-SNAPSHOT) |  |  |
| `imagePullPolicy` _string_ | One of the three policies: `Always`, `IfNotPresent`, `Never` |  |  |
| `env` _object (keys:string, values:string)_ | Environment variables that will be used by initialize the users for runtime |  |  |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#resourcerequirements-v1-core)_ | Resources that will be requested by initialize the users for runtime. <br><br /><br><br />Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources<br />already allocated to the pod. |  | Optional: \{\} <br /> |


#### JindoCompTemplateSpec



JindoCompTemplateSpec is a description of the Jindo commponents



_Appears in:_
- [JindoRuntimeSpec](#jindoruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `replicas` _integer_ | Replicas is the desired number of replicas of the given template.<br />If unspecified, defaults to 1.<br />replicas is the min replicas of dataset in the cluster |  | Minimum: 1 <br />Optional: \{\} <br /> |
| `properties` _object (keys:string, values:string)_ | Configurable properties for the Jindo component. <br> |  | Optional: \{\} <br /> |
| `ports` _object (keys:string, values:integer)_ |  |  | Optional: \{\} <br /> |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#resourcerequirements-v1-core)_ | Resources that will be requested by the Jindo component. <br><br /><br><br />Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources<br />already allocated to the pod. |  | Optional: \{\} <br /> |
| `env` _object (keys:string, values:string)_ | Environment variables that will be used by Jindo component. <br> |  |  |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector which must be true for the master to fit on a node |  | Optional: \{\} <br /> |
| `tolerations` _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#toleration-v1-core) array_ | If specified, the pod's tolerations. |  | Optional: \{\} <br /> |
| `labels` _object (keys:string, values:string)_ | Labels will be added on JindoFS Master or Worker pods.<br />DEPRECATED: This is a deprecated field. Please use PodMetadata instead.<br />Note: this field is set to be exclusive with PodMetadata.Labels |  | Optional: \{\} <br /> |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata defines labels and annotations that will be propagated to Jindo's pods |  | Optional: \{\} <br /> |
| `disabled` _boolean_ | If disable JindoFS master or worker |  | Optional: \{\} <br /> |
| `volumeMounts` _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#volumemount-v1-core) array_ | VolumeMounts specifies the volumes listed in ".spec.volumes" to mount into the jindo runtime component's filesystem. |  | Optional: \{\} <br /> |
| `imagePullSecrets` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#localobjectreference-v1-core) array_ | ImagePullSecrets that will be used to pull images |  |  |


#### JindoFuseSpec



JindoFuseSpec is a description of the Jindo Fuse



_Appears in:_
- [JindoRuntimeSpec](#jindoruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `image` _string_ | Image for Jindo Fuse(e.g. jindo/jindo-fuse) |  |  |
| `imageTag` _string_ | Image Tag for Jindo Fuse(e.g. 2.3.0-SNAPSHOT) |  |  |
| `imagePullPolicy` _string_ | One of the three policies: `Always`, `IfNotPresent`, `Never` |  |  |
| `imagePullSecrets` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#localobjectreference-v1-core) array_ | ImagePullSecrets that will be used to pull images |  |  |
| `properties` _object (keys:string, values:string)_ | Configurable properties for Jindo System. <br> |  |  |
| `env` _object (keys:string, values:string)_ | Environment variables that will be used by Jindo Fuse |  |  |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#resourcerequirements-v1-core)_ | Resources that will be requested by Jindo Fuse. <br><br /><br><br />Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources<br />already allocated to the pod. |  | Optional: \{\} <br /> |
| `args` _string array_ | Arguments that will be passed to Jindo Fuse |  |  |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector which must be true for the fuse client to fit on a node,<br />this option only effect when global is enabled |  | Optional: \{\} <br /> |
| `tolerations` _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#toleration-v1-core) array_ | If specified, the pod's tolerations. |  | Optional: \{\} <br /> |
| `labels` _object (keys:string, values:string)_ | Labels will be added on all the JindoFS pods.<br />DEPRECATED: this is a deprecated field. Please use PodMetadata.Labels instead.<br />Note: this field is set to be exclusive with PodMetadata.Labels |  | Optional: \{\} <br /> |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata defines labels and annotations that will be propagated to Jindo's fuse pods |  | Optional: \{\} <br /> |
| `cleanPolicy` _[FuseCleanPolicy](#fusecleanpolicy)_ | CleanPolicy decides when to clean JindoFS Fuse pods.<br />Currently Fluid supports two policies: OnDemand and OnRuntimeDeleted<br />OnDemand cleans fuse pod once th fuse pod on some node is not needed<br />OnRuntimeDeleted cleans fuse pod only when the cache runtime is deleted<br />Defaults to OnRuntimeDeleted |  | Optional: \{\} <br /> |
| `disabled` _boolean_ | If disable JindoFS fuse |  | Optional: \{\} <br /> |
| `logConfig` _object (keys:string, values:string)_ |  |  | Optional: \{\} <br /> |
| `metrics` _[ClientMetrics](#clientmetrics)_ | Define whether fuse metrics will be enabled. |  | Optional: \{\} <br /> |


#### JindoRuntime



JindoRuntime is the Schema for the jindoruntimes API





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `data.fluid.io/v1alpha1` | | |
| `kind` _string_ | `JindoRuntime` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[JindoRuntimeSpec](#jindoruntimespec)_ |  |  |  |


#### JindoRuntimeSpec



JindoRuntimeSpec defines the desired state of JindoRuntime



_Appears in:_
- [JindoRuntime](#jindoruntime)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `jindoVersion` _[VersionSpec](#versionspec)_ | The version information that instructs fluid to orchestrate a particular version of Jindo. |  |  |
| `master` _[JindoCompTemplateSpec](#jindocomptemplatespec)_ | The component spec of Jindo master |  |  |
| `worker` _[JindoCompTemplateSpec](#jindocomptemplatespec)_ | The component spec of Jindo worker |  |  |
| `fuse` _[JindoFuseSpec](#jindofusespec)_ | The component spec of Jindo Fuse |  |  |
| `properties` _object (keys:string, values:string)_ | Configurable properties for Jindo system. <br> |  |  |
| `tieredstore` _[TieredStore](#tieredstore)_ | Tiered storage used by Jindo |  |  |
| `replicas` _integer_ | The replicas of the worker, need to be specified |  |  |
| `runAs` _[User](#user)_ | Manage the user to run Jindo Runtime |  |  |
| `user` _string_ |  |  |  |
| `hadoopConfig` _string_ | Name of the configMap used to support HDFS configurations when using HDFS as Jindo's UFS. The configMap<br />must be in the same namespace with the JindoRuntime. The configMap should contain user-specific HDFS conf files in it.<br />For now, only "hdfs-site.xml" and "core-site.xml" are supported. It must take the filename of the conf file as the key and content<br />of the file as the value. |  | Optional: \{\} <br /> |
| `secret` _string_ |  |  |  |
| `labels` _object (keys:string, values:string)_ | Labels will be added on all the JindoFS pods.<br />DEPRECATED: this is a deprecated field. Please use PodMetadata.Labels instead.<br />Note: this field is set to be exclusive with PodMetadata.Labels |  | Optional: \{\} <br /> |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata defines labels and annotations that will be propagated to all Jindo's fuse pods |  | Optional: \{\} <br /> |
| `logConfig` _object (keys:string, values:string)_ |  |  | Optional: \{\} <br /> |
| `networkmode` _[NetworkMode](#networkmode)_ | Whether to use hostnetwork or not |  | Enum: [HostNetwork  ContainerNetwork] <br />Optional: \{\} <br /> |
| `cleanCachePolicy` _[CleanCachePolicy](#cleancachepolicy)_ | CleanCachePolicy defines cleanCache Policy |  | Optional: \{\} <br /> |
| `volumes` _[Volume](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#volume-v1-core) array_ | Volumes is the list of Kubernetes volumes that can be mounted by the jindo runtime components and/or fuses. |  | Optional: \{\} <br /> |
| `imagePullSecrets` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#localobjectreference-v1-core) array_ | ImagePullSecrets that will be used to pull images |  | Optional: \{\} <br /> |


#### JobProcessor







_Appears in:_
- [Processor](#processor)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `podSpec` _[PodSpec](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#podspec-v1-core)_ | PodSpec defines Pod specification of the DataProcess job. |  | Optional: \{\} <br /> |


#### JuiceFSCompTemplateSpec



JuiceFSCompTemplateSpec is a description of the JuiceFS components



_Appears in:_
- [JuiceFSRuntimeSpec](#juicefsruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `replicas` _integer_ | Replicas is the desired number of replicas of the given template.<br />If unspecified, defaults to 1.<br />replicas is the min replicas of dataset in the cluster |  | Minimum: 1 <br />Optional: \{\} <br /> |
| `ports` _[ContainerPort](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#containerport-v1-core) array_ | Ports used by JuiceFS |  | Optional: \{\} <br /> |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#resourcerequirements-v1-core)_ | Resources that will be requested by the JuiceFS component. |  | Optional: \{\} <br /> |
| `options` _object (keys:string, values:string)_ | Options |  |  |
| `env` _[EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#envvar-v1-core) array_ | Environment variables that will be used by JuiceFS component. |  |  |
| `enabled` _boolean_ | Enabled or Disabled for the components. |  | Optional: \{\} <br /> |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector |  | Optional: \{\} <br /> |
| `volumeMounts` _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#volumemount-v1-core) array_ | VolumeMounts specifies the volumes listed in ".spec.volumes" to mount into runtime component's filesystem. |  | Optional: \{\} <br /> |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata defines labels and annotations that will be propagated to JuiceFs's pods. |  | Optional: \{\} <br /> |
| `networkMode` _[NetworkMode](#networkmode)_ | Whether to use hostnetwork or not |  | Enum: [HostNetwork  ContainerNetwork] <br />Optional: \{\} <br /> |


#### JuiceFSFuseSpec







_Appears in:_
- [JuiceFSRuntimeSpec](#juicefsruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `image` _string_ | Image for JuiceFS fuse |  |  |
| `imageTag` _string_ | Image for JuiceFS fuse |  |  |
| `imagePullPolicy` _string_ | One of the three policies: `Always`, `IfNotPresent`, `Never` |  |  |
| `env` _[EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#envvar-v1-core) array_ | Environment variables that will be used by JuiceFS Fuse |  |  |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#resourcerequirements-v1-core)_ | Resources that will be requested by JuiceFS Fuse. |  |  |
| `options` _object (keys:string, values:string)_ | Options mount options that fuse pod will use |  | Optional: \{\} <br /> |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector which must be true for the fuse client to fit on a node,<br />this option only effect when global is enabled |  | Optional: \{\} <br /> |
| `volumeMounts` _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#volumemount-v1-core) array_ | VolumeMounts specifies the volumes listed in ".spec.volumes" to mount into runtime component's filesystem. |  | Optional: \{\} <br /> |
| `cleanPolicy` _[FuseCleanPolicy](#fusecleanpolicy)_ | CleanPolicy decides when to clean Juicefs Fuse pods.<br />Currently Fluid supports three policies: OnDemand, OnRuntimeDeleted and OnFuseChangedCleanPolicy<br />OnDemand cleans fuse pod once the fuse pod on some node is not needed<br />OnRuntimeDeleted cleans fuse pod only when the cache runtime is deleted<br />OnFuseChangedCleanPolicy cleans fuse pod once the fuse pod on some node is not needed and the fuse in runtime is updated<br />Defaults to OnRuntimeDeleted |  | Optional: \{\} <br /> |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata defines labels and annotations that will be propagated to JuiceFs's pods. |  | Optional: \{\} <br /> |
| `networkMode` _[NetworkMode](#networkmode)_ | Whether to use hostnetwork or not |  | Enum: [HostNetwork  ContainerNetwork] <br />Optional: \{\} <br /> |


#### JuiceFSRuntime



JuiceFSRuntime is the Schema for the juicefsruntimes API





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `data.fluid.io/v1alpha1` | | |
| `kind` _string_ | `JuiceFSRuntime` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[JuiceFSRuntimeSpec](#juicefsruntimespec)_ |  |  |  |


#### JuiceFSRuntimeSpec



JuiceFSRuntimeSpec defines the desired state of JuiceFSRuntime



_Appears in:_
- [JuiceFSRuntime](#juicefsruntime)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `juicefsVersion` _[VersionSpec](#versionspec)_ | The version information that instructs fluid to orchestrate a particular version of JuiceFS. |  |  |
| `initUsers` _[InitUsersSpec](#initusersspec)_ | The spec of init users |  |  |
| `master` _[JuiceFSCompTemplateSpec](#juicefscomptemplatespec)_ | The component spec of JuiceFS master |  |  |
| `worker` _[JuiceFSCompTemplateSpec](#juicefscomptemplatespec)_ | The component spec of JuiceFS worker |  |  |
| `jobWorker` _[JuiceFSCompTemplateSpec](#juicefscomptemplatespec)_ | The component spec of JuiceFS job Worker |  |  |
| `fuse` _[JuiceFSFuseSpec](#juicefsfusespec)_ | Desired state for JuiceFS Fuse |  |  |
| `tieredstore` _[TieredStore](#tieredstore)_ | Tiered storage used by JuiceFS |  |  |
| `configs` _string_ | Configs of JuiceFS |  |  |
| `replicas` _integer_ | The replicas of the worker, need to be specified |  |  |
| `runAs` _[User](#user)_ | Manage the user to run Juicefs Runtime |  |  |
| `disablePrometheus` _boolean_ | Disable monitoring for JuiceFS Runtime<br />Prometheus is enabled by default |  | Optional: \{\} <br /> |
| `volumes` _[Volume](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#volume-v1-core) array_ | Volumes is the list of Kubernetes volumes that can be mounted by the alluxio runtime components and/or fuses. |  | Optional: \{\} <br /> |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata defines labels and annotations that will be propagated to JuiceFs's pods. |  | Optional: \{\} <br /> |
| `management` _[RuntimeManagement](#runtimemanagement)_ | RuntimeManagement defines policies when managing the runtime |  | Optional: \{\} <br /> |


#### Level



Level describes configurations a tier needs. <br>
Refer to <a href="https://docs.alluxio.io/os/user/stable/en/core-services/Caching.html#configuring-tiered-storage">Configuring Tiered Storage</a> for more info



_Appears in:_
- [TieredStore](#tieredstore)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `mediumtype` _[MediumType](#mediumtype)_ | Medium Type of the tier. One of the three types: `MEM`, `SSD`, `HDD` |  | Enum: [MEM SSD HDD] <br />Required: \{\} <br /> |
| `volumeType` _[VolumeType](#volumetype)_ | VolumeType is the volume type of the tier. Should be one of the three types: `hostPath`, `emptyDir` and `volumeTemplate`.<br />If not set, defaults to hostPath. | hostPath | Enum: [hostPath emptyDir] <br />Optional: \{\} <br /> |
| `path` _string_ | File paths to be used for the tier. Multiple paths are supported.<br />Multiple paths should be separated with comma. For example: "/mnt/cache1,/mnt/cache2". |  | MinLength: 1 <br />Optional: \{\} <br /> |
| `quota` _[Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#quantity-resource-api)_ | Quota for the whole tier. (e.g. 100Gi)<br />Please note that if there're multiple paths used for this tierstore,<br />the quota will be equally divided into these paths. If you'd like to<br />set quota for each, path, see QuotaList for more information. |  | Optional: \{\} <br /> |
| `quotaList` _string_ | QuotaList are quotas used to set quota on multiple paths. Quotas should be separated with comma.<br />Quotas in this list will be set to paths with the same order in Path.<br />For example, with Path defined with "/mnt/cache1,/mnt/cache2" and QuotaList set to "100Gi, 50Gi",<br />then we get 100GiB cache storage under "/mnt/cache1" and 50GiB under "/mnt/cache2".<br />Also note that num of quotas must be consistent with the num of paths defined in Path. |  | Pattern: `^((\+\|-)?(([0-9]+(\.[0-9]*)?)\|(\.[0-9]+))(([KMGTPE]i)\|[numkMGTPE]\|([eE](\+\|-)?(([0-9]+(\.[0-9]*)?)\|(\.[0-9]+)))),)+((\+\|-)?(([0-9]+(\.[0-9]*)?)\|(\.[0-9]+))(([KMGTPE]i)\|[numkMGTPE]\|([eE](\+\|-)?(([0-9]+(\.[0-9]*)?)\|(\.[0-9]+))))?)$` <br />Optional: \{\} <br /> |
| `high` _string_ | Ratio of high watermark of the tier (e.g. 0.9) |  |  |
| `low` _string_ | Ratio of low watermark of the tier (e.g. 0.7) |  |  |


#### MasterSpec



MasterSpec defines the configurations for Vineyard Master component
which is also regarded as the Etcd component in Vineyard.
For more info about Vineyard, refer to <a href="https://v6d.io/">Vineyard</a>



_Appears in:_
- [VineyardRuntimeSpec](#vineyardruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `replicas` _integer_ | The replicas of Vineyard component.<br />If not specified, defaults to 1.<br />For worker, the replicas should not be greater than the number of nodes in the cluster |  | Minimum: 1 <br />Optional: \{\} <br /> |
| `image` _string_ | The image of Vineyard component.<br />For Master, the default image is `registry.aliyuncs.com/vineyard/vineyardd`<br />For Worker, the default image is `registry.aliyuncs.com/vineyard/vineyardd`<br />The default container registry is `docker.io`, you can change it by setting the image field |  | Optional: \{\} <br /> |
| `imageTag` _string_ | The image tag of Vineyard component.<br />For Master, the default image tag is `v0.22.2`.<br />For Worker, the default image tag is `v0.22.2`. |  | Optional: \{\} <br /> |
| `imagePullPolicy` _string_ | The image pull policy of Vineyard component.<br />Default is `IfNotPresent`. |  | Optional: \{\} <br /> |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector to choose which nodes to launch the Vineyard component.<br />E,g. \{"disktype": "ssd"\} |  | Optional: \{\} <br /> |
| `ports` _object (keys:string, values:integer)_ | Ports used by Vineyard component.<br />For Master, the default client port is 2379 and peer port is 2380.<br />For Worker, the default rpc port is 9600 and the default exporter port is 9144. |  | Optional: \{\} <br /> |
| `env` _object (keys:string, values:string)_ | Environment variables that will be used by Vineyard component.<br />For Master, refer to <a href="https://etcd.io/docs/v3.5/op-guide/configuration/">Etcd Configuration</a> for more info<br />Default is not set. |  | Optional: \{\} <br /> |
| `options` _object (keys:string, values:string)_ | Configurable options for Vineyard component.<br />For Master, there is no configurable options.<br />For Worker, support the following options.<br />  vineyardd.reserve.memory: (Bool) where to reserve memory for vineyardd<br />                            If set to true, the memory quota will be counted to the vineyardd rather than the application.<br />  etcd.prefix: (String) the prefix of etcd key for vineyard objects<br />  wait.etcd.timeout: (String) the timeout period before waiting the etcd to be ready, in seconds<br />  Default value is as follows.<br />    vineyardd.reserve.memory: "true"<br />    etcd.prefix: "/vineyard"<br />    wait.etcd.timeout: "120" |  | Optional: \{\} <br /> |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#resourcerequirements-v1-core)_ | Resources contains the resource requirements and limits for the Vineyard component.<br />Default is not set.<br />For Worker, when the options contains vineyardd.reserve.memory=true,<br />the resources.request.memory for worker should be greater than tieredstore.levels[0].quota(aka vineyardd shared memory) |  | Optional: \{\} <br /> |
| `volumeMounts` _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#volumemount-v1-core) array_ | VolumeMounts specifies the volumes listed in ".spec.volumes" to mount into the vineyard runtime component's filesystem.<br />It is useful for specifying a persistent storage.<br />Default is not set. |  | Optional: \{\} <br /> |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata defines labels and annotations that will be propagated to Vineyard's pods including Master and Worker.<br />Default is not set. |  | Optional: \{\} <br /> |
| `networkMode` _[NetworkMode](#networkmode)_ | Whether to use hostnetwork or not<br />Default is HostNetwork |  | Enum: [HostNetwork  ContainerNetwork] <br />Optional: \{\} <br /> |
| `endpoint` _[ExternalEndpointSpec](#externalendpointspec)_ | ExternalEndpoint defines the configurations for external etcd cluster<br />Default is not set<br />If set, the Vineyard Master component will not be deployed,<br />which means the Vineyard Worker component will use an external Etcd cluster.<br />E,g.<br />  endpoint:<br />    uri: "etcd-svc.etcd-namespace.svc.cluster.local:2379"<br />    encryptOptions:<br />      - name: access-key<br />		   valueFrom:<br />          secretKeyRef:<br />            name: etcd-secret<br />			   key: accesskey |  | Optional: \{\} <br /> |


#### MediumSource



MediumSource describes the storage medium type for tiered store.
Only one of its members may be specified.



_Appears in:_
- [RuntimeTieredStoreLevel](#runtimetieredstorelevel)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `processMemory` _[ProcessMemoryMediumSource](#processmemorymediumsource)_ | ProcessMemory indicates that process memory should be used as the storage medium.<br />The cache will be stored in the process's memory space. |  | Optional: \{\} <br /> |
| `volume` _[VolumeMediumSource](#volumemediumsource)_ | Volume indicates that a Kubernetes volume should be used as the storage medium.<br />Supported volume types include hostPath, emptyDir, and ephemeral volumes. |  | Optional: \{\} <br /> |




#### MetadataSyncPolicy



MetadataSyncPolicy defines policies when syncing metadata



_Appears in:_
- [RuntimeManagement](#runtimemanagement)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `autoSync` _boolean_ | AutoSync enables automatic metadata sync when setting up a runtime. If not set, it defaults to true. |  | Optional: \{\} <br /> |


#### Mount



Mount describes a mounting. <br>
Refer to <a href="https://docs.alluxio.io/os/user/stable/en/ufs/S3.html">Alluxio Storage Integrations</a> for more info



_Appears in:_
- [DatasetSpec](#datasetspec)
- [DatasetStatus](#datasetstatus)
- [MountPointStatus](#mountpointstatus)
- [RuntimeStatus](#runtimestatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `mountPoint` _string_ | MountPoint is the mount point of source. |  | MinLength: 5 <br />Required: \{\} <br /> |
| `options` _object (keys:string, values:string)_ | The Mount Options. <br><br />Refer to <a href="https://docs.alluxio.io/os/user/stable/en/reference/Properties-List.html">Mount Options</a>.  <br><br />The option has Prefix 'fs.' And you can Learn more from<br /><a href="https://docs.alluxio.io/os/user/stable/en/ufs/S3.html">The Storage Integrations</a> |  | Optional: \{\} <br /> |
| `name` _string_ | The name of mount |  | MinLength: 0 <br />Optional: \{\} <br /> |
| `path` _string_ | The path of mount, if not set will be /\{Name\} |  | Optional: \{\} <br /> |
| `readOnly` _boolean_ | Optional: Defaults to false (read-write). |  | Optional: \{\} <br /> |
| `shared` _boolean_ | Optional: Defaults to false (shared). |  | Optional: \{\} <br /> |
| `encryptOptions` _[EncryptOption](#encryptoption) array_ | The secret information |  | Optional: \{\} <br /> |


#### MountPointStatus



MountPointStatus describes the status of a single mount point in the dataset



_Appears in:_
- [CacheRuntimeStatus](#cacheruntimestatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `mount` _[Mount](#mount)_ | Mount contains the mount point configuration from the bound dataset.<br />This includes the remote path, mount options, and other mount-specific settings. |  |  |
| `mountTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#time-v1-meta)_ | MountTime is the timestamp of the last successful mount operation.<br />If MountTime is earlier than the master component's start time, a remount will be required. |  | Optional: \{\} <br /> |


#### NetworkMode

_Underlying type:_ _string_





_Appears in:_
- [AlluxioCompTemplateSpec](#alluxiocomptemplatespec)
- [AlluxioFuseSpec](#alluxiofusespec)
- [EFCCompTemplateSpec](#efccomptemplatespec)
- [EFCFuseSpec](#efcfusespec)
- [JindoRuntimeSpec](#jindoruntimespec)
- [JuiceFSCompTemplateSpec](#juicefscomptemplatespec)
- [JuiceFSFuseSpec](#juicefsfusespec)
- [MasterSpec](#masterspec)
- [ThinCompTemplateSpec](#thincomptemplatespec)
- [ThinFuseSpec](#thinfusespec)
- [VineyardClientSocketSpec](#vineyardclientsocketspec)
- [VineyardCompTemplateSpec](#vineyardcomptemplatespec)

| Field | Description |
| --- | --- |
| `HostNetwork` |  |
| `ContainerNetwork` |  |
| `` | DefaultNetworkMode is Host<br /> |


#### NodePublishSecretPolicy

_Underlying type:_ _string_





_Appears in:_
- [ThinRuntimeProfileSpec](#thinruntimeprofilespec)

| Field | Description |
| --- | --- |
| `NotMountNodePublishSecret` |  |
| `MountNodePublishSecretIfExists` |  |
| `CopyNodePublishSecretAndMountIfNotExists` |  |


#### OSAdvise



OSAdvise is a description of choices to have optimization on specific operating system



_Appears in:_
- [EFCRuntimeSpec](#efcruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `osVersion` _string_ | Specific operating system version that can have optimization. |  | Optional: \{\} <br /> |
| `enabled` _boolean_ | Enable operating system optimization<br />not enabled by default. |  | Optional: \{\} <br /> |


#### ObjectRef







_Appears in:_
- [AffinityStrategy](#affinitystrategy)
- [OperationRef](#operationref)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | API version of the referent operation |  | Optional: \{\} <br /> |
| `kind` _string_ | Kind specifies the type of the referent operation |  | Enum: [DataLoad DataBackup DataMigrate DataProcess] <br />Required: \{\} <br /> |
| `name` _string_ | Name specifies the name of the referent operation |  | Required: \{\} <br /> |
| `namespace` _string_ | Namespace specifies the namespace of the referent operation. |  | Optional: \{\} <br /> |


#### OperationRef







_Appears in:_
- [DataBackupSpec](#databackupspec)
- [DataLoadSpec](#dataloadspec)
- [DataMigrateSpec](#datamigratespec)
- [DataProcessSpec](#dataprocessspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | API version of the referent operation |  | Optional: \{\} <br /> |
| `kind` _string_ | Kind specifies the type of the referent operation |  | Enum: [DataLoad DataBackup DataMigrate DataProcess] <br />Required: \{\} <br /> |
| `name` _string_ | Name specifies the name of the referent operation |  | Required: \{\} <br /> |
| `namespace` _string_ | Namespace specifies the namespace of the referent operation. |  | Optional: \{\} <br /> |
| `affinityStrategy` _[AffinityStrategy](#affinitystrategy)_ | AffinityStrategy specifies the pod affinity strategy with the referent operation. |  | Optional: \{\} <br /> |




#### PlacementMode

_Underlying type:_ _string_





_Appears in:_
- [DatasetSpec](#datasetspec)

| Field | Description |
| --- | --- |
| `Exclusive` |  |
| `Shared` |  |
| `` | DefaultMode is exclusive<br /> |


#### PodMetadata



PodMetadata defines subgroup properties of metav1.ObjectMeta



_Appears in:_
- [AlluxioCompTemplateSpec](#alluxiocomptemplatespec)
- [AlluxioFuseSpec](#alluxiofusespec)
- [AlluxioRuntimeSpec](#alluxioruntimespec)
- [CacheRuntimeClientSpec](#cacheruntimeclientspec)
- [CacheRuntimeMasterSpec](#cacheruntimemasterspec)
- [CacheRuntimeSpec](#cacheruntimespec)
- [CacheRuntimeWorkerSpec](#cacheruntimeworkerspec)
- [DataLoadSpec](#dataloadspec)
- [DataMigrateSpec](#datamigratespec)
- [EFCCompTemplateSpec](#efccomptemplatespec)
- [EFCFuseSpec](#efcfusespec)
- [EFCRuntimeSpec](#efcruntimespec)
- [JindoCompTemplateSpec](#jindocomptemplatespec)
- [JindoFuseSpec](#jindofusespec)
- [JindoRuntimeSpec](#jindoruntimespec)
- [JuiceFSCompTemplateSpec](#juicefscomptemplatespec)
- [JuiceFSFuseSpec](#juicefsfusespec)
- [JuiceFSRuntimeSpec](#juicefsruntimespec)
- [MasterSpec](#masterspec)
- [Metadata](#metadata)
- [Processor](#processor)
- [RuntimeComponentCommonSpec](#runtimecomponentcommonspec)
- [ThinFuseSpec](#thinfusespec)
- [VineyardClientSocketSpec](#vineyardclientsocketspec)
- [VineyardCompTemplateSpec](#vineyardcomptemplatespec)
- [VineyardRuntimeSpec](#vineyardruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `labels` _object (keys:string, values:string)_ | Labels are labels of pod specification |  |  |
| `annotations` _object (keys:string, values:string)_ | Annotations are annotations of pod specification |  |  |


#### Policy

_Underlying type:_ _string_





_Appears in:_
- [DataLoadSpec](#dataloadspec)
- [DataMigrateSpec](#datamigratespec)

| Field | Description |
| --- | --- |
| `Once` | Once run data migrate once, default policy is Once<br /> |
| `Cron` | Cron run data migrate by cron<br /> |
| `OnEvent` | OnEvent run data migrate when event occurs<br /> |


#### Prefer



Prefer defines the label key and weight for generating a PreferredSchedulingTerm.



_Appears in:_
- [AffinityStrategy](#affinitystrategy)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ |  |  |  |
| `weight` _integer_ |  |  |  |


#### ProcessMemoryMediumSource



ProcessMemoryMediumSource describes process memory as a storage medium.
When specified, cache data will be stored in the process's memory space,
and the quota will be added to the container's resource requests and limits.



_Appears in:_
- [MediumSource](#mediumsource)



#### Processor



Processor defines the actual processor for DataProcess. Processor can be either of a Job or a Shell script.



_Appears in:_
- [DataProcessSpec](#dataprocessspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `serviceAccountName` _string_ | ServiceAccountName defiens the serviceAccountName of the container |  | Optional: \{\} <br /> |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata defines labels and annotations on the processor pod. |  | Optional: \{\} <br /> |
| `job` _[JobProcessor](#jobprocessor)_ | Job represents a processor which runs DataProcess as a job. |  | Optional: \{\} <br /> |
| `script` _[ScriptProcessor](#scriptprocessor)_ | Shell represents a processor which executes shell script |  |  |


#### Require



Require defines the label key for generating a NodeSelectorTerm.



_Appears in:_
- [AffinityStrategy](#affinitystrategy)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ |  |  |  |


#### Runtime



Runtime describes a runtime to be used to support dataset



_Appears in:_
- [DatasetSpec](#datasetspec)
- [DatasetStatus](#datasetstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name of the runtime object |  |  |
| `namespace` _string_ | Namespace of the runtime object |  |  |
| `category` _[Category](#category)_ | Category the runtime object belongs to (e.g. Accelerate) |  |  |
| `type` _string_ | Runtime object's type (e.g. Alluxio) |  |  |
| `masterReplicas` _integer_ | Runtime master replicas |  |  |


#### RuntimeComponentCommonSpec



RuntimeComponentCommonSpec describes the common configuration for CacheRuntime components



_Appears in:_
- [CacheRuntimeClientSpec](#cacheruntimeclientspec)
- [CacheRuntimeMasterSpec](#cacheruntimemasterspec)
- [CacheRuntimeWorkerSpec](#cacheruntimeworkerspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `disabled` _boolean_ | Disabled indicates whether the component should be disabled.<br />If set to true, the component will not be created. |  | Optional: \{\} <br /> |
| `runtimeVersion` _[VersionSpec](#versionspec)_ | RuntimeVersion is the version information that instructs Fluid to orchestrate a particular version of the runtime. |  | Optional: \{\} <br /> |
| `options` _object (keys:string, values:string)_ | Options is a set of key-value pairs that provide additional configuration for the cache system. |  | Optional: \{\} <br /> |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata contains labels and annotations that will be propagated to the component's pods. |  | Optional: \{\} <br /> |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#resourcerequirements-v1-core)_ | Resources describes the compute resource requirements.<br />More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |  | Optional: \{\} <br /> |
| `env` _[EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#envvar-v1-core) array_ | Env is a list of environment variables to set in the container. |  | Optional: \{\} <br /> |
| `volumeMounts` _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#volumemount-v1-core) array_ | VolumeMounts are the volumes to mount into the container's filesystem.<br />Cannot be updated. |  | Optional: \{\} <br /> |
| `args` _string array_ | Args are arguments to the entrypoint.<br />The container image's CMD is used if this is not provided. |  | Optional: \{\} <br /> |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector which must be true for the pod to fit on a node.<br />Selector which must match a node's labels for the pod to be scheduled on that node.<br />More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/ |  | Optional: \{\} <br /> |
| `tolerations` _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#toleration-v1-core) array_ | Tolerations are the pod's tolerations.<br />If specified, the pod's tolerations.<br />More info: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/ |  | Optional: \{\} <br /> |


#### RuntimeComponentDefinition



RuntimeComponentDefinition defines the configuration for a CacheRuntime component



_Appears in:_
- [RuntimeTopology](#runtimetopology)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `workloadType` _[TypeMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#typemeta-v1-meta)_ | WorkloadType is the default workload type of the component |  | Optional: \{\} <br /> |
| `options` _object (keys:string, values:string)_ | Options is a set of key-value pairs that provide additional configuration for the component |  | Optional: \{\} <br /> |
| `template` _[PodTemplateSpec](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#podtemplatespec-v1-core)_ | Template describes the pods that will be created.<br />The template follows the standard PodTemplateSpec from Kubernetes core. |  | Optional: \{\} <br /> |
| `service` _[RuntimeComponentService](#runtimecomponentservice)_ | Service is the service configuration for the component |  | Optional: \{\} <br /> |
| `dependencies` _[RuntimeComponentDependencies](#runtimecomponentdependencies)_ | Dependencies specifies the dependencies required by the component |  | Optional: \{\} <br /> |


#### RuntimeComponentDependencies



RuntimeComponentDependencies defines the dependencies required by a CacheRuntime component



_Appears in:_
- [RuntimeComponentDefinition](#runtimecomponentdefinition)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `encryptOption` _[EncryptOptionComponentDependency](#encryptoptioncomponentdependency)_ | EncryptOption is the configuration for encrypt option secret mount |  | Optional: \{\} <br /> |
| `extraResources` _[ExtraResourcesComponentDependency](#extraresourcescomponentdependency)_ | ExtraResources specifies the usage of extra resources such as ConfigMaps |  | Optional: \{\} <br /> |


#### RuntimeComponentService



RuntimeComponentService describes the service configuration for a CacheRuntime component



_Appears in:_
- [RuntimeComponentDefinition](#runtimecomponentdefinition)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `headless` _[HeadlessRuntimeComponentService](#headlessruntimecomponentservice)_ | Headless enables a headless service for the component.<br />A headless service does not allocate a cluster IP and allows direct pod-to-pod communication. |  | Optional: \{\} <br /> |


#### RuntimeComponentStatus



RuntimeComponentStatus describes the observed state of a runtime component.
It follows the standard Kubernetes pattern for tracking workload status.



_Appears in:_
- [CacheRuntimeStatus](#cacheruntimestatus)
- [RuntimeComponentStatusCollection](#runtimecomponentstatuscollection)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `phase` _[RuntimePhase](#runtimephase)_ | Phase is the current lifecycle phase of the component.<br />Known phases include: "Pending", "Running", "Failed", etc. |  | Required: \{\} <br /> |
| `reason` _string_ | Reason is a brief, machine-readable string that gives the reason for the current phase.<br />This is useful for understanding why a component is in a particular state. |  | Optional: \{\} <br /> |
| `message` _string_ | Message is a human-readable message indicating details about the component's status. |  | Optional: \{\} <br /> |
| `readyReplicas` _integer_ | ReadyReplicas is the number of pods with a Ready condition. |  | Optional: \{\} <br /> |
| `currentReplicas` _integer_ | CurrentReplicas is the current number of replicas running for the component |  | Optional: \{\} <br /> |
| `desiredReplicas` _integer_ | DesiredReplicas is the desired number of replicas as specified in the spec. |  | Optional: \{\} <br /> |
| `unavailableReplicas` _integer_ | UnavailableReplicas is the number of pods that are not available.<br />A pod is considered unavailable if it is not ready or has been terminated. |  | Optional: \{\} <br /> |
| `availableReplicas` _integer_ | AvailableReplicas is the number of pods that are available and ready to serve requests.<br />This count includes only pods that have passed readiness checks. |  | Optional: \{\} <br /> |


#### RuntimeComponentStatusCollection



RuntimeComponentStatusCollection describes the status of all runtime components.
It provides a unified view of master, worker, and client component states.



_Appears in:_
- [CacheRuntimeStatus](#cacheruntimestatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `master` _[RuntimeComponentStatus](#runtimecomponentstatus)_ | Master is the observed state of the master component. |  | Optional: \{\} <br /> |
| `worker` _[RuntimeComponentStatus](#runtimecomponentstatus)_ | Worker is the observed state of the worker component. |  | Optional: \{\} <br /> |
| `client` _[RuntimeComponentStatus](#runtimecomponentstatus)_ | Client is the observed state of the client (FUSE) component. |  | Optional: \{\} <br /> |


#### RuntimeCondition



Condition describes the state of the cache at a certain point.



_Appears in:_
- [CacheRuntimeStatus](#cacheruntimestatus)
- [RuntimeStatus](#runtimestatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[RuntimeConditionType](#runtimeconditiontype)_ | Type of cache condition. |  |  |
| `reason` _string_ | The reason for the condition's last transition. |  |  |
| `message` _string_ | A human readable message indicating details about the transition. |  |  |
| `lastProbeTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#time-v1-meta)_ | The last time this condition was updated. |  |  |
| `lastTransitionTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#time-v1-meta)_ | Last time the condition transitioned from one status to another. |  |  |


#### RuntimeConditionType

_Underlying type:_ _string_

RuntimeConditionType indicates valid conditions type of a runtime



_Appears in:_
- [RuntimeCondition](#runtimecondition)

| Field | Description |
| --- | --- |
| `MasterInitialized` | RuntimeMasterInitialized means the master of runtime is initialized<br /> |
| `MasterReady` | RuntimeMasterReady means the master of runtime is ready<br /> |
| `WorkersInitialized` | RuntimeWorkersInitialized means the workers of runtime are initialized<br /> |
| `WorkersReady` | RuntimeWorkersReady means the workers of runtime are ready<br /> |
| `WorkersScaledIn` | RuntimeWorkerScaledIn means the workers of runtime just scaled in<br /> |
| `WorkersScaledOut` | RuntimeWorkerScaledIn means the workers of runtime just scaled out<br /> |
| `FusesInitialized` | RuntimeFusesInitialized means the fuses of runtime are initialized<br /> |
| `FusesReady` | RuntimeFusesReady means the fuses of runtime are ready<br /> |
| `FusesScaledIn` | RuntimeFusesScaledIn means the fuses of runtime just scaled in<br /> |
| `FusesScaledOut` | RuntimeFusesScaledOut means the fuses of runtime just scaled out<br /> |


#### RuntimeExtraResources



RuntimeExtraResources defines the extra resources for CacheRuntime



_Appears in:_
- [CacheRuntimeClass](#cacheruntimeclass)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `configMaps` _[ConfigMapRuntimeExtraResource](#configmapruntimeextraresource) array_ | ConfigMaps is a list of ConfigMaps that will be created in the runtime's namespace.<br />These ConfigMaps can be referenced and mounted by runtime components. |  | Optional: \{\} <br /> |


#### RuntimeManagement



RuntimeManagement defines suggestions for runtime controllers to manage the runtime



_Appears in:_
- [AlluxioRuntimeSpec](#alluxioruntimespec)
- [JuiceFSRuntimeSpec](#juicefsruntimespec)
- [ThinRuntimeSpec](#thinruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `cleanCachePolicy` _[CleanCachePolicy](#cleancachepolicy)_ | CleanCachePolicy defines the policy of cleaning cache when shutting down the runtime |  | Optional: \{\} <br /> |
| `metadataSyncPolicy` _[MetadataSyncPolicy](#metadatasyncpolicy)_ | MetadataSyncPolicy defines the policy of syncing metadata when setting up the runtime. If not set, |  | Optional: \{\} <br /> |


#### RuntimePhase

_Underlying type:_ _string_





_Appears in:_
- [RuntimeComponentStatus](#runtimecomponentstatus)
- [RuntimeStatus](#runtimestatus)

| Field | Description |
| --- | --- |
| `` |  |
| `NotReady` |  |
| `PartialReady` |  |
| `Ready` |  |




#### RuntimeTieredStore



RuntimeTieredStore describes the tiered storage configuration for cache runtime



_Appears in:_
- [CacheRuntimeClientSpec](#cacheruntimeclientspec)
- [CacheRuntimeWorkerSpec](#cacheruntimeworkerspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `levels` _[RuntimeTieredStoreLevel](#runtimetieredstorelevel) array_ | Levels is the list of cache storage tiers from highest priority to lowest.<br />Each tier can use different storage media (e.g., memory, SSD, HDD). |  | Optional: \{\} <br /> |


#### RuntimeTieredStoreLevel



RuntimeTieredStoreLevel describes the configuration for a single tier in the tiered storage



_Appears in:_
- [RuntimeTieredStore](#runtimetieredstore)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `medium` _[MediumSource](#mediumsource)_ | Medium describes the storage medium type for this tier.<br />Supported types include process memory and various volume types. |  | Optional: \{\} <br /> |
| `path` _string array_ | Path is a list of file paths to be used for the cache tier.<br />Multiple paths can be specified to distribute cache across different mount points.<br />For example: ["/mnt/cache1", "/mnt/cache2"]. |  | MinItems: 1 <br />Optional: \{\} <br /> |
| `quota` _[Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#quantity-resource-api) array_ | Quota is a list of storage quotas for each path in the tier.<br />The length of Quota should match the length of Path.<br />Each quota corresponds to the path at the same index.<br />For example: ["100Gi", "50Gi"] allocates 100GiB to the first path and 50GiB to the second path. |  | Optional: \{\} <br /> |
| `high` _string_ | High is the ratio of high watermark of the tier (e.g., "0.9").<br />When cache usage exceeds this ratio, eviction will be triggered. |  | Optional: \{\} <br /> |
| `low` _string_ | Low is the ratio of low watermark of the tier (e.g., "0.7").<br />Eviction will continue until cache usage falls below this ratio. |  | Optional: \{\} <br /> |


#### RuntimeTopology



RuntimeTopology defines the topology structure of CacheRuntime components



_Appears in:_
- [CacheRuntimeClass](#cacheruntimeclass)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `master` _[RuntimeComponentDefinition](#runtimecomponentdefinition)_ | Master is the configuration for master component |  | Optional: \{\} <br /> |
| `worker` _[RuntimeComponentDefinition](#runtimecomponentdefinition)_ | Worker is the configuration for worker component |  | Optional: \{\} <br /> |
| `client` _[RuntimeComponentDefinition](#runtimecomponentdefinition)_ | Client is the configuration for client component |  | Optional: \{\} <br /> |


#### ScriptProcessor







_Appears in:_
- [Processor](#processor)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `image` _string_ | Image (e.g. alluxio/alluxio) |  |  |
| `imageTag` _string_ | Image tag (e.g. 2.3.0-SNAPSHOT) |  |  |
| `imagePullPolicy` _string_ | One of the three policies: `Always`, `IfNotPresent`, `Never` |  |  |
| `restartPolicy` _[RestartPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#restartpolicy-v1-core)_ | RestartPolicy specifies the processor job's restart policy. Only "Never", "OnFailure" is allowed. | Never | Enum: [Never OnFailure] <br />Optional: \{\} <br /> |
| `command` _string array_ | Entrypoint command for ScriptProcessor. |  | Optional: \{\} <br /> |
| `source` _string_ | Script source for ScriptProcessor |  | Required: \{\} <br /> |
| `env` _[EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#envvar-v1-core) array_ | List of environment variables to set in the container. |  | Optional: \{\} <br /> |
| `volumeMounts` _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#volumemount-v1-core) array_ | Pod volumes to mount into the container's filesystem. |  | Optional: \{\} <br /> |
| `volumes` _[Volume](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#volume-v1-core) array_ | List of volumes that can be mounted by containers belonging to the pod. |  | Optional: \{\} <br /> |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#resourcerequirements-v1-core)_ | Resources that will be requested by the DataProcess job. <br> |  | Optional: \{\} <br /> |


#### SecretKeySelector







_Appears in:_
- [EncryptOptionSource](#encryptoptionsource)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | The name of required secret |  | Required: \{\} <br /> |
| `key` _string_ | The required key in the secret |  | Optional: \{\} <br /> |


#### TargetDataset



TargetDataset defines the target dataset of the DataLoad



_Appears in:_
- [DataLoadSpec](#dataloadspec)
- [TargetDatasetWithMountPath](#targetdatasetwithmountpath)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name defines name of the target dataset |  |  |
| `namespace` _string_ | Namespace defines namespace of the target dataset |  |  |


#### TargetDatasetWithMountPath



TargetDataset defines which dataset will be processed by DataProcess.
Under the hood, the dataset's pvc will be mounted to the given mountPath of the DataProcess's containers.



_Appears in:_
- [DataProcessSpec](#dataprocessspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name defines name of the target dataset |  |  |
| `namespace` _string_ | Namespace defines namespace of the target dataset |  |  |
| `mountPath` _string_ | MountPath defines where the Dataset should be mounted in DataProcess's containers. |  | Required: \{\} <br /> |
| `subPath` _string_ | SubPath defines subpath of the target dataset to mount. |  | Optional: \{\} <br /> |


#### TargetPath



TargetPath defines the target path of the DataLoad



_Appears in:_
- [DataLoadSpec](#dataloadspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `path` _string_ | Path defines path to be load |  |  |
| `replicas` _integer_ | Replicas defines how many replicas will be loaded |  |  |


#### ThinCompTemplateSpec



ThinCompTemplateSpec is a description of the thinRuntime components



_Appears in:_
- [ThinRuntimeProfileSpec](#thinruntimeprofilespec)
- [ThinRuntimeSpec](#thinruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `image` _string_ | Image for thinRuntime fuse |  |  |
| `imageTag` _string_ | Image for thinRuntime fuse |  |  |
| `imagePullPolicy` _string_ | One of the three policies: `Always`, `IfNotPresent`, `Never` |  |  |
| `imagePullSecrets` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#localobjectreference-v1-core) array_ | ImagePullSecrets that will be used to pull images |  | Optional: \{\} <br /> |
| `replicas` _integer_ | Replicas is the desired number of replicas of the given template.<br />If unspecified, defaults to 1.<br />replicas is the min replicas of dataset in the cluster |  | Minimum: 1 <br />Optional: \{\} <br /> |
| `ports` _[ContainerPort](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#containerport-v1-core) array_ | Ports used thinRuntime |  | Optional: \{\} <br /> |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#resourcerequirements-v1-core)_ | Resources that will be requested by thinRuntime component. |  | Optional: \{\} <br /> |
| `env` _[EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#envvar-v1-core) array_ | Environment variables that will be used by thinRuntime component. |  |  |
| `enabled` _boolean_ | Enabled or Disabled for the components. |  | Optional: \{\} <br /> |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector |  | Optional: \{\} <br /> |
| `volumeMounts` _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#volumemount-v1-core) array_ | VolumeMounts specifies the volumes listed in ".spec.volumes" to mount into runtime component's filesystem. |  | Optional: \{\} <br /> |
| `livenessProbe` _[Probe](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#probe-v1-core)_ | livenessProbe of thin fuse pod |  | Optional: \{\} <br /> |
| `readinessProbe` _[Probe](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#probe-v1-core)_ | readinessProbe of thin fuse pod |  | Optional: \{\} <br /> |
| `networkMode` _[NetworkMode](#networkmode)_ | Whether to use hostnetwork or not |  | Enum: [HostNetwork  ContainerNetwork] <br />Optional: \{\} <br /> |


#### ThinFuseSpec







_Appears in:_
- [ThinRuntimeProfileSpec](#thinruntimeprofilespec)
- [ThinRuntimeSpec](#thinruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `image` _string_ | Image for thinRuntime fuse |  |  |
| `imageTag` _string_ | Image for thinRuntime fuse |  |  |
| `imagePullPolicy` _string_ | One of the three policies: `Always`, `IfNotPresent`, `Never` |  |  |
| `imagePullSecrets` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#localobjectreference-v1-core) array_ | ImagePullSecrets that will be used to pull images |  | Optional: \{\} <br /> |
| `ports` _[ContainerPort](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#containerport-v1-core) array_ | Ports used thinRuntime |  | Optional: \{\} <br /> |
| `env` _[EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#envvar-v1-core) array_ | Environment variables that will be used by thinRuntime Fuse |  |  |
| `command` _string array_ | Command that will be passed to thinRuntime Fuse |  |  |
| `args` _string array_ | Arguments that will be passed to thinRuntime Fuse |  |  |
| `options` _object (keys:string, values:string)_ | Options configurable options of FUSE client, performance parameters usually.<br />will be merged with Dataset.spec.mounts.options into fuse pod. |  |  |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#resourcerequirements-v1-core)_ | Resources that will be requested by thinRuntime Fuse. |  |  |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector which must be true for the fuse client to fit on a node,<br />this option only effect when global is enabled |  | Optional: \{\} <br /> |
| `cleanPolicy` _[FuseCleanPolicy](#fusecleanpolicy)_ | CleanPolicy decides when to clean thinRuntime Fuse pods.<br />Currently Fluid supports two policies: OnDemand and OnRuntimeDeleted<br />OnDemand cleans fuse pod once the fuse pod on some node is not needed<br />OnRuntimeDeleted cleans fuse pod only when the cache runtime is deleted<br />Defaults to OnDemand |  | Optional: \{\} <br /> |
| `networkMode` _[NetworkMode](#networkmode)_ | Whether to use hostnetwork or not |  | Enum: [HostNetwork  ContainerNetwork] <br />Optional: \{\} <br /> |
| `livenessProbe` _[Probe](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#probe-v1-core)_ | livenessProbe of thin fuse pod |  | Optional: \{\} <br /> |
| `readinessProbe` _[Probe](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#probe-v1-core)_ | readinessProbe of thin fuse pod |  | Optional: \{\} <br /> |
| `volumeMounts` _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#volumemount-v1-core) array_ | VolumeMounts specifies the volumes listed in ".spec.volumes" to mount into the thinruntime component's filesystem. |  | Optional: \{\} <br /> |
| `lifecycle` _[Lifecycle](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#lifecycle-v1-core)_ | Lifecycle describes actions that the management system should take in response to container lifecycle events. |  |  |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata defines labels and annotations that will be propagated to ThinRuntime's FUSE pods. |  | Optional: \{\} <br /> |


#### ThinRuntime



ThinRuntime is the Schema for the thinruntimes API





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `data.fluid.io/v1alpha1` | | |
| `kind` _string_ | `ThinRuntime` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[ThinRuntimeSpec](#thinruntimespec)_ |  |  |  |


#### ThinRuntimeProfile



ThinRuntimeProfile is the Schema for the ThinRuntimeProfiles API





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `data.fluid.io/v1alpha1` | | |
| `kind` _string_ | `ThinRuntimeProfile` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[ThinRuntimeProfileSpec](#thinruntimeprofilespec)_ |  |  |  |


#### ThinRuntimeProfileSpec



ThinRuntimeProfileSpec defines the desired state of ThinRuntimeProfile



_Appears in:_
- [ThinRuntimeProfile](#thinruntimeprofile)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `fileSystemType` _string_ | file system of thinRuntime |  | Required: \{\} <br /> |
| `imagePullSecrets` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#localobjectreference-v1-core) array_ | ImagePullSecrets that will be used to pull images |  | Optional: \{\} <br /> |
| `worker` _[ThinCompTemplateSpec](#thincomptemplatespec)_ | The component spec of worker |  |  |
| `fuse` _[ThinFuseSpec](#thinfusespec)_ | The component spec of thinRuntime |  |  |
| `volumes` _[Volume](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#volume-v1-core) array_ | Volumes is the list of Kubernetes volumes that can be mounted by runtime components and/or fuses. |  | Optional: \{\} <br /> |
| `nodePublishSecretPolicy` _[NodePublishSecretPolicy](#nodepublishsecretpolicy)_ | NodePublishSecretPolicy describes the policy to decide which to do with node publish secret when mounting an existing persistent volume. | MountNodePublishSecretIfExists | Enum: [NotMountNodePublishSecret MountNodePublishSecretIfExists CopyNodePublishSecretAndMountIfNotExists] <br /> |




#### ThinRuntimeSpec



ThinRuntimeSpec defines the desired state of ThinRuntime



_Appears in:_
- [ThinRuntime](#thinruntime)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `profileName` _string_ | The specific runtime profile name, empty value is used for handling datasets which mount another dataset |  |  |
| `worker` _[ThinCompTemplateSpec](#thincomptemplatespec)_ | The component spec of worker |  |  |
| `fuse` _[ThinFuseSpec](#thinfusespec)_ | The component spec of thinRuntime |  |  |
| `tieredstore` _[TieredStore](#tieredstore)_ | Tiered storage |  |  |
| `replicas` _integer_ | The replicas of the worker, need to be specified |  |  |
| `runAs` _[User](#user)_ | Manage the user to run Runtime |  |  |
| `disablePrometheus` _boolean_ | Disable monitoring for Runtime<br />Prometheus is enabled by default |  | Optional: \{\} <br /> |
| `volumes` _[Volume](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#volume-v1-core) array_ | Volumes is the list of Kubernetes volumes that can be mounted by runtime components and/or fuses. |  | Optional: \{\} <br /> |
| `management` _[RuntimeManagement](#runtimemanagement)_ | RuntimeManagement defines policies when managing the runtime |  | Optional: \{\} <br /> |
| `imagePullSecrets` _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#localobjectreference-v1-core) array_ | ImagePullSecrets that will be used to pull images |  | Optional: \{\} <br /> |


#### TieredStore



TieredStore is a description of the tiered store



_Appears in:_
- [AlluxioRuntimeSpec](#alluxioruntimespec)
- [EFCRuntimeSpec](#efcruntimespec)
- [GooseFSRuntimeSpec](#goosefsruntimespec)
- [JindoRuntimeSpec](#jindoruntimespec)
- [JuiceFSRuntimeSpec](#juicefsruntimespec)
- [ThinRuntimeSpec](#thinruntimespec)
- [VineyardRuntimeSpec](#vineyardruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `levels` _[Level](#level) array_ | configurations for multiple tiers |  |  |


#### User



User explains the user and group to run a Container



_Appears in:_
- [AlluxioRuntimeSpec](#alluxioruntimespec)
- [DataBackupSpec](#databackupspec)
- [DatasetSpec](#datasetspec)
- [GooseFSRuntimeSpec](#goosefsruntimespec)
- [JindoRuntimeSpec](#jindoruntimespec)
- [JuiceFSRuntimeSpec](#juicefsruntimespec)
- [ThinRuntimeSpec](#thinruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `uid` _integer_ | The uid to run the alluxio runtime |  |  |
| `gid` _integer_ | The gid to run the alluxio runtime |  |  |
| `user` _string_ | The user name to run the alluxio runtime |  |  |
| `group` _string_ | The group name to run the alluxio runtime |  |  |


#### VersionSpec



VersionSpec represents the settings for the  version that fluid is orchestrating.



_Appears in:_
- [AlluxioRuntimeSpec](#alluxioruntimespec)
- [CacheRuntimeClientSpec](#cacheruntimeclientspec)
- [CacheRuntimeMasterSpec](#cacheruntimemasterspec)
- [CacheRuntimeWorkerSpec](#cacheruntimeworkerspec)
- [DataMigrateSpec](#datamigratespec)
- [EFCCompTemplateSpec](#efccomptemplatespec)
- [EFCFuseSpec](#efcfusespec)
- [GooseFSRuntimeSpec](#goosefsruntimespec)
- [InitFuseSpec](#initfusespec)
- [JindoRuntimeSpec](#jindoruntimespec)
- [JuiceFSRuntimeSpec](#juicefsruntimespec)
- [RuntimeComponentCommonSpec](#runtimecomponentcommonspec)
- [ScriptProcessor](#scriptprocessor)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `image` _string_ | Image (e.g. alluxio/alluxio) |  |  |
| `imageTag` _string_ | Image tag (e.g. 2.3.0-SNAPSHOT) |  |  |
| `imagePullPolicy` _string_ | One of the three policies: `Always`, `IfNotPresent`, `Never` |  |  |


#### VineyardClientSocketSpec



VineyardClientSocketSpec holds the configurations for vineyard client socket



_Appears in:_
- [VineyardRuntimeSpec](#vineyardruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `image` _string_ | Image for Vineyard Fuse<br />Default is `registry.aliyuncs.com/vineyard/vineyard-fluid-fuse` |  | Optional: \{\} <br /> |
| `imageTag` _string_ | Image Tag for Vineyard Fuse<br />Default is `v0.22.2` |  | Optional: \{\} <br /> |
| `imagePullPolicy` _string_ | Image pull policy for Vineyard Fuse<br />Default is `IfNotPresent`<br />Available values are `Always`, `IfNotPresent`, `Never` |  | Optional: \{\} <br /> |
| `env` _object (keys:string, values:string)_ | Environment variables that will be used by Vineyard Fuse.<br />Default is not set. |  | Optional: \{\} <br /> |
| `cleanPolicy` _[FuseCleanPolicy](#fusecleanpolicy)_ | CleanPolicy decides when to clean Vineyard Fuse pods.<br />Currently Fluid supports two policies: OnDemand and OnRuntimeDeleted<br />OnDemand cleans fuse pod once th fuse pod on some node is not needed<br />OnRuntimeDeleted cleans fuse pod only when the cache runtime is deleted<br />Defaults to OnRuntimeDeleted |  | Optional: \{\} <br /> |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#resourcerequirements-v1-core)_ | Resources contains the resource requirements and limits for the Vineyard Fuse.<br />Default is not set. |  | Optional: \{\} <br /> |
| `networkMode` _[NetworkMode](#networkmode)_ | Whether to use hostnetwork or not<br />Default is HostNetwork |  | Enum: [HostNetwork  ContainerNetwork] <br />Optional: \{\} <br /> |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata defines labels and annotations that will be propagated to Vineyard's pods. |  | Optional: \{\} <br /> |
| `options` _object (keys:string, values:string)_ | Options for configuring vineyardd parameters.<br />Supported options are as follows.<br />  reserve_memory: (Bool) Whether to reserving enough physical memory pages for vineyardd.<br />                  Default is true.<br />  allocator: (String) The allocator used by vineyardd, could be "dlmalloc" or "mimalloc".<br />             Default is "dlmalloc".<br />  compression: (Bool) Compress before migration or spilling.<br />               Default is true.<br />  coredump: (Bool) Enable coredump core dump when been aborted.<br />            Default is false.<br />  meta_timeout: (Int) Timeout period before waiting the metadata service to be ready, in seconds<br />				   Default is 60.<br />  etcd_endpoint: (String) The endpoint of etcd.<br />                 Default is same as the etcd endpoint of vineyard worker.<br />  etcd_prefix: (String) Metadata path prefix in etcd.<br />               Default is "/vineyard".<br />  size: (String) shared memory size for vineyardd.<br />                 1024M, 1024000, 1G, or 1Gi.<br />                 Default is "0", which means no cache.<br />                 When the size is not set to "0", it should be greater than the 2048 bytes(2K).<br />  spill_path: (String) Path to spill temporary files, if not set, spilling will be disabled.<br />              Default is "".<br />  spill_lower_rate: (Double) The lower rate of memory usage to trigger spilling.<br />					   Default is 0.3.<br />  spill_upper_rate: (Double) The upper rate of memory usage to stop spilling.<br />					   Default is 0.8.<br />Default is as follows.<br />fuse:<br />  options:<br />    size: "0"<br />    etcd_endpoint: "http://\{\{Name\}\}-master-0.\{\{Name\}\}-master.\{\{Namespace\}\}:\{\{EtcdClientPort\}\}"<br />	   etcd_prefix: "/vineyard" |  | Optional: \{\} <br /> |


#### VineyardCompTemplateSpec



VineyardCompTemplateSpec is the common configurations for vineyard components including Master and Worker.



_Appears in:_
- [MasterSpec](#masterspec)
- [VineyardRuntimeSpec](#vineyardruntimespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `replicas` _integer_ | The replicas of Vineyard component.<br />If not specified, defaults to 1.<br />For worker, the replicas should not be greater than the number of nodes in the cluster |  | Minimum: 1 <br />Optional: \{\} <br /> |
| `image` _string_ | The image of Vineyard component.<br />For Master, the default image is `registry.aliyuncs.com/vineyard/vineyardd`<br />For Worker, the default image is `registry.aliyuncs.com/vineyard/vineyardd`<br />The default container registry is `docker.io`, you can change it by setting the image field |  | Optional: \{\} <br /> |
| `imageTag` _string_ | The image tag of Vineyard component.<br />For Master, the default image tag is `v0.22.2`.<br />For Worker, the default image tag is `v0.22.2`. |  | Optional: \{\} <br /> |
| `imagePullPolicy` _string_ | The image pull policy of Vineyard component.<br />Default is `IfNotPresent`. |  | Optional: \{\} <br /> |
| `nodeSelector` _object (keys:string, values:string)_ | NodeSelector is a selector to choose which nodes to launch the Vineyard component.<br />E,g. \{"disktype": "ssd"\} |  | Optional: \{\} <br /> |
| `ports` _object (keys:string, values:integer)_ | Ports used by Vineyard component.<br />For Master, the default client port is 2379 and peer port is 2380.<br />For Worker, the default rpc port is 9600 and the default exporter port is 9144. |  | Optional: \{\} <br /> |
| `env` _object (keys:string, values:string)_ | Environment variables that will be used by Vineyard component.<br />For Master, refer to <a href="https://etcd.io/docs/v3.5/op-guide/configuration/">Etcd Configuration</a> for more info<br />Default is not set. |  | Optional: \{\} <br /> |
| `options` _object (keys:string, values:string)_ | Configurable options for Vineyard component.<br />For Master, there is no configurable options.<br />For Worker, support the following options.<br />  vineyardd.reserve.memory: (Bool) where to reserve memory for vineyardd<br />                            If set to true, the memory quota will be counted to the vineyardd rather than the application.<br />  etcd.prefix: (String) the prefix of etcd key for vineyard objects<br />  wait.etcd.timeout: (String) the timeout period before waiting the etcd to be ready, in seconds<br />  Default value is as follows.<br />    vineyardd.reserve.memory: "true"<br />    etcd.prefix: "/vineyard"<br />    wait.etcd.timeout: "120" |  | Optional: \{\} <br /> |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#resourcerequirements-v1-core)_ | Resources contains the resource requirements and limits for the Vineyard component.<br />Default is not set.<br />For Worker, when the options contains vineyardd.reserve.memory=true,<br />the resources.request.memory for worker should be greater than tieredstore.levels[0].quota(aka vineyardd shared memory) |  | Optional: \{\} <br /> |
| `volumeMounts` _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#volumemount-v1-core) array_ | VolumeMounts specifies the volumes listed in ".spec.volumes" to mount into the vineyard runtime component's filesystem.<br />It is useful for specifying a persistent storage.<br />Default is not set. |  | Optional: \{\} <br /> |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata defines labels and annotations that will be propagated to Vineyard's pods including Master and Worker.<br />Default is not set. |  | Optional: \{\} <br /> |
| `networkMode` _[NetworkMode](#networkmode)_ | Whether to use hostnetwork or not<br />Default is HostNetwork |  | Enum: [HostNetwork  ContainerNetwork] <br />Optional: \{\} <br /> |


#### VineyardRuntime



VineyardRuntime is the Schema for the VineyardRuntimes API





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `data.fluid.io/v1alpha1` | | |
| `kind` _string_ | `VineyardRuntime` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[VineyardRuntimeSpec](#vineyardruntimespec)_ |  |  |  |


#### VineyardRuntimeSpec



VineyardRuntimeSpec defines the desired state of VineyardRuntime



_Appears in:_
- [VineyardRuntime](#vineyardruntime)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `master` _[MasterSpec](#masterspec)_ | Master holds the configurations for Vineyard Master component<br />Represents the Etcd component in Vineyard |  | Optional: \{\} <br /> |
| `worker` _[VineyardCompTemplateSpec](#vineyardcomptemplatespec)_ | Worker holds the configurations for Vineyard Worker component<br />Represents the Vineyardd component in Vineyard |  | Optional: \{\} <br /> |
| `replicas` _integer_ | The replicas of the worker, need to be specified<br />If worker.replicas and the field are both specified, the field will be respected |  |  |
| `fuse` _[VineyardClientSocketSpec](#vineyardclientsocketspec)_ | Fuse holds the configurations for Vineyard client socket.<br />Note that the "Fuse" here is kept just for API consistency, VineyardRuntime mount a socket file instead of a FUSE filesystem to make data cache available.<br />Applications can connect to the vineyard runtime components through IPC or RPC.<br />IPC is the default way to connect to vineyard runtime components, which is more efficient than RPC.<br />If the socket file is not mounted, the connection will fall back to RPC. |  | Optional: \{\} <br /> |
| `tieredstore` _[TieredStore](#tieredstore)_ | Tiered storage used by vineyardd<br />The MediumType can only be `MEM` and `SSD`<br />`MEM` actually represents the shared memory of vineyardd.<br />`SSD` represents the external storage of vineyardd.<br />Default is as follows.<br />  tieredstore:<br />    levels:<br />    - level: 0<br />      mediumtype: MEM<br />      quota: 4Gi<br />Choose hostpath as the external storage of vineyardd.<br />  tieredstore:<br />    levels:<br />	   - level: 0<br />      mediumtype: MEM<br />      quota: 4Gi<br />		 high: "0.8"<br />      low: "0.3"<br />    - level: 1<br />      mediumtype: SSD<br />      quota: 10Gi<br />      volumeType: Hostpath<br />      path: /var/spill-path |  | Optional: \{\} <br /> |
| `disablePrometheus` _boolean_ | Disable monitoring metrics for Vineyard Runtime<br />Default is false |  | Optional: \{\} <br /> |
| `podMetadata` _[PodMetadata](#podmetadata)_ | PodMetadata defines labels and annotations that will be propagated to Vineyard's pods. |  | Optional: \{\} <br /> |
| `volumes` _[Volume](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#volume-v1-core) array_ | Volumes is the list of Kubernetes volumes that can be mounted by the vineyard components (Master and Worker).<br />Default is null. |  | Optional: \{\} <br /> |


#### VolumeMediumSource



VolumeMediumSource describes a Kubernetes volume as a storage medium.
Only one of its members may be specified.



_Appears in:_
- [MediumSource](#mediumsource)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `hostPath` _[HostPathVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#hostpathvolumesource-v1-core)_ | HostPath represents a pre-existing file or directory on the host machine that is directly exposed to the container.<br />More info: https://kubernetes.io/docs/concepts/storage/volumes#hostpath |  | Optional: \{\} <br /> |
| `emptyDir` _[EmptyDirVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#emptydirvolumesource-v1-core)_ | EmptyDir represents a temporary directory that shares a pod's lifetime.<br />More info: https://kubernetes.io/docs/concepts/storage/volumes#emptydir |  | Optional: \{\} <br /> |
| `ephemeral` _[EphemeralVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.35/#ephemeralvolumesource-v1-core)_ | Ephemeral represents a volume that is handled by a cluster storage driver.<br />The volume's lifecycle is tied to the pod that defines it.<br />More info: https://kubernetes.io/docs/concepts/storage/ephemeral-volumes/ |  | Optional: \{\} <br /> |




#### WaitingStatus







_Appears in:_
- [OperationStatus](#operationstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `operationComplete` _boolean_ | OperationComplete indicates if the preceding operation is complete |  |  |


