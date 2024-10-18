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

package alluxio

import (
	"context"
	"reflect"

	"github.com/fluid-cloudnative/fluid/pkg/types/cacheworkerset"
	openkruise "github.com/openkruise/kruise-api/apps/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

// CheckMasterReady checks if the master is ready
func (e *AlluxioEngine) CheckMasterReady() (ready bool, err error) {

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
	if runtime.Spec.ScaleConfig.WorkerType == "" || runtime.Spec.ScaleConfig.WorkerType == cacheworkerset.StatefulSetType {

	} else if runtime.Spec.ScaleConfig.WorkerType == cacheworkerset.AdvancedStatefulSetType {
		status := openkruise.StatefulSetStatus{
			Replicas:          master.Status.Replicas,
			ReadyReplicas:     master.Status.ReadyReplicas,
			UpdatedReplicas:   master.Status.UpdatedReplicas,
			CurrentReplicas:   master.Status.CurrentReplicas,
			CurrentRevision:   master.Status.CurrentRevision,
			UpdateRevision:    master.Status.UpdateRevision,
			CollisionCount:    master.Status.CollisionCount,
			Conditions:        master.Status.Conditions,
			AvailableReplicas: master.Status.AvailableReplicas,

			LabelSelector: master.Status.LabelSelector,
		}
		spec := openkruise.StatefulSetSpec{
			Replicas:             master.Spec.Replicas,
			PodManagementPolicy:  master.Spec.PodManagementPolicy,
			ServiceName:          master.Spec.ServiceName,
			Template:             master.Spec.Template,
			VolumeClaimTemplates: master.Spec.VolumeClaimTemplates,
			Selector:             master.Spec.Selector,
			RevisionHistoryLimit: master.Spec.RevisionHistoryLimit,
			UpdateStrategy:       master.Spec.UpdateStrategy,
		}
		set := openkruise.StatefulSet{
			ObjectMeta: master.ObjectMeta,
			Status:     status,
			Spec:       spec,
			TypeMeta:   master.TypeMeta,
		}
	} else if runtime.Spec.ScaleConfig.WorkerType == cacheworkerset.DaemonSetType {

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

			if runtime.Spec.APIGateway.Enabled {
				if runtimeToUpdate.Status.APIGatewayStatus == nil {
					runtimeToUpdate.Status.APIGatewayStatus, err = e.GetAPIGatewayStatus()
					if err != nil {
						return err
					}
				} else {
					e.Log.Info("No need to update APIGateway status")
				}
			} else {
				e.Log.Info("No need to update APIGateway status")
			}

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
	return
}

// ShouldSetupMaster checks if we need setup the master
func (e *AlluxioEngine) ShouldSetupMaster() (should bool, err error) {

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
// It returns any cache error encountered
func (e *AlluxioEngine) SetupMaster() (err error) {
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
