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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Use fake client because of it will be maintained in the long term
// due to https://github.com/kubernetes-sigs/controller-runtime/pull/1101
var _ = Describe("IsPersistentVolumeClaimExist", func() {
	var (
		namespace     string
		testPVCInputs []*v1.PersistentVolumeClaim
		client        client.Client
	)

	BeforeEach(func() {
		namespace = "default"
		testPVCInputs = []*v1.PersistentVolumeClaim{{
			ObjectMeta: metav1.ObjectMeta{Name: "notCreatedByFluid",
				Namespace: namespace},
			Spec: v1.PersistentVolumeClaimSpec{},
		}, {
			ObjectMeta: metav1.ObjectMeta{Name: "createdByFluid",
				Annotations: common.GetExpectedFluidAnnotations(),
				Namespace:   namespace},
			Spec: v1.PersistentVolumeClaimSpec{},
		}}

		testPVCs := []runtime.Object{}
		for _, pvc := range testPVCInputs {
			testPVCs = append(testPVCs, pvc.DeepCopy())
		}

		client = fake.NewFakeClientWithScheme(testScheme, testPVCs...)
	})

	It("should return false when volume doesn't exist", func() {
		got, _ := IsPersistentVolumeClaimExist(client, "notExist", namespace, map[string]string{})
		Expect(got).To(BeFalse())
	})

	It("should return false when volume is not created by fluid", func() {
		got, _ := IsPersistentVolumeClaimExist(client, "notCreatedByFluid", namespace, map[string]string{})
		Expect(got).To(BeFalse())
	})

	It("should return true when volume is created by fluid", func() {
		got, _ := IsPersistentVolumeClaimExist(client, "createdByFluid", namespace, common.GetExpectedFluidAnnotations())
		Expect(got).To(BeTrue())
	})

	It("should return false when volume is not created by fluid with different annotations", func() {
		got, _ := IsPersistentVolumeClaimExist(client, "notCreatedByFluid2", "", map[string]string{
			"test1": "test1",
		})
		Expect(got).To(BeFalse())
	})
})

var _ = Describe("DeletePersistentVolumeClaim", func() {
	var (
		namespace     string
		testPVCInputs []*v1.PersistentVolumeClaim
		client        client.Client
	)

	BeforeEach(func() {
		namespace = "default"
		testPVCInputs = []*v1.PersistentVolumeClaim{{
			ObjectMeta: metav1.ObjectMeta{Name: "aaa",
				Namespace: namespace},
			Spec: v1.PersistentVolumeClaimSpec{},
		}, {
			ObjectMeta: metav1.ObjectMeta{Name: "bbb",
				Annotations: common.GetExpectedFluidAnnotations(),
				Namespace:   namespace},
			Spec: v1.PersistentVolumeClaimSpec{},
		}}

		testPVCs := []runtime.Object{}
		for _, pvc := range testPVCInputs {
			testPVCs = append(testPVCs, pvc.DeepCopy())
		}

		client = fake.NewFakeClientWithScheme(testScheme, testPVCs...)
	})

	It("should not error when volume doesn't exist", func() {
		err := DeletePersistentVolumeClaim(client, "notfound", namespace)
		Expect(err).To(BeNil())
	})

	It("should not error when volume exists", func() {
		err := DeletePersistentVolumeClaim(client, "found", namespace)
		Expect(err).To(BeNil())
	})
})

