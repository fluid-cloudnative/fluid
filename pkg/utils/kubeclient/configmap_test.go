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
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ConfigMap Operations", func() {
	var (
		namespace  string
		testClient client.Client
		configMaps []runtime.Object
	)

	BeforeEach(func() {
		namespace = "default"

		configMaps = []runtime.Object{
			&v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: namespace,
				},
				Data: map[string]string{
					"key1": "value1",
				},
			},
			&v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test2",
					Namespace: namespace,
				},
				Data: map[string]string{
					"key2": "value2",
				},
			},
		}

		testClient = fake.NewFakeClientWithScheme(testScheme, configMaps...)
	})

	Describe("IsConfigMapExist", func() {
		Context("when ConfigMap exists", func() {
			It("should return true and no error", func() {
				found, err := IsConfigMapExist(testClient, "test1", namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
			})
		})

		Context("when ConfigMap does not exist", func() {
			It("should return false and no error", func() {
				found, err := IsConfigMapExist(testClient, "notExist", namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Context("when namespace is different", func() {
			It("should return false for ConfigMap in different namespace", func() {
				found, err := IsConfigMapExist(testClient, "test1", "other-namespace")
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})
	})

	Describe("GetConfigmapByName", func() {
		Context("when ConfigMap exists", func() {
			It("should return the ConfigMap successfully", func() {
				cm, err := GetConfigmapByName(testClient, "test1", namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(cm).NotTo(BeNil())
				Expect(cm.Name).To(Equal("test1"))
				Expect(cm.Namespace).To(Equal(namespace))
				Expect(cm.Data).To(HaveKey("key1"))
				Expect(cm.Data["key1"]).To(Equal("value1"))
			})
		})

		Context("when ConfigMap does not exist", func() {
			It("should return nil ConfigMap and no error", func() {
				cm, err := GetConfigmapByName(testClient, "notExist", namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(cm).To(BeNil())
			})
		})

		Context("when retrieving multiple ConfigMaps", func() {
			It("should return correct ConfigMap for each name", func() {
				cm1, err1 := GetConfigmapByName(testClient, "test1", namespace)
				cm2, err2 := GetConfigmapByName(testClient, "test2", namespace)

				Expect(err1).NotTo(HaveOccurred())
				Expect(err2).NotTo(HaveOccurred())
				Expect(cm1.Name).To(Equal("test1"))
				Expect(cm2.Name).To(Equal("test2"))
				Expect(cm1.Data["key1"]).To(Equal("value1"))
				Expect(cm2.Data["key2"]).To(Equal("value2"))
			})
		})
	})

	Describe("DeleteConfigMap", func() {
		Context("when ConfigMap exists", func() {
			It("should delete the ConfigMap successfully", func() {
				err := DeleteConfigMap(testClient, "test1", namespace)
				Expect(err).NotTo(HaveOccurred())

				// Verify deletion
				found, _ := IsConfigMapExist(testClient, "test1", namespace)
				Expect(found).To(BeFalse())
			})
		})

		Context("when ConfigMap does not exist", func() {
			It("should not return an error", func() {
				err := DeleteConfigMap(testClient, "notExist", namespace)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when deleting multiple ConfigMaps", func() {
			It("should delete both ConfigMaps successfully", func() {
				err1 := DeleteConfigMap(testClient, "test1", namespace)
				err2 := DeleteConfigMap(testClient, "test2", namespace)

				Expect(err1).NotTo(HaveOccurred())
				Expect(err2).NotTo(HaveOccurred())

				found1, _ := IsConfigMapExist(testClient, "test1", namespace)
				found2, _ := IsConfigMapExist(testClient, "test2", namespace)
				Expect(found1).To(BeFalse())
				Expect(found2).To(BeFalse())
			})
		})
	})

	Describe("CopyConfigMap", func() {
		var (
			srcNamespace string
			dstNamespace string
			ownerRef     metav1.OwnerReference
		)

		BeforeEach(func() {
			srcNamespace = "src"
			dstNamespace = "dst"

			ownerRef = metav1.OwnerReference{
				APIVersion: "v1",
				Kind:       "Dataset",
				Name:       "test-dataset",
				UID:        "test-uid-123",
			}

			// Create source ConfigMap
			srcConfigMap := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "src-config",
					Namespace: srcNamespace,
					Labels: map[string]string{
						"app": "test",
					},
					Annotations: map[string]string{
						"description": "test config",
					},
				},
				Data: map[string]string{
					"check.sh": "/bin/sh check",
					"config":   "test config data",
				},
			}

			testClient = fake.NewFakeClient(srcConfigMap)
		})

		Context("when copying ConfigMap successfully", func() {
			It("should create a copy in destination namespace", func() {
				src := types.NamespacedName{
					Name:      "src-config",
					Namespace: srcNamespace,
				}
				dst := types.NamespacedName{
					Name:      "dst-config",
					Namespace: dstNamespace,
				}

				err := CopyConfigMap(testClient, src, dst, ownerRef)
				Expect(err).NotTo(HaveOccurred())

				// Verify copied ConfigMap exists
				copiedCM, err := GetConfigmapByName(testClient, dst.Name, dst.Namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(copiedCM).NotTo(BeNil())
				Expect(copiedCM.Name).To(Equal("dst-config"))
				Expect(copiedCM.Namespace).To(Equal(dstNamespace))
				Expect(copiedCM.Data).To(HaveKey("check.sh"))
				Expect(copiedCM.Data["check.sh"]).To(Equal("/bin/sh check"))
				Expect(copiedCM.Labels).To(HaveKey("app"))
				Expect(copiedCM.Labels["app"]).To(Equal("test"))
				Expect(copiedCM.Annotations).To(HaveKey("description"))
				Expect(copiedCM.OwnerReferences).To(HaveLen(1))
				Expect(copiedCM.OwnerReferences[0].Name).To(Equal("test-dataset"))
			})

			It("should add dataset ID label to copied ConfigMap", func() {
				src := types.NamespacedName{
					Name:      "src-config",
					Namespace: srcNamespace,
				}
				dst := types.NamespacedName{
					Name:      "dst-config",
					Namespace: dstNamespace,
				}

				err := CopyConfigMap(testClient, src, dst, ownerRef)
				Expect(err).NotTo(HaveOccurred())

				copiedCM, _ := GetConfigmapByName(testClient, dst.Name, dst.Namespace)
				Expect(copiedCM.Labels).To(HaveKey(common.LabelAnnotationDatasetId))
			})
		})

		Context("when destination ConfigMap already exists", func() {
			It("should not return an error and skip copying", func() {
				// Create destination ConfigMap first
				dstConfigMap := &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dst-config",
						Namespace: dstNamespace,
					},
					Data: map[string]string{
						"existing": "data",
					},
				}
				testClient = fake.NewFakeClient(
					&v1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "src-config",
							Namespace: srcNamespace,
						},
						Data: map[string]string{
							"check.sh": "/bin/sh check",
						},
					},
					dstConfigMap,
				)

				src := types.NamespacedName{
					Name:      "src-config",
					Namespace: srcNamespace,
				}
				dst := types.NamespacedName{
					Name:      "dst-config",
					Namespace: dstNamespace,
				}

				err := CopyConfigMap(testClient, src, dst, ownerRef)
				Expect(err).NotTo(HaveOccurred())

				// Verify original data is preserved
				copiedCM, _ := GetConfigmapByName(testClient, dst.Name, dst.Namespace)
				Expect(copiedCM.Data).To(HaveKey("existing"))
			})
		})

		Context("when source ConfigMap does not exist", func() {
			It("should return an error", func() {
				src := types.NamespacedName{
					Name:      "non-existent",
					Namespace: srcNamespace,
				}
				dst := types.NamespacedName{
					Name:      "dst-config",
					Namespace: dstNamespace,
				}

				err := CopyConfigMap(testClient, src, dst, ownerRef)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("do not exist"))
			})
		})
	})

	Describe("CreateConfigMap", func() {
		var (
			cmName         string
			cmNamespace    string
			dataKey        string
			dataValue      []byte
			ownerDatasetId string
		)

		BeforeEach(func() {
			cmName = "new-configmap"
			cmNamespace = "default"
			dataKey = "config.yaml"
			dataValue = []byte("key: value\ntest: data")
			ownerDatasetId = "dataset-123"
			testClient = fake.NewFakeClientWithScheme(testScheme)
		})

		Context("when creating a new ConfigMap", func() {
			It("should create ConfigMap successfully", func() {
				err := CreateConfigMap(testClient, cmName, cmNamespace, dataKey, dataValue, ownerDatasetId)
				Expect(err).NotTo(HaveOccurred())

				// Verify ConfigMap was created
				cm, err := GetConfigmapByName(testClient, cmName, cmNamespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(cm).NotTo(BeNil())
				Expect(cm.Name).To(Equal(cmName))
				Expect(cm.Namespace).To(Equal(cmNamespace))
				Expect(cm.Data).To(HaveKey(dataKey))
				Expect(cm.Data[dataKey]).To(Equal(string(dataValue)))
			})

			It("should set dataset ID label correctly", func() {
				err := CreateConfigMap(testClient, cmName, cmNamespace, dataKey, dataValue, ownerDatasetId)
				Expect(err).NotTo(HaveOccurred())

				cm, _ := GetConfigmapByName(testClient, cmName, cmNamespace)
				Expect(cm.Labels).To(HaveKey(common.LabelAnnotationDatasetId))
				Expect(cm.Labels[common.LabelAnnotationDatasetId]).To(Equal(ownerDatasetId))
			})

			It("should handle empty owner dataset ID", func() {
				err := CreateConfigMap(testClient, cmName, cmNamespace, dataKey, dataValue, "")
				Expect(err).NotTo(HaveOccurred())

				cm, _ := GetConfigmapByName(testClient, cmName, cmNamespace)
				Expect(cm.Labels[common.LabelAnnotationDatasetId]).To(Equal(""))
			})

			It("should handle large text data", func() {
				largeData := []byte("This is a large configuration file with multiple lines\nLine 2\nLine 3\nLine 4")
				err := CreateConfigMap(testClient, "large-cm", cmNamespace, "config.txt", largeData, ownerDatasetId)
				Expect(err).NotTo(HaveOccurred())

				cm, _ := GetConfigmapByName(testClient, "large-cm", cmNamespace)
				Expect(cm.Data["config.txt"]).To(Equal(string(largeData)))
			})
		})

		Context("when creating multiple ConfigMaps", func() {
			It("should create all ConfigMaps successfully", func() {
				err1 := CreateConfigMap(testClient, "cm1", cmNamespace, "key1", []byte("data1"), "id1")
				err2 := CreateConfigMap(testClient, "cm2", cmNamespace, "key2", []byte("data2"), "id2")

				Expect(err1).NotTo(HaveOccurred())
				Expect(err2).NotTo(HaveOccurred())

				cm1, _ := GetConfigmapByName(testClient, "cm1", cmNamespace)
				cm2, _ := GetConfigmapByName(testClient, "cm2", cmNamespace)

				Expect(cm1.Data["key1"]).To(Equal("data1"))
				Expect(cm2.Data["key2"]).To(Equal("data2"))
			})
		})
	})

	Describe("UpdateConfigMap", func() {
		var existingCM *v1.ConfigMap

		BeforeEach(func() {
			existingCM = &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "update-test",
					Namespace: namespace,
					Labels: map[string]string{
						"version": "v1",
					},
				},
				Data: map[string]string{
					"config": "original data",
				},
			}
			testClient = fake.NewFakeClient(existingCM)
		})

		Context("when updating an existing ConfigMap", func() {
			It("should update ConfigMap data successfully", func() {
				cm, _ := GetConfigmapByName(testClient, "update-test", namespace)
				cm.Data["config"] = "updated data"
				cm.Data["newkey"] = "newvalue"

				err := UpdateConfigMap(testClient, cm)
				Expect(err).NotTo(HaveOccurred())

				// Verify update
				updatedCM, _ := GetConfigmapByName(testClient, "update-test", namespace)
				Expect(updatedCM.Data["config"]).To(Equal("updated data"))
				Expect(updatedCM.Data["newkey"]).To(Equal("newvalue"))
			})

			It("should update ConfigMap labels successfully", func() {
				cm, _ := GetConfigmapByName(testClient, "update-test", namespace)
				cm.Labels["version"] = "v2"
				cm.Labels["environment"] = "production"

				err := UpdateConfigMap(testClient, cm)
				Expect(err).NotTo(HaveOccurred())

				updatedCM, _ := GetConfigmapByName(testClient, "update-test", namespace)
				Expect(updatedCM.Labels["version"]).To(Equal("v2"))
				Expect(updatedCM.Labels["environment"]).To(Equal("production"))
			})

			It("should update ConfigMap annotations successfully", func() {
				cm, _ := GetConfigmapByName(testClient, "update-test", namespace)
				cm.Annotations = map[string]string{
					"updated":   "true",
					"timestamp": "2024-01-01",
				}

				err := UpdateConfigMap(testClient, cm)
				Expect(err).NotTo(HaveOccurred())

				updatedCM, _ := GetConfigmapByName(testClient, "update-test", namespace)
				Expect(updatedCM.Annotations).To(HaveKey("updated"))
				Expect(updatedCM.Annotations["updated"]).To(Equal("true"))
			})
		})
	})

	Describe("Integration Tests", func() {
		Context("complete ConfigMap lifecycle", func() {
			It("should create, read, update, and delete ConfigMap", func() {
				cmName := "lifecycle-test"
				testClient = fake.NewFakeClientWithScheme(testScheme)

				// Create
				err := CreateConfigMap(testClient, cmName, namespace, "key", []byte("value"), "dataset-id")
				Expect(err).NotTo(HaveOccurred())

				// Read
				cm, err := GetConfigmapByName(testClient, cmName, namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(cm).NotTo(BeNil())

				// Update
				cm.Data["key"] = "updated-value"
				err = UpdateConfigMap(testClient, cm)
				Expect(err).NotTo(HaveOccurred())

				updatedCM, _ := GetConfigmapByName(testClient, cmName, namespace)
				Expect(updatedCM.Data["key"]).To(Equal("updated-value"))

				// Delete
				err = DeleteConfigMap(testClient, cmName, namespace)
				Expect(err).NotTo(HaveOccurred())

				found, _ := IsConfigMapExist(testClient, cmName, namespace)
				Expect(found).To(BeFalse())
			})
		})

		Context("working with multiple namespaces", func() {
			It("should handle ConfigMaps in different namespaces independently", func() {
				testClient = fake.NewFakeClientWithScheme(testScheme)

				// Create in different namespaces
				err1 := CreateConfigMap(testClient, "cm", "ns1", "key", []byte("data1"), "id1")
				err2 := CreateConfigMap(testClient, "cm", "ns2", "key", []byte("data2"), "id2")

				Expect(err1).NotTo(HaveOccurred())
				Expect(err2).NotTo(HaveOccurred())

				// Verify isolation
				cm1, _ := GetConfigmapByName(testClient, "cm", "ns1")
				cm2, _ := GetConfigmapByName(testClient, "cm", "ns2")

				Expect(cm1.Data["key"]).To(Equal("data1"))
				Expect(cm2.Data["key"]).To(Equal("data2"))
				Expect(cm1.Labels[common.LabelAnnotationDatasetId]).To(Equal("id1"))
				Expect(cm2.Labels[common.LabelAnnotationDatasetId]).To(Equal("id2"))
			})
		})
	})
})
