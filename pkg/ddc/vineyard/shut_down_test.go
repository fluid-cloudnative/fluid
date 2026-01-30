/*
Copyright 2024 The Fluid Authors.
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

package vineyard

import (
	"reflect"

	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/net"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

const mockConfigMapData = `
master:
  hostNetwork: true
  ports:
    client: 14001
    peer: 14002
worker:
  hostNetwork: true
  ports:
    rpc: 14003
    exporter: 14004`

var testScheme *runtime.Scheme

func init() {
	testScheme = runtime.NewScheme()
	_ = corev1.AddToScheme(testScheme)
	_ = datav1alpha1.AddToScheme(testScheme)
	_ = appsv1.AddToScheme(testScheme)
}

var _ = Describe("VineyardEngine Shutdown Tests", func() {
	Describe("destroyWorkers", func() {
		const dummyDatasetUID = "dummy-dataset-uid"

		var (
			runtimeInfoSpark  base.RuntimeInfoInterface
			runtimeInfoHadoop base.RuntimeInfoInterface
			nodeInputs        []*corev1.Node
			fakeClient        client.Client
		)

		BeforeEach(func() {
			var err error
			runtimeInfoSpark, err = base.BuildRuntimeInfo("spark", "fluid", common.VineyardRuntime)
			Expect(err).NotTo(HaveOccurred())
			runtimeInfoSpark.SetupWithDataset(&datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
			})
			runtimeInfoSpark.SetOwnerDatasetUID(dummyDatasetUID)

			runtimeInfoHadoop, err = base.BuildRuntimeInfo("hadoop", "fluid", common.VineyardRuntime)
			Expect(err).NotTo(HaveOccurred())
			runtimeInfoHadoop.SetupWithDataset(&datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ShareMode},
			})
			runtimeInfoHadoop.SetFuseNodeSelector(map[string]string{"node-select": "true"})
			runtimeInfoHadoop.SetOwnerDatasetUID(dummyDatasetUID)

			nodeInputs = []*corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-node-spark",
						Labels: map[string]string{
							"fluid.io/dataset-num":                "1",
							"fluid.io/s-vineyard-fluid-spark":     "true",
							"fluid.io/s-fluid-spark":              "true",
							"fluid.io/s-h-vineyard-d-fluid-spark": "5B",
							"fluid.io/s-h-vineyard-m-fluid-spark": "1B",
							"fluid.io/s-h-vineyard-t-fluid-spark": "6B",
							"fluid_exclusive":                     "fluid_spark",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-node-share",
						Labels: map[string]string{
							"fluid.io/dataset-num":                 "2",
							"fluid.io/s-vineyard-fluid-hadoop":     "true",
							"fluid.io/s-fluid-hadoop":              "true",
							"fluid.io/s-h-vineyard-d-fluid-hadoop": "5B",
							"fluid.io/s-h-vineyard-m-fluid-hadoop": "1B",
							"fluid.io/s-h-vineyard-t-fluid-hadoop": "6B",
							"fluid.io/s-vineyard-fluid-hbase":      "true",
							"fluid.io/s-fluid-hbase":               "true",
							"fluid.io/s-h-vineyard-d-fluid-hbase":  "5B",
							"fluid.io/s-h-vineyard-m-fluid-hbase":  "1B",
							"fluid.io/s-h-vineyard-t-fluid-hbase":  "6B",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-node-hadoop",
						Labels: map[string]string{
							"fluid.io/dataset-num":                 "1",
							"fluid.io/s-vineyard-fluid-hadoop":     "true",
							"fluid.io/s-fluid-hadoop":              "true",
							"fluid.io/s-h-vineyard-d-fluid-hadoop": "5B",
							"fluid.io/s-h-vineyard-m-fluid-hadoop": "1B",
							"fluid.io/s-h-vineyard-t-fluid-hadoop": "6B",
							"node-select":                          "true",
						},
					},
				},
			}

			testNodes := []runtime.Object{}
			for _, nodeInput := range nodeInputs {
				testNodes = append(testNodes, nodeInput.DeepCopy())
			}
			fakeClient = fake.NewFakeClientWithScheme(testScheme, testNodes...)
		})

		Context("when destroying workers in exclusive mode", func() {
			It("should remove all labels from exclusive node", func() {
				engine := &VineyardEngine{
					Log:         fake.NullLogger(),
					runtimeInfo: runtimeInfoSpark,
					Client:      fakeClient,
					name:        runtimeInfoSpark.GetName(),
					namespace:   runtimeInfoSpark.GetNamespace(),
				}
				engine.Helper = ctrl.BuildHelper(runtimeInfoSpark, fakeClient, engine.Log)

				err := engine.destroyWorkers()
				Expect(err).NotTo(HaveOccurred())

				newNode, err := kubeclient.GetNode(fakeClient, "test-node-spark")
				Expect(err).NotTo(HaveOccurred())
				Expect(newNode.Labels).To(BeEmpty())

				newNode, err = kubeclient.GetNode(fakeClient, "test-node-share")
				Expect(err).NotTo(HaveOccurred())
				Expect(newNode.Labels).To(Equal(map[string]string{
					"fluid.io/dataset-num":                 "2",
					"fluid.io/s-vineyard-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":              "true",
					"fluid.io/s-h-vineyard-d-fluid-hadoop": "5B",
					"fluid.io/s-h-vineyard-m-fluid-hadoop": "1B",
					"fluid.io/s-h-vineyard-t-fluid-hadoop": "6B",
					"fluid.io/s-vineyard-fluid-hbase":      "true",
					"fluid.io/s-fluid-hbase":               "true",
					"fluid.io/s-h-vineyard-d-fluid-hbase":  "5B",
					"fluid.io/s-h-vineyard-m-fluid-hbase":  "1B",
					"fluid.io/s-h-vineyard-t-fluid-hbase":  "6B",
				}))
			})
		})

		Context("when destroying workers in share mode", func() {
			It("should decrement shared dataset labels", func() {
				engine := &VineyardEngine{
					Log:         fake.NullLogger(),
					runtimeInfo: runtimeInfoHadoop,
					Client:      fakeClient,
					name:        runtimeInfoHadoop.GetName(),
					namespace:   runtimeInfoHadoop.GetNamespace(),
				}
				engine.Helper = ctrl.BuildHelper(runtimeInfoHadoop, fakeClient, engine.Log)

				err := engine.destroyWorkers()
				Expect(err).NotTo(HaveOccurred())

				newNode, err := kubeclient.GetNode(fakeClient, "test-node-share")
				Expect(err).NotTo(HaveOccurred())
				Expect(newNode.Labels).To(Equal(map[string]string{
					"fluid.io/dataset-num":                "1",
					"fluid.io/s-vineyard-fluid-hbase":     "true",
					"fluid.io/s-fluid-hbase":              "true",
					"fluid.io/s-h-vineyard-d-fluid-hbase": "5B",
					"fluid.io/s-h-vineyard-m-fluid-hbase": "1B",
					"fluid.io/s-h-vineyard-t-fluid-hbase": "6B",
				}))

				newNode, err = kubeclient.GetNode(fakeClient, "test-node-hadoop")
				Expect(err).NotTo(HaveOccurred())
				Expect(newNode.Labels).To(Equal(map[string]string{
					"node-select": "true",
				}))
			})
		})
	})

	Describe("cleanAll", func() {
		Context("when cleaning all resources", func() {
			It("should clean up successfully", func() {
				cm := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-vineyard-values",
						Namespace: "fluid",
					},
					Data: map[string]string{"data": mockConfigMapData},
				}
				testObjs := []runtime.Object{cm.DeepCopy()}
				fakeClient := fake.NewFakeClientWithScheme(testScheme, testObjs...)

				helper := &ctrl.Helper{}
				patch := ApplyMethod(reflect.TypeOf(helper), "CleanUpFuse", func(_ *ctrl.Helper) (int, error) {
					return 0, nil
				})
				defer patch.Reset()

				engine := &VineyardEngine{
					name:        "spark",
					namespace:   "fluid",
					runtimeType: "vineyard",
					Client:      fakeClient,
					Log:         fake.NullLogger(),
				}

				err := engine.cleanAll()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("cleanupCache", func() {
		Context("when cleaning up cache", func() {
			It("should return nil", func() {
				engine := &VineyardEngine{
					name:      "spark",
					namespace: "fluid",
					Log:       fake.NullLogger(),
				}

				err := engine.cleanupCache()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("destroyMaster", func() {
		Context("when helm release exists", func() {
			It("should delete release", func() {
				engine := &VineyardEngine{
					name:      "spark",
					namespace: "fluid",
				}

				patch1 := ApplyFunc(helm.CheckRelease,
					func(_ string, _ string) (bool, error) {
						return true, nil
					})
				defer patch1.Reset()

				patch2 := ApplyFunc(helm.DeleteRelease,
					func(_ string, _ string) error {
						return nil
					})
				defer patch2.Reset()

				err := engine.destroyMaster()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("releasePorts", func() {
		Context("when configmap exists with port data", func() {
			It("should release reserved ports", func() {
				cm := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "vineyard-vineyard-values",
						Namespace: "fluid",
					},
					Data: map[string]string{"data": mockConfigMapData},
				}
				testObjs := []runtime.Object{cm.DeepCopy()}
				fakeClient := fake.NewFakeClientWithScheme(testScheme, testObjs...)

				portRange := "14000-16000"
				pr, err := net.ParsePortRange(portRange)
				Expect(err).NotTo(HaveOccurred())

				dummyPorts := func(_ client.Client) ([]int, error) {
					return []int{14000, 14001, 14002, 14003}, nil
				}

				err = portallocator.SetupRuntimePortAllocator(fakeClient, pr, "bitmap", dummyPorts)
				Expect(err).NotTo(HaveOccurred())

				allocator, err := portallocator.GetRuntimePortAllocator()
				Expect(err).NotTo(HaveOccurred())

				patch := ApplyMethod(reflect.TypeOf(allocator), "ReleaseReservedPorts",
					func(_ *portallocator.RuntimePortAllocator, _ []int) {
						// No-op for test mock
					})
				defer patch.Reset()

				engine := &VineyardEngine{
					name:        "vineyard",
					namespace:   "fluid",
					runtimeType: "vineyard",
					Client:      fakeClient,
					Log:         fake.NullLogger(),
				}

				err = engine.releasePorts()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
