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
	ctx.Log.V(1).Info("DataProcess job still running", "namespace", ctx.Namespace, "jobName", jobName)
	return
}
