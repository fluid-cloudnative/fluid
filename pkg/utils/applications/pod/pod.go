/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
