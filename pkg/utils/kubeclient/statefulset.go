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

package kubeclient

import (
	"context"
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
		if replicas == sts.Status.ReadyReplicas {
			phase = datav1alpha1.RuntimePhaseReady
		} else {
			phase = datav1alpha1.RuntimePhasePartialReady
		}
	} else {
		phase = datav1alpha1.RuntimePhaseNotReady
	}

	return

}

// GetUnavailablePodsStatefulSet gets unavailable pods of the specified statefulset
func GetUnavailablePodsStatefulSet(c client.Client, sts *appsv1.StatefulSet, selector labels.Selector) (unavailablePods []*v1.Pod, err error) {

	pods, err := GetPodsForStatefulSet(c, sts, selector)
	if err != nil {
		return
	}

	for _, pod := range pods {
		if !isRunningAndReady(&pod) {
			unavailablePods = append(unavailablePods, &pod)
		}
	}

	return
}

// GetUnavailablePodNamesForStatefulSet gets pod names of the specified statefulset
func GetUnavailablePodNamesForStatefulSet(c client.Client, sts *appsv1.StatefulSet, selector labels.Selector) (names []types.NamespacedName, err error) {

	pods, err := GetUnavailablePodsStatefulSet(c, sts, selector)
	if err != nil {
		return
	}

	for _, pod := range pods {
		names = append(names, types.NamespacedName{
			Namespace: pod.Namespace,
			Name:      pod.Name,
		})
	}

	return
}
