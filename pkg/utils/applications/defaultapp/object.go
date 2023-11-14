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

package defaultapp

import (
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type DefaultApplication struct {
	obj runtime.Object
	app *DefaultObject
}

func NewApplication(runtimeObject runtime.Object) common.FluidApplication {
	return &DefaultApplication{
		obj: runtimeObject,
		app: &DefaultObject{runtimeObject: runtimeObject},
	}
}

func (u *DefaultApplication) GetPodSpecs() (specs []common.FluidObject, err error) {
	specs = []common.FluidObject{}
	err = fmt.Errorf("not implemented")
	return
}

func (u *DefaultApplication) SetPodSpecs(specs []common.FluidObject) (err error) {
	err = fmt.Errorf("not implemented")
	return
}

func (u *DefaultApplication) GetObject() runtime.Object {
	return u.obj
}

type DefaultObject struct {
	runtimeObject runtime.Object
}

func (o *DefaultObject) GetRoot() runtime.Object {
	return o.runtimeObject
}

func (o *DefaultObject) GetVolumes() (volumes []corev1.Volume, err error) {
	err = fmt.Errorf("not implemented")
	return
}

func (o *DefaultObject) SetVolumes(volumes []corev1.Volume) (err error) {
	err = fmt.Errorf("not implemented")
	return
}

func (o *DefaultObject) GetContainers() (containers []corev1.Container, err error) {
	err = fmt.Errorf("not implemented")
	return
}

func (o *DefaultObject) SetContainers(containers []corev1.Container) (err error) {
	err = fmt.Errorf("not implemented")
	return
}

func (o *DefaultApplication) GetVolumeMounts() (volumeMounts []corev1.VolumeMount, err error) {
	err = fmt.Errorf("not implemented")
	return
}

func (o *DefaultApplication) SetMetaObject(metaObject metav1.ObjectMeta) (err error) {
	err = fmt.Errorf("not implemented")
	return
}

func (o *DefaultApplication) GetMetaObject() (metaObject metav1.ObjectMeta, err error) {
	err = fmt.Errorf("not implemented")
	return
}
