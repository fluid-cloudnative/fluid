/*
  Copyright 2022 The Fluid Authors.

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

package common

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ComponentType string

const (
	CacheRuntime = "cache"

	CacheEngineImpl = CacheRuntime
)

const (
	ComponentTypeMaster ComponentType = "master"
	ComponentTypeWorker ComponentType = "worker"
	ComponentTypeClient ComponentType = "client"
)

type CacheRuntimeValue struct {
	// RuntimeIdentity is used to identify the runtime (name/namespace)
	RuntimeIdentity RuntimeIdentity `json:"runtimeIdentity"`

	Master *CacheRuntimeComponentValue `json:"master,omitempty"`
	Worker *CacheRuntimeComponentValue `json:"worker,omitempty"`
	Client *CacheRuntimeComponentValue `json:"client,omitempty"`
}

// CacheRuntimeComponentValue is the common value for building CacheRuntimeValue.
type CacheRuntimeComponentValue struct {
	// Component name, not Runtime name
	Name            string
	Namespace       string
	Enabled         bool
	WorkloadType    metav1.TypeMeta
	Replicas        int32
	PodTemplateSpec corev1.PodTemplateSpec
	Owner           *OwnerReference
	ComponentType   ComponentType `json:"componentType,omitempty"`

	// Service name, can be not same as Component name
	Service *CacheRuntimeComponentServiceConfig
}

// CacheRuntimeConfig defines the config of runtime, will be auto mounted by configmap in the component pod.
type CacheRuntimeConfig struct {
	// Mounts from Dataset Spec
	Mounts []MountConfig `json:"mounts,omitempty"`
	// AccessModes from Dataset Spec
	AccessModes []corev1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`
	// fuse mount path, used in Worker or Client Pod according to Topology.
	TargetPath string `json:"targetPath,omitempty"`

	Master *CacheRuntimeComponentConfig `json:"master,omitempty"`
	Worker *CacheRuntimeComponentConfig `json:"worker,omitempty"`
	Client *CacheRuntimeComponentConfig `json:"client,omitempty"`
}

// MountConfig defines the mount config about dataset Mounts
type MountConfig struct {
	MountPoint string `json:"mountPoint"`
	// Non-encrypted mount options, key is the option name, value is the option value.
	Options map[string]string `json:"options,omitempty"`
	// Encrypted mount options, key is the option name, value is the secret mount path in container.
	EncryptOptions map[string]string `json:"encryptOptions,omitempty"`
	Name           string            `json:"name,omitempty"`
	Path           string            `json:"path,omitempty"`
	ReadOnly       bool              `json:"readOnly,omitempty"`
	Shared         bool              `json:"shared,omitempty"`
}

type CacheRuntimeComponentConfig struct {
	Enabled  bool              `json:"enabled,omitempty"`
	Name     string            `json:"name,omitempty"`
	Options  map[string]string `json:"options,omitempty"`
	Replicas int32             `json:"replicas,omitempty"`

	Service CacheRuntimeComponentServiceConfig `json:"service,omitempty"`
}

type CacheRuntimeComponentServiceConfig struct {
	Name string `json:"name"`
}
