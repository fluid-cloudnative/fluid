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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestTransformWorkerResourcesQuotaListPositive(t *testing.T) {
	rt := &datav1alpha1.JindoRuntime{
		ObjectMeta: metav1.ObjectMeta{Name: "jindo", Namespace: "fluid"},
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
	s := runtime.NewScheme()
	s.AddKnownTypes(datav1alpha1.GroupVersion, rt)
	_ = corev1.AddToScheme(s)
	engine := &JindoFSxEngine{
		name:      "jindo",
		namespace: "fluid",
		Client:    fake.NewFakeClientWithScheme(s, rt.DeepCopy()),
		Log:       fake.NullLogger(),
	}

	// Jindo hands the quotas over with Mi written as "m"; 512Mi+512Mi must come out as 1Gi.
	value := &Jindo{}
	if err := engine.transformWorkerResources(rt, value, "512m,512m"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := value.Worker.Resources.Requests.Memory; got != "1Gi" {
		t.Errorf("helm value memory request = %q, want 1Gi", got)
	}
	persisted, err := engine.getRuntime()
	if err != nil {
		t.Fatalf("getRuntime: %v", err)
	}
	if got := persisted.Spec.Worker.Resources.Requests.Memory(); got.Cmp(resource.MustParse("1Gi")) != 0 {
		t.Errorf("persisted runtime memory request = %v, want 1Gi", got)
	}
}

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
