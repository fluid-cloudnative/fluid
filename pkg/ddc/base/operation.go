/*
Copyright 2023 The Fluid Authors.

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
	"context"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"reflect"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
)

const cleanupErrorMsg = "Failed to get remaining time to clean up for operation %s"

func (t *TemplateEngine) Operate(ctx cruntime.ReconcileRequestContext, opStatus *datav1alpha1.OperationStatus,
	operation dataoperation.OperationInterface) (ctrl.Result, error) {

	// we can do customized runtime engine override the template engine, implement if needed.

	// use default template engine
	switch opStatus.Phase {
	case common.PhaseNone:
		return t.reconcileNone(ctx, opStatus, operation)
	case common.PhasePending:
		return t.reconcilePending(ctx, opStatus, operation)
	case common.PhaseExecuting:
		return t.reconcileExecuting(ctx, opStatus, operation)
	case common.PhaseComplete:
		return t.reconcileComplete(ctx, opStatus, operation)
	case common.PhaseFailed:
		return t.reconcileFailed(ctx, opStatus, operation)
	default:
		ctx.Log.Error(fmt.Errorf("unknown phase"), "won't reconcile it", "phase", opStatus.Phase)
		return utils.NoRequeue()
	}
}

func (t *TemplateEngine) reconcileNone(ctx cruntime.ReconcileRequestContext, opStatus *datav1alpha1.OperationStatus,
	operation dataoperation.OperationInterface) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileNone")

	// 0. check the object spec valid or not
	conditions, err := operation.Validate(ctx)
	if err != nil {
		object := operation.GetOperationObject()
		log.Error(err, "validate failed", "operationName", object.GetName(), "namespace", object.GetNamespace())
		ctx.Recorder.Event(object, v1.EventTypeWarning, common.DataOperationNotValid, err.Error())

		opStatus.Conditions = conditions
		opStatus.Phase = common.PhaseFailed
		if err = operation.UpdateOperationApiStatus(opStatus); err != nil {
			return utils.RequeueIfError(err)
		}
		// update operation status would trigger requeue, no need to requeue here
		return utils.NoRequeue()
	}

	// 1. update status to pending
	opStatus.Phase = common.PhasePending
	if len(opStatus.Conditions) == 0 {
		opStatus.Conditions = []datav1alpha1.Condition{}
	}
	opStatus.Duration = "Unfinished"
	opStatus.Infos = map[string]string{}

	if operation.HasPrecedingOperation() {
		opStatus.WaitingFor.OperationComplete = ptr.To(true)
	}

	if err = operation.UpdateOperationApiStatus(opStatus); err != nil {
		log.Error(err, fmt.Sprintf("failed to update the %s", operation.GetOperationType()))
		return utils.RequeueIfError(err)
	}
	log.V(1).Info(fmt.Sprintf("Update phase of the %s to Pending successfully", operation.GetOperationType()))
	// update opreation status would trigger requeue, no need to requeue here
	return utils.NoRequeue()
}

func (t *TemplateEngine) reconcilePending(ctx cruntime.ReconcileRequestContext, opStatus *datav1alpha1.OperationStatus,
	operation dataoperation.OperationInterface) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcilePending")

	// 1. check preceding operation status
	if opStatus.WaitingFor.OperationComplete != nil && *opStatus.WaitingFor.OperationComplete {
		// when operationComplete set back to false, a new reconciliation loop will be triggered, so no requeue here.
		return utils.NoRequeue()
	}

	// 2. set current data operation to dataset
	err := SetDataOperationInTargetDataset(ctx, operation, t)
	if err != nil {
		return utils.RequeueAfterInterval(20 * time.Second)
	}

	log.Info("Set data operation on target dataset, try to update phase")
	opStatus.Phase = common.PhaseExecuting
	if err = operation.UpdateOperationApiStatus(opStatus); err != nil {
		log.Error(err, fmt.Sprintf("failed to update %s status to Executing, will retry", operation.GetOperationType()))
		return utils.RequeueIfError(err)
	}
	log.V(1).Info(fmt.Sprintf("update %s status to Executing successfully", operation.GetOperationType()))
	// update operation status would trigger requeue, no need to requeue here
	return utils.NoRequeue()
}

func (t *TemplateEngine) reconcileExecuting(ctx cruntime.ReconcileRequestContext, opStatus *datav1alpha1.OperationStatus,
	operation dataoperation.OperationInterface) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileExecuting")

	// 1. Install the helm chart if not exists
	err := InstallDataOperationHelmIfNotExist(ctx, operation, t.Implement)
	if err != nil {
		object := operation.GetOperationObject()
		// runtime does not support current data operation, set status to failed
		if fluiderrs.IsNotSupported(err) {
			log.Error(err, "not support current data operation, set status to failed")
			ctx.Recorder.Eventf(object, v1.EventTypeWarning, common.DataOperationNotSupport,
				"RuntimeType %s not support %s", ctx.RuntimeType, operation.GetOperationType())

			opStatus.Phase = common.PhaseFailed
			if err = operation.UpdateOperationApiStatus(opStatus); err != nil {
				log.Error(err, "failed to update api status")
				return utils.RequeueIfError(err)
			}
			// goto failed case
			// opreation status updated would trigger requeue, no need to requeue here
			return utils.NoRequeue()
		}
		ctx.Recorder.Eventf(object, v1.EventTypeWarning, common.DataOperationExecutionFailed, "fail to execute data operation: %v", err)
		return utils.RequeueAfterInterval(20 * time.Second)
	}

	// 2. update data operation's status by helm status
	statusHandler := operation.GetStatusHandler()
	if statusHandler == nil {
		err = fmt.Errorf("fail to get status handler")
		log.Error(err, "status handler is nil")
		return utils.RequeueIfError(err)
	}
	opStatusToUpdate, err := statusHandler.GetOperationStatus(ctx, opStatus)
	if err != nil {
		log.Error(err, "failed to update status")
		return utils.RequeueIfError(err)
	}
	if !reflect.DeepEqual(opStatus, opStatusToUpdate) {
		if err = operation.UpdateOperationApiStatus(opStatusToUpdate); err != nil {
			log.Error(err, "failed to update api status")
			return utils.RequeueIfError(err)
		}
		log.V(1).Info(fmt.Sprintf("update operation status to %s successfully", opStatusToUpdate.Phase), "opstatus", opStatusToUpdate)
	}

	return utils.RequeueAfterInterval(20 * time.Second)
}

func (t *TemplateEngine) reconcileComplete(ctx cruntime.ReconcileRequestContext, opStatus *datav1alpha1.OperationStatus,
	operation dataoperation.OperationInterface) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileComplete")

	// 0. clean up if ttl after finished expired
	var ttl *time.Duration
	if utils.NeedCleanUp(opStatus, operation) {
		var err error
		ttl, err = t.processTTL(opStatus, operation, log, ctx)
		if err != nil {
			log.Error(err, fmt.Sprintf(cleanupErrorMsg, operation.GetOperationType()))
			return utils.RequeueIfError(err)
		} else if ttl != nil && *ttl <= 0 {
			return utils.NoRequeue()
		}
	}

	// 1. Update the infos field if complete
	if opStatus.Infos == nil {
		opStatus.Infos = map[string]string{}
	}
	// different data operation may set different key-value
	err := operation.UpdateStatusInfoForCompleted(opStatus.Infos)
	if err != nil {
		return utils.RequeueIfError(err)
	}

	if err = operation.UpdateOperationApiStatus(opStatus); err != nil {
		log.Error(err, fmt.Sprintf("failed to update the %s status", operation.GetOperationType()))
		return utils.RequeueIfError(err)
	}

	// 2. remove current data operation on target dataset if complete
	err = ReleaseTargetDataset(ctx, operation)
	if err != nil {
		return utils.RequeueIfError(err)
	}

	// 3. check and update data operation's status by helm status
	statusHandler := operation.GetStatusHandler()
	if statusHandler == nil {
		err := fmt.Errorf("fail to get status handler")
		log.Error(err, "status handler is nil")
		return utils.RequeueIfError(err)
	}
	opStatusToUpdate, err := statusHandler.GetOperationStatus(ctx, opStatus)
	if err != nil {
		log.Error(err, "failed to update status")
		return utils.RequeueIfError(err)
	}
	if !reflect.DeepEqual(opStatus, opStatusToUpdate) {
		if err = operation.UpdateOperationApiStatus(opStatusToUpdate); err != nil {
			log.Error(err, fmt.Sprintf("failed to update the %s status", operation.GetOperationType()))
			return utils.RequeueIfError(err)
		}
		log.V(1).Info(fmt.Sprintf("update operation status to %s successfully", opStatusToUpdate.Phase), "opstatus", opStatusToUpdate)
	}

	// 4. record and no requeue
	// For cron operations, the phase may be updated to pending here, and we only log bellow messages in complete phase
	if opStatusToUpdate.Phase == common.PhaseComplete {
		object := operation.GetOperationObject()
		ctx.Recorder.Eventf(object, v1.EventTypeNormal, common.DataOperationSucceed,
			"%s %s succeeded", operation.GetOperationType(), object.GetName())

		// scale the statefulset replicas to 0 for parallel data operation
		// if the status not Complete, there is a new starting job, not scale the statefulset to zero.
		if operation.GetParallelTaskNumber() > 1 {
			releaseNameSpacedName := operation.GetReleaseNameSpacedName()
			err = kubeclient.ScaleCacheWorkerSet(t.Client, utils.GetParallelOperationWorkersName(releaseNameSpacedName.Name), releaseNameSpacedName.Namespace, 0)
			if err != nil {
				return utils.RequeueIfError(err)
			}
		}
	}

	// 5. Requeue if data operation set ttl after finished and has not expired
	if ttl != nil && *ttl > 0 {
		log.V(1).Info("requeue after remaining time to clean up data operation", "timeToLive", ttl)
		return utils.RequeueAfterInterval(*ttl)
	}

	return utils.NoRequeue()
}

// processTTL processes the operations that need to be cleaned up based on the TTL.
func (t *TemplateEngine) processTTL(opStatus *datav1alpha1.OperationStatus, operation dataoperation.OperationInterface, log logr.Logger, ctx cruntime.ReconcileRequestContext) (ttl *time.Duration, err error) {
	// Get the remaining time to clean up for the operation.
	ttl, err = utils.Timeleft(opStatus, operation)
	if err != nil {
		log.Error(err, fmt.Sprintf(cleanupErrorMsg, operation.GetOperationType()))
		return
	}

	// If the remaining time is not nil and less than or equal to 0, clean up the data operation.
	if ttl != nil && *ttl <= 0 {
		if err = ctx.Client.Delete(context.TODO(), operation.GetOperationObject()); err != nil && utils.IgnoreNotFound(err) != nil {
			log.Error(err, "Failed to clean up data operation %s", operation.GetOperationType())
			return
		}
		log.Info("Data operation has been clean up")
	}
	return
}

func (t *TemplateEngine) reconcileFailed(ctx cruntime.ReconcileRequestContext, opStatus *datav1alpha1.OperationStatus, operation dataoperation.OperationInterface) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileFailed")

	// 0. clean up if ttl after finished expired
	var ttl *time.Duration
	if utils.NeedCleanUp(opStatus, operation) {
		var err error
		ttl, err = t.processTTL(opStatus, operation, log, ctx)
		if err != nil {
			log.Error(err, fmt.Sprintf(cleanupErrorMsg, operation.GetOperationType()))
			return utils.RequeueIfError(err)
		} else if ttl != nil && *ttl <= 0 {
			return utils.NoRequeue()
		}
	}

	// 1. remove current data operation on target dataset
	err := ReleaseTargetDataset(ctx, operation)
	if err != nil {
		return utils.RequeueIfError(err)
	}

	// 2. check and update data operation's status by helm status
	statusHandler := operation.GetStatusHandler()
	if statusHandler == nil {
		err := fmt.Errorf("fail to get status handler")
		log.Error(err, "status handler is nil")
		return utils.RequeueIfError(err)
	}
	opStatusToUpdate, err := statusHandler.GetOperationStatus(ctx, opStatus)
	if err != nil {
		log.Error(err, "failed to update status")
		return utils.RequeueIfError(err)
	}
	if !reflect.DeepEqual(opStatus, opStatusToUpdate) {
		if err = operation.UpdateOperationApiStatus(opStatusToUpdate); err != nil {
			log.Error(err, fmt.Sprintf("failed to update the %s status", operation.GetOperationType()))
			return utils.RequeueIfError(err)
		}
		log.V(1).Info(fmt.Sprintf("update operation status to %s successfully", opStatusToUpdate.Phase), "opstatus", opStatusToUpdate)
	}

	// 2. record and no requeue
	// For cron operations, the phase may be updated to pending here, and we only log bellow messages in failed phase
	if opStatusToUpdate.Phase == common.PhaseFailed {
		object := operation.GetOperationObject()
		ctx.Recorder.Eventf(object, v1.EventTypeWarning, common.DataOperationFailed, "%s %s failed", operation.GetOperationType(), object.GetName())

		// scale the statefulset replicas to 0 for parallel data operation
		// if the status not PhaseFailed, there is a new starting job, not scale the statefulset to zero.
		if operation.GetParallelTaskNumber() > 1 {
			releaseNameSpacedName := operation.GetReleaseNameSpacedName()
			err = kubeclient.ScaleCacheWorkerSet(t.Client, utils.GetParallelOperationWorkersName(releaseNameSpacedName.Name), releaseNameSpacedName.Namespace, 0)
			if err != nil {
				return utils.RequeueIfError(err)
			}
		}
	}

	// 3. Requeue if data operation set ttl after finished and has not expired
	if ttl != nil && *ttl > 0 {
		log.V(1).Info("get remaining time to clean up data operation", "timeToLive", ttl)
		return utils.RequeueAfterInterval(*ttl)
	}
	return utils.NoRequeue()
}
