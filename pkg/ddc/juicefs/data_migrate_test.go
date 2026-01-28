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

package juicefs

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cdatamigrate "github.com/fluid-cloudnative/fluid/pkg/datamigrate"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// TestJuiceFSEngine_generateDataMigrateValueFile tests the generateDataMigrateValueFile method of JuiceFSEngine.
// This function verifies that the value file for data migration is generated correctly based on different DataMigrate specifications.
// Test cases cover the following scenarios:
// - Data migration without a target path.
// - Data migration with a specified target path.
// - Data migration with a specified target path and parallel migration options.
var _ = Describe("JuiceFSEngine_generateDataMigrateValueFile", func() {
	var (
		configMap                     *corev1.ConfigMap
		datasetInputs                 []v1alpha1.Dataset
		context                       cruntime.ReconcileRequestContext
		dataMigrateNoTarget           v1alpha1.DataMigrate
		dataMigrateWithTarget         v1alpha1.DataMigrate
		parallelDataMigrateWithTarget v1alpha1.DataMigrate
	)

	BeforeEach(func() {
		configMap = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-dataset-juicefs-values",
				Namespace: "fluid",
			},
			Data: map[string]string{
				"data": ``,
			},
		}

		datasetInputs = []v1alpha1.Dataset{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataset",
					Namespace: "fluid",
				},
				Spec: v1alpha1.DatasetSpec{
					Mounts: []v1alpha1.Mount{{
						MountPoint: "juicefs:///",
					}},
				},
			},
		}

		testObjs := []runtime.Object{}
		testObjs = append(testObjs, configMap)
		for _, datasetInput := range datasetInputs {
			testObjs = append(testObjs, datasetInput.DeepCopy())
		}

		context = cruntime.ReconcileRequestContext{
			Client: fake.NewFakeClientWithScheme(testScheme, testObjs...),
		}

		dataMigrateNoTarget = v1alpha1.DataMigrate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-datamigrate",
				Namespace: "fluid",
			},
			Spec: v1alpha1.DataMigrateSpec{
				From: v1alpha1.DataToMigrate{
					DataSet: &v1alpha1.DatasetToMigrate{
						Name:      "test-dataset",
						Namespace: "fluid",
					},
				},
				To: v1alpha1.DataToMigrate{
					ExternalStorage: &v1alpha1.ExternalStorage{
						URI: "minio://test/test",
					},
				},
			},
		}

		dataMigrateWithTarget = v1alpha1.DataMigrate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-datamigrate",
				Namespace: "fluid",
			},
			Spec: v1alpha1.DataMigrateSpec{
				From: v1alpha1.DataToMigrate{
					DataSet: &v1alpha1.DatasetToMigrate{
						Name:      "test-dataset",
						Namespace: "fluid",
						Path:      "/test/",
					},
				},
				To: v1alpha1.DataToMigrate{
					ExternalStorage: &v1alpha1.ExternalStorage{
						URI: "minio://test/test",
					},
				},
				Options: map[string]string{
					"exclude": "4.png",
				},
			},
		}

		parallelDataMigrateWithTarget = v1alpha1.DataMigrate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-para-datamigrate",
				Namespace: "fluid",
			},
			Spec: v1alpha1.DataMigrateSpec{
				From: v1alpha1.DataToMigrate{
					DataSet: &v1alpha1.DatasetToMigrate{
						Name:      "test-dataset",
						Namespace: "fluid",
						Path:      "/test/",
					},
				},
				To: v1alpha1.DataToMigrate{
					ExternalStorage: &v1alpha1.ExternalStorage{
						URI: "minio://test/test",
					},
				},
				Options: map[string]string{
					"exclude": "4.png",
				},
				Parallelism: 2,
				ParallelOptions: map[string]string{
					cdatamigrate.SSHPort: "120",
				},
			},
		}
	})

	It("should generate value file for data migration without target path", func() {
		runtimeInfo, err := base.BuildRuntimeInfo("juicefs", "fluid", "juicefs")
		Expect(err).NotTo(HaveOccurred())

		engine := JuiceFSEngine{
			name:        "juicefs",
			namespace:   "fluid",
			Client:      context.Client,
			Log:         fake.NullLogger(),
			runtimeInfo: runtimeInfo,
		}

		fileName, err := engine.generateDataMigrateValueFile(context, &dataMigrateNoTarget)
		Expect(err).NotTo(HaveOccurred())
		Expect(fileName).To(ContainSubstring(filepath.Join(os.TempDir(), "fluid-test-datamigrate-migrate-values.yaml")))
	})

	It("should generate value file for data migration with target path", func() {
		runtimeInfo, err := base.BuildRuntimeInfo("juicefs", "fluid", "juicefs")
		Expect(err).NotTo(HaveOccurred())

		engine := JuiceFSEngine{
			name:        "juicefs",
			namespace:   "fluid",
			Client:      context.Client,
			Log:         fake.NullLogger(),
			runtimeInfo: runtimeInfo,
		}

		fileName, err := engine.generateDataMigrateValueFile(context, &dataMigrateWithTarget)
		Expect(err).NotTo(HaveOccurred())
		Expect(fileName).To(ContainSubstring(filepath.Join(os.TempDir(), "fluid-test-datamigrate-migrate-values.yaml")))
	})

	It("should generate value file for parallel data migration with target path", func() {
		runtimeInfo, err := base.BuildRuntimeInfo("juicefs", "fluid", "juicefs")
		Expect(err).NotTo(HaveOccurred())

		engine := JuiceFSEngine{
			name:        "juicefs",
			namespace:   "fluid",
			Client:      context.Client,
			Log:         fake.NullLogger(),
			runtimeInfo: runtimeInfo,
		}

		fileName, err := engine.generateDataMigrateValueFile(context, &parallelDataMigrateWithTarget)
		Expect(err).NotTo(HaveOccurred())
		Expect(fileName).To(ContainSubstring(filepath.Join(os.TempDir(), "fluid-test-para-datamigrate-migrate-values.yaml")))
	})
})

