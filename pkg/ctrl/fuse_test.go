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
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	testScheme *runtime.Scheme
)

var _ = Describe("Ctrl Helper Fuse Tests", Label("pkg.ctrl.fuse_test.go"), func() {
	var helper *Helper
	var resources []runtime.Object
	var fuseDs *appsv1.DaemonSet
	var k8sClient client.Client
	var runtimeInfo base.RuntimeInfoInterface

	BeforeEach(func() {
		fuseDs = mockRuntimeDaemonset("test-helper-fuse", "fluid")
		resources = []runtime.Object{
			fuseDs,
		}
	})

	JustBeforeEach(func() {
		k8sClient = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, resources...)
		helper = BuildHelper(runtimeInfo, k8sClient, fake.NullLogger())
	})

	Describe("Test Helper.CheckAndUpdateFuseStatus()", func() {
		When("fuse ds is ready", func() {
			BeforeEach(func() {
				fuseDs.Status.DesiredNumberScheduled = 1
				fuseDs.Status.CurrentNumberScheduled = 1
				fuseDs.Status.NumberReady = 1
				fuseDs.Status.NumberAvailable = 1
				fuseDs.Status.NumberUnavailable = 0
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

					ready, err := helper.CheckAndSyncFuseStatus(getRuntimeFn, types.NamespacedName{Namespace: fuseDs.Namespace, Name: fuseDs.Name})
					Expect(err).To(BeNil())
					Expect(ready).To(BeTrue())

					gotRuntime := &datav1alpha1.AlluxioRuntime{}
					err = k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: alluxioruntime.Namespace, Name: alluxioruntime.Name}, gotRuntime)
					Expect(err).To(BeNil())
					Expect(gotRuntime.Status.FusePhase).To(Equal(datav1alpha1.RuntimePhaseReady))
					Expect(gotRuntime.Status.DesiredFuseNumberScheduled).To(BeEquivalentTo(1))
					Expect(gotRuntime.Status.CurrentFuseNumberScheduled).To(BeEquivalentTo(1))
					Expect(gotRuntime.Status.FuseNumberReady).To(BeEquivalentTo(1))
					Expect(gotRuntime.Status.FuseNumberAvailable).To(BeEquivalentTo(1))
					Expect(gotRuntime.Status.FuseNumberUnavailable).To(BeEquivalentTo(0))

					Expect(gotRuntime.Status.Conditions).To(HaveLen(1))
					Expect(gotRuntime.Status.Conditions[0].Type).To(Equal(datav1alpha1.RuntimeFusesReady))
					Expect(gotRuntime.Status.Conditions[0].Status).To(Equal(corev1.ConditionTrue))
				})
			})
		})
	})
})

