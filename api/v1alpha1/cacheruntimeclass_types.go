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

// RuntimeTopology defines the topology structure of CacheRuntime components
type RuntimeTopology struct {
	// Master is the configuration for master component
	// +optional
	Master *RuntimeComponentDefinition `json:"master,omitempty"`

	// Worker is the configuration for worker component
	// +optional
	Worker *RuntimeComponentDefinition `json:"worker,omitempty"`

	// Client is the configuration for client component
	// +optional
	Client *RuntimeComponentDefinition `json:"client,omitempty"`
}

// RuntimeComponentDefinition defines the configuration for a CacheRuntime component
type RuntimeComponentDefinition struct {
	// WorkloadType is the default workload type of the component
	// +optional
	WorkloadType metav1.TypeMeta `json:"workloadType,omitempty"`

	// Options is a set of key-value pairs that provide additional configuration for the component
	// +optional
	Options map[string]string `json:"options,omitempty"`

	// Template describes the pods that will be created.
	// The template follows the standard PodTemplateSpec from Kubernetes core.
	// +optional
	Template corev1.PodTemplateSpec `json:"template,omitempty"`

	// Service is the service configuration for the component
	// +optional
	Service RuntimeComponentService `json:"service,omitempty"`

	// Dependencies specifies the dependencies required by the component
	// +optional
	Dependencies RuntimeComponentDependencies `json:"dependencies,omitempty"`
}

// EncryptOptionComponentDependency defines the configuration for encrypt option dependency
type EncryptOptionComponentDependency struct {
}

// ExtraResourcesComponentDependency defines the extra resources configuration for component dependencies
type ExtraResourcesComponentDependency struct {
	// ConfigMaps is a list of ConfigMaps in the same namespace to mount into the component
	// +optional
	ConfigMaps []ConfigMapDependencyConfig `json:"configMaps,omitempty"`
}

// RuntimeComponentDependencies defines the dependencies required by a CacheRuntime component
type RuntimeComponentDependencies struct {
	// EncryptOption is the configuration for encrypt option secret mount
	// +optional
	EncryptOption *EncryptOptionComponentDependency `json:"encryptOption,omitempty"`

	// ExtraResources specifies the usage of extra resources such as ConfigMaps
	// +optional
	ExtraResources *ExtraResourcesComponentDependency `json:"extraResources,omitempty"`
}

// HeadlessRuntimeComponentService defines the configuration for headless service
type HeadlessRuntimeComponentService struct {
}

// ComponentServiceConfig defines the service configuration for runtime components.
// Currently only headless service is supported, but this can be extended in the future
// to support other service types such as ClusterIP, NodePort, LoadBalancer, etc.
type ComponentServiceConfig struct {
	// Headless enables a headless service for the component.
	// A headless service does not allocate a cluster IP and allows direct pod-to-pod communication.
	// +optional
	Headless *HeadlessRuntimeComponentService `json:"headless,omitempty"`
}

// RuntimeComponentService describes the service configuration for a CacheRuntime component
type RuntimeComponentService struct {
	ComponentServiceConfig `json:",inline"`
}

// ConfigMapRuntimeExtraResource defines a ConfigMap template for CacheRuntime extra resources
type ConfigMapRuntimeExtraResource struct {
	// Name is the name of the ConfigMap.
	// This will be used as the actual ConfigMap name when created in the runtime's namespace.
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

// ConfigMapDependencyConfig defines the ConfigMap mount configuration
type ConfigMapDependencyConfig struct {
	// Name is the ConfigMap template name defined in extraResources.configMaps
	// +optional
	Name string `json:"name,omitempty"`

	// MountPath is the path within the container at which the ConfigMap should be mounted.
	// Must not contain ':'.
	// +optional
	MountPath string `json:"mountPath,omitempty"`
}

// RuntimeExtraResources defines the extra resources for CacheRuntime
type RuntimeExtraResources struct {
	// ConfigMaps is a list of ConfigMaps that will be created in the runtime's namespace.
	// These ConfigMaps can be referenced and mounted by runtime components.
	// +optional
	ConfigMaps []ConfigMapRuntimeExtraResource `json:"configMaps,omitempty"`
}

// CacheRuntimeClass is the Schema for the cacheruntimeclasses API.
// CacheRuntimeClass defines a class of cache runtime implementations with specific configurations.
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:resource:categories={fluid}
// +genclient
// +genclient:nonNamespaced
type CacheRuntimeClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// FileSystemType is the file system type of the cache runtime (e.g., "alluxio", "juicefs")
	// +kubebuilder:validation:Required
	FileSystemType string `json:"fileSystemType"`

	// Topology describes the topology of the CacheRuntime components (master, worker, client)
	// +optional
	Topology *RuntimeTopology `json:"topology,omitempty"`

	// ExtraResources specifies additional resources (e.g., ConfigMaps) used by the CacheRuntime components
	// +optional
	ExtraResources RuntimeExtraResources `json:"extraResources,omitempty"`
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
