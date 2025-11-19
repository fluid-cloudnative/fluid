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

package base

import (
	"context"
	"fmt"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
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

	err := engine.loggingErrorExceptConflict(apierrs.NewConflict(schema.GroupResource{Group: "test-group", Resource: "test-resource"}, "myresource", fmt.Errorf("Conflict error")), "test")
	if err != nil {
		t.Errorf("log should ignore conflict error, but error occured %v", err)
	}
}
