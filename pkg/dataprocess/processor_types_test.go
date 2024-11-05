/*
Copyright 2023 The Fluid Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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
