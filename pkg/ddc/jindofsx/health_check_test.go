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

package jindofsx

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/tools/record"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	ctrlhelper "github.com/fluid-cloudnative/fluid/pkg/ctrl"
)

func TestCheckRuntimeHealthy(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.JindoRuntime
		worker    *appsv1.StatefulSet
		master    *appsv1.StatefulSet
		fuse      *appsv1.DaemonSet
		name      string
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "healthy",
			fields: fields{
				name:      "health-data",
				namespace: "big-data",
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "health-data",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Replicas: 1,
						Fuse: datav1alpha1.JindoFuseSpec{
							Global: true,
						},
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "health-data-jindofs-worker",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      1,
						ReadyReplicas: 1,
					},
				},
				master: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "health-data-jindofs-master",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      1,
						ReadyReplicas: 1,
					},
				},
				fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "health-data-jindofs-fuse",
						Namespace: "big-data",
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 0,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "master-nohealthy",
			fields: fields{
				name:      "unhealthy-master",
				namespace: "big-data",
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unhealthy-master",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Replicas: 1,
						Fuse: datav1alpha1.JindoFuseSpec{
							Global: true,
						},
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unhealthy-master-jindofs-worker",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 0,
						Replicas:      1,
					},
				}, master: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unhealthy-master-jindofs-master",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 0,
					},
				}, fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unhealthy-master-jindofs-fuse",
						Namespace: "big-data",
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 1,
					},
				},
			},
			wantErr: true,
		}, {
			name: "worker-nohealthy",
			fields: fields{
				name:      "unhealthy-worker",
				namespace: "big-data",
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unhealthy-worker",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Replicas: 1,
						Fuse: datav1alpha1.JindoFuseSpec{
							Global: true,
						},
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unhealthy-worker-jindofs-worker",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 0,
						Replicas:      1,
					},
				}, master: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unhealthy-worker-jindofs-master",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 1,
						Replicas:      1,
					},
				}, fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unhealthy-worker-jindofs-fuse",
						Namespace: "big-data",
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 1,
					},
				},
			},
			wantErr: true,
		}, {
			name: "fuse-nohealthy",
			fields: fields{
				name:      "unhealthy-fuse",
				namespace: "big-data",
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unhealthy-fuse",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Replicas: 1,
						Fuse: datav1alpha1.JindoFuseSpec{
							Global: true,
						},
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unhealthy-fuse-jindofs-worker",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 1,
						Replicas:      1,
					},
				}, master: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unhealthy-fuse-jindofs-master",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 1,
						Replicas:      1,
					},
				}, fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unhealthy-fuse-jindofs-fuse",
						Namespace: "big-data",
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 1,
					},
				},
			},
			wantErr: true,
		}, {
			name: "no-master-nohealthy",
			fields: fields{
				name:      "unhealthy-no-master",
				namespace: "big-data",
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unhealthy-no-master",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Replicas: 1,
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unhealthy-no-master-jindofs-worker",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 1,
						Replicas:      1,
					},
				}, master: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unhealthy-no-master-jindofs",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 1,
						Replicas:      1,
					},
				}, fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unhealthy-no-master-jindofs-no-master",
						Namespace: "big-data",
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 1,
					},
				},
			},
			wantErr: true,
		}, {
			name: "no-worker-nohealthy",
			fields: fields{
				name:      "unhealthy-no-worker",
				namespace: "big-data",
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unhealthy-no-worker",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Replicas: 1,
					},
				},
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unhealthy-no-worker-jindofs-master",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 1,
						Replicas:      1,
					},
				}, master: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unhealthy-no-worker-jindofs",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 1,
						Replicas:      1,
					},
				}, fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unhealthy-no-worker-jindofs-no-worker",
						Namespace: "big-data",
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 1,
					},
				},
			},
			wantErr: true,
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

			_ = v1.AddToScheme(s)

			runtimeObjs = append(runtimeObjs, tt.fields.runtime, data, tt.fields.worker, tt.fields.master, tt.fields.fuse)
			mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)
			e := &JindoFSxEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
				Log:       ctrl.Log.WithName(tt.fields.name),
				Recorder:  record.NewFakeRecorder(300),
			}

			runtimeInfo, err := base.BuildRuntimeInfo(tt.fields.name, tt.fields.namespace, "jindo", datav1alpha1.TieredStore{})
			if err != nil {
				t.Errorf("JindoFSxEngine.CheckWorkersReady() error = %v", err)
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
