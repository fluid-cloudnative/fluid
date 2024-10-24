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

package volume

import (
	"context"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGetNamespacedNameByVolumeId(t *testing.T) {
	testPVs := []runtime.Object{}
	client := fake.NewFakeClientWithScheme(testScheme, testPVs...)
	testCase := []struct {
		volumeId        string
		pv              *v1.PersistentVolume
		pvc             *v1.PersistentVolumeClaim
		expectName      string
		expectNamespace string
		wantErr         bool
	}{
		{
			volumeId: "default-test",
			pv: &v1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "default-test"},
				Spec: v1.PersistentVolumeSpec{ClaimRef: &v1.ObjectReference{
					Namespace: "default",
					Name:      "test",
				}},
			},
			pvc: &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
					Labels: map[string]string{
						common.LabelAnnotationStorageCapacityPrefix + "default-test": "",
					},
				},
			},
			expectName:      "test",
			expectNamespace: "default",
			wantErr:         false,
		},
	}
	for _, test := range testCase {
		_ = client.Create(context.TODO(), test.pv)
		_ = client.Create(context.TODO(), test.pvc)
		namespace, name, err := GetNamespacedNameByVolumeId(client, test.volumeId)
		if err != nil {
			t.Errorf("failed to call GetNamespacedNameByVolumeId due to %v", err)
		}

		if name != test.expectName && namespace != test.expectNamespace {
			t.Errorf("testcase %s Expect name %s, but got %s, Expect namespace %s, but got %s",
				test.volumeId, test.expectName, name, test.expectNamespace, namespace)
		}
	}
}
