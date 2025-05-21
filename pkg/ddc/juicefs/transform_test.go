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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"reflect"
	"testing"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/net"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

var dummy = func(client client.Client) (ports []int, err error) {
	return []int{14000, 14001}, nil
}

func TestJuiceFSEngine_transform(t *testing.T) {
	juicefsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "fluid",
		},
		Data: map[string][]byte{
			"metaurl": []byte("test"),
		},
	}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, (*juicefsSecret).DeepCopy())

	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
	runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "juicefs")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	engine := JuiceFSEngine{
		name:      "test",
		namespace: "fluid",
		Client:    client,
		Log:       fake.NullLogger(),
		runtime: &datav1alpha1.JuiceFSRuntime{
			Spec: datav1alpha1.JuiceFSRuntimeSpec{
				Fuse: datav1alpha1.JuiceFSFuseSpec{},
			},
		},
		runtimeInfo: runtimeInfo,
	}
	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))

	var tests = []struct {
		runtime *datav1alpha1.JuiceFSRuntime
		dataset *datav1alpha1.Dataset
		value   *JuiceFS
	}{
		{&datav1alpha1.JuiceFSRuntime{
			Spec: datav1alpha1.JuiceFSRuntimeSpec{
				Fuse: datav1alpha1.JuiceFSFuseSpec{},
				Worker: datav1alpha1.JuiceFSCompTemplateSpec{
					Replicas:     2,
					Resources:    corev1.ResourceRequirements{},
					Options:      nil,
					Env:          nil,
					Enabled:      false,
					NodeSelector: nil,
				},
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "juicefs:///mnt/test",
					Name:       "test",
					EncryptOptions: []datav1alpha1.EncryptOption{{
						Name: "metaurl",
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{
								Name: "test",
								Key:  "metaurl",
							},
						},
					}},
				}},
			},
		}, &JuiceFS{}},
	}
	for _, test := range tests {
		err := engine.transformFuse(test.runtime, test.dataset, test.value)
		if err != nil {
			t.Errorf("error %v", err)
		}
	}
}

func TestJuiceFSEngine_transformTolerations(t *testing.T) {
	type fields struct {
		name      string
		namespace string
	}
	type args struct {
		dataset *datav1alpha1.Dataset
		value   *JuiceFS
	}
	var tests = []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test",
			fields: fields{
				name:      "",
				namespace: "",
			},
			args: args{
				dataset: &datav1alpha1.Dataset{Spec: datav1alpha1.DatasetSpec{
					Tolerations: []corev1.Toleration{{
						Key:      "a",
						Operator: corev1.TolerationOpEqual,
						Value:    "b",
					}},
				}},
				value: &JuiceFS{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JuiceFSEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
			}
			j.transformTolerations(tt.args.dataset, tt.args.value)
			if len(tt.args.value.Tolerations) != len(tt.args.dataset.Spec.Tolerations) {
				t.Errorf("transformTolerations() tolerations = %v", tt.args.value.Tolerations)
			}
		})
	}
}

