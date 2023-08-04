package dataflow

import (
	"context"
	"reflect"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	utilpointer "k8s.io/utils/pointer"
)

func reconcileDataLoad(ctx reconcileRequestContext) (needRequeue bool, err error) {
	dataLoad, err := utils.GetDataLoad(ctx.Client, ctx.Name, ctx.Namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.V(1).Info("Data operation not found, skip reconciling")
			return false, nil
		}
		return true, errors.Wrap(err, "failed to get dataload")
	}

	if dataLoad.Status.WaitingFor.OperationComplete == nil || !*dataLoad.Status.WaitingFor.OperationComplete {
		ctx.Log.Info("Data operation not waiting for any other operation, skip reconciling")
		return false, nil
	}

	precedingOpStatus, err := utils.GetPrecedingOperationStatus(ctx.Client, dataLoad.Spec.RunAfter)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			// preceding operation not found
			ctx.Recorder.Eventf(dataLoad, corev1.EventTypeWarning, common.DataOperationNotFound, "Preceding operation %s \"%s/%s\" not found",
				dataLoad.Spec.RunAfter.OperationKind,
				dataLoad.Spec.RunAfter.Namespace,
				dataLoad.Spec.RunAfter.Name)
			return true, nil
		}
		return true, errors.Wrapf(err, "failed to get preceding operation status (DataLoad.Spec.RunAfter: %v)", dataLoad.Spec.RunAfter)
	}

	if precedingOpStatus != nil && precedingOpStatus.Phase != common.PhaseComplete {
		ctx.Recorder.Eventf(dataLoad, corev1.EventTypeNormal, common.DataOperationWaiting, "Waiting for operation %s \"%s/%s\" to complete",
			dataLoad.Spec.RunAfter.OperationKind,
			dataLoad.Spec.RunAfter.Namespace,
			dataLoad.Spec.RunAfter.Name)
		return true, nil
	}

	// set opStatus.waitingFor.operationComplete back to false
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		tmpDataLoad, err := utils.GetDataLoad(ctx.Client, ctx.Name, ctx.Namespace)
		if err != nil {
			if utils.IgnoreNotFound(err) == nil {
				return nil
			}
			return err
		}

		toUpdate := tmpDataLoad.DeepCopy()
		toUpdate.Status.WaitingFor.OperationComplete = utilpointer.Bool(false)
		if !reflect.DeepEqual(toUpdate.Status, tmpDataLoad.Status) {
			return ctx.Client.Status().Update(context.TODO(), toUpdate)
		}

		return nil
	})

	if err != nil {
		return true, errors.Wrapf(err, "failed to update operation DataLoad status waitingFor.OperationComplete=false")
	}

	return false, nil
}

func reconcileDataBackup(ctx reconcileRequestContext) (needRequeue bool, err error) {
	return false, nil
}

func reconcileDataMigrate(ctx reconcileRequestContext) (needRequeue bool, err error) {
	return false, nil
}

func reconcileDataProcess(ctx reconcileRequestContext) (needRequeue bool, err error) {
	return false, nil
}
