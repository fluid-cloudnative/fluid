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

package juicefs

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestJuiceFSEngine_generateDataProcessValueFile(t *testing.T) {
	dataset := &datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-dataset",
			Namespace: "default",
		},
	}

	dataProcess := &datav1alpha1.DataProcess{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-dataprocess",
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
				Script: &datav1alpha1.ScriptProcessor{
					RestartPolicy: corev1.RestartPolicyNever,
					VersionSpec: datav1alpha1.VersionSpec{
						Image:    "test-image",
						ImageTag: "latest",
					},
				},
			},
		},
	}

	type args struct {
		engine *JuiceFSEngine
		ctx    cruntime.ReconcileRequestContext
		object client.Object
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "TestNotOfTypeDataProcess",
			args: args{
				engine: &JuiceFSEngine{},
				ctx:    cruntime.ReconcileRequestContext{},
				object: &datav1alpha1.Dataset{}, // not of type DataProcess
			},
			wantErr: true,
		},
		{
			name: "TestTargetDatasetNotFound",
			args: args{
				engine: &JuiceFSEngine{
					Client: fake.NewFakeClientWithScheme(testScheme), // No dataset
				},
				ctx: cruntime.ReconcileRequestContext{},
				object: &datav1alpha1.DataProcess{
					Spec: datav1alpha1.DataProcessSpec{
						Dataset: datav1alpha1.TargetDatasetWithMountPath{
							TargetDataset: datav1alpha1.TargetDataset{
								Name:      "demo-dataset-notfound",
								Namespace: "default",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "TestGenerateDataProcessValueFile",
			args: args{
				engine: &JuiceFSEngine{
					Client: fake.NewFakeClientWithScheme(testScheme, dataset),
				},
				ctx:    cruntime.ReconcileRequestContext{},
				object: dataProcess,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.args.engine.generateDataProcessValueFile(tt.args.ctx, tt.args.object)
			if (err != nil) != tt.wantErr {
				t.Errorf("JuiceFSEngine.generateDataProcessValueFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
