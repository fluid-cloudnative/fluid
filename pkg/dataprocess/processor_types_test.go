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

package dataprocess

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func TestScriptProcessorImpl_ValidateDatasetMountPath(t *testing.T) {
	type fields struct {
		ScriptProcessor *datav1alpha1.ScriptProcessor
	}
	type args struct {
		datasetMountPath string
	}
	tests := []struct {
		name                      string
		fields                    fields
		args                      args
		wantPass                  bool
		wantConflictVolName       string
		wantConflictContainerName string
	}{
		{
			name: "TestConflictVolMountPath",
			fields: fields{
				ScriptProcessor: &datav1alpha1.ScriptProcessor{
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "myvol",
							MountPath: "/fluid-data",
						},
					},
				},
			},
			args: args{
				datasetMountPath: "/fluid-data",
			},
			wantPass:                  false,
			wantConflictVolName:       "myvol",
			wantConflictContainerName: DataProcessScriptProcessorContainerName,
		},
		{
			name: "TestRegular",
			fields: fields{
				ScriptProcessor: &datav1alpha1.ScriptProcessor{
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "myvol",
							MountPath: "/mydata",
						},
					},
				},
			},
			args: args{
				datasetMountPath: "/fluid-data",
			},
			wantPass:                  true,
			wantConflictVolName:       "",
			wantConflictContainerName: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ScriptProcessorImpl{
				ScriptProcessor: tt.fields.ScriptProcessor,
			}
			gotPass, gotConflictVolName, gotConflictContainerName := p.ValidateDatasetMountPath(tt.args.datasetMountPath)
			if gotPass != tt.wantPass {
				t.Errorf("ScriptProcessorImpl.ValidateDatasetMountPath() gotPass = %v, want %v", gotPass, tt.wantPass)
			}
			if gotConflictVolName != tt.wantConflictVolName {
				t.Errorf("ScriptProcessorImpl.ValidateDatasetMountPath() gotConflictVolName = %v, want %v", gotConflictVolName, tt.wantConflictVolName)
			}
			if gotConflictContainerName != tt.wantConflictContainerName {
				t.Errorf("ScriptProcessorImpl.ValidateDatasetMountPath() gotConflictContainerName = %v, want %v", gotConflictContainerName, tt.wantConflictContainerName)
			}
		})
	}
}

func TestJobProcessorImpl_ValidateDatasetMountPath(t *testing.T) {
	type fields struct {
		JobProcessor *datav1alpha1.JobProcessor
	}
	type args struct {
		datasetMountPath string
	}
	tests := []struct {
		name                      string
		fields                    fields
		args                      args
		wantPass                  bool
		wantConflictVolName       string
		wantConflictContainerName string
	}{
		{
			name: "TestContainerConflictVolMountPath",
			fields: fields{
				JobProcessor: &datav1alpha1.JobProcessor{
					PodSpec: &corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "test-container",
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "myvol",
										MountPath: "/fluid-data",
									},
								},
							},
						},
					},
				},
			},
			args: args{
				datasetMountPath: "/fluid-data",
			},
			wantPass:                  false,
			wantConflictVolName:       "myvol",
			wantConflictContainerName: "test-container",
		},
		{
			name: "TestInitContainerConflictVolMountPath",
			fields: fields{
				JobProcessor: &datav1alpha1.JobProcessor{
					PodSpec: &corev1.PodSpec{
						InitContainers: []corev1.Container{
							{
								Name: "test-init-container",
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "myvol",
										MountPath: "/fluid-data",
									},
								},
							},
						},
					},
				},
			},
			args: args{
				datasetMountPath: "/fluid-data",
			},
			wantPass:                  false,
			wantConflictVolName:       "myvol",
			wantConflictContainerName: "test-init-container",
		},
		{
			name: "TestValidateDatasetMountPath",
			fields: fields{
				JobProcessor: &datav1alpha1.JobProcessor{
					PodSpec: &corev1.PodSpec{
						InitContainers: []corev1.Container{
							{
								Name: "test-init-container",
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "myvol",
										MountPath: "/mydata",
									},
								},
							},
						},
						Containers: []corev1.Container{
							{
								Name: "test-container",
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "myvol",
										MountPath: "/mydata",
									},
								},
							},
						},
					},
				},
			},
			args: args{
				datasetMountPath: "/fluid-data",
			},
			wantPass:                  true,
			wantConflictVolName:       "",
			wantConflictContainerName: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &JobProcessorImpl{
				JobProcessor: tt.fields.JobProcessor,
			}
			gotPass, gotConflictVolName, gotConflictContainerName := p.ValidateDatasetMountPath(tt.args.datasetMountPath)
			if gotPass != tt.wantPass {
				t.Errorf("JobProcessorImpl.ValidateDatasetMountPath() gotPass = %v, want %v", gotPass, tt.wantPass)
			}
			if gotConflictVolName != tt.wantConflictVolName {
				t.Errorf("JobProcessorImpl.ValidateDatasetMountPath() gotConflictVolName = %v, want %v", gotConflictVolName, tt.wantConflictVolName)
			}
			if gotConflictContainerName != tt.wantConflictContainerName {
				t.Errorf("JobProcessorImpl.ValidateDatasetMountPath() gotConflictContainerName = %v, want %v", gotConflictContainerName, tt.wantConflictContainerName)
			}
		})
	}
}
