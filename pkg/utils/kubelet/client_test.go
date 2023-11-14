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
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		Convey("GetNodeRunningPods client err", func() {
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
		Convey("GetNodeRunningPods json err", func() {
			httpClient := &http.Client{}
			patch1 := ApplyMethod(reflect.TypeOf(httpClient), "Get", func(_ *http.Client, url string) (resp *http.Response, err error) {
				return &http.Response{Body: io.NopCloser(strings.NewReader("abc"))}, nil
			})
			defer patch1.Reset()

			kubeletClient := KubeletClient{}
			got, err := kubeletClient.GetNodeRunningPods()
			So(err, ShouldNotBeNil)
			So(got, ShouldBeNil)
		})
	})
}
