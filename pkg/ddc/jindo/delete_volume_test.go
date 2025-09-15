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

// newTestJindoEngine creates a mock JindoEngine instance for testing purposes.
// This helper function is used to simulate JindoEngine behavior under different runtime conditions.
// It accepts a Kubernetes client, a runtime name and namespace, and a boolean flag indicating whether
// to initialize the engine with a runtime.
//
// Parameters:
//   - client: a fake or real Kubernetes client used by the engine to interact with cluster resources.
//   - name: the name of the JindoRuntime, used to build runtime metadata.
//   - namespace: the namespace of the JindoRuntime.
//   - withRunTime: if true, the engine will be initialized with a valid JindoRuntime and runtimeInfo;
//     if false, the runtime and runtimeInfo will be set to nil, simulating a missing runtime.
//
// Returns:
// - A pointer to the initialized JindoEngine instance, ready for use in unit tests.
func newTestJindoEngine(client client.Client, name string, namespace string, withRunTime bool) *JindoEngine {
	runTime := &datav1alpha1.JindoRuntime{}
	runTimeInfo, _ := base.BuildRuntimeInfo(name, namespace, common.JindoRuntime)
	runTimeInfo.SetOwnerDatasetUID("dummy-dataset-uid")
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

// TestJindoEngine_DeleteVolume tests the DeleteVolume method of JindoEngine under various scenarios.
//
// Parameters:
//   - t *testing.T : Go testing framework context for test assertions and logging
//
// Test Strategy:
// 1. Setup test environment with mocked Kubernetes resources:
//   - PVs with 'CreatedBy=fluid' annotation to simulate Fluid-managed persistent volumes
//   - PVCs with different configurations (normal vs error cases)
//     2. Create 3 test scenarios using parameterized test pattern:
//     Case 1: Normal deletion (JindoEngineCommon)
//   - Input: Valid PV/PVC without conflicting annotations
//   - Expected: Successful deletion (isDeleted=true, isErr=false)
//     Case 2: Protected PVC deletion (JindoEngineErr)
//   - Input: PVC with 'CreatedBy=fluid' annotation (protected resource)
//   - Expected: Failed deletion (isErr=true)
//     Case 3: Missing runtime scenario (JindoEngineNoRunTime)
//   - Input: Engine with runtime=false configuration
//   - Expected: Failed deletion (isErr=true)
//     3. Verification: Uses doTestCases() helper to validate deletion outcomes against expectations
//
// Test Resources:
// - fake.NewFakeClientWithScheme: Simulates Kubernetes API server with predefined resources
// - testScheme: Runtime scheme for Kubernetes API objects
// - testPVInputs/testPVCInputs: Predefined PersistentVolume and PersistentVolumeClaim configurations
//
// Note: This test focuses on edge cases for PVC/PV cleanup workflow in Fluid orchestration system.
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

// TestJindoEngine_DeleteFusePersistentVolume tests the deletion of a Fuse Persistent Volume in the JindoEngine.
// It creates a fake client with a predefined Persistent Volume and initializes two instances of JindoEngine,
// one with runtime enabled and one without. It then runs test cases to verify the deletion behavior of the
// Persistent Volume in both scenarios.
//
// Test cases:
// - JindoEngine with runtime enabled: expects the Persistent Volume to be deleted without errors.
// - JindoEngine without runtime enabled: expects the Persistent Volume to be deleted with an error.
//
// The function uses the doTestCases helper to execute the test cases and validate the results.
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
