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

package base

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestGetFuseDaemonset(t *testing.T) {
	type testCase struct {
		name        string
		namespace   string
		runtimeType string
		ds          *appsv1.DaemonSet
		setClient   bool
		wantErr     bool
	}

	tests := []testCase{
		{
			name:        "alluxio",
			namespace:   "default",
			runtimeType: common.AlluxioRuntime,
			ds: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "alluxio-fuse",
					Namespace: "default",
				},
			},
			setClient: true,
			wantErr:   false,
		}, {
			name:        "jindo",
			namespace:   "default",
			runtimeType: common.JindoRuntime,
			ds: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "jindo-jindofs-fuse",
					Namespace: "default",
				},
			},
			setClient: true,
			wantErr:   false,
		}, {
			name:        "noclient",
			namespace:   "default",
			runtimeType: common.JindoRuntime,
			ds: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "jindo-jindofs-fuse",
					Namespace: "default",
				},
			},
			setClient: false,
			wantErr:   true,
		},
	}

	for _, test := range tests {
		var fakeClient client.Client
		if test.setClient {
			objs := []runtime.Object{}
			s := runtime.NewScheme()
			_ = corev1.AddToScheme(s)
			_ = datav1alpha1.AddToScheme(s)
			_ = appsv1.AddToScheme(s)
			objs = append(objs, test.ds)
			fakeClient = fake.NewFakeClientWithScheme(s, objs...)
		}

		runtimeInfo := RuntimeInfo{
			name:        test.name,
			namespace:   test.namespace,
			runtimeType: test.runtimeType,
		}

		if fakeClient != nil {
			runtimeInfo.SetAPIReader(fakeClient)
		}

		_, err := runtimeInfo.getFuseDaemonset()
		if (err == nil) == test.wantErr {
			t.Errorf("testcase %s is failed, want err %v, got err %v", test.name, test.wantErr, err)
		}
	}
}

func TestGetMountInfoFromVolumeClaim(t *testing.T) {
	namespace := "default"
	testPVCInputs := []*corev1.PersistentVolumeClaim{{
		ObjectMeta: metav1.ObjectMeta{Name: "fluid-dataset",
			Namespace: namespace},
		Spec: corev1.PersistentVolumeClaimSpec{
			VolumeName: "default-fluid-dataset",
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "nonfluidpvc",
			Annotations: common.GetExpectedFluidAnnotations(),
			Namespace:   namespace},
		Spec: corev1.PersistentVolumeClaimSpec{
			VolumeName: "nonfluidpv",
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "nopv",
			Annotations: common.GetExpectedFluidAnnotations(),
			Namespace:   namespace},
		Spec: corev1.PersistentVolumeClaimSpec{
			VolumeName: "nopv",
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "fluid-dataset-subpath",
			Annotations: common.GetExpectedFluidAnnotations(),
			Namespace:   namespace},
		Spec: corev1.PersistentVolumeClaimSpec{
			VolumeName: "default-fluid-dataset-subpath",
		},
	}}

	objs := []runtime.Object{}

	for _, pvc := range testPVCInputs {
		objs = append(objs, pvc.DeepCopy())
	}

	testPVInputs := []*corev1.PersistentVolume{{
		ObjectMeta: metav1.ObjectMeta{Name: "default-fluid-dataset"},
		Spec: corev1.PersistentVolumeSpec{
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				CSI: &corev1.CSIPersistentVolumeSource{
					Driver: "fuse.csi.fluid.io",
					VolumeAttributes: map[string]string{
						common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/nofounddataset/jindofs-fuse",
						common.VolumeAttrMountType: common.JindoRuntime,
					},
				},
			},
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "nonfluidpv", Annotations: common.GetExpectedFluidAnnotations()},
		Spec:       corev1.PersistentVolumeSpec{},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "default-fluid-dataset-subpath"},
		Spec: corev1.PersistentVolumeSpec{
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				CSI: &corev1.CSIPersistentVolumeSource{
					Driver: "fuse.csi.fluid.io",
					VolumeAttributes: map[string]string{
						common.VolumeAttrFluidPath:    "/runtime-mnt/jindo/big-data/nofounddataset/jindofs-fuse",
						common.VolumeAttrMountType:    common.JindoRuntime,
						common.VolumeAttrFluidSubPath: "subtest",
					},
				},
			},
		},
	}}

	for _, pv := range testPVInputs {
		objs = append(objs, pv.DeepCopy())
	}

	type args struct {
		name      string
		namespace string
	}
	tests := []struct {
		name        string
		args        args
		wantError   bool
		wantPath    string
		wantType    string
		wantSubPath string
	}{{
		name: "volumeClaim doesn't exist",
		args: args{
			name:      "notExist",
			namespace: namespace,
		},
		wantError: true,
	}, {
		name: "non fluid pv",
		args: args{
			name:      "nonfluidpvc",
			namespace: namespace,
		},
		wantError: true,
	}, {
		name: " fluid pv",
		args: args{
			name:      "fluid-dataset",
			namespace: namespace,
		},
		wantError: false,
		wantPath:  "/runtime-mnt/jindo/big-data/nofounddataset/jindofs-fuse",
		wantType:  common.JindoRuntime,
	}, {
		name: "no pv",
		args: args{
			name:      "nopv",
			namespace: namespace,
		},
		wantError: true,
	}, {
		name: "sub pv",
		args: args{
			name:      "fluid-dataset-subpath",
			namespace: namespace,
		},
		wantError:   false,
		wantPath:    "/runtime-mnt/jindo/big-data/nofounddataset/jindofs-fuse",
		wantType:    common.JindoRuntime,
		wantSubPath: "subtest",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testScheme := runtime.NewScheme()
			_ = corev1.AddToScheme(testScheme)
			runtimeInfo := RuntimeInfo{
				name:        tt.args.name,
				namespace:   tt.args.namespace,
				runtimeType: common.JindoRuntime,
				apiReader:   fake.NewFakeClientWithScheme(testScheme, objs...),
			}

			path, mountType, subpath, err := runtimeInfo.getMountInfo()
			got := err != nil

			if got != tt.wantError {
				t.Errorf("testcase %v getMountInfo() for %v in %v = %v, err = %v", tt.name,
					tt.args.name,
					tt.args.namespace,
					got,
					err)
			}

			if path != tt.wantPath {
				t.Errorf("testcase %v GetMountInfoFromVolumeClaim() for %v in %v  got path  %v, want path = %v", tt.name,
					tt.args.name,
					tt.args.namespace,
					path,
					tt.wantPath)
			}

			if mountType != tt.wantType {
				t.Errorf("testcase %v GetMountInfoFromVolumeClaim() for %v in %v  got mountType  %v, want mountType = %v", tt.name,
					tt.args.name,
					tt.args.namespace,
					mountType,
					tt.wantType)
			}

			if subpath != tt.wantSubPath {
				t.Errorf("testcase %v GetMountInfoFromVolumeClaim() for %v in %v  got subpath  %v, want subpath = %v", tt.name,
					tt.args.name,
					tt.args.namespace,
					subpath,
					tt.wantSubPath)
			}

		})
	}

}
