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

package watch

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type podEventHandler struct {
}

func (h *podEventHandler) onCreateFunc(r Controller) func(e event.CreateEvent) bool {
	return func(e event.CreateEvent) (onCreate bool) {
		// ignore create event
		return false
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
		if !shouldRequeue(podNew) {
			return false
		}

		log.V(1).Info("podEventHandler.onUpdateFunc", "name", podNew.GetName(), "namespace", podNew.GetNamespace())
		return true
	}
}

func (h *podEventHandler) onDeleteFunc(r Controller) func(e event.DeleteEvent) bool {
	return func(e event.DeleteEvent) bool {
		// ignore delete event
		return false
	}
}

func shouldRequeue(pod *corev1.Pod) bool {
	if pod == nil {
		return false
	}

	// ignore if it's not fluid label pod
	if !utils.ServerlessEnabled(pod.Labels) {
		return false
	}

	// ignore if restartPolicy is Always
	if pod.Spec.RestartPolicy == corev1.RestartPolicyAlways {
		return false
	}

	// ignore if no fuse container
	exist := false
	for _, cn := range pod.Spec.Containers {
		if cn.Name == common.FuseContainerName {
			exist = true
		}
	}
	if !exist {
		log.Info("There are no fluid fuse sidecar in pod.", "name", pod.Name, "namespace", pod.Namespace)
		return false
	}

	// reconcile if app container exit 0 and fuse container not exit
	appExited := false
	fuseExited := false
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name != common.FuseContainerName {
			if containerStatus.State.Terminated != nil && containerStatus.State.Terminated.ExitCode == 0 {
				appExited = true
			}
		}
		if containerStatus.Name == common.FuseContainerName {
			if containerStatus.State.Terminated != nil && containerStatus.State.Terminated.ExitCode == 0 {
				fuseExited = true
			}
		}
	}
	return appExited && !fuseExited
}
