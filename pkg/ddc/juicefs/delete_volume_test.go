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

package juicefs

import (
	"context"
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
)

type TestCase struct {
	engine    *JuiceFSEngine
	isDeleted bool
	isErr     bool
}

// newTestJuiceEngine creates a JuiceFSEngine for testing
// Parameters:
// 	 - client: fake client
//   - name: the name of the JuiceFS engine
//   - namespace: the namespace of the JuiceFS engine
//   - withRunTime: whether the JuiceFS engine has runtime
// Returns:
//   - JuiceFSEngine: the JuiceFS engine for testing
func newTestJuiceEngine(client client.Client, name string, namespace string, withRunTime bool) *JuiceFSEngine {
	// 1. Create a JuiceFSRuntime and RuntimeInfo 
	runTime := &datav1alpha1.JuiceFSRuntime{}
	runTimeInfo, _ := base.BuildRuntimeInfo(name, namespace, common.JuiceFSRuntime)
	// 2. If the JuiceFS engine does not have runtime, set the runtime and runtimeInfo to nil
	if !withRunTime {
		runTimeInfo = nil
		runTime = nil
	}
	// 3. Create a JuiceFSEngine
	engine := &JuiceFSEngine{
		runtime:     runTime,
		name:        name,
		namespace:   namespace,
		Client:      client,
		runtimeInfo: runTimeInfo,
		Log:         fake.NullLogger(),
	}
	// 4. Return the JuiceFSEngine
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

func TestJuiceFSEngine_DeleteVolume(t *testing.T) {
	testPVInputs := []*v1.PersistentVolume{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "fluid-juicefs",
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
				Name:       "juicefs",
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
	juicefsEngineCommon := newTestJuiceEngine(fakeClient, "juicefs", "fluid", true)
	juicefsEngineErr := newTestJuiceEngine(fakeClient, "error", "fluid", true)
	juicefsEngineNoRunTime := newTestJuiceEngine(fakeClient, "juicefs", "fluid", false)
	var testCases = []TestCase{
		{
			engine:    juicefsEngineCommon,
			isDeleted: true,
			isErr:     false,
		},
		{
			engine:    juicefsEngineErr,
			isDeleted: true,
			isErr:     true,
		},
		{
			engine:    juicefsEngineNoRunTime,
			isDeleted: true,
			isErr:     true,
		},
	}
	doTestCases(testCases, t)
}

func TestJuiceFSEngine_deleteFusePersistentVolume(t *testing.T) {
	testPVInputs := []*v1.PersistentVolume{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "fluid-juicefs",
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
	juicefsEngine := newTestJuiceEngine(fakeClient, "juicefs", "fluid", true)
	juicefsEngineNoRuntime := newTestJuiceEngine(fakeClient, "juicefs", "fluid", false)
	testCases := []TestCase{
		{
			engine:    juicefsEngine,
			isDeleted: true,
			isErr:     false,
		},
		{
			engine:    juicefsEngineNoRuntime,
			isDeleted: true,
			isErr:     true,
		},
	}
	doTestCases(testCases, t)
}

func TestJuiceFSEngine_deleteFusePersistentVolumeClaim(t *testing.T) {
	testPVCInputs := []*v1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "juicefs",
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
	juicefsEngine := newTestJuiceEngine(fakeClient, "hbase", "fluid", true)
	juicefsEngineNoRuntime := newTestJuiceEngine(fakeClient, "hbase", "fluid", false)
	testCases := []TestCase{
		{
			engine:    juicefsEngine,
			isDeleted: true,
			isErr:     false,
		},
		{
			engine:    juicefsEngineNoRuntime,
			isDeleted: true,
			isErr:     true,
		},
	}
	doTestCases(testCases, t)
}
