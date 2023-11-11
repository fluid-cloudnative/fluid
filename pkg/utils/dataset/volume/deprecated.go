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
