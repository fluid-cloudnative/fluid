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
)

// TestParseReportSummary tests the parseReportSummary method of AlluxioEngine.
// It verifies that the method correctly parses the Alluxio report summary string
// and extracts the cache capacity and cached size information.
//
// The test case includes:
// - A mock Alluxio report summary string (mockAlluxioReportSummaryForParseReport)
// - Expected cacheStates output with cacheCapacity and cached values
//
// The test compares the parsed output with expected values and reports any discrepancies.
// This ensures the parsing logic handles the Alluxio report summary format correctly.
func TestParseReportSummary(t *testing.T) {
	testCases := map[string]struct {
		summary string
		want    cacheStates
	}{
		"test parseReportSummary case 1": {
			summary: mockAlluxioReportSummaryForParseReport(),
			want: cacheStates{
				cacheCapacity: "19.07MiB",
				cached:        "9.69MiB",
			},
		},
	}

	for k, item := range testCases {
		got := AlluxioEngine{}.parseReportSummary(item.summary)
		if !reflect.DeepEqual(item.want, got) {
			t.Errorf("%s check failure,want:%+v,got:%+v", k, item.want, got)
		}
	}
}

func TestParseReportMetric(t *testing.T) {
	testCases := map[string]struct {
		metrics            string
		want               cacheHitStates
		lastCacheHitStates *cacheHitStates
	}{
		"test ParseReportMetric case 1": {
			metrics:            mockAlluxioReportMetricsForParseMetric(),
			lastCacheHitStates: nil,
			want: cacheHitStates{
				bytesReadLocal:  20310917,
				bytesReadUfsAll: 32243712,
			},
		},
		"test ParseReportMetric case 2": {
			metrics: mockAlluxioReportMetricsForParseMetric(),
			lastCacheHitStates: &cacheHitStates{
				bytesReadLocal:  10000,
				bytesReadUfsAll: 40000,
			},
			want: cacheHitStates{
				bytesReadLocal:        20310917,
				bytesReadUfsAll:       32243712,
				cacheHitRatio:         "38.7%",
				localHitRatio:         "38.7%",
				remoteHitRatio:        "0.0%",
				localThroughputRatio:  "38.7%",
				remoteThroughputRatio: "0.0%",
				cacheThroughputRatio:  "38.7%",
			},
		},
	}

	for k, item := range testCases {
		e := &AlluxioEngine{}
		if item.lastCacheHitStates != nil {
			e.lastCacheHitStates = item.lastCacheHitStates
			got := cacheHitStates{}
			e.ParseReportMetric(item.metrics, &got, item.lastCacheHitStates)

			// skip timestamp check
			item.want.timestamp = got.timestamp

			if !reflect.DeepEqual(item.want, got) {
				t.Errorf("%s check failure,\n want:%+v \n,got:%+v", k, item.want, got)
			}
		} else {
			got := cacheHitStates{}
			e.ParseReportMetric(item.metrics, &got, item.lastCacheHitStates)
			if item.want.bytesReadLocal != got.bytesReadLocal {
				t.Errorf("%s bytesReadLocal check failure,want:%+v,got:%+v", k, item.want.bytesReadLocal, got.bytesReadLocal)
			}
			if item.want.bytesReadUfsAll != got.bytesReadUfsAll {
				t.Errorf("%s bytesReadUfsAll check failure,want:%+v,got:%+v", k, item.want.bytesReadUfsAll, got.bytesReadUfsAll)
			}
		}

	}
}

func mockAlluxioReportMetricsForParseMetric() string {
	return `Cluster.BytesReadAlluxio  (Type: COUNTER, Value: 0B)
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
	Master.registerWorker.User:root  (Type: TIMER, Value: 1)`
}

func mockAlluxioReportSummaryForParseReport() string {
	summary := `Alluxio cluster summary: 
    Master Address: 172.18.0.2:20000
    Web Port: 20001
    Rpc Port: 20000
    Started: 07-02-2021 11:15:25:107
    Uptime: 0 day(s), 1 hour(s), 3 minute(s), and 35 second(s)
    Version: 2.3.1-SNAPSHOT
    Safe Mode: false
    Zookeeper Enabled: false
    Live Workers: 1
    Lost Workers: 0
    Total Capacity: 19.07MB
        Tier: MEM  Size: 19.07MB
    Used Capacity: 9.69MB
        Tier: MEM  Size: 9.69MB
    Free Capacity: 9.39MB
	`

	return summary
}
