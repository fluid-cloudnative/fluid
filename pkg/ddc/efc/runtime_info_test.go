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

package efc

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
)

func newEFCEngineRT(client client.Client, name string, namespace string, withRuntimeInfo bool, unittest bool) *EFCEngine {
	runTimeInfo, _ := base.BuildRuntimeInfo(name, namespace, common.EFCRuntime, datav1alpha1.TieredStore{})
	engine := &EFCEngine{
		runtime:     nil,
		name:        name,
		namespace:   namespace,
		Client:      client,
		runtimeInfo: nil,
		UnitTest:    unittest,
		Log:         fake.NullLogger(),
	}

	if withRuntimeInfo {
		engine.runtimeInfo = runTimeInfo
	}
	return engine
}

func TestEFCEngine_getRuntimeInfo(t *testing.T) {
	runtimeInputs := []*datav1alpha1.EFCRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "runtime1",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.EFCRuntimeSpec{
				Fuse: datav1alpha1.EFCFuseSpec{
					CleanPolicy: datav1alpha1.OnDemandCleanPolicy,
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "runtime2",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.EFCRuntimeSpec{
				Fuse: datav1alpha1.EFCFuseSpec{
					CleanPolicy: datav1alpha1.OnDemandCleanPolicy,
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "runtime3",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.EFCRuntimeSpec{
				Fuse: datav1alpha1.EFCFuseSpec{
					CleanPolicy: datav1alpha1.OnDemandCleanPolicy,
				},
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{
						{
							Path:      "/mnt/cache1,/mnt/cache2",
							QuotaList: "100ST,50Gi",
						},
					},
				},
			},
		},
	}
	daemonSetInputs := []*v1.DaemonSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "runtime1-fuse",
				Namespace: "fluid",
			},
			Spec: v1.DaemonSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{NodeSelector: map[string]string{"data.fluid.io/storage-fluid-runtime1": "selector"}},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "runtime2-fuse",
				Namespace: "fluid",
			},
			Spec: v1.DaemonSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{NodeSelector: map[string]string{"data.fluid.io/storage-fluid-runtime2": "selector"}},
				},
			},
		},
	}
	dataSetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "runtime",
				Namespace: "fluid",
			},
		},
	}
	objs := []runtime.Object{}
	for _, runtimeInput := range runtimeInputs {
		objs = append(objs, runtimeInput.DeepCopy())
	}
	for _, daemonSetInput := range daemonSetInputs {
		objs = append(objs, daemonSetInput.DeepCopy())
	}
	for _, dataSetInput := range dataSetInputs {
		objs = append(objs, dataSetInput.DeepCopy())
	}
	//scheme := runtime.NewScheme()
	//scheme.AddKnownTypes(v1.SchemeGroupVersion, daemonSetWithSelector)
	//scheme.AddKnownTypes(v1alpha1.GroupVersion,runtimeInput)
	fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)

	testCases := []struct {
		name            string
		namespace       string
		withRuntimeInfo bool
		unittest        bool
		isErr           bool
		isNil           bool
	}{
		{
			name:            "runtime1",
			namespace:       "fluid",
			withRuntimeInfo: false,
			unittest:        false,
			isErr:           false,
			isNil:           false,
		},
		{
			name:            "runtime2",
			namespace:       "fluid",
			withRuntimeInfo: false,
			unittest:        true,
			isErr:           false,
			isNil:           false,
		},
		{
			name:            "runtime1",
			namespace:       "fluid",
			withRuntimeInfo: true,
			unittest:        false,
			isErr:           false,
			isNil:           false,
		},
		{
			name:            "runtime2",
			namespace:       "fluid",
			withRuntimeInfo: false,
			unittest:        false,
			isErr:           false,
			isNil:           false,
		},
		{
			name:            "runtime3",
			namespace:       "fluid",
			withRuntimeInfo: false,
			unittest:        false,
			isErr:           true,
			isNil:           true,
		},
		{
			name:            "runtime4",
			namespace:       "fluid",
			withRuntimeInfo: false,
			unittest:        false,
			isErr:           true,
			isNil:           true,
		},
	}
	for _, testCase := range testCases {
		engine := newEFCEngineRT(fakeClient, testCase.name, testCase.namespace, testCase.withRuntimeInfo, testCase.unittest)
		runtimeInfo, err := engine.getRuntimeInfo()
		isNil := runtimeInfo == nil
		isErr := err != nil
		if isNil != testCase.isNil {
			t.Errorf(" want %t, got %t", testCase.isNil, isNil)
		}
		if isErr != testCase.isErr {
			t.Errorf(" want %t, got %t", testCase.isErr, isErr)
		}
	}
}
