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

package ctrl

import (
	"context"
	"reflect"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (e *Helper) CheckAndSyncMasterStatus(getRuntimeFn func(client.Client) (base.RuntimeInterface, error), masterStsNamespacedName types.NamespacedName) (ready bool, err error) {
	masterSts, err := kubeclient.GetStatefulSet(e.client, masterStsNamespacedName.Name, masterStsNamespacedName.Namespace)
	if err != nil {
		return
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := getRuntimeFn(e.client)
		if err != nil {
			return err
		}

		oldStatus := runtime.GetStatus().DeepCopy()
		statusToUpdate := runtime.GetStatus()
		var expectReplicas int32
		if masterSts.Spec.Replicas != nil {
			expectReplicas = *masterSts.Spec.Replicas
		} else {
			expectReplicas = 0
		}

		statusToUpdate.DesiredMasterNumberScheduled = expectReplicas
		statusToUpdate.CurrentMasterNumberScheduled = masterSts.Status.Replicas
		statusToUpdate.MasterNumberReady = masterSts.Status.ReadyReplicas

		phase := kubeclient.GetPhaseFromStatefulset(expectReplicas, *masterSts)
		statusToUpdate.MasterPhase = phase

		var cond datav1alpha1.RuntimeCondition
		if len(statusToUpdate.Conditions) == 0 {
			statusToUpdate.Conditions = []datav1alpha1.RuntimeCondition{}
		}

		// add masterInitCond to indicate master is initialized, this is idempotent because RuntimeMasterInitialized is a OnceActionType condition
		masterInitCond := utils.NewRuntimeCondition(datav1alpha1.RuntimeMasterInitialized, datav1alpha1.RuntimeMasterInitializedReason,
			"The master is initialized.", corev1.ConditionTrue)
		statusToUpdate.Conditions =
			utils.UpdateRuntimeCondition(statusToUpdate.Conditions,
				masterInitCond)

		switch phase {
		case datav1alpha1.RuntimePhaseReady:
			ready = true
			cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeMasterReady, datav1alpha1.RuntimeMasterReadyReason,
				"The master is ready.", corev1.ConditionTrue)
		case datav1alpha1.RuntimePhasePartialReady:
			ready = true
			cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeMasterReady, datav1alpha1.RuntimeMasterReadyReason,
				"The master is partially ready.", corev1.ConditionTrue)
		case datav1alpha1.RuntimePhaseNotReady:
			ready = false
			cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeMasterReady, datav1alpha1.RuntimeMasterReadyReason,
				"The master is not ready.", corev1.ConditionFalse)
		}

		if len(cond.Type) != 0 {
			statusToUpdate.Conditions = utils.UpdateRuntimeCondition(statusToUpdate.Conditions, cond)
		}

		if !reflect.DeepEqual(oldStatus, statusToUpdate) {
			return e.client.Status().Update(context.TODO(), runtime)
		}

		return nil
	})

	if err != nil {
		return false, errors.Wrap(err, "failed to update master ready status in runtime status")
	}

	return ready, nil
}
