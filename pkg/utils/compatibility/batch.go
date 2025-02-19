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
	"github.com/blang/semver/v4"
	"log"

	"github.com/fluid-cloudnative/fluid/pkg/utils/testutil"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
)

var batchV1CronJobCompatible = false
var nodeBindingTokenSupported = false

// Beta release, default enabled.
const nodeBindingTokenSupportedVersion = "v1.30.0"

func init() {
	if testutil.IsUnitTest() {
		return
	}
	discoveryCompatibility()
}

func discoveryCompatibility() {

	restConfig := ctrl.GetConfigOrDie()
	discoveryClient := discovery.NewDiscoveryClientForConfigOrDie(restConfig)

	discoverBatchAPICompatibility(discoveryClient)

	discoverNodeBindingTokenCompatibility(discoveryClient)
}

// DiscoverBatchAPICompatibility discovers compatibility of the batch API group in the cluster and set in batchV1CronJobCompatible variable.
func discoverBatchAPICompatibility(discoveryClient *discovery.DiscoveryClient) {
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

func discoverNodeBindingTokenCompatibility(discoveryClient *discovery.DiscoveryClient) {
	serverVersion, err := discoveryClient.ServerVersion()
	if err != nil && !errors.IsNotFound(err) {
		log.Fatalf("failed to discover batch/v1 group version: %v", err)
	}
	// transform to semver.Version and compare
	currentVersion, err := semver.ParseTolerant(serverVersion.GitVersion)
	if err != nil {
		log.Fatalf("Failed to parse current version: %v", err)
	}
	targetVersion, err := semver.ParseTolerant(nodeBindingTokenSupportedVersion)
	if err != nil {
		log.Fatalf("Failed to parse target version: %v", err)
	}

	if currentVersion.GT(targetVersion) {
		nodeBindingTokenSupported = true
	}
}

func IsBatchV1CronJobSupported() bool {
	return batchV1CronJobCompatible
}

func IsNodeBindingTokenSupported() bool {
	return nodeBindingTokenSupported
}
