package discovery

import (
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestResourceEnabled(t *testing.T) {
	globalDiscovery = map[string]bool{
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
			if got := globalDiscovery.ResourceEnabled(tt.args.resourceSingularName); got != tt.want {
				t.Errorf("ResourceEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_discoverFluidResourcesInCluster(t *testing.T) {
	patch1 := gomonkey.ApplyFunc(ctrl.GetConfigOrDie, func() *rest.Config {
		return nil
	})
	defer patch1.Reset()

	patch2 := gomonkey.ApplyFunc(discovery.NewDiscoveryClientForConfigOrDie, func(_ *rest.Config) *discovery.DiscoveryClient {
		return nil
	})
	defer patch2.Reset()

	tests := []struct {
		name          string
		patchedFunc   func(groupVersion string) (*metav1.APIResourceList, error)
		wantResources fluidDiscovery
	}{
		{
			name: "test",
			patchedFunc: func(groupVersion string) (*metav1.APIResourceList, error) {
				return &metav1.APIResourceList{
					APIResources: []metav1.APIResource{
						{
							SingularName: "dataset",
						},
						{
							SingularName: "alluxioruntime",
						},
						{
							SingularName: "dataload",
						},
					},
				}, nil
			},
			wantResources: fluidDiscovery(map[string]bool{
				"dataset":        true,
				"alluxioruntime": true,
				"dataload":       true,
			}),
		},
	}
	for _, tt := range tests {
		// clear global-wise enabledFluidResources for following tests
		globalDiscovery = map[string]bool{}
		t.Run(tt.name, func(t *testing.T) {
			var fakeClient *discovery.DiscoveryClient
			patch3 := gomonkey.ApplyMethodFunc(fakeClient, "ServerResourcesForGroupVersion", tt.patchedFunc)
			defer patch3.Reset()

			discoverFluidResourcesInCluster()

			if !reflect.DeepEqual(tt.wantResources, globalDiscovery) {
				t.Fatalf("failed to discoverFluidResourcesInCluster, got %v, want %v", globalDiscovery, tt.wantResources)
			}
		})
	}
}

func TestGetFluidDiscovery(t *testing.T) {
	want1 := fluidDiscovery(map[string]bool{
		"foo": true,
		"bar": true,
	})
	patch := gomonkey.ApplyFunc(initDiscovery, func() {
		globalDiscovery = want1
	})
	defer patch.Reset()

	t.Run("first time globalDiscovery", func(t *testing.T) {
		got := GetFluidDiscovery()
		if !reflect.DeepEqual(got, want1) {
			t.Errorf("GetFluidDiscovery() = %v, want %v", got, want1)
		}
	})

	// For the second time, global Discovery will not be rewritten to "want1" as we use once.Do()
	want2 := fluidDiscovery(map[string]bool{
		"foo2": true,
		"bar2": true,
	})
	globalDiscovery = want2

	t.Run("second time globalDiscovery", func(t *testing.T) {

		got := GetFluidDiscovery()
		if !reflect.DeepEqual(got, want2) {
			t.Errorf("GetFluidDiscovery() = %v, want %v", got, globalDiscovery)
		}
	})
}
