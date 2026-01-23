/*
Copyright 2023 The Fluid Authors.

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

package alluxio

import (
	"fmt"
	"os"
	"reflect"

	"github.com/agiledragon/gomonkey/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/dataflow"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AlluxioEngine DataBackup Tests", Label("pkg.ddc.alluxio.backup_data_test.go"), func() {
	var (
		dataset        *datav1alpha1.Dataset
		alluxioruntime *datav1alpha1.AlluxioRuntime
		engine         *AlluxioEngine
		fakeClient     client.Client
		resources      []runtime.Object
		patches        *gomonkey.Patches
	)

	BeforeEach(func() {
		dataset, alluxioruntime = mockFluidObjectsForTests(types.NamespacedName{Namespace: "fluid", Name: "hbase"})
		engine = mockAlluxioEngineForTests(dataset, alluxioruntime)
	})

	AfterEach(func() {
		if patches != nil {
			patches.Reset()
		}
	})

	JustBeforeEach(func() {
		fakeClient = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, resources...)
		engine.Client = fakeClient
	})

	Describe("Test AlluxioEngine.generateDataBackupValueFile()", func() {
		var (
			databackup *datav1alpha1.DataBackup
			masterPod  *corev1.Pod
			ctx        cruntime.ReconcileRequestContext
		)

		BeforeEach(func() {
			databackup = &datav1alpha1.DataBackup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-backup",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DataBackupSpec{
					Dataset:    "hbase",
					BackupPath: "pvc://backup-pvc/path/to/backup",
				},
			}

			masterPod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase-master-0",
					Namespace: "fluid",
				},
				Status: corev1.PodStatus{
					PodIP: "192.168.1.100",
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "alluxio-master",
							Ports: []corev1.ContainerPort{
								{
									Name:          "rpc",
									ContainerPort: 19998,
								},
							},
						},
					},
				},
			}

			ctx = cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{
					Namespace: "fluid",
					Name:      "hbase",
				},
				Log: fake.NullLogger(),
			}
		})

		Context("when object is not a DataBackup", func() {
			BeforeEach(func() {
				resources = []runtime.Object{dataset, alluxioruntime, masterPod}
			})

			It("should return an error", func() {
				nonDataBackupObject := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "not-a-databackup",
						Namespace: "fluid",
					},
				}

				valueFileName, err := engine.generateDataBackupValueFile(ctx, nonDataBackupObject)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("is not a DataBackup"))
				Expect(valueFileName).To(BeEmpty())
			})
		})

		Context("when getRuntime fails", func() {
			BeforeEach(func() {
				// Do not include alluxioruntime in resources so getRuntime fails
				resources = []runtime.Object{dataset, masterPod}
			})

			It("should return an error", func() {
				valueFileName, err := engine.generateDataBackupValueFile(ctx, databackup)
				Expect(err).To(HaveOccurred())
				Expect(valueFileName).To(BeEmpty())
			})
		})

		Context("when getMasterPod fails", func() {
			BeforeEach(func() {
				// Do not include masterPod in resources so getMasterPod fails
				resources = []runtime.Object{dataset, alluxioruntime}
			})

			It("should return an error", func() {
				valueFileName, err := engine.generateDataBackupValueFile(ctx, databackup)
				Expect(err).To(HaveOccurred())
				Expect(valueFileName).To(BeEmpty())
			})
		})

		Context("when GetDataset fails", func() {
			BeforeEach(func() {
				// Do not include dataset in resources so GetDataset fails
				resources = []runtime.Object{alluxioruntime, masterPod}
			})

			It("should return an error", func() {
				patches = gomonkey.ApplyFunc(dataflow.InjectAffinityByRunAfterOp, func(_ client.Client, _ *datav1alpha1.OperationRef, _ string, _ *corev1.Affinity) (*corev1.Affinity, error) {
					return nil, nil
				})

				valueFileName, err := engine.generateDataBackupValueFile(ctx, databackup)
				Expect(err).To(HaveOccurred())
				Expect(valueFileName).To(BeEmpty())
			})
		})

		Context("when InjectAffinityByRunAfterOp fails", func() {
			BeforeEach(func() {
				resources = []runtime.Object{dataset, alluxioruntime, masterPod}
				alluxioruntime.Spec.Replicas = 1
			})

			It("should return an error", func() {
				patches = gomonkey.ApplyFunc(dataflow.InjectAffinityByRunAfterOp, func(_ client.Client, _ *datav1alpha1.OperationRef, _ string, _ *corev1.Affinity) (*corev1.Affinity, error) {
					return nil, fmt.Errorf("InjectAffinityByRunAfterOp error")
				})

				valueFileName, err := engine.generateDataBackupValueFile(ctx, databackup)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("InjectAffinityByRunAfterOp error"))
				Expect(valueFileName).To(BeEmpty())
			})
		})

		Context("when ParseBackupRestorePath fails", func() {
			BeforeEach(func() {
				resources = []runtime.Object{dataset, alluxioruntime, masterPod}
				// Set an invalid backup path
				databackup.Spec.BackupPath = "invalid-path-without-scheme"
			})

			It("should return an error", func() {
				patches = gomonkey.ApplyFunc(dataflow.InjectAffinityByRunAfterOp, func(_ client.Client, _ *datav1alpha1.OperationRef, _ string, _ *corev1.Affinity) (*corev1.Affinity, error) {
					return nil, nil
				})

				valueFileName, err := engine.generateDataBackupValueFile(ctx, databackup)
				Expect(err).To(HaveOccurred())
				Expect(valueFileName).To(BeEmpty())
			})
		})

		Context("when DataBackup is valid with single replica runtime", func() {
			BeforeEach(func() {
				resources = []runtime.Object{dataset, alluxioruntime, masterPod}
				alluxioruntime.Spec.Replicas = 1
			})

			It("should generate value file successfully", func() {
				patches = gomonkey.ApplyFunc(dataflow.InjectAffinityByRunAfterOp, func(_ client.Client, _ *datav1alpha1.OperationRef, _ string, _ *corev1.Affinity) (*corev1.Affinity, error) {
					return nil, nil
				})

				valueFileName, err := engine.generateDataBackupValueFile(ctx, databackup)
				Expect(err).NotTo(HaveOccurred())
				Expect(valueFileName).NotTo(BeEmpty())

				// Verify the file was created
				fileInfo, statErr := os.Stat(valueFileName)
				Expect(statErr).NotTo(HaveOccurred())
				Expect(fileInfo.Size()).To(BeNumerically(">", 0))

				// Cleanup
				cleanupErr := os.Remove(valueFileName)
				Expect(cleanupErr).NotTo(HaveOccurred())
			})
		})

		Context("when runtime has multiple replicas (HA mode)", func() {
			BeforeEach(func() {
				resources = []runtime.Object{dataset, alluxioruntime, masterPod}
				alluxioruntime.Spec.Replicas = 3
			})

			Context("when MasterPodName lookup succeeds", func() {
				It("should generate value file successfully", func() {
					patches = gomonkey.NewPatches()
					patches.ApplyFunc(dataflow.InjectAffinityByRunAfterOp, func(_ client.Client, _ *datav1alpha1.OperationRef, _ string, _ *corev1.Affinity) (*corev1.Affinity, error) {
						return nil, nil
					})
					patches.ApplyMethod(reflect.TypeOf(operations.AlluxioFileUtils{}), "MasterPodName", func(_ operations.AlluxioFileUtils) (string, error) {
						return "hbase-master-0", nil
					})

					valueFileName, err := engine.generateDataBackupValueFile(ctx, databackup)
					Expect(err).NotTo(HaveOccurred())
					Expect(valueFileName).NotTo(BeEmpty())

					// Cleanup
					cleanupErr := os.Remove(valueFileName)
					Expect(cleanupErr).NotTo(HaveOccurred())
				})
			})

			Context("when MasterPodName lookup fails", func() {
				It("should return an error", func() {
					patches = gomonkey.NewPatches()
					patches.ApplyMethod(reflect.TypeOf(operations.AlluxioFileUtils{}), "MasterPodName", func(_ operations.AlluxioFileUtils) (string, error) {
						return "", fmt.Errorf("master pod not found")
					})

					valueFileName, err := engine.generateDataBackupValueFile(ctx, databackup)
					Expect(err).To(HaveOccurred())
					Expect(valueFileName).To(BeEmpty())
				})
			})
		})

		Context("when runtime.Spec.RunAs is set", func() {
			BeforeEach(func() {
				resources = []runtime.Object{dataset, alluxioruntime, masterPod}
				alluxioruntime.Spec.Replicas = 1
				uid := int64(1000)
				gid := int64(1000)
				alluxioruntime.Spec.RunAs = &datav1alpha1.User{
					UID: &uid,
					GID: &gid,
				}
			})

			It("should include RunAs info in the generated file", func() {
				patches = gomonkey.ApplyFunc(dataflow.InjectAffinityByRunAfterOp, func(_ client.Client, _ *datav1alpha1.OperationRef, _ string, _ *corev1.Affinity) (*corev1.Affinity, error) {
					return nil, nil
				})

				valueFileName, err := engine.generateDataBackupValueFile(ctx, databackup)
				Expect(err).NotTo(HaveOccurred())
				Expect(valueFileName).NotTo(BeEmpty())

				// Read file content to verify RunAs info is included
				content, readErr := os.ReadFile(valueFileName)
				Expect(readErr).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("enabled: true"))
				Expect(string(content)).To(ContainSubstring("user: 1000"))
				Expect(string(content)).To(ContainSubstring("group: 1000"))

				// Cleanup
				cleanupErr := os.Remove(valueFileName)
				Expect(cleanupErr).NotTo(HaveOccurred())
			})
		})

		Context("when databackup.Spec.RunAs overrides runtime.Spec.RunAs", func() {
			BeforeEach(func() {
				resources = []runtime.Object{dataset, alluxioruntime, masterPod}
				alluxioruntime.Spec.Replicas = 1
				runtimeUID := int64(1000)
				runtimeGID := int64(1000)
				alluxioruntime.Spec.RunAs = &datav1alpha1.User{
					UID: &runtimeUID,
					GID: &runtimeGID,
				}
				databackupUID := int64(2000)
				databackupGID := int64(2000)
				databackup.Spec.RunAs = &datav1alpha1.User{
					UID: &databackupUID,
					GID: &databackupGID,
				}
			})

			It("should use databackup RunAs values", func() {
				patches = gomonkey.ApplyFunc(dataflow.InjectAffinityByRunAfterOp, func(_ client.Client, _ *datav1alpha1.OperationRef, _ string, _ *corev1.Affinity) (*corev1.Affinity, error) {
					return nil, nil
				})

				valueFileName, err := engine.generateDataBackupValueFile(ctx, databackup)
				Expect(err).NotTo(HaveOccurred())
				Expect(valueFileName).NotTo(BeEmpty())

				// Read file content to verify databackup RunAs is used
				content, readErr := os.ReadFile(valueFileName)
				Expect(readErr).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("user: 2000"))
				Expect(string(content)).To(ContainSubstring("group: 2000"))

				// Cleanup
				cleanupErr := os.Remove(valueFileName)
				Expect(cleanupErr).NotTo(HaveOccurred())
			})
		})

		Context("when FLUID_WORKDIR environment variable is set", func() {
			BeforeEach(func() {
				resources = []runtime.Object{dataset, alluxioruntime, masterPod}
				alluxioruntime.Spec.Replicas = 1
				setErr := os.Setenv("FLUID_WORKDIR", "/custom/workdir")
				Expect(setErr).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				unsetErr := os.Unsetenv("FLUID_WORKDIR")
				Expect(unsetErr).NotTo(HaveOccurred())
			})

			It("should use custom workdir in generated file", func() {
				patches = gomonkey.ApplyFunc(dataflow.InjectAffinityByRunAfterOp, func(_ client.Client, _ *datav1alpha1.OperationRef, _ string, _ *corev1.Affinity) (*corev1.Affinity, error) {
					return nil, nil
				})

				valueFileName, err := engine.generateDataBackupValueFile(ctx, databackup)
				Expect(err).NotTo(HaveOccurred())
				Expect(valueFileName).NotTo(BeEmpty())

				// Read file content to verify custom workdir is used
				content, readErr := os.ReadFile(valueFileName)
				Expect(readErr).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("workdir: /custom/workdir"))

				// Cleanup
				cleanupErr := os.Remove(valueFileName)
				Expect(cleanupErr).NotTo(HaveOccurred())
			})
		})

		Context("when local path backup is specified", func() {
			BeforeEach(func() {
				resources = []runtime.Object{dataset, alluxioruntime, masterPod}
				alluxioruntime.Spec.Replicas = 1
				databackup.Spec.BackupPath = "local:///backup/path"
			})

			It("should generate value file with empty PVCName", func() {
				patches = gomonkey.ApplyFunc(dataflow.InjectAffinityByRunAfterOp, func(_ client.Client, _ *datav1alpha1.OperationRef, _ string, _ *corev1.Affinity) (*corev1.Affinity, error) {
					return nil, nil
				})

				valueFileName, err := engine.generateDataBackupValueFile(ctx, databackup)
				Expect(err).NotTo(HaveOccurred())
				Expect(valueFileName).NotTo(BeEmpty())

				// Read file content to verify local path handling
				content, readErr := os.ReadFile(valueFileName)
				Expect(readErr).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("runtimeType: " + common.AlluxioRuntime))

				// Cleanup
				cleanupErr := os.Remove(valueFileName)
				Expect(cleanupErr).NotTo(HaveOccurred())
			})
		})
	})
})
