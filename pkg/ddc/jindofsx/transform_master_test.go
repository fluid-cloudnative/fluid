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

package jindofsx

import (
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestTransformToken(t *testing.T) {
	var tests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		dataset    *datav1alpha1.Dataset
		jindoValue *Jindo
		expect     string
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Secret: "secret",
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &Jindo{Secret: "test"}, "secrets:///token/"},
	}
	for _, test := range tests {
		engine := &JindoFSxEngine{Log: fake.NullLogger()}
		engine.transformToken(test.jindoValue)
		if test.jindoValue.Master.TokenProperties["default.credential.provider"] != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Master.MasterProperties["default.credential.provider"])
		}
	}
}

func TestTransformMasterMountPath(t *testing.T) {
	var tests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		dataset    *datav1alpha1.Dataset
		jindoValue *Jindo
		expect     *Level
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Secret: "secret",
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &Jindo{}, &Level{
			Path:       "/mnt/disk1",
			Type:       string(common.VolumeTypeHostPath),
			MediumType: string(common.Memory),
		}},
	}
	for _, test := range tests {
		engine := &JindoFSxEngine{Log: fake.NullLogger()}
		properties := engine.transformMasterMountPath("/mnt/disk1", common.Memory, common.VolumeTypeHostPath)
		if !reflect.DeepEqual(properties["1"], test.expect) {
			t.Errorf("expected value %v, but got %v", test.expect, properties["1"])
		}
	}
}

