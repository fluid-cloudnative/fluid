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

package juicefs

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"

	"github.com/brahma-adshonor/gohook"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cdataload "github.com/fluid-cloudnative/fluid/pkg/dataload"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
)

var valuesConfigMapData = `
cacheDirs:
  "1":
    path: /jfs/cache3:/jfs/cache4
    type: hostPath
configs:
  accesskeySecret: jfs-secret
  accesskeySecretKey: accesskey
  bucket: http://minio.default.svc.cluster.local:9000/minio/test2
  formatCmd: /usr/local/bin/juicefs format --trash-days=0 --access-key=${ACCESS_KEY}
    --secret-key=${SECRET_KEY} --storage=minio --bucket=http://minio.default.svc.cluster.local:9000/minio/test2
    ${METAURL} minio
  metaurlSecret: jfs-secret
  metaurlSecretKey: metaurl
  name: minio
  secretkeySecret: jfs-secret
  secretkeySecretKey: secretkey
  storage: minio
edition: community
fsGroup: 0
fullnameOverride: jfsdemo
fuse:
  metricsPort: 14001
  command: /bin/mount.juicefs ${METAURL} /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse
    -o metrics=0.0.0.0:9567,cache-size=1024,free-space-ratio=0.1,cache-dir=/jfs/cache3:/jfs/cache4
  criticalPod: true
  enabled: true
  hostMountPath: /runtime-mnt/juicefs/default/jfsdemo
  hostNetwork: true
  image: registry.cn-hangzhou.aliyuncs.com/juicefs/juicefs-fuse
  imagePullPolicy: IfNotPresent
  imageTag: v1.0.0-4.8.0
  mountPath: /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse
  nodeSelector:
    fluid.io/f-default-jfsdemo: "true"
  privileged: true
  resources: {}
  statCmd: stat -c %i /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse
  volumeMounts:
  - mountPath: /jfs/cache3:/jfs/cache4
    name: cache-dir-1
  volumes:
  - hostPath:
      path: /jfs/cache3:/jfs/cache4
      type: DirectoryOrCreate
    name: cache-dir-1
group: 0
image: registry.cn-hangzhou.aliyuncs.com/juicefs/juicefs-fuse
imagePullPolicy: IfNotPresent
imagePullSecrets: null
imageTag: v1.0.0-4.8.0
owner:
  apiVersion: data.fluid.io/v1alpha1
  blockOwnerDeletion: false
  controller: true
  enabled: true
  kind: JuiceFSRuntime
  name: jfsdemo
  uid: 9ae3312b-d5b6-4a3d-895c-7712bfa7d74e
placement: Exclusive
runtimeIdentity:
  name: jfsdemo
  namespace: default
source: ${METAURL}
user: 0
worker:
  metricsPort: 14000
  command: /bin/mount.juicefs ${METAURL} /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse
    -o cache-size=1024,free-space-ratio=0.1,cache-dir=/jfs/cache1:/jfs/cache2,metrics=0.0.0.0:9567
  hostNetwork: true
  mountPath: /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse
  privileged: true
  resources: {}
  statCmd: stat -c %i /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse
  volumeMounts:
  - mountPath: /jfs/cache1:/jfs/cache2
    name: cache-dir-1
  volumes:
  - hostPath:
      path: /jfs/cache1:/jfs/cache2
      type: DirectoryOrCreate
    name: cache-dir-1
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
	statefulsetInputs := []appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "juicefs-worker",
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
	engine := &JuiceFSEngine{
		name:      "juicefs",
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

	statefulsetInputs := []appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "juicefs-worker",
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
				Namespace: "fluid",
				Labels:    map[string]string{"a": "b"},
			},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
				Conditions: []v1.PodCondition{{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				}},
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

	testCases := []struct {
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
			name:      "juicefs",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		}
		if fileName, err := engine.generateDataLoadValueFile(context, &test.dataLoad); err != nil || !strings.Contains(fileName, test.expectFileName) {
			t.Errorf("fail to generate the dataload value file: %v", err)
		}
	}
}

func TestJuiceFSEngine_GenerateDataLoadValueFileWithRuntime(t *testing.T) {
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset-juicefs-values",
			Namespace: "fluid",
		},
		Data: map[string]string{
			"data": ``,
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

	statefulsetInputs := []appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "juicefs-worker",
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
				Namespace: "fluid",
				Labels:    map[string]string{"a": "b"},
			},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
				Conditions: []v1.PodCondition{{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				}},
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
			Target: []datav1alpha1.TargetPath{{
				Path:     "/dir",
				Replicas: 1,
			}},
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

	testCases := []struct {
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
			name:      "juicefs",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		}
		if fileName, err := engine.generateDataLoadValueFile(context, &test.dataLoad); err != nil || !strings.Contains(fileName, test.expectFileName) {
			t.Errorf("fail to generate the dataload value file: %v", err)
		}
	}
}

func TestJuiceFSEngine_CheckRuntimeReady(t *testing.T) {
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
			name: "test",
			fields: fields{
				name:      "juicefs-test",
				namespace: "fluid",
			},
			sts: appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "juicefs-test-worker",
					Namespace: "fluid",
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
						Labels:    map[string]string{"a": "b"},
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
			name: "test-err",
			fields: fields{
				name:      "juicefs",
				namespace: "fluid",
			},
			sts: appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "juicefs-worker",
					Namespace: "fluid",
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
						Labels:    map[string]string{"a": "b"},
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
	}
	for _, tt := range tests {
		testObjs := []runtime.Object{}
		t.Run(tt.name, func(t *testing.T) {
			testObjs = append(testObjs, tt.sts.DeepCopy())
			testObjs = append(testObjs, tt.podList.DeepCopy())
			client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
			e := &JuiceFSEngine{
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
}

func TestJuiceFSEngine_genDataLoadValue(t *testing.T) {
	testCases := map[string]struct {
		image         string
		runtimeName   string
		targetDataset *datav1alpha1.Dataset
		dataload      *datav1alpha1.DataLoad
		cacheInfo     map[string]string
		pods          []v1.Pod
		want          *cdataload.DataLoadValue
	}{
		"test case with scheduler name": {
			image:       "fluid:v0.0.1",
			runtimeName: "juicefs",
			targetDataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataset",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							Name:       "spark",
							MountPoint: "local://mnt/data0",
							Path:       "/mnt",
						},
					},
				},
			},
			dataload: &datav1alpha1.DataLoad{
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
							Path:     "/test",
							Replicas: 1,
						},
					},
					SchedulerName: "scheduler-test",
					Options: map[string]string{
						"dl-opts-k-1": "dl-opts-v-1",
					},
				},
			},
			pods: []v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pods-1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pods-2",
					},
				},
			},
			cacheInfo: map[string]string{
				Edition:         CommunityEdition,
				"cache-info-k1": "cache-info-v1",
				"cache-info-k2": "cache-info-v2",
				"cache-info-k3": "cache-info-v3",
			},
			want: &cdataload.DataLoadValue{
				Name: "test-dataload",
				Owner: &common.OwnerReference{
					APIVersion:         "/",
					Enabled:            true,
					Name:               "test-dataload",
					BlockOwnerDeletion: false,
					Controller:         true,
				},
				DataLoadInfo: cdataload.DataLoadInfo{
					BackoffLimit:  3,
					Image:         "fluid:v0.0.1",
					TargetDataset: "test-dataset",
					SchedulerName: "scheduler-test",
					TargetPaths: []cdataload.TargetPath{
						{
							Path:     "/test",
							Replicas: 1,
						},
					},
					ImagePullSecrets: []v1.LocalObjectReference{},
					Options: map[string]string{
						// dataload spec options
						"dl-opts-k-1": "dl-opts-v-1",
						// cache info
						Edition:         CommunityEdition,
						"cache-info-k1": "cache-info-v1",
						"cache-info-k2": "cache-info-v2",
						"cache-info-k3": "cache-info-v3",
						"podNames":      "pods-1:pods-2",
						"runtimeName":   "juicefs",
						"timeout":       DefaultDataLoadTimeout,
					},
				},
			},
		},
		"test case with affinity": {
			image:       "fluid:v0.0.1",
			runtimeName: "juicefs",
			targetDataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataset",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							Name:       "spark",
							MountPoint: "local://mnt/data0",
							Path:       "/mnt",
						},
					},
				},
			},
			dataload: &datav1alpha1.DataLoad{
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
							Path:     "/test",
							Replicas: 1,
						},
					},
					Options: map[string]string{
						"dl-opts-k-1": "dl-opts-v-1",
					},
					SchedulerName: "scheduler-test",
					Affinity: &v1.Affinity{
						NodeAffinity: &v1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
								NodeSelectorTerms: []v1.NodeSelectorTerm{
									{
										MatchExpressions: []v1.NodeSelectorRequirement{
											{
												Key:      "topology.kubernetes.io/zone",
												Operator: v1.NodeSelectorOpIn,
												Values: []string{
													"antarctica-east1",
													"antarctica-west1",
												},
											},
										},
									},
								},
							},
							PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
								{
									Weight: 1,
									Preference: v1.NodeSelectorTerm{
										MatchExpressions: []v1.NodeSelectorRequirement{
											{
												Key:      "another-node-label-key",
												Operator: v1.NodeSelectorOpIn,
												Values: []string{
													"another-node-label-value",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			pods: []v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pods-1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pods-2",
					},
				},
			},
			cacheInfo: map[string]string{
				Edition:         CommunityEdition,
				"cache-info-k1": "cache-info-v1",
				"cache-info-k2": "cache-info-v2",
				"cache-info-k3": "cache-info-v3",
			},
			want: &cdataload.DataLoadValue{
				Name: "test-dataload",
				Owner: &common.OwnerReference{
					APIVersion:         "/",
					Enabled:            true,
					Name:               "test-dataload",
					BlockOwnerDeletion: false,
					Controller:         true,
				},
				DataLoadInfo: cdataload.DataLoadInfo{
					BackoffLimit:  3,
					Image:         "fluid:v0.0.1",
					TargetDataset: "test-dataset",
					SchedulerName: "scheduler-test",
					TargetPaths: []cdataload.TargetPath{
						{
							Path:     "/test",
							Replicas: 1,
						},
					},
					ImagePullSecrets: []v1.LocalObjectReference{},
					Affinity: &v1.Affinity{
						NodeAffinity: &v1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
								NodeSelectorTerms: []v1.NodeSelectorTerm{
									{
										MatchExpressions: []v1.NodeSelectorRequirement{
											{
												Key:      "topology.kubernetes.io/zone",
												Operator: v1.NodeSelectorOpIn,
												Values: []string{
													"antarctica-east1",
													"antarctica-west1",
												},
											},
										},
									},
								},
							},
							PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
								{
									Weight: 1,
									Preference: v1.NodeSelectorTerm{
										MatchExpressions: []v1.NodeSelectorRequirement{
											{
												Key:      "another-node-label-key",
												Operator: v1.NodeSelectorOpIn,
												Values: []string{
													"another-node-label-value",
												},
											},
										},
									},
								},
							},
						},
					},
					Options: map[string]string{
						// dataload spec options
						"dl-opts-k-1": "dl-opts-v-1",
						// cache info
						Edition:         CommunityEdition,
						"cache-info-k1": "cache-info-v1",
						"cache-info-k2": "cache-info-v2",
						"cache-info-k3": "cache-info-v3",
						"podNames":      "pods-1:pods-2",
						"runtimeName":   "juicefs",
						"timeout":       DefaultDataLoadTimeout,
					},
				},
			},
		},
		"test case with node selector": {
			image:       "fluid:v0.0.1",
			runtimeName: "juicefs",
			targetDataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataset",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							Name:       "spark",
							MountPoint: "local://mnt/data0",
							Path:       "/mnt",
						},
					},
				},
			},
			dataload: &datav1alpha1.DataLoad{
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
							Path:     "/test",
							Replicas: 1,
						},
					},
					SchedulerName: "scheduler-test",
					NodeSelector: map[string]string{
						"diskType": "ssd",
					},
					Options: map[string]string{
						"dl-opts-k-1": "dl-opts-v-1",
					},
				},
			},
			pods: []v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pods-1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pods-2",
					},
				},
			},
			cacheInfo: map[string]string{
				Edition:         CommunityEdition,
				"cache-info-k1": "cache-info-v1",
				"cache-info-k2": "cache-info-v2",
				"cache-info-k3": "cache-info-v3",
			},
			want: &cdataload.DataLoadValue{
				Name: "test-dataload",
				Owner: &common.OwnerReference{
					APIVersion:         "/",
					Enabled:            true,
					Name:               "test-dataload",
					BlockOwnerDeletion: false,
					Controller:         true,
				},
				DataLoadInfo: cdataload.DataLoadInfo{
					BackoffLimit:  3,
					Image:         "fluid:v0.0.1",
					TargetDataset: "test-dataset",
					SchedulerName: "scheduler-test",
					TargetPaths: []cdataload.TargetPath{
						{
							Path:     "/test",
							Replicas: 1,
						},
					},
					ImagePullSecrets: []v1.LocalObjectReference{},
					NodeSelector: map[string]string{
						"diskType": "ssd",
					},
					Options: map[string]string{
						// dataload spec options
						"dl-opts-k-1": "dl-opts-v-1",
						// cache info
						Edition:         CommunityEdition,
						"cache-info-k1": "cache-info-v1",
						"cache-info-k2": "cache-info-v2",
						"cache-info-k3": "cache-info-v3",
						"podNames":      "pods-1:pods-2",
						"runtimeName":   "juicefs",
						"timeout":       DefaultDataLoadTimeout,
					},
				},
			},
		},
		"test case with tolerations": {
			image:       "fluid:v0.0.1",
			runtimeName: "juicefs",
			targetDataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataset",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							Name:       "spark",
							MountPoint: "local://mnt/data0",
							Path:       "/mnt",
						},
					},
				},
			},
			dataload: &datav1alpha1.DataLoad{
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
							Path:     "/test",
							Replicas: 1,
						},
					},
					SchedulerName: "scheduler-test",
					Tolerations: []v1.Toleration{
						{
							Key:      "example-key",
							Operator: v1.TolerationOpExists,
							Effect:   v1.TaintEffectNoSchedule,
						},
					},
					Options: map[string]string{
						"dl-opts-k-1": "dl-opts-v-1",
					},
				},
			},
			pods: []v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pods-1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pods-2",
					},
				},
			},
			cacheInfo: map[string]string{
				Edition:         CommunityEdition,
				"cache-info-k1": "cache-info-v1",
				"cache-info-k2": "cache-info-v2",
				"cache-info-k3": "cache-info-v3",
			},
			want: &cdataload.DataLoadValue{
				Name: "test-dataload",
				Owner: &common.OwnerReference{
					APIVersion:         "/",
					Enabled:            true,
					Name:               "test-dataload",
					BlockOwnerDeletion: false,
					Controller:         true,
				},
				DataLoadInfo: cdataload.DataLoadInfo{
					BackoffLimit:  3,
					Image:         "fluid:v0.0.1",
					TargetDataset: "test-dataset",
					SchedulerName: "scheduler-test",
					TargetPaths: []cdataload.TargetPath{
						{
							Path:     "/test",
							Replicas: 1,
						},
					},
					ImagePullSecrets: []v1.LocalObjectReference{},
					Tolerations: []v1.Toleration{
						{
							Key:      "example-key",
							Operator: v1.TolerationOpExists,
							Effect:   v1.TaintEffectNoSchedule,
						},
					},
					Options: map[string]string{
						// dataload spec options
						"dl-opts-k-1": "dl-opts-v-1",
						// cache info
						Edition:         CommunityEdition,
						"cache-info-k1": "cache-info-v1",
						"cache-info-k2": "cache-info-v2",
						"cache-info-k3": "cache-info-v3",
						"podNames":      "pods-1:pods-2",
						"runtimeName":   "juicefs",
						"timeout":       DefaultDataLoadTimeout,
					},
				},
			},
		},
	}

	for k, item := range testCases {
		engine := JuiceFSEngine{
			namespace: "fluid",
			name:      item.runtimeName,
			Log:       fake.NullLogger(),
		}
		got := engine.genDataLoadValue(item.image, item.cacheInfo, item.pods, item.targetDataset, item.dataload)
		if !reflect.DeepEqual(got, item.want) {
			t.Errorf("case %s, got %v,want:%v", k, got, item.want)
		}
	}
}
