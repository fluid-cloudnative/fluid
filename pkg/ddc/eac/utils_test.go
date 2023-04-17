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

package eac

import (
	"fmt"
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	ctrlhelper "github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilpointer "k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var valuesConfigMapData = `
fullnameOverride: efcdemo
placement: Exclusive
master:
  image: registry.cn-zhangjiakou.aliyuncs.com/nascache/eac-fuse
  imageTag: update
  imagePullPolicy: IfNotPresent
  imagePullSecrets: []
  mountPoint: 123456-abcd.cn-zhangjiakou.nas.aliyuncs.com:/test-fluid-3/
  count: 1
  enabled: true
  option: client_owner=default-efcdemo-master,assign_uuid=default-efcdemo-master,g_tier_EnableDadi=true,g_tier_DadiEnablePrefetch=true
  hostNetwork: true
  tieredstore:
    levels:
    - level: 0
      mediumtype: MEM
      type: emptyDir
      path: /dev/shm
worker:
  image: registry.cn-zhangjiakou.aliyuncs.com/nascache/eac-worker
  imageTag: update
  imagePullPolicy: IfNotPresent
  imagePullSecrets: []
  port:
    rpc: 17673
  enabled: true
  option: cache_capacity_gb=2,cache_media=/dev/efc-worker-cache-path/default/efcdemo,server_port=17673
  resources:
    requests:
      memory: 1953125Ki
  hostNetwork: true
  tieredstore:
    levels:
    - alias: SSD
      level: 0
      mediumtype: SSD
      type: emptyDir
      path: /dev/efc-worker-cache-path/default/efcdemo
      quota: 2GB
fuse:
  image: registry.cn-zhangjiakou.aliyuncs.com/nascache/eac-fuse
  imageTag: update
  imagePullPolicy: IfNotPresent
  imagePullSecrets: []
  mountPoint: 123456-abcd.cn-zhangjiakou.nas.aliyuncs.com:/test-fluid-3/
  hostMountPath: /runtime-mnt/efc/default/efcdemo
  port:
    monitor: 17645
  option: assign_uuid=default-efcdemo-fuse,g_tier_EnableDadi=true,g_tier_DadiEnablePrefetch=true
  nodeSelector:
    fluid.io/f-default-efcdemo: "true"
  hostNetwork: true
  tieredstore:
    levels:
    - level: 0
      mediumtype: MEM
      type: emptyDir
      path: /dev/shm
  criticalPod: true
initFuse:
  image: registry.cn-zhangjiakou.aliyuncs.com/nascache/init-alifuse
  imageTag: update
  imagePullPolicy: IfNotPresent
  imagePullSecrets: []
osAdvise:
  osVersion: centos
  enabled: true
`

var workerEndpointsConfigMapData = `
{"containerendpoints":[]}
`

func newEACEngine(client client.Client, name string, namespace string) *EACEngine {
	runTimeInfo, _ := base.BuildRuntimeInfo(name, namespace, common.EFCRuntime, datav1alpha1.TieredStore{})
	engine := &EACEngine{
		runtime:     &datav1alpha1.EFCRuntime{},
		name:        name,
		namespace:   namespace,
		Client:      client,
		runtimeInfo: runTimeInfo,
		Log:         fake.NullLogger(),
	}
	engine.Helper = ctrlhelper.BuildHelper(runTimeInfo, client, engine.Log)
	return engine
}

func Test_parsePortsFromConfigMap(t *testing.T) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase-efc-values",
			Namespace: "fluid",
		},
		Data: map[string]string{
			"data": valuesConfigMapData,
		},
	}
	runtimeObjs := []runtime.Object{}

	s := runtime.NewScheme()
	s.AddKnownTypes(corev1.SchemeGroupVersion, configMap)
	_ = corev1.AddToScheme(s)

	runtimeObjs = append(runtimeObjs, configMap)
	mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)
	e := &EACEngine{
		name:        "hbase",
		namespace:   "fluid",
		Client:      mockClient,
		runtimeType: common.EFCRuntime,
		Log:         ctrl.Log.WithName("hbase"),
	}
	configMap, err := kubeclient.GetConfigmapByName(mockClient, e.getConfigmapName(), e.namespace)
	if err != nil {
		t.Errorf("fail to exec")
	}

	if configMap == nil {
		t.Errorf("fail to exec")
	}

	reservedPorts, err := parsePortsFromConfigMap(configMap)
	if err != nil || len(reservedPorts) != 1 || reservedPorts[0] != 17673 {
		t.Errorf("fail to exec")
	}
}

