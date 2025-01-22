/*
Copyright 2020 The Fluid Author.

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

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced
// +genclient

type CacheRuntimeClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// PluginName is the name of the plugin used by the runtime class
	PluginName string `json:"pluginName,omitempty"`

	// Topology defines the topology of the CacheRuntime components
	Topology CacheRuntimeTopology `json:"topology,omitempty"`

	// ExtraResources defines the extra resources used by the CacheRuntime components
	// +optional
	ExtraResources CacheRuntimeExtraResources `json:"extraResources,omitempty"`

	//TODO DataOperation Definition
}

type CacheRuntimeExtraResources struct {
	// ConfigMap holds configuration data for pods to consume.
	Configmaps []CacheRuntimeResourceConfigmap `json:"configmaps,omitempty"`
}

type CacheRuntimeResourceConfigmap struct {
	// Name of the configmap
	Name string `json:"name,omitempty"`

	// TemplateRef indicates the template reference of the configmap
	TemplateRef CacheRuntimeResourceConfigmapTemplateRef `json:"templateRef,omitempty"`
}

type CacheRuntimeResourceConfigmapTemplateRef struct {
	// Name of the configmap
	Name string `json:"name,omitempty"`

	// Namespace of the configmap
	Namespace string `json:"namespace,omitempty"`
}

type CacheRuntimeTopology struct {
	// The component spec of Cache master
	// +optional
	Master CacheRuntimeTopologyComponentDefinition `json:"master,omitempty"`

	// The component spec of Cache worker
	// +optional
	Worker CacheRuntimeTopologyComponentDefinition `json:"worker,omitempty"`

	// The component spec of Cache client
	// +optional
	Client CacheRuntimeTopologyComponentDefinition `json:"client,omitempty"`
}

type CacheRuntimeTopologyComponentDefinition struct {
	// Replicas defines the default replicas of the component
	Replicas int `json:"replicas,omitempty"`

	// WorkloadType defines the default workload type of the component
	WorkloadType metav1.TypeMeta `json:"workloadType,omitempty"`

	// EncryptOptionSecretMountConfig defines the config of the encrypt option secret mount
	// +optional
	EncryptOptionSecretMountConfig EncryptOptionSecretMountConfig `json:"encryptOptionSecretMountConfig,omitempty"`

	// CacheDirMountConfig defines the config of the cache dir mount
	// +optional
	CacheDirMountConfig bool `json:"cacheDirMountConfig,omitempty"`

	// Services defines the services used by the components
	// +optional
	Services []CacheRuntimeTopologyComponentService `json:"services,omitempty"`

	// PodSpec defines the pod spec of the components
	PodSpec corev1.PodSpec `json:"podSpec,omitempty"`
}

type CacheRuntimeTopologyComponentService struct {
	// Name of the service
	Name string `json:"name,omitempty"`

	// Type of the service
	// +kubebuilder:validation:Enum=headless;"";clusterIP;nodeIP
	// +optional
	Type string `json:"type,omitempty"`

	// Ports of the service
	Ports []corev1.ContainerPort `json:"ports,omitempty"`
}

type EncryptOptionSecretMountConfig struct {
	// MountPath of the encrypt option secret volume
	MountPath string `json:"mountPath,omitempty"`

	// SetAsEnv indicates whether the encrypt option secret volume should be set as env in component pod's containers
	SetAsEnv bool `json:"setAsEnv,omitempty"`
}

type CacheDirMountConfig struct {
	// Enable indicates whether the cache dir mount should be mounted to the component
	Enable bool `json:"enable,omitempty"`
}
