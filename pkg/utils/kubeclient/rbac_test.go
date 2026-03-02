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
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("DeleteServiceAccount", func() {
	var (
		namespace           string
		testSAInput         []*corev1.ServiceAccount
		testServiceAccounts []runtime.Object
		fakeClient          client.Client
	)

	BeforeEach(func() {
		namespace = "default"
		testSAInput = []*corev1.ServiceAccount{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "test1", Namespace: namespace},
			},
		}

		testServiceAccounts = []runtime.Object{}
		for _, ns := range testSAInput {
			testServiceAccounts = append(testServiceAccounts, ns.DeepCopy())
		}

		fakeClient = fake.NewFakeClientWithScheme(testScheme, testServiceAccounts...)
	})

	Context("when service account doesn't exist", func() {
		It("should not return an error", func() {
			err := DeleteServiceAccount(fakeClient, "notExist", namespace)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when service account exists", func() {
		It("should delete successfully without error", func() {
			err := DeleteServiceAccount(fakeClient, "test1", namespace)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("DeleteRole", func() {
	var (
		namespace     string
		testRoleInput []*rbacv1.Role
		testRole      []runtime.Object
		fakeClient    client.Client
	)

	BeforeEach(func() {
		namespace = "default"
		testRoleInput = []*rbacv1.Role{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: namespace},
			},
		}

		testRole = []runtime.Object{}
		for _, ns := range testRoleInput {
			testRole = append(testRole, ns.DeepCopy())
		}

		fakeClient = fake.NewFakeClientWithScheme(testScheme, testRole...)
	})

	Context("when role doesn't exist", func() {
		It("should not return an error", func() {
			err := DeleteRole(fakeClient, "notExist", namespace)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when role exists", func() {
		It("should delete successfully without error", func() {
			err := DeleteRole(fakeClient, "test", namespace)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("DeleteRoleBinding", func() {
	var (
		namespace            string
		testRoleBindingInput []*rbacv1.RoleBinding
		testRoleBinding      []runtime.Object
		fakeClient           client.Client
	)

	BeforeEach(func() {
		namespace = "default"
		testRoleBindingInput = []*rbacv1.RoleBinding{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: namespace},
			},
		}

		testRoleBinding = []runtime.Object{}
		for _, ns := range testRoleBindingInput {
			testRoleBinding = append(testRoleBinding, ns.DeepCopy())
		}

		fakeClient = fake.NewFakeClientWithScheme(testScheme, testRoleBinding...)
	})

	Context("when role binding doesn't exist", func() {
		It("should not return an error", func() {
			err := DeleteRoleBinding(fakeClient, "notExist", namespace)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when role binding exists", func() {
		It("should delete successfully without error", func() {
			err := DeleteRoleBinding(fakeClient, "test", namespace)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
