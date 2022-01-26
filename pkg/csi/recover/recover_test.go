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

package recover

import (
	"errors"
	. "github.com/agiledragon/gomonkey"
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubelet"
	"github.com/fluid-cloudnative/fluid/pkg/utils/mountinfo"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachineryRuntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	k8sexec "k8s.io/utils/exec"
	"k8s.io/utils/mount"
	"os"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
	"time"
)

const testfuseRecoverPeriod = 30

var mockPod = v1.Pod{
	ObjectMeta: metav1.ObjectMeta{
		Labels:    map[string]string{"role": "juicefs-fuse"},
		Name:      "test-pod",
		Namespace: "default",
		OwnerReferences: []metav1.OwnerReference{{
			Kind: "DaemonSet",
			Name: "test-juicefs-fuse",
		}},
	},
	Spec: v1.PodSpec{},
	Status: v1.PodStatus{
		Conditions: []v1.PodCondition{{
			Type:   v1.PodReady,
			Status: v1.ConditionTrue,
		}},
		ContainerStatuses: []v1.ContainerStatus{{
			Name: "test-container",
			State: v1.ContainerState{Running: &v1.ContainerStateRunning{
				StartedAt: metav1.Time{Time: time.Now()},
			}},
		}},
	},
}

func Test_initializeKubeletClient(t *testing.T) {
	Convey("Test_initializeKubeletClient", t, func() {
		Convey("initialize success with non-default kubelet timeout", func() {
			const (
				fakeToken          = "fakeToken"
				fakeNodeIP         = "fakeNodeIP"
				fakeClientCert     = ""
				fakeClientKey      = ""
				fakeKubeletTimeout = "120"
			)
			patch1 := ApplyFunc(ioutil.ReadFile, func(filename string) ([]byte, error) {
				return []byte(fakeToken), nil
			})
			defer patch1.Reset()

			os.Setenv("NODE_IP", fakeNodeIP)
			os.Setenv("KUBELET_CLIENT_CERT", fakeClientCert)
			os.Setenv("KUBELET_CLIENT_KEY", fakeClientKey)
			os.Setenv("KUBELET_TIMEOUT", fakeKubeletTimeout)

			kubeletClient, err := initializeKubeletClient()
			So(err, ShouldBeNil)
			So(kubeletClient, ShouldNotBeNil)
		})
	})
}

func TestRecover_run(t *testing.T) {
	Convey("TestRecover_run", t, func() {
		Convey("run success", func() {
			kubeclient := &kubelet.KubeletClient{}
			patch1 := ApplyMethod(reflect.TypeOf(kubeclient), "GetNodeRunningPods", func(_ *kubelet.KubeletClient) (*v1.PodList, error) {
				return &v1.PodList{Items: []v1.Pod{mockPod}}, nil
			})
			defer patch1.Reset()
			patch2 := ApplyFunc(mountinfo.GetBrokenMountPoints, func() ([]mountinfo.MountPoint, error) {
				return []mountinfo.MountPoint{{
					SourcePath:            "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse",
					MountPath:             "/var/lib/kubelet/pods/1140aa96-18c2-4896-a14f-7e3965a51406/volumes/kubernetes.io~csi/default-jfsdemo/mount",
					FilesystemType:        "fuse.juicefs",
					ReadOnly:              false,
					Count:                 0,
					NamespacedDatasetName: "default-jfsdemo",
				}}, nil
			})
			defer patch2.Reset()

			r := &FuseRecover{
				SafeFormatAndMount: mount.SafeFormatAndMount{
					Interface: &mount.FakeMounter{},
				},
				KubeClient:        fake.NewFakeClient(),
				KubeletClient:     kubeclient,
				Recorder:          record.NewFakeRecorder(1),
				containers:        make(map[string]*containerStat),
				recoverFusePeriod: testfuseRecoverPeriod,
			}
			r.runOnce()
		})
		Convey("GetNodeRunningPods error", func() {
			kubeclient := &kubelet.KubeletClient{}
			patch1 := ApplyMethod(reflect.TypeOf(kubeclient), "GetNodeRunningPods", func(_ *kubelet.KubeletClient) (*v1.PodList, error) {
				return &v1.PodList{}, errors.New("test")
			})
			defer patch1.Reset()
			patch2 := ApplyFunc(mountinfo.GetBrokenMountPoints, func() ([]mountinfo.MountPoint, error) {
				return []mountinfo.MountPoint{}, nil
			})
			defer patch2.Reset()

			r := FuseRecover{
				SafeFormatAndMount: mount.SafeFormatAndMount{},
				KubeClient:         fake.NewFakeClient(),
				KubeletClient:      &kubelet.KubeletClient{},
				Recorder:           record.NewFakeRecorder(1),
			}
			r.runOnce()
		})
		Convey("container restart", func() {
			kubeclient := &kubelet.KubeletClient{}
			patch1 := ApplyMethod(reflect.TypeOf(kubeclient), "GetNodeRunningPods", func(_ *kubelet.KubeletClient) (*v1.PodList, error) {
				return &v1.PodList{Items: []v1.Pod{mockPod}}, nil
			})
			defer patch1.Reset()
			patch2 := ApplyFunc(mountinfo.GetBrokenMountPoints, func() ([]mountinfo.MountPoint, error) {
				return []mountinfo.MountPoint{}, nil
			})
			defer patch2.Reset()

			r := &FuseRecover{
				SafeFormatAndMount: mount.SafeFormatAndMount{
					Interface: &mount.FakeMounter{},
				},
				KubeClient:        fake.NewFakeClient(),
				KubeletClient:     kubeclient,
				Recorder:          record.NewFakeRecorder(1),
				containers:        make(map[string]*containerStat),
				recoverFusePeriod: testfuseRecoverPeriod,
			}

			r.containers = map[string]*containerStat{
				"test-container-test-juicefs-fuse-default": {
					name:          "test-container",
					podName:       "test-pod",
					namespace:     "default",
					daemonSetName: "test-juicefs-fuse",
					startAt: metav1.Time{
						Time: time.Now().Add(-1 * time.Minute),
					},
				},
			}
			r.runOnce()
		})
	})
}

