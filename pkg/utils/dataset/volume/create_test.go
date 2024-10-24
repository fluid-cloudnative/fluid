package volume

import (
	"context"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestCreatePersistentVolumeForRuntime(t *testing.T) {
	// runtimeInfoExclusive is a runtimeInfo with ExclusiveMode with a PV already in use.
	runtimeInfoHbase, err := base.BuildRuntimeInfo("hbase", "fluid", "alluxio")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	// runtimeInfoExclusive is a runtimeInfo in global mode with no correspond PV.
	runtimeInfoSpark, err := base.BuildRuntimeInfo("spark", "fluid", "alluxio")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfoSpark.SetFuseNodeSelector(map[string]string{"test-node": "true"})

	// runtimeInfoShare is a runtimeInfo in non global mode with no correspond PV.
	runtimeInfoHadoop, err := base.BuildRuntimeInfo("hadoop", "fluid", "alluxio")
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
	testObjs := []runtime.Object{}
	for _, pvInput := range testPVInputs {
		testObjs = append(testObjs, pvInput.DeepCopy())
	}

	testDatasetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		},
	}
	for _, datasetInput := range testDatasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	var testCase = []struct {
		runtimeInfo   base.RuntimeInfoInterface
		mountPath     string
		mountType     string
		expectedPVNum int
	}{
		{
			runtimeInfo:   runtimeInfoHbase,
			mountPath:     "/runtimeInfoHbase",
			mountType:     "alluxio",
			expectedPVNum: 1,
		},
		{
			runtimeInfo:   runtimeInfoSpark,
			mountPath:     "/runtimeInfoSpark",
			mountType:     "alluxio",
			expectedPVNum: 2,
		},
		{
			runtimeInfo:   runtimeInfoHadoop,
			mountPath:     "/runtimeInfoHadoop",
			mountType:     "alluxio",
			expectedPVNum: 3,
		},
	}

	for _, test := range testCase {
		var log = ctrl.Log.WithName("delete")
		err := CreatePersistentVolumeForRuntime(client, test.runtimeInfo, test.mountPath, test.mountType, log)
		if err != nil {
			t.Errorf("fail to exec the function with error %v", err)
			return
		}
		var pvs v1.PersistentVolumeList
		err = client.List(context.TODO(), &pvs)
		if err != nil {
			t.Errorf("fail to exec the function with error %v", err)
			return
		}
		if len(pvs.Items) != test.expectedPVNum {
			t.Errorf("fail to create the pv")
		}
	}
}

func TestCreatePersistentVolumeClaimForRuntime(t *testing.T) {
	runtimeInfoHbase, err := base.BuildRuntimeInfo("hbase", "fluid", "alluxio")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	runtimeInfoSpark, err := base.BuildRuntimeInfo("spark", "fluid", "alluxio")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	testPVCInputs := []*v1.PersistentVolumeClaim{{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
			Annotations: map[string]string{
				"CreatedBy": "fluid",
			},
		},
		Spec: v1.PersistentVolumeClaimSpec{},
	}}
	testObjs := []runtime.Object{}
	for _, pvcInput := range testPVCInputs {
		testObjs = append(testObjs, pvcInput.DeepCopy())
	}

	testDatasetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		},
	}
	for _, datasetInput := range testDatasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
	var testCase = []struct {
		runtimeInfo    base.RuntimeInfoInterface
		expectedPVCNum int
	}{
		{
			runtimeInfo:    runtimeInfoHbase,
			expectedPVCNum: 1,
		},
		{
			runtimeInfo:    runtimeInfoSpark,
			expectedPVCNum: 2,
		},
	}

	for _, test := range testCase {
		var log = ctrl.Log.WithName("delete")
		err := CreatePersistentVolumeClaimForRuntime(client, test.runtimeInfo, log)
		if err != nil {
			t.Errorf("fail to exec the function with error %v", err)
			return
		}
		var pvs v1.PersistentVolumeClaimList
		err = client.List(context.TODO(), &pvs)
		if err != nil {
			t.Errorf("fail to exec the function with error %v", err)
			return
		}
		if len(pvs.Items) != test.expectedPVCNum {
			t.Errorf("fail to create the pv")
		}
	}
}
