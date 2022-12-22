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
	"bytes"
	"context"
	"io"

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

// IsSucceededPod determines if the pod is Succeeded
func IsSucceededPod(pod *corev1.Pod) bool {
	return pod != nil && pod.Status.Phase == corev1.PodSucceeded
}

// IsFailedPod determines if the pod is failed
func IsFailedPod(pod *corev1.Pod) bool {
	return pod != nil && pod.Status.Phase == corev1.PodFailed
}

// TailPodLogs tail pod's log.
// Given name and namespace of pod, lines eq: "tail -n <lines>".
func TailPodLogs(name, namespace string, lines int64) (logstr string, err error) {
	err = initClient()
	if err != nil {
		return "", err
	}
	req := clientset.CoreV1().Pods(namespace).GetLogs(name, &corev1.PodLogOptions{TailLines: &lines})
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}
	logstr = buf.String()
	return
}

// ListPodsByLabel gets podList with given namespace and labels.
func ListPodsByLabel(clientClient client.Client, namespace string, label map[string]string) (*corev1.PodList, error) {
	podList := &corev1.PodList{}
	err := clientClient.List(context.TODO(), podList, client.InNamespace(namespace), client.MatchingLabels(label))
	if err != nil {
		return nil, err
	}
	return podList, nil
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
