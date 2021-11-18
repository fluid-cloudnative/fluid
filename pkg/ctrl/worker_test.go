package ctrl

import (
	"testing"

	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilpointer "k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetWorkersAsStatefulset(t *testing.T) {

	statefulsetInputs := []*appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sts-jindofs-worker",
				Namespace: "big-data",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32Ptr(2),
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		},
	}

	daemonsetInputs := []*appsv1.DaemonSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ds-jindofs-worker",
				Namespace: "big-data",
			},
		},
	}

	objs := []runtime.Object{}

	for _, runtimeInput := range daemonsetInputs {
		objs = append(objs, runtimeInput.DeepCopy())
	}
	for _, statefulsetInput := range statefulsetInputs {
		objs = append(objs, statefulsetInput.DeepCopy())
	}

	s := runtime.NewScheme()
	_ = appsv1.AddToScheme(s)
	fakeClient := fake.NewFakeClientWithScheme(s, objs...)
	testCases := []struct {
		name            string
		key             types.NamespacedName
		success         bool
		deprecatedError bool
	}{
		{
			name: "noError",
			key: types.NamespacedName{
				Name:      "sts-jindofs-worker",
				Namespace: "big-data",
			},
			success:         true,
			deprecatedError: false,
		}, {
			name: "deprecatedError",
			key: types.NamespacedName{
				Name:      "ds-jindofs-worker",
				Namespace: "big-data",
			},
			success:         false,
			deprecatedError: true,
		}, {
			name: "otherError",
			key: types.NamespacedName{
				Name:      "test-jindofs-worker",
				Namespace: "big-data",
			},
			success:         false,
			deprecatedError: false,
		},
	}

	for _, testCase := range testCases {
		_, err := GetWorkersAsStatefulset(fakeClient, testCase.key)

		if testCase.success != (err == nil) {
			t.Errorf("testcase %s failed due to expect succcess %v, got error %v", testCase.name, testCase.success, err)
		}

		if !testCase.success {
			if testCase.deprecatedError != fluiderrs.IsDeprecated(err) {
				t.Errorf("testcase %s failed due to expect isdeprecated  %v, got  %v", testCase.name, testCase.deprecatedError, fluiderrs.IsDeprecated(err))
			}
		}
	}

}
