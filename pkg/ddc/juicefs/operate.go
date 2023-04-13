/*
Copyright 2022 The Fluid Authors.

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

package juicefs

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/errors"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (j *JuiceFSEngine) GetDataOperationValueFile(ctx cruntime.ReconcileRequestContext, object client.Object,
	operation dataoperation.OperationInterface) (valueFileName string, err error) {
	operationType := operation.GetOperationType()
	dataMigrate, ok := object.(*v1alpha1.DataMigrate)
	if !ok {
		return "", fmt.Errorf("object %v is not a DataMigrate", object)
	}

	if operationType == dataoperation.DataMigrate {
		valueFileName, err = j.generateDataMigrateValueFile(ctx, *dataMigrate)
		return valueFileName, err
	}

	return "", errors.NewNotSupported(
		schema.GroupResource{
			Group:    object.GetObjectKind().GroupVersionKind().Group,
			Resource: object.GetObjectKind().GroupVersionKind().Kind,
		}, "JuiceFSRuntime")
}
