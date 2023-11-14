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

package thin

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
)

func (t ThinEngine) CheckWorkersReady() (ready bool, err error) {
	var (
		workerName string = t.getWorkerName()
		namespace  string = t.namespace
	)

	workers, err := kubeclient.GetStatefulSet(t.Client, workerName, namespace)
	if err != nil {
		return ready, err
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := t.getRuntime()
		if err != nil {
			return err
		}
		runtimeToUpdate := runtime.DeepCopy()
		ready, err = t.Helper.CheckWorkersReady(runtimeToUpdate, runtimeToUpdate.Status, workers)
		if err != nil {
			_ = utils.LoggingErrorExceptConflict(t.Log, err, "Failed to setup worker",
				types.NamespacedName{Namespace: t.namespace, Name: t.name})
		}
		return err
	})

	return
}

func (t ThinEngine) ShouldSetupWorkers() (should bool, err error) {
	runtime, err := t.getRuntime()
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

func (t ThinEngine) SetupWorkers() (err error) {
	var (
		workerName string = t.getWorkerName()
		namespace  string = t.namespace
	)

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		workers, err := kubeclient.GetStatefulSet(t.Client, workerName, namespace)
		if err != nil {
			return err
		}
		runtime, err := t.getRuntime()
		if err != nil {
			return err
		}
		runtimeToUpdate := runtime.DeepCopy()
		err = t.Helper.SetupWorkers(runtimeToUpdate, runtimeToUpdate.Status, workers)
		return err
	})
	if err != nil {
		return utils.LoggingErrorExceptConflict(t.Log, err, "Failed to setup worker",
			types.NamespacedName{Namespace: t.namespace, Name: t.name})
	}
	return
}

// getWorkerSelectors gets the selector of the worker
func (t *ThinEngine) getWorkerSelectors() string {
	labels := map[string]string{
		"release":          t.name,
		common.PodRoleType: workerPodRole,
		"app":              common.ThinRuntime,
	}
	labelSelector := &metav1.LabelSelector{
		MatchLabels: labels,
	}

	selectorValue := ""
	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		t.Log.Error(err, "Failed to parse the labelSelector of the runtime", "labels", labels)
	} else {
		selectorValue = selector.String()
	}
	return selectorValue
}
