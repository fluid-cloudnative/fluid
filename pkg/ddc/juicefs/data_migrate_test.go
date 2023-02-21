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
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cdatamigrate "github.com/fluid-cloudnative/fluid/pkg/datamigrate"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestJuiceFSEngine_genDataUrl(t *testing.T) {
	juicefsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Data: map[string][]byte{
			"access-key": []byte(base64.StdEncoding.EncodeToString([]byte("test"))),
			"secret-key": []byte(base64.StdEncoding.EncodeToString([]byte("test"))),
		},
	}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, (*juicefsSecret).DeepCopy())
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	type args struct {
		data v1alpha1.DataToMigrate
		info *cdatamigrate.DataMigrateInfo
	}
	tests := []struct {
		name        string
		args        args
		wantDataUrl string
		wantErr     bool
	}{
		{
			name: "test",
			args: args{
				data: v1alpha1.DataToMigrate{
					ExternalStorage: &v1alpha1.ExternalStorage{
						URI: "http://minio/test/",
						EncryptOptions: []v1alpha1.EncryptOption{{
							Name: "access-key",
							ValueFrom: v1alpha1.EncryptOptionSource{
								SecretKeyRef: v1alpha1.SecretKeySelector{
									Name: "test",
									Key:  "access-key",
								},
							},
						}},
					},
				},
				info: &cdatamigrate.DataMigrateInfo{
					EncryptOptions: []v1alpha1.EncryptOption{{
						Name: "access-key",
						ValueFrom: v1alpha1.EncryptOptionSource{
							SecretKeyRef: v1alpha1.SecretKeySelector{
								Name: "test",
								Key:  "access-key",
							},
						},
					}},
				},
			},
			wantDataUrl: "http://${ACCESS_KEY}:@minio/test/",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JuiceFSEngine{
				Client: client,
				Log:    fake.NullLogger(),
			}
			gotDataUrl, err := j.genDataUrl(tt.args.data, tt.args.info)
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
