package datasetusageinjector

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("DatasetUsageInjector", func() {
	var (
		runtimeInfo1 base.RuntimeInfoInterface
		runtimeInfo2 base.RuntimeInfoInterface
		injector     *DatasetUsageInjector
	)

	BeforeEach(func() {
		var err error
		runtimeInfo1, err = base.BuildRuntimeInfo("demo-dataset-1", "fluid-test", "")
		Expect(err).NotTo(HaveOccurred())

		runtimeInfo2, err = base.BuildRuntimeInfo("demo-dataset-2", "fluid-test", "")
		Expect(err).NotTo(HaveOccurred())

		plugin, err := NewPlugin(fake.NewFakeClient(), "")
		Expect(err).NotTo(HaveOccurred())

		var ok bool
		injector, ok = plugin.(*DatasetUsageInjector)
		Expect(ok).To(BeTrue())
	})

	Describe("Mutate", func() {
		Context("when one dataset is mounted", func() {
			It("should add the dataset annotation", func() {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "fluid-test",
					},
				}

				runtimeInfos := map[string]base.RuntimeInfoInterface{
					"demo-dataset-1": runtimeInfo1,
				}

				shouldStop, err := injector.Mutate(pod, runtimeInfos)
				Expect(err).NotTo(HaveOccurred())
				Expect(shouldStop).To(BeFalse())

				expectedPod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "fluid-test",
						Annotations: map[string]string{
							common.LabelAnnotationDatasetsInUse: "demo-dataset-1",
						},
					},
				}
				Expect(pod).To(Equal(expectedPod))
			})
		})

		Context("when multiple datasets are mounted", func() {
			It("should add comma-separated dataset names in sorted order", func() {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "fluid-test",
					},
				}

				runtimeInfos := map[string]base.RuntimeInfoInterface{
					"demo-dataset-2": runtimeInfo2,
					"demo-dataset-1": runtimeInfo1,
				}

				shouldStop, err := injector.Mutate(pod, runtimeInfos)
				Expect(err).NotTo(HaveOccurred())
				Expect(shouldStop).To(BeFalse())

				expectedPod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "fluid-test",
						Annotations: map[string]string{
							common.LabelAnnotationDatasetsInUse: "demo-dataset-1,demo-dataset-2",
						},
					},
				}
				Expect(pod).To(Equal(expectedPod))
			})
		})

		Context("when no datasets are mounted", func() {
			It("should not add any annotations and return early", func() {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-no-dataset",
						Namespace: "fluid-test",
					},
				}

				runtimeInfos := map[string]base.RuntimeInfoInterface{}

				shouldStop, err := injector.Mutate(pod, runtimeInfos)
				Expect(err).NotTo(HaveOccurred())
				Expect(shouldStop).To(BeFalse())

				expectedPod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-no-dataset",
						Namespace: "fluid-test",
					},
				}
				Expect(pod).To(Equal(expectedPod))
			})
		})

		Context("when pod has GenerateName instead of Name", func() {
			It("should still add the dataset annotation", func() {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: "test-generate-",
						Namespace:    "fluid-test",
					},
				}

				runtimeInfos := map[string]base.RuntimeInfoInterface{
					"demo-dataset-1": runtimeInfo1,
				}

				shouldStop, err := injector.Mutate(pod, runtimeInfos)
				Expect(err).NotTo(HaveOccurred())
				Expect(shouldStop).To(BeFalse())

				Expect(pod.Annotations).To(HaveKeyWithValue(
					common.LabelAnnotationDatasetsInUse,
					"demo-dataset-1",
				))
			})
		})

		Context("when pod already has the same annotation", func() {
			It("should not modify the annotation", func() {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-existing",
						Namespace: "fluid-test",
						Annotations: map[string]string{
							common.LabelAnnotationDatasetsInUse: "demo-dataset-1",
						},
					},
				}

				runtimeInfos := map[string]base.RuntimeInfoInterface{
					"demo-dataset-1": runtimeInfo1,
				}

				shouldStop, err := injector.Mutate(pod, runtimeInfos)
				Expect(err).NotTo(HaveOccurred())
				Expect(shouldStop).To(BeFalse())

				Expect(pod.Annotations[common.LabelAnnotationDatasetsInUse]).To(Equal("demo-dataset-1"))
			})
		})

		Context("when pod has different annotation value", func() {
			It("should update the annotation", func() {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-update",
						Namespace: "fluid-test",
						Annotations: map[string]string{
							common.LabelAnnotationDatasetsInUse: "old-dataset",
						},
					},
				}

				runtimeInfos := map[string]base.RuntimeInfoInterface{
					"demo-dataset-1": runtimeInfo1,
				}

				shouldStop, err := injector.Mutate(pod, runtimeInfos)
				Expect(err).NotTo(HaveOccurred())
				Expect(shouldStop).To(BeFalse())

				Expect(pod.Annotations[common.LabelAnnotationDatasetsInUse]).To(Equal("demo-dataset-1"))
			})
		})
	})

	Describe("GetName", func() {
		It("should return the correct name", func() {
			name := injector.GetName()
			Expect(name).To(Equal(Name))
			Expect(name).To(Equal("DatasetUsageInjector"))
		})
	})
})
