/*
Copyright 2023 The Fluid Author.

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
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	mockDatasetName      = "fluid-data-set"
	mockDatasetNamespace = "default"
	mockDataset1Name     = "dataset-1"
	mockMountPathSpark   = "/mnt/data0"
	mockMountNameSpark   = "spark"
	mockMountNameHbase   = "hbase"
	mockMountNameHadoop  = "hadoop"
	mockPath1            = "/path1"
	mockPath2            = "/path2"
)

var _ = Describe("GetDataset", func() {
	var (
		scheme      *runtime.Scheme
		initDataset *datav1alpha1.Dataset
	)

	BeforeEach(func() {
		initDataset = &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      mockDatasetName,
				Namespace: mockDatasetNamespace,
			},
		}
		scheme = runtime.NewScheme()
		scheme.AddKnownTypes(datav1alpha1.GroupVersion, initDataset)
	})

	Context("when dataset exists", func() {
		It("should return the dataset successfully", func() {
			fakeClient := fake.NewFakeClientWithScheme(scheme, initDataset)

			gotDataset, err := GetDataset(fakeClient, mockDatasetName, mockDatasetNamespace)

			Expect(err).NotTo(HaveOccurred())
			Expect(gotDataset).NotTo(BeNil())
			Expect(gotDataset.Name).To(Equal(mockDatasetName))
		})
	})

	Context("when dataset does not exist", func() {
		It("should return an error", func() {
			fakeClient := fake.NewFakeClientWithScheme(scheme, initDataset)

			gotDataset, err := GetDataset(fakeClient, mockDatasetName+"not-exist", mockDatasetNamespace)

			Expect(err).To(HaveOccurred())
			Expect(gotDataset).To(BeNil())
		})
	})
})

var _ = Describe("IsSetupDone", func() {
	DescribeTable("should correctly determine if dataset setup is done",
		func(conditions []datav1alpha1.DatasetCondition, wantDone bool) {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mockDataset1Name,
					Namespace: mockDatasetNamespace,
				},
				Status: datav1alpha1.DatasetStatus{
					Conditions: conditions,
				},
			}

			gotDone := IsSetupDone(dataset)

			Expect(gotDone).To(Equal(wantDone))
		},
		Entry("dataset is ready",
			[]datav1alpha1.DatasetCondition{
				{Type: datav1alpha1.DatasetReady},
			},
			true,
		),
		Entry("dataset is only initialized",
			[]datav1alpha1.DatasetCondition{
				{Type: datav1alpha1.DatasetInitialized},
			},
			false,
		),
		Entry("dataset has no conditions",
			nil,
			false,
		),
	)
})

var _ = Describe("GetAccessModesOfDataset", func() {
	DescribeTable("should return correct access modes",
		func(name string, getName string, accessModes []corev1.PersistentVolumeAccessMode, wantAccessModes []corev1.PersistentVolumeAccessMode, notFound bool) {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: mockDatasetNamespace,
				},
				Spec: datav1alpha1.DatasetSpec{
					AccessModes: accessModes,
				},
			}
			scheme := runtime.NewScheme()
			scheme.AddKnownTypes(datav1alpha1.GroupVersion, dataset)
			fakeClient := fake.NewFakeClientWithScheme(scheme, dataset)

			gotAccessModes, err := GetAccessModesOfDataset(fakeClient, getName, mockDatasetNamespace)

			if notFound {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(gotAccessModes).To(Equal(wantAccessModes))
			}
		},
		Entry("dataset with ReadWriteMany access mode",
			mockDataset1Name,
			mockDataset1Name,
			[]corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			[]corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			false,
		),
		Entry("dataset with no access mode defaults to ReadOnlyMany",
			mockDataset1Name,
			mockDataset1Name,
			nil,
			[]corev1.PersistentVolumeAccessMode{corev1.ReadOnlyMany},
			false,
		),
		Entry("dataset not found",
			mockDataset1Name,
			mockDataset1Name+"-notexist",
			nil,
			nil,
			true,
		),
	)
})

var _ = Describe("GetPVCStorageCapacityOfDataset", func() {
	DescribeTable("should return correct storage capacity",
		func(name string, getName string, storageCapacity string, wantStorageCapacity resource.Quantity, notFound bool) {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:        name,
					Namespace:   mockDatasetNamespace,
					Annotations: map[string]string{PVCStorageAnnotation: storageCapacity},
				},
			}
			scheme := runtime.NewScheme()
			scheme.AddKnownTypes(datav1alpha1.GroupVersion, dataset)
			fakeClient := fake.NewFakeClientWithScheme(scheme, dataset)

			gotStorageCapacity, err := GetPVCStorageCapacityOfDataset(fakeClient, getName, mockDatasetNamespace)

			if notFound {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(gotStorageCapacity).To(Equal(wantStorageCapacity))
			}
		},
		Entry("dataset with empty storage capacity defaults to 100Pi",
			mockDataset1Name,
			mockDataset1Name,
			"",
			resource.MustParse("100Pi"),
			false,
		),
		Entry("dataset with 1Gi storage capacity",
			mockDataset1Name,
			mockDataset1Name,
			"1Gi",
			resource.MustParse("1Gi"),
			false,
		),
		Entry("dataset not found",
			mockDataset1Name,
			mockDataset1Name+"-notexist",
			"",
			resource.Quantity{},
			true,
		),
		Entry("dataset with invalid storage capacity format defaults to 100Pi",
			mockDataset1Name,
			mockDataset1Name,
			"formatError",
			resource.MustParse("100Pi"),
			false,
		),
	)
})

var _ = Describe("IsTargetPathUnderFluidNativeMounts", func() {
	DescribeTable("should correctly determine if target path is under fluid native mounts",
		func(targetPath string, mount datav1alpha1.Mount, wantIsTarget bool) {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mockDataset1Name,
					Namespace: mockDatasetNamespace,
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{mount},
				},
			}

			gotIsTarget := IsTargetPathUnderFluidNativeMounts(targetPath, *dataset)

			Expect(gotIsTarget).To(Equal(wantIsTarget))
		},
		Entry("local scheme with path containing target",
			mockMountPathSpark,
			datav1alpha1.Mount{
				Name:       mockMountNameSpark,
				MountPoint: "local://mnt/data0",
				Path:       "/mnt",
			},
			true,
		),
		Entry("local scheme with exact path match",
			mockMountPathSpark,
			datav1alpha1.Mount{
				Name:       mockMountNameSpark,
				MountPoint: "local://mnt/data0",
				Path:       mockMountPathSpark,
			},
			true,
		),
		Entry("local scheme with path deeper than target",
			mockMountPathSpark,
			datav1alpha1.Mount{
				Name:       mockMountNameSpark,
				MountPoint: "local://mnt/data0",
				Path:       "/mnt/data0/spark",
			},
			false,
		),
		Entry("pvc scheme with subpath of mount",
			"/mnt/data0/spark/part-1",
			datav1alpha1.Mount{
				Name:       mockMountNameSpark,
				MountPoint: "pvc://mnt/data0",
				Path:       "/mnt/data0/spark",
			},
			true,
		),
		Entry("pvc scheme with exact path match",
			mockMountPathSpark,
			datav1alpha1.Mount{
				Name:       mockMountNameSpark,
				MountPoint: "pvc://mnt/data0",
				Path:       mockMountPathSpark,
			},
			true,
		),
		Entry("pvc scheme with path deeper than target",
			mockMountPathSpark,
			datav1alpha1.Mount{
				Name:       mockMountNameSpark,
				MountPoint: "pvc://mnt/data0",
				Path:       "/mnt/data0/spark",
			},
			false,
		),
		Entry("local scheme without path subpath match",
			"/spark/part-1",
			datav1alpha1.Mount{
				Name:       mockMountNameSpark,
				MountPoint: "local://mnt/data0",
			},
			true,
		),
		Entry("local scheme without path no match",
			"/sparks/part-1",
			datav1alpha1.Mount{
				Name:       mockMountNameSpark,
				MountPoint: "local://mnt/data0",
			},
			false,
		),
		Entry("local scheme without path exact match",
			"/spark",
			datav1alpha1.Mount{
				Name:       mockMountNameSpark,
				MountPoint: "local://mnt/data0",
			},
			true,
		),
		Entry("http scheme not fluid native",
			"/mnt/spark",
			datav1alpha1.Mount{
				Name:       mockMountNameSpark,
				MountPoint: "http://mnt/data0",
				Path:       "/mnt",
			},
			false,
		),
		Entry("https scheme not fluid native",
			"/mnt/spark",
			datav1alpha1.Mount{
				Name:       mockMountNameSpark,
				MountPoint: "https://mnt/data0",
				Path:       "/mnt",
			},
			false,
		),
	)
})

var _ = Describe("UpdateMountStatus", func() {
	var (
		scheme *runtime.Scheme
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
	})

	Context("when updating to UpdatingDatasetPhase", func() {
		It("should update status successfully", func() {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mockDatasetName,
					Namespace: mockDatasetNamespace,
				},
			}
			scheme.AddKnownTypes(datav1alpha1.GroupVersion, dataset)
			fakeClient := fake.NewFakeClientWithScheme(scheme, dataset)

			err := UpdateMountStatus(fakeClient, mockDatasetName, mockDatasetNamespace, datav1alpha1.UpdatingDatasetPhase)

			Expect(err).NotTo(HaveOccurred())
			updatedDataset, getErr := GetDataset(fakeClient, mockDatasetName, mockDatasetNamespace)
			Expect(getErr).NotTo(HaveOccurred())
			Expect(updatedDataset.Status.Phase).To(Equal(datav1alpha1.UpdatingDatasetPhase))
			Expect(updatedDataset.Status.Conditions[0].Message).To(Equal("The ddc runtime is updating."))
		})
	})

	Context("when updating to BoundDatasetPhase", func() {
		It("should update status successfully", func() {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mockDatasetName,
					Namespace: mockDatasetNamespace,
				},
			}
			scheme.AddKnownTypes(datav1alpha1.GroupVersion, dataset)
			fakeClient := fake.NewFakeClientWithScheme(scheme, dataset)

			err := UpdateMountStatus(fakeClient, mockDatasetName, mockDatasetNamespace, datav1alpha1.BoundDatasetPhase)

			Expect(err).NotTo(HaveOccurred())
			updatedDataset, getErr := GetDataset(fakeClient, mockDatasetName, mockDatasetNamespace)
			Expect(getErr).NotTo(HaveOccurred())
			Expect(updatedDataset.Status.Phase).To(Equal(datav1alpha1.BoundDatasetPhase))
			Expect(updatedDataset.Status.Conditions[0].Message).To(Equal("The ddc runtime has updated completely."))
		})
	})

	DescribeTable("should return error for unsupported phases",
		func(phase datav1alpha1.DatasetPhase) {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mockDatasetName,
					Namespace: mockDatasetNamespace,
				},
			}
			scheme.AddKnownTypes(datav1alpha1.GroupVersion, dataset)
			fakeClient := fake.NewFakeClientWithScheme(scheme, dataset)

			err := UpdateMountStatus(fakeClient, mockDatasetName, mockDatasetNamespace, phase)

			Expect(err).To(HaveOccurred())
		},
		Entry("NotBoundDatasetPhase", datav1alpha1.NotBoundDatasetPhase),
		Entry("FailedDatasetPhase", datav1alpha1.FailedDatasetPhase),
		Entry("NoneDatasetPhase", datav1alpha1.NoneDatasetPhase),
	)
})

var _ = Describe("UFSToUpdate", func() {
	Describe("AnalyzePathsDelta", func() {
		DescribeTable("should correctly analyze paths delta",
			func(specMounts []datav1alpha1.Mount, statusMounts []datav1alpha1.Mount, wantToAdd []string, wantToRemove []string, wantUpdate bool) {
				dataset := &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{
						Mounts: specMounts,
					},
					Status: datav1alpha1.DatasetStatus{
						Mounts: statusMounts,
					},
				}

				ufsToUpdate := NewUFSToUpdate(dataset)
				ufsToUpdate.AnalyzePathsDelta()

				Expect(ufsToUpdate.ToAdd()).To(ConsistOf(wantToAdd))
				Expect(ufsToUpdate.ToRemove()).To(ConsistOf(wantToRemove))
				Expect(ufsToUpdate.ShouldUpdate()).To(Equal(wantUpdate))
			},
			Entry("add new mounts",
				[]datav1alpha1.Mount{
					{Name: mockMountNameHbase},
					{Name: mockMountNameSpark},
				},
				[]datav1alpha1.Mount{},
				[]string{"/hbase", "/spark"},
				[]string{},
				true,
			),
			Entry("add and remove mounts",
				[]datav1alpha1.Mount{
					{Name: mockMountNameHbase},
				},
				[]datav1alpha1.Mount{
					{Name: mockMountNameSpark},
				},
				[]string{"/hbase"},
				[]string{"/spark"},
				true,
			),
			Entry("remove multiple mounts",
				[]datav1alpha1.Mount{
					{Name: mockMountNameHbase},
				},
				[]datav1alpha1.Mount{
					{Name: mockMountNameSpark},
					{Name: mockMountNameHbase},
					{Name: mockMountNameHadoop},
				},
				[]string{},
				[]string{"/spark", "/hadoop"},
				true,
			),
			Entry("no changes needed",
				[]datav1alpha1.Mount{
					{Name: mockMountNameSpark},
					{Name: mockMountNameHbase},
					{Name: mockMountNameHadoop},
				},
				[]datav1alpha1.Mount{
					{Name: mockMountNameSpark},
					{Name: mockMountNameHbase},
					{Name: mockMountNameHadoop},
				},
				[]string{},
				[]string{},
				false,
			),
		)
	})
})

var _ = Describe("AddMountPaths", func() {
	DescribeTable("should correctly add mount paths",
		func(originAdd []string, toAdd []string, result []string) {
			ufsToUpdate := NewUFSToUpdate(&datav1alpha1.Dataset{})
			// Directly set unexported field to isolate testing of AddMountPaths logic
			ufsToUpdate.toAdd = originAdd

			ufsToUpdate.AddMountPaths(toAdd)

			if len(result) == 0 {
				Expect(ufsToUpdate.ToAdd()).To(BeEmpty())
			} else {
				Expect(ufsToUpdate.ToAdd()).To(ConsistOf(result))
			}
		},
		Entry("add new path to existing",
			[]string{mockPath1},
			[]string{mockPath2},
			[]string{mockPath1, mockPath2},
		),
		Entry("add duplicate path",
			[]string{mockPath1},
			[]string{mockPath1},
			[]string{mockPath1},
		),
		Entry("add to empty",
			[]string{},
			[]string{mockPath1},
			[]string{mockPath1},
		),
		Entry("add empty to existing",
			[]string{mockPath1},
			[]string{},
			[]string{mockPath1},
		),
	)
})
