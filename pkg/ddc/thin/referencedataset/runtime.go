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

package referencedataset

import (
	"context"
	"fmt"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// getPhysicalDatasetRuntimeStatus get the runtime status of the physical dataset
func (e *ReferenceDatasetEngine) getPhysicalDatasetRuntimeStatus() (status *datav1alpha1.RuntimeStatus, err error) {
	physicalRuntimeInfo, err := e.getPhysicalRuntimeInfo()
	if err != nil {
		return status, err
	}

	// if physicalRuntimeInfo is nil and no err, the runtime is deleting.
	if physicalRuntimeInfo == nil {
		return nil, nil
	}

	return base.GetRuntimeStatus(e.Client, physicalRuntimeInfo.GetRuntimeType(),
		physicalRuntimeInfo.GetName(), physicalRuntimeInfo.GetNamespace())
}

// getRuntime get the current runtime
func (e *ReferenceDatasetEngine) getRuntime() (*datav1alpha1.ThinRuntime, error) {
	key := types.NamespacedName{
		Name:      e.name,
		Namespace: e.namespace,
	}

	var runtime datav1alpha1.ThinRuntime
	if err := e.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}

	return &runtime, nil
}

func (e *ReferenceDatasetEngine) getRuntimeInfo() (base.RuntimeInfoInterface, error) {
	if e.runtimeInfo != nil {
		return e.runtimeInfo, nil
	}

	runtime, err := e.getRuntime()
	if err != nil {
		return e.runtimeInfo, err
	}

	e.runtimeInfo, err = base.BuildRuntimeInfo(e.name, e.namespace, e.runtimeType, runtime.Spec.TieredStore, base.WithMetadataList(base.GetMetadataListFromAnnotation(runtime)))
	if err != nil {
		return e.runtimeInfo, err
	}

	// Setup Fuse Deploy Mode
	e.runtimeInfo.SetupFuseDeployMode(true, runtime.Spec.Fuse.NodeSelector)

	// Ignore the deprecated common labels and PersistentVolumes, use physical runtime

	// Setup with Dataset Info
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			e.Log.Info("Dataset is notfound", "name", e.name, "namespace", e.namespace)
			return e.runtimeInfo, nil
		}

		e.Log.Info("Failed to get dataset when get runtimeInfo")
		return e.runtimeInfo, err
	}

	// set exclusive mode
	// TODO: how to handle the exclusive mode ?
	e.runtimeInfo.SetupWithDataset(dataset)

	e.Log.Info("Setup with dataset done", "exclusive", e.runtimeInfo.IsExclusive())

	return e.runtimeInfo, nil
}

// getPhysicalRuntimeInfo get physicalRuntimeInfo from dataset.
// If could not get dataset, getPhysicalRuntimeInfo try to get physicalRuntimeInfo from runtime status.
func (e *ReferenceDatasetEngine) getPhysicalRuntimeInfo() (base.RuntimeInfoInterface, error) {
	// If already have physicalRuntimeInfo, return it directly
	if e.physicalRuntimeInfo != nil {
		return e.physicalRuntimeInfo, nil
	}

	var physicalNameSpacedNames []types.NamespacedName

	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil && utils.IgnoreNotFound(err) != nil {
		// return if it is not a not-found error
		return e.physicalRuntimeInfo, err
	}

	if dataset != nil {
		// get physicalRuntimeInfo from dataset
		physicalNameSpacedNames = base.GetPhysicalDatasetFromMounts(dataset.Spec.Mounts)
	} else {
		// try to get physicalRuntimeInfo from runtime status
		runtime, err := e.getRuntime()
		if err != nil {
			return e.physicalRuntimeInfo, err
		}
		if len(runtime.Status.Mounts) != 0 {
			physicalNameSpacedNames = base.GetPhysicalDatasetFromMounts(runtime.Status.Mounts)
		}
	}

	if len(physicalNameSpacedNames) == 0 {
		// dataset is nil and len(runtime.Status.Mounts) is 0, return a not-found error
		return e.physicalRuntimeInfo, &k8serrors.StatusError{
			ErrStatus: metav1.Status{
				Reason:  metav1.StatusReasonNotFound,
				Code:    http.StatusNotFound,
				Message: "can't get physical runtime info from either dataset or runtime",
			},
		}
	}
	if len(physicalNameSpacedNames) > 1 {
		return e.physicalRuntimeInfo, fmt.Errorf("ThinEngine with no profile name can only handle dataset only mounting one dataset but get %v", len(physicalNameSpacedNames))
	}
	namespacedName := physicalNameSpacedNames[0]

	physicalRuntimeInfo, err := base.GetRuntimeInfo(e.Client, namespacedName.Name, namespacedName.Namespace)
	if err != nil {
		return e.physicalRuntimeInfo, err
	}

	e.physicalRuntimeInfo = physicalRuntimeInfo

	return e.physicalRuntimeInfo, nil
}
