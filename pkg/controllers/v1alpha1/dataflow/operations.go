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

package dataflow

import (
	"context"
	"reflect"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func reconcileDataLoad(ctx reconcileRequestContext) (needRequeue bool, err error) {
	dataLoad, err := utils.GetDataLoad(ctx.Client, ctx.Name, ctx.Namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.V(1).Info("DataLoad not found, skip reconciling")
			return false, nil
		}
		return true, errors.Wrap(err, "failed to get dataload")
	}

	updateStatusFn := func() error {
		tmp, err := utils.GetDataLoad(ctx.Client, ctx.Name, ctx.Namespace)
		if err != nil {
			if utils.IgnoreNotFound(err) == nil {
				return nil
			}
			return err
		}

		toUpdate := tmp.DeepCopy()
		toUpdate.Status.WaitingFor.OperationComplete = ptr.To(false)
		if !reflect.DeepEqual(toUpdate.Status, tmp.Status) {
			return ctx.Client.Status().Update(context.TODO(), toUpdate)
		}

		return nil
	}

	return reconcileOperationDataFlow(ctx, dataLoad, dataLoad.Spec.RunAfter, dataLoad.Status, updateStatusFn)
}

func reconcileDataBackup(ctx reconcileRequestContext) (needRequeue bool, err error) {
	dataBackup, err := utils.GetDataBackup(ctx.Client, ctx.Name, ctx.Namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.V(1).Info("DataBackup not found, skip reconciling")
			return false, nil
		}
		return true, errors.Wrap(err, "failed to get databackup")
	}

	updateStatusFn := func() error {
		tmp, err := utils.GetDataBackup(ctx.Client, ctx.Name, ctx.Namespace)
		if err != nil {
			if utils.IgnoreNotFound(err) == nil {
				return nil
			}
			return err
		}

		toUpdate := tmp.DeepCopy()
		toUpdate.Status.WaitingFor.OperationComplete = ptr.To(false)
		if !reflect.DeepEqual(toUpdate.Status, tmp.Status) {
			return ctx.Client.Status().Update(context.TODO(), toUpdate)
		}

		return nil
	}

	return reconcileOperationDataFlow(ctx, dataBackup, dataBackup.Spec.RunAfter, dataBackup.Status, updateStatusFn)
}

func reconcileDataMigrate(ctx reconcileRequestContext) (needRequeue bool, err error) {
	dataMigrate, err := utils.GetDataMigrate(ctx.Client, ctx.Name, ctx.Namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.V(1).Info("DataMigrate not found, skip reconciling")
			return false, nil
		}
		return true, errors.Wrap(err, "failed to get datamigrate")
	}

	updateStatusFn := func() error {
		tmp, err := utils.GetDataMigrate(ctx.Client, ctx.Name, ctx.Namespace)
		if err != nil {
			if utils.IgnoreNotFound(err) == nil {
				return nil
			}
			return err
		}

		toUpdate := tmp.DeepCopy()
		toUpdate.Status.WaitingFor.OperationComplete = ptr.To(false)
		if !reflect.DeepEqual(toUpdate.Status, tmp.Status) {
			return ctx.Client.Status().Update(context.TODO(), toUpdate)
		}

		return nil
	}

	return reconcileOperationDataFlow(ctx, dataMigrate, dataMigrate.Spec.RunAfter, dataMigrate.Status, updateStatusFn)
}

func reconcileDataProcess(ctx reconcileRequestContext) (needRequeue bool, err error) {
	dataProcess, err := utils.GetDataProcess(ctx.Client, ctx.Name, ctx.Namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.V(1).Info("DataMigrate not found, skip reconciling")
			return false, nil
		}
		return true, errors.Wrap(err, "failed to get datamigrate")
	}

	updateStatusFn := func() error {
		tmp, err := utils.GetDataProcess(ctx.Client, ctx.Name, ctx.Namespace)
		if err != nil {
			if utils.IgnoreNotFound(err) == nil {
				return nil
			}
			return err
		}

		toUpdate := tmp.DeepCopy()
		toUpdate.Status.WaitingFor.OperationComplete = ptr.To(false)
		if !reflect.DeepEqual(toUpdate.Status, tmp.Status) {
			return ctx.Client.Status().Update(context.TODO(), toUpdate)
		}

		return nil
	}

	return reconcileOperationDataFlow(ctx, dataProcess, dataProcess.Spec.RunAfter, dataProcess.Status, updateStatusFn)
}

func reconcileOperationDataFlow(ctx reconcileRequestContext,
	object client.Object,
	runAfter *datav1alpha1.OperationRef,
	opStatus datav1alpha1.OperationStatus,
	updateStatusFn func() error) (needRequeue bool, err error) {

	opRefNamespace := ctx.Namespace
	if len(runAfter.Namespace) != 0 {
		opRefNamespace = runAfter.Namespace
	}

	precedingOpStatus, err := utils.GetPrecedingOperationStatus(ctx.Client, runAfter, opRefNamespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			// preceding operation not found
			ctx.Recorder.Eventf(object, corev1.EventTypeWarning, common.DataOperationNotFound, "Preceding operation %s \"%s/%s\" not found",
				runAfter.Kind,
				opRefNamespace,
				runAfter.Name)
			return true, nil
		}
		return true, errors.Wrapf(err, "failed to get preceding operation status (DataLoad.Spec.RunAfter: %v)", runAfter)
	}

	if precedingOpStatus != nil && precedingOpStatus.Phase != common.PhaseComplete {
		ctx.Recorder.Eventf(object, corev1.EventTypeNormal, common.DataOperationWaiting, "Waiting for operation %s \"%s/%s\" to complete",
			runAfter.Kind,
			opRefNamespace,
			runAfter.Name)
		return true, nil
	}

	// set opStatus.waitingFor.operationComplete back to false
	err = retry.RetryOnConflict(retry.DefaultBackoff, updateStatusFn)

	if err != nil {
		return true, errors.Wrapf(err, "failed to update operation status waitingFor.OperationComplete=false")
	}

	return false, nil
}
