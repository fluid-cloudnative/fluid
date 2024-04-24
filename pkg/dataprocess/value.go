/*
  Copyright 2023 The Fluid Authors.

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

package dataprocess

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

type DataProcessValue struct {
	Name string `json:"name"`

	Owner *common.OwnerReference `json:"owner,omitempty"`

	DataProcessInfo DataProcessInfo `json:"dataProcess"`
}

type DataProcessInfo struct {
	TargetDataset string `json:"targetDataset,omitempty"`

	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	Labels map[string]string `json:"labels,omitempty"`

	Annotations map[string]string `json:"annotations,omitempty"`

	JobProcessor *JobProcessor `json:"jobProcessor,omitempty"`

	ScriptProcessor *ScriptProcessor `json:"scriptProcessor,omitempty"`
}

type ScriptProcessor struct {
	Image string `json:"image,omitempty"`

	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	RestartPolicy corev1.RestartPolicy `json:"restartPolicy,omitempty"`

	Command []string `json:"command,omitempty"`

	Source string `json:"source,omitempty"`

	Envs []corev1.EnvVar `json:"envs,omitempty"`

	Volumes []corev1.Volume `json:"volumes,omitempty"`

	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`

	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type JobProcessor struct {
	PodSpec *corev1.PodSpec `json:"podSpec,omitempty"`
}
