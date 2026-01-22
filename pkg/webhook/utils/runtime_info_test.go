package utils

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type mockRuntimeInfo struct {
	name      string
	namespace string
}

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

		It("should return empty for a non-dataset PVC", func() {
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

		It("should error if PVC not found", func() {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

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

		It("should error if runtime info build fails", func() {
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

	Describe("checkDatasetBound", func() {
		It("should succeed if dataset is bound", func() {
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

		It("should error if dataset is not bound", func() {
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

		It("should error if dataset phase is None", func() {
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

		It("should error if dataset has NotReady condition", func() {
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

		It("should error if dataset is not found", func() {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

			err := checkDatasetBound(fakeClient, "nonexistent", "default")
			Expect(err).To(HaveOccurred())
		})

		It("should succeed if dataset is in Pending phase", func() {
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

// Benchmark tests
func BenchmarkCollectRuntimeInfosFromPVCs(b *testing.B) {
	scheme := setupScheme()
	setupLog := zap.New(zap.UseDevMode(false))

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bench-pvc",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(pvc).
		Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = CollectRuntimeInfosFromPVCs(
			fakeClient,
			[]string{"bench-pvc"},
			"default",
			setupLog,
			false,
		)
	}
}

func BenchmarkCheckDatasetBound(b *testing.B) {
	scheme := setupScheme()

	dataset := &v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bench-dataset",
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = checkDatasetBound(fakeClient, "bench-dataset", "default")
	}
}
