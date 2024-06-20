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
	"fmt"
	"reflect"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/fluid-cloudnative/fluid/pkg/utils/patch"
)

// CheckMasterHealthy checks the sts healthy with role
func (e *Helper) CheckMasterHealthy(recorder record.EventRecorder, runtime base.RuntimeInterface,
	currentStatus datav1alpha1.RuntimeStatus,
	sts *appsv1.StatefulSet) (err error) {
	var (
		healthy             bool
		selector            labels.Selector
		unavailablePodNames []types.NamespacedName
	)
	if sts.Status.Replicas == sts.Status.ReadyReplicas {
		healthy = true
	}

	statusToUpdate := runtime.GetStatus()
	if len(statusToUpdate.Conditions) == 0 {
		statusToUpdate.Conditions = []datav1alpha1.RuntimeCondition{}
	}

	if healthy {
		realHealthy, readyTime, err := e.recheckMasterHealthyByEachContainerStartedTime(sts)
		if err != nil {
			e.log.Error(err, "failed to recheck master healthy by each container started time")
			return err
		}

		if !readyTime.IsZero() && sts.Annotations[common.AnnotationLatestMasterStartedTime] != readyTime.Format(time.RFC3339) {
			toPatch := patch.NewStrategicPatch().InsertAnnotation(common.AnnotationLatestMasterStartedTime, readyTime.Format(time.RFC3339))
			if err = e.client.Patch(context.Background(), sts, toPatch); err != nil {
				return err
			}
		}

		if !realHealthy {
			return fmt.Errorf("the master %s in %s is not ready. Master pod must have restarted since last ready time",
				sts.Name,
				sts.Namespace,
			)
		}

		cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeMasterReady, "The master is ready.",
			"The master is ready.", corev1.ConditionTrue)
		_, oldCond := utils.GetRuntimeCondition(statusToUpdate.Conditions, cond.Type)

		if oldCond == nil || oldCond.Type != cond.Type {
			statusToUpdate.Conditions =
				utils.UpdateRuntimeCondition(statusToUpdate.Conditions,
					cond)
		}
		statusToUpdate.MasterPhase = datav1alpha1.RuntimePhaseReady

	} else {
		// 1. Update the status
		cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeMasterReady, "The master is not ready.",
			fmt.Sprintf("The master %s in %s is not ready.", sts.Name, sts.Namespace), corev1.ConditionFalse)
		_, oldCond := utils.GetRuntimeCondition(statusToUpdate.Conditions, cond.Type)

		if oldCond == nil || oldCond.Type != cond.Type {
			statusToUpdate.Conditions =
				utils.UpdateRuntimeCondition(statusToUpdate.Conditions,
					cond)
		}
		statusToUpdate.MasterPhase = datav1alpha1.RuntimePhaseNotReady

		// 2. Record the event

		selector, err = metav1.LabelSelectorAsSelector(sts.Spec.Selector)
		if err != nil {
			return fmt.Errorf("error converting StatefulSet %s in namespace %s selector: %v", sts.Name, sts.Namespace, err)
		}

		unavailablePodNames, err = kubeclient.GetUnavailablePodNamesForStatefulSet(e.client, sts, selector)
		if err != nil {
			return err
		}

		// 3. Set event
		err = fmt.Errorf("the master %s in %s is not ready. The expected number is %d, the actual number is %d, the unhealthy pods are %v",
			sts.Name,
			sts.Namespace,
			sts.Status.Replicas,
			sts.Status.ReadyReplicas,
			unavailablePodNames)

		recorder.Eventf(runtime, corev1.EventTypeWarning, "MasterUnhealthy", err.Error())

	}

	status := *statusToUpdate
	if !reflect.DeepEqual(status, currentStatus) {
		updateErr := e.client.Status().Update(context.TODO(), runtime)
		if updateErr != nil {
			return updateErr
		}
	}

	if err != nil {
		return
	}

	return

}

func (e *Helper) StopBindDataSetInHealthCheck() bool {
	ds, err := utils.GetDataset(e.client, e.runtimeInfo.GetName(), e.runtimeInfo.GetNamespace())
	if err != nil {
		return false
	}
	// When current dataset is Failed, we expect to re-execute `Setup` in
	// runtime controller to recover the corrupted dataset, and runtime controller
	// will update dataset as Bound after successfully setup, that is, health check
	// does not do the bind operation to dataset.
	return ds.Status.Phase == datav1alpha1.FailedDatasetPhase
}

func (e *Helper) recheckMasterHealthyByEachContainerStartedTime(masterSts *appsv1.StatefulSet) (ready bool, readyTime metav1.Time, err error) {
	lastLatest := time.Time{}

	if len(masterSts.Annotations[common.AnnotationLatestMasterStartedTime]) > 0 {
		lastLatest, err = time.Parse(time.RFC3339, masterSts.Annotations[common.AnnotationLatestMasterStartedTime])
		if err != nil {
			e.log.Error(err, "failed to parse last latest master started time from annotation",
				"value", masterSts.Annotations[common.AnnotationLatestMasterStartedTime])
			return true, metav1.Time{}, err
		}
	}

	selector, err := metav1.LabelSelectorAsSelector(masterSts.Spec.Selector)
	if err != nil {
		return false, metav1.Time{}, fmt.Errorf("error converting StatefulSet %s in namespace %s selector: %v", masterSts.Name, masterSts.Namespace, err)
	}
	pods, err := kubeclient.GetPodsForStatefulSet(e.client, masterSts, selector)
	if err != nil {
		return false, metav1.Time{}, err
	}
	if len(pods) < int(*masterSts.Spec.Replicas) {
		return false, metav1.Time{}, nil
	}

	latest := metav1.Time{}
	for pi := range pods {
		curLatest := findLatestContainersStartedTime(&pods[pi])
		if curLatest.After(latest.Time) {
			latest = curLatest
		}
	}
	return !lastLatest.IsZero() && lastLatest.Equal(latest.Time), latest, nil
}

func findLatestContainersStartedTime(pod *corev1.Pod) metav1.Time {
	latest := metav1.Time{}
	for csi := range pod.Status.ContainerStatuses {
		cs := &pod.Status.ContainerStatuses[csi]
		if cs.State.Running != nil && cs.State.Running.StartedAt.After(latest.Time) {
			latest = cs.State.Running.StartedAt
		}
	}
	return latest
}
