package volume

import (
	"context"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestDeleteFusePersistentVolume(t *testing.T) {
	runtimeInfoHbase, err := base.BuildRuntimeInfo("hbase", "fluid", "alluxio", datav1alpha1.Tieredstore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	runtimeInfoHadoop, err := base.BuildRuntimeInfo("hadoop", "fluid", "alluxio", datav1alpha1.Tieredstore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	testPVInputs := []*v1.PersistentVolume{{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fluid-hbase",
			Annotations: map[string]string{
				"CreatedBy": "fluid",
			},
		},
		Spec: v1.PersistentVolumeSpec{},
	}}

	testPVs := []runtime.Object{}
	for _, pvInput := range testPVInputs {
		testPVs = append(testPVs, pvInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testPVs...)
	var testCase = []struct {
		runtimeInfo    base.RuntimeInfoInterface
		expectedResult v1.PersistentVolume
	}{
		{
			runtimeInfo: runtimeInfoHadoop,
			expectedResult: v1.PersistentVolume{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PersistentVolume",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "fluid-hbase",
					Annotations: map[string]string{
						"CreatedBy": "fluid",
					},
				},
				Spec: v1.PersistentVolumeSpec{},
			},
		},
		{
			runtimeInfo:    runtimeInfoHbase,
			expectedResult: v1.PersistentVolume{},
		},
	}
	for _, test := range testCase {
		var log = ctrl.Log.WithName("delete")
		_ = DeleteFusePersistentVolume(client, test.runtimeInfo, log)

		key := types.NamespacedName{
			Name: "fluid-hbase",
		}
		pv := &v1.PersistentVolume{}
		_ = client.Get(context.TODO(), key, pv)
		if !reflect.DeepEqual(test.expectedResult, *pv) {
			t.Errorf("fail to exec the function with the error")
		}
	}
}

func TestDeleteFusePersistentVolumeIfExists(t *testing.T) {
	testPVInputs := []*v1.PersistentVolume{{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hbase",
			Annotations: map[string]string{
				"CreatedBy": "fluid",
			},
		},
		Spec: v1.PersistentVolumeSpec{},
	}}

	testPVs := []runtime.Object{}
	for _, pvInput := range testPVInputs {
		testPVs = append(testPVs, pvInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testPVs...)
	var testCase = []struct {
		pvName         string
		expectedResult v1.PersistentVolume
	}{
		{
			pvName: "hadoop",
			expectedResult: v1.PersistentVolume{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PersistentVolume",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "hbase",
					Annotations: map[string]string{
						"CreatedBy": "fluid",
					},
				},
				Spec: v1.PersistentVolumeSpec{},
			},
		},
		{
			pvName:         "hbase",
			expectedResult: v1.PersistentVolume{},
		},
	}
	for _, test := range testCase {
		var log = ctrl.Log.WithName("delete")
		_ = deleteFusePersistentVolumeIfExists(client, test.pvName, log)

		key := types.NamespacedName{
			Name: "hbase",
		}
		pv := &v1.PersistentVolume{}
		_ = client.Get(context.TODO(), key, pv)
		if !reflect.DeepEqual(test.expectedResult, *pv) {
			t.Errorf("fail to exec the function with the error")
		}
	}
}

func TestDeleteFusePersistentVolumeClaim(t *testing.T) {
	runtimeInfoHbase, err := base.BuildRuntimeInfo("hbase", "fluid", "alluxio", datav1alpha1.Tieredstore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	runtimeInfoHadoop, err := base.BuildRuntimeInfo("hadoop", "fluid", "alluxio", datav1alpha1.Tieredstore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	testPVCInputs := []*v1.PersistentVolumeClaim{{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "hbase",
			Namespace:  "fluid",
			Finalizers: []string{"kubernetes.io/pvc-protection"},
		},
		Spec: v1.PersistentVolumeClaimSpec{},
	}}

	testPVCs := []runtime.Object{}
	for _, pvInput := range testPVCInputs {
		testPVCs = append(testPVCs, pvInput.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testPVCs...)

	var testCase = []struct {
		runtimeInfo    base.RuntimeInfoInterface
		expectedResult error
	}{
		{
			runtimeInfo:    runtimeInfoHadoop,
			expectedResult: nil,
		},
		{
			runtimeInfo:    runtimeInfoHbase,
			expectedResult: nil,
		},
	}
	for _, test := range testCase {
		var log = ctrl.Log.WithName("delete")
		if err := DeleteFusePersistentVolumeClaim(client, test.runtimeInfo, log); err != test.expectedResult {
			t.Errorf("fail to exec the function with the error %v", err)
		}
	}

}
