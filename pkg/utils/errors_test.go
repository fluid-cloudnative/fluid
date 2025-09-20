/*
Copyright 2023 The Fluid Author.

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

package utils

import (
	"fmt"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func TestIgnoreAlreadyExists(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name: "already_exists_error",
			err: apierrs.NewAlreadyExists(schema.GroupResource{
				Group:    "",
				Resource: "pod",
			}, "mypod"),
			wantErr: false,
		},
		{
			name: "not_found_error",
			err: apierrs.NewNotFound(schema.GroupResource{
				Group:    "",
				Resource: "pod",
			}, "mypod"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := IgnoreAlreadyExists(tt.err); (err != nil) != tt.wantErr {
				t.Errorf("IgnoreAlreadyExists() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIgnoreNotFound(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name: "already_exists_error",
			err: apierrs.NewAlreadyExists(schema.GroupResource{
				Group:    "",
				Resource: "pod",
			}, "mypod"),
			wantErr: true,
		},
		{
			name: "not_found_error",
			err: apierrs.NewNotFound(schema.GroupResource{
				Group:    "",
				Resource: "pod",
			}, "mypod"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := IgnoreNotFound(tt.err); (err != nil) != tt.wantErr {
				t.Errorf("IgnoreNotFound() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIgnoreNoKindMatchError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name: "not_found_error",
			err: apierrs.NewNotFound(schema.GroupResource{
				Group:    "",
				Resource: "pod",
			}, "mypod"),
			wantErr: true,
		},
		{
			name: "no_kind_match_error",
			err: &apimeta.NoKindMatchError{
				GroupKind: schema.GroupKind{
					Group: "data.fluid.io",
					Kind:  "AlluxioRuntime",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := IgnoreNoKindMatchError(tt.err); (err != nil) != tt.wantErr {
				t.Errorf("IgnoreNoKindMatchError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoggingErrorExceptConflict(t *testing.T) {
	logger := fake.NullLogger()
	result := LoggingErrorExceptConflict(logger,
		apierrs.NewConflict(schema.GroupResource{},
			"test",
			fmt.Errorf("the object has been modified; please apply your changes to the latest version and try again")),
		"Failed to setup worker",
		types.NamespacedName{
			Namespace: "test",
			Name:      "test",
		})
	if result != nil {
		t.Errorf("Expected error result is null, but got %v", result)
	}
	result = LoggingErrorExceptConflict(logger,
		apierrs.NewNotFound(schema.GroupResource{}, "test"),
		"Failed to setup worker", types.NamespacedName{
			Namespace: "test",
			Name:      "test",
		})
	if result == nil {
		t.Errorf("Expected error result is not null, but got %v", result)
	}
}
