/*
Copyright 2025 The Fluid Author.

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

// +k8s:deepcopy-gen=package
// +groupName=data.fluid.io
package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CacheRuntimeTopology defines the topology structure of CacheRuntime components
type CacheRuntimeTopology struct {
	// Master specifies the configuration for master component
	// +optional
	Master *CacheRuntimeTopologyComponentDefinition `json:"master,omitempty"`

	// Worker specifies the configuration for worker component
	// +optional
	Worker *CacheRuntimeTopologyComponentDefinition `json:"worker,omitempty"`

	// Client specifies the configuration for client component
	// +optional
	Client *CacheRuntimeTopologyComponentDefinition `json:"client,omitempty"`
}

// CacheRuntimeTopologyComponentDefinition defines the configuration for a CacheRuntime component
type CacheRuntimeTopologyComponentDefinition struct {
	// WorkloadType defines the default workload type of the component
	// +optional
	WorkloadType metav1.TypeMeta `json:"workloadType,omitempty"`

	// Options specifies additional options for the component
	// +optional
	Options map[string]string `json:"options,omitempty"`

	// PodTemplateSpec defines the pod template spec of the component
	// +optional
	PodTemplateSpec corev1.PodTemplateSpec `json:"podTemplateSpec,omitempty"`

	// Service defines the service configuration for the component
	// +optional
	Service CacheRuntimeComponentService `json:"service,omitempty"`

	// Dependencies defines the dependencies required by the component
	// +optional
	Dependencies Dependencies `json:"dependencies,omitempty"`
}

// Dependencies defines the dependencies required by a CacheRuntime component
type Dependencies struct {
	// EncryptOption defines the configuration for encrypt option secret mount
	// +optional
	EncryptOption *EncryptOptionConfig `json:"encryptOption,omitempty"`

	// ExtraResources defines the usage of extra resources
	// +optional
	ExtraResources *ExtraResources `json:"extraResources,omitempty"`
}

// CacheRuntimeComponentService defines the service configuration for a CacheRuntime component
type CacheRuntimeComponentService struct {
	// Headless specifies the headless service configuration
	// +optional
	Headless *CacheRuntimeComponentHeadlessService `json:"headless,omitempty"`
}

// CacheRuntimeComponentHeadlessService defines the headless service configuration for a CacheRuntime component
type CacheRuntimeComponentHeadlessService struct {
}

// EncryptOptionConfig defines the configuration for encrypt option
type EncryptOptionConfig struct {
}

// ExtraResources defines the extra resources configuration
type ExtraResources struct {
	// Configmaps defines the configmap templates which will be used to mount configmaps
	// +optional
	Configmaps []ExtraResourceConfigmapConfig `json:"configmaps,omitempty"`
}

// CacheRuntimeExtraResources defines the extra resources for CacheRuntime
type CacheRuntimeExtraResources struct {
	// Configmaps defines the configmap templates which will be created in runtime's namespace
	// +optional
	Configmaps []CacheRuntimeExtraResourceConfigmap `json:"configmaps,omitempty"`
}

// CacheRuntimeExtraResourceConfigmap defines a configmap template for CacheRuntime extra resources
type CacheRuntimeExtraResourceConfigmap struct {
	// Name specifies the name of the configmap
	// +optional
	Name string `json:"name,omitempty"`

	// Data contains the configuration data.
	// Each key must consist of alphanumeric characters, '-', '_' or '.'.
	// Values with non-UTF-8 byte sequences must use the BinaryData field.
	// The keys stored in Data must not overlap with the keys in
	// the BinaryData field, this is enforced during validation process.
	// +optional
	Data map[string]string `json:"data,omitempty"`
}

// ExtraResourceConfigmapConfig defines the configmap mount configuration
type ExtraResourceConfigmapConfig struct {
	// Name specifies the configmap template name defined in extraResources.configmaps
	// +optional
	Name string `json:"name,omitempty"`

	// MountPath specifies the path where the configmap will be mounted
	// +optional
	MountPath string `json:"mountPath,omitempty"`
}

// CacheRuntimeClass is the Schema for the cacheruntimeclasses API
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +genclient
// +genclient:nonNamespaced
type CacheRuntimeClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// FileSystemType specifies the file system type of cache runtime
	// +kubebuilder:validation:Required
	FileSystemType string `json:"fileSystemType"`

	// Topology defines the topology of the CacheRuntime components
	// +optional
	Topology *CacheRuntimeTopology `json:"topology,omitempty"`

	// ExtraResources defines the extra resources used by the CacheRuntime components
	// +optional
	ExtraResources CacheRuntimeExtraResources `json:"extraResources,omitempty"`
}

// CacheRuntimeClassList contains a list of CacheRuntimeClass
// +kubebuilder:object:root=true
type CacheRuntimeClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CacheRuntimeClass `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CacheRuntimeClass{}, &CacheRuntimeClassList{})
}
