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

package jindocache

import (
	"context"
	"reflect"

	. "github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	netutils "k8s.io/apimachinery/pkg/util/net"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	testScheme *runtime.Scheme
)

func init() {
	testScheme = runtime.NewScheme()
	_ = v1.AddToScheme(testScheme)
	_ = datav1alpha1.AddToScheme(testScheme)
	_ = appsv1.AddToScheme(testScheme)
}

var _ = Describe("JindoCacheEngine destroyWorkers", func() {
	buildRuntimeInfo := func(name string) base.RuntimeInfoInterface {
		runtimeInfo, err := base.BuildRuntimeInfo(name, "fluid", common.JindoRuntime)
		Expect(err).NotTo(HaveOccurred())
		runtimeInfo.SetupWithDataset(&datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
		})
		return runtimeInfo
	}

	buildNodes := func() []*v1.Node {
		return []*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node-spark",
					Labels: map[string]string{
						"fluid.io/dataset-num":             "1",
						"fluid.io/s-jindo-fluid-spark":     "true",
						"fluid.io/s-fluid-spark":           "true",
						"fluid.io/s-h-jindo-d-fluid-spark": "5B",
						"fluid.io/s-h-jindo-m-fluid-spark": "1B",
						"fluid.io/s-h-jindo-t-fluid-spark": "6B",
						"fluid_exclusive":                  "fluid_spark",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node-share",
					Labels: map[string]string{
						"fluid.io/dataset-num":              "2",
						"fluid.io/s-jindo-fluid-hadoop":     "true",
						"fluid.io/s-fluid-hadoop":           "true",
						"fluid.io/s-h-jindo-d-fluid-hadoop": "5B",
						"fluid.io/s-h-jindo-m-fluid-hadoop": "1B",
						"fluid.io/s-h-jindo-t-fluid-hadoop": "6B",
						"fluid.io/s-jindo-fluid-hbase":      "true",
						"fluid.io/s-fluid-hbase":            "true",
						"fluid.io/s-h-jindo-d-fluid-hbase":  "5B",
						"fluid.io/s-h-jindo-m-fluid-hbase":  "1B",
						"fluid.io/s-h-jindo-t-fluid-hbase":  "6B",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node-hadoop",
					Labels: map[string]string{
						"fluid.io/dataset-num":              "1",
						"fluid.io/s-jindo-fluid-hadoop":     "true",
						"fluid.io/s-fluid-hadoop":           "true",
						"fluid.io/s-h-jindo-d-fluid-hadoop": "5B",
						"fluid.io/s-h-jindo-m-fluid-hadoop": "1B",
						"fluid.io/s-h-jindo-t-fluid-hadoop": "6B",
						"node-select":                       "true",
					},
				},
			},
		}
	}

	DescribeTable("should tear down worker labels when the runtime exists in the client",
		func(runtimeInfo base.RuntimeInfoInterface, runtimeObject *datav1alpha1.JindoRuntime, wantedNodeLabels map[string]map[string]string) {
			runtimeObjects := []runtime.Object{}
			for _, node := range buildNodes() {
				runtimeObjects = append(runtimeObjects, node.DeepCopy())
			}
			runtimeObjects = append(runtimeObjects, runtimeObject.DeepCopy())

			client := fake.NewFakeClientWithScheme(testScheme, runtimeObjects...)
			engine := &JindoCacheEngine{
				Log:         fake.NullLogger(),
				runtimeInfo: runtimeInfo,
				Client:      client,
				name:        runtimeInfo.GetName(),
				namespace:   runtimeInfo.GetNamespace(),
			}
			engine.Helper = ctrl.BuildHelper(runtimeInfo, client, engine.Log)

			err := engine.destroyWorkers()

			Expect(err).NotTo(HaveOccurred())
			for _, node := range buildNodes() {
				newNode, getErr := kubeclient.GetNode(client, node.Name)
				Expect(getErr).NotTo(HaveOccurred())
				if len(wantedNodeLabels[node.Name]) == 0 {
					Expect(newNode.Labels).To(BeEmpty())
					continue
				}
				Expect(newNode.Labels).To(HaveLen(len(wantedNodeLabels[node.Name])))
				Expect(newNode.Labels).To(Equal(wantedNodeLabels[node.Name]))
			}
		},
		Entry("for spark",
			buildRuntimeInfo("spark"),
			&datav1alpha1.JindoRuntime{ObjectMeta: metav1.ObjectMeta{Name: "spark", Namespace: "fluid"}},
			map[string]map[string]string{
				"test-node-spark": {},
				"test-node-share": {
					"fluid.io/dataset-num":              "2",
					"fluid.io/s-jindo-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":           "true",
					"fluid.io/s-h-jindo-d-fluid-hadoop": "5B",
					"fluid.io/s-h-jindo-m-fluid-hadoop": "1B",
					"fluid.io/s-h-jindo-t-fluid-hadoop": "6B",
					"fluid.io/s-jindo-fluid-hbase":      "true",
					"fluid.io/s-fluid-hbase":            "true",
					"fluid.io/s-h-jindo-d-fluid-hbase":  "5B",
					"fluid.io/s-h-jindo-m-fluid-hbase":  "1B",
					"fluid.io/s-h-jindo-t-fluid-hbase":  "6B",
				},
				"test-node-hadoop": {
					"fluid.io/dataset-num":              "1",
					"fluid.io/s-jindo-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":           "true",
					"fluid.io/s-h-jindo-d-fluid-hadoop": "5B",
					"fluid.io/s-h-jindo-m-fluid-hadoop": "1B",
					"fluid.io/s-h-jindo-t-fluid-hadoop": "6B",
					"node-select":                       "true",
				},
			},
		),
		Entry("for hadoop", func() base.RuntimeInfoInterface {
			runtimeInfo := buildRuntimeInfo("hadoop")
			runtimeInfo.SetFuseNodeSelector(map[string]string{"node-select": "true"})
			return runtimeInfo
		}(),
			&datav1alpha1.JindoRuntime{ObjectMeta: metav1.ObjectMeta{Name: "hadoop", Namespace: "fluid"}},
			map[string]map[string]string{
				"test-node-spark": {
					"fluid.io/dataset-num":             "1",
					"fluid.io/s-jindo-fluid-spark":     "true",
					"fluid.io/s-fluid-spark":           "true",
					"fluid.io/s-h-jindo-d-fluid-spark": "5B",
					"fluid.io/s-h-jindo-m-fluid-spark": "1B",
					"fluid.io/s-h-jindo-t-fluid-spark": "6B",
					"fluid_exclusive":                  "fluid_spark",
				},
				"test-node-share": {
					"fluid.io/dataset-num":             "1",
					"fluid.io/s-jindo-fluid-hbase":     "true",
					"fluid.io/s-fluid-hbase":           "true",
					"fluid.io/s-h-jindo-d-fluid-hbase": "5B",
					"fluid.io/s-h-jindo-m-fluid-hbase": "1B",
					"fluid.io/s-h-jindo-t-fluid-hbase": "6B",
				},
				"test-node-hadoop": {
					"node-select": "true",
				},
			},
		),
	)
})

