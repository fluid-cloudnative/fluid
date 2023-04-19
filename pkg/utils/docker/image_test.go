/*
Copyright 2023 The Fluid Author.

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
package docker

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

var testScheme *runtime.Scheme

func init() {
	testScheme = runtime.NewScheme()
	_ = v1.AddToScheme(testScheme)
}

func TestParseDockerImage(t *testing.T) {
	testCases := []struct {
		input string
		image string
		tag   string
	}{
		{"test:abc", "test", "abc"},
		{"test", "test", "latest"},
		{"test:35000/test:abc", "test:35000/test", "abc"},
	}
	for _, tc := range testCases {
		image, tag := ParseDockerImage(tc.input)
		if tc.image != image {
			t.Errorf("expected image %#v, got %#v",
				tc.image, image)
		}

		if tc.tag != tag {
			t.Errorf("expected tag %#v, got %#v",
				tc.tag, tag)
		}
	}
}

func TestGetImageRepoFromEnv(t *testing.T) {
	t.Setenv("FLUID_IMAGE_ENV", "fluid:0.6.0")
	t.Setenv("ALLUXIO_IMAGE_ENV", "alluxio")

	testCase := []struct {
		envName string
		want    string
	}{
		{
			envName: "FLUID_IMAGE_ENV",
			want:    "fluid",
		},
		{
			envName: "NOT EXIST",
			want:    "",
		},
		{
			envName: "ALLUXIO_IMAGE_ENV",
			want:    "",
		},
	}

	for _, test := range testCase {
		if result := GetImageRepoFromEnv(test.envName); result != test.want {
			t.Errorf("expected %v, got %v", test.want, result)
		}
	}
}

func TestGetImageTagFromEnv(t *testing.T) {
	t.Setenv("FLUID_IMAGE_ENV", "fluid:0.6.0")
	t.Setenv("ALLUXIO_IMAGE_ENV", "alluxio")

	testCase := []struct {
		envName string
		want    string
	}{
		{
			envName: "FLUID_IMAGE_ENV",
			want:    "0.6.0",
		},
		{
			envName: "NOT EXIST",
			want:    "",
		},
		{
			envName: "ALLUXIO_IMAGE_ENV",
			want:    "",
		},
	}
	for _, test := range testCase {
		if result := GetImageTagFromEnv(test.envName); result != test.want {
			t.Errorf("expected %v, got %v", test.want, result)
		}
	}
}

func TestGetImagePullSecrets(t *testing.T) {
	testCases := map[string]struct {
		envName       string
		envMockValues string
		want          []v1.LocalObjectReference
	}{
		"test with env value case 1": {
			envName:       common.EnvImagePullSecretsKey,
			envMockValues: "test1,test2",
			want: []v1.LocalObjectReference{
				{
					Name: "test1",
				},
				{
					Name: "test2",
				},
			},
		},
		"test with env value case 2": {
			envName:       common.EnvImagePullSecretsKey,
			envMockValues: "",
			want:          []v1.LocalObjectReference{},
		},
		"test with env value case 3": {
			envName:       common.EnvImagePullSecretsKey,
			envMockValues: "str1",
			want:          []v1.LocalObjectReference{{Name: "str1"}},
		},
		"test with env value case 4": {
			envName:       common.EnvImagePullSecretsKey,
			envMockValues: "str1,",
			want:          []v1.LocalObjectReference{{Name: "str1"}},
		},
		"test with env value case 5": {
			envName:       common.EnvImagePullSecretsKey,
			envMockValues: ",,,str1,",
			want:          []v1.LocalObjectReference{{Name: "str1"}},
		},
		"test with env value case 6": {
			envName:       common.EnvImagePullSecretsKey,
			envMockValues: ",,,str1,,,str2,,",
			want:          []v1.LocalObjectReference{{Name: "str1"}, {Name: "str2"}},
		},
	}

	for k, v := range testCases {
		t.Setenv(v.envName, v.envMockValues)
		got := GetImagePullSecretsFromEnv(v.envName)
		if !reflect.DeepEqual(got, v.want) {
			t.Errorf("%s: expected: %s, got: %s", k, v.want, got)
		}
	}
}

func TestParseInitImage(t *testing.T) {
	t.Setenv("FLUID_IMAGE_ENV", "fluid:0.6.0")

	testCase := []struct {
		image               string
		tag                 string
		imagePullPolicy     string
		envName             string
		wantImage           string
		wantTag             string
		wantImagePullPolicy string
	}{
		{
			image:               "fluid",
			tag:                 "0.6.0",
			imagePullPolicy:     "",
			envName:             "FLUID_IMAGE_ENV",
			wantImage:           "fluid",
			wantTag:             "0.6.0",
			wantImagePullPolicy: common.DefaultImagePullPolicy,
		},
		{
			image:               "",
			tag:                 "0.6.0",
			imagePullPolicy:     "Always",
			envName:             "FLUID_IMAGE_ENV",
			wantImage:           "fluid",
			wantTag:             "0.6.0",
			wantImagePullPolicy: "Always",
		},
		{
			image:               "fluid",
			tag:                 "",
			imagePullPolicy:     "Always",
			envName:             "FLUID_IMAGE_ENV",
			wantImage:           "fluid",
			wantTag:             "0.6.0",
			wantImagePullPolicy: "Always",
		},
		{
			image:               "fluid",
			tag:                 "0.6.0",
			imagePullPolicy:     "Always",
			envName:             "FLUID_IMAGE_ENV",
			wantImage:           "fluid",
			wantTag:             "0.6.0",
			wantImagePullPolicy: "Always",
		},
	}
	for _, test := range testCase {
		resultImage, resultTag, resultImagePullPolicy := ParseInitImage(test.image, test.tag, test.imagePullPolicy, test.envName)
		if resultImage != test.wantImage {
			t.Errorf("expected %v, got %v", test.wantImage, resultImage)
		}
		if resultTag != test.wantTag {
			t.Errorf("expected %v, got %v", test.wantTag, resultTag)
		}
		if resultImagePullPolicy != test.wantImagePullPolicy {
			t.Errorf("expected %v, got %v", test.wantImagePullPolicy, resultImagePullPolicy)
		}
	}
}

func TestGetWorkerImage(t *testing.T) {
	configMapInputs := []*v1.ConfigMap{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "hbase-alluxio-values", Namespace: "default"},
			Data: map[string]string{
				"data": "image: fluid\nimageTag: 0.6.0",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "spark-alluxio-values", Namespace: "default"},
			Data: map[string]string{
				"test-data": "image: fluid\n imageTag: 0.6.0",
			},
		},
	}

	testConfigMaps := []runtime.Object{}
	for _, cm := range configMapInputs {
		testConfigMaps = append(testConfigMaps, cm.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testConfigMaps...)

	testCase := []struct {
		datasetName   string
		runtimeType   string
		namespace     string
		wantImageName string
		wantImageTag  string
	}{
		{
			datasetName:   "hbase",
			runtimeType:   "jindoruntime",
			namespace:     "fluid",
			wantImageName: "",
			wantImageTag:  "",
		},
		{
			datasetName:   "hbase",
			runtimeType:   "alluxio",
			namespace:     "default",
			wantImageName: "fluid",
			wantImageTag:  "0.6.0",
		},
		{
			datasetName:   "spark",
			runtimeType:   "alluxio",
			namespace:     "default",
			wantImageName: "",
			wantImageTag:  "",
		},
	}

	for _, test := range testCase {
		resultImageName, resultImageTag := GetWorkerImage(client, test.datasetName, test.runtimeType, test.namespace)
		if resultImageName != test.wantImageName {
			t.Errorf("expected %v, got %v", test.wantImageName, resultImageName)
		}
		if resultImageTag != test.wantImageTag {
			t.Errorf("expected %v, got %v", test.wantImageTag, resultImageTag)
		}
	}
}
