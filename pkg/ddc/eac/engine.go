/*
  Copyright 2022 The Fluid Authors.

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

package eac

import (
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type EACEngine struct {
	runtime     *datav1alpha1.EACRuntime
	name        string
	namespace   string
	runtimeType string
	runtimeInfo base.RuntimeInfoInterface
	UnitTest    bool
	Log         logr.Logger
	*ctrl.Helper
	client.Client
	//When reaching this gracefulShutdownLimits, the system is forced to clean up.
	gracefulShutdownLimits int32
	retryShutdown          int32
	// initImage              string
}

func Build(id string, ctx cruntime.ReconcileRequestContext) (base.Engine, error) {
	engine := &EACEngine{
		name:                   ctx.Name,
		namespace:              ctx.Namespace,
		Client:                 ctx.Client,
		Log:                    ctx.Log,
		runtimeType:            ctx.RuntimeType,
		gracefulShutdownLimits: 5,
		retryShutdown:          0,
	}
	err := engine.parseRuntime(ctx)
	if err != nil {
		return nil, err
	}

	// Build and setup runtime info
	runtimeInfo, err := engine.getRuntimeInfo()
	if err != nil {
		return nil, fmt.Errorf("engine %s failed to get runtime info", ctx.Name)
	}

	// Build the helper
	engine.Helper = ctrl.BuildHelper(runtimeInfo, ctx.Client, engine.Log)

	template := base.NewTemplateEngine(engine, id, ctx)

	err = kubeclient.EnsureNamespace(ctx.Client, ctx.Namespace)
	return template, err
}

// Precheck checks if the given key can be found in the current runtime types
func Precheck(client client.Client, key types.NamespacedName) (found bool, err error) {
	var obj datav1alpha1.GooseFSRuntime
	return utils.CheckObject(client, key, &obj)
}

func (e *EACEngine) parseRuntime(ctx cruntime.ReconcileRequestContext) error {
	if ctx.Runtime != nil {
		runtime, ok := ctx.Runtime.(*datav1alpha1.EACRuntime)
		if !ok {
			return fmt.Errorf("engine %s is failed to parse", ctx.Name)
		}
		e.runtime = runtime
	} else {
		return fmt.Errorf("engine %s is failed to parse", ctx.Name)
	}
	return nil
}

func (e *EACEngine) UsedStorageBytes() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) FreeStorageBytes() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) TotalStorageBytes() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) TotalFileNums() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) CheckMasterReady() (ready bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) CheckWorkersReady() (ready bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) ShouldSetupMaster() (should bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) ShouldSetupWorkers() (should bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) ShouldCheckUFS() (should bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) SetupMaster() (err error) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) SetupWorkers() (err error) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) UpdateDatasetStatus(phase datav1alpha1.DatasetPhase) (err error) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) PrepareUFS() (err error) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) ShouldUpdateUFS() (ufsToUpdate *utils.UFSToUpdate) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) UpdateOnUFSChange(ufsToUpdate *utils.UFSToUpdate) (ready bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) Shutdown() error {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) AssignNodesToCache(desiredNum int32) (currentNum int32, err error) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) CheckRuntimeHealthy() (err error) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) UpdateCacheOfDataset() (err error) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) CheckAndUpdateRuntimeStatus() (ready bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) CreateVolume() error {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) SyncReplicas(ctx cruntime.ReconcileRequestContext) error {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) SyncMetadata() (err error) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) DeleteVolume() (err error) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) BindToDataset() (err error) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) CreateDataLoadJob(ctx cruntime.ReconcileRequestContext, targetDataload datav1alpha1.DataLoad) error {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) CheckRuntimeReady() (ready bool) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) SyncRuntime(ctx cruntime.ReconcileRequestContext) (changed bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) CheckExistenceOfPath(targetDataload datav1alpha1.DataLoad) (notExist bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (e *EACEngine) SyncScheduleInfoToCacheNodes() (err error) {
	//TODO implement me
	panic("implement me")
}
