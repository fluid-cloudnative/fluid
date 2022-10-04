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

package thin

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

type ThinValue struct {
	FullnameOverride string `yaml:"fullnameOverride"`

	common.ImageInfo `yaml:",inline"`
	common.UserInfo  `yaml:",inline"`

	Fuse          Fuse                   `yaml:"fuse,omitempty"`
	Worker        Worker                 `yaml:"worker,omitempty"`
	NodeSelector  map[string]string      `yaml:"nodeSelector,omitempty"`
	Tolerations   []corev1.Toleration    `json:"tolerations,omitempty"`
	PlacementMode string                 `yaml:"placement,omitempty"`
	Owner         *common.OwnerReference `yaml:"owner,omitempty"`
}

type Worker struct {
	Image           string                 `yaml:"image,omitempty"`
	ImageTag        string                 `yaml:"imageTag,omitempty"`
	ImagePullPolicy string                 `yaml:"imagePullPolicy,omitempty"`
	Resources       common.Resources       `yaml:"resources,omitempty"`
	NodeSelector    map[string]string      `yaml:"nodeSelector,omitempty"`
	HostNetwork     bool                   `yaml:"hostNetwork,omitempty"`
	Envs            []corev1.EnvVar        `yaml:"envs,omitempty"`
	Ports           []corev1.ContainerPort `yaml:"ports,omitempty"`
	Volumes         []corev1.Volume        `yaml:"volumes,omitempty"`
	VolumeMounts    []corev1.VolumeMount   `yaml:"volumeMounts,omitempty"`
	LivenessProbe   *corev1.Probe          `yaml:"livenessProbe,omitempty"`
	ReadinessProbe  *corev1.Probe          `yaml:"readinessProbe,omitempty"`
	CacheDir        string                 `yaml:"cacheDir,omitempty"`
}

type Fuse struct {
	Enabled         bool                   `yaml:"enabled,omitempty"`
	Image           string                 `yaml:"image,omitempty"`
	ImageTag        string                 `yaml:"imageTag,omitempty"`
	ImagePullPolicy string                 `yaml:"imagePullPolicy,omitempty"`
	Resources       common.Resources       `yaml:"resources,omitempty"`
	Ports           []corev1.ContainerPort `yaml:"ports,omitempty"`
	CriticalPod     bool                   `yaml:"criticalPod,omitempty"`
	HostNetwork     bool                   `yaml:"hostNetwork,omitempty"`
	MountPath       string                 `yaml:"mountPath,omitempty"`
	NodeSelector    map[string]string      `yaml:"nodeSelector,omitempty"`
	Envs            []corev1.EnvVar        `yaml:"envs,omitempty"`
	Command         []string               `yaml:"command,omitempty"`
	Args            []string               `yaml:"args,omitempty"`
	Volumes         []corev1.Volume        `yaml:"volumes,omitempty"`
	VolumeMounts    []corev1.VolumeMount   `yaml:"volumeMounts,omitempty"`
	LivenessProbe   *corev1.Probe          `yaml:"livenessProbe,omitempty"`
	ReadinessProbe  *corev1.Probe          `yaml:"readinessProbe,omitempty"`
	CacheDir        string                 `yaml:"cacheDir,omitempty"`
	ConfigValue     string                 `yaml:"configValue"`
}

type Config struct {
	Mounts         []datav1alpha1.Mount `json:"mounts"`
	RuntimeOptions map[string]string    `json:"runtimeOptions"`
}

type RuntimeStatus struct {
	Workers []string `json:"workers"`
	Fuses   []string `json:"fuses"`
}
