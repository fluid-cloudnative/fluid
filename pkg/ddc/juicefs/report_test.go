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
	"github.com/brahma-adshonor/gohook"
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/juicefs/operations"
)

func mockJuiceFSMetric() string {
	return `# HELP juicefs_blockcache_blocks number of cached blocks
# TYPE juicefs_blockcache_blocks gauge
juicefs_blockcache_blocks{mp="/jfs/pvc-60cd26e1-e2f4-45dd-948b-6aafc7f7e8ce",vol_name="minio"} 51
# HELP juicefs_blockcache_bytes number of cached bytes
# TYPE juicefs_blockcache_bytes gauge
juicefs_blockcache_bytes{mp="/jfs/pvc-60cd26e1-e2f4-45dd-948b-6aafc7f7e8ce",vol_name="minio"} 396462
# HELP juicefs_blockcache_drops dropped block
# TYPE juicefs_blockcache_drops counter
juicefs_blockcache_drops{mp="/jfs/pvc-60cd26e1-e2f4-45dd-948b-6aafc7f7e8ce",vol_name="minio"} 0
# HELP juicefs_blockcache_evicts evicted cache blocks
# TYPE juicefs_blockcache_evicts counter
juicefs_blockcache_evicts{mp="/jfs/pvc-60cd26e1-e2f4-45dd-948b-6aafc7f7e8ce",vol_name="minio"} 0
# HELP juicefs_blockcache_hit_bytes read bytes from cached block
# TYPE juicefs_blockcache_hit_bytes counter
juicefs_blockcache_hit_bytes{mp="/jfs/pvc-60cd26e1-e2f4-45dd-948b-6aafc7f7e8ce",vol_name="minio"} 3.288597e+06
# HELP juicefs_blockcache_hits read from cached block
# TYPE juicefs_blockcache_hits counter
juicefs_blockcache_hits{mp="/jfs/pvc-60cd26e1-e2f4-45dd-948b-6aafc7f7e8ce",vol_name="minio"} 3861
# HELP juicefs_blockcache_miss missed read from cached block
# TYPE juicefs_blockcache_miss counter
juicefs_blockcache_miss{mp="/jfs/pvc-60cd26e1-e2f4-45dd-948b-6aafc7f7e8ce",vol_name="minio"} 0
# HELP juicefs_blockcache_miss_bytes missed bytes from cached block
# TYPE juicefs_blockcache_miss_bytes counter
juicefs_blockcache_miss_bytes{mp="/jfs/pvc-60cd26e1-e2f4-45dd-948b-6aafc7f7e8ce",vol_name="minio"} 0
# HELP juicefs_blockcache_write_bytes write bytes of cached block
# TYPE juicefs_blockcache_write_bytes counter
juicefs_blockcache_write_bytes{mp="/jfs/pvc-60cd26e1-e2f4-45dd-948b-6aafc7f7e8ce",vol_name="minio"} 3.454749e+06
# HELP juicefs_blockcache_writes written cached block
# TYPE juicefs_blockcache_writes counter
juicefs_blockcache_writes{mp="/jfs/pvc-60cd26e1-e2f4-45dd-948b-6aafc7f7e8ce",vol_name="minio"} 3903
# HELP juicefs_cpu_usage Accumulated CPU usage in seconds.
# TYPE juicefs_cpu_usage gauge
juicefs_cpu_usage{mp="/jfs/pvc-60cd26e1-e2f4-45dd-948b-6aafc7f7e8ce",vol_name="minio"} 77.154679
# HELP juicefs_fuse_open_handlers number of open files and directories.
# TYPE juicefs_fuse_open_handlers gauge
juicefs_fuse_open_handlers{mp="/jfs/pvc-60cd26e1-e2f4-45dd-948b-6aafc7f7e8ce",vol_name="minio"} 0
# HELP juicefs_go_build_info Build information about the main Go module.
# TYPE juicefs_go_build_info gauge
juicefs_go_build_info{checksum="",mp="/jfs/pvc-60cd26e1-e2f4-45dd-948b-6aafc7f7e8ce",path="github.com/juicedata/juicefs",version="(devel)",vol_name="minio"} 1
# HELP juicefs_memory Used memory in bytes.
# TYPE juicefs_memory gauge
juicefs_memory{mp="/jfs/pvc-60cd26e1-e2f4-45dd-948b-6aafc7f7e8ce",vol_name="minio"} 1.15068928e+08
# HELP juicefs_meta_ops_durations_histogram_seconds Operation latency distributions.
# TYPE juicefs_meta_ops_durations_histogram_seconds histogram
# HELP juicefs_object_request_data_bytes Object requests size in bytes.
# TYPE juicefs_object_request_data_bytes counter
juicefs_object_request_data_bytes{method="PUT",mp="/jfs/pvc-60cd26e1-e2f4-45dd-948b-6aafc7f7e8ce",vol_name="minio"} 3.454749e+06
# HELP juicefs_object_request_durations_histogram_seconds Object requests latency distributions.
# TYPE juicefs_object_request_durations_histogram_seconds histogram
# HELP juicefs_object_request_errors failed requests to object store
# TYPE juicefs_object_request_errors counter
juicefs_object_request_errors{mp="/jfs/pvc-60cd26e1-e2f4-45dd-948b-6aafc7f7e8ce",vol_name="minio"} 0
# HELP juicefs_store_cache_size_bytes size of store cache.
# TYPE juicefs_store_cache_size_bytes gauge
juicefs_store_cache_size_bytes{mp="/jfs/pvc-60cd26e1-e2f4-45dd-948b-6aafc7f7e8ce",vol_name="minio"} 0
# HELP juicefs_transaction_restart The number of times a transaction is restarted.
# TYPE juicefs_transaction_restart counter
juicefs_transaction_restart{mp="/jfs/pvc-60cd26e1-e2f4-45dd-948b-6aafc7f7e8ce",vol_name="minio"} 0
# HELP juicefs_uptime Total running time in seconds.
# TYPE juicefs_uptime gauge
juicefs_uptime{mp="/jfs/pvc-60cd26e1-e2f4-45dd-948b-6aafc7f7e8ce",vol_name="minio"} 9692.14011537
# HELP juicefs_used_buffer_size_bytes size of currently used buffer.
# TYPE juicefs_used_buffer_size_bytes gauge
juicefs_used_buffer_size_bytes{mp="/jfs/pvc-60cd26e1-e2f4-45dd-948b-6aafc7f7e8ce",vol_name="minio"} 0
# HELP juicefs_used_inodes Total number of inodes.
# TYPE juicefs_used_inodes gauge
juicefs_used_inodes{mp="/jfs/pvc-60cd26e1-e2f4-45dd-948b-6aafc7f7e8ce",vol_name="minio"} 7
# HELP juicefs_used_space Total used space in bytes.
# TYPE juicefs_used_space gauge
juicefs_used_space{mp="/jfs/pvc-60cd26e1-e2f4-45dd-948b-6aafc7f7e8ce",vol_name="minio"} 262144
# HELP process_cpu_seconds_total Total user and system CPU time spent in seconds.
# TYPE process_cpu_seconds_total counter
process_cpu_seconds_total 77.15
# HELP process_max_fds Maximum number of open file descriptors.
# TYPE process_max_fds gauge
process_max_fds 1.048576e+06
# HELP process_open_fds Number of open file descriptors.
# TYPE process_open_fds gauge
process_open_fds 17
# HELP process_resident_memory_bytes Resident memory size in bytes.
# TYPE process_resident_memory_bytes gauge
process_resident_memory_bytes 1.15068928e+08
# HELP process_start_time_seconds Start time of the process since unix epoch in seconds.
# TYPE process_start_time_seconds gauge
process_start_time_seconds 1.63185359198e+09
# HELP process_virtual_memory_bytes Virtual memory size in bytes.
# TYPE process_virtual_memory_bytes gauge
process_virtual_memory_bytes 1.74823424e+09
# HELP process_virtual_memory_max_bytes Maximum amount of virtual memory available in bytes.
# TYPE process_virtual_memory_max_bytes gauge
process_virtual_memory_max_bytes -1`
}

