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

package dataprocess

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OnceStatusHandler struct {
	client.Client
	dataProcess *datav1alpha1.DataProcess
}

var _ dataoperation.StatusHandler = &OnceStatusHandler{}

// GetOperationStatus get operation status according to helm chart status
func (handler *OnceStatusHandler) GetOperationStatus(ctx runtime.ReconcileRequestContext, opStatus *datav1alpha1.OperationStatus) (result *datav1alpha1.OperationStatus, err error) {
	result = opStatus.DeepCopy()
	object := handler.dataProcess

	releaseName := utils.GetDataProcessReleaseName(object.GetName())
	jobName := utils.GetDataProcessJobName(releaseName)

	ctx.Log.V(1).Info("DataProcess chart already existed, check its running status")
	job, err := kubeclient.GetJob(handler.Client, jobName, ctx.Namespace)
	if err != nil {
		// In case of NotFound error
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.Info("Related job missing, will delete helm chart and retry", "namespace", ctx.Namespace, "jobName", jobName)
			if err = helm.DeleteReleaseIfExists(releaseName, ctx.Namespace); err != nil {
				ctx.Log.Error(err, "failed to delete dataprocess helm release", "namespace", ctx.Namespace, "releaseName", releaseName)
				return
			}
		}

		// In cases of other error
		ctx.Log.Error(err, "can't get dataprocess job", "namespace", ctx.Namespace, "jobName", jobName)
		return
	}

	finishedJobCondition := kubeclient.GetFinishedJobCondition(job)
	if finishedJobCondition == nil {
		ctx.Log.V(1).Info("DataProcess job still running", "namespace", ctx.Namespace, "jobName", jobName)
		return
	}
	isJobSucceed := finishedJobCondition.Type == batchv1.JobComplete

	// job either failed or complete, update DataLoad's phase status
	jobCondition := job.Status.Conditions[0]

	result.Conditions = []datav1alpha1.Condition{
		{
			Type:               common.ConditionType(jobCondition.Type),
			Status:             jobCondition.Status,
			Reason:             jobCondition.Reason,
			Message:            jobCondition.Message,
			LastProbeTime:      jobCondition.LastProbeTime,
			LastTransitionTime: jobCondition.LastTransitionTime,
		},
	}

	if isJobSucceed {
		result.Phase = common.PhaseComplete
	} else {
		result.Phase = common.PhaseFailed
	}
	result.Duration = utils.CalculateDuration(job.CreationTimestamp.Time, jobCondition.LastTransitionTime.Time)

	return
}

type CronStatusHandler struct {
	client.Client
	dataProcess *datav1alpha1.DataProcess
}

var _ dataoperation.StatusHandler = &CronStatusHandler{}

