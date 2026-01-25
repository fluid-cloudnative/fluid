/*
  Copyright 2024 The Fluid Authors.

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

package thin

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func TestValidateDuplicateDatasetMounts(t *testing.T) {
	const (
		mountPointPath1 = "s3://bucket/path1"
		mountPointPath2 = "s3://bucket/path2"
		dataPath        = "/data"
		dataPath1       = "/data1"
		dataPath2       = "/data2"
	)

	testCases := []struct {
		name      string
		dataset   *datav1alpha1.Dataset
		expectErr bool
	}{
		{
			name:      "nil dataset",
			dataset:   nil,
			expectErr: false,
		},
		{
			name: "empty mounts",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{},
				},
			},
			expectErr: false,
		},
		{
			name: "single mount",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: "s3://bucket/path",
							Name:       "data",
							Path:       "/data",
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "multiple mounts with unique names and paths",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: mountPointPath1,
							Name:       "data1",
							Path:       dataPath1,
						},
						{
							MountPoint: mountPointPath2,
							Name:       "data2",
							Path:       dataPath2,
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "duplicate mount names",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: mountPointPath1,
							Name:       "data",
							Path:       dataPath1,
						},
						{
							MountPoint: mountPointPath2,
							Name:       "data",
							Path:       dataPath2,
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "duplicate mount paths",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: mountPointPath1,
							Name:       "data1",
							Path:       dataPath,
						},
						{
							MountPoint: mountPointPath2,
							Name:       "data2",
							Path:       dataPath,
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "empty path defaults to name based path",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: mountPointPath1,
							Name:       "data1",
						},
						{
							MountPoint: mountPointPath2,
							Name:       "data2",
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "empty paths with same names create duplicates",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: mountPointPath1,
							Name:       "data",
						},
						{
							MountPoint: mountPointPath2,
							Name:       "data",
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "explicit path conflicts with default name path",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: mountPointPath1,
							Name:       "data1",
						},
						{
							MountPoint: mountPointPath2,
							Name:       "data2",
							Path:       dataPath1,
						},
					},
				},
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := cruntime.ReconcileRequestContext{
				Dataset: tc.dataset,
			}
			err := validateDuplicateDatasetMounts(ctx)
			if (err != nil) != tc.expectErr {
				t.Errorf("validateDuplicateDatasetMounts() error = %v, wantErr %v", err, tc.expectErr)
			}
		})
	}
}

func TestThinEngineValidate(t *testing.T) {
	namespace := "fluid"
	runtimeName := "thin-test"

	runtimeInput := &datav1alpha1.ThinRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      runtimeName,
			Namespace: namespace,
		},
		Spec: datav1alpha1.ThinRuntimeSpec{
			Fuse: datav1alpha1.ThinFuseSpec{},
		},
	}

	datasetInput := &datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      runtimeName,
			Namespace: namespace,
			UID:       "test-dataset-uid",
		},
		Spec: datav1alpha1.DatasetSpec{
			PlacementMode: datav1alpha1.ExclusiveMode,
			Mounts: []datav1alpha1.Mount{
				{
					MountPoint: "s3://bucket/path",
					Name:       "data",
					Path:       "/data",
				},
			},
		},
	}

	testObjs := []runtime.Object{}
	testObjs = append(testObjs, runtimeInput.DeepCopy())
	testObjs = append(testObjs, datasetInput.DeepCopy())

	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	runtimeInfo, err := base.BuildRuntimeInfo(runtimeName, namespace, common.ThinRuntime)
	if err != nil {
		t.Fatalf("failed to build runtime info: %v", err)
	}
	runtimeInfo.SetupWithDataset(datasetInput)
	runtimeInfo.SetOwnerDatasetUID(datasetInput.UID)

	engine := &ThinEngine{
		runtime:     runtimeInput,
		name:        runtimeName,
		namespace:   namespace,
		Client:      client,
		runtimeInfo: runtimeInfo,
		Log:         fake.NullLogger(),
	}

	testCases := []struct {
		name      string
		dataset   *datav1alpha1.Dataset
		expectErr bool
	}{
		{
			name:      "valid dataset",
			dataset:   datasetInput,
			expectErr: false,
		},
		{
			name: "dataset with duplicate mount names",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      runtimeName,
					Namespace: namespace,
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: "s3://bucket/path1",
							Name:       "data",
							Path:       "/data1",
						},
						{
							MountPoint: "s3://bucket/path2",
							Name:       "data",
							Path:       "/data2",
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "dataset with duplicate mount paths",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      runtimeName,
					Namespace: namespace,
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: "s3://bucket/path1",
							Name:       "data1",
							Path:       "/data",
						},
						{
							MountPoint: "s3://bucket/path2",
							Name:       "data2",
							Path:       "/data",
						},
					},
				},
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{
					Name:      runtimeName,
					Namespace: namespace,
				},
				Dataset: tc.dataset,
			}
			err := engine.Validate(ctx)
			if (err != nil) != tc.expectErr {
				t.Errorf("engine.Validate() error = %v, wantErr %v", err, tc.expectErr)
			}
		})
	}
}

func TestThinEngineValidateWithRuntimeInfoError(t *testing.T) {
	namespace := "fluid"
	runtimeName := "thin-test"

	runtimeInput := &datav1alpha1.ThinRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      runtimeName,
			Namespace: namespace,
		},
	}

	testObjs := []runtime.Object{}
	testObjs = append(testObjs, runtimeInput.DeepCopy())

	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engine := &ThinEngine{
		runtime:     runtimeInput,
		name:        runtimeName,
		namespace:   namespace,
		Client:      client,
		runtimeInfo: nil,
		Log:         fake.NullLogger(),
	}

	ctx := cruntime.ReconcileRequestContext{
		NamespacedName: types.NamespacedName{
			Name:      runtimeName,
			Namespace: namespace,
		},
	}

	err := engine.Validate(ctx)
	if err == nil {
		t.Errorf("expected error due to missing runtime info but got nil")
	}
}
