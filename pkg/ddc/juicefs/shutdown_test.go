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
	"reflect"
	"testing"

	"github.com/go-logr/logr"

	"github.com/fluid-cloudnative/fluid/pkg/ctrl"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/juicefs/operations"

	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

const (
	CommonStatus = `{
  "Setting": {
    "Name": "zww-juicefs",
    "UUID": "73416457-6f3f-490b-abb6-cbc1f837944e",
    "Storage": "minio",
    "Bucket": "http://10.98.166.242:9000/zww-juicefs",
    "AccessKey": "minioadmin",
    "SecretKey": "removed",
    "BlockSize": 4096,
    "Compression": "none",
    "Shards": 0,
    "HashPrefix": false,
    "Capacity": 0,
    "Inodes": 0,
    "KeyEncrypted": false,
    "TrashDays": 2,
    "MetaVersion": 0,
    "MinClientVersion": "",
    "MaxClientVersion": ""
  },
  "Sessions": [
    {
      "Sid": 14,
      "Expire": "2022-02-09T10:01:50Z",
      "Version": "1.0-dev (2022-02-09 748949ac)",
      "HostName": "juicefs-pvc-33d9bdf3-5fb5-42fe-bf48-d3d6156b424b-createvol2dv4j",
      "MountPoint": "/mnt/jfs",
      "ProcessID": 20
    }
  ]
}`
)

func mockRunningPodsOfDaemonSet() (pods []corev1.Pod) {
	return []corev1.Pod{{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "fluid",
		},
		Spec:   corev1.PodSpec{},
		Status: corev1.PodStatus{},
	}}
}

func mockRunningPodsOfStatefulSet() (pods []corev1.Pod) {
	return []corev1.Pod{{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "fluid",
		},
		Spec:   corev1.PodSpec{},
		Status: corev1.PodStatus{},
	}}
}

