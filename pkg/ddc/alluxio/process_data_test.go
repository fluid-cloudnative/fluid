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

package alluxio

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TestAlluxioEngine_generateDataProcessValueFile tests the generateDataProcessValueFile 
// function of AlluxioEngine under different input scenarios.
//
// Parameters:
// - dataset: a mock Dataset object for simulating an existing dataset.
// - dataProcess: a mock DataProcess object containing processor and target dataset info.
// - args: includes the engine instance, request context, and the input object to test.
//
// Return:
// - Verifies whether generateDataProcessValueFile returns an error as expected 
//   in each scenario.

func TestAlluxioEngine_generateDataProcessValueFile(t *testing.T) {
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
		engine *AlluxioEngine
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
				engine: &AlluxioEngine{},
				ctx:    cruntime.ReconcileRequestContext{},
				object: &datav1alpha1.Dataset{}, // not of type DataProcess
			},
			wantErr: true,
		},
		{
			name: "TestTargetDatasetNotFound",
			args: args{
				engine: &AlluxioEngine{
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
				engine: &AlluxioEngine{
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
				t.Errorf("AlluxioEngine.generateDataProcessValueFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
