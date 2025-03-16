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

package alluxio

import (
	"reflect"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/go-logr/logr"
	. "github.com/smartystreets/goconvey/convey"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/utils/ptr"
)

// TestQueryCacheStatus is a unit test for the queryCacheStatus function.
// It verifies the correctness of the function under different dataset conditions.
func TestQueryCacheStatus(t *testing.T) {
	Convey("test queryCacheStatus ", t, func() {
		Convey("with dataset UFSTotal is not empty ", func() {
			var engine *AlluxioEngine
			patch1 := ApplyMethod(reflect.TypeOf(engine), "GetReportSummary",
				func(_ *AlluxioEngine) (string, error) {
					summary := mockAlluxioReportSummary()
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

			patch3 := ApplyMethod(reflect.TypeOf(engine), "GetCacheHitStates",
				func(_ *AlluxioEngine) cacheHitStates {
					return cacheHitStates{
						bytesReadLocal:  20310917,
						bytesReadUfsAll: 32243712,
					}
				})
			defer patch3.Reset()

			e := &AlluxioEngine{}
			got, err := e.queryCacheStatus()
			want := cacheStates{
				cacheCapacity:    "19.07MiB",
				cached:           "0.00B",
				cachedPercentage: "0.0%",
				cacheHitStates: cacheHitStates{
					bytesReadLocal:  20310917,
					bytesReadUfsAll: 32243712,
				},
			}

			So(got, ShouldResemble, want)
			So(err, ShouldEqual, nil)
		})

		Convey("with dataset UFSTotal is: [Calculating]", func() {
			var engine *AlluxioEngine
			patch1 := ApplyMethod(reflect.TypeOf(engine), "GetReportSummary",
				func(_ *AlluxioEngine) (string, error) {
					summary := mockAlluxioReportSummary()
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

			patch3 := ApplyMethod(reflect.TypeOf(engine), "GetCacheHitStates",
				func(_ *AlluxioEngine) cacheHitStates {
					return cacheHitStates{}
				})
			defer patch3.Reset()

			e := &AlluxioEngine{}
			got, err := e.queryCacheStatus()
			want := cacheStates{
				cacheCapacity: "19.07MiB",
				cached:        "0.00B",
			}

			So(got, ShouldResemble, want)
			So(err, ShouldEqual, nil)
		})

		Convey("with dataset UFSTotal is empty", func() {
			var engine *AlluxioEngine
			patch1 := ApplyMethod(reflect.TypeOf(engine), "GetReportSummary",
				func(_ *AlluxioEngine) (string, error) {
					summary := mockAlluxioReportSummary()
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

			patch3 := ApplyMethod(reflect.TypeOf(engine), "GetCacheHitStates",
				func(_ *AlluxioEngine) cacheHitStates {
					return cacheHitStates{}
				})
			defer patch3.Reset()

			e := &AlluxioEngine{}
			got, err := e.queryCacheStatus()
			want := cacheStates{
				cacheCapacity: "19.07MiB",
				cached:        "0.00B",
			}

			So(got, ShouldResemble, want)
			So(err, ShouldEqual, nil)
		})
	})
}

// TestGetCacheHitStates verifies that the GetCacheHitStates method of the AlluxioEngine
// correctly extracts and maps the cache hit metrics from the engine's report.
// It overrides the GetReportMetrics method to return a controlled report, then asserts that
// the returned cache hit state contains the expected values for both local cache hits (bytesReadLocal)
// and unified file system hits (bytesReadUfsAll), ensuring correct metric parsing.
//
// Parameters:
//   - t (*testing.T): The testing context used for running and reporting the test.
//
// Returns:
//   - None: This test function does not return a value but uses assertions to verify correctness.
func TestGetCacheHitStates(t *testing.T) {
	Convey("Test GetCacheHitStates ", t, func() {
		Convey("with data ", func() {
			var engine *AlluxioEngine
			patch1 := ApplyMethod(reflect.TypeOf(engine), "GetReportMetrics",
				func(_ *AlluxioEngine) (string, error) {
					r := mockAlluxioReportMetrics()
					return r, nil
				})
			defer patch1.Reset()

			e := &AlluxioEngine{}

			got := e.GetCacheHitStates()
			want := cacheHitStates{
				bytesReadLocal:  20310917,
				bytesReadUfsAll: 32243712,
			}
			So(got.bytesReadLocal, ShouldEqual, want.bytesReadLocal)
			So(got.bytesReadUfsAll, ShouldEqual, want.bytesReadUfsAll)
		})

	})
}

func TestPatchDatasetStatus(t *testing.T) {
	engine := &AlluxioEngine{}
	testCases := []struct {
		total      string
		cached     string
		percentage string
	}{
		{
			total:      "100",
			cached:     "10",
			percentage: "10.0%",
		},
		{
			total:      "100",
			cached:     "50",
			percentage: "50.0%",
		},
	}
	for _, testCase := range testCases {
		dataset := &datav1alpha1.Dataset{
			Status: datav1alpha1.DatasetStatus{
				UfsTotal: testCase.total,
			},
		}
		states := &cacheStates{
			cached: testCase.cached,
		}
		engine.patchDatasetStatus(dataset, states)
		if states.cachedPercentage != testCase.percentage {
			t.Errorf(" want %s, got %s", testCase.percentage, states.cachedPercentage)
		}
	}
}

func TestInvokeCleanCache(t *testing.T) {
	masterInputs := []*appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop-master",
				Namespace: "fluid",
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 0,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-master",
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
		engine := &AlluxioEngine{
			Client:    fakeClient,
			namespace: testCase.namespace,
			name:      testCase.name,
			Log:       fake.NullLogger(),
		}
		err := engine.invokeCleanCache("")
		isErr := err != nil
		if isErr != testCase.isErr {
			t.Errorf("test-name:%s want %t, got %t", testCase.name, testCase.isErr, isErr)
		}
	}
}

// $ alluxio fsadmin report summary
func mockAlluxioReportSummary() string {
	s := `Alluxio cluster summary: 
	Master Address: 172.18.0.2:20000
	Web Port: 20001
	Rpc Port: 20000
	Started: 06-29-2021 13:43:56:297
	Uptime: 0 day(s), 0 hour(s), 4 minute(s), and 13 second(s)
	Version: 2.3.1-SNAPSHOT
	Safe Mode: false
	Zookeeper Enabled: false
	Live Workers: 1
	Lost Workers: 0
	Total Capacity: 19.07MB
		Tier: MEM  Size: 19.07MB
	Used Capacity: 0B
		Tier: MEM  Size: 0B
	Free Capacity: 19.07MB
	`
	return s
}

func mockAlluxioReportMetrics() string {
	r := `Cluster.BytesReadAlluxio  (Type: COUNTER, Value: 0B)
	Cluster.BytesReadAlluxioThroughput  (Type: GAUGE, Value: 0B/MIN)
	Cluster.BytesReadDomain  (Type: COUNTER, Value: 0B)
	Cluster.BytesReadDomainThroughput  (Type: GAUGE, Value: 0B/MIN)
	Cluster.BytesReadLocal  (Type: COUNTER, Value: 19.37MB)
	Cluster.BytesReadLocalThroughput  (Type: GAUGE, Value: 495.97KB/MIN)
	Cluster.BytesReadPerUfs.UFS:s3:%2F%2Ffluid  (Type: COUNTER, Value: 30.75MB)
	Cluster.BytesReadUfsAll  (Type: COUNTER, Value: 30.75MB)
	Cluster.BytesReadUfsThroughput  (Type: GAUGE, Value: 787.17KB/MIN)
	Cluster.BytesWrittenAlluxio  (Type: COUNTER, Value: 0B)
	Cluster.BytesWrittenAlluxioThroughput  (Type: GAUGE, Value: 0B/MIN)
	Cluster.BytesWrittenDomain  (Type: COUNTER, Value: 0B)
	Cluster.BytesWrittenDomainThroughput  (Type: GAUGE, Value: 0B/MIN)
	Cluster.BytesWrittenLocal  (Type: COUNTER, Value: 0B)
	Cluster.BytesWrittenLocalThroughput  (Type: GAUGE, Value: 0B/MIN)
	Cluster.BytesWrittenUfsAll  (Type: COUNTER, Value: 0B)
	Cluster.BytesWrittenUfsThroughput  (Type: GAUGE, Value: 0B/MIN)
	Cluster.CapacityFree  (Type: GAUGE, Value: 9,842,601)
	Cluster.CapacityFreeTierHDD  (Type: GAUGE, Value: 0)
	Cluster.CapacityFreeTierMEM  (Type: GAUGE, Value: 9,842,601)
	Cluster.CapacityFreeTierSSD  (Type: GAUGE, Value: 0)
	Cluster.CapacityTotal  (Type: GAUGE, Value: 20,000,000)
	Cluster.CapacityTotalTierHDD  (Type: GAUGE, Value: 0)
	Cluster.CapacityTotalTierMEM  (Type: GAUGE, Value: 20,000,000)
	Cluster.CapacityTotalTierSSD  (Type: GAUGE, Value: 0)
	Cluster.CapacityUsed  (Type: GAUGE, Value: 10,157,399)
	Cluster.CapacityUsedTierHDD  (Type: GAUGE, Value: 0)
	Cluster.CapacityUsedTierMEM  (Type: GAUGE, Value: 10,157,399)
	Cluster.CapacityUsedTierSSD  (Type: GAUGE, Value: 0)
	Cluster.RootUfsCapacityFree  (Type: GAUGE, Value: -1)
	Cluster.RootUfsCapacityTotal  (Type: GAUGE, Value: -1)
	Cluster.RootUfsCapacityUsed  (Type: GAUGE, Value: -1)
	Cluster.Workers  (Type: GAUGE, Value: 1)
	Master.CompleteFileOps  (Type: COUNTER, Value: 0)
	Master.ConnectFromMaster.UFS:s3:%2F%2Ffluid.UFS_TYPE:s3  (Type: TIMER, Value: 0)
	Master.Create.UFS:%2Fjournal%2FBlockMaster.UFS_TYPE:local  (Type: TIMER, Value: 1)
	Master.Create.UFS:%2Fjournal%2FFileSystemMaster.UFS_TYPE:local  (Type: TIMER, Value: 1)
	Master.Create.UFS:%2Fjournal%2FMetaMaster.UFS_TYPE:local  (Type: TIMER, Value: 1)
	Master.CreateDirectoryOps  (Type: COUNTER, Value: 0)
	Master.CreateFileOps  (Type: COUNTER, Value: 0)
	Master.DeletePathOps  (Type: COUNTER, Value: 0)
	Master.DirectoriesCreated  (Type: COUNTER, Value: 0)
	Master.EdgeCacheSize  (Type: GAUGE, Value: 7)
	Master.Exists.UFS:s3:%2F%2Ffluid.UFS_TYPE:s3  (Type: TIMER, Value: 2)
	Master.FileBlockInfosGot  (Type: COUNTER, Value: 0)
	Master.FileInfosGot  (Type: COUNTER, Value: 25)
	Master.FilesCompleted  (Type: COUNTER, Value: 7)
	Master.FilesCreated  (Type: COUNTER, Value: 7)
	Master.FilesFreed  (Type: COUNTER, Value: 0)
	Master.FilesPersisted  (Type: COUNTER, Value: 0)
	Master.FilesPinned  (Type: GAUGE, Value: 0)
	Master.FreeFileOps  (Type: COUNTER, Value: 0)
	Master.GetAcl.UFS:s3:%2F%2Ffluid.UFS_TYPE:s3  (Type: TIMER, Value: 7)
	Master.GetBlockInfo.User:root  (Type: TIMER, Value: 3)
	Master.GetBlockMasterInfo.User:root  (Type: TIMER, Value: 173)
	Master.GetConfigHash.User:root  (Type: TIMER, Value: 40)
	Master.GetFileBlockInfoOps  (Type: COUNTER, Value: 0)
	Master.GetFileInfoOps  (Type: COUNTER, Value: 9)
	Master.GetFileLocations.User:root.UFS:s3:%2F%2Ffluid.UFS_TYPE:s3  (Type: TIMER, Value: 24)
	Master.GetFingerprint.User:root.UFS:s3:%2F%2Ffluid.UFS_TYPE:s3  (Type: TIMER, Value: 1)
	Master.GetMountTable.User:root  (Type: TIMER, Value: 2)
	Master.GetNewBlockOps  (Type: COUNTER, Value: 0)
	Master.GetSpace.UFS:s3:%2F%2Ffluid.UFS_TYPE:s3  (Type: TIMER, Value: 18)
	Master.GetSpace.User:root.UFS:s3:%2F%2Ffluid.UFS_TYPE:s3  (Type: TIMER, Value: 103)
	Master.GetStatus.User:root  (Type: TIMER, Value: 6)
	Master.GetStatus.User:root.UFS:s3:%2F%2Ffluid.UFS_TYPE:s3  (Type: TIMER, Value: 3)
	Master.GetStatusFailures.User:root.UFS:s3:%2F%2Ffluid.UFS_TYPE:s3  (Type: COUNTER, Value: 2)
	Master.GetWorkerInfoList.User:root  (Type: TIMER, Value: 2)
	Master.InodeCacheSize  (Type: GAUGE, Value: 8)
	Master.JournalFlushTimer  (Type: TIMER, Value: 22)
	Master.LastBackupEntriesCount  (Type: GAUGE, Value: -1)
	Master.LastBackupRestoreCount  (Type: GAUGE, Value: -1)
	Master.LastBackupRestoreTimeMs  (Type: GAUGE, Value: -1)
	Master.LastBackupTimeMs  (Type: GAUGE, Value: -1)
	Master.ListStatus.UFS:%2Fjournal%2FBlockMaster.UFS_TYPE:local  (Type: TIMER, Value: 63)
	Master.ListStatus.UFS:%2Fjournal%2FFileSystemMaster.UFS_TYPE:local  (Type: TIMER, Value: 63)
	Master.ListStatus.UFS:%2Fjournal%2FMetaMaster.UFS_TYPE:local  (Type: TIMER, Value: 63)
	Master.ListStatus.UFS:%2Fjournal%2FMetricsMaster.UFS_TYPE:local  (Type: TIMER, Value: 63)
	Master.ListStatus.UFS:%2Fjournal%2FTableMaster.UFS_TYPE:local  (Type: TIMER, Value: 63)
	Master.ListStatus.User:root  (Type: TIMER, Value: 3)
	Master.ListStatus.User:root.UFS:s3:%2F%2Ffluid.UFS_TYPE:s3  (Type: TIMER, Value: 1)
	Master.ListingCacheSize  (Type: GAUGE, Value: 8)
	Master.MountOps  (Type: COUNTER, Value: 0)
	Master.NewBlocksGot  (Type: COUNTER, Value: 0)
	Master.PathsDeleted  (Type: COUNTER, Value: 0)
	Master.PathsMounted  (Type: COUNTER, Value: 0)
	Master.PathsRenamed  (Type: COUNTER, Value: 0)
	Master.PathsUnmounted  (Type: COUNTER, Value: 0)
	Master.PerUfsOpConnectFromMaster.UFS:s3:%2F%2Ffluid  (Type: GAUGE, Value: 0)
	Master.PerUfsOpCreate.UFS:%2Fjournal%2FBlockMaster  (Type: GAUGE, Value: 1)
	Master.PerUfsOpCreate.UFS:%2Fjournal%2FFileSystemMaster  (Type: GAUGE, Value: 1)
	Master.PerUfsOpCreate.UFS:%2Fjournal%2FMetaMaster  (Type: GAUGE, Value: 1)
	Master.PerUfsOpExists.UFS:s3:%2F%2Ffluid  (Type: GAUGE, Value: 2)
	Master.PerUfsOpGetFileLocations.UFS:s3:%2F%2Ffluid  (Type: GAUGE, Value: 24)
	Master.PerUfsOpGetFingerprint.UFS:s3:%2F%2Ffluid  (Type: GAUGE, Value: 1)
	Master.PerUfsOpGetSpace.UFS:s3:%2F%2Ffluid  (Type: GAUGE, Value: 116)
	Master.PerUfsOpGetStatus.UFS:s3:%2F%2Ffluid  (Type: GAUGE, Value: 3)
	Master.PerUfsOpListStatus.UFS:%2Fjournal%2FBlockMaster  (Type: GAUGE, Value: 60)
	Master.PerUfsOpListStatus.UFS:%2Fjournal%2FFileSystemMaster  (Type: GAUGE, Value: 60)
	Master.PerUfsOpListStatus.UFS:%2Fjournal%2FMetaMaster  (Type: GAUGE, Value: 60)
	Master.PerUfsOpListStatus.UFS:%2Fjournal%2FMetricsMaster  (Type: GAUGE, Value: 60)
	Master.PerUfsOpListStatus.UFS:%2Fjournal%2FTableMaster  (Type: GAUGE, Value: 60)
	Master.PerUfsOpListStatus.UFS:s3:%2F%2Ffluid  (Type: GAUGE, Value: 1)
	Master.RenamePathOps  (Type: COUNTER, Value: 0)
	Master.SetAclOps  (Type: COUNTER, Value: 0)
	Master.SetAttributeOps  (Type: COUNTER, Value: 0)
	Master.TotalPaths  (Type: GAUGE, Value: 8)
	Master.UfsSessionCount-Ufs:s3:%2F%2Ffluid  (Type: COUNTER, Value: 0)
	Master.UnmountOps  (Type: COUNTER, Value: 0)
	Master.blockHeartbeat.User:root  (Type: TIMER, Value: 2,410)
	Master.commitBlock.User:root  (Type: TIMER, Value: 1)
	Master.getConfigHash  (Type: TIMER, Value: 4)
	Master.getConfigHash.User:root  (Type: TIMER, Value: 239)
	Master.getConfiguration  (Type: TIMER, Value: 20)
	Master.getConfiguration.User:root  (Type: TIMER, Value: 428)
	Master.getMasterInfo.User:root  (Type: TIMER, Value: 173)
	Master.getMetrics.User:root  (Type: TIMER, Value: 33)
	Master.getPinnedFileIds.User:root  (Type: TIMER, Value: 2,410)
	Master.getUfsInfo.User:root  (Type: TIMER, Value: 1)
	Master.getWorkerId.User:root  (Type: TIMER, Value: 1)
	Master.metricsHeartbeat.User:root  (Type: TIMER, Value: 4)
	Master.registerWorker.User:root  (Type: TIMER, Value: 1)
	`
	return r
}

func TestAlluxioEngine_getGracefulShutdownLimits(t *testing.T) {
	type fields struct {
		runtime            *datav1alpha1.AlluxioRuntime
		name               string
		namespace          string
		runtimeType        string
		Log                logr.Logger
		Client             client.Client
		retryShutdown      int32
		initImage          string
		MetadataSyncDoneCh chan base.MetadataSyncResult
		UnitTest           bool
		Recorder           record.EventRecorder
	}
	tests := []struct {
		name                       string
		fields                     fields
		wantGracefulShutdownLimits int32
		wantErr                    bool
	}{
		// TODO: Add test cases.
		{
			name: "no_clean_cache_policy",
			fields: fields{
				name:      "noCleanCache",
				namespace: "default",
				runtime: &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "noCleanCache",
						Namespace: "default",
					},
					Spec: datav1alpha1.AlluxioRuntimeSpec{},
				},
			},
			wantGracefulShutdownLimits: 3,
			wantErr:                    false,
		}, {
			name: "clean_cache_policy",
			fields: fields{
				name:      "cleanCache",
				namespace: "default",
				runtime: &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cleanCache",
						Namespace: "default",
					},
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						RuntimeManagement: datav1alpha1.RuntimeManagement{
							CleanCachePolicy: datav1alpha1.CleanCachePolicy{
								MaxRetryAttempts: ptr.To[int32](12),
							},
						},
					},
				},
			},
			wantGracefulShutdownLimits: 12,
			wantErr:                    false,
		}, {
			name: "test_err",
			fields: fields{
				name:      "err",
				namespace: "default",
				runtime: &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "notFoundErr",
						Namespace: "default",
					},
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						RuntimeManagement: datav1alpha1.RuntimeManagement{
							CleanCachePolicy: datav1alpha1.CleanCachePolicy{
								MaxRetryAttempts: ptr.To[int32](12),
							},
						},
					},
				},
			},
			wantGracefulShutdownLimits: 0,
			wantErr:                    true,
		},
	}
	objs := []runtime.Object{}
	for _, tt := range tests {
		objs = append(objs, tt.fields.runtime.DeepCopy())
	}
	fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AlluxioEngine{
				runtime:            tt.fields.runtime,
				name:               tt.fields.name,
				namespace:          tt.fields.namespace,
				runtimeType:        tt.fields.runtimeType,
				Log:                tt.fields.Log,
				Client:             fakeClient,
				retryShutdown:      tt.fields.retryShutdown,
				initImage:          tt.fields.initImage,
				MetadataSyncDoneCh: tt.fields.MetadataSyncDoneCh,
				Recorder:           tt.fields.Recorder,
			}
			gotGracefulShutdownLimits, err := e.getGracefulShutdownLimits()
			if (err != nil) != tt.wantErr {
				t.Errorf("AlluxioEngine.getGracefulShutdownLimits() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotGracefulShutdownLimits != tt.wantGracefulShutdownLimits {
				t.Errorf("AlluxioEngine.getGracefulShutdownLimits() = %v, want %v", gotGracefulShutdownLimits, tt.wantGracefulShutdownLimits)
			}
		})
	}
}

