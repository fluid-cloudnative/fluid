/*
  Copyright 2026 The Fluid Authors.

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

package efc

import (
	"context"
	"os"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	sigclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type mockOperation struct {
	operationType dataoperation.OperationType
	object        sigclient.Object
}

var _ dataoperation.OperationInterface = (*mockOperation)(nil)

func (m *mockOperation) HasPrecedingOperation() bool { return false }

func (m *mockOperation) GetOperationObject() sigclient.Object { return m.object }

func (m *mockOperation) GetPossibleTargetDatasetNamespacedNames() []types.NamespacedName { return nil }

func (m *mockOperation) GetTargetDataset() (*datav1alpha1.Dataset, error) { return nil, nil }

func (m *mockOperation) GetReleaseNameSpacedName() types.NamespacedName {
	return types.NamespacedName{}
}

func (m *mockOperation) GetChartsDirectory() string { return "" }

func (m *mockOperation) GetOperationType() dataoperation.OperationType { return m.operationType }

func (m *mockOperation) UpdateOperationApiStatus(*datav1alpha1.OperationStatus) error { return nil }

func (m *mockOperation) Validate(cruntime.ReconcileRequestContext) ([]datav1alpha1.Condition, error) {
	return nil, nil
}

func (m *mockOperation) UpdateStatusInfoForCompleted(map[string]string) error { return nil }

func (m *mockOperation) SetTargetDatasetStatusInProgress(*datav1alpha1.Dataset) {}

func (m *mockOperation) RemoveTargetDatasetStatusInProgress(*datav1alpha1.Dataset) {}

func (m *mockOperation) GetStatusHandler() dataoperation.StatusHandler { return nil }

func (m *mockOperation) GetTTL() (*int32, error) { return nil, nil }

func (m *mockOperation) GetParallelTaskNumber() int32 { return 1 }

var _ = Describe("EFCEngine operate", func() {
	Describe("GetDataOperationValueFile", func() {
		It("returns not supported for non-dataprocess operations", func() {
			engine := &EFCEngine{Log: fake.NullLogger()}
			operation := &mockOperation{
				operationType: dataoperation.DataBackupType,
				object: &datav1alpha1.DataBackup{
					TypeMeta: metav1.TypeMeta{
						Kind:       "DataBackup",
						APIVersion: "data.fluid.io/v1alpha1",
					},
				},
			}

			valueFileName, err := engine.GetDataOperationValueFile(cruntime.ReconcileRequestContext{}, operation)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not supported"))
			Expect(valueFileName).To(BeEmpty())
		})

		It("delegates dataprocess operations to the data process value generator", func() {
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
			engine := &EFCEngine{
				Client: fake.NewFakeClientWithScheme(testScheme, dataset.DeepCopy()),
				Log:    fake.NullLogger(),
			}
			operation := &mockOperation{
				operationType: dataoperation.DataProcessType,
				object:        dataProcess,
			}

			valueFileName, err := engine.GetDataOperationValueFile(cruntime.ReconcileRequestContext{}, operation)

			Expect(err).NotTo(HaveOccurred())
			Expect(valueFileName).NotTo(BeEmpty())
			defer os.Remove(valueFileName)
		})
	})
})

var _ = Describe("EFCEngine cleanupCache", func() {
	It("returns without error when the values configmap is absent", func() {
		runtimeObj := &datav1alpha1.EFCRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
		}
		engine := &EFCEngine{
			name:       "spark",
			namespace:  "fluid",
			engineImpl: common.EFCEngineImpl,
			Client:     fake.NewFakeClientWithScheme(testScheme, runtimeObj.DeepCopy()),
			Log:        fake.NullLogger(),
		}

		Expect(engine.cleanupCache()).To(Succeed())
	})

	It("skips worker cleanup when the cache uses emptyDir", func() {
		runtimeObj := &datav1alpha1.EFCRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
		}
		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark-efc-values",
				Namespace: "fluid",
			},
			Data: map[string]string{"data": valuesConfigMapData},
		}
		client := fake.NewFakeClientWithScheme(testScheme, runtimeObj.DeepCopy(), configMap.DeepCopy())
		engine := &EFCEngine{
			name:       "spark",
			namespace:  "fluid",
			engineImpl: common.EFCEngineImpl,
			Client:     client,
			Log:        fake.NullLogger(),
		}

		Expect(engine.cleanupCache()).To(Succeed())

		stored := &corev1.ConfigMap{}
		Expect(client.Get(context.TODO(), types.NamespacedName{Name: "spark-efc-values", Namespace: "fluid"}, stored)).To(Succeed())
	})

	It("returns an error when the values configmap cannot be parsed", func() {
		runtimeObj := &datav1alpha1.EFCRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
		}
		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark-efc-values",
				Namespace: "fluid",
			},
			Data: map[string]string{"data": "not: [valid"},
		}
		engine := &EFCEngine{
			name:       "spark",
			namespace:  "fluid",
			engineImpl: common.EFCEngineImpl,
			Client:     fake.NewFakeClientWithScheme(testScheme, runtimeObj.DeepCopy(), configMap.DeepCopy()),
			Log:        fake.NullLogger(),
		}

		err := engine.cleanupCache()

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("parseCacheDirFromConfigMap fail when cleanupCache"))
	})
})
