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

// GooseFSCompTemplateSpec is a description of the GooseFS commponents
type GooseFSCompTemplateSpec struct {
	// Replicas is the desired number of replicas of the given template.
	// If unspecified, defaults to 1.
	// +kubebuilder:validation:Minimum=1
	// replicas is the min replicas of dataset in the cluster
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Options for JVM
	JvmOptions []string `json:"jvmOptions,omitempty"`

	// Configurable properties for the GOOSEFS component. <br>
	// Refer to <a href="https://cloud.tencent.com/document/product/436/56415">GOOSEFS Configuration Properties</a> for more info
	// +optional
	Properties map[string]string `json:"properties,omitempty"`

	// Ports used by GooseFS(e.g. rpc: 19998 for master)
	// +optional
	Ports map[string]int `json:"ports,omitempty"`

	// Resources that will be requested by the GooseFS component. <br>
	// <br>
	// Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources
	// already allocated to the pod.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Environment variables that will be used by GooseFS component. <br>
	Env map[string]string `json:"env,omitempty"`

	// Enabled or Disabled for the components. For now, only  API Gateway is enabled or disabled.
	// +optional
	Enabled bool `json:"enabled,omitempty"`

	// NodeSelector is a selector which must be true for the master to fit on a node
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// GooseFSFuseSpec is a description of the GooseFS Fuse
type GooseFSFuseSpec struct {

	// Image for GooseFS Fuse(e.g. goosefs/goosefs-fuse)
	Image string `json:"image,omitempty"`

	// Image Tag for GooseFS Fuse(e.g. v1.0.1)
	ImageTag string `json:"imageTag,omitempty"`

	// One of the three policies: `Always`, `IfNotPresent`, `Never`
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	// Options for JVM
	JvmOptions []string `json:"jvmOptions,omitempty"`

	// Configurable properties for the GOOSEFS component. <br>
	// Refer to <a href="https://cloud.tencent.com/document/product/436/56415">GOOSEFS Configuration Properties</a> for more info
	Properties map[string]string `json:"properties,omitempty"`

	// Environment variables that will be used by GooseFS Fuse
	Env map[string]string `json:"env,omitempty"`

	// Resources that will be requested by GooseFS Fuse. <br>
	// <br>
	// Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources
	// already allocated to the pod.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Arguments that will be passed to GooseFS Fuse
	Args []string `json:"args,omitempty"`

	// If the fuse client should be deployed in global mode,
	// otherwise the affinity should be considered
	// +optional
	Global bool `json:"global,omitempty"`

	// NodeSelector is a selector which must be true for the fuse client to fit on a node,
	// this option only effect when global is enabled
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// GooseFSRuntimeSpec defines the desired state of GooseFSRuntime
type GooseFSRuntimeSpec struct {
	// The version information that instructs fluid to orchestrate a particular version of GooseFS.
	GooseFSVersion VersionSpec `json:"goosefsVersion,omitempty"`

	// Desired state for GooseFS master
	Master GooseFSCompTemplateSpec `json:"master,omitempty"`

	// Desired state for GooseFS job master
	JobMaster GooseFSCompTemplateSpec `json:"jobMaster,omitempty"`

	// Desired state for GooseFS worker
	Worker GooseFSCompTemplateSpec `json:"worker,omitempty"`

	// Desired state for GooseFS job Worker
	JobWorker GooseFSCompTemplateSpec `json:"jobWorker,omitempty"`

	// Desired state for GooseFS API Gateway
	APIGateway GooseFSCompTemplateSpec `json:"apiGateway,omitempty"`

	// The spec of init users
	InitUsers InitUsersSpec `json:"initUsers,omitempty"`

	// Desired state for GooseFS Fuse
	Fuse GooseFSFuseSpec `json:"fuse,omitempty"`

	// Configurable properties for the GOOSEFS component. <br>
	// Refer to <a href="https://cloud.tencent.com/document/product/436/56415">GOOSEFS Configuration Properties</a> for more info
	Properties map[string]string `json:"properties,omitempty"`

	// Options for JVM
	JvmOptions []string `json:"jvmOptions,omitempty"`

	// Tiered storage used by GooseFS
	TieredStore TieredStore `json:"tieredstore,omitempty"`

	// Management strategies for the dataset to which the runtime is bound
	Data Data `json:"data,omitempty"`

	// The replicas of the worker, need to be specified
	Replicas int32 `json:"replicas,omitempty"`

	// Manage the user to run GooseFS Runtime
	// GooseFS support POSIX-ACL and Apache Ranger to manager authorization
	// TODO(chrisydxie@tencent.com) Support Apache Ranger.
	RunAs *User `json:"runAs,omitempty"`

	// Disable monitoring for GooseFS Runtime
	// Prometheus is enabled by default
	// +optional
	DisablePrometheus bool `json:"disablePrometheus,omitempty"`

	// Name of the configMap used to support HDFS configurations when using HDFS as GooseFS's UFS. The configMap
	// must be in the same namespace with the GooseFSRuntime. The configMap should contain user-specific HDFS conf files in it.
	// For now, only "hdfs-site.xml" and "core-site.xml" are supported. It must take the filename of the conf file as the key and content
	// of the file as the value.
	// +optional
	HadoopConfig string `json:"hadoopConfig,omitempty"`
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
// +genclient

// GooseFSRuntime is the Schema for the goosefsruntimes API
type GooseFSRuntime struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GooseFSRuntimeSpec `json:"spec,omitempty"`
	Status RuntimeStatus      `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GooseFSRuntimeList contains a list of GooseFSRuntime
type GooseFSRuntimeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GooseFSRuntime `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GooseFSRuntime{}, &GooseFSRuntimeList{})
}

// Replicas gets the replicas of runtime worker
func (runtime *GooseFSRuntime) Replicas() int32 {
	return runtime.Spec.Replicas
}
