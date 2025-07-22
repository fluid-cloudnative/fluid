/*
Copyright 2021 The Fluid Authors.

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

package alluxio

import (
	"context"
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TestCase struct {
	engine    *AlluxioEngine
	isDeleted bool
	isErr     bool
}

// newTestAlluxioEngine creates an instance of AlluxioEngine for testing purposes.
//
// Parameters:
//
//	client     - A Kubernetes client used for interactions with the API server.
//	name       - The name for the AlluxioEngine (used for associated PV/PVC resources in tests).
//	namespace  - The namespace where the resources reside.
//	withRunTime - A flag indicating whether the engine should be initialized with runtime information.
//	              When true, an AlluxioRuntime instance and corresponding runtimeInfo are created;
//	              when false, both runtime and runtimeInfo are set to nil to simulate a scenario without runtime support.
//
// Returns:
//
//	A pointer to an AlluxioEngine instance configured according to the provided parameters.
func newTestAlluxioEngine(client client.Client, name string, namespace string, withRunTime bool) *AlluxioEngine {
	runTime := &datav1alpha1.AlluxioRuntime{}
	runTimeInfo, _ := base.BuildRuntimeInfo(name, namespace, "alluxio")
	runTimeInfo.SetOwnerDatasetUID("dummy-dataset-uid")
	if !withRunTime {
		runTimeInfo = nil
		runTime = nil
	}
	engine := &AlluxioEngine{
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

// TestAlluxioEngine_DeleteVolume tests the DeleteVolume function of the AlluxioEngine.
// It sets up test cases with different PersistentVolume (PV) and PersistentVolumeClaim (PVC) inputs,
// including scenarios with and without errors. The function uses a fake Kubernetes client to simulate
// the behavior of the AlluxioEngine when deleting volumes. The test cases include:
// 1. A common scenario where the volume should be deleted without errors.
// 2. A scenario where an error is expected due to specific annotations on the PVC.
// 3. A scenario where an error is expected because the AlluxioEngine is not running.
// The function then runs these test cases using the doTestCases helper function to verify the expected outcomes.
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
	alluxioEngineCommon := newTestAlluxioEngine(fakeClient, "hbase", "fluid", true)
	alluxioEngineErr := newTestAlluxioEngine(fakeClient, "error", "fluid", true)
	alluxioEngineNoRunTime := newTestAlluxioEngine(fakeClient, "hbase", "fluid", false)
	var testCases = []TestCase{
		{
			engine:    alluxioEngineCommon,
			isDeleted: true,
			isErr:     false,
		},
		{
			engine:    alluxioEngineErr,
			isDeleted: true,
			isErr:     true,
		},
		{
			engine:    alluxioEngineNoRunTime,
			isDeleted: true,
			isErr:     true,
		},
	}
	doTestCases(testCases, t)
}

// TestAlluxioEngine_DeleteFusePersistentVolume tests the functionality of deleting Fuse PersistentVolume in the AlluxioEngine.
// This function is mainly responsible for:
// - Setting up test cases with different configurations of PersistentVolume.
// - Creating a fake client to simulate interactions with the Kubernetes API for testing.
// - Initializing different AlluxioEngine instances with and without runtime settings.
// - Executing test cases and verifying whether the deletion operations of PersistentVolume succeed as expected.

// Parameters:
// - t (*testing.T): The testing framework's testing object, used to report test results and handle test failures.

// Returns:
// - None. The function reports test failures directly through the *testing.T object passed in.

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
	alluxioEngine := newTestAlluxioEngine(fakeClient, "hbase", "fluid", true)
	alluxioEngineNoRuntime := newTestAlluxioEngine(fakeClient, "hbase", "fluid", false)
	testCases := []TestCase{
		{
			engine:    alluxioEngine,
			isDeleted: true,
			isErr:     false,
		},
		{
			engine:    alluxioEngineNoRuntime,
			isDeleted: true,
			isErr:     true,
		},
	}
	doTestCases(testCases, t)
}

// TestAlluxioEngine_DeleteFusePersistentVolumeClaim tests the functionality of deleting Fuse PersistentVolumeClaim in the AlluxioEngine.
// This function is mainly responsible for:
// - Setting up multiple test cases with different configurations of PersistentVolumeClaim.
// - Creating a fake client to simulate interactions with Kubernetes API for testing.
// - Initializing different AlluxioEngine instances with various runtime settings.
// - Executing test cases and verifying whether the deletion operations of PersistentVolumeClaim succeed as expected.

// Parameters:
// - t (*testing.T): The testing framework's testing object, used to report test results and handle test failures.

// Returns:
// - None. The function reports test failures directly through the *testing.T object passed in.

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
	alluxioEngine := newTestAlluxioEngine(fakeClient, "hbase", "fluid", true)
	alluxioEngineNoRuntime := newTestAlluxioEngine(fakeClient, "hbase", "fluid", false)
	testCases := []TestCase{
		{
			engine:    alluxioEngine,
			isDeleted: true,
			isErr:     false,
		},
		{
			engine:    alluxioEngineNoRuntime,
			isDeleted: true,
			isErr:     true,
		},
	}
	doTestCases(testCases, t)
}
