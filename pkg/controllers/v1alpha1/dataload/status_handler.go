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
	dataLoad *datav1alpha1.DataLoad
}

var _ dataoperation.StatusHandler = &OnceStatusHandler{}

type CronStatusHandler struct {
	client.Client
	dataLoad *datav1alpha1.DataLoad
}

var _ dataoperation.StatusHandler = &CronStatusHandler{}

type OnEventStatusHandler struct {
	client.Client
	dataLoad *datav1alpha1.DataLoad
}

func (r *OnceStatusHandler) GetOperationStatus(ctx cruntime.ReconcileRequestContext, opStatus *datav1alpha1.OperationStatus) (result *datav1alpha1.OperationStatus, err error) {
	result = opStatus.DeepCopy()
	// 2. Check running status of the DataLoad job
	releaseName := utils.GetDataLoadReleaseName(r.dataLoad.GetName())
	jobName := utils.GetDataLoadJobName(releaseName)

	ctx.Log.V(1).Info("DataLoad chart already existed, check its running status")
	job, err := kubeclient.GetJob(r.Client, jobName, ctx.Namespace)
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
			result.Duration = utils.CalculateDuration(job.CreationTimestamp.Time, jobCondition.LastTransitionTime.Time)

			return
		}
	}
	ctx.Log.V(1).Info("DataLoad job still running", "namespace", ctx.Namespace, "jobName", jobName)
	return
}

func (c *CronStatusHandler) GetOperationStatus(ctx cruntime.ReconcileRequestContext, opStatus *datav1alpha1.OperationStatus) (result *datav1alpha1.OperationStatus, err error) {
	result = opStatus.DeepCopy()

	// 1. Check running status of the DataLoad job
	releaseName := utils.GetDataLoadReleaseName(c.dataLoad.GetName())
	cronjobName := utils.GetDataLoadJobName(releaseName)

	cronjobStatus, err := kubeclient.GetCronJobStatus(c.Client, types.NamespacedName{Namespace: c.dataLoad.GetNamespace(), Name: cronjobName})

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

	// update LastScheduleTime and LastSuccessfulTime
	result.LastScheduleTime = cronjobStatus.LastScheduleTime
	result.LastSuccessfulTime = cronjobStatus.LastSuccessfulTime

	jobs, err := utils.ListDataOperationJobByCronjob(c.Client, types.NamespacedName{Namespace: c.dataLoad.GetNamespace(), Name: cronjobName})
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
			result.Duration = utils.CalculateDuration(currentJob.CreationTimestamp.Time, jobCondition.LastTransitionTime.Time)
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

func (o *OnEventStatusHandler) GetOperationStatus(ctx cruntime.ReconcileRequestContext, opStatus *datav1alpha1.OperationStatus) (result *datav1alpha1.OperationStatus, err error) {
	//TODO implement me
	return nil, nil
}
