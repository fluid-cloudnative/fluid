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
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cdatamigrate "github.com/fluid-cloudnative/fluid/pkg/datamigrate"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestJuiceFSEngine_generateDataMigrateValueFile(t *testing.T) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset-juicefs-values",
			Namespace: "fluid",
		},
		Data: map[string]string{
			"data": ``,
		},
	}

	datasetInputs := []v1alpha1.Dataset{
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
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	context := cruntime.ReconcileRequestContext{
		Client: client,
	}

	dataMigrateNoTarget := v1alpha1.DataMigrate{
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
	dataMigrateWithTarget := v1alpha1.DataMigrate{
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

	parallelDataMigrateWithTarget := v1alpha1.DataMigrate{
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

	var testCases = []struct {
		dataMigrate    v1alpha1.DataMigrate
		expectFileName string
	}{
		{
			dataMigrate:    dataMigrateNoTarget,
			expectFileName: filepath.Join(os.TempDir(), "fluid-test-datamigrate-migrate-values.yaml"),
		},
		{
			dataMigrate:    dataMigrateWithTarget,
			expectFileName: filepath.Join(os.TempDir(), "fluid-test-datamigrate-migrate-values.yaml"),
		},
		{
			dataMigrate:    parallelDataMigrateWithTarget,
			expectFileName: filepath.Join(os.TempDir(), "fluid-test-para-datamigrate-migrate-values.yaml"),
		},
	}

	for _, test := range testCases {
		runtimeInfo, err := base.BuildRuntimeInfo("juicefs", "fluid", "juicefs")
		if err != nil {
			t.Errorf("fail to create the runtimeInfo with error %v", err)
		}
		engine := JuiceFSEngine{
			name:        "juicefs",
			namespace:   "fluid",
			Client:      client,
			Log:         fake.NullLogger(),
			runtimeInfo: runtimeInfo,
		}
		fileName, err := engine.generateDataMigrateValueFile(context, &test.dataMigrate)
		if err != nil {
			t.Errorf("fail to generate the datamigrate value file: %v", err)
		}
		if !strings.Contains(fileName, test.expectFileName) {
			t.Errorf("got value: %v, want value: %v", fileName, test.expectFileName)
		}
	}
}

func TestJuiceFSEngine_genDataUrl_PVC(t *testing.T) {
	type args struct {
		data          v1alpha1.DataToMigrate
		targetDataset *v1alpha1.Dataset
		info          *cdatamigrate.DataMigrateInfo
	}

	tests := []struct {
		name        string
		args        args
		wantDataUrl string
		wantErr     bool
		wantInfo    *cdatamigrate.DataMigrateInfo
	}{
		{
			name: "test-external-pvc",
			args: args{
				data: v1alpha1.DataToMigrate{
					ExternalStorage: &v1alpha1.ExternalStorage{
						URI: "pvc://my-pvc",
					},
				},
				info: &cdatamigrate.DataMigrateInfo{},
			},
			wantDataUrl: NativeVolumeMigratePath,
			wantErr:     false,
			wantInfo: &cdatamigrate.DataMigrateInfo{
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
			},
		},
		{
			name: "test-external-pvc-subpath",
			args: args{
				data: v1alpha1.DataToMigrate{
					ExternalStorage: &v1alpha1.ExternalStorage{
						URI: "pvc://my-pvc/path/to/dir",
					},
				},
				info: &cdatamigrate.DataMigrateInfo{},
			},
			wantDataUrl: NativeVolumeMigratePath,
			wantErr:     false,
			wantInfo: &cdatamigrate.DataMigrateInfo{
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
			},
		},
		{
			name: "test-external-pvc-rootpath",
			args: args{
				data: v1alpha1.DataToMigrate{
					ExternalStorage: &v1alpha1.ExternalStorage{
						URI: "pvc://my-pvc/",
					},
				},
				info: &cdatamigrate.DataMigrateInfo{},
			},
			wantDataUrl: NativeVolumeMigratePath,
			wantErr:     false,
			wantInfo: &cdatamigrate.DataMigrateInfo{
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JuiceFSEngine{
				Client: fake.NewFakeClient(),
				Log:    fake.NullLogger(),
			}
			gotDataUrl, err := j.genDataUrl(tt.args.data, tt.args.targetDataset, tt.args.info)
			if (err != nil) != tt.wantErr {
				t.Errorf("genDataUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotDataUrl != tt.wantDataUrl {
				t.Errorf("genDataUrl() gotDataUrl = %v, want %v", gotDataUrl, tt.wantDataUrl)
			}
			if !reflect.DeepEqual(tt.args.info, tt.wantInfo) {
				t.Errorf("genDataUrl() got DataMigrateInfo = %v, want %v", tt.args.info, tt.wantInfo)
			}
		})
	}
}

// TestJuiceFSEngine_genDataUrl 测试 JuiceFSEngine 的 genDataUrl 函数
//
// 该函数用于生成数据迁移的 URL，支持以下两种数据源：
// 1. 外部存储（ExternalStorage）：通过 URI 和加密选项生成带认证信息的 URL。
// 2. 数据集存储（DataSet）：通过数据集信息和挂载点生成 JuiceFS 格式的 URL。
//
// 参数：
//   - t *testing.T：测试框架的上下文对象，用于报告测试结果和错误。
//
// 测试用例：
//   - test-external：测试外部存储的 URL 生成，验证是否正确地替换了加密选项。
//   - test-external-subpath：测试外部存储子路径的 URL 生成。
//   - test-external-subpath-file：测试外部存储子路径文件的 URL 生成。
//   - test-dataset：测试数据集存储的 URL 生成，验证是否正确地处理了子路径。
//   - test-dataset-no-path：测试数据集存储的 URL 生成（无子路径）。
//   - test-dataset-subpath-file：测试数据集存储子路径文件的 URL 生成。
//   - test-dataset-subpath-file2：测试数据集存储子路径文件的 URL 生成（带挂载点）。
//   - test-dataset-subpath-file3：测试数据集存储子路径文件的 URL 生成（带挂载点，无子路径）。
//
// 返回值：
//   - 无返回值，通过 t.Errorf 报告测试失败。
func TestJuiceFSEngine_genDataUrl(t *testing.T) {
	juicefsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "juicefs-secret",
		},
		Data: map[string][]byte{
			"access-key": []byte(base64.StdEncoding.EncodeToString([]byte("test"))),
			"secret-key": []byte(base64.StdEncoding.EncodeToString([]byte("test"))),
			"metaurl":    []byte(base64.StdEncoding.EncodeToString([]byte("test"))),
		},
	}
	juicefsConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-juicefs-values",
			Namespace: "default",
		},
		Data: map[string]string{
			"data": valuesConfigMapData,
		},
	}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, (*juicefsSecret).DeepCopy(), juicefsConfigMap)
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	type args struct {
		data          v1alpha1.DataToMigrate
		targetDataset *v1alpha1.Dataset
		info          *cdatamigrate.DataMigrateInfo
	}
	tests := []struct {
		name        string
		args        args
		wantDataUrl string
		wantErr     bool
	}{
		{
			name: "test-external",
			args: args{
				data: v1alpha1.DataToMigrate{ExternalStorage: &v1alpha1.ExternalStorage{
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
				}},
				info: &cdatamigrate.DataMigrateInfo{
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
				},
			},
			wantDataUrl: "http://${EXTERNAL_ACCESS_KEY}:${EXTERNAL_SECRET_KEY}:${EXTERNAL_TOKEN}@minio/",
			wantErr:     false,
		},
		{
			name: "test-external-subpath",
			args: args{
				data: v1alpha1.DataToMigrate{ExternalStorage: &v1alpha1.ExternalStorage{
					URI: "http://minio/test/",
					EncryptOptions: []v1alpha1.EncryptOption{{
						Name: "access-key",
						ValueFrom: v1alpha1.EncryptOptionSource{SecretKeyRef: v1alpha1.SecretKeySelector{
							Name: "juicefs-secret",
							Key:  "access-key",
						}},
					}},
				}},
				info: &cdatamigrate.DataMigrateInfo{
					EncryptOptions: []v1alpha1.EncryptOption{{
						Name: "access-key",
						ValueFrom: v1alpha1.EncryptOptionSource{SecretKeyRef: v1alpha1.SecretKeySelector{
							Name: "juicefs-secret",
							Key:  "access-key",
						}},
					}},
					Options: map[string]string{},
				},
			},
			wantDataUrl: "http://${EXTERNAL_ACCESS_KEY}:@minio/test/",
			wantErr:     false,
		},
		{
			name: "test-external-subpath-file",
			args: args{
				data: v1alpha1.DataToMigrate{ExternalStorage: &v1alpha1.ExternalStorage{
					URI: "http://minio/test",
					EncryptOptions: []v1alpha1.EncryptOption{{
						Name: "access-key",
						ValueFrom: v1alpha1.EncryptOptionSource{SecretKeyRef: v1alpha1.SecretKeySelector{
							Name: "juicefs-secret",
							Key:  "access-key",
						}},
					}},
				}},
				info: &cdatamigrate.DataMigrateInfo{
					EncryptOptions: []v1alpha1.EncryptOption{{
						Name: "access-key",
						ValueFrom: v1alpha1.EncryptOptionSource{SecretKeyRef: v1alpha1.SecretKeySelector{
							Name: "juicefs-secret",
							Key:  "access-key",
						}},
					}},
					Options: map[string]string{},
				},
			},
			wantDataUrl: "http://${EXTERNAL_ACCESS_KEY}:@minio/test",
			wantErr:     false,
		},
		{
			name: "test-dataset",
			args: args{
				data: v1alpha1.DataToMigrate{
					DataSet: &v1alpha1.DatasetToMigrate{
						Name:      "test",
						Namespace: "default",
						Path:      "/subpath/",
					},
				},
				targetDataset: &v1alpha1.Dataset{
					Spec: v1alpha1.DatasetSpec{
						Mounts: []v1alpha1.Mount{{MountPoint: "juicefs:///"}},
					},
				},
				info: &cdatamigrate.DataMigrateInfo{
					EncryptOptions: []v1alpha1.EncryptOption{},
					Options:        map[string]string{},
				},
			},
			wantDataUrl: "jfs://FLUID_METAURL/subpath/",
			wantErr:     false,
		},
		{
			name: "test-dataset-no-path",
			args: args{
				data: v1alpha1.DataToMigrate{
					DataSet: &v1alpha1.DatasetToMigrate{
						Name:      "test",
						Namespace: "default",
					},
				},
				targetDataset: &v1alpha1.Dataset{
					Spec: v1alpha1.DatasetSpec{
						Mounts: []v1alpha1.Mount{{MountPoint: "juicefs:///"}},
					},
				},
				info: &cdatamigrate.DataMigrateInfo{
					EncryptOptions: []v1alpha1.EncryptOption{},
					Options:        map[string]string{},
				},
			},
			wantDataUrl: "jfs://FLUID_METAURL/",
			wantErr:     false,
		},
		{
			name: "test-dataset-subpath-file",
			args: args{
				data: v1alpha1.DataToMigrate{
					DataSet: &v1alpha1.DatasetToMigrate{
						Name:      "test",
						Namespace: "default",
						Path:      "/subpath",
					},
				},
				targetDataset: &v1alpha1.Dataset{
					Spec: v1alpha1.DatasetSpec{
						Mounts: []v1alpha1.Mount{{MountPoint: "juicefs:///"}},
					},
				},
				info: &cdatamigrate.DataMigrateInfo{
					EncryptOptions: []v1alpha1.EncryptOption{},
					Options:        map[string]string{},
				},
			},
			wantDataUrl: "jfs://FLUID_METAURL/subpath",
			wantErr:     false,
		},
		{
			name: "test-dataset-subpath-file2",
			args: args{
				data: v1alpha1.DataToMigrate{
					DataSet: &v1alpha1.DatasetToMigrate{
						Name:      "test",
						Namespace: "default",
						Path:      "/subpath",
					},
				},
				targetDataset: &v1alpha1.Dataset{
					Spec: v1alpha1.DatasetSpec{
						Mounts: []v1alpha1.Mount{{MountPoint: "juicefs:///demo"}},
					},
				},
				info: &cdatamigrate.DataMigrateInfo{
					EncryptOptions: []v1alpha1.EncryptOption{},
					Options:        map[string]string{},
				},
			},
			wantDataUrl: "jfs://FLUID_METAURL/demo/subpath",
			wantErr:     false,
		},
		{
			name: "test-dataset-subpath-file3",
			args: args{
				data: v1alpha1.DataToMigrate{
					DataSet: &v1alpha1.DatasetToMigrate{
						Name:      "test",
						Namespace: "default",
					},
				},
				targetDataset: &v1alpha1.Dataset{
					Spec: v1alpha1.DatasetSpec{
						Mounts: []v1alpha1.Mount{{MountPoint: "juicefs:///demo"}},
					},
				},
				info: &cdatamigrate.DataMigrateInfo{
					EncryptOptions: []v1alpha1.EncryptOption{},
					Options:        map[string]string{},
				},
			},
			wantDataUrl: "jfs://FLUID_METAURL/demo",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JuiceFSEngine{
				Client: client,
				Log:    fake.NullLogger(),
			}
			gotDataUrl, err := j.genDataUrl(tt.args.data, tt.args.targetDataset, tt.args.info)
			if (err != nil) != tt.wantErr {
				t.Errorf("genDataUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotDataUrl != tt.wantDataUrl {
				t.Errorf("genDataUrl() gotDataUrl = %v, want %v", gotDataUrl, tt.wantDataUrl)
			}
		})
	}
}

