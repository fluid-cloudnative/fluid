/*
Copyright 2021 The Fluid Authors.

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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StatefulSet related unit tests", Label("pkg.utils.kubeclient.statefulset_test.go"), func() {
	var (
		client    client.Client
		resources []runtime.Object
	)

	JustBeforeEach(func() {
		client = fake.NewFakeClientWithScheme(testScheme, resources...)
	})

	Describe("Test GetStatefulSet()", func() {
		sts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-sts",
				Namespace: "test-ns",
			},
		}
		When("statefulset exists", func() {
			BeforeEach(func() {
				resources = []runtime.Object{sts}
			})
			It("should successfully get statefulset", func() {
				gotSts, err := GetStatefulSet(client, "test-sts", "test-ns")
				Expect(err).To(BeNil())
				Expect(gotSts).To(Equal(sts))
			})
		})

		When("statefulset does not exist", func() {
			BeforeEach(func() {
				resources = []runtime.Object{}
			})
			It("should return not found error", func() {
				_, err := GetStatefulSet(client, "not-exist", "test-ns")
				Expect(err).NotTo(BeNil())
				Expect(apierrs.IsNotFound(err)).To(BeTrue())
			})
		})
	})

	Describe("Test ScaleStatefulSet()", func() {
		sts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-sts",
				Namespace: "test-ns",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: new(int32),
			},
		}

		When("works as expected", func() {
			BeforeEach(func() {
				*sts.Spec.Replicas = 1
				resources = []runtime.Object{sts}
			})

			It("should successfully scale statefulset", func() {
				Expect(ScaleStatefulSet(client, "test-sts", "test-ns", 3)).To(Succeed())

				updatedSts := &appsv1.StatefulSet{}
				Expect(client.Get(context.TODO(), types.NamespacedName{
					Name:      "test-sts",
					Namespace: "test-ns",
				}, updatedSts)).To(Succeed())
				Expect(*updatedSts.Spec.Replicas).To(Equal(int32(3)))
			})
		})

		When("statefulset is not exists", func() {
			BeforeEach(func() {
				resources = []runtime.Object{}
			})

			It("should return not found error", func() {
				err := ScaleStatefulSet(client, "not-exist", "test-ns", 3)
				Expect(err).NotTo(BeNil())
				Expect(apierrs.IsNotFound(err)).To(BeTrue())
			})
		})
	})

	Describe("Test GetPodsForStatefulSet()", func() {
		sts := &appsv1.StatefulSet{
			TypeMeta: metav1.TypeMeta{
				Kind:       "StatefulSet",
				APIVersion: "apps/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-sts",
				Namespace: "test-ns",
				UID:       "test-sts-uid",
			},
			Spec: appsv1.StatefulSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "test",
					},
				},
			},
		}

		pod1 := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-sts-0",
				Namespace: "test-ns",
				Labels: map[string]string{
					"app": "test",
				},
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "apps/v1",
						Kind:       "StatefulSet",
						Name:       "test-sts",
						UID:        "test-sts-uid",
						Controller: ptr.To(true),
					},
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "test-container",
						Image: "nginx:1.0",
					},
				},
			},
		}

		pod2 := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-sts-1",
				Namespace: "test-ns",
				Labels: map[string]string{
					"app": "test",
				},
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "apps/v1",
						Kind:       "StatefulSet",
						Name:       "test-sts",
						UID:        "test-sts-uid",
						Controller: ptr.To(true),
					},
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "test-container",
						Image: "nginx:1.0",
					},
				},
			},
		}

		When("pods exist for statefulset", func() {
			BeforeEach(func() {
				resources = []runtime.Object{sts, pod1, pod2}
			})

			It("should successfully return pods", func() {
				selector, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
				Expect(err).To(BeNil())

				pods, err := GetPodsForStatefulSet(client, sts, selector)
				Expect(err).To(BeNil())
				Expect(len(pods)).To(Equal(2))
				// Note: Since we're using a fake client, the owner reference matching might not work exactly as in real cluster
			})
		})

		When("no pods exist for statefulset", func() {
			BeforeEach(func() {
				resources = []runtime.Object{sts}
			})

			It("should return empty pod list", func() {
				selector, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
				Expect(err).To(BeNil())

				pods, err := GetPodsForStatefulSet(client, sts, selector)
				Expect(err).To(BeNil())
				Expect(len(pods)).To(Equal(0))
			})
		})

		When("statefulset does not exist", func() {
			BeforeEach(func() {
				resources = []runtime.Object{pod1}
			})

			It("should return empty list without error", func() {
				stsNotExists := &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "not-exist",
						Namespace: "test-ns",
					},
					Spec: appsv1.StatefulSetSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "not-exist",
							},
						},
					},
				}

				selector, err := metav1.LabelSelectorAsSelector(stsNotExists.Spec.Selector)
				Expect(err).To(BeNil())

				pods, err := GetPodsForStatefulSet(client, stsNotExists, selector)
				Expect(err).To(BeNil())
				Expect(len(pods)).To(Equal(0))
			})
		})
	})

	Describe("Test GetPhaseFromStatefulset()", func() {
		When("statefulset has 0 replicas", func() {
			It("should return RuntimePhaseReady", func() {
				sts := appsv1.StatefulSet{
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 0,
					},
				}
				phase := GetPhaseFromStatefulset(0, sts)
				Expect(phase).To(Equal(datav1alpha1.RuntimePhaseReady))
			})
		})

		When("statefulset has ready replicas equal to desired replicas", func() {
			It("should return RuntimePhaseReady", func() {
				sts := appsv1.StatefulSet{
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 2,
					},
				}
				phase := GetPhaseFromStatefulset(2, sts)
				Expect(phase).To(Equal(datav1alpha1.RuntimePhaseReady))
			})
		})

		When("statefulset has ready replicas less than desired replicas", func() {
			It("should return RuntimePhasePartialReady when some are ready", func() {
				sts := appsv1.StatefulSet{
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 1,
					},
				}
				phase := GetPhaseFromStatefulset(3, sts)
				Expect(phase).To(Equal(datav1alpha1.RuntimePhasePartialReady))
			})

			It("should return RuntimePhaseNotReady when none are ready", func() {
				sts := appsv1.StatefulSet{
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas: 0,
					},
				}
				phase := GetPhaseFromStatefulset(2, sts)
				Expect(phase).To(Equal(datav1alpha1.RuntimePhaseNotReady))
			})
		})
	})

	Describe("Test getParentNameAndOrdinal()", func() {
		When("pod name matches statefulset pattern", func() {
			It("should extract parent name and ordinal correctly", func() {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-sts-1",
					},
				}
				parent, ordinal := getParentNameAndOrdinal(pod)
				Expect(parent).To(Equal("test-sts"))
				Expect(ordinal).To(Equal(1))
			})
		})

		When("pod name does not match statefulset pattern", func() {
			It("should return empty parent and -1 ordinal", func() {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-pod",
					},
				}
				parent, ordinal := getParentNameAndOrdinal(pod)
				Expect(parent).To(Equal(""))
				Expect(ordinal).To(Equal(-1))
			})
		})
	})

	Describe("Test getParentName()", func() {
		When("pod has parent statefulset", func() {
			It("should return parent name", func() {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-sts-1",
					},
				}
				parent := getParentName(pod)
				Expect(parent).To(Equal("test-sts"))
			})
		})

		When("pod does not have parent statefulset", func() {
			It("should return empty string", func() {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-pod",
					},
				}
				parent := getParentName(pod)
				Expect(parent).To(Equal(""))
			})
		})
	})

	Describe("Test isMemberOf()", func() {
		sts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-sts",
			},
		}

		When("pod is member of statefulset", func() {
			It("should return true", func() {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-sts-0",
					},
				}
				isMember := isMemberOf(sts, pod)
				Expect(isMember).To(BeTrue())
			})
		})

		When("pod is not member of statefulset", func() {
			It("should return false", func() {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "other-sts-0",
					},
				}
				isMember := isMemberOf(sts, pod)
				Expect(isMember).To(BeFalse())
			})
		})
	})
})
