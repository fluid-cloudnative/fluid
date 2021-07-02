package alluxio

import (
	"context"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type TestCase struct {
	engine 			*AlluxioEngine
	isDeleted 		bool
	isErr 			bool
}



func newTestAlluxioEngine(client client.Client, name string, namespace string, withRunTime bool) *AlluxioEngine {
	runTime := &datav1alpha1.AlluxioRuntime{}
	runTimeInfo,_ := base.BuildRuntimeInfo(name,namespace,"alluxio",datav1alpha1.Tieredstore{})
	if !withRunTime{
		runTimeInfo = nil
		runTime = nil
	}
	engine := &AlluxioEngine{
		runtime:                runTime,
		name:                   name,
		namespace:              namespace,
		Client:                 client,
		runtimeInfo:            runTimeInfo,
		Log:                    log.NullLogger{},
	}
	return engine
}

func doTestCases(testCases []TestCase, t *testing.T){
	for _, test := range testCases {
		err := test.engine.DeleteVolume()
		pv := &v1.PersistentVolume{}
		nullPV := v1.PersistentVolume{}
		key := types.NamespacedName{
			Namespace: test.engine.namespace,
			Name: test.engine.name,
		}
		_ = test.engine.Client.Get(context.TODO(), key, pv)
		if test.isDeleted != reflect.DeepEqual(nullPV, *pv){
			t.Errorf("PV/PVC still exist after delete.")
		}
		isErr := err != nil
		if isErr != test.isErr{
			t.Errorf("expected %t, got %t.", test.isErr, isErr)
		}
	}
}

func TestAlluxioEngine_DeleteVolume(t *testing.T) {
	testPVInputs := []*v1.PersistentVolume{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "fluid-hbase",
				Annotations: map[string]string{
					"CreatedBy": "fluid",
				},
			},
			Spec: v1.PersistentVolumeSpec{},
		},
	}

	tests := []runtime.Object{}

	for _, pvInput := range testPVInputs {
		tests = append(tests, pvInput.DeepCopy())
	}

	testPVCInputs := []*v1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "hbase",
				Namespace:  "fluid",
				Finalizers: []string{"kubernetes.io/pvc-protection"}, // no err
			},
			Spec: v1.PersistentVolumeClaimSpec{},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "error",
				Namespace:  "fluid",
				Finalizers: []string{"kubernetes.io/pvc-protection"},
				Annotations: map[string]string{
					"CreatedBy": "fluid", // have err
				},
			},
			Spec: v1.PersistentVolumeClaimSpec{},
		},

	}

	for _, pvcInput := range testPVCInputs {
		tests = append(tests, pvcInput.DeepCopy())
	}

	fakeClient := fake.NewFakeClientWithScheme(testScheme, tests...)
	alluxioEngineCommon := newTestAlluxioEngine(fakeClient,"hbase","fluid",true)
	alluxioEngineErr := newTestAlluxioEngine(fakeClient,"error","fluid",true)
	alluxioEngineNoRunTime := newTestAlluxioEngine(fakeClient,"hbase","fluid",false)
	var testCases = []TestCase {
		{
			engine:    				alluxioEngineCommon,
			isDeleted: 				true,
			isErr: 					false,
		},
		{
			engine:    				alluxioEngineErr,
			isDeleted: 				true,
			isErr: 					true,
		},
		{
			engine:					alluxioEngineNoRunTime,
			isDeleted: 				true,
			isErr: 					true,
		},
	}
	doTestCases(testCases,t)
}

func TestAlluxioEngine_DeleteFusePersistentVolume(t *testing.T) {
	testPVInputs := []*v1.PersistentVolume{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "fluid-hbase",
				Annotations: map[string]string{
					"CreatedBy": "fluid",
				},
			},
			Spec: v1.PersistentVolumeSpec{},
		},
	}

	tests := []runtime.Object{}

	for _, pvInput := range testPVInputs {
		tests = append(tests, pvInput.DeepCopy())
	}

	fakeClient := fake.NewFakeClientWithScheme(testScheme, tests...)
	alluxioEngine := newTestAlluxioEngine(fakeClient, "hbase", "fluid",true)
	alluxioEngineNoRuntime := newTestAlluxioEngine(fakeClient, "hbase", "fluid",false)
	testCases := []TestCase{
		{
			engine: 			alluxioEngine,
			isDeleted: 			true,
			isErr: 				false,
		},
		{
			engine:   			alluxioEngineNoRuntime,
			isDeleted: 			true,
			isErr: 				true,
		},
	}
	doTestCases(testCases,t)
}

func TestAlluxioEngine_DeleteFusePersistentVolumeClaim(t *testing.T) {
	testPVCInputs := []*v1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "hbase",
				Namespace:  "fluid",
				Finalizers: []string{"kubernetes.io/pvc-protection"}, // no err
			},
			Spec: v1.PersistentVolumeClaimSpec{},
		},
	}

	tests := []runtime.Object{}

	for _, pvcInput := range testPVCInputs {
		tests = append(tests, pvcInput.DeepCopy())
	}

	fakeClient := fake.NewFakeClientWithScheme(testScheme, tests...)
	alluxioEngine := newTestAlluxioEngine(fakeClient, "hbase", "fluid",true)
	alluxioEngineNoRuntime := newTestAlluxioEngine(fakeClient, "hbase", "fluid",false)
	testCases := []TestCase{
		{
			engine: 			alluxioEngine,
			isDeleted: 			true,
			isErr: 				false,
		},
		{
			engine:   			alluxioEngineNoRuntime,
			isDeleted: 			true,
			isErr: 				true,
		},
	}
	doTestCases(testCases,t)
}



