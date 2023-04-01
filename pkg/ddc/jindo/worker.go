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

package jindo

import (
	"context"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
)

// SetupWorkers checks the desired and current replicas of workers and makes an update
// over the status by setting phases and conditions. The function
// calls for a status update and finally returns error if anything unexpected happens.
func (e *JindoEngine) SetupWorkers() (err error) {

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
func (e *JindoEngine) ShouldSetupWorkers() (should bool, err error) {
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
func (e *JindoEngine) CheckWorkersReady() (ready bool, err error) {

	workers, err := ctrl.GetWorkersAsStatefulset(e.Client,
		types.NamespacedName{Namespace: e.namespace, Name: e.getWorkerName()})
	if err != nil {
		if fluiderrs.IsDeprecated(err) {
			e.Log.Info("Warning: Deprecated mode is not support, so skip handling", "details", err)
			ready = true
			return ready, nil
		}
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
			_ = utils.LoggingErrorExceptConflict(e.Log,
				err,
				"Failed to setup worker",
				types.NamespacedName{
					Namespace: e.namespace,
					Name:      e.name,
				})
		}
		return err
	})

	return
}

// getWorkerSelectors gets the selector of the worker
func (e *JindoEngine) getWorkerSelectors() string {
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

// buildWorkersAffinity builds workers affinity if it doesn't have
func (e *JindoEngine) buildWorkersAffinity(workers *v1.StatefulSet) (workersToUpdate *v1.StatefulSet, err error) {
	// TODO: for now, runtime affinity can't be set by user, so we can assume the affinity is nil in the first time.
	// We need to enhance it in future
	workersToUpdate = workers.DeepCopy()

	if workersToUpdate.Spec.Template.Spec.Affinity == nil {
		workersToUpdate.Spec.Template.Spec.Affinity = &corev1.Affinity{}
		dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
		if err != nil {
			return workersToUpdate, err
		}
		// 1. Set pod anti affinity(required) for same dataset (Current using port conflict for scheduling, no need to do)

		// 2. Set pod anti affinity for the different dataset
		if dataset.IsExclusiveMode() {
			workersToUpdate.Spec.Template.Spec.Affinity.PodAntiAffinity = &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "fluid.io/dataset",
									Operator: metav1.LabelSelectorOpExists,
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			}
		} else {
			workersToUpdate.Spec.Template.Spec.Affinity.PodAntiAffinity = &corev1.PodAntiAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
					{
						// The default weight is 50
						Weight: 50,
						PodAffinityTerm: corev1.PodAffinityTerm{
							LabelSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      "fluid.io/dataset",
										Operator: metav1.LabelSelectorOpExists,
									},
								},
							},
							TopologyKey: "kubernetes.io/hostname",
						},
					},
				},
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "fluid.io/dataset-placement",
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{string(datav1alpha1.ExclusiveMode)},
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			}
		}

		// 3. Prefer to locate on the node which already has fuse on it
		if workersToUpdate.Spec.Template.Spec.Affinity.NodeAffinity == nil {
			workersToUpdate.Spec.Template.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{}
		}

		if len(workersToUpdate.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) == 0 {
			workersToUpdate.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = []corev1.PreferredSchedulingTerm{}
		}

		workersToUpdate.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution =
			append(workersToUpdate.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
				corev1.PreferredSchedulingTerm{
					Weight: 100,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      e.getFuseLabelname(),
								Operator: corev1.NodeSelectorOpIn,
								Values:   []string{"true"},
							},
						},
					},
				})

		// 3. set node affinity if possible
		if dataset.Spec.NodeAffinity != nil {
			if dataset.Spec.NodeAffinity.Required != nil {
				workersToUpdate.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution =
					dataset.Spec.NodeAffinity.Required
			}
		}

		err = e.Client.Update(context.TODO(), workersToUpdate)
		if err != nil {
			return workersToUpdate, err
		}

	}

	return
}
