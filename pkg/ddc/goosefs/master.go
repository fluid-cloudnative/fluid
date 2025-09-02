/*
Copyright 2022 The Fluid Authors.

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

package goosefs

import (
	"context"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

// CheckMasterReady checks if the master is ready
func (e *GooseFSEngine) CheckMasterReady() (ready bool, err error) {
	getRuntimeFn := func(client client.Client) (base.RuntimeInterface, error) {
		return utils.GetGooseFSRuntime(client, e.name, e.namespace)
	}

	ready, err = e.Helper.CheckAndUpdateMasterStatus(getRuntimeFn, types.NamespacedName{Namespace: e.namespace, Name: e.getMasterName()})
	if err != nil {
		e.Log.Error(err, "fail to check and update master status")
		return
	}

	if !ready {
		e.Log.Info("master is not ready")
	}

	return
}

// ShouldSetupMaster checks if we need setup the master
func (e *GooseFSEngine) ShouldSetupMaster() (should bool, err error) {

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

// SetupMaster setups the master and updates the status of the runtime.
//
// This function performs the following steps:
// 1. Checks if the master StatefulSet exists:
//   - If not found, initializes the master via `setupMasterInternal()`.
//   - If found, logs the current ready replicas.
//
// 2. Updates the runtime status:
//   - Sets the master phase to `RuntimePhaseNotReady`.
//   - Records desired master replicas (defaulting to 1 if unspecified).
//   - Initializes worker selectors and sets the value file configmap.
//   - Adds a condition indicating the master is initialized.
//
// 3. Uses retry logic to handle concurrent updates to the runtime status.
//
//	Parameters:
//	- e: *GooseFSEngine
//	   The engine instance containing client, logger, namespace, and configuration for the GooseFS runtime.
//
//	Returns:
//	- error
//	   Returns an error if the master setup fails or the runtime status update encounters an issue.
func (e *GooseFSEngine) SetupMaster() (err error) {
	masterName := e.getMasterName()

	// 1. Setup the master
	master, err := kubeclient.GetStatefulSet(e.Client, masterName, e.namespace)
	if err != nil && apierrs.IsNotFound(err) {
		//1. Is not found error
		e.Log.V(1).Info("SetupMaster", "master", masterName)
		return e.setupMasterInternal()
	} else if err != nil {
		//2. Other errors
		return
	} else {
		//3.The master has been set up
		e.Log.V(1).Info("The master has been set.", "replicas", master.Status.ReadyReplicas)
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
		runtimeToUpdate.Status.ValueFileConfigmap = e.getHelmValuesConfigMapName()

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
