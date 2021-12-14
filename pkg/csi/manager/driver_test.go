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

package manager

import (
	"errors"
	. "github.com/agiledragon/gomonkey"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/csi/util"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	. "github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestPodDriver_podReadyHandler(t *testing.T) {
	Convey("TestPodDriver_podReadyHandler", t, func() {
		Convey("podReadyHandler success", func() {
			patch1 := ApplyFunc(kubeclient.GetPersistentVolume, func(client client.Client, name string) (pv *v1.PersistentVolume, err error) {
				return &v1.PersistentVolume{
					Spec: v1.PersistentVolumeSpec{
						PersistentVolumeSource: v1.PersistentVolumeSource{
							CSI: &v1.CSIPersistentVolumeSource{
								VolumeAttributes: map[string]string{common.FLUID_PATH: "/test"},
							},
						},
						AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadOnlyMany},
					},
				}, nil
			})
			defer patch1.Reset()

			patch2 := ApplyFunc(util.GetPVMountPoint, func(pvName string) (mountPoints []string, err error) {
				return []string{"/targetPath"}, nil
			})
			defer patch2.Reset()
			patch3 := ApplyFunc(util.CheckMountPointBroken, func(mountPath string) (broken bool, err error) {
				return false, nil
			})
			defer patch3.Reset()

			k8sClient := fake.NewFakeClient()
			p := NewPodDriver(k8sClient)
			pod := &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-fuse-xxx",
				},
			}
			err := p.podReadyHandler(pod)
			So(err, ShouldBeNil)
		})
		Convey("podReadyHandler pod nil", func() {
			k8sClient := fake.NewFakeClient()
			p := NewPodDriver(k8sClient)
			err := p.podReadyHandler(nil)
			So(err, ShouldBeNil)
		})
		Convey("podReadyHandler pod runtimeName err", func() {
			k8sClient := fake.NewFakeClient()
			p := NewPodDriver(k8sClient)
			pod := &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-xxx",
				},
			}
			err := p.podReadyHandler(pod)
			So(err, ShouldNotBeNil)
		})
		Convey("podReadyHandler GetPVMountPoint err", func() {
			patch1 := ApplyFunc(kubeclient.GetPersistentVolume, func(client client.Client, name string) (pv *v1.PersistentVolume, err error) {
				return &v1.PersistentVolume{
					Spec: v1.PersistentVolumeSpec{
						PersistentVolumeSource: v1.PersistentVolumeSource{
							CSI: &v1.CSIPersistentVolumeSource{
								VolumeAttributes: map[string]string{common.FLUID_PATH: "/test"},
							},
						},
						AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteMany},
					},
				}, nil
			})
			defer patch1.Reset()

			patch2 := ApplyFunc(util.GetPVMountPoint, func(pvName string) (mountPoints []string, err error) {
				return nil, errors.New("test")
			})
			defer patch2.Reset()
			k8sClient := fake.NewFakeClient()
			p := NewPodDriver(k8sClient)
			pod := &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-fuse-xxx",
				},
			}
			err := p.podReadyHandler(pod)
			So(err, ShouldNotBeNil)
		})
		Convey("podReadyHandler CheckMountPointBroken err", func() {
			patch1 := ApplyFunc(kubeclient.GetPersistentVolume, func(client client.Client, name string) (pv *v1.PersistentVolume, err error) {
				return &v1.PersistentVolume{
					Spec: v1.PersistentVolumeSpec{
						PersistentVolumeSource: v1.PersistentVolumeSource{
							CSI: &v1.CSIPersistentVolumeSource{
								VolumeAttributes: map[string]string{common.FLUID_PATH: "/test"},
							},
						},
						AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadOnlyMany},
					},
				}, nil
			})
			defer patch1.Reset()

			patch2 := ApplyFunc(util.GetPVMountPoint, func(pvName string) (mountPoints []string, err error) {
				return []string{"/targetPath"}, nil
			})
			defer patch2.Reset()
			patch3 := ApplyFunc(util.CheckMountPointBroken, func(mountPath string) (broken bool, err error) {
				return false, errors.New("test")
			})
			defer patch3.Reset()

			k8sClient := fake.NewFakeClient()
			p := NewPodDriver(k8sClient)
			pod := &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-fuse-xxx",
				},
			}
			err := p.podReadyHandler(pod)
			So(err, ShouldNotBeNil)
		})
	})
}
