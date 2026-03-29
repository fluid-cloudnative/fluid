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

package dataprocess

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	cdataprocess "github.com/fluid-cloudnative/fluid/pkg/dataprocess"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
)

// newTestDataProcessOperation creates a dataProcessOperation for unit tests.
func newTestDataProcessOperation(s *runtime.Scheme, dp *datav1alpha1.DataProcess) *dataProcessOperation {
	if s == nil {
		s = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(s)
	}
	fakeClient := fake.NewFakeClientWithScheme(s, dp)
	recorder := record.NewFakeRecorder(32)
	return &dataProcessOperation{
		Client:      fakeClient,
		Log:         fake.NullLogger(),
		Recorder:    recorder,
		dataProcess: dp,
	}
}

var _ = Describe("dataProcessOperation", func() {

	Describe("GetOperationObject", func() {
		It("should return the DataProcess object", func() {
			dp := &datav1alpha1.DataProcess{
				ObjectMeta: v1.ObjectMeta{Name: "test", Namespace: "default"},
			}
			op := newTestDataProcessOperation(nil, dp)
			obj := op.GetOperationObject()
			Expect(obj).NotTo(BeNil())
			Expect(obj.GetName()).To(Equal("test"))
			Expect(obj.GetNamespace()).To(Equal("default"))
		})
	})

	Describe("HasPrecedingOperation", func() {
		It("should return false when RunAfter is nil", func() {
			dp := &datav1alpha1.DataProcess{
				ObjectMeta: v1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec:       datav1alpha1.DataProcessSpec{},
			}
			op := newTestDataProcessOperation(nil, dp)
			Expect(op.HasPrecedingOperation()).To(BeFalse())
		})

		It("should return true when RunAfter is set", func() {
			dp := &datav1alpha1.DataProcess{
				ObjectMeta: v1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec: datav1alpha1.DataProcessSpec{
					RunAfter: &datav1alpha1.OperationRef{
						ObjectRef: datav1alpha1.ObjectRef{Name: "prev-op", Kind: "DataLoad"},
					},
				},
			}
			op := newTestDataProcessOperation(nil, dp)
			Expect(op.HasPrecedingOperation()).To(BeTrue())
		})
	})

	Describe("GetPossibleTargetDatasetNamespacedNames", func() {
		It("should return a single namespaced name from spec.dataset", func() {
			dp := &datav1alpha1.DataProcess{
				ObjectMeta: v1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec: datav1alpha1.DataProcessSpec{
					Dataset: datav1alpha1.TargetDatasetWithMountPath{
						TargetDataset: datav1alpha1.TargetDataset{
							Name:      "my-dataset",
							Namespace: "data-ns",
						},
					},
				},
			}
			op := newTestDataProcessOperation(nil, dp)
			names := op.GetPossibleTargetDatasetNamespacedNames()
			Expect(names).To(HaveLen(1))
			Expect(names[0]).To(Equal(types.NamespacedName{Name: "my-dataset", Namespace: "data-ns"}))
		})
	})

	Describe("GetReleaseNameSpacedName", func() {
		It("should return the helm release name in the DataProcess namespace", func() {
			dp := &datav1alpha1.DataProcess{
				ObjectMeta: v1.ObjectMeta{Name: "myproc", Namespace: "default"},
			}
			op := newTestDataProcessOperation(nil, dp)
			nn := op.GetReleaseNameSpacedName()
			Expect(nn.Namespace).To(Equal("default"))
			Expect(nn.Name).To(Equal(utils.GetDataProcessReleaseName("myproc")))
		})
	})

	Describe("GetChartsDirectory", func() {
		It("should return a path containing the DataProcess chart name", func() {
			dp := &datav1alpha1.DataProcess{
				ObjectMeta: v1.ObjectMeta{Name: "test", Namespace: "default"},
			}
			op := newTestDataProcessOperation(nil, dp)
			dir := op.GetChartsDirectory()
			Expect(dir).To(ContainSubstring(cdataprocess.DataProcessChart))
		})
	})

	Describe("GetOperationType", func() {
		It("should return DataProcessType", func() {
			dp := &datav1alpha1.DataProcess{
				ObjectMeta: v1.ObjectMeta{Name: "test", Namespace: "default"},
			}
			op := newTestDataProcessOperation(nil, dp)
			Expect(op.GetOperationType()).To(Equal(dataoperation.DataProcessType))
		})
	})

	Describe("GetStatusHandler", func() {
		It("should return a non-nil OnceStatusHandler", func() {
			dp := &datav1alpha1.DataProcess{
				ObjectMeta: v1.ObjectMeta{Name: "test", Namespace: "default"},
			}
			op := newTestDataProcessOperation(nil, dp)
			handler := op.GetStatusHandler()
			Expect(handler).NotTo(BeNil())
			_, ok := handler.(*OnceStatusHandler)
			Expect(ok).To(BeTrue())
		})
	})

	Describe("GetTTL", func() {
		It("should return nil TTL when not set", func() {
			dp := &datav1alpha1.DataProcess{
				ObjectMeta: v1.ObjectMeta{Name: "test", Namespace: "default"},
			}
			op := newTestDataProcessOperation(nil, dp)
			ttl, err := op.GetTTL()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(BeNil())
		})

		It("should return the configured TTL when set", func() {
			var ttlVal int32 = 300
			dp := &datav1alpha1.DataProcess{
				ObjectMeta: v1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec: datav1alpha1.DataProcessSpec{
					TTLSecondsAfterFinished: &ttlVal,
				},
			}
			op := newTestDataProcessOperation(nil, dp)
			ttl, err := op.GetTTL()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).NotTo(BeNil())
			Expect(*ttl).To(Equal(int32(300)))
		})
	})

	Describe("GetParallelTaskNumber", func() {
		It("should always return 1", func() {
			dp := &datav1alpha1.DataProcess{
				ObjectMeta: v1.ObjectMeta{Name: "test", Namespace: "default"},
			}
			op := newTestDataProcessOperation(nil, dp)
			Expect(op.GetParallelTaskNumber()).To(Equal(int32(1)))
		})
	})

	Describe("UpdateStatusInfoForCompleted", func() {
		It("should return nil (no-op)", func() {
			dp := &datav1alpha1.DataProcess{
				ObjectMeta: v1.ObjectMeta{Name: "test", Namespace: "default"},
			}
			op := newTestDataProcessOperation(nil, dp)
			err := op.UpdateStatusInfoForCompleted(map[string]string{"key": "val"})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("SetTargetDatasetStatusInProgress", func() {
		It("should not panic (no-op)", func() {
			dp := &datav1alpha1.DataProcess{
				ObjectMeta: v1.ObjectMeta{Name: "test", Namespace: "default"},
			}
			op := newTestDataProcessOperation(nil, dp)
			dataset := &datav1alpha1.Dataset{}
			Expect(func() { op.SetTargetDatasetStatusInProgress(dataset) }).NotTo(Panic())
		})
	})

	Describe("RemoveTargetDatasetStatusInProgress", func() {
		It("should not panic (no-op)", func() {
			dp := &datav1alpha1.DataProcess{
				ObjectMeta: v1.ObjectMeta{Name: "test", Namespace: "default"},
			}
			op := newTestDataProcessOperation(nil, dp)
			dataset := &datav1alpha1.Dataset{}
			Expect(func() { op.RemoveTargetDatasetStatusInProgress(dataset) }).NotTo(Panic())
		})
	})

	Describe("Validate", func() {
		var (
			testScheme *runtime.Scheme
			ctx        cruntime.ReconcileRequestContext
		)

		BeforeEach(func() {
			testScheme = runtime.NewScheme()
			_ = datav1alpha1.AddToScheme(testScheme)
			ctx = cruntime.ReconcileRequestContext{
				Log: fake.NullLogger(),
			}
		})

		Context("when DataProcess namespace differs from spec.dataset.namespace", func() {
			It("should return a condition and error about namespace mismatch", func() {
				dp := &datav1alpha1.DataProcess{
					ObjectMeta: v1.ObjectMeta{Name: "test", Namespace: "default"},
					Spec: datav1alpha1.DataProcessSpec{
						Dataset: datav1alpha1.TargetDatasetWithMountPath{
							TargetDataset: datav1alpha1.TargetDataset{
								Name:      "ds",
								Namespace: "other-namespace", // mismatch
							},
						},
						Processor: datav1alpha1.Processor{
							Job: &datav1alpha1.JobProcessor{},
						},
					},
				}
				op := newTestDataProcessOperation(testScheme, dp)
				conditions, err := op.Validate(ctx)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("namespace"))
				Expect(conditions).To(HaveLen(1))
				Expect(string(conditions[0].Reason)).To(Equal("TargetDatasetNamespaceNotSame"))
			})
		})

		Context("when no processor is set (both Job and Script are nil)", func() {
			It("should return a condition and error about missing processor", func() {
				dp := &datav1alpha1.DataProcess{
					ObjectMeta: v1.ObjectMeta{Name: "test", Namespace: "default"},
					Spec: datav1alpha1.DataProcessSpec{
						Dataset: datav1alpha1.TargetDatasetWithMountPath{
							TargetDataset: datav1alpha1.TargetDataset{
								Name:      "ds",
								Namespace: "default", // same namespace
							},
						},
						Processor: datav1alpha1.Processor{
							Job:    nil,
							Script: nil,
						},
					},
				}
				op := newTestDataProcessOperation(testScheme, dp)
				conditions, err := op.Validate(ctx)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("spec.processor"))
				Expect(conditions).To(HaveLen(1))
				Expect(string(conditions[0].Reason)).To(Equal("ProcessorNotSpecified"))
			})
		})

		Context("when both Job and Script processors are set", func() {
			It("should return a condition and error about multiple processors", func() {
				dp := &datav1alpha1.DataProcess{
					ObjectMeta: v1.ObjectMeta{Name: "test", Namespace: "default"},
					Spec: datav1alpha1.DataProcessSpec{
						Dataset: datav1alpha1.TargetDatasetWithMountPath{
							TargetDataset: datav1alpha1.TargetDataset{
								Name:      "ds",
								Namespace: "default",
							},
						},
						Processor: datav1alpha1.Processor{
							Job:    &datav1alpha1.JobProcessor{},
							Script: &datav1alpha1.ScriptProcessor{},
						},
					},
				}
				op := newTestDataProcessOperation(testScheme, dp)
				conditions, err := op.Validate(ctx)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("multiple processors"))
				Expect(conditions).To(HaveLen(1))
				Expect(string(conditions[0].Reason)).To(Equal("MultipleProcessorSpecified"))
			})
		})

		Context("when script processor has a conflicting mountPath with dataset mountPath", func() {
			It("should return a condition and error about conflicting mount path", func() {
				conflictPath := "/data"
				dp := &datav1alpha1.DataProcess{
					ObjectMeta: v1.ObjectMeta{Name: "test", Namespace: "default"},
					Spec: datav1alpha1.DataProcessSpec{
						Dataset: datav1alpha1.TargetDatasetWithMountPath{
							TargetDataset: datav1alpha1.TargetDataset{
								Name:      "ds",
								Namespace: "default",
							},
							MountPath: conflictPath,
						},
						Processor: datav1alpha1.Processor{
							Script: &datav1alpha1.ScriptProcessor{
								Source: "echo hello",
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "conflict-vol",
										MountPath: conflictPath, // same path as dataset mountPath
									},
								},
							},
						},
					},
				}
				op := newTestDataProcessOperation(testScheme, dp)
				conditions, err := op.Validate(ctx)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("conflict"))
				Expect(conditions).To(HaveLen(1))
				Expect(string(conditions[0].Reason)).To(Equal("ConflictMountPath"))
			})
		})

		Context("when DataProcess has a valid script processor with no conflicts", func() {
			It("should return no conditions and no error", func() {
				dp := &datav1alpha1.DataProcess{
					ObjectMeta: v1.ObjectMeta{Name: "test", Namespace: "default"},
					Spec: datav1alpha1.DataProcessSpec{
						Dataset: datav1alpha1.TargetDatasetWithMountPath{
							TargetDataset: datav1alpha1.TargetDataset{
								Name:      "ds",
								Namespace: "default",
							},
							MountPath: "/data",
						},
						Processor: datav1alpha1.Processor{
							Script: &datav1alpha1.ScriptProcessor{
								Source: "echo hello",
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "other-vol",
										MountPath: "/other", // no conflict
									},
								},
							},
						},
					},
				}
				op := newTestDataProcessOperation(testScheme, dp)
				conditions, err := op.Validate(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(conditions).To(BeNil())
			})
		})
	})
})
