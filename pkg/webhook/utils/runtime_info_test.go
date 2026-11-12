package utils

import (
	"strings"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func setupScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)
	return scheme
}

var _ = Describe("RuntimeInfo Utilities", func() {
	var (
		scheme   *runtime.Scheme
		setupLog = zap.New(zap.UseDevMode(true))
	)

	BeforeEach(func() {
		scheme = setupScheme()
	})

	Describe("CollectRuntimeInfosFromPVCs", func() {

		Context("when PVC is not a dataset PVC", func() {
			It("should return empty runtime infos", func() {
				pvc := &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "regular-pvc",
						Namespace: "default",
						Labels:    map[string]string{},
					},
				}

				fakeClient := fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(pvc).
					Build()

				runtimeInfos, err := CollectRuntimeInfosFromPVCs(
					fakeClient,
					[]string{"regular-pvc"},
					"default",
					setupLog,
					false,
				)

				Expect(err).NotTo(HaveOccurred())
				Expect(runtimeInfos).To(HaveLen(0))
			})
		})

		Context("when PVC does not exist", func() {
			It("should return an error", func() {
				fakeClient := fake.NewClientBuilder().
					WithScheme(scheme).
					Build()

				runtimeInfos, err := CollectRuntimeInfosFromPVCs(
					fakeClient,
					[]string{"nonexistent"},
					"default",
					setupLog,
					false,
				)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to get the following PVCs"))
				Expect(runtimeInfos).To(HaveLen(0))
			})
		})

		Context("when runtime info build fails", func() {
			It("should return an error", func() {
				pvc := &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pvc",
						Namespace: "default",
						Labels: map[string]string{
							common.LabelAnnotationStorageCapacityPrefix: "true",
						},
					},
				}

				dataset := &v1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pvc",
						Namespace: "default",
					},
					Status: v1alpha1.DatasetStatus{
						Phase: v1alpha1.NotBoundDatasetPhase,
					},
				}

				fakeClient := fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(pvc, dataset).
					Build()

				_, err := CollectRuntimeInfosFromPVCs(
					fakeClient,
					[]string{"test-pvc"},
					"default",
					setupLog,
					false,
				)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to build runtime info"))
			})
		})
	})

	Describe("checkDatasetBound", func() {

		Context("when dataset is bound", func() {
			It("should succeed", func() {
				dataset := &v1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-dataset",
						Namespace: "default",
					},
					Status: v1alpha1.DatasetStatus{
						Phase: v1alpha1.BoundDatasetPhase,
					},
				}

				fakeClient := fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(dataset).
					Build()

				err := checkDatasetBound(fakeClient, "test-dataset", "default")
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when dataset is not bound", func() {
			It("should return an error", func() {
				dataset := &v1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-dataset",
						Namespace: "default",
					},
					Status: v1alpha1.DatasetStatus{
						Phase: v1alpha1.NotBoundDatasetPhase,
					},
				}

				fakeClient := fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(dataset).
					Build()

				err := checkDatasetBound(fakeClient, "test-dataset", "default")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not bound"))
			})
		})

		Context("when dataset phase is None", func() {
			It("should return an error", func() {
				dataset := &v1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-dataset",
						Namespace: "default",
					},
					Status: v1alpha1.DatasetStatus{
						Phase: v1alpha1.NoneDatasetPhase,
					},
				}

				fakeClient := fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(dataset).
					Build()

				err := checkDatasetBound(fakeClient, "test-dataset", "default")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not bound"))
			})
		})

		Context("when dataset has NotReady condition", func() {
			It("should return a descriptive error", func() {
				dataset := &v1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-dataset",
						Namespace: "default",
					},
					Status: v1alpha1.DatasetStatus{
						Phase: v1alpha1.NotBoundDatasetPhase,
						Conditions: []v1alpha1.DatasetCondition{
							{
								Type:    v1alpha1.DatasetNotReady,
								Status:  corev1.ConditionTrue,
								Message: "Dataset is not ready due to initialization",
							},
						},
					},
				}

				fakeClient := fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(dataset).
					Build()

				err := checkDatasetBound(fakeClient, "test-dataset", "default")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not ready because"))
				Expect(err.Error()).To(ContainSubstring("initialization"))
			})
		})

		Context("when dataset does not exist", func() {
			It("should return an error", func() {
				fakeClient := fake.NewClientBuilder().
					WithScheme(scheme).
					Build()

				err := checkDatasetBound(fakeClient, "nonexistent", "default")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when dataset is in Pending phase", func() {
			It("should succeed", func() {
				dataset := &v1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-dataset",
						Namespace: "default",
					},
					Status: v1alpha1.DatasetStatus{
						Phase: v1alpha1.PendingDatasetPhase,
					},
				}

				fakeClient := fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(dataset).
					Build()

				err := checkDatasetBound(fakeClient, "test-dataset", "default")
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})

var _ = Describe("CollectRuntimeInfosFromAnnotations", func() {
	const namespace = "default"

	var (
		scheme   *runtime.Scheme
		setupLog = zap.New(zap.UseDevMode(true))
	)

	// cacheDataset builds a dataset backed by a cache runtime of the same name.
	cacheDataset := func(name string, phase v1alpha1.DatasetPhase) *v1alpha1.Dataset {
		return &v1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
			Status: v1alpha1.DatasetStatus{
				Phase: phase,
				Runtimes: []v1alpha1.Runtime{
					{Name: name, Namespace: namespace, Type: common.CacheRuntime},
				},
			},
		}
	}

	cacheRuntime := func(name string) *v1alpha1.CacheRuntime {
		return &v1alpha1.CacheRuntime{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		}
	}

	collect := func(annotation string, skipPrecheck bool, objs ...client.Object) (map[string]base.RuntimeInfoInterface, error) {
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
		return CollectRuntimeInfosFromAnnotations(fakeClient,
			map[string]string{common.LabelAnnotationDatasets: annotation},
			namespace, setupLog, skipPrecheck)
	}

	BeforeEach(func() {
		scheme = setupScheme()
	})

	Context("when the annotation is absent", func() {
		It("should return no runtime infos", func() {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
			runtimeInfos, err := CollectRuntimeInfosFromAnnotations(fakeClient, nil, namespace, setupLog, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(runtimeInfos).To(BeEmpty())
		})
	})

	Context("when a bound dataset is declared", func() {
		It("should key its runtime info by the dataset name", func() {
			runtimeInfos, err := collect("demo", false,
				cacheDataset("demo", v1alpha1.BoundDatasetPhase), cacheRuntime("demo"))
			Expect(err).NotTo(HaveOccurred())
			Expect(runtimeInfos).To(HaveLen(1))
			Expect(runtimeInfos).To(HaveKey("demo"))
			Expect(runtimeInfos["demo"].GetName()).To(Equal("demo"))
			Expect(runtimeInfos["demo"].GetNamespace()).To(Equal(namespace))
			Expect(runtimeInfos["demo"].GetRuntimeType()).To(Equal(common.CacheRuntime))
		})
	})

	Context("when the annotation lists several datasets loosely", func() {
		It("should trim spaces, skip empty entries and drop duplicates", func() {
			runtimeInfos, err := collect(" demo , ,other,demo,", false,
				cacheDataset("demo", v1alpha1.BoundDatasetPhase), cacheRuntime("demo"),
				cacheDataset("other", v1alpha1.BoundDatasetPhase), cacheRuntime("other"))
			Expect(err).NotTo(HaveOccurred())
			Expect(runtimeInfos).To(HaveLen(2))
			Expect(runtimeInfos).To(HaveKey("demo"))
			Expect(runtimeInfos).To(HaveKey("other"))
		})
	})

	Context("when a dataset name is not a valid object name", func() {
		It("should name the annotation and the dataset in the error", func() {
			_, err := collect("demo.v1", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("demo.v1"))
			Expect(err.Error()).To(ContainSubstring(common.LabelAnnotationDatasets))
			Expect(err.Error()).To(ContainSubstring("a DNS-1035 label must consist of"))
		})
	})

	Context("when a dataset name leaves no room for the config map prefix", func() {
		It("should reject it even though the dataset is bound", func() {
			longName := strings.Repeat("a", common.MaxClientlessDatasetNameLength+1)
			_, err := collect(longName, false,
				cacheDataset(longName, v1alpha1.BoundDatasetPhase), cacheRuntime(longName))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(longName))
			// The denial carries the reason, not only the name.
			Expect(err.Error()).To(ContainSubstring("dataset name is too long"))
		})

		It("should accept the longest name that still fits", func() {
			name := strings.Repeat("a", common.MaxClientlessDatasetNameLength)
			runtimeInfos, err := collect(name, false,
				cacheDataset(name, v1alpha1.BoundDatasetPhase), cacheRuntime(name))
			Expect(err).NotTo(HaveOccurred())
			Expect(runtimeInfos).To(HaveKey(name))
		})
	})

	Context("when a declared dataset does not exist", func() {
		It("should return an error naming the dataset", func() {
			_, err := collect("missing", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("missing"))
		})
	})

	Context("when a declared dataset is not bound", func() {
		It("should return an error", func() {
			_, err := collect("demo", false,
				cacheDataset("demo", v1alpha1.NotBoundDatasetPhase), cacheRuntime("demo"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("demo"))
		})

		It("should resolve it anyway when the precheck is skipped", func() {
			runtimeInfos, err := collect("demo", true,
				cacheDataset("demo", v1alpha1.NotBoundDatasetPhase), cacheRuntime("demo"))
			Expect(err).NotTo(HaveOccurred())
			Expect(runtimeInfos).To(HaveKey("demo"))
		})
	})

	Context("when a bound dataset has no runtime yet", func() {
		It("should return an error", func() {
			_, err := collect("demo", false, cacheDataset("demo", v1alpha1.BoundDatasetPhase))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("demo"))
		})
	})

	Context("when only some of the declared datasets resolve", func() {
		It("should report only the failed ones", func() {
			_, err := collect("demo,missing", false,
				cacheDataset("demo", v1alpha1.BoundDatasetPhase), cacheRuntime("demo"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("[missing]"))
		})
	})
})
