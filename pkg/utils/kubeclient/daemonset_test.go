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

package kubeclient

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGetDaemonset(t *testing.T) {
	name := "test"
	namespace := "default"
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DaemonSetSpec{},
		Status: appsv1.DaemonSetStatus{
			NumberUnavailable: 1,
			NumberReady:       1,
		},
	}

	objs := []runtime.Object{}
	objs = append(objs, ds.DeepCopy())
	s := runtime.NewScheme()
	_ = appsv1.AddToScheme(s)
	fakeClient := fake.NewFakeClientWithScheme(s, objs...)

	_, err := GetDaemonset(fakeClient, name, namespace)
	if err != nil {
		t.Errorf("failed to call GetDaemonset due to %v with name %s and namespace %s",
			err,
			name,
			namespace)
	}

	_, err = GetDaemonset(fakeClient, "notFound", namespace)
	if err == nil {
		t.Errorf("failed to call GetDaemonset due to %v with name %s and namespace %s",
			err,
			name,
			namespace)
	}

}