func TestJindoFSxEngine_transformMasterWithMultipleOSSEncryptOptions(t *testing.T) {
	secretA := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret-a",
			Namespace: "fluid",
		},
		Data: map[string][]byte{
			"fs.oss.accessKeyId":     []byte("ak-a"),
			"fs.oss.accessKeySecret": []byte("sk-a"),
		},
	}
	secretB := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret-b",
			Namespace: "fluid",
		},
		Data: map[string][]byte{
			"fs.oss.accessKeyId":     []byte("ak-b"),
			"fs.oss.accessKeySecret": []byte("sk-b"),
		},
	}

	s := runtime.NewScheme()
	s.AddKnownTypes(datav1alpha1.GroupVersion, &datav1alpha1.JindoRuntime{}, &datav1alpha1.Dataset{})
	_ = corev1.AddToScheme(s)

	client := fake.NewFakeClientWithScheme(s, secretA.DeepCopy(), secretB.DeepCopy())
	engine := JindoFSxEngine{
		name:      "test",
		namespace: "fluid",
		Client:    client,
		Log:       fake.NullLogger(),
		runtime: &datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Fuse: datav1alpha1.JindoFuseSpec{},
			},
		},
	}

	dataset := &datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{
				{
					MountPoint: "oss://bucket-a/data",
					Name:       "mount-a",
					Options: map[string]string{
						"fs.oss.endpoint": "oss-cn-shanghai.aliyuncs.com",
					},
					EncryptOptions: []datav1alpha1.EncryptOption{
						{
							Name: "fs.oss.accessKeyId",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{Name: "secret-a", Key: "fs.oss.accessKeyId"},
							},
						},
						{
							Name: "fs.oss.accessKeySecret",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{Name: "secret-a", Key: "fs.oss.accessKeySecret"},
							},
						},
					},
				},
				{
					MountPoint: "oss://bucket-b/data",
					Name:       "mount-b",
					Options: map[string]string{
						"fs.oss.endpoint": "oss-cn-shanghai.aliyuncs.com",
					},
					EncryptOptions: []datav1alpha1.EncryptOption{
						{
							Name: "fs.oss.accessKeyId",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{Name: "secret-b", Key: "fs.oss.accessKeyId"},
							},
						},
						{
							Name: "fs.oss.accessKeySecret",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{Name: "secret-b", Key: "fs.oss.accessKeySecret"},
							},
						},
					},
				},
			},
		},
	}

	value := &Jindo{}
	if err := engine.transformMaster(engine.runtime, "/test", value, dataset, true); err != nil {
		t.Fatalf("transformMaster() error = %v", err)
	}

	if _, ok := value.Master.FileStoreProperties["jindofsx.oss.bucket.bucket-a.accessKeyId"]; ok {
		t.Fatalf("expected bucket-a accessKeyId to stay out of fileStoreProperties")
	}
	if _, ok := value.Master.FileStoreProperties["jindofsx.oss.bucket.bucket-a.accessKeySecret"]; ok {
		t.Fatalf("expected bucket-a accessKeySecret to stay out of fileStoreProperties")
	}
	if _, ok := value.Master.FileStoreProperties["jindofsx.oss.bucket.bucket-a.provider.endpoint"]; ok {
		t.Fatalf("expected bucket-a provider endpoint to stay out of fileStoreProperties")
	}
	if _, ok := value.Master.FileStoreProperties["jindofsx.oss.bucket.bucket-b.provider.endpoint"]; ok {
		t.Fatalf("expected bucket-b provider endpoint to stay out of fileStoreProperties")
	}
	if value.Secret != "" || value.SecretKey != "" || value.SecretValue != "" {
		t.Fatalf("expected mount-level OSS encryptOptions not to populate single secret fields, got secret=%q secretKey=%q secretValue=%q", value.Secret, value.SecretKey, value.SecretValue)
	}
	if len(value.SecretProjections) != 2 {
		t.Fatalf("expected 2 grouped secret projections, got %d", len(value.SecretProjections))
	}
	projectedPaths := map[string]string{}
	for _, projection := range value.SecretProjections {
		for _, item := range projection.Items {
			projectedPaths[item.Path] = projection.Name + ":" + item.Key
		}
	}
	if projectedPaths["bucket-a/AccessKeyId"] != "secret-a:fs.oss.accessKeyId" {
		t.Fatalf("unexpected projection for bucket-a AccessKeyId: %q", projectedPaths["bucket-a/AccessKeyId"])
	}
	if projectedPaths["bucket-a/AccessKeySecret"] != "secret-a:fs.oss.accessKeySecret" {
		t.Fatalf("unexpected projection for bucket-a AccessKeySecret: %q", projectedPaths["bucket-a/AccessKeySecret"])
	}
	if projectedPaths["bucket-b/AccessKeyId"] != "secret-b:fs.oss.accessKeyId" {
		t.Fatalf("unexpected projection for bucket-b AccessKeyId: %q", projectedPaths["bucket-b/AccessKeyId"])
	}
	if projectedPaths["bucket-b/AccessKeySecret"] != "secret-b:fs.oss.accessKeySecret" {
		t.Fatalf("unexpected projection for bucket-b AccessKeySecret: %q", projectedPaths["bucket-b/AccessKeySecret"])
	}

	engine.transformToken(value)
	if got := value.Master.TokenProperties["jindofsx.oss.bucket.bucket-a.provider.endpoint"]; got != "secrets:///token/bucket-a/" {
		t.Fatalf("expected bucket-a token provider endpoint, got %q", got)
	}
	if got := value.Master.TokenProperties["jindofsx.oss.bucket.bucket-a.provider.format"]; got != jindoSecretProviderFormat {
		t.Fatalf("expected bucket-a token provider format %q, got %q", jindoSecretProviderFormat, got)
	}
	if got := value.Master.TokenProperties["jindofsx.oss.bucket.bucket-b.provider.endpoint"]; got != "secrets:///token/bucket-b/" {
		t.Fatalf("expected bucket-b token provider endpoint, got %q", got)
	}

	value.Worker.Port.Rpc = 6101
	engine.transformFuse(engine.runtime, value)
	if got := value.Fuse.FuseProperties["fs.oss.provider.endpoint"]; got != "secrets:///token/" {
		t.Fatalf("expected generic fuse provider endpoint, got %q", got)
	}
	if got := value.Fuse.FuseProperties["fs.oss.provider.format"]; got != jindoSecretProviderFormat {
		t.Fatalf("expected generic fuse provider format %q, got %q", jindoSecretProviderFormat, got)
	}
	if got := value.Fuse.FuseProperties["fs.oss.endpoint"]; got != "oss-cn-shanghai.aliyuncs.com" {
		t.Fatalf("expected generic fuse endpoint, got %q", got)
	}
	if got := value.Fuse.FuseProperties["aliyun.oss.bucket.bucket-a.provider.url"]; got != "secrets:///token/bucket-a/" {
		t.Fatalf("expected bucket-a fuse provider url, got %q", got)
	}
	if got := value.Fuse.FuseProperties["fs.oss.bucket.bucket-b.credentials.provider"]; got != jindoOSSCredentialsProvider {
		t.Fatalf("expected bucket-b fuse credentials provider %q, got %q", jindoOSSCredentialsProvider, got)
	}
}

