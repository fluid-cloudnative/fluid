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

package alluxio

import (
	"fmt"

	"github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AlluxioEngine HCFS Tests", Label("pkg.ddc.alluxio.hcfs_test.go"), func() {
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
			mockedObjects.Services[fmt.Sprintf("%s-master-0", alluxioruntime.Name)],
		}
	})

	// JustBeforeEach is guaranteed to run after every BeforeEach()
	// So it's easy to modify resources' specs with an extra BeforeEach()
	JustBeforeEach(func() {
		client = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, resources...)
		engine.Client = client
	})

	Describe("Test AlluxioEngine.GetHCFSStatus()", func() {
		When("given AlluxioEngine works as expected", func() {
			portNum := 19999
			BeforeEach(func() {
				mockedObjects.Services[fmt.Sprintf("%s-master-0", alluxioruntime.Name)].Spec.Ports = []corev1.ServicePort{
					{
						Name: "rpc",
						Port: int32(portNum),
					},
				}
			})
			It("should return the correct HCFSStatus", func() {
				patch := gomonkey.ApplyPrivateMethod(engine, "queryCompatibleUFSVersion", func() (string, error) {
					compatibleVersion := "2.7.0"
					return compatibleVersion, nil
				})
				defer patch.Reset()
				hcfsStatus, err := engine.GetHCFSStatus()
				Expect(err).To(BeNil())

				Expect(hcfsStatus.Endpoint).To(Equal(fmt.Sprintf("alluxio://%s-master-0.%s:%d", alluxioruntime.Name, alluxioruntime.Namespace, portNum)))
				Expect(hcfsStatus.UnderlayerFileSystemVersion).To(Equal("2.7.0"))
			})
		})
	})

	Describe("Test AlluxioEngine.queryHCFSEndpoint()", func() {
		When("given AlluxioEngine works as expected", func() {
			portNum := 19999
			BeforeEach(func() {
				mockedObjects.Services[fmt.Sprintf("%s-master-0", alluxioruntime.Name)].Spec.Ports = []corev1.ServicePort{
					{
						Name: "rpc",
						Port: int32(portNum),
					},
				}
			})

			It("should return the correct endpoint", func() {
				endpoint, err := engine.queryHCFSEndpoint()
				Expect(err).To(BeNil())
				Expect(endpoint).To(Equal(fmt.Sprintf("alluxio://%s-master-0.%s:%d", alluxioruntime.Name, alluxioruntime.Namespace, portNum)))
			})
		})

		When("service does not have a rpc port", func() {
			It("should return an empty endpoint", func() {
				endpoint, err := engine.queryHCFSEndpoint()
				Expect(err).To(BeNil())
				Expect(endpoint).To(Equal(""))
			})
		})

		When("service does not exist", func() {
			BeforeEach(func() {
				resources = []runtime.Object{
					dataset,
					alluxioruntime,
					mockedObjects.MasterSts,
					mockedObjects.WorkerSts,
					mockedObjects.FuseDs,
					// mockedObjects.Services[fmt.Sprintf("%s-master-0", alluxioruntime.Name)],
				}
			})
			It("should ignore the error and return empty endpoint", func() {
				endpoint, err := engine.queryHCFSEndpoint()
				Expect(err).To(BeNil())
				Expect(endpoint).To(Equal(""))
			})
		})
	})

	Describe("Test AlluxioEngine.queryCompatibleUFSVersion()", func() {
		When("given AlluxioEngine works as expected", func() {
			It("should return the compatible UFS version", func() {
				podName := fmt.Sprintf("%s-master-0", alluxioruntime.Name)
				fileUtils := operations.NewAlluxioFileUtils(podName, "alluxio-master", alluxioruntime.Namespace, engine.Log)
				patch := gomonkey.ApplyMethodFunc(fileUtils, "GetConf", func(string) (string, error) {
					return "2.7.1", nil
				})
				defer patch.Reset()

				version, err := engine.queryCompatibleUFSVersion()
				Expect(err).Should(BeNil())
				Expect(version).Should(Equal("2.7.1"))
			})
		})

		When("fail to getConf from Alluxio master pod", func() {
			It("should return error", func() {
				podName := fmt.Sprintf("%s-master-0", alluxioruntime.Name)
				fileUtils := operations.NewAlluxioFileUtils(podName, "alluxio-master", alluxioruntime.Namespace, engine.Log)
				patch := gomonkey.ApplyMethodFunc(fileUtils, "GetConf", func(string) (string, error) {
					return "", fmt.Errorf("fail to getConf")
				})
				defer patch.Reset()

				version, err := engine.queryCompatibleUFSVersion()
				Expect(err).To(HaveOccurred())
				Expect(version).To(Equal(""))
			})
		})
	})
})
