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
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/errors"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (j *JuiceFSEngine) GetDataOperationValueFile(ctx cruntime.ReconcileRequestContext, operation dataoperation.OperationInterface) (valueFileName string, err error) {
	operationType := operation.GetOperationType()
	object := operation.GetOperationObject()

	switch operationType {
	case datav1alpha1.DataMigrateType:
		valueFileName, err = j.generateDataMigrateValueFile(ctx, object)
		return valueFileName, err
	case datav1alpha1.DataLoadType:
		valueFileName, err = j.generateDataLoadValueFile(ctx, object)
		return valueFileName, err
	case datav1alpha1.DataProcessType:
		valueFileName, err = j.generateDataProcessValueFile(ctx, object)
		return valueFileName, err
	default:
		return "", errors.NewNotSupported(
			schema.GroupResource{
				Group:    object.GetObjectKind().GroupVersionKind().Group,
				Resource: object.GetObjectKind().GroupVersionKind().Kind,
			}, "JuiceFSRuntime")
	}
}