func TestJindoFSxEngine_transformMasterDedupesSameBucketSecretProjection(t *testing.T) {
	s := runtime.NewScheme()
	s.AddKnownTypes(datav1alpha1.GroupVersion, &datav1alpha1.JindoRuntime{}, &datav1alpha1.Dataset{})
	_ = corev1.AddToScheme(s)

	engine := JindoFSxEngine{
		name:      "test",
		namespace: "fluid",
		Client:    fake.NewFakeClientWithScheme(s),
		Log:       fake.NullLogger(),
		runtime: &datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Fuse: datav1alpha1.JindoFuseSpec{},
			},
		},
	}

	dataset := &datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{
				{
					MountPoint: "oss://bucket-a/data-1",
					Name:       "mount-a",
					Options: map[string]string{
						"fs.oss.endpoint": "oss-cn-shanghai.aliyuncs.com",
					},
					EncryptOptions: []datav1alpha1.EncryptOption{
						{
							Name: "fs.oss.accessKeyId",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{Name: "secret-a", Key: "fs.oss.accessKeyId"},
							},
						},
						{
							Name: "fs.oss.accessKeySecret",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{Name: "secret-a", Key: "fs.oss.accessKeySecret"},
							},
						},
					},
				},
				{
					MountPoint: "oss://bucket-a/data-2",
					Name:       "mount-b",
					Options: map[string]string{
						"fs.oss.endpoint": "oss-cn-shanghai.aliyuncs.com",
					},
					EncryptOptions: []datav1alpha1.EncryptOption{
						{
							Name: "fs.oss.accessKeyId",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{Name: "secret-a", Key: "fs.oss.accessKeyId"},
							},
						},
						{
							Name: "fs.oss.accessKeySecret",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{Name: "secret-a", Key: "fs.oss.accessKeySecret"},
							},
						},
					},
				},
			},
		},
	}

	value := &Jindo{}
	if err := engine.transformMaster(engine.runtime, "/test", value, dataset, true); err != nil {
		t.Fatalf("transformMaster() error = %v", err)
	}

	if len(value.SecretProjections) != 1 {
		t.Fatalf("expected same-bucket projections to dedupe to 1 grouped entry, got %d", len(value.SecretProjections))
	}
}

func TestJindoFSxEngine_transformMasterRejectsConflictingSameBucketSecretProjection(t *testing.T) {
	s := runtime.NewScheme()
	s.AddKnownTypes(datav1alpha1.GroupVersion, &datav1alpha1.JindoRuntime{}, &datav1alpha1.Dataset{})
	_ = corev1.AddToScheme(s)

	engine := JindoFSxEngine{
		name:      "test",
		namespace: "fluid",
		Client:    fake.NewFakeClientWithScheme(s),
		Log:       fake.NullLogger(),
		runtime: &datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Fuse: datav1alpha1.JindoFuseSpec{},
			},
		},
	}

	dataset := &datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{
				{
					MountPoint: "oss://bucket-a/data-1",
					Name:       "mount-a",
					Options: map[string]string{
						"fs.oss.endpoint": "oss-cn-shanghai.aliyuncs.com",
					},
					EncryptOptions: []datav1alpha1.EncryptOption{
						{
							Name: "fs.oss.accessKeyId",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{Name: "secret-a", Key: "fs.oss.accessKeyId"},
							},
						},
						{
							Name: "fs.oss.accessKeySecret",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{Name: "secret-a", Key: "fs.oss.accessKeySecret"},
							},
						},
					},
				},
				{
					MountPoint: "oss://bucket-a/data-2",
					Name:       "mount-b",
					Options: map[string]string{
						"fs.oss.endpoint": "oss-cn-shanghai.aliyuncs.com",
					},
					EncryptOptions: []datav1alpha1.EncryptOption{
						{
							Name: "fs.oss.accessKeyId",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{Name: "secret-b", Key: "fs.oss.accessKeyId"},
							},
						},
						{
							Name: "fs.oss.accessKeySecret",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{Name: "secret-b", Key: "fs.oss.accessKeySecret"},
							},
						},
					},
				},
			},
		},
	}

	if err := engine.transformMaster(engine.runtime, "/test", &Jindo{}, dataset, true); err == nil {
		t.Fatalf("expected transformMaster() to reject conflicting same-bucket secret projections")
	}
}

