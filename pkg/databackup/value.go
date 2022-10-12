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
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

// DataBackupValue defines the value json file used in DataBackup helm chart
type DataBackupValue struct {
	DataBackup      DataBackup `json:"dataBackup"`
	common.UserInfo `json:",inline"`
	InitUsers       common.InitUsers `json:"initUsers,omitempty"`
}

// DataBackup defines values used in DataBackup helm chart
type DataBackup struct {
	Namespace   string `json:"namespace,omitempty"`
	Dataset     string `json:"dataset,omitempty"`
	Name        string `json:"name,omitempty"`
	NodeName    string `json:"nodeName,omitempty"`
	Image       string `json:"image,omitempty"`
	JavaEnv     string `json:"javaEnv,omitempty"`
	Workdir     string `json:"workdir,omitempty"`
	PVCName     string `json:"pvcName,omitempty"`
	Path        string `json:"path,omitempty"`
	RuntimeType string `json:"runtimeType,omitempty"`
}
