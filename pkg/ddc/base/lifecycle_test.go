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

	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

func TestDefaultExtendedLifecycleManager_LoadData(t *testing.T) {
	manager := &DefaultExtendedLifecycleManager{}
	ctx := cruntime.ReconcileRequestContext{}
	spec := &DataLoadSpec{
		Paths: []string{"/test/path"},
	}

	result, err := manager.LoadData(ctx, spec)
	if err == nil {
		t.Errorf("Expected error for LoadData, got nil")
	}
	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}
	if err.Error() != "LoadData not implemented by this engine" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestDefaultExtendedLifecycleManager_ProcessData(t *testing.T) {
	manager := &DefaultExtendedLifecycleManager{}
	ctx := cruntime.ReconcileRequestContext{}
	spec := &DataProcessSpec{
		Processor:  "test-processor",
		InputPaths: []string{"/input"},
		OutputPath: "/output",
	}

	result, err := manager.ProcessData(ctx, spec)
	if err == nil {
		t.Errorf("Expected error for ProcessData, got nil")
	}
	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}
}

func TestDefaultExtendedLifecycleManager_MutateData(t *testing.T) {
	manager := &DefaultExtendedLifecycleManager{}
	ctx := cruntime.ReconcileRequestContext{}
	spec := &DataMutationSpec{
		MutationType: MutationTypeCreate,
		Path:         "/test/path",
		Data:         []byte("test data"),
	}

	result, err := manager.MutateData(ctx, spec)
	if err == nil {
		t.Errorf("Expected error for MutateData, got nil")
	}
	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}
}

func TestDefaultExtendedLifecycleManager_GetDataOperationStatus(t *testing.T) {
	manager := &DefaultExtendedLifecycleManager{}

	status, err := manager.GetDataOperationStatus("test-op-id")
	if err == nil {
		t.Errorf("Expected error for GetDataOperationStatus, got nil")
	}
	if status != nil {
		t.Errorf("Expected nil status, got %v", status)
	}
}
