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

package volume

import (
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func HasDeprecatedPersistentVolumeName(client client.Client, runtime base.RuntimeInfoInterface, log logr.Logger) (deprecated bool, err error) {
	deprecated, err = kubeclient.IsPersistentVolumeExist(client, runtime.GetName(), common.ExpectedFluidAnnotations)
	if err != nil {
		log.Error(err, "Failed to check if deprecated PV exists", "expeceted PV name", runtime.GetName())
		return
	}

	if deprecated {
		log.Info("Found deprecated PV", "pv name", runtime.GetName())
	} else {
		log.Info("No deprecated PV found", "pv name", fmt.Sprintf("%s-%s", runtime.GetNamespace(), runtime.GetName()))
	}

	return
}
