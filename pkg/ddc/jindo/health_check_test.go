package jindo

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	appsv1 "k8s.io/api/apps/v1"

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
		name      string
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				name:      "spark",
				namespace: "big-data",
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark",
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
						Name:      "spark-jindofs-worker",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "test1",
			fields: fields{
				name:      "hbase",
				namespace: "big-data",
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase",
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
						Name:      "hbase-jindofs-worker",
						Namespace: "big-data",
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 0,
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

			_ = v1.AddToScheme(s)

			runtimeObjs = append(runtimeObjs, tt.fields.runtime, data, tt.fields.worker)
			mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)
			e := &JindoEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
				Log:       ctrl.Log.WithName(tt.fields.name),
			}

			runtimeInfo, err := base.BuildRuntimeInfo(tt.fields.name, tt.fields.namespace, "jindo", datav1alpha1.TieredStore{})
			if err != nil {
				t.Errorf("JindoEngine.CheckWorkersReady() error = %v", err)
			}

			e.Helper = ctrlhelper.BuildHelper(runtimeInfo, mockClient, e.Log)

			err = e.CheckRuntimeHealthy()
			hasErr := (err != nil)
			if tt.wantErr != hasErr {
				t.Errorf("testcase %s check runtime healthy ,hasErr %v, wantErr %v, err:%s", tt.name, hasErr, tt.wantErr, err)
			}

		})
	}

}
