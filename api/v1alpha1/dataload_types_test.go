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

var _ = Describe("DataLoad types", func() {
	Describe("scheme registration", func() {
		It("registers DataLoad and DataLoadList with the package group version", func() {
			dataLoadGVK, err := apiGVKFor(&DataLoad{})
			Expect(err).NotTo(HaveOccurred())
			Expect(dataLoadGVK).To(Equal(GroupVersion.WithKind("DataLoad")))

			dataLoadListGVK, err := apiGVKFor(&DataLoadList{})
			Expect(err).NotTo(HaveOccurred())
			Expect(dataLoadListGVK).To(Equal(GroupVersion.WithKind("DataLoadList")))
		})
	})

	Describe("DeepCopyObject", func() {
		It("returns a distinct runtime object for DataLoad and DataLoadList", func() {
			dataLoad := &DataLoad{
				TypeMeta:   metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: "DataLoad"},
				ObjectMeta: metav1.ObjectMeta{Name: "example", Namespace: "fluid"},
				Spec: DataLoadSpec{
					Dataset: TargetDataset{Name: "dataset", Namespace: "fluid"},
					Target:  []TargetPath{{Path: "/training", Replicas: 2}},
					Options: map[string]string{"threads": "8"},
					RunAfter: &OperationRef{ObjectRef: ObjectRef{
						Kind: "DataLoad",
						Name: "prepare",
					}},
					Policy: Once,
				},
				Status: OperationStatus{Phase: common.PhaseComplete, Duration: "1m"},
			}

			copiedObject := dataLoad.DeepCopyObject()
			copiedDataLoad, ok := copiedObject.(*DataLoad)
			Expect(ok).To(BeTrue())
			Expect(copiedDataLoad).NotTo(BeIdenticalTo(dataLoad))
			Expect(copiedDataLoad.Spec).To(Equal(dataLoad.Spec))
			// Verify deep copy of nested pointers.
			if dataLoad.Spec.RunAfter != nil {
				Expect(copiedDataLoad.Spec.RunAfter).NotTo(BeIdenticalTo(dataLoad.Spec.RunAfter))
			}
			Expect(copiedDataLoad.Status).To(Equal(dataLoad.Status))

			dataLoadList := &DataLoadList{Items: []DataLoad{*dataLoad}}
			copiedListObject := dataLoadList.DeepCopyObject()
			copiedList, ok := copiedListObject.(*DataLoadList)
			Expect(ok).To(BeTrue())
			Expect(copiedList).NotTo(BeIdenticalTo(dataLoadList))
			Expect(copiedList.Items).To(HaveLen(1))
			Expect(copiedList.Items[0].Spec.Dataset).To(Equal(dataLoad.Spec.Dataset))
		})
	})

	Describe("representative spec and status construction", func() {
		It("captures the dataset target, workflow dependency, and operation progress", func() {
			ttlSeconds := int32(60)
			dataLoad := DataLoad{
				Spec: DataLoadSpec{
					Dataset:      TargetDataset{Name: "imagenet", Namespace: "fluid"},
					LoadMetadata: true,
					Target:       []TargetPath{{Path: "/train", Replicas: 2}},
					Options:      map[string]string{"format": "csv"},
					RunAfter: &OperationRef{ObjectRef: ObjectRef{
						Kind: "DataLoad",
						Name: "prepare",
					}},
					TTLSecondsAfterFinished: &ttlSeconds,
					Policy:                  Once,
				},
				Status: OperationStatus{
					Phase:    common.PhaseComplete,
					Duration: "45s",
					Infos: map[string]string{
						"cached": "true",
					},
				},
			}

			Expect(dataLoad.Spec.Dataset.Name).To(Equal("imagenet"))
			Expect(dataLoad.Spec.Target).To(ContainElement(TargetPath{Path: "/train", Replicas: 2}))
			Expect(dataLoad.Spec.RunAfter).NotTo(BeNil())
			Expect(dataLoad.Spec.RunAfter.ObjectRef).To(Equal(ObjectRef{Kind: "DataLoad", Name: "prepare"}))
			Expect(*dataLoad.Spec.TTLSecondsAfterFinished).To(Equal(ttlSeconds))
			Expect(dataLoad.Status.Phase).To(Equal(common.PhaseComplete))
			Expect(dataLoad.Status.Infos).To(HaveKeyWithValue("cached", "true"))
		})
	})
})
