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

package jindo

import (
	"context"
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TestCase struct {
	engine    *JindoEngine
	isDeleted bool
	isErr     bool
}

func newTestJindoEngine(client client.Client, name string, namespace string, withRunTime bool) *JindoEngine {
	runTime := &datav1alpha1.JindoRuntime{}
	runTimeInfo, _ := base.BuildRuntimeInfo(name, namespace, common.JindoRuntime)
	if !withRunTime {
		runTimeInfo = nil
		runTime = nil
	}
	engine := &JindoEngine{
		runtime:     runTime,
		name:        name,
		namespace:   namespace,
		Client:      client,
		runtimeInfo: runTimeInfo,
		Log:         fake.NullLogger(),
	}
	return engine
}

func doTestCases(testCases []TestCase, t *testing.T) {
	for _, test := range testCases {
		err := test.engine.DeleteVolume()
		pv := &v1.PersistentVolume{}
		nullPV := v1.PersistentVolume{}
		key := types.NamespacedName{
			Namespace: test.engine.namespace,
			Name:      test.engine.name,
		}
		_ = test.engine.Client.Get(context.TODO(), key, pv)
		if test.isDeleted != reflect.DeepEqual(nullPV, *pv) {
			t.Errorf("PV/PVC still exist after delete.")
		}
		isErr := err != nil
		if isErr != test.isErr {
			t.Errorf("expected %t, got %t.", test.isErr, isErr)
		}
	}
}

func TestJindoEngine_DeleteVolume(t *testing.T) {
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
	JindoEngineCommon := newTestJindoEngine(fakeClient, "hbase", "fluid", true)
	JindoEngineErr := newTestJindoEngine(fakeClient, "error", "fluid", true)
	JindoEngineNoRunTime := newTestJindoEngine(fakeClient, "hbase", "fluid", false)
	var testCases = []TestCase{
		{
			engine:    JindoEngineCommon,
			isDeleted: true,
			isErr:     false,
		},
		{
			engine:    JindoEngineErr,
			isDeleted: true,
			isErr:     true,
		},
		{
			engine:    JindoEngineNoRunTime,
			isDeleted: true,
			isErr:     true,
		},
	}
	doTestCases(testCases, t)
}

func TestJindoEngine_DeleteFusePersistentVolume(t *testing.T) {
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
	JindoEngine := newTestJindoEngine(fakeClient, "hbase", "fluid", true)
	JindoEngineNoRuntime := newTestJindoEngine(fakeClient, "hbase", "fluid", false)
	testCases := []TestCase{
		{
			engine:    JindoEngine,
			isDeleted: true,
			isErr:     false,
		},
		{
			engine:    JindoEngineNoRuntime,
			isDeleted: true,
			isErr:     true,
		},
	}
	doTestCases(testCases, t)
}

// TestJindoEngine_DeleteFusePersistentVolumeClaim tests the DeleteFusePersistentVolumeClaim method of JindoEngine.  
// This test verifies the behavior when deleting a PersistentVolumeClaim (PVC) under two scenarios:  
// 1. With a functional JindoEngine instance, expecting successful deletion of the PVC.  
// 2. With a JindoEngine instance that lacks a runtime, expecting an error upon attempting to delete the PVC.  
// The test initializes a fake Kubernetes client and the corresponding PVC inputs to execute these scenarios.  
func TestJindoEngine_DeleteFusePersistentVolumeClaim(t *testing.T) {  
	// Define test PVC inputs with metadata and finalizers  
	testPVCInputs := []*v1.PersistentVolumeClaim{  
		{  
			ObjectMeta: metav1.ObjectMeta{  
				Name:       "hbase",  
				Namespace:  "fluid",  
				Finalizers: []string{"kubernetes.io/pvc-protection"}, // This finalizer prevents deletion errors  
			},  
			Spec: v1.PersistentVolumeClaimSpec{},  
		},  
	}  

	// Initialize an empty slice to hold runtime objects for testing  
	tests := []runtime.Object{}  

	// Deep copy each PVC input and append to tests for use in the fake client  
	for _, pvcInput := range testPVCInputs {  
		tests = append(tests, pvcInput.DeepCopy())  
	}  

	// Create a fake Kubernetes client using the defined test scheme and PVCs  
	fakeClient := fake.NewFakeClientWithScheme(testScheme, tests...)  

	// Initialize JindoEngine instances for testing with and without runtime  
	JindoEngine := newTestJindoEngine(fakeClient, "hbase", "fluid", true)  
	JindoEngineNoRuntime := newTestJindoEngine(fakeClient, "hbase", "fluid", false)  

	// Define test cases with expected outcomes for both JindoEngine instances  
	testCases := []TestCase{  
		{  
			engine:    JindoEngine,  
			isDeleted: true,  // Expect deletion to succeed  
			isErr:     false, // Expect no error  
		},  
		{  
			engine:    JindoEngineNoRuntime,  
			isDeleted: true,  // Expect deletion to be attempted  
			isErr:     true,  // Expect an error due to lack of runtime  
		},  
	}  

	// Execute the defined test cases, passing in the testing object  
	doTestCases(testCases, t)  
}  
