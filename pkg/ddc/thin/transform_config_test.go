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
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestThinEngine_extractVolumeInfo(t *testing.T) {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pvc",
			Namespace: "fluid",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			VolumeName: "test-pv",
		},
		Status: corev1.PersistentVolumeClaimStatus{
			Phase: corev1.ClaimBound,
		},
	}

	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pv",
		},
		Spec: corev1.PersistentVolumeSpec{
			MountOptions: []string{"rw", "noexec"},
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				CSI: &corev1.CSIPersistentVolumeSource{
					NodePublishSecretRef: &corev1.SecretReference{
						Name:      "my-secret",
						Namespace: "node-publish-secrets",
					},
					VolumeHandle: "test-pv",
					VolumeAttributes: map[string]string{
						"test-attr":  "true",
						"test-attr2": "foobar",
					},
				},
			},
		},
	}

	client := fake.NewFakeClientWithScheme(testScheme, pvc, pv)

	engine := ThinEngine{
		name:      "thin-test",
		namespace: "fluid",
		Client:    client,
		Log:       fake.NullLogger(),
	}

	tests := []struct {
		name             string
		pvcName          string
		wantCsiInfo      *corev1.CSIPersistentVolumeSource
		wantMountOptions []string
		wantErr          bool
	}{
		{
			name:    "testExtractVolumeInfo",
			pvcName: "test-pvc",
			wantCsiInfo: &corev1.CSIPersistentVolumeSource{
				NodePublishSecretRef: &corev1.SecretReference{
					Name:      "my-secret",
					Namespace: "node-publish-secrets",
				},
				VolumeHandle: "test-pv",
				VolumeAttributes: map[string]string{
					"test-attr":  "true",
					"test-attr2": "foobar",
				},
			},
			wantMountOptions: []string{"rw", "noexec"},
			wantErr:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCsiInfo, gotMountOptions, err := engine.extractVolumeInfo(tt.pvcName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ThinEngine.extractVolumeInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCsiInfo, tt.wantCsiInfo) {
				t.Errorf("ThinEngine.extractVolumeInfo() gotCsiInfo = %v, want %v", gotCsiInfo, tt.wantCsiInfo)
			}
			if !reflect.DeepEqual(gotMountOptions, tt.wantMountOptions) {
				t.Errorf("ThinEngine.extractVolumeInfo() gotMountOptions = %v, want %v", gotMountOptions, tt.wantMountOptions)
			}
		})
	}
}

func TestThinEngine_extractVolumeMountOptions(t *testing.T) {
	engine := ThinEngine{}

	tests := []struct {
		name             string
		pv               *corev1.PersistentVolume
		wantMountOptions []string
		wantErr          bool
	}{
		{
			name: "test_mount_options_in_annotation",
			pv: &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						corev1.MountOptionAnnotation: "rw,noexec,testOpts",
					},
				},
			},
			wantMountOptions: []string{"rw", "noexec", "testOpts"},
			wantErr:          false,
		},
		{
			name: "test_mount_options_in_proerty",
			pv: &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
				Spec: corev1.PersistentVolumeSpec{
					MountOptions: []string{"ro", "noexec"},
				},
			},
			wantMountOptions: []string{"ro", "noexec"},
			wantErr:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMountOptions, err := engine.extractVolumeMountOptions(tt.pv)
			if (err != nil) != tt.wantErr {
				t.Errorf("ThinEngine.extractVolumeMountOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotMountOptions, tt.wantMountOptions) {
				t.Errorf("ThinEngine.extractVolumeMountOptions() = %v, want %v", gotMountOptions, tt.wantMountOptions)
			}
		})
	}
}