func TestAlluxioEngine_getCleanCacheGracePeriodSeconds(t *testing.T) {
	type fields struct {
		runtime       *datav1alpha1.AlluxioRuntime
		name          string
		namespace     string
		runtimeType   string
		Log           logr.Logger
		Client        client.Client
		retryShutdown int32
		initImage     string
	}
	tests := []struct {
		name                             string
		fields                           fields
		wantCleanCacheGracePeriodSeconds int32
		wantErr                          bool
	}{
		// TODO: Add test cases.
		{
			name: "no_clean_cache_policy",
			fields: fields{
				name:      "noCleanCache",
				namespace: "default",
				runtime: &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "noCleanCache",
						Namespace: "default",
					},
					Spec: datav1alpha1.AlluxioRuntimeSpec{},
				},
			},
			wantCleanCacheGracePeriodSeconds: 60,
			wantErr:                          false,
		}, {
			name: "clean_cache_policy",
			fields: fields{
				name:      "cleanCache",
				namespace: "default",
				runtime: &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cleanCache",
						Namespace: "default",
					},
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						RuntimeManagement: datav1alpha1.RuntimeManagement{
							CleanCachePolicy: datav1alpha1.CleanCachePolicy{
								GracePeriodSeconds: ptr.To[int32](12),
							},
						},
					},
				},
			},
			wantCleanCacheGracePeriodSeconds: 12,
			wantErr:                          false,
		}, {
			name: "test_err",
			fields: fields{
				name:      "notFoundError",
				namespace: "default",
				runtime: &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "notFound",
						Namespace: "default",
					},
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						RuntimeManagement: datav1alpha1.RuntimeManagement{
							CleanCachePolicy: datav1alpha1.CleanCachePolicy{
								MaxRetryAttempts: ptr.To[int32](12),
							},
						},
					},
				},
			},
			wantCleanCacheGracePeriodSeconds: 0,
			wantErr:                          true,
		},
	}

	objs := []runtime.Object{}
	for _, tt := range tests {
		objs = append(objs, tt.fields.runtime.DeepCopy())
	}
	fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AlluxioEngine{
				runtime:       tt.fields.runtime,
				name:          tt.fields.name,
				namespace:     tt.fields.namespace,
				runtimeType:   tt.fields.runtimeType,
				Log:           tt.fields.Log,
				Client:        fakeClient,
				retryShutdown: tt.fields.retryShutdown,
				initImage:     tt.fields.initImage,
			}
			gotCleanCacheGracePeriodSeconds, err := e.getCleanCacheGracePeriodSeconds()
			if (err != nil) != tt.wantErr {
				t.Errorf("testcase %v AlluxioEngine.getCleanCacheGracePeriodSeconds() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if gotCleanCacheGracePeriodSeconds != tt.wantCleanCacheGracePeriodSeconds {
				t.Errorf("testcase %v AlluxioEngine.getCleanCacheGracePeriodSeconds() = %v, want %v", tt.name, gotCleanCacheGracePeriodSeconds, tt.wantCleanCacheGracePeriodSeconds)
			}
		})
	}
}
