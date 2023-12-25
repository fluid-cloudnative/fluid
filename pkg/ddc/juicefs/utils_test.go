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
	"fmt"
	"reflect"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

func TestJuiceFSEngine_getDaemonset(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.JuiceFSRuntime
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
				runtime: &datav1alpha1.JuiceFSRuntime{
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
			e := &JuiceFSEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
			}
			gotDaemonset, err := e.getDaemonset(tt.fields.name, tt.fields.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("JuicefsEngine.getDaemonset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotDaemonset, tt.wantDaemonset) {
				t.Errorf("JuiceFSEngine.getDaemonset() = %#v, want %#v", gotDaemonset, tt.wantDaemonset)
			}
		})
	}
}

func TestJuiceFSEngine_getFuseDaemonsetName(t *testing.T) {
	type fields struct {
		name string
	}
	tests := []struct {
		name       string
		fields     fields
		wantDsName string
	}{
		{
			name: "test",
			fields: fields{
				name: "juicefs",
			},
			wantDsName: "juicefs-fuse",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &JuiceFSEngine{
				name: tt.fields.name,
			}
			if gotDsName := e.getFuseDaemonsetName(); gotDsName != tt.wantDsName {
				t.Errorf("JuiceFSEngine.getFuseDaemonsetName() = %v, want %v", gotDsName, tt.wantDsName)
			}
		})
	}
}

func TestJuiceFSEngine_getMountPoint(t *testing.T) {
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
				name:      "juicefs",
				namespace: "default",
				Log:       fake.NullLogger(),
				MountRoot: "/tmp",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &JuiceFSEngine{
				Log:       tt.fields.Log,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
			}
			t.Setenv("MOUNT_ROOT", tt.fields.MountRoot)
			wantMountPath := fmt.Sprintf("%s/%s/%s/juicefs-fuse", tt.fields.MountRoot+"/juicefs", tt.fields.namespace, e.name)
			if gotMountPath := e.getMountPoint(); gotMountPath != wantMountPath {
				t.Errorf("JuiceFSEngine.getMountPoint() = %v, want %v", gotMountPath, wantMountPath)
			}
		})
	}
}

func TestJuiceFSEngine_getHostMountPoint(t *testing.T) {
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
				name:      "juicefs",
				namespace: "default",
				Log:       fake.NullLogger(),
				MountRoot: "/tmp",
			},
			wantMountPath: "/tmp/juicefs/default/juicefs",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JuiceFSEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       tt.fields.Log,
			}
			t.Setenv("MOUNT_ROOT", tt.fields.MountRoot)
			if gotMountPath := j.getHostMountPoint(); gotMountPath != tt.wantMountPath {
				t.Errorf("getHostMountPoint() = %v, want %v", gotMountPath, tt.wantMountPath)
			}
		})
	}
}

