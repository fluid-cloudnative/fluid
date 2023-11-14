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
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type PodApplication struct {
	pod *corev1.Pod
	app *PodObject
}

func NewApplication(pod *corev1.Pod) common.FluidApplication {
	return &PodApplication{
		pod: pod,
		app: &PodObject{pod: pod},
	}
}

func (u *PodApplication) GetPodSpecs() (specs []common.FluidObject, err error) {
	specs = []common.FluidObject{&PodObject{
		pod: u.pod,
	}}
	return
}

func (u *PodApplication) SetPodSpecs(specs []common.FluidObject) (err error) {
	if len(specs) != 1 {
		return fmt.Errorf("the length of specs should be 1, but %v", specs)
	}
	obj, ok := specs[0].(*PodObject)
	if !ok {
		return fmt.Errorf("the length of specs should be 1, but %v", specs)
	}
	u.pod = obj.pod
	return
}

func (u *PodApplication) GetObject() runtime.Object {
	return u.pod
}

type PodObject struct {
	pod *corev1.Pod
}

func (o *PodObject) GetRoot() runtime.Object {
	return o.pod
}

func (o *PodObject) GetVolumes() (volumes []corev1.Volume, err error) {
	volumes = o.pod.Spec.Volumes
	return
}

func (o *PodObject) SetVolumes(volumes []corev1.Volume) (err error) {
	o.pod.Spec.Volumes = volumes
	return
}

func (o *PodObject) GetContainers() (containers []corev1.Container, err error) {
	containers = o.pod.Spec.Containers
	return
}

func (o *PodObject) SetContainers(containers []corev1.Container) (err error) {
	o.pod.Spec.Containers = containers
	return
}

func (o *PodObject) GetInitContainers() (containers []corev1.Container, err error) {
	containers = o.pod.Spec.InitContainers
	return
}

func (o *PodObject) SetInitContainers(containers []corev1.Container) (err error) {
	o.pod.Spec.InitContainers = containers
	return
}

func (o *PodObject) GetVolumeMounts() (volumeMounts []corev1.VolumeMount, err error) {
	volumeMounts = []corev1.VolumeMount{}
	for _, container := range o.pod.Spec.Containers {
		volumeMounts = append(volumeMounts, container.VolumeMounts...)
	}

	for _, initContainer := range o.pod.Spec.InitContainers {
		volumeMounts = append(volumeMounts, initContainer.VolumeMounts...)
	}

	return
}

func (o *PodObject) SetMetaObject(metaObject metav1.ObjectMeta) (err error) {
	// o.pod.ObjectMeta = metaObject
	metaObject.DeepCopyInto(&o.pod.ObjectMeta)
	return
}

func (o *PodObject) GetMetaObject() (metaObject metav1.ObjectMeta, err error) {
	return o.pod.ObjectMeta, nil
}