var _ = Describe("JuiceFSEngine_genDataUrl_PVC", func() {
	It("should handle external PVC without subpath", func() {
		info := &cdatamigrate.DataMigrateInfo{}
		data := v1alpha1.DataToMigrate{
			ExternalStorage: &v1alpha1.ExternalStorage{
				URI: "pvc://my-pvc",
			},
		}

		j := &JuiceFSEngine{
			Client: fake.NewFakeClient(),
			Log:    fake.NullLogger(),
		}

		gotDataUrl, err := j.genDataUrl(data, nil, info)
		Expect(err).NotTo(HaveOccurred())
		Expect(gotDataUrl).To(Equal(NativeVolumeMigratePath))

		wantInfo := &cdatamigrate.DataMigrateInfo{
			NativeVolumes: []corev1.Volume{
				{
					Name: "native-vol",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "my-pvc",
						},
					},
				},
			},
			NativeVolumeMounts: []corev1.VolumeMount{
				{
					Name:      "native-vol",
					MountPath: NativeVolumeMigratePath,
					SubPath:   "",
				},
			},
		}
		Expect(info).To(Equal(wantInfo))
	})

	It("should handle external PVC with subpath", func() {
		info := &cdatamigrate.DataMigrateInfo{}
		data := v1alpha1.DataToMigrate{
			ExternalStorage: &v1alpha1.ExternalStorage{
				URI: "pvc://my-pvc/path/to/dir",
			},
		}

		j := &JuiceFSEngine{
			Client: fake.NewFakeClient(),
			Log:    fake.NullLogger(),
		}

		gotDataUrl, err := j.genDataUrl(data, nil, info)
		Expect(err).NotTo(HaveOccurred())
		Expect(gotDataUrl).To(Equal(NativeVolumeMigratePath))

		wantInfo := &cdatamigrate.DataMigrateInfo{
			NativeVolumes: []corev1.Volume{
				{
					Name: "native-vol",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "my-pvc",
						},
					},
				},
			},
			NativeVolumeMounts: []corev1.VolumeMount{
				{
					Name:      "native-vol",
					MountPath: NativeVolumeMigratePath,
					SubPath:   "path/to/dir",
				},
			},
		}
		Expect(info).To(Equal(wantInfo))
	})

	It("should handle external PVC with root path", func() {
		info := &cdatamigrate.DataMigrateInfo{}
		data := v1alpha1.DataToMigrate{
			ExternalStorage: &v1alpha1.ExternalStorage{
				URI: "pvc://my-pvc/",
			},
		}

		j := &JuiceFSEngine{
			Client: fake.NewFakeClient(),
			Log:    fake.NullLogger(),
		}

		gotDataUrl, err := j.genDataUrl(data, nil, info)
		Expect(err).NotTo(HaveOccurred())
		Expect(gotDataUrl).To(Equal(NativeVolumeMigratePath))

		wantInfo := &cdatamigrate.DataMigrateInfo{
			NativeVolumes: []corev1.Volume{
				{
					Name: "native-vol",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "my-pvc",
						},
					},
				},
			},
			NativeVolumeMounts: []corev1.VolumeMount{
				{
					Name:      "native-vol",
					MountPath: NativeVolumeMigratePath,
					SubPath:   "",
				},
			},
		}
		Expect(info).To(Equal(wantInfo))
	})
})

