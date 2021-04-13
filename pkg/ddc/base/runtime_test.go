package base

import (
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
	"reflect"
	"testing"
)

func Test_convertToTieredstoreInfo(t *testing.T) {
	type args struct {
		tieredstore v1alpha1.Tieredstore
	}

	quota20Gi := resource.MustParse("20Gi")
	quota10Gi := resource.MustParse("10Gi")

	tests := []struct {
		name    string
		args    args
		want    TieredstoreInfo
		wantErr bool
	}{
		{
			name: "Test: No quota config set err",
			args: args{tieredstore: v1alpha1.Tieredstore{Levels: []v1alpha1.Level{
				v1alpha1.Level{
					Quota:     nil,
					QuotaList: "",
				},
			}}},
			want:    TieredstoreInfo{},
			wantErr: true,
		},
		{
			name: "Test: Inconsistent length of quotas and paths",
			args: args{tieredstore: v1alpha1.Tieredstore{Levels: []v1alpha1.Level{
				v1alpha1.Level{
					Path:      "/path/to/cache1/,/path/to/cache2",
					QuotaList: "10Gi,20Gi,30Gi",
				},
			}}},
			want:    TieredstoreInfo{},
			wantErr: true,
		},
		{
			name: "Test: Only quota is set, divide quota equally",
			args: args{tieredstore: v1alpha1.Tieredstore{Levels: []v1alpha1.Level{
				v1alpha1.Level{
					Path:  "/path/to/cache1/,/path/to/cache2",
					Quota: resource.NewQuantity(1024, resource.BinarySI),
				},
			}}},
			want: TieredstoreInfo{Levels: []Level{
				{
					CachePaths: []CachePath{
						{
							Path:  "/path/to/cache1",
							Quota: resource.NewQuantity(512, resource.BinarySI),
						},
						{
							Path:  "/path/to/cache2",
							Quota: resource.NewQuantity(512, resource.BinarySI),
						},
					},
				},
			}},
			wantErr: false,
		},
		{
			name: "Test: quotaList for configs",
			args: args{tieredstore: v1alpha1.Tieredstore{Levels: []v1alpha1.Level{
				v1alpha1.Level{
					Path: "/path/to/cache1/,/path/to/cache2/",
					// QuotaList Overwrites Quota
					Quota:     resource.NewQuantity(124, resource.BinarySI),
					QuotaList: "10Gi,20Gi",
				},
			}}},
			want: TieredstoreInfo{Levels: []Level{
				{
					CachePaths: []CachePath{
						{
							Path:  "/path/to/cache1",
							Quota: &quota10Gi,
						},
						{
							Path:  "/path/to/cache2",
							Quota: &quota20Gi,
						},
					},
				},
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToTieredstoreInfo(tt.args.tieredstore)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertToTieredstoreInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToTieredstoreInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildRuntimeInfo(t *testing.T) {
	const runtimetype = "alluxio"
	type args struct {
		name        string
		namespace   string
		runtimeType string
		tieredstore v1alpha1.Tieredstore
	}

	tieredstore := v1alpha1.Tieredstore{
		Levels: []v1alpha1.Level{
			v1alpha1.Level{
				MediumType: "MEM",
				Path:       "/dev/shm/cache/cache1,/dev/shm/cache/cache2/",
				Quota:      nil,
				QuotaList:  "3Gi,2Gi",
				High:       "0.95",
				Low:        "0.7",
			},
			v1alpha1.Level{
				MediumType: "SSD",
				Path:       "/mnt/cache",
				// 1 << 29 == 1Gi
				Quota:     resource.NewQuantity(1024*1024*1024, resource.BinarySI),
				QuotaList: "",
				High:      "0.99",
				Low:       "0.8",
			},
		},
	}

	tests := []struct {
		name        string
		args        args
		wantRuntime RuntimeInfoInterface
		wantErr     bool
	}{
		{
			name: "TestBuildRuntimeInfo",
			args: args{
				name:        "dataset",
				namespace:   "default",
				runtimeType: runtimetype,
				tieredstore: tieredstore,
			},
			wantRuntime: &RuntimeInfo{
				name:        "dataset",
				namespace:   "default",
				runtimeType: runtimetype,
				exclusive:   false,
				//setup:       false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRuntime, err := BuildRuntimeInfo(tt.args.name, tt.args.namespace, tt.args.runtimeType, tt.args.tieredstore)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildRuntimeInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			tieredstoreInfo, err := convertToTieredstoreInfo(tieredstore)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertToTieredstoreInfo() error = %v, wantErr %v", err, tt.wantErr)
			}

			if gotRuntime.GetName() != tt.wantRuntime.GetName() ||
				gotRuntime.GetNamespace() != tt.wantRuntime.GetNamespace() ||
				gotRuntime.GetRuntimeType() != tt.wantRuntime.GetRuntimeType() ||
				!reflect.DeepEqual(gotRuntime.GetTieredstoreInfo(), tieredstoreInfo) {
				t.Errorf("BuildRuntimeInfo() gotRuntime = %v, want %v", gotRuntime, tt.wantRuntime)
			}

		})
	}
}
