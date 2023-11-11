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

package volume

import (
	"context"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGetNamespacedNameByVolumeId(t *testing.T) {
	testPVs := []runtime.Object{}
	client := fake.NewFakeClientWithScheme(testScheme, testPVs...)
	testCase := []struct {
		volumeId        string
		pv              *v1.PersistentVolume
		pvc             *v1.PersistentVolumeClaim
		expectName      string
		expectNamespace string
		wantErr         bool
	}{
		{
			volumeId: "default-test",
			pv: &v1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "default-test"},
				Spec: v1.PersistentVolumeSpec{ClaimRef: &v1.ObjectReference{
					Namespace: "default",
					Name:      "test",
				}},
			},
			pvc: &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
					Labels: map[string]string{
						common.LabelAnnotationStorageCapacityPrefix + "default-test": "",
					},
				},
			},
			expectName:      "test",
			expectNamespace: "default",
			wantErr:         false,
		},
	}
	for _, test := range testCase {
		_ = client.Create(context.TODO(), test.pv)
		_ = client.Create(context.TODO(), test.pvc)
		namespace, name, err := GetNamespacedNameByVolumeId(client, test.volumeId)
		if err != nil {
			t.Errorf("failed to call GetNamespacedNameByVolumeId due to %v", err)
		}

		if name != test.expectName && namespace != test.expectNamespace {
			t.Errorf("testcase %s Expect name %s, but got %s, Expect namespace %s, but got %s",
				test.volumeId, test.expectName, name, test.expectNamespace, namespace)
		}
	}
}
