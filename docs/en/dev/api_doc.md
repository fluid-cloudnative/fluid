# API Reference
<h2 id="data.fluid.io/v1alpha1">data.fluid.io/v1alpha1</h2>
<p>
<p>Package v1alpha1 is the v1alpha1 version of the API.</p>
</p>
Resource Types:
<ul><li>
<a href="#data.fluid.io/v1alpha1.AlluxioRuntime">AlluxioRuntime</a>
</li><li>
<a href="#data.fluid.io/v1alpha1.DataLoad">DataLoad</a>
</li><li>
<a href="#data.fluid.io/v1alpha1.Dataset">Dataset</a>
</li></ul>
<h3 id="data.fluid.io/v1alpha1.AlluxioRuntime">AlluxioRuntime
</h3>
<p>
<p>AlluxioRuntime is the Schema for the alluxioruntimes API</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
data.fluid.io/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>AlluxioRuntime</code></td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.AlluxioRuntimeSpec">
AlluxioRuntimeSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>alluxioVersion</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.VersionSpec">
VersionSpec
</a>
</em>
</td>
<td>
<p>The version information that instructs fluid to orchestrate a particular version of Alluxio.</p>
</td>
</tr>
<tr>
<td>
<code>master</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.AlluxioCompTemplateSpec">
AlluxioCompTemplateSpec
</a>
</em>
</td>
<td>
<p>Desired state for Alluxio master</p>
</td>
</tr>
<tr>
<td>
<code>jobMaster</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.AlluxioCompTemplateSpec">
AlluxioCompTemplateSpec
</a>
</em>
</td>
<td>
<p>Desired state for Alluxio job master</p>
</td>
</tr>
<tr>
<td>
<code>worker</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.AlluxioCompTemplateSpec">
AlluxioCompTemplateSpec
</a>
</em>
</td>
<td>
<p>Desired state for Alluxio worker</p>
</td>
</tr>
<tr>
<td>
<code>jobWorker</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.AlluxioCompTemplateSpec">
AlluxioCompTemplateSpec
</a>
</em>
</td>
<td>
<p>Desired state for Alluxio job Worker</p>
</td>
</tr>
<tr>
<td>
<code>initUsers</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.InitUsersSpec">
InitUsersSpec
</a>
</em>
</td>
<td>
<p>The spec of init users</p>
</td>
</tr>
<tr>
<td>
<code>fuse</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.AlluxioFuseSpec">
AlluxioFuseSpec
</a>
</em>
</td>
<td>
<p>Desired state for Alluxio Fuse</p>
</td>
</tr>
<tr>
<td>
<code>properties</code></br>
<em>
map[string]string
</em>
</td>
<td>
<p>Configurable properties for Alluxio system. <br>
Refer to <a href="https://docs.alluxio.io/os/user/stable/en/reference/Properties-List.html">Alluxio Configuration Properties</a> for more info</p>
</td>
</tr>
<tr>
<td>
<code>jvmOptions</code></br>
<em>
[]string
</em>
</td>
<td>
<p>Options for JVM</p>
</td>
</tr>
<tr>
<td>
<code>tieredstore</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.Tieredstore">
Tieredstore
</a>
</em>
</td>
<td>
<p>Tiered storage used by Alluxio</p>
</td>
</tr>
<tr>
<td>
<code>data</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.Data">
Data
</a>
</em>
</td>
<td>
<p>Management strategies for the dataset to which the runtime is bound</p>
</td>
</tr>
<tr>
<td>
<code>replicas</code></br>
<em>
int32
</em>
</td>
<td>
<p>The replicas of the worker, need to be specified</p>
</td>
</tr>
<tr>
<td>
<code>runAs</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.User">
User
</a>
</em>
</td>
<td>
<p>Manage the user to run Alluxio Runtime</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.RuntimeStatus">
RuntimeStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.DataLoad">DataLoad
</h3>
<p>
<p>DataLoad is the Schema for the dataloads API</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
data.fluid.io/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>DataLoad</code></td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.DataLoadSpec">
DataLoadSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>dataset</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.TargetDataset">
TargetDataset
</a>
</em>
</td>
<td>
<p>Dataset defines the target dataset of the DataLoad</p>
</td>
</tr>
<tr>
<td>
<code>loadMetadata</code></br>
<em>
bool
</em>
</td>
<td>
<p>LoadMetadata specifies if the dataload job should load metadata</p>
</td>
</tr>
<tr>
<td>
<code>target</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.TargetPath">
[]TargetPath
</a>
</em>
</td>
<td>
<p>Target defines target paths that needs to be loaded</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.DataLoadStatus">
DataLoadStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.Dataset">Dataset
</h3>
<p>
<p>Dataset is the Schema for the datasets API</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
data.fluid.io/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>Dataset</code></td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.DatasetSpec">
DatasetSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>mounts</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.Mount">
[]Mount
</a>
</em>
</td>
<td>
<p>Mount Points to be mounted on Alluxio.</p>
</td>
</tr>
<tr>
<td>
<code>owner</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.User">
User
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The owner of the dataset</p>
</td>
</tr>
<tr>
<td>
<code>nodeAffinity</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.CacheableNodeAffinity">
CacheableNodeAffinity
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeAffinity defines constraints that limit what nodes this dataset can be cached to.
This field influences the scheduling of pods that use the cached dataset.</p>
</td>
</tr>
<tr>
<td>
<code>accessModes</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#persistentvolumeaccessmode-v1-core">
[]Kubernetes core/v1.PersistentVolumeAccessMode
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AccessModes contains all ways the volume backing the PVC can be mounted</p>
</td>
</tr>
<tr>
<td>
<code>runtimes</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.Runtime">
[]Runtime
</a>
</em>
</td>
<td>
<p>Runtimes for supporting dataset (e.g. AlluxioRuntime)</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.DatasetStatus">
DatasetStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.AlluxioCompTemplateSpec">AlluxioCompTemplateSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.AlluxioRuntimeSpec">AlluxioRuntimeSpec</a>)
</p>
<p>
<p>AlluxioCompTemplateSpec is a description of the Alluxio commponents</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>replicas</code></br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Replicas is the desired number of replicas of the given template.
If unspecified, defaults to 1.
replicas is the min replicas of dataset in the cluster</p>
</td>
</tr>
<tr>
<td>
<code>jvmOptions</code></br>
<em>
[]string
</em>
</td>
<td>
<p>Options for JVM</p>
</td>
</tr>
<tr>
<td>
<code>properties</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configurable properties for the Alluxio component. <br>
Refer to <a href="https://docs.alluxio.io/os/user/stable/en/reference/Properties-List.html">Alluxio Configuration Properties</a> for more info</p>
</td>
</tr>
<tr>
<td>
<code>ports</code></br>
<em>
map[string]int
</em>
</td>
<td>
<em>(Optional)</em>
<p>Ports used by Alluxio(e.g. rpc: 19998 for master)</p>
</td>
</tr>
<tr>
<td>
<code>resources</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Resources that will be requested by the Alluxio component. <br>
<br>
Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources
already allocated to the pod.</p>
</td>
</tr>
<tr>
<td>
<code>env</code></br>
<em>
map[string]string
</em>
</td>
<td>
<p>Environment variables that will be used by Alluxio component. <br></p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.AlluxioFuseSpec">AlluxioFuseSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.AlluxioRuntimeSpec">AlluxioRuntimeSpec</a>)
</p>
<p>
<p>AlluxioFuseSpec is a description of the Alluxio Fuse</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>image</code></br>
<em>
string
</em>
</td>
<td>
<p>Image for Alluxio Fuse(e.g. alluxio/alluxio-fuse)</p>
</td>
</tr>
<tr>
<td>
<code>imageTag</code></br>
<em>
string
</em>
</td>
<td>
<p>Image Tag for Alluxio Fuse(e.g. 2.3.0-SNAPSHOT)</p>
</td>
</tr>
<tr>
<td>
<code>imagePullPolicy</code></br>
<em>
string
</em>
</td>
<td>
<p>One of the three policies: <code>Always</code>, <code>IfNotPresent</code>, <code>Never</code></p>
</td>
</tr>
<tr>
<td>
<code>jvmOptions</code></br>
<em>
[]string
</em>
</td>
<td>
<p>Options for JVM</p>
</td>
</tr>
<tr>
<td>
<code>properties</code></br>
<em>
map[string]string
</em>
</td>
<td>
<p>Configurable properties for Alluxio System. <br>
Refer to <a href="https://docs.alluxio.io/os/user/stable/en/reference/Properties-List.html">Alluxio Configuration Properties</a> for more info</p>
</td>
</tr>
<tr>
<td>
<code>env</code></br>
<em>
map[string]string
</em>
</td>
<td>
<p>Environment variables that will be used by Alluxio Fuse</p>
</td>
</tr>
<tr>
<td>
<code>resources</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Resources that will be requested by Alluxio Fuse. <br>
<br>
Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources
already allocated to the pod.</p>
</td>
</tr>
<tr>
<td>
<code>args</code></br>
<em>
[]string
</em>
</td>
<td>
<p>Arguments that will be passed to Alluxio Fuse</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.AlluxioRuntimeRole">AlluxioRuntimeRole
(<code>string</code> alias)</p></h3>
<p>
</p>
<h3 id="data.fluid.io/v1alpha1.AlluxioRuntimeSpec">AlluxioRuntimeSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.AlluxioRuntime">AlluxioRuntime</a>)
</p>
<p>
<p>AlluxioRuntimeSpec defines the desired state of AlluxioRuntime</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>alluxioVersion</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.VersionSpec">
VersionSpec
</a>
</em>
</td>
<td>
<p>The version information that instructs fluid to orchestrate a particular version of Alluxio.</p>
</td>
</tr>
<tr>
<td>
<code>master</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.AlluxioCompTemplateSpec">
AlluxioCompTemplateSpec
</a>
</em>
</td>
<td>
<p>Desired state for Alluxio master</p>
</td>
</tr>
<tr>
<td>
<code>jobMaster</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.AlluxioCompTemplateSpec">
AlluxioCompTemplateSpec
</a>
</em>
</td>
<td>
<p>Desired state for Alluxio job master</p>
</td>
</tr>
<tr>
<td>
<code>worker</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.AlluxioCompTemplateSpec">
AlluxioCompTemplateSpec
</a>
</em>
</td>
<td>
<p>Desired state for Alluxio worker</p>
</td>
</tr>
<tr>
<td>
<code>jobWorker</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.AlluxioCompTemplateSpec">
AlluxioCompTemplateSpec
</a>
</em>
</td>
<td>
<p>Desired state for Alluxio job Worker</p>
</td>
</tr>
<tr>
<td>
<code>initUsers</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.InitUsersSpec">
InitUsersSpec
</a>
</em>
</td>
<td>
<p>The spec of init users</p>
</td>
</tr>
<tr>
<td>
<code>fuse</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.AlluxioFuseSpec">
AlluxioFuseSpec
</a>
</em>
</td>
<td>
<p>Desired state for Alluxio Fuse</p>
</td>
</tr>
<tr>
<td>
<code>properties</code></br>
<em>
map[string]string
</em>
</td>
<td>
<p>Configurable properties for Alluxio system. <br>
Refer to <a href="https://docs.alluxio.io/os/user/stable/en/reference/Properties-List.html">Alluxio Configuration Properties</a> for more info</p>
</td>
</tr>
<tr>
<td>
<code>jvmOptions</code></br>
<em>
[]string
</em>
</td>
<td>
<p>Options for JVM</p>
</td>
</tr>
<tr>
<td>
<code>tieredstore</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.Tieredstore">
Tieredstore
</a>
</em>
</td>
<td>
<p>Tiered storage used by Alluxio</p>
</td>
</tr>
<tr>
<td>
<code>data</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.Data">
Data
</a>
</em>
</td>
<td>
<p>Management strategies for the dataset to which the runtime is bound</p>
</td>
</tr>
<tr>
<td>
<code>replicas</code></br>
<em>
int32
</em>
</td>
<td>
<p>The replicas of the worker, need to be specified</p>
</td>
</tr>
<tr>
<td>
<code>runAs</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.User">
User
</a>
</em>
</td>
<td>
<p>Manage the user to run Alluxio Runtime</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.CacheableNodeAffinity">CacheableNodeAffinity
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DatasetSpec">DatasetSpec</a>)
</p>
<p>
<p>CacheableNodeAffinity defines constraints that limit what nodes this dataset can be cached to.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>required</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#nodeselector-v1-core">
Kubernetes core/v1.NodeSelector
</a>
</em>
</td>
<td>
<p>Required specifies hard node constraints that must be met.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.Data">Data
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.AlluxioRuntimeSpec">AlluxioRuntimeSpec</a>)
</p>
<p>
<p>Data management strategies</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>replicas</code></br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>The copies of the dataset</p>
</td>
</tr>
<tr>
<td>
<code>pin</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Pin the dataset or not. Refer to <a href="https://docs.alluxio.io/os/user/stable/en/operation/User-CLI.html#pin">Alluxio User-CLI pin</a></p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.DataLoadCondition">DataLoadCondition
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DataLoadStatus">DataLoadStatus</a>)
</p>
<p>
<p>DataLoadCondition describes conditions that explains transitions on phase</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code></br>
<em>
dataload.DataLoadConditionType
</em>
</td>
<td>
<p>Type of condition, either <code>Complete</code> or <code>Failed</code></p>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#conditionstatus-v1-core">
Kubernetes core/v1.ConditionStatus
</a>
</em>
</td>
<td>
<p>Status of the condition, one of <code>True</code>, <code>False</code> or <code>Unknown</code></p>
</td>
</tr>
<tr>
<td>
<code>reason</code></br>
<em>
string
</em>
</td>
<td>
<p>Reason for the condition&rsquo;s last transition</p>
</td>
</tr>
<tr>
<td>
<code>message</code></br>
<em>
string
</em>
</td>
<td>
<p>Message is a human-readable message indicating details about the transition</p>
</td>
</tr>
<tr>
<td>
<code>lastProbeTime</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>LastProbeTime describes last time this condition was updated.</p>
</td>
</tr>
<tr>
<td>
<code>lastTransitionTime</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>LastTransitionTime describes last time the condition transitioned from one status to another.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.DataLoadSpec">DataLoadSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DataLoad">DataLoad</a>)
</p>
<p>
<p>DataLoadSpec defines the desired state of DataLoad</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>dataset</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.TargetDataset">
TargetDataset
</a>
</em>
</td>
<td>
<p>Dataset defines the target dataset of the DataLoad</p>
</td>
</tr>
<tr>
<td>
<code>loadMetadata</code></br>
<em>
bool
</em>
</td>
<td>
<p>LoadMetadata specifies if the dataload job should load metadata</p>
</td>
</tr>
<tr>
<td>
<code>target</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.TargetPath">
[]TargetPath
</a>
</em>
</td>
<td>
<p>Target defines target paths that needs to be loaded</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.DataLoadStatus">DataLoadStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DataLoad">DataLoad</a>)
</p>
<p>
<p>DataLoadStatus defines the observed state of DataLoad</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>phase</code></br>
<em>
dataload.DataLoadPhase
</em>
</td>
<td>
<p>Phase describes current phase of DataLoad</p>
</td>
</tr>
<tr>
<td>
<code>conditions</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.DataLoadCondition">
[]DataLoadCondition
</a>
</em>
</td>
<td>
<p>Conditions consists of transition information on DataLoad&rsquo;s Phase</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.DatasetCondition">DatasetCondition
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DatasetStatus">DatasetStatus</a>)
</p>
<p>
<p>Condition describes the state of the cache at a certain point.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.DatasetConditionType">
DatasetConditionType
</a>
</em>
</td>
<td>
<p>Type of cache condition.</p>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#conditionstatus-v1-core">
Kubernetes core/v1.ConditionStatus
</a>
</em>
</td>
<td>
<p>Status of the condition, one of True, False, Unknown.</p>
</td>
</tr>
<tr>
<td>
<code>reason</code></br>
<em>
string
</em>
</td>
<td>
<p>The reason for the condition&rsquo;s last transition.</p>
</td>
</tr>
<tr>
<td>
<code>message</code></br>
<em>
string
</em>
</td>
<td>
<p>A human readable message indicating details about the transition.</p>
</td>
</tr>
<tr>
<td>
<code>lastUpdateTime</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>The last time this condition was updated.</p>
</td>
</tr>
<tr>
<td>
<code>lastTransitionTime</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>Last time the condition transitioned from one status to another.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.DatasetSpec">DatasetSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.Dataset">Dataset</a>)
</p>
<p>
<p>DatasetSpec defines the desired state of Dataset</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>mounts</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.Mount">
[]Mount
</a>
</em>
</td>
<td>
<p>Mount Points to be mounted on Alluxio.</p>
</td>
</tr>
<tr>
<td>
<code>owner</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.User">
User
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The owner of the dataset</p>
</td>
</tr>
<tr>
<td>
<code>nodeAffinity</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.CacheableNodeAffinity">
CacheableNodeAffinity
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeAffinity defines constraints that limit what nodes this dataset can be cached to.
This field influences the scheduling of pods that use the cached dataset.</p>
</td>
</tr>
<tr>
<td>
<code>accessModes</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#persistentvolumeaccessmode-v1-core">
[]Kubernetes core/v1.PersistentVolumeAccessMode
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AccessModes contains all ways the volume backing the PVC can be mounted</p>
</td>
</tr>
<tr>
<td>
<code>runtimes</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.Runtime">
[]Runtime
</a>
</em>
</td>
<td>
<p>Runtimes for supporting dataset (e.g. AlluxioRuntime)</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.DatasetStatus">DatasetStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.Dataset">Dataset</a>)
</p>
<p>
<p>DatasetStatus defines the observed state of Dataset</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>ufsTotal</code></br>
<em>
string
</em>
</td>
<td>
<p>Total in GB of dataset in the cluster</p>
</td>
</tr>
<tr>
<td>
<code>phase</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.DatasetPhase">
DatasetPhase
</a>
</em>
</td>
<td>
<p>Dataset Phase. One of the four phases: <code>Pending</code>, <code>Bound</code>, <code>NotBound</code> and <code>Failed</code></p>
</td>
</tr>
<tr>
<td>
<code>runtimes</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.Runtime">
[]Runtime
</a>
</em>
</td>
<td>
<p>Runtimes for supporting dataset</p>
</td>
</tr>
<tr>
<td>
<code>conditions</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.DatasetCondition">
[]DatasetCondition
</a>
</em>
</td>
<td>
<p>Conditions is an array of current observed conditions.</p>
</td>
</tr>
<tr>
<td>
<code>cacheStates</code></br>
<em>
common.CacheStateList
</em>
</td>
<td>
<p>CacheStatus represents the total resources of the dataset.</p>
</td>
</tr>
<tr>
<td>
<code>hcfs</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.HCFSStatus">
HCFSStatus
</a>
</em>
</td>
<td>
<p>HCFSStatus represents hcfs info</p>
</td>
</tr>
<tr>
<td>
<code>fileNum</code></br>
<em>
string
</em>
</td>
<td>
<p>FileNum represents the file numbers of the dataset</p>
</td>
</tr>
<tr>
<td>
<code>dataLoadRef</code></br>
<em>
string
</em>
</td>
<td>
<p>DataLoadRef specifies the running DataLoad job that targets this Dataset.
This is mainly used as a lock to prevent concurrent DataLoad jobs.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.HCFSStatus">HCFSStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DatasetStatus">DatasetStatus</a>)
</p>
<p>
<p>HCFS Endpoint info</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>endpoint</code></br>
<em>
string
</em>
</td>
<td>
<p>Endpoint for accessing</p>
</td>
</tr>
<tr>
<td>
<code>underlayerFileSystemVersion</code></br>
<em>
string
</em>
</td>
<td>
<p>Underlayer HCFS Compatible Version</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.InitUsersSpec">InitUsersSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.AlluxioRuntimeSpec">AlluxioRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.JindoRuntimeSpec">JindoRuntimeSpec</a>)
</p>
<p>
<p>InitUsersSpec is a description of the initialize the users for runtime</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>image</code></br>
<em>
string
</em>
</td>
<td>
<p>Image for initialize the users for runtime(e.g. alluxio/alluxio-User init)</p>
</td>
</tr>
<tr>
<td>
<code>imageTag</code></br>
<em>
string
</em>
</td>
<td>
<p>Image Tag for initialize the users for runtime(e.g. 2.3.0-SNAPSHOT)</p>
</td>
</tr>
<tr>
<td>
<code>imagePullPolicy</code></br>
<em>
string
</em>
</td>
<td>
<p>One of the three policies: <code>Always</code>, <code>IfNotPresent</code>, <code>Never</code></p>
</td>
</tr>
<tr>
<td>
<code>env</code></br>
<em>
map[string]string
</em>
</td>
<td>
<p>Environment variables that will be used by initialize the users for runtime</p>
</td>
</tr>
<tr>
<td>
<code>resources</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Resources that will be requested by initialize the users for runtime. <br>
<br>
Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources
already allocated to the pod.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.JindoCompTemplateSpec">JindoCompTemplateSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.JindoRuntimeSpec">JindoRuntimeSpec</a>)
</p>
<p>
<p>JindoCompTemplateSpec is a description of the Jindo commponents</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>replicas</code></br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Replicas is the desired number of replicas of the given template.
If unspecified, defaults to 1.
replicas is the min replicas of dataset in the cluster</p>
</td>
</tr>
<tr>
<td>
<code>properties</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configurable properties for the Jindo component. <br></p>
</td>
</tr>
<tr>
<td>
<code>ports</code></br>
<em>
map[string]int
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>resources</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Resources that will be requested by the Jindo component. <br>
<br>
Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources
already allocated to the pod.</p>
</td>
</tr>
<tr>
<td>
<code>env</code></br>
<em>
map[string]string
</em>
</td>
<td>
<p>Environment variables that will be used by Jindo component. <br></p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.JindoFuseSpec">JindoFuseSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.JindoRuntimeSpec">JindoRuntimeSpec</a>)
</p>
<p>
<p>JindoFuseSpec is a description of the Jindo Fuse</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>image</code></br>
<em>
string
</em>
</td>
<td>
<p>Image for Jindo Fuse(e.g. jindo/jindo-fuse)</p>
</td>
</tr>
<tr>
<td>
<code>imageTag</code></br>
<em>
string
</em>
</td>
<td>
<p>Image Tag for Jindo Fuse(e.g. 2.3.0-SNAPSHOT)</p>
</td>
</tr>
<tr>
<td>
<code>imagePullPolicy</code></br>
<em>
string
</em>
</td>
<td>
<p>One of the three policies: <code>Always</code>, <code>IfNotPresent</code>, <code>Never</code></p>
</td>
</tr>
<tr>
<td>
<code>properties</code></br>
<em>
map[string]int
</em>
</td>
<td>
<p>Configurable properties for Jindo System. <br></p>
</td>
</tr>
<tr>
<td>
<code>env</code></br>
<em>
map[string]string
</em>
</td>
<td>
<p>Environment variables that will be used by Jindo Fuse</p>
</td>
</tr>
<tr>
<td>
<code>resources</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Resources that will be requested by Jindo Fuse. <br>
<br>
Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources
already allocated to the pod.</p>
</td>
</tr>
<tr>
<td>
<code>args</code></br>
<em>
[]string
</em>
</td>
<td>
<p>Arguments that will be passed to Jindo Fuse</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.JindoRuntime">JindoRuntime
</h3>
<p>
<p>JindoRuntime is the Schema for the jindoruntimes API</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.JindoRuntimeSpec">
JindoRuntimeSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>jindoVersion</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.VersionSpec">
VersionSpec
</a>
</em>
</td>
<td>
<p>The version information that instructs fluid to orchestrate a particular version of Jindo.</p>
</td>
</tr>
<tr>
<td>
<code>master</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.JindoCompTemplateSpec">
JindoCompTemplateSpec
</a>
</em>
</td>
<td>
<p>Desired state for Jindo master</p>
</td>
</tr>
<tr>
<td>
<code>jobMaster</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.JindoCompTemplateSpec">
JindoCompTemplateSpec
</a>
</em>
</td>
<td>
<p>Desired state for Jindo job master</p>
</td>
</tr>
<tr>
<td>
<code>worker</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.JindoCompTemplateSpec">
JindoCompTemplateSpec
</a>
</em>
</td>
<td>
<p>Desired state for Jindo worker</p>
</td>
</tr>
<tr>
<td>
<code>jobWorker</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.JindoCompTemplateSpec">
JindoCompTemplateSpec
</a>
</em>
</td>
<td>
<p>Desired state for Jindo job Worker</p>
</td>
</tr>
<tr>
<td>
<code>initUsers</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.InitUsersSpec">
InitUsersSpec
</a>
</em>
</td>
<td>
<p>The spec of init users</p>
</td>
</tr>
<tr>
<td>
<code>fuse</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.JindoFuseSpec">
JindoFuseSpec
</a>
</em>
</td>
<td>
<p>Desired state for Jindo Fuse</p>
</td>
</tr>
<tr>
<td>
<code>properties</code></br>
<em>
map[string]string
</em>
</td>
<td>
<p>Configurable properties for Jindo system. <br></p>
</td>
</tr>
<tr>
<td>
<code>tieredstore</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.Tieredstore">
Tieredstore
</a>
</em>
</td>
<td>
<p>Tiered storage used by Jindo</p>
</td>
</tr>
<tr>
<td>
<code>replicas</code></br>
<em>
int32
</em>
</td>
<td>
<p>The replicas of the worker, need to be specified</p>
</td>
</tr>
<tr>
<td>
<code>runAs</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.User">
User
</a>
</em>
</td>
<td>
<p>Manage the user to run Jindo Runtime</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.RuntimeStatus">
RuntimeStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.JindoRuntimeSpec">JindoRuntimeSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.JindoRuntime">JindoRuntime</a>)
</p>
<p>
<p>JindoRuntimeSpec defines the desired state of JindoRuntime</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>jindoVersion</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.VersionSpec">
VersionSpec
</a>
</em>
</td>
<td>
<p>The version information that instructs fluid to orchestrate a particular version of Jindo.</p>
</td>
</tr>
<tr>
<td>
<code>master</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.JindoCompTemplateSpec">
JindoCompTemplateSpec
</a>
</em>
</td>
<td>
<p>Desired state for Jindo master</p>
</td>
</tr>
<tr>
<td>
<code>jobMaster</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.JindoCompTemplateSpec">
JindoCompTemplateSpec
</a>
</em>
</td>
<td>
<p>Desired state for Jindo job master</p>
</td>
</tr>
<tr>
<td>
<code>worker</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.JindoCompTemplateSpec">
JindoCompTemplateSpec
</a>
</em>
</td>
<td>
<p>Desired state for Jindo worker</p>
</td>
</tr>
<tr>
<td>
<code>jobWorker</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.JindoCompTemplateSpec">
JindoCompTemplateSpec
</a>
</em>
</td>
<td>
<p>Desired state for Jindo job Worker</p>
</td>
</tr>
<tr>
<td>
<code>initUsers</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.InitUsersSpec">
InitUsersSpec
</a>
</em>
</td>
<td>
<p>The spec of init users</p>
</td>
</tr>
<tr>
<td>
<code>fuse</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.JindoFuseSpec">
JindoFuseSpec
</a>
</em>
</td>
<td>
<p>Desired state for Jindo Fuse</p>
</td>
</tr>
<tr>
<td>
<code>properties</code></br>
<em>
map[string]string
</em>
</td>
<td>
<p>Configurable properties for Jindo system. <br></p>
</td>
</tr>
<tr>
<td>
<code>tieredstore</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.Tieredstore">
Tieredstore
</a>
</em>
</td>
<td>
<p>Tiered storage used by Jindo</p>
</td>
</tr>
<tr>
<td>
<code>replicas</code></br>
<em>
int32
</em>
</td>
<td>
<p>The replicas of the worker, need to be specified</p>
</td>
</tr>
<tr>
<td>
<code>runAs</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.User">
User
</a>
</em>
</td>
<td>
<p>Manage the user to run Jindo Runtime</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.Level">Level
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.Tieredstore">Tieredstore</a>)
</p>
<p>
<p>Level describes configurations a tier needs. <br>
Refer to <a href="https://docs.alluxio.io/os/user/stable/en/core-services/Caching.html#configuring-tiered-storage">Configuring Tiered Storage</a> for more info</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>mediumtype</code></br>
<em>
common.MediumType
</em>
</td>
<td>
<p>Medium Type of the tier. One of the three types: <code>MEM</code>, <code>SSD</code>, <code>HDD</code></p>
</td>
</tr>
<tr>
<td>
<code>path</code></br>
<em>
string
</em>
</td>
<td>
<p>File path to be used for the tier (e.g. /mnt/ramdisk)</p>
</td>
</tr>
<tr>
<td>
<code>quota</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#quantity-resource-core">
Kubernetes resource.Quantity
</a>
</em>
</td>
<td>
<p>Quota for the tier. (e.g. 100GB)</p>
</td>
</tr>
<tr>
<td>
<code>high</code></br>
<em>
string
</em>
</td>
<td>
<p>Ratio of high watermark of the tier (e.g. 0.9)</p>
</td>
</tr>
<tr>
<td>
<code>low</code></br>
<em>
string
</em>
</td>
<td>
<p>Ratio of low watermark of the tier (e.g. 0.7)</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.Mount">Mount
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DatasetSpec">DatasetSpec</a>)
</p>
<p>
<p>Mount describes a mounting. <br>
Refer to <a href="https://docs.alluxio.io/os/user/stable/en/ufs/S3.html">Alluxio Storage Integrations</a> for more info</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>mountPoint</code></br>
<em>
string
</em>
</td>
<td>
<p>MountPoint is the mount point of source.</p>
</td>
</tr>
<tr>
<td>
<code>options</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The Mount Options. <br>
Refer to <a href="https://docs.alluxio.io/os/user/stable/en/reference/Properties-List.html">Mount Options</a>.  <br>
The option has Prefix &lsquo;fs.&rsquo; And you can Learn more from
<a href="https://docs.alluxio.io/os/user/stable/en/ufs/S3.html">The Storage Integrations</a></p>
</td>
</tr>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<p>The name of mount</p>
</td>
</tr>
<tr>
<td>
<code>path</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The path of mount, if not set will be /{Name}</p>
</td>
</tr>
<tr>
<td>
<code>readOnly</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional: Defaults to false (read-write).</p>
</td>
</tr>
<tr>
<td>
<code>shared</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional: Defaults to false (shared).</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.Runtime">Runtime
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DatasetSpec">DatasetSpec</a>, 
<a href="#data.fluid.io/v1alpha1.DatasetStatus">DatasetStatus</a>)
</p>
<p>
<p>Runtime describes a runtime to be used to support dataset</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<p>Name of the runtime object</p>
</td>
</tr>
<tr>
<td>
<code>namespace</code></br>
<em>
string
</em>
</td>
<td>
<p>Namespace of the runtime object</p>
</td>
</tr>
<tr>
<td>
<code>category</code></br>
<em>
common.Category
</em>
</td>
<td>
<p>Category the runtime object belongs to (e.g. Accelerate)</p>
</td>
</tr>
<tr>
<td>
<code>type</code></br>
<em>
string
</em>
</td>
<td>
<p>Runtime object&rsquo;s type (e.g. Alluxio)</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.RuntimeCondition">RuntimeCondition
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.RuntimeStatus">RuntimeStatus</a>)
</p>
<p>
<p>Condition describes the state of the cache at a certain point.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.RuntimeConditionType">
RuntimeConditionType
</a>
</em>
</td>
<td>
<p>Type of cache condition.</p>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#conditionstatus-v1-core">
Kubernetes core/v1.ConditionStatus
</a>
</em>
</td>
<td>
<p>Status of the condition, one of True, False, Unknown.</p>
</td>
</tr>
<tr>
<td>
<code>reason</code></br>
<em>
string
</em>
</td>
<td>
<p>The reason for the condition&rsquo;s last transition.</p>
</td>
</tr>
<tr>
<td>
<code>message</code></br>
<em>
string
</em>
</td>
<td>
<p>A human readable message indicating details about the transition.</p>
</td>
</tr>
<tr>
<td>
<code>lastProbeTime</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>The last time this condition was updated.</p>
</td>
</tr>
<tr>
<td>
<code>lastTransitionTime</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>Last time the condition transitioned from one status to another.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.RuntimeStatus">RuntimeStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.AlluxioRuntime">AlluxioRuntime</a>, 
<a href="#data.fluid.io/v1alpha1.JindoRuntime">JindoRuntime</a>)
</p>
<p>
<p>RuntimeStatus defines the observed state of Runtime</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>valueFile</code></br>
<em>
string
</em>
</td>
<td>
<p>config map used to set configurations</p>
</td>
</tr>
<tr>
<td>
<code>masterPhase</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.RuntimePhase">
RuntimePhase
</a>
</em>
</td>
<td>
<p>MasterPhase is the master running phase</p>
</td>
</tr>
<tr>
<td>
<code>masterReason</code></br>
<em>
string
</em>
</td>
<td>
<p>Reason for Master&rsquo;s condition transition</p>
</td>
</tr>
<tr>
<td>
<code>workerPhase</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.RuntimePhase">
RuntimePhase
</a>
</em>
</td>
<td>
<p>WorkerPhase is the worker running phase</p>
</td>
</tr>
<tr>
<td>
<code>workerReason</code></br>
<em>
string
</em>
</td>
<td>
<p>Reason for Worker&rsquo;s condition transition</p>
</td>
</tr>
<tr>
<td>
<code>desiredWorkerNumberScheduled</code></br>
<em>
int32
</em>
</td>
<td>
<p>The total number of nodes that should be running the runtime worker
pod (including nodes correctly running the runtime worker pod).</p>
</td>
</tr>
<tr>
<td>
<code>currentWorkerNumberScheduled</code></br>
<em>
int32
</em>
</td>
<td>
<p>The total number of nodes that can be running the runtime worker
pod (including nodes correctly running the runtime worker pod).</p>
</td>
</tr>
<tr>
<td>
<code>workerNumberReady</code></br>
<em>
int32
</em>
</td>
<td>
<p>The number of nodes that should be running the runtime worker pod and have one
or more of the runtime worker pod running and ready.</p>
</td>
</tr>
<tr>
<td>
<code>workerNumberAvailable</code></br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>The number of nodes that should be running the
runtime worker pod and have one or more of the runtime worker pod running and
available (ready for at least spec.minReadySeconds)</p>
</td>
</tr>
<tr>
<td>
<code>workerNumberUnavailable</code></br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>The number of nodes that should be running the
runtime worker pod and have none of the runtime worker pod running and available
(ready for at least spec.minReadySeconds)</p>
</td>
</tr>
<tr>
<td>
<code>desiredMasterNumberScheduled</code></br>
<em>
int32
</em>
</td>
<td>
<p>The total number of nodes that should be running the runtime
pod (including nodes correctly running the runtime master pod).</p>
</td>
</tr>
<tr>
<td>
<code>currentMasterNumberScheduled</code></br>
<em>
int32
</em>
</td>
<td>
<p>The total number of nodes that should be running the runtime
pod (including nodes correctly running the runtime master pod).</p>
</td>
</tr>
<tr>
<td>
<code>masterNumberReady</code></br>
<em>
int32
</em>
</td>
<td>
<p>The number of nodes that should be running the runtime worker pod and have zero
or more of the runtime master pod running and ready.</p>
</td>
</tr>
<tr>
<td>
<code>fusePhase</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.RuntimePhase">
RuntimePhase
</a>
</em>
</td>
<td>
<p>FusePhase is the Fuse running phase</p>
</td>
</tr>
<tr>
<td>
<code>fuseReason</code></br>
<em>
string
</em>
</td>
<td>
<p>Reason for the condition&rsquo;s last transition.</p>
</td>
</tr>
<tr>
<td>
<code>currentFuseNumberScheduled</code></br>
<em>
int32
</em>
</td>
<td>
<p>The total number of nodes that can be running the runtime Fuse
pod (including nodes correctly running the runtime Fuse pod).</p>
</td>
</tr>
<tr>
<td>
<code>desiredFuseNumberScheduled</code></br>
<em>
int32
</em>
</td>
<td>
<p>The total number of nodes that should be running the runtime Fuse
pod (including nodes correctly running the runtime Fuse pod).</p>
</td>
</tr>
<tr>
<td>
<code>fuseNumberReady</code></br>
<em>
int32
</em>
</td>
<td>
<p>The number of nodes that should be running the runtime Fuse pod and have one
or more of the runtime Fuse pod running and ready.</p>
</td>
</tr>
<tr>
<td>
<code>fuseNumberUnavailable</code></br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>The number of nodes that should be running the
runtime fuse pod and have none of the runtime fuse pod running and available
(ready for at least spec.minReadySeconds)</p>
</td>
</tr>
<tr>
<td>
<code>fuseNumberAvailable</code></br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>The number of nodes that should be running the
runtime Fuse pod and have one or more of the runtime Fuse pod running and
available (ready for at least spec.minReadySeconds)</p>
</td>
</tr>
<tr>
<td>
<code>conditions</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.RuntimeCondition">
[]RuntimeCondition
</a>
</em>
</td>
<td>
<p>Represents the latest available observations of a ddc runtime&rsquo;s current state.</p>
</td>
</tr>
<tr>
<td>
<code>cacheStates</code></br>
<em>
common.CacheStateList
</em>
</td>
<td>
<p>CacheStatus represents the total resources of the dataset.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.TargetDataset">TargetDataset
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DataLoadSpec">DataLoadSpec</a>)
</p>
<p>
<p>TargetDataset defines the target dataset of the DataLoad</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<p>Name defines name of the target dataset</p>
</td>
</tr>
<tr>
<td>
<code>namespace</code></br>
<em>
string
</em>
</td>
<td>
<p>Namespace defines namespace of the target dataset</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.TargetPath">TargetPath
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DataLoadSpec">DataLoadSpec</a>)
</p>
<p>
<p>TargetPath defines the target path of the DataLoad</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>path</code></br>
<em>
string
</em>
</td>
<td>
<p>Path defines path to be load</p>
</td>
</tr>
<tr>
<td>
<code>replicas</code></br>
<em>
int32
</em>
</td>
<td>
<p>Replicas defines how many replicas will be loaded</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.Tieredstore">Tieredstore
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.AlluxioRuntimeSpec">AlluxioRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.JindoRuntimeSpec">JindoRuntimeSpec</a>)
</p>
<p>
<p>Tieredstore is a description of the tiered store</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>levels</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.Level">
[]Level
</a>
</em>
</td>
<td>
<p>configurations for multiple tiers</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.User">User
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.AlluxioRuntimeSpec">AlluxioRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.DatasetSpec">DatasetSpec</a>, 
<a href="#data.fluid.io/v1alpha1.JindoRuntimeSpec">JindoRuntimeSpec</a>)
</p>
<p>
<p>Run as</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>uid</code></br>
<em>
int64
</em>
</td>
<td>
<p>The uid to run the alluxio runtime</p>
</td>
</tr>
<tr>
<td>
<code>gid</code></br>
<em>
int64
</em>
</td>
<td>
<p>The gid to run the alluxio runtime</p>
</td>
</tr>
<tr>
<td>
<code>user</code></br>
<em>
string
</em>
</td>
<td>
<p>The user name to run the alluxio runtime</p>
</td>
</tr>
<tr>
<td>
<code>group</code></br>
<em>
string
</em>
</td>
<td>
<p>The group name to run the alluxio runtime</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.VersionSpec">VersionSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.AlluxioRuntimeSpec">AlluxioRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.JindoRuntimeSpec">JindoRuntimeSpec</a>)
</p>
<p>
<p>VersionSpec represents the settings for the  version that fluid is orchestrating.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>image</code></br>
<em>
string
</em>
</td>
<td>
<p>Image (e.g. alluxio/alluxio)</p>
</td>
</tr>
<tr>
<td>
<code>imageTag</code></br>
<em>
string
</em>
</td>
<td>
<p>Image tag (e.g. 2.3.0-SNAPSHOT)</p>
</td>
</tr>
<tr>
<td>
<code>imagePullPolicy</code></br>
<em>
string
</em>
</td>
<td>
<p>One of the three policies: <code>Always</code>, <code>IfNotPresent</code>, <code>Never</code></p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>dfe295c</code>.
</em></p>
