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

package alluxio

import (
	"github.com/fluid-cloudnative/fluid/pkg/common/deprecated"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
)

func (e *AlluxioEngine) getDeprecatedCommonLabelname() string {
	return deprecated.LabelAnnotationStorageCapacityPrefix + e.namespace + "-" + e.name
}

func (e *AlluxioEngine) HasDeprecatedCommonLabelname() (deprecated bool, err error) {

	// return deprecated.LabelAnnotationStorageCapacityPrefix + e.namespace + "-" + e.name

	var (
		workerName string = e.getWorkerName()
		namespace  string = e.namespace
	)

	// runtime, err := e.getRuntime()
	// if err != nil {
	// 	return
	// }

	workers, err := e.getDaemonset(workerName, namespace)
	if err != nil {
		if apierrs.IsNotFound(err) {
			e.Log.Info("Workers with deprecated label not found")
			deprecated = false
			err = nil
			return
		}
		e.Log.Error(err, "Failed to get worker", "workerName", workerName)
		return deprecated, err
	}

	nodeSelectors := workers.Spec.Template.Spec.NodeSelector
	e.Log.Info("The current node selectors for worker", "workerName", workerName, "nodeSelector", nodeSelectors)

	if _, deprecated = nodeSelectors[e.getDeprecatedCommonLabelname()]; deprecated {
		//
		e.Log.Info("the deprecated node selector exists", "nodeselector", e.getDeprecatedCommonLabelname())
	} else {
		e.Log.Info("The deprecated node selector doesn't exist", "nodeselector", e.getDeprecatedCommonLabelname())
	}

	return
}
