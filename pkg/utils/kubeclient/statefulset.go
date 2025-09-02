/*
Copyright 2023 The Fluid Authors.

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

package kubeclient

import (
	"context"
	"k8s.io/client-go/util/retry"
	"reflect"
	"regexp"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

// ScaleStatefulSet scale the statefulset replicas
func ScaleStatefulSet(client client.Client, name string, namespace string, replicas int32) error {
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		workers, err := GetStatefulSet(client, name, namespace)
		if err != nil {
			return err
		}
		workersToUpdate := workers.DeepCopy()
		workersToUpdate.Spec.Replicas = &replicas
		if !reflect.DeepEqual(workers, workersToUpdate) {
			err = client.Update(context.TODO(), workersToUpdate)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// GetStatefulset gets the statefulset by name and namespace
func GetStatefulSet(c client.Client, name string, namespace string) (master *appsv1.StatefulSet, err error) {
	master = &appsv1.StatefulSet{}
	err = c.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, master)

	return master, err
}

// GetPodsForStatefulSet gets pods of the specified statefulset
func GetPodsForStatefulSet(c client.Client, sts *appsv1.StatefulSet, selector labels.Selector) (pods []v1.Pod, err error) {

	podList := &v1.PodList{}
	err = c.List(context.TODO(), podList, &client.ListOptions{
		Namespace:     sts.Namespace,
		LabelSelector: selector,
	})

	if err != nil {
		log.Error(err, "Failed to list pods for statefulset")
		return
	}

	for _, pod := range podList.Items {
		if isMemberOf(sts, &pod) {
			controllerRef := metav1.GetControllerOf(&pod)
			if controllerRef != nil {
				// No controller should care about orphans being deleted.
				matched, err := compareOwnerRefMatcheWithExpected(c, controllerRef, pod.Namespace, sts)
				if err != nil {
					return pods, err
				}
				if matched {
					pods = append(pods, pod)
				}
				// wantedSet, err := resolveControllerRef(c, controllerRef, set.Namespace, statefulSetControllerKind)
			}
		}
	}

	return
}

// statefulPodRegex is a regular expression that extracts the parent StatefulSet and ordinal from the Name of a Pod
var statefulPodRegex = regexp.MustCompile("(.*)-([0-9]+)$")

// getParentNameAndOrdinal gets the name of pod's parent StatefulSet and pod's ordinal as extracted from its Name. If
// the Pod was not created by a StatefulSet, its parent is considered to be empty string, and its ordinal is considered
// to be -1.
func getParentNameAndOrdinal(pod *v1.Pod) (string, int) {
	parent := ""
	ordinal := -1
	subMatches := statefulPodRegex.FindStringSubmatch(pod.Name)
	if len(subMatches) < 3 {
		return parent, ordinal
	}
	parent = subMatches[1]
	if i, err := strconv.ParseInt(subMatches[2], 10, 32); err == nil {
		ordinal = int(i)
	}
	return parent, ordinal
}

// getParentName gets the name of pod's parent StatefulSet. If pod has not parent, the empty string is returned.
func getParentName(pod *v1.Pod) string {
	parent, _ := getParentNameAndOrdinal(pod)
	return parent
}

// isMemberOf tests if pod is a member of statefulset sts.
func isMemberOf(sts *appsv1.StatefulSet, pod *v1.Pod) bool {
	return getParentName(pod) == sts.Name
}

// GetPhaseFromStatefulset gets the phase from statefulset
func GetPhaseFromStatefulset(replicas int32, sts appsv1.StatefulSet) (phase datav1alpha1.RuntimePhase) {
	if replicas == 0 {
		phase = datav1alpha1.RuntimePhaseReady
		return
	}
	if sts.Status.ReadyReplicas > 0 {
		if replicas <= sts.Status.ReadyReplicas {
			phase = datav1alpha1.RuntimePhaseReady
		} else {
			phase = datav1alpha1.RuntimePhasePartialReady
		}
	} else {
		phase = datav1alpha1.RuntimePhaseNotReady
	}

	return

}
