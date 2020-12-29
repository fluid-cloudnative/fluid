/*

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

package base

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

// Runtime Information interface defines the interfaces that should be implemented
// by Alluxio Runtime or other implementation .
// Thread safety is required from implementations of this interface.
type RuntimeInfoInterface interface {
	GetTieredstore() datav1alpha1.Tieredstore

	GetName() string

	GetNamespace() string

	GetRuntimeType() string

	GetStoragetLabelname(read common.ReadType, storage common.StorageType) string

	GetCommonLabelname() string

	GetRuntimeLabelname() string

	IsExclusive() bool

	SetupWithDataset(dataset *datav1alpha1.Dataset)
}

// The real Runtime Info should implement
type RuntimeInfo struct {
	name        string
	namespace   string
	runtimeType string

	tieredstore datav1alpha1.Tieredstore
	exclusive   bool
	// Check if the runtime info is already setup by the dataset
	setup bool
}

func BuildRuntimeInfo(name string,
	namespace string,
	runtimeType string,
	tieredstore datav1alpha1.Tieredstore) (runtime RuntimeInfoInterface) {
	runtime = &RuntimeInfo{
		name:        name,
		namespace:   namespace,
		runtimeType: runtimeType,
		tieredstore: tieredstore,
	}
	return
}

// GetTieredstore gets Tieredstore
func (info *RuntimeInfo) GetTieredstore() datav1alpha1.Tieredstore {
	return info.tieredstore
}

// GetName gets name
func (info *RuntimeInfo) GetName() string {
	return info.name
}

// GetNamespace gets namespace
func (info *RuntimeInfo) GetNamespace() string {
	return info.namespace
}

// GetRuntimeType gets runtime type
func (info *RuntimeInfo) GetRuntimeType() string {
	return info.runtimeType
}

// IsExclusive determines if the runtime is exlusive
func (info *RuntimeInfo) IsExclusive() bool {
	return info.exclusive
}

// SetupWithDataset determines if need to setup with the info of dataset
func (info *RuntimeInfo) SetupWithDataset(dataset *datav1alpha1.Dataset) {
	if !info.setup {
		info.exclusive = dataset.Spec.ExclusiveMode
		info.setup = true
	}
}
