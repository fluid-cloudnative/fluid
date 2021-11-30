package jindo

import (
	"testing"

	"github.com/brahma-adshonor/gohook"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubectl"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/net"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestSetupMasterInternal(t *testing.T) {
	mockExecCreateConfigMapFromFileCommon := func(name string, key, fileName string, namespace string) (err error) {
		return nil
	}
	mockExecCreateConfigMapFromFileErr := func(name string, key, fileName string, namespace string) (err error) {
		return errors.New("fail to exec command")
	}
	mockExecCheckReleaseCommonFound := func(name string, namespace string) (exist bool, err error) {
		return true, nil
	}
	mockExecCheckReleaseCommonNotFound := func(name string, namespace string) (exist bool, err error) {
		return false, nil
	}
	mockExecCheckReleaseErr := func(name string, namespace string) (exist bool, err error) {
		return false, errors.New("fail to check release")
	}
	mockExecInstallReleaseCommon := func(name string, namespace string, valueFile string, chartName string) error {
		return nil
	}
	mockExecInstallReleaseErr := func(name string, namespace string, valueFile string, chartName string) error {
		return errors.New("fail to install dataload chart")
	}

	wrappedUnhookCreateConfigMapFromFile := func() {
		err := gohook.UnHook(kubectl.CreateConfigMapFromFile)
		if err != nil {
			t.Fatal(err.Error())
		}
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

	allixioruntime := &datav1alpha1.JindoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
	}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, (*allixioruntime).DeepCopy())

	datasetInputs := []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
		},
	}
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engine := JindoEngine{
		name:      "hbase",
		namespace: "fluid",
		Client:    client,
		Log:       log.NullLogger{},
		runtime: &datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Master: datav1alpha1.JindoCompTemplateSpec{
					Replicas: 2,
				},
			},
		},
	}
	portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 100}, GetReservedPorts)
	err := gohook.Hook(kubectl.CreateConfigMapFromFile, mockExecCreateConfigMapFromFileErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.setupMasterInernal()
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookCreateConfigMapFromFile()

	// create configmap successfully
	err = gohook.Hook(kubectl.CreateConfigMapFromFile, mockExecCreateConfigMapFromFileCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	// check release found
	err = gohook.Hook(helm.CheckRelease, mockExecCheckReleaseCommonFound, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_ = engine.setupMasterInernal()
	wrappedUnhookCheckRelease()

	// check release error
	err = gohook.Hook(helm.CheckRelease, mockExecCheckReleaseErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.setupMasterInernal()
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookCheckRelease()

	// check release not found
	err = gohook.Hook(helm.CheckRelease, mockExecCheckReleaseCommonNotFound, nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	// install release with error
	err = gohook.Hook(helm.InstallRelease, mockExecInstallReleaseErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.setupMasterInernal()
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookInstallRelease()

	// install release successfully
	err = gohook.Hook(helm.InstallRelease, mockExecInstallReleaseCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_ = engine.setupMasterInernal()
	wrappedUnhookInstallRelease()
	wrappedUnhookCreateConfigMapFromFile()
}

func TestGenerateJindoValueFile(t *testing.T) {
	mockExecCreateConfigMapFromFileCommon := func(name string, key, fileName string, namespace string) (err error) {
		return nil
	}
	mockExecCreateConfigMapFromFileErr := func(name string, key, fileName string, namespace string) (err error) {
		return errors.New("fail to exec command")
	}

	wrappedUnhookCreateConfigMapFromFile := func() {
		err := gohook.UnHook(kubectl.CreateConfigMapFromFile)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	jindoruntime := &datav1alpha1.JindoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
	}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, (*jindoruntime).DeepCopy())

	datasetInputs := []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
		},
	}
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engine := JindoEngine{
		name:      "hbase",
		namespace: "fluid",
		Client:    client,
		Log:       log.NullLogger{},
		runtime: &datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Master: datav1alpha1.JindoCompTemplateSpec{
					Replicas: 2,
				},
			},
		},
	}

	portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 50}, GetReservedPorts)
	err := gohook.Hook(kubectl.CreateConfigMapFromFile, mockExecCreateConfigMapFromFileErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = engine.generateJindoValueFile()
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookCreateConfigMapFromFile()

	err = gohook.Hook(kubectl.CreateConfigMapFromFile, mockExecCreateConfigMapFromFileCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	wrappedUnhookCreateConfigMapFromFile()
}

func TestGetConfigmapName(t *testing.T) {
	engine := JindoEngine{
		name:        "hbase",
		runtimeType: "Jindo",
	}
	expectedResult := "hbase-Jindo-values"
	if engine.getConfigmapName() != expectedResult {
		t.Errorf("fail to get the configmap name")
	}
}