func TestCleanUpFuse(t *testing.T) {
	var testCase = []struct {
		name             string
		namespace        string
		wantedNodeLabels map[string]map[string]string
		wantedCount      int
		context          cruntime.ReconcileRequestContext
		log              logr.Logger
		runtimeType      string
		nodeInputs       []*corev1.Node
	}{
		{
			wantedCount: 1,
			name:        "hbase",
			namespace:   "fluid",
			wantedNodeLabels: map[string]map[string]string{
				"no-fuse": {},
				"multiple-fuse": {
					"fluid.io/f-fluid-hadoop":          "true",
					"node-select":                      "true",
					"fluid.io/s-fluid-hbase":           "true",
					"fluid.io/s-h-jindo-d-fluid-hbase": "5B",
					"fluid.io/s-h-jindo-m-fluid-hbase": "1B",
					"fluid.io/s-h-jindo-t-fluid-hbase": "6B",
				},
				"fuse": {
					"fluid.io/dataset-num":    "1",
					"fluid.io/f-fluid-hadoop": "true",
					"node-select":             "true",
				},
			},
			log:         fake.NullLogger(),
			runtimeType: "jindo",
			nodeInputs: []*corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "no-fuse",
						Labels: map[string]string{},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "multiple-fuse",
						Labels: map[string]string{
							"fluid.io/f-fluid-hadoop":          "true",
							"node-select":                      "true",
							"fluid.io/f-fluid-hbase":           "true",
							"fluid.io/s-fluid-hbase":           "true",
							"fluid.io/s-h-jindo-d-fluid-hbase": "5B",
							"fluid.io/s-h-jindo-m-fluid-hbase": "1B",
							"fluid.io/s-h-jindo-t-fluid-hbase": "6B",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fuse",
						Labels: map[string]string{
							"fluid.io/dataset-num":    "1",
							"fluid.io/f-fluid-hadoop": "true",
							"node-select":             "true",
						},
					},
				},
			},
		},
		{
			wantedCount: 2,
			name:        "spark",
			namespace:   "fluid",
			wantedNodeLabels: map[string]map[string]string{
				"no-fuse": {},
				"multiple-fuse": {
					"node-select":                        "true",
					"fluid.io/s-fluid-hbase":             "true",
					"fluid.io/f-fluid-hbase":             "true",
					"fluid.io/s-h-alluxio-d-fluid-hbase": "5B",
					"fluid.io/s-h-alluxio-m-fluid-hbase": "1B",
					"fluid.io/s-h-alluxio-t-fluid-hbase": "6B",
				},
				"fuse": {
					"fluid.io/dataset-num": "1",
					"node-select":          "true",
				},
			},
			log:         fake.NullLogger(),
			runtimeType: "alluxio",
			nodeInputs: []*corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "no-fuse",
						Labels: map[string]string{},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "multiple-fuse",
						Labels: map[string]string{
							"fluid.io/f-fluid-spark":             "true",
							"node-select":                        "true",
							"fluid.io/f-fluid-hbase":             "true",
							"fluid.io/s-fluid-hbase":             "true",
							"fluid.io/s-h-alluxio-d-fluid-hbase": "5B",
							"fluid.io/s-h-alluxio-m-fluid-hbase": "1B",
							"fluid.io/s-h-alluxio-t-fluid-hbase": "6B",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fuse",
						Labels: map[string]string{
							"fluid.io/dataset-num":   "1",
							"fluid.io/f-fluid-spark": "true",
							"node-select":            "true",
						},
					},
				},
			},
		},
		{
			wantedCount: 0,
			name:        "hbase",
			namespace:   "fluid",
			wantedNodeLabels: map[string]map[string]string{
				"no-fuse": {},
				"multiple-fuse": {
					"fluid.io/f-fluid-spark":              "true",
					"node-select":                         "true",
					"fluid.io/s-fluid-hadoop":             "true",
					"fluid.io/f-fluid-hadoop":             "true",
					"fluid.io/s-h-goosefs-d-fluid-hadoop": "5B",
					"fluid.io/s-h-goosefs-m-fluid-hadoop": "1B",
					"fluid.io/s-h-goosefs-t-fluid-hadoop": "6B",
				},
				"fuse": {
					"fluid.io/dataset-num":   "1",
					"fluid.io/f-fluid-spark": "true",
					"node-select":            "true",
				},
			},
			log:         fake.NullLogger(),
			runtimeType: "goosefs",
			nodeInputs: []*corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "no-fuse",
						Labels: map[string]string{},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "multiple-fuse",
						Labels: map[string]string{
							"fluid.io/f-fluid-spark":              "true",
							"node-select":                         "true",
							"fluid.io/f-fluid-hadoop":             "true",
							"fluid.io/s-fluid-hadoop":             "true",
							"fluid.io/s-h-goosefs-d-fluid-hadoop": "5B",
							"fluid.io/s-h-goosefs-m-fluid-hadoop": "1B",
							"fluid.io/s-h-goosefs-t-fluid-hadoop": "6B",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fuse",
						Labels: map[string]string{
							"fluid.io/dataset-num":   "1",
							"fluid.io/f-fluid-spark": "true",
							"node-select":            "true",
						},
					},
				},
			},
		},
	}
	for _, test := range testCase {

		testNodes := []runtime.Object{}
		for _, nodeInput := range test.nodeInputs {
			testNodes = append(testNodes, nodeInput.DeepCopy())
		}

		fakeClient := fake.NewFakeClientWithScheme(testScheme, testNodes...)

		nodeList := &corev1.NodeList{}
		runtimeInfo, err := base.BuildRuntimeInfo(
			test.name,
			test.namespace,
			test.runtimeType,
		)
		if err != nil {
			t.Errorf("build runtime info error %v", err)
		}
		h := &Helper{
			runtimeInfo: runtimeInfo,
			client:      fakeClient,
			log:         test.log,
		}

		count, err := h.CleanUpFuse()
		if err != nil {
			t.Errorf("fail to exec the function with the error %v", err)
		}
		if count != test.wantedCount {
			t.Errorf("with the wrong number of the fuse ,count %v", count)
		}

		err = fakeClient.List(context.TODO(), nodeList, &client.ListOptions{})
		if err != nil {
			t.Errorf("testcase %s: fail to get the node with the error %v  ", test.name, err)
		}

		for _, node := range nodeList.Items {
			if len(node.Labels) != len(test.wantedNodeLabels[node.Name]) {
				t.Errorf("testcase %s: fail to clean up the labels for node %s  expected %v, got %v", test.name, node.Name, test.wantedNodeLabels[node.Name], node.Labels)
			}
			if len(node.Labels) != 0 && !reflect.DeepEqual(node.Labels, test.wantedNodeLabels[node.Name]) {
				t.Errorf("testcase %s: fail to clean up the labels for node  %s  expected %v, got %v", test.name, node.Name, test.wantedNodeLabels[node.Name], node.Labels)
			}
		}

	}
}
