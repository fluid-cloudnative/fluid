package discovery

import (
	"fmt"
	nativelog "log"
	"strings"
	"sync"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
)

type fluidDiscovery map[string]bool

var (
	globalDiscovery fluidDiscovery = nil
	once            sync.Once
)

// ResourceEnabled checks the map to determine whether a Fluid resource is installed and enabled in the cluster.
func (discovery fluidDiscovery) ResourceEnabled(resourceSingularName string) bool {
	if enabled, exists := discovery[resourceSingularName]; exists && enabled {
		return true
	}

	return false
}

// GetFluidDiscoery returns a global-level singleton of fluidDiscovery.
func GetFluidDiscovery() fluidDiscovery {
	once.Do(func() {
		initDiscovery() // initDiscovery() fails with os.Exit(1), no need to handle error here.
	})
	return globalDiscovery
}

// exponentail backoff with a interval of 0.2s/1s/5s/25s/30s
var backOff = wait.Backoff{
	Steps:    5,
	Duration: 200 * time.Millisecond,
	Factor:   5.0,
	Jitter:   0.1,
	Cap:      30 * time.Second,
}

// initDiscovery initializes the pacakge by discovering all the Fluid CRDs and record them into fluidDiscovery.
// Further calls of ResourceEnabled checks fluidDiscovery to know if the resource is installed in the cluster.
func initDiscovery() {
	discoverFluidResourcesInCluster()
	allEnabledResources := []string{}
	for resource := range globalDiscovery {
		allEnabledResources = append(allEnabledResources, resource)
	}

	nativelog.Printf("Discovered Fluid CRDs in cluster: %v, enable related reconcilers only", allEnabledResources)
}

func discoverFluidResourcesInCluster() {
	restConfig := ctrl.GetConfigOrDie()
	var discoveryClient discovery.DiscoveryInterface = discovery.NewDiscoveryClientForConfigOrDie(restConfig)
	fluidGroupVersion := fmt.Sprintf("%s/%s", datav1alpha1.Group, datav1alpha1.Version)

	err := retry.OnError(backOff, func(err error) bool { return true }, func() error {
		resources, discoverErr := discoveryClient.ServerResourcesForGroupVersion(fluidGroupVersion)
		if discoverErr != nil {
			return discoverErr
		}

		if len(resources.APIResources) == 0 {
			return fmt.Errorf("none of Fluid CRDs is found installed in the cluster")
		}

		globalDiscovery = make(map[string]bool, len(resources.APIResources))
		for _, res := range resources.APIResources {
			lowerResName := strings.ToLower(res.SingularName)
			globalDiscovery[lowerResName] = true
		}

		return nil
	})
	if err != nil {
		nativelog.Fatalf("failed to discover installed fluid runtime CRDs under %s: %v", fluidGroupVersion, err)
	}
}