func TestDestroyWorker(t *testing.T) {
	// runtimeInfoSpark tests destroy Worker in exclusive mode.
	runtimeInfoSpark, err := base.BuildRuntimeInfo("spark", "fluid", common.JuiceFSRuntime)
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfoSpark.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
	})

	// runtimeInfoSpark tests destroy Worker in shareMode mode.
	runtimeInfoHadoop, err := base.BuildRuntimeInfo("hadoop", "fluid", common.JuiceFSRuntime)
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

	var nodeInputs = []*corev1.Node{
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
		engine := &JuiceFSEngine{Log: fake.NullLogger(), runtimeInfo: test.runtimeInfo}
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

	fakeClient := fake.NewFakeClientWithScheme(testScheme)
	engine := JuiceFSEngine{
		name:      "test",
		namespace: "fluid",
		Client:    fakeClient,
		Log:       fake.NullLogger(),
		runtime: &datav1alpha1.JuiceFSRuntime{
			Spec: datav1alpha1.JuiceFSRuntimeSpec{
				Fuse: datav1alpha1.JuiceFSFuseSpec{},
			},
		},
	}

	// check release found & delete common
	checkReleasePatch := ApplyFunc(helm.CheckRelease, mockExecCheckReleaseCommonFound)
	deleteReleasePatch := ApplyFunc(helm.DeleteRelease, mockExecDeleteReleaseCommon)
	err := engine.destroyMaster()
	if err != nil {
		t.Errorf("fail to exec check helm release: %v", err)
	}
	checkReleasePatch.Reset()
	deleteReleasePatch.Reset()

	// check release not found
	checkReleasePatch.ApplyFunc(helm.CheckRelease, mockExecCheckReleaseCommonNotFound)
	err = engine.destroyMaster()
	if err != nil {
		t.Errorf("fail to exec check helm release: %v", err)
	}
	checkReleasePatch.Reset()

	// check release error
	checkReleasePatch.ApplyFunc(helm.CheckRelease, mockExecCheckReleaseErr)
	err = engine.destroyMaster()
	if err == nil {
		t.Errorf("fail to exec check helm release: %v", err)
	}
	checkReleasePatch.Reset()

	// check release found & delete common error
	checkReleasePatch.ApplyFunc(helm.CheckRelease, mockExecCheckReleaseCommonFound)
	deleteReleasePatch.ApplyFunc(helm.DeleteRelease, mockExecDeleteReleaseErr)
	err = engine.destroyMaster()
	if err == nil {
		t.Errorf("fail to exec check helm release: %v", err)
	}
	checkReleasePatch.Reset()
	deleteReleasePatch.Reset()
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
	testRuntimeWithTiredStore := &datav1alpha1.JuiceFSRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test2",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.JuiceFSRuntimeSpec{
			Replicas: 1,
			TieredStore: datav1alpha1.TieredStore{
				Levels: []datav1alpha1.Level{{
					MediumType: "MEM",
					Path:       "/data",
					Quota:      resource.NewQuantity(1024, resource.BinarySI),
				}},
			},
		},
		Status: datav1alpha1.RuntimeStatus{
			CacheStates: map[common.CacheStateName]string{
				common.Cached: "true",
			},
		},
	}
	testConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "test-juicefs-values", Namespace: "fluid"},
		Data:       map[string]string{"data": "{\"edition\": \"enterprise\", \"source\": \"test\"}"},
	}

	testObjs := []runtime.Object{testRuntime.DeepCopy(), testRuntimeWithTiredStore.DeepCopy(), testConfigMap.DeepCopy()}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	Convey("Test CleanupCache ", t, func() {
		Convey("cleanup success", func() {
			var engine *JuiceFSEngine
			patch1 := ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
				func(_ *JuiceFSEngine, dsName string, namespace string) ([]corev1.Pod, error) {
					r := mockRunningPodsOfStatefulSet()
					return r, nil
				})
			defer patch1.Reset()
			patch2 := ApplyMethod(reflect.TypeOf(operations.JuiceFileUtils{}), "DeleteCacheDirs",
				func(_ operations.JuiceFileUtils, cacheDirs []string) error {
					return nil
				})
			defer patch2.Reset()

			e := &JuiceFSEngine{
				name:        "test",
				namespace:   "fluid",
				Client:      client,
				runtime:     testRuntime,
				runtimeType: common.JuiceFSRuntime,
				Log:         fake.NullLogger(),
			}

			got := e.cleanupCache()
			So(got, ShouldEqual, nil)
		})
		Convey("test1", func() {
			var engine *JuiceFSEngine
			patch1 := ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
				func(_ *JuiceFSEngine, dsName string, namespace string) ([]corev1.Pod, error) {
					r := mockRunningPodsOfStatefulSet()
					return r, nil
				})
			defer patch1.Reset()
			patch2 := ApplyMethod(reflect.TypeOf(operations.JuiceFileUtils{}), "DeleteCacheDirs",
				func(_ operations.JuiceFileUtils, cacheDirs []string) error {
					return errors.New("delete dir error")
				})
			defer patch2.Reset()

			e := &JuiceFSEngine{
				name:      "test",
				namespace: "fluid",
				Client:    client,
				runtime:   testRuntime,
				Log:       fake.NullLogger(),
			}

			got := e.cleanupCache()
			So(got, ShouldNotBeNil)
		})
		Convey("test2", func() {
			var engine *JuiceFSEngine
			patch1 := ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
				func(_ *JuiceFSEngine, dsName string, namespace string) ([]corev1.Pod, error) {
					r := mockRunningPodsOfStatefulSet()
					return r, nil
				})
			defer patch1.Reset()
			patch2 := ApplyMethod(reflect.TypeOf(operations.JuiceFileUtils{}), "DeleteCacheDirs",
				func(_ operations.JuiceFileUtils, cacheDirs []string) error {
					return errors.New("delete dir error")
				})
			defer patch2.Reset()

			e := &JuiceFSEngine{
				name:      "test",
				namespace: "fluid",
				Client:    client,
				runtime:   testRuntimeWithTiredStore,
				Log:       fake.NullLogger(),
			}

			got := e.cleanupCache()
			So(got, ShouldNotBeNil)
		})
		Convey("test3", func() {
			var engine *JuiceFSEngine
			patch1 := ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
				func(_ *JuiceFSEngine, dsName string, namespace string) ([]corev1.Pod, error) {
					return []corev1.Pod{}, apierrs.NewNotFound(schema.GroupResource{}, "test")
				})
			defer patch1.Reset()

			e := &JuiceFSEngine{
				name:      "test",
				namespace: "fluid",
				Client:    client,
				runtime:   testRuntimeWithTiredStore,
				Log:       fake.NullLogger(),
			}

			got := e.cleanupCache()
			So(got, ShouldEqual, nil)
		})
		Convey("test4", func() {
			var engine *JuiceFSEngine
			patch1 := ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
				func(_ *JuiceFSEngine, dsName string, namespace string) ([]corev1.Pod, error) {
					return []corev1.Pod{}, errors.New("new error")
				})
			defer patch1.Reset()

			e := &JuiceFSEngine{
				name:      "test",
				namespace: "fluid",
				Client:    client,
				runtime:   testRuntimeWithTiredStore,
				Log:       fake.NullLogger(),
			}

			got := e.cleanupCache()
			So(got, ShouldNotBeNil)
		})
	})
}

