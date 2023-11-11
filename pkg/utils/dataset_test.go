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

package utils

import (
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGetDataset(t *testing.T) {
	mockDatasetName := "fluid-data-set"
	mockDatasetNamespace := "default"
	initDataset := &datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mockDatasetName,
			Namespace: mockDatasetNamespace,
		},
	}
	s := runtime.NewScheme()
	s.AddKnownTypes(datav1alpha1.GroupVersion, initDataset)

	fakeClient := fake.NewFakeClientWithScheme(s, initDataset)

	testCases := map[string]struct {
		name      string
		namespace string
		wantName  string
		notFound  bool
	}{
		"get Dataset test case 1": {
			name:      mockDatasetName,
			namespace: mockDatasetNamespace,
			wantName:  mockDatasetName,
			notFound:  false,
		},
		"get Dataset test case 2": {
			name:      mockDatasetName + "not-exist",
			namespace: mockDatasetNamespace,
			wantName:  "",
			notFound:  true,
		},
	}

	for k, item := range testCases {
		gotDataset, err := GetDataset(fakeClient, item.name, item.namespace)
		if item.notFound {
			if err == nil && gotDataset != nil {
				t.Errorf("%s check failure, want got nil", k)
			}
		} else {
			if gotDataset.Name != item.wantName {
				t.Errorf("%s check failure,got Dataset name:%s,want name:%s", k, gotDataset.Name, item.wantName)
			}
		}
	}

}

func TestIsSetupDone(t *testing.T) {
	testCases := map[string]struct {
		conditions []datav1alpha1.DatasetCondition
		wantDone   bool
	}{
		"test dataset is setup done case 1": {
			conditions: []datav1alpha1.DatasetCondition{
				{Type: datav1alpha1.DatasetReady},
			},
			wantDone: true,
		},
		"test dataset is setup done case 2": {
			conditions: []datav1alpha1.DatasetCondition{
				{Type: datav1alpha1.DatasetInitialized},
			},
			wantDone: false,
		},
		"test dataset is setup done case 3": {
			conditions: nil,
			wantDone:   false,
		},
	}

	for k, item := range testCases {
		dataset := mockDatasetWithCondition("dataset-1", "default", item.conditions)
		gotDone := IsSetupDone(dataset)

		if gotDone != item.wantDone {
			t.Errorf("%s check failure, want:%t,got:%t", k, item.wantDone, gotDone)
		}

	}

}

func TestGetAccessModesOfDataset(t *testing.T) {

	testCases := map[string]struct {
		name           string
		getName        string
		namespace      string
		accessMode     []v1.PersistentVolumeAccessMode
		wantAccessMode []v1.PersistentVolumeAccessMode
		notFound       bool
	}{
		"test get dataset access model case 1": {
			name:      "dataset-1",
			getName:   "dataset-1",
			notFound:  false,
			namespace: "default",
			accessMode: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteMany,
			},
			wantAccessMode: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteMany,
			},
		},
		"test get dataset access model case 2": {
			name:       "dataset-1",
			getName:    "dataset-1",
			notFound:   false,
			namespace:  "default",
			accessMode: nil,
			wantAccessMode: []v1.PersistentVolumeAccessMode{
				v1.ReadOnlyMany,
			},
		},
		"test get dataset access model case 3": {
			name:           "dataset-1",
			getName:        "dataset-1-notexist",
			notFound:       true,
			namespace:      "default",
			accessMode:     nil,
			wantAccessMode: nil,
		},
	}

	for k, item := range testCases {
		dataset := mockDatasetWithAccessModel(item.name, item.namespace, item.accessMode)
		s := runtime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, dataset)

		fakeClient := fake.NewFakeClientWithScheme(s, dataset)

		gotAccessModel, err := GetAccessModesOfDataset(fakeClient, item.getName, item.namespace)

		if item.notFound {
			if err == nil {
				t.Errorf("%s check failure,want err but got nil", k)
			}
		} else {
			if !reflect.DeepEqual(gotAccessModel, item.wantAccessMode) {
				t.Errorf("%s check failure, want:%v,got:%v", k, item.wantAccessMode, gotAccessModel)
			}
		}
	}
}

