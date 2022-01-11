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

package object

import (
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"k8s.io/apimachinery/pkg/runtime"

	corev1 "k8s.io/api/core/v1"
)

type RuntimeApplication struct {
	obj runtime.Object
}

type RuntimeAnchor struct {
	fields []string
}

func NewRuntimeAnchor(fields []string, end string) common.Pointer {
	fieldsToAdd := []string{}
	if len(end) > 0 {
		for _, field := range fields {
			fieldsToAdd = append(fieldsToAdd, field)
			if field == end {
				break
			}
		}
	} else {
		fieldsToAdd = fields
	}

	return &RuntimeAnchor{
		fields: fieldsToAdd,
	}
}

func (a RuntimeAnchor) Key() (id string) {
	return strings.Join(a.fields, "/")
}

func (a RuntimeAnchor) Path() (paths []string) {
	return a.fields
}

func (a RuntimeAnchor) String() string {
	return a.Key()
}

func NewRuntimeApplication(obj runtime.Object) common.Application {
	return &RuntimeApplication{
		obj: obj,
	}
}

func (u *RuntimeApplication) GetObject() (obj runtime.Object) {
	return u.obj
}

func (u *RuntimeApplication) SetContainers(containers []corev1.Container, fields ...string) {

}

func (u *RuntimeApplication) SetVolumes(volumes []corev1.Volume, fields ...string) {

}

func (u *RuntimeApplication) GetVolumes(fields ...string) (volumes []corev1.Volume) {
	return
}

func (u *RuntimeApplication) GetContainers(fields ...string) (containers []corev1.Container) {
	return
}

func (u *RuntimeApplication) LocateContainers() (anchors []common.Pointer, err error) {
	return
}

func (u *RuntimeApplication) LocateVolumes() (anchors []common.Pointer, err error) {
	return
}

func (u *RuntimeApplication) LocateVolumeMounts() (anchors []common.Pointer, err error) {
	return
}

func (u *RuntimeApplication) LocateRuntimeSpec() (anchors []common.Pointer, err error) {
	return
}

func (u *RuntimeApplication) LocatePodSpecs() (anchors []common.Pointer, err error) {
	return
}
