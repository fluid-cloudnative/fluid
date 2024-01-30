package utils

import (
	"fmt"
	nativelog "log"
	"strings"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/testutil"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
)

var enabledFluidResources map[string]bool = nil

// exponentail backoff with a interval of 0.2s/1s/5s/25s/30s
var backOff = wait.Backoff{
	Steps:    5,
	Duration: 200 * time.Millisecond,
	Factor:   5.0,
	Jitter:   0.1,
	Cap:      30 * time.Second,
}

func init() {
	if testutil.IsUnitTest() {
		return
	}
	discoverFluidResourcesInCluster()
	allEnabledResources := []string{}
	for resource := range enabledFluidResources {
		allEnabledResources = append(allEnabledResources, resource)
	}

	nativelog.Printf("Discovered Fluid CRDs in cluster: %v, enable related reconcilers only", allEnabledResources)
}

func discoverFluidResourcesInCluster() {
	restConfig := ctrl.GetConfigOrDie()
	discoveryClient := discovery.NewDiscoveryClientForConfigOrDie(restConfig)
	fluidGroupVersion := fmt.Sprintf("%s/%s", datav1alpha1.Group, datav1alpha1.Version)

	err := retry.OnError(backOff, func(err error) bool { return true }, func() error {
		resources, discoverErr := discoveryClient.ServerResourcesForGroupVersion(fluidGroupVersion)
		if discoverErr != nil {
			return discoverErr
		}

		if len(resources.APIResources) == 0 {
			return fmt.Errorf("none of Fluid CRDs is found installed in the cluster")
		}

		enabledFluidResources = make(map[string]bool, len(resources.APIResources))
		for _, res := range resources.APIResources {
			lowerResName := strings.ToLower(res.SingularName)
			enabledFluidResources[lowerResName] = true
		}

		return nil
	})
	if err != nil {
		nativelog.Fatalf("failed to discover installed fluid runtime CRDs under %s: %v", fluidGroupVersion, err)
	}
}

func ResourceEnabled(resourceSingularName string) bool {
	if enabled, exists := enabledFluidResources[resourceSingularName]; exists && enabled {
		return true
	}

	return false
}
