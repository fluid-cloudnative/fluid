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

package base

import (
	"fmt"
	"reflect"
	"time"

	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

func (t *TemplateEngine) Operate(ctx cruntime.ReconcileRequestContext, object client.Object, opStatus *datav1alpha1.OperationStatus,
	operation dataoperation.OperationInterface) (ctrl.Result, error) {
	operateType := operation.GetOperationType()

	// runtime engine override the template engine
	switch operateType {
	case dataoperation.DataBackup:
		ownImpl, ok := t.Implement.(Databackuper)
		if ok {
			targetDataBackup, success := object.(*datav1alpha1.DataBackup)
			if !success {
				return utils.RequeueIfError(fmt.Errorf("object %v is not a DataBackup", object))
			}
			return ownImpl.BackupData(ctx, *targetDataBackup)
		}
	}

	// use default template engine
	switch opStatus.Phase {
	case common.PhaseNone:
		return t.reconcileNone(ctx, object, opStatus, operation)
	case common.PhasePending:
		return t.reconcilePending(ctx, object, opStatus, operation)
	case common.PhaseExecuting:
		return t.reconcileExecuting(ctx, object, opStatus, operation)
	case common.PhaseComplete:
		return t.reconcileComplete(ctx, object, opStatus, operation)
	case common.PhaseFailed:
		return t.reconcileFailed(ctx, object, opStatus, operation)
	default:
		ctx.Log.Error(fmt.Errorf("unknown phase"), "won't reconcile it", "phase", opStatus.Phase)
		return utils.NoRequeue()
	}
}

func (t *TemplateEngine) reconcileNone(ctx cruntime.ReconcileRequestContext, object client.Object,
	opStatus *datav1alpha1.OperationStatus, operation dataoperation.OperationInterface) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileNone")

	// 0. check the object spec valid or not
	conditions, err := operation.Validate(ctx, object)
	if err != nil {
		log.Error(err, "validate failed", "operationName", object.GetName(), "namespace", object.GetNamespace())
		ctx.Recorder.Event(object, v1.EventTypeWarning, common.DataOperationNotValid, err.Error())

		opStatus.Conditions = conditions
		opStatus.Phase = common.PhaseFailed
		if err = operation.UpdateOperationApiStatus(object, opStatus); err != nil {
			return utils.RequeueIfError(err)
		}
		return utils.RequeueImmediately()
	}

	// 1. update status to pending
	opStatus.Phase = common.PhasePending
	if len(opStatus.Conditions) == 0 {
		opStatus.Conditions = []datav1alpha1.Condition{}
	}
	opStatus.Duration = "Unfinished"
	opStatus.Infos = map[string]string{}
	if err = operation.UpdateOperationApiStatus(object, opStatus); err != nil {
		log.Error(err, fmt.Sprintf("failed to update the %s", operation.GetOperationType()))
		return utils.RequeueIfError(err)
	}
	log.V(1).Info(fmt.Sprintf("Update phase of the %s to Pending successfully", operation.GetOperationType()))
	return utils.RequeueImmediately()
}

func (t *TemplateEngine) reconcilePending(ctx cruntime.ReconcileRequestContext, object client.Object,
	opStatus *datav1alpha1.OperationStatus, operation dataoperation.OperationInterface) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcilePending")

	// 1. set current data operation to dataset
	err := SetDataOperationInTargetDataset(ctx, object, operation, t)
	if err != nil {
		return utils.RequeueAfterInterval(20 * time.Second)
	}

	log.Info("Set data operation on target dataset, try to update phase")
	opStatus.Phase = common.PhaseExecuting
	if err = operation.UpdateOperationApiStatus(object, opStatus); err != nil {
		log.Error(err, fmt.Sprintf("failed to update %s status to Executing, will retry", operation.GetOperationType()))
		return utils.RequeueIfError(err)
	}
	log.V(1).Info(fmt.Sprintf("update %s status to Executing successfully", operation.GetOperationType()))
	return utils.RequeueImmediately()
}

func (t *TemplateEngine) reconcileExecuting(ctx cruntime.ReconcileRequestContext, object client.Object,
	opStatus *datav1alpha1.OperationStatus, operation dataoperation.OperationInterface) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileExecuting")

	// 1. Install the helm chart if not exists
	err := InstallDataOperationHelmIfNotExist(ctx, object, operation, t.Implement)
	if err != nil {
		// runtime does not support current data operation, set status to failed
		if fluiderrs.IsNotSupported(err) {
			log.Error(err, "not support current data operation, set status to failed")
			ctx.Recorder.Eventf(object, v1.EventTypeWarning, common.DataOperationNotSupport,
				"RuntimeType %s not support %s", ctx.RuntimeType, operation.GetOperationType())

			opStatus.Phase = common.PhaseFailed
			if err = operation.UpdateOperationApiStatus(object, opStatus); err != nil {
				log.Error(err, "failed to update api status")
				return utils.RequeueIfError(err)
			}
			// goto failed case
			return utils.RequeueImmediately()
		}
		return utils.RequeueAfterInterval(20 * time.Second)
	}

	// 2. update data operation's status by helm status
	statusHandler := operation.GetStatusHandler(object)
	if statusHandler == nil {
		err = fmt.Errorf("fail to get status handler")
		log.Error(err, "status handler is nil")
		return utils.RequeueIfError(err)
	}
	opStatusToUpdate, err := statusHandler.GetOperationStatus(ctx, object, opStatus)
	if err != nil {
		log.Error(err, "failed to update status")
		return utils.RequeueIfError(err)
	}
	if !reflect.DeepEqual(opStatus, opStatusToUpdate) {
		if err = operation.UpdateOperationApiStatus(object, opStatusToUpdate); err != nil {
			log.Error(err, "failed to update api status")
			return utils.RequeueIfError(err)
		}
		if opStatusToUpdate.Phase != common.PhaseExecuting {
			log.V(1).Info(fmt.Sprintf("Update operation phase to %s", opStatusToUpdate.Phase), "opStatus", opStatusToUpdate)
			// return immediately if phase change
			return utils.RequeueImmediately()
		}
	}

	return utils.RequeueAfterInterval(20 * time.Second)
}

