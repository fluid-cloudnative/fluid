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

package dataload

import (
	"time"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/compatibility"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

const (
	defaultNamespace = "default"
	dataLoadName     = "test-dataload"
	loaderJobName    = "test-dataload-loader-job"
	targetDataset    = "hadoop"
	nodeName         = "node-1"
)

var _ = Describe("OnceStatusHandler", func() {
	var (
		testScheme   *runtime.Scheme
		mockDataload v1alpha1.DataLoad
	)

	BeforeEach(func() {
		testScheme = runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(testScheme)).To(Succeed())
		Expect(batchv1.AddToScheme(testScheme)).To(Succeed())

		mockDataload = v1alpha1.DataLoad{
			ObjectMeta: v1.ObjectMeta{
				Name:      dataLoadName,
				Namespace: defaultNamespace,
			},
			Spec: v1alpha1.DataLoadSpec{
				Dataset: v1alpha1.TargetDataset{
					Name:      targetDataset,
					Namespace: defaultNamespace,
				},
			},
			Status: v1alpha1.OperationStatus{
				Phase:    common.PhasePending,
				Duration: "3s",
				Conditions: []v1alpha1.Condition{{
					Type:   common.Complete,
					Status: corev1.ConditionTrue,
				}},
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{{
							MatchExpressions: []corev1.NodeSelectorRequirement{{
								Key:      "kubernetes.io/hostname",
								Operator: corev1.NodeSelectorOpIn,
								Values:   []string{nodeName},
							}},
						}},
					},
				},
			},
		}
	})

	DescribeTable("GetOperationStatus",
		func(job batchv1.Job, expectedPhase common.Phase, expectedConditionType common.ConditionType) {
			client := fake.NewFakeClientWithScheme(testScheme, &mockDataload, &job)
			handler := &OnceStatusHandler{Client: client, dataLoad: &mockDataload}
			ctx := cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{Namespace: defaultNamespace, Name: ""},
				Log:            fake.NullLogger(),
			}

			opStatus, err := handler.GetOperationStatus(ctx, &mockDataload.Status)
			Expect(err).NotTo(HaveOccurred())
			Expect(opStatus.Phase).To(Equal(expectedPhase))
			Expect(opStatus.Conditions).To(HaveLen(1))
			Expect(opStatus.Conditions[0].Type).To(Equal(expectedConditionType))
			Expect(opStatus.Duration).NotTo(BeEmpty())
		},
		Entry("job success yields PhaseComplete",
			batchv1.Job{
				ObjectMeta: v1.ObjectMeta{Name: loaderJobName, Namespace: defaultNamespace},
				Status: batchv1.JobStatus{
					Conditions: []batchv1.JobCondition{
						{
							Type:               batchv1.JobComplete,
							LastProbeTime:      v1.NewTime(time.Now()),
							LastTransitionTime: v1.NewTime(time.Now()),
						},
					},
				},
			},
			common.PhaseComplete,
			common.Complete,
		),
		Entry("job failed yields PhaseFailed",
			batchv1.Job{
				ObjectMeta: v1.ObjectMeta{Name: loaderJobName, Namespace: defaultNamespace},
				Status: batchv1.JobStatus{
					Conditions: []batchv1.JobCondition{
						{
							Type:               batchv1.JobFailed,
							LastProbeTime:      v1.NewTime(time.Now()),
							LastTransitionTime: v1.NewTime(time.Now()),
						},
					},
				},
			},
			common.PhaseFailed,
			common.Failed,
		),
	)

	It("generates NodeAffinity when completed job injects affinity and current status has none", func() {
		mockDataload.Status.NodeAffinity = nil
		completedJob := batchv1.Job{
			ObjectMeta: v1.ObjectMeta{
				Name:      loaderJobName,
				Namespace: defaultNamespace,
				Annotations: map[string]string{
					common.AnnotationDataFlowAffinityInject:                                        "true",
					common.AnnotationDataFlowCustomizedAffinityPrefix + common.K8sNodeNameLabelKey: nodeName,
				},
			},
			Status: batchv1.JobStatus{
				Conditions: []batchv1.JobCondition{{
					Type:               batchv1.JobComplete,
					Status:             corev1.ConditionTrue,
					LastProbeTime:      v1.NewTime(time.Now()),
					LastTransitionTime: v1.NewTime(time.Now()),
				}},
			},
		}
		client := fake.NewFakeClientWithScheme(testScheme, &mockDataload, &completedJob)
		handler := &OnceStatusHandler{Client: client, dataLoad: &mockDataload}
		ctx := cruntime.ReconcileRequestContext{
			NamespacedName: types.NamespacedName{Namespace: defaultNamespace, Name: ""},
			Log:            fake.NullLogger(),
		}

		opStatus, err := handler.GetOperationStatus(ctx, &mockDataload.Status)
		Expect(err).NotTo(HaveOccurred())
		Expect(opStatus.Phase).To(Equal(common.PhaseComplete))
		Expect(opStatus.NodeAffinity).NotTo(BeNil())
		Expect(opStatus.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution).NotTo(BeNil())
		Expect(opStatus.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms).To(HaveLen(1))
		Expect(opStatus.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions).To(HaveLen(1))
		expression := opStatus.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0]
		Expect(expression.Key).To(Equal(common.K8sNodeNameLabelKey))
		Expect(expression.Operator).To(Equal(corev1.NodeSelectorOpIn))
		Expect(expression.Values).To(Equal([]string{nodeName}))
	})

	It("GetOperationStatus returns current status when job is still running (no finished condition)", func() {
		runningJob := batchv1.Job{
			ObjectMeta: v1.ObjectMeta{Name: loaderJobName, Namespace: defaultNamespace},
			Status:     batchv1.JobStatus{}, // no conditions = still running
		}
		client := fake.NewFakeClientWithScheme(testScheme, &mockDataload, &runningJob)
		handler := &OnceStatusHandler{Client: client, dataLoad: &mockDataload}
		ctx := cruntime.ReconcileRequestContext{
			NamespacedName: types.NamespacedName{Namespace: defaultNamespace, Name: ""},
			Log:            fake.NullLogger(),
		}
		originalStatus := mockDataload.Status.DeepCopy()

		opStatus, err := handler.GetOperationStatus(ctx, &mockDataload.Status)
		Expect(err).NotTo(HaveOccurred())
		// when job is still running, status is returned as an unchanged DeepCopy
		Expect(opStatus).NotTo(BeNil())
		Expect(opStatus).NotTo(BeIdenticalTo(&mockDataload.Status))
		Expect(*opStatus).To(Equal(*originalStatus))
	})
})