var _ = Describe("GetPvcMountNodes", func() {
	var (
		namespace     string
		volumeName1   string
		volumeName2   string
		testPodInputs []*v1.Pod
		client        client.Client
	)

	BeforeEach(func() {
		namespace = "test"
		volumeName1 = "found"
		volumeName2 = "found1"
		testPodInputs = []*v1.Pod{{
			ObjectMeta: metav1.ObjectMeta{Name: "found"},
			Spec:       v1.PodSpec{},
		}, {
			ObjectMeta: metav1.ObjectMeta{Name: "bbb", Namespace: namespace},
			Spec: v1.PodSpec{
				Volumes: []v1.Volume{
					{
						Name: volumeName1,
						VolumeSource: v1.VolumeSource{
							PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
								ClaimName: volumeName1,
								ReadOnly:  true,
							}},
					},
				},
				NodeName: "node1",
			},
			Status: v1.PodStatus{
				Phase: v1.PodSucceeded,
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{Name: "ccc", Namespace: namespace},
			Spec: v1.PodSpec{
				Volumes: []v1.Volume{
					{
						Name: volumeName1,
						VolumeSource: v1.VolumeSource{
							PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
								ClaimName: volumeName1,
								ReadOnly:  true,
							}},
					},
				},
				NodeName: "node2",
			},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{Name: "ddd", Namespace: namespace},
			Spec: v1.PodSpec{
				Volumes: []v1.Volume{
					{
						Name: volumeName1,
						VolumeSource: v1.VolumeSource{
							PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
								ClaimName: volumeName1,
								ReadOnly:  true,
							}},
					},
				},
				NodeName: "node3",
			},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{Name: "eee", Namespace: namespace},
			Spec: v1.PodSpec{
				Volumes: []v1.Volume{
					{
						Name: volumeName2,
						VolumeSource: v1.VolumeSource{
							PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
								ClaimName: volumeName2,
								ReadOnly:  true,
							}},
					},
				},
				NodeName: "node4",
			},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{Name: "fff", Namespace: namespace},
			Spec: v1.PodSpec{
				Volumes: []v1.Volume{
					{
						Name: volumeName2,
						VolumeSource: v1.VolumeSource{
							PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
								ClaimName: volumeName2,
								ReadOnly:  true,
							}},
					},
				},
				NodeName: "",
			},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{Name: "hhh", Namespace: namespace},
			Spec: v1.PodSpec{
				Volumes: []v1.Volume{
					{
						Name: volumeName2,
						VolumeSource: v1.VolumeSource{
							PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
								ClaimName: volumeName1,
								ReadOnly:  true,
							}},
					},
				},
				NodeName: "node3",
			},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
			},
		}}

		testPods := []runtime.Object{}
		for _, pod := range testPodInputs {
			testPods = append(testPods, pod.DeepCopy())
		}

		client = fake.NewFakeClientWithScheme(testScheme, testPods...)
	})

	It("should return empty list when node list is empty", func() {
		pvcMountNodes, _ := GetPvcMountNodes(client, "not found", namespace)
		Expect(len(pvcMountNodes)).To(Equal(0))
	})

	It("should return 1 node when node list is 1", func() {
		pvcMountNodes, _ := GetPvcMountNodes(client, volumeName2, namespace)
		Expect(len(pvcMountNodes)).To(Equal(1))
	})

	It("should return 2 nodes when node list is 2", func() {
		pvcMountNodes, _ := GetPvcMountNodes(client, volumeName1, namespace)
		Expect(len(pvcMountNodes)).To(Equal(2))
	})
})

var _ = Describe("RemoveProtectionFinalizer", func() {
	var (
		namespace     string
		testPVCInputs []*v1.PersistentVolumeClaim
		client        client.Client
	)

	BeforeEach(func() {
		namespace = "default"
		testPVCInputs = []*v1.PersistentVolumeClaim{{
			ObjectMeta: metav1.ObjectMeta{Name: "hasNoFinalizer",
				Namespace: namespace},
			Spec: v1.PersistentVolumeClaimSpec{
				VolumeName: "hasNoFinalizer",
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{Name: "hasFinalizer",
				Annotations: common.GetExpectedFluidAnnotations(),
				Namespace:   namespace,
				Finalizers:  []string{persistentVolumeClaimProtectionFinalizerName}},
			Spec: v1.PersistentVolumeClaimSpec{
				VolumeName: "hasFinalizer",
			},
		}}

		testPVCs := []runtime.Object{}
		for _, pvc := range testPVCInputs {
			testPVCs = append(testPVCs, pvc.DeepCopy())
		}

		client = fake.NewFakeClientWithScheme(testScheme, testPVCs...)
	})

	It("should error when volumeClaim doesn't exist", func() {
		err := RemoveProtectionFinalizer(client, "notExist", namespace)
		Expect(err).NotTo(BeNil())
	})

	It("should error when volumeClaim is not created by fluid", func() {
		err := RemoveProtectionFinalizer(client, "notCreatedByFluid", namespace)
		Expect(err).NotTo(BeNil())
	})

	It("should not error when volumeClaim has no finalizer", func() {
		err := RemoveProtectionFinalizer(client, "hasNoFinalizer", namespace)
		Expect(err).To(BeNil())
	})

	It("should not error when volumeClaim has finalizer", func() {
		err := RemoveProtectionFinalizer(client, "hasFinalizer", namespace)
		Expect(err).To(BeNil())
	})
})