func TestJuiceFSEngine_getUUID_community(t *testing.T) {
	testConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "test-juicefs-values", Namespace: "fluid"},
		Data:       map[string]string{"data": "{\"edition\": \"community\", \"source\": \"test\"}"},
	}

	testObjs := []runtime.Object{testConfigMap.DeepCopy()}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	ExecCommon := func(a operations.JuiceFileUtils, source string) (status string, err error) {
		return CommonStatus, nil
	}
	ExecErr := func(a operations.JuiceFileUtils, source string) (status string, err error) {
		return "", errors.New("fail to run the command")
	}
	pod := corev1.Pod{}
	e := JuiceFSEngine{
		name:        "test",
		namespace:   "fluid",
		runtimeType: "juicefs",
		engineImpl:  "juicefs",
		Log:         fake.NullLogger(),
		Client:      client,
	}

	patches := ApplyMethod(operations.JuiceFileUtils{}, "GetStatus", ExecErr)
	defer patches.Reset()
	_, err := e.getUUID(pod, common.JuiceFSWorkerContainer)
	if err == nil {
		t.Error("getUUID failure, want err, got nil")
	}

	patches.ApplyMethod(operations.JuiceFileUtils{}, "GetStatus", ExecCommon)
	got, err := e.getUUID(pod, common.JuiceFSWorkerContainer)
	if err != nil {
		t.Errorf("getUUID failure, want nil, got err: %v", err)
	}
	if got != "73416457-6f3f-490b-abb6-cbc1f837944e" {
		t.Errorf("getUUID err, got: %v", got)
	}
}

func TestJuiceFSEngine_getUUID_enterprise(t *testing.T) {
	testConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "test-juicefs-values", Namespace: "fluid"},
		Data:       map[string]string{"data": "{\"edition\": \"enterprise\", \"source\": \"test\"}"},
	}

	testObjs := []runtime.Object{testConfigMap.DeepCopy()}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	pod := corev1.Pod{}
	e := JuiceFSEngine{
		name:        "test",
		namespace:   "fluid",
		runtimeType: "juicefs",
		engineImpl:  "juicefs",
		Log:         fake.NullLogger(),
		Client:      client,
	}

	got, err := e.getUUID(pod, common.JuiceFSWorkerContainer)
	if err != nil {
		t.Errorf("getUUID failure, want nil, got err: %v", err)
	}
	if got != "test" {
		t.Errorf("getUUID err, got: %v", got)
	}
}

