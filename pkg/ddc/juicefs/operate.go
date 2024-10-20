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
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/errors"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (j *JuiceFSEngine) GetDataOperationValueFile(ctx cruntime.ReconcileRequestContext, operation dataoperation.OperationInterface) (valueFileName string, err error) {
	operationType := operation.GetOperationType()
	object := operation.GetOperationObject()

	switch operationType {
	case dataoperation.DataMigrateType:
		valueFileName, err = j.generateDataMigrateValueFile(ctx, object)
		return valueFileName, err
	case dataoperation.DataLoadType:
		valueFileName, err = j.generateDataLoadValueFile(ctx, object)
		return valueFileName, err
	case dataoperation.DataProcessType:
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
