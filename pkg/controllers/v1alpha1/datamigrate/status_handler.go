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

package datamigrate

import (
	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

type OnceStatusHandler struct {
	client.Client
	Log         logr.Logger
	dataMigrate *datav1alpha1.DataMigrate
}

var _ dataoperation.StatusHandler = &OnceStatusHandler{}

type CronStatusHandler struct {
	client.Client
	Log         logr.Logger
	dataMigrate *datav1alpha1.DataMigrate
}

var _ dataoperation.StatusHandler = &CronStatusHandler{}

type OnEventStatusHandler struct {
	client.Client
	Log         logr.Logger
	dataMigrate *datav1alpha1.DataMigrate
}

var _ dataoperation.StatusHandler = &OnEventStatusHandler{}

func (m *OnceStatusHandler) GetOperationStatus(ctx cruntime.ReconcileRequestContext, opStatus *datav1alpha1.OperationStatus) (result *datav1alpha1.OperationStatus, err error) {
	result = opStatus.DeepCopy()
	object := m.dataMigrate
	// 1. Check running status of the DataMigrate job
	releaseName := utils.GetDataMigrateReleaseName(object.GetName())
	jobName := utils.GetDataMigrateJobName(releaseName)
	job, err := kubeclient.GetJob(m.Client, jobName, object.GetNamespace())

	if err != nil {
		// helm release found but job missing, delete the helm release and requeue
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.Info("Related Job missing, will delete helm chart and retry", "namespace", ctx.Namespace, "jobName", jobName)
			if err = helm.DeleteReleaseIfExists(releaseName, ctx.Namespace); err != nil {
				m.Log.Error(err, "can't delete DataMigrate release", "namespace", ctx.Namespace, "releaseName", releaseName)
				return
			}
			return
		}
		// other error
		ctx.Log.Error(err, "can't get DataMigrate job", "namespace", ctx.Namespace, "jobName", jobName)
		return
	}

	if len(job.Status.Conditions) != 0 {
		if job.Status.Conditions[0].Type == batchv1.JobFailed ||
			job.Status.Conditions[0].Type == batchv1.JobComplete {
			jobCondition := job.Status.Conditions[0]
			// job either failed or complete, update DataMigrate's phase status
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
			result.Duration = utils.CalculateDuration(job.CreationTimestamp.Time, jobCondition.LastTransitionTime.Time)
			return
		}
	}

	ctx.Log.V(1).Info("DataMigrate job still running", "namespace", ctx.Namespace, "jobName", jobName)
	return
}

func (c *CronStatusHandler) GetOperationStatus(ctx cruntime.ReconcileRequestContext, opStatus *datav1alpha1.OperationStatus) (result *datav1alpha1.OperationStatus, err error) {
	result = opStatus.DeepCopy()
	object := c.dataMigrate
	// 1. Check running status of the DataMigrate job
	releaseName := utils.GetDataMigrateReleaseName(object.GetName())
	cronjobName := utils.GetDataMigrateJobName(releaseName)
	cronjobStatus, err := kubeclient.GetCronJobStatus(c.Client, types.NamespacedName{Namespace: object.GetNamespace(), Name: cronjobName})

	if err != nil {
		// helm release found but cronjob missing, delete the helm release and requeue
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.Info("Related Cronjob missing, will delete helm chart and retry", "namespace", ctx.Namespace, "cronjobName", cronjobName)
			if err = helm.DeleteReleaseIfExists(releaseName, ctx.Namespace); err != nil {
				c.Log.Error(err, "can't delete DataMigrate release", "namespace", ctx.Namespace, "releaseName", releaseName)
				return
			}
			return
		}
		// other error
		ctx.Log.Error(err, "can't get DataMigrate cronjob", "namespace", ctx.Namespace, "cronjobName", cronjobName)
		return
	}

	// update LastScheduleTime and LastSuccessfulTime
	result.LastScheduleTime = cronjobStatus.LastScheduleTime
	result.LastSuccessfulTime = cronjobStatus.LastSuccessfulTime

	jobs, err := utils.ListDataOperationJobByCronjob(c.Client, types.NamespacedName{Namespace: object.GetNamespace(), Name: cronjobName})
	if err != nil {
		ctx.Log.Error(err, "can't list DataMigrate job by cronjob", "namespace", ctx.Namespace, "cronjobName", cronjobName)
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

	// only parallel tasks will set the field Suspend to true, no need to check the parallel task number > 1.
	if currentJob.Spec.Suspend != nil && *currentJob.Spec.Suspend {
		// scale the stateful set, the job acts as a worker.
		err = kubeclient.ScaleStatefulSet(c.Client, utils.GetParallelOperationWorkersName(releaseName), c.dataMigrate.Namespace, c.dataMigrate.Spec.Parallelism-1)
		if err != nil {
			return
		}
		ctx.Log.Info("scale sts success", "name", utils.GetParallelOperationWorkersName(releaseName))
		// set the job suspend field false
		jobToUpdate := currentJob.DeepCopy()
		flag := false
		jobToUpdate.Spec.Suspend = &flag
		err = kubeclient.UpdateJob(c.Client, jobToUpdate)
		// next loop
		ctx.Log.Info("update job", "name", jobToUpdate.GetName(), "error", err)
		if err != nil {
			return
		}
	}

	if len(currentJob.Status.Conditions) != 0 {
		// find the job final status condition. if job is resumed, the first condition type is 'Suspended'
		var jobCondition *batchv1.JobCondition
		for _, condition := range currentJob.Status.Conditions {
			// job is finished.
			if condition.Type == batchv1.JobFailed || condition.Type == batchv1.JobComplete {
				jobCondition = &condition
				break
			}
		}

		if jobCondition != nil {
			// job either failed or complete, update DataMigrate's phase status
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
			result.Duration = utils.CalculateDuration(currentJob.CreationTimestamp.Time, jobCondition.LastTransitionTime.Time)
			// the return statement makes the below code executed in the reconcileCompleted/reconcileFailed.
			// the status for cron data migrate is the correct status Complete/Failed not Pending/Executing before next job is not started.
			return
		}
	}

	ctx.Log.V(1).Info("DataMigrate job still running", "namespace", ctx.Namespace, "cronjobName", cronjobName)
	if opStatus.Phase == common.PhaseComplete || opStatus.Phase == common.PhaseFailed {
		// if datamigrate was complete or failed, but now job is running, set datamigrate pending first
		// dataset will be locked only when datamigrate pending
		result.Phase = common.PhasePending
		result.Duration = "-"
	}
	return
}

func (o *OnEventStatusHandler) GetOperationStatus(ctx cruntime.ReconcileRequestContext, opStatus *datav1alpha1.OperationStatus) (result *datav1alpha1.OperationStatus, err error) {
	//TODO implement me
	return nil, nil
}
