/*
Copyright 2025 The Fluid Authors.

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

package controllers

import (
	"context"
	"fmt"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/ddc"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	jindoutils "github.com/fluid-cloudnative/fluid/pkg/utils/jindo"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// mockOperationInterfaceBuilder implements dataoperation.OperationInterfaceBuilder
type mockOperationInterfaceBuilder struct {
	buildFunc func(object client.Object) (dataoperation.OperationInterface, error)
}

func (m *mockOperationInterfaceBuilder) Build(object client.Object) (dataoperation.OperationInterface, error) {
	if m.buildFunc != nil {
		return m.buildFunc(object)
	}
	return nil, fmt.Errorf("not implemented")
}

// mockOperationInterface implements dataoperation.OperationInterface
type mockOperationInterface struct {
	operationObject                      client.Object
	releaseNamespacedName                types.NamespacedName
	targetDataset                        *datav1alpha1.Dataset
	targetDatasetErr                     error
	possibleTargetDatasetNamespacedNames []types.NamespacedName
	operationType                        dataoperation.OperationType
}

func (m *mockOperationInterface) HasPrecedingOperation() bool { return false }
func (m *mockOperationInterface) GetOperationObject() client.Object {
	return m.operationObject
}
func (m *mockOperationInterface) GetPossibleTargetDatasetNamespacedNames() []types.NamespacedName {
	return m.possibleTargetDatasetNamespacedNames
}
func (m *mockOperationInterface) GetTargetDataset() (*datav1alpha1.Dataset, error) {
	return m.targetDataset, m.targetDatasetErr
}
func (m *mockOperationInterface) GetReleaseNameSpacedName() types.NamespacedName {
	return m.releaseNamespacedName
}
func (m *mockOperationInterface) GetChartsDirectory() string { return "" }
func (m *mockOperationInterface) GetOperationType() dataoperation.OperationType {
	return m.operationType
}
func (m *mockOperationInterface) UpdateOperationApiStatus(opStatus *datav1alpha1.OperationStatus) error {
	return nil
}
func (m *mockOperationInterface) Validate(ctx cruntime.ReconcileRequestContext) ([]datav1alpha1.Condition, error) {
	return nil, nil
}
func (m *mockOperationInterface) UpdateStatusInfoForCompleted(infos map[string]string) error {
	return nil
}
func (m *mockOperationInterface) SetTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {}
func (m *mockOperationInterface) RemoveTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
}
func (m *mockOperationInterface) GetStatusHandler() dataoperation.StatusHandler { return nil }
func (m *mockOperationInterface) GetTTL() (ttl *int32, err error)               { return nil, nil }
func (m *mockOperationInterface) GetParallelTaskNumber() int32                  { return 1 }

var _ = Describe("NewDataOperationReconciler", func() {
	It("should create an OperationReconciler with the provided parameters", func() {
		s := runtime.NewScheme()
		Expect(datav1alpha1.AddToScheme(s)).To(Succeed())

		fakeClient := fakeclient.NewClientBuilder().WithScheme(s).Build()
		fakeRecorder := record.NewFakeRecorder(10)
		log := fake.NullLogger()

		reconciler := NewDataOperationReconciler(nil, fakeClient, log, fakeRecorder)

		Expect(reconciler).NotTo(BeNil())
		Expect(reconciler.Client).To(Equal(fakeClient))
		Expect(reconciler.Log).To(Equal(log))
		Expect(reconciler.Recorder).NotTo(BeNil())
		Expect(reconciler.engines).NotTo(BeNil())
		Expect(reconciler.engines).To(BeEmpty())
		Expect(reconciler.mutex).NotTo(BeNil())
	})
})

var _ = Describe("OperationReconciler engine cache", func() {
	var (
		reconciler *OperationReconciler
		s          *runtime.Scheme
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		Expect(datav1alpha1.AddToScheme(s)).To(Succeed())
		Expect(corev1.AddToScheme(s)).To(Succeed())

		fakeClient := fakeclient.NewClientBuilder().WithScheme(s).Build()
		fakeRecorder := record.NewFakeRecorder(10)
		log := fake.NullLogger()

		reconciler = &OperationReconciler{
			Client:   fakeClient,
			Log:      log,
			Recorder: fakeRecorder,
			engines:  map[string]base.Engine{},
			mutex:    &sync.Mutex{},
		}
	})

	Describe("RemoveEngine", func() {
		It("should remove an engine from the cache by namespaced name", func() {
			nn := types.NamespacedName{Namespace: "default", Name: "test-runtime"}
			// Use ddc.GenerateEngineID to get the correct key format
			id := ddc.GenerateEngineID(nn)
			reconciler.engines[id] = nil // placeholder

			Expect(reconciler.engines).To(HaveLen(1))

			reconciler.RemoveEngine(nn)

			Expect(reconciler.engines).To(BeEmpty())
		})

		It("should not panic when removing a non-existent engine", func() {
			nn := types.NamespacedName{Namespace: "default", Name: "nonexistent"}
			Expect(func() {
				reconciler.RemoveEngine(nn)
			}).NotTo(Panic())
		})

		It("should only remove the targeted engine and leave others intact", func() {
			nn1 := types.NamespacedName{Namespace: "ns1", Name: "runtime1"}
			nn2 := types.NamespacedName{Namespace: "ns2", Name: "runtime2"}
			reconciler.engines[ddc.GenerateEngineID(nn1)] = nil
			reconciler.engines[ddc.GenerateEngineID(nn2)] = nil

			Expect(reconciler.engines).To(HaveLen(2))

			reconciler.RemoveEngine(nn1)

			Expect(reconciler.engines).To(HaveLen(1))
			_, exists := reconciler.engines[ddc.GenerateEngineID(nn2)]
			Expect(exists).To(BeTrue())
		})

		It("should handle concurrent RemoveEngine calls safely", func() {
			done := make(chan struct{})
			go func() {
				defer close(done)
				for i := 0; i < 100; i++ {
					reconciler.RemoveEngine(types.NamespacedName{
						Namespace: "ns",
						Name:      "runtime",
					})
				}
			}()

			for i := 0; i < 100; i++ {
				reconciler.RemoveEngine(types.NamespacedName{
					Namespace: "ns",
					Name:      "runtime",
				})
			}

			Eventually(done).Should(BeClosed())
		})
	})

	Describe("GetOrCreateEngine", func() {
		It("should return cached engine when it already exists", func() {
			nn := types.NamespacedName{Namespace: "default", Name: "test-runtime"}
			id := ddc.GenerateEngineID(nn)
			existingEngine := &fakeEngineCore{id: id}
			reconciler.engines[id] = existingEngine

			ctx := dataoperation.ReconcileRequestContext{
				ReconcileRequestContext: cruntime.ReconcileRequestContext{
					Context: context.Background(),
					NamespacedName: types.NamespacedName{
						Name:      "test-runtime",
						Namespace: "default",
					},
				},
			}

			engine, err := reconciler.GetOrCreateEngine(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(engine).To(Equal(existingEngine))
		})
	})
})

var _ = Describe("OperationReconciler getRuntimeObjectAndEngineImpl", func() {
	var (
		reconciler *OperationReconciler
		s          *runtime.Scheme
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		Expect(datav1alpha1.AddToScheme(s)).To(Succeed())
		Expect(corev1.AddToScheme(s)).To(Succeed())
	})

	It("should return an error for unsupported runtime type", func() {
		fakeClient := fakeclient.NewClientBuilder().WithScheme(s).Build()

		reconciler = &OperationReconciler{
			Client:  fakeClient,
			Log:     fake.NullLogger(),
			engines: map[string]base.Engine{},
			mutex:   &sync.Mutex{},
		}

		_, _, err := reconciler.getRuntimeObjectAndEngineImpl("unsupported-runtime", "test", "default")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("not supported"))
	})

	It("should return an error when alluxio runtime is not found", func() {
		fakeClient := fakeclient.NewClientBuilder().WithScheme(s).Build()

		reconciler = &OperationReconciler{
			Client:  fakeClient,
			Log:     fake.NullLogger(),
			engines: map[string]base.Engine{},
			mutex:   &sync.Mutex{},
		}

		_, _, err := reconciler.getRuntimeObjectAndEngineImpl(common.AlluxioRuntime, "nonexistent", "default")
		Expect(err).To(HaveOccurred())
	})

	It("should return the runtime object and engine impl for an existing alluxio runtime", func() {
		alluxioRuntime := &datav1alpha1.AlluxioRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-alluxio",
				Namespace: "default",
			},
		}

		fakeClient := fakeclient.NewClientBuilder().
			WithScheme(s).
			WithObjects(alluxioRuntime).
			Build()

		reconciler = &OperationReconciler{
			Client:  fakeClient,
			Log:     fake.NullLogger(),
			engines: map[string]base.Engine{},
			mutex:   &sync.Mutex{},
		}

		obj, engineImpl, err := reconciler.getRuntimeObjectAndEngineImpl(common.AlluxioRuntime, "test-alluxio", "default")
		Expect(err).NotTo(HaveOccurred())
		Expect(obj).NotTo(BeNil())
		Expect(engineImpl).To(Equal(common.AlluxioEngineImpl))
	})

	It("should return ThinEngineImpl for thin runtime type", func() {
		thinRuntime := &datav1alpha1.ThinRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-thin",
				Namespace: "default",
			},
		}

		fakeClient := fakeclient.NewClientBuilder().
			WithScheme(s).
			WithObjects(thinRuntime).
			Build()

		reconciler = &OperationReconciler{
			Client:  fakeClient,
			Log:     fake.NullLogger(),
			engines: map[string]base.Engine{},
			mutex:   &sync.Mutex{},
		}

		obj, engineImpl, err := reconciler.getRuntimeObjectAndEngineImpl(common.ThinRuntime, "test-thin", "default")
		Expect(err).NotTo(HaveOccurred())
		Expect(obj).NotTo(BeNil())
		Expect(engineImpl).To(Equal(common.ThinEngineImpl))
	})

	DescribeTable("should return not-found error for missing runtimes of each type",
		func(runtimeType string) {
			fakeClient := fakeclient.NewClientBuilder().WithScheme(s).Build()

			reconciler = &OperationReconciler{
				Client:  fakeClient,
				Log:     fake.NullLogger(),
				engines: map[string]base.Engine{},
				mutex:   &sync.Mutex{},
			}

			_, _, err := reconciler.getRuntimeObjectAndEngineImpl(runtimeType, "nonexistent", "default")
			Expect(err).To(HaveOccurred())
		},
		Entry("alluxio", common.AlluxioRuntime),
		Entry("jindo", common.JindoRuntime),
		Entry("goosefs", common.GooseFSRuntime),
		Entry("juicefs", common.JuiceFSRuntime),
		Entry("efc", common.EFCRuntime),
		Entry("thin", common.ThinRuntime),
		Entry("vineyard", common.VineyardRuntime),
	)

	DescribeTable("should return the correct engine impl for existing runtimes",
		func(runtimeType string, runtimeObj runtime.Object, expectedEngineImpl string) {
			fakeClient := fakeclient.NewClientBuilder().
				WithScheme(s).
				WithRuntimeObjects(runtimeObj).
				Build()

			reconciler = &OperationReconciler{
				Client:  fakeClient,
				Log:     fake.NullLogger(),
				engines: map[string]base.Engine{},
				mutex:   &sync.Mutex{},
			}

			obj, engineImpl, err := reconciler.getRuntimeObjectAndEngineImpl(runtimeType, "test-runtime", "default")
			Expect(err).NotTo(HaveOccurred())
			Expect(obj).NotTo(BeNil())
			Expect(engineImpl).To(Equal(expectedEngineImpl))
		},
		Entry("goosefs", common.GooseFSRuntime, &datav1alpha1.GooseFSRuntime{
			ObjectMeta: metav1.ObjectMeta{Name: "test-runtime", Namespace: "default"},
		}, common.GooseFSEngineImpl),
		Entry("jindo", common.JindoRuntime, &datav1alpha1.JindoRuntime{
			ObjectMeta: metav1.ObjectMeta{Name: "test-runtime", Namespace: "default"},
		}, jindoutils.GetDefaultEngineImpl()),
		Entry("juicefs", common.JuiceFSRuntime, &datav1alpha1.JuiceFSRuntime{
			ObjectMeta: metav1.ObjectMeta{Name: "test-runtime", Namespace: "default"},
		}, common.JuiceFSEngineImpl),
		Entry("efc", common.EFCRuntime, &datav1alpha1.EFCRuntime{
			ObjectMeta: metav1.ObjectMeta{Name: "test-runtime", Namespace: "default"},
		}, common.EFCEngineImpl),
		Entry("vineyard", common.VineyardRuntime, &datav1alpha1.VineyardRuntime{
			ObjectMeta: metav1.ObjectMeta{Name: "test-runtime", Namespace: "default"},
		}, common.VineyardEngineImpl),
	)
})

var _ = Describe("OperationReconciler addFinalizerAndRequeue", func() {
	var (
		reconciler *OperationReconciler
		s          *runtime.Scheme
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		Expect(datav1alpha1.AddToScheme(s)).To(Succeed())
		Expect(corev1.AddToScheme(s)).To(Succeed())
	})

	It("should add a finalizer to the data operation object and requeue", func() {
		dataload := &datav1alpha1.DataLoad{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-dataload",
				Namespace: "default",
			},
		}

		fakeClient := fakeclient.NewClientBuilder().
			WithScheme(s).
			WithObjects(dataload).
			Build()

		reconciler = &OperationReconciler{
			Client:   fakeClient,
			Log:      fake.NullLogger(),
			Recorder: record.NewFakeRecorder(10),
			engines:  map[string]base.Engine{},
			mutex:    &sync.Mutex{},
		}

		ctx := dataoperation.ReconcileRequestContext{
			ReconcileRequestContext: cruntime.ReconcileRequestContext{
				Context: context.Background(),
				Dataset: &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-dataset",
						Namespace: "default",
					},
				},
			},
			DataOpFinalizerName: "fluid.io/dataload-finalizer",
		}
		ctx.Log = fake.NullLogger()

		result, err := reconciler.addFinalizerAndRequeue(ctx, dataload)
		Expect(err).NotTo(HaveOccurred())
		// Should requeue
		Expect(result.Requeue || result.RequeueAfter > 0 || !result.IsZero()).To(BeTrue())

		// Verify the finalizer was added
		updatedDataload := &datav1alpha1.DataLoad{}
		Expect(fakeClient.Get(context.Background(), types.NamespacedName{
			Name:      "test-dataload",
			Namespace: "default",
		}, updatedDataload)).To(Succeed())
		Expect(updatedDataload.GetFinalizers()).To(ContainElement("fluid.io/dataload-finalizer"))
	})
})

var _ = Describe("OperationReconciler addOwnerAndRequeue", func() {
	var (
		reconciler *OperationReconciler
		s          *runtime.Scheme
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		Expect(datav1alpha1.AddToScheme(s)).To(Succeed())
		Expect(corev1.AddToScheme(s)).To(Succeed())
	})

	It("should add an owner reference and requeue", func() {
		dataload := &datav1alpha1.DataLoad{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-dataload",
				Namespace: "default",
			},
		}

		dataset := &datav1alpha1.Dataset{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Dataset",
				APIVersion: "data.fluid.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-dataset",
				Namespace: "default",
				UID:       "dataset-uid-456",
			},
		}

		fakeClient := fakeclient.NewClientBuilder().
			WithScheme(s).
			WithObjects(dataload, dataset).
			Build()

		reconciler = &OperationReconciler{
			Client:   fakeClient,
			Log:      fake.NullLogger(),
			Recorder: record.NewFakeRecorder(10),
			engines:  map[string]base.Engine{},
			mutex:    &sync.Mutex{},
		}

		ctx := dataoperation.ReconcileRequestContext{
			ReconcileRequestContext: cruntime.ReconcileRequestContext{
				Context: context.Background(),
			},
		}
		ctx.Log = fake.NullLogger()

		result, err := reconciler.addOwnerAndRequeue(ctx, dataload, dataset)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Requeue).To(BeTrue())

		// Verify the owner reference was added
		updatedDataload := &datav1alpha1.DataLoad{}
		Expect(fakeClient.Get(context.Background(), types.NamespacedName{
			Name:      "test-dataload",
			Namespace: "default",
		}, updatedDataload)).To(Succeed())
		Expect(updatedDataload.GetOwnerReferences()).To(HaveLen(1))
		Expect(updatedDataload.GetOwnerReferences()[0].Name).To(Equal("test-dataset"))
		Expect(updatedDataload.GetOwnerReferences()[0].UID).To(Equal(types.UID("dataset-uid-456")))
		Expect(updatedDataload.GetOwnerReferences()[0].APIVersion).To(Equal("data.fluid.io/v1alpha1"))
		Expect(updatedDataload.GetOwnerReferences()[0].Kind).To(Equal("Dataset"))
	})
})

var _ = Describe("OperationReconciler ReconcileInternal", func() {
	var (
		reconciler *OperationReconciler
		s          *runtime.Scheme
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		Expect(datav1alpha1.AddToScheme(s)).To(Succeed())
		Expect(corev1.AddToScheme(s)).To(Succeed())
	})

	It("returns error when building operation interface fails", func() {
		obj := &datav1alpha1.DataLoad{ObjectMeta: metav1.ObjectMeta{Name: "load", Namespace: "default"}}
		fakeClient := fakeclient.NewClientBuilder().WithScheme(s).Build()

		reconciler = NewDataOperationReconciler(&mockOperationInterfaceBuilder{
			buildFunc: func(object client.Object) (dataoperation.OperationInterface, error) {
				return nil, fmt.Errorf("build failed")
			},
		}, fakeClient, fake.NullLogger(), record.NewFakeRecorder(10))

		ctx := dataoperation.ReconcileRequestContext{
			ReconcileRequestContext: cruntime.ReconcileRequestContext{Context: context.Background()},
			DataObject:              obj,
		}
		ctx.Log = fake.NullLogger()

		result, err := reconciler.ReconcileInternal(ctx)
		Expect(err).To(HaveOccurred())
		Expect(result).To(Equal(ctrl.Result{}))
	})

	It("requeues when target dataset is not found", func() {
		obj := &datav1alpha1.DataLoad{ObjectMeta: metav1.ObjectMeta{Name: "load", Namespace: "default"}}
		fakeClient := fakeclient.NewClientBuilder().WithScheme(s).Build()

		reconciler = NewDataOperationReconciler(&mockOperationInterfaceBuilder{
			buildFunc: func(object client.Object) (dataoperation.OperationInterface, error) {
				return &mockOperationInterface{
					operationObject:  obj,
					operationType:    dataoperation.DataLoadType,
					targetDatasetErr: apierrors.NewNotFound(schema.GroupResource{Group: "data.fluid.io", Resource: "datasets"}, "missing"),
				}, nil
			},
		}, fakeClient, fake.NullLogger(), record.NewFakeRecorder(10))

		ctx := dataoperation.ReconcileRequestContext{
			ReconcileRequestContext: cruntime.ReconcileRequestContext{Context: context.Background()},
			DataObject:              obj,
		}
		ctx.Log = fake.NullLogger()

		result, err := reconciler.ReconcileInternal(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RequeueAfter).To(BeNumerically(">", 0))
	})

	It("requeues when no accelerate runtime is bound on target dataset", func() {
		obj := &datav1alpha1.DataLoad{ObjectMeta: metav1.ObjectMeta{Name: "load", Namespace: "default"}}
		dataset := &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{Name: "dataset", Namespace: "default"},
		}
		fakeClient := fakeclient.NewClientBuilder().WithScheme(s).WithObjects(dataset).Build()

		reconciler = NewDataOperationReconciler(&mockOperationInterfaceBuilder{
			buildFunc: func(object client.Object) (dataoperation.OperationInterface, error) {
				return &mockOperationInterface{
					operationObject: obj,
					operationType:   dataoperation.DataLoadType,
					targetDataset:   dataset,
				}, nil
			},
		}, fakeClient, fake.NullLogger(), record.NewFakeRecorder(10))

		ctx := dataoperation.ReconcileRequestContext{
			ReconcileRequestContext: cruntime.ReconcileRequestContext{Context: context.Background()},
			DataObject:              obj,
		}
		ctx.Log = fake.NullLogger()

		result, err := reconciler.ReconcileInternal(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RequeueAfter).To(BeNumerically(">", 0))
	})

	It("returns without requeue when bound runtime is not found", func() {
		obj := &datav1alpha1.DataLoad{ObjectMeta: metav1.ObjectMeta{Name: "load", Namespace: "default"}}
		dataset := &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{Name: "dataset", Namespace: "default"},
			Status: datav1alpha1.DatasetStatus{Runtimes: []datav1alpha1.Runtime{{
				Name:      "missing-runtime",
				Namespace: "default",
				Type:      common.AlluxioRuntime,
				Category:  common.AccelerateCategory,
			}}},
		}
		fakeClient := fakeclient.NewClientBuilder().WithScheme(s).WithObjects(dataset).Build()

		reconciler = NewDataOperationReconciler(&mockOperationInterfaceBuilder{
			buildFunc: func(object client.Object) (dataoperation.OperationInterface, error) {
				return &mockOperationInterface{
					operationObject: obj,
					operationType:   dataoperation.DataLoadType,
					targetDataset:   dataset,
				}, nil
			},
		}, fakeClient, fake.NullLogger(), record.NewFakeRecorder(10))

		ctx := dataoperation.ReconcileRequestContext{
			ReconcileRequestContext: cruntime.ReconcileRequestContext{Context: context.Background()},
			DataObject:              obj,
		}
		ctx.Log = fake.NullLogger()

		result, err := reconciler.ReconcileInternal(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(ctrl.Result{}))
	})
})