func TestJindoFSxEngine_transformMasterSupportsInlineOSSCredentialsCompatibility(t *testing.T) {
	s := runtime.NewScheme()
	s.AddKnownTypes(datav1alpha1.GroupVersion, &datav1alpha1.JindoRuntime{}, &datav1alpha1.Dataset{})
	_ = corev1.AddToScheme(s)

	engine := JindoFSxEngine{
		name:      "test",
		namespace: "fluid",
		Client:    fake.NewFakeClientWithScheme(s),
		Log:       fake.NullLogger(),
		runtime: &datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Fuse: datav1alpha1.JindoFuseSpec{},
			},
		},
	}

	dataset := &datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{{
				MountPoint: "oss://bucket-a/data",
				Name:       "mount-a",
				Options: map[string]string{
					"fs.oss.endpoint":        "oss-cn-shanghai.aliyuncs.com",
					"fs.oss.accessKeyId":     "inline-ak",
					"fs.oss.accessKeySecret": "inline-sk",
				},
			}},
		},
	}

	value := &Jindo{}
	if err := engine.transformMaster(engine.runtime, "/test", value, dataset, true); err != nil {
		t.Fatalf("transformMaster() error = %v", err)
	}

	if got := value.Master.FileStoreProperties["jindofsx.oss.bucket.bucket-a.endpoint"]; got != "oss-cn-shanghai.aliyuncs.com" {
		t.Fatalf("expected bucket-a endpoint to be preserved, got %q", got)
	}
	if got := value.Master.FileStoreProperties["jindofsx.oss.bucket.bucket-a.accessKeyId"]; got != "inline-ak" {
		t.Fatalf("expected inline bucket-a accessKeyId to be preserved, got %q", got)
	}
	if got := value.Master.FileStoreProperties["jindofsx.oss.bucket.bucket-a.accessKeySecret"]; got != "inline-sk" {
		t.Fatalf("expected inline bucket-a accessKeySecret to be preserved, got %q", got)
	}
	if len(value.SecretProjections) != 0 {
		t.Fatalf("expected no secret projections for inline credentials, got %d", len(value.SecretProjections))
	}
}

func TestJindoFSxEngine_transformMasterIgnoresNonOSSCredentialEncryptOptionForBucketSecretProjection(t *testing.T) {
	s := runtime.NewScheme()
	s.AddKnownTypes(datav1alpha1.GroupVersion, &datav1alpha1.JindoRuntime{}, &datav1alpha1.Dataset{})
	_ = corev1.AddToScheme(s)

	engine := JindoFSxEngine{
		name:      "test",
		namespace: "fluid",
		Client:    fake.NewFakeClientWithScheme(s),
		Log:       fake.NullLogger(),
		runtime: &datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Fuse: datav1alpha1.JindoFuseSpec{},
			},
		},
	}

	dataset := &datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{{
				MountPoint: "oss://bucket-a/data",
				Name:       "mount-a",
				Options: map[string]string{
					"fs.oss.endpoint": "oss-cn-shanghai.aliyuncs.com",
				},
				EncryptOptions: []datav1alpha1.EncryptOption{{
					Name: "fs.oss.sessionToken",
					ValueFrom: datav1alpha1.EncryptOptionSource{
						SecretKeyRef: datav1alpha1.SecretKeySelector{Name: "secret-a", Key: "token"},
					},
				}},
			}},
		},
	}

	value := &Jindo{}
	if err := engine.transformMaster(engine.runtime, "/test", value, dataset, true); err != nil {
		t.Fatalf("transformMaster() error = %v", err)
	}

	if len(value.SecretProjections) != 0 {
		t.Fatalf("expected no secret projections for non-AK/SK encryptOptions, got %d", len(value.SecretProjections))
	}
	if len(value.BucketSecretPaths) != 0 {
		t.Fatalf("expected no bucket secret paths for non-AK/SK encryptOptions, got %#v", value.BucketSecretPaths)
	}
}

