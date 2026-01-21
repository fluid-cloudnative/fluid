/*
Copyright 2024 The Fluid Authors.

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
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestValidateRuntimeInfo(t *testing.T) {
	tests := []struct {
		name               string
		ownerDatasetUID    string
		placementModeSet   bool
		wantErr            bool
		wantTemporaryError bool
	}{
		{
			name:               "valid runtime info",
			ownerDatasetUID:    "uid-12345",
			placementModeSet:   true,
			wantErr:            false,
			wantTemporaryError: false,
		},
		{
			name:               "empty OwnerDatasetUID returns temporary error",
			ownerDatasetUID:    "",
			placementModeSet:   true,
			wantErr:            true,
			wantTemporaryError: true,
		},
		{
			name:               "placement mode not set returns temporary error",
			ownerDatasetUID:    "uid-12345",
			placementModeSet:   false,
			wantErr:            true,
			wantTemporaryError: true,
		},
		{
			name:               "both invalid returns error for OwnerDatasetUID first",
			ownerDatasetUID:    "",
			placementModeSet:   false,
			wantErr:            true,
			wantTemporaryError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockRuntimeInfoForValidate{
				ownerDatasetUID:  tt.ownerDatasetUID,
				placementModeSet: tt.placementModeSet,
			}

			err := ValidateRuntimeInfo(mock)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRuntimeInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.wantTemporaryError {
				if !fluiderrs.IsTemporaryValidationFailed(err) {
					t.Errorf("ValidateRuntimeInfo() error should be TemporaryValidationFailed, got %T", err)
				}
			}
		})
	}
}

// mockRuntimeInfoForValidate implements RuntimeInfoInterface for testing ValidateRuntimeInfo
type mockRuntimeInfoForValidate struct {
	ownerDatasetUID  string
	placementModeSet bool
}

// Methods used by ValidateRuntimeInfo
func (m *mockRuntimeInfoForValidate) GetOwnerDatasetUID() string { return m.ownerDatasetUID }
func (m *mockRuntimeInfoForValidate) IsPlacementModeSet() bool   { return m.placementModeSet }

// Conventions interface methods (stub implementations)
func (m *mockRuntimeInfoForValidate) GetPersistentVolumeName() string  { return "" }
func (m *mockRuntimeInfoForValidate) GetLabelNameForMemory() string    { return "" }
func (m *mockRuntimeInfoForValidate) GetLabelNameForDisk() string      { return "" }
func (m *mockRuntimeInfoForValidate) GetLabelNameForTotal() string     { return "" }
func (m *mockRuntimeInfoForValidate) GetCommonLabelName() string       { return "" }
func (m *mockRuntimeInfoForValidate) GetFuseLabelName() string         { return "" }
func (m *mockRuntimeInfoForValidate) GetRuntimeLabelName() string      { return "" }
func (m *mockRuntimeInfoForValidate) GetDatasetNumLabelName() string   { return "" }
func (m *mockRuntimeInfoForValidate) GetWorkerStatefulsetName() string { return "" }
func (m *mockRuntimeInfoForValidate) GetExclusiveLabelValue() string   { return "" }

// RuntimeInfoInterface methods (stub implementations)
func (m *mockRuntimeInfoForValidate) GetTieredStoreInfo() TieredStoreInfo { return TieredStoreInfo{} }
func (m *mockRuntimeInfoForValidate) GetName() string                     { return "" }
func (m *mockRuntimeInfoForValidate) GetNamespace() string                { return "" }
func (m *mockRuntimeInfoForValidate) GetRuntimeType() string              { return "" }
func (m *mockRuntimeInfoForValidate) GetPlacementModeWithDefault(defaultMode datav1alpha1.PlacementMode) datav1alpha1.PlacementMode {
	return defaultMode
}
func (m *mockRuntimeInfoForValidate) SetFuseNodeSelector(nodeSelector map[string]string) {
	// No-op for test mock
}
func (m *mockRuntimeInfoForValidate) SetFuseName(fuseName string) {
	// No-op for test mock
}
func (m *mockRuntimeInfoForValidate) SetupFuseCleanPolicy(policy datav1alpha1.FuseCleanPolicy) {
	// No-op for test mock
}
func (m *mockRuntimeInfoForValidate) SetupWithDataset(dataset *datav1alpha1.Dataset) {
	// No-op for test mock
}
func (m *mockRuntimeInfoForValidate) SetOwnerDatasetUID(alias types.UID) {
	// No-op for test mock
}
func (m *mockRuntimeInfoForValidate) GetFuseNodeSelector() map[string]string { return nil }
func (m *mockRuntimeInfoForValidate) GetFuseName() string                    { return "" }
func (m *mockRuntimeInfoForValidate) GetFuseCleanPolicy() datav1alpha1.FuseCleanPolicy {
	return datav1alpha1.OnDemandCleanPolicy
}
func (m *mockRuntimeInfoForValidate) GetFuseContainerTemplate() (*common.FuseInjectionTemplate, error) {
	return nil, nil
}
func (m *mockRuntimeInfoForValidate) SetAPIReader(apiReader client.Reader) {
	// No-op for test mock
}
func (m *mockRuntimeInfoForValidate) GetMetadataList() []datav1alpha1.Metadata { return nil }
func (m *mockRuntimeInfoForValidate) GetAnnotations() map[string]string        { return nil }
func (m *mockRuntimeInfoForValidate) GetFuseMetricsScrapeTarget() mountModeSelector {
	return mountModeSelector{}
}
