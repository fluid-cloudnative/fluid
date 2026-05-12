package base

import (
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

// DataLifecycleManager defines interfaces for complete data lifecycle management
type DataLifecycleManager interface {
	LoadData(ctx cruntime.ReconcileRequestContext, spec *DataLoadSpec) (*DataLoadResult, error)
	GetDataOperationStatus(operationID string) (*DataOperationStatus, error)
	ProcessData(ctx cruntime.ReconcileRequestContext, spec *DataProcessSpec) (*DataProcessResult, error)
	MutateData(ctx cruntime.ReconcileRequestContext, spec *DataMutationSpec) (*DataMutationResult, error)
}

// --- Data Loading ---
type DataLoadSpec struct {
	Paths    []string
	Options  map[string]string
	Priority int
}

type DataLoadResult struct {
	OperationID string
	Status      DataOperationStatus
	LoadedBytes int64
	LoadedFiles int64
}

// --- Data Processing ---
type DataProcessSpec struct {
	Processor string
	Source    string
}

type DataProcessResult struct {
	OperationID string
	Status      DataOperationStatus
}

// --- Data Mutation ---
type DataMutationSpec struct {
	Action string
	Path   string
}

type DataMutationResult struct {
	OperationID string
	Status      DataOperationStatus
}

// --- Common Status ---
type DataOperationStatus struct {
	Phase    DataOperationPhase
	Message  string
	Progress int
	Error    string
}

type DataOperationPhase string

const (
	DataOperationPhasePending   DataOperationPhase = "Pending"
	DataOperationPhaseRunning   DataOperationPhase = "Running"
	DataOperationPhaseCompleted DataOperationPhase = "Completed"
	DataOperationPhaseFailed    DataOperationPhase = "Failed"
)