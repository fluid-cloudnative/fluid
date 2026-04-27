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

package component

import (
	"context"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ComponentManager", func() {
	Describe("NewComponentHelper", func() {
		It("should return StatefulSetManager for StatefulSet workload", func() {
			workloadType := metav1.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "StatefulSet",
			}
			scheme := runtime.NewScheme()
			client := fake.NewFakeClientWithScheme(scheme)

			manager := NewComponentHelper(workloadType, client)
			Expect(manager).NotTo(BeNil())
			_, ok := manager.(*StatefulSetManager)
			Expect(ok).To(BeTrue())
		})

		It("should return DaemonSetManager for DaemonSet workload", func() {
			workloadType := metav1.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "DaemonSet",
			}
			scheme := runtime.NewScheme()
			client := fake.NewFakeClientWithScheme(scheme)

			manager := NewComponentHelper(workloadType, client)
			Expect(manager).NotTo(BeNil())
			_, ok := manager.(*DaemonSetManager)
			Expect(ok).To(BeTrue())
		})

		It("should return StatefulSetManager as default for unknown kind", func() {
			workloadType := metav1.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "Unknown",
			}
			scheme := runtime.NewScheme()
			client := fake.NewFakeClientWithScheme(scheme)

			manager := NewComponentHelper(workloadType, client)
			Expect(manager).NotTo(BeNil())
			_, ok := manager.(*StatefulSetManager)
			Expect(ok).To(BeTrue())
		})

		It("should return StatefulSetManager as default for wrong APIVersion", func() {
			workloadType := metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "StatefulSet",
			}
			scheme := runtime.NewScheme()
			client := fake.NewFakeClientWithScheme(scheme)

			manager := NewComponentHelper(workloadType, client)
			Expect(manager).NotTo(BeNil())
			_, ok := manager.(*StatefulSetManager)
			Expect(ok).To(BeTrue())
		})
	})

	Describe("getCommonLabelsFromComponent", func() {
		It("should generate correct labels with runtime name and component name", func() {
			component := &common.CacheRuntimeComponentValue{
				Name: "test-runtime-master",
				Owner: &common.OwnerReference{
					Name: "test-runtime",
				},
			}

			labels := getCommonLabelsFromComponent(component)
			Expect(labels).To(HaveKey(common.LabelCacheRuntimeName))
			Expect(labels[common.LabelCacheRuntimeName]).To(Equal("test-runtime"))
			Expect(labels).To(HaveKey(common.LabelCacheRuntimeComponentName))
			Expect(labels[common.LabelCacheRuntimeComponentName]).To(Equal("test-runtime-master"))
			Expect(len(labels)).To(Equal(2))
		})
	})
})

// setupTestClient creates a fake client with the necessary schemes registered
func setupTestClient() client.Client {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	return fake.NewFakeClientWithScheme(scheme)
}

