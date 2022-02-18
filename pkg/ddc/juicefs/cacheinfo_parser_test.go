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

package juicefs

import (
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	wantCacheInfo := map[string]string{"cachedir": "/tmp/jfs-cache", "mountpath": "/runtime-mnt/juicefs/fluid/test-dataset/juicefs-fuse", "command": "/bin/mount.juicefs redis://xx.xx.xx.xx:6379/1 /runtime-mnt/juicefs/fluid/test-dataset/juicefs-fuse -o metrics=0.0.0.0:9567,subdir=/dir1,cache-size=4096,free-space-ratio=0.1,cache-dir=/tmp/jfs-cache"}
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
			wantCacheInfo: map[string]string{"cachedir": "/tmp/jfs-cache", "mountpath": "/runtime-mnt/juicefs/fluid/test-dataset/juicefs-fuse", "command": "/bin/mount.juicefs redis://xx.xx.xx.xx:6379/1 /runtime-mnt/juicefs/fluid/test-dataset/juicefs-fuse -o metrics=0.0.0.0:9567,subdir=/dir1,cache-size=4096,free-space-ratio=0.1,cache-dir=/tmp/jfs-cache"},
			wantErr:       false,
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
