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
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/transfromer"
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

	dataProcessScriptProcessorWithoutMountPath := dataProcessScriptProcessor.DeepCopy()
	dataProcessScriptProcessorWithoutMountPath.Spec.Dataset.MountPath = ""

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

	dataProcessJobProcessorWithoutMountPath := dataProcessJobProcessor.DeepCopy()
	dataProcessJobProcessorWithoutMountPath.Spec.Dataset.MountPath = ""

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
	modifiedPodSpecWithoutDatasetVol := modifiedPodSpec.DeepCopy()
	modifiedPodSpecWithoutDatasetVol.Volumes = nil
	modifiedPodSpecWithoutDatasetVol.Containers[0].VolumeMounts = nil

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
				Name:  dataProcessScriptProcessor.Name,
				Owner: transfromer.GenerateOwnerReferenceFromObject(dataProcessScriptProcessor),
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
				Name:  dataProcessJobProcessor.Name,
				Owner: transfromer.GenerateOwnerReferenceFromObject(dataProcessJobProcessor),
				DataProcessInfo: DataProcessInfo{
					TargetDataset:   dataset.Name,
					ScriptProcessor: nil,
					JobProcessor: &JobProcessor{
						PodSpec: modifiedPodSpec,
					},
				},
			},
		},
		{
			name: "TestScriptProcessorWithoutMountPath",
			args: args{
				dataset:     dataset,
				dataProcess: dataProcessScriptProcessorWithoutMountPath,
			},
			want: &DataProcessValue{
				Name:  dataProcessScriptProcessor.Name,
				Owner: transfromer.GenerateOwnerReferenceFromObject(dataProcessScriptProcessorWithoutMountPath),
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
						Volumes:         dataProcessScriptProcessor.Spec.Processor.Script.Volumes,
						VolumeMounts:    dataProcessScriptProcessor.Spec.Processor.Script.VolumeMounts,
					},
				},
			},
		},
		{
			name: "TestJobProcessorWithoutMountPath",
			args: args{
				dataset:     dataset,
				dataProcess: dataProcessJobProcessorWithoutMountPath,
			},
			want: &DataProcessValue{
				Name:  dataProcessJobProcessor.Name,
				Owner: transfromer.GenerateOwnerReferenceFromObject(dataProcessJobProcessorWithoutMountPath),
				DataProcessInfo: DataProcessInfo{
					TargetDataset:   dataset.Name,
					ScriptProcessor: nil,
					JobProcessor: &JobProcessor{
						PodSpec: modifiedPodSpecWithoutDatasetVol,
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
