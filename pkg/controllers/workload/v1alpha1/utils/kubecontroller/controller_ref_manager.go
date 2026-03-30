/*
Copyright 2016 The Kubernetes Authors.
Copyright 2024 The Fluid Authors.

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

package kubecontroller

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"
)

// BaseControllerRefManager provides common functionality for controller ref managers.
type BaseControllerRefManager struct {
	Controller metav1.Object
	Selector   labels.Selector

	canAdoptErr  error
	canAdoptOnce sync.Once
	CanAdoptFunc func(ctx context.Context) error
}

// CanAdopt checks whether the controller can adopt the object.
func (m *BaseControllerRefManager) CanAdopt(ctx context.Context) error {
	m.canAdoptOnce.Do(func() {
		if m.CanAdoptFunc != nil {
			m.canAdoptErr = m.CanAdoptFunc(ctx)
		}
	})
	return m.canAdoptErr
}

// ClaimObject tries to take ownership of an object for this controller.
func (m *BaseControllerRefManager) ClaimObject(ctx context.Context, obj metav1.Object, match func(metav1.Object) bool, adopt, release func(context.Context, metav1.Object) error) (bool, error) {
	controllerRef := metav1.GetControllerOfNoCopy(obj)
	if controllerRef != nil {
		if controllerRef.UID != m.Controller.GetUID() {
			return false, nil
		}
		if match(obj) {
			return true, nil
		}
		if m.Controller.GetDeletionTimestamp() != nil {
			return false, nil
		}
		if err := release(ctx, obj); err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
		return false, nil
	}

	if m.Controller.GetDeletionTimestamp() != nil || !match(obj) {
		return false, nil
	}
	if obj.GetDeletionTimestamp() != nil {
		return false, nil
	}

	if len(m.Controller.GetNamespace()) > 0 && m.Controller.GetNamespace() != obj.GetNamespace() {
		return false, nil
	}

	if err := adopt(ctx, obj); err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// PodControllerRefManager manages the controllerRef of pods.
type PodControllerRefManager struct {
	BaseControllerRefManager
	controllerKind schema.GroupVersionKind
	podControl     PodControlInterface
	finalizers     []string
}

// NewPodControllerRefManager returns a PodControllerRefManager.
func NewPodControllerRefManager(
	podControl PodControlInterface,
	controller metav1.Object,
	selector labels.Selector,
	controllerKind schema.GroupVersionKind,
	canAdopt func(ctx context.Context) error,
	finalizers ...string,
) *PodControllerRefManager {
	return &PodControllerRefManager{
		BaseControllerRefManager: BaseControllerRefManager{
			Controller:   controller,
			Selector:     selector,
			CanAdoptFunc: canAdopt,
		},
		controllerKind: controllerKind,
		podControl:     podControl,
		finalizers:     finalizers,
	}
}

// ClaimPods tries to take ownership of a list of Pods.
func (m *PodControllerRefManager) ClaimPods(ctx context.Context, pods []*v1.Pod, filters ...func(*v1.Pod) bool) ([]*v1.Pod, error) {
	var claimed []*v1.Pod
	var errlist []error

	match := func(obj metav1.Object) bool {
		pod := obj.(*v1.Pod)
		if !m.Selector.Matches(labels.Set(pod.Labels)) {
			return false
		}
		for _, filter := range filters {
			if !filter(pod) {
				return false
			}
		}
		return true
	}
	adopt := func(ctx context.Context, obj metav1.Object) error {
		return m.AdoptPod(ctx, obj.(*v1.Pod))
	}
	release := func(ctx context.Context, obj metav1.Object) error {
		return m.ReleasePod(ctx, obj.(*v1.Pod))
	}

	for _, pod := range pods {
		ok, err := m.ClaimObject(ctx, pod, match, adopt, release)
		if err != nil {
			errlist = append(errlist, err)
			continue
		}
		if ok {
			claimed = append(claimed, pod)
		}
	}
	return claimed, utilerrors.NewAggregate(errlist)
}

// AdoptPod sends a patch to take control of the pod.
func (m *PodControllerRefManager) AdoptPod(ctx context.Context, pod *v1.Pod) error {
	if err := m.CanAdopt(ctx); err != nil {
		return fmt.Errorf("can't adopt Pod %v/%v (%v): %v", pod.Namespace, pod.Name, pod.UID, err)
	}
	addControllerPatch, err := ownerRefControllerPatch(m.Controller, m.controllerKind, pod.UID, m.finalizers...)
	if err != nil {
		return err
	}
	return m.podControl.PatchPod(ctx, pod.Namespace, pod.Name, addControllerPatch)
}

// ReleasePod sends a patch to free the pod from the control of this controller.
func (m *PodControllerRefManager) ReleasePod(ctx context.Context, pod *v1.Pod) error {
	logger := klog.FromContext(ctx)
	logger.V(2).Info("Patching pod to remove its controllerRef", "pod", klog.KObj(pod), "gvk", m.controllerKind, "controller", m.Controller.GetName())
	patchBytes, err := GenerateDeleteOwnerRefStrategicMergeBytes(pod.UID, []types.UID{m.Controller.GetUID()}, m.finalizers...)
	if err != nil {
		return err
	}
	err = m.podControl.PatchPod(ctx, pod.Namespace, pod.Name, patchBytes)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		if errors.IsInvalid(err) {
			return nil
		}
	}
	return err
}

// RecheckDeletionTimestamp returns a CanAdopt() function to recheck deletion.
func RecheckDeletionTimestamp(getObject func(context.Context) (metav1.Object, error)) func(context.Context) error {
	return func(ctx context.Context) error {
		obj, err := getObject(ctx)
		if err != nil {
			return fmt.Errorf("can't recheck DeletionTimestamp: %v", err)
		}
		if obj.GetDeletionTimestamp() != nil {
			return fmt.Errorf("%v/%v has just been deleted at %v", obj.GetNamespace(), obj.GetName(), obj.GetDeletionTimestamp())
		}
		return nil
	}
}

type objectForOwnerRefPatch struct {
	Metadata objectMetaForPatch `json:"metadata"`
}

type objectMetaForPatch struct {
	OwnerReferences []metav1.OwnerReference `json:"ownerReferences"`
	UID             types.UID               `json:"uid"`
	Finalizers      []string                `json:"finalizers,omitempty"`
}

func ownerRefControllerPatch(controller metav1.Object, gvk schema.GroupVersionKind, uid types.UID, finalizers ...string) ([]byte, error) {
	blockOwnerDeletion := true
	isController := true
	addControllerPatch := objectForOwnerRefPatch{
		Metadata: objectMetaForPatch{
			UID: uid,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         gvk.GroupVersion().String(),
					Kind:               gvk.Kind,
					Name:               controller.GetName(),
					UID:                controller.GetUID(),
					Controller:         &isController,
					BlockOwnerDeletion: &blockOwnerDeletion,
				},
			},
		},
	}
	if len(finalizers) > 0 {
		addControllerPatch.Metadata.Finalizers = finalizers
	}
	patchBytes, err := json.Marshal(&addControllerPatch)
	if err != nil {
		return nil, err
	}
	return patchBytes, nil
}

// GenerateDeleteOwnerRefStrategicMergeBytes generates the patch bytes to remove
// the owner references with given uids from an object.
func GenerateDeleteOwnerRefStrategicMergeBytes(objUID types.UID, ownerUIDs []types.UID, finalizers ...string) ([]byte, error) {
	var ownerRefs []map[string]interface{}
	for _, ownerUID := range ownerUIDs {
		ownerRefs = append(ownerRefs, map[string]interface{}{
			"uid":    ownerUID,
			"$patch": "delete",
		})
	}
	patch := map[string]interface{}{
		"metadata": map[string]interface{}{
			"uid":             objUID,
			"ownerReferences": ownerRefs,
		},
	}
	if len(finalizers) > 0 {
		var finalizersToDelete []map[string]interface{}
		for _, f := range finalizers {
			finalizersToDelete = append(finalizersToDelete, map[string]interface{}{
				"$patch": "delete",
				"value":  f,
			})
		}
		patch["metadata"].(map[string]interface{})["$deleteFromPrimitiveList/finalizers"] = finalizers
	}
	return json.Marshal(patch)
}