// TestJuiceFSEngine_setParallelMigrateOptions tests the setParallelMigrateOptions method of JuiceFSEngine.
// This method is responsible for setting parallel migration options, including SSH port and SSH ready timeout.
// The test cases cover:
// 1. Normal scenario where parallel migration options are set correctly.
// 2. Default values are used when no parallel options are provided.
// 3. Error scenario where invalid parallel options are provided.
func TestJuiceFSEngine_setParallelMigrateOptions(t *testing.T) {
	type args struct {
		dataMigrateInfo *cdatamigrate.DataMigrateInfo
		dataMigrate     *v1alpha1.DataMigrate
	}
	tests := []struct {
		name    string
		args    args
		want    cdatamigrate.ParallelOptions
		wanterr bool
	}{
		{
			name: "test-parallel-migrate-options",
			args: args{
				dataMigrateInfo: &cdatamigrate.DataMigrateInfo{},
				dataMigrate: &v1alpha1.DataMigrate{
					Spec: v1alpha1.DataMigrateSpec{
						Parallelism: 2,
						ParallelOptions: map[string]string{
							cdatamigrate.SSHPort:                "120",
							cdatamigrate.SSHReadyTimeoutSeconds: "20",
						},
					},
				},
			},
			want: cdatamigrate.ParallelOptions{
				SSHPort:                120,
				SSHReadyTimeoutSeconds: 20,
			},
			wanterr: false,
		},
		{
			name: "test-parallel-migrate-options-default",
			args: args{
				dataMigrateInfo: &cdatamigrate.DataMigrateInfo{},
				dataMigrate: &v1alpha1.DataMigrate{
					Spec: v1alpha1.DataMigrateSpec{
						Parallelism:     2,
						ParallelOptions: map[string]string{},
					},
				},
			},
			want: cdatamigrate.ParallelOptions{
				SSHPort:                cdatamigrate.DefaultSSHPort,
				SSHReadyTimeoutSeconds: cdatamigrate.DefaultSSHReadyTimeoutSeconds,
			},
			wanterr: false,
		},
		{
			name: "test-parallel-migrate-options-wrong",
			args: args{
				dataMigrateInfo: &cdatamigrate.DataMigrateInfo{},
				dataMigrate: &v1alpha1.DataMigrate{
					Spec: v1alpha1.DataMigrateSpec{
						Parallelism: 2,
						ParallelOptions: map[string]string{
							cdatamigrate.SSHPort:                "120SS",
							cdatamigrate.SSHReadyTimeoutSeconds: "20",
						},
					},
				},
			},
			want:    cdatamigrate.ParallelOptions{},
			wanterr: true,
		},
	}
	client := fake.NewFakeClientWithScheme(testScheme)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JuiceFSEngine{
				name:      "juicefs",
				namespace: "fluid",
				Client:    client,
				Log:       fake.NullLogger(),
			}
			err := j.setParallelMigrateOptions(tt.args.dataMigrateInfo, tt.args.dataMigrate)
			if (err != nil) != tt.wanterr {
				t.Errorf("setParallelMigrateOptions() error = %v, wantErr %v", err, tt.wanterr)
				return
			}
			if err == nil && !reflect.DeepEqual(tt.want, tt.args.dataMigrateInfo.ParallelOptions) {
				t.Errorf("setParallelMigrateOptions() got = %v, want %v", tt.args.dataMigrateInfo.ParallelOptions, tt.want)
			}
		})
	}
}

