/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
		ready, err = e.Helper.CheckWorkersReady(runtimeToUpdate, runtimeToUpdate.Status, workers)
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
