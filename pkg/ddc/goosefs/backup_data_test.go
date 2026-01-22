/*
Copyright 2022 The Fluid Authors.

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

package goosefs

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/dataflow"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/goosefs/operations"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	testNamespace   = "fluid"
	testDatasetName = "test-dataset"
	testBackupName  = "test-backup"
	testMasterName  = "test-dataset-master-0"
	testNodeName    = "test-node"
	testHostIP      = "192.168.1.100"
	testBackupPath  = "pvc://backup-pvc/path"
	testMasterImage = "goosefs-master"
)

var _ = Describe("GooseFSEngine Data Backup", func() {
	var engine *GooseFSEngine

	BeforeEach(func() {
		engine = &GooseFSEngine{
			name:      testDatasetName,
			namespace: testNamespace,
			Log:       fake.NullLogger(),
		}
	})

	Describe("generateDataBackupValueFile", func() {
		Context("when object is not a DataBackup", func() {
			It("should return an error", func() {
				ctx := cruntime.ReconcileRequestContext{
					Log: fake.NullLogger(),
				}

				valueFileName, err := engine.generateDataBackupValueFile(ctx, &datav1alpha1.Dataset{})
				Expect(err).To(HaveOccurred())
				Expect(valueFileName).To(BeEmpty())
			})
		})

		Context("when runtime is not found", func() {
			It("should return an error", func() {
				testObjs := []runtime.Object{}
				fakeClient := fake.NewFakeClientWithScheme(testScheme, testObjs...)
				engine.Client = fakeClient

				ctx := cruntime.ReconcileRequestContext{
					Log: fake.NullLogger(),
					NamespacedName: types.NamespacedName{
						Namespace: testNamespace,
						Name:      testDatasetName,
					},
				}

				databackup := &datav1alpha1.DataBackup{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testBackupName,
						Namespace: testNamespace,
					},
					Spec: datav1alpha1.DataBackupSpec{
						Dataset:    testDatasetName,
						BackupPath: testBackupPath,
					},
				}

				valueFileName, err := engine.generateDataBackupValueFile(ctx, databackup)
				Expect(err).To(HaveOccurred())
				Expect(valueFileName).To(BeEmpty())
			})
		})

		Context("when HA runtime MasterPodName fails", func() {
			It("should return an error", func() {
				goosefsRuntime := &datav1alpha1.GooseFSRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testDatasetName,
						Namespace: testNamespace,
					},
					Spec: datav1alpha1.GooseFSRuntimeSpec{
						Replicas: 3,
					},
				}

				testObjs := []runtime.Object{goosefsRuntime.DeepCopy()}
				fakeClient := fake.NewFakeClientWithScheme(testScheme, testObjs...)
				engine.Client = fakeClient

				ctx := cruntime.ReconcileRequestContext{
					Log: fake.NullLogger(),
					NamespacedName: types.NamespacedName{
						Namespace: testNamespace,
						Name:      testDatasetName,
					},
				}

				databackup := &datav1alpha1.DataBackup{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testBackupName,
						Namespace: testNamespace,
					},
					Spec: datav1alpha1.DataBackupSpec{
						Dataset:    testDatasetName,
						BackupPath: testBackupPath,
					},
				}

				goosefsFileUtils := operations.GooseFSFileUtils{}
				patch := ApplyMethod(reflect.TypeOf(goosefsFileUtils), "MasterPodName", func(_ operations.GooseFSFileUtils) (string, error) {
					return "", fmt.Errorf("mock error")
				})
				defer patch.Reset()

				valueFileName, err := engine.generateDataBackupValueFile(ctx, databackup)
				Expect(err).To(HaveOccurred())
				Expect(valueFileName).To(BeEmpty())
			})
		})

		Context("when backup path is invalid", func() {
			var (
				fakeClient client.Client
				patch      *Patches
			)

			BeforeEach(func() {
				goosefsRuntime := &datav1alpha1.GooseFSRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testDatasetName,
						Namespace: testNamespace,
					},
					Spec: datav1alpha1.GooseFSRuntimeSpec{
						Replicas: 1,
					},
				}

				masterPod := createMasterPod()
				testObjs := []runtime.Object{goosefsRuntime.DeepCopy(), masterPod.DeepCopy()}
				fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)
				engine.Client = fakeClient

				patch = ApplyFunc(docker.GetWorkerImage, func(_ client.Client, _ string, _ string, _ string) (string, string) {
					return "goosefs", "v1.0"
				})
			})

			AfterEach(func() {
				patch.Reset()
			})

			It("should return an error", func() {
				ctx := cruntime.ReconcileRequestContext{
					Log:    fake.NullLogger(),
					Client: fakeClient,
					NamespacedName: types.NamespacedName{
						Namespace: testNamespace,
						Name:      testDatasetName,
					},
				}

				databackup := &datav1alpha1.DataBackup{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testBackupName,
						Namespace: testNamespace,
					},
					Spec: datav1alpha1.DataBackupSpec{
						Dataset:    testDatasetName,
						BackupPath: "invalid-path",
					},
				}

				valueFileName, err := engine.generateDataBackupValueFile(ctx, databackup)
				Expect(err).To(HaveOccurred())
				Expect(valueFileName).To(BeEmpty())
			})
		})

		Context("when affinity injection fails", func() {
			var (
				fakeClient client.Client
				patch      *Patches
				patch2     *Patches
			)

			BeforeEach(func() {
				goosefsRuntime := &datav1alpha1.GooseFSRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testDatasetName,
						Namespace: testNamespace,
					},
					Spec: datav1alpha1.GooseFSRuntimeSpec{
						Replicas: 1,
					},
				}

				masterPod := createMasterPod()
				testObjs := []runtime.Object{goosefsRuntime.DeepCopy(), masterPod.DeepCopy()}
				fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)
				engine.Client = fakeClient

				patch = ApplyFunc(docker.GetWorkerImage, func(_ client.Client, _ string, _ string, _ string) (string, string) {
					return "goosefs", "v1.0"
				})

				patch2 = ApplyFunc(dataflow.InjectAffinityByRunAfterOp, func(_ client.Client, _ *datav1alpha1.OperationRef, _ string, _ *corev1.Affinity) (*corev1.Affinity, error) {
					return nil, fmt.Errorf("mock affinity error")
				})
			})

			AfterEach(func() {
				patch.Reset()
				patch2.Reset()
			})

			It("should return an error", func() {
				ctx := cruntime.ReconcileRequestContext{
					Log:    fake.NullLogger(),
					Client: fakeClient,
					NamespacedName: types.NamespacedName{
						Namespace: testNamespace,
						Name:      testDatasetName,
					},
				}

				databackup := &datav1alpha1.DataBackup{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testBackupName,
						Namespace: testNamespace,
					},
					Spec: datav1alpha1.DataBackupSpec{
						Dataset:    testDatasetName,
						BackupPath: testBackupPath,
						RunAfter: &datav1alpha1.OperationRef{
							ObjectRef: datav1alpha1.ObjectRef{
								Kind: "DataLoad",
								Name: "nonexistent",
							},
							AffinityStrategy: datav1alpha1.AffinityStrategy{
								Policy: datav1alpha1.RequireAffinityStrategy,
							},
						},
					},
				}

				valueFileName, err := engine.generateDataBackupValueFile(ctx, databackup)
				Expect(err).To(HaveOccurred())
				Expect(valueFileName).To(BeEmpty())
			})
		})
	})
})

var _ = Describe("GooseFSEngine Data Backup Success Cases", func() {
	var (
		engine     *GooseFSEngine
		fakeClient client.Client
		patch      *Patches
	)

	BeforeEach(func() {
		engine = &GooseFSEngine{
			name:      testDatasetName,
			namespace: testNamespace,
			Log:       fake.NullLogger(),
		}
	})

	Describe("generateDataBackupValueFile with valid configuration", func() {
		Context("when runtime has RunAs and InitUsers configured", func() {
			var valueFileName string

			BeforeEach(func() {
				uid := int64(1000)
				gid := int64(1000)

				goosefsRuntime := &datav1alpha1.GooseFSRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testDatasetName,
						Namespace: testNamespace,
					},
					Spec: datav1alpha1.GooseFSRuntimeSpec{
						Replicas: 1,
						RunAs: &datav1alpha1.User{
							UID: &uid,
							GID: &gid,
						},
						InitUsers: datav1alpha1.InitUsersSpec{
							Image:           "init-users",
							ImageTag:        "v1.0",
							ImagePullPolicy: "IfNotPresent",
						},
					},
				}

				masterPod := createMasterPod()
				testObjs := []runtime.Object{goosefsRuntime.DeepCopy(), masterPod.DeepCopy()}
				fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)
				engine.Client = fakeClient

				patch = ApplyFunc(docker.GetWorkerImage, func(_ client.Client, _ string, _ string, _ string) (string, string) {
					return "goosefs", "v1.0"
				})
			})

			AfterEach(func() {
				patch.Reset()
				if valueFileName != "" {
					_ = os.Remove(valueFileName)
				}
			})

			It("should generate a valid value file", func() {
				ctx := cruntime.ReconcileRequestContext{
					Log:    fake.NullLogger(),
					Client: fakeClient,
					NamespacedName: types.NamespacedName{
						Namespace: testNamespace,
						Name:      testDatasetName,
					},
				}

				databackup := &datav1alpha1.DataBackup{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testBackupName,
						Namespace: testNamespace,
					},
					Spec: datav1alpha1.DataBackupSpec{
						Dataset:    testDatasetName,
						BackupPath: testBackupPath,
					},
				}

				var err error
				valueFileName, err = engine.generateDataBackupValueFile(ctx, databackup)
				Expect(err).NotTo(HaveOccurred())
				Expect(valueFileName).NotTo(BeEmpty())
				Expect(valueFileName).To(ContainSubstring("backuper-values.yaml"))

				content, readErr := os.ReadFile(valueFileName)
				Expect(readErr).NotTo(HaveOccurred())
				Expect(content).NotTo(BeEmpty())
			})
		})

		Context("when databackup has its own RunAs", func() {
			var valueFileName string

			BeforeEach(func() {
				runtimeUID := int64(1000)
				runtimeGID := int64(1000)

				goosefsRuntime := &datav1alpha1.GooseFSRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testDatasetName,
						Namespace: testNamespace,
					},
					Spec: datav1alpha1.GooseFSRuntimeSpec{
						Replicas: 1,
						RunAs: &datav1alpha1.User{
							UID: &runtimeUID,
							GID: &runtimeGID,
						},
					},
				}

				masterPod := createMasterPod()
				testObjs := []runtime.Object{goosefsRuntime.DeepCopy(), masterPod.DeepCopy()}
				fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)
				engine.Client = fakeClient

				patch = ApplyFunc(docker.GetWorkerImage, func(_ client.Client, _ string, _ string, _ string) (string, string) {
					return "goosefs", "v1.0"
				})
			})

			AfterEach(func() {
				patch.Reset()
				if valueFileName != "" {
					_ = os.Remove(valueFileName)
				}
			})

			It("should use databackup RunAs over runtime RunAs", func() {
				backupUID := int64(2000)
				backupGID := int64(2000)

				ctx := cruntime.ReconcileRequestContext{
					Log:    fake.NullLogger(),
					Client: fakeClient,
					NamespacedName: types.NamespacedName{
						Namespace: testNamespace,
						Name:      testDatasetName,
					},
				}

				databackup := &datav1alpha1.DataBackup{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testBackupName,
						Namespace: testNamespace,
					},
					Spec: datav1alpha1.DataBackupSpec{
						Dataset:    testDatasetName,
						BackupPath: "local:///backup/path",
						RunAs: &datav1alpha1.User{
							UID: &backupUID,
							GID: &backupGID,
						},
					},
				}

				var err error
				valueFileName, err = engine.generateDataBackupValueFile(ctx, databackup)
				Expect(err).NotTo(HaveOccurred())
				Expect(valueFileName).NotTo(BeEmpty())
			})
		})

		Context("when runtime has HA replicas", func() {
			var (
				valueFileName string
				masterPatch   *Patches
			)

			BeforeEach(func() {
				goosefsRuntime := &datav1alpha1.GooseFSRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testDatasetName,
						Namespace: testNamespace,
					},
					Spec: datav1alpha1.GooseFSRuntimeSpec{
						Replicas: 3,
					},
				}

				masterPod := createMasterPod()
				testObjs := []runtime.Object{goosefsRuntime.DeepCopy(), masterPod.DeepCopy()}
				fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)
				engine.Client = fakeClient

				goosefsFileUtils := operations.GooseFSFileUtils{}
				masterPatch = ApplyMethod(reflect.TypeOf(goosefsFileUtils), "MasterPodName", func(_ operations.GooseFSFileUtils) (string, error) {
					return testMasterName, nil
				})

				patch = ApplyFunc(docker.GetWorkerImage, func(_ client.Client, _ string, _ string, _ string) (string, string) {
					return "goosefs", "v1.0"
				})
			})

			AfterEach(func() {
				masterPatch.Reset()
				patch.Reset()
				if valueFileName != "" {
					_ = os.Remove(valueFileName)
				}
			})

			It("should successfully generate value file", func() {
				ctx := cruntime.ReconcileRequestContext{
					Log:    fake.NullLogger(),
					Client: fakeClient,
					NamespacedName: types.NamespacedName{
						Namespace: testNamespace,
						Name:      testDatasetName,
					},
				}

				databackup := &datav1alpha1.DataBackup{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testBackupName,
						Namespace: testNamespace,
					},
					Spec: datav1alpha1.DataBackupSpec{
						Dataset:    testDatasetName,
						BackupPath: testBackupPath,
					},
				}

				var err error
				valueFileName, err = engine.generateDataBackupValueFile(ctx, databackup)
				Expect(err).NotTo(HaveOccurred())
				Expect(valueFileName).NotTo(BeEmpty())
			})
		})

		Context("when worker image is not configured", func() {
			var valueFileName string

			BeforeEach(func() {
				goosefsRuntime := &datav1alpha1.GooseFSRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testDatasetName,
						Namespace: testNamespace,
					},
					Spec: datav1alpha1.GooseFSRuntimeSpec{
						Replicas: 1,
					},
				}

				masterPod := createMasterPod()
				testObjs := []runtime.Object{goosefsRuntime.DeepCopy(), masterPod.DeepCopy()}
				fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)
				engine.Client = fakeClient

				patch = ApplyFunc(docker.GetWorkerImage, func(_ client.Client, _ string, _ string, _ string) (string, string) {
					return "", ""
				})
			})

			AfterEach(func() {
				patch.Reset()
				if valueFileName != "" {
					_ = os.Remove(valueFileName)
				}
			})

			It("should use default image", func() {
				ctx := cruntime.ReconcileRequestContext{
					Log:    fake.NullLogger(),
					Client: fakeClient,
					NamespacedName: types.NamespacedName{
						Namespace: testNamespace,
						Name:      testDatasetName,
					},
				}

				databackup := &datav1alpha1.DataBackup{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testBackupName,
						Namespace: testNamespace,
					},
					Spec: datav1alpha1.DataBackupSpec{
						Dataset:    testDatasetName,
						BackupPath: testBackupPath,
					},
				}

				var err error
				valueFileName, err = engine.generateDataBackupValueFile(ctx, databackup)
				Expect(err).NotTo(HaveOccurred())
				Expect(valueFileName).NotTo(BeEmpty())
			})
		})
	})
})

func createMasterPod() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testMasterName,
			Namespace: testNamespace,
		},
		Spec: corev1.PodSpec{
			NodeName: testNodeName,
			Containers: []corev1.Container{
				{
					Name: testMasterImage,
					Ports: []corev1.ContainerPort{
						{
							Name:          "rpc",
							ContainerPort: 19998,
						},
					},
				},
			},
		},
		Status: corev1.PodStatus{
			HostIP: testHostIP,
		},
	}
}

var _ = Describe("GooseFSEngine Data Backup Value File Content", func() {
	It("should contain expected yaml structure", func() {
		uid := int64(1000)
		gid := int64(1000)

		goosefsRuntime := &datav1alpha1.GooseFSRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testDatasetName,
				Namespace: testNamespace,
			},
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Replicas: 1,
				RunAs: &datav1alpha1.User{
					UID: &uid,
					GID: &gid,
				},
			},
		}

		masterPod := createMasterPod()
		testObjs := []runtime.Object{goosefsRuntime.DeepCopy(), masterPod.DeepCopy()}
		fakeClient := fake.NewFakeClientWithScheme(testScheme, testObjs...)

		engine := &GooseFSEngine{
			name:      testDatasetName,
			namespace: testNamespace,
			Log:       fake.NullLogger(),
			Client:    fakeClient,
		}

		patch := ApplyFunc(docker.GetWorkerImage, func(_ client.Client, _ string, _ string, _ string) (string, string) {
			return "goosefs", "v1.0"
		})
		defer patch.Reset()

		ctx := cruntime.ReconcileRequestContext{
			Log:    fake.NullLogger(),
			Client: fakeClient,
			NamespacedName: types.NamespacedName{
				Namespace: testNamespace,
				Name:      testDatasetName,
			},
		}

		databackup := &datav1alpha1.DataBackup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testBackupName,
				Namespace: testNamespace,
			},
			Spec: datav1alpha1.DataBackupSpec{
				Dataset:    testDatasetName,
				BackupPath: testBackupPath,
			},
		}

		valueFileName, err := engine.generateDataBackupValueFile(ctx, databackup)
		Expect(err).NotTo(HaveOccurred())
		Expect(valueFileName).NotTo(BeEmpty())
		defer func() { _ = os.Remove(valueFileName) }()

		content, readErr := os.ReadFile(valueFileName)
		Expect(readErr).NotTo(HaveOccurred())

		contentStr := string(content)
		Expect(strings.Contains(contentStr, "name:")).To(BeTrue())
		Expect(strings.Contains(contentStr, "namespace:")).To(BeTrue())
	})
})
