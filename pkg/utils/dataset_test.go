/*

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
package utils

import (
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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
