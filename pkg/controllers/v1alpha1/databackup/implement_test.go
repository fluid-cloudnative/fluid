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

package databackup

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdatabackup "github.com/fluid-cloudnative/fluid/pkg/databackup"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("dataBackupOperation", func() {
	var (
		testScheme *runtime.Scheme
		dataBackup *datav1alpha1.DataBackup
		op         *dataBackupOperation
	)

	BeforeEach(func() {
		testScheme = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(testScheme)
		_ = corev1.AddToScheme(testScheme)

		dataBackup = &datav1alpha1.DataBackup{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-backup",
				Namespace: "default",
			},
			Spec: datav1alpha1.DataBackupSpec{
				Dataset:    "test-dataset",
				BackupPath: "pvc://test-pvc/path",
			},
		}

		fakeClient := fake.NewFakeClientWithScheme(testScheme, dataBackup)
		log := ctrl.Log.WithName("test")
		recorder := record.NewFakeRecorder(10)

		op = &dataBackupOperation{
			Client:     fakeClient,
			Log:        log,
			Recorder:   recorder,
			dataBackup: dataBackup,
		}
	})

	Describe("GetOperationObject", func() {
		It("should return the dataBackup object", func() {
			obj := op.GetOperationObject()
			Expect(obj).To(Equal(dataBackup))
		})
	})

	Describe("GetChartsDirectory", func() {
		It("should contain the DatabackupChart constant", func() {
			dir := op.GetChartsDirectory()
			Expect(dir).To(ContainSubstring(cdatabackup.DatabackupChart))
		})
	})

	Describe("HasPrecedingOperation", func() {
		It("should return false when RunAfter is nil", func() {
			dataBackup.Spec.RunAfter = nil
			Expect(op.HasPrecedingOperation()).To(BeFalse())
		})

		It("should return true when RunAfter is set", func() {
			dataBackup.Spec.RunAfter = &datav1alpha1.OperationRef{}
			Expect(op.HasPrecedingOperation()).To(BeTrue())
		})
	})

	Describe("GetOperationType", func() {
		It("should return DataBackupType", func() {
			Expect(op.GetOperationType()).To(Equal(dataoperation.DataBackupType))
		})
	})

	Describe("GetPossibleTargetDatasetNamespacedNames", func() {
		It("should return a single NamespacedName matching the dataBackup", func() {
			names := op.GetPossibleTargetDatasetNamespacedNames()
			Expect(names).To(HaveLen(1))
			Expect(names[0]).To(Equal(types.NamespacedName{
				Namespace: "default",
				Name:      "test-backup",
			}))
		})
	})

	Describe("GetReleaseNameSpacedName", func() {
		It("should return NamespacedName with the release name derived from the backup name", func() {
			nsn := op.GetReleaseNameSpacedName()
			Expect(nsn.Namespace).To(Equal("default"))
			Expect(nsn.Name).NotTo(BeEmpty())
		})
	})

	Describe("GetStatusHandler", func() {
		It("should return an OnceHandler", func() {
			handler := op.GetStatusHandler()
			Expect(handler).NotTo(BeNil())
			_, ok := handler.(*OnceHandler)
			Expect(ok).To(BeTrue())
		})
	})

	Describe("GetTTL", func() {
		It("should return nil when TTLSecondsAfterFinished is not set", func() {
			dataBackup.Spec.TTLSecondsAfterFinished = nil
			ttl, err := op.GetTTL()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(BeNil())
		})

		It("should return the TTL value when set", func() {
			ttlVal := int32(300)
			dataBackup.Spec.TTLSecondsAfterFinished = &ttlVal
			ttl, err := op.GetTTL()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).NotTo(BeNil())
			Expect(*ttl).To(Equal(int32(300)))
		})
	})

	Describe("GetParallelTaskNumber", func() {
		It("should return 1", func() {
			Expect(op.GetParallelTaskNumber()).To(Equal(int32(1)))
		})
	})

	Describe("SetTargetDatasetStatusInProgress", func() {
		It("should not panic and be a no-op", func() {
			dataset := &datav1alpha1.Dataset{}
			Expect(func() { op.SetTargetDatasetStatusInProgress(dataset) }).NotTo(Panic())
		})
	})

	Describe("RemoveTargetDatasetStatusInProgress", func() {
		It("should not panic and be a no-op", func() {
			dataset := &datav1alpha1.Dataset{}
			Expect(func() { op.RemoveTargetDatasetStatusInProgress(dataset) }).NotTo(Panic())
		})
	})

	Describe("Validate", func() {
		It("should return nil conditions and no error for a valid pvc:// path", func() {
			dataBackup.Spec.BackupPath = "pvc://my-pvc/path"
			ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
			conditions, err := op.Validate(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(conditions).To(BeNil())
		})

		It("should return nil conditions and no error for a valid local:// path", func() {
			dataBackup.Spec.BackupPath = "local:///tmp/backup"
			ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
			conditions, err := op.Validate(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(conditions).To(BeNil())
		})

		It("should return error and Failed condition for an unsupported path format", func() {
			dataBackup.Spec.BackupPath = "s3://my-bucket/path"
			ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
			conditions, err := op.Validate(ctx)
			Expect(err).To(HaveOccurred())
			Expect(conditions).To(HaveLen(1))
			Expect(conditions[0].Type).To(Equal(common.Failed))
			Expect(conditions[0].Reason).To(Equal("PathNotSupported"))
		})
	})

	Describe("UpdateStatusInfoForCompleted", func() {
		It("should set BackupLocationPath and BackupLocationNodeName=NA for pvc path", func() {
			dataBackup.Spec.BackupPath = "pvc://my-pvc/path"
			infos := map[string]string{}
			err := op.UpdateStatusInfoForCompleted(infos)
			Expect(err).NotTo(HaveOccurred())
			Expect(infos[cdatabackup.BackupLocationPath]).To(Equal("pvc://my-pvc/path"))
			Expect(infos[cdatabackup.BackupLocationNodeName]).To(Equal("NA"))
		})

		It("should set BackupLocationNodeName from pod for local:// path", func() {
			podName := dataBackup.GetName() + "-pod"
			backupPod := &corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Name:      podName,
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					NodeName: "node-1",
				},
			}
			testScheme2 := runtime.NewScheme()
			_ = datav1alpha1.AddToScheme(testScheme2)
			_ = corev1.AddToScheme(testScheme2)
			dataBackup2 := dataBackup.DeepCopy()
			dataBackup2.Spec.BackupPath = "local:///tmp/backup"
			fakeClient2 := fake.NewFakeClientWithScheme(testScheme2, dataBackup2, backupPod)
			op2 := &dataBackupOperation{
				Client:     fakeClient2,
				Log:        ctrl.Log.WithName("test"),
				Recorder:   record.NewFakeRecorder(10),
				dataBackup: dataBackup2,
			}

			infos := map[string]string{}
			err := op2.UpdateStatusInfoForCompleted(infos)
			Expect(err).NotTo(HaveOccurred())
			Expect(infos[cdatabackup.BackupLocationPath]).To(Equal("local:///tmp/backup"))
			Expect(infos[cdatabackup.BackupLocationNodeName]).To(Equal("node-1"))
		})
	})

	Describe("UpdateOperationApiStatus", func() {
		It("should update the dataBackup status without error", func() {
			testScheme4 := runtime.NewScheme()
			_ = datav1alpha1.AddToScheme(testScheme4)
			_ = corev1.AddToScheme(testScheme4)
			fakeClient4 := fake.NewFakeClientWithScheme(testScheme4, dataBackup)
			op4 := &dataBackupOperation{
				Client:     fakeClient4,
				Log:        ctrl.Log.WithName("test"),
				Recorder:   record.NewFakeRecorder(10),
				dataBackup: dataBackup,
			}
			opStatus := &datav1alpha1.OperationStatus{
				Phase: "Complete",
			}
			err := op4.UpdateOperationApiStatus(opStatus)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("GetTargetDataset", func() {
		It("should return error when dataset does not exist", func() {
			_, err := op.GetTargetDataset()
			Expect(err).To(HaveOccurred())
		})

		It("should return dataset when it exists", func() {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: v1.ObjectMeta{
					Name:      dataBackup.Spec.Dataset,
					Namespace: "default",
				},
			}
			testScheme3 := runtime.NewScheme()
			_ = datav1alpha1.AddToScheme(testScheme3)
			_ = corev1.AddToScheme(testScheme3)
			fakeClient3 := fake.NewFakeClientWithScheme(testScheme3, dataBackup, dataset)
			op3 := &dataBackupOperation{
				Client:     fakeClient3,
				Log:        ctrl.Log.WithName("test"),
				Recorder:   record.NewFakeRecorder(10),
				dataBackup: dataBackup,
			}

			got, err := op3.GetTargetDataset()
			Expect(err).NotTo(HaveOccurred())
			Expect(got).NotTo(BeNil())
			Expect(got.Name).To(Equal(dataBackup.Spec.Dataset))
		})
	})
})
