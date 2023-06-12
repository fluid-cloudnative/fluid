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
	if resources, err := discoveryClient.ServerResourcesForGroupVersion("batch/v1"); err != nil {
		if !errors.IsNotFound(err) {
			log.Fatalf("failed to discover batch/v1 group version: %v", err)
		}
	} else {
		for _, res := range resources.APIResources {
			if res.Name == "cronjobs" {
				batchV1CronJobCompatible = true
				break
			}
		}
	}

	if !batchV1CronJobCompatible {
		log.Print("batch/v1 cronjobs not found in the cluster, fall back to use batch/v1beta1 cronjobs")
	}
}

func IsBatchV1CronJobSupported() bool {
	return batchV1CronJobCompatible
}
