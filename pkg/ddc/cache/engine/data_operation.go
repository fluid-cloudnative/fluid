/*
  Copyright 2026 The Fluid Authors.

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

package engine

import (
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	fluiderrors "github.com/fluid-cloudnative/fluid/pkg/errors"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (e *CacheEngine) Operate(ctx cruntime.ReconcileRequestContext, opStatus *v1alpha1.OperationStatus, operation dataoperation.OperationInterface) (ctrl.Result, error) {

	r := base.EngineOperationReconciler{
		Engine:  e,
		CClient: e.Client,
	}

	// use default template engine
	return r.ReconcileOperation(ctx, opStatus, operation)
}

// CheckRuntimeReady checks if the runtime is ready
func (e *CacheEngine) CheckRuntimeReady() bool {
	// For CacheEngine, we check if both master and worker are ready
	runtime, err := e.getRuntime()
	if err != nil {
		return false
	}
	return runtime.Status.Master.Phase == v1alpha1.RuntimePhaseReady &&
		(runtime.Status.Worker.Phase == v1alpha1.RuntimePhaseReady || runtime.Status.Worker.Phase == v1alpha1.RuntimePhasePartialReady)
}

func (e *CacheEngine) GetDataOperationValueFile(ctx cruntime.ReconcileRequestContext, operation dataoperation.OperationInterface) (valueFileName string, err error) {

	operationType := operation.GetOperationType()
	object := operation.GetOperationObject()

	switch operationType {
	case dataoperation.DataLoadType:
		valueFileName, err = e.generateDataLoadValueFile(ctx, object)
		return valueFileName, err
	case dataoperation.DataProcessType:
		valueFileName, err = e.generateDataProcessValueFile(ctx, object)
		return valueFileName, err
	case dataoperation.DataMigrateType:
		valueFileName, err = e.generateDataMigrateValueFile(ctx, object)
		return valueFileName, err
	case dataoperation.DataBackupType:
		valueFileName, err = e.generateDataBackupValueFile(ctx, object)
		return valueFileName, err
	default:
		return "", fluiderrors.NewNotSupported(
			schema.GroupResource{
				Group:    object.GetObjectKind().GroupVersionKind().Group,
				Resource: object.GetObjectKind().GroupVersionKind().Kind,
			}, "CacheRuntime["+e.name+"]")
	}
}
