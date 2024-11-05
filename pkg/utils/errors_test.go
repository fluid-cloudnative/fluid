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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func TestLoggingErrorExceptConflict(t *testing.T) {
	logger := fake.NullLogger()
	result := LoggingErrorExceptConflict(logger,
		apierrors.NewConflict(schema.GroupResource{},
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
		apierrors.NewNotFound(schema.GroupResource{}, "test"),
		"Failed to setup worker", types.NamespacedName{
			Namespace: "test",
			Name:      "test",
		})
	if result == nil {
		t.Errorf("Expected error result is not null, but got %v", result)
	}
}
