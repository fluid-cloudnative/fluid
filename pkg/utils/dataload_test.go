/*

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
package utils

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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

func TestGetDataLoadJob(t *testing.T) {
	mockJobName := "fluid-test-job"
	mockJobNamespace := "default"
	initJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mockJobName,
			Namespace: mockJobNamespace,
		},
	}

	fakeClient := fake.NewFakeClient(initJob)

	testCases := map[string]struct {
		name      string
		namespace string
		wantName  string
		notFound  bool
	}{
		"test get DataLoad Job case 1": {
			name:      mockJobName,
			namespace: mockJobNamespace,
			wantName:  mockJobName,
			notFound:  false,
		},
		"test get DataLoad Job case 2": {
			name:      mockJobName + "not-exist",
			namespace: mockJobNamespace,
			wantName:  "",
			notFound:  true,
		},
	}

	for k, item := range testCases {
		gotJob, err := GetDataLoadJob(fakeClient, item.name, item.namespace)
		if item.notFound {
			if err == nil && gotJob != nil {
				t.Errorf("%s check failure, want get err, but get nil", k)
			}
		} else {
			if gotJob.Name != item.wantName {
				t.Errorf("%s check failure, want DataLoad Job name:%s, got DataLoad Job name:%s", k, item.wantName, gotJob.Name)
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

func TestGetDataLoadRef(t *testing.T) {
	type args struct {
		name      string
		namespace string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				name:      "hbase",
				namespace: "default",
			},
			want: "default-hbase",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDataLoadRef(tt.args.name, tt.args.namespace); got != tt.want {
				t.Errorf("GetDataLoadRef() = %v, want %v", got, tt.want)
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
