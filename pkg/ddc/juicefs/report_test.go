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
	return `blockcache.blocks: 359926
blockcache.bytes: 536315271087
blockcache.evict: 631631
blockcache.evictBytes: 1678993012494
blockcache.evictDur: 781200
blockcache.hitBytes: 8341055360094
blockcache.hits: 2590312
blockcache.miss: 1002930
blockcache.missBytes: 1935703021973
blockcache.readDuration: 9841510309
blockcache.transfer: 0
blockcache.transferBytes: 0
blockcache.transferDur: 0
blockcache.transferScanDur: 0
blockcache.write: 878685
blockcache.writeBytes: 1929020439752
blockcache.writeDuration: 1912409077
cpuusage: 45146399768
fuse_ops.access: 0
fuse_ops.copy_file_range: 0
fuse_ops.create: 343128
fuse_ops.fallocate: 8432
fuse_ops.flock: 10209
fuse_ops.flush: 1321489
fuse_ops.fsync: 2654
fuse_ops.getattr: 12723492
fuse_ops.getlk: 0
fuse_ops.getxattr: 64245
fuse_ops.link: 0
fuse_ops.listxattr: 34246
fuse_ops.lookup.cache: 0
fuse_ops.lookup: 11108473
fuse_ops.mkdir: 7320
fuse_ops.mknod: 0
fuse_ops.open: 938395
fuse_ops.opendir: 4331813
fuse_ops.read: 29264056
fuse_ops.readdir: 9860174
fuse_ops.readlink: 83667
fuse_ops.release: 1280310
fuse_ops.releasedir: 4331813
fuse_ops.removexattr: 0
fuse_ops.rename: 5940
fuse_ops.resolve: 0
fuse_ops.rmdir: 7009
fuse_ops.setattr: 334082
fuse_ops.setlk: 1
fuse_ops.setxattr: 260
fuse_ops.statfs: 122772
fuse_ops.summary: 0
fuse_ops.symlink: 1268
fuse_ops.truncate: 118975
fuse_ops.unlink: 238395
fuse_ops.write: 50020133
fuse_ops: 126562751
gcPause: 6235533455
get_bytes: 4137587
goroutines: 106
handles: 1213
heapCacheUsed: 0
heapInuse: 2603769856
heapSys: 8219760008
memusage: 3745714176
meta.bytes_received: 4044902811
meta.bytes_sent: 764854860
meta.dircache.access: 115642
meta.dircache.add: 1804428
meta.dircache.addEntry: 357021
meta.dircache.getattr: 9397422
meta.dircache.lookup: 8193661
meta.dircache.newDir: 0
meta.dircache.open: 801417
meta.dircache.readdir: 873385
meta.dircache.remove: 1020588
meta.dircache.removeEntry: 251293
meta.dircache.setattr: 2134567
meta.dircache0.dirs: 2
meta.dircache0.inodes: 16
meta.dircache0: 0
meta.dircache: 24949424
meta.packets_received: 15212342
meta.packets_sent: 11933806
meta.reconnects: 2
meta.usec_ping: [407]
meta.usec_timediff: [31157]
meta: 4861853245
metaDuration: 9618560452
metaRequest: 11478098
object.copy: 0
object.delete: 533229
object.error: 73
object.get: 41
object.head: 0
object.list: 0
object.put: 1173014
object: 1706357
objectDuration.delete: 15028921520
objectDuration.get: 1107924
objectDuration.head: 0
objectDuration.list: 0
objectDuration.put: 117350542328
objectDuration: 132380571772
offHeapCacheUsed: 0
openfiles: 9
operationDuration: 69875736987
operations: 126562750
put_bytes: 1134075243126
readBufferUsed: 1472635601
read_bytes: 3269189737736
remotecache.errors: 78
remotecache.get: 1002956
remotecache.getBytes: 1935698827777
remotecache.getDuration: 66132691974
remotecache.put: 1172971
remotecache.putBytes: 2450127940984
remotecache.putDuration: 9764079869
remotecache.receive: 0
remotecache.receiveBytes: 0
remotecache.recvDuration: 0
remotecache.send: 0
remotecache.sendBytes: 0
remotecache.sendDuration: 0
symlink_cache.inserts: 416
symlink_cache.search_hits: 83251
symlink_cache.search_misses: 416
symlink_cache: 84083
threads: 419
totalBufferUsed: 1505634733
uptime: 614133
write_bytes: 2410711988610`
}

