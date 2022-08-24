/*
Copyright 2022 The Fluid Author.

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

package runtime

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilpointer "k8s.io/utils/pointer"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func Test_scaleoutRuntimeControllerIfNeeded(t *testing.T) {
	type args struct {
		key types.NamespacedName
		log logr.Logger
	}
	tests := []struct {
		name      string
		args      args
		wantScale bool
		wantErr   bool
	}{
		// TODO: Add test cases.
		{},
	}

	objs := []runtime.Object{}
	s := runtime.NewScheme()
	_ = appsv1.AddToScheme(s)
	deployments := []*appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "alluxioruntime-controller",
				Namespace: common.NamespaceFluidSystem,
			}, Spec: appsv1.DeploymentSpec{
				Replicas: utilpointer.Int32Ptr(0),
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "jindoruntime-controller",
				Namespace: common.NamespaceFluidSystem,
			}, Spec: appsv1.DeploymentSpec{
				Replicas: utilpointer.Int32Ptr(1),
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "juicefsruntime-controller",
				Namespace: common.NamespaceFluidSystem,
			}, Spec: appsv1.DeploymentSpec{
				Replicas: utilpointer.Int32Ptr(0),
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "goosefsruntime-controller",
				Namespace: common.NamespaceFluidSystem,
				Annotations: map[string]string{
					"managers.fluid.io/replicas": "3",
				},
			}, Spec: appsv1.DeploymentSpec{
				Replicas: utilpointer.Int32Ptr(0),
			},
		},
	}

	for _, deployment := range deployments {
		objs = append(objs, deployment)
	}

	fakeClient := fake.NewFakeClientWithScheme(s, objs...)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotScale, err := scaleoutRuntimeControllerIfNeeded(fakeClient, tt.args.key, tt.args.log)
			if (err != nil) != tt.wantErr {
				t.Errorf("scaleoutRuntimeControllerIfNeeded() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotScale != tt.wantScale {
				t.Errorf("scaleoutRuntimeControllerIfNeeded() = %v, want %v", gotScale, tt.wantScale)
			}
		})
	}
}