func TestJuiceFSEngine_getRuntime(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.JuiceFSRuntime
		name      string
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		want    *datav1alpha1.JuiceFSRuntime
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				runtime: &datav1alpha1.JuiceFSRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "juicefs",
						Namespace: "default",
					},
				},
				name:      "juicefs",
				namespace: "default",
			},
			want: &datav1alpha1.JuiceFSRuntime{
				TypeMeta: metav1.TypeMeta{
					Kind:       "JuiceFSRuntime",
					APIVersion: "data.fluid.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "juicefs",
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
			e := &JuiceFSEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
			}
			got, err := e.getRuntime()
			if (err != nil) != tt.wantErr {
				t.Errorf("JuiceFSEngine.getRuntime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JuiceFSEngine.getRuntime() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestJuiceFSEngine_parseJuiceFSImage(t *testing.T) {
	type args struct {
		edition         string
		image           string
		tag             string
		imagePullPolicy string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		want2   string
		wantErr bool
	}{
		{
			name: "test0",
			args: args{
				edition:         "community",
				image:           "juicedata/mount",
				tag:             "ce-v1.0.4",
				imagePullPolicy: "IfNotPresent",
			},
			want:    "juicedata/mount",
			want1:   "ce-v1.0.4",
			want2:   "IfNotPresent",
			wantErr: false,
		},
		{
			name: "test1",
			args: args{
				edition:         "community",
				image:           "",
				tag:             "",
				imagePullPolicy: "IfNotPresent",
			},
			want:    "juicedata/mount",
			want1:   "ce-v1.1.0-beta2",
			want2:   "IfNotPresent",
			wantErr: false,
		},
		{
			name: "test2",
			args: args{
				edition:         "enterprise",
				image:           "juicedata/mount",
				tag:             "ee-4.9.15",
				imagePullPolicy: "IfNotPresent",
			},
			want:    "juicedata/mount",
			want1:   "ee-4.9.15",
			want2:   "IfNotPresent",
			wantErr: false,
		},
		{
			name: "test3",
			args: args{
				edition:         "enterprise",
				image:           "",
				tag:             "",
				imagePullPolicy: "IfNotPresent",
			},
			want:    "juicedata/mount",
			want1:   "ee-4.9.14",
			want2:   "IfNotPresent",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &JuiceFSEngine{}
			t.Setenv(common.JuiceFSCEImageEnv, "juicedata/mount:ce-v1.1.0-beta2")
			t.Setenv(common.JuiceFSEEImageEnv, "juicedata/mount:ee-4.9.14")
			got, got1, got2, err := e.parseJuiceFSImage(tt.args.edition, tt.args.image, tt.args.tag, tt.args.imagePullPolicy)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseJuiceFSImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("JuiceFSEngine.parseJuiceFSImage() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("JuiceFSEngine.parseJuiceFSImage() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("JuiceFSEngine.parseJuiceFSImage() got2 = %v, want %v", got2, tt.want2)
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
			wantPath: "/tmp/juicefs",
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

func Test_parseInt64Size(t *testing.T) {
	type args struct {
		sizeStr string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				sizeStr: "10",
			},
			want:    10,
			wantErr: false,
		},
		{
			name: "test2",
			args: args{
				sizeStr: "v",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInt64Size(tt.args.sizeStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInt64Size() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseInt64Size() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseSubPathFromMountPoint(t *testing.T) {
	type args struct {
		mountPoint string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test-correct",
			args: args{
				mountPoint: "juicefs:///abc",
			},
			want:    "/abc",
			wantErr: false,
		},
		{
			name: "test-wrong",
			args: args{
				mountPoint: "/abc",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSubPathFromMountPoint(tt.args.mountPoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSubPathFromMountPoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseSubPathFromMountPoint() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJuiceFSEngine_GetRunningPodsOfStatefulSet(t *testing.T) {
	type args struct {
		stsName   string
		namespace string
	}
	tests := []struct {
		name     string
		args     args
		sts      *appsv1.StatefulSet
		podLists *corev1.PodList
		wantPods []corev1.Pod
		wantErr  bool
	}{
		{
			name: "test1",
			args: args{
				stsName:   "test1",
				namespace: "fluid",
			},
			sts: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: "fluid",
				},
				Spec: appsv1.StatefulSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"sts": "test1"},
					},
				},
			},
			podLists: &corev1.PodList{
				Items: []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test1-pod",
							Namespace: "fluid",
							Labels:    map[string]string{"sts": "test1"},
						},
						Status: corev1.PodStatus{
							Phase: corev1.PodRunning,
							Conditions: []corev1.PodCondition{{
								Type:   corev1.PodReady,
								Status: corev1.ConditionTrue,
							}},
						},
					},
				},
			},
			wantPods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1-pod",
						Namespace: "fluid",
						Labels:    map[string]string{"sts": "test1"},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
						Conditions: []corev1.PodCondition{{
							Type:   corev1.PodReady,
							Status: corev1.ConditionTrue,
						}},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "test2",
			args: args{
				stsName:   "app",
				namespace: "fluid",
			},
			sts:      &appsv1.StatefulSet{},
			podLists: &corev1.PodList{},
			wantPods: []corev1.Pod{},
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := runtime.NewScheme()
			s.AddKnownTypes(appsv1.SchemeGroupVersion, &appsv1.StatefulSet{})
			s.AddKnownTypes(corev1.SchemeGroupVersion, &corev1.PodList{})
			_ = corev1.AddToScheme(s)
			mockClient := fake.NewFakeClientWithScheme(s, tt.sts.DeepCopy(), tt.podLists.DeepCopy())
			j := &JuiceFSEngine{
				Client: mockClient,
			}
			gotPods, err := j.GetRunningPodsOfStatefulSet(tt.args.stsName, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRunningPodsOfStatefulSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(gotPods[0].Status, tt.wantPods[0].Status) {
					t.Errorf("testcase %s GetRunningPodsOfStatefulSet() gotPods = %v, want %v", tt.name, gotPods, tt.wantPods)
				}
			}
		})
	}
}

func TestJuiceFSEngine_getValuesConfigMap(t *testing.T) {
	type fields struct {
		runtime     *datav1alpha1.JuiceFSRuntime
		runtimeType string
		engineImpl  string
		name        string
		namespace   string
		Client      client.Client
	}
	tests := []struct {
		name    string
		fields  fields
		wantCm  *corev1.ConfigMap
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				runtime: &datav1alpha1.JuiceFSRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "runtime",
						Namespace: "default",
					},
				},
				name:        "test",
				namespace:   "default",
				runtimeType: "juicefs",
				engineImpl:  "juicefs",
			},
			wantCm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-juicefs-values",
					Namespace: "default",
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := runtime.NewScheme()
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.runtime)
			s.AddKnownTypes(corev1.SchemeGroupVersion, &corev1.ConfigMap{})
			_ = corev1.AddToScheme(s)
			mockClient := fake.NewFakeClientWithScheme(s, tt.fields.runtime, tt.wantCm)
			e := &JuiceFSEngine{
				runtime:     tt.fields.runtime,
				runtimeType: tt.fields.runtimeType,
				engineImpl:  tt.fields.engineImpl,
				name:        tt.fields.name,
				namespace:   tt.fields.namespace,
				Client:      mockClient,
			}
			gotCm, err := e.GetValuesConfigMap()
			if (err != nil) != tt.wantErr {
				t.Errorf("JuicefsEngine.getValuesConfigMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCm, tt.wantCm) {
				t.Errorf("JuiceFSEngine.getValuesConfigMap() = %#v, want %#v", gotCm, tt.wantCm)
			}
		})
	}
}

