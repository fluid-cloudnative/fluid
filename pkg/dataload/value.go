/*
Copyright 2021 The Fluid Authors.

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

package dataload

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

// DataLoadValue defines the value yaml file used in DataLoad helm chart
type DataLoadValue struct {
	Name         string                 `json:"name"`
	Owner        *common.OwnerReference `json:"owner,omitempty"`
	DataLoadInfo DataLoadInfo           `json:"dataloader"`
}

// DataLoadInfo defines values used in DataLoad helm chart
type DataLoadInfo struct {
	// Policy including None, Once, Cron, OnEvent
	Policy string `json:"policy"`

	// Schedule The schedule in Cron format, only set when policy is cron, see https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule,omitempty"`

	// BackoffLimit specifies the upper limit times when the DataLoad job fails
	BackoffLimit int32 `json:"backoffLimit,omitempty"`

	// TargetDataset specifies the dataset that the DataLoad targets
	TargetDataset string `json:"targetDataset,omitempty"`

	// LoadMetadata specifies if the DataLoad job should load metadata from UFS when doing data load
	LoadMetadata bool `json:"loadMetadata,omitempty"`

	// TargetPaths specifies which paths should the DataLoad load
	TargetPaths []TargetPath `json:"targetPaths,omitempty"`

	// Image specifies the image that the DataLoad job uses
	Image string `json:"image,omitempty"`

	// Options specifies the extra dataload properties for runtime
	Options map[string]string `json:"options,omitempty"`

	// Labels defines labels in DataLoad's pod metadata
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations defines annotations in DataLoad's pod metadata
	Annotations map[string]string `json:"annotations,omitempty"`

	// image pull secrets
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// pod affinity
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// pod tolerations
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// node selector
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// scheduler name
	SchedulerName string `json:"schedulerName,omitempty"`

	// Resources that will be requested by DataLoad job.
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type TargetPath struct {
	// Path specifies the path should be loaded
	Path string `json:"path,omitempty"`

	// Replicas specifies how many replicas should be loaded
	Replicas int32 `json:"replicas,omitempty"`

	// FluidNative specifies if the path is a native mountPoint(e.g. hostpath or pvc)
	FluidNative bool `json:"fluidNative,omitempty"`
}
