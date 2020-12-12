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
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

// Engine interface defines the interfaces that should be implemented
// by a distributed data caching Engine.
// Thread safety is required from implementations of this interface.
type Engine interface {
	// ID returns the id
	ID() string

	// Shutdown and clean up the engine
	Shutdown() error

	// Setup the engine
	Setup(ctx cruntime.ReconcileRequestContext) (ready bool, err error)

	// Setup the Volume
	CreateVolume() (err error)

	// Destroy the Volume
	DeleteVolume() (err error)

	// Sync syncs the alluxio runtime
	Sync(ctx cruntime.ReconcileRequestContext) error
}

// The real engine should implement
type Implement interface {
	UnderFileSystemService
	// Is the master ready
	CheckMasterReady() (ready bool, err error)
	// are the workers ready
	CheckWorkersReady() (ready bool, err error)
	// ShouldSetupMaster checks if we need setup the master
	ShouldSetupMaster() (should bool, err error)
	// ShouldSetupWorkers checks if we need setup the workers
	ShouldSetupWorkers() (should bool, err error)

	ShouldCheckUFS() (should bool, err error)
	// setup the cache master
	SetupMaster() (err error)
	// setup the cache worker
	SetupWorkers() (err error)
	// check if it's Bound to the dataset
	// IsBoundToDataset() (bound bool, err error)
	// Bind to the dataset
	UpdateDatasetStatus(phase datav1alpha1.DatasetPhase) (err error)
	// Prepare the mounts and metadata if it's not ready
	PrepareUFS() (err error)
	// Shutdown and clean up the engine
	Shutdown() error

	// AssignNodesToCache picks up the nodes for replicas
	AssignNodesToCache(desiredNum int32) (currentNum int32, err error)
	// CheckRuntimeHealthy checks runtime healthy
	CheckRuntimeHealthy() (err error)
	// UpdateCacheOfDataset updates cache of the dataset
	UpdateCacheOfDataset() (err error)
	// CheckAndUpdateRuntimeStatus checks and updates the status
	CheckAndUpdateRuntimeStatus() (ready bool, err error)

	CreateVolume() error

	// SyncReplicas syncs the replicas
	SyncReplicas(ctx cruntime.ReconcileRequestContext) error

	// SyncMetadata syncs all metadata from UFS
	SyncMetadata() (err error)

	// Destroy the Volume
	DeleteVolume() (err error)

	// BindToDataset binds the engine to dataset
	BindToDataset() (err error)
}

// UnderFileSystemService interface defines the interfaces that should be implemented
// by a underlayer fileSystem service for the data. The implementation is the underlayer file system connector.
// It is responsible for checking ufs and preload the data.
// Thread safety is required from implementations of this interface.
type UnderFileSystemService interface {
	UsedStorageBytes() (int64, error)

	FreeStorageBytes() (int64, error)

	TotalStorageBytes() (int64, error)

	TotalFileNums() (int64, error)
}
