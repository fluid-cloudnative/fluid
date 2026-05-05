/*
Copyright 2020 The Kruise Authors.

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

package lifecycle

import (
	"fmt"
	"strings"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/controllers/workload/v1alpha1/utils/podadapter"
	"github.com/fluid-cloudnative/fluid/pkg/controllers/workload/v1alpha1/utils/podreadiness"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	workloadv1alpha1 "github.com/fluid-cloudnative/fluid/api/workload/v1alpha1"
)

const (
	// these keys for MarkPodNotReady Policy of pod lifecycle
	preparingDeleteHookKey = "preDeleteHook"
	preparingUpdateHookKey = "preUpdateHook"
)

// Interface for managing pods lifecycle.
type Interface interface {
	UpdatePodLifecycle(pod *v1.Pod, state workloadv1alpha1.LifecycleStateType, markPodNotReady bool) (bool, *v1.Pod, error)
	UpdatePodLifecycleWithHandler(pod *v1.Pod, state workloadv1alpha1.LifecycleStateType, inPlaceUpdateHandler *workloadv1alpha1.LifecycleHook) (bool, *v1.Pod, error)
}

type realControl struct {
	adp                 podadapter.Adapter
	podReadinessControl podreadiness.Interface
}

func New(c client.Client) Interface {
	adp := &podadapter.AdapterRuntimeClient{Client: c}
	return &realControl{
		adp:                 adp,
		podReadinessControl: podreadiness.NewForAdapter(adp),
	}
}

func NewForTypedClient(c clientset.Interface) Interface {
	adp := &podadapter.AdapterTypedClient{Client: c}
	return &realControl{
		adp:                 adp,
		podReadinessControl: podreadiness.NewForAdapter(adp),
	}
}

func NewForInformer(informer coreinformers.PodInformer) Interface {
	adp := &podadapter.AdapterInformer{PodInformer: informer}
	return &realControl{
		adp:                 adp,
		podReadinessControl: podreadiness.NewForAdapter(adp),
	}
}

func GetPodLifecycleState(pod *v1.Pod) workloadv1alpha1.LifecycleStateType {
	if pod == nil || pod.Labels == nil {
		return ""
	}
	return workloadv1alpha1.LifecycleStateType(pod.Labels[workloadv1alpha1.LifecycleStateKey])
}

func IsHookMarkPodNotReady(lifecycleHook *workloadv1alpha1.LifecycleHook) bool {
	if lifecycleHook == nil {
		return false
	}
	return lifecycleHook.MarkPodNotReady
}

func IsLifecycleMarkPodNotReady(lifecycle *workloadv1alpha1.Lifecycle) bool {
	if lifecycle == nil {
		return false
	}
	return IsHookMarkPodNotReady(lifecycle.PreDelete) || IsHookMarkPodNotReady(lifecycle.InPlaceUpdate)
}

func SetPodLifecycle(state workloadv1alpha1.LifecycleStateType) func(*v1.Pod) {
	return func(pod *v1.Pod) {
		if pod == nil {
			return
		}
		if pod.Labels == nil {
			pod.Labels = make(map[string]string)
		}
		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}
		pod.Labels[workloadv1alpha1.LifecycleStateKey] = string(state)
		pod.Annotations[workloadv1alpha1.LifecycleTimestampKey] = time.Now().Format(time.RFC3339)
	}
}

func (c *realControl) executePodNotReadyPolicy(pod *v1.Pod, state workloadv1alpha1.LifecycleStateType) (err error) {
	switch state {
	case workloadv1alpha1.LifecycleStatePreparingDelete:
		err = c.podReadinessControl.AddNotReadyKey(pod, getReadinessMessage(preparingDeleteHookKey))
	case workloadv1alpha1.LifecycleStatePreparingUpdate:
		err = c.podReadinessControl.AddNotReadyKey(pod, getReadinessMessage(preparingUpdateHookKey))
	case workloadv1alpha1.LifecycleStateUpdated:
		err = c.podReadinessControl.RemoveNotReadyKey(pod, getReadinessMessage(preparingUpdateHookKey))
	}

	if err != nil {
		klog.ErrorS(err, "Failed to set pod Ready/NotReady at lifecycle state",
			"pod", client.ObjectKeyFromObject(pod), "state", state)
	}
	return
}

func (c *realControl) UpdatePodLifecycle(pod *v1.Pod, state workloadv1alpha1.LifecycleStateType, markPodNotReady bool) (updated bool, gotPod *v1.Pod, err error) {
	if markPodNotReady {
		if err = c.executePodNotReadyPolicy(pod, state); err != nil {
			return false, nil, err
		}
	}

	if GetPodLifecycleState(pod) == state {
		return false, pod, nil
	}

	pod = pod.DeepCopy()
	if adp, ok := c.adp.(podadapter.AdapterWithPatch); ok {
		body := fmt.Sprintf(
			`{"metadata":{"labels":{"%s":"%s"},"annotations":{"%s":"%s"}}}`,
			workloadv1alpha1.LifecycleStateKey,
			string(state),
			workloadv1alpha1.LifecycleTimestampKey,
			time.Now().Format(time.RFC3339),
		)
		gotPod, err = adp.PatchPod(pod, client.RawPatch(types.StrategicMergePatchType, []byte(body)))
	} else {
		SetPodLifecycle(state)(pod)
		gotPod, err = c.adp.UpdatePod(pod)
	}

	return true, gotPod, err
}

func (c *realControl) UpdatePodLifecycleWithHandler(pod *v1.Pod, state workloadv1alpha1.LifecycleStateType, inPlaceUpdateHandler *workloadv1alpha1.LifecycleHook) (updated bool, gotPod *v1.Pod, err error) {
	if inPlaceUpdateHandler == nil || pod == nil {
		return false, pod, nil
	}

	if inPlaceUpdateHandler.MarkPodNotReady {
		if err = c.executePodNotReadyPolicy(pod, state); err != nil {
			return false, nil, err
		}
	}

	if GetPodLifecycleState(pod) == state {
		return false, pod, nil
	}

	pod = pod.DeepCopy()
	if adp, ok := c.adp.(podadapter.AdapterWithPatch); ok {
		var labelsHandler, finalizersHandler string
		for k, v := range inPlaceUpdateHandler.LabelsHandler {
			labelsHandler = fmt.Sprintf(`%s,"%s":"%s"`, labelsHandler, k, v)
		}
		for _, v := range inPlaceUpdateHandler.FinalizersHandler {
			finalizersHandler = fmt.Sprintf(`%s,"%s"`, finalizersHandler, v)
		}
		finalizersHandler = fmt.Sprintf(`[%s]`, strings.TrimLeft(finalizersHandler, ","))

		body := fmt.Sprintf(
			`{"metadata":{"labels":{"%s":"%s"%s},"annotations":{"%s":"%s"},"finalizers":%s}}`,
			workloadv1alpha1.LifecycleStateKey,
			string(state),
			labelsHandler,
			workloadv1alpha1.LifecycleTimestampKey,
			time.Now().Format(time.RFC3339),
			finalizersHandler,
		)
		gotPod, err = adp.PatchPod(pod, client.RawPatch(types.StrategicMergePatchType, []byte(body)))
	} else {
		if pod.Labels == nil {
			pod.Labels = make(map[string]string)
		}
		for k, v := range inPlaceUpdateHandler.LabelsHandler {
			pod.Labels[k] = v
		}
		pod.Finalizers = append(pod.Finalizers, inPlaceUpdateHandler.FinalizersHandler...)

		SetPodLifecycle(state)(pod)
		gotPod, err = c.adp.UpdatePod(pod)
	}

	return true, gotPod, err
}

func IsPodHooked(hook *workloadv1alpha1.LifecycleHook, pod *v1.Pod) bool {
	if hook == nil || pod == nil {
		return false
	}
	for _, f := range hook.FinalizersHandler {
		if controllerutil.ContainsFinalizer(pod, f) {
			return true
		}
	}
	for k, v := range hook.LabelsHandler {
		if pod.Labels[k] == v {
			return true
		}
	}
	return false
}

func IsPodAllHooked(hook *workloadv1alpha1.LifecycleHook, pod *v1.Pod) bool {
	if hook == nil || pod == nil {
		return false
	}
	for _, f := range hook.FinalizersHandler {
		if !controllerutil.ContainsFinalizer(pod, f) {
			return false
		}
	}
	for k, v := range hook.LabelsHandler {
		if pod.Labels[k] != v {
			return false
		}
	}
	return true
}

func getReadinessMessage(key string) podreadiness.Message {
	return podreadiness.Message{UserAgent: "Lifecycle", Key: key}
}