func TestFuseRecover_compareOrRecordContainerStat(t *testing.T) {
	type fields struct {
		key       string
		container *containerStat
	}
	type args struct {
		pod v1.Pod
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantRestarted bool
	}{
		{
			name: "test1",
			fields: fields{
				key: "test-container-test-juicefs-fuse-default",
				container: &containerStat{
					name:          "test-container",
					podName:       "test-pod",
					namespace:     "default",
					daemonSetName: "test-juicefs-fuse",
					startAt: metav1.Time{
						Time: time.Now().Add(-1 * time.Minute),
					},
				},
			},
			args: args{
				pod: mockPod,
			},
			wantRestarted: true,
		},
		{
			name: "test2",
			fields: fields{
				key: "test-container-test-juicefs-fuse-default",
				container: &containerStat{
					name:          "test-container",
					podName:       "test-pod",
					namespace:     "default",
					daemonSetName: "test-juicefs-fuse",
					startAt: metav1.Time{
						Time: time.Now(),
					},
				},
			},
			args: args{
				pod: mockPod,
			},
			wantRestarted: false,
		},
		{
			name:   "test-nods",
			fields: fields{},
			args: args{
				pod: v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
				},
			},
			wantRestarted: false,
		},
		{
			name: "test-cn-not-running",
			fields: fields{
				key: "test-container-test-juicefs-fuse-default",
				container: &containerStat{
					name:          "test-container",
					podName:       "test-pod",
					namespace:     "default",
					daemonSetName: "test-juicefs-fuse",
					startAt: metav1.Time{
						Time: time.Now().Add(-1 * time.Minute),
					},
				},
			},
			args: args{
				pod: v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels:    map[string]string{"role": "juicefs-fuse"},
						Name:      "test-pod",
						Namespace: "default",
						OwnerReferences: []metav1.OwnerReference{{
							Kind: "DaemonSet",
							Name: "test-juicefs-fuse",
						}},
					},
					Spec: v1.PodSpec{},
					Status: v1.PodStatus{
						ContainerStatuses: []v1.ContainerStatus{{
							Name: "test-container",
							State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{
								StartedAt: metav1.Time{Time: time.Now()},
							}},
						}},
					}},
			},
			wantRestarted: false,
		},
		{
			name:   "test-no-container-record",
			fields: fields{},
			args: args{
				pod: mockPod,
			},
			wantRestarted: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kubeletClient := &kubelet.KubeletClient{}
			r := &FuseRecover{
				SafeFormatAndMount: mount.SafeFormatAndMount{
					Interface: &mount.FakeMounter{},
				},
				KubeClient:        fake.NewFakeClient(),
				KubeletClient:     kubeletClient,
				Recorder:          record.NewFakeRecorder(1),
				containers:        make(map[string]*containerStat),
				recoverFusePeriod: testfuseRecoverPeriod,
			}
			if tt.fields.container != nil {
				r.containers[tt.fields.key] = tt.fields.container
			}
			if gotRestarted := r.compareOrRecordContainerStat(tt.args.pod); gotRestarted != tt.wantRestarted {
				t.Errorf("compareOrRecordContainerStat() = %v, want %v", gotRestarted, tt.wantRestarted)
			}
		})
	}
}

