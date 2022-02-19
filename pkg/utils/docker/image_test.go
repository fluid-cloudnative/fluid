package docker

import (
	"os"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	testScheme *runtime.Scheme
)

func init() {
	testScheme = runtime.NewScheme()
	_ = v1.AddToScheme(testScheme)
}

func TestParseDockerImage(t *testing.T) {
	var testCases = []struct {
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
	err := os.Setenv("FLUID_IMAGE_ENV", "fluid:0.6.0")
	if err != nil {
		t.Errorf("can't set the environment with error %v", err)
	}

	err = os.Setenv("ALLUXIO_IMAGE_ENV", "alluxio")
	if err != nil {
		t.Errorf("can't set the environment with error %v", err)
	}

	var testCase = []struct {
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
	err := os.Setenv("FLUID_IMAGE_ENV", "fluid:0.6.0")
	if err != nil {
		t.Errorf("can't set the environment with error %v", err)
	}

	err = os.Setenv("ALLUXIO_IMAGE_ENV", "alluxio")
	if err != nil {
		t.Errorf("can't set the environment with error %v", err)
	}

	var testCase = []struct {
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

func TestParseInitImage(t *testing.T) {
	err := os.Setenv("FLUID_IMAGE_ENV", "fluid:0.6.0")
	if err != nil {
		t.Errorf("can't set the environment with error %v", err)
	}

	var testCase = []struct {
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

	var testCase = []struct {
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
