/*
  Copyright 2022 The Fluid Authors.

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

package efc

import (
	"errors"
	"testing"

	"github.com/brahma-adshonor/gohook"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/efc/operations"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	utilpointer "k8s.io/utils/pointer"
)

func TestEFCEngine_CreateDataLoadJob(t *testing.T) {
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "efcdemo-efc-values",
			Namespace: "fluid",
		},
		Data: map[string]string{
			"data": valuesConfigMapData,
		},
	}

	mockExecCheckReleaseCommon := func(name string, namespace string) (exist bool, err error) {
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
			Name:      "efcdemo-dataload",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.DataLoadSpec{
			Dataset: datav1alpha1.TargetDataset{
				Name:      "efcdemo",
				Namespace: "fluid",
			},
		},
	}
	datasetInputs := []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "efcdemo",
				Namespace: "fluid",
			},
		},
	}
	statefulsetInputs := []appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "efcdemo-worker",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"a": "b"},
				},
			},
		},
	}
	podListInputs := []v1.PodList{{
		Items: []v1.Pod{{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"a": "b"},
			},
		}},
	}}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, configMap)
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	for _, statefulsetInput := range statefulsetInputs {
		testObjs = append(testObjs, statefulsetInput.DeepCopy())
	}
	for _, podInput := range podListInputs {
		testObjs = append(testObjs, podInput.DeepCopy())
	}
	testScheme.AddKnownTypes(v1.SchemeGroupVersion, configMap)
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
	engine := &EFCEngine{
		name:      "efcdemo",
		namespace: "fluid",
		Client:    client,
		Log:       fake.NullLogger(),
	}
	ctx := cruntime.ReconcileRequestContext{
		Log:      fake.NullLogger(),
		Client:   client,
		Recorder: record.NewFakeRecorder(1),
	}

	err := gohook.Hook(helm.CheckRelease, mockExecCheckReleaseErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.CreateDataLoadJob(ctx, targetDataLoad)
	if err == nil {
		t.Errorf("fail to catch the error: %v", err)
	}
	wrappedUnhookCheckRelease()

	err = gohook.Hook(helm.CheckRelease, mockExecCheckReleaseCommon, nil)
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
		t.Errorf("fail to exec the function: %v", err)
	}
	wrappedUnhookCheckRelease()
}

func TestEFCEngine_GenerateDataLoadValueFileWithRuntimeHDD(t *testing.T) {
}

func TestEFCEngine_GenerateDataLoadValueFileWithRuntime(t *testing.T) {
}

func TestEFCEngine_CheckRuntimeReady(t *testing.T) {
	type fields struct {
		name      string
		namespace string
	}
	tests := []struct {
		name      string
		fields    fields
		sts       appsv1.StatefulSet
		podList   v1.PodList
		wantReady bool
	}{
		{
			name: "efc-test",
			fields: fields{
				name:      "efc-test",
				namespace: "fluid",
			},
			sts: appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "efc-test-worker",
					Namespace: "fluid",
					UID:       "uid1",
				},
				Spec: appsv1.StatefulSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"a": "b"},
					},
				},
			},
			podList: v1.PodList{
				Items: []v1.Pod{{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "fluid",
						Name:      "efc-test-worker-0",
						Labels:    map[string]string{"a": "b"},
						OwnerReferences: []metav1.OwnerReference{{
							Kind:       "StatefulSet",
							APIVersion: "apps/v1",
							Name:       "efc-test-worker",
							UID:        "uid1",
							Controller: utilpointer.BoolPtr(true),
						}},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
						Conditions: []v1.PodCondition{{
							Type:   v1.PodReady,
							Status: v1.ConditionTrue,
						}},
					},
				}},
			},
			wantReady: true,
		},
		{
			name: "efc-test-err",
			fields: fields{
				name:      "efc-test-err",
				namespace: "fluid",
			},
			sts: appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "efc-test-err-worker",
					Namespace: "fluid",
					UID:       "uid2",
				},
				Spec: appsv1.StatefulSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"a": "b"},
					},
				},
			},
			podList: v1.PodList{
				Items: []v1.Pod{{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "fluid",
						Name:      "efc-test-err-worker-0",
						Labels:    map[string]string{"a": "b"},
						OwnerReferences: []metav1.OwnerReference{{
							Kind:       "StatefulSet",
							APIVersion: "apps/v1",
							Name:       "efc-test-err-worker",
							UID:        "uid2",
							Controller: utilpointer.BoolPtr(true),
						}},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
						Conditions: []v1.PodCondition{{
							Type:   v1.PodReady,
							Status: v1.ConditionFalse,
						}},
					},
				}},
			},
			wantReady: false,
		},
		{
			name: "efc-test-err2",
			fields: fields{
				name:      "efc-test-err2",
				namespace: "fluid",
			},
			sts: appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "efc-test-err2-worker",
					Namespace: "fluid",
					UID:       "uid3",
				},
				Spec: appsv1.StatefulSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"a": "b"},
					},
				},
			},
			podList: v1.PodList{
				Items: []v1.Pod{{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "fluid",
						Name:      "efc-test-err2-worker-0",
						Labels:    map[string]string{"nota": "notb"},
						OwnerReferences: []metav1.OwnerReference{{
							Kind:       "StatefulSet",
							APIVersion: "apps/v1",
							Name:       "efc-test-err2-worker",
							UID:        "uid3",
							Controller: utilpointer.BoolPtr(true),
						}},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
						Conditions: []v1.PodCondition{{
							Type:   v1.PodReady,
							Status: v1.ConditionTrue,
						}},
					},
				}},
			},
			wantReady: false,
		},
	}

	ReadyCommon := func(a operations.EFCFileUtils) (ready bool) {
		return true
	}
	wrappedUnhookReady := func() {
		err := gohook.UnHook(operations.EFCFileUtils.Ready)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(operations.EFCFileUtils.Ready, ReadyCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, tt := range tests {
		testObjs := []runtime.Object{}
		t.Run(tt.name, func(t *testing.T) {
			testObjs = append(testObjs, tt.sts.DeepCopy())
			testObjs = append(testObjs, tt.podList.DeepCopy())
			client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
			e := &EFCEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    client,
				Log:       fake.NullLogger(),
			}
			if gotReady := e.CheckRuntimeReady(); gotReady != tt.wantReady {
				t.Errorf("CheckRuntimeReady() = %v, want %v", gotReady, tt.wantReady)
			}
		})
	}

	wrappedUnhookReady()
}
