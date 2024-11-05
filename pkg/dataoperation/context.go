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
