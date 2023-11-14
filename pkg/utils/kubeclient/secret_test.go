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

package kubeclient

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/utils"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetSecret(t *testing.T) {

	secretName := "mysecret"
	secretNamespace := "default"

	mockSecret1 := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: secretNamespace,
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(testScheme, mockSecret1)

	testCases := map[string]struct {
		name          string
		namespace     string
		wantName      string
		wantNamespace string
		notFound      bool
	}{
		"Case 1: get existed secret": {
			name:          secretName,
			namespace:     secretNamespace,
			wantName:      secretName,
			wantNamespace: secretNamespace,
			notFound:      false,
		},
		"Case 2: get non-existed secret": {
			name:          secretName + "not-exist",
			namespace:     secretNamespace,
			wantName:      "",
			wantNamespace: "",
			notFound:      true,
		},
	}

	for caseName, item := range testCases {
		gotSecret, err := GetSecret(fakeClient, item.name, item.namespace)
		if item.notFound {
			if err == nil || utils.IgnoreNotFound(err) != nil {
				t.Errorf("%s check failure, want not found error, but got %v", caseName, err)
			}
		} else {
			if gotSecret == nil {
				t.Errorf("%s check failure, got nil secret", caseName)
			} else if gotSecret.Name != item.wantName || gotSecret.Namespace != item.wantNamespace {
				t.Errorf("%s check failure, want secret with name %s and namespace %s, but got name %s and namespace %s",
					caseName,
					gotSecret.Name,
					gotSecret.Namespace,
					item.wantName,
					item.wantNamespace)
			}
		}
	}
}

func TestCreateSecret(t *testing.T) {

	mockSecret1 := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret2",
			Namespace: "namespace",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(testScheme, mockSecret1)
	testCases := map[string]struct {
		mockSecret *v1.Secret
		notFound   bool
	}{
		"Case 1: create new secret": {
			mockSecret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret1",
					Namespace: "namespace",
				},
			},
			notFound: true,
		},
		"Case 2: create new secret": {
			mockSecret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret1",
					Namespace: "namespace2",
				},
			},
			notFound: true,
		},
		"Case 3: create existed secret": {
			mockSecret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret2",
					Namespace: "namespace",
				},
			},
			notFound: false,
		},
	}
	for caseName, item := range testCases {
		err := CreateSecret(fakeClient, item.mockSecret)
		if err != nil {
			if !item.notFound {
				continue
			} else {
				t.Errorf("%s check failure, cannot create existed secret", caseName)
			}
		}
		gotSecret, err := GetSecret(fakeClient, item.mockSecret.Name, item.mockSecret.Namespace)
		if err != nil {
			t.Errorf("%s check failure, want not found error, but got %v", caseName, err)
		} else {
			if gotSecret == nil {
				t.Errorf("%s check failure, got nil secret", caseName)
			} else if gotSecret.Name != item.mockSecret.Name || gotSecret.Namespace != item.mockSecret.Namespace {
				t.Errorf("%s check failure, want secret with name %s and namespace %s, but got name %s and namespace %s",
					caseName,
					gotSecret.Name,
					gotSecret.Namespace,
					item.mockSecret.Name,
					item.mockSecret.Namespace)
			}
		}
	}
}

func TestUpdateSecret(t *testing.T) {

	mockSecret1 := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret2",
			Namespace: "namespace",
			Labels: map[string]string{
				"key": "old",
			},
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(testScheme, mockSecret1)
	testCases := map[string]struct {
		mockSecret *v1.Secret
		notFound   bool
	}{
		"Case 1: update new secret": {
			mockSecret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret1",
					Namespace: "namespace",
					Labels: map[string]string{
						"key": "new",
					},
				},
			},
			notFound: true,
		},
		"Case 2: update new secret": {
			mockSecret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret1",
					Namespace: "namespace2",
					Labels: map[string]string{
						"key": "new",
					},
				},
			},
			notFound: true,
		},
		"Case 3: update existed secret": {
			mockSecret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret2",
					Namespace: "namespace",
					Labels: map[string]string{
						"key": "new",
					},
				},
			},
			notFound: false,
		},
	}
	for caseName, item := range testCases {
		err := UpdateSecret(fakeClient, item.mockSecret)
		if err != nil {
			if item.notFound {
				continue
			} else {
				t.Errorf("%s check failure, cannot update unexisted secret", caseName)
			}
		}
		gotSecret, err := GetSecret(fakeClient, item.mockSecret.Name, item.mockSecret.Namespace)
		if err != nil {
			t.Errorf("%s check failure, want not found error, but got %v", caseName, err)
		} else {
			if gotSecret == nil {
				t.Errorf("%s check failure, got nil secret", caseName)
			} else if gotSecret.Labels["key"] != item.mockSecret.Labels["key"] {
				t.Errorf("%s check failure beacuse have not updated the secret", caseName)
			}
		}
	}
}
