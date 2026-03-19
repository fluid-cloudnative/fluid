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

package dataflowaffinity

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("DataOpJobReconciler", func() {
	var testScheme *runtime.Scheme

	BeforeEach(func() {
		testScheme = runtime.NewScheme()
		Expect(v1.AddToScheme(testScheme)).To(Succeed())
		Expect(batchv1.AddToScheme(testScheme)).To(Succeed())
		Expect(datav1alpha1.AddToScheme(testScheme)).To(Succeed())
	})

	Describe("ControllerName", func() {
		It("returns the controller name constant", func() {
			f := &DataOpJobReconciler{Log: fake.NullLogger()}
			Expect(f.ControllerName()).To(Equal(DataOpJobControllerName))
		})
	})

	Describe("ManagedResource", func() {
		It("returns a batchv1.Job object", func() {
			f := &DataOpJobReconciler{Log: fake.NullLogger()}
			obj := f.ManagedResource()
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
			It("returns an error (not-found propagates)", func() {
				c := fake.NewFakeClientWithScheme(testScheme)
				f := &DataOpJobReconciler{Client: c, Log: fake.NullLogger()}
				_, err := f.Reconcile(context.Background(), reconcile.Request{
					NamespacedName: types.NamespacedName{Name: "missing-job", Namespace: "default"},
				})
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when job should not be in queue (cronjob label)", func() {
			It("returns no-requeue without error", func() {
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
				f := &DataOpJobReconciler{Client: c, Log: fake.NullLogger()}
				result, err := f.Reconcile(context.Background(), reconcile.Request{
					NamespacedName: types.NamespacedName{Name: "cron-job", Namespace: "default"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())
			})
		})

		Context("when job is a valid fluid job without affinity annotation", func() {
			It("injects the dataflow affinity annotation and returns no-requeue", func() {
				job := &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-job",
						Namespace: "default",
						Labels: map[string]string{
							common.LabelAnnotationManagedBy: common.Fluid,
						},
					},
				}
				c := fake.NewFakeClientWithScheme(testScheme, job)
				f := &DataOpJobReconciler{Client: c, Log: fake.NullLogger()}
				result, err := f.Reconcile(context.Background(), reconcile.Request{
					NamespacedName: types.NamespacedName{Name: "test-job", Namespace: "default"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())

				updatedJob := &batchv1.Job{}
				Expect(c.Get(context.Background(), types.NamespacedName{Name: "test-job", Namespace: "default"}, updatedJob)).To(Succeed())
				Expect(updatedJob.Annotations).To(HaveKeyWithValue(common.AnnotationDataFlowAffinityInject, "true"))
			})
		})

		Context("when job is complete and has a succeeded pod", func() {
			It("injects node labels and returns no-requeue", func() {
				job := &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "complete-job",
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
								"controller-uid": "abc-123",
							},
						},
					},
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{
							{Type: batchv1.JobComplete},
						},
					},
				}
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "complete-pod",
						Namespace: "default",
						Labels: map[string]string{
							"controller-uid": "abc-123",
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
							common.K8sRegionLabelKey:   "region01",
							common.K8sZoneLabelKey:     "zone01",
						},
					},
				}
				c := fake.NewFakeClientWithScheme(testScheme, job, pod, node)
				f := &DataOpJobReconciler{Client: c, Log: fake.NullLogger()}
				result, err := f.Reconcile(context.Background(), reconcile.Request{
					NamespacedName: types.NamespacedName{Name: "complete-job", Namespace: "default"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())

				updatedJob := &batchv1.Job{}
				Expect(c.Get(context.Background(), types.NamespacedName{Name: "complete-job", Namespace: "default"}, updatedJob)).To(Succeed())
				Expect(updatedJob.Annotations).To(HaveKeyWithValue(common.AnnotationDataFlowCustomizedAffinityPrefix+common.K8sNodeNameLabelKey, "node01"))
				Expect(updatedJob.Annotations).To(HaveKeyWithValue(common.AnnotationDataFlowCustomizedAffinityPrefix+common.K8sRegionLabelKey, "region01"))
				Expect(updatedJob.Annotations).To(HaveKeyWithValue(common.AnnotationDataFlowCustomizedAffinityPrefix+common.K8sZoneLabelKey, "zone01"))
			})
		})
	})

	Describe("injectPodNodeLabelsToJob", func() {
		Context("when job has a succeeded pod", func() {
			It("should inject node labels as annotations onto the job", func() {
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
				f := &DataOpJobReconciler{
					Client: c,
					Log:    fake.NullLogger(),
				}

				err := f.injectPodNodeLabelsToJob(job)
				Expect(err).NotTo(HaveOccurred())

				wantAnnotations := map[string]string{
					common.AnnotationDataFlowCustomizedAffinityPrefix + common.K8sNodeNameLabelKey: "node01",
					common.AnnotationDataFlowCustomizedAffinityPrefix + common.K8sRegionLabelKey:   "region01",
					common.AnnotationDataFlowCustomizedAffinityPrefix + common.K8sZoneLabelKey:     "zone01",
					common.AnnotationDataFlowCustomizedAffinityPrefix + "k8s.gpu":                  "true",
				}
				Expect(job.Annotations).To(Equal(wantAnnotations))
			})
		})

		Context("when job has only a failed pod", func() {
			It("should return an error", func() {
				job := &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-job-failed",
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
						},
					},
				}

				c := fake.NewFakeClientWithScheme(testScheme, job, pod, node)
				f := &DataOpJobReconciler{
					Client: c,
					Log:    fake.NullLogger(),
				}

				err := f.injectPodNodeLabelsToJob(job)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
