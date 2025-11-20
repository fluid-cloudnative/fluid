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

package ctrl

import (
	"context"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ctrl Worker Tests", Label("pkg.ctrl.master_test.go"), func() {
	var helper *Helper
	var resources []runtime.Object
	var workerSts *appsv1.StatefulSet
	var k8sClient client.Client
	var runtimeInfo base.RuntimeInfoInterface
	BeforeEach(func() {
		workerSts = mockRuntimeStatefulset("test-helper-worker", "fluid")
		resources = []runtime.Object{
			workerSts,
		}
	})
	JustBeforeEach(func() {
		k8sClient = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, resources...)
		helper = BuildHelper(runtimeInfo, k8sClient, fake.NullLogger())
	})

	Describe("Test Helper.CheckAndUpdateWorkerStatus()", func() {
		When("master sts is ready", func() {
			BeforeEach(func() {
				workerSts.Spec.Replicas = ptr.To[int32](1)
				workerSts.Status.AvailableReplicas = 1
				workerSts.Status.Replicas = 1
				workerSts.Status.ReadyReplicas = 1
			})

			When("applying to AlluxioRuntime", func() {
				var alluxioruntime *datav1alpha1.AlluxioRuntime
				BeforeEach(func() {
					alluxioruntime = &datav1alpha1.AlluxioRuntime{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-alluxio",
							Namespace: "fluid",
						},
						Spec:   datav1alpha1.AlluxioRuntimeSpec{},
						Status: datav1alpha1.RuntimeStatus{},
					}
					resources = append(resources, alluxioruntime)
				})
				It("should update AlluxioRuntime's status successfully", func() {
					getRuntimeFn := func(k8sClient client.Client) (base.RuntimeInterface, error) {
						runtime := &datav1alpha1.AlluxioRuntime{}
						err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: alluxioruntime.Name}, runtime)
						return runtime, err
					}

					ready, err := helper.CheckAndSyncWorkerStatus(getRuntimeFn, types.NamespacedName{Namespace: workerSts.Namespace, Name: workerSts.Name})
					Expect(err).To(BeNil())
					Expect(ready).To(BeTrue())

					gotRuntime := &datav1alpha1.AlluxioRuntime{}
					err = k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: alluxioruntime.Name}, gotRuntime)
					Expect(err).To(BeNil())
					Expect(gotRuntime.Status.WorkerPhase).To(Equal(datav1alpha1.RuntimePhaseReady))
					Expect(gotRuntime.Status.DesiredWorkerNumberScheduled).To(BeEquivalentTo(1))
					Expect(gotRuntime.Status.CurrentWorkerNumberScheduled).To(BeEquivalentTo(1))
					Expect(gotRuntime.Status.WorkerNumberReady).To(BeEquivalentTo(1))
					Expect(gotRuntime.Status.WorkerNumberAvailable).To(BeEquivalentTo(1))
					Expect(gotRuntime.Status.WorkerNumberUnavailable).To(BeEquivalentTo(0))

					Expect(gotRuntime.Status.Conditions).To(HaveLen(1))
					Expect(gotRuntime.Status.Conditions[0].Type).To(Equal(datav1alpha1.RuntimeWorkersReady))
					Expect(gotRuntime.Status.Conditions[0].Status).To(Equal(corev1.ConditionTrue))
				})
			})
		})

	})

})

func TestGetWorkersAsStatefulset(t *testing.T) {

	statefulsetInputs := []*appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sts-jindofs-worker",
				Namespace: "big-data",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: ptr.To[int32](2),
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		},
	}

	objs := []runtime.Object{}

	for _, statefulsetInput := range statefulsetInputs {
		objs = append(objs, statefulsetInput.DeepCopy())
	}

	s := runtime.NewScheme()
	_ = appsv1.AddToScheme(s)
	fakeClient := fake.NewFakeClientWithScheme(s, objs...)
	testCases := []struct {
		name            string
		key             types.NamespacedName
		success         bool
		deprecatedError bool
	}{
		{
			name: "noError",
			key: types.NamespacedName{
				Name:      "sts-jindofs-worker",
				Namespace: "big-data",
			},
			success:         true,
			deprecatedError: false,
		}, {
			name: "otherError",
			key: types.NamespacedName{
				Name:      "test-jindofs-worker",
				Namespace: "big-data",
			},
			success:         false,
			deprecatedError: false,
		},
	}

	for _, testCase := range testCases {
		_, err := GetWorkersAsStatefulset(fakeClient, testCase.key)

		if testCase.success != (err == nil) {
			t.Errorf("testcase %s failed due to expect succcess %v, got error %v", testCase.name, testCase.success, err)
		}
	}

}