func TestJuiceFSEngine_transformPodMetadata(t *testing.T) {
	engine := &JuiceFSEngine{Log: fake.NullLogger()}

	type testCase struct {
		Name    string
		Runtime *datav1alpha1.JuiceFSRuntime
		Value   *JuiceFS

		wantValue *JuiceFS
	}

	testCases := []testCase{
		{
			Name: "set_common_labels_and_annotations",
			Runtime: &datav1alpha1.JuiceFSRuntime{
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					PodMetadata: datav1alpha1.PodMetadata{
						Labels:      map[string]string{"common-key": "common-value"},
						Annotations: map[string]string{"common-annotation": "val"},
					},
					UpdateStrategy: datav1alpha1.InPlaceIfPossible,
				},
			},
			Value: &JuiceFS{},
			wantValue: &JuiceFS{
				Worker: Worker{
					Labels:      map[string]string{"common-key": "common-value", common.RuntimePodType: common.RuntimeWorkerPod},
					Annotations: map[string]string{"common-annotation": "val", common.AnnotationRuntimeName: ""},
				},
				Fuse: Fuse{
					Labels:      map[string]string{"common-key": "common-value"},
					Annotations: map[string]string{"common-annotation": "val"},
				},
			},
		},
		{
			Name: "set_master_and_workers_labels_and_annotations",
			Runtime: &datav1alpha1.JuiceFSRuntime{
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					PodMetadata: datav1alpha1.PodMetadata{
						Labels:      map[string]string{"common-key": "common-value"},
						Annotations: map[string]string{"common-annotation": "val"},
					},
					UpdateStrategy: datav1alpha1.InPlaceIfPossible,
					Worker: datav1alpha1.JuiceFSCompTemplateSpec{
						PodMetadata: datav1alpha1.PodMetadata{
							Labels:      map[string]string{"common-key": "worker-value", common.RuntimePodType: common.RuntimeWorkerPod},
							Annotations: map[string]string{"common-annotation": "worker-val", common.AnnotationRuntimeName: ""},
						},
					},
				},
			},
			Value: &JuiceFS{},
			wantValue: &JuiceFS{
				Worker: Worker{
					Labels:      map[string]string{"common-key": "worker-value", common.RuntimePodType: common.RuntimeWorkerPod},
					Annotations: map[string]string{"common-annotation": "worker-val", common.AnnotationRuntimeName: ""},
				},
				Fuse: Fuse{
					Labels:      map[string]string{"common-key": "common-value"},
					Annotations: map[string]string{"common-annotation": "val"},
				},
			},
		},
	}

	for _, tt := range testCases {
		err := engine.transformPodMetadata(tt.Runtime, tt.Value)
		if err != nil {
			t.Fatalf("test name: %s. Expect err = nil, but got err = %v", tt.Name, err)
		}

		if !reflect.DeepEqual(tt.Value, tt.wantValue) {
			t.Fatalf("test name: %s. Expect value %v, but got %v", tt.Name, tt.wantValue, tt.Value)
		}
	}
}

func TestJuiceFSEngine_allocatePorts(t *testing.T) {
	pr := net.ParsePortRangeOrDie("14000-15999")
	err := portallocator.SetupRuntimePortAllocator(nil, pr, "bitmap", dummy)
	if err != nil {
		t.Fatal(err.Error())
	}
	type args struct {
		runtime *datav1alpha1.JuiceFSRuntime
		value   *JuiceFS
	}
	tests := []struct {
		name                  string
		args                  args
		wantErr               bool
		wantWorkerMetricsPort bool
		wantFuseMetricsPort   bool
	}{
		{
			name: "test",
			args: args{
				runtime: &datav1alpha1.JuiceFSRuntime{},
				value: &JuiceFS{
					Edition: CommunityEdition,
				},
			},
			wantErr:               false,
			wantWorkerMetricsPort: true,
			wantFuseMetricsPort:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JuiceFSEngine{}
			if err := j.allocatePorts(tt.args.runtime, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("allocatePorts() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantWorkerMetricsPort {
				if tt.args.value.Worker.MetricsPort == nil {
					t.Error("allocatePorts() got worker port nil")
				}
				if *tt.args.value.Worker.MetricsPort < 14000 || *tt.args.value.Worker.MetricsPort > 15999 {
					t.Errorf("allocatePorts() got worker port = %v, but want in range [14000, 15999]", *tt.args.value.Worker.MetricsPort)
				}
			}
			if tt.wantFuseMetricsPort {
				if tt.args.value.Fuse.MetricsPort == nil {
					t.Error("allocatePorts() got fuse port nil")
				}
				if *tt.args.value.Fuse.MetricsPort < 14000 || *tt.args.value.Fuse.MetricsPort > 15999 {
					t.Errorf("allocatePorts() got fuse port = %v, but want in range [14000, 15999]", *tt.args.value.Fuse.MetricsPort)
				}
			}
		})
	}
}

