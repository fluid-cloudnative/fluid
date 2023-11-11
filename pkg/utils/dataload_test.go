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

package utils

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGetDataLoad(t *testing.T) {
	mockDataLoadName := "fluid-test-data-load"
	mockDataLoadNamespace := "default"
	initDataLoad := &datav1alpha1.DataLoad{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mockDataLoadName,
			Namespace: mockDataLoadNamespace,
		},
	}

	s := runtime.NewScheme()
	s.AddKnownTypes(datav1alpha1.GroupVersion, initDataLoad)

	fakeClient := fake.NewFakeClientWithScheme(s, initDataLoad)

	testCases := map[string]struct {
		name      string
		namespace string
		wantName  string
		notFound  bool
	}{
		"test get DataLoad case 1": {
			name:      mockDataLoadName,
			namespace: mockDataLoadNamespace,
			wantName:  mockDataLoadName,
			notFound:  false,
		},
		"test get DataLoad case 2": {
			name:      mockDataLoadName + "not-exist",
			namespace: mockDataLoadNamespace,
			wantName:  "",
			notFound:  true,
		},
	}

	for k, item := range testCases {
		gotDataLoad, err := GetDataLoad(fakeClient, item.name, item.namespace)
		if item.notFound {
			if err == nil && gotDataLoad != nil {
				t.Errorf("%s check failure, want get err, but get nil", k)
			}
		} else {
			if gotDataLoad.Name != item.wantName {
				t.Errorf("%s check failure, want DataLoad name:%s, got DataLoad name:%s", k, item.wantName, gotDataLoad.Name)
			}
		}
	}

}

func TestGetDataLoadReleaseName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				name: "imagenet-dataload",
			},
			want: "imagenet-dataload-loader",
		},
		{
			name: "test2",
			args: args{
				name: "test",
			},
			want: "test-loader",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDataLoadReleaseName(tt.args.name); got != tt.want {
				t.Errorf("GetDataLoadReleaseName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDataLoadJobName(t *testing.T) {
	type args struct {
		releaseName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{releaseName: GetDataLoadReleaseName("hbase")},
			want: "hbase-loader-job",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDataLoadJobName(tt.args.releaseName); got != tt.want {
				t.Errorf("GetDataLoadJobName() = %v, want %v", got, tt.want)
			}
		})
	}
}
