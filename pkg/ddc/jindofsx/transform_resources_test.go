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
	"strings"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestTransformWorkerResourcesWithQuotaList(t *testing.T) {
	newRuntime := func(limit string) *datav1alpha1.JindoRuntime {
		rt := &datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Path:       "/dev/shm,/mnt/ram",
						QuotaList:  "10Gi,5Gi",
					}},
				},
			},
		}
		if limit != "" {
			rt.Spec.Worker.Resources.Limits = corev1.ResourceList{corev1.ResourceMemory: resource.MustParse(limit)}
		}
		return rt
	}

	tests := []struct {
		name       string
		limit      string
		userQuotas string
		expectErr  string
	}{
		// the summed quota is 15Gi, which exceeds the 12Gi limit
		{"sum exceeds limit", "12Gi", "10g,5g", "greater than"},
		{"invalid quota", "", "10g,abc", "invalid quota"},
	}
	for _, test := range tests {
		engine := &JindoFSxEngine{Log: fake.NullLogger()}
		err := engine.transformWorkerResources(newRuntime(test.limit), &Jindo{}, test.userQuotas)
		if err == nil || !strings.Contains(err.Error(), test.expectErr) {
			t.Errorf("%s: expected error containing %q, got %v", test.name, test.expectErr, err)
		}
	}
}
