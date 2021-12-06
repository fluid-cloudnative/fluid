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

	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

// GetWorkersAsStatefulset gets workers as statefulset object. if it returns deprecated errors, it indicates that
// not support anymore.
func GetWorkersAsStatefulset(client client.Client, key types.NamespacedName) (workers *appsv1.StatefulSet, err error) {
	workers, err = kubeclient.GetStatefulSet(client, key.Name, key.Namespace)
	if err != nil {
		if apierrs.IsNotFound(err) {
			_, dsErr := kubeclient.GetDaemonset(client, key.Name, key.Namespace)
			// return workers, fluiderr.NewDeprecated()
			// find the daemonset successfully
			if dsErr == nil {
				return workers, fluiderrs.NewDeprecated(schema.GroupResource{
					Group:    appsv1.SchemeGroupVersion.Group,
					Resource: "daemonsets",
				}, key)
			}
		}
	}

	return
}

// CheckworkersHealthy checks the sts healthy with role
func (e *Helper) CheckWorkersHealthy(ctx cruntime.ReconcileRequestContext, runtime base.RuntimeInterface,
	currentStatus datav1alpha1.RuntimeStatus,
	sts *appsv1.StatefulSet) (err error) {
	var healthy bool
	if sts.Status.Replicas == sts.Status.ReadyReplicas {
		healthy = true
	}

	statusToUpdate := runtime.GetStatus()
	if len(statusToUpdate.Conditions) == 0 {
		statusToUpdate.Conditions = []datav1alpha1.RuntimeCondition{}
	}

	if healthy {
		cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersReady, "The worker are ready.",
			"The worker are ready.", corev1.ConditionTrue)
		_, oldCond := utils.GetRuntimeCondition(statusToUpdate.Conditions, cond.Type)

		if oldCond == nil || oldCond.Type != cond.Type {
			statusToUpdate.Conditions =
				utils.UpdateRuntimeCondition(statusToUpdate.Conditions,
					cond)
		}
		statusToUpdate.WorkerPhase = datav1alpha1.RuntimePhaseReady

	} else {
		// 1. Update the status
		cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersReady, "The workers are not ready.",
			fmt.Sprintf("The workers %s in %s are not ready.", sts.Name, sts.Namespace), corev1.ConditionFalse)
		_, oldCond := utils.GetRuntimeCondition(statusToUpdate.Conditions, cond.Type)

		if oldCond == nil || oldCond.Type != cond.Type {
			statusToUpdate.Conditions =
				utils.UpdateRuntimeCondition(statusToUpdate.Conditions,
					cond)
		}
		statusToUpdate.WorkerPhase = datav1alpha1.RuntimePhaseNotReady

		// 2. Record the event
		selector, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
		if err != nil {
			return fmt.Errorf("error converting StatefulSet %s in namespace %s selector: %v", sts.Name, sts.Namespace, err)
		}

		unavailablePodNames, err := kubeclient.GetunavailablePodNamesForStatefulSet(e.client, sts, selector)
		if err != nil {
			return err
		}

		// 3. Set error
		msg := fmt.Sprintf("the workers %s in %s are not ready. The expected number is %d, the actual number is %d, the unhealthy pods are %v",
			sts.Name,
			sts.Namespace,
			sts.Status.Replicas,
			sts.Status.ReadyReplicas,
			unavailablePodNames)

		ctx.Recorder.Eventf(runtime, corev1.EventTypeWarning, "workersUnhealthy", msg)
	}

	if err != nil {
		return
	}

	status := *statusToUpdate
	if !reflect.DeepEqual(status, currentStatus) {
		return e.client.Status().Update(context.TODO(), runtime)
	}

	return

}
