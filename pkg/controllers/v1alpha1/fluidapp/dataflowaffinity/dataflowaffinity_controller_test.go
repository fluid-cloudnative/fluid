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
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

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
	const (
		controllerUIDKey = "controller-uid"
		testNodeName     = "node01"
		testPodName      = "test-pod"
	)

	var testScheme *runtime.Scheme

	BeforeEach(func() {
		testScheme = newTestScheme()
	})

	Describe("ControllerName", func() {
		It("returns the controller name constant", func() {
			reconciler := &DataOpJobReconciler{Log: fake.NullLogger()}
			Expect(reconciler.ControllerName()).To(Equal(DataOpJobControllerName))
		})
	})

	Describe("ManagedResource", func() {
		It("returns a batchv1.Job object", func() {
			reconciler := &DataOpJobReconciler{Log: fake.NullLogger()}
			obj := reconciler.ManagedResource()
			Expect(obj).To(BeAssignableToTypeOf(&batchv1.Job{}))
		})
	})

	Describe("NewDataOpJobReconciler", func() {
		It("constructs a reconciler with the given client, logger, and recorder", func() {
			c := fake.NewFakeClientWithScheme(testScheme)
			logger := fake.NullLogger()
			r := NewDataOpJobReconciler(c, logger, nil)
			Expect(r).NotTo(BeNil())
			Expect(r.Client).To(Equal(c))
		})
	})

	Describe("Reconcile", func() {
		Context("when the job does not exist", func() {
			It("returns an error when the job is missing", func() {
				c := fake.NewFakeClientWithScheme(testScheme)
				reconciler := &DataOpJobReconciler{Client: c, Log: fake.NullLogger()}
				_, err := reconciler.Reconcile(context.Background(), reconcile.Request{
					NamespacedName: types.NamespacedName{Name: "missing-job", Namespace: "default"},
				})
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when job should not be in queue", func() {
			It("returns no requeue without error for cron jobs", func() {
				job := &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cron-job",
						Namespace: "default",
						Labels: map[string]string{
							common.LabelAnnotationManagedBy: common.Fluid,
							"cronjob":                       "something",
						},
					},
				}
				c := fake.NewFakeClientWithScheme(testScheme, job)
				reconciler := &DataOpJobReconciler{Client: c, Log: fake.NullLogger()}
				result, err := reconciler.Reconcile(context.Background(), reconcile.Request{
					NamespacedName: types.NamespacedName{Name: "cron-job", Namespace: "default"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())
			})
		})

		Context("when job is a valid fluid job without affinity annotation", func() {
			It("injects the dataflow affinity annotation and returns no requeue", func() {
				const testJobName = "test-job"

				job := &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testJobName,
						Namespace: "default",
						Labels: map[string]string{
							common.LabelAnnotationManagedBy: common.Fluid,
						},
					},
				}
				c := fake.NewFakeClientWithScheme(testScheme, job)
				reconciler := &DataOpJobReconciler{Client: c, Log: fake.NullLogger()}
				result, err := reconciler.Reconcile(context.Background(), reconcile.Request{
					NamespacedName: types.NamespacedName{Name: testJobName, Namespace: "default"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())

				updatedJob := &batchv1.Job{}
				Expect(c.Get(context.Background(), types.NamespacedName{Name: testJobName, Namespace: "default"}, updatedJob)).To(Succeed())
				Expect(updatedJob.Annotations).To(HaveKeyWithValue(common.AnnotationDataFlowAffinityInject, "true"))
			})
		})

		Context("when job is complete and has a succeeded pod", func() {
			It("injects node labels and returns no requeue", func() {
				const completeJobName = "complete-job"

				job := &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:      completeJobName,
						Namespace: "default",
						Labels: map[string]string{
							common.LabelAnnotationManagedBy: common.Fluid,
						},
						Annotations: map[string]string{
							common.AnnotationDataFlowAffinityInject: "true",
						},
					},
					Spec: batchv1.JobSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								controllerUIDKey: "abc-123",
							},
						},
					},
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{{Type: batchv1.JobComplete}},
					},
				}
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "complete-pod",
						Namespace: "default",
						Labels: map[string]string{
							controllerUIDKey: "abc-123",
						},
					},
					Spec:   v1.PodSpec{NodeName: testNodeName},
					Status: v1.PodStatus{Phase: v1.PodSucceeded},
				}
				node := &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: testNodeName,
						Labels: map[string]string{
							common.K8sNodeNameLabelKey: testNodeName,
							common.K8sRegionLabelKey:   "region01",
							common.K8sZoneLabelKey:     "zone01",
						},
					},
				}
				c := fake.NewFakeClientWithScheme(testScheme, job, pod, node)
				reconciler := &DataOpJobReconciler{Client: c, Log: fake.NullLogger()}
				result, err := reconciler.Reconcile(context.Background(), reconcile.Request{
					NamespacedName: types.NamespacedName{Name: completeJobName, Namespace: "default"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())

				updatedJob := &batchv1.Job{}
				Expect(c.Get(context.Background(), types.NamespacedName{Name: completeJobName, Namespace: "default"}, updatedJob)).To(Succeed())
				Expect(updatedJob.Annotations).To(HaveKeyWithValue(common.AnnotationDataFlowCustomizedAffinityPrefix+common.K8sNodeNameLabelKey, testNodeName))
				Expect(updatedJob.Annotations).To(HaveKeyWithValue(common.AnnotationDataFlowCustomizedAffinityPrefix+common.K8sRegionLabelKey, "region01"))
				Expect(updatedJob.Annotations).To(HaveKeyWithValue(common.AnnotationDataFlowCustomizedAffinityPrefix+common.K8sZoneLabelKey, "zone01"))
			})
		})
	})

	Describe("fillCustomizedNodeAffinity", func() {
		It("fills annotations with node labels", func() {
			annotationsToInject := map[string]string{}
			nodeLabels := map[string]string{
				common.K8sNodeNameLabelKey: testNodeName,
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
			Expect(annotationsToInject[common.AnnotationDataFlowCustomizedAffinityPrefix+common.K8sNodeNameLabelKey]).To(Equal(testNodeName))
			Expect(annotationsToInject[common.AnnotationDataFlowCustomizedAffinityPrefix+common.K8sRegionLabelKey]).To(Equal("region01"))
			Expect(annotationsToInject[common.AnnotationDataFlowCustomizedAffinityPrefix+common.K8sZoneLabelKey]).To(Equal("zone01"))
			Expect(annotationsToInject[common.AnnotationDataFlowCustomizedAffinityPrefix+"custom-label"]).To(Equal("custom-value"))
		})

		It("skips non-existent labels", func() {
			annotationsToInject := map[string]string{}
			nodeLabels := map[string]string{
				common.K8sNodeNameLabelKey: testNodeName,
			}
			exposedLabelNames := []string{
				common.K8sNodeNameLabelKey,
				"non-existent-label",
			}

			fillCustomizedNodeAffinity(annotationsToInject, nodeLabels, exposedLabelNames)

			Expect(annotationsToInject).To(HaveLen(1))
			Expect(annotationsToInject[common.AnnotationDataFlowCustomizedAffinityPrefix+common.K8sNodeNameLabelKey]).To(Equal(testNodeName))
		})

		It("handles labels with whitespace", func() {
			annotationsToInject := map[string]string{}
			nodeLabels := map[string]string{
				common.K8sNodeNameLabelKey: testNodeName,
			}
			exposedLabelNames := []string{" " + common.K8sNodeNameLabelKey + " "}

			fillCustomizedNodeAffinity(annotationsToInject, nodeLabels, exposedLabelNames)

			Expect(annotationsToInject).To(HaveLen(1))
			Expect(annotationsToInject[common.AnnotationDataFlowCustomizedAffinityPrefix+common.K8sNodeNameLabelKey]).To(Equal(testNodeName))
		})
	})

	Describe("injectPodNodeLabelsToJob", func() {
		const jobControllerUIDValue = "455afc34-93b1-4e75-a6fa-8e13d2c6ca06"

		Context("when job has a succeeded pod", func() {
			It("injects node labels as annotations onto the job", func() {
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
								controllerUIDKey: jobControllerUIDValue,
							},
						},
					},
				}
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: testPodName,
						Labels: map[string]string{
							controllerUIDKey: jobControllerUIDValue,
						},
						Annotations: map[string]string{
							common.AnnotationDataFlowAffinityLabelsName: "k8s.gpu,,",
						},
					},
					Spec: v1.PodSpec{
						NodeName: testNodeName,
						Affinity: &v1.Affinity{
							NodeAffinity: &v1.NodeAffinity{
								RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
									NodeSelectorTerms: []v1.NodeSelectorTerm{{
										MatchExpressions: []v1.NodeSelectorRequirement{{
											Key:      "k8s.gpu",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"true"},
										}},
									}},
								},
							},
						},
					},
					Status: v1.PodStatus{Phase: v1.PodSucceeded},
				}
				node := &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: testNodeName,
						Labels: map[string]string{
							common.K8sNodeNameLabelKey: testNodeName,
							common.K8sRegionLabelKey:   "region01",
							common.K8sZoneLabelKey:     "zone01",
							"k8s.gpu":                  "true",
						},
					},
				}

				c := fake.NewFakeClientWithScheme(testScheme, job, pod, node)
				reconciler := &DataOpJobReconciler{Client: c, Log: fake.NullLogger()}

				err := reconciler.injectPodNodeLabelsToJob(job)
				Expect(err).NotTo(HaveOccurred())

				expectedAnnotations := map[string]string{
					common.AnnotationDataFlowCustomizedAffinityPrefix + common.K8sNodeNameLabelKey: testNodeName,
					common.AnnotationDataFlowCustomizedAffinityPrefix + common.K8sRegionLabelKey:   "region01",
					common.AnnotationDataFlowCustomizedAffinityPrefix + common.K8sZoneLabelKey:     "zone01",
					common.AnnotationDataFlowCustomizedAffinityPrefix + "k8s.gpu":                  "true",
				}
				Expect(job.Annotations).To(Equal(expectedAnnotations))
			})
		})

		Context("when job has only a failed pod", func() {
			It("returns an error", func() {
				job := &batchv1.Job{
					Spec: batchv1.JobSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								controllerUIDKey: jobControllerUIDValue,
							},
						},
					},
				}
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: testPodName,
						Labels: map[string]string{
							controllerUIDKey: jobControllerUIDValue,
						},
					},
					Status: v1.PodStatus{Phase: v1.PodFailed},
				}
				node := &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: testNodeName,
						Labels: map[string]string{
							common.K8sNodeNameLabelKey: testNodeName,
						},
					},
				}

				c := fake.NewFakeClientWithScheme(testScheme, job, pod, node)
				reconciler := &DataOpJobReconciler{Client: c, Log: fake.NullLogger()}

				err := reconciler.injectPodNodeLabelsToJob(job)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when pod has no node name", func() {
			It("returns an error", func() {
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
								controllerUIDKey: "test-uid",
							},
						},
					},
				}
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: testPodName,
						Labels: map[string]string{
							controllerUIDKey: "test-uid",
						},
					},
					Spec:   v1.PodSpec{NodeName: ""},
					Status: v1.PodStatus{Phase: v1.PodSucceeded},
				}

				c := fake.NewFakeClientWithScheme(testScheme, job, pod)
				reconciler := &DataOpJobReconciler{Client: c, Log: fake.NullLogger()}

				err := reconciler.injectPodNodeLabelsToJob(job)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no node name"))
			})
		})

		Context("when job has nil annotations", func() {
			It("creates the annotations map before injecting labels", func() {
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
								controllerUIDKey: "test-uid",
							},
						},
					},
				}
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: testPodName,
						Labels: map[string]string{
							controllerUIDKey: "test-uid",
						},
					},
					Spec:   v1.PodSpec{NodeName: testNodeName},
					Status: v1.PodStatus{Phase: v1.PodSucceeded},
				}
				node := &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: testNodeName,
						Labels: map[string]string{
							common.K8sNodeNameLabelKey: testNodeName,
						},
					},
				}

				c := fake.NewFakeClientWithScheme(testScheme, job, pod, node)
				reconciler := &DataOpJobReconciler{Client: c, Log: fake.NullLogger()}

				err := reconciler.injectPodNodeLabelsToJob(job)
				Expect(err).NotTo(HaveOccurred())
				Expect(job.Annotations).NotTo(BeNil())
			})
		})
	})
})
