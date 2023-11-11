/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
