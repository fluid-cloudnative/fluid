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

var _ = Describe("DataBackup types", func() {
	Describe("scheme registration", func() {
		It("registers DataBackup and DataBackupList with the package group version", func() {
			dataBackupGVK, err := apiGVKFor(&DataBackup{})
			Expect(err).NotTo(HaveOccurred())
			Expect(dataBackupGVK).To(Equal(GroupVersion.WithKind("DataBackup")))

			dataBackupListGVK, err := apiGVKFor(&DataBackupList{})
			Expect(err).NotTo(HaveOccurred())
			Expect(dataBackupListGVK).To(Equal(GroupVersion.WithKind("DataBackupList")))
		})
	})

	Describe("DeepCopyObject", func() {
		It("returns a distinct runtime object for DataBackup and DataBackupList", func() {
			uid := int64(1000)
			gid := int64(1000)
			dataBackup := &DataBackup{
				TypeMeta:   metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: "DataBackup"},
				ObjectMeta: metav1.ObjectMeta{Name: "example", Namespace: "fluid"},
				Spec: DataBackupSpec{
					Dataset:    "imagenet",
					BackupPath: "s3://bucket/checkpoints",
					RunAs:      &User{UID: &uid, GID: &gid},
					RunAfter: &OperationRef{ObjectRef: ObjectRef{
						Kind: "DataLoad",
						Name: "prepare-backup",
					}},
				},
				Status: OperationStatus{Phase: common.PhaseComplete, Duration: "30s"},
			}

			copiedObject := dataBackup.DeepCopyObject()
			copiedDataBackup, ok := copiedObject.(*DataBackup)
			Expect(ok).To(BeTrue())
			Expect(copiedDataBackup).NotTo(BeIdenticalTo(dataBackup))
			Expect(copiedDataBackup.Spec).To(Equal(dataBackup.Spec))
			// Verify deep copy of nested pointers.
			if dataBackup.Spec.RunAs != nil {
				Expect(copiedDataBackup.Spec.RunAs).NotTo(BeIdenticalTo(dataBackup.Spec.RunAs))
			}
			if dataBackup.Spec.RunAfter != nil {
				Expect(copiedDataBackup.Spec.RunAfter).NotTo(BeIdenticalTo(dataBackup.Spec.RunAfter))
			}
			Expect(copiedDataBackup.Status).To(Equal(dataBackup.Status))
			if dataBackup.Spec.RunAs != nil {
				if dataBackup.Spec.RunAs.UID != nil {
					Expect(copiedDataBackup.Spec.RunAs.UID).NotTo(BeIdenticalTo(dataBackup.Spec.RunAs.UID))
				}
				if dataBackup.Spec.RunAs.GID != nil {
					Expect(copiedDataBackup.Spec.RunAs.GID).NotTo(BeIdenticalTo(dataBackup.Spec.RunAs.GID))
				}
			}

			dataBackupList := &DataBackupList{Items: []DataBackup{*dataBackup}}
			copiedListObject := dataBackupList.DeepCopyObject()
			copiedList, ok := copiedListObject.(*DataBackupList)
			Expect(ok).To(BeTrue())
			Expect(copiedList).NotTo(BeIdenticalTo(dataBackupList))
			Expect(copiedList.Items).To(HaveLen(1))
			Expect(copiedList.Items[0].Spec.BackupPath).To(Equal("s3://bucket/checkpoints"))
		})
	})

	Describe("representative spec and status construction", func() {
		It("captures backup target, workflow dependency, and backup location info", func() {
			ttlSeconds := int32(120)
			dataBackup := DataBackup{
				Spec: DataBackupSpec{
					Dataset:    "imagenet",
					BackupPath: "oss://archive/imagenet",
					RunAs:      &User{UserName: "fluid"},
					RunAfter: &OperationRef{ObjectRef: ObjectRef{
						Kind: "DataLoad",
						Name: "freeze-dataset",
					}},
					TTLSecondsAfterFinished: &ttlSeconds,
				},
				Status: OperationStatus{
					Phase:    common.PhaseComplete,
					Duration: "3m",
					Infos: map[string]string{
						"BackupLocationPath":     "/archive/imagenet",
						"BackupLocationNodeName": "worker-0",
					},
				},
			}

			Expect(dataBackup.Spec.Dataset).To(Equal("imagenet"))
			Expect(dataBackup.Spec.RunAfter).NotTo(BeNil())
			Expect(dataBackup.Spec.RunAfter.ObjectRef).To(Equal(ObjectRef{Kind: "DataLoad", Name: "freeze-dataset"}))
			Expect(dataBackup.Spec.RunAs).NotTo(BeNil())
			Expect(dataBackup.Spec.RunAs.UserName).To(Equal("fluid"))
			Expect(dataBackup.Status.Phase).To(Equal(common.PhaseComplete))
			Expect(dataBackup.Status.Infos).To(HaveKeyWithValue("BackupLocationNodeName", "worker-0"))
		})
	})
})
