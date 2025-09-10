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
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AlluxioEngine master component tests", Label("pkg.ddc.alluxio.master_test.go"), func() {
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
		}
	})

	// JustBeforeEach is guaranteed to run after every BeforeEach()
	// So it's easy to modify resources' specs with an extra BeforeEach()
	JustBeforeEach(func() {
		client = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, resources...)
		engine.runtimeInfo.SetFuseName(engine.getFuseName())
		engine.Client = client
	})

	Describe("Test AlluxioEngine.CheckMasterReady()", func() {
		JustBeforeEach(func() {
			engine.Helper = ctrl.BuildHelper(engine.runtimeInfo, engine.Client, engine.Log)
		})
		When("all master replicas are ready", func() {
			BeforeEach(func() {
				mockedObjects.MasterSts.Spec.Replicas = ptr.To[int32](1)
				mockedObjects.MasterSts.Status.Replicas = 1
				mockedObjects.MasterSts.Status.ReadyReplicas = 1
			})

			It("Should return ready as true", func() {
				ready, err := engine.CheckMasterReady()
				Expect(err).To(BeNil())
				Expect(ready).To(BeTrue())

				// Check if the runtime status is updated correctly
				updatedRuntime, err := utils.GetAlluxioRuntime(client, alluxioruntime.Name, alluxioruntime.Namespace)
				Expect(err).To(BeNil())

				// Check if MasterPhase is set correctly
				Expect(updatedRuntime.Status.MasterPhase).To(Equal(datav1alpha1.RuntimePhaseReady))

				// Check if the condition is set correctly
				Expect(len(updatedRuntime.Status.Conditions)).To(Equal(2))
				Expect(updatedRuntime.Status.Conditions[0].Type).To(Equal(datav1alpha1.RuntimeMasterInitialized))
				Expect(updatedRuntime.Status.Conditions[0].Status).To(Equal(corev1.ConditionTrue))
				Expect(updatedRuntime.Status.Conditions[1].Type).To(Equal(datav1alpha1.RuntimeMasterReady))
				Expect(updatedRuntime.Status.Conditions[1].Status).To(Equal(corev1.ConditionTrue))

				Expect(updatedRuntime.Status.DesiredMasterNumberScheduled).To(Equal(int32(1)))
				Expect(updatedRuntime.Status.CurrentMasterNumberScheduled).To(Equal(int32(1)))
				Expect(updatedRuntime.Status.MasterNumberReady).To(Equal(int32(1)))
			})
		})

		When("not all master replicas are ready", func() {
			BeforeEach(func() {
				mockedObjects.MasterSts.Spec.Replicas = ptr.To[int32](3)
				mockedObjects.MasterSts.Status.ReadyReplicas = 2
				mockedObjects.MasterSts.Status.Replicas = 3
			})

			It("Should return ready as true because it's partially ready", func() {
				ready, err := engine.CheckMasterReady()
				Expect(err).To(BeNil())
				Expect(ready).To(BeTrue())

				// Check if the runtime status is updated correctly
				updatedRuntime, err := utils.GetAlluxioRuntime(client, alluxioruntime.Name, alluxioruntime.Namespace)
				Expect(err).To(BeNil())

				// Check if MasterPhase is set correctly
				Expect(updatedRuntime.Status.MasterPhase).To(Equal(datav1alpha1.RuntimePhasePartialReady))

				// Check if the condition is set correctly
				Expect(len(updatedRuntime.Status.Conditions)).To(Equal(2))
				Expect(updatedRuntime.Status.Conditions[0].Type).To(Equal(datav1alpha1.RuntimeMasterInitialized))
				Expect(updatedRuntime.Status.Conditions[0].Status).To(Equal(corev1.ConditionTrue))
				Expect(updatedRuntime.Status.Conditions[1].Type).To(Equal(datav1alpha1.RuntimeMasterReady))
				Expect(updatedRuntime.Status.Conditions[1].Status).To(Equal(corev1.ConditionTrue))
				Expect(updatedRuntime.Status.DesiredMasterNumberScheduled).To(Equal(int32(3)))
				Expect(updatedRuntime.Status.CurrentMasterNumberScheduled).To(Equal(int32(3)))
				Expect(updatedRuntime.Status.MasterNumberReady).To(Equal(int32(2)))
			})
		})

		When("none of the master replicas is ready", func() {
			BeforeEach(func() {
				mockedObjects.MasterSts.Spec.Replicas = ptr.To[int32](1)
				mockedObjects.MasterSts.Status.Replicas = 1
				mockedObjects.MasterSts.Status.ReadyReplicas = 0
			})

			It("should return ready as false", func() {
				ready, err := engine.CheckMasterReady()
				Expect(err).To(BeNil())
				Expect(ready).To(BeFalse())

				// Check if the runtime status is updated correctly
				updatedRuntime, err := utils.GetAlluxioRuntime(client, alluxioruntime.Name, alluxioruntime.Namespace)
				Expect(err).To(BeNil())

				Expect(len(updatedRuntime.Status.Conditions)).To(Equal(2))
				Expect(updatedRuntime.Status.Conditions[0].Type).To(Equal(datav1alpha1.RuntimeMasterInitialized))
				Expect(updatedRuntime.Status.Conditions[0].Status).To(Equal(corev1.ConditionTrue))
				Expect(updatedRuntime.Status.Conditions[1].Type).To(Equal(datav1alpha1.RuntimeMasterReady))
				Expect(updatedRuntime.Status.Conditions[1].Status).To(Equal(corev1.ConditionFalse))
				Expect(updatedRuntime.Status.DesiredMasterNumberScheduled).To(Equal(int32(1)))
				Expect(updatedRuntime.Status.CurrentMasterNumberScheduled).To(Equal(int32(1)))
				Expect(updatedRuntime.Status.MasterNumberReady).To(Equal(int32(0)))
			})
		})

		When("master replicas is set to 0", func() {
			BeforeEach(func() {
				mockedObjects.MasterSts.Spec.Replicas = ptr.To[int32](0)
			})

			It("Should return ready as true", func() {
				ready, err := engine.CheckMasterReady()
				Expect(err).To(BeNil())
				Expect(ready).To(BeTrue())

				// Check if the runtime status is updated correctly
				updatedRuntime, err := utils.GetAlluxioRuntime(client, alluxioruntime.Name, alluxioruntime.Namespace)
				Expect(err).To(BeNil())

				Expect(len(updatedRuntime.Status.Conditions)).To(Equal(2))
				Expect(updatedRuntime.Status.Conditions[0].Type).To(Equal(datav1alpha1.RuntimeMasterInitialized))
				Expect(updatedRuntime.Status.Conditions[0].Status).To(Equal(corev1.ConditionTrue))
				Expect(updatedRuntime.Status.Conditions[1].Type).To(Equal(datav1alpha1.RuntimeMasterReady))
				Expect(updatedRuntime.Status.Conditions[1].Status).To(Equal(corev1.ConditionTrue))
				Expect(updatedRuntime.Status.DesiredMasterNumberScheduled).To(Equal(int32(0)))
				Expect(updatedRuntime.Status.CurrentMasterNumberScheduled).To(Equal(int32(0)))
				Expect(updatedRuntime.Status.MasterNumberReady).To(Equal(int32(0)))
			})
		})

		When("master replicas is not set", func() {
			BeforeEach(func() {
				mockedObjects.MasterSts.Spec.Replicas = nil
			})

			It("Should return ready as true", func() {
				ready, err := engine.CheckMasterReady()
				Expect(err).To(BeNil())
				Expect(ready).To(BeTrue())

				// Check if the runtime status is updated correctly
				updatedRuntime, err := engine.getRuntime()
				Expect(err).To(BeNil())

				// Check if MasterPhase is set correctly
				Expect(updatedRuntime.Status.MasterPhase).To(Equal(datav1alpha1.RuntimePhaseReady))

				// Check if the condition is set correctly
				Expect(len(updatedRuntime.Status.Conditions)).To(Equal(2))
				Expect(updatedRuntime.Status.Conditions[0].Type).To(Equal(datav1alpha1.RuntimeMasterInitialized))
				Expect(updatedRuntime.Status.Conditions[0].Status).To(Equal(corev1.ConditionTrue))
				Expect(updatedRuntime.Status.Conditions[1].Type).To(Equal(datav1alpha1.RuntimeMasterReady))
				Expect(updatedRuntime.Status.Conditions[1].Status).To(Equal(corev1.ConditionTrue))
				Expect(updatedRuntime.Status.DesiredMasterNumberScheduled).To(Equal(int32(0)))
				Expect(updatedRuntime.Status.CurrentMasterNumberScheduled).To(Equal(int32(0)))
				Expect(updatedRuntime.Status.MasterNumberReady).To(Equal(int32(0)))
			})
		})
	})
})

