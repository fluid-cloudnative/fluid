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

package juicefs

import (
	"github.com/fluid-cloudnative/fluid/pkg/common/deprecated"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
)

func (j *JuiceFSEngine) getDeprecatedCommonLabelName() string {
	return deprecated.LabelAnnotationStorageCapacityPrefix + j.namespace + "-" + j.name
}

func (j *JuiceFSEngine) HasDeprecatedCommonLabelName() (deprecated bool, err error) {
	// return deprecated.LabelAnnotationStorageCapacityPrefix + e.namespace + "-" + e.name

	var (
		fuseName  string = j.getFuseDaemonsetName()
		namespace string = j.namespace
	)

	fuses, err := j.getDaemonset(fuseName, namespace)
	if err != nil {
		if apierrs.IsNotFound(err) {
			j.Log.Info("Fuses with deprecated label not found")
			deprecated = false
			err = nil
			return
		}
		j.Log.Error(err, "Failed to get fuse", "fuseName", fuseName)
		return deprecated, err
	}

	nodeSelectors := fuses.Spec.Template.Spec.NodeSelector
	j.Log.Info("The current node selectors for worker", "fuseName", fuseName, "nodeSelector", nodeSelectors)

	if _, deprecated = nodeSelectors[j.getDeprecatedCommonLabelName()]; deprecated {
		j.Log.Info("the deprecated node selector exists", "nodeselector", j.getDeprecatedCommonLabelName())
	} else {
		j.Log.Info("The deprecated node selector doesn't exist", "nodeselector", j.getDeprecatedCommonLabelName())
	}

	return
}
