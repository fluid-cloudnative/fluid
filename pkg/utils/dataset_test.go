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

func TestGetUFSToUpdate(t *testing.T) {
	testCases := map[string]struct {
		dataset    *datav1alpha1.Dataset
		specAdd    []string
		specRemove []string
	}{
		"get UpdateUFSMap test case 1": {
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
			specAdd:    []string{"/hbase", "/spark"},
			specRemove: []string{},
		},
		"get UpdateUFSMap test case 2": {
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
			specAdd:    []string{"/hbase"},
			specRemove: []string{"/spark"},
		},
		"get UpdateUFSMap test case 3": {
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
					Mounts: []datav1alpha1.Mount{
						{
							Name: "hbase",
						},
						{
							Name: "spark",
						},
						{
							Name: "hadoop",
						},
						{
							Name: "test",
						},
					},
				},
			},
			specAdd:    []string{},
			specRemove: []string{"/hadoop", "/test"},
		},
	}
	for k, item := range testCases {
		ufsToUpdate := GetUFSToUpdate(item.dataset)

		if !(len(ufsToUpdate.ToAdd) == 0 && len(item.specAdd) == 0) {
			if !reflect.DeepEqual(ufsToUpdate.ToAdd, item.specAdd) {
				t.Errorf("%s check failure, got ToBeAdded mountPaths %s,want %s", k, ufsToUpdate.ToAdd, item.specAdd)
			}
		}
		if !(len(ufsToUpdate.ToRemove) == 0 && len(item.specRemove) == 0) {
			if !reflect.DeepEqual(ufsToUpdate.ToRemove, item.specRemove) {
				t.Errorf("%s check failure, got ToBeRemoved mountPaths %s,want %s", k, ufsToUpdate.ToRemove, item.specRemove)
			}
		}
	}
}

func TestGetMounts(t *testing.T) {
	testCases := map[string]struct {
		dataset          *datav1alpha1.Dataset
		specMountsWant   []string
		statusMountsWant []string
	}{
		"get Mounts test case 1": {
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
			specMountsWant: []string{
				"/hbase",
				"/spark",
			},
			statusMountsWant: []string{},
		},
		"get Mounts test case 2": {
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
					Mounts: []datav1alpha1.Mount{
						{
							Name: "spark",
						},
					},
				},
			},
			specMountsWant: []string{
				"/hbase",
				"/spark",
			},
			statusMountsWant: []string{
				"/spark",
			},
		},
		"get Mounts test case 3": {
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
			specMountsWant: []string{
				"/hbase",
			},
			statusMountsWant: []string{
				"/spark",
				"/hbase",
				"/hadoop",
			},
		},
	}

	for k, item := range testCases {
		specMountPaths, mountedMountPaths := getMounts(item.dataset)
		if !(len(specMountPaths) == 0 && len(item.specMountsWant) == 0) {
			if !reflect.DeepEqual(specMountPaths, item.specMountsWant) {
				t.Errorf("%s check failure, got spec mountPaths %s,want %s", k, specMountPaths, item.specMountsWant)
			}
		}
		if !(len(mountedMountPaths) == 0 && len(item.statusMountsWant) == 0) {
			if !reflect.DeepEqual(mountedMountPaths, item.statusMountsWant) {
				t.Errorf("%s check failure, got mounted mountPaths %s,want %s", k, mountedMountPaths, item.statusMountsWant)
			}
		}

	}
}

func TestCalculateMountPointsChanges(t *testing.T) {
	testCases := map[string]struct {
		specMountPaths    []string
		mountedMountPaths []string
		expectAdd         []string
		expectRemove      []string
	}{
		"calculate mount point changes test case 1": {
			specMountPaths:    []string{"hadoop3.3.0"},
			mountedMountPaths: []string{"hadoopCurrent", "hadoop3.3.0"},
			expectAdd:         []string{},
			expectRemove:      []string{"hadoopCurrent"},
		},
		"calculate mount point changes test case 2": {
			specMountPaths:    []string{"hadoop3.3.0", "hadoop3.3.0"},
			mountedMountPaths: []string{"hadoopCurrent"},
			expectAdd:         []string{"hadoop3.3.0", "hadoop3.3.0"},
			expectRemove:      []string{"hadoopCurrent"},
		},
		"calculate mount point changes test case 3": {
			specMountPaths:    []string{"hadoop3.3.0", "hadoop3.2.2"},
			mountedMountPaths: []string{"hadoopCurrent", "hadoop3.2.2"},
			expectAdd:         []string{"hadoop3.3.0"},
			expectRemove:      []string{"hadoopCurrent"},
		},
		"calculate mount point changes test case 4": {
			specMountPaths:    []string{"hadoop3.3.0"},
			mountedMountPaths: []string{"hadoop3.3.0"},
			expectAdd:         []string{},
			expectRemove:      []string{},
		},
		"calculate mount point changes test case 5": {
			specMountPaths:    []string{"hadoop3.3.0", "hadoop3.2.2", "hadoop3.3.1"},
			mountedMountPaths: []string{"hadoopCurrent", "hadoop3.2.2"},
			expectAdd:         []string{"hadoop3.3.0", "hadoop3.3.1"},
			expectRemove:      []string{"hadoopCurrent"},
		},
	}

	for k, item := range testCases {
		toAdd, toRemove := calculateMountPointsChanges(item.specMountPaths, item.mountedMountPaths)

		if !(len(toAdd) == 0 && len(item.expectAdd) == 0) {
			if !reflect.DeepEqual(toAdd, item.expectAdd) {
				t.Errorf("%s check failure, expected added %v, got %v", k, item.expectAdd, toAdd)
			}
		}
		if !(len(toRemove) == 0 && len(item.expectRemove) == 0) {
			if !reflect.DeepEqual(toRemove, item.expectRemove) {
				t.Errorf("%s check failure, expected removed %v, got %v", k, item.expectRemove, toRemove)
			}
		}

	}

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
