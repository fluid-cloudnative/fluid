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

// TestJuiceFSEngine_generateDataProcessValueFile tests the generateDataProcessValueFile function of the JuiceFSEngine.
// It verifies the behavior of the function under different scenarios, including invalid input and missing datasets.
//
// Parameters:
//    - t: The testing context used for reporting test failures and logging.
func TestJuiceFSEngine_generateDataProcessValueFile(t *testing.T) {
	// 1. Define a sample dataset and dataProcess object for testing.
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

	// 2. Define the test cases using a struct to encapsulate the test arguments and expected results.
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

	// 3. Run each test case and verify the results.
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
