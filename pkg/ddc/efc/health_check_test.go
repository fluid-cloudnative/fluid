/*
  Copyright 2022 The Fluid Authors.

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

package efc

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	utilpointer "k8s.io/utils/pointer"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	ctrlhelper "github.com/fluid-cloudnative/fluid/pkg/ctrl"
)

func TestCheckRuntimeHealthy(t *testing.T) {
	type fields struct {
		runtime         *datav1alpha1.EFCRuntime
		worker          *appsv1.StatefulSet
		master          *appsv1.StatefulSet
		fuse            *appsv1.DaemonSet
		workerEndPoints *v1.ConfigMap
		name            string
		namespace       string
	}
	tests := []struct {
		name    string
		wantErr bool
		fields  fields
	}{
		{
			name:    "healthy",
			wantErr: false,
			fields: fields{
				name:      "health-data",
				namespace: "big-data",
				runtime: &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "health-data",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.EFCRuntimeSpec{
						Replicas: 1,
						Fuse:     datav1alpha1.EFCFuseSpec{},
					},
				},
				master: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "health-data-master",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      1,
						ReadyReplicas: 1,
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "health-data-worker",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: utilpointer.Int32(1),
						Selector: &metav1.LabelSelector{},
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      1,
						ReadyReplicas: 1,
					},
				},
				fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "health-data-fuse",
						Namespace: "big-data",
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 0,
					},
				},
				workerEndPoints: &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "health-data-worker-endpoints",
						Namespace: "big-data",
					},
					Data: map[string]string{
						WorkerEndpointsDataName: workerEndpointsConfigMapData,
					},
				},
			},
		},
		{
			name:    "master-no-healthy",
			wantErr: true,
			fields: fields{
				name:      "master-no-health-data",
				namespace: "big-data",
				runtime: &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "master-no-health-data",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.EFCRuntimeSpec{
						Replicas: 1,
						Fuse:     datav1alpha1.EFCFuseSpec{},
					},
				},
				master: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "master-no-health-data-master",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      1,
						ReadyReplicas: 0,
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "master-no-health-data-worker",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: utilpointer.Int32(1),
						Selector: &metav1.LabelSelector{},
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      1,
						ReadyReplicas: 1,
					},
				},
				fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "master-no-health-data-fuse",
						Namespace: "big-data",
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 0,
					},
				},
				workerEndPoints: &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "master-no-health-data-worker-endpoints",
						Namespace: "big-data",
					},
					Data: map[string]string{
						WorkerEndpointsDataName: workerEndpointsConfigMapData,
					},
				},
			},
		},
		{
			name:    "worker-no-healthy",
			wantErr: true,
			fields: fields{
				name:      "worker-no-health-data",
				namespace: "big-data",
				runtime: &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "worker-no-health-data",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.EFCRuntimeSpec{
						Replicas: 1,
						Fuse:     datav1alpha1.EFCFuseSpec{},
					},
				},
				master: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "worker-no-health-data-master",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      1,
						ReadyReplicas: 1,
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "worker-no-health-data-worker",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: utilpointer.Int32(2),
						Selector: &metav1.LabelSelector{},
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      2,
						ReadyReplicas: 0,
					},
				},
				fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "worker-no-health-data-fuse",
						Namespace: "big-data",
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 0,
					},
				},
				workerEndPoints: &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "worker-no-health-data-worker-endpoints",
						Namespace: "big-data",
					},
					Data: map[string]string{
						WorkerEndpointsDataName: workerEndpointsConfigMapData,
					},
				},
			},
		},
		{
			name:    "worker-partial-healthy",
			wantErr: false,
			fields: fields{
				name:      "worker-partial-health-data",
				namespace: "big-data",
				runtime: &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "worker-partial-health-data",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.EFCRuntimeSpec{
						Replicas: 1,
						Fuse:     datav1alpha1.EFCFuseSpec{},
					},
				},
				master: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "worker-partial-health-data-master",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      1,
						ReadyReplicas: 1,
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "worker-partial-health-data-worker",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: utilpointer.Int32(2),
						Selector: &metav1.LabelSelector{},
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      2,
						ReadyReplicas: 1,
					},
				},
				fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "worker-partial-health-data-fuse",
						Namespace: "big-data",
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 0,
					},
				},
				workerEndPoints: &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "worker-partial-health-data-worker-endpoints",
						Namespace: "big-data",
					},
					Data: map[string]string{
						WorkerEndpointsDataName: workerEndpointsConfigMapData,
					},
				},
			},
		},
		{
			name:    "fuse-no-healthy",
			wantErr: true,
			fields: fields{
				name:      "fuse-no-health-data",
				namespace: "big-data",
				runtime: &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "fuse-no-health-data",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.EFCRuntimeSpec{
						Replicas: 1,
						Fuse:     datav1alpha1.EFCFuseSpec{},
					},
				},
				master: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "fuse-no-health-data-master",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      1,
						ReadyReplicas: 1,
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "fuse-no-health-data-worker",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: utilpointer.Int32(1),
						Selector: &metav1.LabelSelector{},
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      1,
						ReadyReplicas: 1,
					},
				},
				fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "fuse-no-health-data-fuse",
						Namespace: "big-data",
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 1,
					},
				},
				workerEndPoints: &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "fuse-no-health-data-worker-endpoints",
						Namespace: "big-data",
					},
					Data: map[string]string{
						WorkerEndpointsDataName: workerEndpointsConfigMapData,
					},
				},
			},
		},
		{
			name:    "endpoints-no-healthy",
			wantErr: true,
			fields: fields{
				name:      "endpoints-no-health-data",
				namespace: "big-data",
				runtime: &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "endpoints-no-health-data",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.EFCRuntimeSpec{
						Replicas: 1,
						Fuse:     datav1alpha1.EFCFuseSpec{},
					},
				},
				master: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "endpoints-no-health-data-master",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      1,
						ReadyReplicas: 1,
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "endpoints-no-health-data-worker",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: utilpointer.Int32(1),
						Selector: &metav1.LabelSelector{},
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      1,
						ReadyReplicas: 1,
					},
				},
				fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "endpoints-no-health-data-fuse",
						Namespace: "big-data",
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 0,
					},
				},
				workerEndPoints: &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "123",
						Namespace: "big-data",
					},
					Data: map[string]string{
						WorkerEndpointsDataName: workerEndpointsConfigMapData,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtimeObjs := []runtime.Object{}
			data := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tt.fields.name,
					Namespace: tt.fields.namespace,
				},
			}

			s := runtime.NewScheme()
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.runtime)
			s.AddKnownTypes(datav1alpha1.GroupVersion, data)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, tt.fields.worker)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, tt.fields.fuse)
			s.AddKnownTypes(v1.SchemeGroupVersion, tt.fields.workerEndPoints)

			_ = v1.AddToScheme(s)

			runtimeObjs = append(runtimeObjs, tt.fields.runtime, data, tt.fields.worker, tt.fields.master, tt.fields.fuse, tt.fields.workerEndPoints)
			mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)
			e := &EFCEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
				Log:       ctrl.Log.WithName(tt.fields.name),
			}

			runtimeInfo, err := base.BuildRuntimeInfo(tt.fields.name, tt.fields.namespace, common.EFCRuntime, datav1alpha1.TieredStore{})
			if err != nil {
				t.Errorf("EFCEngine.CheckWorkersReady() error = %v", err)
			}

			e.Helper = ctrlhelper.BuildHelper(runtimeInfo, mockClient, e.Log)

			healthError := e.CheckRuntimeHealthy()
			hasErr := (healthError != nil)
			if tt.wantErr != hasErr {
				t.Errorf("testcase %s check runtime healthy ,hasErr %v, wantErr %v, err:%s", tt.name, hasErr, tt.wantErr, healthError)
			}

		})
	}

}
