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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Secret related unit tests", Label("pkg.utils.kubeclient.secret_test.go"), func() {
	var (
		fakeClient client.Client
		testScheme *runtime.Scheme
	)

	BeforeEach(func() {
		testScheme = runtime.NewScheme()
		_ = v1.AddToScheme(testScheme)
	})

	Context("Test GetSecret()", func() {
		var (
			secretName      = "mysecret"
			secretNamespace = "default"
			mockSecret      *v1.Secret
		)

		BeforeEach(func() {
			mockSecret = &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: secretNamespace,
				},
			}
			fakeClient = fake.NewFakeClientWithScheme(testScheme, mockSecret)
		})

		It("should get existing secret successfully", func() {
			secret, err := GetSecret(fakeClient, secretName, secretNamespace)

			Expect(err).NotTo(HaveOccurred())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Name).To(Equal(secretName))
			Expect(secret.Namespace).To(Equal(secretNamespace))
		})

		It("should return error for non-existing secret", func() {
			secret, err := GetSecret(fakeClient, secretName+"not-exist", secretNamespace)

			Expect(err).To(HaveOccurred())
			Expect(secret).To(BeNil())
		})
	})

	Context("Test CreateSecret()", func() {
		var (
			existingSecret *v1.Secret
		)

		BeforeEach(func() {
			existingSecret = &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret2",
					Namespace: "namespace",
				},
			}
			fakeClient = fake.NewFakeClientWithScheme(testScheme, existingSecret)
		})

		It("should create new secret successfully", func() {
			newSecret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret1",
					Namespace: "namespace",
				},
			}

			err := CreateSecret(fakeClient, newSecret)
			Expect(err).NotTo(HaveOccurred())

			// Verify the secret was created
			createdSecret, err := GetSecret(fakeClient, newSecret.Name, newSecret.Namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(createdSecret).NotTo(BeNil())
			Expect(createdSecret.Name).To(Equal(newSecret.Name))
			Expect(createdSecret.Namespace).To(Equal(newSecret.Namespace))
		})

		It("should create new secret in different namespace", func() {
			newSecret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret1",
					Namespace: "namespace2",
				},
			}

			err := CreateSecret(fakeClient, newSecret)
			Expect(err).NotTo(HaveOccurred())

			// Verify the secret was created
			createdSecret, err := GetSecret(fakeClient, newSecret.Name, newSecret.Namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(createdSecret).NotTo(BeNil())
			Expect(createdSecret.Name).To(Equal(newSecret.Name))
			Expect(createdSecret.Namespace).To(Equal(newSecret.Namespace))
		})

		It("should return error when creating existing secret", func() {
			err := CreateSecret(fakeClient, existingSecret)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Test UpdateSecret()", func() {
		var (
			existingSecret *v1.Secret
		)

		BeforeEach(func() {
			existingSecret = &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret2",
					Namespace: "namespace",
					Labels: map[string]string{
						"key": "old",
					},
				},
			}
			fakeClient = fake.NewFakeClientWithScheme(testScheme, existingSecret)
		})

		It("should return error when updating non-existing secret", func() {
			nonExistingSecret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret1",
					Namespace: "namespace",
					Labels: map[string]string{
						"key": "new",
					},
				},
			}

			err := UpdateSecret(fakeClient, nonExistingSecret)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when updating non-existing secret in different namespace", func() {
			nonExistingSecret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret1",
					Namespace: "namespace2",
					Labels: map[string]string{
						"key": "new",
					},
				},
			}

			err := UpdateSecret(fakeClient, nonExistingSecret)
			Expect(err).To(HaveOccurred())
		})

		It("should update existing secret successfully", func() {
			updatedSecret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret2",
					Namespace: "namespace",
					Labels: map[string]string{
						"key": "new",
					},
				},
			}

			err := UpdateSecret(fakeClient, updatedSecret)
			Expect(err).NotTo(HaveOccurred())

			// Verify the secret was updated
			gotSecret, err := GetSecret(fakeClient, updatedSecret.Name, updatedSecret.Namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(gotSecret).NotTo(BeNil())
			Expect(gotSecret.Labels["key"]).To(Equal("new"))
		})
	})

	Context("Test CopySecretToNamespace()", func() {
		var (
			sourceSecret *v1.Secret
			from         types.NamespacedName
			to           types.NamespacedName
		)

		BeforeEach(func() {
			from = types.NamespacedName{
				Name:      "source-secret",
				Namespace: "source-namespace",
			}
			to = types.NamespacedName{
				Name:      "target-secret",
				Namespace: "target-namespace",
			}

			sourceSecret = &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      from.Name,
					Namespace: from.Namespace,
				},
				Data: map[string][]byte{
					"username": []byte("admin"),
					"password": []byte("secret123"),
				},
				StringData: map[string]string{
					"host": "example.com",
				},
			}
			fakeClient = fake.NewFakeClientWithScheme(testScheme, sourceSecret)
		})

		It("should copy secret to target namespace successfully", func() {
			err := CopySecretToNamespace(fakeClient, from, to, nil)
			Expect(err).NotTo(HaveOccurred())

			// Verify the secret was copied
			copiedSecret, err := GetSecret(fakeClient, to.Name, to.Namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(copiedSecret).NotTo(BeNil())
			Expect(copiedSecret.Name).To(Equal(to.Name))
			Expect(copiedSecret.Namespace).To(Equal(to.Namespace))
			Expect(copiedSecret.Data).To(Equal(sourceSecret.Data))
			Expect(copiedSecret.StringData).To(Equal(sourceSecret.StringData))
			Expect(copiedSecret.Labels[common.LabelAnnotationCopyFrom]).To(Equal("source-namespace_source-secret"))
		})

		It("should copy secret with owner reference", func() {
			ownerRef := &common.OwnerReference{
				APIVersion:         "v1",
				Kind:               "ConfigMap",
				Name:               "test-configmap",
				UID:                "test-uid",
				Controller:         true,
				BlockOwnerDeletion: true,
			}

			err := CopySecretToNamespace(fakeClient, from, to, ownerRef)
			Expect(err).NotTo(HaveOccurred())

			// Verify the secret was copied with owner reference
			copiedSecret, err := GetSecret(fakeClient, to.Name, to.Namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(copiedSecret).NotTo(BeNil())
			Expect(copiedSecret.OwnerReferences).To(HaveLen(1))
			Expect(copiedSecret.OwnerReferences[0].APIVersion).To(Equal(ownerRef.APIVersion))
			Expect(copiedSecret.OwnerReferences[0].Kind).To(Equal(ownerRef.Kind))
			Expect(copiedSecret.OwnerReferences[0].Name).To(Equal(ownerRef.Name))
			Expect(copiedSecret.OwnerReferences[0].UID).To(Equal(types.UID(ownerRef.UID)))
			Expect(*copiedSecret.OwnerReferences[0].Controller).To(Equal(ownerRef.Controller))
			Expect(*copiedSecret.OwnerReferences[0].BlockOwnerDeletion).To(Equal(ownerRef.BlockOwnerDeletion))
		})

		It("should return nil if target secret already exists", func() {
			// Create target secret first
			targetSecret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      to.Name,
					Namespace: to.Namespace,
				},
			}
			fakeClient = fake.NewFakeClientWithScheme(testScheme, sourceSecret, targetSecret)

			err := CopySecretToNamespace(fakeClient, from, to, nil)
			Expect(err).NotTo(HaveOccurred())

			// Verify the original target secret is unchanged
			existingSecret, err := GetSecret(fakeClient, to.Name, to.Namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(existingSecret).NotTo(BeNil())
			Expect(existingSecret.Data).To(BeEmpty())
		})

		It("should return error if source secret does not exist", func() {
			nonExistentFrom := types.NamespacedName{
				Name:      "non-existent-secret",
				Namespace: "source-namespace",
			}

			err := CopySecretToNamespace(fakeClient, nonExistentFrom, to, nil)
			Expect(err).To(HaveOccurred())
		})
	})
})
