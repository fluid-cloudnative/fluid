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

package goosefs

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
	engine    *GooseFSEngine
	isDeleted bool
	isErr     bool
}

func newTestGooseFSEngine(client client.Client, name string, namespace string, withRunTime bool) *GooseFSEngine {
	runTime := &datav1alpha1.GooseFSRuntime{}
	runTimeInfo, _ := base.BuildRuntimeInfo(name, namespace, "goosefs", datav1alpha1.TieredStore{})
	if !withRunTime {
		runTimeInfo = nil
		runTime = nil
	}
	engine := &GooseFSEngine{
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

func TestGooseFSEngine_DeleteVolume(t *testing.T) {
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
	goosefsEngineCommon := newTestGooseFSEngine(fakeClient, "hbase", "fluid", true)
	goosefsEngineErr := newTestGooseFSEngine(fakeClient, "error", "fluid", true)
	goosefsEngineNoRunTime := newTestGooseFSEngine(fakeClient, "hbase", "fluid", false)
	var testCases = []TestCase{
		{
			engine:    goosefsEngineCommon,
			isDeleted: true,
			isErr:     false,
		},
		{
			engine:    goosefsEngineErr,
			isDeleted: true,
			isErr:     true,
		},
		{
			engine:    goosefsEngineNoRunTime,
			isDeleted: true,
			isErr:     true,
		},
	}
	doTestCases(testCases, t)
}

func TestGooseFSEngine_DeleteFusePersistentVolume(t *testing.T) {
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
	goosefsEngine := newTestGooseFSEngine(fakeClient, "hbase", "fluid", true)
	goosefsEngineNoRuntime := newTestGooseFSEngine(fakeClient, "hbase", "fluid", false)
	testCases := []TestCase{
		{
			engine:    goosefsEngine,
			isDeleted: true,
			isErr:     false,
		},
		{
			engine:    goosefsEngineNoRuntime,
			isDeleted: true,
			isErr:     true,
		},
	}
	doTestCases(testCases, t)
}

func TestGooseFSEngine_DeleteFusePersistentVolumeClaim(t *testing.T) {
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
	goosefsEngine := newTestGooseFSEngine(fakeClient, "hbase", "fluid", true)
	goosefsEngineNoRuntime := newTestGooseFSEngine(fakeClient, "hbase", "fluid", false)
	testCases := []TestCase{
		{
			engine:    goosefsEngine,
			isDeleted: true,
			isErr:     false,
		},
		{
			engine:    goosefsEngineNoRuntime,
			isDeleted: true,
			isErr:     true,
		},
	}
	doTestCases(testCases, t)
}
