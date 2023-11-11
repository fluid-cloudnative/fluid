/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package recover

import (
	"os"
	"reflect"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/mountinfo"
	. "github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachineryRuntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	k8sexec "k8s.io/utils/exec"
	"k8s.io/utils/mount"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const testfuseRecoverPeriod = 30

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
			patch1 := ApplyFunc(os.ReadFile, func(filename string) ([]byte, error) {
				return []byte(fakeToken), nil
			})
			defer patch1.Reset()

			t.Setenv("NODE_IP", fakeNodeIP)
			t.Setenv("KUBELET_CLIENT_CERT", fakeClientCert)
			t.Setenv("KUBELET_CLIENT_KEY", fakeClientKey)
			t.Setenv("KUBELET_TIMEOUT", fakeKubeletTimeout)

			kubeletClient, err := initializeKubeletClient()
			So(err, ShouldBeNil)
			So(kubeletClient, ShouldNotBeNil)
		})
	})
}

func TestRecover_run(t *testing.T) {
	Convey("TestRecover_run", t, func() {
		Convey("run success", func() {
			dataset := &v1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "jfsdemo",
					Namespace: "default",
				},
			}

			s := apimachineryRuntime.NewScheme()
			_ = v1alpha1.AddToScheme(s)
			_ = corev1.AddToScheme(s)
			fakeClient := fake.NewFakeClientWithScheme(s, dataset)

			mockedFsMounts := map[string]string{}

			sourcePath := "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse"
			targetPath := "/var/lib/kubelet/pods/1140aa96-18c2-4896-a14f-7e3965a51406/volumes/kubernetes.io~csi/default-jfsdemo/mount"

			fakeMounter := &mount.FakeMounter{}
			r := &FuseRecover{
				SafeFormatAndMount: mount.SafeFormatAndMount{
					Interface: fakeMounter,
				},
				KubeClient:        fakeClient,
				ApiReader:         fakeClient,
				Recorder:          record.NewFakeRecorder(1),
				recoverFusePeriod: testfuseRecoverPeriod,
				locks:             utils.NewVolumeLocks(),
			}

			patch1 := ApplyMethod(reflect.TypeOf(fakeMounter), "Mount", func(_ *mount.FakeMounter, source string, target string, _ string, _ []string) error {
				mockedFsMounts[source] = target
				return nil
			})
			defer patch1.Reset()

			patch2 := ApplyFunc(mountinfo.GetBrokenMountPoints, func() ([]mountinfo.MountPoint, error) {
				return []mountinfo.MountPoint{{
					SourcePath:            sourcePath,
					MountPath:             targetPath,
					FilesystemType:        "fuse.juicefs",
					ReadOnly:              false,
					Count:                 0,
					NamespacedDatasetName: "default-jfsdemo",
				}}, nil
			})
			defer patch2.Reset()

			patch3 := ApplyFunc(volume.GetNamespacedNameByVolumeId, func(client client.Reader, volumeId string) (namespace, name string, err error) {
				return "default", "jfsdemo", nil
			})
			defer patch3.Reset()

			patch4 := ApplyPrivateMethod(r, "shouldRecover", func(mountPath string) (bool, error) {
				return true, nil
			})
			defer patch4.Reset()

			r.runOnce()

			if target, exists := mockedFsMounts[sourcePath]; !exists || target != targetPath {
				t.Errorf("failed to recover mount point")
			}
		})
	})
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
		dataset *v1alpha1.Dataset
		pv      *corev1.PersistentVolume
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
				pv: &corev1.PersistentVolume{
					ObjectMeta: metav1.ObjectMeta{
						Name: "default-jfsdemo",
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
				eventType:   corev1.EventTypeNormal,
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
				pv: &corev1.PersistentVolume{
					ObjectMeta: metav1.ObjectMeta{
						Name: "other-pv",
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
				eventType:   corev1.EventTypeNormal,
				eventReason: common.FuseRecoverSucceed,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := apimachineryRuntime.NewScheme()
			_ = v1alpha1.AddToScheme(s)
			_ = scheme.AddToScheme(s)
			fakeClient := fake.NewFakeClientWithScheme(s, tt.fields.dataset, tt.fields.pv)
			r := &FuseRecover{
				KubeClient: fakeClient,
				ApiReader:  fakeClient,
				Recorder:   record.NewFakeRecorder(1),
			}

			r.eventRecord(tt.args.point, tt.args.eventType, tt.args.eventReason)
		})
	}
}

func TestNewFuseRecover(t *testing.T) {
	type args struct {
		kubeClient        client.Client
		recorder          record.EventRecorder
		recoverFusePeriod string
		locks             *utils.VolumeLocks
	}

	fakeClient := fake.NewFakeClient()
	fakeRecorder := record.NewFakeRecorder(1)
	volumeLocks := utils.NewVolumeLocks()

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
				recoverFusePeriod: "5s",
				locks:             volumeLocks,
			},
			want: &FuseRecover{
				SafeFormatAndMount: mount.SafeFormatAndMount{
					Interface: mount.New(""),
					Exec:      k8sexec.New(),
				},
				KubeClient:              fakeClient,
				ApiReader:               fakeClient,
				Recorder:                fakeRecorder,
				recoverFusePeriod:       defaultFuseRecoveryPeriod,
				recoverWarningThreshold: defaultRecoverWarningThreshold,
				locks:                   volumeLocks,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(utils.MountRoot, "/runtime-mnt")
			t.Setenv(FuseRecoveryPeriod, tt.args.recoverFusePeriod)

			got, err := NewFuseRecover(tt.args.kubeClient, tt.args.recorder, tt.args.kubeClient, tt.args.locks)
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
