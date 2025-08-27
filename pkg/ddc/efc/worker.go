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

package efc

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
)

// getWorkerSelectors gets the selector of the worker
func (e *EFCEngine) getWorkerSelectors() string {
	labels := map[string]string{
		"release":          e.name,
		common.PodRoleType: workerPodRole,
		"app":              common.EFCRuntime,
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

// ShouldSetupWorkers checks if we need setup the workers
func (e *EFCEngine) ShouldSetupWorkers() (should bool, err error) {
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

// SetupWorkers checks the desired and current replicas of workers and makes an update
// over the status by setting phases and conditions. The function
// calls for a status update and finally returns error if anything unexpected happens.
func (e *EFCEngine) SetupWorkers() (err error) {
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		workers, err := ctrl.GetWorkersAsStatefulset(e.Client,
			types.NamespacedName{Namespace: e.namespace, Name: e.getWorkerName()})
		if err != nil {
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
		_ = utils.LoggingErrorExceptConflict(e.Log, err, "Failed to setup workers", types.NamespacedName{Namespace: e.namespace, Name: e.name})
		return err
	}
	return
}

// are the workers ready
func (e *EFCEngine) CheckWorkersReady() (ready bool, err error) {
	workers, err := ctrl.GetWorkersAsStatefulset(e.Client,
		types.NamespacedName{Namespace: e.namespace, Name: e.getWorkerName()})
	if err != nil {
		return ready, err
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}
		runtimeToUpdate := runtime.DeepCopy()
		ready, err = e.Helper.CheckAndUpdateWorkerStatus(runtimeToUpdate, workers)
		if err != nil {
			_ = utils.LoggingErrorExceptConflict(e.Log, err, "Failed to check worker ready", types.NamespacedName{Namespace: e.namespace, Name: e.name})
		}
		return err
	})

	return
}

func (e *EFCEngine) syncWorkersEndpoints() (count int, err error) {
	workerPods, err := e.getWorkerRunningPods()
	if err != nil {
		return 0, err
	}

	_, containerName := e.getWorkerPodInfo()
	workersEndpoints := WorkerEndPoints{}
	for _, pod := range workerPods {
		if !podutil.IsPodReady(&pod) {
			continue
		}
		for _, container := range pod.Spec.Containers {
			if container.Name == containerName {
				for _, port := range container.Ports {
					if port.Name == "rpc" {
						workersEndpoints.ContainerEndPoints = append(workersEndpoints.ContainerEndPoints, pod.Status.PodIP+":"+strconv.Itoa(int(port.ContainerPort)))
					}
				}
			}
		}
	}
	count = len(workersEndpoints.ContainerEndPoints)

	b, _ := json.Marshal(workersEndpoints)
	e.Log.Info("Sync worker endpoints", "worker-endpoints", string(b))

	configMap, err := kubeclient.GetConfigmapByName(e.Client, e.getWorkersEndpointsConfigmapName(), e.namespace)
	if err != nil {
		return count, err
	}
	if configMap == nil {
		return count, fmt.Errorf("fail to find ConfigMap name:%s, namespace:%s ", e.getWorkersEndpointsConfigmapName(), e.namespace)
	}
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		configMapToUpdate := configMap.DeepCopy()
		configMapToUpdate.Data[WorkerEndpointsDataName] = string(b)
		if !reflect.DeepEqual(configMapToUpdate, configMap) {
			err = e.Client.Update(context.TODO(), configMapToUpdate)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return count, err
	}

	return count, nil
}
