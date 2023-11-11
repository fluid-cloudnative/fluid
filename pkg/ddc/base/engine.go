/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package base

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	ctrl "sigs.k8s.io/controller-runtime"
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

	// Dataloader
	// @Deprecated use DataOperator instead.
	Dataloader

	// Datamigrater
	// @Deprecated use DataOperator instead.
	Datamigrater

	// DataOperator is a common interface for Data Operations like DataBackup/DataLoad/DataMigrate etc.
	DataOperator
}

// DataOperator is a common interface of TemplateEngine for Data Operations like DataBackup/DataLoad/DataMigrate etc.
type DataOperator interface {
	Operate(ctx cruntime.ReconcileRequestContext, opStatus *datav1alpha1.OperationStatus, operation dataoperation.OperationInterface) (ctrl.Result, error)
}

// DataOperatorYamlGenerator is the implementation of DataOperator interface for runtime engine
type DataOperatorYamlGenerator interface {
	GetDataOperationValueFile(ctx cruntime.ReconcileRequestContext, operation dataoperation.OperationInterface) (valueFileName string, err error)
}

type Dataloader interface {
	// LoadData generate dataload values and install helm chart
	LoadData(ctx cruntime.ReconcileRequestContext, targetDataload datav1alpha1.DataLoad) (err error)

	// CheckRuntimeReady Check if runtime is ready
	// @Deprecated because it's common for all engine
	CheckRuntimeReady() (ready bool)

	// CheckExistenceOfPath Check existence of path
	// @Deprecated as https://github.com/fluid-cloudnative/fluid/pull/2355
	CheckExistenceOfPath(targetDataload datav1alpha1.DataLoad) (notExist bool, err error)
}

type Databackuper interface {
	BackupData(ctx cruntime.ReconcileRequestContext, targetDataBackup datav1alpha1.DataBackup) (ctrl.Result, error)
}

type Datamigrater interface {
	// MigrateData generate datamigrate values and install helm chart
	MigrateData(ctx cruntime.ReconcileRequestContext, targetDataMigrate datav1alpha1.DataMigrate) (err error)
}

// Implement is what the real engine should implement if it use the TemplateEngine
type Implement interface {
	UnderFileSystemService

	DataOperatorYamlGenerator

	// ShouldSetupMaster checks if the master ready
	CheckMasterReady() (ready bool, err error)

	// CheckWorkersReady checks if the workers ready
	CheckWorkersReady() (ready bool, err error)

	// ShouldSetupMaster checks if we need to setup the master
	ShouldSetupMaster() (should bool, err error)

	// ShouldSetupWorkers checks if we need to setup the workers
	ShouldSetupWorkers() (should bool, err error)

	// ShouldCheckUFS checks if we should check the ufs
	ShouldCheckUFS() (should bool, err error)

	// SetupMaster setup the cache master
	SetupMaster() (err error)

	// SetupWorkers setup the cache worker
	SetupWorkers() (err error)

	// UpdateDatasetStatus update the status of Dataset according to the given phase
	UpdateDatasetStatus(phase datav1alpha1.DatasetPhase) (err error)

	// PrepareUFS prepare the mounts and metadata if it's not ready
	PrepareUFS() (err error)

	// ShouldUpdateUFS check if we need to update the ufs and return all ufs to update
	// If the ufs have changed and the engine supports add/remove mount points dynamically,
	// then we need to UpdateOnUFSChange
	ShouldUpdateUFS() (ufsToUpdate *utils.UFSToUpdate)

	// UpdateOnUFSChange update the mount point of Dataset if ufs change
	// if an engine doesn't support UpdateOnUFSChange, it need to return false
	UpdateOnUFSChange(ufsToUpdate *utils.UFSToUpdate) (ready bool, err error)

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

	// CreateVolume create the pv and pvc for the Dataset
	CreateVolume() error

	// SyncReplicas syncs the replicas
	SyncReplicas(ctx cruntime.ReconcileRequestContext) error

	// SyncMetadata syncs all metadata from UFS
	SyncMetadata() (err error)

	// DeleteVolume Destroy the Volume
	DeleteVolume() (err error)

	// BindToDataset binds the engine to dataset
	BindToDataset() (err error)

	// CreateDataLoadJob creates the job to load data
	// @Deprecated TODO: remove when DataOperator ready
	CreateDataLoadJob(ctx cruntime.ReconcileRequestContext, targetDataload datav1alpha1.DataLoad) error

	// CreateDataMigrateJob creates the job to load data
	// @Deprecated TODO: remove when DataOperator ready
	CreateDataMigrateJob(ctx cruntime.ReconcileRequestContext, targetDataMigrate datav1alpha1.DataMigrate) error

	// checks if the runtime is ready
	CheckRuntimeReady() (ready bool)

	// SyncRuntime syncs the runtime spec
	SyncRuntime(ctx cruntime.ReconcileRequestContext) (changed bool, err error)

	// CheckExistenceOfPath Check existence of targetDataload path
	// useless as https://github.com/fluid-cloudnative/fluid/pull/2355
	// @Deprecated TODO: remove when DataOperator ready
	CheckExistenceOfPath(targetDataload datav1alpha1.DataLoad) (notExist bool, err error)

	// Sync the scheduleInfo to cacheNodes
	SyncScheduleInfoToCacheNodes() (err error)
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