func mockJuiceFSMetricOfCommunity() string {
	return `juicefs_blockcache_blocks 9708
juicefs_blockcache_bytes 40757435762
juicefs_blockcache_drops 0
juicefs_blockcache_evicts 0
juicefs_blockcache_hit_bytes 40717671794
juicefs_blockcache_hits 9709
juicefs_blockcache_miss 0
juicefs_blockcache_miss_bytes 0
juicefs_blockcache_read_hist_seconds_total 9709
juicefs_blockcache_read_hist_seconds_sum 973.9802880359989
juicefs_blockcache_write_bytes 0
juicefs_blockcache_write_hist_seconds_total 0
juicefs_blockcache_write_hist_seconds_sum 0
juicefs_blockcache_writes 0
juicefs_compact_size_histogram_bytes_total 0
juicefs_compact_size_histogram_bytes_sum 0
juicefs_cpu_usage 101.787289
juicefs_fuse_open_handlers 1
juicefs_fuse_ops_durations_histogram_seconds_total 395082
juicefs_fuse_ops_durations_histogram_seconds_sum 438.98124123499315
juicefs_fuse_read_size_bytes_total 310652
juicefs_fuse_read_size_bytes_sum 40717671794
juicefs_fuse_written_size_bytes_total 0
juicefs_fuse_written_size_bytes_sum 0
juicefs_go_build_info__github.com/juicedata/juicefs_(devel) 1
juicefs_go_goroutines 46
juicefs_go_info_go1.16.15 1
juicefs_go_memstats_alloc_bytes 20049512
juicefs_go_memstats_alloc_bytes_total 373878120
juicefs_go_memstats_buck_hash_sys_bytes 1481768
juicefs_go_memstats_frees_total 6271987
juicefs_go_memstats_gc_cpu_fraction 0.000018793731834382145
juicefs_go_memstats_gc_sys_bytes 11826520
juicefs_go_memstats_heap_alloc_bytes 20049512
juicefs_go_memstats_heap_idle_bytes 174145536
juicefs_go_memstats_heap_inuse_bytes 25149440
juicefs_go_memstats_heap_objects 43126
juicefs_go_memstats_heap_released_bytes 171704320
juicefs_go_memstats_heap_sys_bytes 199294976
juicefs_go_memstats_last_gc_time_seconds 1651914570.9444923
juicefs_go_memstats_lookups_total 0
juicefs_go_memstats_mallocs_total 6315113
juicefs_go_memstats_mcache_inuse_bytes 16800
juicefs_go_memstats_mcache_sys_bytes 32768
juicefs_go_memstats_mspan_inuse_bytes 320416
juicefs_go_memstats_mspan_sys_bytes 1277952
juicefs_go_memstats_next_gc_bytes 40455344
juicefs_go_memstats_other_sys_bytes 3066536
juicefs_go_memstats_stack_inuse_bytes 2031616
juicefs_go_memstats_stack_sys_bytes 2031616
juicefs_go_memstats_sys_bytes 219012136
juicefs_go_threads 31
juicefs_memory 82145280
juicefs_meta_ops_durations_histogram_seconds_total 85488
juicefs_meta_ops_durations_histogram_seconds_sum 26.298036121000194
juicefs_object_request_errors 0
juicefs_process_cpu_seconds_total 101.77
juicefs_process_max_fds 1048576
juicefs_process_open_fds 14
juicefs_process_resident_memory_bytes 82145280
juicefs_process_start_time_seconds 1651910490.62
juicefs_process_virtual_memory_bytes 3328643072
juicefs_process_virtual_memory_max_bytes -1
juicefs_staging_block_bytes 0
juicefs_staging_blocks 0
juicefs_store_cache_size_bytes 0
juicefs_transaction_durations_histogram_seconds_total 138
juicefs_transaction_durations_histogram_seconds_sum 0.17865633699999994
juicefs_transaction_restart 0
juicefs_uptime 4096.290097335
juicefs_used_buffer_size_bytes 0
juicefs_used_inodes 1
juicefs_used_space 40717672448`
}

func TestJuiceFSEngine_parseMetric(t *testing.T) {
	type args struct {
		metrics string
		edition string
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
				edition: "enterprise",
			},
			wantPodMetric: fuseMetrics{
				blockCacheBytes:     536315271087,
				blockCacheHits:      2590312,
				blockCacheMiss:      1002930,
				blockCacheHitsBytes: 8341055360094,
				blockCacheMissBytes: 1935703021973,
				usedSpace:           0,
			},
		},
		{
			name: "test",
			args: args{
				metrics: mockJuiceFSMetricOfCommunity(),
				edition: "community",
			},
			wantPodMetric: fuseMetrics{
				blockCacheBytes:     40757435762,
				blockCacheHits:      9709,
				blockCacheMiss:      0,
				blockCacheHitsBytes: 40717671794,
				blockCacheMissBytes: 0,
				usedSpace:           0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := JuiceFSEngine{}
			if gotPodMetric := j.parseMetric(tt.args.metrics, tt.args.edition); !reflect.DeepEqual(gotPodMetric, tt.wantPodMetric) {
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
	gotMetrics, err := j.GetPodMetrics("test", "test")
	if err != nil {
		t.Errorf("getPodMetrics() error = %v", err)
		return
	}
	if gotMetrics != mockJuiceFSMetric() {
		t.Errorf("getPodMetrics() gotMetrics = %v, want %v", gotMetrics, mockJuiceFSMetric())
	}
}
