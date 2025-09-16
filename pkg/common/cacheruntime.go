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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ComponentType string

const (
	ComponentTypeMaster ComponentType = "master"
	ComponentTypeWorker ComponentType = "worker"
	ComponentTypeClient ComponentType = "client"
)

const (
	CacheRuntime = "cache"

	CacheFuseContainer = "cache-fuse"

	CacheWorkerContainer = "cache-worker"

	CacheFuseOptionEnvKey = "MOUNT_OPTIONS"
	CacheFusePointEnvKey  = "MOUNT_POINT"

	CacheMountType = "cache-fuse-mount"

	CacheEngineImpl = CacheRuntime
)

type CacheRuntimeValue struct {
	FullnameOverride string          `json:"fullnameOverride"`
	OwnerDatasetId   string          `json:"ownerDatasetId"`
	RuntimeIdentity  RuntimeIdentity `json:"runtimeIdentity"`
	RuntimeValue     string          `json:"runtimeValue"`
	PlacementMode    string          `json:"placement,omitempty"`

	ImageInfo `json:",inline"`
	UserInfo  `json:",inline"`

	Master *CacheRuntimeComponentValue `json:"master,omitempty"`
	Worker *CacheRuntimeComponentValue `json:"worker,omitempty"`
	Client *CacheRuntimeComponentValue `json:"client,omitempty"`
}

type CacheRuntimeComponentValue struct {
	Enabled       bool              `json:"enabled"`
	Name          string            `json:"name"`
	Namespace     string            `json:"namespace"`
	WorkloadType  metav1.TypeMeta   `json:"workloadType"`
	Replicas      int32             `json:"replicas,omitempty"`
	Options       map[string]string `json:"options,omitempty"`
	CriticalPod   bool              `json:"criticalPod,omitempty"`
	Owner         *OwnerReference   `json:"owner,omitempty"`
	ComponentType ComponentType     `json:"componentType,omitempty"`
	TargetPath    string            `json:"targetPath,omitempty"`
	NodeSelector  map[string]string `json:"odeSelector,omitempty"`

	PodTemplateSpec corev1.PodTemplateSpec `json:"podTemplateSpec,omitempty"`

	Service       *CacheRuntimeComponentServiceConfig `json:"service,omitempty"`
	TieredStore   []TieredStoreOption                 `json:"tieredStore,omitempty"`
	EncryptOption map[string]string                   `json:"encryptOption,omitempty"`
}

type CacheRuntimeComponentServiceConfig struct {
	Name string `json:"name"`
}

type CacheRuntimeComponentCommonConfig struct {
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	Tolerations      []corev1.Toleration           `json:"tolerations,omitempty"`
	NodeSelector     map[string]string             `json:"nodeSelector,omitempty"`
	Envs             []corev1.EnvVar               `json:"envs,omitempty"`
	PlacementMode    string                        `json:"placement,omitempty"`
	Owner            *OwnerReference               `json:"owner,omitempty"`
	Options          map[string]string             `json:"options,omitempty"`

	EncryptOptionConfigs *EncryptOptionVolumeConfig `json:"encryptOptionConfigs,omitempty"`
	TargetPathConfig     *TargetPathVolumeConfig    `json:"targetPathConfig,omitempty"`
	RuntimeConfigConfig  *RuntimeConfigVolumeConfig `json:"runtimeConfigConfig,omitempty"`
}

type CacheRuntimeComponentTieredStoreConfig struct {
	CacheVolumeOptions []TieredStoreOption  `json:"cacheVolumeOptions,omitempty"`
	CacheVolumes       []corev1.Volume      `json:"cacheVolume,omitempty"`
	CacheVolumeMounts  []corev1.VolumeMount `json:"cacheVolumeMount,omitempty"`
}

type TieredStoreOption struct {
	CacheDir               string             `json:"cacheDir,omitempty"`
	CacheCapacity          string             `json:"cacheCapacity,omitempty"`
	Low                    string             `json:"low,omitempty"`
	High                   string             `json:"high,omitempty"`
	MemQuantityRequirement *resource.Quantity `json:"memQuantityRequirement,omitempty"`
}

type EncryptOptionVolumeConfig struct {
	EncryptOptionConfig       map[string]string
	EncryptOptionVolumes      []corev1.Volume
	EncryptOptionVolumeMounts []corev1.VolumeMount
}

type TargetPathVolumeConfig struct {
	TargetPath            string
	TargetPathHostVolume  corev1.Volume
	TargetPathVolumeMount corev1.VolumeMount
}

type RuntimeConfigVolumeConfig struct {
	RuntimeConfigPath        string
	RuntimeConfigVolume      corev1.Volume
	RuntimeConfigVolumeMount corev1.VolumeMount
}

type CacheRuntimeConfig struct {
	Mounts      []MountConfig                       `json:"mounts,omitempty"`
	AccessModes []corev1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`

	Master *CacheRuntimeComponentConfig `json:"master,omitempty"`
	Worker *CacheRuntimeComponentConfig `json:"worker,omitempty"`
	Client *CacheRuntimeComponentConfig `json:"client,omitempty"`

	Topology map[string]TopologyConfig `json:"topology,omitempty"`
}

type MountConfig struct {
	MountPoint string            `json:"mountPoint"`
	Options    map[string]string `json:"options,omitempty"`
	Name       string            `json:"name,omitempty"`
	Path       string            `json:"path,omitempty"`
	ReadOnly   bool              `json:"readOnly,omitempty"`
	Shared     bool              `json:"shared,omitempty"`
}

type CacheRuntimeComponentConfig struct {
	Enabled bool              `json:"enabled,omitempty"`
	Name    string            `json:"name,omitempty"`
	Options map[string]string `json:"options,omitempty"`

	TargetPath    string              `json:"targetPath,omitempty"`
	TieredStore   []TieredStoreOption `json:"tieredStore,omitempty"`
	EncryptOption map[string]string   `json:"encryptOption,omitempty"`
}

type TopologyConfig struct {
	PodConfigs []PodConfig                        `json:"podConfigs,omitempty"`
	Service    CacheRuntimeComponentServiceConfig `json:"service,omitempty"`
}

type PodConfig struct {
	PodName string       `json:"podName,omitempty"`
	PodIP   string       `json:"podIP,omitempty"`
	HostIP  string       `json:"hostIp,omitempty"`
	Ports   []PortConfig `json:"ports,omitempty"`
}

type PortConfig struct {
	Name string `json:"name,omitempty"`
	Port int32  `json:"port,omitempty"`
}

func GetCommonLabelsFromComponent(component *CacheRuntimeComponentValue) map[string]string {
	// Be careful to change these labels.
	// They are used as sts.spec.selector which cannot be updated. If changed, may cause all exist rbgs failed.
	return map[string]string{
		LabelCacheRuntimeName:          component.Name,
		LabelCacheRuntimeComponentName: fmt.Sprintf("%s-%s", component.Name, component.ComponentType),
	}
}
