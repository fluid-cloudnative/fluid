/*
Copyright 2021 The Fluid Authors.

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
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

func TestTransformFuse(t *testing.T) {
	juicefsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "fluid",
		},
		Data: map[string][]byte{
			"metaurl": []byte(base64.StdEncoding.EncodeToString([]byte("test"))),
			"name":    []byte(base64.StdEncoding.EncodeToString([]byte("test"))),
		},
	}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, (*juicefsSecret).DeepCopy())

	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
	engine := JuiceFSEngine{
		name:      "test",
		namespace: "fluid",
		Client:    client,
		Log:       log.NullLogger{},
		runtime: &datav1alpha1.JuiceFSRuntime{
			Spec: datav1alpha1.JuiceFSRuntimeSpec{
				Fuse: datav1alpha1.JuiceFSFuseSpec{},
			},
		},
	}

	var tests = []struct {
		runtime      *datav1alpha1.JuiceFSRuntime
		dataset      *datav1alpha1.Dataset
		juicefsValue *JuiceFS
		expect       string
	}{
		{&datav1alpha1.JuiceFSRuntime{
			Spec: datav1alpha1.JuiceFSRuntimeSpec{
				Fuse: datav1alpha1.JuiceFSFuseSpec{},
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
					EncryptOptions: []datav1alpha1.EncryptOption{{
						Name: "metaurl",
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{
								Name: "test",
								Key:  "metaurl",
							}}}, {
						Name: "name",
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{
								Name: "test",
								Key:  "name",
							}}}},
				}},
			}}, &JuiceFS{}, ""},
	}
	for _, test := range tests {
		err := engine.transformFuse(test.runtime, test.dataset, test.juicefsValue)
		if err != nil {
			t.Errorf("Got err %v", err)
		}
	}
}
