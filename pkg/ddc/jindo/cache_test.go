package jindo

import (
	"reflect"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	. "github.com/smartystreets/goconvey/convey"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestQueryCacheStatus(t *testing.T) {
	Convey("test queryCacheStatus ", t, func() {
		Convey("with dataset UFSTotal is not empty ", func() {
			var enging *JindoEngine
			patch1 := ApplyMethod(reflect.TypeOf(enging), "GetReportSummary",
				func(_ *JindoEngine) (string, error) {
					summary := mockJindoReportSummary()
					return summary, nil
				})
			defer patch1.Reset()

			patch2 := ApplyFunc(utils.GetDataset,
				func(_ client.Client, _ string, _ string) (*datav1alpha1.Dataset, error) {
					d := &datav1alpha1.Dataset{
						Status: datav1alpha1.DatasetStatus{
							UfsTotal: "52.18MiB",
						},
					}
					return d, nil
				})
			defer patch2.Reset()

			e := &JindoEngine{}
			got, err := e.queryCacheStatus()
			want := cacheStates{
				cacheCapacity:    "250.38GiB",
				cached:           "11.72GiB",
				cachedPercentage: "100.0%",
			}

			So(got, ShouldResemble, want)
			So(err, ShouldEqual, nil)
		})

		Convey("with dataset UFSTotal is: [Calculating]", func() {
			var enging *JindoEngine
			patch1 := ApplyMethod(reflect.TypeOf(enging), "GetReportSummary",
				func(_ *JindoEngine) (string, error) {
					summary := mockJindoReportSummary()
					return summary, nil
				})
			defer patch1.Reset()

			patch2 := ApplyFunc(utils.GetDataset,
				func(_ client.Client, _ string, _ string) (*datav1alpha1.Dataset, error) {
					d := &datav1alpha1.Dataset{
						Status: datav1alpha1.DatasetStatus{
							UfsTotal: "[Calculating]",
						},
					}
					return d, nil
				})
			defer patch2.Reset()

			e := &JindoEngine{}
			got, err := e.queryCacheStatus()
			want := cacheStates{
				cacheCapacity: "250.38GiB",
				cached:        "11.72GiB",
			}

			So(got, ShouldResemble, want)
			So(err, ShouldEqual, nil)
		})

		Convey("with dataset UFSTotal is empty", func() {
			var enging *JindoEngine
			patch1 := ApplyMethod(reflect.TypeOf(enging), "GetReportSummary",
				func(_ *JindoEngine) (string, error) {
					summary := mockJindoReportSummary()
					return summary, nil
				})
			defer patch1.Reset()

			patch2 := ApplyFunc(utils.GetDataset,
				func(_ client.Client, _ string, _ string) (*datav1alpha1.Dataset, error) {
					d := &datav1alpha1.Dataset{
						Status: datav1alpha1.DatasetStatus{
							UfsTotal: "",
						},
					}
					return d, nil
				})
			defer patch2.Reset()

			e := &JindoEngine{}
			got, err := e.queryCacheStatus()
			want := cacheStates{
				cacheCapacity: "250.38GiB",
				cached:        "11.72GiB",
			}

			So(got, ShouldResemble, want)
			So(err, ShouldEqual, nil)
		})
	})
}

//
// $ jindo jfs -report
//
func mockJindoReportSummary() string {
	s := `Namespace Address: localhost:18000
	Rpc Port: 8101
	Started: Mon Jul 19 07:41:39 2021
	Version: 3.6.1
	Live Nodes: 2
	Decommission Nodes: 0
	Mode: BLOCK
	Total Capacity: 250.38GB
	Used Capacity: 11.72GB
	`
	return s
}