func TestJuiceFSEngine_parseMetric(t *testing.T) {
	type args struct {
		metrics string
	}
	tests := []struct {
		name          string
		args          args
		wantPodMetric fuseMetrics
	}{
		{
			name: "test",
			args: args{
				metrics: mockJuiceFSMetric(),
			},
			wantPodMetric: fuseMetrics{
				blockCacheBytes:     396462,
				blockCacheHits:      3861,
				blockCacheMiss:      0,
				blockCacheHitsBytes: 3288597,
				blockCacheMissBytes: 0,
				usedSpace:           262144,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := JuiceFSEngine{}
			if gotPodMetric := j.parseMetric(tt.args.metrics); !reflect.DeepEqual(gotPodMetric, tt.wantPodMetric) {
				t.Errorf("parseMetric() = %v, want %v", gotPodMetric, tt.wantPodMetric)
			}
		})
	}
}

func TestJuiceFSEngine_getPodMetrics(t *testing.T) {
	GetMetricCommon := func(a operations.JuiceFileUtils) (metric string, err error) {
		return mockJuiceFSMetric(), nil
	}
	err := gohook.Hook(operations.JuiceFileUtils.GetMetric, GetMetricCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	j := JuiceFSEngine{}
	gotMetrics, err := j.GetPodMetrics("test")
	if err != nil {
		t.Errorf("getPodMetrics() error = %v", err)
		return
	}
	if gotMetrics != mockJuiceFSMetric() {
		t.Errorf("getPodMetrics() gotMetrics = %v, want %v", gotMetrics, mockJuiceFSMetric())
	}
}