func (t *TemplateEngine) reconcileComplete(ctx cruntime.ReconcileRequestContext, object client.Object,
	opStatus *datav1alpha1.OperationStatus, operation dataoperation.OperationInterface) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileComplete")

	// 1. Update the infos field if complete
	if opStatus.Infos == nil {
		opStatus.Infos = map[string]string{}
	}
	// different data operation may set different key-value
	err := operation.UpdateStatusInfoForCompleted(object, opStatus.Infos)
	if err != nil {
		return utils.RequeueIfError(err)
	}

	if err = operation.UpdateOperationApiStatus(object, opStatus); err != nil {
		log.Error(err, fmt.Sprintf("failed to update the %s status", operation.GetOperationType()))
		return utils.RequeueIfError(err)
	}

	// 2. remove current data operation on target dataset if complete
	err = ReleaseTargetDataset(ctx, object, operation)
	if err != nil {
		return utils.RequeueIfError(err)
	}

	// 3. check and update data operation's status by helm status
	statusHandler := operation.GetStatusHandler(object)
	if statusHandler == nil {
		err := fmt.Errorf("fail to get status handler")
		log.Error(err, "status handler is nil")
		return utils.RequeueIfError(err)
	}
	opStatusToUpdate, err := statusHandler.GetOperationStatus(ctx, object, opStatus)
	if err != nil {
		log.Error(err, "failed to update status")
		return utils.RequeueIfError(err)
	}
	if !reflect.DeepEqual(opStatus, opStatusToUpdate) {
		if err = operation.UpdateOperationApiStatus(object, opStatusToUpdate); err != nil {
			log.Error(err, fmt.Sprintf("failed to update the %s status", operation.GetOperationType()))
			return utils.RequeueIfError(err)
		}
		if opStatusToUpdate.Phase != common.PhaseComplete {
			log.V(1).Info(fmt.Sprintf("Update operation phase to %s", opStatusToUpdate.Phase), "opStatus", opStatusToUpdate)
			// return immediately if not complete
			return utils.RequeueImmediately()
		}
	}

	// 4. record and no requeue
	log.Info(fmt.Sprintf("%s success, no need to requeue", operation.GetOperationType()))
	ctx.Recorder.Eventf(object, v1.EventTypeNormal, common.DataOperationSucceed,
		"%s %s succeeded", operation.GetOperationType(), object.GetName())
	return utils.NoRequeue()
}

func (t *TemplateEngine) reconcileFailed(ctx cruntime.ReconcileRequestContext, object client.Object,
	opStatus *datav1alpha1.OperationStatus, operation dataoperation.OperationInterface) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileFailed")

	// 1. remove current data operation on target dataset
	err := ReleaseTargetDataset(ctx, object, operation)
	if err != nil {
		return utils.RequeueIfError(err)
	}

	// 2. check and update data operation's status by helm status
	statusHandler := operation.GetStatusHandler(object)
	if statusHandler == nil {
		err := fmt.Errorf("fail to get status handler")
		log.Error(err, "status handler is nil")
		return utils.RequeueIfError(err)
	}
	opStatusToUpdate, err := statusHandler.GetOperationStatus(ctx, object, opStatus)
	if err != nil {
		log.Error(err, "failed to update status")
		return utils.RequeueIfError(err)
	}
	if !reflect.DeepEqual(opStatus, opStatusToUpdate) {
		if err = operation.UpdateOperationApiStatus(object, opStatusToUpdate); err != nil {
			log.Error(err, fmt.Sprintf("failed to update the %s status", operation.GetOperationType()))
			return utils.RequeueIfError(err)
		}
		if opStatusToUpdate.Phase != common.PhaseFailed {
			log.V(1).Info(fmt.Sprintf("Update operation phase to %s", opStatusToUpdate.Phase), "opStatus", opStatusToUpdate)
			// return immediately if not failed
			return utils.RequeueImmediately()
		}
	}

	// 2. record and no requeue
	log.Info(fmt.Sprintf("%s failed, won't requeue", operation.GetOperationType()))
	ctx.Recorder.Eventf(object, v1.EventTypeWarning, common.DataOperationFailed, "%s %s failed", operation.GetOperationType(), object.GetName())
	return utils.NoRequeue()
}
