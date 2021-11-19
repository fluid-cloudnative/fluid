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

package juicefs

import (
	"errors"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/juicefs/operations"
	"reflect"
	"testing"

	. "github.com/agiledragon/gomonkey"
	"github.com/brahma-adshonor/gohook"
	. "github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	apimachineryRuntime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

func mockRunningPodsOfDaemonSet() (pods []corev1.Pod) {
	return []corev1.Pod{{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "fluid",
		},
		Spec:   v1.PodSpec{},
		Status: v1.PodStatus{},
	}}
}

func TestDestroyWorker(t *testing.T) {
	// runtimeInfoSpark tests destroy Worker in exclusive mode.
	runtimeInfoSpark, err := base.BuildRuntimeInfo("spark", "fluid", "juicefs", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfoSpark.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
	})

	// runtimeInfoSpark tests destroy Worker in shareMode mode.
	runtimeInfoHadoop, err := base.BuildRuntimeInfo("hadoop", "fluid", "juicefs", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfoHadoop.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ShareMode},
	})
	nodeSelector := map[string]string{
		"node-select": "true",
	}
	runtimeInfoHadoop.SetupFuseDeployMode(true, nodeSelector)

	var nodeInputs = []*v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-spark",
				Labels: map[string]string{
					"fluid.io/dataset-num":               "1",
					"fluid.io/s-juicefs-fluid-spark":     "true",
					"fluid.io/s-fluid-spark":             "true",
					"fluid.io/s-h-juicefs-d-fluid-spark": "5B",
					"fluid.io/s-h-juicefs-m-fluid-spark": "1B",
					"fluid.io/s-h-juicefs-t-fluid-spark": "6B",
					"fluid_exclusive":                    "fluid_spark",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-share",
				Labels: map[string]string{
					"fluid.io/dataset-num":                "2",
					"fluid.io/s-juicefs-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":             "true",
					"fluid.io/s-h-juicefs-d-fluid-hadoop": "5B",
					"fluid.io/s-h-juicefs-m-fluid-hadoop": "1B",
					"fluid.io/s-h-juicefs-t-fluid-hadoop": "6B",
					"fluid.io/s-juicefs-fluid-hbase":      "true",
					"fluid.io/s-fluid-hbase":              "true",
					"fluid.io/s-h-juicefs-d-fluid-hbase":  "5B",
					"fluid.io/s-h-juicefs-m-fluid-hbase":  "1B",
					"fluid.io/s-h-juicefs-t-fluid-hbase":  "6B",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-hadoop",
				Labels: map[string]string{
					"fluid.io/dataset-num":                "1",
					"fluid.io/s-juicefs-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":             "true",
					"fluid.io/s-h-juicefs-d-fluid-hadoop": "5B",
					"fluid.io/s-h-juicefs-m-fluid-hadoop": "1B",
					"fluid.io/s-h-juicefs-t-fluid-hadoop": "6B",
					"node-select":                         "true",
				},
			},
		},
	}

	testNodes := []runtime.Object{}
	for _, nodeInput := range nodeInputs {
		testNodes = append(testNodes, nodeInput.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testNodes...)

	var testCase = []struct {
		expectedWorkers  int32
		runtimeInfo      base.RuntimeInfoInterface
		wantedNodeNumber int32
		wantedNodeLabels map[string]map[string]string
	}{
		{
			expectedWorkers:  -1,
			runtimeInfo:      runtimeInfoSpark,
			wantedNodeNumber: 0,
			wantedNodeLabels: map[string]map[string]string{
				"test-node-spark": {},
				"test-node-share": {
					"fluid.io/dataset-num":                "2",
					"fluid.io/s-juicefs-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":             "true",
					"fluid.io/s-h-juicefs-d-fluid-hadoop": "5B",
					"fluid.io/s-h-juicefs-m-fluid-hadoop": "1B",
					"fluid.io/s-h-juicefs-t-fluid-hadoop": "6B",
					"fluid.io/s-juicefs-fluid-hbase":      "true",
					"fluid.io/s-fluid-hbase":              "true",
					"fluid.io/s-h-juicefs-d-fluid-hbase":  "5B",
					"fluid.io/s-h-juicefs-m-fluid-hbase":  "1B",
					"fluid.io/s-h-juicefs-t-fluid-hbase":  "6B",
				},
				"test-node-hadoop": {
					"fluid.io/dataset-num":                "1",
					"fluid.io/s-juicefs-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":             "true",
					"fluid.io/s-h-juicefs-d-fluid-hadoop": "5B",
					"fluid.io/s-h-juicefs-m-fluid-hadoop": "1B",
					"fluid.io/s-h-juicefs-t-fluid-hadoop": "6B",
					"node-select":                         "true",
				},
			},
		},
		{
			expectedWorkers:  -1,
			runtimeInfo:      runtimeInfoHadoop,
			wantedNodeNumber: 0,
			wantedNodeLabels: map[string]map[string]string{
				"test-node-spark": {},
				"test-node-share": {
					"fluid.io/dataset-num":               "1",
					"fluid.io/s-juicefs-fluid-hbase":     "true",
					"fluid.io/s-fluid-hbase":             "true",
					"fluid.io/s-h-juicefs-d-fluid-hbase": "5B",
					"fluid.io/s-h-juicefs-m-fluid-hbase": "1B",
					"fluid.io/s-h-juicefs-t-fluid-hbase": "6B",
				},
				"test-node-hadoop": {
					"node-select": "true",
				},
			},
		},
	}
	for _, test := range testCase {
		engine := &JuiceFSEngine{Log: log.NullLogger{}, runtimeInfo: test.runtimeInfo}
		engine.Client = client
		engine.name = test.runtimeInfo.GetName()
		engine.namespace = test.runtimeInfo.GetNamespace()
		if err != nil {
			t.Errorf("fail to exec the function with the error %v", err)
		}
		currentWorkers, err := engine.destroyWorkers(test.expectedWorkers)
		if err != nil {
			t.Errorf("fail to exec the function with the error %v", err)
		}
		if currentWorkers != test.wantedNodeNumber {
			t.Errorf("shutdown the worker with the wrong number of the workers")
		}
		for _, node := range nodeInputs {
			newNode, err := kubeclient.GetNode(client, node.Name)
			if err != nil {
				t.Errorf("fail to get the node with the error %v", err)
			}

			if len(newNode.Labels) != len(test.wantedNodeLabels[node.Name]) {
				t.Errorf("fail to decrease the labels")
			}
			if len(newNode.Labels) != 0 && !reflect.DeepEqual(newNode.Labels, test.wantedNodeLabels[node.Name]) {
				t.Errorf("fail to decrease the labels")
			}
		}

	}
}

