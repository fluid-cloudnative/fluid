package compatibility

import (
	"log"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
)

var batchV1CronJobCompatible = false

func init() {
	restConfig := ctrl.GetConfigOrDie()

	discoveryClient := discovery.NewDiscoveryClientForConfigOrDie(restConfig)

	resources, err := discoveryClient.ServerResourcesForGroupVersion("batch/v1")
	if err != nil && !errors.IsNotFound(err) {
		log.Fatalf("failed to discover batch/v1 group version: %v", err)
	}

	if len(resources.APIResources) > 0 {
		for _, res := range resources.APIResources {
			if res.Name == "cronjobs" {
				batchV1CronJobCompatible = true
				break
			}
		}
	}
}

func IsBatchV1CronJobSupported() bool {
	return batchV1CronJobCompatible
}
