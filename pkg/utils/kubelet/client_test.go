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

package kubelet

import (
	"errors"
	. "github.com/agiledragon/gomonkey"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

var (
	mockPodLists = "{\n    \"kind\":\"PodList\",\n    \"apiVersion\":\"v1\",\n    \"metadata\":{\n\n    },\n    \"items\":[\n        {\n            \"metadata\":{\n                \"name\":\"jfsdemo-fuse-6cnqt\",\n                \"generateName\":\"jfsdemo-fuse-\",\n                \"namespace\":\"default\",\n                \"labels\":{\n                    \"app\":\"juicefs\",\n                    \"role\":\"juicefs-fuse\"\n                }\n            },\n            \"spec\":{\n                \"containers\":[\n                    {\n                        \"name\":\"juicefs-fuse\",\n                        \"image\":\"juicedata/juicefs-csi-driver:v0.10.5\",\n                        \"command\":[\n                            \"sh\",\n                            \"-c\",\n                            \"/bin/mount.juicefs redis://10.111.100.91:6379/0 /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse -o metrics=0.0.0.0:9567,subdir=/demo,cache-dir=/dev/shm,cache-size=40960,free-space-ratio=0.1\"\n                        ],\n                        \"securityContext\":{\n                            \"privileged\":true\n                        }\n                    }\n                ]\n            }\n        }\n    ]\n}"
	privileged   = true
	mockPod      = corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:         "jfsdemo-fuse-6cnqt",
			GenerateName: "jfsdemo-fuse-",
			Namespace:    "default",
			Labels: map[string]string{
				"app":  "juicefs",
				"role": "juicefs-fuse",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:    "juicefs-fuse",
				Image:   "juicedata/juicefs-csi-driver:v0.10.5",
				Command: []string{"sh", "-c", "/bin/mount.juicefs redis://10.111.100.91:6379/0 /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse -o metrics=0.0.0.0:9567,subdir=/demo,cache-dir=/dev/shm,cache-size=40960,free-space-ratio=0.1"},
				SecurityContext: &corev1.SecurityContext{
					Privileged: &privileged,
				},
			}},
		},
	}
)

func TestKubeletClient_GetNodeRunningPods(t *testing.T) {
	Convey("TestKubeletClient_GetNodeRunningPods", t, func() {
		Convey("GetNodeRunningPods success", func() {
			httpClient := &http.Client{}
			patch1 := ApplyMethod(reflect.TypeOf(httpClient), "Get", func(_ *http.Client, url string) (resp *http.Response, err error) {
				return &http.Response{Body: io.NopCloser(strings.NewReader(mockPodLists))}, nil
			})
			defer patch1.Reset()

			kubeletClient := KubeletClient{}
			got, err := kubeletClient.GetNodeRunningPods()
			So(err, ShouldBeNil)
			So(len(got.Items), ShouldEqual, 1)
			if !reflect.DeepEqual(got.Items[0], mockPod) {
				t.Errorf("got = %v, \nwant %v", got.Items[0], mockPod)
			}
		})
		Convey("GetNodeRunningPods http client get err", func() {
			httpClient := &http.Client{}
			patch1 := ApplyMethod(reflect.TypeOf(httpClient), "Get", func(_ *http.Client, url string) (resp *http.Response, err error) {
				return nil, errors.New("test")
			})
			defer patch1.Reset()

			kubeletClient := KubeletClient{}
			got, err := kubeletClient.GetNodeRunningPods()
			So(err, ShouldNotBeNil)
			So(got, ShouldBeNil)
		})
		Convey("GetNodeRunningPods json parse err", func() {
			httpClient := &http.Client{}
			patch1 := ApplyMethod(reflect.TypeOf(httpClient), "Get", func(_ *http.Client, url string) (resp *http.Response, err error) {
				return &http.Response{Body: io.NopCloser(strings.NewReader("{-}"))}, nil
			})
			defer patch1.Reset()

			kubeletClient := KubeletClient{}
			got, err := kubeletClient.GetNodeRunningPods()
			So(err, ShouldNotBeNil)
			So(got, ShouldBeNil)
		})
	})
}

func TestKubeletClientConfig_transportConfig(t *testing.T) {
	type fields struct {
		Address         string
		Port            uint
		TLSClientConfig rest.TLSClientConfig
		BearerToken     string
		HTTPTimeout     time.Duration
	}
	tests := []struct {
		name   string
		fields fields
		want   *transport.Config
	}{
		{
			name: "test-ca",
			fields: fields{
				Address: "127.0.0.1",
				Port:    10250,
				TLSClientConfig: rest.TLSClientConfig{
					CertFile: "test",
					KeyFile:  "test",
					CAFile:   "test",
					CertData: []byte{},
					KeyData:  []byte{},
					CAData:   []byte{},
				},
				BearerToken: "test",
				HTTPTimeout: 0,
			},
			want: &transport.Config{
				TLS: transport.TLSConfig{
					CAFile:   "test",
					CertFile: "test",
					KeyFile:  "test",
					CAData:   []byte{},
					CertData: []byte{},
					KeyData:  []byte{},
				},
				BearerToken: "test",
			},
		},
		{
			name: "test-insecure",
			fields: fields{
				Address:         "127.0.0.1",
				Port:            10250,
				TLSClientConfig: rest.TLSClientConfig{},
				BearerToken:     "test",
				HTTPTimeout:     0,
			},
			want: &transport.Config{
				TLS: transport.TLSConfig{
					Insecure: true,
				},
				BearerToken: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &KubeletClientConfig{
				Address:         tt.fields.Address,
				Port:            tt.fields.Port,
				TLSClientConfig: tt.fields.TLSClientConfig,
				BearerToken:     tt.fields.BearerToken,
				HTTPTimeout:     tt.fields.HTTPTimeout,
			}
			if got := c.transportConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("transportConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_makeTransport(t *testing.T) {
	type args struct {
		config                *KubeletClientConfig
		insecureSkipTLSVerify bool
	}
	tests := []struct {
		name    string
		args    args
		want    http.RoundTripper
		wantErr bool
	}{
		{
			name: "test-nil",
			args: args{
				config: nil,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeTransport(tt.args.config, tt.args.insecureSkipTLSVerify)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeTransport() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeTransport() got = %v, want %v", got, tt.want)
			}
		})
	}
}
