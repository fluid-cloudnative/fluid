/*
Copyright 2023 The Fluid Author.

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
	"fmt"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Volume related unit tests", Label("pkg.utils.kubeclient.volume_test.go"), func() {
	var (
		client     client.Client
		resources  []runtime.Object
		testScheme *runtime.Scheme
	)

	JustBeforeEach(func() {
		testScheme = runtime.NewScheme()
		_ = v1.AddToScheme(testScheme)
		client = fake.NewFakeClientWithScheme(testScheme, resources...)
	})

	Describe("Test GetPersistentVolume()", func() {
		var (
			pv *v1.PersistentVolume
		)

		BeforeEach(func() {
			pv = &v1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pv",
				},
				Spec: v1.PersistentVolumeSpec{},
			}
		})

		Context("when persistent volume exists", func() {
			BeforeEach(func() {
				resources = []runtime.Object{pv}
			})

			It("should return the persistent volume successfully", func() {
				gotPV, err := GetPersistentVolume(client, pv.Name)
				Expect(err).NotTo(HaveOccurred())
				Expect(gotPV).NotTo(BeNil())
				Expect(gotPV.Name).To(Equal(pv.Name))
			})
		})

		Context("when persistent volume does not exist", func() {
			BeforeEach(func() {
				resources = []runtime.Object{}
			})

			It("should return not found error", func() {
				gotPV, err := GetPersistentVolume(client, "not-exist-pv")
				Expect(err).To(HaveOccurred())
				Expect(gotPV).To(BeNil())
				Expect(apierrs.IsNotFound(err)).To(BeTrue())
			})
		})
	})

	Describe("Test IsPersistentVolumeExist()", func() {
		var (
			pv *v1.PersistentVolume
		)

		BeforeEach(func() {
			pv = &v1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pv",
					Annotations: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
				Spec: v1.PersistentVolumeSpec{},
			}
		})

		Context("when persistent volume exists with matching annotations", func() {
			BeforeEach(func() {
				resources = []runtime.Object{pv}
			})

			It("should return true", func() {
				annotations := map[string]string{
					"key1": "value1",
					"key2": "value2",
				}
				found, err := IsPersistentVolumeExist(client, pv.Name, annotations)
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
			})
		})

		Context("when persistent volume exists without matching annotations", func() {
			BeforeEach(func() {
				resources = []runtime.Object{pv}
			})

			It("should return false", func() {
				annotations := map[string]string{
					"key1": "value1",
					"key3": "value3", // This key doesn't exist in PV
				}
				found, err := IsPersistentVolumeExist(client, pv.Name, annotations)
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Context("when persistent volume does not exist", func() {
			BeforeEach(func() {
				resources = []runtime.Object{}
			})

			It("should return false", func() {
				annotations := map[string]string{
					"key1": "value1",
				}
				found, err := IsPersistentVolumeExist(client, "not-exist-pv", annotations)
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})
	})

	Describe("Test IsPersistentVolumeClaimExist()", func() {
		var (
			pvc       *v1.PersistentVolumeClaim
			namespace string
		)

		BeforeEach(func() {
			namespace = "test-ns"
			pvc = &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pvc",
					Namespace: namespace,
					Annotations: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
				Spec: v1.PersistentVolumeClaimSpec{},
			}
		})

		Context("when persistent volume claim exists with matching annotations", func() {
			BeforeEach(func() {
				resources = []runtime.Object{pvc}
			})

			It("should return true", func() {
				annotations := map[string]string{
					"key1": "value1",
					"key2": "value2",
				}
				found, err := IsPersistentVolumeClaimExist(client, pvc.Name, namespace, annotations)
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
			})
		})

		Context("when persistent volume claim exists without matching annotations", func() {
			BeforeEach(func() {
				resources = []runtime.Object{pvc}
			})

			It("should return false", func() {
				annotations := map[string]string{
					"key1": "value1",
					"key3": "value3", // This key doesn't exist in PVC
				}
				found, err := IsPersistentVolumeClaimExist(client, pvc.Name, namespace, annotations)
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Context("when persistent volume claim does not exist", func() {
			BeforeEach(func() {
				resources = []runtime.Object{}
			})

			It("should return false", func() {
				annotations := map[string]string{
					"key1": "value1",
				}
				found, err := IsPersistentVolumeClaimExist(client, "not-exist-pvc", namespace, annotations)
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})
	})

	Describe("Test DeletePersistentVolume()", func() {
		var (
			pv *v1.PersistentVolume
		)

		BeforeEach(func() {
			pv = &v1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pv",
				},
				Spec: v1.PersistentVolumeSpec{},
			}
		})

		Context("when persistent volume exists", func() {
			BeforeEach(func() {
				resources = []runtime.Object{pv}
			})

			It("should delete the persistent volume successfully", func() {
				err := DeletePersistentVolume(client, pv.Name)
				Expect(err).NotTo(HaveOccurred())

				// Verify PV is deleted
				_, err = GetPersistentVolume(client, pv.Name)
				Expect(err).To(HaveOccurred())
				Expect(apierrs.IsNotFound(err)).To(BeTrue())
			})
		})

		Context("when persistent volume does not exist", func() {
			BeforeEach(func() {
				resources = []runtime.Object{}
			})

			It("should not return error", func() {
				err := DeletePersistentVolume(client, "not-exist-pv")
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Test DeletePersistentVolumeClaim()", func() {
		var (
			pvc       *v1.PersistentVolumeClaim
			namespace string
		)

		BeforeEach(func() {
			namespace = "test-ns"
			pvc = &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pvc",
					Namespace: namespace,
				},
				Spec: v1.PersistentVolumeClaimSpec{},
			}
		})

		Context("when persistent volume claim exists", func() {
			BeforeEach(func() {
				resources = []runtime.Object{pvc}
			})

			It("should delete the persistent volume claim successfully", func() {
				err := DeletePersistentVolumeClaim(client, pvc.Name, namespace)
				Expect(err).NotTo(HaveOccurred())

				// Verify PVC is deleted
				key := types.NamespacedName{
					Name:      pvc.Name,
					Namespace: namespace,
				}
				gotPVC := &v1.PersistentVolumeClaim{}
				err = client.Get(context.TODO(), key, gotPVC)
				Expect(err).To(HaveOccurred())
				Expect(apierrs.IsNotFound(err)).To(BeTrue())
			})
		})

		Context("when persistent volume claim does not exist", func() {
			BeforeEach(func() {
				resources = []runtime.Object{}
			})

			It("should not return error", func() {
				err := DeletePersistentVolumeClaim(client, "not-exist-pvc", namespace)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when persistent volume claim is already being deleted", func() {
			BeforeEach(func() {
				now := metav1.Now()
				pvc.DeletionTimestamp = &now
				pvc.Finalizers = []string{"kubernetes.io/pvc-protection"}
				resources = []runtime.Object{pvc}
			})

			It("should skip deletion", func() {
				err := DeletePersistentVolumeClaim(client, pvc.Name, namespace)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Test GetPVCsFromPod()", func() {
		var pod *v1.Pod

		BeforeEach(func() {
			pod = &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-ns",
				},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							Name: "hostpath-volume",
							VolumeSource: v1.VolumeSource{
								HostPath: &v1.HostPathVolumeSource{
									Path: "/tmp/test",
								},
							},
						},
						{
							Name: "pvc-volume-1",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: "test-pvc-1",
								},
							},
						},
						{
							Name: "pvc-volume-2",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: "test-pvc-2",
								},
							},
						},
					},
				},
			}
		})

		Context("when pod has persistent volume claims", func() {
			It("should return the list of PVC volumes", func() {
				pvcs := GetPVCsFromPod(*pod)
				Expect(len(pvcs)).To(Equal(2))
				Expect(pvcs[0].Name).To(Equal("pvc-volume-1"))
				Expect(pvcs[0].PersistentVolumeClaim.ClaimName).To(Equal("test-pvc-1"))
				Expect(pvcs[1].Name).To(Equal("pvc-volume-2"))
				Expect(pvcs[1].PersistentVolumeClaim.ClaimName).To(Equal("test-pvc-2"))
			})
		})

		Context("when pod has no persistent volume claims", func() {
			BeforeEach(func() {
				pod.Spec.Volumes = []v1.Volume{
					{
						Name: "hostpath-volume",
						VolumeSource: v1.VolumeSource{
							HostPath: &v1.HostPathVolumeSource{
								Path: "/tmp/test",
							},
						},
					},
				}
			})

			It("should return empty list", func() {
				pvcs := GetPVCsFromPod(*pod)
				Expect(len(pvcs)).To(Equal(0))
			})
		})
	})

	Describe("Test GetPvcMountPods()", func() {
		var (
			namespace string
			pvcName   string
			pod1      *v1.Pod
			pod2      *v1.Pod
			pod3      *v1.Pod
		)

		BeforeEach(func() {
			namespace = "test-ns"
			pvcName = "test-pvc"

			pod1 = &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-with-pvc",
					Namespace: namespace,
				},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							Name: "pvc-volume",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvcName,
								},
							},
						},
					},
				},
			}

			pod2 = &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-without-pvc",
					Namespace: namespace,
				},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							Name: "hostpath-volume",
							VolumeSource: v1.VolumeSource{
								HostPath: &v1.HostPathVolumeSource{
									Path: "/tmp/test",
								},
							},
						},
					},
				},
			}

			pod3 = &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-with-different-pvc",
					Namespace: namespace,
				},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							Name: "pvc-volume",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: "different-pvc",
								},
							},
						},
					},
				},
			}
		})

		Context("when pods are mounting the persistent volume claim", func() {
			BeforeEach(func() {
				resources = []runtime.Object{pod1, pod2, pod3}
			})

			It("should return the list of mounting pods", func() {
				pods, err := GetPvcMountPods(client, pvcName, namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(pods)).To(Equal(1))
				Expect(pods[0].Name).To(Equal("pod-with-pvc"))
			})
		})

		Context("when no pods are mounting the persistent volume claim", func() {
			BeforeEach(func() {
				pod2.Spec.Volumes = pod1.Spec.Volumes // Make pod2 also use the PVC
				pod3.Spec.Volumes = pod1.Spec.Volumes // Make pod3 also use the PVC
				resources = []runtime.Object{pod1, pod2, pod3}
			})

			It("should return empty list", func() {
				pods, err := GetPvcMountPods(client, "non-existent-pvc", namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(pods)).To(Equal(0))
			})
		})
	})

	Describe("Test GetPvcMountNodes()", func() {
		var (
			namespace string
			pvcName   string
			pod1      *v1.Pod
			pod2      *v1.Pod
			pod3      *v1.Pod
		)

		BeforeEach(func() {
			namespace = "test-ns"
			pvcName = "test-pvc"

			pod1 = &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "running-pod",
					Namespace: namespace,
				},
				Spec: v1.PodSpec{
					NodeName: "node-1",
					Volumes: []v1.Volume{
						{
							Name: "pvc-volume",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvcName,
								},
							},
						},
					},
				},
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
				},
			}

			pod2 = &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "completed-pod",
					Namespace: namespace,
				},
				Spec: v1.PodSpec{
					NodeName: "node-2",
					Volumes: []v1.Volume{
						{
							Name: "pvc-volume",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvcName,
								},
							},
						},
					},
				},
				Status: v1.PodStatus{
					Phase: v1.PodSucceeded,
				},
			}

			pod3 = &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-without-pvc",
					Namespace: namespace,
				},
				Spec: v1.PodSpec{
					NodeName: "node-3",
					Volumes: []v1.Volume{
						{
							Name: "hostpath-volume",
							VolumeSource: v1.VolumeSource{
								HostPath: &v1.HostPathVolumeSource{
									Path: "/tmp/test",
								},
							},
						},
					},
				},
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
				},
			}
		})

		Context("when pods are mounting the persistent volume claim on nodes", func() {
			BeforeEach(func() {
				resources = []runtime.Object{pod1, pod2, pod3}
			})

			It("should return the map of nodes and pod counts", func() {
				nodes, err := GetPvcMountNodes(client, pvcName, namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(nodes)).To(Equal(1)) // Only running pod should be counted
				Expect(nodes["node-1"]).To(Equal(int64(1)))
			})
		})

		Context("when no pods are mounting the persistent volume claim", func() {
			BeforeEach(func() {
				resources = []runtime.Object{pod3} // Only pod without PVC
			})

			It("should return empty map", func() {
				nodes, err := GetPvcMountNodes(client, pvcName, namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(nodes)).To(Equal(0))
			})
		})

		Context("when only completed pods are mounting the persistent volume claim", func() {
			BeforeEach(func() {
				pod1.Status.Phase = v1.PodSucceeded // Make running pod completed
				resources = []runtime.Object{pod1, pod3}
			})

			It("should return empty map", func() {
				nodes, err := GetPvcMountNodes(client, pvcName, namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(nodes)).To(Equal(0))
			})
		})
	})

	Describe("Test RemoveProtectionFinalizer()", func() {
		var (
			namespace string
			pvc       *v1.PersistentVolumeClaim
		)

		BeforeEach(func() {
			namespace = "test-ns"
			pvc = &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pvc",
					Namespace: namespace,
					Finalizers: []string{
						"kubernetes.io/pvc-protection",
						"another-finalizer",
					},
				},
				Spec: v1.PersistentVolumeClaimSpec{},
			}
		})

		Context("when persistent volume claim has protection finalizer", func() {
			BeforeEach(func() {
				resources = []runtime.Object{pvc}
			})

			It("should remove the protection finalizer", func() {
				err := RemoveProtectionFinalizer(client, pvc.Name, namespace)
				Expect(err).NotTo(HaveOccurred())

				// Verify finalizer is removed
				key := types.NamespacedName{
					Name:      pvc.Name,
					Namespace: namespace,
				}
				updatedPVC := &v1.PersistentVolumeClaim{}
				err = client.Get(context.TODO(), key, updatedPVC)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedPVC.Finalizers).To(ContainElement("another-finalizer"))
				Expect(updatedPVC.Finalizers).NotTo(ContainElement("kubernetes.io/pvc-protection"))
			})
		})

		Context("when persistent volume claim does not exist", func() {
			BeforeEach(func() {
				resources = []runtime.Object{}
			})

			It("should return error", func() {
				err := RemoveProtectionFinalizer(client, "non-existent-pvc", namespace)
				Expect(err).To(HaveOccurred())
				Expect(apierrs.IsNotFound(err)).To(BeTrue())
			})
		})
	})

	Describe("Test ShouldDeleteDataset()", func() {
		var (
			namespace string
			pvcName   string
			pvc       *v1.PersistentVolumeClaim
			pod       *v1.Pod
		)

		BeforeEach(func() {
			namespace = "test-ns"
			pvcName = "test-pvc"
			pvc = &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:        pvcName,
					Namespace:   namespace,
					Annotations: common.GetExpectedFluidAnnotations(),
				},
				Spec: v1.PersistentVolumeClaimSpec{},
			}

			pod = &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: namespace,
				},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							Name: "pvc-volume",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvcName,
								},
							},
						},
					},
				},
			}
		})

		Context("when persistent volume claim does not exist", func() {
			BeforeEach(func() {
				resources = []runtime.Object{}
			})

			It("should not return error", func() {
				err := ShouldDeleteDataset(client, "non-existent-pvc", namespace)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when no pods are using the persistent volume claim", func() {
			BeforeEach(func() {
				resources = []runtime.Object{pvc}
			})

			It("should not return error", func() {
				err := ShouldDeleteDataset(client, pvcName, namespace)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when running pods are using the persistent volume claim", func() {
			BeforeEach(func() {
				pod.Status.Phase = v1.PodRunning
				resources = []runtime.Object{pvc, pod}
			})

			It("should return error", func() {
				err := ShouldDeleteDataset(client, pvcName, namespace)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("is using it"))
			})
		})

		Context("when only completed pods are using the persistent volume claim", func() {
			BeforeEach(func() {
				pod.Status.Phase = v1.PodSucceeded
				resources = []runtime.Object{pvc, pod}
			})

			It("should not return error", func() {
				err := ShouldDeleteDataset(client, pvcName, namespace)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Test ShouldRemoveProtectionFinalizer()", func() {
		var (
			namespace string
			pvcName   string
			pvc       *v1.PersistentVolumeClaim
			now       metav1.Time
		)

		BeforeEach(func() {
			namespace = "test-ns"
			pvcName = "test-pvc"
			now = metav1.Now()

			pvc = &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pvcName,
					Namespace: namespace,
					Finalizers: []string{
						"kubernetes.io/pvc-protection",
					},
				},
				Spec: v1.PersistentVolumeClaimSpec{},
			}
		})

		Context("when persistent volume claim is not in terminating state", func() {
			BeforeEach(func() {
				resources = []runtime.Object{pvc}
			})

			It("should return false", func() {
				should, err := ShouldRemoveProtectionFinalizer(client, pvcName, namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(should).To(BeFalse())
			})
		})

		Context("when persistent volume claim has no protection finalizer", func() {
			BeforeEach(func() {
				pvc.Finalizers = []string{"placeholder-finalizer"}
				pvc.DeletionTimestamp = &now
				resources = []runtime.Object{pvc}
			})

			It("should return false", func() {
				should, err := ShouldRemoveProtectionFinalizer(client, pvcName, namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(should).To(BeFalse())
			})
		})

		Context("when persistent volume claim is in terminating state for less than 30 seconds", func() {
			BeforeEach(func() {
				pvc.DeletionTimestamp = &now
				resources = []runtime.Object{pvc}
			})

			It("should return false", func() {
				should, err := ShouldRemoveProtectionFinalizer(client, pvcName, namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(should).To(BeFalse())
			})
		})

		Context("when persistent volume claim is in terminating state for more than 30 seconds and no active pods", func() {
			BeforeEach(func() {
				then := metav1.NewTime(now.Add(-31 * time.Second)) // 31 seconds ago
				pvc.DeletionTimestamp = &then
				resources = []runtime.Object{pvc}
			})

			It("should return true", func() {
				should, err := ShouldRemoveProtectionFinalizer(client, pvcName, namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(should).To(BeTrue())
			})
		})

		Context("when persistent volume claim is in terminating state for more than 30 seconds but active pods exist", func() {
			var pod *v1.Pod

			BeforeEach(func() {
				then := metav1.NewTime(now.Add(-31 * time.Second)) // 31 seconds ago
				pvc.DeletionTimestamp = &then

				pod = &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "active-pod",
						Namespace: namespace,
					},
					Spec: v1.PodSpec{
						Volumes: []v1.Volume{
							{
								Name: "pvc-volume",
								VolumeSource: v1.VolumeSource{
									PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
										ClaimName: pvcName,
									},
								},
							},
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				}

				resources = []runtime.Object{pvc, pod}
			})

			It("should return false", func() {
				should, err := ShouldRemoveProtectionFinalizer(client, pvcName, namespace)
				Expect(err).To(HaveOccurred())
				Expect(should).To(BeFalse())
			})
		})
	})

	Describe("Test CheckIfPVCIsDataset()", func() {
		var pvc *v1.PersistentVolumeClaim

		BeforeEach(func() {
			pvc = &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pvc",
					Namespace: "test-ns",
				},
				Spec: v1.PersistentVolumeClaimSpec{},
			}
		})

		Context("when pvc has storage capacity label prefix", func() {
			BeforeEach(func() {
				pvc.Labels = map[string]string{
					fmt.Sprintf("%s-test-ns-mydataset", common.LabelAnnotationStorageCapacityPrefix): "true",
				}
			})

			It("should return true", func() {
				isDataset := CheckIfPVCIsDataset(pvc)
				Expect(isDataset).To(BeTrue())
			})
		})

		Context("when pvc has manager dataset label", func() {
			BeforeEach(func() {
				pvc.Labels = map[string]string{
					common.LabelAnnotationManagedBy: "foo",
				}
			})

			It("should return true", func() {
				isDataset := CheckIfPVCIsDataset(pvc)
				Expect(isDataset).To(BeTrue())
			})
		})

		Context("when pvc has no dataset-related labels", func() {
			BeforeEach(func() {
				pvc.Labels = map[string]string{
					"app": "test",
				}
			})

			It("should return false", func() {
				isDataset := CheckIfPVCIsDataset(pvc)
				Expect(isDataset).To(BeFalse())
			})
		})

		Context("when pvc is nil", func() {
			It("should return false", func() {
				isDataset := CheckIfPVCIsDataset(nil)
				Expect(isDataset).To(BeFalse())
			})
		})
	})

	Describe("Test GetReferringDatasetPVCInfo()", func() {
		var pvc *v1.PersistentVolumeClaim

		BeforeEach(func() {
			pvc = &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pvc",
					Namespace: "ref",
				},
				Spec: v1.PersistentVolumeClaimSpec{},
			}
		})

		Context("when pvc is referring to a dataset", func() {
			BeforeEach(func() {
				pvc.Labels = map[string]string{
					common.LabelAnnotationDatasetReferringName:      "dataset",
					common.LabelAnnotationDatasetReferringNameSpace: "fluid",
				}
			})

			It("should return ok as true with name and namespace", func() {
				ok, name, namespace := GetReferringDatasetPVCInfo(pvc)
				Expect(ok).To(BeTrue())
				Expect(name).To(Equal("dataset"))
				Expect(namespace).To(Equal("fluid"))
			})
		})

		Context("when pvc is not referring to any dataset", func() {
			BeforeEach(func() {
				pvc.Labels = map[string]string{}
			})

			It("should return ok as false", func() {
				ok, name, namespace := GetReferringDatasetPVCInfo(pvc)
				Expect(ok).To(BeFalse())
				Expect(name).To(BeEmpty())
				Expect(namespace).To(BeEmpty())
			})
		})
	})

	Describe("Test SetPVCDeleteTimeout()", func() {
		Context("when setting pvc delete timeout", func() {
			It("should update the global timeout value", func() {
				originalTimeout := pvcDeleteTimeout
				defer func() {
					SetPVCDeleteTimeout(originalTimeout) // Restore original value
				}()

				newTimeout := 10 * time.Second
				SetPVCDeleteTimeout(newTimeout)
				Expect(pvcDeleteTimeout).To(Equal(newTimeout))
			})
		})
	})
})
