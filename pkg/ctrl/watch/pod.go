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

package watch

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"

	"strings"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type podEventHandler struct {
}

func (h *podEventHandler) onCreateFunc(r Controller) func(e event.CreateEvent) bool {
	return func(e event.CreateEvent) (onCreate bool) {
		// ignore create event
		pod, ok := e.Object.(*corev1.Pod)
		if !ok {
			log.Info("pod.onCreateFunc Skip", "object", e.Object)
			return false
		}

		if !ShouldInQueue(pod) {
			log.Info("podEventHandler.onCreateFunc skip due to shouldRequeue false")
			return false
		}

		log.V(1).Info("podEventHandler.onCreateFunc", "name", pod.GetName(), "namespace", pod.GetNamespace())
		return true
	}
}

func (h *podEventHandler) onUpdateFunc(r Controller) func(e event.UpdateEvent) bool {
	return func(e event.UpdateEvent) (needUpdate bool) {
		podNew, ok := e.ObjectNew.(*corev1.Pod)
		if !ok {
			log.Info("pod.onUpdateFunc Skip", "object", e.ObjectNew)
			return needUpdate
		}

		podOld, ok := e.ObjectOld.(*corev1.Pod)
		if !ok {
			log.Info("pod.onUpdateFunc Skip", "object", e.ObjectNew)
			return needUpdate
		}

		if podNew.GetResourceVersion() == podOld.GetResourceVersion() {
			log.V(1).Info("pod.onUpdateFunc Skip due to resourceVersion not changed")
			return needUpdate
		}

		// ignore if it's not fluid label pod
		if !ShouldInQueue(podNew) {
			log.Info("podEventHandler.onUpdateFunc skip due to shouldRequeue false")
			return false
		}

		log.Info("podEventHandler.onUpdateFunc", "name", podNew.GetName(), "namespace", podNew.GetNamespace())
		return true
	}
}

func (h *podEventHandler) onDeleteFunc(r Controller) func(e event.DeleteEvent) bool {
	return func(e event.DeleteEvent) bool {
		// ignore delete event
		return false
	}
}

func ShouldInQueue(pod *corev1.Pod) bool {
	if pod == nil {
		return false
	}

	// ignore if it's not fluid label pod
	if !utils.ServerlessEnabled(pod.Labels) {
		log.Info("Serverless not enable.", "labels", pod.Labels)
		return false
	}

	// ignore if it claims to ignore
	if utils.AppControllerDisabled(pod.Labels) {
		log.Info("Calim to make application controller ignore.", "labels", pod.Labels)
		return false
	}

	// ignore if restartPolicy is Always
	if pod.Spec.RestartPolicy == corev1.RestartPolicyAlways {
		log.Info("pod restart policy", "policy", pod.Spec.RestartPolicy)
		return false
	}

	// ignore if no fuse container
	exist := false
	for _, cn := range pod.Spec.Containers {
		if strings.Contains(cn.Name, common.FuseContainerName) {
			exist = true
			break
		}
	}
	if !exist {
		log.Info("There are no fluid fuse sidecar in pod.", "name", pod.Name, "namespace", pod.Namespace)
		return false
	}

	// ignore if pod status is not running
	if pod.Status.Phase != corev1.PodRunning || len(pod.Status.ContainerStatuses) < 2 {
		log.Info("Pod status is not running or containerStatus less than 2.", "name", pod.Name, "namespace", pod.Namespace)
		return false
	}

	// reconcile if all app containers exit 0 and fuse container not exit
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if !strings.Contains(containerStatus.Name, common.FuseContainerName) {
			log.V(1).Info("container status", "status", containerStatus)
			if containerStatus.State.Terminated == nil {
				log.Info("fluid app not exited", "pod", pod.Name, "container", containerStatus.Name, "namespace", pod.Namespace)
				// container not exist
				return false
			}
		}
		if strings.Contains(containerStatus.Name, common.FuseContainerName) {
			if containerStatus.State.Running == nil {
				log.Info("fluid fuse not running", "pod", pod.Name, "container", containerStatus.Name, "namespace", pod.Namespace)
				return false
			}
		}
	}
	return true
}