func TestJuiceFSEngine_cleanAll(t *testing.T) {
	configMaps := []corev1.ConfigMap{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-config",
				Namespace: "fluid",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-fluid-value",
				Namespace: "fluid",
			},
		},
	}
	testObjs := []runtime.Object{}
	for _, cm := range configMaps {
		testObjs = append(testObjs, cm.DeepCopy())
	}

	fakeClient := fake.NewFakeClientWithScheme(testScheme, testObjs...)
	type fields struct {
		name      string
		namespace string
		Client    client.Client
		log       logr.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				name:      "test",
				namespace: "fluid",
				Client:    fakeClient,
				log:       fake.NullLogger(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helper := &ctrl.Helper{}
			patch1 := ApplyMethod(reflect.TypeOf(helper), "CleanUpFuse", func(_ *ctrl.Helper) (int, error) {
				return 0, nil
			})
			defer patch1.Reset()
			j := &JuiceFSEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    fakeClient,
				Log:       tt.fields.log,
			}
			if err := j.cleanAll(); (err != nil) != tt.wantErr {
				t.Errorf("cleanAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJuiceFSEngine_getCacheDirs(t *testing.T) {
	type args struct {
		runtime *datav1alpha1.JuiceFSRuntime
	}
	tests := []struct {
		name          string
		args          args
		wantCacheDirs []string
	}{
		{
			name: "test-default",
			args: args{
				runtime: &datav1alpha1.JuiceFSRuntime{},
			},
			wantCacheDirs: []string{"/var/jfsCache"},
		},
		{
			name: "test-hostpath",
			args: args{
				runtime: &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						TieredStore: datav1alpha1.TieredStore{
							Levels: []datav1alpha1.Level{{
								MediumType: common.Memory,
								VolumeType: common.VolumeTypeHostPath,
								Path:       "/mnt/ramdisk",
							}},
						},
					},
				},
			},
			wantCacheDirs: []string{"/mnt/ramdisk"},
		},
		{
			name: "test-emptydir",
			args: args{
				runtime: &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						TieredStore: datav1alpha1.TieredStore{
							Levels: []datav1alpha1.Level{{
								MediumType: common.Memory,
								VolumeType: common.VolumeTypeEmptyDir,
								Path:       "/mnt/ramdisk",
							}},
						},
					},
				},
			},
			wantCacheDirs: nil,
		},
		{
			name: "test-multipath-tiredstore",
			args: args{
				runtime: &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						TieredStore: datav1alpha1.TieredStore{
							Levels: []datav1alpha1.Level{{
								MediumType: common.Memory,
								VolumeType: common.VolumeTypeHostPath,
								Path:       "/mnt/ramdisk:/mnt/ramdisk2",
							}},
						},
					},
				},
			},
			wantCacheDirs: []string{"/mnt/ramdisk", "/mnt/ramdisk2"},
		},
		{
			name: "test-worker-cache",
			args: args{
				runtime: &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						TieredStore: datav1alpha1.TieredStore{
							Levels: []datav1alpha1.Level{{
								MediumType: common.Memory,
								VolumeType: common.VolumeTypeHostPath,
								Path:       "/mnt/ramdisk:/mnt/ramdisk2",
							}},
						},
						Worker: datav1alpha1.JuiceFSCompTemplateSpec{
							Options: map[string]string{
								"cache-dir": "/worker/ramdisk",
							},
						},
					},
				},
			},
			wantCacheDirs: []string{"/mnt/ramdisk", "/mnt/ramdisk2", "/worker/ramdisk"},
		},
		{
			name: "test-worker-multi-cache",
			args: args{
				runtime: &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						TieredStore: datav1alpha1.TieredStore{
							Levels: []datav1alpha1.Level{{
								MediumType: common.Memory,
								VolumeType: common.VolumeTypeHostPath,
								Path:       "/mnt/ramdisk:/mnt/ramdisk2",
							}},
						},
						Worker: datav1alpha1.JuiceFSCompTemplateSpec{
							Options: map[string]string{
								"cache-dir": "/worker/ramdisk1:/worker/ramdisk2",
							},
						},
					},
				},
			},
			wantCacheDirs: []string{"/mnt/ramdisk", "/mnt/ramdisk2", "/worker/ramdisk1", "/worker/ramdisk2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JuiceFSEngine{}
			if gotCacheDirs := j.getCacheDirs(tt.args.runtime); !reflect.DeepEqual(gotCacheDirs, tt.wantCacheDirs) {
				t.Errorf("getCacheDirs() = %v, want %v", gotCacheDirs, tt.wantCacheDirs)
			}
		})
	}
}