func TestIsTargetPathUnderFluidNativeMounts(t *testing.T) {
	testCases := map[string]struct {
		targetPath   string
		mount        datav1alpha1.Mount
		wantIsTarget bool
	}{
		"test is target with mount path case 1": {
			targetPath: "/mnt/data0",
			mount: datav1alpha1.Mount{
				Name:       "spark",
				MountPoint: "local://mnt/data0",
				Path:       "/mnt",
			},
			wantIsTarget: true,
		},
		"test is target with mount path case 2": {
			targetPath: "/mnt/data0",
			mount: datav1alpha1.Mount{
				Name:       "spark",
				MountPoint: "local://mnt/data0",
				Path:       "/mnt/data0",
			},
			wantIsTarget: true,
		},
		"test is target with mount path case 3": {
			targetPath: "/mnt/data0",
			mount: datav1alpha1.Mount{
				Name:       "spark",
				MountPoint: "local://mnt/data0",
				Path:       "/mnt/data0/spark",
			},
			wantIsTarget: false,
		},
		"test is target with mount path case 4": {
			targetPath: "/mnt/data0/spark/part-1",
			mount: datav1alpha1.Mount{
				Name:       "spark",
				MountPoint: "pvc://mnt/data0",
				Path:       "/mnt/data0/spark",
			},
			wantIsTarget: true,
		},
		"test is target with mount path case 5": {
			targetPath: "/mnt/data0",
			mount: datav1alpha1.Mount{
				Name:       "spark",
				MountPoint: "pvc://mnt/data0",
				Path:       "/mnt/data0",
			},
			wantIsTarget: true,
		},
		"test is target with mount path case 6": {
			targetPath: "/mnt/data0",
			mount: datav1alpha1.Mount{
				Name:       "spark",
				MountPoint: "pvc://mnt/data0",
				Path:       "/mnt/data0/spark",
			},
			wantIsTarget: false,
		},
		"test is target without path case 1": {
			targetPath: "/spark/part-1",
			mount: datav1alpha1.Mount{
				Name:       "spark",
				MountPoint: "local://mnt/data0",
			},
			wantIsTarget: true,
		},
		"test is target without path case 2": {
			targetPath: "/sparks/part-1",
			mount: datav1alpha1.Mount{
				Name:       "spark",
				MountPoint: "local://mnt/data0",
			},
			wantIsTarget: false,
		},
		"test is target without path case 3": {
			targetPath: "/spark",
			mount: datav1alpha1.Mount{
				Name:       "spark",
				MountPoint: "local://mnt/data0",
			},
			wantIsTarget: true,
		},
		"test is target without fluid native path case 1": {
			targetPath: "/mnt/spark",
			mount: datav1alpha1.Mount{
				Name:       "spark",
				MountPoint: "http://mnt/data0",
				Path:       "/mnt",
			},
			wantIsTarget: false,
		},
		"test is target without fluid native path case 2": {
			targetPath: "/mnt/spark",
			mount: datav1alpha1.Mount{
				Name:       "spark",
				MountPoint: "https://mnt/data0",
				Path:       "/mnt",
			},
			wantIsTarget: false,
		},
	}

	for k, item := range testCases {
		dataset := mockDatasetWithMountPath("data-set-1", "default", item.mount)
		gotIsTarget := IsTargetPathUnderFluidNativeMounts(item.targetPath, *dataset)
		if gotIsTarget != item.wantIsTarget {
			t.Errorf("%s check failure, want:%t,got:%t", k, item.wantIsTarget, gotIsTarget)
		}
	}
}

func mockDatasetWithMountPath(name, ns string, mount datav1alpha1.Mount) *datav1alpha1.Dataset {
	dataset := &datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{
				mount,
			},
		},
	}
	return dataset
}

