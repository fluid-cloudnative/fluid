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

package base

import (
	"context"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"

	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
)

func TestLoggingErrorExceptConflict(t *testing.T) {

	engine := NewTemplateEngine(nil, "id", runtime.ReconcileRequestContext{
		Context: context.Background(),
		NamespacedName: types.NamespacedName{
			Namespace: "test",
			Name:      "test",
		},
		Runtime: &datav1alpha1.AlluxioRuntime{},
		Log:     fake.NullLogger(),
	})

	err := engine.loggingErrorExceptConflict(fluiderrs.NewDeprecated(schema.GroupResource{Group: "", Resource: "test"}, types.NamespacedName{}), "test")
	if !fluiderrs.IsDeprecated(err) {
		t.Errorf("Failed to check deprecated error %v", err)
	}
}
