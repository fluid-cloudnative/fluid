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

package datamigrate

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

type DataMigrateValue struct {
	DataMigrateInfo DataMigrateInfo `json:"datamigrate"`
}

type DataMigrateInfo struct {
	// BackoffLimit specifies the upper limit times when the DataMigrate job fails
	BackoffLimit int32 `json:"backoffLimit,omitempty"`

	// TargetDataset specifies the dataset that the DataLoad targets
	TargetDataset string `json:"targetDataset,omitempty"`

	// MigrateFrom specifies the data that the DataMigrate migrate from
	MigrateFrom string `json:"migrateFrom,omitempty"`

	// MigrateTo specifies the data that the DataMigrate migrate to
	MigrateTo string `json:"migrateTo,omitempty"`

	// EncryptOptions specifies the encrypt options that the DataMigrate job uses
	EncryptOptions []v1alpha1.EncryptOption `json:"encryptOptions,omitempty"`

	// Image specifies the image that the DataMigrate job uses
	Image string `json:"image,omitempty"`

	// Options specifies the extra dataMigrate properties for runtime
	Options map[string]string `json:"options,omitempty"`

	// Labels defines labels in DataMigrate's pod metadata
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations defines annotations in DataMigrate's pod metadata
	Annotations map[string]string `json:"annotations,omitempty"`

	// image pull secrets
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// specifies local:// and pvc:// volume
	NativeVolumes []corev1.Volume `json:"nativeVolumes,omitempty"`

	// specifies local:// and pvc:// volume mount
	NativeVolumeMounts []corev1.VolumeMount `json:"nativeVolumeMounts,omitempty"`
}
