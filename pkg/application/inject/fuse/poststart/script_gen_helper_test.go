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
	"fmt"

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
				Expect(cm.Annotations).To(HaveKey(common.AnnotationCheckMountScriptSHA256))
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

var _ = Describe("ScriptGeneratorForApp", func() {
	Describe("NewScriptGeneratorForApp", func() {
		It("should create generator with correct namespace", func() {
			g := NewScriptGeneratorForApp("test-ns")
			Expect(g).NotTo(BeNil())
			Expect(g.namespace).To(Equal("test-ns"))
		})
	})

	Describe("BuildConfigmap", func() {
		It("should create configmap with correct name, namespace and script content", func() {
			g := NewScriptGeneratorForApp("default")
			cm := g.BuildConfigmap()

			Expect(cm.Name).To(Equal(appConfigMapName))
			Expect(cm.Namespace).To(Equal("default"))
			Expect(cm.Data).To(HaveKey(appScriptName))
			Expect(cm.Data[appScriptName]).NotTo(BeEmpty())
		})

		It("should set the LabelCheckMountScriptSHA256 label", func() {
			g := NewScriptGeneratorForApp("default")
			cm := g.BuildConfigmap()

			Expect(cm.Annotations).To(HaveKey(common.AnnotationCheckMountScriptSHA256))
			Expect(cm.Annotations[common.AnnotationCheckMountScriptSHA256]).To(Equal(appScriptContentSHA256))
		})

		It("SHA256 label value should be at most 63 characters", func() {
			g := NewScriptGeneratorForApp("default")
			cm := g.BuildConfigmap()

			sha256Annotation := cm.Annotations[common.AnnotationCheckMountScriptSHA256]
			Expect(len(sha256Annotation)).To(BeNumerically("<=", 63))
		})

		It("should produce the same configmap for different namespaces with only namespace differing", func() {
			g1 := NewScriptGeneratorForApp("ns-a")
			g2 := NewScriptGeneratorForApp("ns-b")
			cm1 := g1.BuildConfigmap()
			cm2 := g2.BuildConfigmap()

			Expect(cm1.Name).To(Equal(cm2.Name))
			Expect(cm1.Namespace).To(Equal("ns-a"))
			Expect(cm2.Namespace).To(Equal("ns-b"))
			Expect(cm1.Data).To(Equal(cm2.Data))
			Expect(cm1.Labels).To(Equal(cm2.Labels))
		})
	})

	Describe("GetScriptSHA256", func() {
		It("should return the same SHA256 as stored in appScriptContentSHA256", func() {
			g := NewScriptGeneratorForApp("default")
			Expect(g.GetScriptSHA256()).To(Equal(appScriptContentSHA256))
		})

		It("should return a non-empty SHA256 string with length <= 63", func() {
			g := NewScriptGeneratorForApp("default")
			sha := g.GetScriptSHA256()
			Expect(sha).NotTo(BeEmpty())
			Expect(len(sha)).To(BeNumerically("<=", 63))
		})
	})

	Describe("GetPostStartCommand", func() {
		It("should return correct exec command with given paths and types", func() {
			g := NewScriptGeneratorForApp("default")
			handler := g.GetPostStartCommand("/data1:/data2", "alluxio:jindo")

			Expect(handler).NotTo(BeNil())
			Expect(handler.Exec).NotTo(BeNil())
			expectedCmd := fmt.Sprintf("time %s %s %s", appScriptPath, "/data1:/data2", "alluxio:jindo")
			Expect(handler.Exec.Command).To(Equal([]string{"bash", "-c", expectedCmd}))
		})

		It("should include the script path in the command", func() {
			g := NewScriptGeneratorForApp("default")
			handler := g.GetPostStartCommand("/mnt/data", "juicefs")

			Expect(handler.Exec.Command[2]).To(ContainSubstring(appScriptPath))
		})
	})

	Describe("GetVolume", func() {
		It("should return volume with correct name and configmap reference", func() {
			g := NewScriptGeneratorForApp("default")
			vol := g.GetVolume()

			Expect(vol.Name).To(Equal(appVolName))
			Expect(vol.VolumeSource.ConfigMap).NotTo(BeNil())
			Expect(vol.VolumeSource.ConfigMap.Name).To(Equal(appConfigMapName))
		})

		It("should set default mode to 0755", func() {
			g := NewScriptGeneratorForApp("default")
			vol := g.GetVolume()

			Expect(vol.VolumeSource.ConfigMap.DefaultMode).NotTo(BeNil())
			Expect(*vol.VolumeSource.ConfigMap.DefaultMode).To(Equal(int32(0755)))
		})
	})

	Describe("GetVolumeMount", func() {
		It("should return volume mount with correct properties", func() {
			g := NewScriptGeneratorForApp("default")
			vm := g.GetVolumeMount()

			Expect(vm.Name).To(Equal(appVolName))
			Expect(vm.MountPath).To(Equal(appScriptPath))
			Expect(vm.SubPath).To(Equal(appScriptName))
			Expect(vm.ReadOnly).To(BeTrue())
		})
	})

	Describe("appScriptContentSHA256 init", func() {
		It("should be initialized with a non-empty value at package init", func() {
			Expect(appScriptContentSHA256).NotTo(BeEmpty())
		})

		It("should be consistent with computeScriptSHA256 of the replaced script content", func() {
			expected := computeScriptSHA256(replacer.Replace(contentCheckMountReadyScript))
			Expect(appScriptContentSHA256).To(Equal(expected))
		})

		It("should match the dataset-level LabelAnnotationDatasetId label format", func() {
			// Verify label value length constraint: SHA256 hex is 64 chars, truncated to 63
			Expect(len(appScriptContentSHA256)).To(Equal(63))
		})
	})

	Describe("BuildConfigmap label consistency with GetScriptSHA256", func() {
		It("should have consistent SHA256 between label and GetScriptSHA256", func() {
			g := NewScriptGeneratorForApp("test-ns")
			cm := g.BuildConfigmap()
			sha := g.GetScriptSHA256()

			Expect(cm.Annotations[common.AnnotationCheckMountScriptSHA256]).To(Equal(sha))
		})
	})

	Describe("integration: BuildConfigmap with dataset-style label check", func() {
		It("should set label from a dataset-independent fixed hash", func() {
			// The app configmap SHA256 is fixed (not dataset-scoped), unlike dataset-level configmaps
			_ = &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sample-dataset",
					Namespace: "default",
					UID:       "sample-uid",
				},
			}
			g := NewScriptGeneratorForApp("default")
			cm := g.BuildConfigmap()

			// App configmap does NOT have LabelAnnotationDatasetId (dataset-independent)
			Expect(cm.Labels).NotTo(HaveKey(common.LabelAnnotationDatasetId))
			// But it MUST have the SHA256 annotation
			Expect(cm.Annotations).To(HaveKey(common.AnnotationCheckMountScriptSHA256))
		})
	})
})