var (
	communityValuesData = `    fullnameOverride: jfscedemo
    edition: community
    source: ${METAURL}
    image: xxx/juicefs/juicefs-fuse
    imageTag: v1.0.0-beta2
    imagePullPolicy: IfNotPresent
    user: 0
    group: 0
    fsGroup: 0
    configs:
      name: community
      accesskeySecret: juicefs-ce-secret
      secretkeySecret: juicefs-ce-secret
      bucket: xxx
      metaurlSecret: juicefs-ce-secret
      formatCmd: /usr/local/bin/juicefs format --access-key=${ACCESS_KEY} --secret-key=${SECRET_KEY}
        --no-update --bucket=http://xxx ${METAURL} community
    fuse:
      enabled: true
      image: xxx/juicefs/juicefs-fuse
      nodeSelector:
        fluid.io/f-kube-system-jfscedemo: "true"
      imageTag: v1.0.0-beta2
      imagePullPolicy: IfNotPresent
      criticalPod: true
      mountPath: /runtime-mnt/juicefs/kube-system/jfscedemo/juicefs-fuse
      cacheDir: /data/cachece
      hostMountPath: /runtime-mnt/juicefs/kube-system/jfscedemo
      command: /bin/mount.juicefs ${METAURL} /runtime-mnt/juicefs/kube-system/jfscedemo/juicefs-fuse
        -o max-uploads=80,cache-size=102400,free-space-ratio=0.1,cache-dir=/data/cachece,metrics=0.0.0.0:9567
      statCmd: stat -c %i /runtime-mnt/juicefs/kube-system/jfscedemo/juicefs-fuse
    worker:
      mountPath: /runtime-mnt/juicefs/kube-system/jfscedemo/juicefs-fuse
      cacheDir: /data/cachece
      statCmd: stat -c %i /runtime-mnt/juicefs/kube-system/jfscedemo/juicefs-fuse
      command: /bin/mount.juicefs ${METAURL} /runtime-mnt/juicefs/kube-system/jfscedemo/juicefs-fuse
        -o max-uploads=80,cache-size=102400,free-space-ratio=0.1,cache-dir=/data/cachece,metrics=0.0.0.0:9567
    placement: Exclusive`

	enterpriseValuesData = `    fullnameOverride: jfsdemo
    edition: enterprise
    source: test
    image: xxx/juicefs/juicefs-fuse
    imageTag: v1.0.0-beta2
    imagePullPolicy: IfNotPresent
    user: 0
    group: 0
    fsGroup: 0
    configs:
      name: test
      accesskeySecret: juicefs-secret
      secretkeySecret: juicefs-secret
      tokenSecret: juicefs-secret
      formatCmd: /usr/bin/juicefs auth --token=${TOKEN} --accesskey=${ACCESS_KEY} --secretkey=${SECRET_KEY}
        test
    fuse:
      enabled: true
      image: xxx/juicefs/juicefs-fuse
      nodeSelector:
        fluid.io/f-kube-system-jfsdemo: "true"
      envs:
      - name: BASE_URL
        value: http://xxx/static
        valuefrom: null
      - name: CFG_URL
        value: http://xxx/volume/%s/mount
        valuefrom: null
      imageTag: v1.0.0-beta2
      imagePullPolicy: IfNotPresent
      criticalPod: true
      subPath: /dataset-bak
      mountPath: /runtime-mnt/juicefs/kube-system/jfsdemo/juicefs-fuse
      cacheDir: /data/cache
      hostMountPath: /runtime-mnt/juicefs/kube-system/jfsdemo
      command: /sbin/mount.juicefs test /runtime-mnt/juicefs/kube-system/jfsdemo/juicefs-fuse
        -o cache-group=jfsdemo,no-sharing,subdir=/dataset-bak,cache-dir=/data/cache,max-uploads=80,free-space-ratio=0.1,foreground,cache-size=102400,max-cached-inodes=10000000
      statCmd: stat -c %i /runtime-mnt/juicefs/kube-system/jfsdemo/juicefs-fuse
    worker:
      envs:
      - name: BASE_URL
        value: http://xxx/static
        valuefrom: null
      - name: CFG_URL
        value: http://xxx/volume/%s/mount
        valuefrom: null
      mountPath: /runtime-mnt/juicefs/kube-system/jfsdemo/juicefs-fuse
      cacheDir: /data/cache
      statCmd: stat -c %i /runtime-mnt/juicefs/kube-system/jfsdemo/juicefs-fuse
      command: /sbin/mount.juicefs test /runtime-mnt/juicefs/kube-system/jfsdemo/juicefs-fuse
        -o foreground,cache-group=jfsdemo,free-space-ratio=0.1,cache-dir=/data/cache,cache-size=102400,max-cached-inodes=10000000,max-uploads=80,subdir=/dataset-bak
    placement: Exclusive`
)

