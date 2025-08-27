/*
Copyright 2022 The Fluid Authors.

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

package alluxio

import (
	"fmt"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AlluxioEngine setup master internal tests", Label("pkg.ddc.alluxio.master_internal_test.go"), func() {
	var (
		dataset        *datav1alpha1.Dataset
		alluxioruntime *datav1alpha1.AlluxioRuntime
		engine         *AlluxioEngine
		mockedObjects  mockedObjects
		client         client.Client
		resources      []runtime.Object
	)

	BeforeEach(func() {
		dataset, alluxioruntime = mockFluidObjectsForTests(types.NamespacedName{Namespace: "fluid", Name: "hbase"})
		engine = mockAlluxioEngineForTests(dataset, alluxioruntime)
		mockedObjects = mockAlluxioObjectsForTests(dataset, alluxioruntime, engine)
		resources = []runtime.Object{
			dataset,
			alluxioruntime,
			mockedObjects.MasterSts,
			mockedObjects.WorkerSts,
			mockedObjects.FuseDs,
			mockedObjects.PersistentVolumeClaim,
			mockedObjects.PersistentVolume,
		}
	})

	// JustBeforeEach is guaranteed to run after every BeforeEach()
	// So it's easy to modify resources' specs with an extra BeforeEach()
	JustBeforeEach(func() {
		client = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, resources...)
		engine.Client = client
	})

	Describe("Test AlluxioEngine.setupMasterInternal", func() {
		// generateAlluxioValueFile is not a test target in this function, patch it with a dummy function
		var transformFuncPatch *gomonkey.Patches
		BeforeEach(func() {
			transformFuncPatch = gomonkey.ApplyPrivateMethod(engine, "generateAlluxioValueFile", func(runtime *datav1alpha1.AlluxioRuntime) (string, error) {
				return "<helm-tmp-value-file>", nil
			})
		})
		AfterEach(func() {
			transformFuncPatch.Reset()
		})
		When("helm release is already installed", func() {
			It("should not install helm release", func() {
				patch := gomonkey.ApplyFunc(helm.CheckRelease, func(name string, namespace string) (exist bool, err error) {
					return true, nil
				})
				defer patch.Reset()

				patch2 := gomonkey.ApplyFunc(helm.InstallRelease, func(name string, namespace string, valueFile string, chartName string) error {
					return fmt.Errorf("should not call helm.InstallRelease")
				})
				defer patch2.Reset()

				err := engine.setupMasterInternal()
				Expect(err).To(BeNil())
			})
		})

		When("helm release is not installed", func() {
			It("should install helm release", func() {
				patch := gomonkey.ApplyFunc(helm.CheckRelease, func(name string, namespace string) (exist bool, err error) {
					return false, nil
				})
				defer patch.Reset()

				patch2 := gomonkey.ApplyFunc(helm.InstallRelease, func(name string, namespace string, valueFile string, chartName string) error {
					return nil
				})
				defer patch2.Reset()

				err := engine.setupMasterInternal()
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("Test AlluxioEngine.generateAlluxioValueFile()", func() {
		When("engine works as expected", func() {
			It("value file should be created and no error should be returned", func() {
				mockedValue := Alluxio{
					FullnameOverride: alluxioruntime.Name,
					ImageInfo: common.ImageInfo{
						Image:    "alluxio/alluxio",
						ImageTag: "2.8.0",
					},
					Master: Master{
						Env: map[string]string{
							"test-master-env": "true",
						},
					},
					Worker: Worker{
						Env: map[string]string{
							"test-worker-env": "true",
						},
					},
					Fuse: Fuse{
						Env: map[string]string{
							"test-fuse-env": "true",
						},
						Image:    "alluxio/alluxio-fuse",
						ImageTag: "2.8.0",
					},
				}

				patch := gomonkey.ApplyPrivateMethod(engine, "transform", func() (value *Alluxio, err error) {
					// mock a simple Alluxio value
					return &mockedValue, nil
				})

				defer patch.Reset()

				valueFile, err := engine.generateAlluxioValueFile(alluxioruntime)
				Expect(err).To(BeNil())
				Expect(valueFile).NotTo(HaveLen(0))

				// value file should exist
				_, err = os.Stat(valueFile)
				Expect(err).To(BeNil())

				// check content in the value file
				expectedValue := Alluxio{}
				valueBytes, err := os.ReadFile(valueFile)
				Expect(err).To(BeNil(), "error when calling os.ReadFile")
				err = yaml.Unmarshal(valueBytes, &expectedValue)
				Expect(err).To(BeNil(), "error when calling yaml.unmarshalling value")
				Expect(expectedValue).To(Equal(mockedValue))

				defer os.Remove(valueFile)

				// configmap should exists
				cm, err := kubeclient.GetConfigmapByName(engine.Client, engine.getHelmValuesConfigMapName(), engine.namespace)
				Expect(err).To(BeNil())
				Expect(cm).NotTo(BeNil())
			})
		})

		When("configmap with same name exists in the cluster", func() {
			BeforeEach(func() {
				helmValueConfigmap := corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      engine.getHelmValuesConfigMapName(),
						Namespace: engine.namespace,
					},
					Data: map[string]string{
						"test": "test",
					},
				}

				resources = append(resources, &helmValueConfigmap)
			})

			It("should delete the old configmap and create a new one", func() {
				patch := gomonkey.ApplyPrivateMethod(engine, "transform", func() (value *Alluxio, err error) {
					// mock a simple Alluxio value
					return &Alluxio{}, nil
				})
				defer patch.Reset()

				valueFile, err := engine.generateAlluxioValueFile(alluxioruntime)
				Expect(err).To(BeNil())
				Expect(valueFile).NotTo(HaveLen(0))

				cm, err := kubeclient.GetConfigmapByName(engine.Client, engine.getHelmValuesConfigMapName(), engine.namespace)
				Expect(err).To(BeNil())
				Expect(cm).NotTo(BeNil())

				Expect(cm.Data).NotTo(HaveKeyWithValue("test", "test"))
			})
		})
	})
})

func TestGetConfigmapName(t *testing.T) {
	engine := AlluxioEngine{
		name:       "hbase",
		engineImpl: "alluxio",
	}
	expectedResult := "hbase-alluxio-values"
	if engine.getHelmValuesConfigMapName() != expectedResult {
		t.Errorf("fail to get the configmap name")
	}
}
