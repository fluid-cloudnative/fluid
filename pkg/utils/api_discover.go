package utils

import (
	"fmt"
	nativelog "log"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/testutil"
	"github.com/pkg/errors"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
)

var enabledFluidResources map[string]bool = nil

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
	resources, err := discoveryClient.ServerResourcesForGroupVersion(fluidGroupVersion)
	if err != nil {
		panic(errors.Wrapf(err, "failed to discover installed fluid runtime CRDs under %s", fluidGroupVersion))
	}

	enabledFluidResources = make(map[string]bool, len(resources.APIResources))
	for _, res := range resources.APIResources {
		lowerResName := strings.ToLower(res.SingularName)
		enabledFluidResources[lowerResName] = true
	}
}

func ResourceEnabled(resourceSingularName string) bool {
	if enabled, exists := enabledFluidResources[resourceSingularName]; exists && enabled {
		return true
	}

	return false
}
