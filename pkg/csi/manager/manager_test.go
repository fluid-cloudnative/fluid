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

package manager

import (
	"errors"
	. "github.com/agiledragon/gomonkey"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubelet"
	. "github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

func TestManager_run(t *testing.T) {
	Convey("TestManager_run", t, func() {
		Convey("run success", func() {
			kubeclient := &kubelet.KubeletClient{}
			patch1 := ApplyMethod(reflect.TypeOf(kubeclient), "GetNodeRunningPods", func(_ *kubelet.KubeletClient) (*v1.PodList, error) {
				return &v1.PodList{}, nil
			})
			defer patch1.Reset()

			driver := NewPodDriver(fake.NewFakeClient())
			manager := Manager{
				KubeletClient: &kubelet.KubeletClient{},
				Driver:        driver,
			}
			manager.run()
		})
		Convey("get pods error", func() {
			kubeclient := &kubelet.KubeletClient{}
			patch1 := ApplyMethod(reflect.TypeOf(kubeclient), "GetNodeRunningPods", func(_ *kubelet.KubeletClient) (*v1.PodList, error) {
				return &v1.PodList{}, errors.New("test")
			})
			defer patch1.Reset()

			driver := NewPodDriver(fake.NewFakeClient())
			manager := Manager{
				KubeletClient: &kubelet.KubeletClient{},
				Driver:        driver,
			}
			manager.run()
		})
		Convey("run pods", func() {
			kubeclient := &kubelet.KubeletClient{}
			patch1 := ApplyMethod(reflect.TypeOf(kubeclient), "GetNodeRunningPods", func(_ *kubelet.KubeletClient) (*v1.PodList, error) {
				return &v1.PodList{
					Items: []v1.Pod{{
						ObjectMeta: metav1.ObjectMeta{Name: "test-fuse-test"},
						Status:     v1.PodStatus{},
					}},
				}, nil
			})
			defer patch1.Reset()

			driver := NewPodDriver(fake.NewFakeClient())
			manager := Manager{
				KubeletClient: &kubelet.KubeletClient{},
				Driver:        driver,
			}
			manager.run()
		})
	})
}
