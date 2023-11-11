/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package thin

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
	engine    *ThinEngine
	isDeleted bool
	isErr     bool
}

func newTestThinEngine(client client.Client, name string, namespace string, withRunTime bool) *ThinEngine {
	runTime := &datav1alpha1.ThinRuntime{}
	runTimeInfo, _ := base.BuildRuntimeInfo(name, namespace, "thin", datav1alpha1.TieredStore{})
	if !withRunTime {
		runTimeInfo = nil
		runTime = nil
	}
	engine := &ThinEngine{
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

func TestThinEngine_DeleteVolume(t *testing.T) {
	testPVInputs := []*v1.PersistentVolume{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "fluid-thin",
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
				Name:       "thin",
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

	testRuntimeInputs := []*datav1alpha1.ThinRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "thin",
				Namespace: "fluid",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "error",
				Namespace: "fluid",
			},
		},
	}

	for _, runtimeInput := range testRuntimeInputs {
		tests = append(tests, runtimeInput)
	}

	fakeClient := fake.NewFakeClientWithScheme(testScheme, tests...)
	ThinEngineCommon := newTestThinEngine(fakeClient, "thin", "fluid", true)
	ThinEngineErr := newTestThinEngine(fakeClient, "error", "fluid", true)
	ThinEngineNoRunTime := newTestThinEngine(fakeClient, "thin-no-runtime", "fluid", false)
	var testCases = []TestCase{
		{
			engine:    ThinEngineCommon,
			isDeleted: true,
			isErr:     false,
		},
		{
			engine:    ThinEngineErr,
			isDeleted: true,
			isErr:     true,
		},
		{
			engine:    ThinEngineNoRunTime,
			isDeleted: true,
			isErr:     true,
		},
	}
	doTestCases(testCases, t)
}

func TestThinEngine_deleteFusePersistentVolume(t *testing.T) {
	testPVInputs := []*v1.PersistentVolume{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "fluid-thin",
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

	testRuntimeInputs := []*datav1alpha1.ThinRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "thin",
				Namespace: "fluid",
			},
		},
	}

	for _, runtimeInput := range testRuntimeInputs {
		tests = append(tests, runtimeInput)
	}

	fakeClient := fake.NewFakeClientWithScheme(testScheme, tests...)
	ThinEngine := newTestThinEngine(fakeClient, "thin", "fluid", true)
	ThinEngineNoRuntime := newTestThinEngine(fakeClient, "thin-no-runtime", "fluid", false)
	testCases := []TestCase{
		{
			engine:    ThinEngine,
			isDeleted: true,
			isErr:     false,
		},
		{
			engine:    ThinEngineNoRuntime,
			isDeleted: true,
			isErr:     true,
		},
	}
	doTestCases(testCases, t)
}

func TestThinEngine_deleteFusePersistentVolumeClaim(t *testing.T) {
	testPVCInputs := []*v1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "thin",
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

	testRuntimeInputs := []*datav1alpha1.ThinRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
		},
	}

	for _, runtimeInput := range testRuntimeInputs {
		tests = append(tests, runtimeInput)
	}

	fakeClient := fake.NewFakeClientWithScheme(testScheme, tests...)
	ThinEngine := newTestThinEngine(fakeClient, "hbase", "fluid", true)
	ThinEngineNoRuntime := newTestThinEngine(fakeClient, "hbase-no-runtime", "fluid", false)
	testCases := []TestCase{
		{
			engine:    ThinEngine,
			isDeleted: true,
			isErr:     false,
		},
		{
			engine:    ThinEngineNoRuntime,
			isDeleted: true,
			isErr:     true,
		},
	}
	doTestCases(testCases, t)
}
