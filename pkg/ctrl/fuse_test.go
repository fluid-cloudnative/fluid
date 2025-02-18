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

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	testScheme *runtime.Scheme
)

func TestCheckFuseHealthy(t *testing.T) {
	runtimeInputs := []*datav1alpha1.JindoRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.JindoRuntimeSpec{
				Replicas: 3, // 2
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.JindoRuntimeSpec{
				Replicas: 2,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.JindoRuntimeSpec{
				Replicas: 2,
			},
		},
	}

	podList := &corev1.PodList{
		Items: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-jindofs-fuse-0",
				Namespace: "fluid",
				Labels:    map[string]string{"a": "b"},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodFailed,
				Conditions: []corev1.PodCondition{{
					Type:   corev1.PodReady,
					Status: corev1.ConditionTrue,
				}},
			},
		}},
	}

	dataSetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop",
				Namespace: "fluid",
			},
		},
	}

	dsInputss := []*appsv1.DaemonSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-jindofs-fuse",
				Namespace: "fluid",
			},
			Spec: appsv1.DaemonSetSpec{},
			Status: appsv1.DaemonSetStatus{
				NumberUnavailable:      0,
				NumberReady:            1,
				NumberAvailable:        1,
				DesiredNumberScheduled: 1,
				CurrentNumberScheduled: 1,
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop-jindofs-fuse",
				Namespace: "fluid",
			},
			Spec: appsv1.DaemonSetSpec{},
			Status: appsv1.DaemonSetStatus{
				NumberUnavailable:      1,
				NumberReady:            1,
				NumberAvailable:        1,
				DesiredNumberScheduled: 2,
				CurrentNumberScheduled: 1,
			},
		},
	}

	objs := []runtime.Object{}

	for _, runtimeInput := range runtimeInputs {
		objs = append(objs, runtimeInput.DeepCopy())
	}
	for _, dataSetInput := range dataSetInputs {
		objs = append(objs, dataSetInput.DeepCopy())
	}
	for _, dsInputs := range dsInputss {
		objs = append(objs, dsInputs.DeepCopy())
	}

	for _, pod := range podList.Items {
		objs = append(objs, &pod)
	}

	// objs = append(objs, podList)

	s := runtime.NewScheme()
	_ = corev1.AddToScheme(s)
	_ = datav1alpha1.AddToScheme(s)
	_ = appsv1.AddToScheme(s)
	fakeClient := fake.NewFakeClientWithScheme(s, objs...)
	testCases := []struct {
		caseName   string
		name       string
		namespace  string
		Phase      datav1alpha1.RuntimePhase
		fuse       *appsv1.DaemonSet
		fuseReason string
		TypeValue  bool
		isErr      bool
	}{
		{
			caseName:  "Healthy",
			name:      "hbase",
			namespace: "fluid",
			fuse: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase-jindofs-fuse",
					Namespace: "fluid",
				},
				Spec: appsv1.DaemonSetSpec{},
				Status: appsv1.DaemonSetStatus{
					NumberUnavailable:      0,
					NumberAvailable:        1,
					NumberReady:            1,
					DesiredNumberScheduled: 1,
					CurrentNumberScheduled: 1,
				},
			},
			Phase:      datav1alpha1.RuntimePhaseReady,
			fuseReason: "The fuse is ready.",
			isErr:      false,
		},
		{
			caseName:  "Unhealthy",
			name:      "hadoop",
			namespace: "fluid",
			fuse: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hadoop-jindofs-fuse",
					Namespace: "fluid",
				},
				Spec: appsv1.DaemonSetSpec{},
				Status: appsv1.DaemonSetStatus{
					NumberUnavailable:      1,
					NumberAvailable:        1,
					NumberReady:            1,
					DesiredNumberScheduled: 2,
					CurrentNumberScheduled: 1,
				},
			},
			Phase:      datav1alpha1.RuntimePhaseNotReady,
			fuseReason: "The fuses are not ready.",
			isErr:      false,
		},
	}
	for _, testCase := range testCases {

		runtimeInfo, err := base.BuildRuntimeInfo(testCase.name, testCase.namespace, common.JindoRuntime)
		if err != nil {
			t.Errorf("testcase %s failed due to %v", testCase.name, err)
		}

		var runtime *datav1alpha1.JindoRuntime = &datav1alpha1.JindoRuntime{}

		err = fakeClient.Get(context.TODO(), types.NamespacedName{
			Namespace: testCase.namespace,
			Name:      testCase.name,
		}, runtime)

		if err != nil {
			t.Errorf("testCase %s sync replicas failed,err:%v", testCase.caseName, err)
		}

		ds := &appsv1.DaemonSet{}
		err = fakeClient.Get(context.TODO(), types.NamespacedName{
			Namespace: testCase.fuse.Namespace,
			Name:      testCase.fuse.Name,
		}, ds)
		if err != nil {
			t.Errorf("sync replicas failed,err:%s", err.Error())
		}

		h := BuildHelper(runtimeInfo, fakeClient, fake.NullLogger())

		err = h.CheckFuseHealthy(record.NewFakeRecorder(300), runtime, ds.Name)

		if testCase.isErr == (err == nil) {
			t.Errorf("check fuse's healthy failed,err:%v", err)
		}

		err = fakeClient.Get(context.TODO(), types.NamespacedName{
			Namespace: testCase.namespace,
			Name:      testCase.name,
		}, runtime)

		if err != nil {
			t.Errorf("check fuse's healthy failed,err:%s", err.Error())
		}

		if runtime.Status.FusePhase != testCase.Phase {
			t.Errorf("testcase %s is failed, expect phase %v, got %v", testCase.caseName,
				testCase.Phase,
				runtime.Status.FusePhase)
		}
		if runtime.Status.DesiredFuseNumberScheduled != testCase.fuse.Status.DesiredNumberScheduled {
			t.Errorf("testcase %s is failed, expect DesiredFuseNumberScheduled %v, got %v", testCase.caseName,
				runtime.Status.DesiredFuseNumberScheduled,
				testCase.fuse.Status.DesiredNumberScheduled)
		}

		if runtime.Status.CurrentFuseNumberScheduled != testCase.fuse.Status.CurrentNumberScheduled {
			t.Errorf("testcase %s is failed, expect CurrentFuseNumberScheduled %v, got %v", testCase.caseName,
				runtime.Status.CurrentFuseNumberScheduled,
				testCase.fuse.Status.CurrentNumberScheduled)
		}

		if runtime.Status.FuseNumberUnavailable != testCase.fuse.Status.NumberUnavailable {
			t.Errorf("testcase %s is failed, expect FuseNumberUnavailable %v, got %v", testCase.caseName,
				runtime.Status.FuseNumberUnavailable,
				testCase.fuse.Status.NumberUnavailable)
		}

		if runtime.Status.FuseNumberAvailable != testCase.fuse.Status.NumberAvailable {
			t.Errorf("testcase %s is failed, expect FuseNumberAvailable %v, got %v", testCase.caseName,
				runtime.Status.FuseNumberAvailable,
				testCase.fuse.Status.NumberAvailable)
		}

		if runtime.Status.FuseNumberReady != testCase.fuse.Status.NumberReady {
			t.Errorf("testcase %s is failed, expect FuseNumberReady %v, got %v", testCase.caseName,
				runtime.Status.FuseNumberReady,
				testCase.fuse.Status.NumberReady)
		}
	}
}

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
