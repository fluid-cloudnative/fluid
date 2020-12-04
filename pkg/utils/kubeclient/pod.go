/*

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

	"k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IsCompletePod determines if the pod is complete
func IsCompletePod(pod *v1.Pod) bool {
	if pod == nil {
		return false
	}

	if pod.DeletionTimestamp != nil {
		return true
	}

	if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
		return true
	}
	return false
}

// IsFailedPod determines if the pod is failed
func IsFailedPod(pod *v1.Pod) bool {
	return pod != nil && pod.Status.Phase == v1.PodFailed
}

// GetPodByName gets pod with given name and namespace of the pod.
func GetPodByName(client client.Client, name, namespace string) (pod *v1.Pod, err error) {
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	pod = &v1.Pod{}

	if err = client.Get(context.TODO(), key, pod); err != nil {
		if apierrs.IsNotFound(err) {
			err = nil
			pod = nil
		}
		return pod, err
	}

	return
}

//DeletePod deletes the given pod if it exists
func DeletePod(client client.Client, pod *v1.Pod) error {
	err := client.Delete(context.TODO(), pod)
	if apierrs.IsNotFound(err) {
		err = nil
	}
	return err
}

// exclude Inactive pod when compute ports
func ExcludeInactivePod(pod *v1.Pod) bool {
	// pod not assigned
	if len(pod.Spec.NodeName) == 0 {
		return true
	}
	// pod is Successed or failed
	if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
		return true
	}
	return false
}
