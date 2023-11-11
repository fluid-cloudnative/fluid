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
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/dataprocess"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (e *AlluxioEngine) generateDataProcessValueFile(ctx cruntime.ReconcileRequestContext, object client.Object) (valueFileName string, err error) {
	dataProcess, ok := object.(*datav1alpha1.DataProcess)
	if !ok {
		err = fmt.Errorf("object %v is not of type DataProcess", object)
		return "", err
	}

	targetDataset, err := utils.GetDataset(e.Client, dataProcess.Spec.Dataset.Name, dataProcess.Spec.Dataset.Namespace)
	if err != nil {
		return "", errors.Wrap(err, "failed to get dataset")
	}

	return dataprocess.GenDataProcessValueFile(targetDataset, dataProcess)
}