func TestJuiceFSEngine_destroyMaster(t *testing.T) {
	mockExecCheckReleaseCommonFound := func(name string, namespace string) (exist bool, err error) {
		return true, nil
	}
	mockExecCheckReleaseCommonNotFound := func(name string, namespace string) (exist bool, err error) {
		return false, nil
	}
	mockExecCheckReleaseErr := func(name string, namespace string) (exist bool, err error) {
		return false, errors.New("fail to check release")
	}
	mockExecDeleteReleaseCommon := func(name string, namespace string) error {
		return nil
	}
	mockExecDeleteReleaseErr := func(name string, namespace string) error {
		return errors.New("fail to delete chart")
	}

	wrappedUnhookCheckRelease := func() {
		err := gohook.UnHook(helm.CheckRelease)
		if err != nil {
			t.Fatal(err.Error())
		}
	}
	wrappedUnhookDeleteRelease := func() {
		err := gohook.UnHook(helm.DeleteRelease)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	engine := JuiceFSEngine{
		name:      "test",
		namespace: "fluid",
		Log:       log.NullLogger{},
		runtime: &datav1alpha1.JuiceFSRuntime{
			Spec: datav1alpha1.JuiceFSRuntimeSpec{
				Fuse: datav1alpha1.JuiceFSFuseSpec{},
			},
		},
	}

	// check release found & delete common
	err := gohook.Hook(helm.CheckRelease, mockExecCheckReleaseCommonFound, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook(helm.DeleteRelease, mockExecDeleteReleaseCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.destroyMaster()
	if err != nil {
		t.Errorf("fail to exec check helm release: %v", err)
	}
	wrappedUnhookCheckRelease()
	wrappedUnhookDeleteRelease()

	// check release not found
	err = gohook.Hook(helm.CheckRelease, mockExecCheckReleaseCommonNotFound, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.destroyMaster()
	if err != nil {
		t.Errorf("fail to exec check helm release: %v", err)
	}

	// check release error
	err = gohook.Hook(helm.CheckRelease, mockExecCheckReleaseErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.destroyMaster()
	if err == nil {
		t.Errorf("fail to exec check helm release: %v", err)
	}

	// check release found & delete common error
	err = gohook.Hook(helm.CheckRelease, mockExecCheckReleaseCommonFound, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook(helm.DeleteRelease, mockExecDeleteReleaseErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.destroyMaster()
	if err == nil {
		t.Errorf("fail to exec check helm release: %v", err)
	}
	wrappedUnhookDeleteRelease()
}

func TestJuiceFSEngine_cleanupCache(t *testing.T) {
	testRuntime := &datav1alpha1.JuiceFSRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.JuiceFSRuntimeSpec{
			Replicas: 1,
		},
		Status: datav1alpha1.RuntimeStatus{
			CacheStates: map[common.CacheStateName]string{
				common.Cached: "true",
			},
		},
	}

	s := apimachineryRuntime.NewScheme()
	s.AddKnownTypes(datav1alpha1.GroupVersion, testRuntime)
	client := fake.NewFakeClientWithScheme(testScheme, testRuntime)

	Convey("Test CleanupCache ", t, func() {
		Convey("cleanup success", func() {
			var engine *JuiceFSEngine
			patch1 := ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfDaemonset",
				func(_ *JuiceFSEngine, dsName string, namespace string) ([]corev1.Pod, error) {
					r := mockRunningPodsOfDaemonSet()
					return r, nil
				})
			defer patch1.Reset()
			patch2 := ApplyMethod(reflect.TypeOf(operations.JuiceFileUtils{}), "DeleteDir",
				func(_ operations.JuiceFileUtils, cacheDir string) error {
					return nil
				})
			defer patch2.Reset()

			e := &JuiceFSEngine{
				name:      "test",
				namespace: "fluid",
				Client:    client,
				runtime:   testRuntime,
				Log:       log.NullLogger{},
			}

			got := e.cleanupCache()
			So(got, ShouldEqual, nil)
		})
		Convey("test", func() {
			var engine *JuiceFSEngine
			patch1 := ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfDaemonset",
				func(_ *JuiceFSEngine, dsName string, namespace string) ([]corev1.Pod, error) {
					r := mockRunningPodsOfDaemonSet()
					return r, nil
				})
			defer patch1.Reset()
			patch2 := ApplyMethod(reflect.TypeOf(operations.JuiceFileUtils{}), "DeleteDir",
				func(_ operations.JuiceFileUtils, cacheDir string) error {
					return errors.New("delete dir error")
				})
			defer patch2.Reset()

			e := &JuiceFSEngine{
				name:      "test",
				namespace: "fluid",
				Client:    client,
				runtime:   testRuntime,
				Log:       log.NullLogger{},
			}

			got := e.cleanupCache()
			So(got, ShouldNotBeNil)
		})
	})
}
