package alluxio

import (
	"context"
	"errors"
	"github.com/brahma-adshonor/gohook"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
	"time"
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
			Log:       log.NullLogger{},
		},
		{
			name:      "spark",
			namespace: "fluid",
			Client:    client,
			Log:       log.NullLogger{},
		},
	}

	for _, engine := range engines {
		err := engine.SyncMetadata()
		if err != nil {
			t.Errorf("fail to exec the function")
		}
	}

	engine := AlluxioEngine{
		name:      "hadoop",
		namespace: "fluid",
		Client:    client,
		Log:       log.NullLogger{},
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

func TestShouldSyncMetadata(t *testing.T) {
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
			Log:       log.NullLogger{},
		},
		{
			name:      "spark",
			namespace: "fluid",
			Client:    client,
			Log:       log.NullLogger{},
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
	}

	for _, test := range testCases {
		should, err := test.engine.shouldSyncMetadata()
		if err != nil || should != test.expectedShould {
			t.Errorf("fail to exec the function")
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
			Log:       log.NullLogger{},
		},
		{
			name:      "spark",
			namespace: "fluid",
			Client:    client,
			Log:       log.NullLogger{},
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
			Log:       log.NullLogger{},
		},
		{
			name:      "hbase",
			namespace: "fluid",
			Client:    client,
			Log:       log.NullLogger{},
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
			name:               "spark",
			namespace:          "fluid",
			Client:             client,
			Log:                log.NullLogger{},
			MetadataSyncDoneCh: nil,
		},
		{
			name:               "hbase",
			namespace:          "fluid",
			Client:             client,
			Log:                log.NullLogger{},
			MetadataSyncDoneCh: make(chan MetadataSyncResult),
		},
	}

	result := MetadataSyncResult{
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
			expectedUfsTotal: METADATA_SYNC_NOT_DONE_MSG,
			expectedFileNum:  METADATA_SYNC_NOT_DONE_MSG,
		},
		{
			engine:           engines[1],
			expectedUfsTotal: "2GB",
			expectedFileNum:  "5",
		},
	}

	for index, test := range testCase {
		if index == 1 {
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