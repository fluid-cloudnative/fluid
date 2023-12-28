/*
Copyright 2023 The Fluid Authors.

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
package vineyard

import (
	"reflect"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestQueryCacheStatus(t *testing.T) {
	Convey("test queryCacheStatus ", t, func() {
		Convey("with total usage is 150Mi", func() {
			var engine *VineyardEngine
			quota := resource.MustParse("100Mi")
			patch := ApplyMethod(reflect.TypeOf(engine), "GetReportSummary",
				func(_ *VineyardEngine) ([]string, error) {
					return []string{"instances_memory_usage_bytes 104857600", "instances_memory_usage_bytes 52428800"}, nil
				})
			defer patch.Reset()

			e := &VineyardEngine{
				runtime: &v1alpha1.VineyardRuntime{
					Spec: v1alpha1.VineyardRuntimeSpec{
						TieredStore: v1alpha1.TieredStore{
							Levels: []v1alpha1.Level{
								{
									MediumType: "MEM",
									Quota:      &quota,
								},
							},
						},
					},
					Status: v1alpha1.RuntimeStatus{
						WorkerNumberReady: 3,
					},
				},
			}
			got, err := e.queryCacheStatus(e.runtime)
			want := cacheStates{
				cacheCapacity:    "300.00MiB",
				cached:           "150.00MiB",
				cachedPercentage: "50.0%",
			}

			So(got, ShouldResemble, want)
			So(err, ShouldEqual, nil)
		})
	})
}
