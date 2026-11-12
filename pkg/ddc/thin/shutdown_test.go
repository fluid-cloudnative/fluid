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

package thin

import (
	"errors"
	"reflect"

	. "github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func buildShutdownRuntime(name, namespace string, placementMode datav1alpha1.PlacementMode, nodeSelector map[string]string) *datav1alpha1.ThinRuntime {
	return &datav1alpha1.ThinRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: datav1alpha1.ThinRuntimeSpec{
			Fuse: datav1alpha1.ThinFuseSpec{NodeSelector: nodeSelector},
		},
		Status: datav1alpha1.RuntimeStatus{},
	}
}

func buildShutdownDataset(name, namespace string, placementMode datav1alpha1.PlacementMode) *datav1alpha1.Dataset {
	return &datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: datav1alpha1.DatasetSpec{
			PlacementMode: placementMode,
		},
	}
}

func buildShutdownEngineWithWorkerTeardownFixture(name, namespace string) *ThinEngine {
	runtimeInfo, err := base.BuildRuntimeInfo(name, namespace, common.ThinRuntime)
	Expect(err).NotTo(HaveOccurred())
	runtimeInfo.SetupWithDataset(buildShutdownDataset(name, namespace, datav1alpha1.ExclusiveMode))

	workerLabelName := runtimeInfo.GetCommonLabelName()
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "shutdown-node",
			Labels: map[string]string{
				runtimeInfo.GetDatasetNumLabelName(): "1",
				workerLabelName:                      "true",
				runtimeInfo.GetRuntimeLabelName():    "true",
			},
		},
	}
	runtimeObj := buildShutdownRuntime(name, namespace, datav1alpha1.ExclusiveMode, nil)
	datasetObj := buildShutdownDataset(name, namespace, datav1alpha1.ExclusiveMode)
	valueConfigMapName := name + "-" + common.ThinEngineImpl + "-values"
	fakeClient := fake.NewFakeClientWithScheme(
		testScheme,
		runtimeObj,
		datasetObj,
		node,
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: name + "-config", Namespace: namespace}},
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: valueConfigMapName, Namespace: namespace}},
	)

	engine := &ThinEngine{
		name:        name,
		namespace:   namespace,
		runtime:     runtimeObj,
		Client:      fakeClient,
		Log:         fake.NullLogger(),
		engineImpl:  common.ThinEngineImpl,
		runtimeInfo: runtimeInfo,
		runtimeType: common.ThinRuntime,
	}
	engine.Helper = ctrl.BuildHelper(runtimeInfo, fakeClient, engine.Log)

	return engine
}

