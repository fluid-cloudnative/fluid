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

package jindocache

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	. "github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	. "github.com/smartystreets/goconvey/convey"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestQueryCacheStatus(t *testing.T) {
	Convey("test queryCacheStatus ", t, func() {
		Convey("with dataset UFSTotal is not empty ", func() {
			var engine *JindoCacheEngine
			patch1 := ApplyMethod(reflect.TypeOf(engine), "GetReportSummary",
				func(_ *JindoCacheEngine) (string, error) {
					summary := mockJindoReportSummary()
					return summary, nil
				})
			defer patch1.Reset()

			patch2 := ApplyFunc(utils.GetDataset,
				func(_ client.Reader, _ string, _ string) (*datav1alpha1.Dataset, error) {
					d := &datav1alpha1.Dataset{
						Status: datav1alpha1.DatasetStatus{
							UfsTotal: "52.18MiB",
						},
					}
					return d, nil
				})
			defer patch2.Reset()

			e := &JindoCacheEngine{
				runtime: &datav1alpha1.JindoRuntime{Spec: datav1alpha1.JindoRuntimeSpec{
					TieredStore: datav1alpha1.TieredStore{
						Levels: []datav1alpha1.Level{
							{
								Path:       "/mnt/jindo0",
								MediumType: common.HDD,
							},
						},
					}},
				},
			}
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
			var engine *JindoCacheEngine
			patch1 := ApplyMethod(reflect.TypeOf(engine), "GetReportSummary",
				func(_ *JindoCacheEngine) (string, error) {
					summary := mockJindoReportSummary()
					return summary, nil
				})
			defer patch1.Reset()

			patch2 := ApplyFunc(utils.GetDataset,
				func(_ client.Reader, _ string, _ string) (*datav1alpha1.Dataset, error) {
					d := &datav1alpha1.Dataset{
						Status: datav1alpha1.DatasetStatus{
							UfsTotal: "[Calculating]",
						},
					}
					return d, nil
				})
			defer patch2.Reset()

			e := &JindoCacheEngine{
				runtime: &datav1alpha1.JindoRuntime{Spec: datav1alpha1.JindoRuntimeSpec{
					TieredStore: datav1alpha1.TieredStore{
						Levels: []datav1alpha1.Level{
							{
								Path:       "/mnt/jindo0",
								MediumType: common.HDD,
							},
						},
					}},
				},
			}
			got, err := e.queryCacheStatus()
			want := cacheStates{
				cacheCapacity: "250.38GiB",
				cached:        "11.72GiB",
			}

			So(got, ShouldResemble, want)
			So(err, ShouldEqual, nil)
		})

		Convey("with dataset UFSTotal is empty", func() {
			var engine *JindoCacheEngine
			patch1 := ApplyMethod(reflect.TypeOf(engine), "GetReportSummary",
				func(_ *JindoCacheEngine) (string, error) {
					summary := mockJindoReportSummary()
					return summary, nil
				})
			defer patch1.Reset()

			patch2 := ApplyFunc(utils.GetDataset,
				func(_ client.Reader, _ string, _ string) (*datav1alpha1.Dataset, error) {
					d := &datav1alpha1.Dataset{
						Status: datav1alpha1.DatasetStatus{
							UfsTotal: "",
						},
					}
					return d, nil
				})
			defer patch2.Reset()

			e := &JindoCacheEngine{
				runtime: &datav1alpha1.JindoRuntime{Spec: datav1alpha1.JindoRuntimeSpec{
					TieredStore: datav1alpha1.TieredStore{
						Levels: []datav1alpha1.Level{
							{
								Path:       "/mnt/jindo0",
								MediumType: common.HDD,
							},
						},
					}},
				},
			}
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
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-no-pod-jindofs-master",
				Namespace: "fluid",
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-exec-error-jindofs-master",
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
		name        string
		namespace   string
		patchExecFn func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (stdout, stderr string, err error)
		isErr       bool
	}{
		{
			name:      "hadoop",
			namespace: "fluid",
			patchExecFn: func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (stdout, stderr string, err error) {
				return "", "", nil
			},
			isErr: false,
		},
		{
			name:      "hbase",
			namespace: "fluid",
			patchExecFn: func(ctx context.Context, podName, containerName, namespace string, cmd []string) (stdout string, stderr string, err error) {
				return "cache cleaned up", "", nil
			},
			isErr: false,
		},
		{
			name:      "hbase-no-pod",
			namespace: "fluid",
			patchExecFn: func(ctx context.Context, podName, containerName, namespace string, cmd []string) (stdout string, stderr string, err error) {
				return "", "", errors.NewNotFound(schema.GroupResource{Group: "v1", Resource: "pods"}, "pod not found")
			},
			isErr: false,
		},
		{
			name:      "hbase-exec-error",
			namespace: "fluid",
			patchExecFn: func(ctx context.Context, podName, containerName, namespace string, cmd []string) (stdout string, stderr string, err error) {
				return "", "", fmt.Errorf("expected exec error")
			},
			isErr: true,
		},
		{
			name:      "none",
			namespace: "fluid",
			isErr:     false,
		},
	}
	for _, testCase := range testCases {
		engine := &JindoCacheEngine{
			Client:    fakeClient,
			namespace: testCase.namespace,
			name:      testCase.name,
			Log:       fake.NullLogger(),
		}

		patch := ApplyFunc(kubeclient.ExecCommandInContainerWithFullOutput, testCase.patchExecFn)

		err := engine.invokeCleanCache()
		isErr := err != nil
		if isErr != testCase.isErr {
			t.Errorf("test-name:%s want %t, got %t", testCase.name, testCase.isErr, isErr)
		}

		patch.Reset()
	}
}

// $ jindo fs -report
func mockJindoReportSummary() string {
	s := `Namespace Address: localhost:18000
	Rpc Port: 8101
	Started: Mon Jul 19 07:41:39 2021
	Version: 3.6.1
	Live Nodes: 2
	Decommission Nodes: 0
	Mode: BLOCK
	Total Disk Capacity: 250.38GB
	Used Disk Capacity: 11.72GB
    Total MEM Capacity: 250.38GB
	Used MEM Capacity: 11.72GB
	`
	return s
}
