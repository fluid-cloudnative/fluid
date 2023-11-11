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
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGetDaemonset(t *testing.T) {
	name := "test"
	namespace := "default"
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DaemonSetSpec{},
		Status: appsv1.DaemonSetStatus{
			NumberUnavailable: 1,
			NumberReady:       1,
		},
	}

	objs := []runtime.Object{}
	objs = append(objs, ds.DeepCopy())
	s := runtime.NewScheme()
	_ = appsv1.AddToScheme(s)
	fakeClient := fake.NewFakeClientWithScheme(s, objs...)

	_, err := GetDaemonset(fakeClient, name, namespace)
	if err != nil {
		t.Errorf("failed to call GetDaemonset due to %v with name %s and namespace %s",
			err,
			name,
			namespace)
	}

	_, err = GetDaemonset(fakeClient, "notFound", namespace)
	if err == nil {
		t.Errorf("failed to call GetDaemonset due to %v with name %s and namespace %s",
			err,
			name,
			namespace)
	}

}
