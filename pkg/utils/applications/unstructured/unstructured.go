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
