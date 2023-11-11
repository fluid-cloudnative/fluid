/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
