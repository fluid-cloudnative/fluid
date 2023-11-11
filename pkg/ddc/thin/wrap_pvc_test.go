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
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestThinEngine_wrapMountedPersistentVolumeClaim(t *testing.T) {
	testObjs := []runtime.Object{}
	testDatasetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dataset1",
				Namespace: "default",
			},
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{
					{
						Name:       "native-pvc",
						MountPoint: "pvc://my-pvc-1",
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dataset2",
				Namespace: "default",
			},
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{
					{
						Name:       "native-pvc",
						MountPoint: "pvc://my-pvc-2",
					},
				},
			},
		},
	}
	for _, datasetInput := range testDatasetInputs {
		testObjs = append(testObjs, datasetInput)
	}

	testPVCInputs := []*corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-pvc-1",
				Namespace: "default",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-pvc-2",
				Namespace: "default",
				Labels: map[string]string{
					common.LabelAnnotationManagedBy: "dataset2",
				},
			},
		},
	}
	for _, pvcInput := range testPVCInputs {
		testObjs = append(testObjs, pvcInput)
	}

	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	type fields struct {
		name      string
		namespace string
	}

	tests := []struct {
		name        string
		fields      fields
		wantErr     bool
		wantPvcName string
	}{
		{
			name: "wrap_native_pvc",
			fields: fields{
				name:      "dataset1",
				namespace: "default",
			},
			wantErr:     false,
			wantPvcName: "my-pvc-1",
		},
		{
			name: "wrap_native_pvc_with_existed_label",
			fields: fields{
				name:      "dataset2",
				namespace: "default",
			},
			wantErr:     false,
			wantPvcName: "my-pvc-2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &ThinEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    client,
				Log:       fake.NullLogger(),
			}

			if err := engine.wrapMountedPersistentVolumeClaim(); (err != nil) != tt.wantErr {
				t.Errorf("ThinEngine.wrapMountedPersistentVolumeClaim() error = %v, wantErr %v", err, tt.wantErr)
			}

			pvc, err := kubeclient.GetPersistentVolumeClaim(client, tt.wantPvcName, engine.namespace)
			if err != nil {
				t.Errorf("Got error when checking pvc labels: %v", err)
			}

			if wrappedBy, exists := pvc.Labels[common.LabelAnnotationManagedBy]; !exists {
				t.Errorf("Expect get label \"%s=%s\" on pvc, but not exists", common.LabelAnnotationManagedBy, engine.name)
			} else if wrappedBy != engine.name {
				t.Errorf("Expect get label \"%s=%s\" on pvc, but got %s", common.LabelAnnotationManagedBy, engine.name, wrappedBy)
			}
		})
	}
}

func TestThinEngine_unwrapMountedPersistentVolumeClaims(t *testing.T) {
	testObjs := []runtime.Object{}

	testPVCInputs := []*corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-pvc-1",
				Namespace: "default",
				Labels: map[string]string{
					common.LabelAnnotationManagedBy: "dataset1",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-pvc-2",
				Namespace: "default",
				Labels:    map[string]string{},
			},
		},
	}
	for _, pvcInput := range testPVCInputs {
		testObjs = append(testObjs, pvcInput)
	}

	testRuntimeInputs := []*datav1alpha1.ThinRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dataset1",
				Namespace: "default",
			},
			Status: datav1alpha1.RuntimeStatus{
				Mounts: []datav1alpha1.Mount{
					{
						MountPoint: "pvc://my-pvc-1",
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dataset2",
				Namespace: "default",
			},
			Status: datav1alpha1.RuntimeStatus{
				Mounts: []datav1alpha1.Mount{
					{
						MountPoint: "pvc://my-pvc-2",
					},
				},
			},
		},
	}
	for _, runtimeInput := range testRuntimeInputs {
		testObjs = append(testObjs, runtimeInput)
	}

	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	type fields struct {
		runtime   *datav1alpha1.ThinRuntime
		name      string
		namespace string
	}
	tests := []struct {
		name        string
		fields      fields
		wantErr     bool
		wantPvcName string
	}{
		{
			name: "unwrap_native_pvc",
			fields: fields{
				name:      "dataset1",
				namespace: "default",
			},
			wantErr:     false,
			wantPvcName: "my-pvc-1",
		},
		{
			name: "unwrap_native_pvc_without_label",
			fields: fields{
				name:      "dataset2",
				namespace: "default",
			},
			wantErr:     false,
			wantPvcName: "my-pvc-2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &ThinEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       fake.NullLogger(),
				Client:    client,
			}

			if err := engine.unwrapMountedPersistentVolumeClaims(); (err != nil) != tt.wantErr {
				t.Errorf("ThinEngine.unwrapMountedPersistentVolumeClaims() error = %v, wantErr %v", err, tt.wantErr)
			}

			pvc, err := kubeclient.GetPersistentVolumeClaim(client, tt.wantPvcName, engine.namespace)
			if err != nil {
				t.Errorf("Got error when checking pvc labels: %v", err)
			}

			if _, exists := pvc.Labels[common.LabelAnnotationManagedBy]; exists {
				t.Errorf("Expect no label \"%s\" on pvc, but it exists. pvc Labels: %v", common.LabelAnnotationManagedBy, pvc.Labels)
			}
		})
	}
}
