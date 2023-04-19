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

package efc

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGetReservedPorts(t *testing.T) {
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase-efc-values",
			Namespace: "fluid",
		},
		Data: map[string]string{
			"data": valuesConfigMapData,
		},
	}
	dataSets := []*v1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Status: v1alpha1.DatasetStatus{
				Runtimes: []v1alpha1.Runtime{
					{
						Name:      "hbase",
						Namespace: "fluid",
						Type:      common.EFCRuntime,
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "no-runtime",
				Namespace: "fluid",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "runtime-type",
				Namespace: "fluid",
			},
			Status: v1alpha1.DatasetStatus{
				Runtimes: []v1alpha1.Runtime{
					{
						Type: "not-efc",
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "no-map",
				Namespace: "fluid",
			},
			Status: v1alpha1.DatasetStatus{
				Runtimes: []v1alpha1.Runtime{
					{
						Type: common.EFCRuntime,
					},
				},
			},
		},
	}
	var runtimeObjs []runtime.Object
	runtimeObjs = append(runtimeObjs, configMap)
	for _, dataSet := range dataSets {
		runtimeObjs = append(runtimeObjs, dataSet.DeepCopy())
	}
	fakeClient := fake.NewFakeClientWithScheme(testScheme, runtimeObjs...)
	_, err := GetReservedPorts(fakeClient)
	if err != nil {
		t.Errorf("GetReservedPorts failed.")
	}
}
