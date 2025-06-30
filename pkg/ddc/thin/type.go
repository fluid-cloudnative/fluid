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
	FullnameOverride string `json:"fullnameOverride"`
	OwnerDatasetId   string `json:"ownerDatasetId"`

	common.ImageInfo `json:",inline"`
	common.UserInfo  `json:",inline"`

	Fuse             Fuse                          `json:"fuse,omitempty"`
	Worker           Worker                        `json:"worker,omitempty"`
	NodeSelector     map[string]string             `json:"nodeSelector,omitempty"`
	Tolerations      []corev1.Toleration           `json:"tolerations,omitempty"`
	PlacementMode    string                        `json:"placement,omitempty"`
	Owner            *common.OwnerReference        `json:"owner,omitempty"`
	RuntimeIdentity  common.RuntimeIdentity        `json:"runtimeIdentity"`
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

type Worker struct {
	Enabled          bool                          `json:"enabled,omitempty"`
	Image            string                        `json:"image,omitempty"`
	ImageTag         string                        `json:"imageTag,omitempty"`
	ImagePullPolicy  string                        `json:"imagePullPolicy,omitempty"`
	Resources        common.Resources              `json:"resources,omitempty"`
	NodeSelector     map[string]string             `json:"nodeSelector,omitempty"`
	HostNetwork      bool                          `json:"hostNetwork,omitempty"`
	Envs             []corev1.EnvVar               `json:"envs,omitempty"`
	Ports            []corev1.ContainerPort        `json:"ports,omitempty"`
	Volumes          []corev1.Volume               `json:"volumes,omitempty"`
	VolumeMounts     []corev1.VolumeMount          `json:"volumeMounts,omitempty"`
	LivenessProbe    *corev1.Probe                 `json:"livenessProbe,omitempty"`
	ReadinessProbe   *corev1.Probe                 `json:"readinessProbe,omitempty"`
	CacheDir         string                        `json:"cacheDir,omitempty"`
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

type Fuse struct {
	Enabled          bool                          `json:"enabled,omitempty"`
	Image            string                        `json:"image,omitempty"`
	ImageTag         string                        `json:"imageTag,omitempty"`
	ImagePullPolicy  string                        `json:"imagePullPolicy,omitempty"`
	Resources        common.Resources              `json:"resources,omitempty"`
	Ports            []corev1.ContainerPort        `json:"ports,omitempty"`
	CriticalPod      bool                          `json:"criticalPod,omitempty"`
	HostNetwork      bool                          `json:"hostNetwork"`
	HostPID          bool                          `json:"hostPID,omitempty"`
	TargetPath       string                        `json:"targetPath,omitempty"`
	NodeSelector     map[string]string             `json:"nodeSelector,omitempty"`
	Envs             []corev1.EnvVar               `json:"envs,omitempty"`
	Command          []string                      `json:"command,omitempty"`
	Args             []string                      `json:"args,omitempty"`
	Volumes          []corev1.Volume               `json:"volumes,omitempty"`
	VolumeMounts     []corev1.VolumeMount          `json:"volumeMounts,omitempty"`
	LivenessProbe    *corev1.Probe                 `json:"livenessProbe,omitempty"`
	ReadinessProbe   *corev1.Probe                 `json:"readinessProbe,omitempty"`
	CacheDir         string                        `json:"cacheDir,omitempty"`
	ConfigValue      string                        `json:"configValue"`
	ConfigStorage    string                        `json:"configStorage"`
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	Lifecycle        *corev1.Lifecycle             `json:"lifecycle,omitempty"`
}

type Config struct {
	Mounts                       []datav1alpha1.Mount                         `json:"mounts"`
	TargetPath                   string                                       `json:"targetPath,omitempty"`
	RuntimeOptions               map[string]string                            `json:"runtimeOptions,omitempty"`
	PersistentVolumeAttrs        map[string]*corev1.CSIPersistentVolumeSource `json:"persistentVolumeAttrs,omitempty"`
	PersistentVolumeMountOptions map[string][]string                          `json:"persistentVolumeMountOptions,omitempty"`
	AccessModes                  []corev1.PersistentVolumeAccessMode          `json:"accessModes,omitempty"`
}
