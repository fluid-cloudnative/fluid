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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	unstructuredtype "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type UnstructuredApplication struct {
	obj *unstructuredtype.Unstructured
	app *UnstructuredObject
}

func NewApplication(unstructured *unstructuredtype.Unstructured) common.FluidApplication {
	return &UnstructuredApplication{
		obj: unstructured,
		app: &UnstructuredObject{unstructured: unstructured},
	}
}

func (u *UnstructuredApplication) GetPodSpecs() (specs []common.FluidObject, err error) {
	specs = []common.FluidObject{}
	return
}

func (u *UnstructuredApplication) SetPodSpecs(specs []common.FluidObject) (err error) {
	err = fmt.Errorf("not implemented")
	return
}

func (u *UnstructuredApplication) GetObject() runtime.Object {
	return u.obj
}

type UnstructuredObject struct {
	unstructured *unstructuredtype.Unstructured
}

func (o *UnstructuredObject) GetRoot() runtime.Object {
	return o.unstructured
}

func (o *UnstructuredObject) GetVolumes() (volumes []corev1.Volume, err error) {
	err = fmt.Errorf("not implemented")
	return
}

func (o *UnstructuredObject) SetVolumes(volumes []corev1.Volume) (err error) {
	err = fmt.Errorf("not implemented")
	return
}

func (o *UnstructuredObject) GetContainers() (containers []corev1.Container, err error) {
	err = fmt.Errorf("not implemented")
	return
}

func (o *UnstructuredObject) SetContainers(containers []corev1.Container) (err error) {
	err = fmt.Errorf("not implemented")
	return
}

func (o *UnstructuredObject) GetVolumeMounts() (volumeMounts []corev1.VolumeMount, err error) {
	err = fmt.Errorf("not implemented")
	return
}

func (o *UnstructuredObject) SetMetaObject(metaObject metav1.ObjectMeta) (err error) {
	err = fmt.Errorf("not implemented")
	return
}

func (o *UnstructuredObject) GetMetaObject() (metaObject metav1.ObjectMeta, err error) {
	err = fmt.Errorf("not implemented")
	return
}
