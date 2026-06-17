/*
Copyright 2026 The Fluid Authors.

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

package v1alpha1

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("DataMigrate types", func() {
	Describe("scheme registration", func() {
		It("registers DataMigrate and DataMigrateList with the package group version", func() {
			dataMigrateGVK, err := apiGVKFor(&DataMigrate{})
			Expect(err).NotTo(HaveOccurred())
			Expect(dataMigrateGVK).To(Equal(GroupVersion.WithKind("DataMigrate")))

			dataMigrateListGVK, err := apiGVKFor(&DataMigrateList{})
			Expect(err).NotTo(HaveOccurred())
			Expect(dataMigrateListGVK).To(Equal(GroupVersion.WithKind("DataMigrateList")))
		})
	})

	Describe("DeepCopyObject", func() {
		It("returns a distinct runtime object for DataMigrate and DataMigrateList", func() {
			dataMigrate := &DataMigrate{
				TypeMeta:   metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: "DataMigrate"},
				ObjectMeta: metav1.ObjectMeta{Name: "example", Namespace: "fluid"},
				Spec: DataMigrateSpec{
					From:            DataToMigrate{DataSet: &DatasetToMigrate{Name: "source", Namespace: "fluid", Path: "/raw"}},
					To:              DataToMigrate{ExternalStorage: &ExternalStorage{URI: "s3://bucket/archive"}},
					RuntimeType:     "alluxio",
					Options:         map[string]string{"bandwidth": "high"},
					Policy:          Once,
					Parallelism:     4,
					ParallelOptions: map[string]string{"sshPort": "22"},
				},
				Status: OperationStatus{Phase: common.PhaseComplete, Duration: "2m"},
			}

			copiedObject := dataMigrate.DeepCopyObject()
			copiedDataMigrate, ok := copiedObject.(*DataMigrate)
			Expect(ok).To(BeTrue())
			Expect(copiedDataMigrate).NotTo(BeIdenticalTo(dataMigrate))
			Expect(copiedDataMigrate.Spec.From).To(Equal(dataMigrate.Spec.From))
			// Verify deep copy of nested pointers.
			Expect(copiedDataMigrate.Spec.From.DataSet).NotTo(BeIdenticalTo(dataMigrate.Spec.From.DataSet))
			Expect(copiedDataMigrate.Spec.To).To(Equal(dataMigrate.Spec.To))
			Expect(copiedDataMigrate.Spec.To.ExternalStorage).NotTo(BeIdenticalTo(dataMigrate.Spec.To.ExternalStorage))
			Expect(copiedDataMigrate.Status).To(Equal(dataMigrate.Status))

			dataMigrateList := &DataMigrateList{Items: []DataMigrate{*dataMigrate}}
			copiedListObject := dataMigrateList.DeepCopyObject()
			copiedList, ok := copiedListObject.(*DataMigrateList)
			Expect(ok).To(BeTrue())
			Expect(copiedList).NotTo(BeIdenticalTo(dataMigrateList))
			Expect(copiedList.Items).To(HaveLen(1))
			Expect(copiedList.Items[0].Spec.RuntimeType).To(Equal("alluxio"))
		})
	})

	Describe("representative spec and status construction", func() {
		It("captures migration endpoints, workflow dependency, and launcher settings", func() {
			ttlSeconds := int32(300)
			dataMigrate := DataMigrate{
				Spec: DataMigrateSpec{
					From:        DataToMigrate{DataSet: &DatasetToMigrate{Name: "source", Namespace: "fluid", Path: "/images"}},
					To:          DataToMigrate{DataSet: &DatasetToMigrate{Name: "target", Namespace: "fluid", Path: "/archive"}},
					Block:       true,
					RuntimeType: "alluxio",
					Options:     map[string]string{"overwrite": "true"},
					RunAfter: &OperationRef{ObjectRef: ObjectRef{
						Kind: "DataLoad",
						Name: "load-source",
					}},
					TTLSecondsAfterFinished: &ttlSeconds,
					Parallelism:             2,
					ParallelOptions:         map[string]string{"sshSecret": "migrate-ssh"},
				},
				Status: OperationStatus{
					Phase:    common.PhaseComplete,
					Duration: "5m",
					Infos:    map[string]string{"launcher": "parallel"},
				},
			}

			Expect(dataMigrate.Spec.From.DataSet).NotTo(BeNil())
			Expect(*dataMigrate.Spec.From.DataSet).To(Equal(DatasetToMigrate{Name: "source", Namespace: "fluid", Path: "/images"}))
			Expect(dataMigrate.Spec.To.DataSet).NotTo(BeNil())
			Expect(*dataMigrate.Spec.To.DataSet).To(Equal(DatasetToMigrate{Name: "target", Namespace: "fluid", Path: "/archive"}))
			Expect(dataMigrate.Spec.RunAfter).NotTo(BeNil())
			Expect(dataMigrate.Spec.RunAfter.ObjectRef).To(Equal(ObjectRef{Kind: "DataLoad", Name: "load-source"}))
			Expect(dataMigrate.Spec.Parallelism).To(Equal(int32(2)))
			Expect(dataMigrate.Status.Infos).To(HaveKeyWithValue("launcher", "parallel"))
			Expect(dataMigrate.Status.Phase).To(Equal(common.PhaseComplete))
		})
	})
})
