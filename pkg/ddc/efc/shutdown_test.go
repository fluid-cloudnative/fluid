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

package efc

import (
	"context"
	"reflect"

	"github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/net"
)

func newShutdownRuntime(name string, placement datav1alpha1.PlacementMode, nodeSelector map[string]string) *datav1alpha1.EFCRuntime {
	return &datav1alpha1.EFCRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "fluid",
		},
		Spec: datav1alpha1.EFCRuntimeSpec{
			Fuse: datav1alpha1.EFCFuseSpec{
				NodeSelector: nodeSelector,
			},
		},
	}
}

func newShutdownRuntimeInfo(name string, placement datav1alpha1.PlacementMode, nodeSelector map[string]string) base.RuntimeInfoInterface {
	runtimeInfo, err := base.BuildRuntimeInfo(name, "fluid", common.EFCRuntime)
	Expect(err).NotTo(HaveOccurred())
	runtimeInfo.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: placement},
	})
	if len(nodeSelector) > 0 {
		runtimeInfo.SetFuseNodeSelector(nodeSelector)
	}
	return runtimeInfo
}

func newShutdownNode(name string, labels map[string]string) *corev1.Node {
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
}

var _ = Describe("EFCEngine shutdown", func() {
	Describe("destroyWorkers", func() {
		It("tears down exclusive runtime worker labels after seeding the runtime object", func() {
			runtimeInfo := newShutdownRuntimeInfo("spark", datav1alpha1.ExclusiveMode, nil)
			runtimeObj := newShutdownRuntime("spark", datav1alpha1.ExclusiveMode, nil)
			nodes := []*corev1.Node{
				newShutdownNode("test-node-spark", map[string]string{
					"fluid.io/dataset-num":           "1",
					"fluid.io/s-efc-fluid-spark":     "true",
					"fluid.io/s-fluid-spark":         "true",
					"fluid.io/s-h-efc-d-fluid-spark": "5B",
					"fluid.io/s-h-efc-m-fluid-spark": "1B",
					"fluid.io/s-h-efc-t-fluid-spark": "6B",
					"fluid_exclusive":                "fluid_spark",
				}),
				newShutdownNode("test-node-share", map[string]string{
					"fluid.io/dataset-num":            "2",
					"fluid.io/s-efc-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":         "true",
					"fluid.io/s-h-efc-d-fluid-hadoop": "5B",
					"fluid.io/s-h-efc-m-fluid-hadoop": "1B",
					"fluid.io/s-h-efc-t-fluid-hadoop": "6B",
					"fluid.io/s-efc-fluid-hbase":      "true",
					"fluid.io/s-fluid-hbase":          "true",
					"fluid.io/s-h-efc-d-fluid-hbase":  "5B",
					"fluid.io/s-h-efc-m-fluid-hbase":  "1B",
					"fluid.io/s-h-efc-t-fluid-hbase":  "6B",
				}),
			}

			objects := []runtime.Object{runtimeObj}
			for _, node := range nodes {
				objects = append(objects, node.DeepCopy())
			}

			client := fake.NewFakeClientWithScheme(testScheme, objects...)
			engine := &EFCEngine{
				name:        runtimeInfo.GetName(),
				namespace:   runtimeInfo.GetNamespace(),
				runtimeType: common.EFCRuntime,
				runtimeInfo: runtimeInfo,
				Client:      client,
				Log:         fake.NullLogger(),
			}
			engine.Helper = ctrl.BuildHelper(runtimeInfo, client, engine.Log)

			Expect(engine.destroyWorkers()).To(Succeed())

			node := &corev1.Node{}
			Expect(client.Get(context.TODO(), types.NamespacedName{Name: "test-node-spark"}, node)).To(Succeed())
			Expect(node.Labels).To(BeEmpty())

			Expect(client.Get(context.TODO(), types.NamespacedName{Name: "test-node-share"}, node)).To(Succeed())
			Expect(node.Labels).To(Equal(map[string]string{
				"fluid.io/dataset-num":            "2",
				"fluid.io/s-efc-fluid-hadoop":     "true",
				"fluid.io/s-fluid-hadoop":         "true",
				"fluid.io/s-h-efc-d-fluid-hadoop": "5B",
				"fluid.io/s-h-efc-m-fluid-hadoop": "1B",
				"fluid.io/s-h-efc-t-fluid-hadoop": "6B",
				"fluid.io/s-efc-fluid-hbase":      "true",
				"fluid.io/s-fluid-hbase":          "true",
				"fluid.io/s-h-efc-d-fluid-hbase":  "5B",
				"fluid.io/s-h-efc-m-fluid-hbase":  "1B",
				"fluid.io/s-h-efc-t-fluid-hbase":  "6B",
			}))
		})

		It("tears down shared runtime worker labels after seeding the runtime object", func() {
			nodeSelector := map[string]string{"node-select": "true"}
			runtimeInfo := newShutdownRuntimeInfo("hadoop", datav1alpha1.ShareMode, nodeSelector)
			runtimeObj := newShutdownRuntime("hadoop", datav1alpha1.ShareMode, nodeSelector)
			nodes := []*corev1.Node{
				newShutdownNode("test-node-share", map[string]string{
					"fluid.io/dataset-num":            "2",
					"fluid.io/s-efc-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":         "true",
					"fluid.io/s-h-efc-d-fluid-hadoop": "5B",
					"fluid.io/s-h-efc-m-fluid-hadoop": "1B",
					"fluid.io/s-h-efc-t-fluid-hadoop": "6B",
					"fluid.io/s-efc-fluid-hbase":      "true",
					"fluid.io/s-fluid-hbase":          "true",
					"fluid.io/s-h-efc-d-fluid-hbase":  "5B",
					"fluid.io/s-h-efc-m-fluid-hbase":  "1B",
					"fluid.io/s-h-efc-t-fluid-hbase":  "6B",
				}),
				newShutdownNode("test-node-hadoop", map[string]string{
					"fluid.io/dataset-num":            "1",
					"fluid.io/s-efc-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":         "true",
					"fluid.io/s-h-efc-d-fluid-hadoop": "5B",
					"fluid.io/s-h-efc-m-fluid-hadoop": "1B",
					"fluid.io/s-h-efc-t-fluid-hadoop": "6B",
					"node-select":                     "true",
				}),
			}

			objects := []runtime.Object{runtimeObj}
			for _, node := range nodes {
				objects = append(objects, node.DeepCopy())
			}

			client := fake.NewFakeClientWithScheme(testScheme, objects...)
			engine := &EFCEngine{
				name:        runtimeInfo.GetName(),
				namespace:   runtimeInfo.GetNamespace(),
				runtimeType: common.EFCRuntime,
				runtimeInfo: runtimeInfo,
				Client:      client,
				Log:         fake.NullLogger(),
			}
			engine.Helper = ctrl.BuildHelper(runtimeInfo, client, engine.Log)

			Expect(engine.destroyWorkers()).To(Succeed())

			node := &corev1.Node{}
			Expect(client.Get(context.TODO(), types.NamespacedName{Name: "test-node-share"}, node)).To(Succeed())
			Expect(node.Labels).To(Equal(map[string]string{
				"fluid.io/dataset-num":           "1",
				"fluid.io/s-efc-fluid-hbase":     "true",
				"fluid.io/s-fluid-hbase":         "true",
				"fluid.io/s-h-efc-d-fluid-hbase": "5B",
				"fluid.io/s-h-efc-m-fluid-hbase": "1B",
				"fluid.io/s-h-efc-t-fluid-hbase": "6B",
			}))

			Expect(client.Get(context.TODO(), types.NamespacedName{Name: "test-node-hadoop"}, node)).To(Succeed())
			Expect(node.Labels).To(Equal(map[string]string{
				"node-select": "true",
			}))
		})
	})

	Describe("cleanAll", func() {
		It("cleans fuse resources and removes the values configmap", func() {
			configMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "spark-efc-values",
					Namespace: "fluid",
				},
				Data: map[string]string{"data": valuesConfigMapData},
			}
			client := fake.NewFakeClientWithScheme(testScheme, configMap.DeepCopy())
			helper := &ctrl.Helper{}
			cleaned := false
			patches := gomonkey.ApplyMethod(reflect.TypeOf(helper), "CleanUpFuse", func(_ *ctrl.Helper) (int, error) {
				cleaned = true
				return 1, nil
			})
			defer patches.Reset()

			engine := &EFCEngine{
				name:       "spark",
				namespace:  "fluid",
				engineImpl: common.EFCEngineImpl,
				Client:     client,
				Helper:     helper,
				Log:        fake.NullLogger(),
			}

			Expect(engine.cleanAll()).To(Succeed())
			Expect(cleaned).To(BeTrue())

			deletedConfigMap := &corev1.ConfigMap{}
			err := client.Get(context.TODO(), types.NamespacedName{Name: "spark-efc-values", Namespace: "fluid"}, deletedConfigMap)
			Expect(err).To(HaveOccurred())
			Expect(k8serrors.IsNotFound(err)).To(BeTrue())
		})
	})

	Describe("releasePorts", func() {
		It("releases ports parsed from the values configmap", func() {
			portRange, err := net.ParsePortRange("17673-17674")
			Expect(err).NotTo(HaveOccurred())

			configMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "spark-efc-values",
					Namespace: "fluid",
				},
				Data: map[string]string{"data": valuesConfigMapData},
			}
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "spark",
					Namespace: "fluid",
				},
				Status: datav1alpha1.DatasetStatus{
					Runtimes: []datav1alpha1.Runtime{{
						Name:      "spark",
						Namespace: "fluid",
						Type:      common.EFCRuntime,
					}},
				},
			}
			client := fake.NewFakeClientWithScheme(testScheme, configMap.DeepCopy(), dataset.DeepCopy())
			Expect(portallocator.SetupRuntimePortAllocator(client, portRange, "bitmap", GetReservedPorts)).To(Succeed())

			allocator, err := portallocator.GetRuntimePortAllocator()
			Expect(err).NotTo(HaveOccurred())

			released := []int{}
			patches := gomonkey.ApplyMethod(reflect.TypeOf(allocator), "ReleaseReservedPorts", func(_ *portallocator.RuntimePortAllocator, ports []int) {
				released = append(released, ports...)
			})
			defer patches.Reset()

			engine := &EFCEngine{
				name:       "spark",
				namespace:  "fluid",
				engineImpl: common.EFCEngineImpl,
				Client:     client,
				Log:        fake.NullLogger(),
			}

			Expect(engine.releasePorts()).To(Succeed())
			Expect(released).To(Equal([]int{17673}))
		})

		It("returns without error when the values configmap is absent", func() {
			portRange, err := net.ParsePortRange("26000-32000")
			Expect(err).NotTo(HaveOccurred())

			client := fake.NewFakeClientWithScheme(testScheme)
			Expect(portallocator.SetupRuntimePortAllocator(client, portRange, "bitmap", GetReservedPorts)).To(Succeed())

			engine := &EFCEngine{
				name:       "spark",
				namespace:  "fluid",
				engineImpl: common.EFCEngineImpl,
				Client:     client,
				Log:        fake.NullLogger(),
			}

			Expect(engine.releasePorts()).To(Succeed())
		})
	})

	Describe("destroyMaster", func() {
		It("deletes the helm release when it exists", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			deleted := false
			patches.ApplyFunc(helm.CheckRelease, func(_ string, _ string) (bool, error) {
				return true, nil
			})
			patches.ApplyFunc(helm.DeleteRelease, func(_ string, _ string) error {
				deleted = true
				return nil
			})

			engine := &EFCEngine{name: "spark", namespace: "fluid"}

			Expect(engine.destroyMaster()).To(Succeed())
			Expect(deleted).To(BeTrue())
		})

		It("skips deletion when no helm release exists", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			deleted := false
			patches.ApplyFunc(helm.CheckRelease, func(_ string, _ string) (bool, error) {
				return false, nil
			})
			patches.ApplyFunc(helm.DeleteRelease, func(_ string, _ string) error {
				deleted = true
				return nil
			})

			engine := &EFCEngine{name: "spark", namespace: "fluid"}

			Expect(engine.destroyMaster()).To(Succeed())
			Expect(deleted).To(BeFalse())
		})
	})
})
