# API Reference
<h2 id="data.fluid.io/v1alpha1">data.fluid.io/v1alpha1</h2>
<p>
<p>Package v1alpha1 is the v1alpha1 version of the API.</p>
</p>
Resource Types:
<ul><li>
<a href="#data.fluid.io/v1alpha1.AlluxioRuntime">AlluxioRuntime</a>
</li><li>
<a href="#data.fluid.io/v1alpha1.DataBackup">DataBackup</a>
</li><li>
<a href="#data.fluid.io/v1alpha1.DataLoad">DataLoad</a>
</li><li>
<a href="#data.fluid.io/v1alpha1.DataMigrate">DataMigrate</a>
</li><li>
<a href="#data.fluid.io/v1alpha1.Dataset">Dataset</a>
</li><li>
<a href="#data.fluid.io/v1alpha1.EFCRuntime">EFCRuntime</a>
</li><li>
<a href="#data.fluid.io/v1alpha1.GooseFSRuntime">GooseFSRuntime</a>
</li><li>
<a href="#data.fluid.io/v1alpha1.JindoRuntime">JindoRuntime</a>
</li><li>
<a href="#data.fluid.io/v1alpha1.JuiceFSRuntime">JuiceFSRuntime</a>
</li><li>
<a href="#data.fluid.io/v1alpha1.VineyardRuntime">VineyardRuntime</a>
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
<p>The component spec of Alluxio master</p>
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
<p>The component spec of Alluxio job master</p>
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
<p>The component spec of Alluxio worker</p>
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
<p>The component spec of Alluxio job Worker</p>
</td>
</tr>
<tr>
<td>
<code>apiGateway</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.AlluxioCompTemplateSpec">
AlluxioCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of Alluxio API Gateway</p>
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
<p>The component spec of Alluxio Fuse</p>
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
<a href="#data.fluid.io/v1alpha1.TieredStore">
TieredStore
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
<tr>
<td>
<code>disablePrometheus</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Disable monitoring for Alluxio Runtime
Prometheus is enabled by default</p>
</td>
</tr>
<tr>
<td>
<code>hadoopConfig</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Name of the configMap used to support HDFS configurations when using HDFS as Alluxio&rsquo;s UFS. The configMap
must be in the same namespace with the AlluxioRuntime. The configMap should contain user-specific HDFS conf files in it.
For now, only &ldquo;hdfs-site.xml&rdquo; and &ldquo;core-site.xml&rdquo; are supported. It must take the filename of the conf file as the key and content
of the file as the value.</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Volumes is the list of Kubernetes volumes that can be mounted by the alluxio runtime components and/or fuses.</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PodMetadata defines labels and annotations that will be propagated to Alluxio&rsquo;s pods</p>
</td>
</tr>
<tr>
<td>
<code>management</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.RuntimeManagement">
RuntimeManagement
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RuntimeManagement defines policies when managing the runtime</p>
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
<h3 id="data.fluid.io/v1alpha1.DataBackup">DataBackup
</h3>
<p>
<p>DataBackup is the Schema for the backup API</p>
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
<td><code>DataBackup</code></td>
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
<a href="#data.fluid.io/v1alpha1.DataBackupSpec">
DataBackupSpec
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
string
</em>
</td>
<td>
<p>Dataset defines the target dataset of the DataBackup</p>
</td>
</tr>
<tr>
<td>
<code>backupPath</code></br>
<em>
string
</em>
</td>
<td>
<p>BackupPath defines the target path to save data of the DataBackup</p>
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
<p>Manage the user to run Alluxio DataBackup</p>
</td>
</tr>
<tr>
<td>
<code>runAfter</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.OperationRef">
OperationRef
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies that the preceding operation in a workflow</p>
</td>
</tr>
<tr>
<td>
<code>ttlSecondsAfterFinished</code></br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>TTLSecondsAfterFinished is the time second to clean up data operations after finished or failed</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.OperationStatus">
OperationStatus
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
<tr>
<td>
<code>options</code></br>
<em>
map[string]string
</em>
</td>
<td>
<p>Options specifies the extra dataload properties for runtime</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<p>PodMetadata defines labels and annotations that will be propagated to DataLoad pods</p>
</td>
</tr>
<tr>
<td>
<code>affinity</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#affinity-v1-core">
Kubernetes core/v1.Affinity
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Affinity defines affinity for DataLoad pod</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Tolerations defines tolerations for DataLoad pod</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeSelector defiens node selector for DataLoad pod</p>
</td>
</tr>
<tr>
<td>
<code>schedulerName</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>SchedulerName sets the scheduler to be used for DataLoad pod</p>
</td>
</tr>
<tr>
<td>
<code>policy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.Policy">
Policy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>including Once, Cron, OnEvent</p>
</td>
</tr>
<tr>
<td>
<code>schedule</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The schedule in Cron format, only set when policy is cron, see <a href="https://en.wikipedia.org/wiki/Cron">https://en.wikipedia.org/wiki/Cron</a>.</p>
</td>
</tr>
<tr>
<td>
<code>runAfter</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.OperationRef">
OperationRef
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies that the preceding operation in a workflow</p>
</td>
</tr>
<tr>
<td>
<code>ttlSecondsAfterFinished</code></br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>TTLSecondsAfterFinished is the time second to clean up data operations after finished or failed</p>
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
<p>Resources that will be requested by the DataLoad job. <br></p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.OperationStatus">
OperationStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.DataMigrate">DataMigrate
</h3>
<p>
<p>DataMigrate is the Schema for the datamigrates API</p>
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
<td><code>DataMigrate</code></td>
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
<a href="#data.fluid.io/v1alpha1.DataMigrateSpec">
DataMigrateSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>VersionSpec</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.VersionSpec">
VersionSpec
</a>
</em>
</td>
<td>
<p>
(Members of <code>VersionSpec</code> are embedded into this type.)
</p>
<em>(Optional)</em>
<p>The version information that instructs fluid to orchestrate a particular version for data migrate.</p>
</td>
</tr>
<tr>
<td>
<code>from</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.DataToMigrate">
DataToMigrate
</a>
</em>
</td>
<td>
<p>data to migrate source, including dataset and external storage</p>
</td>
</tr>
<tr>
<td>
<code>to</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.DataToMigrate">
DataToMigrate
</a>
</em>
</td>
<td>
<p>data to migrate destination, including dataset and external storage</p>
</td>
</tr>
<tr>
<td>
<code>block</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>if dataMigrate blocked dataset usage, default is false</p>
</td>
</tr>
<tr>
<td>
<code>runtimeType</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>using which runtime to migrate data; if none, take dataset runtime as default</p>
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
<p>options for migrate, different for each runtime</p>
</td>
</tr>
<tr>
<td>
<code>policy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.Policy">
Policy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>policy for migrate, including Once, Cron, OnEvent</p>
</td>
</tr>
<tr>
<td>
<code>schedule</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The schedule in Cron format, only set when policy is cron, see <a href="https://en.wikipedia.org/wiki/Cron">https://en.wikipedia.org/wiki/Cron</a>.</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<p>PodMetadata defines labels and annotations that will be propagated to DataMigrate pods</p>
</td>
</tr>
<tr>
<td>
<code>affinity</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#affinity-v1-core">
Kubernetes core/v1.Affinity
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Affinity defines affinity for DataMigrate pod</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Tolerations defines tolerations for DataMigrate pod</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeSelector defiens node selector for DataMigrate pod</p>
</td>
</tr>
<tr>
<td>
<code>schedulerName</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>SchedulerName sets the scheduler to be used for DataMigrate pod</p>
</td>
</tr>
<tr>
<td>
<code>runAfter</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.OperationRef">
OperationRef
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies that the preceding operation in a workflow</p>
</td>
</tr>
<tr>
<td>
<code>ttlSecondsAfterFinished</code></br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>TTLSecondsAfterFinished is the time second to clean up data operations after finished or failed</p>
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
<p>Resources that will be requested by the DataMigrate job. <br></p>
</td>
</tr>
<tr>
<td>
<code>parallelism</code></br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Parallelism defines the parallelism tasks numbers for DataMigrate. If the value is greater than 1, the job acts
as a launcher, and users should define the WorkerSpec.</p>
</td>
</tr>
<tr>
<td>
<code>parallelOptions</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ParallelOptions defines options like ssh port and ssh secret name when parallelism is greater than 1.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.OperationStatus">
OperationStatus
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
<em>(Optional)</em>
<p>Mount Points to be mounted on cache runtime. <br>
This field can be empty because some runtimes don&rsquo;t need to mount external storage (e.g.
<a href="https://v6d.io/">Vineyard</a>).</p>
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
<code>tolerations</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, the pod&rsquo;s tolerations.</p>
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
<tr>
<td>
<code>placement</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PlacementMode">
PlacementMode
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Manage switch for opening Multiple datasets single node deployment or not
TODO(xieydd) In future, evaluate node resources and runtime resources to decide whether to turn them on</p>
</td>
</tr>
<tr>
<td>
<code>dataRestoreLocation</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.DataRestoreLocation">
DataRestoreLocation
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DataRestoreLocation is the location to load data of dataset  been backuped</p>
</td>
</tr>
<tr>
<td>
<code>sharedOptions</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>SharedOptions is the options to all mount</p>
</td>
</tr>
<tr>
<td>
<code>sharedEncryptOptions</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.EncryptOption">
[]EncryptOption
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SharedEncryptOptions is the encryptOption to all mount</p>
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
<h3 id="data.fluid.io/v1alpha1.EFCRuntime">EFCRuntime
</h3>
<p>
<p>EFCRuntime is the Schema for the efcruntimes API</p>
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
<td><code>EFCRuntime</code></td>
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
<a href="#data.fluid.io/v1alpha1.EFCRuntimeSpec">
EFCRuntimeSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>master</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.EFCCompTemplateSpec">
EFCCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of EFC master</p>
</td>
</tr>
<tr>
<td>
<code>worker</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.EFCCompTemplateSpec">
EFCCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of EFC worker</p>
</td>
</tr>
<tr>
<td>
<code>initFuse</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.InitFuseSpec">
InitFuseSpec
</a>
</em>
</td>
<td>
<p>The spec of init alifuse</p>
</td>
</tr>
<tr>
<td>
<code>fuse</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.EFCFuseSpec">
EFCFuseSpec
</a>
</em>
</td>
<td>
<p>The component spec of EFC Fuse</p>
</td>
</tr>
<tr>
<td>
<code>tieredstore</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.TieredStore">
TieredStore
</a>
</em>
</td>
<td>
<p>Tiered storage used by EFC worker</p>
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
<code>osAdvise</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.OSAdvise">
OSAdvise
</a>
</em>
</td>
<td>
<p>Operating system optimization for EFC</p>
</td>
</tr>
<tr>
<td>
<code>cleanCachePolicy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.CleanCachePolicy">
CleanCachePolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CleanCachePolicy defines cleanCache Policy</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PodMetadata defines labels and annotations that will be propagated to all EFC&rsquo;s pods</p>
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
<h3 id="data.fluid.io/v1alpha1.GooseFSRuntime">GooseFSRuntime
</h3>
<p>
<p>GooseFSRuntime is the Schema for the goosefsruntimes API</p>
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
<td><code>GooseFSRuntime</code></td>
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
<a href="#data.fluid.io/v1alpha1.GooseFSRuntimeSpec">
GooseFSRuntimeSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>goosefsVersion</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.VersionSpec">
VersionSpec
</a>
</em>
</td>
<td>
<p>The version information that instructs fluid to orchestrate a particular version of GooseFS.</p>
</td>
</tr>
<tr>
<td>
<code>master</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.GooseFSCompTemplateSpec">
GooseFSCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of GooseFS master</p>
</td>
</tr>
<tr>
<td>
<code>jobMaster</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.GooseFSCompTemplateSpec">
GooseFSCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of GooseFS job master</p>
</td>
</tr>
<tr>
<td>
<code>worker</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.GooseFSCompTemplateSpec">
GooseFSCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of GooseFS worker</p>
</td>
</tr>
<tr>
<td>
<code>jobWorker</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.GooseFSCompTemplateSpec">
GooseFSCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of GooseFS job Worker</p>
</td>
</tr>
<tr>
<td>
<code>apiGateway</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.GooseFSCompTemplateSpec">
GooseFSCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of GooseFS API Gateway</p>
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
<a href="#data.fluid.io/v1alpha1.GooseFSFuseSpec">
GooseFSFuseSpec
</a>
</em>
</td>
<td>
<p>The component spec of GooseFS Fuse</p>
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
<p>Configurable properties for the GOOSEFS component. <br>
Refer to <a href="https://cloud.tencent.com/document/product/436/56415">GOOSEFS Configuration Properties</a> for more info</p>
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
<a href="#data.fluid.io/v1alpha1.TieredStore">
TieredStore
</a>
</em>
</td>
<td>
<p>Tiered storage used by GooseFS</p>
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
<p>Manage the user to run GooseFS Runtime
GooseFS support POSIX-ACL and Apache Ranger to manager authorization
TODO(chrisydxie@tencent.com) Support Apache Ranger.</p>
</td>
</tr>
<tr>
<td>
<code>disablePrometheus</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Disable monitoring for GooseFS Runtime
Prometheus is enabled by default</p>
</td>
</tr>
<tr>
<td>
<code>hadoopConfig</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Name of the configMap used to support HDFS configurations when using HDFS as GooseFS&rsquo;s UFS. The configMap
must be in the same namespace with the GooseFSRuntime. The configMap should contain user-specific HDFS conf files in it.
For now, only &ldquo;hdfs-site.xml&rdquo; and &ldquo;core-site.xml&rdquo; are supported. It must take the filename of the conf file as the key and content
of the file as the value.</p>
</td>
</tr>
<tr>
<td>
<code>cleanCachePolicy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.CleanCachePolicy">
CleanCachePolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CleanCachePolicy defines cleanCache Policy</p>
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
<td><code>JindoRuntime</code></td>
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
<p>The component spec of Jindo master</p>
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
<p>The component spec of Jindo worker</p>
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
<p>The component spec of Jindo Fuse</p>
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
<a href="#data.fluid.io/v1alpha1.TieredStore">
TieredStore
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
<tr>
<td>
<code>user</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>hadoopConfig</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Name of the configMap used to support HDFS configurations when using HDFS as Jindo&rsquo;s UFS. The configMap
must be in the same namespace with the JindoRuntime. The configMap should contain user-specific HDFS conf files in it.
For now, only &ldquo;hdfs-site.xml&rdquo; and &ldquo;core-site.xml&rdquo; are supported. It must take the filename of the conf file as the key and content
of the file as the value.</p>
</td>
</tr>
<tr>
<td>
<code>secret</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>labels</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Labels will be added on all the JindoFS pods.
DEPRECATED: this is a deprecated field. Please use PodMetadata.Labels instead.
Note: this field is set to be exclusive with PodMetadata.Labels</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PodMetadata defines labels and annotations that will be propagated to all Jindo&rsquo;s fuse pods</p>
</td>
</tr>
<tr>
<td>
<code>logConfig</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>networkmode</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.NetworkMode">
NetworkMode
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use hostnetwork or not</p>
</td>
</tr>
<tr>
<td>
<code>cleanCachePolicy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.CleanCachePolicy">
CleanCachePolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CleanCachePolicy defines cleanCache Policy</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Volumes is the list of Kubernetes volumes that can be mounted by the jindo runtime components and/or fuses.</p>
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
<h3 id="data.fluid.io/v1alpha1.JuiceFSRuntime">JuiceFSRuntime
</h3>
<p>
<p>JuiceFSRuntime is the Schema for the juicefsruntimes API</p>
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
<td><code>JuiceFSRuntime</code></td>
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
<a href="#data.fluid.io/v1alpha1.JuiceFSRuntimeSpec">
JuiceFSRuntimeSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>juicefsVersion</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.VersionSpec">
VersionSpec
</a>
</em>
</td>
<td>
<p>The version information that instructs fluid to orchestrate a particular version of JuiceFS.</p>
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
<code>master</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.JuiceFSCompTemplateSpec">
JuiceFSCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of JuiceFS master</p>
</td>
</tr>
<tr>
<td>
<code>worker</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.JuiceFSCompTemplateSpec">
JuiceFSCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of JuiceFS worker</p>
</td>
</tr>
<tr>
<td>
<code>jobWorker</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.JuiceFSCompTemplateSpec">
JuiceFSCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of JuiceFS job Worker</p>
</td>
</tr>
<tr>
<td>
<code>fuse</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.JuiceFSFuseSpec">
JuiceFSFuseSpec
</a>
</em>
</td>
<td>
<p>Desired state for JuiceFS Fuse</p>
</td>
</tr>
<tr>
<td>
<code>tieredstore</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.TieredStore">
TieredStore
</a>
</em>
</td>
<td>
<p>Tiered storage used by JuiceFS</p>
</td>
</tr>
<tr>
<td>
<code>configs</code></br>
<em>
string
</em>
</td>
<td>
<p>Configs of JuiceFS</p>
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
<p>Manage the user to run Juicefs Runtime</p>
</td>
</tr>
<tr>
<td>
<code>disablePrometheus</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Disable monitoring for JuiceFS Runtime
Prometheus is enabled by default</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Volumes is the list of Kubernetes volumes that can be mounted by the alluxio runtime components and/or fuses.</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PodMetadata defines labels and annotations that will be propagated to JuiceFs&rsquo;s pods.</p>
</td>
</tr>
<tr>
<td>
<code>cleanCachePolicy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.CleanCachePolicy">
CleanCachePolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CleanCachePolicy defines cleanCache Policy</p>
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
<h3 id="data.fluid.io/v1alpha1.VineyardRuntime">VineyardRuntime
</h3>
<p>
<p>VineyardRuntime is the Schema for the VineyardRuntimes API</p>
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
<td><code>VineyardRuntime</code></td>
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
<a href="#data.fluid.io/v1alpha1.VineyardRuntimeSpec">
VineyardRuntimeSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>master</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.MasterSpec">
MasterSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Master holds the configurations for Vineyard Master component
Represents the Etcd component in Vineyard</p>
</td>
</tr>
<tr>
<td>
<code>worker</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.VineyardCompTemplateSpec">
VineyardCompTemplateSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Worker holds the configurations for Vineyard Worker component
Represents the Vineyardd component in Vineyard</p>
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
<p>The replicas of the worker, need to be specified
If worker.replicas and the field are both specified, the field will be respected</p>
</td>
</tr>
<tr>
<td>
<code>fuse</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.VineyardClientSocketSpec">
VineyardClientSocketSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Fuse holds the configurations for Vineyard client socket.
Note that the &ldquo;Fuse&rdquo; here is kept just for API consistency, VineyardRuntime mount a socket file instead of a FUSE filesystem to make data cache available.
Applications can connect to the vineyard runtime components through IPC or RPC.
IPC is the default way to connect to vineyard runtime components, which is more efficient than RPC.
If the socket file is not mounted, the connection will fall back to RPC.</p>
</td>
</tr>
<tr>
<td>
<code>tieredstore</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.TieredStore">
TieredStore
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Tiered storage used by vineyardd
The MediumType can only be <code>MEM</code> and <code>SSD</code>
<code>MEM</code> actually represents the shared memory of vineyardd.
<code>SSD</code> represents the external storage of vineyardd.
Default is as follows.
tieredstore:
levels:
- level: 0
mediumtype: MEM
quota: 4Gi</p>
<p>Choose hostpath as the external storage of vineyardd.
tieredstore:
levels:
- level: 0
mediumtype: MEM
quota: 4Gi
high: &ldquo;0.8&rdquo;
low: &ldquo;0.3&rdquo;
- level: 1
mediumtype: SSD
quota: 10Gi
volumeType: Hostpath
path: /var/spill-path</p>
</td>
</tr>
<tr>
<td>
<code>disablePrometheus</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Disable monitoring metrics for Vineyard Runtime
Default is false</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PodMetadata defines labels and annotations that will be propagated to Vineyard&rsquo;s pods.</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Volumes is the list of Kubernetes volumes that can be mounted by the vineyard components (Master and Worker).
Default is null.</p>
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
<h3 id="data.fluid.io/v1alpha1.APIGatewayStatus">APIGatewayStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.RuntimeStatus">RuntimeStatus</a>)
</p>
<p>
<p>API Gateway</p>
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
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.AffinityPolicy">AffinityPolicy
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.AffinityStrategy">AffinityStrategy</a>)
</p>
<p>
<p>AffinityPolicy the strategy for the affinity between Data Operation Pods.</p>
</p>
<h3 id="data.fluid.io/v1alpha1.AffinityStrategy">AffinityStrategy
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.OperationRef">OperationRef</a>)
</p>
<p>
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
<code>policy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.AffinityPolicy">
AffinityPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Policy one of: &ldquo;&rdquo;, &ldquo;Require&rdquo;, &ldquo;Prefer&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>prefers</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.Prefer">
[]Prefer
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>requires</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.Require">
[]Require
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
<tr>
<td>
<code>enabled</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Enabled or Disabled for the components. For now, only  API Gateway is enabled or disabled.</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeSelector is a selector which must be true for the master to fit on a node</p>
</td>
</tr>
<tr>
<td>
<code>networkMode</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.NetworkMode">
NetworkMode
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use hostnetwork or not</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>VolumeMounts specifies the volumes listed in &ldquo;.spec.volumes&rdquo; to mount into the alluxio runtime component&rsquo;s filesystem.</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PodMetadata defines labels and annotations that will be propagated to Alluxio&rsquo;s pods</p>
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
<tr>
<td>
<code>nodeSelector</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeSelector is a selector which must be true for the fuse client to fit on a node,
this option only effect when global is enabled</p>
</td>
</tr>
<tr>
<td>
<code>cleanPolicy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.FuseCleanPolicy">
FuseCleanPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CleanPolicy decides when to clean Alluxio Fuse pods.
Currently Fluid supports two policies: OnDemand and OnRuntimeDeleted
OnDemand cleans fuse pod once the fuse pod on some node is not needed
OnRuntimeDeleted cleans fuse pod only when the cache runtime is deleted
Defaults to OnRuntimeDeleted</p>
</td>
</tr>
<tr>
<td>
<code>networkMode</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.NetworkMode">
NetworkMode
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use hostnetwork or not</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>VolumeMounts specifies the volumes listed in &ldquo;.spec.volumes&rdquo; to mount into the alluxio runtime component&rsquo;s filesystem.</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PodMetadata defines labels and annotations that will be propagated to Alluxio&rsquo;s fuse pods</p>
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
<p>The component spec of Alluxio master</p>
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
<p>The component spec of Alluxio job master</p>
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
<p>The component spec of Alluxio worker</p>
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
<p>The component spec of Alluxio job Worker</p>
</td>
</tr>
<tr>
<td>
<code>apiGateway</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.AlluxioCompTemplateSpec">
AlluxioCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of Alluxio API Gateway</p>
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
<p>The component spec of Alluxio Fuse</p>
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
<a href="#data.fluid.io/v1alpha1.TieredStore">
TieredStore
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
<tr>
<td>
<code>disablePrometheus</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Disable monitoring for Alluxio Runtime
Prometheus is enabled by default</p>
</td>
</tr>
<tr>
<td>
<code>hadoopConfig</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Name of the configMap used to support HDFS configurations when using HDFS as Alluxio&rsquo;s UFS. The configMap
must be in the same namespace with the AlluxioRuntime. The configMap should contain user-specific HDFS conf files in it.
For now, only &ldquo;hdfs-site.xml&rdquo; and &ldquo;core-site.xml&rdquo; are supported. It must take the filename of the conf file as the key and content
of the file as the value.</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Volumes is the list of Kubernetes volumes that can be mounted by the alluxio runtime components and/or fuses.</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PodMetadata defines labels and annotations that will be propagated to Alluxio&rsquo;s pods</p>
</td>
</tr>
<tr>
<td>
<code>management</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.RuntimeManagement">
RuntimeManagement
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RuntimeManagement defines policies when managing the runtime</p>
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
<h3 id="data.fluid.io/v1alpha1.CleanCachePolicy">CleanCachePolicy
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.EFCRuntimeSpec">EFCRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.GooseFSRuntimeSpec">GooseFSRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.JindoRuntimeSpec">JindoRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.JuiceFSRuntimeSpec">JuiceFSRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.RuntimeManagement">RuntimeManagement</a>)
</p>
<p>
<p>CleanCachePolicy defines policies when cleaning cache</p>
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
<code>gracePeriodSeconds</code></br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional duration in seconds the cache needs to clean gracefully. May be decreased in delete runtime request.
Value must be non-negative integer. The value zero indicates clean immediately via the timeout
command (no opportunity to shut down).
If this value is nil, the default grace period will be used instead.
The grace period is the duration in seconds after the processes running in the pod are sent
a termination signal and the time when the processes are forcibly halted with timeout command.
Set this value longer than the expected cleanup time for your process.</p>
</td>
</tr>
<tr>
<td>
<code>maxRetryAttempts</code></br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional max retry Attempts when cleanCache function returns an error after execution, runtime attempts
to run it three more times by default. With Maximum Retry Attempts, you can customize the maximum number
of retries. This gives you the option to continue processing retries.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.Condition">Condition
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.OperationStatus">OperationStatus</a>)
</p>
<p>
<p>Condition explains the transitions on phase</p>
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
common.ConditionType
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
<h3 id="data.fluid.io/v1alpha1.Data">Data
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.AlluxioRuntimeSpec">AlluxioRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.GooseFSRuntimeSpec">GooseFSRuntimeSpec</a>)
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
<h3 id="data.fluid.io/v1alpha1.DataBackupSpec">DataBackupSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DataBackup">DataBackup</a>)
</p>
<p>
<p>DataBackupSpec defines the desired state of DataBackup</p>
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
string
</em>
</td>
<td>
<p>Dataset defines the target dataset of the DataBackup</p>
</td>
</tr>
<tr>
<td>
<code>backupPath</code></br>
<em>
string
</em>
</td>
<td>
<p>BackupPath defines the target path to save data of the DataBackup</p>
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
<p>Manage the user to run Alluxio DataBackup</p>
</td>
</tr>
<tr>
<td>
<code>runAfter</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.OperationRef">
OperationRef
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies that the preceding operation in a workflow</p>
</td>
</tr>
<tr>
<td>
<code>ttlSecondsAfterFinished</code></br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>TTLSecondsAfterFinished is the time second to clean up data operations after finished or failed</p>
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
<tr>
<td>
<code>options</code></br>
<em>
map[string]string
</em>
</td>
<td>
<p>Options specifies the extra dataload properties for runtime</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<p>PodMetadata defines labels and annotations that will be propagated to DataLoad pods</p>
</td>
</tr>
<tr>
<td>
<code>affinity</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#affinity-v1-core">
Kubernetes core/v1.Affinity
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Affinity defines affinity for DataLoad pod</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Tolerations defines tolerations for DataLoad pod</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeSelector defiens node selector for DataLoad pod</p>
</td>
</tr>
<tr>
<td>
<code>schedulerName</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>SchedulerName sets the scheduler to be used for DataLoad pod</p>
</td>
</tr>
<tr>
<td>
<code>policy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.Policy">
Policy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>including Once, Cron, OnEvent</p>
</td>
</tr>
<tr>
<td>
<code>schedule</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The schedule in Cron format, only set when policy is cron, see <a href="https://en.wikipedia.org/wiki/Cron">https://en.wikipedia.org/wiki/Cron</a>.</p>
</td>
</tr>
<tr>
<td>
<code>runAfter</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.OperationRef">
OperationRef
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies that the preceding operation in a workflow</p>
</td>
</tr>
<tr>
<td>
<code>ttlSecondsAfterFinished</code></br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>TTLSecondsAfterFinished is the time second to clean up data operations after finished or failed</p>
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
<p>Resources that will be requested by the DataLoad job. <br></p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.DataMigrateSpec">DataMigrateSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DataMigrate">DataMigrate</a>)
</p>
<p>
<p>DataMigrateSpec defines the desired state of DataMigrate</p>
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
<code>VersionSpec</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.VersionSpec">
VersionSpec
</a>
</em>
</td>
<td>
<p>
(Members of <code>VersionSpec</code> are embedded into this type.)
</p>
<em>(Optional)</em>
<p>The version information that instructs fluid to orchestrate a particular version for data migrate.</p>
</td>
</tr>
<tr>
<td>
<code>from</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.DataToMigrate">
DataToMigrate
</a>
</em>
</td>
<td>
<p>data to migrate source, including dataset and external storage</p>
</td>
</tr>
<tr>
<td>
<code>to</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.DataToMigrate">
DataToMigrate
</a>
</em>
</td>
<td>
<p>data to migrate destination, including dataset and external storage</p>
</td>
</tr>
<tr>
<td>
<code>block</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>if dataMigrate blocked dataset usage, default is false</p>
</td>
</tr>
<tr>
<td>
<code>runtimeType</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>using which runtime to migrate data; if none, take dataset runtime as default</p>
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
<p>options for migrate, different for each runtime</p>
</td>
</tr>
<tr>
<td>
<code>policy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.Policy">
Policy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>policy for migrate, including Once, Cron, OnEvent</p>
</td>
</tr>
<tr>
<td>
<code>schedule</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The schedule in Cron format, only set when policy is cron, see <a href="https://en.wikipedia.org/wiki/Cron">https://en.wikipedia.org/wiki/Cron</a>.</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<p>PodMetadata defines labels and annotations that will be propagated to DataMigrate pods</p>
</td>
</tr>
<tr>
<td>
<code>affinity</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#affinity-v1-core">
Kubernetes core/v1.Affinity
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Affinity defines affinity for DataMigrate pod</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Tolerations defines tolerations for DataMigrate pod</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeSelector defiens node selector for DataMigrate pod</p>
</td>
</tr>
<tr>
<td>
<code>schedulerName</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>SchedulerName sets the scheduler to be used for DataMigrate pod</p>
</td>
</tr>
<tr>
<td>
<code>runAfter</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.OperationRef">
OperationRef
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies that the preceding operation in a workflow</p>
</td>
</tr>
<tr>
<td>
<code>ttlSecondsAfterFinished</code></br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>TTLSecondsAfterFinished is the time second to clean up data operations after finished or failed</p>
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
<p>Resources that will be requested by the DataMigrate job. <br></p>
</td>
</tr>
<tr>
<td>
<code>parallelism</code></br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Parallelism defines the parallelism tasks numbers for DataMigrate. If the value is greater than 1, the job acts
as a launcher, and users should define the WorkerSpec.</p>
</td>
</tr>
<tr>
<td>
<code>parallelOptions</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ParallelOptions defines options like ssh port and ssh secret name when parallelism is greater than 1.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.DataProcess">DataProcess
</h3>
<p>
<p>DataProcess is the Schema for the dataprocesses API</p>
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
<a href="#data.fluid.io/v1alpha1.DataProcessSpec">
DataProcessSpec
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
<a href="#data.fluid.io/v1alpha1.TargetDatasetWithMountPath">
TargetDatasetWithMountPath
</a>
</em>
</td>
<td>
<p>Dataset specifies the target dataset and its mount path.</p>
</td>
</tr>
<tr>
<td>
<code>processor</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.Processor">
Processor
</a>
</em>
</td>
<td>
<p>Processor specify how to process data.</p>
</td>
</tr>
<tr>
<td>
<code>runAfter</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.OperationRef">
OperationRef
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies that the preceding operation in a workflow</p>
</td>
</tr>
<tr>
<td>
<code>ttlSecondsAfterFinished</code></br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>TTLSecondsAfterFinished is the time second to clean up data operations after finished or failed</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.OperationStatus">
OperationStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.DataProcessSpec">DataProcessSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DataProcess">DataProcess</a>)
</p>
<p>
<p>DataProcessSpec defines the desired state of DataProcess</p>
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
<a href="#data.fluid.io/v1alpha1.TargetDatasetWithMountPath">
TargetDatasetWithMountPath
</a>
</em>
</td>
<td>
<p>Dataset specifies the target dataset and its mount path.</p>
</td>
</tr>
<tr>
<td>
<code>processor</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.Processor">
Processor
</a>
</em>
</td>
<td>
<p>Processor specify how to process data.</p>
</td>
</tr>
<tr>
<td>
<code>runAfter</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.OperationRef">
OperationRef
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies that the preceding operation in a workflow</p>
</td>
</tr>
<tr>
<td>
<code>ttlSecondsAfterFinished</code></br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>TTLSecondsAfterFinished is the time second to clean up data operations after finished or failed</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.DataRestoreLocation">DataRestoreLocation
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DatasetSpec">DatasetSpec</a>)
</p>
<p>
<p>DataRestoreLocation describes the spec restore location of  Dataset</p>
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
<em>(Optional)</em>
<p>Path describes the path of restore, in the form of  local://subpath or pvc://<pvcName>/subpath</p>
</td>
</tr>
<tr>
<td>
<code>nodeName</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeName describes the nodeName of restore if Path is  in the form of local://subpath</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.DataToMigrate">DataToMigrate
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DataMigrateSpec">DataMigrateSpec</a>)
</p>
<p>
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
<a href="#data.fluid.io/v1alpha1.DatasetToMigrate">
DatasetToMigrate
</a>
</em>
</td>
<td>
<p>dataset to migrate</p>
</td>
</tr>
<tr>
<td>
<code>externalStorage</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.ExternalStorage">
ExternalStorage
</a>
</em>
</td>
<td>
<p>external storage for data migrate</p>
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
<em>(Optional)</em>
<p>Mount Points to be mounted on cache runtime. <br>
This field can be empty because some runtimes don&rsquo;t need to mount external storage (e.g.
<a href="https://v6d.io/">Vineyard</a>).</p>
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
<code>tolerations</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, the pod&rsquo;s tolerations.</p>
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
<tr>
<td>
<code>placement</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PlacementMode">
PlacementMode
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Manage switch for opening Multiple datasets single node deployment or not
TODO(xieydd) In future, evaluate node resources and runtime resources to decide whether to turn them on</p>
</td>
</tr>
<tr>
<td>
<code>dataRestoreLocation</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.DataRestoreLocation">
DataRestoreLocation
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DataRestoreLocation is the location to load data of dataset  been backuped</p>
</td>
</tr>
<tr>
<td>
<code>sharedOptions</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>SharedOptions is the options to all mount</p>
</td>
</tr>
<tr>
<td>
<code>sharedEncryptOptions</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.EncryptOption">
[]EncryptOption
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SharedEncryptOptions is the encryptOption to all mount</p>
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
<code>mounts</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.Mount">
[]Mount
</a>
</em>
</td>
<td>
<p>the info of mount points have been mounted</p>
</td>
</tr>
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
This is mainly used as a lock to prevent concurrent DataLoad jobs.
Deprecated, use OperationRef instead</p>
</td>
</tr>
<tr>
<td>
<code>dataBackupRef</code></br>
<em>
string
</em>
</td>
<td>
<p>DataBackupRef specifies the running Backup job that targets this Dataset.
This is mainly used as a lock to prevent concurrent DataBackup jobs.
Deprecated, use OperationRef instead</p>
</td>
</tr>
<tr>
<td>
<code>operationRef</code></br>
<em>
map[string]string
</em>
</td>
<td>
<p>OperationRef specifies the Operation that targets this Dataset.
This is mainly used as a lock to prevent concurrent same Operation jobs.</p>
</td>
</tr>
<tr>
<td>
<code>datasetRef</code></br>
<em>
[]string
</em>
</td>
<td>
<p>DatasetRef specifies the datasets namespaced name mounting this Dataset.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.DatasetToMigrate">DatasetToMigrate
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DataToMigrate">DataToMigrate</a>)
</p>
<p>
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
<p>name of dataset</p>
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
<p>namespace of dataset</p>
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
<p>path to migrate</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.EFCCompTemplateSpec">EFCCompTemplateSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.EFCRuntimeSpec">EFCRuntimeSpec</a>)
</p>
<p>
<p>EFCCompTemplateSpec is a description of the EFC components</p>
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
<code>version</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.VersionSpec">
VersionSpec
</a>
</em>
</td>
<td>
<p>The version information that instructs fluid to orchestrate a particular version of EFC Comp</p>
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
<p>Configurable properties for the EFC component.</p>
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
<p>Ports used by EFC(e.g. rpc: 19998 for master).</p>
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
<p>Resources that will be requested by the EFC component. <br>
<br>
Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources
already allocated to the pod.</p>
</td>
</tr>
<tr>
<td>
<code>disabled</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Enabled or Disabled for the components.
Default enable.</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeSelector is a selector which must be true for the component to fit on a node.</p>
</td>
</tr>
<tr>
<td>
<code>networkMode</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.NetworkMode">
NetworkMode
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use host network or not.</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PodMetadata defines labels and annotations that will be propagated to EFC&rsquo;s master and worker pods</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.EFCFuseSpec">EFCFuseSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.EFCRuntimeSpec">EFCRuntimeSpec</a>)
</p>
<p>
<p>EFCFuseSpec is a description of the EFC Fuse</p>
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
<code>version</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.VersionSpec">
VersionSpec
</a>
</em>
</td>
<td>
<p>The version information that instructs fluid to orchestrate a particular version of EFC Fuse</p>
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
<p>Configurable properties for EFC fuse</p>
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
<p>Resources that will be requested by EFC Fuse. <br>
<br>
Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources
already allocated to the pod.</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeSelector is a selector which must be true for the fuse client to fit on a node,
this option only effect when global is enabled</p>
</td>
</tr>
<tr>
<td>
<code>cleanPolicy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.FuseCleanPolicy">
FuseCleanPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CleanPolicy decides when to clean EFC Fuse pods.
Currently Fluid supports two policies: OnDemand and OnRuntimeDeleted
OnDemand cleans fuse pod once th fuse pod on some node is not needed
OnRuntimeDeleted cleans fuse pod only when the cache runtime is deleted
Defaults to OnRuntimeDeleted</p>
</td>
</tr>
<tr>
<td>
<code>networkMode</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.NetworkMode">
NetworkMode
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use hostnetwork or not</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PodMetadata defines labels and annotations that will be propagated to EFC&rsquo;s fuse pods</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.EFCRuntimeSpec">EFCRuntimeSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.EFCRuntime">EFCRuntime</a>)
</p>
<p>
<p>EFCRuntimeSpec defines the desired state of EFCRuntime</p>
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
<code>master</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.EFCCompTemplateSpec">
EFCCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of EFC master</p>
</td>
</tr>
<tr>
<td>
<code>worker</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.EFCCompTemplateSpec">
EFCCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of EFC worker</p>
</td>
</tr>
<tr>
<td>
<code>initFuse</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.InitFuseSpec">
InitFuseSpec
</a>
</em>
</td>
<td>
<p>The spec of init alifuse</p>
</td>
</tr>
<tr>
<td>
<code>fuse</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.EFCFuseSpec">
EFCFuseSpec
</a>
</em>
</td>
<td>
<p>The component spec of EFC Fuse</p>
</td>
</tr>
<tr>
<td>
<code>tieredstore</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.TieredStore">
TieredStore
</a>
</em>
</td>
<td>
<p>Tiered storage used by EFC worker</p>
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
<code>osAdvise</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.OSAdvise">
OSAdvise
</a>
</em>
</td>
<td>
<p>Operating system optimization for EFC</p>
</td>
</tr>
<tr>
<td>
<code>cleanCachePolicy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.CleanCachePolicy">
CleanCachePolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CleanCachePolicy defines cleanCache Policy</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PodMetadata defines labels and annotations that will be propagated to all EFC&rsquo;s pods</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.EncryptOption">EncryptOption
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DatasetSpec">DatasetSpec</a>, 
<a href="#data.fluid.io/v1alpha1.ExternalEndpointSpec">ExternalEndpointSpec</a>, 
<a href="#data.fluid.io/v1alpha1.ExternalStorage">ExternalStorage</a>, 
<a href="#data.fluid.io/v1alpha1.Mount">Mount</a>)
</p>
<p>
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
<p>The name of encryptOption</p>
</td>
</tr>
<tr>
<td>
<code>valueFrom</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.EncryptOptionSource">
EncryptOptionSource
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The valueFrom of encryptOption</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.EncryptOptionSource">EncryptOptionSource
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.EncryptOption">EncryptOption</a>)
</p>
<p>
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
<code>secretKeyRef</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.SecretKeySelector">
SecretKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The encryptInfo obtained from secret</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.ExternalEndpointSpec">ExternalEndpointSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.MasterSpec">MasterSpec</a>)
</p>
<p>
<p>ExternalEndpointSpec defines the configurations for external etcd cluster</p>
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
<code>uri</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>URI specifies the endpoint of external Etcd cluster
E,g. &ldquo;etcd-svc.etcd-namespace.svc.cluster.local:2379&rdquo;
Default is not set and use http protocol to connect to external etcd cluster</p>
</td>
</tr>
<tr>
<td>
<code>encryptOptions</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.EncryptOption">
[]EncryptOption
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>encrypt info for accessing the external etcd cluster</p>
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
<p>Configurable options for External Etcd cluster.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.ExternalStorage">ExternalStorage
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DataToMigrate">DataToMigrate</a>)
</p>
<p>
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
<code>uri</code></br>
<em>
string
</em>
</td>
<td>
<p>type of external storage, including s3, oss, gcs, ceph, nfs, pvc, etc. (related to runtime)</p>
</td>
</tr>
<tr>
<td>
<code>encryptOptions</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.EncryptOption">
[]EncryptOption
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>encrypt info for external storage</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.FuseCleanPolicy">FuseCleanPolicy
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.AlluxioFuseSpec">AlluxioFuseSpec</a>, 
<a href="#data.fluid.io/v1alpha1.EFCFuseSpec">EFCFuseSpec</a>, 
<a href="#data.fluid.io/v1alpha1.GooseFSFuseSpec">GooseFSFuseSpec</a>, 
<a href="#data.fluid.io/v1alpha1.JindoFuseSpec">JindoFuseSpec</a>, 
<a href="#data.fluid.io/v1alpha1.JuiceFSFuseSpec">JuiceFSFuseSpec</a>, 
<a href="#data.fluid.io/v1alpha1.ThinFuseSpec">ThinFuseSpec</a>, 
<a href="#data.fluid.io/v1alpha1.VineyardClientSocketSpec">VineyardClientSocketSpec</a>)
</p>
<p>
</p>
<h3 id="data.fluid.io/v1alpha1.GooseFSCompTemplateSpec">GooseFSCompTemplateSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.GooseFSRuntimeSpec">GooseFSRuntimeSpec</a>)
</p>
<p>
<p>GooseFSCompTemplateSpec is a description of the GooseFS commponents</p>
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
<p>Configurable properties for the GOOSEFS component. <br>
Refer to <a href="https://cloud.tencent.com/document/product/436/56415">GOOSEFS Configuration Properties</a> for more info</p>
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
<p>Ports used by GooseFS(e.g. rpc: 19998 for master)</p>
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
<p>Resources that will be requested by the GooseFS component. <br>
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
<p>Environment variables that will be used by GooseFS component. <br></p>
</td>
</tr>
<tr>
<td>
<code>enabled</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Enabled or Disabled for the components. For now, only  API Gateway is enabled or disabled.</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeSelector is a selector which must be true for the master to fit on a node</p>
</td>
</tr>
<tr>
<td>
<code>annotations</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Annotations is an unstructured key value map stored with a resource that may be
set by external tools to store and retrieve arbitrary metadata. They are not
queryable and should be preserved when modifying objects.
More info: <a href="http://kubernetes.io/docs/user-guide/annotations">http://kubernetes.io/docs/user-guide/annotations</a></p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.GooseFSFuseSpec">GooseFSFuseSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.GooseFSRuntimeSpec">GooseFSRuntimeSpec</a>)
</p>
<p>
<p>GooseFSFuseSpec is a description of the GooseFS Fuse</p>
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
<p>Image for GooseFS Fuse(e.g. goosefs/goosefs-fuse)</p>
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
<p>Image Tag for GooseFS Fuse(e.g. v1.0.1)</p>
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
<p>Configurable properties for the GOOSEFS component. <br>
Refer to <a href="https://cloud.tencent.com/document/product/436/56415">GOOSEFS Configuration Properties</a> for more info</p>
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
<p>Environment variables that will be used by GooseFS Fuse</p>
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
<p>Resources that will be requested by GooseFS Fuse. <br>
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
<p>Arguments that will be passed to GooseFS Fuse</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeSelector is a selector which must be true for the fuse client to fit on a node,
this option only effect when global is enabled</p>
</td>
</tr>
<tr>
<td>
<code>cleanPolicy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.FuseCleanPolicy">
FuseCleanPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CleanPolicy decides when to clean GooseFS Fuse pods.
Currently Fluid supports two policies: OnDemand and OnRuntimeDeleted
OnDemand cleans fuse pod once th fuse pod on some node is not needed
OnRuntimeDeleted cleans fuse pod only when the cache runtime is deleted
Defaults to OnRuntimeDeleted</p>
</td>
</tr>
<tr>
<td>
<code>annotations</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Annotations is an unstructured key value map stored with a resource that may be
set by external tools to store and retrieve arbitrary metadata. They are not
queryable and should be preserved when modifying objects.
More info: <a href="http://kubernetes.io/docs/user-guide/annotations">http://kubernetes.io/docs/user-guide/annotations</a></p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.GooseFSRuntimeSpec">GooseFSRuntimeSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.GooseFSRuntime">GooseFSRuntime</a>)
</p>
<p>
<p>GooseFSRuntimeSpec defines the desired state of GooseFSRuntime</p>
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
<code>goosefsVersion</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.VersionSpec">
VersionSpec
</a>
</em>
</td>
<td>
<p>The version information that instructs fluid to orchestrate a particular version of GooseFS.</p>
</td>
</tr>
<tr>
<td>
<code>master</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.GooseFSCompTemplateSpec">
GooseFSCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of GooseFS master</p>
</td>
</tr>
<tr>
<td>
<code>jobMaster</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.GooseFSCompTemplateSpec">
GooseFSCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of GooseFS job master</p>
</td>
</tr>
<tr>
<td>
<code>worker</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.GooseFSCompTemplateSpec">
GooseFSCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of GooseFS worker</p>
</td>
</tr>
<tr>
<td>
<code>jobWorker</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.GooseFSCompTemplateSpec">
GooseFSCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of GooseFS job Worker</p>
</td>
</tr>
<tr>
<td>
<code>apiGateway</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.GooseFSCompTemplateSpec">
GooseFSCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of GooseFS API Gateway</p>
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
<a href="#data.fluid.io/v1alpha1.GooseFSFuseSpec">
GooseFSFuseSpec
</a>
</em>
</td>
<td>
<p>The component spec of GooseFS Fuse</p>
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
<p>Configurable properties for the GOOSEFS component. <br>
Refer to <a href="https://cloud.tencent.com/document/product/436/56415">GOOSEFS Configuration Properties</a> for more info</p>
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
<a href="#data.fluid.io/v1alpha1.TieredStore">
TieredStore
</a>
</em>
</td>
<td>
<p>Tiered storage used by GooseFS</p>
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
<p>Manage the user to run GooseFS Runtime
GooseFS support POSIX-ACL and Apache Ranger to manager authorization
TODO(chrisydxie@tencent.com) Support Apache Ranger.</p>
</td>
</tr>
<tr>
<td>
<code>disablePrometheus</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Disable monitoring for GooseFS Runtime
Prometheus is enabled by default</p>
</td>
</tr>
<tr>
<td>
<code>hadoopConfig</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Name of the configMap used to support HDFS configurations when using HDFS as GooseFS&rsquo;s UFS. The configMap
must be in the same namespace with the GooseFSRuntime. The configMap should contain user-specific HDFS conf files in it.
For now, only &ldquo;hdfs-site.xml&rdquo; and &ldquo;core-site.xml&rdquo; are supported. It must take the filename of the conf file as the key and content
of the file as the value.</p>
</td>
</tr>
<tr>
<td>
<code>cleanCachePolicy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.CleanCachePolicy">
CleanCachePolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CleanCachePolicy defines cleanCache Policy</p>
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
<h3 id="data.fluid.io/v1alpha1.InitFuseSpec">InitFuseSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.EFCRuntimeSpec">EFCRuntimeSpec</a>)
</p>
<p>
<p>InitFuseSpec is a description of initialize the fuse kernel module for runtime</p>
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
<code>version</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.VersionSpec">
VersionSpec
</a>
</em>
</td>
<td>
<p>The version information that instructs fluid to orchestrate a particular version of Alifuse</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.InitUsersSpec">InitUsersSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.AlluxioRuntimeSpec">AlluxioRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.GooseFSRuntimeSpec">GooseFSRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.JuiceFSRuntimeSpec">JuiceFSRuntimeSpec</a>)
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
<tr>
<td>
<code>nodeSelector</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeSelector is a selector which must be true for the master to fit on a node</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, the pod&rsquo;s tolerations.</p>
</td>
</tr>
<tr>
<td>
<code>labels</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Labels will be added on JindoFS Master or Worker pods.
DEPRECATED: This is a deprecated field. Please use PodMetadata instead.
Note: this field is set to be exclusive with PodMetadata.Labels</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PodMetadata defines labels and annotations that will be propagated to Jindo&rsquo;s pods</p>
</td>
</tr>
<tr>
<td>
<code>disabled</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>If disable JindoFS master or worker</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>VolumeMounts specifies the volumes listed in &ldquo;.spec.volumes&rdquo; to mount into the jindo runtime component&rsquo;s filesystem.</p>
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
map[string]string
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
<tr>
<td>
<code>nodeSelector</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeSelector is a selector which must be true for the fuse client to fit on a node,
this option only effect when global is enabled</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, the pod&rsquo;s tolerations.</p>
</td>
</tr>
<tr>
<td>
<code>labels</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Labels will be added on all the JindoFS pods.
DEPRECATED: this is a deprecated field. Please use PodMetadata.Labels instead.
Note: this field is set to be exclusive with PodMetadata.Labels</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PodMetadata defines labels and annotations that will be propagated to Jindo&rsquo;s fuse pods</p>
</td>
</tr>
<tr>
<td>
<code>cleanPolicy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.FuseCleanPolicy">
FuseCleanPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CleanPolicy decides when to clean JindoFS Fuse pods.
Currently Fluid supports two policies: OnDemand and OnRuntimeDeleted
OnDemand cleans fuse pod once th fuse pod on some node is not needed
OnRuntimeDeleted cleans fuse pod only when the cache runtime is deleted
Defaults to OnRuntimeDeleted</p>
</td>
</tr>
<tr>
<td>
<code>disabled</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>If disable JindoFS fuse</p>
</td>
</tr>
<tr>
<td>
<code>logConfig</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
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
<p>The component spec of Jindo master</p>
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
<p>The component spec of Jindo worker</p>
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
<p>The component spec of Jindo Fuse</p>
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
<a href="#data.fluid.io/v1alpha1.TieredStore">
TieredStore
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
<tr>
<td>
<code>user</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>hadoopConfig</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Name of the configMap used to support HDFS configurations when using HDFS as Jindo&rsquo;s UFS. The configMap
must be in the same namespace with the JindoRuntime. The configMap should contain user-specific HDFS conf files in it.
For now, only &ldquo;hdfs-site.xml&rdquo; and &ldquo;core-site.xml&rdquo; are supported. It must take the filename of the conf file as the key and content
of the file as the value.</p>
</td>
</tr>
<tr>
<td>
<code>secret</code></br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>labels</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Labels will be added on all the JindoFS pods.
DEPRECATED: this is a deprecated field. Please use PodMetadata.Labels instead.
Note: this field is set to be exclusive with PodMetadata.Labels</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PodMetadata defines labels and annotations that will be propagated to all Jindo&rsquo;s fuse pods</p>
</td>
</tr>
<tr>
<td>
<code>logConfig</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>networkmode</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.NetworkMode">
NetworkMode
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use hostnetwork or not</p>
</td>
</tr>
<tr>
<td>
<code>cleanCachePolicy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.CleanCachePolicy">
CleanCachePolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CleanCachePolicy defines cleanCache Policy</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Volumes is the list of Kubernetes volumes that can be mounted by the jindo runtime components and/or fuses.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.JobProcessor">JobProcessor
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.Processor">Processor</a>)
</p>
<p>
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
<code>podSpec</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#podspec-v1-core">
Kubernetes core/v1.PodSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PodSpec defines Pod specification of the DataProcess job.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.JuiceFSCompTemplateSpec">JuiceFSCompTemplateSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.JuiceFSRuntimeSpec">JuiceFSRuntimeSpec</a>)
</p>
<p>
<p>JuiceFSCompTemplateSpec is a description of the JuiceFS components</p>
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
<code>ports</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#containerport-v1-core">
[]Kubernetes core/v1.ContainerPort
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Ports used by JuiceFS</p>
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
<p>Resources that will be requested by the JuiceFS component.</p>
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
<p>Options</p>
</td>
</tr>
<tr>
<td>
<code>env</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#envvar-v1-core">
[]Kubernetes core/v1.EnvVar
</a>
</em>
</td>
<td>
<p>Environment variables that will be used by JuiceFS component.</p>
</td>
</tr>
<tr>
<td>
<code>enabled</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Enabled or Disabled for the components.</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeSelector is a selector</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>VolumeMounts specifies the volumes listed in &ldquo;.spec.volumes&rdquo; to mount into runtime component&rsquo;s filesystem.</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PodMetadata defines labels and annotations that will be propagated to JuiceFs&rsquo;s pods.</p>
</td>
</tr>
<tr>
<td>
<code>networkMode</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.NetworkMode">
NetworkMode
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use hostnetwork or not</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.JuiceFSFuseSpec">JuiceFSFuseSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.JuiceFSRuntimeSpec">JuiceFSRuntimeSpec</a>)
</p>
<p>
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
<p>Image for JuiceFS fuse</p>
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
<p>Image for JuiceFS fuse</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#envvar-v1-core">
[]Kubernetes core/v1.EnvVar
</a>
</em>
</td>
<td>
<p>Environment variables that will be used by JuiceFS Fuse</p>
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
<p>Resources that will be requested by JuiceFS Fuse.</p>
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
<p>Options mount options that fuse pod will use</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeSelector is a selector which must be true for the fuse client to fit on a node,
this option only effect when global is enabled</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>VolumeMounts specifies the volumes listed in &ldquo;.spec.volumes&rdquo; to mount into runtime component&rsquo;s filesystem.</p>
</td>
</tr>
<tr>
<td>
<code>cleanPolicy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.FuseCleanPolicy">
FuseCleanPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CleanPolicy decides when to clean Juicefs Fuse pods.
Currently Fluid supports two policies: OnDemand and OnRuntimeDeleted
OnDemand cleans fuse pod once th fuse pod on some node is not needed
OnRuntimeDeleted cleans fuse pod only when the cache runtime is deleted
Defaults to OnDemand</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PodMetadata defines labels and annotations that will be propagated to JuiceFs&rsquo;s pods.</p>
</td>
</tr>
<tr>
<td>
<code>networkMode</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.NetworkMode">
NetworkMode
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use hostnetwork or not</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.JuiceFSRuntimeSpec">JuiceFSRuntimeSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.JuiceFSRuntime">JuiceFSRuntime</a>)
</p>
<p>
<p>JuiceFSRuntimeSpec defines the desired state of JuiceFSRuntime</p>
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
<code>juicefsVersion</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.VersionSpec">
VersionSpec
</a>
</em>
</td>
<td>
<p>The version information that instructs fluid to orchestrate a particular version of JuiceFS.</p>
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
<code>master</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.JuiceFSCompTemplateSpec">
JuiceFSCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of JuiceFS master</p>
</td>
</tr>
<tr>
<td>
<code>worker</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.JuiceFSCompTemplateSpec">
JuiceFSCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of JuiceFS worker</p>
</td>
</tr>
<tr>
<td>
<code>jobWorker</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.JuiceFSCompTemplateSpec">
JuiceFSCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of JuiceFS job Worker</p>
</td>
</tr>
<tr>
<td>
<code>fuse</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.JuiceFSFuseSpec">
JuiceFSFuseSpec
</a>
</em>
</td>
<td>
<p>Desired state for JuiceFS Fuse</p>
</td>
</tr>
<tr>
<td>
<code>tieredstore</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.TieredStore">
TieredStore
</a>
</em>
</td>
<td>
<p>Tiered storage used by JuiceFS</p>
</td>
</tr>
<tr>
<td>
<code>configs</code></br>
<em>
string
</em>
</td>
<td>
<p>Configs of JuiceFS</p>
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
<p>Manage the user to run Juicefs Runtime</p>
</td>
</tr>
<tr>
<td>
<code>disablePrometheus</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Disable monitoring for JuiceFS Runtime
Prometheus is enabled by default</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Volumes is the list of Kubernetes volumes that can be mounted by the alluxio runtime components and/or fuses.</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PodMetadata defines labels and annotations that will be propagated to JuiceFs&rsquo;s pods.</p>
</td>
</tr>
<tr>
<td>
<code>cleanCachePolicy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.CleanCachePolicy">
CleanCachePolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CleanCachePolicy defines cleanCache Policy</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.Level">Level
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.TieredStore">TieredStore</a>)
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
<code>volumeType</code></br>
<em>
common.VolumeType
</em>
</td>
<td>
<em>(Optional)</em>
<p>VolumeType is the volume type of the tier. Should be one of the three types: <code>hostPath</code>, <code>emptyDir</code> and <code>volumeTemplate</code>.
If not set, defaults to hostPath.</p>
</td>
</tr>
<tr>
<td>
<code>volumeSource</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.VolumeSource">
VolumeSource
</a>
</em>
</td>
<td>
<p>VolumeSource is the volume source of the tier. It follows the form of corev1.VolumeSource.
For now, users should only specify VolumeSource when VolumeType is set to emptyDir.</p>
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
<p>File paths to be used for the tier. Multiple paths are supported.
Multiple paths should be separated with comma. For example: &ldquo;/mnt/cache1,/mnt/cache2&rdquo;.</p>
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
<em>(Optional)</em>
<p>Quota for the whole tier. (e.g. 100Gi)
Please note that if there&rsquo;re multiple paths used for this tierstore,
the quota will be equally divided into these paths. If you&rsquo;d like to
set quota for each, path, see QuotaList for more information.</p>
</td>
</tr>
<tr>
<td>
<code>quotaList</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>QuotaList are quotas used to set quota on multiple paths. Quotas should be separated with comma.
Quotas in this list will be set to paths with the same order in Path.
For example, with Path defined with &ldquo;/mnt/cache1,/mnt/cache2&rdquo; and QuotaList set to &ldquo;100Gi, 50Gi&rdquo;,
then we get 100GiB cache storage under &ldquo;/mnt/cache1&rdquo; and 50GiB under &ldquo;/mnt/cache2&rdquo;.
Also note that num of quotas must be consistent with the num of paths defined in Path.</p>
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
<h3 id="data.fluid.io/v1alpha1.MasterSpec">MasterSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.VineyardRuntimeSpec">VineyardRuntimeSpec</a>)
</p>
<p>
<p>MasterSpec defines the configurations for Vineyard Master component
which is also regarded as the Etcd component in Vineyard.
For more info about Vineyard, refer to <a href="https://v6d.io/">Vineyard</a></p>
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
<code>VineyardCompTemplateSpec</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.VineyardCompTemplateSpec">
VineyardCompTemplateSpec
</a>
</em>
</td>
<td>
<p>
(Members of <code>VineyardCompTemplateSpec</code> are embedded into this type.)
</p>
<em>(Optional)</em>
<p>The component configurations for Vineyard Master</p>
</td>
</tr>
<tr>
<td>
<code>endpoint</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.ExternalEndpointSpec">
ExternalEndpointSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ExternalEndpoint defines the configurations for external etcd cluster
Default is not set
If set, the Vineyard Master component will not be deployed,
which means the Vineyard Worker component will use an external Etcd cluster.
E,g.
endpoint:
uri: &ldquo;etcd-svc.etcd-namespace.svc.cluster.local:2379&rdquo;
encryptOptions:
- name: access-key
valueFrom:
secretKeyRef:
name: etcd-secret
key: accesskey</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.Metadata">Metadata
</h3>
<p>
<p>Metadata defines subgroup properties of metav1.ObjectMeta</p>
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
<code>PodMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<p>
(Members of <code>PodMetadata</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>selector</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#groupkind-v1-meta">
Kubernetes meta/v1.GroupKind
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.MetadataSyncPolicy">MetadataSyncPolicy
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.RuntimeManagement">RuntimeManagement</a>)
</p>
<p>
<p>MetadataSyncPolicy defines policies when syncing metadata</p>
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
<code>autoSync</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>AutoSync enables automatic metadata sync when setting up a runtime. If not set, it defaults to true.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.Mount">Mount
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DatasetSpec">DatasetSpec</a>, 
<a href="#data.fluid.io/v1alpha1.DatasetStatus">DatasetStatus</a>, 
<a href="#data.fluid.io/v1alpha1.RuntimeStatus">RuntimeStatus</a>)
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
<em>(Optional)</em>
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
<tr>
<td>
<code>encryptOptions</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.EncryptOption">
[]EncryptOption
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The secret information</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.NetworkMode">NetworkMode
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.AlluxioCompTemplateSpec">AlluxioCompTemplateSpec</a>, 
<a href="#data.fluid.io/v1alpha1.AlluxioFuseSpec">AlluxioFuseSpec</a>, 
<a href="#data.fluid.io/v1alpha1.EFCCompTemplateSpec">EFCCompTemplateSpec</a>, 
<a href="#data.fluid.io/v1alpha1.EFCFuseSpec">EFCFuseSpec</a>, 
<a href="#data.fluid.io/v1alpha1.JindoRuntimeSpec">JindoRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.JuiceFSCompTemplateSpec">JuiceFSCompTemplateSpec</a>, 
<a href="#data.fluid.io/v1alpha1.JuiceFSFuseSpec">JuiceFSFuseSpec</a>, 
<a href="#data.fluid.io/v1alpha1.ThinCompTemplateSpec">ThinCompTemplateSpec</a>, 
<a href="#data.fluid.io/v1alpha1.ThinFuseSpec">ThinFuseSpec</a>, 
<a href="#data.fluid.io/v1alpha1.VineyardClientSocketSpec">VineyardClientSocketSpec</a>, 
<a href="#data.fluid.io/v1alpha1.VineyardCompTemplateSpec">VineyardCompTemplateSpec</a>)
</p>
<p>
</p>
<h3 id="data.fluid.io/v1alpha1.NodePublishSecretPolicy">NodePublishSecretPolicy
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.ThinRuntimeProfileSpec">ThinRuntimeProfileSpec</a>)
</p>
<p>
</p>
<h3 id="data.fluid.io/v1alpha1.OSAdvise">OSAdvise
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.EFCRuntimeSpec">EFCRuntimeSpec</a>)
</p>
<p>
<p>OSAdvise is a description of choices to have optimization on specific operating system</p>
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
<code>osVersion</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specific operating system version that can have optimization.</p>
</td>
</tr>
<tr>
<td>
<code>enabled</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Enable operating system optimization
not enabled by default.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.OperationRef">OperationRef
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DataBackupSpec">DataBackupSpec</a>, 
<a href="#data.fluid.io/v1alpha1.DataLoadSpec">DataLoadSpec</a>, 
<a href="#data.fluid.io/v1alpha1.DataMigrateSpec">DataMigrateSpec</a>, 
<a href="#data.fluid.io/v1alpha1.DataProcessSpec">DataProcessSpec</a>)
</p>
<p>
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
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>API version of the referent operation</p>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
<em>
string
</em>
</td>
<td>
<p>Kind specifies the type of the referent operation</p>
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
<p>Name specifies the name of the referent operation</p>
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
<em>(Optional)</em>
<p>Namespace specifies the namespace of the referent operation.</p>
</td>
</tr>
<tr>
<td>
<code>affinityStrategy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.AffinityStrategy">
AffinityStrategy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AffinityStrategy specifies the pod affinity strategy with the referent operation.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.OperationStatus">OperationStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DataBackup">DataBackup</a>, 
<a href="#data.fluid.io/v1alpha1.DataLoad">DataLoad</a>, 
<a href="#data.fluid.io/v1alpha1.DataMigrate">DataMigrate</a>, 
<a href="#data.fluid.io/v1alpha1.DataProcess">DataProcess</a>)
</p>
<p>
<p>OperationStatus defines the observed state of operation</p>
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
common.Phase
</em>
</td>
<td>
<p>Phase describes current phase of operation</p>
</td>
</tr>
<tr>
<td>
<code>duration</code></br>
<em>
string
</em>
</td>
<td>
<p>Duration tell user how much time was spent to operation</p>
</td>
</tr>
<tr>
<td>
<code>conditions</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.Condition">
[]Condition
</a>
</em>
</td>
<td>
<p>Conditions consists of transition information on operation&rsquo;s Phase</p>
</td>
</tr>
<tr>
<td>
<code>infos</code></br>
<em>
map[string]string
</em>
</td>
<td>
<p>Infos operation customized name-value</p>
</td>
</tr>
<tr>
<td>
<code>lastScheduleTime</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>LastScheduleTime is the last time the cron operation was scheduled</p>
</td>
</tr>
<tr>
<td>
<code>lastSuccessfulTime</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>LastSuccessfulTime is the last time the cron operation successfully completed</p>
</td>
</tr>
<tr>
<td>
<code>waitingFor</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.WaitingStatus">
WaitingStatus
</a>
</em>
</td>
<td>
<p>WaitingStatus stores information about waiting operation.</p>
</td>
</tr>
<tr>
<td>
<code>nodeAffinity</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#nodeaffinity-v1-core">
Kubernetes core/v1.NodeAffinity
</a>
</em>
</td>
<td>
<p>NodeAffinity records the node affinity for operation pods</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.OperationType">OperationType
(<code>string</code> alias)</p></h3>
<p>
</p>
<h3 id="data.fluid.io/v1alpha1.PlacementMode">PlacementMode
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DatasetSpec">DatasetSpec</a>)
</p>
<p>
</p>
<h3 id="data.fluid.io/v1alpha1.PodMetadata">PodMetadata
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.AlluxioCompTemplateSpec">AlluxioCompTemplateSpec</a>, 
<a href="#data.fluid.io/v1alpha1.AlluxioFuseSpec">AlluxioFuseSpec</a>, 
<a href="#data.fluid.io/v1alpha1.AlluxioRuntimeSpec">AlluxioRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.DataLoadSpec">DataLoadSpec</a>, 
<a href="#data.fluid.io/v1alpha1.DataMigrateSpec">DataMigrateSpec</a>, 
<a href="#data.fluid.io/v1alpha1.EFCCompTemplateSpec">EFCCompTemplateSpec</a>, 
<a href="#data.fluid.io/v1alpha1.EFCFuseSpec">EFCFuseSpec</a>, 
<a href="#data.fluid.io/v1alpha1.EFCRuntimeSpec">EFCRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.JindoCompTemplateSpec">JindoCompTemplateSpec</a>, 
<a href="#data.fluid.io/v1alpha1.JindoFuseSpec">JindoFuseSpec</a>, 
<a href="#data.fluid.io/v1alpha1.JindoRuntimeSpec">JindoRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.JuiceFSCompTemplateSpec">JuiceFSCompTemplateSpec</a>, 
<a href="#data.fluid.io/v1alpha1.JuiceFSFuseSpec">JuiceFSFuseSpec</a>, 
<a href="#data.fluid.io/v1alpha1.JuiceFSRuntimeSpec">JuiceFSRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.Metadata">Metadata</a>, 
<a href="#data.fluid.io/v1alpha1.Processor">Processor</a>, 
<a href="#data.fluid.io/v1alpha1.VineyardClientSocketSpec">VineyardClientSocketSpec</a>, 
<a href="#data.fluid.io/v1alpha1.VineyardCompTemplateSpec">VineyardCompTemplateSpec</a>, 
<a href="#data.fluid.io/v1alpha1.VineyardRuntimeSpec">VineyardRuntimeSpec</a>)
</p>
<p>
<p>PodMetadata defines subgroup properties of metav1.ObjectMeta</p>
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
<code>labels</code></br>
<em>
map[string]string
</em>
</td>
<td>
<p>Labels are labels of pod specification</p>
</td>
</tr>
<tr>
<td>
<code>annotations</code></br>
<em>
map[string]string
</em>
</td>
<td>
<p>Annotations are annotations of pod specification</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.Policy">Policy
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DataLoadSpec">DataLoadSpec</a>, 
<a href="#data.fluid.io/v1alpha1.DataMigrateSpec">DataMigrateSpec</a>)
</p>
<p>
</p>
<h3 id="data.fluid.io/v1alpha1.Prefer">Prefer
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.AffinityStrategy">AffinityStrategy</a>)
</p>
<p>
<p>Prefer defines the label key and weight for generating a PreferredSchedulingTerm.</p>
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
</td>
</tr>
<tr>
<td>
<code>weight</code></br>
<em>
int32
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.Processor">Processor
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DataProcessSpec">DataProcessSpec</a>)
</p>
<p>
<p>Processor defines the actual processor for DataProcess. Processor can be either of a Job or a Shell script.</p>
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
<code>serviceAccountName</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ServiceAccountName defiens the serviceAccountName of the container</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PodMetadata defines labels and annotations on the processor pod.</p>
</td>
</tr>
<tr>
<td>
<code>job</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.JobProcessor">
JobProcessor
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Job represents a processor which runs DataProcess as a job.</p>
</td>
</tr>
<tr>
<td>
<code>script</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.ScriptProcessor">
ScriptProcessor
</a>
</em>
</td>
<td>
<p>Shell represents a processor which executes shell script</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.Require">Require
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.AffinityStrategy">AffinityStrategy</a>)
</p>
<p>
<p>Require defines the label key for generating a NodeSelectorTerm.</p>
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
<tr>
<td>
<code>masterReplicas</code></br>
<em>
int32
</em>
</td>
<td>
<p>Runtime master replicas</p>
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
<h3 id="data.fluid.io/v1alpha1.RuntimeManagement">RuntimeManagement
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.AlluxioRuntimeSpec">AlluxioRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.ThinRuntimeSpec">ThinRuntimeSpec</a>)
</p>
<p>
<p>RuntimeManagement defines suggestions for runtime controllers to manage the runtime</p>
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
<code>cleanCachePolicy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.CleanCachePolicy">
CleanCachePolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CleanCachePolicy defines the policy of cleaning cache when shutting down the runtime</p>
</td>
</tr>
<tr>
<td>
<code>metadataSyncPolicy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.MetadataSyncPolicy">
MetadataSyncPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>MetadataSyncPolicy defines the policy of syncing metadata when setting up the runtime. If not set,</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.RuntimeStatus">RuntimeStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.AlluxioRuntime">AlluxioRuntime</a>, 
<a href="#data.fluid.io/v1alpha1.EFCRuntime">EFCRuntime</a>, 
<a href="#data.fluid.io/v1alpha1.GooseFSRuntime">GooseFSRuntime</a>, 
<a href="#data.fluid.io/v1alpha1.JindoRuntime">JindoRuntime</a>, 
<a href="#data.fluid.io/v1alpha1.JuiceFSRuntime">JuiceFSRuntime</a>, 
<a href="#data.fluid.io/v1alpha1.VineyardRuntime">VineyardRuntime</a>, 
<a href="#data.fluid.io/v1alpha1.ThinRuntime">ThinRuntime</a>)
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
<code>setupDuration</code></br>
<em>
string
</em>
</td>
<td>
<p>Duration tell user how much time was spent to setup the runtime</p>
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
<tr>
<td>
<code>selector</code></br>
<em>
string
</em>
</td>
<td>
<p>Selector is used for auto-scaling</p>
</td>
</tr>
<tr>
<td>
<code>apiGateway</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.APIGatewayStatus">
APIGatewayStatus
</a>
</em>
</td>
<td>
<p>APIGatewayStatus represents rest api gateway status</p>
</td>
</tr>
<tr>
<td>
<code>mountTime</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>MountTime represents time last mount happened
if Mounttime is earlier than master starting time, remount will be required</p>
</td>
</tr>
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
<p>MountPoints represents the mount points specified in the bounded dataset</p>
</td>
</tr>
<tr>
<td>
<code>cacheAffinity</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#nodeaffinity-v1-core">
Kubernetes core/v1.NodeAffinity
</a>
</em>
</td>
<td>
<p>CacheAffinity represents the runtime worker pods node affinity including node selector</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.ScriptProcessor">ScriptProcessor
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.Processor">Processor</a>)
</p>
<p>
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
<code>VersionSpec</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.VersionSpec">
VersionSpec
</a>
</em>
</td>
<td>
<p>
(Members of <code>VersionSpec</code> are embedded into this type.)
</p>
<p>VersionSpec specifies the container&rsquo;s image info.</p>
</td>
</tr>
<tr>
<td>
<code>restartPolicy</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#restartpolicy-v1-core">
Kubernetes core/v1.RestartPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RestartPolicy specifies the processor job&rsquo;s restart policy. Only &ldquo;Never&rdquo;, &ldquo;OnFailure&rdquo; is allowed.</p>
</td>
</tr>
<tr>
<td>
<code>command</code></br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Entrypoint command for ScriptProcessor.</p>
</td>
</tr>
<tr>
<td>
<code>source</code></br>
<em>
string
</em>
</td>
<td>
<p>Script source for ScriptProcessor</p>
</td>
</tr>
<tr>
<td>
<code>env</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#envvar-v1-core">
[]Kubernetes core/v1.EnvVar
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of environment variables to set in the container.</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Pod volumes to mount into the container&rsquo;s filesystem.</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of volumes that can be mounted by containers belonging to the pod.</p>
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
<p>Resources that will be requested by the DataProcess job. <br></p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.SecretKeySelector">SecretKeySelector
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.EncryptOptionSource">EncryptOptionSource</a>)
</p>
<p>
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
<p>The name of required secret</p>
</td>
</tr>
<tr>
<td>
<code>key</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The required key in the secret</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.TargetDataset">TargetDataset
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DataLoadSpec">DataLoadSpec</a>, 
<a href="#data.fluid.io/v1alpha1.TargetDatasetWithMountPath">TargetDatasetWithMountPath</a>)
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
<h3 id="data.fluid.io/v1alpha1.TargetDatasetWithMountPath">TargetDatasetWithMountPath
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.DataProcessSpec">DataProcessSpec</a>)
</p>
<p>
<p>TargetDataset defines which dataset will be processed by DataProcess.
Under the hood, the dataset&rsquo;s pvc will be mounted to the given mountPath of the DataProcess&rsquo;s containers.</p>
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
<code>TargetDataset</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.TargetDataset">
TargetDataset
</a>
</em>
</td>
<td>
<p>
(Members of <code>TargetDataset</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>mountPath</code></br>
<em>
string
</em>
</td>
<td>
<p>MountPath defines where the Dataset should be mounted in DataProcess&rsquo;s containers.</p>
</td>
</tr>
<tr>
<td>
<code>subPath</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>SubPath defines subpath of the target dataset to mount.</p>
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
<h3 id="data.fluid.io/v1alpha1.ThinCompTemplateSpec">ThinCompTemplateSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.ThinRuntimeProfileSpec">ThinRuntimeProfileSpec</a>, 
<a href="#data.fluid.io/v1alpha1.ThinRuntimeSpec">ThinRuntimeSpec</a>)
</p>
<p>
<p>ThinCompTemplateSpec is a description of the thinRuntime components</p>
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
<p>Image for thinRuntime fuse</p>
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
<p>Image for thinRuntime fuse</p>
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
<code>ports</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#containerport-v1-core">
[]Kubernetes core/v1.ContainerPort
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Ports used thinRuntime</p>
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
<p>Resources that will be requested by thinRuntime component.</p>
</td>
</tr>
<tr>
<td>
<code>env</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#envvar-v1-core">
[]Kubernetes core/v1.EnvVar
</a>
</em>
</td>
<td>
<p>Environment variables that will be used by thinRuntime component.</p>
</td>
</tr>
<tr>
<td>
<code>enabled</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Enabled or Disabled for the components.</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeSelector is a selector</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>VolumeMounts specifies the volumes listed in &ldquo;.spec.volumes&rdquo; to mount into runtime component&rsquo;s filesystem.</p>
</td>
</tr>
<tr>
<td>
<code>livenessProbe</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#probe-v1-core">
Kubernetes core/v1.Probe
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>livenessProbe of thin fuse pod</p>
</td>
</tr>
<tr>
<td>
<code>readinessProbe</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#probe-v1-core">
Kubernetes core/v1.Probe
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>readinessProbe of thin fuse pod</p>
</td>
</tr>
<tr>
<td>
<code>networkMode</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.NetworkMode">
NetworkMode
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use hostnetwork or not</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.ThinFuseSpec">ThinFuseSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.ThinRuntimeProfileSpec">ThinRuntimeProfileSpec</a>, 
<a href="#data.fluid.io/v1alpha1.ThinRuntimeSpec">ThinRuntimeSpec</a>)
</p>
<p>
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
<p>Image for thinRuntime fuse</p>
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
<p>Image for thinRuntime fuse</p>
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
<code>ports</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#containerport-v1-core">
[]Kubernetes core/v1.ContainerPort
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Ports used thinRuntime</p>
</td>
</tr>
<tr>
<td>
<code>env</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#envvar-v1-core">
[]Kubernetes core/v1.EnvVar
</a>
</em>
</td>
<td>
<p>Environment variables that will be used by thinRuntime Fuse</p>
</td>
</tr>
<tr>
<td>
<code>command</code></br>
<em>
[]string
</em>
</td>
<td>
<p>Command that will be passed to thinRuntime Fuse</p>
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
<p>Arguments that will be passed to thinRuntime Fuse</p>
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
<p>Options configurable options of FUSE client, performance parameters usually.
will be merged with Dataset.spec.mounts.options into fuse pod.</p>
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
<p>Resources that will be requested by thinRuntime Fuse.</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeSelector is a selector which must be true for the fuse client to fit on a node,
this option only effect when global is enabled</p>
</td>
</tr>
<tr>
<td>
<code>cleanPolicy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.FuseCleanPolicy">
FuseCleanPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CleanPolicy decides when to clean thinRuntime Fuse pods.
Currently Fluid supports two policies: OnDemand and OnRuntimeDeleted
OnDemand cleans fuse pod once the fuse pod on some node is not needed
OnRuntimeDeleted cleans fuse pod only when the cache runtime is deleted
Defaults to OnDemand</p>
</td>
</tr>
<tr>
<td>
<code>networkMode</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.NetworkMode">
NetworkMode
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use hostnetwork or not</p>
</td>
</tr>
<tr>
<td>
<code>livenessProbe</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#probe-v1-core">
Kubernetes core/v1.Probe
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>livenessProbe of thin fuse pod</p>
</td>
</tr>
<tr>
<td>
<code>readinessProbe</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#probe-v1-core">
Kubernetes core/v1.Probe
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>readinessProbe of thin fuse pod</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>VolumeMounts specifies the volumes listed in &ldquo;.spec.volumes&rdquo; to mount into the thinruntime component&rsquo;s filesystem.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.ThinRuntime">ThinRuntime
</h3>
<p>
<p>ThinRuntime is the Schema for the thinruntimes API</p>
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
<a href="#data.fluid.io/v1alpha1.ThinRuntimeSpec">
ThinRuntimeSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>profileName</code></br>
<em>
string
</em>
</td>
<td>
<p>The specific runtime profile name, empty value is used for handling datasets which mount another dataset</p>
</td>
</tr>
<tr>
<td>
<code>worker</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.ThinCompTemplateSpec">
ThinCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of worker</p>
</td>
</tr>
<tr>
<td>
<code>fuse</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.ThinFuseSpec">
ThinFuseSpec
</a>
</em>
</td>
<td>
<p>The component spec of thinRuntime</p>
</td>
</tr>
<tr>
<td>
<code>tieredstore</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.TieredStore">
TieredStore
</a>
</em>
</td>
<td>
<p>Tiered storage</p>
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
<p>Manage the user to run Runtime</p>
</td>
</tr>
<tr>
<td>
<code>disablePrometheus</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Disable monitoring for Runtime
Prometheus is enabled by default</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Volumes is the list of Kubernetes volumes that can be mounted by runtime components and/or fuses.</p>
</td>
</tr>
<tr>
<td>
<code>management</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.RuntimeManagement">
RuntimeManagement
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RuntimeManagement defines policies when managing the runtime</p>
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
<h3 id="data.fluid.io/v1alpha1.ThinRuntimeProfile">ThinRuntimeProfile
</h3>
<p>
<p>ThinRuntimeProfile is the Schema for the ThinRuntimeProfiles API</p>
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
<a href="#data.fluid.io/v1alpha1.ThinRuntimeProfileSpec">
ThinRuntimeProfileSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>fileSystemType</code></br>
<em>
string
</em>
</td>
<td>
<p>file system of thinRuntime</p>
</td>
</tr>
<tr>
<td>
<code>worker</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.ThinCompTemplateSpec">
ThinCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of worker</p>
</td>
</tr>
<tr>
<td>
<code>fuse</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.ThinFuseSpec">
ThinFuseSpec
</a>
</em>
</td>
<td>
<p>The component spec of thinRuntime</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Volumes is the list of Kubernetes volumes that can be mounted by runtime components and/or fuses.</p>
</td>
</tr>
<tr>
<td>
<code>nodePublishSecretPolicy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.NodePublishSecretPolicy">
NodePublishSecretPolicy
</a>
</em>
</td>
<td>
<p>NodePublishSecretPolicy describes the policy to decide which to do with node publish secret when mounting an existing persistent volume.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.ThinRuntimeProfileStatus">
ThinRuntimeProfileStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.ThinRuntimeProfileSpec">ThinRuntimeProfileSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.ThinRuntimeProfile">ThinRuntimeProfile</a>)
</p>
<p>
<p>ThinRuntimeProfileSpec defines the desired state of ThinRuntimeProfile</p>
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
<code>fileSystemType</code></br>
<em>
string
</em>
</td>
<td>
<p>file system of thinRuntime</p>
</td>
</tr>
<tr>
<td>
<code>worker</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.ThinCompTemplateSpec">
ThinCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of worker</p>
</td>
</tr>
<tr>
<td>
<code>fuse</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.ThinFuseSpec">
ThinFuseSpec
</a>
</em>
</td>
<td>
<p>The component spec of thinRuntime</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Volumes is the list of Kubernetes volumes that can be mounted by runtime components and/or fuses.</p>
</td>
</tr>
<tr>
<td>
<code>nodePublishSecretPolicy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.NodePublishSecretPolicy">
NodePublishSecretPolicy
</a>
</em>
</td>
<td>
<p>NodePublishSecretPolicy describes the policy to decide which to do with node publish secret when mounting an existing persistent volume.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.ThinRuntimeProfileStatus">ThinRuntimeProfileStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.ThinRuntimeProfile">ThinRuntimeProfile</a>)
</p>
<p>
<p>ThinRuntimeProfileStatus defines the observed state of ThinRuntimeProfile</p>
</p>
<h3 id="data.fluid.io/v1alpha1.ThinRuntimeSpec">ThinRuntimeSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.ThinRuntime">ThinRuntime</a>)
</p>
<p>
<p>ThinRuntimeSpec defines the desired state of ThinRuntime</p>
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
<code>profileName</code></br>
<em>
string
</em>
</td>
<td>
<p>The specific runtime profile name, empty value is used for handling datasets which mount another dataset</p>
</td>
</tr>
<tr>
<td>
<code>worker</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.ThinCompTemplateSpec">
ThinCompTemplateSpec
</a>
</em>
</td>
<td>
<p>The component spec of worker</p>
</td>
</tr>
<tr>
<td>
<code>fuse</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.ThinFuseSpec">
ThinFuseSpec
</a>
</em>
</td>
<td>
<p>The component spec of thinRuntime</p>
</td>
</tr>
<tr>
<td>
<code>tieredstore</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.TieredStore">
TieredStore
</a>
</em>
</td>
<td>
<p>Tiered storage</p>
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
<p>Manage the user to run Runtime</p>
</td>
</tr>
<tr>
<td>
<code>disablePrometheus</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Disable monitoring for Runtime
Prometheus is enabled by default</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Volumes is the list of Kubernetes volumes that can be mounted by runtime components and/or fuses.</p>
</td>
</tr>
<tr>
<td>
<code>management</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.RuntimeManagement">
RuntimeManagement
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RuntimeManagement defines policies when managing the runtime</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.TieredStore">TieredStore
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.AlluxioRuntimeSpec">AlluxioRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.EFCRuntimeSpec">EFCRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.GooseFSRuntimeSpec">GooseFSRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.JindoRuntimeSpec">JindoRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.JuiceFSRuntimeSpec">JuiceFSRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.ThinRuntimeSpec">ThinRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.VineyardRuntimeSpec">VineyardRuntimeSpec</a>)
</p>
<p>
<p>TieredStore is a description of the tiered store</p>
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
<a href="#data.fluid.io/v1alpha1.DataBackupSpec">DataBackupSpec</a>, 
<a href="#data.fluid.io/v1alpha1.DatasetSpec">DatasetSpec</a>, 
<a href="#data.fluid.io/v1alpha1.GooseFSRuntimeSpec">GooseFSRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.JindoRuntimeSpec">JindoRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.JuiceFSRuntimeSpec">JuiceFSRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.ThinRuntimeSpec">ThinRuntimeSpec</a>)
</p>
<p>
<p>User explains the user and group to run a Container</p>
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
<a href="#data.fluid.io/v1alpha1.DataMigrateSpec">DataMigrateSpec</a>, 
<a href="#data.fluid.io/v1alpha1.EFCCompTemplateSpec">EFCCompTemplateSpec</a>, 
<a href="#data.fluid.io/v1alpha1.EFCFuseSpec">EFCFuseSpec</a>, 
<a href="#data.fluid.io/v1alpha1.GooseFSRuntimeSpec">GooseFSRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.InitFuseSpec">InitFuseSpec</a>, 
<a href="#data.fluid.io/v1alpha1.JindoRuntimeSpec">JindoRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.JuiceFSRuntimeSpec">JuiceFSRuntimeSpec</a>, 
<a href="#data.fluid.io/v1alpha1.ScriptProcessor">ScriptProcessor</a>)
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
<h3 id="data.fluid.io/v1alpha1.VineyardClientSocketSpec">VineyardClientSocketSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.VineyardRuntimeSpec">VineyardRuntimeSpec</a>)
</p>
<p>
<p>VineyardClientSocketSpec holds the configurations for vineyard client socket</p>
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
<em>(Optional)</em>
<p>Image for Vineyard Fuse
Default is <code>vineyardcloudnative/vineyard-fluid-fuse</code></p>
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
<em>(Optional)</em>
<p>Image Tag for Vineyard Fuse
Default is <code>v0.21.5</code></p>
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
<em>(Optional)</em>
<p>Image pull policy for Vineyard Fuse
Default is <code>IfNotPresent</code>
Available values are <code>Always</code>, <code>IfNotPresent</code>, <code>Never</code></p>
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
<em>(Optional)</em>
<p>Environment variables that will be used by Vineyard Fuse.
Default is not set.</p>
</td>
</tr>
<tr>
<td>
<code>cleanPolicy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.FuseCleanPolicy">
FuseCleanPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CleanPolicy decides when to clean Vineyard Fuse pods.
Currently Fluid supports two policies: OnDemand and OnRuntimeDeleted
OnDemand cleans fuse pod once th fuse pod on some node is not needed
OnRuntimeDeleted cleans fuse pod only when the cache runtime is deleted
Defaults to OnRuntimeDeleted</p>
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
<p>Resources contains the resource requirements and limits for the Vineyard Fuse.
Default is not set.</p>
</td>
</tr>
<tr>
<td>
<code>networkMode</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.NetworkMode">
NetworkMode
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use hostnetwork or not
Default is HostNetwork</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PodMetadata defines labels and annotations that will be propagated to Vineyard&rsquo;s pods.</p>
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
<p>Options for configuring vineyardd parameters.
Supported options are as follows.
reserve_memory: (Bool) Whether to reserving enough physical memory pages for vineyardd.
Default is true.
allocator: (String) The allocator used by vineyardd, could be &ldquo;dlmalloc&rdquo; or &ldquo;mimalloc&rdquo;.
Default is &ldquo;dlmalloc&rdquo;.
compression: (Bool) Compress before migration or spilling.
Default is true.
coredump: (Bool) Enable coredump core dump when been aborted.
Default is false.
meta_timeout: (Int) Timeout period before waiting the metadata service to be ready, in seconds
Default is 60.
etcd_endpoint: (String) The endpoint of etcd.
Default is same as the etcd endpoint of vineyard worker.
etcd_prefix: (String) Metadata path prefix in etcd.
Default is &ldquo;/vineyard&rdquo;.
size: (String) shared memory size for vineyardd.
1024M, 1024000, 1G, or 1Gi.
Default is &ldquo;0&rdquo;, which means no cache.
When the size is not set to &ldquo;0&rdquo;, it should be greater than the 2048 bytes(2K).
spill_path: (String) Path to spill temporary files, if not set, spilling will be disabled.
Default is &ldquo;&rdquo;.
spill_lower_rate: (Double) The lower rate of memory usage to trigger spilling.
Default is 0.3.
spill_upper_rate: (Double) The upper rate of memory usage to stop spilling.
Default is 0.8.
Default is as follows.
fuse:
options:
size: &ldquo;0&rdquo;
etcd_endpoint: &ldquo;http://{{Name}}-master-0.{{Name}}-master.{{Namespace}}:{{EtcdClientPort}}&rdquo;
etcd_prefix: &ldquo;/vineyard&rdquo;</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.VineyardCompTemplateSpec">VineyardCompTemplateSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.MasterSpec">MasterSpec</a>, 
<a href="#data.fluid.io/v1alpha1.VineyardRuntimeSpec">VineyardRuntimeSpec</a>)
</p>
<p>
<p>VineyardCompTemplateSpec is the common configurations for vineyard components including Master and Worker.</p>
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
<p>The replicas of Vineyard component.
If not specified, defaults to 1.
For worker, the replicas should not be greater than the number of nodes in the cluster</p>
</td>
</tr>
<tr>
<td>
<code>image</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The image of Vineyard component.
For Master, the default image is <code>registry.aliyuncs.com/vineyard/vineyardd</code>
For Worker, the default image is <code>registry.aliyuncs.com/vineyard/vineyardd</code>
The default container registry is <code>docker.io</code>, you can change it by setting the image field</p>
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
<em>(Optional)</em>
<p>The image tag of Vineyard component.
For Master, the default image tag is <code>v0.22.1</code>.
For Worker, the default image tag is <code>v0.22.1</code>.</p>
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
<em>(Optional)</em>
<p>The image pull policy of Vineyard component.
Default is <code>IfNotPresent</code>.</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeSelector is a selector to choose which nodes to launch the Vineyard component.
E,g. {&ldquo;disktype&rdquo;: &ldquo;ssd&rdquo;}</p>
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
<p>Ports used by Vineyard component.
For Master, the default client port is 2379 and peer port is 2380.
For Worker, the default rpc port is 9600 and the default exporter port is 9144.</p>
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
<em>(Optional)</em>
<p>Environment variables that will be used by Vineyard component.
For Master, refer to <a href="https://etcd.io/docs/v3.5/op-guide/configuration/">Etcd Configuration</a> for more info
Default is not set.</p>
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
<p>Configurable options for Vineyard component.
For Master, there is no configurable options.
For Worker, support the following options.</p>
<p>vineyardd.reserve.memory: (Bool) where to reserve memory for vineyardd
If set to true, the memory quota will be counted to the vineyardd rather than the application.
etcd.prefix: (String) the prefix of etcd key for vineyard objects
wait.etcd.timeout: (String) the timeout period before waiting the etcd to be ready, in seconds</p>
<p>Default value is as follows.</p>
<pre><code>vineyardd.reserve.memory: &quot;true&quot;
etcd.prefix: &quot;/vineyard&quot;
wait.etcd.timeout: &quot;120&quot;
</code></pre>
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
<p>Resources contains the resource requirements and limits for the Vineyard component.
Default is not set.
For Worker, when the options contains vineyardd.reserve.memory=true,
the resources.request.memory for worker should be greater than tieredstore.levels[0].quota(aka vineyardd shared memory)</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>VolumeMounts specifies the volumes listed in &ldquo;.spec.volumes&rdquo; to mount into the vineyard runtime component&rsquo;s filesystem.
It is useful for specifying a persistent storage.
Default is not set.</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PodMetadata defines labels and annotations that will be propagated to Vineyard&rsquo;s pods including Master and Worker.
Default is not set.</p>
</td>
</tr>
<tr>
<td>
<code>networkMode</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.NetworkMode">
NetworkMode
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Whether to use hostnetwork or not
Default is HostNetwork</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.VineyardRuntimeSpec">VineyardRuntimeSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.VineyardRuntime">VineyardRuntime</a>)
</p>
<p>
<p>VineyardRuntimeSpec defines the desired state of VineyardRuntime</p>
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
<code>master</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.MasterSpec">
MasterSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Master holds the configurations for Vineyard Master component
Represents the Etcd component in Vineyard</p>
</td>
</tr>
<tr>
<td>
<code>worker</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.VineyardCompTemplateSpec">
VineyardCompTemplateSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Worker holds the configurations for Vineyard Worker component
Represents the Vineyardd component in Vineyard</p>
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
<p>The replicas of the worker, need to be specified
If worker.replicas and the field are both specified, the field will be respected</p>
</td>
</tr>
<tr>
<td>
<code>fuse</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.VineyardClientSocketSpec">
VineyardClientSocketSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Fuse holds the configurations for Vineyard client socket.
Note that the &ldquo;Fuse&rdquo; here is kept just for API consistency, VineyardRuntime mount a socket file instead of a FUSE filesystem to make data cache available.
Applications can connect to the vineyard runtime components through IPC or RPC.
IPC is the default way to connect to vineyard runtime components, which is more efficient than RPC.
If the socket file is not mounted, the connection will fall back to RPC.</p>
</td>
</tr>
<tr>
<td>
<code>tieredstore</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.TieredStore">
TieredStore
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Tiered storage used by vineyardd
The MediumType can only be <code>MEM</code> and <code>SSD</code>
<code>MEM</code> actually represents the shared memory of vineyardd.
<code>SSD</code> represents the external storage of vineyardd.
Default is as follows.
tieredstore:
levels:
- level: 0
mediumtype: MEM
quota: 4Gi</p>
<p>Choose hostpath as the external storage of vineyardd.
tieredstore:
levels:
- level: 0
mediumtype: MEM
quota: 4Gi
high: &ldquo;0.8&rdquo;
low: &ldquo;0.3&rdquo;
- level: 1
mediumtype: SSD
quota: 10Gi
volumeType: Hostpath
path: /var/spill-path</p>
</td>
</tr>
<tr>
<td>
<code>disablePrometheus</code></br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Disable monitoring metrics for Vineyard Runtime
Default is false</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Volumes is the list of Kubernetes volumes that can be mounted by the vineyard components (Master and Worker).
Default is null.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.VineyardSockSpec">VineyardSockSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.VineyardRuntimeSpec">VineyardRuntimeSpec</a>)
</p>
<p>
<p>VineyardSockSpec holds the configurations for vineyard client socket</p>
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
<em>(Optional)</em>
<p>Image for Vineyard Fuse
Default is <code>registry.aliyuncs.com/vineyard/vineyard-fluid-fuse</code></p>
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
<em>(Optional)</em>
<p>Image Tag for Vineyard Fuse
Default is <code>v0.22.1</code></p>
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
<em>(Optional)</em>
<p>Image pull policy for Vineyard Fuse
Default is <code>IfNotPresent</code>
Available values are <code>Always</code>, <code>IfNotPresent</code>, <code>Never</code></p>
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
<em>(Optional)</em>
<p>Environment variables that will be used by Vineyard Fuse.
Default is not set.</p>
</td>
</tr>
<tr>
<td>
<code>cleanPolicy</code></br>
<em>
<a href="#data.fluid.io/v1alpha1.PodMetadata">
PodMetadata
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PodMetadata defines labels and annotations that will be propagated to Vineyard&rsquo;s pods.</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Volumes is the list of Kubernetes volumes that can be mounted by the vineyard components (Master and Worker).
Default is null.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.VolumeSource">VolumeSource
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.Level">Level</a>)
</p>
<p>
<p>VolumeSource defines volume source and volume claim template.</p>
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
<code>VolumeSource</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volumesource-v1-core">
Kubernetes core/v1.VolumeSource
</a>
</em>
</td>
<td>
<p>
(Members of <code>VolumeSource</code> are embedded into this type.)
</p>
</td>
</tr>
</tbody>
</table>
<h3 id="data.fluid.io/v1alpha1.WaitingStatus">WaitingStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#data.fluid.io/v1alpha1.OperationStatus">OperationStatus</a>)
</p>
<p>
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
<code>operationComplete</code></br>
<em>
bool
</em>
</td>
<td>
<p>OperationComplete indicates if the preceding operation is complete</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>f0f1ed0</code>.
</em></p>
