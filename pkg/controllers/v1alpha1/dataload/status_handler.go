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

package dataload

import (
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
)

type OnceStatusHandler struct {
	client.Client
}

var _ dataoperation.StatusHandler = &OnceStatusHandler{}

type CronStatusHandler struct {
	client.Client
}

var _ dataoperation.StatusHandler = &CronStatusHandler{}

type OnEventStatusHandler struct {
	client.Client
}

func (r *OnceStatusHandler) GetOperationStatus(ctx cruntime.ReconcileRequestContext, object client.Object, opStatus *datav1alpha1.OperationStatus) (result *datav1alpha1.OperationStatus, err error) {
	result = opStatus.DeepCopy()
	// 2. Check running status of the DataLoad job
	releaseName := utils.GetDataLoadReleaseName(object.GetName())
	jobName := utils.GetDataLoadJobName(releaseName)

	ctx.Log.V(1).Info("DataLoad chart already existed, check its running status")
	job, err := utils.GetJob(r.Client, jobName, ctx.Namespace)
	if err != nil {
		// helm release found but job missing, delete the helm release and requeue
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.Info("Related Job missing, will delete helm chart and retry", "namespace", ctx.Namespace, "jobName", jobName)
			if err = helm.DeleteReleaseIfExists(releaseName, ctx.Namespace); err != nil {
				ctx.Log.Error(err, "can't delete dataload release", "namespace", ctx.Namespace, "releaseName", releaseName)
				return
			}
		}
		// other error
		ctx.Log.Error(err, "can't get dataload job", "namespace", ctx.Namespace, "jobName", jobName)
		return
	}

	if len(job.Status.Conditions) != 0 {
		if job.Status.Conditions[0].Type == batchv1.JobFailed ||
			job.Status.Conditions[0].Type == batchv1.JobComplete {
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
			if jobCondition.Type == batchv1.JobFailed {
				result.Phase = common.PhaseFailed
			} else {
				result.Phase = common.PhaseComplete
			}
			result.Duration = utils.CalculateDuration(object.GetCreationTimestamp().Time, jobCondition.LastTransitionTime.Time)

			return
		}
	}
	ctx.Log.V(1).Info("DataLoad job still running", "namespace", ctx.Namespace, "jobName", jobName)
	return
}

func (c *CronStatusHandler) GetOperationStatus(ctx cruntime.ReconcileRequestContext, object client.Object, opStatus *datav1alpha1.OperationStatus) (result *datav1alpha1.OperationStatus, err error) {
	result = opStatus.DeepCopy()

	// 1. Check running status of the DataLoad job
	releaseName := utils.GetDataLoadReleaseName(object.GetName())
	cronjobName := utils.GetDataLoadJobName(releaseName)

	cronjobStatus, err := kubeclient.GetCronJobStatus(c.Client, types.NamespacedName{Namespace: object.GetNamespace(), Name: cronjobName})

	if err != nil {
		// helm release found but cronjob missing, delete the helm release and requeue
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.Info("Related Cronjob missing, will delete helm chart and retry", "namespace", ctx.Namespace, "cronjobName", cronjobName)
			if err = helm.DeleteReleaseIfExists(releaseName, ctx.Namespace); err != nil {
				ctx.Log.Error(err, "can't delete DataLoad release", "namespace", ctx.Namespace, "releaseName", releaseName)
				return
			}
			return
		}
		// other error
		ctx.Log.Error(err, "can't get DataLoad cronjob", "namespace", ctx.Namespace, "cronjobName", cronjobName)
		return
	}

	jobs, err := utils.ListDataOperationJobByCronjob(c.Client, types.NamespacedName{Namespace: object.GetNamespace(), Name: cronjobName})
	if err != nil {
		ctx.Log.Error(err, "can't list DataLoad job by cronjob", "namespace", ctx.Namespace, "cronjobName", cronjobName)
		return
	}

	// get the newest job
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

	if len(currentJob.Status.Conditions) != 0 {
		if currentJob.Status.Conditions[0].Type == batchv1.JobFailed ||
			currentJob.Status.Conditions[0].Type == batchv1.JobComplete {
			jobCondition := currentJob.Status.Conditions[0]
			// job either failed or complete, update dataload's phase status
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
			if jobCondition.Type == batchv1.JobFailed {
				result.Phase = common.PhaseFailed
			} else {
				result.Phase = common.PhaseComplete
			}
			result.Duration = utils.CalculateDuration(object.GetCreationTimestamp().Time, jobCondition.LastTransitionTime.Time)
			return
		}
	}

	ctx.Log.V(1).Info("DataLoad job still running", "namespace", ctx.Namespace, "cronjobName", cronjobName)
	if opStatus.Phase == common.PhaseComplete || opStatus.Phase == common.PhaseFailed {
		// if dataload was complete or failed, but now job is running, set dataload pending first
		// dataset will be locked only when dataload pending
		result.Phase = common.PhasePending
		result.Duration = "-"
	}
	return
}

func (o *OnEventStatusHandler) GetOperationStatus(ctx cruntime.ReconcileRequestContext, object client.Object, opStatus *datav1alpha1.OperationStatus) (result *datav1alpha1.OperationStatus, err error) {
	//TODO implement me
	return nil, nil
}
