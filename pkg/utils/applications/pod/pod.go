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

package pod

import (
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type PodApplication struct {
	obj *corev1.Pod
}

type PodAnchor struct {
	fields []string
}

func NewPodAnchor(fields []string, end string) common.Anchor {
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

	return &PodAnchor{
		fields: fieldsToAdd,
	}
}

func (a PodAnchor) Key() (id string) {
	return strings.Join(a.fields, "/")
}

func (a PodAnchor) Path() (paths []string) {
	return a.fields
}

func (a PodAnchor) String() string {
	return a.Key()
}

func NewPodApplication(obj *corev1.Pod) *PodApplication {
	return &PodApplication{
		obj: obj,
	}
}

func (u *PodApplication) GetObject() (obj runtime.Object) {
	return u.obj
}

func (u *PodApplication) SetContainers(containers []corev1.Container, fields ...string) {

}

func (u *PodApplication) SetVolumes(volumes []corev1.Volume, fields ...string) {

}

func (u *PodApplication) GetVolumes(fields ...string) (volumes []corev1.Volume) {
	return
}

func (u *PodApplication) GetContainers(fields ...string) (containers []corev1.Container) {
	return
}

func (u *PodApplication) LocateContainers() (anchors []common.Anchor, err error) {
	return
}

func (u *PodApplication) LocateVolumes() (anchors []common.Anchor, err error) {
	return
}

func (u *PodApplication) LocateVolumeMounts() (anchors []common.Anchor, err error) {
	return
}

func (u *PodApplication) LocatePodSpec() (anchors []common.Anchor, err error) {
	return
}
