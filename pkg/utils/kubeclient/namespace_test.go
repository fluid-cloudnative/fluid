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

package kubeclient

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestIsNamespaceExist(t *testing.T) {

	testNamespaceInputs := []*v1.Namespace{{
		ObjectMeta: metav1.ObjectMeta{Name: "test1"},
		Spec:       v1.NamespaceSpec{},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "test2"},
		Spec:       v1.NamespaceSpec{},
	}}

	testNamespaces := []runtime.Object{}

	for _, ns := range testNamespaceInputs {
		testNamespaces = append(testNamespaces, ns.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testNamespaces...)

	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "namespace doesn't exist",
			args: args{
				name: "notExist",
			},
		},
		{
			name: "namespace exists",
			args: args{
				name: "test1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := EnsureNamespace(client, tt.args.name); err != nil {
				t.Errorf("testcase %v EnsureNamespace()'s err is %v", tt.name, err)
			}
		})
	}

}

func TestCreateNamespace(t *testing.T) {

	testNamespaceInputs := []*v1.Namespace{}

	testNamespaces := []runtime.Object{}

	for _, ns := range testNamespaceInputs {
		testNamespaces = append(testNamespaces, ns.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testNamespaces...)

	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "namespace doesn't exist",
			args: args{
				name: "notExist",
			},
		},
		{
			name: "namespace exists",
			args: args{
				name: "test1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createNamespace(client, tt.args.name); err != nil {
				t.Errorf("testcase %v createNamespace()'s err is %v", tt.name, err)
			}
		})
	}

}
