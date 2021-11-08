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

	. "github.com/agiledragon/gomonkey"
	. "github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
)

func TestJuiceFSEngine_queryCacheStatus(t *testing.T) {
	Convey("Test CleanupCache ", t, func() {
		Convey("cleanup success", func() {
			runtimeInfo, err := base.BuildRuntimeInfo("juicefs", "fluid", "juicefs", datav1alpha1.TieredStore{})
			if err != nil {
				t.Errorf("fail to create the runtimeInfo with error %v", err)
			}
			runtimeInfo.SetupFuseDeployMode(false, nil)
			var engine *JuiceFSEngine
			patch1 := ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfDaemonset",
				func(_ *JuiceFSEngine, dsName string, namespace string) ([]corev1.Pod, error) {
					r := mockRunningPodsOfDaemonSet()
					return r, nil
				})
			defer patch1.Reset()
			patch2 := ApplyMethod(reflect.TypeOf(engine), "GetPodMetrics",
				func(_ *JuiceFSEngine, podName string) (string, error) {
					return mockJuiceFSMetric(), nil
				})
			defer patch2.Reset()

			a := &JuiceFSEngine{
				name:        "test",
				namespace:   "default",
				runtimeType: "JuiceFSRuntime",
				Log:         nil,
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
				cached:               "387.17KiB",
				cachedPercentage:     "151.2%",
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
