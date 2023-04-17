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
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/brahma-adshonor/gohook"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	utilpointer "k8s.io/utils/pointer"
)

func TestSyncMetadata(t *testing.T) {
	QueryMetaDataInfoIntoFileCommon := func(a operations.AlluxioFileUtils, key operations.KeyOfMetaDataFile, filename string) (value string, err error) {
		return "1024", nil
	}

	wrappedUnhookQueryMetaDataInfoIntoFile := func() {
		err := gohook.UnHook(operations.AlluxioFileUtils.QueryMetaDataInfoIntoFile)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	datasetInputs := []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Status: datav1alpha1.DatasetStatus{
				UfsTotal: "2Gi",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
			Status: datav1alpha1.DatasetStatus{
				UfsTotal: "",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{
				DataRestoreLocation: &datav1alpha1.DataRestoreLocation{
					Path:     "local:///host1/erf",
					NodeName: "test-node",
				},
			},
			Status: datav1alpha1.DatasetStatus{
				UfsTotal: "",
			},
		},
	}

	runtimeInputs := []datav1alpha1.AlluxioRuntime{
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
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop",
				Namespace: "fluid",
			},
		},
	}

	testObjs := []runtime.Object{}
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}

	for _, runtimeInput := range runtimeInputs {
		testObjs = append(testObjs, runtimeInput.DeepCopy())
	}

	var statefulsetInputs = []appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-master",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32(2),
			},
			Status: appsv1.StatefulSetStatus{
				Replicas:      3,
				ReadyReplicas: 2,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark-master",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32(2),
			},
			Status: appsv1.StatefulSetStatus{
				Replicas:      3,
				ReadyReplicas: 3,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop-master",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32(2),
			},
			Status: appsv1.StatefulSetStatus{
				Replicas:      3,
				ReadyReplicas: 3,
			},
		},
	}

	for _, statefulset := range statefulsetInputs {
		testObjs = append(testObjs, statefulset.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []AlluxioEngine{
		{
			name:      "hbase",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
		{
			name:      "spark",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
	}

	for _, engine := range engines {
		err := engine.SyncMetadata()
		if err != nil {
			t.Errorf("fail to exec the function due to: %v", err)
		}
	}

	engine := AlluxioEngine{
		name:      "hadoop",
		namespace: "fluid",
		Client:    client,
		Log:       fake.NullLogger(),
	}

	err := gohook.Hook(operations.AlluxioFileUtils.QueryMetaDataInfoIntoFile, QueryMetaDataInfoIntoFileCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.SyncMetadata()
	if err != nil {
		t.Errorf("fail to exec function RestoreMetadataInternal")
	}
	wrappedUnhookQueryMetaDataInfoIntoFile()
}

func TestSyncMetadataWithoutMaster(t *testing.T) {
	var statefulsetInputs = []appsv1.StatefulSet{}

	testObjs := []runtime.Object{}
	for _, statefulset := range statefulsetInputs {
		testObjs = append(testObjs, statefulset.DeepCopy())
	}

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
	}

	for _, daemonSet := range daemonSetInputs {
		testObjs = append(testObjs, daemonSet.DeepCopy())
	}

	var alluxioruntimeInputs = []datav1alpha1.AlluxioRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Status: datav1alpha1.RuntimeStatus{
				MasterPhase: datav1alpha1.RuntimePhaseReady,
				WorkerPhase: datav1alpha1.RuntimePhaseReady,
				FusePhase:   datav1alpha1.RuntimePhaseReady,
			},
		},
	}
	for _, alluxioruntimeInput := range alluxioruntimeInputs {
		testObjs = append(testObjs, alluxioruntimeInput.DeepCopy())
	}

	var datasetInputs = []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
			Status: datav1alpha1.DatasetStatus{
				Phase: datav1alpha1.BoundDatasetPhase,
				HCFSStatus: &datav1alpha1.HCFSStatus{
					Endpoint:                    "test Endpoint",
					UnderlayerFileSystemVersion: "Underlayer HCFS Compatible Version",
				},
			},
		},
	}
	for _, dataset := range datasetInputs {
		testObjs = append(testObjs, dataset.DeepCopy())
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
		},
	}

	var testCase = []struct {
		engine                     AlluxioEngine
		expectedErrorNil           bool
		expectedRuntimeMasterPhase datav1alpha1.RuntimePhase
		expectedRuntimeWorkerPhase datav1alpha1.RuntimePhase
		expectedRuntimeFusePhase   datav1alpha1.RuntimePhase
		expectedDatasetPhase       datav1alpha1.DatasetPhase
	}{
		{
			engine:                     engines[0],
			expectedErrorNil:           false,
			expectedRuntimeMasterPhase: datav1alpha1.RuntimePhaseNotReady,
			expectedRuntimeWorkerPhase: datav1alpha1.RuntimePhaseNotReady,
			expectedRuntimeFusePhase:   datav1alpha1.RuntimePhaseNotReady,
			expectedDatasetPhase:       datav1alpha1.FailedDatasetPhase,
		},
	}

	for _, test := range testCase {
		klog.Info("test")
		err := test.engine.SyncMetadata()
		if err != nil && test.expectedErrorNil == true ||
			err == nil && test.expectedErrorNil == false {
			t.Errorf("fail to exec the SyncMetadata function with err %v", err)
			return
		}

		alluxioruntime, err := test.engine.getRuntime()
		if err != nil {
			t.Errorf("fail to get the runtime with the error %v", err)
			return
		}

		if alluxioruntime.Status.MasterPhase != test.expectedRuntimeMasterPhase {
			t.Errorf("fail to update the runtime master status, get %s, expect %s", alluxioruntime.Status.MasterPhase, test.expectedRuntimeMasterPhase)
			return
		}

		if alluxioruntime.Status.WorkerPhase != test.expectedRuntimeWorkerPhase {
			t.Errorf("fail to update the runtime worker status, get %s, expect %s", alluxioruntime.Status.WorkerPhase, test.expectedRuntimeWorkerPhase)
			return
		}

		if alluxioruntime.Status.FusePhase != test.expectedRuntimeFusePhase {
			t.Errorf("fail to update the runtime fuse status, get %s, expect %s", alluxioruntime.Status.FusePhase, test.expectedRuntimeFusePhase)
			return
		}

		var dataset datav1alpha1.Dataset
		key := types.NamespacedName{
			Name:      alluxioruntime.Name,
			Namespace: alluxioruntime.Namespace,
		}
		err = client.Get(context.TODO(), key, &dataset)

		if err != nil {
			t.Errorf("fail to get the dataset with error %v", err)
			return
		}
		if !reflect.DeepEqual(dataset.Status.Phase, test.expectedDatasetPhase) {
			t.Errorf("fail to update the dataset status, get %s, expect %s", dataset.Status.Phase, test.expectedDatasetPhase)
			return
		}

	}
}

