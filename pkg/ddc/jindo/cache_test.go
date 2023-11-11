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

package jindo

import (
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	. "github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	. "github.com/smartystreets/goconvey/convey"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestQueryCacheStatus(t *testing.T) {
	Convey("test queryCacheStatus ", t, func() {
		Convey("with dataset UFSTotal is not empty ", func() {
			var engine *JindoEngine
			patch1 := ApplyMethod(reflect.TypeOf(engine), "GetReportSummary",
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
			var engine *JindoEngine
			patch1 := ApplyMethod(reflect.TypeOf(engine), "GetReportSummary",
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
			var engine *JindoEngine
			patch1 := ApplyMethod(reflect.TypeOf(engine), "GetReportSummary",
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

func TestInvokeCleanCache(t *testing.T) {
	masterInputs := []*appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop-jindofs-master",
				Namespace: "fluid",
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 0,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-jindofs-master",
				Namespace: "fluid",
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		},
	}
	objs := []runtime.Object{}
	for _, masterInput := range masterInputs {
		objs = append(objs, masterInput.DeepCopy())
	}
	fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)
	testCases := []struct {
		name      string
		namespace string
		isErr     bool
	}{
		{
			name:      "hadoop",
			namespace: "fluid",
			isErr:     false,
		},
		{
			name:      "hbase",
			namespace: "fluid",
			isErr:     true,
		},
		{
			name:      "none",
			namespace: "fluid",
			isErr:     false,
		},
	}
	for _, testCase := range testCases {
		engine := &JindoEngine{
			Client:    fakeClient,
			namespace: testCase.namespace,
			name:      testCase.name,
			Log:       fake.NullLogger(),
		}
		err := engine.invokeCleanCache()
		isErr := err != nil
		if isErr != testCase.isErr {
			t.Errorf("test-name:%s want %t, got %t", testCase.name, testCase.isErr, isErr)
		}
	}
}

// $ jindo jfs -report
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
