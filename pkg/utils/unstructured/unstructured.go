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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// UnstructuredApp allows objects that do not have Golang structs registered to be manipulated
// generically. This can be used to deal with the API objects from a plug-in. UnstructuredApp
// objects can handle the common object like Container, Volume
type UnstructuredApplication struct {
	obj *unstructured.Unstructured
}

type Anchor struct {
	fields []string
}

func (a Anchor) Key() (id string) {
	for _, field := range a.fields {
		id = id + "/" + field
	}
	return
}

func (a Anchor) Path() []string {
	return a.fields
}

func NewUnstructuredApplication(obj *unstructured.Unstructured) *UnstructuredApplication {
	return &UnstructuredApplication{
		obj: obj,
	}
}

func ()

func (u *UnstructuredApplication) SetContainers(containers []corev1.Container, fields ...string) {

}

func (u *UnstructuredApplication) SetVolumes(volumes []corev1.Volume, fields ...string) {

}

func (u *UnstructuredApplication) GetVolumes(fields ...string) (volumes []corev1.Volume) {
	return
}

func (u *UnstructuredApplication) GetContainers(fields ...string) (containers []corev1.Container) {
	return
}

func (u *UnstructuredApplication) LocateContainers() (anchors []Anchor) {
	return
}

func (u *UnstructuredApplication) LocateVolumes() (anchors []Anchor) {
	return
}