func TestJuiceFSEngine_GetEdition(t *testing.T) {
	type args struct {
		cmName    string
		namespace string
	}
	tests := []struct {
		name      string
		args      args
		cm        *corev1.ConfigMap
		wantValue string
	}{
		{
			name: "test1",
			args: args{
				cmName:    "test1",
				namespace: "fluid",
			},
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: "fluid",
				},
				Data: map[string]string{
					"data": communityValuesData,
				},
			},
			wantValue: "community",
		},
		{
			name: "test2",
			args: args{
				cmName:    "test2",
				namespace: "fluid",
			},
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test2",
					Namespace: "fluid",
				},
				Data: map[string]string{
					"data": enterpriseValuesData,
				},
			},
			wantValue: "enterprise",
		},
		{
			name: "test3",
			args: args{
				cmName:    "test3",
				namespace: "fluid",
			},
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test3",
					Namespace: "fluid",
				},
				Data: map[string]string{
					"data": "test",
				},
			},
			wantValue: "",
		},
		{
			name: "test4",
			args: args{
				cmName:    "test4",
				namespace: "fluid",
			},
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test4",
					Namespace: "fluid",
				},
				Data: map[string]string{
					"data": "a: b",
				},
			},
			wantValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var engine *JuiceFSEngine
			patch1 := ApplyMethod(reflect.TypeOf(engine), "GetValuesConfigMap",
				func(_ *JuiceFSEngine) (*corev1.ConfigMap, error) {
					return tt.cm, nil
				})
			defer patch1.Reset()

			j := JuiceFSEngine{
				name:      "test",
				namespace: tt.args.namespace,
				Log:       fake.NullLogger(),
				runtime: &datav1alpha1.JuiceFSRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "fluid",
					},
				},
			}
			gotValue := j.GetEdition()
			if gotValue != tt.wantValue {
				t.Errorf("JuiceFSEngine.GetEdition() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestGetMetricsPort(t *testing.T) {
	type args struct {
		options map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				options: map[string]string{
					"metrics": "0.0.0.0:9567",
				},
			},
			want:    9567,
			wantErr: false,
		},
		{
			name: "test-default",
			args: args{
				options: map[string]string{},
			},
			want:    9567,
			wantErr: false,
		},
		{
			name: "test-wrong1",
			args: args{
				options: map[string]string{
					"metrics": "0.0.0.0:test",
				},
			},
			want:    9567,
			wantErr: true,
		},
		{
			name: "test-wrong2",
			args: args{
				options: map[string]string{
					"metrics": "0.0.0.0",
				},
			},
			want:    9567,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMetricsPort(tt.args.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMetricsPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetMetricsPort() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseVersion(t *testing.T) {
	type args struct {
		version string
	}
	var tests = []struct {
		name    string
		args    args
		want    *ClientVersion
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				version: "v1.0.0",
			},
			want: &ClientVersion{
				Major: 1,
				Minor: 0,
				Patch: 0,
				Tag:   "",
			},
			wantErr: false,
		},
		{
			name: "test2",
			args: args{
				version: "nightly",
			},
			want: &ClientVersion{
				Tag: "nightly",
			},
			wantErr: false,
		},
		{
			name: "test3",
			args: args{
				version: "4.9.0",
			},
			want: &ClientVersion{
				Major: 4,
				Minor: 9,
				Patch: 0,
				Tag:   "",
			},
			wantErr: false,
		},
		{
			name: "test4",
			args: args{
				version: "1.0.0-rc1",
			},
			want: &ClientVersion{
				Major: 1,
				Minor: 0,
				Patch: 0,
				Tag:   "rc1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseVersion(tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseVersion() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJuiceFSEngine_getWorkerCommand(t *testing.T) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-worker-script",
			Namespace: "fluid",
		},
		Data: map[string]string{
			"script.sh": `#!/bin/bash

if [ enterprise = community ]; then
echo "$(date '+%Y/%m/%d %H:%M:%S').$(printf "%03d" $(($(date '+%N')/1000))) juicefs format start."
/usr/bin/juicefs auth --token=${TOKEN} --accesskey=${ACCESS_KEY} --secretkey=${SECRET_KEY} --bucket=http://test4.minio.default.svc.cluster.local:9000 test-fluid-2
elif [ ! -f /root/.juicefs/test-fluid-2.conf ]; then
echo "$(date '+%Y/%m/%d %H:%M:%S').$(printf "%03d" $(($(date '+%N')/1000))) juicefs auth start."
/usr/bin/juicefs auth --token=${TOKEN} --accesskey=${ACCESS_KEY} --secretkey=${SECRET_KEY} --bucket=http://test4.minio.default.svc.cluster.local:9000 test-fluid-2
fi

echo "$(date '+%Y/%m/%d %H:%M:%S').$(printf "%03d" $(($(date '+%N')/1000))) juicefs mount start."
/sbin/mount.juicefs test-fluid-2 /runtime-mnt/juicefs/default/jfsdemo-ee/juicefs-fuse -o subdir=/demo,cache-size=2048,free-space-ratio=0.1,cache-dir=/dev/shm,foreground,no-update,cache-group=default-jfsdemo-ee
`,
		},
	}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, cm)

	fakeClient := fake.NewFakeClientWithScheme(testScheme, testObjs...)
	tests := []struct {
		name        string
		runtimeName string
		wantCommand string
		wantErr     bool
	}{
		{
			name:        "test-normal",
			runtimeName: "test",
			wantCommand: "/sbin/mount.juicefs test-fluid-2 /runtime-mnt/juicefs/default/jfsdemo-ee/juicefs-fuse -o subdir=/demo,cache-size=2048,free-space-ratio=0.1,cache-dir=/dev/shm,foreground,no-update,cache-group=default-jfsdemo-ee",
			wantErr:     false,
		},
		{
			name:        "test-not-found",
			runtimeName: "test1",
			wantCommand: "",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := JuiceFSEngine{
				name:      tt.runtimeName,
				namespace: "fluid",
				Client:    fakeClient,
				Log:       fake.NullLogger(),
			}
			gotCommand, err := j.getWorkerCommand()
			if (err != nil) != tt.wantErr {
				t.Errorf("getWorkerCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotCommand != tt.wantCommand {
				t.Errorf("getWorkerCommand() gotCommand = %v, want %v", gotCommand, tt.wantCommand)
			}
		})
	}
}

func TestJuiceFSEngine_updateWorkerScript(t *testing.T) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-worker-script",
			Namespace: "fluid",
		},
		Data: map[string]string{
			"script.sh": `#!/bin/bash

if [ enterprise = community ]; then
echo "$(date '+%Y/%m/%d %H:%M:%S').$(printf "%03d" $(($(date '+%N')/1000))) juicefs format start."
/usr/bin/juicefs auth --token=${TOKEN} --accesskey=${ACCESS_KEY} --secretkey=${SECRET_KEY} --bucket=http://test4.minio.default.svc.cluster.local:9000 test-fluid-2
elif [ ! -f /root/.juicefs/test-fluid-2.conf ]; then
echo "$(date '+%Y/%m/%d %H:%M:%S').$(printf "%03d" $(($(date '+%N')/1000))) juicefs auth start."
/usr/bin/juicefs auth --token=${TOKEN} --accesskey=${ACCESS_KEY} --secretkey=${SECRET_KEY} --bucket=http://test4.minio.default.svc.cluster.local:9000 test-fluid-2
fi

echo "$(date '+%Y/%m/%d %H:%M:%S').$(printf "%03d" $(($(date '+%N')/1000))) juicefs mount start."
/sbin/mount.juicefs test-fluid-2 /runtime-mnt/juicefs/default/jfsdemo-ee/juicefs-fuse -o subdir=/demo,cache-size=2048,free-space-ratio=0.1,cache-dir=/dev/shm,foreground,no-update,cache-group=default-jfsdemo-ee
`,
		},
	}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, cm)

	fakeClient := fake.NewFakeClientWithScheme(testScheme, testObjs...)
	tests := []struct {
		name        string
		runtimeName string
		command     string
		wantCommand string
		wantErr     bool
	}{
		{
			name:        "test-normal",
			runtimeName: "test",
			command:     "echo abc",
			wantCommand: "echo abc",
			wantErr:     false,
		},
		{
			name:        "test-not-found",
			runtimeName: "test1",
			command:     "",
			wantCommand: "",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := JuiceFSEngine{
				name:      tt.runtimeName,
				namespace: "fluid",
				Client:    fakeClient,
				Log:       fake.NullLogger(),
			}
			err := j.updateWorkerScript(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateWorkerScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotCommand, err := j.getWorkerCommand()
			if err != nil {
				t.Errorf("updateWorkerScript get configmap command error = %v", err)
				return
			}
			if gotCommand != tt.command {
				t.Errorf("updateWorkerScript() gotCommand = %v, want %v", gotCommand, tt.command)
			}
		})
	}
}

