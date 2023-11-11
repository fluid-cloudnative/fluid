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

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Use fake client because of it will be maintained in the long term
// due to https://github.com/kubernetes-sigs/controller-runtime/pull/1101
func TestGetServiceByName(t *testing.T) {
	namespace := "default"
	testServiceInputs := []*v1.Service{{
		ObjectMeta: metav1.ObjectMeta{Name: "svc1"},
		Spec:       v1.ServiceSpec{},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "svc2", Annotations: common.ExpectedFluidAnnotations},
		Spec:       v1.ServiceSpec{},
	}}

	testServices := []runtime.Object{}

	for _, pv := range testServiceInputs {
		testServices = append(testServices, pv.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testServices...)

	type args struct {
		name      string
		namespace string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "service doesn't exist",
			args: args{
				name:      "notExist",
				namespace: namespace,
			},
		},
		{
			name: "service is not created by fluid",
			args: args{
				name:      "svc1",
				namespace: namespace,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := GetServiceByName(client, tt.args.name, tt.args.namespace); err != nil {
				t.Errorf("testcase %v GetServiceByName() got error %v", tt.name, err)
			}
		})
	}

}
