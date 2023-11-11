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

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetDaemonset gets the daemonset by name and namespace
func GetDaemonset(c client.Client, name string, namespace string) (ds *appsv1.DaemonSet, err error) {
	ds = &appsv1.DaemonSet{}
	err = c.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, ds)

	return ds, err
}

// GetDaemonPods gets pods of the specified daemonset
func GetDaemonPods(c client.Client, ds *appsv1.DaemonSet) (pods []*v1.Pod, err error) {
	selector, err := metav1.LabelSelectorAsSelector(ds.Spec.Selector)
	if err != nil {
		return nil, err
	}

	podList := &v1.PodList{}
	err = c.List(context.TODO(), podList, &client.ListOptions{
		Namespace:     ds.Namespace,
		LabelSelector: selector,
	})

	if err != nil {
		log.Error(err, "Failed to list pods for daemonset")
		return
	}

	for _, pod := range podList.Items {
		pods = append(pods, &pod)
	}

	return
}

// GetUnavailableDaemonPods gets unavailable pods of the specified daemonset
func GetUnavailableDaemonPods(c client.Client, ds *appsv1.DaemonSet) (unavailablePods []*v1.Pod, err error) {
	pods, err := GetDaemonPods(c, ds)
	if err != nil {
		return
	}

	for _, pod := range pods {
		if !isRunningAndReady(pod) {
			unavailablePods = append(unavailablePods, pod)
		}
	}

	return
}

// GetUnavailableDaemonPods gets unavailable pods of the specified daemonset
func GetUnavailableDaemonPodNames(c client.Client, ds *appsv1.DaemonSet) (names []types.NamespacedName, err error) {
	pods, err := GetUnavailableDaemonPods(c, ds)
	if err != nil {
		return
	}

	for _, pod := range pods {
		if !isRunningAndReady(pod) {
			names = append(names, types.NamespacedName{
				Namespace: pod.Namespace,
				Name:      pod.Name,
			})
		}
	}

	return
}
