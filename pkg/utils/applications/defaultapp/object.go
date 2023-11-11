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