var _ = Describe("ThinEngine shutdown", Label("pkg.ddc.thin.shutdown_test.go"), func() {
	Describe("destroyWorkers", func() {
		It("should tear down worker labels after seeding runtime objects required by runtime info", func() {
			runtimeInfoSpark, err := base.BuildRuntimeInfo("spark", "fluid", common.ThinRuntime)
			Expect(err).NotTo(HaveOccurred())
			runtimeInfoSpark.SetupWithDataset(buildShutdownDataset("spark", "fluid", datav1alpha1.ExclusiveMode))

			nodeSelector := map[string]string{"node-select": "true"}
			runtimeInfoHadoop, err := base.BuildRuntimeInfo("hadoop", "fluid", common.ThinRuntime)
			Expect(err).NotTo(HaveOccurred())
			runtimeInfoHadoop.SetupWithDataset(buildShutdownDataset("hadoop", "fluid", datav1alpha1.ShareMode))
			runtimeInfoHadoop.SetFuseNodeSelector(nodeSelector)

			nodeInputs := []*corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-node-spark",
						Labels: map[string]string{
							"fluid.io/dataset-num":            "1",
							"fluid.io/s-thin-fluid-spark":     "true",
							"fluid.io/s-fluid-spark":          "true",
							"fluid.io/s-h-thin-d-fluid-spark": "5B",
							"fluid.io/s-h-thin-m-fluid-spark": "1B",
							"fluid.io/s-h-thin-t-fluid-spark": "6B",
							"fluid_exclusive":                 "fluid_spark",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-node-share",
						Labels: map[string]string{
							"fluid.io/dataset-num":             "2",
							"fluid.io/s-thin-fluid-hadoop":     "true",
							"fluid.io/s-fluid-hadoop":          "true",
							"fluid.io/s-h-thin-d-fluid-hadoop": "5B",
							"fluid.io/s-h-thin-m-fluid-hadoop": "1B",
							"fluid.io/s-h-thin-t-fluid-hadoop": "6B",
							"fluid.io/s-thin-fluid-hbase":      "true",
							"fluid.io/s-fluid-hbase":           "true",
							"fluid.io/s-h-thin-d-fluid-hbase":  "5B",
							"fluid.io/s-h-thin-m-fluid-hbase":  "1B",
							"fluid.io/s-h-thin-t-fluid-hbase":  "6B",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-node-hadoop",
						Labels: map[string]string{
							"fluid.io/dataset-num":             "1",
							"fluid.io/s-thin-fluid-hadoop":     "true",
							"fluid.io/s-fluid-hadoop":          "true",
							"fluid.io/s-h-thin-d-fluid-hadoop": "5B",
							"fluid.io/s-h-thin-m-fluid-hadoop": "1B",
							"fluid.io/s-h-thin-t-fluid-hadoop": "6B",
							"node-select":                      "true",
						},
					},
				},
			}

			testObjects := []runtime.Object{
				buildShutdownRuntime("spark", "fluid", datav1alpha1.ExclusiveMode, nil),
				buildShutdownDataset("spark", "fluid", datav1alpha1.ExclusiveMode),
				buildShutdownRuntime("hadoop", "fluid", datav1alpha1.ShareMode, nodeSelector),
				buildShutdownDataset("hadoop", "fluid", datav1alpha1.ShareMode),
			}
			for _, nodeInput := range nodeInputs {
				testObjects = append(testObjects, nodeInput.DeepCopy())
			}

			fakeClient := fake.NewFakeClientWithScheme(testScheme, testObjects...)

			tests := []struct {
				runtimeInfo      base.RuntimeInfoInterface
				wantedNodeLabels map[string]map[string]string
			}{
				{
					runtimeInfo: runtimeInfoSpark,
					wantedNodeLabels: map[string]map[string]string{
						"test-node-spark": nil,
						"test-node-share": {
							"fluid.io/dataset-num":             "2",
							"fluid.io/s-thin-fluid-hadoop":     "true",
							"fluid.io/s-fluid-hadoop":          "true",
							"fluid.io/s-h-thin-d-fluid-hadoop": "5B",
							"fluid.io/s-h-thin-m-fluid-hadoop": "1B",
							"fluid.io/s-h-thin-t-fluid-hadoop": "6B",
							"fluid.io/s-thin-fluid-hbase":      "true",
							"fluid.io/s-fluid-hbase":           "true",
							"fluid.io/s-h-thin-d-fluid-hbase":  "5B",
							"fluid.io/s-h-thin-m-fluid-hbase":  "1B",
							"fluid.io/s-h-thin-t-fluid-hbase":  "6B",
						},
						"test-node-hadoop": {
							"fluid.io/dataset-num":             "1",
							"fluid.io/s-thin-fluid-hadoop":     "true",
							"fluid.io/s-fluid-hadoop":          "true",
							"fluid.io/s-h-thin-d-fluid-hadoop": "5B",
							"fluid.io/s-h-thin-m-fluid-hadoop": "1B",
							"fluid.io/s-h-thin-t-fluid-hadoop": "6B",
							"node-select":                      "true",
						},
					},
				},
				{
					runtimeInfo: runtimeInfoHadoop,
					wantedNodeLabels: map[string]map[string]string{
						"test-node-spark": nil,
						"test-node-share": {
							"fluid.io/dataset-num":            "1",
							"fluid.io/s-thin-fluid-hbase":     "true",
							"fluid.io/s-fluid-hbase":          "true",
							"fluid.io/s-h-thin-d-fluid-hbase": "5B",
							"fluid.io/s-h-thin-m-fluid-hbase": "1B",
							"fluid.io/s-h-thin-t-fluid-hbase": "6B",
						},
						"test-node-hadoop": {
							"node-select": "true",
						},
					},
				},
			}

			for _, test := range tests {
				engine := &ThinEngine{
					Log:         fake.NullLogger(),
					Client:      fakeClient,
					Helper:      ctrl.BuildHelper(test.runtimeInfo, fakeClient, fake.NullLogger()),
					runtimeInfo: test.runtimeInfo,
					name:        test.runtimeInfo.GetName(),
					namespace:   test.runtimeInfo.GetNamespace(),
					runtimeType: common.ThinRuntime,
				}

				Expect(engine.destroyWorkers()).To(Succeed())

				for _, node := range nodeInputs {
					newNode, err := kubeclient.GetNode(fakeClient, node.Name)
					Expect(err).NotTo(HaveOccurred())
					Expect(newNode.Labels).To(Equal(test.wantedNodeLabels[node.Name]))
				}
			}
		})
	})

	Describe("destroyMaster", func() {
		It("should delete the release when the helm release exists", func() {
			client := fake.NewFakeClientWithScheme(testScheme)
			engine := &ThinEngine{
				name:      "test",
				namespace: "fluid",
				Log:       fake.NullLogger(),
				Client:    client,
				runtime: &datav1alpha1.ThinRuntime{
					Spec: datav1alpha1.ThinRuntimeSpec{Fuse: datav1alpha1.ThinFuseSpec{}},
				},
			}

			checkReleasePatch := ApplyFunc(helm.CheckRelease, func(name string, namespace string) (bool, error) {
				Expect(name).To(Equal("test"))
				Expect(namespace).To(Equal("fluid"))
				return true, nil
			})
			defer checkReleasePatch.Reset()

			deleteReleasePatch := ApplyFunc(helm.DeleteRelease, func(name string, namespace string) error {
				Expect(name).To(Equal("test"))
				Expect(namespace).To(Equal("fluid"))
				return nil
			})
			defer deleteReleasePatch.Reset()

			Expect(engine.destroyMaster()).To(Succeed())
		})

		It("should clean orphaned resources when the helm release does not exist", func() {
			orphanedCm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "fluid",
					Name:      "test-runtimeset",
				},
			}
			client := fake.NewFakeClientWithScheme(testScheme, orphanedCm)
			engine := &ThinEngine{
				name:      "test",
				namespace: "fluid",
				Log:       fake.NullLogger(),
				Client:    client,
				runtime: &datav1alpha1.ThinRuntime{
					Spec: datav1alpha1.ThinRuntimeSpec{Fuse: datav1alpha1.ThinFuseSpec{}},
				},
			}

			checkReleasePatch := ApplyFunc(helm.CheckRelease, func(name string, namespace string) (bool, error) {
				return false, nil
			})
			defer checkReleasePatch.Reset()

			Expect(engine.destroyMaster()).To(Succeed())

			cm, err := kubeclient.GetConfigmapByName(engine.Client, orphanedCm.Name, orphanedCm.Namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(cm).To(BeNil())
		})

		It("should return an error when checking the helm release fails", func() {
			client := fake.NewFakeClientWithScheme(testScheme)
			engine := &ThinEngine{
				name:      "test",
				namespace: "fluid",
				Log:       fake.NullLogger(),
				Client:    client,
				runtime: &datav1alpha1.ThinRuntime{
					Spec: datav1alpha1.ThinRuntimeSpec{Fuse: datav1alpha1.ThinFuseSpec{}},
				},
			}

			checkReleasePatch := ApplyFunc(helm.CheckRelease, func(name string, namespace string) (bool, error) {
				return false, errors.New("fail to check release")
			})
			defer checkReleasePatch.Reset()

			Expect(engine.destroyMaster()).To(MatchError("fail to check release"))
		})

		It("should return an error when deleting an installed release fails", func() {
			client := fake.NewFakeClientWithScheme(testScheme)
			engine := &ThinEngine{
				name:      "test",
				namespace: "fluid",
				Log:       fake.NullLogger(),
				Client:    client,
				runtime: &datav1alpha1.ThinRuntime{
					Spec: datav1alpha1.ThinRuntimeSpec{Fuse: datav1alpha1.ThinFuseSpec{}},
				},
			}

			checkReleasePatch := ApplyFunc(helm.CheckRelease, func(name string, namespace string) (bool, error) {
				return true, nil
			})
			defer checkReleasePatch.Reset()

			deleteReleasePatch := ApplyFunc(helm.DeleteRelease, func(name string, namespace string) error {
				return errors.New("fail to delete chart")
			})
			defer deleteReleasePatch.Reset()

			Expect(engine.destroyMaster()).To(MatchError("fail to delete chart"))
		})
	})

	Describe("cleanAll", func() {
		It("should clean fuse artifacts and configmaps", func() {
			configMaps := []runtime.Object{
				&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "test-config", Namespace: "fluid"}},
				&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "test-thin-values", Namespace: "fluid"}},
			}
			fakeClient := fake.NewFakeClientWithScheme(testScheme, configMaps...)

			helper := &ctrl.Helper{}
			patches := ApplyMethod(reflect.TypeOf(helper), "CleanUpFuse", func(_ *ctrl.Helper) (int, error) {
				return 0, nil
			})
			defer patches.Reset()

			engine := &ThinEngine{
				name:       "test",
				namespace:  "fluid",
				engineImpl: common.ThinEngineImpl,
				Client:     fakeClient,
				Log:        fake.NullLogger(),
				Helper:     helper,
			}

			Expect(engine.cleanAll()).To(Succeed())

			for _, name := range []string{"test-config", "test-thin-values"} {
				cm, err := kubeclient.GetConfigmapByName(fakeClient, name, "fluid")
				Expect(err).NotTo(HaveOccurred())
				Expect(cm).To(BeNil())
			}
		})
	})

	Describe("Shutdown", func() {
		It("should return the destroyWorkers error after the cache retry limit is reached", func() {
			engine := &ThinEngine{
				name:                   "missing-runtime",
				namespace:              "fluid",
				Log:                    fake.NullLogger(),
				Client:                 fake.NewFakeClientWithScheme(testScheme),
				gracefulShutdownLimits: 1,
				retryShutdown:          1,
				runtimeType:            common.ThinRuntime,
			}

			err := engine.Shutdown()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})

		It("should return the destroyMaster error after workers are torn down successfully", func() {
			engine := buildShutdownEngineWithWorkerTeardownFixture("shutdown-master-error", "fluid")
			engine.gracefulShutdownLimits = 1
			engine.retryShutdown = 1

			checkReleasePatch := ApplyFunc(helm.CheckRelease, func(name string, namespace string) (bool, error) {
				return false, errors.New("check release failed")
			})
			defer checkReleasePatch.Reset()

			Expect(engine.Shutdown()).To(MatchError("check release failed"))
		})

		It("should complete shutdown successfully after the cache retry limit is reached", func() {
			engine := buildShutdownEngineWithWorkerTeardownFixture("shutdown-success", "fluid")
			engine.gracefulShutdownLimits = 1
			engine.retryShutdown = 1

			checkReleasePatch := ApplyFunc(helm.CheckRelease, func(name string, namespace string) (bool, error) {
				return false, nil
			})
			defer checkReleasePatch.Reset()

			cleanUpFusePatch := ApplyMethod(reflect.TypeOf(&ctrl.Helper{}), "CleanUpFuse", func(_ *ctrl.Helper) (int, error) {
				return 0, nil
			})
			defer cleanUpFusePatch.Reset()

			Expect(engine.Shutdown()).To(Succeed())

			for _, cmName := range []string{"shutdown-success-config", "shutdown-success-thin-values"} {
				cm, err := kubeclient.GetConfigmapByName(engine.Client, cmName, "fluid")
				Expect(err).NotTo(HaveOccurred())
				Expect(cm).To(BeNil())
			}
		})
	})
})