func TestEscapeBashStr(t *testing.T) {
	cases := [][]string{
		{"abc", "abc"},
		{"test-volume", "test-volume"},
		{"http://minio.kube-system:9000/minio/dynamic-ce", "http://minio.kube-system:9000/minio/dynamic-ce"},
		{"$(cat /proc/self/status | grep CapEff > /test.txt)", "$'$(cat /proc/self/status | grep CapEff > /test.txt)'"},
		{"hel`cat /proc/self/status`lo", "$'hel`cat /proc/self/status`lo'"},
		{"'h'el`cat /proc/self/status`lo", "$'\\'h\\'el`cat /proc/self/status`lo'"},
		{"\\'h\\'el`cat /proc/self/status`lo", "$'\\'h\\'el`cat /proc/self/status`lo'"},
		{"$'h'el`cat /proc/self/status`lo", "$'$\\'h\\'el`cat /proc/self/status`lo'"},
		{"hel\\`cat /proc/self/status`lo", "$'hel\\\\`cat /proc/self/status`lo'"},
		{"hel\\\\`cat /proc/self/status`lo", "$'hel\\\\`cat /proc/self/status`lo'"},
		{"hel\\'`cat /proc/self/status`lo", "$'hel\\'`cat /proc/self/status`lo'"},
	}
	for _, c := range cases {
		escaped := escapeBashStr(c[0])
		if escaped != c[1] {
			t.Errorf("escapeBashVar(%s) = %s, want %s", c[0], escaped, c[1])
		}
	}
}
