package common

import "testing"

func Test_parseResolvConf(t *testing.T) {
	type args struct {
		conf string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			args: args{
				conf: `nameserver 10.255.0.10
search default.svc.cluster.local svc.cluster.local cluster.local
options ndots:5`,
			},
			want:    "cluster.local",
			wantErr: false,
		},
		{
			args: args{
				conf: `nameserver 10.255.0.10
search default.svc.cluster.local svc.cluster.local cluster.local foo.bar
options ndots:5`,
			},
			want:    "cluster.local",
			wantErr: false,
		},
		{
			args: args{
				conf: `nameserver 10.255.0.10
search default.svc.clusterxx.local svc.clusterxx.local clusterxx.local
options ndots:5`,
			},
			want:    "clusterxx.local",
			wantErr: false,
		},
		{
			args: args{
				conf: `nameserver 10.255.0.10
options ndots:5`,
			},
			want:    "",
			wantErr: true,
		},
		{
			args: args{
				conf: `nameserver 10.255.0.10
search a
options ndots:5`,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseResolvConf(tt.args.conf)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseResolvConf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseResolvConf() = %v, want %v", got, tt.want)
			}
		})
	}
}
