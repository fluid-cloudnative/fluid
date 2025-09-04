/*

Copyright 2021 The Fluid Author.

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
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AlluxioEngine Health Check Tests", Label("pkg.ddc.alluxio.health_check_test.go"), func() {
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

	Describe("Test AlluxioEngine.CheckRuntimeHealthy()", func() {
		JustBeforeEach(func() {
			engine.Helper = ctrl.BuildHelper(engine.runtimeInfo, client, fake.NullLogger())
		})

		// TODO: AlluxioEngine.UpdateDatasetStatus() now relies on getHCFSStatus().
		// Remove this gomonkey patch after refactoring AlluxioEngine.UpdateDatasetStatus()
		var patch *gomonkey.Patches
		JustBeforeEach(func() {
			patch = gomonkey.ApplyMethodFunc(engine, "UpdateDatasetStatus", func(phase datav1alpha1.DatasetPhase) error {
				gotDataset := &datav1alpha1.Dataset{}
				err := client.Get(context.TODO(), types.NamespacedName{Namespace: dataset.Namespace, Name: dataset.Name}, gotDataset)
				if err != nil {
					return err
				}

				gotDataset.Status.Phase = phase
				return client.Status().Update(context.TODO(), gotDataset)
			})
		})
		JustAfterEach(func() {
			if patch != nil {
				patch.Reset()
			}
		})
		When("all components are healthy", func() {
			BeforeEach(func() {
				// Make sure all components are ready
				mockedObjects.MasterSts.Spec.Replicas = ptr.To[int32](1)
				mockedObjects.MasterSts.Status.Replicas = 1
				mockedObjects.MasterSts.Status.ReadyReplicas = 1
				mockedObjects.MasterSts.Status.AvailableReplicas = 1

				mockedObjects.WorkerSts.Spec.Replicas = ptr.To[int32](3)
				mockedObjects.WorkerSts.Status.Replicas = 3
				mockedObjects.WorkerSts.Status.ReadyReplicas = 3
				mockedObjects.WorkerSts.Status.AvailableReplicas = 3

				mockedObjects.FuseDs.Status.NumberReady = 1
				mockedObjects.FuseDs.Status.NumberAvailable = 1
				mockedObjects.FuseDs.Status.NumberUnavailable = 0
			})

			It("Should update runtime and dataset status to healthy", func() {
				err := engine.CheckRuntimeHealthy()
				Expect(err).To(BeNil())

				// Check runtime status
				gotRuntime := &datav1alpha1.AlluxioRuntime{}
				err = client.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: alluxioruntime.Name}, gotRuntime)
				Expect(err).To(BeNil())
				Expect(gotRuntime.Status.MasterPhase).To(Equal(datav1alpha1.RuntimePhaseReady))
				Expect(gotRuntime.Status.WorkerPhase).To(Equal(datav1alpha1.RuntimePhaseReady))
				Expect(gotRuntime.Status.FusePhase).To(Equal(datav1alpha1.RuntimePhaseReady))

				// Check dataset status
				gotDataset := &datav1alpha1.Dataset{}
				err = client.Get(context.TODO(), types.NamespacedName{Namespace: dataset.Namespace, Name: dataset.Name}, gotDataset)
				Expect(err).To(BeNil())
				Expect(gotDataset.Status.Phase).To(Equal(datav1alpha1.BoundDatasetPhase))
			})
		})

		When("master is not ready", func() {
			BeforeEach(func() {
				// Make master not ready
				mockedObjects.MasterSts.Spec.Replicas = ptr.To[int32](1)
				mockedObjects.MasterSts.Status.Replicas = 1
				mockedObjects.MasterSts.Status.ReadyReplicas = 0

				dataset.Status.Phase = datav1alpha1.BoundDatasetPhase
			})
			It("Should return error but dataset status will not changed", func() {
				err := engine.CheckRuntimeHealthy()
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("master"))
				Expect(err.Error()).To(ContainSubstring("not healthy"))

				// Check dataset status
				gotDataset := &datav1alpha1.Dataset{}
				err = client.Get(context.TODO(), types.NamespacedName{Namespace: dataset.Namespace, Name: dataset.Name}, gotDataset)
				Expect(err).To(BeNil())
				Expect(gotDataset.Status.Phase).To(Equal(datav1alpha1.BoundDatasetPhase))
			})
		})

		When("worker is not ready", func() {
			BeforeEach(func() {
				// Make sure master is ready but worker is not
				mockedObjects.MasterSts.Spec.Replicas = ptr.To[int32](1)
				mockedObjects.MasterSts.Status.Replicas = 1
				mockedObjects.MasterSts.Status.ReadyReplicas = 1

				mockedObjects.WorkerSts.Spec.Replicas = ptr.To[int32](1)
				mockedObjects.WorkerSts.Status.Replicas = 1
				mockedObjects.WorkerSts.Status.ReadyReplicas = 0

				dataset.Status.Phase = datav1alpha1.BoundDatasetPhase
			})
			It("Should return error but dataset status will not changed", func() {
				err := engine.CheckRuntimeHealthy()
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("worker"))
				Expect(err.Error()).To(ContainSubstring("not healthy"))

				// Check dataset status
				gotDataset := &datav1alpha1.Dataset{}
				err = client.Get(context.TODO(), types.NamespacedName{Namespace: dataset.Namespace, Name: dataset.Name}, gotDataset)
				Expect(err).To(BeNil())
				Expect(gotDataset.Status.Phase).To(Equal(datav1alpha1.BoundDatasetPhase))
			})
		})

		When("master sts does not exist", func() {
			BeforeEach(func() {
				resources = []runtime.Object{
					dataset,
					alluxioruntime,
					// mockedObjects.MasterSts,
					mockedObjects.WorkerSts,
					mockedObjects.FuseDs,
					mockedObjects.PersistentVolumeClaim,
					mockedObjects.PersistentVolume,
				}
			})

			It("should return error and update dataset status to Failed", func() {
				err := engine.CheckRuntimeHealthy()
				Expect(err).NotTo(BeNil())
				Expect(apierrs.IsNotFound(err)).To(BeTrue())

				// Check dataset status
				gotDataset := &datav1alpha1.Dataset{}
				err = client.Get(context.TODO(), types.NamespacedName{Namespace: dataset.Namespace, Name: dataset.Name}, gotDataset)
				Expect(err).To(BeNil())
				Expect(gotDataset.Status.Phase).To(Equal(datav1alpha1.FailedDatasetPhase))
			})
		})

		When("worker sts does not exist", func() {
			BeforeEach(func() {
				// Make sure master is ready
				mockedObjects.MasterSts.Spec.Replicas = ptr.To[int32](1)
				mockedObjects.MasterSts.Status.Replicas = 1
				mockedObjects.MasterSts.Status.ReadyReplicas = 1

				resources = []runtime.Object{
					dataset,
					alluxioruntime,
					mockedObjects.MasterSts,
					// mockedObjects.WorkerSts,
					mockedObjects.FuseDs,
					mockedObjects.PersistentVolumeClaim,
					mockedObjects.PersistentVolume,
				}
			})

			It("should return error and update dataset status to Failed", func() {
				err := engine.CheckRuntimeHealthy()
				Expect(err).NotTo(BeNil())
				Expect(apierrs.IsNotFound(err)).To(BeTrue())

				// Check dataset status
				gotDataset := &datav1alpha1.Dataset{}
				err = client.Get(context.TODO(), types.NamespacedName{Namespace: dataset.Namespace, Name: dataset.Name}, gotDataset)
				Expect(err).To(BeNil())
				Expect(gotDataset.Status.Phase).To(Equal(datav1alpha1.FailedDatasetPhase))
			})
		})

		When("fuse ds does not exist", func() {
			BeforeEach(func() {
				// Make sure master and worker are ready
				mockedObjects.MasterSts.Spec.Replicas = ptr.To[int32](1)
				mockedObjects.MasterSts.Status.Replicas = 1
				mockedObjects.MasterSts.Status.ReadyReplicas = 1

				mockedObjects.WorkerSts.Spec.Replicas = ptr.To[int32](3)
				mockedObjects.WorkerSts.Status.Replicas = 3
				mockedObjects.WorkerSts.Status.ReadyReplicas = 3

				resources = []runtime.Object{
					dataset,
					alluxioruntime,
					mockedObjects.MasterSts,
					mockedObjects.WorkerSts,
					// mockedObjects.FuseDs,
					mockedObjects.PersistentVolumeClaim,
					mockedObjects.PersistentVolume,
				}
			})

			It("should return error and update dataset status to Failed", func() {
				err := engine.CheckRuntimeHealthy()
				Expect(err).NotTo(BeNil())
				Expect(apierrs.IsNotFound(err)).To(BeTrue())

				// Check dataset status
				gotDataset := &datav1alpha1.Dataset{}
				err = client.Get(context.TODO(), types.NamespacedName{Namespace: dataset.Namespace, Name: dataset.Name}, gotDataset)
				Expect(err).To(BeNil())
				Expect(gotDataset.Status.Phase).To(Equal(datav1alpha1.FailedDatasetPhase))
			})
		})
	})
})

func TestCheckFuseHealthy(t *testing.T) {
	var daemonSetInputs = []appsv1.DaemonSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-fuse",
				Namespace: "fluid",
			},
			Status: appsv1.DaemonSetStatus{
				NumberUnavailable: 1,
				NumberReady:       1,
				NumberAvailable:   1,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark-fuse",
				Namespace: "fluid",
			},
			Status: appsv1.DaemonSetStatus{
				NumberUnavailable: 0,
				NumberReady:       1,
				NumberAvailable:   1,
			},
		},
	}

	testObjs := []runtime.Object{}
	for _, daemonSet := range daemonSetInputs {
		testObjs = append(testObjs, daemonSet.DeepCopy())
	}

	var alluxioruntimeInputs = []datav1alpha1.AlluxioRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
		},
	}
	for _, alluxioruntimeInput := range alluxioruntimeInputs {
		testObjs = append(testObjs, alluxioruntimeInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []AlluxioEngine{
		{
			Client:    client,
			Log:       fake.NullLogger(),
			namespace: "fluid",
			name:      "hbase",
			runtime: &datav1alpha1.AlluxioRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
			},
			Recorder: record.NewFakeRecorder(1),
		},
		{
			Client:    client,
			Log:       fake.NullLogger(),
			namespace: "fluid",
			name:      "spark",
			runtime: &datav1alpha1.AlluxioRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "spark",
					Namespace: "fluid",
				},
			},
			Recorder: record.NewFakeRecorder(1),
		},
	}

	var testCase = []struct {
		engine                               AlluxioEngine
		expectedWorkerPhase                  datav1alpha1.RuntimePhase
		expectedErrorNil                     bool
		expectedRuntimeFuseNumberReady       int32
		expectedRuntimeFuseNumberAvailable   int32
		expectedRuntimeFuseNumberUnavailable int32
	}{
		{
			engine:                               engines[0],
			expectedWorkerPhase:                  datav1alpha1.RuntimePhaseNotReady,
			expectedErrorNil:                     true,
			expectedRuntimeFuseNumberReady:       1,
			expectedRuntimeFuseNumberAvailable:   1,
			expectedRuntimeFuseNumberUnavailable: 1,
		},
		{
			engine:                               engines[1],
			expectedWorkerPhase:                  "",
			expectedErrorNil:                     true,
			expectedRuntimeFuseNumberReady:       1,
			expectedRuntimeFuseNumberAvailable:   1,
			expectedRuntimeFuseNumberUnavailable: 0,
		},
	}

	for _, test := range testCase {
		runtimeInfo, _ := base.BuildRuntimeInfo(test.engine.name, test.engine.namespace, common.AlluxioRuntime)
		test.engine.Helper = ctrl.BuildHelper(runtimeInfo, client, test.engine.Log)
		_, err := test.engine.checkFuseHealthy()
		if err != nil && test.expectedErrorNil == true ||
			err == nil && test.expectedErrorNil == false {
			t.Errorf("fail to exec the checkMasterHealthy function with err %v", err)
			return
		}

		alluxioruntime, err := test.engine.getRuntime()
		if err != nil {
			t.Errorf("fail to get the runtime with the error %v", err)
			return
		}

		if alluxioruntime.Status.FuseNumberReady != test.expectedRuntimeFuseNumberReady ||
			alluxioruntime.Status.FuseNumberAvailable != test.expectedRuntimeFuseNumberAvailable ||
			alluxioruntime.Status.FuseNumberUnavailable != test.expectedRuntimeFuseNumberUnavailable {
			t.Errorf("fail to update the runtime")
			return
		}

		_, cond := utils.GetRuntimeCondition(alluxioruntime.Status.Conditions, datav1alpha1.RuntimeFusesReady)
		if cond == nil {
			t.Errorf("fail to update the condition")
			return
		}
	}
}
