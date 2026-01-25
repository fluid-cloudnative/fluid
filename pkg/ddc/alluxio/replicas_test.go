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
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Constants for test values
const (
	testNamespaceReplicas = "fluid"
	testNodeSpark         = "test-node-spark"
	testNodeShare         = "test-node-share"
	testNodeHadoop        = "test-node-hadoop"
	testRuntimeHbase      = "hbase"
	testRuntimeHadoop     = "hadoop"
	testRuntimeObj        = "obj"
	testWorkerSuffix      = "-worker"
	testFuseSuffix        = "-fuse"
	testEndpoint          = "test Endpoint"
	testHCFSVersion       = "Underlayer HCFS Compatible Version"
)

func newAlluxioEngineREP(client client.Client, name string, namespace string) *AlluxioEngine {
	runTimeInfo, _ := base.BuildRuntimeInfo(name, namespace, "alluxio")
	engine := &AlluxioEngine{
		runtime:     &v1alpha1.AlluxioRuntime{},
		name:        name,
		namespace:   namespace,
		Client:      client,
		runtimeInfo: runTimeInfo,
		Log:         fake.NullLogger(),
	}
	engine.Helper = ctrl.BuildHelper(runTimeInfo, client, engine.Log)
	return engine
}

var _ = Describe("AlluxioEngine Replicas Tests", Label("pkg.ddc.alluxio.replicas_test.go"), func() {
	Describe("SyncReplicas", func() {
		var (
			fakeClient client.Client
			objs       []runtime.Object
		)

		BeforeEach(func() {
			// Setup test nodes
			nodeInputs := []*corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: testNodeSpark,
						Labels: map[string]string{
							"fluid.io/dataset-num":               "1",
							"fluid.io/s-alluxio-fluid-spark":     "true",
							"fluid.io/s-fluid-spark":             "true",
							"fluid.io/s-h-alluxio-d-fluid-spark": "5B",
							"fluid.io/s-h-alluxio-m-fluid-spark": "1B",
							"fluid.io/s-h-alluxio-t-fluid-spark": "6B",
							"fluid_exclusive":                    "fluid_spark",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: testNodeShare,
						Labels: map[string]string{
							"fluid.io/dataset-num":                "2",
							"fluid.io/s-alluxio-fluid-hadoop":     "true",
							"fluid.io/s-fluid-hadoop":             "true",
							"fluid.io/s-h-alluxio-d-fluid-hadoop": "5B",
							"fluid.io/s-h-alluxio-m-fluid-hadoop": "1B",
							"fluid.io/s-h-alluxio-t-fluid-hadoop": "6B",
							"fluid.io/s-alluxio-fluid-hbase":      "true",
							"fluid.io/s-fluid-hbase":              "true",
							"fluid.io/s-h-alluxio-d-fluid-hbase":  "5B",
							"fluid.io/s-h-alluxio-m-fluid-hbase":  "1B",
							"fluid.io/s-h-alluxio-t-fluid-hbase":  "6B",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: testNodeHadoop,
						Labels: map[string]string{
							"fluid.io/dataset-num":                "1",
							"fluid.io/s-alluxio-fluid-hadoop":     "true",
							"fluid.io/s-fluid-hadoop":             "true",
							"fluid.io/s-h-alluxio-d-fluid-hadoop": "5B",
							"fluid.io/s-h-alluxio-m-fluid-hadoop": "1B",
							"fluid.io/s-h-alluxio-t-fluid-hadoop": "6B",
							"node-select":                         "true",
						},
					},
				},
			}

			// Setup test runtimes
			runtimeInputs := []*v1alpha1.AlluxioRuntime{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testRuntimeHbase,
						Namespace: testNamespaceReplicas,
					},
					Spec: v1alpha1.AlluxioRuntimeSpec{
						Replicas: 3,
					},
					Status: v1alpha1.RuntimeStatus{
						DesiredWorkerNumberScheduled: 2,
						Conditions: []v1alpha1.RuntimeCondition{
							utils.NewRuntimeCondition(v1alpha1.RuntimeWorkersInitialized, v1alpha1.RuntimeWorkersInitializedReason, "The workers are initialized.", corev1.ConditionTrue),
							utils.NewRuntimeCondition(v1alpha1.RuntimeFusesInitialized, v1alpha1.RuntimeFusesInitializedReason, "The fuses are initialized.", corev1.ConditionTrue),
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testRuntimeHadoop,
						Namespace: testNamespaceReplicas,
					},
					Spec: v1alpha1.AlluxioRuntimeSpec{
						Replicas: 1,
					},
					Status: v1alpha1.RuntimeStatus{
						DesiredWorkerNumberScheduled: 2,
						Conditions: []v1alpha1.RuntimeCondition{
							utils.NewRuntimeCondition(v1alpha1.RuntimeWorkersInitialized, v1alpha1.RuntimeWorkersInitializedReason, "The workers are initialized.", corev1.ConditionTrue),
							utils.NewRuntimeCondition(v1alpha1.RuntimeFusesInitialized, v1alpha1.RuntimeFusesInitializedReason, "The fuses are initialized.", corev1.ConditionTrue),
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testRuntimeObj,
						Namespace: testNamespaceReplicas,
					},
					Spec: v1alpha1.AlluxioRuntimeSpec{
						Replicas: 2,
					},
					Status: v1alpha1.RuntimeStatus{
						DesiredWorkerNumberScheduled: 2,
					},
				},
			}

			// Setup test workers (StatefulSets)
			workersInputs := []*appsv1.StatefulSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testRuntimeHbase + testWorkerSuffix,
						Namespace: testNamespaceReplicas,
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](2),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testRuntimeHadoop + testWorkerSuffix,
						Namespace: testNamespaceReplicas,
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](2),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testRuntimeObj + testWorkerSuffix,
						Namespace: testNamespaceReplicas,
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](2),
					},
				},
			}

			// Setup test datasets
			dataSetInputs := []*v1alpha1.Dataset{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testRuntimeHbase,
						Namespace: testNamespaceReplicas,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testRuntimeHadoop,
						Namespace: testNamespaceReplicas,
					},
				},
			}

			// Setup test fuse DaemonSets
			fuseInputs := []*appsv1.DaemonSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testRuntimeHbase + testFuseSuffix,
						Namespace: testNamespaceReplicas,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testRuntimeHadoop + testFuseSuffix,
						Namespace: testNamespaceReplicas,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testRuntimeObj + testFuseSuffix,
						Namespace: testNamespaceReplicas,
					},
				},
			}

			// Build objects list
			objs = []runtime.Object{}
			for _, nodeInput := range nodeInputs {
				objs = append(objs, nodeInput.DeepCopy())
			}
			for _, runtimeInput := range runtimeInputs {
				objs = append(objs, runtimeInput.DeepCopy())
			}
			for _, workerInput := range workersInputs {
				objs = append(objs, workerInput.DeepCopy())
			}
			for _, fuseInput := range fuseInputs {
				objs = append(objs, fuseInput.DeepCopy())
			}
			for _, dataSetInput := range dataSetInputs {
				objs = append(objs, dataSetInput.DeepCopy())
			}

			fakeClient = fake.NewFakeClientWithScheme(testScheme, objs...)
		})

		Context("when scaling out workers", func() {
			It("should add RuntimeWorkerScaledOut condition", func() {
				engine := newAlluxioEngineREP(fakeClient, testRuntimeHbase, testNamespaceReplicas)
				err := engine.SyncReplicas(cruntime.ReconcileRequestContext{
					Log:      fake.NullLogger(),
					Recorder: record.NewFakeRecorder(300),
				})
				Expect(err).NotTo(HaveOccurred())

				rt, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())

				found := false
				for _, cond := range rt.Status.Conditions {
					if cond.Type == v1alpha1.RuntimeWorkerScaledOut {
						found = true
						break
					}
				}
				Expect(found).To(BeTrue(), "expected RuntimeWorkerScaledOut condition to be present")
				Expect(len(rt.Status.Conditions)).To(Equal(3))
			})
		})

		Context("when scaling in workers", func() {
			It("should add RuntimeWorkerScaledIn condition", func() {
				engine := newAlluxioEngineREP(fakeClient, testRuntimeHadoop, testNamespaceReplicas)
				err := engine.SyncReplicas(cruntime.ReconcileRequestContext{
					Log:      fake.NullLogger(),
					Recorder: record.NewFakeRecorder(300),
				})
				Expect(err).NotTo(HaveOccurred())

				rt, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())

				found := false
				for _, cond := range rt.Status.Conditions {
					if cond.Type == v1alpha1.RuntimeWorkerScaledIn {
						found = true
						break
					}
				}
				Expect(found).To(BeTrue(), "expected RuntimeWorkerScaledIn condition to be present")
				Expect(len(rt.Status.Conditions)).To(Equal(3))
			})
		})

		Context("when no scaling is needed", func() {
			It("should not add any scaling condition", func() {
				engine := newAlluxioEngineREP(fakeClient, testRuntimeObj, testNamespaceReplicas)
				err := engine.SyncReplicas(cruntime.ReconcileRequestContext{
					Log:      fake.NullLogger(),
					Recorder: record.NewFakeRecorder(300),
				})
				Expect(err).NotTo(HaveOccurred())

				rt, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())

				// Check that no scaling conditions are present
				Expect(rt.Status.Conditions).To(BeEmpty())
			})
		})
	})

	Describe("SyncReplicas without worker StatefulSet", func() {
		var (
			fakeClient client.Client
			engine     *AlluxioEngine
		)

		BeforeEach(func() {
			// Setup test objects without worker StatefulSet
			testObjs := []runtime.Object{}

			// DaemonSet for fuse
			daemonSetInputs := []appsv1.DaemonSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testRuntimeHbase + testFuseSuffix,
						Namespace: testNamespaceReplicas,
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 1,
						NumberReady:       1,
						NumberAvailable:   1,
					},
				},
			}
			for _, daemonSet := range daemonSetInputs {
				testObjs = append(testObjs, daemonSet.DeepCopy())
			}

			// AlluxioRuntime
			alluxioruntimeInputs := []v1alpha1.AlluxioRuntime{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testRuntimeHbase,
						Namespace: testNamespaceReplicas,
					},
					Status: v1alpha1.RuntimeStatus{
						MasterPhase: v1alpha1.RuntimePhaseReady,
						WorkerPhase: v1alpha1.RuntimePhaseReady,
						FusePhase:   v1alpha1.RuntimePhaseReady,
					},
				},
			}
			for _, alluxioruntimeInput := range alluxioruntimeInputs {
				testObjs = append(testObjs, alluxioruntimeInput.DeepCopy())
			}

			// Dataset
			datasetInputs := []*v1alpha1.Dataset{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testRuntimeHbase,
						Namespace: testNamespaceReplicas,
					},
					Spec: v1alpha1.DatasetSpec{},
					Status: v1alpha1.DatasetStatus{
						Phase: v1alpha1.BoundDatasetPhase,
						HCFSStatus: &v1alpha1.HCFSStatus{
							Endpoint:                    testEndpoint,
							UnderlayerFileSystemVersion: testHCFSVersion,
						},
					},
				},
			}
			for _, dataset := range datasetInputs {
				testObjs = append(testObjs, dataset.DeepCopy())
			}

			fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)

			engine = &AlluxioEngine{
				Client:    fakeClient,
				Log:       fake.NullLogger(),
				namespace: testNamespaceReplicas,
				name:      testRuntimeHbase,
				runtime: &v1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testRuntimeHbase,
						Namespace: testNamespaceReplicas,
					},
				},
			}
		})

		Context("when worker StatefulSet is missing", func() {
			It("should update runtime and dataset status to reflect failure", func() {
				err := engine.SyncReplicas(cruntime.ReconcileRequestContext{
					Log:      fake.NullLogger(),
					Recorder: record.NewFakeRecorder(300),
				})
				// Error is expected when worker is not found
				Expect(err).To(HaveOccurred())

				// Verify runtime status
				alluxioruntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(alluxioruntime.Status.MasterPhase).To(Equal(v1alpha1.RuntimePhaseReady))
				Expect(alluxioruntime.Status.WorkerPhase).To(Equal(v1alpha1.RuntimePhaseNotReady))
				Expect(alluxioruntime.Status.FusePhase).To(Equal(v1alpha1.RuntimePhaseReady))

				// Verify dataset status
				var dataset v1alpha1.Dataset
				key := types.NamespacedName{
					Name:      alluxioruntime.Name,
					Namespace: alluxioruntime.Namespace,
				}
				err = fakeClient.Get(context.TODO(), key, &dataset)
				Expect(err).NotTo(HaveOccurred())
				Expect(dataset.Status.Phase).To(Equal(v1alpha1.FailedDatasetPhase))
			})
		})
	})
})
