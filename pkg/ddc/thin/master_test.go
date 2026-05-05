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

package thin

import (
	"context"
	"errors"
	. "github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Master Tests", func() {
	Describe("ThinEngine.CheckMasterReady", func() {
		Context("when fuse daemonset exists", func() {
			It("should return ready", func() {
				daemonSet := &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-fuse",
						Namespace: "fluid",
					},
				}
				testObjs := []runtime.Object{daemonSet.DeepCopy()}
				client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

				engine := ThinEngine{
					name:      "test",
					namespace: "fluid",
					Client:    client,
				}

				ready, err := engine.CheckMasterReady()
				Expect(err).To(BeNil())
				Expect(ready).To(BeTrue())
			})
		})

		Context("when fuse daemonset does not exist", func() {
			It("should return error", func() {
				testObjs := []runtime.Object{}
				client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

				engine := ThinEngine{
					name:      "notexist",
					namespace: "fluid",
					Client:    client,
				}

				ready, err := engine.CheckMasterReady()
				Expect(err).NotTo(BeNil())
				Expect(ready).To(BeFalse())
			})
		})
	})

	Describe("ThinEngine.ShouldSetupMaster", func() {
		DescribeTable("should correctly determine if master setup is needed",
			func(phase datav1alpha1.RuntimePhase, expected bool) {
				runtimeInput := &datav1alpha1.ThinRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-" + string(phase),
						Namespace: "fluid",
					},
					Status: datav1alpha1.RuntimeStatus{
						FusePhase: phase,
					},
				}
				testObjs := []runtime.Object{runtimeInput.DeepCopy()}
				client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

				engine := ThinEngine{
					name:      "test-" + string(phase),
					namespace: "fluid",
					Client:    client,
				}

				should, err := engine.ShouldSetupMaster()
				Expect(err).To(BeNil())
				Expect(should).To(Equal(expected))
			},
			Entry("None phase - should setup", datav1alpha1.RuntimePhaseNone, true),
			Entry("Ready phase - should not setup", datav1alpha1.RuntimePhaseReady, false),
			Entry("NotReady phase - should not setup", datav1alpha1.RuntimePhaseNotReady, false),
		)
	})

	Describe("ThinEngine.SetupMaster", func() {
		Context("when fuse daemonset already exists", func() {
			It("should return no error", func() {
				daemonSet := &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "existing-fuse",
						Namespace: "fluid",
					},
				}
				testObjs := []runtime.Object{daemonSet.DeepCopy()}
				client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

				engine := ThinEngine{
					name:      "existing",
					namespace: "fluid",
					Client:    client,
					Log:       fake.NullLogger(),
				}

				err := engine.SetupMaster()
				Expect(err).To(BeNil())
			})
		})

	})

	Describe("ThinEngine.generateThinValueFile", func() {
		It("generates values when the runtime profile lookup falls back to nil", func() {
			dataset, runtimeObj, _ := mockFluidObjectsForTests(types.NamespacedName{Name: "test-dataset", Namespace: "default"})
			runtimeObj.Spec.Fuse = datav1alpha1.ThinFuseSpec{
				Image: "runtime-fuse",
			}

			client := fake.NewFakeClientWithScheme(testScheme, dataset, runtimeObj)
			engine := mockThinEngineForTests(dataset, runtimeObj, nil)
			engine.Client = client
			engine.runtime = runtimeObj

			generatedProfile, err := engine.getThinRuntimeProfile()
			Expect(err).To(HaveOccurred())
			Expect(generatedProfile).To(BeNil())

			valueFile, err := engine.generateThinValueFile(runtimeObj, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(valueFile).To(BeAnExistingFile())

			configMap := &corev1.ConfigMap{}
			err = engine.Client.Get(context.TODO(), types.NamespacedName{
				Name:      engine.getHelmValuesConfigMapName(),
				Namespace: engine.namespace,
			}, configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap.Labels).To(HaveKeyWithValue(common.LabelAnnotationDatasetId, dataset.Labels[common.LabelAnnotationDatasetId]))
			Expect(configMap.Data).To(HaveKey("data"))
		})

		It("skips storing runtime helm values when runtime config map generation is disabled", func() {
			dataset, runtimeObj, profile := mockFluidObjectsForTests(types.NamespacedName{Name: "test-dataset", Namespace: "default"})
			runtimeObj.Spec.Fuse = datav1alpha1.ThinFuseSpec{Image: "runtime-fuse"}
			runtimeObj.Annotations = map[string]string{common.AnnotationDisableRuntimeHelmValueConfig: "false"}

			engine := mockThinEngineForTests(dataset, runtimeObj, profile)
			engine.Client = fake.NewFakeClientWithScheme(testScheme, dataset, runtimeObj, profile)
			engine.runtime = runtimeObj

			valueFile, err := engine.generateThinValueFile(runtimeObj, profile)
			Expect(err).NotTo(HaveOccurred())
			Expect(valueFile).To(BeAnExistingFile())

			configMap := &corev1.ConfigMap{}
			err = engine.Client.Get(context.TODO(), types.NamespacedName{
				Name:      engine.getHelmValuesConfigMapName(),
				Namespace: engine.namespace,
			}, configMap)
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
		})

	})

	Describe("ThinEngine.setupMasterInternal", func() {
		It("continues with a missing runtime profile when the release already exists", func() {
			dataset, runtimeObj, _ := mockFluidObjectsForTests(types.NamespacedName{Name: "test-dataset", Namespace: "default"})
			runtimeObj.Spec.Fuse = datav1alpha1.ThinFuseSpec{Image: "runtime-fuse"}

			engine := mockThinEngineForTests(dataset, runtimeObj, nil)
			engine.Client = fake.NewFakeClientWithScheme(testScheme, dataset, runtimeObj)
			engine.runtime = runtimeObj

			checkReleasePatch := ApplyFunc(helm.CheckRelease, func(name string, namespace string) (bool, error) {
				Expect(name).To(Equal(engine.name))
				Expect(namespace).To(Equal(engine.namespace))
				return true, nil
			})
			installReleasePatch := ApplyFunc(helm.InstallRelease, func(string, string, string, string) error {
				Fail("InstallRelease should not be called when the release already exists")
				return nil
			})
			defer checkReleasePatch.Reset()
			defer installReleasePatch.Reset()

			Expect(engine.setupMasterInternal()).To(Succeed())
		})

		It("installs the release when it is not already present", func() {
			dataset, runtimeObj, profile := mockFluidObjectsForTests(types.NamespacedName{Name: "test-dataset", Namespace: "default"})
			runtimeObj.Spec.Fuse = datav1alpha1.ThinFuseSpec{Image: "runtime-fuse"}

			engine := mockThinEngineForTests(dataset, runtimeObj, profile)
			engine.Client = fake.NewFakeClientWithScheme(testScheme, dataset, runtimeObj, profile)
			engine.runtime = runtimeObj

			checkReleasePatch := ApplyFunc(helm.CheckRelease, func(string, string) (bool, error) {
				return false, nil
			})
			installReleasePatch := ApplyFunc(helm.InstallRelease, func(name string, namespace string, valueFile string, chart string) error {
				Expect(name).To(Equal(engine.name))
				Expect(namespace).To(Equal(engine.namespace))
				Expect(valueFile).To(BeAnExistingFile())
				Expect(chart).To(ContainSubstring("thin"))
				return nil
			})
			defer checkReleasePatch.Reset()
			defer installReleasePatch.Reset()

			Expect(engine.setupMasterInternal()).To(Succeed())
		})

		It("returns the install release error", func() {
			dataset, runtimeObj, profile := mockFluidObjectsForTests(types.NamespacedName{Name: "test-dataset", Namespace: "default"})
			runtimeObj.Spec.Fuse = datav1alpha1.ThinFuseSpec{Image: "runtime-fuse"}

			engine := mockThinEngineForTests(dataset, runtimeObj, profile)
			engine.Client = fake.NewFakeClientWithScheme(testScheme, dataset, runtimeObj, profile)
			engine.runtime = runtimeObj

			checkReleasePatch := ApplyFunc(helm.CheckRelease, func(string, string) (bool, error) {
				return false, nil
			})
			installReleasePatch := ApplyFunc(helm.InstallRelease, func(string, string, string, string) error {
				return errors.New("install failed")
			})
			defer checkReleasePatch.Reset()
			defer installReleasePatch.Reset()

			Expect(engine.setupMasterInternal()).To(MatchError("install failed"))
		})
	})

	Describe("ThinEngine.getThinRuntimeProfile", func() {
		It("returns nil when the engine runtime is nil", func() {
			engine := &ThinEngine{}

			profile, err := engine.getThinRuntimeProfile()

			Expect(err).NotTo(HaveOccurred())
			Expect(profile).To(BeNil())
		})

		It("loads the referenced runtime profile", func() {
			dataset, runtimeObj, profile := mockFluidObjectsForTests(types.NamespacedName{Name: "test-dataset", Namespace: "default"})
			engine := mockThinEngineForTests(dataset, runtimeObj, profile)
			engine.Client = fake.NewFakeClientWithScheme(testScheme, dataset, runtimeObj, profile)

			loadedProfile, err := engine.getThinRuntimeProfile()

			Expect(err).NotTo(HaveOccurred())
			Expect(loadedProfile).NotTo(BeNil())
			Expect(loadedProfile.Name).To(Equal(profile.Name))
		})
	})

	Describe("ThinEngine.ifRuntimeHelmValueEnable", func() {
		It("defaults to enabled when runtime is nil or the annotation is invalid", func() {
			Expect((&ThinEngine{}).ifRuntimeHelmValueEnable()).To(BeTrue())

			engine := &ThinEngine{runtime: &datav1alpha1.ThinRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{common.AnnotationDisableRuntimeHelmValueConfig: "not-a-bool"},
				},
			}}

			Expect(engine.ifRuntimeHelmValueEnable()).To(BeTrue())
		})

		It("follows the parsed runtime annotation value", func() {
			engine := &ThinEngine{runtime: &datav1alpha1.ThinRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{common.AnnotationDisableRuntimeHelmValueConfig: "false"},
				},
			}}

			Expect(engine.ifRuntimeHelmValueEnable()).To(BeFalse())
		})
	})
})