var _ = Describe("JindoCacheEngine shutdown orchestration", func() {
	It("should stop after clean cache retry fails before destroying workers", func() {
		engine := &JindoCacheEngine{
			Log:                    fake.NullLogger(),
			retryShutdown:          0,
			gracefulShutdownLimits: 1,
		}

		cleanCachePatch := ApplyPrivateMethod(engine, "invokeCleanCache", func() error {
			return context.DeadlineExceeded
		})
		defer cleanCachePatch.Reset()

		destroyWorkersCalled := false
		destroyWorkersPatch := ApplyPrivateMethod(engine, "destroyWorkers", func() error {
			destroyWorkersCalled = true
			return nil
		})
		defer destroyWorkersPatch.Reset()

		err := engine.Shutdown()

		Expect(err).To(MatchError(context.DeadlineExceeded))
		Expect(engine.retryShutdown).To(Equal(int32(1)))
		Expect(destroyWorkersCalled).To(BeFalse())
	})

	It("should stop at the first teardown error after skipping clean cache retries", func() {
		engine := &JindoCacheEngine{
			Log:                    fake.NullLogger(),
			retryShutdown:          1,
			gracefulShutdownLimits: 1,
		}

		destroyWorkersPatch := ApplyPrivateMethod(engine, "destroyWorkers", func() error {
			return context.Canceled
		})
		defer destroyWorkersPatch.Reset()

		releasePortsCalled := false
		releasePortsPatch := ApplyPrivateMethod(engine, "releasePorts", func() error {
			releasePortsCalled = true
			return nil
		})
		defer releasePortsPatch.Reset()

		err := engine.Shutdown()

		Expect(err).To(MatchError(context.Canceled))
		Expect(releasePortsCalled).To(BeFalse())
	})

	It("should release ports from the runtime configmap", func() {
		configMap := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark-jindofs-config",
				Namespace: "fluid",
			},
			Data: map[string]string{
				"jindocache.cfg": cfg,
			},
		}

		fakeClient := fake.NewFakeClientWithScheme(testScheme, configMap)
		portallocator.SetupRuntimePortAllocatorWithType(fakeClient, &netutils.PortRange{Base: 18000, Size: 2}, portallocator.BitMap, func(ctrlclient.Client) ([]int, error) {
			return nil, nil
		})
		allocator, err := portallocator.GetRuntimePortAllocator()
		Expect(err).NotTo(HaveOccurred())
		allocatorPatch := ApplyFunc(portallocator.GetRuntimePortAllocator, func() (*portallocator.RuntimePortAllocator, error) {
			return allocator, nil
		})
		defer allocatorPatch.Reset()

		engine := &JindoCacheEngine{
			Log:       fake.NullLogger(),
			Client:    fakeClient,
			name:      "spark",
			namespace: "fluid",
		}

		allocated, err := allocator.GetAvailablePorts(2)
		Expect(err).NotTo(HaveOccurred())
		Expect(allocated).To(ConsistOf(18000, 18001))

		Expect(engine.releasePorts()).To(Succeed())
		reallocated, err := allocator.GetAvailablePorts(2)
		Expect(err).NotTo(HaveOccurred())
		Expect(reallocated).To(ConsistOf(18000, 18001))
	})

	It("should delete helm-managed configmaps during cleanAll", func() {
		configMaps := []runtime.Object{
			&v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "spark-jindo-values", Namespace: "fluid"}},
			&v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "spark-jindofs-config", Namespace: "fluid"}},
			&v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "spark-jindofs-client-config", Namespace: "fluid"}},
		}
		fakeClient := fake.NewFakeClientWithScheme(testScheme, configMaps...)
		engine := &JindoCacheEngine{
			Log:        fake.NullLogger(),
			Client:     fakeClient,
			name:       "spark",
			namespace:  "fluid",
			engineImpl: common.JindoRuntime,
			Helper:     ctrl.BuildHelper(buildRuntimeInfoForCleanAll("spark"), fakeClient, fake.NullLogger()),
		}

		Expect(engine.cleanAll()).To(Succeed())

		for _, name := range []string{"spark-jindo-values", "spark-jindofs-config", "spark-jindofs-client-config"} {
			configMap := &v1.ConfigMap{}
			err := fakeClient.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: "fluid"}, configMap)
			Expect(err).To(HaveOccurred())
			Expect(utils.IgnoreNotFound(err)).To(BeNil())
		}
	})

	It("should clean configmaps without error whether they exist or not", func() {
		namespace := "default"
		configMapInputs := []*v1.ConfigMap{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "hbase-alluxio-values", Namespace: namespace},
				Data: map[string]string{
					"data": "image: fluid\nimageTag: 0.6.0",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "hbase-alluxio-config", Namespace: namespace},
				Data:       map[string]string{},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "spark-alluxio-values", Namespace: namespace},
				Data: map[string]string{
					"test-data": "image: fluid\n imageTag: 0.6.0",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "hadoop-alluxio-config", Namespace: namespace},
			},
		}

		testConfigMaps := []runtime.Object{}
		for _, cm := range configMapInputs {
			testConfigMaps = append(testConfigMaps, cm.DeepCopy())
		}

		client := fake.NewFakeClientWithScheme(testScheme, testConfigMaps...)
		for _, tc := range []struct {
			name string
		}{
			{name: "notExist"},
			{name: "test1"},
		} {
			engine := &JindoCacheEngine{
				Log:        fake.NullLogger(),
				name:       tc.name,
				namespace:  namespace,
				engineImpl: "jindo",
				Client:     client,
			}

			Expect(engine.cleanConfigMap()).To(Succeed())
		}
	})

	It("should clean all without error for nodes with and without fuse labels", func() {
		nodeInputs := []*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "no-fuse",
					Labels: map[string]string{},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "fuse",
					Labels: map[string]string{
						"fluid.io/f-jindo-fluid-hadoop":    "true",
						"node-select":                      "true",
						"fluid.io/f-jindo-fluid-hbase":     "true",
						"fluid.io/s-fluid-hbase":           "true",
						"fluid.io/s-h-jindo-d-fluid-hbase": "5B",
						"fluid.io/s-h-jindo-m-fluid-hbase": "1B",
						"fluid.io/s-h-jindo-t-fluid-hbase": "6B",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "multiple-fuse",
					Labels: map[string]string{
						"fluid.io/dataset-num":            "1",
						"fluid.io/f-jindo-fluid-hadoop":   "true",
						"fluid.io/f-jindo-fluid-hadoop-1": "true",
						"node-select":                     "true",
					},
				},
			},
		}

		testNodes := []runtime.Object{}
		for _, nodeInput := range nodeInputs {
			testNodes = append(testNodes, nodeInput.DeepCopy())
		}

		client := fake.NewFakeClientWithScheme(testScheme, testNodes...)
		helper := &ctrl.Helper{}
		patch := ApplyMethod(reflect.TypeOf(helper), "CleanUpFuse", func(_ *ctrl.Helper) (int, error) {
			return 0, nil
		})
		defer patch.Reset()

		engine := &JindoCacheEngine{
			Log:       fake.NullLogger(),
			Client:    client,
			name:      "fluid-hadoop",
			namespace: "default",
		}

		Expect(engine.cleanAll()).To(Succeed())
	})
})

func buildRuntimeInfoForCleanAll(name string) base.RuntimeInfoInterface {
	runtimeInfo, err := base.BuildRuntimeInfo(name, "fluid", common.JindoRuntime)
	if err != nil {
		panic(err)
	}
	runtimeInfo.SetupWithDataset(&datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "fluid"},
		Spec: datav1alpha1.DatasetSpec{
			PlacementMode: datav1alpha1.ExclusiveMode,
		},
	})
	return runtimeInfo
}
