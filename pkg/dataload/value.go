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

// DataLoadValue defines the value yaml file used in DataLoad helm chart
type DataLoadValue struct {
	DataLoadInfo DataLoadInfo `yaml:"dataloader"`
}

// DataLoadInfo defines values used in DataLoad helm chart
type DataLoadInfo struct {
	// BackoffLimit specifies the upper limit times when the DataLoad job fails
	BackoffLimit int32 `yaml:"backoffLimit,omitempty"`

	// TargetDataset specifies the dataset that the DataLoad targets
	TargetDataset string `yaml:"targetDataset,omitempty"`

	// LoadMetadata specifies if the DataLoad job should load metadata from UFS when doing data load
	LoadMetadata bool `yaml:"loadMetadata,omitempty"`

	// TargetPaths specifies which paths should the DataLoad load
	TargetPaths []TargetPath `yaml:"targetPaths,omitempty"`

	// Image specifies the image that the DataLoad job uses
	Image string `yaml:"image,omitempty"`

	// Options specifies the extra dataload properties for runtime
	Options map[string]string `yaml:"options,omitempty"`
}

type TargetPath struct {
	// Path specifies the path should be loaded
	Path string `yaml:"path,omitempty"`

	// Replicas specifies how many replicas should be loaded
	Replicas int32 `yaml:"replicas,omitempty"`

	// FluidNative specifies if the path is a native mountPoint(e.g. hostpath or pvc)
	FluidNative bool `yaml:"fluidNative,omitempty"`
}
