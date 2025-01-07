/*
Copyright 2021 The Fluid Author.

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

package databackup

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/fluid-cloudnative/fluid/pkg/common"
)

// DataBackupValue defines the value yaml file used in DataBackup helm chart
type DataBackupValue struct {
	DataBackup      DataBackup `yaml:"dataBackup"`
	common.UserInfo `yaml:",inline"`
	InitUsers       common.InitUsers `yaml:"initUsers,omitempty"`
}

// DataBackup defines values used in DataBackup helm chart
type DataBackup struct {
	Namespace      string `yaml:"namespace,omitempty"`
	Dataset        string `yaml:"dataset,omitempty"`
	OwnerDatasetId string `yaml:"ownerDatasetId,omitempty"`
	Name           string `yaml:"name,omitempty"`
	NodeName       string `yaml:"nodeName,omitempty"`
	Image          string `yaml:"image,omitempty"`
	JavaEnv        string `yaml:"javaEnv,omitempty"`
	Workdir        string `yaml:"workdir,omitempty"`
	PVCName        string `yaml:"pvcName,omitempty"`
	Path           string `yaml:"path,omitempty"`
	RuntimeType    string `yaml:"runtimeType,omitempty"`
	// image pull secrets
	ImagePullSecrets []corev1.LocalObjectReference `yaml:"imagePullSecrets,omitempty"`
	Affinity         *corev1.Affinity              `yaml:"affinity,omitempty"`
}