var _ = Describe("CronStatusHandler", func() {
	var (
		testScheme         *runtime.Scheme
		mockCronDataload   v1alpha1.DataLoad
		lastScheduleTime   v1.Time
		lastSuccessfulTime v1.Time
		mockCronJob        batchv1.CronJob
		patches            *gomonkey.Patches
	)

	BeforeEach(func() {
		testScheme = runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(testScheme)).To(Succeed())
		Expect(batchv1.AddToScheme(testScheme)).To(Succeed())

		startTime := time.Date(2023, 8, 1, 12, 0, 0, 0, time.Local)
		lastScheduleTime = v1.NewTime(startTime)
		lastSuccessfulTime = v1.NewTime(startTime.Add(time.Second * 10))

		mockCronDataload = v1alpha1.DataLoad{
			ObjectMeta: v1.ObjectMeta{Name: dataLoadName, Namespace: defaultNamespace},
			Spec: v1alpha1.DataLoadSpec{
				Dataset:  v1alpha1.TargetDataset{Name: targetDataset, Namespace: defaultNamespace},
				Policy:   "Cron",
				Schedule: "* * * * *",
			},
			Status: v1alpha1.OperationStatus{
				Phase: common.PhaseComplete,
			},
		}

		mockCronJob = batchv1.CronJob{
			ObjectMeta: v1.ObjectMeta{Name: loaderJobName, Namespace: defaultNamespace},
			Spec:       batchv1.CronJobSpec{Schedule: "* * * * *"},
			Status: batchv1.CronJobStatus{
				LastScheduleTime:   &lastScheduleTime,
				LastSuccessfulTime: &lastSuccessfulTime,
			},
		}

		// Patch IsBatchV1CronJobSupported to avoid ctrl.GetConfigOrDie() panic in unit tests.
		patches = gomonkey.ApplyFunc(compatibility.IsBatchV1CronJobSupported, func() bool {
			return true
		})
	})

	AfterEach(func() {
		patches.Reset()
	})

	It("completed job yields PhaseComplete", func() {
		job := batchv1.Job{
			ObjectMeta: v1.ObjectMeta{
				Name:              "test-dataload-loader-job-1",
				Namespace:         defaultNamespace,
				Labels:            map[string]string{"cronjob": loaderJobName},
				CreationTimestamp: lastScheduleTime,
			},
			Status: batchv1.JobStatus{
				Conditions: []batchv1.JobCondition{
					{
						Type:               batchv1.JobComplete,
						LastProbeTime:      lastSuccessfulTime,
						LastTransitionTime: lastSuccessfulTime,
					},
				},
			},
		}
		client := fake.NewFakeClientWithScheme(testScheme, &mockCronDataload, &mockCronJob, &job)
		handler := &CronStatusHandler{Client: client, dataLoad: &mockCronDataload}
		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}

		opStatus, err := handler.GetOperationStatus(ctx, &mockCronDataload.Status)
		Expect(err).NotTo(HaveOccurred())
		Expect(opStatus.LastScheduleTime).To(Equal(&lastScheduleTime))
		Expect(opStatus.LastSuccessfulTime).To(Equal(&lastSuccessfulTime))
		Expect(opStatus.Phase).To(Equal(common.PhaseComplete))
	})

	It("running job yields PhasePending", func() {
		job := batchv1.Job{
			ObjectMeta: v1.ObjectMeta{
				Name:              "test-dataload-loader-job-1",
				Namespace:         defaultNamespace,
				Labels:            map[string]string{"cronjob": loaderJobName},
				CreationTimestamp: lastScheduleTime,
			},
			Status: batchv1.JobStatus{},
		}
		client := fake.NewFakeClientWithScheme(testScheme, &mockCronDataload, &mockCronJob, &job)
		handler := &CronStatusHandler{Client: client, dataLoad: &mockCronDataload}
		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}

		opStatus, err := handler.GetOperationStatus(ctx, &mockCronDataload.Status)
		Expect(err).NotTo(HaveOccurred())
		Expect(opStatus.LastScheduleTime).To(Equal(&lastScheduleTime))
		Expect(opStatus.LastSuccessfulTime).To(Equal(&lastSuccessfulTime))
		Expect(opStatus.Phase).To(Equal(common.PhasePending))
	})

	It("returns status with timestamps when no current job matches schedule time", func() {
		// no jobs in the fake client that match lastScheduleTime
		client := fake.NewFakeClientWithScheme(testScheme, &mockCronDataload, &mockCronJob)
		handler := &CronStatusHandler{Client: client, dataLoad: &mockCronDataload}
		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}

		opStatus, err := handler.GetOperationStatus(ctx, &mockCronDataload.Status)
		Expect(err).NotTo(HaveOccurred())
		// should still update the schedule/successful times from the cronjob status
		Expect(opStatus.LastScheduleTime).To(Equal(&lastScheduleTime))
		Expect(opStatus.LastSuccessfulTime).To(Equal(&lastSuccessfulTime))
	})
})

var _ = Describe("OnEventStatusHandler", func() {
	It("GetOperationStatus returns nil result and nil error (stub)", func() {
		testScheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(testScheme)).To(Succeed())
		dataload := &v1alpha1.DataLoad{
			ObjectMeta: v1.ObjectMeta{Name: dataLoadName, Namespace: defaultNamespace},
		}
		c := fake.NewFakeClientWithScheme(testScheme, dataload)
		handler := &OnEventStatusHandler{Client: c, dataLoad: dataload}
		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}

		result, err := handler.GetOperationStatus(ctx, &dataload.Status)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeNil())
	})
})
