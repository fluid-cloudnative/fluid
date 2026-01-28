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
	appsv1 "k8s.io/api/apps/v1"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("CompareOwnerRefMatcheWithExpected", func() {
	var (
		controller    *appsv1.StatefulSet
		child         runtime.Object
		mockClient    client.Client
		scheme        *runtime.Scheme
		runtimeObjs   []runtime.Object
		metaObj       metav1.Object
		controllerRef *metav1.OwnerReference
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		runtimeObjs = []runtime.Object{}
	})

	JustBeforeEach(func() {
		scheme.AddKnownTypes(appsv1.SchemeGroupVersion, controller)
		err := v1.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())

		runtimeObjs = append(runtimeObjs, controller)
		runtimeObjs = append(runtimeObjs, child)
		mockClient = fake.NewFakeClientWithScheme(scheme, runtimeObjs...)

		var err2 error
		metaObj, err2 = meta.Accessor(child)
		Expect(err2).NotTo(HaveOccurred())

		controllerRef = metav1.GetControllerOf(metaObj)
	})

	Context("when there is no controller reference", func() {
		BeforeEach(func() {
			controller = &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: "big-data",
				},
				Spec: appsv1.StatefulSetSpec{},
			}
			child = &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1-0",
					Namespace: "big-data",
				},
				Spec: v1.PodSpec{},
			}
		})

		It("should return false without error", func() {
			result, err := compareOwnerRefMatcheWithExpected(mockClient, controllerRef, metaObj.GetNamespace(), controller)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeFalse())
		})
	})

	Context("when the controller UID does not match", func() {
		BeforeEach(func() {
			controller = &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test2",
					Namespace: "big-data",
					UID:       "uid",
				},
				Spec: appsv1.StatefulSetSpec{},
			}
			child = &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test2-0",
					Namespace: "big-data",
					OwnerReferences: []metav1.OwnerReference{{
						Kind:       "StatefulSet",
						APIVersion: "app/v1",
						UID:        "uid1",
						Controller: ptr.To(true),
					}},
				},
				Spec: v1.PodSpec{},
			}
		})

		It("should return false without error", func() {
			result, err := compareOwnerRefMatcheWithExpected(mockClient, controllerRef, metaObj.GetNamespace(), controller)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeFalse())
		})
	})

	Context("when the controller matches", func() {
		BeforeEach(func() {
			controller = &appsv1.StatefulSet{
				TypeMeta: metav1.TypeMeta{
					Kind:       "StatefulSet",
					APIVersion: "app/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test2",
					Namespace: "big-data",
					UID:       "uid2",
				},
				Spec: appsv1.StatefulSetSpec{},
			}
			child = &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test2-0",
					Namespace: "big-data",
					OwnerReferences: []metav1.OwnerReference{{
						Kind:       "StatefulSet",
						APIVersion: "app/v1",
						UID:        "uid2",
						Name:       "test2",
						Controller: ptr.To(true),
					}},
				},
				Spec: v1.PodSpec{},
			}
		})

		It("should return true without error", func() {
			result, err := compareOwnerRefMatcheWithExpected(mockClient, controllerRef, metaObj.GetNamespace(), controller)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeTrue())
		})
	})
})
