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
	"os"
	"strings"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cdatabackup "github.com/fluid-cloudnative/fluid/pkg/databackup"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"
)

const (
	testDatasetName      = "test-dataset"
	testDatasetNamespace = "default"
)

func TestGenerateDataBackupValueFile(t *testing.T) {
	testCases := []struct {
		name             string
		dataBackup       *datav1alpha1.DataBackup
		runtime          *datav1alpha1.AlluxioRuntime
		dataset          *datav1alpha1.Dataset
		masterPod        *corev1.Pod
		expectError      bool
		errorMsg         string
		expectValueCheck func(t *testing.T, values *cdatabackup.DataBackupValue)
	}{
		{
			name: "valid databackup with single master",
			dataBackup: &datav1alpha1.DataBackup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-backup",
					Namespace: testDatasetNamespace,
				},
				Spec: datav1alpha1.DataBackupSpec{
					Dataset:    testDatasetName,
					BackupPath: "pvc://backup-pvc/data",
				},
			},
			runtime: &datav1alpha1.AlluxioRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testDatasetName,
					Namespace: testDatasetNamespace,
				},
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Replicas: 1,
				},
			},
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testDatasetName,
					Namespace: testDatasetNamespace,
					UID:       "test-uid",
				},
			},
			masterPod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataset-master-0",
					Namespace: testDatasetNamespace,
				},
				Spec: corev1.PodSpec{
					NodeName: "test-node",
					Containers: []corev1.Container{
						{
							Name: "alluxio-master",
							Ports: []corev1.ContainerPort{
								{
									Name:          "rpc",
									ContainerPort: 19998,
								},
							},
						},
					},
				},
				Status: corev1.PodStatus{
					PodIP: "10.0.0.1",
				},
			},
			expectError: false,
			expectValueCheck: func(t *testing.T, values *cdatabackup.DataBackupValue) {
				if values.UserInfo.User != 0 {
					t.Errorf("expected default user 0, got %d", values.UserInfo.User)
				}
				if values.InitUsers.Enabled {
					t.Error("expected InitUsers to be disabled, but it was enabled")
				}
			},
		},
		{
			name: "valid databackup with RunAs",
			dataBackup: &datav1alpha1.DataBackup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-backup-runas",
					Namespace: testDatasetNamespace,
				},
				Spec: datav1alpha1.DataBackupSpec{
					Dataset:    testDatasetName,
					BackupPath: "pvc://backup-pvc/data",
					RunAs: &datav1alpha1.User{
						UID: ptr.To(int64(1000)),
						GID: ptr.To(int64(1000)),
					},
				},
			},
			runtime: &datav1alpha1.AlluxioRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testDatasetName,
					Namespace: testDatasetNamespace,
				},
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Replicas: 1,
				},
			},
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testDatasetName,
					Namespace: testDatasetNamespace,
					UID:       "test-uid",
				},
			},
			masterPod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataset-master-0",
					Namespace: testDatasetNamespace,
				},
				Spec: corev1.PodSpec{
					NodeName: "test-node",
					Containers: []corev1.Container{
						{
							Name: "alluxio-master",
							Ports: []corev1.ContainerPort{
								{
									Name:          "rpc",
									ContainerPort: 19998,
								},
							},
						},
					},
				},
				Status: corev1.PodStatus{
					PodIP: "10.0.0.1",
				},
			},
			expectError: false,
			expectValueCheck: func(t *testing.T, values *cdatabackup.DataBackupValue) {
				if values.UserInfo.User != 1000 {
					t.Errorf("expected user 1000, got %d", values.UserInfo.User)
				}
				if values.UserInfo.Group != 1000 {
					t.Errorf("expected group 1000, got %d", values.UserInfo.Group)
				}
				if !values.InitUsers.Enabled {
					t.Error("expected InitUsers to be enabled")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			objs := []runtime.Object{}
			if tc.runtime != nil {
				objs = append(objs, tc.runtime)
			}
			if tc.dataset != nil {
				objs = append(objs, tc.dataset)
			}
			if tc.masterPod != nil {
				objs = append(objs, tc.masterPod)
			}

			client := fake.NewFakeClientWithScheme(testScheme, objs...)
			runtimeInfo, err := base.BuildRuntimeInfo(testDatasetName, testDatasetNamespace, "alluxio")
			if err != nil {
				t.Fatalf("failed to build runtime info: %v", err)
			}

			engine := &AlluxioEngine{
				Client:      client,
				name:        testDatasetName,
				namespace:   testDatasetNamespace,
				runtime:     tc.runtime,
				runtimeInfo: runtimeInfo,
				Log:         fake.NullLogger(),
			}

			ctx := cruntime.ReconcileRequestContext{
				Log: fake.NullLogger(),
			}

			valueFileName, err := engine.generateDataBackupValueFile(ctx, tc.dataBackup)

			if tc.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tc.errorMsg)
				} else if tc.errorMsg != "" && !strings.Contains(err.Error(), tc.errorMsg) {
					t.Errorf("expected error containing %q, got %q", tc.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if valueFileName == "" {
					t.Error("expected non-empty value file name")
				} else {
					defer os.Remove(valueFileName)
					if _, err := os.Stat(valueFileName); os.IsNotExist(err) {
						t.Errorf("value file %s was not created", valueFileName)
					}

					if tc.expectValueCheck != nil {
						data, err := os.ReadFile(valueFileName)
						if err != nil {
							t.Fatalf("failed to read value file: %v", err)
						}

						var values cdatabackup.DataBackupValue
						err = yaml.Unmarshal(data, &values)
						if err != nil {
							t.Fatalf("failed to unmarshal value file: %v", err)
						}
						tc.expectValueCheck(t, &values)
					}
				}
			}
		})
	}
}