func mockDatasetWithAccessModel(name, ns string, accessModel []v1.PersistentVolumeAccessMode) *datav1alpha1.Dataset {
	dataset := &datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: datav1alpha1.DatasetSpec{
			AccessModes: accessModel,
		},
	}
	return dataset
}

func mockDatasetWithCondition(name, ns string, conditions []datav1alpha1.DatasetCondition) *datav1alpha1.Dataset {
	dataset := &datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Status: datav1alpha1.DatasetStatus{
			Conditions: conditions,
		},
	}
	return dataset
}

func TestUpdateMountStatus(t *testing.T) {
	mockDatasetName := "fluid-data-set"
	mockDatasetNamespace := "default"

	testCases := map[string]struct {
		dataset         *datav1alpha1.Dataset
		phase           datav1alpha1.DatasetPhase
		shouldReturnErr bool
	}{
		"Update MountStatus test case 1": {
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mockDatasetName,
					Namespace: mockDatasetNamespace,
				},
			},
			phase:           datav1alpha1.UpdatingDatasetPhase,
			shouldReturnErr: false,
		},
		"Update MountStatus test case 2": {
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mockDatasetName,
					Namespace: mockDatasetNamespace,
				},
			},
			phase:           datav1alpha1.BoundDatasetPhase,
			shouldReturnErr: false,
		},
		"Update MountStatus test case 3": {
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mockDatasetName,
					Namespace: mockDatasetNamespace,
				},
			},
			phase:           datav1alpha1.NotBoundDatasetPhase,
			shouldReturnErr: true,
		},
		"Update MountStatus test case 4": {
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mockDatasetName,
					Namespace: mockDatasetNamespace,
				},
			},
			phase:           datav1alpha1.FailedDatasetPhase,
			shouldReturnErr: true,
		},
		"Update MountStatus test case 5": {
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mockDatasetName,
					Namespace: mockDatasetNamespace,
				},
			},
			phase:           datav1alpha1.NoneDatasetPhase,
			shouldReturnErr: true,
		},
	}
	for k, item := range testCases {
		s := runtime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, item.dataset)
		fakeClient := fake.NewFakeClientWithScheme(s, item.dataset)
		err := UpdateMountStatus(fakeClient, mockDatasetName, mockDatasetNamespace, item.phase)
		if item.phase != datav1alpha1.BoundDatasetPhase && item.phase != datav1alpha1.UpdatingDatasetPhase {
			if err == nil {
				t.Errorf("%s check failure, should not change dataset phase to %s", k, item.phase)
			}
		} else {
			if err != nil {
				t.Errorf("%s check failure", k)
			}
			dataset, err := GetDataset(fakeClient, mockDatasetName, mockDatasetNamespace)
			if err != nil || dataset == nil {
				t.Errorf("%s check failure because cannot get dataset", k)
			} else {
				if dataset.Status.Phase != item.phase {
					t.Errorf("%s check failure, expected %v, got %v", k, item.phase, dataset.Status.Phase)
				}
				if item.phase == datav1alpha1.BoundDatasetPhase && dataset.Status.Conditions[0].Message != "The ddc runtime has updated completely." {
					t.Errorf("%s check failure, expected \"The ddc runtime has updated completely.\", got %v", k, dataset.Status.Conditions[0].Message)
				}
				if item.phase == datav1alpha1.UpdatingDatasetPhase && dataset.Status.Conditions[0].Message != "The ddc runtime is updating." {
					t.Errorf("%s check failure, expected \"The ddc runtime is updating.\", got %v", k, dataset.Status.Conditions[0].Message)
				}
			}

		}

	}

}

