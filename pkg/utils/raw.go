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

package utils

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

var (
	kinds = []struct {
		groupVersion schema.GroupVersion
		obj          runtime.Object
	}{
		{corev1.SchemeGroupVersion, &corev1.ReplicationController{}},
		{corev1.SchemeGroupVersion, &corev1.Pod{}},
		{appsv1.SchemeGroupVersion, &appsv1.Deployment{}},
		{appsv1.SchemeGroupVersion, &appsv1.DaemonSet{}},
		{appsv1.SchemeGroupVersion, &appsv1.ReplicaSet{}},
		{batchv1.SchemeGroupVersion, &batchv1.Job{}},
		{appsv1.SchemeGroupVersion, &appsv1.StatefulSet{}},
		{corev1.SchemeGroupVersion, &corev1.List{}},
	}
	injectScheme = runtime.NewScheme()
)

func init() {
	for _, kind := range kinds {
		injectScheme.AddKnownTypes(kind.groupVersion, kind.obj)
		injectScheme.AddUnversionedTypes(kind.groupVersion, kind.obj)
	}
}

// FromRawToObject is used to convert from raw to the runtime object
func FromRawToObject(raw []byte) (obj runtime.Object, err error) {
	json, err := yaml.ToJSON(raw)
	if err != nil {
		return
	}

	obj, err = runtime.Decode(unstructured.UnstructuredJSONScheme, json)
	if err != nil {
		return
	}
	unstructured, ok := obj.(*unstructured.Unstructured)
	if !ok {
		err = fmt.Errorf("unstructured.Unstructured expected")
		return
	}

	gvk := schema.FromAPIVersionAndKind(unstructured.GetAPIVersion(), unstructured.GroupVersionKind().Kind)
	typedObj, err := injectScheme.New(gvk)
	if err == nil {
		if err = yaml.Unmarshal(raw, typedObj); err != nil {
			return nil, err
		}
		return typedObj, err
	} else if runtime.IsNotRegisteredError(err) {
		return unstructured, nil
	}

	return
}