func Test_addWorkerPodAntiAffinity(t *testing.T) {
	type args struct {
		dataMigrateInfo *cdatamigrate.DataMigrateInfo
		dataMigrate     *v1alpha1.DataMigrate
	}
	tests := []struct {
		name string
		args args
		want *cdatamigrate.DataMigrateInfo
	}{
		{
			name: "no affinity",
			args: args{
				dataMigrateInfo: &cdatamigrate.DataMigrateInfo{
					Affinity: nil,
				},
				dataMigrate: &v1alpha1.DataMigrate{
					ObjectMeta: metav1.ObjectMeta{
						Name: "dataset-migrate",
					},
					Spec: v1alpha1.DataMigrateSpec{},
				},
			},
			want: &cdatamigrate.DataMigrateInfo{
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
			},
		},
		{
			name: "no pod anti affinity",
			args: args{
				dataMigrateInfo: &cdatamigrate.DataMigrateInfo{
					Affinity: &corev1.Affinity{
						PodAffinity: &corev1.PodAffinity{},
					},
				},
				dataMigrate: &v1alpha1.DataMigrate{
					ObjectMeta: metav1.ObjectMeta{
						Name: "dataset-migrate",
					},
					Spec: v1alpha1.DataMigrateSpec{},
				},
			},
			want: &cdatamigrate.DataMigrateInfo{
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
			},
		},
		{
			name: "no pod anti PreferredDuringSchedulingIgnoredDuringExecution ",
			args: args{
				dataMigrateInfo: &cdatamigrate.DataMigrateInfo{
					Affinity: &corev1.Affinity{
						PodAffinity:     &corev1.PodAffinity{},
						PodAntiAffinity: &corev1.PodAntiAffinity{},
					},
				},
				dataMigrate: &v1alpha1.DataMigrate{
					ObjectMeta: metav1.ObjectMeta{
						Name: "dataset-migrate",
					},
					Spec: v1alpha1.DataMigrateSpec{},
				},
			},
			want: &cdatamigrate.DataMigrateInfo{
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
			},
		},
		{
			name: "pod anti PreferredDuringSchedulingIgnoredDuringExecution ",
			args: args{
				dataMigrateInfo: &cdatamigrate.DataMigrateInfo{
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
				},
				dataMigrate: &v1alpha1.DataMigrate{
					ObjectMeta: metav1.ObjectMeta{
						Name: "dataset-migrate",
					},
					Spec: v1alpha1.DataMigrateSpec{},
				},
			},
			want: &cdatamigrate.DataMigrateInfo{
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
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addWorkerPodPreferredAntiAffinity(tt.args.dataMigrateInfo, tt.args.dataMigrate)
			if !reflect.DeepEqual(tt.args.dataMigrateInfo, tt.want) {
				t.Errorf("addWorkerPodPreferredAntiAffinity() got = %v, want %v", tt.args.dataMigrateInfo, tt.want)
			}
		})
	}
}
