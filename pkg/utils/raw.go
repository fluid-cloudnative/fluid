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

package utils

import (
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/types/cacheworkerset"
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
		{appsv1.SchemeGroupVersion, &cacheworkerset.CacheWorkerSet{}},
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
