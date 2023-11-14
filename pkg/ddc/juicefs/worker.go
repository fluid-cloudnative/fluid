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
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

func (j JuiceFSEngine) CheckWorkersReady() (ready bool, err error) {
	var (
		workerName string = j.getWorkerName()
		namespace  string = j.namespace
	)

	workers, err := kubeclient.GetStatefulSet(j.Client, workerName, namespace)
	if err != nil {
		return ready, err
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := j.getRuntime()
		if err != nil {
			return err
		}
		runtimeToUpdate := runtime.DeepCopy()
		ready, err = j.Helper.CheckWorkersReady(runtimeToUpdate, runtimeToUpdate.Status, workers)
		if err != nil {
			_ = utils.LoggingErrorExceptConflict(j.Log, err, "Failed to setup worker",
				types.NamespacedName{Namespace: j.namespace, Name: j.name})
		}
		return err
	})

	return
}

func (j JuiceFSEngine) ShouldSetupWorkers() (should bool, err error) {
	runtime, err := j.getRuntime()
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

func (j JuiceFSEngine) SetupWorkers() (err error) {
	var (
		workerName string = j.getWorkerName()
		namespace  string = j.namespace
	)

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		workers, err := kubeclient.GetStatefulSet(j.Client, workerName, namespace)
		if err != nil {
			return err
		}
		runtime, err := j.getRuntime()
		if err != nil {
			return err
		}
		runtimeToUpdate := runtime.DeepCopy()
		err = j.Helper.SetupWorkers(runtimeToUpdate, runtimeToUpdate.Status, workers)
		return err
	})
	if err != nil {
		return utils.LoggingErrorExceptConflict(j.Log, err, "Failed to setup worker",
			types.NamespacedName{Namespace: j.namespace, Name: j.name})
	}
	return
}

// getWorkerSelectors gets the selector of the worker
func (j *JuiceFSEngine) getWorkerSelectors() string {
	labels := map[string]string{
		"release":          j.name,
		common.PodRoleType: workerPodRole,
		"app":              common.JuiceFSRuntime,
	}
	labelSelector := &metav1.LabelSelector{
		MatchLabels: labels,
	}

	selectorValue := ""
	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		j.Log.Error(err, "Failed to parse the labelSelector of the runtime", "labels", labels)
	} else {
		selectorValue = selector.String()
	}
	return selectorValue
}
