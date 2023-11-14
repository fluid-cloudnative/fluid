/*
Copyright 2023 The Fluid Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package compatibility

import (
	"log"

	"github.com/fluid-cloudnative/fluid/pkg/utils/testutil"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
)

var batchV1CronJobCompatible = false

func init() {
	if testutil.IsUnitTest() {
		return
	}
	discoverBatchAPICompatibility()
}

// DiscoverBatchAPICompatibility discovers compatibility of the batch API group in the cluster and set in batchV1CronJobCompatible variable.
func discoverBatchAPICompatibility() {
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
