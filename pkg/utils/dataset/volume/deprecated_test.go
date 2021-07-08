package volume

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

var (
	testScheme *runtime.Scheme
)

func init() {
	testScheme = runtime.NewScheme()
	_ = v1.AddToScheme(testScheme)
	_ = datav1alpha1.AddToScheme(testScheme)
}

func TestHasDeprecatedPersistentVolumeName(t *testing.T) {
	testPVInputs := []*v1.PersistentVolume{{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hbase",
			Annotations: map[string]string{
				"CreatedBy": "fluid",
			},
		},
		Spec: v1.PersistentVolumeSpec{},
	}}

	runtimeInfoSpark, err := base.BuildRuntimeInfo("spark", "fluid", "alluxio", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	runtimeInfoHbase, err := base.BuildRuntimeInfo("hbase", "fluid", "alluxio", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	testPVs := []runtime.Object{}
	for _, pvInput := range testPVInputs {
		testPVs = append(testPVs, pvInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testPVs...)

	var testCase = []struct {
		runtimeInfo    base.RuntimeInfoInterface
		expectedResult bool
	}{
		{
			runtimeInfo:    runtimeInfoSpark,
			expectedResult: false,
		},
		{
			runtimeInfo:    runtimeInfoHbase,
			expectedResult: true,
		},
	}
	for _, test := range testCase {
		var log = ctrl.Log.WithName("deprecated")
		if result, _ := HasDeprecatedPersistentVolumeName(client, test.runtimeInfo, log); result != test.expectedResult {
			t.Errorf("fail to exec the function with the error %v", err)
		}
	}
}