// TestShouldSetupMaster tests the ShouldSetupMaster function of AlluxioEngine.
// Functionality: Verifies if the Alluxio master should be set up based on runtime status.
// Parameters:
//   - t *testing.T: Standard testing object for test reporting and logging.
//
// Return: None (testing function).
// Notes:
//   - Uses fake client to simulate interactions with Kubernetes API.
//   - Tests two scenarios:
//     1. Runtime with MasterPhase "NotReady" should return false.
//     2. Runtime with MasterPhase "None" should return true.
//   - Fails the test if actual results mismatch expectations.
func TestShouldSetupMaster(t *testing.T) {
	alluxioruntimeInputs := []datav1alpha1.AlluxioRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
			Status: datav1alpha1.RuntimeStatus{
				MasterPhase: datav1alpha1.RuntimePhaseNotReady,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Status: datav1alpha1.RuntimeStatus{
				MasterPhase: datav1alpha1.RuntimePhaseNone,
			},
		},
	}
	testObjs := []runtime.Object{}
	for _, alluxioruntime := range alluxioruntimeInputs {
		testObjs = append(testObjs, alluxioruntime.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []AlluxioEngine{
		{
			name:      "spark",
			namespace: "fluid",
			Client:    client,
		},
		{
			name:      "hbase",
			namespace: "fluid",
			Client:    client,
		},
	}

	var testCases = []struct {
		engine         AlluxioEngine
		expectedResult bool
	}{
		{
			engine:         engines[0],
			expectedResult: false,
		},
		{
			engine:         engines[1],
			expectedResult: true,
		},
	}

	for _, test := range testCases {
		if should, _ := test.engine.ShouldSetupMaster(); should != test.expectedResult {
			t.Errorf("fail to exec the function")
			return
		}
	}
}

// TestSetupMaster tests the SetupMaster function of the AlluxioEngine.
// It initializes a fake Kubernetes client with predefined StatefulSet and AlluxioRuntime objects,
// then verifies whether SetupMaster correctly updates the runtime's status fields.
// The test ensures that:
//  1. The SetupMaster function executes without errors.
//  2. The runtime object is correctly retrieved after execution.
//  3. The runtime's status is properly updated, including the selector,
//     configuration map name, and the presence of conditions.。
func TestSetupMaster(t *testing.T) {
	statefulSetInputs := []appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark-master",
				Namespace: "fluid",
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		},
	}

	testObjs := []runtime.Object{}
	for _, statefulSet := range statefulSetInputs {
		testObjs = append(testObjs, statefulSet.DeepCopy())
	}

	alluxioruntimeInputs := []datav1alpha1.AlluxioRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
		},
	}
	for _, alluxioruntime := range alluxioruntimeInputs {
		testObjs = append(testObjs, alluxioruntime.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []AlluxioEngine{
		{
			name:      "spark",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
	}

	var testCases = []struct {
		engine                AlluxioEngine
		expectedSelector      string
		expectedConfigMapName string
	}{
		{
			engine:                engines[0],
			expectedConfigMapName: "spark--values",
			expectedSelector:      "app=alluxio,release=spark,role=alluxio-worker",
		},
	}

	for _, test := range testCases {
		if err := test.engine.SetupMaster(); err != nil {
			t.Errorf("fail to exec the func with error %v", err)
			return
		}
		alluxioruntime, err := test.engine.getRuntime()
		if err != nil {
			t.Errorf("fail to get the runtime")
			return
		}
		if alluxioruntime.Status.Selector != test.expectedSelector || alluxioruntime.Status.ValueFileConfigmap != test.expectedConfigMapName ||
			len(alluxioruntime.Status.Conditions) == 0 {
			t.Errorf("fail to update the runtime")
			return
		}
	}
}
