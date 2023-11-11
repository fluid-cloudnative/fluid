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

package jindofsx

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
)

var cfg = `
[bigboot]
logger.dir =  /dev/shm/default/oss-tf-dataset/bigboot/log
logger.cleanner.enable = true

[jindofsx-common]
jfs.namespaces = spark
jfs.namespaces.spark.oss.uri = oss://tensorflow-datasets.oss-cn-shanghai-internal.aliyuncs.com/
namespace.backend.type = rocksdb
namespace.blocklet.cache.size = 1000000
namespace.filelet.cache.size = 100000
namespace.meta-dir = /dev/shm/default/oss-tf-dataset/bigboot/server
namespace.rpc.port = 18000
namespace.filelet.atime.enable = false

[jindofsx-storage]
namespace.meta-dir = /dev/shm/default/oss-tf-dataset/bigboot/bignode
storage.data-dirs = /dev/shm/default/oss-tf-dataset/bigboot
storage.data-dirs.capacities = 10g
storage.ram.cache.size = 10g
storage.rpc.port = 18001
namespace.meta-dir = /dev/shm/default/oss-tf-dataset/bigboot/bignode
storage.compaction.enable = false

[jindofsx-namespace]
client.oss.upload.queue.size = 5
client.oss.upload.threads = 4
client.storage.rpc.port = 18001
`

func Test_parsePortsFromConfigMap(t *testing.T) {
	type args struct {
		configMap *v1.ConfigMap
	}
	tests := []struct {
		name      string
		args      args
		wantPorts []int
		wantErr   bool
	}{
		{
			name: "parse configMap",
			args: args{configMap: &v1.ConfigMap{Data: map[string]string{
				"jindofsx.cfg": cfg,
			}}},
			wantPorts: []int{18000, 18001},
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPorts, err := parsePortsFromConfigMap(tt.args.configMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePortsFromConfigMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPorts, tt.wantPorts) {
				t.Errorf("parsePortsFromConfigMap() gotPorts = %v, want %v", gotPorts, tt.wantPorts)
			}
		})
	}
}
