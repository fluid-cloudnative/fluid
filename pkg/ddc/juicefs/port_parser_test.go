/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/
package juicefs

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestGetReservedPorts(t *testing.T) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myjfs-juicefs-values",
			Namespace: "fluid",
		},
		Data: map[string]string{
			"data": valuesConfigMapData,
		},
	}
	dataSets := []*v1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "myjfs", Namespace: "fluid"},
			Status: v1alpha1.DatasetStatus{
				Runtimes: []v1alpha1.Runtime{{
					Name:      "myjfs",
					Namespace: "fluid",
					Type:      "juicefs",
				}},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "no-runtime", Namespace: "fluid"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "runtime-type", Namespace: "fluid"},
			Status: v1alpha1.DatasetStatus{
				Runtimes: []v1alpha1.Runtime{{Type: "not-juicefs"}},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "no-map", Namespace: "fluid"},
			Status: v1alpha1.DatasetStatus{
				Runtimes: []v1alpha1.Runtime{{Type: "juicefs"}},
			},
		},
	}
	runtimeObjs := []runtime.Object{}
	runtimeObjs = append(runtimeObjs, configMap)
	for _, dataSet := range dataSets {
		runtimeObjs = append(runtimeObjs, dataSet.DeepCopy())
	}
	fakeClient := fake.NewFakeClientWithScheme(testScheme, runtimeObjs...)
	wantPorts := []int{14000, 14001}
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
					"data": valuesConfigMapData,
				},
			}},
			wantPorts: []int{14000, 14001},
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