func TestShouldSyncMetadata(t *testing.T) {
	datasetInputs := []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-metadata-sync-done",
				Namespace: "fluid",
			},
			Status: datav1alpha1.DatasetStatus{
				UfsTotal: "2Gi",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-metadata-sync-default",
				Namespace: "fluid",
			},
			Status: datav1alpha1.DatasetStatus{
				UfsTotal: "",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-metadata-sync-autosync-enabled",
				Namespace: "fluid",
			},
			Status: datav1alpha1.DatasetStatus{
				UfsTotal: metadataSyncNotDoneMsg,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-metadata-sync-autosync-disabled",
				Namespace: "fluid",
			},
			Status: datav1alpha1.DatasetStatus{
				UfsTotal: "",
			},
		},
	}
	runtimeInputs := []datav1alpha1.AlluxioRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-metadata-sync-done",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.AlluxioRuntimeSpec{},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-metadata-sync-default",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				RuntimeManagement: datav1alpha1.RuntimeManagement{
					MetadataSyncPolicy: datav1alpha1.MetadataSyncPolicy{
						AutoSync: nil,
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-metadata-sync-autosync-enabled",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				RuntimeManagement: datav1alpha1.RuntimeManagement{
					MetadataSyncPolicy: datav1alpha1.MetadataSyncPolicy{
						AutoSync: utilpointer.BoolPtr(true),
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-metadata-sync-autosync-disabled",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				RuntimeManagement: datav1alpha1.RuntimeManagement{
					MetadataSyncPolicy: datav1alpha1.MetadataSyncPolicy{
						AutoSync: utilpointer.BoolPtr(false),
					},
				},
			},
		},
	}
	testObjs := []runtime.Object{}
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	for _, runtimeInput := range runtimeInputs {
		testObjs = append(testObjs, runtimeInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []AlluxioEngine{
		{
			name:      "test-metadata-sync-done",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
		{
			name:      "test-metadata-sync-default",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
		{
			name:      "test-metadata-sync-autosync-enabled",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
		{
			name:      "test-metadata-sync-autosync-disabled",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
	}

	var testCases = []struct {
		engine         AlluxioEngine
		expectedShould bool
	}{
		{
			engine:         engines[0],
			expectedShould: false,
		},
		{
			engine:         engines[1],
			expectedShould: true,
		},
		{
			engine:         engines[2],
			expectedShould: true,
		},
		{
			engine:         engines[3],
			expectedShould: false,
		},
	}

	for _, test := range testCases {
		should, err := test.engine.shouldSyncMetadata()
		if err != nil || should != test.expectedShould {
			t.Errorf("fail to exec the function due to: %v", err)
		}
	}
}

func TestShouldRestoreMetadata(t *testing.T) {
	datasetInputs := []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{
				DataRestoreLocation: &datav1alpha1.DataRestoreLocation{
					Path:     "local:///host1/erf",
					NodeName: "test-node",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
		},
	}
	testObjs := []runtime.Object{}
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []AlluxioEngine{
		{
			name:      "hbase",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
		{
			name:      "spark",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
	}

	var testCases = []struct {
		engine         AlluxioEngine
		expectedShould bool
	}{
		{
			engine:         engines[0],
			expectedShould: true,
		},
		{
			engine:         engines[1],
			expectedShould: false,
		},
	}
	for _, test := range testCases {
		should, err := test.engine.shouldRestoreMetadata()
		if err != nil || should != test.expectedShould {
			t.Errorf("fail to exec the function")
		}
	}
}

func TestRestoreMetadataInternal(t *testing.T) {
	QueryMetaDataInfoIntoFileCommon := func(a operations.AlluxioFileUtils, key operations.KeyOfMetaDataFile, filename string) (value string, err error) {
		return "1024", nil
	}
	QueryMetaDataInfoIntoFileErr := func(a operations.AlluxioFileUtils, key operations.KeyOfMetaDataFile, filename string) (value string, err error) {
		return "", errors.New("fail to query MetaDataInfo")
	}

	wrappedUnhookQueryMetaDataInfoIntoFile := func() {
		err := gohook.UnHook(operations.AlluxioFileUtils.QueryMetaDataInfoIntoFile)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	datasetInputs := []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{
				DataRestoreLocation: &datav1alpha1.DataRestoreLocation{
					Path:     "local:///host1/erf",
					NodeName: "test-node",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{
				DataRestoreLocation: &datav1alpha1.DataRestoreLocation{
					Path:     "pvc://pvc1/erf",
					NodeName: "test-node",
				},
			},
		},
	}
	testObjs := []runtime.Object{}
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []AlluxioEngine{
		{
			name:      "hbase",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
		{
			name:      "hbase",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
	}

	err := gohook.Hook(operations.AlluxioFileUtils.QueryMetaDataInfoIntoFile, QueryMetaDataInfoIntoFileErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engines[0].RestoreMetadataInternal()
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookQueryMetaDataInfoIntoFile()

	err = gohook.Hook(operations.AlluxioFileUtils.QueryMetaDataInfoIntoFile, QueryMetaDataInfoIntoFileCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	var testCases = []struct {
		engine                  AlluxioEngine
		expectedDatasetUfsTotal string
		expectedDatasetFileNum  string
	}{
		{
			engine:                  engines[0],
			expectedDatasetUfsTotal: "1.00KiB",
			expectedDatasetFileNum:  "1024",
		},
		{
			engine:                  engines[1],
			expectedDatasetUfsTotal: "1.00KiB",
			expectedDatasetFileNum:  "1024",
		},
	}

	for _, test := range testCases {
		err = test.engine.RestoreMetadataInternal()
		if err != nil {
			t.Errorf("fail to exec function RestoreMetadataInternal")
		}
	}
	wrappedUnhookQueryMetaDataInfoIntoFile()
}

func TestSyncMetadataInternal(t *testing.T) {
	datasetInputs := []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		},
	}
	testObjs := []runtime.Object{}
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []AlluxioEngine{
		{
			name:               "hbase",
			namespace:          "fluid",
			Client:             client,
			Log:                fake.NullLogger(),
			MetadataSyncDoneCh: make(chan base.MetadataSyncResult),
		},
		{
			name:               "spark",
			namespace:          "fluid",
			Client:             client,
			Log:                fake.NullLogger(),
			MetadataSyncDoneCh: nil,
		},
	}

	result := base.MetadataSyncResult{
		StartTime: time.Now(),
		UfsTotal:  "2GB",
		Done:      true,
		FileNum:   "5",
	}

	var testCase = []struct {
		engine           AlluxioEngine
		expectedResult   bool
		expectedUfsTotal string
		expectedFileNum  string
	}{
		{
			engine:           engines[0],
			expectedUfsTotal: "2GB",
			expectedFileNum:  "5",
		},
	}

	for index, test := range testCase {
		if index == 0 {
			go func() {
				test.engine.MetadataSyncDoneCh <- result
			}()
		}

		err := test.engine.syncMetadataInternal()
		if err != nil {
			t.Errorf("fail to exec the function with error %v", err)
		}

		key := types.NamespacedName{
			Namespace: test.engine.namespace,
			Name:      test.engine.name,
		}

		dataset := &datav1alpha1.Dataset{}
		err = client.Get(context.TODO(), key, dataset)
		if err != nil {
			t.Errorf("failt to get the dataset with error %v", err)
		}

		if dataset.Status.UfsTotal != test.expectedUfsTotal || dataset.Status.FileNum != test.expectedFileNum {
			t.Errorf("expected UfsTotal %s, get UfsTotal %s, expected FileNum %s, get FileNum %s", test.expectedUfsTotal, dataset.Status.UfsTotal, test.expectedFileNum, dataset.Status.FileNum)
		}
	}
}
