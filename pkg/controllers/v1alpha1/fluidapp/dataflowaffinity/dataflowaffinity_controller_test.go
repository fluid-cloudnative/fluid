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

package dataflowaffinity

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func newTestScheme() *runtime.Scheme {
	testScheme := runtime.NewScheme()
	_ = v1.AddToScheme(testScheme)
	_ = batchv1.AddToScheme(testScheme)
	_ = datav1alpha1.AddToScheme(testScheme)
	return testScheme
}

var _ = Describe("DataOpJobReconciler", func() {
	Describe("ControllerName", func() {
		It("should return DataOpJobController", func() {
			reconciler := &DataOpJobReconciler{}
			Expect(reconciler.ControllerName()).To(Equal("DataOpJobController"))
		})
	})

	Describe("ManagedResource", func() {
		It("should return a Job object", func() {
			reconciler := &DataOpJobReconciler{}
			obj := reconciler.ManagedResource()
			Expect(obj).NotTo(BeNil())
			_, ok := obj.(*batchv1.Job)
			Expect(ok).To(BeTrue())
		})
	})

	Describe("NewDataOpJobReconciler", func() {
		It("should create a new reconciler", func() {
			reconciler := NewDataOpJobReconciler(nil, fake.NullLogger(), nil)
			Expect(reconciler).NotTo(BeNil())
			Expect(reconciler.Log).NotTo(BeNil())
		})
	})

	Describe("fillCustomizedNodeAffinity", func() {
		It("should fill annotations with node labels", func() {
			annotationsToInject := map[string]string{}
			nodeLabels := map[string]string{
				common.K8sNodeNameLabelKey: "node01",
				common.K8sRegionLabelKey:   "region01",
				common.K8sZoneLabelKey:     "zone01",
				"custom-label":             "custom-value",
			}
			exposedLabelNames := []string{
				common.K8sNodeNameLabelKey,
				common.K8sRegionLabelKey,
				common.K8sZoneLabelKey,
				"custom-label",
			}

			fillCustomizedNodeAffinity(annotationsToInject, nodeLabels, exposedLabelNames)

			Expect(annotationsToInject).To(HaveLen(4))
			Expect(annotationsToInject[common.AnnotationDataFlowCustomizedAffinityPrefix+common.K8sNodeNameLabelKey]).To(Equal("node01"))
			Expect(annotationsToInject[common.AnnotationDataFlowCustomizedAffinityPrefix+common.K8sRegionLabelKey]).To(Equal("region01"))
			Expect(annotationsToInject[common.AnnotationDataFlowCustomizedAffinityPrefix+common.K8sZoneLabelKey]).To(Equal("zone01"))
			Expect(annotationsToInject[common.AnnotationDataFlowCustomizedAffinityPrefix+"custom-label"]).To(Equal("custom-value"))
		})

		It("should skip non-existent labels", func() {
			annotationsToInject := map[string]string{}
			nodeLabels := map[string]string{
				common.K8sNodeNameLabelKey: "node01",
			}
			exposedLabelNames := []string{
				common.K8sNodeNameLabelKey,
				"non-existent-label",
			}

			fillCustomizedNodeAffinity(annotationsToInject, nodeLabels, exposedLabelNames)

			Expect(annotationsToInject).To(HaveLen(1))
			Expect(annotationsToInject[common.AnnotationDataFlowCustomizedAffinityPrefix+common.K8sNodeNameLabelKey]).To(Equal("node01"))
		})

		It("should handle labels with whitespace", func() {
			annotationsToInject := map[string]string{}
			nodeLabels := map[string]string{
				common.K8sNodeNameLabelKey: "node01",
			}
			exposedLabelNames := []string{
				" " + common.K8sNodeNameLabelKey + " ",
			}

			fillCustomizedNodeAffinity(annotationsToInject, nodeLabels, exposedLabelNames)

			Expect(annotationsToInject).To(HaveLen(1))
			Expect(annotationsToInject[common.AnnotationDataFlowCustomizedAffinityPrefix+common.K8sNodeNameLabelKey]).To(Equal("node01"))
		})
	})

	Describe("injectPodNodeLabelsToJob", func() {
		var testScheme *runtime.Scheme

		BeforeEach(func() {
			testScheme = newTestScheme()
		})

		Context("when job has succeeded pods", func() {
			It("should inject pod node labels to job annotations", func() {
				job := &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-job",
						Labels: map[string]string{
							common.LabelAnnotationManagedBy: common.Fluid,
						},
					},
					Spec: batchv1.JobSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"controller-uid": "455afc34-93b1-4e75-a6fa-8e13d2c6ca06",
							},
						},
					},
				}
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-pod",
						Labels: map[string]string{
							"controller-uid": "455afc34-93b1-4e75-a6fa-8e13d2c6ca06",
						},
						Annotations: map[string]string{
							common.AnnotationDataFlowAffinityLabelsName: "k8s.gpu,,",
						},
					},
					Spec: v1.PodSpec{
						NodeName: "node01",
						Affinity: &v1.Affinity{
							NodeAffinity: &v1.NodeAffinity{
								RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
									NodeSelectorTerms: []v1.NodeSelectorTerm{
										{
											MatchExpressions: []v1.NodeSelectorRequirement{
												{
													Key:      "k8s.gpu",
													Operator: v1.NodeSelectorOpIn,
													Values:   []string{"true"},
												},
											},
										},
									},
								},
							},
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodSucceeded,
					},
				}
				node := &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node01",
						Labels: map[string]string{
							common.K8sNodeNameLabelKey: "node01",
							common.K8sRegionLabelKey:   "region01",
							common.K8sZoneLabelKey:     "zone01",
							"k8s.gpu":                  "true",
						},
					},
				}

				c := fake.NewFakeClientWithScheme(testScheme, job, pod, node)
				reconciler := &DataOpJobReconciler{
					Client: c,
					Log:    fake.NullLogger(),
				}

				err := reconciler.injectPodNodeLabelsToJob(job)
				Expect(err).NotTo(HaveOccurred())

				expectedAnnotations := map[string]string{
					common.AnnotationDataFlowCustomizedAffinityPrefix + common.K8sNodeNameLabelKey: "node01",
					common.AnnotationDataFlowCustomizedAffinityPrefix + common.K8sRegionLabelKey:   "region01",
					common.AnnotationDataFlowCustomizedAffinityPrefix + common.K8sZoneLabelKey:     "zone01",
					common.AnnotationDataFlowCustomizedAffinityPrefix + "k8s.gpu":                  "true",
				}
				Expect(job.Annotations).To(Equal(expectedAnnotations))
			})
		})

		Context("when job has failed pods", func() {
			It("should return an error", func() {
				job := &batchv1.Job{
					Spec: batchv1.JobSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"controller-uid": "455afc34-93b1-4e75-a6fa-8e13d2c6ca06",
							},
						},
					},
				}
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-pod",
						Labels: map[string]string{
							"controller-uid": "455afc34-93b1-4e75-a6fa-8e13d2c6ca06",
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodFailed,
					},
				}
				node := &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node01",
						Labels: map[string]string{
							common.K8sNodeNameLabelKey: "node01",
							common.K8sRegionLabelKey:   "region01",
							common.K8sZoneLabelKey:     "zone01",
							"k8s.gpu":                  "true",
						},
					},
				}

				c := fake.NewFakeClientWithScheme(testScheme, job, pod, node)
				reconciler := &DataOpJobReconciler{
					Client: c,
					Log:    fake.NullLogger(),
				}

				err := reconciler.injectPodNodeLabelsToJob(job)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when pod has no node name", func() {
			It("should return an error", func() {
				job := &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-job",
						Labels: map[string]string{
							common.LabelAnnotationManagedBy: common.Fluid,
						},
					},
					Spec: batchv1.JobSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"controller-uid": "test-uid",
							},
						},
					},
				}
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-pod",
						Labels: map[string]string{
							"controller-uid": "test-uid",
						},
					},
					Spec: v1.PodSpec{
						NodeName: "",
					},
					Status: v1.PodStatus{
						Phase: v1.PodSucceeded,
					},
				}

				c := fake.NewFakeClientWithScheme(testScheme, job, pod)
				reconciler := &DataOpJobReconciler{
					Client: c,
					Log:    fake.NullLogger(),
				}

				err := reconciler.injectPodNodeLabelsToJob(job)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no node name"))
			})
		})

		Context("when job has nil annotations", func() {
			It("should create annotations map and inject labels", func() {
				job := &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-job",
						Labels: map[string]string{
							common.LabelAnnotationManagedBy: common.Fluid,
						},
						Annotations: nil,
					},
					Spec: batchv1.JobSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"controller-uid": "test-uid",
							},
						},
					},
				}
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-pod",
						Labels: map[string]string{
							"controller-uid": "test-uid",
						},
					},
					Spec: v1.PodSpec{
						NodeName: "node01",
					},
					Status: v1.PodStatus{
						Phase: v1.PodSucceeded,
					},
				}
				node := &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node01",
						Labels: map[string]string{
							common.K8sNodeNameLabelKey: "node01",
						},
					},
				}

				c := fake.NewFakeClientWithScheme(testScheme, job, pod, node)
				reconciler := &DataOpJobReconciler{
					Client: c,
					Log:    fake.NullLogger(),
				}

				err := reconciler.injectPodNodeLabelsToJob(job)
				Expect(err).NotTo(HaveOccurred())
				Expect(job.Annotations).NotTo(BeNil())
			})
		})
	})
})
