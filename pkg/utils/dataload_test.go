package utils

import "testing"

func TestGetDataLoadReleaseName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				name: "imagenet-dataload",
			},
			want: "imagenet-dataload-loader",
		},
		{
			name: "test2",
			args: args{
				name: "test",
			},
			want: "test-loader",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDataLoadReleaseName(tt.args.name); got != tt.want {
				t.Errorf("GetDataLoadReleaseName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDataLoadRef(t *testing.T) {
	type args struct {
		name      string
		namespace string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				name:      "hbase",
				namespace: "default",
			},
			want: "default-hbase",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDataLoadRef(tt.args.name, tt.args.namespace); got != tt.want {
				t.Errorf("GetDataLoadRef() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDataLoadJobName(t *testing.T) {
	type args struct {
		releaseName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{releaseName: GetDataLoadReleaseName("hbase")},
			want: "hbase-loader-job",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDataLoadJobName(tt.args.releaseName); got != tt.want {
				t.Errorf("GetDataLoadJobName() = %v, want %v", got, tt.want)
			}
		})
	}
}
