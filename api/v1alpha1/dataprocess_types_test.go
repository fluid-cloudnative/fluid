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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("DataProcess types", func() {
	Describe("scheme registration", func() {
		It("registers DataProcess and DataProcessList with the package group version", func() {
			dataProcessGVK, err := apiGVKFor(&DataProcess{})
			Expect(err).NotTo(HaveOccurred())
			Expect(dataProcessGVK).To(Equal(GroupVersion.WithKind("DataProcess")))

			dataProcessListGVK, err := apiGVKFor(&DataProcessList{})
			Expect(err).NotTo(HaveOccurred())
			Expect(dataProcessListGVK).To(Equal(GroupVersion.WithKind("DataProcessList")))
		})
	})

	Describe("DeepCopyObject", func() {
		It("returns a distinct runtime object for DataProcess and DataProcessList", func() {
			dataProcess := &DataProcess{
				TypeMeta:   metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: "DataProcess"},
				ObjectMeta: metav1.ObjectMeta{Name: "example", Namespace: "fluid"},
				Spec: DataProcessSpec{
					Dataset: TargetDatasetWithMountPath{
						TargetDataset: TargetDataset{Name: "dataset", Namespace: "fluid"},
						MountPath:     "/data",
					},
					Processor: Processor{
						Script: &ScriptProcessor{
							Source:        "python /scripts/train.py",
							RestartPolicy: corev1.RestartPolicyNever,
							Command:       []string{"/bin/sh", "-c"},
						},
					},
				},
				Status: OperationStatus{Phase: common.PhaseComplete, Duration: "90s"},
			}

			copiedObject := dataProcess.DeepCopyObject()
			copiedDataProcess, ok := copiedObject.(*DataProcess)
			Expect(ok).To(BeTrue())
			Expect(copiedDataProcess).NotTo(BeIdenticalTo(dataProcess))
			Expect(copiedDataProcess.Spec.Dataset).To(Equal(dataProcess.Spec.Dataset))
			// Verify deep copy of nested pointers.
			Expect(copiedDataProcess.Spec.Processor.Script).NotTo(BeIdenticalTo(dataProcess.Spec.Processor.Script))
			Expect(*copiedDataProcess.Spec.Processor.Script).To(Equal(*dataProcess.Spec.Processor.Script))
			Expect(copiedDataProcess.Status).To(Equal(dataProcess.Status))

			dataProcessList := &DataProcessList{Items: []DataProcess{*dataProcess}}
			copiedListObject := dataProcessList.DeepCopyObject()
			copiedList, ok := copiedListObject.(*DataProcessList)
			Expect(ok).To(BeTrue())
			Expect(copiedList).NotTo(BeIdenticalTo(dataProcessList))
			Expect(copiedList.Items).To(HaveLen(1))
			Expect(copiedList.Items[0].Spec.Dataset.MountPath).To(Equal("/data"))
		})
	})

	Describe("representative spec and status construction", func() {
		It("captures dataset mounting, script processor configuration, and workflow dependency", func() {
			ttlSeconds := int32(600)
			dataProcess := DataProcess{
				Spec: DataProcessSpec{
					Dataset: TargetDatasetWithMountPath{
						TargetDataset: TargetDataset{Name: "imagenet", Namespace: "fluid"},
						MountPath:     "/workspace/dataset",
						SubPath:       "training",
					},
					Processor: Processor{
						Script: &ScriptProcessor{
							Source:        "python /workspace/train.py",
							RestartPolicy: corev1.RestartPolicyOnFailure,
							Command:       []string{"/bin/bash", "-lc"},
							Env:           []corev1.EnvVar{{Name: "EPOCHS", Value: "5"}},
						},
					},
					RunAfter: &OperationRef{ObjectRef: ObjectRef{
						Kind: "DataLoad",
						Name: "load-training-data",
					}},
					TTLSecondsAfterFinished: &ttlSeconds,
				},
				Status: OperationStatus{
					Phase:    common.PhaseComplete,
					Duration: "8m",
					Infos:    map[string]string{"processor": "script"},
				},
			}

			Expect(dataProcess.Spec.Dataset.TargetDataset.Name).To(Equal("imagenet"))
			Expect(dataProcess.Spec.Dataset.MountPath).To(Equal("/workspace/dataset"))
			Expect(dataProcess.Spec.Processor.Script).NotTo(BeNil())
			Expect(dataProcess.Spec.Processor.Script.Env).To(ContainElement(corev1.EnvVar{Name: "EPOCHS", Value: "5"}))
			Expect(dataProcess.Spec.RunAfter).NotTo(BeNil())
			Expect(dataProcess.Spec.RunAfter.ObjectRef).To(Equal(ObjectRef{Kind: "DataLoad", Name: "load-training-data"}))
			Expect(dataProcess.Status.Infos).To(HaveKeyWithValue("processor", "script"))
		})
	})
})