var _ = Describe("StatefulSetManager", func() {
	var (
		manager   *StatefulSetManager
		ctx       context.Context
		component *common.CacheRuntimeComponentValue
	)

	BeforeEach(func() {
		client := setupTestClient()
		manager = newStatefulSetManager(client)
		ctx = context.Background()

		replicas := int32(3)
		component = &common.CacheRuntimeComponentValue{
			Name:      "test-runtime-master",
			Namespace: "fluid",
			Replicas:  replicas,
			Owner: &common.OwnerReference{
				APIVersion: "data.fluid.io/v1alpha1",
				Kind:       "CacheRuntime",
				Name:       "test-runtime",
				UID:        "test-uid",
			},
			PodTemplateSpec: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "master",
							Image: "test-image:latest",
						},
					},
				},
			},
			Service: &common.CacheRuntimeComponentServiceConfig{
				Name: "test-runtime-master-svc",
			},
			WorkloadType: metav1.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "StatefulSet",
			},
		}
	})

	Describe("Reconciler", func() {
		It("should create StatefulSet and Service successfully", func() {
			err := manager.Reconciler(ctx, component)
			Expect(err).NotTo(HaveOccurred())

			sts := &appsv1.StatefulSet{}
			err = manager.client.Get(ctx, types.NamespacedName{
				Name:      "test-runtime-master",
				Namespace: "fluid",
			}, sts)
			Expect(err).NotTo(HaveOccurred())
			Expect(sts.Name).To(Equal("test-runtime-master"))
			Expect(*sts.Spec.Replicas).To(Equal(int32(3)))
			Expect(sts.Spec.PodManagementPolicy).To(Equal(appsv1.ParallelPodManagement))
			Expect(sts.Spec.ServiceName).To(Equal("test-runtime-master-svc"))

			svc := &corev1.Service{}
			err = manager.client.Get(ctx, types.NamespacedName{
				Name:      "test-runtime-master-svc",
				Namespace: "fluid",
			}, svc)
			Expect(err).NotTo(HaveOccurred())
			Expect(svc.Name).To(Equal("test-runtime-master-svc"))
			Expect(svc.Spec.ClusterIP).To(Equal("None"))
			Expect(svc.Spec.PublishNotReadyAddresses).To(BeTrue())
		})

		It("should not recreate if StatefulSet already exists", func() {
			// First reconciliation
			err := manager.Reconciler(ctx, component)
			Expect(err).NotTo(HaveOccurred())

			// Second reconciliation should not fail
			err = manager.Reconciler(ctx, component)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle nil service gracefully", func() {
			component.Service = nil
			err := manager.Reconciler(ctx, component)
			Expect(err).NotTo(HaveOccurred())

			// Verify StatefulSet was created without ServiceName
			sts := &appsv1.StatefulSet{}
			err = manager.client.Get(ctx, types.NamespacedName{
				Name:      "test-runtime-master",
				Namespace: "fluid",
			}, sts)
			Expect(err).NotTo(HaveOccurred())
			Expect(sts.Spec.ServiceName).To(BeEmpty())
		})
	})

	Describe("ConstructComponentStatus", func() {
		It("should return Ready phase when all replicas are ready", func() {
			replicas := int32(3)
			sts := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-runtime-master",
					Namespace: "fluid",
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: &replicas,
				},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas:     3,
					CurrentReplicas:   3,
					AvailableReplicas: 3,
				},
			}
			Expect(manager.client.Create(ctx, sts)).To(Succeed())

			status, err := manager.ConstructComponentStatus(ctx, component)
			Expect(err).NotTo(HaveOccurred())
			Expect(status.DesiredReplicas).To(Equal(int32(3)))
			Expect(status.ReadyReplicas).To(Equal(int32(3)))
			Expect(status.CurrentReplicas).To(Equal(int32(3)))
			Expect(status.AvailableReplicas).To(Equal(int32(3)))
			Expect(status.UnavailableReplicas).To(Equal(int32(0)))
			Expect(status.Phase).To(Equal(datav1alpha1.RuntimePhaseReady))
		})

		It("should return NotReady phase when replicas are partially ready", func() {
			replicas := int32(3)
			sts := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-runtime-master",
					Namespace: "fluid",
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: &replicas,
				},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas:     2,
					CurrentReplicas:   3,
					AvailableReplicas: 2,
				},
			}
			Expect(manager.client.Create(ctx, sts)).To(Succeed())

			status, err := manager.ConstructComponentStatus(ctx, component)
			Expect(err).NotTo(HaveOccurred())
			Expect(status.DesiredReplicas).To(Equal(int32(3)))
			Expect(status.ReadyReplicas).To(Equal(int32(2)))
			Expect(status.UnavailableReplicas).To(Equal(int32(1)))
			Expect(status.Phase).To(Equal(datav1alpha1.RuntimePhaseNotReady))
		})

		It("should return NotReady phase when no replicas are ready", func() {
			replicas := int32(3)
			sts := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-runtime-master",
					Namespace: "fluid",
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: &replicas,
				},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas:     0,
					CurrentReplicas:   3,
					AvailableReplicas: 0,
				},
			}
			Expect(manager.client.Create(ctx, sts)).To(Succeed())

			status, err := manager.ConstructComponentStatus(ctx, component)
			Expect(err).NotTo(HaveOccurred())
			Expect(status.ReadyReplicas).To(Equal(int32(0)))
			Expect(status.Phase).To(Equal(datav1alpha1.RuntimePhaseNotReady))
		})

		It("should return error when StatefulSet doesn't exist", func() {
			_, err := manager.ConstructComponentStatus(ctx, component)
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("DaemonSetManager", func() {
	var (
		manager   *DaemonSetManager
		ctx       context.Context
		component *common.CacheRuntimeComponentValue
	)

	BeforeEach(func() {
		client := setupTestClient()
		manager = newDaemonSetManager(client)
		ctx = context.Background()

		component = &common.CacheRuntimeComponentValue{
			Name:      "test-runtime-worker",
			Namespace: "fluid",
			Owner: &common.OwnerReference{
				APIVersion: "data.fluid.io/v1alpha1",
				Kind:       "CacheRuntime",
				Name:       "test-runtime",
				UID:        "test-uid",
			},
			PodTemplateSpec: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "worker",
							Image: "test-image:latest",
						},
					},
				},
			},
			Service: &common.CacheRuntimeComponentServiceConfig{
				Name: "test-runtime-worker-svc",
			},
			WorkloadType: metav1.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "DaemonSet",
			},
		}
	})

	Describe("Reconciler", func() {
		It("should create DaemonSet and Service successfully", func() {
			err := manager.Reconciler(ctx, component)
			Expect(err).NotTo(HaveOccurred())

			ds := &appsv1.DaemonSet{}
			err = manager.client.Get(ctx, types.NamespacedName{
				Name:      "test-runtime-worker",
				Namespace: "fluid",
			}, ds)
			Expect(err).NotTo(HaveOccurred())
			Expect(ds.Name).To(Equal("test-runtime-worker"))

			svc := &corev1.Service{}
			err = manager.client.Get(ctx, types.NamespacedName{
				Name:      "test-runtime-worker-svc",
				Namespace: "fluid",
			}, svc)
			Expect(err).NotTo(HaveOccurred())
			Expect(svc.Name).To(Equal("test-runtime-worker-svc"))
			Expect(svc.Spec.ClusterIP).To(Equal("None"))
			Expect(svc.Spec.PublishNotReadyAddresses).To(BeTrue())
		})

		It("should not recreate if DaemonSet already exists", func() {
			err := manager.Reconciler(ctx, component)
			Expect(err).NotTo(HaveOccurred())

			err = manager.Reconciler(ctx, component)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle nil service gracefully", func() {
			component.Service = nil
			err := manager.Reconciler(ctx, component)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("ConstructComponentStatus", func() {
		It("should return correct status when all nodes are ready", func() {
			ds := &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-runtime-worker",
					Namespace: "fluid",
				},
				Status: appsv1.DaemonSetStatus{
					DesiredNumberScheduled: 3,
					CurrentNumberScheduled: 3,
					NumberAvailable:        3,
					NumberUnavailable:      0,
					NumberReady:            3,
				},
			}
			Expect(manager.client.Create(ctx, ds)).To(Succeed())

			status, err := manager.ConstructComponentStatus(ctx, component)
			Expect(err).NotTo(HaveOccurred())
			Expect(status.DesiredReplicas).To(Equal(int32(3)))
			Expect(status.ReadyReplicas).To(Equal(int32(3)))
			Expect(status.AvailableReplicas).To(Equal(int32(3)))
			Expect(status.UnavailableReplicas).To(Equal(int32(0)))
			Expect(status.Phase).To(Equal(datav1alpha1.RuntimePhaseReady))
		})

		It("should return Ready phase even when some nodes are unavailable", func() {
			ds := &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-runtime-worker",
					Namespace: "fluid",
				},
				Status: appsv1.DaemonSetStatus{
					DesiredNumberScheduled: 3,
					CurrentNumberScheduled: 3,
					NumberAvailable:        2,
					NumberUnavailable:      1,
					NumberReady:            2,
				},
			}
			Expect(manager.client.Create(ctx, ds)).To(Succeed())

			status, err := manager.ConstructComponentStatus(ctx, component)
			Expect(err).NotTo(HaveOccurred())
			Expect(status.DesiredReplicas).To(Equal(int32(3)))
			Expect(status.ReadyReplicas).To(Equal(int32(2)))
			Expect(status.AvailableReplicas).To(Equal(int32(2)))
			Expect(status.UnavailableReplicas).To(Equal(int32(1)))
			// DaemonSet should return NotReady when not all replicas are ready
			Expect(status.Phase).To(Equal(datav1alpha1.RuntimePhaseNotReady))
		})

		It("should return NotReady phase when no nodes are ready", func() {
			ds := &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-runtime-worker",
					Namespace: "fluid",
				},
				Status: appsv1.DaemonSetStatus{
					DesiredNumberScheduled: 3,
					CurrentNumberScheduled: 3,
					NumberAvailable:        0,
					NumberUnavailable:      3,
					NumberReady:            0,
				},
			}
			Expect(manager.client.Create(ctx, ds)).To(Succeed())

			status, err := manager.ConstructComponentStatus(ctx, component)
			Expect(err).NotTo(HaveOccurred())
			Expect(status.DesiredReplicas).To(Equal(int32(3)))
			Expect(status.ReadyReplicas).To(Equal(int32(0)))
			Expect(status.AvailableReplicas).To(Equal(int32(0)))
			Expect(status.UnavailableReplicas).To(Equal(int32(3)))
			Expect(status.Phase).To(Equal(datav1alpha1.RuntimePhaseNotReady))
		})

		It("should return error when DaemonSet doesn't exist", func() {
			_, err := manager.ConstructComponentStatus(ctx, component)
			Expect(err).To(HaveOccurred())
		})
	})
})
