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

package juicefs

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
)

func getTestJuiceFSEngine(client client.Client, name string, namespace string) *JuiceFSEngine {
	runTime := &datav1alpha1.JuiceFSRuntime{}
	runTimeInfo, _ := base.BuildRuntimeInfo(name, namespace, "juicefs", datav1alpha1.TieredStore{})
	engine := &JuiceFSEngine{
		runtime:     runTime,
		name:        name,
		namespace:   namespace,
		Client:      client,
		runtimeInfo: runTimeInfo,
		Log:         fake.NullLogger(),
	}
	return engine
}

func TestJuiceFSEngine_HasDeprecatedCommonLabelName(t *testing.T) {
	daemonSetWithSelector := &v1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fuse1-fuse",
			Namespace: "fluid",
		},
		Spec: v1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{NodeSelector: map[string]string{"data.fluid.io/storage-fluid-fuse1": "selector"}},
			},
		},
	}
	daemonSetWithoutSelector := &v1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fuse2-fuse",
			Namespace: "fluid",
		},
		Spec: v1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{NodeSelector: map[string]string{"data.fluid.io/storage-fluid-fuse1": "selector"}},
			},
		},
	}
	runtimeObjs := []runtime.Object{}
	runtimeObjs = append(runtimeObjs, daemonSetWithSelector)
	runtimeObjs = append(runtimeObjs, daemonSetWithoutSelector)
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, daemonSetWithSelector)
	fakeClient := fake.NewFakeClientWithScheme(scheme, runtimeObjs...)

	testCases := []struct {
		name      string
		namespace string
		out       bool
		isErr     bool
	}{
		{
			name:      "fuse1",
			namespace: "fluid",
			out:       true,
			isErr:     false,
		},
		{
			name:      "none",
			namespace: "fluid",
			out:       false,
			isErr:     false,
		},
		{
			name:      "fuse2",
			namespace: "fluid",
			out:       false,
			isErr:     false,
		},
	}

	for _, test := range testCases {
		engine := getTestJuiceFSEngine(fakeClient, test.name, test.namespace)
		out, err := engine.HasDeprecatedCommonLabelName()
		if out != test.out {
			t.Errorf("input parameter is %s-%s,expected %t, got %t", test.namespace, test.name, test.out, out)
		}
		isErr := err != nil
		if isErr != test.isErr {
			t.Errorf("input parameter is %s-%s,expected %t, got %t", test.namespace, test.name, test.isErr, isErr)
		}
	}
}

func TestJuiceFSEngine_getDeprecatedCommonLabelName(t *testing.T) {
	testCases := []struct {
		name      string
		namespace string
		out       string
	}{
		{
			name:      "fuse1",
			namespace: "fluid",
			out:       "data.fluid.io/storage-fluid-fuse1",
		},
		{
			name:      "fuse2",
			namespace: "fluid",
			out:       "data.fluid.io/storage-fluid-fuse2",
		},
		{
			name:      "fluid",
			namespace: "test",
			out:       "data.fluid.io/storage-test-fluid",
		},
	}
	fakeClient := fake.NewFakeClientWithScheme(testScheme)
	for _, test := range testCases {
		engine := getTestJuiceFSEngine(fakeClient, test.name, test.namespace)
		out := engine.getDeprecatedCommonLabelName()
		if out != test.out {
			t.Errorf("input parameter is %s-%s,expected %s, got %s", test.namespace, test.name, test.out, out)
		}
	}
}
