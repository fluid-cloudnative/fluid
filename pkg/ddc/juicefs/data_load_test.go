package juicefs

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"

	"github.com/brahma-adshonor/gohook"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/log"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
)

var valuesConfigMapData = `
fullnameOverride: test-dataset
image: juicedata/juicefs-csi-driver
imageTag: v0.11.0
imagePullPolicy: IfNotPresent
user: 0
group: 0
fsGroup: 0
fuse:
  prepare:
    subPath: /dir1
    name: pics
    accesskeySecret: juicefs-secret
    secretkeySecret: juicefs-secret
    bucket: http://xx.xx.xx.xx/pics
    metaurlSecret: juicefs-secret
    storage: minio
  image: juicedata/juicefs-csi-driver
  nodeSelector:
    fluid.io/f-fluid-test-dataset: "true"
  imageTag: v0.11.0
  mountPath: /runtime-mnt/juicefs/fluid/test-dataset/juicefs-fuse
  cacheDir: /tmp/jfs-cache
  hostMountPath: /runtime-mnt/juicefs/fluid/test-dataset
  command: /bin/mount.juicefs redis://xx.xx.xx.xx:6379/1 /runtime-mnt/juicefs/fluid/test-dataset/juicefs-fuse -o metrics=0.0.0.0:9567,subdir=/dir1,cache-size=4096,free-space-ratio=0.1,cache-dir=/tmp/jfs-cache
  statCmd: stat -c %i /runtime-mnt/juicefs/fluid/test-dataset/juicefs-fuse
  enabled: true
  criticalPod: true
worker:
  cacheDir: /tmp/jfs-cache
placement: Exclusive

Events:  <none>
`

func TestJuiceFSEngine_CreateDataLoadJob(t *testing.T) {
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset-juicefs-values",
			Namespace: "fluid",
		},
		Data: map[string]string{
			"data": valuesConfigMapData,
		},
	}

	mockExecCheckReleaseCommon := func(name string, namespace string) (exist bool, err error) {
		return false, nil
	}
	//mockExecCheckReleaseErr := func(name string, namespace string) (exist bool, err error) {
	//	return false, errors.New("fail to check release")
	//}
	mockExecInstallReleaseCommon := func(name string, namespace string, valueFile string, chartName string) error {
		return nil
	}
	mockExecInstallReleaseErr := func(name string, namespace string, valueFile string, chartName string) error {
		return errors.New("fail to install dataload chart")
	}

	wrappedUnhookCheckRelease := func() {
		err := gohook.UnHook(helm.CheckRelease)
		if err != nil {
			t.Fatal(err.Error())
		}
	}
	wrappedUnhookInstallRelease := func() {
		err := gohook.UnHook(helm.InstallRelease)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	targetDataLoad := datav1alpha1.DataLoad{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.DataLoadSpec{
			Dataset: datav1alpha1.TargetDataset{
				Name:      "test-dataset",
				Namespace: "fluid",
			},
		},
	}
	datasetInputs := []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-dataset",
				Namespace: "fluid",
			},
		},
	}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, configMap)
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	testScheme.AddKnownTypes(v1.SchemeGroupVersion, configMap)
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
	engine := &JuiceFSEngine{
		name:   "hbase",
		Client: client,
	}
	ctx := cruntime.ReconcileRequestContext{
		Log:      log.NullLogger{},
		Client:   client,
		Recorder: record.NewFakeRecorder(1),
	}

	err := gohook.Hook(helm.CheckRelease, mockExecCheckReleaseCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook(helm.InstallRelease, mockExecInstallReleaseErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.CreateDataLoadJob(ctx, targetDataLoad)
	if err == nil {
		t.Errorf("fail to catch the error: %v", err)
	}
	wrappedUnhookInstallRelease()

	err = gohook.Hook(helm.InstallRelease, mockExecInstallReleaseCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.CreateDataLoadJob(ctx, targetDataLoad)
	if err != nil {
		t.Errorf("fail to exec the function")
	}
	wrappedUnhookCheckRelease()
}

func TestJuiceFSEngine_GenerateDataLoadValueFileWithRuntimeHDD(t *testing.T) {
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset-juicefs-values",
			Namespace: "fluid",
		},
		Data: map[string]string{
			"data": valuesConfigMapData,
		},
	}

	datasetInputs := []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-dataset",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		},
	}

	testObjs := []runtime.Object{}
	testObjs = append(testObjs, configMap)
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	context := cruntime.ReconcileRequestContext{
		Client: client,
	}

	dataLoadNoTarget := datav1alpha1.DataLoad{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataload",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.DataLoadSpec{
			Dataset: datav1alpha1.TargetDataset{
				Name:      "test-dataset",
				Namespace: "fluid",
			},
		},
	}
	dataLoadWithTarget := datav1alpha1.DataLoad{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataload",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.DataLoadSpec{
			Dataset: datav1alpha1.TargetDataset{
				Name:      "test-dataset",
				Namespace: "fluid",
			},
			Target: []datav1alpha1.TargetPath{
				{
					Path: "/test",
				},
			},
		},
	}

	var testCases = []struct {
		dataLoad       datav1alpha1.DataLoad
		expectFileName string
	}{
		{
			dataLoad:       dataLoadNoTarget,
			expectFileName: filepath.Join(os.TempDir(), "fluid-test-dataload-loader-values.yaml"),
		},
		{
			dataLoad:       dataLoadWithTarget,
			expectFileName: filepath.Join(os.TempDir(), "fluid-test-dataload-loader-values.yaml"),
		},
	}

	for _, test := range testCases {
		engine := JuiceFSEngine{
			Client: client,
		}
		if fileName, _ := engine.generateDataLoadValueFile(context, test.dataLoad); !strings.Contains(fileName, test.expectFileName) {
			t.Errorf("fail to generate the dataload value file")
		}
	}
}

func TestJuiceFSEngine_CheckExistenceOfPath(t *testing.T) {
	mockExecNotExist := func(podName string, containerName string, namespace string, cmd []string) (stdout string, stderr string, e error) {
		return "does not exist", "", errors.New("other error")
	}
	wrappedUnhook := func() {
		err := gohook.UnHook(kubeclient.ExecCommandInContainer)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	engine := JuiceFSEngine{
		namespace: "fluid",
		name:      "hbase",
		Log:       log.NullLogger{},
	}

	err := gohook.Hook(kubeclient.ExecCommandInContainer, mockExecNotExist, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	targetDataload := datav1alpha1.DataLoad{
		Spec: datav1alpha1.DataLoadSpec{
			Target: []datav1alpha1.TargetPath{
				{
					Path:     "/tmp",
					Replicas: 1,
				},
			},
		},
	}
	notExist, err := engine.CheckExistenceOfPath(targetDataload)
	if !(err == nil && notExist == true) {
		t.Errorf("fail to exec the function")
	}
	wrappedUnhook()
}