func TestJindoFSxEngine_transformMasterUsesReferencedSecretKeysForNonOSSMounts(t *testing.T) {
	s := runtime.NewScheme()
	s.AddKnownTypes(datav1alpha1.GroupVersion, &datav1alpha1.JindoRuntime{}, &datav1alpha1.Dataset{})
	_ = corev1.AddToScheme(s)

	engine := JindoFSxEngine{
		name:      "test",
		namespace: "fluid",
		Client:    fake.NewFakeClientWithScheme(s),
		Log:       fake.NullLogger(),
		runtime: &datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Fuse: datav1alpha1.JindoFuseSpec{},
			},
		},
	}

	dataset := &datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{{
				MountPoint: "s3://bucket-a/data",
				Name:       "mount-a",
				Options: map[string]string{
					"fs.s3.endpoint": "s3.us-west-1.amazonaws.com",
				},
				EncryptOptions: []datav1alpha1.EncryptOption{
					{
						Name: "fs.s3.accessKeyId",
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{Name: "secret-a", Key: "custom-ak"},
						},
					},
					{
						Name: "fs.s3.accessKeySecret",
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{Name: "secret-a", Key: "custom-sk"},
						},
					},
				},
			}},
		},
	}

	value := &Jindo{}
	if err := engine.transformMaster(engine.runtime, "/test", value, dataset, true); err != nil {
		t.Fatalf("transformMaster() error = %v", err)
	}

	if value.Secret != "secret-a" {
		t.Fatalf("expected secret name %q, got %q", "secret-a", value.Secret)
	}
	if value.SecretKey != "custom-ak" {
		t.Fatalf("expected SecretKey to use referenced key %q, got %q", "custom-ak", value.SecretKey)
	}
	if value.SecretValue != "custom-sk" {
		t.Fatalf("expected SecretValue to use referenced key %q, got %q", "custom-sk", value.SecretValue)
	}
}

func TestJindoFSxEngine_transformMasterReturnsErrorWhenReferencedSecretMissing(t *testing.T) {
	s := runtime.NewScheme()
	s.AddKnownTypes(datav1alpha1.GroupVersion, &datav1alpha1.JindoRuntime{}, &datav1alpha1.Dataset{})
	_ = corev1.AddToScheme(s)

	engine := JindoFSxEngine{
		name:      "test",
		namespace: "fluid",
		Client:    fake.NewFakeClientWithScheme(s),
		Log:       fake.NullLogger(),
		runtime: &datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Fuse: datav1alpha1.JindoFuseSpec{},
			},
		},
	}

	dataset := &datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{{
				MountPoint: "oss://bucket-a/data",
				Name:       "mount-a",
				Options: map[string]string{
					"fs.oss.endpoint": "oss-cn-shanghai.aliyuncs.com",
				},
				EncryptOptions: []datav1alpha1.EncryptOption{{
					Name: "fs.oss.accessKeyId",
					ValueFrom: datav1alpha1.EncryptOptionSource{
						SecretKeyRef: datav1alpha1.SecretKeySelector{Name: "missing-secret", Key: "fs.oss.accessKeyId"},
					},
				}},
			}},
		},
	}

	if err := engine.transformMaster(engine.runtime, "/test", &Jindo{}, dataset, false); err == nil {
		t.Fatalf("expected transformMaster() to fail when the referenced secret is missing")
	}
}

func TestJindoFSxEngine_transformMasterRejectsEmptyReferencedSecretKey(t *testing.T) {
	s := runtime.NewScheme()
	s.AddKnownTypes(datav1alpha1.GroupVersion, &datav1alpha1.JindoRuntime{}, &datav1alpha1.Dataset{})
	_ = corev1.AddToScheme(s)

	engine := JindoFSxEngine{
		name:      "test",
		namespace: "fluid",
		Client:    fake.NewFakeClientWithScheme(s),
		Log:       fake.NullLogger(),
		runtime: &datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Fuse: datav1alpha1.JindoFuseSpec{},
			},
		},
	}

	dataset := &datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{{
				MountPoint: "oss://bucket-a/data",
				Name:       "mount-a",
				Options: map[string]string{
					"fs.oss.endpoint": "oss-cn-shanghai.aliyuncs.com",
				},
				EncryptOptions: []datav1alpha1.EncryptOption{
					{
						Name: "fs.oss.accessKeyId",
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{Name: "secret-a"},
						},
					},
					{
						Name: "fs.oss.accessKeySecret",
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{Name: "secret-a", Key: "fs.oss.accessKeySecret"},
						},
					},
				},
			}},
		},
	}

	if err := engine.transformMaster(engine.runtime, "/test", &Jindo{}, dataset, true); err == nil {
		t.Fatalf("expected transformMaster() to fail when the referenced secret key is empty")
	}
}

