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

package kubeclient

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestDeleteServiceAccount(t *testing.T) {
	namespace := "default"
	testSAInput := []*corev1.ServiceAccount{{
		ObjectMeta: metav1.ObjectMeta{Name: "test1", Namespace: namespace},
	}}

	testServiceAccounts := []runtime.Object{}

	for _, ns := range testSAInput {
		testServiceAccounts = append(testServiceAccounts, ns.DeepCopy())
	}

	fakeClient := fake.NewFakeClientWithScheme(testScheme, testServiceAccounts...)
	type args struct {
		name      string
		namespace string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "sa doesn't exist",
			args: args{
				name:      "notExist",
				namespace: namespace,
			},
			wantErr: false,
		},
		{
			name: "sa exist",
			args: args{
				name:      "test1",
				namespace: namespace,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteServiceAccount(fakeClient, tt.args.name, tt.args.namespace); (err != nil) != tt.wantErr {
				t.Errorf("DeleteServiceAccount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteRole(t *testing.T) {
	namespace := "default"
	testRoleInput := []*rbacv1.Role{{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: namespace},
	}}

	testRole := []runtime.Object{}

	for _, ns := range testRoleInput {
		testRole = append(testRole, ns.DeepCopy())
	}

	fakeClient := fake.NewFakeClientWithScheme(testScheme, testRole...)
	type args struct {
		name      string
		namespace string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test role not exist",
			args: args{
				name:      "notExist",
				namespace: namespace,
			},
			wantErr: false,
		},
		{
			name: "test role exist",
			args: args{
				name:      "test",
				namespace: namespace,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteRole(fakeClient, tt.args.name, tt.args.namespace); (err != nil) != tt.wantErr {
				t.Errorf("DeleteRole() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteRoleBinding(t *testing.T) {
	namespace := "default"
	testRoleBindingInput := []*rbacv1.RoleBinding{{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: namespace},
	}}

	testRoleBinding := []runtime.Object{}

	for _, ns := range testRoleBindingInput {
		testRoleBinding = append(testRoleBinding, ns.DeepCopy())
	}

	fakeClient := fake.NewFakeClientWithScheme(testScheme, testRoleBinding...)
	type args struct {
		name      string
		namespace string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test rolebinding not exist",
			args: args{
				name:      "notExist",
				namespace: namespace,
			},
			wantErr: false,
		},
		{
			name: "test rolebinding exist",
			args: args{
				name:      "test",
				namespace: namespace,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteRoleBinding(fakeClient, tt.args.name, tt.args.namespace); (err != nil) != tt.wantErr {
				t.Errorf("DeleteRoleBinding() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
