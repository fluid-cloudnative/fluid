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

package referencedataset

import (
	"context"
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestReferenceDatasetEngine_CreateVolume(t *testing.T) {
	testScheme := runtime.NewScheme()
	_ = corev1.AddToScheme(testScheme)
	_ = datav1alpha1.AddToScheme(testScheme)
	_ = appsv1.AddToScheme(testScheme)

	testObjs := []runtime.Object{}

	var dataset = datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "done",
			Namespace: "big-data",
		},
		Status: datav1alpha1.DatasetStatus{
			Runtimes: []datav1alpha1.Runtime{
				{
					Name:      "done",
					Namespace: "big-data",
					Type:      common.AlluxioRuntime,
				},
			},
		},
	}
	var runtime = datav1alpha1.AlluxioRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "done",
			Namespace: "big-data",
		},
	}
	var runtimeInfo, err = base.BuildRuntimeInfo(runtime.Name, runtime.Namespace, common.AlluxioRuntime, datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	var pv = corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: runtimeInfo.GetPersistentVolumeName(),
		},
	}
	var pvc = corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      runtime.GetName(),
			Namespace: runtime.GetNamespace(),
		},
	}

	var refRuntime = datav1alpha1.ThinRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
	}
	var refDataset = datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{
				{
					MountPoint: "dataset://big-data/done",
				},
			},
		},
	}

	testObjs = append(testObjs, &dataset, &refDataset)
	testObjs = append(testObjs, &runtime, &refRuntime)
	testObjs = append(testObjs, &pv, &pvc)

	fakeClient := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	type fields struct {
		Id        string
		Client    client.Client
		Log       logr.Logger
		name      string
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "create-volume",
			fields: fields{
				Log:       fake.NullLogger(),
				namespace: refRuntime.Namespace,
				name:      refRuntime.Name,
				Client:    fakeClient,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		e := &ReferenceDatasetEngine{
			Id:        tt.fields.Id,
			Client:    tt.fields.Client,
			Log:       tt.fields.Log,
			name:      tt.fields.name,
			namespace: tt.fields.namespace,
		}
		if err := e.CreateVolume(); (err != nil) != tt.wantErr {
			t.Errorf("CreateVolume() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		var pvs corev1.PersistentVolumeList
		err = fakeClient.List(context.TODO(), &pvs)
		if err != nil {
			t.Errorf("fail to exec the function with error %v", err)
			return
		}
		if len(pvs.Items) != 2 {
			t.Errorf("fail to create the pv")
		}

		var pvcs corev1.PersistentVolumeClaimList
		err = fakeClient.List(context.TODO(), &pvcs)
		if err != nil {
			t.Errorf("fail to exec the function with error %v", err)
			return
		}
		if len(pvcs.Items) != 2 {
			t.Errorf("fail to create the pvc")
		}
	}
}

func TestReferenceDatasetEngine_DeleteVolume(t *testing.T) {
	testScheme := runtime.NewScheme()
	_ = corev1.AddToScheme(testScheme)
	_ = datav1alpha1.AddToScheme(testScheme)
	_ = appsv1.AddToScheme(testScheme)

	testObjs := []runtime.Object{}

	var refRuntime = datav1alpha1.ThinRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
	}
	var refDataset = datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{
				{
					MountPoint: "dataset://big-data/done",
				},
			},
		},
	}

	var runtimeInfo, err = base.BuildRuntimeInfo(refRuntime.Name, refRuntime.Namespace, common.ThinRuntime, datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	var pv = corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        runtimeInfo.GetPersistentVolumeName(),
			Annotations: common.ExpectedFluidAnnotations,
		},
	}
	var pvc = corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:        refRuntime.GetName(),
			Namespace:   refRuntime.GetNamespace(),
			Annotations: common.ExpectedFluidAnnotations,
			Finalizers:  []string{"kubernetes.io/pvc-protection"},
		},
	}

	testObjs = append(testObjs, &refDataset, &refRuntime)
	testObjs = append(testObjs, &pv, &pvc)

	fakeClient := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	type fields struct {
		Id        string
		Client    client.Client
		Log       logr.Logger
		name      string
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "delete-volume",
			fields: fields{
				Log:       fake.NullLogger(),
				namespace: refRuntime.Namespace,
				name:      refRuntime.Name,
				Client:    fakeClient,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		e := &ReferenceDatasetEngine{
			Id:        tt.fields.Id,
			Client:    tt.fields.Client,
			Log:       tt.fields.Log,
			name:      tt.fields.name,
			namespace: tt.fields.namespace,
		}
		kubeclient.SetPVCDeleteTimeout(0)
		// pvc is designed to delete until timeout, so ignore the error
		_ = e.DeleteVolume()
		var pvs corev1.PersistentVolumeList
		err = fakeClient.List(context.TODO(), &pvs)
		if err != nil {
			t.Errorf("fail to exec the function with error %v", err)
			return
		}
		if len(pvs.Items) != 0 {
			t.Errorf("fail to delete the pv")
		}

		var pvcs corev1.PersistentVolumeClaimList
		err = fakeClient.List(context.TODO(), &pvcs)
		if err != nil {
			t.Errorf("fail to exec the function with error %v", err)
			return
		}
		if len(pvcs.Items) != 0 {
			t.Errorf("fail to delete the pvc")
		}
	}
}

func Test_accessModesForVirtualDataset(t *testing.T) {
	type args struct {
		virtualDataset *datav1alpha1.Dataset
		copiedPvSpec   *corev1.PersistentVolumeSpec
	}
	tests := []struct {
		name string
		args args
		want []corev1.PersistentVolumeAccessMode
	}{
		// TODO: Add test cases.
		{
			name: "no_access_mode",
			args: args{
				virtualDataset: &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name: "v1",
					},
					Spec: datav1alpha1.DatasetSpec{},
				},
				copiedPvSpec: &corev1.PersistentVolumeSpec{},
			},
			want: []corev1.PersistentVolumeAccessMode{corev1.ReadOnlyMany},
		}, {
			name: "read_access_mode",
			args: args{
				virtualDataset: &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name: "v1",
					},
					Spec: datav1alpha1.DatasetSpec{},
				},
				copiedPvSpec: &corev1.PersistentVolumeSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
				},
			},
			want: []corev1.PersistentVolumeAccessMode{corev1.ReadOnlyMany},
		}, {
			name: "read_access_mode",
			args: args{
				virtualDataset: &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name: "v1",
					},
					Spec: datav1alpha1.DatasetSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
					},
				},
				copiedPvSpec: &corev1.PersistentVolumeSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
				},
			},
			want: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := accessModesForVirtualDataset(tt.args.virtualDataset, tt.args.copiedPvSpec.AccessModes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("name %v accessModesForVirtualDataset() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
