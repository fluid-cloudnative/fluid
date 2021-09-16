package base

import (
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_convertToTieredstoreInfo(t *testing.T) {
	type args struct {
		tieredstore v1alpha1.TieredStore
	}

	quota20Gi := resource.MustParse("20Gi")
	quota10Gi := resource.MustParse("10Gi")

	tests := []struct {
		name    string
		args    args
		want    TieredStoreInfo
		wantErr bool
	}{
		{
			name: "Test: No quota config set err",
			args: args{tieredstore: v1alpha1.TieredStore{Levels: []v1alpha1.Level{
				{
					Quota:     nil,
					QuotaList: "",
				},
			}}},
			want:    TieredStoreInfo{},
			wantErr: true,
		},
		{
			name: "Test: Inconsistent length of quotas and paths",
			args: args{tieredstore: v1alpha1.TieredStore{Levels: []v1alpha1.Level{
				{
					Path:      "/path/to/cache1/,/path/to/cache2",
					QuotaList: "10Gi,20Gi,30Gi",
				},
			}}},
			want:    TieredStoreInfo{},
			wantErr: true,
		},
		{
			name: "Test: Only quota is set, divide quota equally",
			args: args{tieredstore: v1alpha1.TieredStore{Levels: []v1alpha1.Level{
				{
					Path:  "/path/to/cache1/,/path/to/cache2",
					Quota: resource.NewQuantity(1024, resource.BinarySI),
				},
			}}},
			want: TieredStoreInfo{Levels: []Level{
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
			args: args{tieredstore: v1alpha1.TieredStore{Levels: []v1alpha1.Level{
				{
					Path: "/path/to/cache1/,/path/to/cache2/",
					// QuotaList Overwrites Quota
					Quota:     resource.NewQuantity(124, resource.BinarySI),
					QuotaList: "10Gi,20Gi",
				},
			}}},
			want: TieredStoreInfo{Levels: []Level{
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
		tieredstore v1alpha1.TieredStore
	}

	tieredstore := v1alpha1.TieredStore{
		Levels: []v1alpha1.Level{
			{
				MediumType: "MEM",
				Path:       "/dev/shm/cache/cache1,/dev/shm/cache/cache2/",
				Quota:      nil,
				QuotaList:  "3Gi,2Gi",
				High:       "0.95",
				Low:        "0.7",
			},
			{
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
				!reflect.DeepEqual(gotRuntime.GetTieredStoreInfo(), tieredstoreInfo) {
				t.Errorf("BuildRuntimeInfo() gotRuntime = %v, want %v", gotRuntime, tt.wantRuntime)
			}

		})
	}
}

func TestGetRuntimeInfo(t *testing.T) {
	s := runtime.NewScheme()

	alluxioRuntime := v1alpha1.AlluxioRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "alluxio",
			Namespace: "default",
		},
	}

	dataAlluxio := v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "alluxio",
			Namespace: "default",
		},
		Status: v1alpha1.DatasetStatus{
			Runtimes: []v1alpha1.Runtime{
				v1alpha1.Runtime{
					Name:      "alluxio",
					Namespace: "default",
					Type:      common.ALLUXIO_RUNTIME,
				},
			},
		},
	}

	goosefsRuntime := v1alpha1.GooseFSRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "goosefs",
			Namespace: "default",
		},
	}

	dataGooseFS := v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "goosefs",
			Namespace: "default",
		},
		Status: v1alpha1.DatasetStatus{
			Runtimes: []v1alpha1.Runtime{
				v1alpha1.Runtime{
					Name:      "goosefs",
					Namespace: "default",
					Type:      common.GooseFSRuntime,
				},
			},
		},
	}

	jindoRuntime := v1alpha1.JindoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "jindo",
			Namespace: "default",
		},
	}

	dataJindo := v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "jindo",
			Namespace: "default",
		},
		Status: v1alpha1.DatasetStatus{
			Runtimes: []v1alpha1.Runtime{
				v1alpha1.Runtime{
					Name:      "jindo",
					Namespace: "default",
					Type:      common.JINDO_RUNTIME,
				},
			},
		},
	}
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.AlluxioRuntime{})
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.GooseFSRuntime{})
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.JindoRuntime{})
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.Dataset{})
	_ = v1.AddToScheme(s)
	alluxioRuntimeObjs := []runtime.Object{}
	goosefsRuntimeObjs := []runtime.Object{}
	jindoRuntimeObjs := []runtime.Object{}

	alluxioRuntimeObjs = append(alluxioRuntimeObjs, &alluxioRuntime, &dataAlluxio)
	goosefsRuntimeObjs = append(goosefsRuntimeObjs, &goosefsRuntime, &dataGooseFS)
	jindoRuntimeObjs = append(jindoRuntimeObjs, &jindoRuntime, &dataJindo)
	type args struct {
		client    client.Client
		name      string
		namespace string
	}
	tests := []struct {
		name    string
		args    args
		want    RuntimeInfoInterface
		wantErr bool
	}{
		{
			name: "alluxio_test",
			args: args{
				client:    fake.NewFakeClientWithScheme(s, alluxioRuntimeObjs...),
				name:      "alluxio",
				namespace: "default",
			},
			want: &RuntimeInfo{
				name:        "alluxio",
				namespace:   "default",
				runtimeType: common.ALLUXIO_RUNTIME,
				// fuse global is set to true since v0.7.0
				fuse: Fuse{
					Global: true,
				},
			},
			wantErr: false,
		},
		{
			name: "goosefs_test",
			args: args{
				client:    fake.NewFakeClientWithScheme(s, goosefsRuntimeObjs...),
				name:      "goosefs",
				namespace: "default",
			},
			want: &RuntimeInfo{
				name:        "goosefs",
				namespace:   "default",
				runtimeType: common.GooseFSRuntime,
				// fuse global is set to true since v0.7.0
				fuse: Fuse{
					Global: true,
				},
			},
			wantErr: false,
		},
		{
			name: "jindo_test",
			args: args{
				client:    fake.NewFakeClientWithScheme(s, jindoRuntimeObjs...),
				name:      "jindo",
				namespace: "default",
			},
			want: &RuntimeInfo{
				name:        "jindo",
				namespace:   "default",
				runtimeType: common.JINDO_RUNTIME,
				// fuse global is set to true since v0.7.0
				fuse: Fuse{
					Global: true,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRuntimeInfo(tt.args.client, tt.args.name, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRuntimeInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRuntimeInfo() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