func TestFuseRecover_umountDuplicate(t *testing.T) {
	type args struct {
		point mountinfo.MountPoint
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test",
			args: args{
				point: mountinfo.MountPoint{
					SourcePath:            "/test",
					MountPath:             "/test",
					FilesystemType:        "test",
					ReadOnly:              false,
					Count:                 3,
					NamespacedDatasetName: "test",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := FuseRecover{SafeFormatAndMount: mount.SafeFormatAndMount{
				Interface: &mount.FakeMounter{},
			}}
			r.umountDuplicate(tt.args.point)
		})
	}
}

func TestFuseRecover_recoverBrokenMount(t *testing.T) {
	type args struct {
		point mountinfo.MountPoint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				point: mountinfo.MountPoint{
					SourcePath:            "/test",
					MountPath:             "/test",
					FilesystemType:        "test",
					ReadOnly:              false,
					Count:                 1,
					NamespacedDatasetName: "test",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := FuseRecover{SafeFormatAndMount: mount.SafeFormatAndMount{
				Interface: &mount.FakeMounter{},
			}}
			if err := r.recoverBrokenMount(tt.args.point); (err != nil) != tt.wantErr {
				t.Errorf("recoverBrokenMount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFuseRecover_eventRecord(t *testing.T) {
	type fields struct {
		containers map[string]*containerStat
		dataset    *v1alpha1.Dataset
	}
	type args struct {
		point       mountinfo.MountPoint
		eventType   string
		eventReason string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test",
			fields: fields{
				dataset: &v1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "jfsdemo",
						Namespace: "default",
					},
				},
			},
			args: args{
				point: mountinfo.MountPoint{
					SourcePath:            "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse",
					MountPath:             "/var/lib/kubelet/pods/1140aa96-18c2-4896-a14f-7e3965a51406/volumes/kubernetes.io~csi/default-jfsdemo/mount",
					FilesystemType:        "fuse.juicefs",
					ReadOnly:              false,
					Count:                 0,
					NamespacedDatasetName: "default-jfsdemo",
				},
				eventType:   v1.EventTypeNormal,
				eventReason: common.FuseRecoverSucceed,
			},
		},
		{
			name: "test-err",
			fields: fields{
				dataset: &v1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "jfsdemo",
						Namespace: "default",
					},
				},
			},
			args: args{
				point: mountinfo.MountPoint{
					SourcePath:            "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse",
					MountPath:             "/var/lib/kubelet/pods/1140aa96-18c2-4896-a14f-7e3965a51406/volumes/kubernetes.io~csi/default-jfsdemo/mount",
					FilesystemType:        "fuse.juicefs",
					ReadOnly:              false,
					Count:                 0,
					NamespacedDatasetName: "jfsdemo",
				},
				eventType:   v1.EventTypeNormal,
				eventReason: common.FuseRecoverSucceed,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := apimachineryRuntime.NewScheme()
			s.AddKnownTypes(v1alpha1.GroupVersion, tt.fields.dataset)
			fakeClient := fake.NewFakeClientWithScheme(s, tt.fields.dataset)
			r := &FuseRecover{
				KubeClient:    fakeClient,
				KubeletClient: nil,
				Recorder:      record.NewFakeRecorder(1),
				containers:    tt.fields.containers,
			}
			r.eventRecord(tt.args.point, tt.args.eventType, tt.args.eventReason)
		})
	}
}

func TestNewFuseRecover(t *testing.T) {
	type args struct {
		kubeClient        client.Client
		recorder          record.EventRecorder
		recoverFusePeriod int
	}

	fakeClient := fake.NewFakeClient()
	fakeRecorder := record.NewFakeRecorder(1)
	fakeKubeletClient := &kubelet.KubeletClient{}
	fakeContainersMap := make(map[string]*containerStat)
	fakeRecoverFusePeriod := 20

	tests := []struct {
		name    string
		args    args
		want    *FuseRecover
		wantErr bool
	}{
		{
			name: "test_newFuseRecover",
			args: args{
				kubeClient:        fakeClient,
				recorder:          fakeRecorder,
				recoverFusePeriod: fakeRecoverFusePeriod,
			},
			want: &FuseRecover{
				SafeFormatAndMount: mount.SafeFormatAndMount{
					Interface: mount.New(""),
					Exec:      k8sexec.New(),
				},
				KubeClient:        fakeClient,
				KubeletClient:     fakeKubeletClient,
				Recorder:          fakeRecorder,
				containers:        fakeContainersMap,
				recoverFusePeriod: fakeRecoverFusePeriod,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(utils.MountRoot, "/runtime-mnt")

			patch := ApplyFunc(initializeKubeletClient, func() (*kubelet.KubeletClient, error) {
				return fakeKubeletClient, nil
			})
			defer patch.Reset()

			got, err := NewFuseRecover(tt.args.kubeClient, tt.args.recorder, tt.args.recoverFusePeriod)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFuseRecover() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFuseRecover() got = %v, want %v", got, tt.want)
			}
		})
	}
}
