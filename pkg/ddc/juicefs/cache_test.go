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

	"k8s.io/apimachinery/pkg/api/resource"

	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestJuiceFSEngine_queryCacheStatus(t *testing.T) {
	Convey("Test queryCacheStatus ", t, func() {
		Convey("queryCacheStatus success", func() {
			runtimeInfo, err := base.BuildRuntimeInfo("juicefs", "fluid", "juicefs", datav1alpha1.TieredStore{})
			if err != nil {
				t.Errorf("fail to create the runtimeInfo with error %v", err)
			}
			var engine *JuiceFSEngine
			patch1 := ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfDaemonset",
				func(_ *JuiceFSEngine, dsName string, namespace string) ([]corev1.Pod, error) {
					r := mockRunningPodsOfDaemonSet()
					return r, nil
				})
			defer patch1.Reset()
			patch2 := ApplyMethod(reflect.TypeOf(engine), "GetPodMetrics",
				func(_ *JuiceFSEngine, podName, containerName string) (string, error) {
					return mockJuiceFSMetric(), nil
				})
			defer patch2.Reset()
			patch3 := ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
				func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
					r := mockRunningPodsOfStatefulSet()
					return r, nil
				})
			defer patch3.Reset()
			patch4 := ApplyMethod(reflect.TypeOf(engine), "GetEdition",
				func(_ *JuiceFSEngine) string {
					return "enterprise"
				})
			defer patch4.Reset()

			a := &JuiceFSEngine{
				name:        "test",
				namespace:   "default",
				runtimeType: "JuiceFSRuntime",
				Log:         fake.NullLogger(),
				runtimeInfo: runtimeInfo,
				runtime: &datav1alpha1.JuiceFSRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						Replicas: 1,
						Worker: datav1alpha1.JuiceFSCompTemplateSpec{Options: map[string]string{
							"cache-size": "102400",
						}},
					},
					Status: datav1alpha1.RuntimeStatus{
						WorkerNumberReady: 1,
					},
				},
			}
			want := cacheStates{
				cacheCapacity:        "100.00GiB",
				cached:               "37.96GiB",
				cachedPercentage:     "0.0%",
				cacheHitRatio:        "100.0%",
				cacheThroughputRatio: "100.0%",
			}
			got, err := a.queryCacheStatus()
			if err != nil {
				t.Error("check failure, want err, got nil")
			}
			if want != got {
				t.Errorf("got=%v, want=%v", got, want)
			}
		})
		Convey("queryCacheStatus", func() {
			runtimeInfo, err := base.BuildRuntimeInfo("juicefs", "fluid", "juicefs", datav1alpha1.TieredStore{
				Levels: []datav1alpha1.Level{{
					MediumType: "MEM",
					Path:       "/data",
					Quota:      resource.NewQuantity(100, resource.BinarySI),
				}},
			})
			if err != nil {
				t.Errorf("fail to create the runtimeInfo with error %v", err)
			}
			var engine *JuiceFSEngine
			patch1 := ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfDaemonset",
				func(_ *JuiceFSEngine, dsName string, namespace string) ([]corev1.Pod, error) {
					r := mockRunningPodsOfDaemonSet()
					return r, nil
				})
			defer patch1.Reset()
			patch2 := ApplyMethod(reflect.TypeOf(engine), "GetPodMetrics",
				func(_ *JuiceFSEngine, podName, containerName string) (string, error) {
					return mockJuiceFSMetric(), nil
				})
			defer patch2.Reset()
			patch3 := ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
				func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
					r := mockRunningPodsOfStatefulSet()
					return r, nil
				})
			defer patch3.Reset()
			patch4 := ApplyMethod(reflect.TypeOf(engine), "GetEdition",
				func(_ *JuiceFSEngine) string {
					return "enterprise"
				})
			defer patch4.Reset()

			a := &JuiceFSEngine{
				name:        "test",
				namespace:   "default",
				runtimeType: "JuiceFSRuntime",
				Log:         fake.NullLogger(),
				runtimeInfo: runtimeInfo,
				runtime: &datav1alpha1.JuiceFSRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "fluid",
					},
				},
			}
			want := cacheStates{
				cacheCapacity:        "",
				cached:               "37.96GiB",
				cachedPercentage:     "0.0%",
				cacheHitRatio:        "100.0%",
				cacheThroughputRatio: "100.0%",
			}
			got, err := a.queryCacheStatus()
			if err != nil {
				t.Error("check failure, want err, got nil")
			}
			if want != got {
				t.Errorf("got=%v, want=%v", got, want)
			}
		})
	})
}
