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
	"fmt"
	"k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetJob gets the job given its name and namespace
func GetJob(client client.Client, name, namespace string) (*v1.Job, error) {
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var job v1.Job
	if err := client.Get(context.TODO(), key, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func UpdateJob(client client.Client, job *v1.Job) error {
	return client.Update(context.TODO(), job)
}

// GetSucceedPodForJob get the first finished pod for the job, if no succeed pod, return nil with no error.
func GetSucceedPodForJob(c client.Client, job *v1.Job) (*corev1.Pod, error) {
	var podList corev1.PodList
	selector, err := metav1.LabelSelectorAsSelector(job.Spec.Selector)
	if err != nil {
		return nil, fmt.Errorf("error converting Job %s in namespace %s selector: %v", job.Name, job.Namespace, err)
	}
	err = c.List(context.TODO(), &podList, &client.ListOptions{
		Namespace:     job.Namespace,
		LabelSelector: selector,
	})

	for _, pod := range podList.Items {
		if pod.Status.Phase == corev1.PodSucceeded {
			return &pod, nil
		}
	}
	// no succeed job, return nil with no error.
	return nil, nil
}
