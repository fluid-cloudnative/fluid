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
		wantResources map[string]bool
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
			wantResources: map[string]bool{
				"dataset":        true,
				"alluxioruntime": true,
				"dataload":       true,
			},
		},
	}
	for _, tt := range tests {
		// clear global-wise enabledFluidResources for following tests
		enabledFluidResources = map[string]bool{}
		t.Run(tt.name, func(t *testing.T) {
			var fakeClient *discovery.DiscoveryClient
			patch3 := gomonkey.ApplyMethodFunc(fakeClient, "ServerResourcesForGroupVersion", tt.patchedFunc)
			defer patch3.Reset()

			discoverFluidResourcesInCluster()

			if !reflect.DeepEqual(tt.wantResources, enabledFluidResources) {
				t.Fatalf("failed to discoverFluidResourcesInCluster, got %v, want %v", enabledFluidResources, tt.wantResources)
			}
		})
	}
}