// TestJuiceFSEngine_genDataUrl Test JuiceFSEngine's genDataUrl function
// This function is used to generate URLs for data migration and supports the following two data sources:
// 1. ExternalStorage: Generate URLs with authentication information via URIs and encryption options.
// 2. DataSet: Generate a URL in JuiceFS format from the dataset information and mount points.
//
// Test Cases:
// - test-external: Tests externally stored URL generation to verify that the encryption option is correctly replaced.
// - test-external-subpath: Tests the URL generation of the external storage subpath.
// - test-external-subpath-file: tests the URL generation of the external storage subpath file.
// - test-dataset: Test the URL generation of the dataset store to verify that the subpath is handled correctly.
// - test-dataset-no-path: The URL generated by the test dataset store (no subpath).
// - test-dataset-subpath-file: The URL of the test dataset storage subpath file is generated.
// - test-dataset-subpath-file2: The URL of the test dataset storage subpath file (with mount point) is generated.
// - test-dataset-subpath-file3: The URL of the test dataset storage subpath file (with mount points, no subpaths).
var _ = Describe("JuiceFSEngine_genDataUrl", func() {
	var (
		juicefsSecret    *corev1.Secret
		juicefsConfigMap *corev1.ConfigMap
	)

	BeforeEach(func() {
		juicefsSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: "juicefs-secret",
			},
			Data: map[string][]byte{
				"access-key": []byte(base64.StdEncoding.EncodeToString([]byte("test"))),
				"secret-key": []byte(base64.StdEncoding.EncodeToString([]byte("test"))),
				"metaurl":    []byte(base64.StdEncoding.EncodeToString([]byte("test"))),
			},
		}
		juicefsConfigMap = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-juicefs-values",
				Namespace: "default",
			},
			Data: map[string]string{
				"data": valuesConfigMapData,
			},
		}
	})

	It("should generate URL for external storage with full encryption options", func() {
		testObjs := []runtime.Object{}
		testObjs = append(testObjs, (*juicefsSecret).DeepCopy(), juicefsConfigMap)
		client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

		info := &cdatamigrate.DataMigrateInfo{
			EncryptOptions: []v1alpha1.EncryptOption{
				{
					Name: "access-key",
					ValueFrom: v1alpha1.EncryptOptionSource{SecretKeyRef: v1alpha1.SecretKeySelector{
						Name: "juicefs-secret",
						Key:  "access-key",
					}},
				},
				{
					Name: "secret-key",
					ValueFrom: v1alpha1.EncryptOptionSource{SecretKeyRef: v1alpha1.SecretKeySelector{
						Name: "juicefs-secret",
						Key:  "secret-key",
					}},
				},
				{
					Name: "token",
					ValueFrom: v1alpha1.EncryptOptionSource{SecretKeyRef: v1alpha1.SecretKeySelector{
						Name: "juicefs-secret",
						Key:  "token",
					}},
				},
			},
			Options: map[string]string{},
		}
		data := v1alpha1.DataToMigrate{ExternalStorage: &v1alpha1.ExternalStorage{
			URI: "http://minio/",
			EncryptOptions: []v1alpha1.EncryptOption{
				{
					Name: "access-key",
					ValueFrom: v1alpha1.EncryptOptionSource{SecretKeyRef: v1alpha1.SecretKeySelector{
						Name: "juicefs-secret",
						Key:  "access-key",
					}},
				},
				{
					Name: "secret-key",
					ValueFrom: v1alpha1.EncryptOptionSource{SecretKeyRef: v1alpha1.SecretKeySelector{
						Name: "juicefs-secret",
						Key:  "secret-key",
					}},
				},
				{
					Name: "token",
					ValueFrom: v1alpha1.EncryptOptionSource{SecretKeyRef: v1alpha1.SecretKeySelector{
						Name: "juicefs-secret",
						Key:  "token",
					}},
				},
			},
		}}

		j := &JuiceFSEngine{
			Client: client,
			Log:    fake.NullLogger(),
		}
		gotDataUrl, err := j.genDataUrl(data, nil, info)
		Expect(err).NotTo(HaveOccurred())
		Expect(gotDataUrl).To(Equal("http://${EXTERNAL_ACCESS_KEY}:${EXTERNAL_SECRET_KEY}:${EXTERNAL_TOKEN}@minio/"))
	})

	It("should generate URL for external storage with subpath", func() {
		testObjs := []runtime.Object{}
		testObjs = append(testObjs, (*juicefsSecret).DeepCopy(), juicefsConfigMap)
		client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

		info := &cdatamigrate.DataMigrateInfo{
			EncryptOptions: []v1alpha1.EncryptOption{{
				Name: "access-key",
				ValueFrom: v1alpha1.EncryptOptionSource{SecretKeyRef: v1alpha1.SecretKeySelector{
					Name: "juicefs-secret",
					Key:  "access-key",
				}},
			}},
			Options: map[string]string{},
		}
		data := v1alpha1.DataToMigrate{ExternalStorage: &v1alpha1.ExternalStorage{
			URI: "http://minio/test/",
			EncryptOptions: []v1alpha1.EncryptOption{{
				Name: "access-key",
				ValueFrom: v1alpha1.EncryptOptionSource{SecretKeyRef: v1alpha1.SecretKeySelector{
					Name: "juicefs-secret",
					Key:  "access-key",
				}},
			}},
		}}

		j := &JuiceFSEngine{
			Client: client,
			Log:    fake.NullLogger(),
		}
		gotDataUrl, err := j.genDataUrl(data, nil, info)
		Expect(err).NotTo(HaveOccurred())
		Expect(gotDataUrl).To(Equal("http://${EXTERNAL_ACCESS_KEY}:@minio/test/"))
	})

	It("should generate URL for external storage with subpath file", func() {
		testObjs := []runtime.Object{}
		testObjs = append(testObjs, (*juicefsSecret).DeepCopy(), juicefsConfigMap)
		client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

		info := &cdatamigrate.DataMigrateInfo{
			EncryptOptions: []v1alpha1.EncryptOption{{
				Name: "access-key",
				ValueFrom: v1alpha1.EncryptOptionSource{SecretKeyRef: v1alpha1.SecretKeySelector{
					Name: "juicefs-secret",
					Key:  "access-key",
				}},
			}},
			Options: map[string]string{},
		}
		data := v1alpha1.DataToMigrate{ExternalStorage: &v1alpha1.ExternalStorage{
			URI: "http://minio/test",
			EncryptOptions: []v1alpha1.EncryptOption{{
				Name: "access-key",
				ValueFrom: v1alpha1.EncryptOptionSource{SecretKeyRef: v1alpha1.SecretKeySelector{
					Name: "juicefs-secret",
					Key:  "access-key",
				}},
			}},
		}}

		j := &JuiceFSEngine{
			Client: client,
			Log:    fake.NullLogger(),
		}
		gotDataUrl, err := j.genDataUrl(data, nil, info)
		Expect(err).NotTo(HaveOccurred())
		Expect(gotDataUrl).To(Equal("http://${EXTERNAL_ACCESS_KEY}:@minio/test"))
	})

	It("should generate URL for dataset with subpath", func() {
		testObjs := []runtime.Object{}
		testObjs = append(testObjs, (*juicefsSecret).DeepCopy(), juicefsConfigMap)
		client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

		info := &cdatamigrate.DataMigrateInfo{
			EncryptOptions: []v1alpha1.EncryptOption{},
			Options:        map[string]string{},
		}
		data := v1alpha1.DataToMigrate{
			DataSet: &v1alpha1.DatasetToMigrate{
				Name:      "test",
				Namespace: "default",
				Path:      "/subpath/",
			},
		}
		targetDataset := &v1alpha1.Dataset{
			Spec: v1alpha1.DatasetSpec{
				Mounts: []v1alpha1.Mount{{MountPoint: "juicefs:///"}},
			},
		}

		j := &JuiceFSEngine{
			Client: client,
			Log:    fake.NullLogger(),
		}
		gotDataUrl, err := j.genDataUrl(data, targetDataset, info)
		Expect(err).NotTo(HaveOccurred())
		Expect(gotDataUrl).To(Equal("jfs://FLUID_METAURL/subpath/"))
	})

	It("should generate URL for dataset without path", func() {
		testObjs := []runtime.Object{}
		testObjs = append(testObjs, (*juicefsSecret).DeepCopy(), juicefsConfigMap)
		client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

		info := &cdatamigrate.DataMigrateInfo{
			EncryptOptions: []v1alpha1.EncryptOption{},
			Options:        map[string]string{},
		}
		data := v1alpha1.DataToMigrate{
			DataSet: &v1alpha1.DatasetToMigrate{
				Name:      "test",
				Namespace: "default",
			},
		}
		targetDataset := &v1alpha1.Dataset{
			Spec: v1alpha1.DatasetSpec{
				Mounts: []v1alpha1.Mount{{MountPoint: "juicefs:///"}},
			},
		}

		j := &JuiceFSEngine{
			Client: client,
			Log:    fake.NullLogger(),
		}
		gotDataUrl, err := j.genDataUrl(data, targetDataset, info)
		Expect(err).NotTo(HaveOccurred())
		Expect(gotDataUrl).To(Equal("jfs://FLUID_METAURL/"))
	})

	It("should generate URL for dataset with subpath file", func() {
		testObjs := []runtime.Object{}
		testObjs = append(testObjs, (*juicefsSecret).DeepCopy(), juicefsConfigMap)
		client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

		info := &cdatamigrate.DataMigrateInfo{
			EncryptOptions: []v1alpha1.EncryptOption{},
			Options:        map[string]string{},
		}
		data := v1alpha1.DataToMigrate{
			DataSet: &v1alpha1.DatasetToMigrate{
				Name:      "test",
				Namespace: "default",
				Path:      "/subpath",
			},
		}
		targetDataset := &v1alpha1.Dataset{
			Spec: v1alpha1.DatasetSpec{
				Mounts: []v1alpha1.Mount{{MountPoint: "juicefs:///"}},
			},
		}

		j := &JuiceFSEngine{
			Client: client,
			Log:    fake.NullLogger(),
		}
		gotDataUrl, err := j.genDataUrl(data, targetDataset, info)
		Expect(err).NotTo(HaveOccurred())
		Expect(gotDataUrl).To(Equal("jfs://FLUID_METAURL/subpath"))
	})

	It("should generate URL for dataset with subpath file and mount point", func() {
		testObjs := []runtime.Object{}
		testObjs = append(testObjs, (*juicefsSecret).DeepCopy(), juicefsConfigMap)
		client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

		info := &cdatamigrate.DataMigrateInfo{
			EncryptOptions: []v1alpha1.EncryptOption{},
			Options:        map[string]string{},
		}
		data := v1alpha1.DataToMigrate{
			DataSet: &v1alpha1.DatasetToMigrate{
				Name:      "test",
				Namespace: "default",
				Path:      "/subpath",
			},
		}
		targetDataset := &v1alpha1.Dataset{
			Spec: v1alpha1.DatasetSpec{
				Mounts: []v1alpha1.Mount{{MountPoint: "juicefs:///demo"}},
			},
		}

		j := &JuiceFSEngine{
			Client: client,
			Log:    fake.NullLogger(),
		}
		gotDataUrl, err := j.genDataUrl(data, targetDataset, info)
		Expect(err).NotTo(HaveOccurred())
		Expect(gotDataUrl).To(Equal("jfs://FLUID_METAURL/demo/subpath"))
	})

	It("should generate URL for dataset with mount point and no subpath", func() {
		testObjs := []runtime.Object{}
		testObjs = append(testObjs, (*juicefsSecret).DeepCopy(), juicefsConfigMap)
		client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

		info := &cdatamigrate.DataMigrateInfo{
			EncryptOptions: []v1alpha1.EncryptOption{},
			Options:        map[string]string{},
		}
		data := v1alpha1.DataToMigrate{
			DataSet: &v1alpha1.DatasetToMigrate{
				Name:      "test",
				Namespace: "default",
			},
		}
		targetDataset := &v1alpha1.Dataset{
			Spec: v1alpha1.DatasetSpec{
				Mounts: []v1alpha1.Mount{{MountPoint: "juicefs:///demo"}},
			},
		}

		j := &JuiceFSEngine{
			Client: client,
			Log:    fake.NullLogger(),
		}
		gotDataUrl, err := j.genDataUrl(data, targetDataset, info)
		Expect(err).NotTo(HaveOccurred())
		Expect(gotDataUrl).To(Equal("jfs://FLUID_METAURL/demo"))
	})
})

