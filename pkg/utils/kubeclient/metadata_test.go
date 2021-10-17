package kubeclient

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCompareOwnerRefMatcheWithExpected(t *testing.T) {

	type fields struct {
		controller *appsv1.StatefulSet
		child      runtime.Object
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{name: "No controller",
			fields: fields{
				controller: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{},
				},
				child: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1-0",
						Namespace: "big-data",
					},
					Spec: corev1.PodSpec{},
				},
			},
		}, {name: "the controller uid is not matched",
			fields: fields{
				controller: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{},
				},
				child: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2-0",
						Namespace: "big-data",
					},
					Spec: corev1.PodSpec{},
				},
			},
			want: false,
		},
		{name: "Is Controller",
			fields: fields{
				controller: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{},
				},
				child: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2-0",
						Namespace: "big-data",
					},
					Spec: corev1.PodSpec{},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := runtime.NewScheme()
			s.AddKnownTypes(appsv1.SchemeGroupVersion, tt.fields.controller)
			_ = v1.AddToScheme(s)
			runtimeObjs := []runtime.Object{}
			runtimeObjs = append(runtimeObjs, tt.fields.controller)
			runtimeObjs = append(runtimeObjs, tt.fields.child)
			mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)

			metaObj, err := meta.Accessor(tt.fields.child)
			if err != nil {
				t.Errorf(" meta.Accessor = %v", err)
			}
			controllerRef := metav1.GetControllerOf(metaObj)
			want, err := compareOwnerRefMatcheWithExpected(mockClient, controllerRef, metaObj.GetNamespace(), tt.fields.controller)
			if err != nil {
				t.Errorf("compareOwnerRefMatcheWithExpected = %v", err)
			}

			if want != tt.want {
				t.Errorf("compareOwnerRefMatcheWithExpected() = %v, want %v", want, tt.want)
			}
		})
	}
}