func TestJuiceFSEngine_genWorkerMount(t *testing.T) {
	type fields struct {
		name      string
		namespace string
		Log       logr.Logger
	}
	type args struct {
		value   *JuiceFS
		runtime *datav1alpha1.JuiceFSRuntime
	}
	tests := []struct {
		name              string
		fields            fields
		args              args
		wantErr           bool
		wantWorkerCommand string
		wantWorkerStatCmd string
	}{
		{
			name: "test-community",
			fields: fields{
				name:      "test",
				namespace: "fluid",
				Log:       fake.NullLogger(),
			},
			args: args{
				value: &JuiceFS{
					FullnameOverride: "test-community",
					Edition:          "community",
					Source:           "redis://127.0.0.1:6379",
					Configs: Configs{
						Name:            "test-community",
						AccessKeySecret: "test",
						SecretKeySecret: "test",
						Bucket:          "http://127.0.0.1:9000/minio/test",
						MetaUrlSecret:   "test",
						Storage:         "minio",
					},
					Worker: Worker{
						MountPath: "/test-worker",
					},
				},
				runtime: &datav1alpha1.JuiceFSRuntime{},
			},
			wantErr:           false,
			wantWorkerCommand: "exec /bin/mount.juicefs redis://127.0.0.1:6379 /test-worker -o metrics=0.0.0.0:9567",
			wantWorkerStatCmd: "stat -c %i /test-worker",
		},
		{
			name: "test-community-options",
			fields: fields{
				name:      "test",
				namespace: "fluid",
				Log:       fake.NullLogger(),
			},
			args: args{
				value: &JuiceFS{
					FullnameOverride: "test-community",
					Edition:          "community",
					Source:           "redis://127.0.0.1:6379",
					Configs: Configs{
						Name:            "test-community",
						AccessKeySecret: "test",
						SecretKeySecret: "test",
						Bucket:          "http://127.0.0.1:9000/minio/test",
						MetaUrlSecret:   "test",
						Storage:         "minio",
					},
					Fuse: Fuse{
						SubPath:       "/",
						MountPath:     "/test",
						HostMountPath: "/test",
					},
					Worker: Worker{
						MountPath: "/test-worker",
					},
				},
				runtime: &datav1alpha1.JuiceFSRuntime{Spec: datav1alpha1.JuiceFSRuntimeSpec{Worker: datav1alpha1.JuiceFSCompTemplateSpec{
					Options: map[string]string{"metrics": "127.0.0.1:9567"},
				}}},
			},
			wantErr:           false,
			wantWorkerCommand: "exec /bin/mount.juicefs redis://127.0.0.1:6379 /test-worker -o metrics=127.0.0.1:9567",
			wantWorkerStatCmd: "stat -c %i /test-worker",
		},
		{
			name: "test-enterprise",
			fields: fields{
				name:      "test",
				namespace: "fluid",
				Log:       fake.NullLogger(),
			},
			args: args{
				value: &JuiceFS{
					FullnameOverride: "test-enterprise",
					Edition:          "enterprise",
					Source:           "test-enterprise",
					Configs: Configs{
						Name:            "test-enterprise",
						AccessKeySecret: "test",
						SecretKeySecret: "test",
						Bucket:          "http://127.0.0.1:9000/minio/test",
						TokenSecret:     "test",
					},
					Fuse: Fuse{
						SubPath:       "/",
						MountPath:     "/test",
						HostMountPath: "/test",
					},
					Worker: Worker{
						MountPath: "/test",
					},
				},
				runtime: &datav1alpha1.JuiceFSRuntime{},
			},
			wantErr:           false,
			wantWorkerCommand: "exec /sbin/mount.juicefs test-enterprise /test -o foreground,no-update,cache-group=fluid-test-enterprise",
			wantWorkerStatCmd: "stat -c %i /test",
		},
		{
			name: "test-enterprise-options",
			fields: fields{
				name:      "test",
				namespace: "fluid",
				Log:       fake.NullLogger(),
			},
			args: args{
				value: &JuiceFS{
					FullnameOverride: "test-enterprise",
					Edition:          "enterprise",
					Source:           "test-enterprise",
					Configs: Configs{
						Name:            "test-enterprise",
						AccessKeySecret: "test",
						SecretKeySecret: "test",
						Bucket:          "http://127.0.0.1:9000/minio/test",
						TokenSecret:     "test",
					},
					Fuse: Fuse{
						SubPath:       "/",
						MountPath:     "/test",
						HostMountPath: "/test",
					},
					Worker: Worker{
						MountPath: "/test",
					},
				},
				runtime: &datav1alpha1.JuiceFSRuntime{Spec: datav1alpha1.JuiceFSRuntimeSpec{Worker: datav1alpha1.JuiceFSCompTemplateSpec{
					Options: map[string]string{"verbose": "", "cache-group": "test"},
				}}},
			},
			wantErr:           false,
			wantWorkerCommand: "exec /sbin/mount.juicefs test-enterprise /test -o verbose,foreground,no-update,cache-group=test",
			wantWorkerStatCmd: "stat -c %i /test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tt.fields.name,
					Namespace: tt.fields.namespace,
				},
			}
			client := fake.NewFakeClientWithScheme(testScheme, dataset)
			j := &JuiceFSEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       tt.fields.Log,
				Client:    client,
			}
			j.genWorkerMount(tt.args.value, tt.args.runtime.Spec.Worker.Options)
			if len(tt.args.value.Worker.Command) != len(tt.wantWorkerCommand) ||
				tt.args.value.Worker.StatCmd != tt.wantWorkerStatCmd {
				t.Errorf("command got = %v\ncommand want = %v", tt.args.value.Worker.Command, tt.wantWorkerCommand)
				t.Errorf("stat cmd got = %v\nstat cmd want = %v", tt.args.value.Worker.StatCmd, tt.wantWorkerStatCmd)
			}
		})
	}
}

