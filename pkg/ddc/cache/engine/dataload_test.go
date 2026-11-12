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
	"context"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdataload "github.com/fluid-cloudnative/fluid/pkg/dataload"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("DataLoad genDataLoadValue Tests", Label("pkg.ddc.cache.engine.dataload_test.go"), func() {
	var (
		scheme       *runtime.Scheme
		engine       *CacheEngine
		ctx          cruntime.ReconcileRequestContext
		dataset      *datav1alpha1.Dataset
		runtimeObj   *datav1alpha1.CacheRuntime
		runtimeClass *datav1alpha1.CacheRuntimeClass
		dataload     *datav1alpha1.DataLoad
		baseClient   client.Client
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		_ = corev1.AddToScheme(scheme)
		_ = datav1alpha1.AddToScheme(scheme)

		runtimeObj = &datav1alpha1.CacheRuntime{
			TypeMeta: metav1.TypeMeta{APIVersion: "data.fluid.io/v1alpha1", Kind: datav1alpha1.CacheRuntimeKind},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "demo",
				Namespace: "default",
				UID:       "demo-uid",
			},
			Spec: datav1alpha1.CacheRuntimeSpec{
				RuntimeClassName: "test-class",
				Master:           datav1alpha1.CacheRuntimeMasterSpec{RuntimeComponentCommonSpec: datav1alpha1.RuntimeComponentCommonSpec{Disabled: true}},
				Worker:           datav1alpha1.CacheRuntimeWorkerSpec{RuntimeComponentCommonSpec: datav1alpha1.RuntimeComponentCommonSpec{Disabled: true}},
				Client:           datav1alpha1.CacheRuntimeClientSpec{RuntimeComponentCommonSpec: datav1alpha1.RuntimeComponentCommonSpec{Disabled: true}},
			},
		}

		runtimeClass = &datav1alpha1.CacheRuntimeClass{
			ObjectMeta: metav1.ObjectMeta{Name: "test-class"},
			Topology: &datav1alpha1.RuntimeTopology{
				Worker: &datav1alpha1.RuntimeComponentDefinition{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{Image: "fluidio/fluid:v1.0.0"},
							},
						},
					},
				},
			},
			DataOperationSpecs: []datav1alpha1.DataOperationSpec{
				{
					Name:    "DataLoad",
					Command: []string{"/usr/local/bin/dataload"},
					Args:    []string{"--config", "/etc/fluid/config/runtime.json"},
				},
			},
		}

		dataset = &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{Name: "demo", Namespace: "default"},
			Spec: datav1alpha1.DatasetSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadOnlyMany},
				Mounts: []datav1alpha1.Mount{
					{
						Name:       "hbase",
						MountPoint: "local:///data",
						Path:       "/data",
						ReadOnly:   true,
						Shared:     true,
					},
				},
			},
			Status: datav1alpha1.DatasetStatus{Runtimes: []datav1alpha1.Runtime{{Name: "demo", Type: common.CacheRuntime}}},
		}

		dataload = &datav1alpha1.DataLoad{
			ObjectMeta: metav1.ObjectMeta{Name: "test-load", Namespace: "default"},
			Spec: datav1alpha1.DataLoadSpec{
				Dataset:      datav1alpha1.TargetDataset{Name: "demo", Namespace: "default"},
				LoadMetadata: true,
				Policy:       datav1alpha1.Once,
				Target:       []datav1alpha1.TargetPath{{Path: "/data"}},
				Resources:    corev1.ResourceRequirements{},
			},
		}

		baseClient = fake.NewFakeClientWithScheme(scheme, runtimeObj, runtimeClass, dataset)
		engine = &CacheEngine{Client: baseClient, name: "demo", namespace: "default"}
		ctx = cruntime.ReconcileRequestContext{Context: context.Background()}
	})

	Describe("genDataLoadValue", func() {
		Context("when volume/mount wiring is correct", func() {
			It("should produce volumes referencing the correct ConfigMap", func() {
				val, err := engine.genDataLoadValue(ctx, dataset, runtimeObj, runtimeClass, dataload)
				Expect(err).NotTo(HaveOccurred())

				Expect(val.DataLoadInfo.Volumes).To(HaveLen(1))
				vol := val.DataLoadInfo.Volumes[0]
				Expect(vol.Name).To(Equal(engine.getRuntimeConfigVolumeName()))
				Expect(vol.ConfigMap).NotTo(BeNil())
				Expect(vol.ConfigMap.LocalObjectReference.Name).To(Equal(common.GetCacheRuntimeConfigConfigMapName(engine.name)))
			})

			It("should produce volume mounts at the correct directory", func() {
				val, err := engine.genDataLoadValue(ctx, dataset, runtimeObj, runtimeClass, dataload)
				Expect(err).NotTo(HaveOccurred())

				Expect(val.DataLoadInfo.VolumeMounts).To(HaveLen(1))
				vm := val.DataLoadInfo.VolumeMounts[0]
				Expect(vm.Name).To(Equal(engine.getRuntimeConfigVolumeName()))
				Expect(vm.MountPath).To(Equal(engine.getRuntimeConfigDir()))
				Expect(vm.ReadOnly).To(BeTrue())
			})

			It("should set FLUID_RUNTIME_CONFIG_PATH env to match mount path + config file name", func() {
				val, err := engine.genDataLoadValue(ctx, dataset, runtimeObj, runtimeClass, dataload)
				Expect(err).NotTo(HaveOccurred())

				var foundEnv *cdataload.Env
				for i, env := range val.DataLoadInfo.Envs {
					if env.Name == "FLUID_RUNTIME_CONFIG_PATH" {
						foundEnv = &val.DataLoadInfo.Envs[i]
						break
					}
				}
				Expect(foundEnv).NotTo(BeNil())
				Expect(foundEnv.Value).To(Equal(engine.getRuntimeConfigPath()))
			})

			It("should keep mount dir and runtime config path in sync", func() {
				_, err := engine.genDataLoadValue(ctx, dataset, runtimeObj, runtimeClass, dataload)
				Expect(err).NotTo(HaveOccurred())

				// Verify the mount path + config file name equals the config path
				mountPath := engine.getRuntimeConfigDir()
				configFileName := engine.getRuntimeConfigFileName()
				configPath := engine.getRuntimeConfigPath()

				Expect(configPath).To(Equal(mountPath + "/" + configFileName))
			})
		})

		Context("when DataLoad spec is not defined in runtime class", func() {
			BeforeEach(func() {
				runtimeClass.DataOperationSpecs = nil
			})

			It("should return a not-supported error", func() {
				_, err := engine.genDataLoadValue(ctx, dataset, runtimeObj, runtimeClass, dataload)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when DataLoad command and args are both empty", func() {
			BeforeEach(func() {
				runtimeClass.DataOperationSpecs = []datav1alpha1.DataOperationSpec{
					{
						Name:    "DataLoad",
						Command: []string{},
						Args:    []string{},
					},
				}
			})

			It("should return an error", func() {
				recorder := &record.FakeRecorder{}
				testCtx := cruntime.ReconcileRequestContext{Context: context.Background(), Recorder: recorder}
				_, err := engine.genDataLoadValue(testCtx, dataset, runtimeObj, runtimeClass, dataload)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
