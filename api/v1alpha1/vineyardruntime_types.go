/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VineyardCompTemplateSpec is the common configurations for vineyard components including Master and Worker.
type VineyardCompTemplateSpec struct {
	// The replicas of Vineyard component.
	// If not specified, defaults to 1.
	// For worker, the replicas should not be greater than the number of nodes in the cluster
	// +kubebuilder:validation:Minimum=1
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// The image of Vineyard component.
	// For Master, the default image is `bitnami/etcd`
	// For Worker, the default image is `vineyardcloudnative/vineyardd`
	// The default container registry is `docker.io`, you can change it by setting the image field
	// +optional
	Image string `json:"image,omitempty"`

	// The image tag of Vineyard component.
	// For Master, the default image tag is `3.5.10`.
	// For Worker, the default image tag is `latest`.
	// +optional
	ImageTag string `json:"imageTag,omitempty"`

	// The image pull policy of Vineyard component.
	// Default is `IfNotPresent`.
	// +optional
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	// NodeSelector is a selector to choose which nodes to launch the Vineyard component.
	// E,g. {"disktype": "ssd"}
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Ports used by Vineyard component.
	// For Master, the default client port is 2379 and peer port is 2380.
	// For Worker, the default rpc port is 9600 and the default exporter port is 9144.
	// +optional
	Ports map[string]int `json:"ports,omitempty"`

	// Environment variables that will be used by Vineyard component.
	// For Master, refer to <a href="https://etcd.io/docs/v3.5/op-guide/configuration/">Etcd Configuration</a> for more info
	// Default is not set.
	// +optional
	Env map[string]string `json:"env,omitempty"`

	// Configurable options for Vineyard component.
	// For Master, there is no configurable options.
	// For Worker, support the following options.
	//
	//   vineyardd.reserve.memory: (Bool) where to reserve memory for vineyardd
	//                             If set to true, the memory quota will be counted to the vineyardd rather than the application.
	//   etcd.prefix: (String) the prefix of etcd key for vineyard objects
	//   wait.etcd.timeout: (String) the timeout period before waiting the etcd to be ready, in seconds
	//
	//
	//   Default value is as follows.
	//
	//     vineyardd.reserve.memory: "true"
	//     etcd.prefix: "/vineyard"
	//     wait.etcd.timeout: "120"
	//
	// +optional
	Options map[string]string `json:"options,omitempty"`

	// Resources contains the resource requirements and limits for the Vineyard component.
	// Default is not set.
	// For Worker, when the options contains vineyardd.reserve.memory=true,
	// the resources.request.memory for worker should be greater than tieredstore.levels[0].quota(aka vineyardd shared memory)
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// VolumeMounts specifies the volumes listed in ".spec.volumes" to mount into the vineyard runtime component's filesystem.
	// It is useful for specifying a persistent storage.
	// Default is not set.
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
}

// ExternalEndpointSpec defines the configurations for external etcd cluster
type ExternalEndpointSpec struct {
	// URI specifies the endpoint of external Etcd cluster
	// E,g. "etcd-svc.etcd-namespace.svc.cluster.local:2379"
	// Default is not set and use http protocol to connect to external etcd cluster
	// +optional
	URI string `json:"uri"`

	// encrypt info for accessing the external etcd cluster
	// +optional
	EncryptOptions []EncryptOption `json:"encryptOptions,omitempty"`

	// Configurable options for External Etcd cluster.
	// +optional
	Options map[string]string `json:"options,omitempty"`
}

// MasterSpec defines the configurations for Vineyard Master component
// which is also regarded as the Etcd component in Vineyard.
// For more info about Vineyard, refer to <a href="https://v6d.io/">Vineyard</a>
type MasterSpec struct {
	// The component configurations for Vineyard Master
	// +optional
	VineyardCompTemplateSpec `json:",inline"`

	// ExternalEndpoint defines the configurations for external etcd cluster
	// Default is not set
	// If set, the Vineyard Master component will not be deployed,
	// which means the Vineyard Worker component will use an external Etcd cluster.
	// E,g.
	//   endpoint:
	//     uri: "etcd-svc.etcd-namespace.svc.cluster.local:2379"
	//     encryptOptions:
	//       - name: access-key
	// 		   valueFrom:
	//           secretKeyRef:
	//             name: etcd-secret
	//			   key: accesskey
	// +optional
	ExternalEndpoint ExternalEndpointSpec `json:"endpoint,omitempty"`
}