func Test_parseCacheDirFromConfigMap(t *testing.T) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase-efc-values",
			Namespace: "fluid",
		},
		Data: map[string]string{
			"data": valuesConfigMapData,
		},
	}
	runtimeObjs := []runtime.Object{}

	s := runtime.NewScheme()
	s.AddKnownTypes(corev1.SchemeGroupVersion, configMap)
	_ = corev1.AddToScheme(s)

	runtimeObjs = append(runtimeObjs, configMap)
	mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)
	e := &EACEngine{
		name:        "hbase",
		namespace:   "fluid",
		Client:      mockClient,
		runtimeType: common.EFCRuntime,
		Log:         ctrl.Log.WithName("hbase"),
	}
	configMap, err := kubeclient.GetConfigmapByName(mockClient, e.getConfigmapName(), e.namespace)
	if err != nil {
		t.Errorf("fail to exec")
	}

	if configMap == nil {
		t.Errorf("fail to exec")
	}

	cacheDir, cacheType, err := parseCacheDirFromConfigMap(configMap)
	if err != nil || cacheDir != "/dev/efc-worker-cache-path/default/efcdemo" || cacheType != common.VolumeTypeEmptyDir {
		t.Errorf("fail to exec")
	}
}

func TestEACEngine_getDaemonset(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.EFCRuntime
		name      string
		namespace string
		Client    client.Client
	}
	tests := []struct {
		name          string
		fields        fields
		wantDaemonset *appsv1.DaemonSet
		wantErr       bool
	}{
		{
			name: "test",
			fields: fields{
				runtime: &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "runtime1",
						Namespace: "default",
					},
				},
				name:      "runtime1",
				namespace: "default",
			},
			wantDaemonset: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runtime1",
					Namespace: "default",
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "DaemonSet",
					APIVersion: "apps/v1",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := runtime.NewScheme()
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.runtime)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, &appsv1.DaemonSet{})
			_ = corev1.AddToScheme(s)
			mockClient := fake.NewFakeClientWithScheme(s, tt.fields.runtime, tt.wantDaemonset)
			e := &EACEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
			}
			gotDaemonset, err := e.getDaemonset(tt.fields.name, tt.fields.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("EACEngine.getDaemonset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotDaemonset, tt.wantDaemonset) {
				t.Errorf("EACEngine.getDaemonset() = %#v, want %#v", gotDaemonset, tt.wantDaemonset)
			}
		})
	}
}

func TestEACEngine_getMountPath(t *testing.T) {
	type fields struct {
		name      string
		namespace string
		Log       logr.Logger
		MountRoot string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "test",
			fields: fields{
				name:      "efc",
				namespace: "default",
				Log:       fake.NullLogger(),
				MountRoot: "/tmp",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &EACEngine{
				Log:       tt.fields.Log,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
			}
			t.Setenv("MOUNT_ROOT", tt.fields.MountRoot)
			wantMountPath := fmt.Sprintf("%s/%s/%s/eac-fuse", tt.fields.MountRoot+"/efc", tt.fields.namespace, e.name)
			if gotMountPath := e.getMountPath(); gotMountPath != wantMountPath {
				t.Errorf("EACEngine.getMountPoint() = %v, want %v", gotMountPath, wantMountPath)
			}
		})
	}
}

func TestEACEngine_getHostMountPath(t *testing.T) {
	type fields struct {
		name      string
		namespace string
		Log       logr.Logger
		MountRoot string
	}
	var tests = []struct {
		name          string
		fields        fields
		wantMountPath string
	}{
		{
			name: "test",
			fields: fields{
				name:      "efc",
				namespace: "default",
				Log:       fake.NullLogger(),
				MountRoot: "/tmp",
			},
			wantMountPath: "/tmp/efc/default/efc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &EACEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       tt.fields.Log,
			}
			t.Setenv("MOUNT_ROOT", tt.fields.MountRoot)
			if gotMountPath := j.getHostMountPath(); gotMountPath != tt.wantMountPath {
				t.Errorf("getHostMountPoint() = %v, want %v", gotMountPath, tt.wantMountPath)
			}
		})
	}
}

