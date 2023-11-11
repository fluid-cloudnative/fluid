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
