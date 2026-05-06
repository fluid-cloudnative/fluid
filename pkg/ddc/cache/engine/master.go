/*
  Copyright 2026 The Fluid Authors.

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

package engine

import (
	"context"
	"errors"
	"reflect"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/cache/component"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
)

func (e *CacheEngine) SetupMasterComponent(masterValue *common.CacheRuntimeComponentValue) (bool, error) {
	shouldSetupMaster, err := e.shouldSetupMaster()
	if err != nil {
		return false, err
	}
	if shouldSetupMaster {
		if err = e.setupMasterInternal(masterValue); err != nil {
			e.Log.Error(err, "failed to setup master")
			return false, err
		}
	}

	return true, nil
}

func (e *CacheEngine) shouldSetupMaster() (bool, error) {
	runtime, err := e.getRuntime()
	if err != nil {
		return false, err
	}

	if runtime.Status.Master.Phase == datav1alpha1.RuntimePhaseNone {
		return true, nil
	}

	return false, nil
}

func (e *CacheEngine) setupMasterInternal(masterValue *common.CacheRuntimeComponentValue) error {
	manager := component.NewComponentHelper(masterValue.WorkloadType, e.Client)
	err := manager.Reconciler(context.TODO(), masterValue)
	if err != nil {
		return err
	}

	// update status of master
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}

		masterStatus, err := manager.ConstructComponentStatus(context.TODO(), masterValue)
		if err != nil {
			return err
		}
		// from RuntimePhaseNone to RuntimePhaseNotReady, not reconcile the master component the next time.
		masterStatus.Phase = datav1alpha1.RuntimePhaseNotReady

		runtimeToUpdate := runtime.DeepCopy()
		runtimeToUpdate.Status.Master = masterStatus

		// TODO(cache runtime): figure out how to use this selector
		// runtimeToUpdate.Status.Selector = e.getWorkerSelectors()

		if len(runtimeToUpdate.Status.Conditions) == 0 {
			runtimeToUpdate.Status.Conditions = []datav1alpha1.RuntimeCondition{}
		}
		cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeMasterInitialized, datav1alpha1.RuntimeMasterInitializedReason,
			"The master is initialized.", corev1.ConditionTrue)
		runtimeToUpdate.Status.Conditions =
			utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
				cond)

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			return e.Client.Status().Update(context.TODO(), runtimeToUpdate)
		}

		return nil
	})
	if err != nil {
		e.Log.Error(err, "failed to update runtime status")
		return err
	}

	return nil
}

func (e *CacheEngine) getMasterPodInfo(value *common.CacheRuntimeValue) (podName string, containerName string, err error) {
	// pod name is auto generated
	podName = GetComponentName(e.name, common.ComponentTypeMaster) + "-0"
	// container name, use the first container name
	if value.Master == nil || len(value.Master.PodTemplateSpec.Spec.Containers) == 0 {
		return "", "", errors.New("no container in master pod template")
	}
	containerName = value.Master.PodTemplateSpec.Spec.Containers[0].Name

	return
}