func TestJuiceFSEngine_genEdition(t *testing.T) {
	type args struct {
		mount                datav1alpha1.Mount
		value                *JuiceFS
		SharedEncryptOptions []datav1alpha1.EncryptOption
	}
	tests := []struct {
		name        string
		args        args
		wantEdition string
	}{
		{
			name: "test-community-1",
			args: args{
				mount: datav1alpha1.Mount{
					EncryptOptions: []datav1alpha1.EncryptOption{{
						Name:      JuiceMetaUrl,
						ValueFrom: datav1alpha1.EncryptOptionSource{},
					}},
				},
				value:                &JuiceFS{},
				SharedEncryptOptions: nil,
			},
			wantEdition: "community",
		},
		{
			name: "test-community-2",
			args: args{
				value: &JuiceFS{},
				SharedEncryptOptions: []datav1alpha1.EncryptOption{{
					Name:      JuiceMetaUrl,
					ValueFrom: datav1alpha1.EncryptOptionSource{},
				}},
			},
			wantEdition: "community",
		},
		{
			name: "test-enterprise-1",
			args: args{
				mount: datav1alpha1.Mount{
					EncryptOptions: []datav1alpha1.EncryptOption{{
						Name:      JuiceToken,
						ValueFrom: datav1alpha1.EncryptOptionSource{},
					}},
				},
				value:                &JuiceFS{},
				SharedEncryptOptions: nil,
			},
			wantEdition: "enterprise",
		},
		{
			name: "test-enterprise-2",
			args: args{
				value: &JuiceFS{},
				SharedEncryptOptions: []datav1alpha1.EncryptOption{{
					Name:      JuiceToken,
					ValueFrom: datav1alpha1.EncryptOptionSource{},
				}},
			},
			wantEdition: "enterprise",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JuiceFSEngine{}
			j.genEdition(tt.args.mount, tt.args.value, tt.args.SharedEncryptOptions)
		})
	}
}
