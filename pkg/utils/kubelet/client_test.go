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
