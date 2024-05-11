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

package vineyard

import (
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	dummyConfigMapData = `master:
  hostNetwork: true
  ports:
    client: 14001
    peer: 14002
worker:
  hostNetwork: true
  ports:
    rpc: 14003
    exporter: 14004
`
)

func TestGetReservedPorts(t *testing.T) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-vineyard-values",
			Namespace: "fluid",
		},
		Data: map[string]string{
			"data": mockConfigMapData,
		},
	}
	dataSets := []*v1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "fluid"},
			Status: v1alpha1.DatasetStatus{
				Runtimes: []v1alpha1.Runtime{{
					Name:      "test",
					Namespace: "fluid",
					Type:      "vineyard",
				}},
			},
		},
	}
	runtimeObjs := []runtime.Object{}
	runtimeObjs = append(runtimeObjs, configMap)
	for _, dataSet := range dataSets {
		runtimeObjs = append(runtimeObjs, dataSet.DeepCopy())
	}
	fakeClient := fake.NewFakeClientWithScheme(testScheme, runtimeObjs...)
	wantPorts := []int{14001, 14002, 14003, 14004}
	ports, err := GetReservedPorts(fakeClient)
	if err != nil {
		t.Errorf("GetReservedPorts failed.")
	}
	if !reflect.DeepEqual(ports, wantPorts) {
		t.Errorf("gotPorts = %v, want %v", ports, wantPorts)
	}

}

func Test_parsePortsFromConfigMap(t *testing.T) {
	type args struct {
		configMap *corev1.ConfigMap
	}
	tests := []struct {
		name      string
		args      args
		wantPorts []int
		wantErr   bool
	}{
		{
			name: "parsePortsFromConfigMap",
			args: args{configMap: &corev1.ConfigMap{
				Data: map[string]string{
					"data": dummyConfigMapData,
				},
			}},
			wantPorts: []int{14001, 14002, 14003, 14004},
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPorts, err := parsePortsFromConfigMap(tt.args.configMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePortsFromConfigMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPorts, tt.wantPorts) {
				t.Errorf("parsePortsFromConfigMap() gotPorts = %v, want %v", gotPorts, tt.wantPorts)
			}
		})
	}
}
