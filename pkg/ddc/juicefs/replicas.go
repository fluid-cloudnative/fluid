/*
Copyright 2021 The Fluid Authors.

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

package juicefs

import (
	"context"
	"reflect"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
)

// SyncReplicas syncs the replicas
func (j *JuiceFSEngine) SyncReplicas(ctx cruntime.ReconcileRequestContext) error {
	runtime, err := j.getRuntime()
	if err != nil {
		return err
	}

	desireReplicas := runtime.Replicas()
	if desireReplicas > runtime.Status.CurrentWorkerNumberScheduled {

		err = j.SetupWorkers()
		if err != nil {
			return err
		}
		_, err = j.CheckWorkersReady()
		if err != nil {
			j.Log.Error(err, "Check if the workers are ready")
			return err
		}

		// update conditions
		err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			runtime, err := j.getRuntime()
			if err != nil {
				return err
			}

			runtimeToUpdate := runtime.DeepCopy()
			cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersScaledOutReason, datav1alpha1.RuntimeWorkersScaledOutReason,
				"The workers are scale out.", corev1.ConditionTrue)
			runtimeToUpdate.Status.Conditions =
				utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
					cond)

			if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
				return j.Client.Status().Update(context.TODO(), runtimeToUpdate)
			}

			return nil
		})

		if err != nil {
			return err
		}

		// add the event
		runtime, err := j.getRuntime()
		if err != nil {
			return err
		}
		currentReplicas := runtime.Status.CurrentWorkerNumberScheduled
		desireReplicas := runtime.Status.DesiredWorkerNumberScheduled
		ctx.Recorder.Eventf(runtime, corev1.EventTypeNormal, common.Succeed, "JuiceFS runtime scaled out. current replicas: %d, desired replicas: %d.", currentReplicas, desireReplicas)

	} else if desireReplicas < runtime.Status.CurrentWorkerNumberScheduled {
		replicas := runtime.Replicas()
		j.Log.Info("Scaling in JuiceFS workers", "expectedReplicas", replicas)
		curReplicas, err := j.destroyWorkers(replicas)
		if err != nil {
			return err
		}

		if curReplicas > replicas {
			ctx.Recorder.Eventf(runtime, corev1.EventTypeWarning, common.RuntimeScaleInFailed,
				"JuiceFS workers are being used by some pods, can't scale in (expected replicas: %v, current replicas: %v)",
				replicas, curReplicas)
		} else {
			ctx.Recorder.Eventf(runtime, corev1.EventTypeNormal, common.Succeed, "JuiceFS runtime scaled in. current replicas: %d, desired replicas: %d.", curReplicas, desireReplicas)
		}

		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			runtime, err := j.getRuntime()
			if err != nil {
				j.Log.Error(err, "scale in when sync replicas")
				return err
			}

			runtimeToUpdate := runtime.DeepCopy()

			if len(runtimeToUpdate.Status.Conditions) == 0 {
				runtimeToUpdate.Status.Conditions = []datav1alpha1.RuntimeCondition{}
			}

			runtimeToUpdate.Status.DesiredWorkerNumberScheduled = replicas
			runtimeToUpdate.Status.CurrentWorkerNumberScheduled = curReplicas
			cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkerScaledIn, datav1alpha1.RuntimeWorkersScaledInReason,
				"The workers scaled in.", corev1.ConditionTrue)
			runtimeToUpdate.Status.Conditions =
				utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions, cond)

			if !runtimeToUpdate.Spec.Fuse.Global {
				runtimeToUpdate.Status.DesiredFuseNumberScheduled = replicas
				runtimeToUpdate.Status.CurrentWorkerNumberScheduled = curReplicas
				fuseCond := utils.NewRuntimeCondition(datav1alpha1.RuntimeFusesScaledIn, datav1alpha1.RuntimeFusesScaledInReason,
					"The fuses scaled in.", corev1.ConditionTrue)
				runtimeToUpdate.Status.Conditions =
					utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions, fuseCond)
			}

			if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
				return j.Client.Status().Update(context.TODO(), runtimeToUpdate)
			}

			return nil
		})

		if err != nil {
			return err
		}

	} else {
		j.Log.V(1).Info("Nothing to do")
	}

	return nil
}