func TestEACEngine_getRuntime(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.EFCRuntime
		name      string
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		want    *datav1alpha1.EFCRuntime
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				runtime: &datav1alpha1.EFCRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "efc",
						Namespace: "default",
					},
				},
				name:      "efc",
				namespace: "default",
			},
			want: &datav1alpha1.EFCRuntime{
				TypeMeta: metav1.TypeMeta{
					Kind:       "EFCRuntime",
					APIVersion: "data.fluid.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "efc",
					Namespace: "default",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := runtime.NewScheme()
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.runtime)
			_ = corev1.AddToScheme(s)
			mockClient := fake.NewFakeClientWithScheme(s, tt.want)
			e := &EACEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
			}
			got, err := e.getRuntime()
			if (err != nil) != tt.wantErr {
				t.Errorf("EACEngine.getRuntime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EACEngine.getRuntime() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_getMountRoot(t *testing.T) {
	tests := []struct {
		name     string
		wantPath string
	}{
		{
			name:     "test",
			wantPath: "/tmp/efc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MOUNT_ROOT", "/tmp")
			if gotPath := getMountRoot(); gotPath != tt.wantPath {
				t.Errorf("getMountRoot() = %v, want %v", gotPath, tt.wantPath)
			}
		})
	}
}

func TestEACEngine_getWorkerRunningPods(t *testing.T) {
	type fields struct {
		worker    *appsv1.StatefulSet
		pods      []*corev1.Pod
		configMap *corev1.ConfigMap
		name      string
		namespace string
	}
	tests := []struct {
		name      string
		fields    fields
		wantErr   bool
		wantCount int
	}{
		{
			name:      "test",
			wantErr:   false,
			wantCount: 1,
			fields: fields{
				name:      "spark",
				namespace: "big-data",
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-worker",
						Namespace: "big-data",
						UID:       "uid1",
					},
					Spec: appsv1.StatefulSetSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app":              "efc",
								"role":             "eac-worker",
								"fluid.io/dataset": "big-data-spark",
							},
						},
					},
				},
				pods: []*corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "spark-worker-0",
							Namespace: "big-data",
							OwnerReferences: []metav1.OwnerReference{{
								Kind:       "StatefulSet",
								APIVersion: "apps/v1",
								Name:       "spark-worker",
								UID:        "uid1",
								Controller: utilpointer.BoolPtr(true),
							}},
							Labels: map[string]string{
								"app":              "efc",
								"role":             "eac-worker",
								"fluid.io/dataset": "big-data-spark",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "eac-worker",
									Ports: []corev1.ContainerPort{
										{
											Name:          "rpc",
											ContainerPort: 7788,
										},
									},
								},
							},
						},
						Status: corev1.PodStatus{
							PodIP: "127.0.0.1",
							Phase: corev1.PodRunning,
							Conditions: []corev1.PodCondition{{
								Type:   corev1.PodReady,
								Status: corev1.ConditionTrue,
							}},
						},
					},
				},
				configMap: &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-worker-endpoints",
						Namespace: "big-data",
					},
					Data: map[string]string{
						WorkerEndpointsDataName: workerEndpointsConfigMapData,
					},
				},
			},
		},
		{
			name:      "test2",
			wantErr:   false,
			wantCount: 0,
			fields: fields{
				name:      "spark",
				namespace: "big-data",
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-worker",
						Namespace: "big-data",
						UID:       "uid1",
					},
					Spec: appsv1.StatefulSetSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app":              "efc",
								"role":             "eac-worker",
								"fluid.io/dataset": "big-data-spark",
							},
						},
					},
				},
				pods: []*corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "spark-worker-0",
							Namespace: "big-data",
							OwnerReferences: []metav1.OwnerReference{{
								Kind:       "StatefulSet",
								APIVersion: "apps/v1",
								Name:       "spark-worker",
								UID:        "uid1",
								Controller: utilpointer.BoolPtr(true),
							}},
							Labels: map[string]string{
								"app":  "efc",
								"role": "eac-worker",
								//"fluid.io/dataset": "big-data-spark",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "eac-worker",
									Ports: []corev1.ContainerPort{
										{
											Name:          "rpc",
											ContainerPort: 7788,
										},
									},
								},
							},
						},
						Status: corev1.PodStatus{
							PodIP: "127.0.0.1",
							Phase: corev1.PodRunning,
							Conditions: []corev1.PodCondition{{
								Type:   corev1.PodReady,
								Status: corev1.ConditionTrue,
							}},
						},
					},
				},
				configMap: &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase-worker-endpoints",
						Namespace: "big-data",
					},
					Data: map[string]string{
						WorkerEndpointsDataName: workerEndpointsConfigMapData,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtimeObjs := []runtime.Object{}

			s := runtime.NewScheme()
			s.AddKnownTypes(appsv1.SchemeGroupVersion, tt.fields.worker)
			s.AddKnownTypes(corev1.SchemeGroupVersion, tt.fields.configMap)
			_ = corev1.AddToScheme(s)

			runtimeObjs = append(runtimeObjs, tt.fields.worker)
			runtimeObjs = append(runtimeObjs, tt.fields.configMap)
			for _, pod := range tt.fields.pods {
				runtimeObjs = append(runtimeObjs, pod)
			}
			mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)
			e := &EACEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
				Log:       ctrl.Log.WithName(tt.fields.name),
			}
			runtimeInfo, err := base.BuildRuntimeInfo(tt.fields.name, tt.fields.namespace, "efc", datav1alpha1.TieredStore{})
			if err != nil {
				t.Errorf("EACEngine.CheckWorkersReady() error = %v", err)
			}

			e.Helper = ctrlhelper.BuildHelper(runtimeInfo, mockClient, e.Log)

			pods, err := e.getWorkerRunningPods()
			if (err != nil) != tt.wantErr {
				t.Errorf("EACEngine.syncWorkersEndpoints() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(pods) != tt.wantCount {
				t.Errorf("EACEngine.syncWorkersEndpoints() count = %v, wantCount %v", len(pods), tt.wantCount)
				return
			}
		})
	}
}

