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
	"fmt"

	corev1 "k8s.io/api/core/v1"
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

// MinExecutionTimeoutSeconds is the minimum timeout for execution in Component pods.
const MinExecutionTimeoutSeconds = 20

type CacheRuntimeValue struct {
	Master *CacheRuntimeComponentValue
	Worker *CacheRuntimeComponentValue
	Client *CacheRuntimeComponentValue
}

// CacheRuntimeComponentValue is the common value for building CacheRuntimeValue.
type CacheRuntimeComponentValue struct {
	// Component name, not Runtime name
	Name            string
	Namespace       string
	Enabled         bool
	Replicas        int32
	PodTemplateSpec corev1.PodTemplateSpec
	Owner           *OwnerReference
	ComponentType   ComponentType

	// Service name, can be not same as Component name
	Service *CacheRuntimeComponentServiceConfig
}

// CacheRuntimeStatusValue contains only the fields needed for status update
type CacheRuntimeStatusValue struct {
	Master *ComponentStatusInfo
	Worker *ComponentStatusInfo
	Client *ComponentStatusInfo
}

// ComponentIdentity contains minimal identity information for component status queries
type ComponentIdentity struct {
	Name      string
	Namespace string
}

// ComponentStatusInfo contains the minimal information needed for status updates
type ComponentStatusInfo struct {
	ComponentIdentity
	Enabled bool
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

	// TieredStoreLevels contains the tiered storage configuration for worker
	// This field is primarily used by Worker component to configure cache storage tiers
	TieredStoreLevels []TieredStoreLevelConfig `json:"tieredStoreLevels,omitempty"`
}

type CacheRuntimeComponentServiceConfig struct {
	Name string `json:"name"`
}

// GetCacheComponentName gets the component name using runtime name and component type.
func GetCacheComponentName(runtimeName string, componentType ComponentType) string {
	return fmt.Sprintf("%s-%s", runtimeName, componentType)
}

// GetCacheRuntimeConfigConfigMapName get the configmap name of the runtime config.
func GetCacheRuntimeConfigConfigMapName(name string) string {
	return fmt.Sprintf("fluid-runtime-config-%s", name)
}

// TieredStoreLevelConfig defines the configuration for a single tier in the tiered storage.
// This config will be mounted into the worker container via ConfigMap.
type TieredStoreLevelConfig struct {
	// MountPaths contains the mount paths inside the container for this tier
	// For processMemory: single element array with the mount path (e.g. ["/dev/shm"])
	// For emptyDir: single element array with the mount path
	// For hostPath: array of mount paths corresponding to each host path
	MountPaths []string `json:"mountPaths,omitempty"`

	// MediumType indicates the storage medium type: "MEM", "SSD", "HDD", etc.
	// This represents the performance characteristics of the storage medium
	MediumType MediumType `json:"mediumType,omitempty"`

	// Quotas is the storage capacity for this tier (e.g., ["100Gi"])
	// The length of Quotas matches the length of MountPaths
	// For processMemory/emptyDir: single element array with the quota
	// For hostPath: array of quotas corresponding to each mount path
	Quotas []string `json:"quotas,omitempty"`

	// High is the high watermark ratio (e.g., "0.9")
	// When cache usage exceeds this ratio, eviction will be triggered
	High string `json:"high,omitempty"`

	// Low is the low watermark ratio (e.g., "0.7")
	// Eviction will continue until cache usage falls below this ratio
	Low string `json:"low,omitempty"`
}