func TestJindoFSxEngine_transformMasterReturnsErrorWhenReferencedSecretDataKeyMissing(t *testing.T) {
	s := runtime.NewScheme()
	s.AddKnownTypes(datav1alpha1.GroupVersion, &datav1alpha1.JindoRuntime{}, &datav1alpha1.Dataset{})
	_ = corev1.AddToScheme(s)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret-a",
			Namespace: "fluid",
		},
		Data: map[string][]byte{
			"unexpected-key": []byte("ak-a"),
		},
	}

	engine := JindoFSxEngine{
		name:      "test",
		namespace: "fluid",
		Client:    fake.NewFakeClientWithScheme(s, secret.DeepCopy()),
		Log:       fake.NullLogger(),
		runtime: &datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Fuse: datav1alpha1.JindoFuseSpec{},
			},
		},
	}

	dataset := &datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{{
				MountPoint: "oss://bucket-a/data",
				Name:       "mount-a",
				Options: map[string]string{
					"fs.oss.endpoint": "oss-cn-shanghai.aliyuncs.com",
				},
				EncryptOptions: []datav1alpha1.EncryptOption{{
					Name: "fs.oss.accessKeyId",
					ValueFrom: datav1alpha1.EncryptOptionSource{
						SecretKeyRef: datav1alpha1.SecretKeySelector{Name: "secret-a", Key: "fs.oss.accessKeyId"},
					},
				}},
			}},
		},
	}

	if err := engine.transformMaster(engine.runtime, "/test", &Jindo{}, dataset, false); err == nil {
		t.Fatalf("expected transformMaster() to fail when the referenced secret key is missing from secret data")
	}
}

func TestJindoFSxEngine_transformMasterAcceptsBucketRootMountPoint(t *testing.T) {
	s := runtime.NewScheme()
	s.AddKnownTypes(datav1alpha1.GroupVersion, &datav1alpha1.JindoRuntime{}, &datav1alpha1.Dataset{})
	_ = corev1.AddToScheme(s)

	engine := JindoFSxEngine{
		name:      "test",
		namespace: "fluid",
		Client:    fake.NewFakeClientWithScheme(s),
		Log:       fake.NullLogger(),
		runtime: &datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Fuse: datav1alpha1.JindoFuseSpec{},
			},
		},
	}

	dataset := &datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{{
				MountPoint: "oss://bucket-root",
				Name:       "mount-a",
				Options: map[string]string{
					"fs.oss.endpoint": "oss-cn-shanghai.aliyuncs.com",
				},
			}},
		},
	}

	value := &Jindo{}
	if err := engine.transformMaster(engine.runtime, "/test", value, dataset, true); err != nil {
		t.Fatalf("transformMaster() error = %v", err)
	}

	if got := value.Master.FileStoreProperties["jindofsx.oss.bucket.bucket-root.endpoint"]; got != "oss-cn-shanghai.aliyuncs.com" {
		t.Fatalf("expected bucket-root endpoint to be preserved, got %q", got)
	}
}

func TestJindoFSxEngine_transformMasterRejectsMixedOSSInlineAndSecretProjection(t *testing.T) {
	s := runtime.NewScheme()
	s.AddKnownTypes(datav1alpha1.GroupVersion, &datav1alpha1.JindoRuntime{}, &datav1alpha1.Dataset{})
	_ = corev1.AddToScheme(s)

	engine := JindoFSxEngine{
		name:      "test",
		namespace: "fluid",
		Client:    fake.NewFakeClientWithScheme(s),
		Log:       fake.NullLogger(),
		runtime: &datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Fuse: datav1alpha1.JindoFuseSpec{},
			},
		},
	}

	dataset := &datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{{
				MountPoint: "oss://bucket-a/data",
				Name:       "mount-a",
				Options: map[string]string{
					"fs.oss.endpoint":        "oss-cn-shanghai.aliyuncs.com",
					"fs.oss.accessKeySecret": "inline-sk",
				},
				EncryptOptions: []datav1alpha1.EncryptOption{{
					Name: "fs.oss.accessKeyId",
					ValueFrom: datav1alpha1.EncryptOptionSource{
						SecretKeyRef: datav1alpha1.SecretKeySelector{Name: "secret-a", Key: "custom-ak"},
					},
				}},
			}},
		},
	}

	if err := engine.transformMaster(engine.runtime, "/test", &Jindo{}, dataset, true); err == nil {
		t.Fatalf("expected transformMaster() to reject mixed inline and secret-projected OSS credentials")
	}
}
