/*
Copyright 2023 The Fluid Author.

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

package jindofsx

import (
	"context"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

func (e *JindoFSxEngine) CheckMasterReady() (ready bool, err error) {
	if e.runtime.Spec.Master.Disabled {
		ready = true
		err = nil
		return
	}
	masterName := e.getMasterName()
	// 1. Check the status
	runtime, err := e.getRuntime()
	if err != nil {
		return
	}

	master, err := kubeclient.GetStatefulSet(e.Client, masterName, e.namespace)
	if err != nil {
		return
	}

	masterReplicas := runtime.Spec.Master.Replicas
	if masterReplicas == 0 {
		masterReplicas = 1
	}
	if masterReplicas == master.Status.ReadyReplicas {
		ready = true
	} else {
		e.Log.Info("The master is not ready.", "replicas", masterReplicas,
			"readyReplicas", master.Status.ReadyReplicas)
	}

	// 2. Update the phase
	if ready {
		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			runtime, err := e.getRuntime()
			if err != nil {
				return err
			}
			runtimeToUpdate := runtime.DeepCopy()

			runtimeToUpdate.Status.CurrentMasterNumberScheduled = int32(master.Status.ReadyReplicas)

			runtimeToUpdate.Status.MasterPhase = datav1alpha1.RuntimePhaseReady

			if len(runtimeToUpdate.Status.Conditions) == 0 {
				runtimeToUpdate.Status.Conditions = []datav1alpha1.RuntimeCondition{}
			}
			cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeMasterReady, datav1alpha1.RuntimeMasterReadyReason,
				"The master is ready.", corev1.ConditionTrue)
			runtimeToUpdate.Status.Conditions =
				utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
					cond)

			if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
				return e.Client.Status().Update(context.TODO(), runtimeToUpdate)
			}

			return nil
		})

		if err != nil {
			e.Log.Error(err, "Update runtime status")
			return
		}
	}

	if err != nil {
		return
	}

	return
}

// ShouldSetupMaster checks if we need setup the master
func (e *JindoFSxEngine) ShouldSetupMaster() (should bool, err error) {
	runtime, err := e.getRuntime()
	if err != nil {
		return
	}

	switch runtime.Status.MasterPhase {
	case datav1alpha1.RuntimePhaseNone:
		should = true
	default:
		should = false
	}

	return
}

// SetupMaster setups the master and updates the status
// It will print the information in the Debug window according to the Master status
// It return any cache error encountered
func (e *JindoFSxEngine) SetupMaster() (err error) {

	// Setup the Jindo cluster
	masterName := e.getMasterName()
	master, err := kubeclient.GetStatefulSet(e.Client, masterName, e.namespace)
	if err != nil && apierrs.IsNotFound(err) {
		//1. Is not found error
		e.Log.Info("SetupMaster", "master", e.name+"-master")
		return e.setupMasterInernal()
	} else if err != nil {
		//2. Other errors
		return
	} else {
		//3.The master has been set up
		e.Log.Info("The master has been set.", "replicas", master.Status.ReadyReplicas)
	}

	if e.runtime.Spec.Master.Disabled {
		return nil
	}

	// 2. Update the status of the runtime
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}
		runtimeToUpdate := runtime.DeepCopy()
		runtimeToUpdate.Status.MasterPhase = datav1alpha1.RuntimePhaseNotReady
		replicas := runtimeToUpdate.Spec.Master.Replicas
		if replicas == 0 {
			replicas = 1
		}

		// Init selector for worker
		runtimeToUpdate.Status.Selector = e.getWorkerSelectors()
		runtimeToUpdate.Status.DesiredMasterNumberScheduled = replicas
		runtimeToUpdate.Status.ValueFileConfigmap = e.getConfigmapName()

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
		e.Log.Error(err, "Update runtime status")
		return err
	}

	return
}
