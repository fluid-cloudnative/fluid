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
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test Daemonset", func() {
	var (
		name      string
		namespace string
		ds        *appsv1.DaemonSet
		objs      []runtime.Object
	)

	BeforeEach(func() {
		name = "test"
		namespace = "default"
		ds = &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: appsv1.DaemonSetSpec{},
			Status: appsv1.DaemonSetStatus{
				NumberUnavailable: 1,
				NumberReady:       1,
			},
		}

		objs = []runtime.Object{}
		objs = append(objs, ds.DeepCopy())
	})

	Context("Test GetDaemonset", func() {
		It("Should get the daemonset successfully", func() {
			fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)
			_, err := GetDaemonset(fakeClient, name, namespace)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should fail to get the daemonset because of not found", func() {
			fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)
			_, err := GetDaemonset(fakeClient, "notFound", namespace)
			Expect(err).To(HaveOccurred())
			Expect(apierrs.IsNotFound(err)).To(BeTrue())
		})
	})

	Context("Test UpdateDaemonSetUpdateStrategy", func() {
		var (
			fakeClient client.Client
		)

		BeforeEach(func() {
			fakeClient = fake.NewFakeClientWithScheme(testScheme, objs...)
		})

		It("Should update to OnDelete strategy successfully", func() {
			// Update to OnDelete strategy
			newStrategy := appsv1.DaemonSetUpdateStrategy{
				Type: appsv1.OnDeleteDaemonSetStrategyType,
			}

			err := UpdateDaemonSetUpdateStrategy(fakeClient, name, namespace, newStrategy)
			Expect(err).NotTo(HaveOccurred())

			// Verify the update
			updatedDS, err := GetDaemonset(fakeClient, name, namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedDS.Spec.UpdateStrategy.Type).To(Equal(appsv1.OnDeleteDaemonSetStrategyType))
		})

		It("Should update RollingUpdate parameters successfully", func() {
			// Update RollingUpdate parameters
			maxSurge := &intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "20%",
			}
			newRollingUpdateStrategy := appsv1.DaemonSetUpdateStrategy{
				Type: appsv1.RollingUpdateDaemonSetStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDaemonSet{
					MaxSurge: maxSurge,
				},
			}

			err := UpdateDaemonSetUpdateStrategy(fakeClient, name, namespace, newRollingUpdateStrategy)
			Expect(err).NotTo(HaveOccurred())

			// Verify the update
			updatedDS, err := GetDaemonset(fakeClient, name, namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedDS.Spec.UpdateStrategy.Type).To(Equal(appsv1.RollingUpdateDaemonSetStrategyType))
			Expect(updatedDS.Spec.UpdateStrategy.RollingUpdate).NotTo(BeNil())
			Expect(updatedDS.Spec.UpdateStrategy.RollingUpdate.MaxSurge).NotTo(BeNil())
			Expect(updatedDS.Spec.UpdateStrategy.RollingUpdate.MaxSurge.StrVal).To(Equal("20%"))
		})

		It("Should fail to update non-existent DaemonSet", func() {
			newStrategy := appsv1.DaemonSetUpdateStrategy{
				Type: appsv1.OnDeleteDaemonSetStrategyType,
			}

			err := UpdateDaemonSetUpdateStrategy(fakeClient, "non-existent", namespace, newStrategy)
			Expect(err).To(HaveOccurred())
		})

		It("Should update with same strategy (no actual change)", func() {
			// First get the current DS
			currentDS, err := GetDaemonset(fakeClient, name, namespace)
			Expect(err).NotTo(HaveOccurred())

			// Update with same strategy
			err = UpdateDaemonSetUpdateStrategy(fakeClient, name, namespace, currentDS.Spec.UpdateStrategy)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
