/*
Copyright 2021 The Fluid Authors.

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

package unstructured

import (
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Object interface {
	// getRoot() *unstructured.Unstructured
	common.Object
}

type unstructuredObject struct {
	Object
}

func (u *unstructuredObject) getRoot() (root *unstructured.Unstructured, err error) {
	original := u.GetRoot()

	root, ok := original.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("failed to convert %v to unstructured.Unstructure", original)
	}

	return
}

func (u *unstructuredObject) GetVolumes() (volumes []corev1.Volume, err error) {
	root, err := u.getRoot()
	if err != nil {
		return nil, err
	}

	field, found, err := unstructured.NestedFieldCopy(root.Object, u.GetVolumesPtr().Paths()...)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("failed to find the volumes from %v", u.GetVolumesPtr().Paths())
	}
	original, ok := field.([]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to parse %v", field)
	}

	volumes = make([]corev1.Volume, 0, len(original))
	for _, obj := range original {
		o, ok := obj.(map[string]interface{})
		if !ok {
			// expected map[string]interface{}, got something else
			return nil, fmt.Errorf("failed to parse %v", obj)
		}

		volume, err := convertMapToVolume(o)
		if err != nil {
			return nil, err
		}
		volumes = append(volumes, volume)
	}

	return
}

func (u *unstructuredObject) SetVolumes(volumes []corev1.Volume) (err error) {
	root, err := u.getRoot()
	if err != nil {
		return err
	}

	if len(volumes) == 0 {
		unstructured.RemoveNestedField(root.Object, u.GetVolumesPtr().Paths()...)
		return
	}

	newVolumes := make([]interface{}, 0, len(volumes))
	for _, volume := range volumes {
		out, err := convertVolumeToMap(&volume)
		if err != nil {
			return err
		}
		newVolumes = append(newVolumes, out)
	}

	unstructured.SetNestedSlice(root.Object, newVolumes, u.GetVolumesPtr().Paths()...)

	return
}

func (u *unstructuredObject) GetContainers() (containers []corev1.Container, err error) {
	root, err := u.getRoot()
	if err != nil {
		return nil, err
	}

	field, found, err := unstructured.NestedFieldCopy(root.Object, u.GetContainersPtr().Paths()...)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("failed to find the containers from %v", u.GetContainersPtr().Paths())
	}
	original, ok := field.([]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to parse %v", field)
	}
	containers = make([]corev1.Container, 0, len(original))
	for _, obj := range original {
		o, ok := obj.(map[string]interface{})
		if !ok {
			// expected map[string]interface{}, got something else
			return nil, fmt.Errorf("failed to parse %v", obj)
		}
		container, err := convertMapToContainer(o)
		if err != nil {
			return nil, err
		}
		containers = append(containers, container)
	}

	return
}

func (u *unstructuredObject) SetContainers(containers []corev1.Container) (err error) {
	root, err := u.getRoot()
	if err != nil {
		return err
	}

	if len(containers) == 0 {
		unstructured.RemoveNestedField(root.Object, u.GetContainersPtr().Paths()...)
		return
	}

	newContainers := make([]interface{}, 0, len(containers))
	for _, container := range containers {
		out, err := convertContainerToMap(&container)
		if err != nil {
			return err
		}
		newContainers = append(newContainers, out)
	}

	unstructured.SetNestedSlice(root.Object, newContainers, u.GetContainersPtr().Paths()...)
	return
}