func (handler *CronStatusHandler) GetOperationStatus(ctx runtime.ReconcileRequestContext, opStatus *datav1alpha1.OperationStatus) (result *datav1alpha1.OperationStatus, err error) {
	result = opStatus.DeepCopy()
	object := handler.dataProcess

	releaseName := utils.GetDataProcessReleaseName(object.GetName())
	cronjobName := utils.GetDataProcessJobName(releaseName)

	cronjobStatus, err := kubeclient.GetCronJobStatus(handler.Client, types.NamespacedName{Namespace: object.GetNamespace(), Name: cronjobName})
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.Info("Related CronJob missing, will delete helm chart and retry", "namespace", ctx.Namespace, "cronjobName", cronjobName)
			if err = helm.DeleteReleaseIfExists(releaseName, ctx.Namespace); err != nil {
				ctx.Log.Error(err, "failed to delete dataprocess helm release", "namespace", ctx.Namespace, "releaseName", releaseName)
				return
			}
			return
		}
		ctx.Log.Error(err, "can't get dataprocess cronjob", "namespace", ctx.Namespace, "cronjobName", cronjobName)
		return
	}

	if cronjobStatus.LastScheduleTime == nil {
		ctx.Log.Info("CronJob has not been scheduled yet", "namespace", ctx.Namespace, "cronjobName", cronjobName)
		return
	}

	result.LastScheduleTime = cronjobStatus.LastScheduleTime
	result.LastSuccessfulTime = cronjobStatus.LastSuccessfulTime

	jobs, err := utils.ListDataOperationJobByCronjob(handler.Client, types.NamespacedName{Namespace: object.GetNamespace(), Name: cronjobName})
	if err != nil {
		ctx.Log.Error(err, "can't list dataprocess jobs by cronjob", "namespace", ctx.Namespace, "cronjobName", cronjobName)
		return
	}

	var currentJob *batchv1.Job
	for _, job := range jobs {
		if job.CreationTimestamp == *cronjobStatus.LastScheduleTime || job.CreationTimestamp.After(cronjobStatus.LastScheduleTime.Time) {
			currentJob = &job
			break
		}
	}
	if currentJob == nil {
		ctx.Log.Info("can't get newest job by cronjob, skip", "namespace", ctx.Namespace, "cronjobName", cronjobName)
		return
	}

	finishedJobCondition := kubeclient.GetFinishedJobCondition(currentJob)
	if finishedJobCondition == nil {
		ctx.Log.V(1).Info("DataProcess job still running", "namespace", ctx.Namespace, "cronjobName", cronjobName)
		if opStatus.Phase == common.PhaseComplete || opStatus.Phase == common.PhaseFailed {
			result.Phase = common.PhasePending
			result.Duration = "-"
		}
		return
	}

	result.Conditions = []datav1alpha1.Condition{
		{
			Type:               common.ConditionType(finishedJobCondition.Type),
			Status:             finishedJobCondition.Status,
			Reason:             finishedJobCondition.Reason,
			Message:            finishedJobCondition.Message,
			LastProbeTime:      finishedJobCondition.LastProbeTime,
			LastTransitionTime: finishedJobCondition.LastTransitionTime,
		},
	}
	if finishedJobCondition.Type == batchv1.JobFailed {
		result.Phase = common.PhaseFailed
	} else {
		result.Phase = common.PhaseComplete
	}
	result.Duration = utils.CalculateDuration(currentJob.CreationTimestamp.Time, finishedJobCondition.LastTransitionTime.Time)
	return
}

type OnEventStatusHandler struct {
	client.Client
	dataProcess *datav1alpha1.DataProcess
}

var _ dataoperation.StatusHandler = &OnEventStatusHandler{}

func (handler *OnEventStatusHandler) GetOperationStatus(ctx runtime.ReconcileRequestContext, opStatus *datav1alpha1.OperationStatus) (result *datav1alpha1.OperationStatus, err error) {
	result = opStatus.DeepCopy()
	object := handler.dataProcess

	releaseName := utils.GetDataProcessReleaseName(object.GetName())
	jobName := utils.GetDataProcessJobName(releaseName)

	ctx.Log.V(1).Info("DataProcess chart already existed, check its running status")
	job, err := kubeclient.GetJob(handler.Client, jobName, object.GetNamespace())
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.Info("Related job missing, will delete helm chart and retry", "namespace", object.GetNamespace(), "jobName", jobName)
			if err = helm.DeleteReleaseIfExists(releaseName, object.GetNamespace()); err != nil {
				ctx.Log.Error(err, "failed to delete dataprocess helm release", "namespace", object.GetNamespace(), "releaseName", releaseName)
				return
			}
			return
		}
		ctx.Log.Error(err, "can't get dataprocess job", "namespace", object.GetNamespace(), "jobName", jobName)
		return
	}

	finishedJobCondition := kubeclient.GetFinishedJobCondition(job)
	if finishedJobCondition == nil {
		ctx.Log.V(1).Info("DataProcess job still running", "namespace", object.GetNamespace(), "jobName", jobName)
		return
	}
	isJobSucceed := finishedJobCondition.Type == batchv1.JobComplete

	result.Conditions = []datav1alpha1.Condition{
		{
			Type:               common.ConditionType(finishedJobCondition.Type),
			Status:             finishedJobCondition.Status,
			Reason:             finishedJobCondition.Reason,
			Message:            finishedJobCondition.Message,
			LastProbeTime:      finishedJobCondition.LastProbeTime,
			LastTransitionTime: finishedJobCondition.LastTransitionTime,
		},
	}
	if isJobSucceed {
		result.Phase = common.PhaseComplete
	} else {
		result.Phase = common.PhaseFailed
	}
	result.Duration = utils.CalculateDuration(job.CreationTimestamp.Time, finishedJobCondition.LastTransitionTime.Time)
	return
}
