/*
Copyright 2023 The Fluid Author.

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

package updatedbconf

import "testing"

func Test_configUpdate(t *testing.T) {
	type args struct {
		content  string
		newFs    []string
		newPaths []string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "add new path and fs",
			args: args{
				newFs:    []string{"fuse.alluxio-fuse", "fuse.jindofs-fuse", "JuiceFS", "fuse.goosefs-fuse"},
				newPaths: []string{"/runtime-mnt"},
				content: `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/tmp /var/spool /media /var/lib/os-prober /var/lib/ceph /home/.ecryptfs /var/lib/schroot"
PRUNEFS="foo bar"`,
			},
			want: `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/tmp /var/spool /media /var/lib/os-prober /var/lib/ceph /home/.ecryptfs /var/lib/schroot /runtime-mnt"
PRUNEFS="foo bar fuse.alluxio-fuse fuse.jindofs-fuse JuiceFS fuse.goosefs-fuse"`,
			wantErr: false,
		},
		{
			name: "no new path",
			args: args{
				newFs:    []string{"fuse.alluxio-fuse", "fuse.jindofs-fuse", "JuiceFS", "fuse.goosefs-fuse"},
				newPaths: []string{"/runtime-mnt"},
				content: `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/tmp /var/spool /media /var/lib/os-prober /var/lib/ceph /home/.ecryptfs /var/lib/schroot /runtime-mnt"
PRUNEFS="foo bar"`,
			},
			want: `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/tmp /var/spool /media /var/lib/os-prober /var/lib/ceph /home/.ecryptfs /var/lib/schroot /runtime-mnt"
PRUNEFS="foo bar fuse.alluxio-fuse fuse.jindofs-fuse JuiceFS fuse.goosefs-fuse"`,
			wantErr: false,
		},
		{
			name: "empty path or fs config",
			args: args{
				newFs:    []string{"fuse.alluxio-fuse", "fuse.jindofs-fuse", "JuiceFS", "fuse.goosefs-fuse"},
				newPaths: []string{"/runtime-mnt"},
				content:  `PRUNE_BIND_MOUNTS="yes"`,
			},
			want: `PRUNE_BIND_MOUNTS="yes"
PRUNEFS="fuse.alluxio-fuse fuse.jindofs-fuse JuiceFS fuse.goosefs-fuse"
PRUNEPATHS="/runtime-mnt"`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := updateConfig(tt.args.content, tt.args.newFs, tt.args.newPaths)
			if (err != nil) != tt.wantErr {
				t.Errorf("configUpdate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("configUpdate() = \n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}
