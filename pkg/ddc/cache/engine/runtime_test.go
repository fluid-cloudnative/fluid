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

package engine

import (
	"context"
	"fmt"

	workloadv1alpha1 "github.com/fluid-cloudnative/advanced-statefulset/api/workload/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("CacheRuntimeInfo GetWorkerPods Tests", Label("pkg.ddc.cache.engine.runtime_test.go"), func() {
	var (
		cacheRuntimeInfo *CacheRuntimeInfo
		fakeClient       client.Client
	)

	BeforeEach(func() {
		scheme := CacheEngineTestScheme
		fakeClient = fake.NewFakeClientWithScheme(scheme)
	})

	Context("When getting worker pods", func() {
		It("should return worker pods matching the selector", func() {
			// Create runtime info
			runtimeName := "test-cache"
			namespace := "default"

			runtimeInfo, err := base.BuildRuntimeInfo(runtimeName, namespace, "cache")
			Expect(err).NotTo(HaveOccurred())

			cacheRuntimeInfo = &CacheRuntimeInfo{RuntimeInfoInterface: runtimeInfo}

			// Create AdvancedStatefulSet with specific labels
			workerName := common.GetCacheComponentName(runtimeName, common.ComponentTypeWorker)

			advancedSts := &workloadv1alpha1.AdvancedStatefulSet{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps/v1",
					Kind:       "StatefulSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      workerName,
					Namespace: namespace,
					UID:       "test-worker-uid",
				},
				Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							common.LabelCacheRuntimeName:          runtimeName,
							common.LabelCacheRuntimeComponentName: workerName,
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								common.LabelCacheRuntimeName:          runtimeName,
								common.LabelCacheRuntimeComponentName: workerName,
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "worker",
									Image: "test-image",
								},
							},
						},
					},
				},
			}

			// Create matching pods
			pod1 := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cache-worker-0",
					Namespace: namespace,
					Labels: map[string]string{
						common.LabelCacheRuntimeName:          runtimeName,
						common.LabelCacheRuntimeComponentName: workerName,
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         "apps/v1",
							Kind:               "StatefulSet",
							Name:               workerName,
							UID:                "test-worker-uid",
							Controller:         func() *bool { b := true; return &b }(),
							BlockOwnerDeletion: func() *bool { b := true; return &b }(),
						},
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "worker",
							Image: "test-image",
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			}

			pod2 := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cache-worker-1",
					Namespace: namespace,
					Labels: map[string]string{
						common.LabelCacheRuntimeName:          runtimeName,
						common.LabelCacheRuntimeComponentName: workerName,
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         "apps/v1",
							Kind:               "StatefulSet",
							Name:               workerName,
							UID:                "test-worker-uid",
							Controller:         func() *bool { b := true; return &b }(),
							BlockOwnerDeletion: func() *bool { b := true; return &b }(),
						},
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "worker",
							Image: "test-image",
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			}

			// Create non-matching pod (master pod)
			masterName := common.GetCacheComponentName(runtimeName, common.ComponentTypeMaster)
			nonMatchingPod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cache-master-0",
					Namespace: namespace,
					Labels: map[string]string{
						common.LabelCacheRuntimeName:          runtimeName,
						common.LabelCacheRuntimeComponentName: masterName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "master",
							Image: "test-image",
						},
					},
				},
			}

			// Add objects to fake client
			Expect(fakeClient.Create(context.TODO(), advancedSts)).To(Succeed())
			Expect(fakeClient.Create(context.TODO(), pod1)).To(Succeed())
			Expect(fakeClient.Create(context.TODO(), pod2)).To(Succeed())
			Expect(fakeClient.Create(context.TODO(), nonMatchingPod)).To(Succeed())

			// Test the GetWorkerPods function
			pods, err := cacheRuntimeInfo.GetWorkerPods(fakeClient)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(pods)).To(Equal(2))

			// Verify that only matching pods are returned
			podNames := []string{}
			for _, pod := range pods {
				podNames = append(podNames, pod.Name)
			}
			Expect(podNames).To(ContainElements("test-cache-worker-0", "test-cache-worker-1"))
			Expect(podNames).NotTo(ContainElement("test-cache-master-0"))
		})

		It("should return empty list when no pods match the selector", func() {
			// Create runtime info
			runtimeName := "test-cache"
			namespace := "default"

			runtimeInfo, err := base.BuildRuntimeInfo(runtimeName, namespace, "cache")
			Expect(err).NotTo(HaveOccurred())

			cacheRuntimeInfo = &CacheRuntimeInfo{RuntimeInfoInterface: runtimeInfo}

			// Create AdvancedStatefulSet with specific labels
			workerName := common.GetCacheComponentName(runtimeName, common.ComponentTypeWorker)

			advancedSts := &workloadv1alpha1.AdvancedStatefulSet{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps/v1",
					Kind:       "StatefulSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      workerName,
					Namespace: namespace,
					UID:       "test-worker-uid",
				},
				Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							common.LabelCacheRuntimeName:          runtimeName,
							common.LabelCacheRuntimeComponentName: workerName,
						},
					},
				},
			}

			// Create non-matching pods only (master pods)
			masterName := common.GetCacheComponentName(runtimeName, common.ComponentTypeMaster)
			nonMatchingPod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cache-master-0",
					Namespace: namespace,
					Labels: map[string]string{
						common.LabelCacheRuntimeName:          runtimeName,
						common.LabelCacheRuntimeComponentName: masterName,
					},
				},
			}

			// Add objects to fake client
			Expect(fakeClient.Create(context.TODO(), advancedSts)).To(Succeed())
			Expect(fakeClient.Create(context.TODO(), nonMatchingPod)).To(Succeed())

			pods, err := cacheRuntimeInfo.GetWorkerPods(fakeClient)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(pods)).To(Equal(0))
		})

		It("should return error when AdvancedStatefulSet does not exist", func() {
			// Create runtime info
			runtimeName := "test-cache"
			namespace := "default"

			runtimeInfo, err := base.BuildRuntimeInfo(runtimeName, namespace, "cache")
			Expect(err).NotTo(HaveOccurred())

			cacheRuntimeInfo = &CacheRuntimeInfo{RuntimeInfoInterface: runtimeInfo}

			// Don't create any AdvancedStatefulSet
			pods, err := cacheRuntimeInfo.GetWorkerPods(fakeClient)
			Expect(err).To(HaveOccurred())
			Expect(pods).To(BeNil())
		})

		It("should handle pods from different namespaces correctly", func() {
			// Create runtime info
			runtimeName := "test-cache"
			namespace := "default"
			otherNamespace := "other-ns"

			runtimeInfo, err := base.BuildRuntimeInfo(runtimeName, namespace, "cache")
			Expect(err).NotTo(HaveOccurred())

			cacheRuntimeInfo = &CacheRuntimeInfo{RuntimeInfoInterface: runtimeInfo}

			// Create AdvancedStatefulSet in default namespace
			workerName := common.GetCacheComponentName(runtimeName, common.ComponentTypeWorker)

			advancedSts := &workloadv1alpha1.AdvancedStatefulSet{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps/v1",
					Kind:       "StatefulSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      workerName,
					Namespace: namespace,
					UID:       "test-worker-uid",
				},
				Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							common.LabelCacheRuntimeName:          runtimeName,
							common.LabelCacheRuntimeComponentName: workerName,
						},
					},
				},
			}

			// Create pod in the same namespace with matching labels
			matchingPod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cache-worker-0",
					Namespace: namespace,
					Labels: map[string]string{
						common.LabelCacheRuntimeName:          runtimeName,
						common.LabelCacheRuntimeComponentName: workerName,
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         "apps/v1",
							Kind:               "StatefulSet",
							Name:               workerName,
							UID:                "test-worker-uid",
							Controller:         func() *bool { b := true; return &b }(),
							BlockOwnerDeletion: func() *bool { b := true; return &b }(),
						},
					},
				},
			}

			// Create pod in different namespace with same labels
			otherNsPod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cache-worker-0",
					Namespace: otherNamespace,
					Labels: map[string]string{
						common.LabelCacheRuntimeName:          runtimeName,
						common.LabelCacheRuntimeComponentName: workerName,
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         "apps/v1",
							Kind:               "StatefulSet",
							Name:               workerName,
							UID:                "test-worker-uid",
							Controller:         func() *bool { b := true; return &b }(),
							BlockOwnerDeletion: func() *bool { b := true; return &b }(),
						},
					},
				},
			}

			// Add objects to fake client
			Expect(fakeClient.Create(context.TODO(), advancedSts)).To(Succeed())
			Expect(fakeClient.Create(context.TODO(), matchingPod)).To(Succeed())
			Expect(fakeClient.Create(context.TODO(), otherNsPod)).To(Succeed())

			pods, err := cacheRuntimeInfo.GetWorkerPods(fakeClient)
			Expect(err).NotTo(HaveOccurred())
			// Should only return pods from the correct namespace
			Expect(len(pods)).To(Equal(1))
			Expect(pods[0].Name).To(Equal("test-cache-worker-0"))
			Expect(pods[0].Namespace).To(Equal(namespace))
		})

		It("should return all worker pods when multiple replicas exist", func() {
			// Create runtime info
			runtimeName := "test-cache"
			namespace := "default"

			runtimeInfo, err := base.BuildRuntimeInfo(runtimeName, namespace, "cache")
			Expect(err).NotTo(HaveOccurred())

			cacheRuntimeInfo = &CacheRuntimeInfo{RuntimeInfoInterface: runtimeInfo}

			// Create AdvancedStatefulSet
			workerName := common.GetCacheComponentName(runtimeName, common.ComponentTypeWorker)

			advancedSts := &workloadv1alpha1.AdvancedStatefulSet{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps/v1",
					Kind:       "StatefulSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      workerName,
					Namespace: namespace,
					UID:       "test-worker-uid",
				},
				Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
					Replicas: func() *int32 { i := int32(3); return &i }(),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							common.LabelCacheRuntimeName:          runtimeName,
							common.LabelCacheRuntimeComponentName: workerName,
						},
					},
				},
			}

			// Create 3 worker pods
			for i := 0; i < 3; i++ {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("test-cache-worker-%d", i),
						Namespace: namespace,
						Labels: map[string]string{
							common.LabelCacheRuntimeName:          runtimeName,
							common.LabelCacheRuntimeComponentName: workerName,
						},
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion:         "apps/v1",
								Kind:               "StatefulSet",
								Name:               workerName,
								UID:                "test-worker-uid",
								Controller:         func() *bool { b := true; return &b }(),
								BlockOwnerDeletion: func() *bool { b := true; return &b }(),
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				}
				Expect(fakeClient.Create(context.TODO(), pod)).To(Succeed())
			}

			Expect(fakeClient.Create(context.TODO(), advancedSts)).To(Succeed())

			pods, err := cacheRuntimeInfo.GetWorkerPods(fakeClient)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(pods)).To(Equal(3))
		})
	})

	Context("When testing GetComponentName helper", func() {
		It("should generate correct worker component name", func() {
			runtimeName := "my-cache-runtime"
			expectedWorkerName := runtimeName + "-" + string(common.ComponentTypeWorker)

			actualWorkerName := common.GetCacheComponentName(runtimeName, common.ComponentTypeWorker)
			Expect(actualWorkerName).To(Equal(expectedWorkerName))
			Expect(actualWorkerName).To(Equal("my-cache-runtime-worker"))
		})

		It("should generate correct master component name", func() {
			runtimeName := "my-cache-runtime"
			expectedMasterName := runtimeName + "-" + string(common.ComponentTypeMaster)

			actualMasterName := common.GetCacheComponentName(runtimeName, common.ComponentTypeMaster)
			Expect(actualMasterName).To(Equal(expectedMasterName))
			Expect(actualMasterName).To(Equal("my-cache-runtime-master"))
		})
	})
})
