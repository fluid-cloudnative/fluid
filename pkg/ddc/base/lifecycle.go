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
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

// DataLifecycleManager defines interfaces for complete data lifecycle management
// including data loading, data processing workflows, and cache-aware data mutations
type DataLifecycleManager interface {
	// LoadData loads data into cache from UFS
	// This is a cache-aware operation that optimizes data placement
	LoadData(ctx cruntime.ReconcileRequestContext, spec *DataLoadSpec) (*DataLoadResult, error)

	// ProcessData executes data processing workflows on cached data
	// Returns processing result and any errors encountered
	ProcessData(ctx cruntime.ReconcileRequestContext, spec *DataProcessSpec) (*DataProcessResult, error)

	// MutateData performs cache-aware data mutations
	// Ensures mutations are performed efficiently using cache locality
	MutateData(ctx cruntime.ReconcileRequestContext, spec *DataMutationSpec) (*DataMutationResult, error)

	// GetDataOperationStatus retrieves the current status of a data operation
	GetDataOperationStatus(operationID string) (*DataOperationStatus, error)
}

// DataLoadSpec specifies parameters for data loading operation
type DataLoadSpec struct {
	// Paths specifies the data paths to load
	Paths []string

	// Options contains additional options for data loading
	Options map[string]string

	// Priority indicates the priority of the load operation
	Priority int
}

// DataLoadResult contains the result of a data load operation
type DataLoadResult struct {
	// OperationID uniquely identifies this operation
	OperationID string

	// Status indicates the current status
	Status DataOperationStatus

	// LoadedBytes indicates how many bytes were loaded
	LoadedBytes int64

	// LoadedFiles indicates how many files were loaded
	LoadedFiles int64
}

// DataProcessSpec specifies parameters for data processing operation
type DataProcessSpec struct {
	// Processor defines the processing logic to apply
	Processor string

	// InputPaths specifies input data paths
	InputPaths []string

	// OutputPath specifies where to store processed data
	OutputPath string

	// Options contains additional processing options
	Options map[string]string
}

// DataProcessResult contains the result of a data processing operation
type DataProcessResult struct {
	// OperationID uniquely identifies this operation
	OperationID string

	// Status indicates the current status
	Status DataOperationStatus

	// ProcessedBytes indicates how many bytes were processed
	ProcessedBytes int64

	// OutputPath is the final output path
	OutputPath string
}

// DataMutationSpec specifies parameters for data mutation operation
type DataMutationSpec struct {
	// MutationType specifies the type of mutation (create, update, delete)
	MutationType MutationType

	// Path specifies the data path to mutate
	Path string

	// Data contains the data to write (for create/update)
	Data []byte

	// Options contains additional mutation options
	Options map[string]string
}

// DataMutationResult contains the result of a data mutation operation
type DataMutationResult struct {
	// OperationID uniquely identifies this operation
	OperationID string

	// Status indicates the current status
	Status DataOperationStatus

	// MutatedPath is the path that was mutated
	MutatedPath string
}

// MutationType represents the type of data mutation
type MutationType string

const (
	MutationTypeCreate MutationType = "create"
	MutationTypeUpdate MutationType = "update"
	MutationTypeDelete MutationType = "delete"
)

// DataOperationStatus represents the status of a data operation
type DataOperationStatus struct {
	// Phase indicates the current phase of the operation
	Phase DataOperationPhase

	// Message contains a human-readable message about the status
	Message string

	// Progress indicates operation progress (0-100)
	Progress int

	// Error contains error information if the operation failed
	Error string
}

// DataOperationPhase represents the phase of a data operation
type DataOperationPhase string

const (
	DataOperationPhasePending   DataOperationPhase = "Pending"
	DataOperationPhaseRunning   DataOperationPhase = "Running"
	DataOperationPhaseCompleted DataOperationPhase = "Completed"
	DataOperationPhaseFailed    DataOperationPhase = "Failed"
	DataOperationPhaseCancelled DataOperationPhase = "Cancelled"
)