func TestUfsToUpdate(t *testing.T) {
	testCases := map[string]struct {
		dataset      *datav1alpha1.Dataset
		wantToAdd    []string
		wantToRemove []string
		wantUpdate   bool
	}{
		"ufsToUpdate test case 1": {
			dataset: &datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							Name: "hbase",
						},
						{
							Name: "spark",
						},
					},
				},
				Status: datav1alpha1.DatasetStatus{
					Mounts: []datav1alpha1.Mount{},
				},
			},
			wantToAdd:    []string{"/hbase", "/spark"},
			wantToRemove: []string{},
			wantUpdate:   true,
		},
		"ufsToUpdate test case 2": {
			dataset: &datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							Name: "hbase",
						},
					},
				},
				Status: datav1alpha1.DatasetStatus{
					Mounts: []datav1alpha1.Mount{
						{
							Name: "spark",
						},
					},
				},
			},
			wantToAdd:    []string{"/hbase"},
			wantToRemove: []string{"/spark"},
			wantUpdate:   true,
		},
		"ufsToUpdate test case 3": {
			dataset: &datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							Name: "hbase",
						},
					},
				},
				Status: datav1alpha1.DatasetStatus{
					Mounts: []datav1alpha1.Mount{
						{
							Name: "spark",
						},
						{
							Name: "hbase",
						},
						{
							Name: "hadoop",
						},
					},
				},
			},
			wantToAdd:    []string{},
			wantToRemove: []string{"/spark", "/hadoop"},
			wantUpdate:   true,
		},
		"ufsToUpdate test case 4": {
			dataset: &datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							Name: "spark",
						},
						{
							Name: "hbase",
						},
						{
							Name: "hadoop",
						},
					},
				},
				Status: datav1alpha1.DatasetStatus{
					Mounts: []datav1alpha1.Mount{
						{
							Name: "spark",
						},
						{
							Name: "hbase",
						},
						{
							Name: "hadoop",
						},
					},
				},
			},
			wantToAdd:    []string{},
			wantToRemove: []string{},
			wantUpdate:   false,
		},
	}
	for k, item := range testCases {
		ufsToUpdate := NewUFSToUpdate(item.dataset)
		ufsToUpdate.AnalyzePathsDelta()
		if len(item.wantToAdd) != 0 || len(ufsToUpdate.ToAdd()) != 0 {
			if !reflect.DeepEqual(item.wantToAdd, ufsToUpdate.ToAdd()) {
				t.Errorf("%s check fail, wantToAdd is %s, get %s", k, item.wantToAdd, ufsToUpdate.ToAdd())
			}
		}
		if len(item.wantToRemove) != 0 || len(ufsToUpdate.ToRemove()) != 0 {
			if !reflect.DeepEqual(item.wantToRemove, ufsToUpdate.ToRemove()) {
				t.Errorf("%s check fail, wantToRemove is %s, get %s", k, item.wantToRemove, ufsToUpdate.ToRemove())
			}
		}
		if item.wantUpdate != ufsToUpdate.ShouldUpdate() {
			t.Errorf("%s check fail, shouldUpdate is %v, get %v", k, item.wantUpdate, ufsToUpdate.ShouldUpdate())
		}

	}
}

func TestAddMountPaths(t *testing.T) {
	testCases := []struct {
		originAdd []string
		toAdd     []string
		result    []string
	}{
		{
			originAdd: []string{"/path1"},
			toAdd:     []string{"/path2"},
			result:    []string{"/path1", "/path2"},
		},
		{
			originAdd: []string{"/path1"},
			toAdd:     []string{"/path1"},
			result:    []string{"/path1"},
		},
		{
			originAdd: []string{},
			toAdd:     []string{"/path1"},
			result:    []string{"/path1"},
		},
		{
			originAdd: []string{"/path1"},
			toAdd:     []string{},
			result:    []string{"/path1"},
		},
	}

	for k, item := range testCases {
		ufsToUpdate := NewUFSToUpdate(&datav1alpha1.Dataset{})
		ufsToUpdate.toAdd = item.originAdd

		ufsToUpdate.AddMountPaths(item.toAdd)

		if len(item.result) != 0 || len(ufsToUpdate.ToAdd()) != 0 {
			if !reflect.DeepEqual(item.result, ufsToUpdate.ToAdd()) {
				t.Errorf("%d check fail, expected mountpath is %s, get %s", k, item.result, ufsToUpdate.ToAdd())
			}
		}
	}
}
