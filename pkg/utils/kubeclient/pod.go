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

func MergeNodeSelectorAndNodeAffinity(nodeSelector map[string]string, podAffinity *corev1.Affinity) (nodeAffinity *corev1.NodeAffinity) {
	if podAffinity != nil && podAffinity.NodeAffinity != nil {
		nodeAffinity = podAffinity.NodeAffinity.DeepCopy()
	}

	// no node affinity
	if nodeAffinity == nil {
		nodeAffinity = &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				// required field, can not be omitted
				NodeSelectorTerms: []corev1.NodeSelectorTerm{},
			},
		}
	}

	// has preferred affinity, but no required affinity
	if nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{
			// required field, can not be omitted
			NodeSelectorTerms: []corev1.NodeSelectorTerm{},
		}
	}

	var expressions []corev1.NodeSelectorRequirement
	for key, value := range nodeSelector {
		expressions = append(expressions,
			corev1.NodeSelectorRequirement{
				Key:      key,
				Operator: corev1.NodeSelectorOpIn,
				Values:   []string{value},
			},
		)
	}
	// inject to MatchExpressions for And relation.
	if len(expressions) > 0 {
		if len(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) == 0 {
			nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = []corev1.NodeSelectorTerm{
				{
					MatchExpressions: expressions,
				},
			}
		} else {
			for idx := range nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
				nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[idx].MatchExpressions =
					append(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[idx].MatchExpressions, expressions...)
			}
		}
	}
	return
}
