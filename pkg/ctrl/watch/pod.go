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

  package utils


  import (
          "math/rand"
          "testing"
          "time"


          corev1 "k8s.io/api/core/v1"
  )

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
	if !utils.FuseSidecarPrivileged(pod.ObjectMeta) {
		log.Info("Privileged fuse sidecar is not enabled.", "labels", pod.Labels)
		return false
	}

	// ignore if not done
	if !utils.InjectSidecarDone(pod.Labels) {
		log.Info("Serverless inject not finished.", "labels", pod.Labels)
		return false
	}

	// ignore if it claims to ignore
	if utils.AppControllerDisabled(pod.Labels) {
		log.Info("Claim to make application controller ignore.", "labels", pod.Labels)
		return false
	}

	// ignore if restartPolicy is Always
	if pod.Spec.RestartPolicy == corev1.RestartPolicyAlways {
		log.Info("Pod restart policy", "policy", pod.Spec.RestartPolicy)
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
