/*
  Copyright 2023 The Fluid Authors.

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

package thin

import (
	"reflect"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/brahma-adshonor/gohook"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func mockExecCommandInContainerForTotalFileNums() (stdout string, stderr string, err error) {
	r := `1331167`
	return r, "", nil
}

func mockExecCommandInContainerForUsedStorageBytes() (stdout string, stderr string, err error) {
	r := `nfs:test 207300683100160  41460043776 207259223056384   1% /data`
	return r, "", nil
}

func TestTotalStorageBytes(t *testing.T) {
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-fuse",
			Namespace: "fluid",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"a": "b"},
			},
		},
	}
	var pod = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-fuse-0",
			Namespace: "fluid",
			Labels:    map[string]string{"a": "b"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			Conditions: []corev1.PodCondition{{
				Type:   corev1.PodReady,
				Status: corev1.ConditionTrue,
			}},
		},
	}

	podList := &corev1.PodList{
		Items: []corev1.Pod{*pod},
	}
	runtimeObjs := []runtime.Object{}
	runtimeObjs = append(runtimeObjs, ds, pod)
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(appsv1.SchemeGroupVersion, ds)
	scheme.AddKnownTypes(corev1.SchemeGroupVersion, pod)
	scheme.AddKnownTypes(corev1.SchemeGroupVersion, podList)
	fakeClient := fake.NewFakeClientWithScheme(scheme, runtimeObjs...)
	type fields struct {
		runtime   *datav1alpha1.ThinRuntime
		name      string
		namespace string
	}
	tests := []struct {
		name      string
		fields    fields
		wantValue int64
		wantErr   bool
	}{
		{
			name: "test",
			fields: fields{
				name:      "test",
				namespace: "fluid",
				runtime: &datav1alpha1.ThinRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "fluid",
					},
				},
			},
			wantValue: 41460043776,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := ThinEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    fakeClient,
				Log:       fake.NullLogger(),
			}
			patch1 := ApplyFunc(kubeclient.ExecCommandInContainer, func(podName string, containerName string, namespace string, cmd []string) (string, string, error) {
				stdout, stderr, err := mockExecCommandInContainerForUsedStorageBytes()
				return stdout, stderr, err
			})
			defer patch1.Reset()
			gotValue, err := e.TotalStorageBytes()
			if (err != nil) != tt.wantErr {
				t.Errorf("ThinEngine.TotalStorageBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValue != tt.wantValue {
				t.Errorf("ThinEngine.TotalStorageBytes() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestTotalFileNums(t *testing.T) {
	statefulSet := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-fuse",
			Namespace: "fluid",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"a": "b"},
			},
		},
	}
	var pod = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-fuse-0",
			Namespace: "fluid",
			Labels:    map[string]string{"a": "b"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			Conditions: []corev1.PodCondition{{
				Type:   corev1.PodReady,
				Status: corev1.ConditionTrue,
			}},
		},
	}
	runtimeObjs := []runtime.Object{}
	runtimeObjs = append(runtimeObjs, statefulSet, pod)
	fakeClient := fake.NewFakeClientWithScheme(testScheme, runtimeObjs...)
	type fields struct {
		runtime   *datav1alpha1.ThinRuntime
		name      string
		namespace string
	}
	tests := []struct {
		name      string
		fields    fields
		wantValue int64
		wantErr   bool
	}{
		{
			name: "test",
			fields: fields{
				name:      "test",
				namespace: "fluid",
				runtime: &datav1alpha1.ThinRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name: "spark",
					},
				},
			},
			wantValue: 1331167,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := ThinEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       fake.NullLogger(),
				Client:    fakeClient,
			}
			patch1 := ApplyFunc(kubeclient.ExecCommandInContainer, func(podName string, containerName string, namespace string, cmd []string) (string, string, error) {
				stdout, stderr, err := mockExecCommandInContainerForTotalFileNums()
				return stdout, stderr, err
			})
			defer patch1.Reset()
			gotValue, err := e.TotalFileNums()
			if (err != nil) != tt.wantErr {
				t.Errorf("ThinEngine.TotalFileNums() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValue != tt.wantValue {
				t.Errorf("ThinEngine.TotalFileNums() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestThinEngine_FreeStorageBytes(t *testing.T) {
	type fields struct{}
	tests := []struct {
		name    string
		fields  fields
		want    int64
		wantErr bool
	}{
		{
			name:    "test",
			fields:  fields{},
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := ThinEngine{}
			got, err := j.FreeStorageBytes()
			if (err != nil) != tt.wantErr {
				t.Errorf("FreeStorageBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FreeStorageBytes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThinEngine_PrepareUFS(t *testing.T) {
	type fields struct{}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "test",
			fields:  fields{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := ThinEngine{}
			if err := j.PrepareUFS(); (err != nil) != tt.wantErr {
				t.Errorf("PrepareUFS() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestThinEngine_ShouldCheckUFS(t *testing.T) {
	type fields struct {
	}
	tests := []struct {
		name       string
		fields     fields
		wantShould bool
		wantErr    bool
	}{
		{
			name:       "test",
			fields:     fields{},
			wantShould: false,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := ThinEngine{}
			gotShould, err := j.ShouldCheckUFS()
			if (err != nil) != tt.wantErr {
				t.Errorf("ShouldCheckUFS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotShould != tt.wantShould {
				t.Errorf("ShouldCheckUFS() gotShould = %v, want %v", gotShould, tt.wantShould)
			}
		})
	}
}

func TestThinEngine_ShouldUpdateUFS(t *testing.T) {
	type fields struct{}
	tests := []struct {
		name            string
		fields          fields
		wantUfsToUpdate *utils.UFSToUpdate
	}{
		{
			name:            "test",
			fields:          fields{},
			wantUfsToUpdate: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := ThinEngine{Client: fake.NewFakeClient()}
			if gotUfsToUpdate := j.ShouldUpdateUFS(); !reflect.DeepEqual(gotUfsToUpdate, tt.wantUfsToUpdate) {
				t.Errorf("ShouldUpdateUFS() = %v, want %v", gotUfsToUpdate, tt.wantUfsToUpdate)
			}
		})
	}
}

func TestThinEngine_UpdateOnUFSChange(t *testing.T) {
	type fields struct{}
	type args struct {
		ufsToUpdate *utils.UFSToUpdate
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantReady bool
		wantErr   bool
	}{
		{
			name:      "test",
			fields:    fields{},
			args:      args{},
			wantReady: true,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := ThinEngine{}
			gotReady, err := j.UpdateOnUFSChange(tt.args.ufsToUpdate)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateOnUFSChange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotReady != tt.wantReady {
				t.Errorf("UpdateOnUFSChange() gotReady = %v, want %v", gotReady, tt.wantReady)
			}
		})
	}
}

func TestThinEngine_UsedStorageBytes(t *testing.T) {
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-fuse",
			Namespace: "thin",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"a": "b"},
			},
		},
		Status: appsv1.DaemonSetStatus{},
	}

	var pod = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-fuse-0",
			Namespace: "thin",
			Labels:    map[string]string{"a": "b"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			Conditions: []corev1.PodCondition{{
				Type:   corev1.PodReady,
				Status: corev1.ConditionTrue,
			}},
		},
	}
	podList := &corev1.PodList{
		Items: []corev1.Pod{*pod},
	}
	runtimeObjs := []runtime.Object{}
	runtimeObjs = append(runtimeObjs, ds, pod)
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(appsv1.SchemeGroupVersion, ds)
	scheme.AddKnownTypes(corev1.SchemeGroupVersion, pod)
	scheme.AddKnownTypes(corev1.SchemeGroupVersion, podList)
	fakeClient := fake.NewFakeClientWithScheme(scheme, runtimeObjs...)
	type fields struct {
		runtime   *datav1alpha1.ThinRuntime
		name      string
		namespace string
	}
	tests := []struct {
		name      string
		fields    fields
		wantValue int64
		wantErr   bool
	}{
		{
			name: "test",
			fields: fields{
				name:      "test",
				namespace: "thin",
				runtime: &datav1alpha1.ThinRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name: "spark",
					},
				},
			},
			wantValue: 41460043776,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := ThinEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       fake.NullLogger(),
				Client:    fakeClient,
			}
			patch1 := ApplyFunc(kubeclient.ExecCommandInContainer, func(podName string, containerName string, namespace string, cmd []string) (string, string, error) {
				stdout, stderr, err := mockExecCommandInContainerForUsedStorageBytes()
				return stdout, stderr, err
			})
			defer patch1.Reset()
			gotValue, err := j.UsedStorageBytes()
			if (err != nil) != tt.wantErr {
				t.Errorf("ThinEngine.UsedStorageBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValue != tt.wantValue {
				t.Errorf("ThinEngine.UsedStorageBytes() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestThinEngine_updateFuseConfigOnChange(t *testing.T) {
	mockUpdate := func(client client.Client, cm *corev1.ConfigMap) error {
		return nil
	}
	err := gohook.Hook(kubeclient.UpdateConfigMap, mockUpdate, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	wrappedUnhook := func() {
		err := gohook.UnHook(kubeclient.UpdateConfigMap)
		if err != nil {
			t.Fatal(err.Error())
		}
	}
	defer wrappedUnhook()

	dataset := datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{
				{
					MountPoint: "https://mirrors.bit.edu.cn/apache/zookeeper/stable/",
					Name:       "zookeeper",
				},
			},
		},
	}
	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-fuse-conf",
			Namespace: "default",
		},
		Data: map[string]string{
			"config.json": "{\"mounts\":[{\"mountPoint\":\"https://mirrors.bit.edu.cn/apache/zookeeper/stable/\",\"name\":\"zookeeper\"}],\"targetPath\":\"/thin/default/test/thin-fuse\"}",
		},
	}
	newDataset := datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{
				{
					MountPoint: "https://mirrors.bit.edu.cn/apache/zookeeper/stable/",
					Name:       "zookeeper",
				},
				{
					MountPoint: "https://mirrors.bit.edu.cn/apache/hbase/stable/",
					Name:       "hbase",
				},
			},
		},
	}
	thinruntime := datav1alpha1.ThinRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	testcase := []struct {
		name       string
		dataset    *datav1alpha1.Dataset
		cm         *corev1.ConfigMap
		wantUpdate bool
		wantErr    bool
	}{
		{
			name:       "not update configmap",
			dataset:    &dataset,
			cm:         &cm,
			wantUpdate: false,
			wantErr:    false,
		},
		{
			name:       "update configmap",
			dataset:    &newDataset,
			cm:         &cm,
			wantUpdate: true,
			wantErr:    false,
		},
		{
			name:       "no configmap",
			dataset:    &datav1alpha1.Dataset{},
			cm:         &corev1.ConfigMap{},
			wantUpdate: false,
			wantErr:    false,
		},
	}

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = datav1alpha1.AddToScheme(scheme)

	for _, tt := range testcase {
		client := fake.NewFakeClientWithScheme(scheme, tt.dataset, tt.cm)
		thinEngine := ThinEngine{
			Client:    client,
			name:      "test",
			namespace: "default",
			runtime:   &thinruntime,
			Log:       fake.NullLogger(),
		}
		update, err := thinEngine.updateFuseConfigOnChange(thinEngine.runtime, tt.dataset)
		if (err != nil) != tt.wantErr {
			t.Errorf("testcase %s failed due to error %v", tt.name, err)
		}
		if update != tt.wantUpdate {
			t.Errorf("testcase %s failed, want update %v, get update %v", tt.name, tt.wantUpdate, update)
		}
	}
}
