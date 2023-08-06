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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OnceStatusHandler struct {
	client.Client
}

var _ dataoperation.StatusHandler = &OnceStatusHandler{}

// GetOperationStatus get operation status according to helm chart status
func (handler *OnceStatusHandler) GetOperationStatus(ctx runtime.ReconcileRequestContext, object client.Object, opStatus *datav1alpha1.OperationStatus) (result *datav1alpha1.OperationStatus, err error) {
	result = opStatus.DeepCopy()

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