var _ = Describe("scriptGeneratorHelper.RefreshConfigMapContents", func() {
	var (
		dataset      *datav1alpha1.Dataset
		configMapKey types.NamespacedName
		helper       *scriptGeneratorHelper
	)

	BeforeEach(func() {
		dataset = &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-dataset",
				Namespace: "default",
				UID:       "test-uid",
			},
		}
		configMapKey = types.NamespacedName{Name: "test-cm", Namespace: "default"}
		helper = &scriptGeneratorHelper{
			configMapName:  "test-config",
			scriptFileName: "check-mount.sh",
			scriptContent:  "#!/bin/bash\necho hello",
			scriptSHA256:   computeScriptSHA256("#!/bin/bash\necho hello"),
		}
	})

	It("should overwrite Data with the new script content", func() {
		existing := helper.BuildConfigMap(dataset, configMapKey)
		existing.Data[helper.scriptFileName] = "old content"

		helper.RefreshConfigMapContents(dataset, configMapKey, existing)

		Expect(existing.Data[helper.scriptFileName]).To(Equal("#!/bin/bash\necho hello"))
	})

	It("should set the SHA256 annotation on the existing configmap", func() {
		existing := helper.BuildConfigMap(dataset, configMapKey)
		delete(existing.Annotations, common.AnnotationCheckMountScriptSHA256)

		helper.RefreshConfigMapContents(dataset, configMapKey, existing)

		Expect(existing.Annotations).To(HaveKeyWithValue(common.AnnotationCheckMountScriptSHA256, helper.scriptSHA256))
	})

	It("should overwrite a stale SHA256 annotation", func() {
		existing := helper.BuildConfigMap(dataset, configMapKey)
		existing.Annotations[common.AnnotationCheckMountScriptSHA256] = "stale-sha"

		helper.RefreshConfigMapContents(dataset, configMapKey, existing)

		Expect(existing.Annotations[common.AnnotationCheckMountScriptSHA256]).To(Equal(helper.scriptSHA256))
	})

	It("should preserve extra labels not managed by the generator", func() {
		existing := helper.BuildConfigMap(dataset, configMapKey)
		existing.Labels["user-defined-label"] = "preserved"

		helper.RefreshConfigMapContents(dataset, configMapKey, existing)

		Expect(existing.Labels).To(HaveKeyWithValue("user-defined-label", "preserved"))
	})

	It("should preserve extra annotations not managed by the generator", func() {
		existing := helper.BuildConfigMap(dataset, configMapKey)
		existing.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = "some-value"

		helper.RefreshConfigMapContents(dataset, configMapKey, existing)

		Expect(existing.Annotations).To(HaveKey("kubectl.kubernetes.io/last-applied-configuration"))
	})

	It("should initialize nil Labels before merging", func() {
		existing := helper.BuildConfigMap(dataset, configMapKey)
		existing.Labels = nil

		helper.RefreshConfigMapContents(dataset, configMapKey, existing)

		Expect(existing.Labels).NotTo(BeNil())
		Expect(existing.Labels).To(HaveKey(common.LabelAnnotationDatasetId))
	})

	It("should initialize nil Annotations before merging", func() {
		existing := helper.BuildConfigMap(dataset, configMapKey)
		existing.Annotations = nil

		helper.RefreshConfigMapContents(dataset, configMapKey, existing)

		Expect(existing.Annotations).NotTo(BeNil())
		Expect(existing.Annotations).To(HaveKey(common.AnnotationCheckMountScriptSHA256))
	})

	It("should return the same object pointer", func() {
		existing := helper.BuildConfigMap(dataset, configMapKey)
		result := helper.RefreshConfigMapContents(dataset, configMapKey, existing)

		Expect(result).To(BeIdenticalTo(existing))
	})
})
