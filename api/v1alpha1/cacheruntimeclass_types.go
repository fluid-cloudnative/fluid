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

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +genclient

type CacheRuntimeClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// file system of cacheRuntime
	// +required
	FileSystemType string `json:"fileSystemType"`

	// Topology defines the topology of the CacheRuntime components
	Topology *CacheRuntimeTopology `json:"topology,omitempty"`

	// ExtraResources defines the extra resources used by the CacheRuntime components
	// +optional
	ExtraResources CacheRuntimeExtraResources `json:"extraResources,omitempty"`
}

type CacheRuntimeTopology struct {
	// The component spec of master component
	// +optional
	Master *CacheRuntimeTopologyComponentDefinition `json:"master,omitempty"`

	// The component spec of worker component
	// +optional
	Worker *CacheRuntimeTopologyComponentDefinition `json:"worker,omitempty"`

	// The component spec of client component
	// +optional
	Client *CacheRuntimeTopologyComponentDefinition `json:"client,omitempty"`
}

type CacheRuntimeTopologyComponentDefinition struct {
	// WorkloadType defines the default workload type of the component
	WorkloadType metav1.TypeMeta `json:"workloadType,omitempty"`

	Options map[string]string `json:"options,omitempty"`

	// PodTemplateSpec defines the pod spec of the components
	PodTemplateSpec corev1.PodTemplateSpec `json:"podTemplateSpec,omitempty"`

	// Service defines the service used by the components
	// +optional
	Service CacheRuntimeComponentService `json:"service,omitempty"`

	Dependencies Dependencies `json:"dependencies,omitempty"`
}

type Dependencies struct {
	// EncryptOptionConfig defines the config of the encrypt option secret mount
	// +optional
	EncryptOption *EncryptOptionConfig `json:"encryptOption,omitempty"`

	// ExtraResources define the usage of extraResources
	// +optional
	ExtraResources *ExtraResources `json:"extraResources,omitempty"`
}

type CacheRuntimeComponentService struct {
	Headless *CacheRuntimeComponentHeadlessService `json:"headless,omitempty"`

	// TODO: ClusterIPService *CacheRuntimeComponentHeadlessService `json:"clusterIPService,omitempty"`
}

type CacheRuntimeComponentHeadlessService struct {
}

type EncryptOptionConfig struct {
}

type ExtraResources struct {
	// Configmaps define the template of configmaps which will be used to create a configmap in runtime's namespace
	Configmaps []ExtraResourceConfigmapConfig `json:"configmaps,omitempty"`
}

type CacheRuntimeExtraResources struct {
	// Configmaps define the template of configmaps which will be used to create a configmap in runtime's namespace
	Configmaps []CacheRuntimeExtraResourceConfigmap `json:"configmaps,omitempty"`
}

type CacheRuntimeExtraResourceConfigmap struct {
	// Name of the configmap
	Name string `json:"name,omitempty"`

	// Data contains the configuration data.
	// Each key must consist of alphanumeric characters, '-', '_' or '.'.
	// Values with non-UTF-8 byte sequences must use the BinaryData field.
	// The keys stored in Data must not overlap with the keys in
	// the BinaryData field, this is enforced during validation process.
	// +optional
	Data map[string]string `json:"data,omitempty"`
}

type ExtraResourceConfigmapConfig struct {
	// Name indicates the configmap template name defined in extraResources.configmaps
	Name string `json:"name,omitempty"`

	// MountPath define the configmap volume mountPath
	MountPath string `json:"mountPath,omitempty"`
}

//+kubebuilder:object:root=true

// CacheRuntimeClassList contains a list of RuntimeClass
type CacheRuntimeClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CacheRuntimeClass `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CacheRuntimeClass{}, &CacheRuntimeClassList{})
}
