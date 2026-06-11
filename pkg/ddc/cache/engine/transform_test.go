/*
Copyright 2026 The Fluid Authors.

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

package engine

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("CacheEngine Transform Tests", Label("pkg.ddc.cache.engine.transform_test.go"), func() {
	var (
		engine       *CacheEngine
		dataset      *datav1alpha1.Dataset
		runtimeObj   *datav1alpha1.CacheRuntime
		runtimeClass *datav1alpha1.CacheRuntimeClass
	)

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(scheme)

		dataset = &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-dataset",
				Namespace: "default",
				UID:       "test-dataset-uid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		}

		runtimeObj = &datav1alpha1.CacheRuntime{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "data.fluid.io/v1alpha1",
				Kind:       "CacheRuntime",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-runtime",
				Namespace: "default",
				UID:       "test-runtime-uid",
			},
			Spec: datav1alpha1.CacheRuntimeSpec{
				RuntimeClassName: "test-class",
				Master:           datav1alpha1.CacheRuntimeMasterSpec{Replicas: 1},
				Worker:           datav1alpha1.CacheRuntimeWorkerSpec{Replicas: 2},
				Client:           datav1alpha1.CacheRuntimeClientSpec{},
			},
		}

		runtimeClass = &datav1alpha1.CacheRuntimeClass{
			ObjectMeta:     metav1.ObjectMeta{Name: "test-class"},
			FileSystemType: "test-fs",
			Topology: &datav1alpha1.RuntimeTopology{
				Master: &datav1alpha1.RuntimeComponentDefinition{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "master", Image: "test-master:latest"}},
						},
					},
				},
				Worker: &datav1alpha1.RuntimeComponentDefinition{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "worker", Image: "test-worker:latest"}},
						},
					},
				},
				Client: &datav1alpha1.RuntimeComponentDefinition{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "client", Image: "test-client:latest"}},
						},
					},
				},
			},
		}

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(dataset, runtimeObj).Build()

		engine = &CacheEngine{
			name:      "test-runtime",
			namespace: "default",
			Client:    fakeClient,
			Log:       ctrl.Log.WithName("test"),
		}
	})

	Describe("transform", func() {
		Context("when topology is nil", func() {
			BeforeEach(func() {
				runtimeClass.Topology = nil
			})

			It("should return error", func() {
				value, err := engine.transform(dataset, runtimeObj, runtimeClass)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("at least one component should be defined"))
				Expect(value).To(BeNil())
			})
		})

		Context("when all components are nil", func() {
			BeforeEach(func() {
				runtimeClass.Topology.Master = nil
				runtimeClass.Topology.Worker = nil
				runtimeClass.Topology.Client = nil
			})

			It("should return error", func() {
				value, err := engine.transform(dataset, runtimeObj, runtimeClass)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("at least one component should be defined"))
				Expect(value).To(BeNil())
			})
		})

		Context("when only Master is defined", func() {
			BeforeEach(func() {
				runtimeClass.Topology.Worker = nil
				runtimeClass.Topology.Client = nil
			})

			It("should transform successfully with only Master", func() {
				value, err := engine.transform(dataset, runtimeObj, runtimeClass)
				Expect(err).NotTo(HaveOccurred())
				Expect(value).NotTo(BeNil())
				Expect(value.Master).NotTo(BeNil())
				Expect(value.Master.Enabled).To(BeTrue())
			})
		})

		Context("when only Worker is defined", func() {
			BeforeEach(func() {
				runtimeClass.Topology.Master = nil
				runtimeClass.Topology.Client = nil
			})

			It("should transform successfully with only Worker", func() {
				value, err := engine.transform(dataset, runtimeObj, runtimeClass)
				Expect(err).NotTo(HaveOccurred())
				Expect(value).NotTo(BeNil())
				Expect(value.Worker).NotTo(BeNil())
				Expect(value.Worker.Enabled).To(BeTrue())
			})
		})

		Context("when only Client is defined", func() {
			BeforeEach(func() {
				runtimeClass.Topology.Master = nil
				runtimeClass.Topology.Worker = nil
			})

			It("should transform successfully with only Client", func() {
				value, err := engine.transform(dataset, runtimeObj, runtimeClass)
				Expect(err).NotTo(HaveOccurred())
				Expect(value).NotTo(BeNil())
				Expect(value.Client).NotTo(BeNil())
				Expect(value.Client.Enabled).To(BeTrue())
			})
		})

		Context("when all components are defined", func() {
			It("should transform all components successfully", func() {
				value, err := engine.transform(dataset, runtimeObj, runtimeClass)
				Expect(err).NotTo(HaveOccurred())
				Expect(value).NotTo(BeNil())
				Expect(value.Master).NotTo(BeNil())
				Expect(value.Master.Enabled).To(BeTrue())
				Expect(value.Worker).NotTo(BeNil())
				Expect(value.Worker.Enabled).To(BeTrue())
				Expect(value.Client).NotTo(BeNil())
				Expect(value.Client.Enabled).To(BeTrue())
			})
		})

		Context("with ExtraResources ConfigMaps", func() {
			BeforeEach(func() {
				runtimeClass.ExtraResources.ConfigMaps = []datav1alpha1.ConfigMapRuntimeExtraResource{
					{Name: "extra-config-1", Data: map[string]string{"key1": "value1"}},
					{Name: "extra-config-2", Data: map[string]string{"key2": "value2"}},
				}
			})

			It("should include extra configmap names in common config", func() {
				value, err := engine.transform(dataset, runtimeObj, runtimeClass)
				Expect(err).NotTo(HaveOccurred())
				Expect(value).NotTo(BeNil())
				Expect(value.Master).NotTo(BeNil())
			})
		})

		Context("common config generation", func() {
			It("should generate correct owner reference", func() {
				value, err := engine.transform(dataset, runtimeObj, runtimeClass)
				Expect(err).NotTo(HaveOccurred())
				Expect(value).NotTo(BeNil())
				Expect(value.Master.Owner).NotTo(BeNil())
				Expect(value.Master.Owner.Name).To(Equal("test-runtime"))
				Expect(value.Master.Owner.Kind).To(Equal("CacheRuntime"))
			})

			It("should generate runtime config volume", func() {
				value, err := engine.transform(dataset, runtimeObj, runtimeClass)
				Expect(err).NotTo(HaveOccurred())
				Expect(value).NotTo(BeNil())
				volumeNames := make([]string, len(value.Master.PodTemplateSpec.Spec.Volumes))
				for i, vol := range value.Master.PodTemplateSpec.Spec.Volumes {
					volumeNames[i] = vol.Name
				}
				Expect(volumeNames).To(ContainElement(ContainSubstring("runtime-config")))
			})
		})
	})

	Describe("getRuntimeStatusValue", func() {
		Context("when topology is nil", func() {
			BeforeEach(func() {
				runtimeClass.Topology = nil
			})

			It("should return error", func() {
				statusValue, err := engine.getRuntimeStatusValue(runtimeObj, runtimeClass)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("at least one component should be defined"))
				Expect(statusValue).To(BeNil())
			})
		})

		Context("when all components are nil", func() {
			BeforeEach(func() {
				runtimeClass.Topology.Master = nil
				runtimeClass.Topology.Worker = nil
				runtimeClass.Topology.Client = nil
			})

			It("should return error", func() {
				statusValue, err := engine.getRuntimeStatusValue(runtimeObj, runtimeClass)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("at least one component should be defined"))
				Expect(statusValue).To(BeNil())
			})
		})

		Context("when Master is disabled", func() {
			BeforeEach(func() {
				runtimeObj.Spec.Master.Disabled = true
			})

			It("should set Master as disabled in status value", func() {
				statusValue, err := engine.getRuntimeStatusValue(runtimeObj, runtimeClass)
				Expect(err).NotTo(HaveOccurred())
				Expect(statusValue).NotTo(BeNil())
				Expect(statusValue.Master).NotTo(BeNil())
				Expect(statusValue.Master.Enabled).To(BeFalse())
			})
		})

		Context("when Worker is disabled", func() {
			BeforeEach(func() {
				runtimeObj.Spec.Worker.Disabled = true
			})

			It("should set Worker as disabled in status value", func() {
				statusValue, err := engine.getRuntimeStatusValue(runtimeObj, runtimeClass)
				Expect(err).NotTo(HaveOccurred())
				Expect(statusValue).NotTo(BeNil())
				Expect(statusValue.Worker).NotTo(BeNil())
				Expect(statusValue.Worker.Enabled).To(BeFalse())
			})
		})

		Context("when Client is disabled", func() {
			BeforeEach(func() {
				runtimeObj.Spec.Client.Disabled = true
			})

			It("should set Client as disabled in status value", func() {
				statusValue, err := engine.getRuntimeStatusValue(runtimeObj, runtimeClass)
				Expect(err).NotTo(HaveOccurred())
				Expect(statusValue).NotTo(BeNil())
				Expect(statusValue.Client).NotTo(BeNil())
				Expect(statusValue.Client.Enabled).To(BeFalse())
			})
		})

		Context("when all components are enabled", func() {
			It("should extract status info for all components", func() {
				statusValue, err := engine.getRuntimeStatusValue(runtimeObj, runtimeClass)
				Expect(err).NotTo(HaveOccurred())
				Expect(statusValue).NotTo(BeNil())

				Expect(statusValue.Master).NotTo(BeNil())
				Expect(statusValue.Master.Enabled).To(BeTrue())
				Expect(statusValue.Master.ComponentIdentity.Name).To(Equal("test-runtime-master"))
				Expect(statusValue.Master.ComponentIdentity.Namespace).To(Equal("default"))

				Expect(statusValue.Worker).NotTo(BeNil())
				Expect(statusValue.Worker.Enabled).To(BeTrue())
				Expect(statusValue.Worker.ComponentIdentity.Name).To(Equal("test-runtime-worker"))
				Expect(statusValue.Worker.ComponentIdentity.Namespace).To(Equal("default"))

				Expect(statusValue.Client).NotTo(BeNil())
				Expect(statusValue.Client.Enabled).To(BeTrue())
				Expect(statusValue.Client.ComponentIdentity.Name).To(Equal("test-runtime-client"))
				Expect(statusValue.Client.ComponentIdentity.Namespace).To(Equal("default"))
			})
		})

		Context("component name generation", func() {
			It("should generate correct component names", func() {
				statusValue, err := engine.getRuntimeStatusValue(runtimeObj, runtimeClass)
				Expect(err).NotTo(HaveOccurred())
				Expect(statusValue.Master.ComponentIdentity.Name).To(Equal(common.GetCacheComponentName("test-runtime", common.ComponentTypeMaster)))
				Expect(statusValue.Worker.ComponentIdentity.Name).To(Equal(common.GetCacheComponentName("test-runtime", common.ComponentTypeWorker)))
				Expect(statusValue.Client.ComponentIdentity.Name).To(Equal(common.GetCacheComponentName("test-runtime", common.ComponentTypeClient)))
			})
		})
	})

	Describe("transformComponentCommonConfig", func() {
		It("should generate owner reference from runtime object", func() {
			config, err := engine.transformComponentCommonConfig(runtimeObj, runtimeClass)
			Expect(err).NotTo(HaveOccurred())
			Expect(config).NotTo(BeNil())
			Expect(config.Owner).NotTo(BeNil())
			Expect(config.Owner.Name).To(Equal("test-runtime"))
			Expect(config.Owner.Kind).To(Equal("CacheRuntime"))
			Expect(config.Owner.APIVersion).To(Equal("data.fluid.io/v1alpha1"))
		})

		It("should generate runtime config volume", func() {
			config, err := engine.transformComponentCommonConfig(runtimeObj, runtimeClass)
			Expect(err).NotTo(HaveOccurred())
			Expect(config.RuntimeConfigs).NotTo(BeNil())
			Expect(config.RuntimeConfigs.RuntimeConfigVolume.Name).To(ContainSubstring("runtime-config"))
			Expect(config.RuntimeConfigs.RuntimeConfigVolumeMount.Name).To(ContainSubstring("runtime-config"))
			Expect(config.RuntimeConfigs.RuntimeConfigVolumeMount.MountPath).To(Equal("/etc/fluid/config"))
			Expect(config.RuntimeConfigs.RuntimeConfigVolumeMount.ReadOnly).To(BeTrue())
		})

		Context("with ExtraResources ConfigMaps", func() {
			BeforeEach(func() {
				runtimeClass.ExtraResources.ConfigMaps = []datav1alpha1.ConfigMapRuntimeExtraResource{
					{Name: "config-1"},
					{Name: "config-2"},
				}
			})

			It("should populate ExtraConfigMapNames", func() {
				config, err := engine.transformComponentCommonConfig(runtimeObj, runtimeClass)
				Expect(err).NotTo(HaveOccurred())
				Expect(config.RuntimeConfigs.ExtraConfigMapNames).NotTo(BeNil())
				Expect(config.RuntimeConfigs.ExtraConfigMapNames).To(HaveKey("config-1"))
				Expect(config.RuntimeConfigs.ExtraConfigMapNames).To(HaveKey("config-2"))
				Expect(len(config.RuntimeConfigs.ExtraConfigMapNames)).To(Equal(2))
			})
		})

		Context("without ExtraResources", func() {
			It("should not initialize ExtraConfigMapNames", func() {
				config, err := engine.transformComponentCommonConfig(runtimeObj, runtimeClass)
				Expect(err).NotTo(HaveOccurred())
				Expect(config.RuntimeConfigs.ExtraConfigMapNames).To(BeNil())
			})
		})
	})
})
