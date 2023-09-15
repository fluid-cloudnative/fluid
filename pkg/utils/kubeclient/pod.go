/*
  Copyright 2022 The Fluid Authors.

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

	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
)

// IsCompletePod determines if the pod is complete
func IsCompletePod(pod *corev1.Pod) bool {
	if pod == nil {
		return false
	}

	if pod.DeletionTimestamp != nil {
		return true
	}

	if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
		return true
	}
	return false
}

// IsFinishedPod determines if the pod is succeeded or failed
func IsFinishedPod(pod *corev1.Pod) bool {
	return pod != nil && (pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed)
}

// IsSucceededPod determines if the pod is Succeeded
func IsSucceededPod(pod *corev1.Pod) bool {
	return pod != nil && pod.Status.Phase == corev1.PodSucceeded
}

// IsFailedPod determines if the pod is failed
func IsFailedPod(pod *corev1.Pod) bool {
	return pod != nil && pod.Status.Phase == corev1.PodFailed
}

// GetPodByName gets pod with given name and namespace of the pod.
func GetPodByName(client client.Client, name, namespace string) (pod *corev1.Pod, err error) {
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	pod = &corev1.Pod{}

	if err = client.Get(context.TODO(), key, pod); err != nil {
		if apierrs.IsNotFound(err) {
			err = nil
			pod = nil
		}
		return pod, err
	}

	return
}

// GetPVCNamesFromPod get names of pvc mounted by Pod
func GetPVCNamesFromPod(pod *corev1.Pod) (pvcNames []string) {
	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim == nil {
			continue
		}
		pvcNames = append(pvcNames, volume.PersistentVolumeClaim.ClaimName)
	}
	return
}

// isRunningAndReady returns true if pod is in the PodRunning Phase, if it has a condition of PodReady.
func isRunningAndReady(pod *corev1.Pod) bool {
	return pod.Status.Phase == corev1.PodRunning && podutil.IsPodReady(pod)
}

// GetIpAddressesOfPods gets the ipAddresses of pods
func GetIpAddressesOfPods(client client.Client, pods []corev1.Pod) (ipAddresses []string, err error) {
	// nodes := []corev1.Node{}
	nodes := make([]corev1.Node, 0, len(pods))
	for _, pod := range pods {
		nodeName := pod.Spec.NodeName
		node, err := GetNode(client, nodeName)
		if err != nil {
			return ipAddresses, err
		}
		nodes = append(nodes, *node)
	}

	return GetIpAddressesOfNodes(nodes), err
}

func AppendNodeSelectorToNodeAffinity(nodeSelector map[string]string, nodeAffinity *corev1.NodeAffinity) *corev1.NodeAffinity {
	if nodeAffinity == nil {
		nodeAffinity = &corev1.NodeAffinity{}
	}

	if nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{},
		}
	}

	for key, value := range nodeSelector {
		term := corev1.NodeSelectorTerm{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{
					Key:      key,
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{value},
				},
			},
		}
		nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms =
			append(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, term)
	}
	return nodeAffinity
}