func TestGenerateDataBackupValueFileInvalidObject(t *testing.T) {
	client := fake.NewFakeClientWithScheme(testScheme)
	runtimeInfo, err := base.BuildRuntimeInfo("test", testDatasetNamespace, "alluxio")
	if err != nil {
		t.Fatalf("failed to build runtime info: %v", err)
	}

	engine := &AlluxioEngine{
		Client:      client,
		name:        "test",
		namespace:   testDatasetNamespace,
		runtimeInfo: runtimeInfo,
		Log:         fake.NullLogger(),
	}

	ctx := cruntime.ReconcileRequestContext{
		Log: fake.NullLogger(),
	}

	wrongTypeObject := &datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "not-a-databackup",
			Namespace: testDatasetNamespace,
		},
	}

	valueFileName, err := engine.generateDataBackupValueFile(ctx, wrongTypeObject)

	if err == nil {
		t.Error("expected error for invalid object type, got nil")
	}
	if valueFileName != "" {
		t.Errorf("expected empty value file name, got %s", valueFileName)
	}
}

func TestGenerateDataBackupValueFileRuntimeNotFound(t *testing.T) {
	dataBackup := &datav1alpha1.DataBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-backup",
			Namespace: testDatasetNamespace,
		},
		Spec: datav1alpha1.DataBackupSpec{
			Dataset:    "nonexistent-dataset",
			BackupPath: "pvc://backup-pvc/data",
		},
	}

	client := fake.NewFakeClientWithScheme(testScheme)
	runtimeInfo, err := base.BuildRuntimeInfo("nonexistent-dataset", testDatasetNamespace, "alluxio")
	if err != nil {
		t.Fatalf("failed to build runtime info: %v", err)
	}

	engine := &AlluxioEngine{
		Client:      client,
		name:        "nonexistent-dataset",
		namespace:   testDatasetNamespace,
		runtimeInfo: runtimeInfo,
		Log:         fake.NullLogger(),
	}

	ctx := cruntime.ReconcileRequestContext{
		Log: fake.NullLogger(),
	}

	valueFileName, err := engine.generateDataBackupValueFile(ctx, dataBackup)

	if err == nil {
		t.Error("expected error for runtime not found, got nil")
	}
	if valueFileName != "" {
		t.Errorf("expected empty value file name, got %s", valueFileName)
	}
}