// VineyardSockSpec holds the configurations for vineyard client socket
type VineyardSockSpec struct {
	// Image for Vineyard Fuse
	// Default is `vineyardcloudnative/vineyard-mount-socket`
	// +optional
	Image string `json:"image,omitempty"`

	// Image Tag for Vineyard Fuse
	// Default is `latest`
	// +optional
	ImageTag string `json:"imageTag,omitempty"`

	// Image pull policy for Vineyard Fuse
	// Default is `IfNotPresent`
	// Available values are `Always`, `IfNotPresent`, `Never`
	// +optional
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	// CleanPolicy decides when to clean Vineyard Fuse pods.
	// Currently Fluid supports two policies: OnDemand and OnRuntimeDeleted
	// OnDemand cleans fuse pod once th fuse pod on some node is not needed
	// OnRuntimeDeleted cleans fuse pod only when the cache runtime is deleted
	// Defaults to OnRuntimeDeleted
	// +optional
	CleanPolicy FuseCleanPolicy `json:"cleanPolicy,omitempty"`

	// Resources contains the resource requirements and limits for the Vineyard Fuse.
	// Default is not set.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// VineyardRuntimeSpec defines the desired state of VineyardRuntime
type VineyardRuntimeSpec struct {
	// Master holds the configurations for Vineyard Master component
	// Represents the Etcd component in Vineyard
	// +optional
	Master MasterSpec `json:"master,omitempty"`

	// Worker holds the configurations for Vineyard Worker component
	// Represents the Vineyardd component in Vineyard
	// +optional
	Worker VineyardCompTemplateSpec `json:"worker,omitempty"`

	// Fuse holds the configurations for Vineyard client socket.
	// Note that the "Fuse" here is kept just for API consistency, VineyardRuntime mount a socket file instead of a FUSE filesystem to make data cache available.
	// Applications can connect to the vineyard runtime components through IPC or RPC.
	// IPC is the default way to connect to vineyard runtime components, which is more efficient than RPC.
	// If the socket file is not mounted, the connection will fall back to RPC.
	// +optional
	Fuse VineyardSockSpec `json:"fuse,omitempty"`

	// Tiered storage used by vineyardd
	// The MediumType can only be `MEM` and `SSD`
	// `MEM` actually represents the shared memory of vineyardd.
	// `SSD` represents the external storage of vineyardd.
	// Default is as follows.
	//   tieredstore:
	//     levels:
	//     - level: 0
	//       mediumtype: MEM
	//       quota: 4Gi
	//
	// Choose hostpath as the external storage of vineyardd.
	//   tieredstore:
	//     levels:
	//	   - level: 0
	//       mediumtype: MEM
	//       quota: 4Gi
	//		 high: "0.8"
	//       low: "0.3"
	//     - level: 1
	//       mediumtype: SSD
	//       quota: 10Gi
	//       volumeType: Hostpath
	//       path: /var/spill-path
	// +optional
	TieredStore TieredStore `json:"tieredstore,omitempty"`

	// Disable monitoring metrics for Vineyard Runtime
	// Default is false
	// +optional
	DisablePrometheus bool `json:"disablePrometheus,omitempty"`

	// Volumes is the list of Kubernetes volumes that can be mounted by the vineyard components (Master and Worker).
	// Default is null.
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.currentWorkerNumberScheduled,selectorpath=.status.selector
// +kubebuilder:printcolumn:name="Ready Masters",type="integer",JSONPath=`.status.masterNumberReady`,priority=10
// +kubebuilder:printcolumn:name="Desired Masters",type="integer",JSONPath=`.status.desiredMasterNumberScheduled`,priority=10
// +kubebuilder:printcolumn:name="Master Phase",type="string",JSONPath=`.status.masterPhase`,priority=0
// +kubebuilder:printcolumn:name="Ready Workers",type="integer",JSONPath=`.status.workerNumberReady`,priority=10
// +kubebuilder:printcolumn:name="Desired Workers",type="integer",JSONPath=`.status.desiredWorkerNumberScheduled`,priority=10
// +kubebuilder:printcolumn:name="Worker Phase",type="string",JSONPath=`.status.workerPhase`,priority=0
// +kubebuilder:printcolumn:name="Ready Fuses",type="integer",JSONPath=`.status.fuseNumberReady`,priority=10
// +kubebuilder:printcolumn:name="Desired Fuses",type="integer",JSONPath=`.status.desiredFuseNumberScheduled`,priority=10
// +kubebuilder:printcolumn:name="Fuse Phase",type="string",JSONPath=`.status.fusePhase`,priority=0
// +kubebuilder:printcolumn:name="API Gateway",type="string",JSONPath=`.status.apiGateway.endpoint`,priority=10
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`,priority=0
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:resource:categories={fluid},shortName=v6d
// +genclient

// VineyardRuntime is the Schema for the VineyardRuntimes API
type VineyardRuntime struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VineyardRuntimeSpec `json:"spec,omitempty"`
	Status RuntimeStatus       `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced

// VineyardRuntimeList contains a list of VineyardRuntime
type VineyardRuntimeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VineyardRuntime `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VineyardRuntime{}, &VineyardRuntimeList{})
}

// Replicas gets the replicas of runtime worker
func (runtime *VineyardRuntime) Replicas() int32 {
	return runtime.Spec.Worker.Replicas
}

func (runtime *VineyardRuntime) GetStatus() *RuntimeStatus {
	return &runtime.Status
}
