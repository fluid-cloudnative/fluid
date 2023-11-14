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
