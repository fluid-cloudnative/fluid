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

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestGetCacheInfoFromConfigmap(t *testing.T) {
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset-juicefs-values",
			Namespace: "fluid",
		},
		Data: map[string]string{
			"data": valuesConfigMapData,
		},
	}
	dataSet := &v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset",
			Namespace: "fluid",
		},
		Status: v1alpha1.DatasetStatus{
			Runtimes: []v1alpha1.Runtime{
				{
					Name:      "test-dataset",
					Namespace: "fluid",
					Type:      "juicefs",
				},
			},
		},
	}
	runtimeObjs := []runtime.Object{}
	runtimeObjs = append(runtimeObjs, configMap)
	runtimeObjs = append(runtimeObjs, dataSet.DeepCopy())
	fakeClient := fake.NewFakeClientWithScheme(testScheme, runtimeObjs...)
	wantCacheInfo := map[string]string{"mountpath": "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse", "edition": "community"}
	cacheinfo, err := GetCacheInfoFromConfigmap(fakeClient, dataSet.Name, dataSet.Namespace)
	if err != nil {
		t.Errorf("GetCacheInfoFromConfigmap failed.")
	}
	if !reflect.DeepEqual(cacheinfo, wantCacheInfo) {
		t.Errorf("gotcacheinfo = %v, want %v", cacheinfo, wantCacheInfo)
	}

}

func Test_parseCacheInfoFromConfigMap(t *testing.T) {
	type args struct {
		configMap *v1.ConfigMap
	}
	tests := []struct {
		name          string
		args          args
		wantCacheInfo map[string]string
		wantErr       bool
	}{
		{
			name: "parseCacheInfoFromConfigMap",
			args: args{configMap: &v1.ConfigMap{
				Data: map[string]string{
					"data": valuesConfigMapData,
				},
			}},
			wantCacheInfo: map[string]string{"mountpath": "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse", "edition": "community"},
			wantErr:       false,
		},
		{
			name: "parseCacheInfoFromConfigMap-err",
			args: args{configMap: &v1.ConfigMap{
				Data: map[string]string{
					"data": `test`,
				},
			}},
			wantCacheInfo: nil,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPorts, err := parseCacheInfoFromConfigMap(tt.args.configMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCacheInfoFromConfigMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPorts, tt.wantCacheInfo) {
				t.Errorf("parseCacheInfoFromConfigMap() gotPorts = %v, want %v", gotPorts, tt.wantCacheInfo)
			}
		})
	}
}

func TestGetFSInfoFromConfigMap(t *testing.T) {
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset-juicefs-values",
			Namespace: "fluid",
		},
		Data: map[string]string{
			"data": valuesConfigMapData,
		},
	}
	dataSet := &v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset",
			Namespace: "fluid",
		},
		Status: v1alpha1.DatasetStatus{
			Runtimes: []v1alpha1.Runtime{
				{
					Name:      "test-dataset",
					Namespace: "fluid",
					Type:      "juicefs",
				},
			},
		},
	}
	runtimeObjs := []runtime.Object{}
	runtimeObjs = append(runtimeObjs, configMap)
	runtimeObjs = append(runtimeObjs, dataSet.DeepCopy())
	fakeClient := fake.NewFakeClientWithScheme(testScheme, runtimeObjs...)
	wantMetaurlInfo := map[string]string{
		MetaurlSecret:      "jfs-secret",
		MetaurlSecretKey:   "metaurl",
		SecretKeySecret:    "jfs-secret",
		SecretKeySecretKey: "secretkey",
		TokenSecret:        "",
		TokenSecretKey:     "",
		AccessKeySecret:    "jfs-secret",
		AccessKeySecretKey: "accesskey",
		FormatCmd:          "/usr/local/bin/juicefs format --trash-days=0 --access-key=${ACCESS_KEY} --secret-key=${SECRET_KEY} --storage=minio --bucket=http://minio.default.svc.cluster.local:9000/minio/test2 ${METAURL} minio",
		Name:               "minio",
		Edition:            "community",
	}
	metaurlInfo, err := GetFSInfoFromConfigMap(fakeClient, dataSet.Name, dataSet.Namespace)
	if err != nil {
		t.Errorf("GetMetaUrlInfoFromConfigMap failed.")
	}
	if len(metaurlInfo) != len(wantMetaurlInfo) {
		t.Errorf("parseCacheInfoFromConfigMap() gotMetaurlInfo = %v,\n want %v", metaurlInfo, wantMetaurlInfo)
	}
	for k, v := range metaurlInfo {
		if v != wantMetaurlInfo[k] {
			t.Errorf("parseCacheInfoFromConfigMap() got %s = %v,\n want %v", k, v, wantMetaurlInfo[k])
		}
	}
}

func Test_parseFSInfoFromConfigMap(t *testing.T) {
	type args struct {
		configMap *v1.ConfigMap
	}
	tests := []struct {
		name            string
		args            args
		wantMetaurlInfo map[string]string
		wantErr         bool
	}{
		{
			name: "test",
			args: args{
				configMap: &v1.ConfigMap{
					Data: map[string]string{
						"data": valuesConfigMapData,
					},
				},
			},
			wantMetaurlInfo: map[string]string{
				MetaurlSecret:      "jfs-secret",
				MetaurlSecretKey:   "metaurl",
				SecretKeySecret:    "jfs-secret",
				SecretKeySecretKey: "secretkey",
				TokenSecret:        "",
				TokenSecretKey:     "",
				AccessKeySecret:    "jfs-secret",
				AccessKeySecretKey: "accesskey",
				FormatCmd:          "/usr/local/bin/juicefs format --trash-days=0 --access-key=${ACCESS_KEY} --secret-key=${SECRET_KEY} --storage=minio --bucket=http://minio.default.svc.cluster.local:9000/minio/test2 ${METAURL} minio",
				Name:               "minio",
				Edition:            "community",
			},
			wantErr: false,
		},
		{
			name: "test-err",
			args: args{
				configMap: &v1.ConfigMap{
					Data: map[string]string{
						"data": "test",
					},
				},
			},
			wantMetaurlInfo: map[string]string{},
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMetaurlInfo, err := parseFSInfoFromConfigMap(tt.args.configMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFSInfoFromConfigMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotMetaurlInfo, tt.wantMetaurlInfo) {
				t.Errorf("parseFSInfoFromConfigMap() gotMetaurlInfo = %v, want %v", gotMetaurlInfo, tt.wantMetaurlInfo)
			}
			if len(gotMetaurlInfo) != len(tt.wantMetaurlInfo) {
				t.Errorf("parseCacheInfoFromConfigMap() gotMetaurlInfo = %v,\n want %v", gotMetaurlInfo, tt.wantMetaurlInfo)
			}
			for k, v := range gotMetaurlInfo {
				if v != tt.wantMetaurlInfo[k] {
					t.Errorf("parseCacheInfoFromConfigMap() got %s = %v,\n want %v", k, v, tt.wantMetaurlInfo[k])
				}
			}
		})
	}
}
