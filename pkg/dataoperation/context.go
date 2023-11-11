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

package dataoperation

import (
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// type OperationType string

// const (
// 	DataLoad    OperationType = "DataLoad"
// 	DataBackup  OperationType = "DataBackup"
// 	DataMigrate OperationType = "DataMigrate"
// 	DataProcess OperationType = "DataProcess"
// )

// ReconcileRequestContext loads or applys the configuration state of a service.
type ReconcileRequestContext struct {
	// used for create engine
	cruntime.ReconcileRequestContext

	// object for dataset operation
	DataObject          client.Object
	OpStatus            *v1alpha1.OperationStatus
	DataOpFinalizerName string
}
