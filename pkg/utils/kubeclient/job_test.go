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

package kubeclient

import (
	"context"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Job related unit tests", Label("pkg.utils.kubeclient.job_test.go"), func() {
	var client client.Client
	var resources []runtime.Object

	JustBeforeEach(func() {
		client = fake.NewFakeClientWithScheme(testScheme, resources...)
	})

	Describe("Test GetSuccessPodForJob()", func() {
		jobPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-job-pod",
				Namespace: "test-ns",
				Labels: map[string]string{
					"job-name": "test-job",
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodSucceeded,
			},
		}

		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-job",
				Namespace: "test-ns",
			},
			Spec: batchv1.JobSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"job-name": "test-job",
					},
				},
			},
		}

		When("everything works as expected", func() {
			BeforeEach(func() {
				resources = []runtime.Object{job, jobPod}
			})

			It("should return the succeed pod successfully", func() {
				gotPod, err := GetSucceedPodForJob(client, job)
				Expect(err).To(BeNil())
				Expect(gotPod).To(Equal(jobPod))
			})
		})

		When("job pod does not succeed", func() {
			BeforeEach(func() {
				jobPod.Status.Phase = corev1.PodFailed
				resources = []runtime.Object{job, jobPod}
			})

			It("should not return the pod", func() {
				gotPod, err := GetSucceedPodForJob(client, job)
				Expect(err).To(BeNil())
				Expect(gotPod).To(BeNil())
			})
		})

		When("job pod does not exist", func() {
			BeforeEach(func() {
				resources = []runtime.Object{job}
			})

			It("should return nil because it is not found", func() {
				gotPod, err := GetSucceedPodForJob(client, job)
				Expect(err).To(BeNil())
				Expect(gotPod).To(BeNil())
			})
		})
	})

	Describe("Test UpdateJob()", func() {
		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-job",
				Namespace: "test-ns",
			},
			Spec: batchv1.JobSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"job-name": "test-job",
					},
				},
			},
		}

		When("works as expected", func() {
			BeforeEach(func() {
				resources = []runtime.Object{job}
			})

			It("should successfully update job", func() {
				jobToUpdate := job.DeepCopy()
				jobToUpdate.Spec.Selector = &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"updated-label": "xxx",
					},
				}

				Expect(UpdateJob(client, jobToUpdate)).To(Succeed())
				jobUpdated := &batchv1.Job{}
				Expect(client.Get(context.TODO(), types.NamespacedName{Name: job.Name, Namespace: job.Namespace}, jobUpdated)).To(Succeed())
				Expect(jobUpdated.Spec.Selector).To(Equal(jobToUpdate.Spec.Selector))
			})
		})

		When("job is not exists", func() {
			BeforeEach(func() {
				resources = []runtime.Object{}
			})

			It("should return not found error", func() {
				err := UpdateJob(client, job)
				Expect(err).NotTo(BeNil())
				Expect(apierrs.IsNotFound(err)).To(BeTrue())
			})
		})
	})

	Describe("Test GetJob()", func() {
		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-job",
				Namespace: "test-ns",
			},
			Spec: batchv1.JobSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"job-name": "test-job",
					},
				},
			},
		}

		When("job exists", func() {
			BeforeEach(func() {
				resources = []runtime.Object{job}
			})

			It("should successfully get job", func() {
				gotJob, err := GetJob(client, job.Name, job.Namespace)
				Expect(err).To(BeNil())
				Expect(gotJob).To(Equal(job))
			})
		})

		When("job does not exist", func() {
			BeforeEach(func() {
				resources = []runtime.Object{}
			})

			It("should return not found error", func() {
				gotJob, err := GetJob(client, job.Name, job.Namespace)
				Expect(err).NotTo(BeNil())
				Expect(gotJob).To(BeNil())
				Expect(apierrs.IsNotFound(err)).To(BeTrue())
			})
		})
	})

	Describe("Test GetFinishedJobCondition()", func() {
		When("job has no conditions", func() {
			It("Should return nil", func() {
				job := &batchv1.Job{
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{},
					},
				}
				condition := GetFinishedJobCondition(job)
				Expect(condition).To(BeNil())
			})
		})

		When("job has JobComplete condition", func() {
			It("Should return the JobComplete condition", func() {
				job := &batchv1.Job{
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{
							{
								Type: batchv1.JobSuspended,
							},
							{
								Type: batchv1.JobComplete,
							},
						},
					},
				}
				condition := GetFinishedJobCondition(job)
				Expect(condition).NotTo(BeNil())
				Expect(condition.Type).To(Equal(batchv1.JobComplete))
			})
		})

		When("job has JobFailed condition", func() {
			It("Should return the JobFailed condition", func() {
				job := &batchv1.Job{
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{
							{
								Type: batchv1.JobSuspended,
							},
							{
								Type: batchv1.JobFailed,
							},
						},
					},
				}
				condition := GetFinishedJobCondition(job)
				Expect(condition).NotTo(BeNil())
				Expect(condition.Type).To(Equal(batchv1.JobFailed))
			})
		})

		When("job has other conditions", func() {
			It("Should return nil", func() {
				job := &batchv1.Job{
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{
							{
								Type: batchv1.JobSuspended,
							},
						},
					},
				}
				condition := GetFinishedJobCondition(job)
				Expect(condition).To(BeNil())
			})
		})
	})
})
