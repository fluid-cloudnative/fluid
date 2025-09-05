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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"k8s.io/client-go/util/retry"
)

// SetupWorkers checks the desired and current replicas of workers and makes an update
// over the status by setting phases and conditions. The function
// calls for a status update and finally returns error if anything unexpected happens.
func (e *JindoFSxEngine) SetupWorkers() (err error) {

	if e.runtime.Spec.Worker.Disabled {
		err = nil
		return
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		workers, err := ctrl.GetWorkersAsStatefulset(e.Client,
			types.NamespacedName{Namespace: e.namespace, Name: e.getWorkerName()})
		if err != nil {
			if fluiderrs.IsDeprecated(err) {
				e.Log.Info("Warning: Deprecated mode is not support, so skip handling", "details", err)
				return nil
			}
			return err
		}

		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}
		runtimeToUpdate := runtime.DeepCopy()
		return e.Helper.SetupWorkers(runtimeToUpdate, runtimeToUpdate.Status, workers)
	})

	if err != nil {
		_ = utils.LoggingErrorExceptConflict(e.Log,
			err,
			"Failed to setup worker",
			types.NamespacedName{
				Namespace: e.namespace,
				Name:      e.name,
			})
	}
	return
}

// ShouldSetupWorkers checks if we need setup the workers
func (e *JindoFSxEngine) ShouldSetupWorkers() (should bool, err error) {
	runtime, err := e.getRuntime()
	if err != nil {
		return
	}

	switch runtime.Status.WorkerPhase {
	case datav1alpha1.RuntimePhaseNone:
		should = true
	default:
		should = false
	}

	return
}

// CheckWorkersReady checks if the workers are ready
func (e *JindoFSxEngine) CheckWorkersReady() (ready bool, err error) {
	if e.runtime.Spec.Worker.Disabled {
		ready = true
		err = nil
		return
	}

	getRuntimeFn := func(client client.Client) (base.RuntimeInterface, error) {
		return utils.GetJindoRuntime(client, e.name, e.namespace)
	}

	ready, err = e.Helper.CheckAndSyncWorkerStatus(getRuntimeFn, types.NamespacedName{Namespace: e.namespace, Name: e.getWorkerName()})
	if err != nil {
		e.Log.Error(err, "fail to check and update worker status")
		return
	}

	if !ready {
		e.Log.Info("workers are not ready")
	}

	return
}

// getWorkerSelectors gets the selector of the worker
func (e *JindoFSxEngine) getWorkerSelectors() string {
	labels := map[string]string{
		"release":          e.name,
		common.PodRoleType: workerPodRole,
		"app":              common.JindoRuntime,
	}
	labelSelector := &metav1.LabelSelector{
		MatchLabels: labels,
	}

	selectorValue := ""
	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		e.Log.Error(err, "Failed to parse the labelSelector of the runtime", "labels", labels)
	} else {
		selectorValue = selector.String()
	}
	return selectorValue
}