func TestEACEngine_getMountInfoAndSecret(t *testing.T) {
	dataSetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "check",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{
					{
						MountPoint: "efc://volume-uuid.region.nas.aliyuncs.com:/test-fluid-3",
						EncryptOptions: []datav1alpha1.EncryptOption{
							{
								Name: "eac.nas.accessKeyId",
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: "mysecret",
										Key:  "eac.nas.accessKeyId",
									},
								},
							},
							{
								Name: "eac.nas.accessKeySecret",
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: "mysecret",
										Key:  "eac.nas.accessKeySecret",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nocheck",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{
					{
						MountPoint: "efc://volume-uuid.region.nas.aliyuncs.com:/test-fluid-3",
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "errorcheck",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{
					{
						MountPoint: "efc://volume-uuid.region.nas.aliyuncs.com:/test-fluid-3",
						EncryptOptions: []datav1alpha1.EncryptOption{
							{
								Name: AccessKeyIDName,
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: "check-none",
										Key:  "id",
									},
								},
							},
							{
								Name: AccessKeySecretName,
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: "check",
										Key:  "secret",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	secretInputs := []corev1.Secret{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mysecret",
				Namespace: "fluid",
			},
			Data: map[string][]byte{
				"eac.nas.accessKeyId":     []byte("123"),
				"eac.nas.accessKeySecret": []byte("321"),
			},
		},
	}

	objs := []runtime.Object{}
	for _, d := range dataSetInputs {
		objs = append(objs, d.DeepCopy())
	}
	for _, s := range secretInputs {
		objs = append(objs, s.DeepCopy())
	}

	fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)

	tests := []struct {
		Name            string
		Namespace       string
		WantErr         bool
		MountPoint      string
		ServiceAddr     string
		FileSystemId    string
		DirPath         string
		AccessKeyID     string
		AccessKeySecret string
	}{
		{
			Name:            "check",
			Namespace:       "fluid",
			WantErr:         false,
			MountPoint:      "volume-uuid.region.nas.aliyuncs.com:/test-fluid-3/",
			ServiceAddr:     "region",
			FileSystemId:    "volume",
			DirPath:         "/test-fluid-3/",
			AccessKeyID:     "123",
			AccessKeySecret: "321",
		},
		{
			Name:            "nocheck",
			Namespace:       "fluid",
			WantErr:         false,
			MountPoint:      "volume-uuid.region.nas.aliyuncs.com:/test-fluid-3/",
			ServiceAddr:     "region",
			FileSystemId:    "volume",
			DirPath:         "/test-fluid-3/",
			AccessKeyID:     "",
			AccessKeySecret: "",
		},
		{
			Name:      "errorcheck",
			Namespace: "fluid",
			WantErr:   true,
		},
	}

	for _, te := range tests {
		e := newEACEngine(fakeClient, te.Name, te.Namespace)
		info, err := e.getMountInfo()
		if (err != nil) != te.WantErr {
			t.Errorf("fail to exec func")
		}
		if err != nil {
			continue
		}
		if info.MountPoint != te.MountPoint || info.ServiceAddr != te.ServiceAddr || info.FileSystemId != te.FileSystemId || info.DirPath != te.DirPath || info.AccessKeyID != te.AccessKeyID || info.AccessKeySecret != te.AccessKeySecret {
			t.Errorf("fail to exec func")
		}

	}
}