func TestSetupWorkers(t *testing.T) {

	// runtimeInfoSpark tests create worker in exclusive mode.

	runtimeInfoSpark, err := base.BuildRuntimeInfo("spark", "big-data", common.JindoRuntime)

	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfoSpark.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
	})

	// runtimeInfoSpark tests create worker in shareMode mode.
	runtimeInfoHadoop, err := base.BuildRuntimeInfo("hadoop", "big-data", common.JindoRuntime)
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfoHadoop.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ShareMode},
	})
	nodeSelector := map[string]string{
		"node-select": "true",
	}
	runtimeInfoHadoop.SetFuseNodeSelector(nodeSelector)

	type fields struct {
		replicas    int32
		nodeInputs  []*corev1.Node
		worker      appsv1.StatefulSet
		runtime     *datav1alpha1.JindoRuntime
		runtimeInfo base.RuntimeInfoInterface
		name        string
		namespace   string
	}
	tests := []struct {
		name             string
		fields           fields
		wantedNodeLabels map[string]map[string]string
	}{
		{
			name: "test0",
			fields: fields{
				replicas: 1,
				nodeInputs: []*corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-node-spark",
						},
					},
				},
				worker: appsv1.StatefulSet{

					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-jindofs-worker",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
				},
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Replicas: 1,
					},
				},
				runtimeInfo: runtimeInfoSpark,
				name:        "spark",
				namespace:   "big-data",
			},
			wantedNodeLabels: map[string]map[string]string{
				"test-node-spark": {
					"fluid.io/dataset-num":                "1",
					"fluid.io/s-jindo-big-data-spark":     "true",
					"fluid.io/s-big-data-spark":           "true",
					"fluid.io/s-h-jindo-t-big-data-spark": "0B",
					"fluid_exclusive":                     "big-data_spark",
				},
			},
		},
		{
			name: "test1",
			fields: fields{
				replicas: 3,
				worker: appsv1.StatefulSet{

					ObjectMeta: metav1.ObjectMeta{
						Name:      "hadoop-jindofs-worker",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
				},
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hadoop",
						Namespace: "big-data",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Replicas: 3,
					},
				},
				runtimeInfo: runtimeInfoHadoop,
				name:        "hadoop",
				namespace:   "big-data",
			},
			wantedNodeLabels: map[string]map[string]string{
				"test-node-hadoop": {
					"fluid.io/dataset-num":                 "1",
					"fluid.io/s-jindo-big-data-hadoop":     "true",
					"fluid.io/s-big-data-hadoop":           "true",
					"fluid.io/s-h-jindo-t-big-data-hadoop": "0B",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtimeObjs := []runtime.Object{}
			for _, nodeInput := range tt.fields.nodeInputs {
				runtimeObjs = append(runtimeObjs, nodeInput.DeepCopy())
			}
			runtimeObjs = append(runtimeObjs, tt.fields.worker.DeepCopy())

			s := runtime.NewScheme()
			data := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tt.fields.name,
					Namespace: tt.fields.namespace,
				},
			}
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.runtime)
			s.AddKnownTypes(datav1alpha1.GroupVersion, data)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, &tt.fields.worker)
			_ = corev1.AddToScheme(s)
			runtimeObjs = append(runtimeObjs, tt.fields.runtime)
			runtimeObjs = append(runtimeObjs, data)
			mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)

			h := BuildHelper(tt.fields.runtimeInfo, mockClient, fake.NullLogger())

			err := h.SetupWorkers(tt.fields.runtime, tt.fields.runtime.Status, &tt.fields.worker)

			if err != nil {
				t.Errorf("test case %s h.SetupWorkers() error = %v", t.Name(), err)
			}

			worker := &appsv1.StatefulSet{}
			key := types.NamespacedName{
				Namespace: tt.fields.worker.Namespace,
				Name:      tt.fields.worker.Name,
			}

			err = mockClient.Get(context.TODO(), key, worker)
			if err != nil {
				t.Errorf("test case %s mockClient.Get() error = %v", t.Name(), err)
			}

			if tt.fields.replicas != *worker.Spec.Replicas {
				t.Errorf("Failed to scale %v for %v", tt.name, tt.fields)
			}

			// for _, node := range tt.fields.nodeInputs {
			// 	newNode, err := kubeclient.GetNode(mockClient, node.Name)
			// 	if err != nil {
			// 		t.Errorf("fail to get the node with the error %v", err)
			// 	}

			// 	if len(newNode.Labels) != len(tt.wantedNodeLabels[node.Name]) {
			// 		t.Errorf("fail to decrease the labels, newNode labels is %v", newNode.Labels)
			// 	}
			// 	if len(newNode.Labels) != 0 && !reflect.DeepEqual(newNode.Labels, tt.wantedNodeLabels[node.Name]) {
			// 		t.Errorf("fail to decrease the labels, newNode labels is %v", newNode.Labels)
			// 	}
			// }
		})
	}
}
