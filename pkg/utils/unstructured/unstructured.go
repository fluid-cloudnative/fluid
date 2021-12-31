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
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/mitchellh/mapstructure"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/nqd/flat"
)

const delimiter string = ":"

// UnstructuredApp allows objects that do not have Golang structs registered to be manipulated
// generically. This can be used to deal with the API objects from a plug-in. UnstructuredApp
// objects can handle the common object like Container, Volume
type UnstructuredApplication struct {
	obj *unstructured.Unstructured
}

type UnstructuredAnchor struct {
	fields []string
}

func NewUnstructuredAnchor(fields []string) common.Anchor {
	return &UnstructuredAnchor{
		fields: fields,
	}
}

func (a UnstructuredAnchor) Key() (id string) {
	for _, field := range a.fields {
		id = id + "/" + field
	}
	return
}

func (a UnstructuredAnchor) Path() []string {
	return a.fields
}

func NewUnstructuredApplication(obj *unstructured.Unstructured) *UnstructuredApplication {
	return &UnstructuredApplication{
		obj: obj,
	}
}

func (u *UnstructuredApplication) GetObject() (obj *unstructured.Unstructured) {
	return u.obj
}

func (u *UnstructuredApplication) SetContainers(containers []corev1.Container, fields ...string) {
	if len(containers) == 0 {
		unstructured.RemoveNestedField(u.obj.Object, fields...)
		return
	}

	newContainers := make([]interface{}, 0, len(containers))
	for _, container := range containers {
		out, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&container)
		if err != nil {
			utilruntime.HandleError(fmt.Errorf("unable to convert Container: %v", err))
			continue
		}
		newContainers = append(newContainers, out)
	}
	unstructured.SetNestedSlice(u.obj.Object, newContainers, fields...)
}

func (u *UnstructuredApplication) SetVolumes(volumes []corev1.Volume, fields ...string) {
	if len(volumes) == 0 {
		unstructured.RemoveNestedField(u.obj.Object, fields...)
		return
	}

	newVolumes := make([]interface{}, 0, len(volumes))
	for _, volume := range volumes {
		out, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&volume)
		if err != nil {
			utilruntime.HandleError(fmt.Errorf("unable to convert Volume: %v", err))
			continue
		}
		newVolumes = append(newVolumes, out)
	}
	unstructured.SetNestedSlice(u.obj.Object, newVolumes, fields...)
}

func (u *UnstructuredApplication) GetVolumes(fields ...string) (volumes []corev1.Volume) {
	field, found, err := unstructured.NestedFieldNoCopy(u.obj.Object, fields...)
	if !found || err != nil {
		return nil
	}
	original, ok := field.([]interface{})
	if !ok {
		return nil
	}
	vol := make([]corev1.Volume, 0, len(original))
	for _, obj := range original {
		var volume corev1.Volume
		o, ok := obj.(map[string]interface{})
		if !ok {
			// expected map[string]interface{}, got something else
			return nil
		}
		mapstructure.Decode(o, &volume)
		vol = append(vol, volume)
	}
	return vol
}

func (u *UnstructuredApplication) GetContainers(fields ...string) (containers []corev1.Container) {
	field, found, err := unstructured.NestedFieldNoCopy(u.obj.Object, fields...)
	if !found || err != nil {
		return nil
	}
	original, ok := field.([]interface{})
	if !ok {
		return nil
	}
	vol := make([]corev1.Container, 0, len(original))
	for _, obj := range original {
		var container corev1.Container
		o, ok := obj.(map[string]interface{})
		if !ok {
			// expected map[string]interface{}, got something else
			return nil
		}
		mapstructure.Decode(o, &container)
		vol = append(vol, container)
	}
	return vol
}

func (u *UnstructuredApplication) LocateContainers() (anchors []common.Anchor, err error) {

	out, err := flat.Flatten(u.obj.Object, &flat.Options{
		Delimiter: delimiter,
	})
	if err != nil {
		return anchors, err
	}
	for key, _ := range out {
		if strings.Contains(key, "containers") {
			anchors = append(anchors, NewUnstructuredAnchor(strings.Split(key, ":")))
		}
	}
	return anchors, err

}

func (u *UnstructuredApplication) LocateVolumes() (anchors []common.Anchor, err error) {
	return
}
