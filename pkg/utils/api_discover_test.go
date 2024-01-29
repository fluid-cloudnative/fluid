package utils

import "testing"

func TestResourceEnabled(t *testing.T) {
	enabledFluidResources = map[string]bool{
		"dataload":       true,
		"alluxioruntime": true,
		"dataset":        true,
	}

	type args struct {
		resourceSingularName string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "dataset is enabled",
			args: args{
				resourceSingularName: "dataset",
			},
			want: true,
		},
		{
			name: "alluxioruntime is enabled",
			args: args{
				resourceSingularName: "alluxioruntime",
			},
			want: true,
		},
		{
			name: "databackup is disabled",
			args: args{
				resourceSingularName: "databackup",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResourceEnabled(tt.args.resourceSingularName); got != tt.want {
				t.Errorf("ResourceEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}
