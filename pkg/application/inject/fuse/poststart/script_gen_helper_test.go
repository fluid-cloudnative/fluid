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

package poststart

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

var _ = Describe("ScriptGeneratorHelper", func() {
	Describe("BuildConfigMap", func() {
		Context("basic configmap creation", func() {
			It("should create configmap with correct properties", func() {
				helper := &scriptGeneratorHelper{
					configMapName:   "test-config",
					scriptContent:   "#!/bin/bash\necho 'test'",
					scriptFileName:  "init.sh",
					scriptMountPath: "/scripts",
				}
				dataset := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-dataset",
						Namespace: "default",
						UID:       "test-uid-123",
					},
				}
				configMapKey := types.NamespacedName{
					Name:      "my-configmap",
					Namespace: "default",
				}

				cm := helper.BuildConfigMap(dataset, configMapKey)

				Expect(cm.Name).To(Equal("my-configmap"))
				Expect(cm.Namespace).To(Equal("default"))
				Expect(cm.Data).To(HaveKeyWithValue("init.sh", "#!/bin/bash\necho 'test'"))
				Expect(cm.Labels).To(HaveKeyWithValue(common.LabelAnnotationDatasetId, "default-test-dataset"))
			})
		})

		Context("configmap with different namespace", func() {
			It("should create configmap with correct namespace and properties", func() {
				helper := &scriptGeneratorHelper{
					configMapName:   "poststart-script",
					scriptContent:   "echo 'hello world'",
					scriptFileName:  "startup.sh",
					scriptMountPath: "/opt/scripts",
				}
				dataset := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-dataset",
						Namespace: "prod-ns",
						UID:       "uid-456",
					},
				}
				configMapKey := types.NamespacedName{
					Name:      "prod-configmap",
					Namespace: "prod-ns",
				}

				cm := helper.BuildConfigMap(dataset, configMapKey)

				Expect(cm.Name).To(Equal("prod-configmap"))
				Expect(cm.Namespace).To(Equal("prod-ns"))
				Expect(cm.Data).To(HaveKeyWithValue("startup.sh", "echo 'hello world'"))
				Expect(cm.Labels).To(HaveKeyWithValue(common.LabelAnnotationDatasetId, "prod-ns-my-dataset"))
			})
		})

		Context("configmap with empty script content", func() {
			It("should create configmap with empty script", func() {
				helper := &scriptGeneratorHelper{
					configMapName:   "empty-script",
					scriptContent:   "",
					scriptFileName:  "empty.sh",
					scriptMountPath: "/scripts",
				}
				dataset := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "empty-dataset",
						Namespace: "test",
						UID:       "empty-uid",
					},
				}
				configMapKey := types.NamespacedName{
					Name:      "empty-cm",
					Namespace: "test",
				}

				cm := helper.BuildConfigMap(dataset, configMapKey)

				Expect(cm.Name).To(Equal("empty-cm"))
				Expect(cm.Namespace).To(Equal("test"))
				Expect(cm.Data).To(HaveKeyWithValue("empty.sh", ""))
				Expect(cm.Labels).To(HaveKeyWithValue(common.LabelAnnotationDatasetId, "test-empty-dataset"))
			})
		})
	})

	Describe("GetNamespacedConfigMapKey", func() {
		Context("alluxio runtime type", func() {
			It("should return correct configmap key", func() {
				helper := &scriptGeneratorHelper{
					configMapName: "poststart-script",
				}
				datasetKey := types.NamespacedName{
					Name:      "test-dataset",
					Namespace: "default",
				}

				key := helper.GetNamespacedConfigMapKey(datasetKey, "Alluxio")

				Expect(key.Name).To(Equal("alluxio-poststart-script"))
				Expect(key.Namespace).To(Equal("default"))
			})
		})

		Context("juicefs runtime type", func() {
			It("should return correct configmap key", func() {
				helper := &scriptGeneratorHelper{
					configMapName: "init-config",
				}
				datasetKey := types.NamespacedName{
					Name:      "my-dataset",
					Namespace: "prod",
				}

				key := helper.GetNamespacedConfigMapKey(datasetKey, "JuiceFS")

				Expect(key.Name).To(Equal("juicefs-init-config"))
				Expect(key.Namespace).To(Equal("prod"))
			})
		})

		Context("lowercase runtime type", func() {
			It("should return correct configmap key", func() {
				helper := &scriptGeneratorHelper{
					configMapName: "script",
				}
				datasetKey := types.NamespacedName{
					Name:      "dataset",
					Namespace: "ns",
				}

				key := helper.GetNamespacedConfigMapKey(datasetKey, "jindo")

				Expect(key.Name).To(Equal("jindo-script"))
				Expect(key.Namespace).To(Equal("ns"))
			})
		})

		Context("mixed case runtime type", func() {
			It("should return correct configmap key", func() {
				helper := &scriptGeneratorHelper{
					configMapName: "my-config",
				}
				datasetKey := types.NamespacedName{
					Name:      "test",
					Namespace: "test-ns",
				}

				key := helper.GetNamespacedConfigMapKey(datasetKey, "GooseFS")

				Expect(key.Name).To(Equal("goosefs-my-config"))
				Expect(key.Namespace).To(Equal("test-ns"))
			})
		})
	})

	Describe("GetVolume", func() {
		Context("basic volume creation", func() {
			It("should create volume with correct properties", func() {
				helper := &scriptGeneratorHelper{
					configMapName: "test-config",
				}
				configMapKey := types.NamespacedName{
					Name:      "my-configmap",
					Namespace: "default",
				}

				volume := helper.GetVolume(configMapKey)

				Expect(volume.Name).To(Equal("test-config"))
				Expect(volume.VolumeSource.ConfigMap).NotTo(BeNil())
				Expect(volume.VolumeSource.ConfigMap.Name).To(Equal("my-configmap"))
				Expect(volume.VolumeSource.ConfigMap.DefaultMode).NotTo(BeNil())
				Expect(*volume.VolumeSource.ConfigMap.DefaultMode).To(Equal(int32(0755)))
			})
		})

		Context("different configmap name", func() {
			It("should create volume with correct configmap reference", func() {
				helper := &scriptGeneratorHelper{
					configMapName: "poststart-vol",
				}
				configMapKey := types.NamespacedName{
					Name:      "prod-cm",
					Namespace: "production",
				}

				volume := helper.GetVolume(configMapKey)

				Expect(volume.Name).To(Equal("poststart-vol"))
				Expect(volume.VolumeSource.ConfigMap).NotTo(BeNil())
				Expect(volume.VolumeSource.ConfigMap.Name).To(Equal("prod-cm"))
				Expect(volume.VolumeSource.ConfigMap.DefaultMode).NotTo(BeNil())
				Expect(*volume.VolumeSource.ConfigMap.DefaultMode).To(Equal(int32(0755)))
			})
		})
	})

	Describe("GetVolumeMount", func() {
		Context("basic volume mount", func() {
			It("should create volume mount with correct properties", func() {
				helper := &scriptGeneratorHelper{
					configMapName:   "test-config",
					scriptFileName:  "init.sh",
					scriptMountPath: "/scripts/init.sh",
				}

				vm := helper.GetVolumeMount()

				Expect(vm.Name).To(Equal("test-config"))
				Expect(vm.MountPath).To(Equal("/scripts/init.sh"))
				Expect(vm.SubPath).To(Equal("init.sh"))
				Expect(vm.ReadOnly).To(BeTrue())
			})
		})

		Context("different mount path", func() {
			It("should create volume mount with correct mount path", func() {
				helper := &scriptGeneratorHelper{
					configMapName:   "poststart-config",
					scriptFileName:  "startup.sh",
					scriptMountPath: "/opt/scripts/startup.sh",
				}

				vm := helper.GetVolumeMount()

				Expect(vm.Name).To(Equal("poststart-config"))
				Expect(vm.MountPath).To(Equal("/opt/scripts/startup.sh"))
				Expect(vm.SubPath).To(Equal("startup.sh"))
				Expect(vm.ReadOnly).To(BeTrue())
			})
		})

		Context("root path mount", func() {
			It("should create volume mount at root path", func() {
				helper := &scriptGeneratorHelper{
					configMapName:   "root-config",
					scriptFileName:  "run.sh",
					scriptMountPath: "/run.sh",
				}

				vm := helper.GetVolumeMount()

				Expect(vm.Name).To(Equal("root-config"))
				Expect(vm.MountPath).To(Equal("/run.sh"))
				Expect(vm.SubPath).To(Equal("run.sh"))
				Expect(vm.ReadOnly).To(BeTrue())
			})
		})
	})
})