// TestJuiceFSEngine_setParallelMigrateOptions tests the setParallelMigrateOptions method of JuiceFSEngine.
// This method is responsible for setting parallel migration options, including SSH port and SSH ready timeout.
// The test cases cover:
// 1. Normal scenario where parallel migration options are set correctly.
// 2. Default values are used when no parallel options are provided.
// 3. Error scenario where invalid parallel options are provided.
var _ = Describe("JuiceFSEngine_setParallelMigrateOptions", func() {
	It("should set parallel migrate options correctly", func() {
		client := fake.NewFakeClientWithScheme(testScheme)
		dataMigrateInfo := &cdatamigrate.DataMigrateInfo{}
		dataMigrate := &v1alpha1.DataMigrate{
			Spec: v1alpha1.DataMigrateSpec{
				Parallelism: 2,
				ParallelOptions: map[string]string{
					cdatamigrate.SSHPort:                "120",
					cdatamigrate.SSHReadyTimeoutSeconds: "20",
				},
			},
		}

		j := &JuiceFSEngine{
			name:      "juicefs",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		}

		err := j.setParallelMigrateOptions(dataMigrateInfo, dataMigrate)
		Expect(err).NotTo(HaveOccurred())

		want := cdatamigrate.ParallelOptions{
			SSHPort:                120,
			SSHReadyTimeoutSeconds: 20,
		}
		Expect(dataMigrateInfo.ParallelOptions).To(Equal(want))
	})

	It("should use default values when no parallel options are provided", func() {
		client := fake.NewFakeClientWithScheme(testScheme)
		dataMigrateInfo := &cdatamigrate.DataMigrateInfo{}
		dataMigrate := &v1alpha1.DataMigrate{
			Spec: v1alpha1.DataMigrateSpec{
				Parallelism:     2,
				ParallelOptions: map[string]string{},
			},
		}

		j := &JuiceFSEngine{
			name:      "juicefs",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		}

		err := j.setParallelMigrateOptions(dataMigrateInfo, dataMigrate)
		Expect(err).NotTo(HaveOccurred())

		want := cdatamigrate.ParallelOptions{
			SSHPort:                cdatamigrate.DefaultSSHPort,
			SSHReadyTimeoutSeconds: cdatamigrate.DefaultSSHReadyTimeoutSeconds,
		}
		Expect(dataMigrateInfo.ParallelOptions).To(Equal(want))
	})

	It("should return error for invalid parallel options", func() {
		client := fake.NewFakeClientWithScheme(testScheme)
		dataMigrateInfo := &cdatamigrate.DataMigrateInfo{}
		dataMigrate := &v1alpha1.DataMigrate{
			Spec: v1alpha1.DataMigrateSpec{
				Parallelism: 2,
				ParallelOptions: map[string]string{
					cdatamigrate.SSHPort:                "120SS",
					cdatamigrate.SSHReadyTimeoutSeconds: "20",
				},
			},
		}

		j := &JuiceFSEngine{
			name:      "juicefs",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		}

		err := j.setParallelMigrateOptions(dataMigrateInfo, dataMigrate)
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("addWorkerPodPreferredAntiAffinity", func() {
	It("should add anti-affinity when no affinity exists", func() {
		dataMigrateInfo := &cdatamigrate.DataMigrateInfo{
			Affinity: nil,
		}
		dataMigrate := &v1alpha1.DataMigrate{
			ObjectMeta: metav1.ObjectMeta{
				Name: "dataset-migrate",
			},
			Spec: v1alpha1.DataMigrateSpec{},
		}

		addWorkerPodPreferredAntiAffinity(dataMigrateInfo, dataMigrate)

		want := &cdatamigrate.DataMigrateInfo{
			Affinity: &corev1.Affinity{
				PodAntiAffinity: &corev1.PodAntiAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
						{
							Weight: 100,
							PodAffinityTerm: corev1.PodAffinityTerm{
								LabelSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										dataoperation.OperationLabel: fmt.Sprintf("migrate-%s-%s", "", utils.GetDataMigrateReleaseName("dataset-migrate")),
									},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
				},
			},
		}
		Expect(dataMigrateInfo).To(Equal(want))
	})

	It("should add anti-affinity when pod affinity exists but no pod anti-affinity", func() {
		dataMigrateInfo := &cdatamigrate.DataMigrateInfo{
			Affinity: &corev1.Affinity{
				PodAffinity: &corev1.PodAffinity{},
			},
		}
		dataMigrate := &v1alpha1.DataMigrate{
			ObjectMeta: metav1.ObjectMeta{
				Name: "dataset-migrate",
			},
			Spec: v1alpha1.DataMigrateSpec{},
		}

		addWorkerPodPreferredAntiAffinity(dataMigrateInfo, dataMigrate)

		want := &cdatamigrate.DataMigrateInfo{
			Affinity: &corev1.Affinity{
				PodAffinity: &corev1.PodAffinity{},
				PodAntiAffinity: &corev1.PodAntiAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
						{
							Weight: 100,
							PodAffinityTerm: corev1.PodAffinityTerm{
								LabelSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										dataoperation.OperationLabel: fmt.Sprintf("migrate-%s-%s", "", utils.GetDataMigrateReleaseName("dataset-migrate")),
									},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
				},
			},
		}
		Expect(dataMigrateInfo).To(Equal(want))
	})

	It("should add anti-affinity when pod anti-affinity exists but no preferred scheduling terms", func() {
		dataMigrateInfo := &cdatamigrate.DataMigrateInfo{
			Affinity: &corev1.Affinity{
				PodAffinity:     &corev1.PodAffinity{},
				PodAntiAffinity: &corev1.PodAntiAffinity{},
			},
		}
		dataMigrate := &v1alpha1.DataMigrate{
			ObjectMeta: metav1.ObjectMeta{
				Name: "dataset-migrate",
			},
			Spec: v1alpha1.DataMigrateSpec{},
		}

		addWorkerPodPreferredAntiAffinity(dataMigrateInfo, dataMigrate)

		want := &cdatamigrate.DataMigrateInfo{
			Affinity: &corev1.Affinity{
				PodAffinity: &corev1.PodAffinity{},
				PodAntiAffinity: &corev1.PodAntiAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
						{
							Weight: 100,
							PodAffinityTerm: corev1.PodAffinityTerm{
								LabelSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										dataoperation.OperationLabel: fmt.Sprintf("migrate-%s-%s", "", utils.GetDataMigrateReleaseName("dataset-migrate")),
									},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
				},
			},
		}
		Expect(dataMigrateInfo).To(Equal(want))
	})

	It("should append anti-affinity when preferred anti-affinity terms already exist", func() {
		dataMigrateInfo := &cdatamigrate.DataMigrateInfo{
			Affinity: &corev1.Affinity{
				PodAffinity: &corev1.PodAffinity{},
				PodAntiAffinity: &corev1.PodAntiAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
						{
							Weight: 100,
							PodAffinityTerm: corev1.PodAffinityTerm{
								LabelSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"a": "b",
									},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
				},
			},
		}
		dataMigrate := &v1alpha1.DataMigrate{
			ObjectMeta: metav1.ObjectMeta{
				Name: "dataset-migrate",
			},
			Spec: v1alpha1.DataMigrateSpec{},
		}

		addWorkerPodPreferredAntiAffinity(dataMigrateInfo, dataMigrate)

		want := &cdatamigrate.DataMigrateInfo{
			Affinity: &corev1.Affinity{
				PodAffinity: &corev1.PodAffinity{},
				PodAntiAffinity: &corev1.PodAntiAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
						{
							Weight: 100,
							PodAffinityTerm: corev1.PodAffinityTerm{
								LabelSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"a": "b",
									},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
						{
							Weight: 100,
							PodAffinityTerm: corev1.PodAffinityTerm{
								LabelSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										dataoperation.OperationLabel: fmt.Sprintf("migrate-%s-%s", "", utils.GetDataMigrateReleaseName("dataset-migrate")),
									},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
				},
			},
		}
		Expect(dataMigrateInfo).To(Equal(want))
	})
})
