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
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGenDataProcessValue(t *testing.T) {
	dataset := &datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-dataset",
			Namespace: "default",
		},
	}

	dataProcessScriptProcessor := &datav1alpha1.DataProcess{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-process",
			Namespace: "default",
		},
		Spec: datav1alpha1.DataProcessSpec{
			Dataset: datav1alpha1.TargetDatasetWithMountPath{
				TargetDataset: datav1alpha1.TargetDataset{
					Name:      dataset.Name,
					Namespace: dataset.Namespace,
				},
				MountPath: "/data",
			},
			Processor: datav1alpha1.Processor{
				Script: &datav1alpha1.ScriptProcessor{
					VersionSpec: datav1alpha1.VersionSpec{
						Image:           "test-image",
						ImageTag:        "latest",
						ImagePullPolicy: "IfNotPresent",
					},
					RestartPolicy: corev1.RestartPolicyNever,
					Command:       []string{"bash"},
					Source:        "sleep inf",
					Env: []corev1.EnvVar{
						{
							Name:  "TEST_ENV",
							Value: "foobar",
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "mycode",
							MountPath: "/code",
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "mycode",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "mypvc",
								},
							},
						},
					},
				},
			},
		},
	}

	dataProcessJobProcessor := &datav1alpha1.DataProcess{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-process",
			Namespace: "default",
		},
		Spec: datav1alpha1.DataProcessSpec{
			Dataset: datav1alpha1.TargetDatasetWithMountPath{
				TargetDataset: datav1alpha1.TargetDataset{
					Name:      "demo-dataset",
					Namespace: "default",
				},
				MountPath: "/data",
			},
			Processor: datav1alpha1.Processor{
				Job: &datav1alpha1.JobProcessor{
					PodSpec: &corev1.PodSpec{
						RestartPolicy: corev1.RestartPolicyOnFailure,
						Containers: []corev1.Container{
							{
								Image:           "test-image",
								ImagePullPolicy: "IfNotPresent",
							},
						},
					},
				},
			},
		},
	}

	modifiedPodSpec := &corev1.PodSpec{
		RestartPolicy: corev1.RestartPolicyOnFailure,
		Containers: []corev1.Container{
			{
				Image:           "test-image",
				ImagePullPolicy: "IfNotPresent",
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "fluid-dataset-vol",
						MountPath: "/data",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "fluid-dataset-vol",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: dataset.Name,
					},
				},
			},
		},
	}

	type args struct {
		dataset     *datav1alpha1.Dataset
		dataProcess *datav1alpha1.DataProcess
	}
	tests := []struct {
		name string
		args args
		want *DataProcessValue
	}{
		{
			name: "TestScriptProcessor",
			args: args{
				dataset:     dataset,
				dataProcess: dataProcessScriptProcessor,
			},
			want: &DataProcessValue{
				Name: dataProcessScriptProcessor.Name,
				DataProcessInfo: DataProcessInfo{
					TargetDataset: dataset.Name,
					JobProcessor:  nil,
					ScriptProcessor: &ScriptProcessor{
						Image:           "test-image:latest",
						ImagePullPolicy: "IfNotPresent",
						RestartPolicy:   dataProcessScriptProcessor.Spec.Processor.Script.RestartPolicy,
						Envs:            dataProcessScriptProcessor.Spec.Processor.Script.Env,
						Command:         dataProcessScriptProcessor.Spec.Processor.Script.Command,
						Source:          dataProcessScriptProcessor.Spec.Processor.Script.Source,
						Volumes:         append(dataProcessScriptProcessor.Spec.Processor.Script.Volumes, corev1.Volume{Name: "fluid-dataset-vol", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: dataset.Name}}}),
						VolumeMounts:    append(dataProcessScriptProcessor.Spec.Processor.Script.VolumeMounts, corev1.VolumeMount{Name: "fluid-dataset-vol", MountPath: dataProcessScriptProcessor.Spec.Dataset.MountPath}),
					},
				},
			},
		},
		{
			name: "TestJobProcessor",
			args: args{
				dataset:     dataset,
				dataProcess: dataProcessJobProcessor,
			},
			want: &DataProcessValue{
				Name: dataProcessJobProcessor.Name,
				DataProcessInfo: DataProcessInfo{
					TargetDataset:   dataset.Name,
					ScriptProcessor: nil,
					JobProcessor: &JobProcessor{
						PodSpec: modifiedPodSpec,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenDataProcessValue(tt.args.dataset, tt.args.dataProcess); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenDataProcessValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
